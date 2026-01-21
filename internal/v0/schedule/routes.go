package schedule

import (
	"API/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware *auth.Middleware) {
	schedule := rg.Group("/schedule")
	{
		schedule.GET("", authMiddleware.RequireToken("schedule"), h.GetSchedule)
	}

	schedule_admin := rg.Group("/admin")
	schedule_admin.Use(authMiddleware.RequireSession())
	schedule_admin.Use(authMiddleware.RequireRole(auth.RoleAdmin))
	{
		schedule_admin.POST("/foods", h.PostFood)
		schedule_admin.POST("/versions", h.PostVersion)
		schedule_admin.POST("/items", h.PostSchedule)
		schedule_admin.POST("/announcements", h.PostAnnouncement)
	}
}
