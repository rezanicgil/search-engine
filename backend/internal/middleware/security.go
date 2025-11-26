// security.go - Security headers middleware
// Adds security headers to HTTP responses for better security posture

package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to all responses
// This helps protect against common web vulnerabilities
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection (legacy but still useful)
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Referrer policy - control how much referrer information is sent
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy - restrict resource loading
		// Adjust based on your needs
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		// Permissions Policy (formerly Feature-Policy)
		// Restrict browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

