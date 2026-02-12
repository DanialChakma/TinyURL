package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/initializers"
	"go.mod/services"
)

type SecureURLController struct {
	service *services.SecureURLService
}

func NewSecureURLController(service *services.SecureURLService) *SecureURLController {
	return &SecureURLController{service: service}
}

func (u *SecureURLController) CreateSecureShortURL(c *gin.Context) {

	var body struct {
		LongURL string `json:"long_url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortCode, err := u.service.Create(body.LongURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url": initializers.AppBaseURL + "/secure_redirect/" + shortCode,
	})
}

func (u *SecureURLController) SecureRedirectTo(c *gin.Context) {

	shortCode := c.Param("code")

	longURL, err := u.service.Resolve(shortCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, longURL)
}
