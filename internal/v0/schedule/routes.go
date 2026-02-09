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

//   This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
//   API Copyright (C) 2025 OpenSourceDUTH
//       This program is free software: you can redistribute it and/or modify
//       it under the terms of the GNU General Public License as published by
//       the Free Software Foundation, either version 3 of the License, or
//       (at your option) any later version.

//       This program is distributed in the hope that it will be useful,
//       but WITHOUT ANY WARRANTY; without even the implied warranty of
//       MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//       GNU General Public License for more details.

//       You should have received a copy of the GNU General Public License
//       along with this program.  If not, see <https://www.gnu.org/licenses/>.
