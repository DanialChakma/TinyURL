package controllers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.mod/services"
)

type URLController struct {
	service *services.URLService
}

func NewURLController(service *services.URLService) *URLController {
	return &URLController{service: service}
}

var base62Regex = regexp.MustCompile(`^[0-9a-zA-Z]{1,200}$`)

type CreateShortURLRequest struct {
	LongURL  string `json:"long_url" example:"https://example.com/very/long/url"`
	TenantID string `json:"tenant_id,omitempty" example:"tenant_123"`
}

type CreateShortURLResponse struct {
	ShortURL string `json:"short_url" example:"https://short.ly/links/abc123"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}

/*
CreateShortURL handles:
POST /shorten

This endpoint:
1. Accepts a long URL in JSON request body
2. Generates a globally unique numeric ID
3. Encodes the ID into a Base62 short code
4. Stores the short_code → long_url mapping in MongoDB
5. Caches the mapping in Redis for fast future redirects
6. Returns the full shortened URL to the client

Request Body:

	{
	"long_url": "https://example.com/very/long/url"
	}

Response:

	{
	"short_url": "http://localhost:8080/abc123X"
	}
*/

// CreateShortURL godoc
// @Summary Create stateful short URL
// @Description Generates a short URL for the provided long URL. Optionally supports multi-tenant optimization.
// @Tags URL
// @Accept json
// @Produce json
// @Param request body CreateShortURLRequest true "Long URL payload"
// @Success 200 {object} CreateShortURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /links [post]
func (u *URLController) CreateShortURL(c *gin.Context) {

	var body struct {
		LongURL  string `json:"long_url" binding:"required,url,max=2048"`
		TenantID string `json:"tenant_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL, err := u.service.Create(c.Request.Context(), body.LongURL, body.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

/*
Redirect handles:
GET /:code
This endpoint:
1. Extracts the short code from the URL path
2. Attempts to resolve it from Redis cache (fast path)
3. If cache miss, queries MongoDB (source of truth)
4. Re-populates Redis cache on DB hit
5. Redirects the client to the original long URL
If the short code does not exist, returns HTTP 404.
Example:
GET /abc123X  → 302 Redirect to https://example.com
*/

// Redirect godoc
// @Summary Redirect to original URL
// @Description Redirects to the original long URL using the provided short code.
// @Tags URL
// @Param code path string true "Short code (Base62 alphanumeric)"
// @Success 302 {string} string "Redirects to original URL"
// @Failure 400 {object} ErrorResponse "Invalid short code"
// @Failure 404 {object} ErrorResponse "URL not found"
// @Router /links/{code} [get]

func (u *URLController) Redirect(c *gin.Context) {

	shortCode := c.Param("code")

	if !base62Regex.MatchString(shortCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short code"})
		return
	}

	longURL, err := u.service.Resolve(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusFound, longURL)
}

func (u *URLController) CreateShortURLStateless(c *gin.Context) {

	var body struct {
		LongURL    string `json:"long_url" binding:"required,url,max=4096"`
		TenantID   string `json:"tenant_id" binding:"omitempty,max=100"`
		TrimDomain bool   `json:"trim_domain"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL, err := u.service.CreateStateless(
		c.Request.Context(),
		body.LongURL,
		body.TenantID,
		body.TrimDomain,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"short_url": shortURL,
	})
}
func (u *URLController) RedirectStateless(c *gin.Context) {
	shortCode := c.Param("code")     // path param
	tenantID := c.Query("tenant_id") // query param

	// ✅ Validation: allowed characters + max length
	if len(shortCode) == 0 || len(shortCode) > 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code must be 1-16 characters"})
		return
	}

	if !base62Regex.MatchString(shortCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code can only contain letters and digits"})
		return
	}

	// 🔥 Resolve the stateless URL via service layer
	longURL, err := u.service.ResolveStateless(shortCode, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🔄 Redirect to the original URL
	c.Redirect(http.StatusFound, longURL)
}
