package main

import (
	"API/internal/auth"
	"API/internal/common"
	"API/internal/env"
	"API/internal/v0/schedule"
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Schedule database
	scheduleDB, err := sql.Open("sqlite3", "./internal/databases/schedule.db")
	if err != nil {
		log.Fatal(err)
	}
	defer scheduleDB.Close()

	// Auth database
	authDB, err := sql.Open("sqlite3", "./internal/databases/auth.db")
	if err != nil {
		log.Fatal(err)
	}
	defer authDB.Close()

	// Enable WAL mode for auth database (better concurrent performance)
	if _, err := authDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", err)
	}

	// Initialize schedule components
	schedRepo := schedule.NewRepository(scheduleDB)
	schedHandler := schedule.NewHandler(schedRepo)

	// Initialize auth components
	authRepo := auth.NewRepository(authDB)

	// OAuth configuration
	oauthConfig := auth.NewOAuthConfig(
		auth.ProviderConfig{
			ClientID:     env.GetEnv(env.EnvGoogleClientID, ""),
			ClientSecret: env.GetEnv(env.EnvGoogleClientSecret, ""),
		},
		auth.ProviderConfig{
			ClientID:     env.GetEnv(env.EnvGitHubClientID, ""),
			ClientSecret: env.GetEnv(env.EnvGitHubClientSecret, ""),
		},
		env.GetEnv(env.EnvAuthCallbackBaseURL, "http://localhost:9237"),
	)

	// Auth stores
	stateStore := auth.NewOAuthStateStore(authRepo)
	sessionStore := auth.NewSessionStore(
		authRepo,
		env.GetDuration(env.EnvSessionDuration, 7*24*time.Hour),
		env.GetBool(env.EnvSecureCookies, false),
	)
	featureRegistry := auth.NewFeatureRegistry(authRepo)
	tokenStore := auth.NewTokenStore(authRepo, featureRegistry)
	quotaEngine := auth.NewQuotaEngine(authRepo, featureRegistry)
	usageTracker := auth.NewUsageTracker(authRepo, stateStore, sessionStore)

	// Start usage tracker background goroutines
	usageTracker.Start(ctx)

	// Auth handlers
	authHandler := auth.NewHandler(
		authRepo,
		oauthConfig,
		stateStore,
		sessionStore,
		tokenStore,
		featureRegistry,
	)
	adminHandler := auth.NewAdminHandler(
		authRepo,
		tokenStore,
		featureRegistry,
		quotaEngine,
		usageTracker,
	)
	authMiddleware := auth.NewMiddleware(
		tokenStore,
		sessionStore,
		featureRegistry,
		quotaEngine,
		usageTracker,
	)

	router := gin.Default()

	// Global routes
	global := router.Group("/api")
	common.RegisterRoutes(global)

	// Auth routes (public + session-protected + admin)
	auth.RegisterRoutes(global, authHandler, adminHandler, authMiddleware)

	// v0 API routes
	v0Group := router.Group("/api/v0")
	{
		// Schedule routes (protected by token)
		schedule.RegisterRoutes(v0Group, schedHandler, authMiddleware)
	}

	router.StaticFile("/favicon.ico", "./internal/assets/logo.svg")

	// Graceful shutdown handling
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		usageTracker.Stop()
	}()

	err = router.Run(":9237")
	if err != nil {
		log.Fatal(err)
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
