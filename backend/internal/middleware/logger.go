// logger.go - Request logging middleware
// Logs all incoming requests for debugging and monitoring
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggerMiddleware logs basic request/response information and attaches
// a simple trace ID to each request for easier debugging.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate a simple trace ID and add to context + response headers.
		traceID := uuid.New().String()
		c.Set("trace_id", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)

		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		method := c.Request.Method

		// Process request.
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		log.Printf("[REQ] trace=%s | %3d | %13v | %15s | %-7s %s",
			traceID,
			status,
			latency,
			clientIP,
			method,
			path,
		)
	}
}
