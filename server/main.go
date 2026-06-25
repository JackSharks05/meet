// Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
//
// Changes in this fork:
//   - Two run modes via the MODE env var ("public" vs "admin"). The public
//     listener exposes only the respondent-facing API (auth status + events);
//     the admin listener exposes the full API (user/profile, analytics,
//     folders, calendar OAuth, swagger). See PLAN.md for the architecture.
//   - CORS allowed origins are configurable via CORS_ORIGINS (comma-separated).
//   - The listen port is configurable via PORT (default 3002).
//   - Removed Stripe/billing and the Slack command route.
//   - Removed serving the built frontend: the frontend is deployed separately
//     (Vercel for the public site, Tailscale Serve for admin), so this process
//     is an API server only.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"schej.it/server/db"
	"schej.it/server/logger"
	"schej.it/server/routes"
	"schej.it/server/services/gcloud"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "schej.it/server/docs"
)

// @title Schej.it API
// @version 1.0
// @description This is the API for Schej.it!

// @host localhost:3002/api

func main() {
	// Set release flag
	release := flag.Bool("release", false, "Whether this is the release version of the server")
	flag.Parse()
	if *release {
		os.Setenv("GIN_MODE", "release")
		gin.SetMode(gin.ReleaseMode)
	} else {
		os.Setenv("GIN_MODE", "debug")
	}

	// Init logfile
	logFile, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	// Init logger
	logger.Init(logFile)

	// Load .env variables
	loadDotEnv()

	// Determine run mode: "public" (respondent-facing only) or "admin" (full).
	mode := strings.ToLower(os.Getenv("MODE"))
	if mode == "" {
		mode = "admin"
	}
	isAdmin := mode == "admin"

	// Init router
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())

	// Cors — allowed origins come from CORS_ORIGINS (comma-separated). Defaults
	// cover local development.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     getCorsOrigins(),
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Init database
	closeConnection := db.Init()
	defer closeConnection()

	// Init google cloud stuff (no-op without a service-account key)
	closeTasks := gcloud.InitTasks()
	defer closeTasks()

	// Session
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(sessions.Sessions("session", store))

	// Init routes
	apiRouter := router.Group("/api")

	// Routes available to everyone (public respondent flow + admin). Each
	// function registers only its respondent-facing subset when admin is false.
	routes.InitAuth(apiRouter, isAdmin)
	routes.InitEvents(apiRouter, isAdmin)

	// Admin-only routes (account management, analytics, folders, API docs)
	if isAdmin {
		routes.InitUser(apiRouter)
		routes.InitAnalytics(apiRouter)
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	// API-only server: anything unmatched is a 404 (frontend is served elsewhere)
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	logger.StdErr.Printf("Starting meet server in %q mode\n", mode)

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	router.Run(fmt.Sprintf(":%s", port))
}

// getCorsOrigins returns the list of allowed CORS origins from CORS_ORIGINS
// (comma-separated), falling back to local development origins.
func getCorsOrigins() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:8080", "http://localhost:5173"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// Load .env variables
func loadDotEnv() {
	// In containerized/production runs, env vars are injected directly and there
	// is no .env file — that's expected, so a missing file is not fatal.
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found; using environment variables")
	}

	// Validate session secret
	validateSessionSecret()
}

// validateSessionSecret ensures SESSION_SECRET is set and meets security requirements
func validateSessionSecret() {
	secret := os.Getenv("SESSION_SECRET")

	if secret == "" {
		logger.StdErr.Panicln("SESSION_SECRET environment variable is required but not set")
	}

	// Minimum 32 characters for adequate security (256 bits)
	if len(secret) < 32 {
		logger.StdErr.Panicln("SESSION_SECRET must be at least 32 characters long")
	}
}
