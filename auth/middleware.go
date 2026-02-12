package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mod/services"
)

// Pass AuthService via closure
func AuthMiddleware(authService *services.AuthService, tokenService *services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := tokenService.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// ✅ Use service, not direct DB
		user, err := authService.GetUserByUsername(c.Request.Context(), claims.Username)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if claims.TokenVersion != user.TokenVersion {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalidated"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
