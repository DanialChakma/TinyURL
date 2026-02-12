package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/models"
	"go.mod/services"
)

type AuthController struct {
	service *services.AuthService
}

// ----------- Swagger DTOs ------------

// RegisterRequest represents register payload
type RegisterRequest struct {
	Username string `json:"username" example:"john" binding:"required"`
	Password string `json:"password" example:"123456" binding:"required"`
	Email    string `json:"email" example:"john@example.com" binding:"required,email"`
	Role     string `json:"role" example:"user" binding:"required"`
}

// LoginRequest represents login payload
type LoginRequest struct {
	Username string `json:"username" example:"john" binding:"required"`
	Password string `json:"password" example:"123456" binding:"required"`
}

// TokenResponse represents JWT response
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func NewAuthController(s *services.AuthService) *AuthController {
	return &AuthController{service: s}
}

// -------------------- REGISTER --------------------
// Register godoc
// @Summary Register new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register payload"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /register [post]
func (c *AuthController) Register(ctx *gin.Context) {

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Role     string `json:"role" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.Register(ctx.Request.Context(), &models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Role:     req.Role,
	})
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"user": user})
}

// -------------------- LOGIN --------------------
// Login godoc
// @Summary Login user
// @Description Authenticate user and return access & refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := c.service.Login(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// -------------------- REFRESH --------------------
// Refresh godoc
// @Summary Refresh access token
// @Description Generate new access and refresh tokens using refresh token (Token Rotation)
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} TokenResponse
// @Failure 401 {object} map[string]string
// @Router /refresh [post]
func (c *AuthController) Refresh(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		return
	}

	refreshToken := authHeader[len(bearerPrefix):]

	access, refresh, err := c.service.Refresh(ctx.Request.Context(), refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}
