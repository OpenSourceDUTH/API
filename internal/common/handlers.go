package common

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type StatusResponse struct {
	InternalServerLatency string `json:"internal_server_latency"`
	Uptime                string `json:"uptime"`
}

// Uptime Logic
var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

// Ping Logic
func ping() time.Duration {
	start := time.Now()
	duration := time.Since(start)
	return duration
}

func Status(c *gin.Context) {
	data := StatusResponse{
		InternalServerLatency: ping().String(),
		Uptime:                uptime().Truncate(time.Second).String(),
	}
	response := CreateSuccessResponse(data)
	c.JSON(http.StatusOK, response)
}

//This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team as well as helper endpoints to integrate with our apps.
//API Copyright (C) 2025 OpenSourceDUTH
//This program is free software: you can redistribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.
//
//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.
//
//You should have received a copy of the GNU General Public License
//along with this program.  If not, see <https://www.gnu.org/licenses/>.
