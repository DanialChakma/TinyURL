package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.mod/app"
	"go.mod/routes"
)

// @title Scalable Multi-Tenant URL Shortener API
// @version 1.0
// @description A high-performance URL shortening service built with Go (Gin framework), Redis cache, and MongoDB. Supports millions of short URLs, fast lookups with Redis TTL caching, multi-tenancy for multiple organizations, and both stateful and stateless short URLs.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	container := app.Bootstrap()
	// Read PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // fallback
	}

	router := gin.Default()

	// Register all routes
	routes.RegisterRoutes(router, container)

	log.Printf("Server running on %s\n", port)
	router.Run(port)
}
