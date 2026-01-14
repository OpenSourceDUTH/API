package auth

import (
	"net/http"
	"strconv"

	"API/internal/common"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin-only endpoints
type AdminHandler struct {
	repo       *Repository
	tokenStore *TokenStore
	features   *FeatureRegistry
	quota      *QuotaEngine
	usage      *UsageTracker
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	repo *Repository,
	tokenStore *TokenStore,
	features *FeatureRegistry,
	quota *QuotaEngine,
	usage *UsageTracker,
) *AdminHandler {
	return &AdminHandler{
		repo:       repo,
		tokenStore: tokenStore,
		features:   features,
		quota:      quota,
		usage:      usage,
	}
}

// --- Group Management ---

// ListGroups returns all groups
// GET /admin/groups
func (h *AdminHandler) ListGroups(c *gin.Context) {
	groups, err := h.repo.GetAllGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list groups"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"groups": groups,
	}))
}

// GetGroup returns a group by ID
// GET /admin/groups/:id
func (h *AdminHandler) GetGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid group ID"}))
		return
	}

	group, err := h.repo.GetGroupByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to get group"}))
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, common.CreateErrorResponse([]string{"group not found"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"group": group,
	}))
}

// CreateGroup creates a new group
// POST /admin/groups
func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var req GroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	group, err := h.repo.CreateGroup(req.Name, req.DefaultRPM, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{
		"group": group,
	}))
}

// UpdateGroup updates a group
// PATCH /admin/groups/:id
func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid group ID"}))
		return
	}

	var req GroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	if err := h.repo.UpdateGroup(id, req.Name, req.DefaultRPM, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to update group"}))
		return
	}

	group, _ := h.repo.GetGroupByID(id)
	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"group": group,
	}))
}

// DeleteGroup deletes a group
// DELETE /admin/groups/:id
func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid group ID"}))
		return
	}

	if err := h.repo.DeleteGroup(id); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to delete group"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "group deleted",
	}))
}

// GetGroupQuotas returns quotas for a group
// GET /admin/groups/:id/quotas
func (h *AdminHandler) GetGroupQuotas(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid group ID"}))
		return
	}

	quotas, err := h.quota.GetGroupFeatureQuotas(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to get quotas"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"quotas": quotas,
	}))
}

// SetGroupQuotas sets quotas for a group
// PUT /admin/groups/:id/quotas
func (h *AdminHandler) SetGroupQuotas(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid group ID"}))
		return
	}

	var req QuotaSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	if err := h.quota.BulkSetGroupFeatureQuotas(id, req.Quotas); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to set quotas"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "quotas updated",
	}))
}

// --- Feature Management ---

// ListFeatures returns all features
// GET /admin/features
func (h *AdminHandler) ListFeatures(c *gin.Context) {
	features, err := h.features.GetAllFeatures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list features"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"features": features,
	}))
}

// GetFeature returns a feature by ID
// GET /admin/features/:id
func (h *AdminHandler) GetFeature(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid feature ID"}))
		return
	}

	feature, err := h.features.GetFeatureByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to get feature"}))
		return
	}
	if feature == nil {
		c.JSON(http.StatusNotFound, common.CreateErrorResponse([]string{"feature not found"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"feature": feature,
	}))
}

// CreateFeature creates a new feature
// POST /admin/features
func (h *AdminHandler) CreateFeature(c *gin.Context) {
	var req FeatureCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	feature, err := h.features.CreateFeature(req.Slug, req.Name, req.ParentID, req.AdminOnly)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{
		"feature": feature,
	}))
}

// UpdateFeature updates a feature
// PATCH /admin/features/:id
func (h *AdminHandler) UpdateFeature(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid feature ID"}))
		return
	}

	var req FeatureUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	if err := h.features.UpdateFeature(id, req.Name, req.ParentID, req.AdminOnly); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to update feature"}))
		return
	}

	feature, _ := h.features.GetFeatureByID(id)
	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"feature": feature,
	}))
}

// DeleteFeature deletes a feature
// DELETE /admin/features/:id
func (h *AdminHandler) DeleteFeature(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid feature ID"}))
		return
	}

	if err := h.features.DeleteFeature(id); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to delete feature"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "feature deleted",
	}))
}

// --- Academic Domain Management ---

// ListAcademicDomains returns all academic domains
// GET /admin/academic-domains
func (h *AdminHandler) ListAcademicDomains(c *gin.Context) {
	domains, err := h.repo.GetAllAcademicDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list domains"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"domains": domains,
	}))
}

// AddAcademicDomain adds an academic domain
// POST /admin/academic-domains
func (h *AdminHandler) AddAcademicDomain(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	if err := h.repo.AddAcademicDomain(req.Domain); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to add domain"}))
		return
	}

	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{
		"message": "domain added",
	}))
}

// RemoveAcademicDomain removes an academic domain
// DELETE /admin/academic-domains/:domain
func (h *AdminHandler) RemoveAcademicDomain(c *gin.Context) {
	domain := c.Param("domain")

	if err := h.repo.RemoveAcademicDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to remove domain"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "domain removed",
	}))
}

// --- User Management ---

// ListUsers returns all users with pagination
// GET /admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	users, err := h.repo.GetAllUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list users"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	}))
}

// GetUser returns a user by ID
// GET /admin/users/:id
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid user ID"}))
		return
	}

	user, err := h.repo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to get user"}))
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, common.CreateErrorResponse([]string{"user not found"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"user": user,
	}))
}

// UpdateUser updates a user
// PATCH /admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid user ID"}))
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	if err := h.repo.UpdateUser(id, req.Role, req.Status, req.GroupID, req.MaxTokens); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to update user"}))
		return
	}

	user, _ := h.repo.GetUserByID(id)
	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"user": user,
	}))
}

// GetUserQuotas returns quota overrides for a user
// GET /admin/users/:id/quotas
func (h *AdminHandler) GetUserQuotas(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid user ID"}))
		return
	}

	overrides, err := h.quota.GetUserQuotaOverrides(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to get quotas"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"overrides": overrides,
	}))
}

// SetUserQuotas sets quota overrides for a user
// PUT /admin/users/:id/quotas
func (h *AdminHandler) SetUserQuotas(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid user ID"}))
		return
	}

	var req QuotaSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	if err := h.quota.BulkSetUserQuotaOverrides(id, req.Quotas); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to set quotas"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "quotas updated",
	}))
}

// GetUserUsage returns usage statistics for a user
// GET /admin/users/:id/usage
func (h *AdminHandler) GetUserUsage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid user ID"}))
		return
	}

	stats, err := h.usage.GetUsageStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to get usage"}))
		return
	}

	totalRPM, _ := h.usage.GetUserTotalRPM(id)

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"totalRpm":  totalRPM,
		"byFeature": stats,
	}))
}

// --- Token Management ---

// CreateUserToken creates a token for a user (admin)
// POST /admin/users/:id/tokens
func (h *AdminHandler) CreateUserToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"Invalid user ID"}))
		return
	}

	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	// Admin-created tokens can have any features
	token, err := h.tokenStore.CreateAdminToken(id, req.Label, req.Features, req.AllowedIPs, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{
		"token":   token.RawToken,
		"details": token.Token,
		"message": "Admin token created. Save this token now - it will not be shown again.",
	}))
}

// ListUserTokens returns all tokens for a user (admin)
// GET /admin/users/:id/tokens
func (h *AdminHandler) ListUserTokens(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid user ID"}))
		return
	}

	tokens, err := h.tokenStore.ListUserTokens(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{"failed to list tokens"}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"tokens": tokens,
	}))
}

// RevokeToken revokes any token (admin)
// DELETE /admin/tokens/:id
func (h *AdminHandler) RevokeToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"invalid token ID"}))
		return
	}

	if err := h.tokenStore.AdminRevokeToken(id); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(gin.H{
		"message": "token revoked",
	}))
}
