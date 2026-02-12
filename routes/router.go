package routes

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.mod/app"
)

func RegisterRoutes(router *gin.Engine, container *app.Container) {
	// Read API version from env
	apiPrefix := os.Getenv("API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api/v1"
	}

	// Register Swagger directly on root router
	RegisterSwaggerRoute(router)
	RegisterURLRoutes(router, container)

	// Global API group
	api := router.Group(apiPrefix)

	RegisterAuthRoutes(api, container)
	RegisterTenantRoutes(api, container)
}
