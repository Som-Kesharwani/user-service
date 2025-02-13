package middleware

import (
	"github.com/Som-Kesharwani/user-service/helper"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Split the "Bearer" part from the token
		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 || parts[1] == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token missing or malformed"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		if claims, err := helper.ValidateToken(tokenString); err == nil {
			// Check token expiration

			// Attach claims to the request context (optional, but useful)
			c.Set("email", claims.Email)
			c.Set("userID", claims.UserID)

			// Proceed to the next handler
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}
	}
}
