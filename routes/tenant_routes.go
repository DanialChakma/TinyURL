package routes

import (
	"github.com/gin-gonic/gin"
	"go.mod/app"
	"go.mod/auth"
	"go.mod/controllers"
	"go.mod/initializers"
	"go.mod/repo"
	"go.mod/services"
)

func RegisterTenantRoutes(router *gin.RouterGroup, container *app.Container) {

	// Create Repository
	tenantRepo := repo.NewTenantRepository(container.DB)

	// Inject Repository into Service
	tenantService := services.NewTenantService(tenantRepo)

	// Inject Service into Controller
	tenantController := controllers.NewTenantController(tenantService)

	authRepo := repo.NewAuthRepository(container.DB)
	tokenService := services.NewTokenService(initializers.JwtKey, initializers.JwtRefreshKey)
	authService := services.NewAuthService(authRepo, tokenService)
	jwtAuth := auth.AuthMiddleware(authService, tokenService)
	tenantGroup := router.Group("/tenants")
	tenantGroup.Use(jwtAuth)
	{
		tenantGroup.POST("/", tenantController.CreateTenant)
		tenantGroup.GET("/", tenantController.GetAllTenants)
		tenantGroup.GET("/:id", tenantController.GetTenant)
		tenantGroup.PUT("/:id", tenantController.UpdateTenant)
		tenantGroup.DELETE("/:id", tenantController.DeleteTenant)
	}
}
