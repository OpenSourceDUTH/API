package main

import (
	"API/internal/common"
	"API/internal/v0/schedule"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := sql.Open("sqlite3", "./internal/databases/schedule.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Initialize Logic Layers
	schedRepo := schedule.NewRepository(db)
	schedHandler := schedule.NewHandler(schedRepo)

	router := gin.Default()

	// Groups
	global := router.Group("/api")
	common.RegisterRoutes(global)

	v0Group := router.Group("/api/v0")
	{
		// Pass the v0Group and the initialized handler
		schedule.RegisterRoutes(v0Group, schedHandler)
	}

	router.StaticFile("/favicon.ico", "./internal/assets/logo.svg")
	err = router.Run(":9237")
	if err != nil {
		return
	}
}

/*
This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team as well as helper endpoints to integrate with our apps.
API Copyright (C) 2025 OpenSourceDUTH
    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
