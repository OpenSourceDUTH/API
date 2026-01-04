package schedule

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	schedule := rg.Group("/schedule")
	{
		schedule.POST("/foods", h.PostFood)
		schedule.POST("/versions", h.PostVersion)
		schedule.POST("/items", h.PostSchedule)
		schedule.POST("/announcements", h.PostAnnouncement)
	}
}
