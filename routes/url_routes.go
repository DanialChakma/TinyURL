package routes

import (
	"github.com/gin-gonic/gin"
	"go.mod/app"
	"go.mod/controllers"
	"go.mod/repo"
	"go.mod/services"
)

/*Following func is just a wrapper for needed for Swagger documentation generation */

// RedirectShortURL godoc
// @Summary Redirect to original URL
// @Description Redirects to the original long URL using the provided short code.
// @Tags URL
// @Param code path string true "Short code (Base62 alphanumeric)"
// @Success 302 {string} string "Redirects to original URL"
// @Failure 400 {object} controllers.ErrorResponse "Invalid short code"
// @Failure 404 {object} controllers.ErrorResponse "URL not found"
// @Router /links/{code} [get]
func RedirectShortURL(container *app.Container) gin.HandlerFunc {
	urlController := controllers.NewURLController(
		services.NewURLService(
			repo.NewURLRepository(container.DB),
			repo.NewTenantRepository(container.DB),
			container.Cache,
			container.IDGen,
		),
	)

	return func(c *gin.Context) {
		urlController.Redirect(c)
	}
}

func RegisterURLRoutes(router *gin.Engine, container *app.Container) {

	urlRepo := repo.NewURLRepository(container.DB)
	tenantRepo := repo.NewTenantRepository(container.DB)
	urlService := services.NewURLService(
		urlRepo,
		tenantRepo,
		container.Cache,
		container.IDGen,
	)

	urlController := controllers.NewURLController(urlService)

	router.POST("/links", urlController.CreateShortURL)
	// router.GET("/links/:code", urlController.Redirect)
	router.GET("/links/:code", RedirectShortURL(container))
	router.POST("/stateless", urlController.CreateShortURLStateless)
	router.GET("/stateless/:code", urlController.RedirectStateless)

	service := services.NewSecureURLService()
	controller := controllers.NewSecureURLController(service)

	router.POST("/secure", controller.CreateSecureShortURL)
	router.GET("/secure_redirect/:code", controller.SecureRedirectTo)

}
