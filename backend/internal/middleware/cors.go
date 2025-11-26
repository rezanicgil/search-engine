// cors.go - CORS middleware
// Handles cross-origin requests for frontend integration
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware adds CORS headers to responses so that the React frontend
// (running on a different origin, e.g. http://localhost:3000) can call the API.
//
// For a real production app, you would restrict AllowedOrigins instead of "*".
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow all origins for now; tighten this when you know your frontend origin(s).
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Handle preflight requests quickly.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Continue to next middleware/handler.
		c.Next()
	}
}
