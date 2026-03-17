package auth

import (
	"context"
	"net/http"
	"strings"

	"API/internal/common"

	"github.com/gin-gonic/gin"
)

const (
	OAuthStateCookieName = "osduth_oauth_state"
)

// Handler handles authentication endpoints
type Handler struct {
	repo         *Repository
	oauthConfig  *OAuthConfig
	stateStore   *OAuthStateStore
	sessionStore *SessionStore
	tokenStore   *TokenStore
	features     *FeatureRegistry
	frontendURL  string
}

// NewHandler creates a new auth handler
func NewHandler(
	repo *Repository,
	oauthConfig *OAuthConfig,
	stateStore *OAuthStateStore,
	sessionStore *SessionStore,
	tokenStore *TokenStore,
	features *FeatureRegistry,
	frontendURL string,
) *Handler {
	return &Handler{
		repo:         repo,
		oauthConfig:  oauthConfig,
		stateStore:   stateStore,
		sessionStore: sessionStore,
		tokenStore:   tokenStore,
		features:     features,
		frontendURL:  frontendURL,
	}
}

// redirect sends the user to the frontend with optional query params
func (h *Handler) redirect(c *gin.Context, path string, params map[string]string) {
	url := h.frontendURL + path
	if len(params) > 0 {
		q := make([]string, 0, len(params))
		for k, v := range params {
			q = append(q, k+"="+v)
		}
		url += "?" + strings.Join(q, "&")
	}
	c.Redirect(http.StatusFound, url)
}

// Login initiates OAuth flow
// GET /auth/login/:provider?redirect_uri=<path>
func (h *Handler) Login(c *gin.Context) {
	provider := Provider(c.Param("provider"))
	errRedirect := func(msg string) { h.redirect(c, "/auth/error", map[string]string{"error": msg}) }

	if provider != ProviderGoogle && provider != ProviderGitHub {
		errRedirect("unsupported_provider")
		return
	}
	if !h.oauthConfig.IsProviderConfigured(provider) {
		errRedirect("provider_not_configured")
		return
	}

	state, err := h.stateStore.CreateState()
	if err != nil {
		errRedirect("state_failed")
		return
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = "/"
	}

	c.SetSameSite(http.SameSiteLaxMode)
	maxAge := int(OAuthStateExpiry.Seconds())
	secure := h.sessionStore.secureCookie
	c.SetCookie("osduth_auth_redirect", redirectURI, maxAge, "/", "", secure, true)
	c.SetCookie(OAuthStateCookieName, state, maxAge, "/", "", secure, true)

	authURL, err := h.oauthConfig.GetAuthURL(provider, state)
	if err != nil {
		errRedirect("auth_url_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback handles OAuth callback
// GET /auth/callback/:provider
func (h *Handler) Callback(c *gin.Context) {
	provider := Provider(c.Param("provider"))
	secure := h.sessionStore.secureCookie

	// Helper: clear auth cookies and redirect with error
	authError := func(code string) {
		c.SetCookie("osduth_auth_redirect", "", -1, "/", "", secure, true)
		c.SetCookie(OAuthStateCookieName, "", -1, "/", "", secure, true)
		redirectPath, _ := c.Cookie("osduth_auth_redirect")
		if redirectPath == "" {
			redirectPath = "/"
		}
		h.redirect(c, redirectPath, map[string]string{"error": code})
	}

	if provider != ProviderGoogle && provider != ProviderGitHub {
		authError("unsupported_provider")
		return
	}

	queryState := c.Query("state")
	cookieState, err := c.Cookie(OAuthStateCookieName)
	if err != nil || cookieState == "" || queryState != cookieState {
		authError("invalid_state")
		return
	}

	valid, err := h.stateStore.ValidateState(queryState)
	if err != nil || !valid {
		authError("expired_state")
		return
	}

	c.SetCookie(OAuthStateCookieName, "", -1, "/", "", secure, true)

	if errMsg := c.Query("error"); errMsg != "" {
		authError("oauth_denied")
		return
	}

	code := c.Query("code")
	if code == "" {
		authError("missing_code")
		return
	}

	ctx := context.Background()
	token, err := h.oauthConfig.ExchangeCode(ctx, provider, code)
	if err != nil {
		authError("token_exchange_failed")
		return
	}

	userInfo, err := h.oauthConfig.GetUserInfo(ctx, provider, token)
	if err != nil {
		authError("userinfo_failed")
		return
	}

	user, err := h.findOrCreateUser(userInfo, provider, token.AccessToken, token.RefreshToken)
	if err != nil {
		authError("user_creation_failed")
		return
	}

	if user.Status != StatusActive {
		authError("account_" + string(user.Status))
		return
	}

	session, err := h.sessionStore.CreateSession(user.ID)
	if err != nil {
		authError("session_failed")
		return
	}

	h.sessionStore.SetSessionCookie(c, session.ID)

	redirectPath, _ := c.Cookie("osduth_auth_redirect")
	if redirectPath == "" {
		redirectPath = "/"
	}
	c.SetCookie("osduth_auth_redirect", "", -1, "/", "", secure, true)

	h.redirect(c, redirectPath, map[string]string{"auth": "success"})
}

func (h *Handler) findOrCreateUser(info *OAuthUserInfo, provider Provider, accessToken, refreshToken string) (*User, error) {
	// Check if OAuth identity exists
	identity, err := h.repo.GetOAuthIdentity(provider, info.ProviderID)
	if err != nil {
		return nil, err
	}

	if identity != nil {
		// Update tokens
		err := h.repo.UpdateOAuthIdentityTokens(identity.ID, accessToken, refreshToken)
		if err != nil {
			return nil, err
		}
		return h.repo.GetUserByID(identity.UserID)
	}

	// Check if user exists by email
	user, err := h.repo.GetUserByEmail(info.Email)
	if err != nil {
		return nil, err
	}

	if user != nil {
		// Link new OAuth identity to existing user
		_, err = h.repo.CreateOAuthIdentity(user.ID, provider, info.ProviderID, accessToken, refreshToken)
		if err != nil {
			return nil, err
		}
		return h.repo.GetUserByID(user.ID)
	}

	// Create new user
	// Determine group based on email domain
	groupID, err := h.determineGroupForEmail(info.Email)
	if err != nil {
		return nil, err
	}

	user, err = h.repo.CreateUser(info.Email, info.DisplayName, groupID)
	if err != nil {
		return nil, err
	}

	// Create OAuth identity
	_, err = h.repo.CreateOAuthIdentity(user.ID, provider, info.ProviderID, accessToken, refreshToken)
	if err != nil {
		return nil, err
	}

	return h.repo.GetUserByID(user.ID)
}

func (h *Handler) determineGroupForEmail(email string) (int64, error) {
	// Extract domain from email
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		// Default to regular group
		group, err := h.repo.GetGroupByName("regular")
		if err != nil || group == nil {
			return 1, nil // Fallback to ID 1
		}
		return group.ID, nil
	}

	domain := strings.ToLower(parts[1])

	// Check if domain is academic
	isAcademic, err := h.repo.IsAcademicDomain(domain)
	if err != nil {
		return 1, nil
	}

	if isAcademic {
		group, err := h.repo.GetGroupByName("academic")
		if err != nil || group == nil {
			return 1, nil
		}
		return group.ID, nil
	}

	// Default to regular group
	group, err := h.repo.GetGroupByName("regular")
	if err != nil || group == nil {
		return 1, nil
	}
	return group.ID, nil
}

// Me returns the current authenticated user
// GET /auth/me
func (h *Handler) Me(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, common.CreateErrorResponse([]string{"not authenticated"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"role":        user.Role,
			"status":      user.Status,
			"group":       user.Group,
			"maxTokens":   user.MaxTokens,
			"createdAt":   user.CreatedAt,
		},
	}))
}

// Logout logs out the current user
// GET /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	sessionID, _ := h.sessionStore.GetSessionFromCookie(c)
	if sessionID != "" {
		h.sessionStore.DeleteSession(sessionID)
	}
	h.sessionStore.ClearSessionCookie(c)

	// API clients get JSON, browsers get redirect
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{"message": "logged out"}))
	} else {
		h.redirect(c, "/", map[string]string{"auth": "logged_out"})
	}
}

// ListTokens returns all tokens for the current user
// GET /auth/tokens
func (h *Handler) ListTokens(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, common.CreateErrorResponse([]string{"not authenticated"}))
		return
	}

	tokens, err := h.tokenStore.ListUserTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list tokens"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"tokens": tokens,
	}))
}

// ListAssignableFeatures returns features that users can assign to their tokens
// GET /auth/tokens/features
func (h *Handler) ListAssignableFeatures(c *gin.Context) {
	features, err := h.features.GetUserAssignableFeatures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list features"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"features": features,
	}))
}

// CreateToken creates a new token for the current user
// POST /auth/tokens
func (h *Handler) CreateToken(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, common.CreateErrorResponse([]string{"not authenticated"}))
		return
	}

	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	token, err := h.tokenStore.CreateUserToken(user.ID, req.Label, req.Features, req.AllowedIPs, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{
		"token":   token.RawToken,
		"details": token.Token,
		"message": "Token created. Save this token now - it will not be shown again.",
	}))
}

// RevokeToken revokes a token owned by the current user
// DELETE /auth/tokens/:id
func (h *Handler) RevokeToken(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, common.CreateErrorResponse([]string{"Not authenticated"}))
		return
	}

	tokenIDStr := c.Param("id")

	// Parse token ID
	tokenID, err := parseID(tokenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"Invalid token ID"}))
		return
	}

	if err := h.tokenStore.RevokeToken(tokenID, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "Token revoked successfully",
	}))
}

// parseID parses a string ID to int64
func parseID(s string) (int64, error) {
	var id int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{s}
		}
		id = id*10 + int64(c-'0')
	}
	return id, nil
}

type parseError struct {
	s string
}

func (e *parseError) Error() string {
	return "Invalid ID: " + e.s
}
