package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/app"
	"go.mod/controllers"
	"go.mod/initializers"
	"go.mod/repo"
	"go.mod/services"
)

func RegisterAuthRoutes(router *gin.RouterGroup, container *app.Container) {

	// --------- Initialize Auth Dependencies ---------
	// os.Getenv("JWT_SECRET_KEY")
	authRepo := repo.NewAuthRepository(container.DB)
	tokenService := services.NewTokenService(initializers.JwtKey, initializers.JwtRefreshKey)
	authService := services.NewAuthService(authRepo, tokenService)
	authController := controllers.NewAuthController(authService)

	// --------- Public Routes ---------
	router.POST("/register", authController.Register)
	router.POST("/login", authController.Login)
	router.POST("/refresh", authController.Refresh)

	// --------- Protected Routes ---------
	protected := router.Group("/protected")
	// protected.Use(authController.AuthMiddleware()) // JWT middleware
	protected.GET("/hello", func(c *gin.Context) {
		username := c.GetString("username")
		role := c.GetString("role")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello " + username,
			"role":    role,
		})
	})
}
