// Package middleware holds shared Gin middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS is a permissive CORS middleware for local development.
// Restrict the allowed origins before production.
func CORS() gin.HandlerFunc {
	// Return the actual middleware handler function.
	return func(c *gin.Context) {
		// Set permissive CORS headers: allow any origin, common methods, and standard headers.
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		// Preflight requests (OPTIONS) are answered immediately with 204 No Content.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		// Pass the request to the next middleware or handler.
		c.Next()
	}
}
