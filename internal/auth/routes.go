package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all auth-related routes
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	adminHandler *AdminHandler,
	middleware *Middleware,
) {
	auth := router.Group("/auth")
	{
		// Public OAuth routes
		auth.GET("/login/:provider", handler.Login)
		auth.GET("/callback/:provider", handler.Callback)

		// Session-protected routes
		sessionProtected := auth.Group("")
		sessionProtected.Use(middleware.RequireSession())
		{
			sessionProtected.GET("/me", handler.Me)
			sessionProtected.GET("/logout", handler.Logout)

			// Token management
			sessionProtected.GET("/tokens", handler.ListTokens)
			sessionProtected.GET("/tokens/features", handler.ListAssignableFeatures)
			sessionProtected.POST("/tokens", handler.CreateToken)
			sessionProtected.DELETE("/tokens/:id", handler.RevokeToken)
		}
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.RequireSession())
	admin.Use(middleware.RequireRole(RoleAdmin))
	{
		// Group management
		admin.GET("/groups", adminHandler.ListGroups)
		admin.POST("/groups", adminHandler.CreateGroup)
		admin.GET("/groups/:id", adminHandler.GetGroup)
		admin.PATCH("/groups/:id", adminHandler.UpdateGroup)
		admin.DELETE("/groups/:id", adminHandler.DeleteGroup)
		admin.GET("/groups/:id/quotas", adminHandler.GetGroupQuotas)
		admin.PUT("/groups/:id/quotas", adminHandler.SetGroupQuotas)

		// Feature management
		admin.GET("/features", adminHandler.ListFeatures)
		admin.POST("/features", adminHandler.CreateFeature)
		admin.GET("/features/:id", adminHandler.GetFeature)
		admin.PATCH("/features/:id", adminHandler.UpdateFeature)
		admin.DELETE("/features/:id", adminHandler.DeleteFeature)

		// Academic domain management
		admin.GET("/academic-domains", adminHandler.ListAcademicDomains)
		admin.POST("/academic-domains", adminHandler.AddAcademicDomain)
		admin.DELETE("/academic-domains/:domain", adminHandler.RemoveAcademicDomain)

		// User management
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/:id", adminHandler.GetUser)
		admin.PATCH("/users/:id", adminHandler.UpdateUser)
		admin.GET("/users/:id/quotas", adminHandler.GetUserQuotas)
		admin.PUT("/users/:id/quotas", adminHandler.SetUserQuotas)
		admin.GET("/users/:id/usage", adminHandler.GetUserUsage)
		admin.GET("/users/:id/tokens", adminHandler.ListUserTokens)
		admin.POST("/users/:id/tokens", adminHandler.CreateUserToken)

		// Token management (admin)
		admin.DELETE("/tokens/:id", adminHandler.RevokeToken)
	}
}
