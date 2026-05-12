package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haseeb/url-shortener/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the Authorization header
		authHeader := c.GetHeader("Authorization")

		// 2. Check if header exists and starts with "Bearer "
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort() // Stop the request here
			return
		}

		// 3. Extract the token (remove "Bearer " prefix)
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 4. Verify the token
		claims, err := utils.VerifyToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 5. Save user info in context so controllers can use it
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		// 6. Continue to the actual route handler
		c.Next()
	}
}
