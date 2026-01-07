package schedule

import (
	"API/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware *auth.Middleware) {
	schedule := rg.Group("/schedule")
	schedule.Use(authMiddleware.RequireToken("schedule"))
	{
		schedule.POST("/foods", h.PostFood)
		schedule.POST("/versions", h.PostVersion)
		schedule.POST("/items", h.PostSchedule)
		schedule.POST("/announcements", h.PostAnnouncement)
	}
}
