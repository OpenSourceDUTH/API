package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Context keys
	ContextKeyUser  = "auth_user"
	ContextKeyToken = "auth_token"

	// Headers
	HeaderAuthorization      = "Authorization"
	HeaderRateLimitLimit     = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderRateLimitReset     = "X-RateLimit-Reset"
	HeaderRetryAfter         = "Retry-After"
)

// Middleware provides authentication and authorization middleware
type Middleware struct {
	tokenStore   *TokenStore
	sessionStore *SessionStore
	features     *FeatureRegistry
	quota        *QuotaEngine
	usage        *UsageTracker
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(
	tokenStore *TokenStore,
	sessionStore *SessionStore,
	features *FeatureRegistry,
	quota *QuotaEngine,
	usage *UsageTracker,
) *Middleware {
	return &Middleware{
		tokenStore:   tokenStore,
		sessionStore: sessionStore,
		features:     features,
		quota:        quota,
		usage:        usage,
	}
}

// RequireToken returns a middleware that validates bearer tokens and checks quotas
func (m *Middleware) RequireToken(featureSlug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract Authorization header
		authHeader := c.GetHeader(HeaderAuthorization)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// 2. Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}
		rawToken := parts[1]

		// 3. Validate token
		validated, err := m.tokenStore.ValidateToken(rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 4. Get the feature being accessed
		feature, err := m.features.GetFeatureBySlug(featureSlug)
		if err != nil || feature == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "feature not found",
			})
			return
		}

		// 5. Live admin-only check: if feature is admin-only and token is not admin-created, deny
		adminOnly, err := m.features.IsFeatureAdminOnly(feature.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check feature permissions",
			})
			return
		}
		if adminOnly && !validated.Token.AdminCreated {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "this feature requires an admin-issued token",
			})
			return
		}

		// 6. Check if token has access to this feature (including parent features)
		hasAccess, err := m.features.TokenHasFeatureAccess(validated.FeatureIDs, featureSlug)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check feature access",
			})
			return
		}
		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("token does not have access to feature '%s'", featureSlug),
			})
			return
		}

		// 7. Check IP whitelist
		if len(validated.AllowedIPs) > 0 {
			clientIP := c.ClientIP()
			canonicalIP, err := CanonicalizeIP(clientIP)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "invalid client IP",
				})
				return
			}

			if !IsIPAllowed(canonicalIP, validated.AllowedIPs) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "IP address not allowed for this token",
				})
				return
			}
		}

		// 8. Check RPM quota
		effectiveRPM, err := m.quota.GetEffectiveRPM(validated.User.ID, feature.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check quota",
			})
			return
		}

		// If not unlimited, check usage
		if effectiveRPM != UnlimitedRPM {
			currentRPM, err := m.usage.GetFeatureRPM(validated.User.ID, feature.ID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "failed to check usage",
				})
				return
			}

			// Set rate limit headers
			remaining := effectiveRPM - currentRPM - 1 // -1 for this request
			if remaining < 0 {
				remaining = 0
			}
			resetTime := time.Now().Add(60 * time.Second).Unix()

			c.Header(HeaderRateLimitLimit, strconv.Itoa(effectiveRPM))
			c.Header(HeaderRateLimitRemaining, strconv.Itoa(remaining))
			c.Header(HeaderRateLimitReset, strconv.FormatInt(resetTime, 10))

			if currentRPM >= effectiveRPM {
				c.Header(HeaderRetryAfter, "60")
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":      "rate limit exceeded",
					"limit":      effectiveRPM,
					"retryAfter": 60,
				})
				return
			}
		}

		// 9. Record usage (non-blocking)
		m.usage.RecordRequest(validated.User.ID, feature.ID)

		// 10. Set context values
		c.Set(ContextKeyUser, validated.User)
		c.Set(ContextKeyToken, validated.Token)

		c.Next()
	}
}

// RequireSession returns a middleware that validates session cookies
func (m *Middleware) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := m.sessionStore.GetSessionFromCookie(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
			})
			return
		}

		user, err := m.sessionStore.GetUserFromSession(sessionID)
		if err != nil || user == nil {
			m.sessionStore.ClearSessionCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "session expired or invalid",
			})
			return
		}

		// Check user status
		if user.Status != StatusActive {
			m.sessionStore.ClearSessionCookie(c)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("account is %s", user.Status),
			})
			return
		}

		c.Set(ContextKeyUser, user)
		c.Next()
	}
}

// RequireRole returns a middleware that checks if the user has the required role
func (m *Middleware) RequireRole(role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get(ContextKeyUser)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
			})
			return
		}

		user, ok := userVal.(*User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid user context",
			})
			return
		}

		if user.Role != role && user.Role != RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("requires %s role", role),
			})
			return
		}

		c.Next()
	}
}

// OptionalSession attempts to load a session but doesn't fail if none exists
func (m *Middleware) OptionalSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := m.sessionStore.GetSessionFromCookie(c)
		if err != nil {
			c.Next()
			return
		}

		user, err := m.sessionStore.GetUserFromSession(sessionID)
		if err == nil && user != nil && user.Status == StatusActive {
			c.Set(ContextKeyUser, user)
		}

		c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from the context
func GetUserFromContext(c *gin.Context) *User {
	userVal, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil
	}
	user, ok := userVal.(*User)
	if !ok {
		return nil
	}
	return user
}

// GetTokenFromContext retrieves the validated token from the context
func GetTokenFromContext(c *gin.Context) *Token {
	tokenVal, exists := c.Get(ContextKeyToken)
	if !exists {
		return nil
	}
	token, ok := tokenVal.(*Token)
	if !ok {
		return nil
	}
	return token
}
