// rate_limiter_redis.go - Redis-based rate limiting middleware
// Provides distributed rate limiting across multiple instances
package middleware

import (
	"log"
	"net/http"
	"search-engine/backend/pkg/ratelimit"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiterConfig controls Redis-based rate limiting behavior
type RedisRateLimiterConfig struct {
	Client            *redis.Client
	RequestsPerMinute int
	KeyPrefix         string // Optional prefix for Redis keys
}

// NewRedisRateLimiterMiddleware creates a Redis-based rate limiter middleware
// This provides distributed rate limiting across multiple instances
func NewRedisRateLimiterMiddleware(cfg RedisRateLimiterConfig) gin.HandlerFunc {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60
	}
	if cfg.Client == nil {
		// Fallback to in-memory if Redis client is not provided
		return NewIPRateLimiterMiddleware(RateLimiterConfig{
			RequestsPerMinute: cfg.RequestsPerMinute,
		})
	}

	limiter := ratelimit.NewRedisRateLimiter(cfg.Client, cfg.KeyPrefix)
	window := time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := c.Request.Context()

		// Check rate limit
		allowed, remaining, resetTime, err := limiter.Allow(ctx, ip, cfg.RequestsPerMinute, window)
		if err != nil {
			// On Redis error, allow the request but log the error
			// This prevents Redis failures from blocking all requests
			log.Printf("Rate limit Redis error: %v", err)
			c.Next()
			return
		}

		// Add rate limit headers (RFC 6585)
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"message":     "Too many requests, please try again later.",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			return
		}

		c.Next()
	}
}
