// rate_limiter.go - Rate limiting middleware
// Prevents API abuse and manages request limits
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// simpleTokenBucket is a very small in-memory token bucket
// used for IP-based or key-based rate limiting.
type simpleTokenBucket struct {
	capacity     int
	tokens       int
	refillRate   int          // tokens per interval
	refillTicker *time.Ticker // refill interval
	mu           sync.Mutex
}

func newSimpleTokenBucket(capacity, refillRate int, interval time.Duration) *simpleTokenBucket {
	if capacity <= 0 {
		capacity = 60
	}
	if refillRate <= 0 {
		refillRate = capacity
	}
	tb := &simpleTokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
	}
	tb.refillTicker = time.NewTicker(interval)
	go func() {
		for range tb.refillTicker.C {
			tb.mu.Lock()
			tb.tokens += tb.refillRate
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			tb.mu.Unlock()
		}
	}()
	return tb
}

func (tb *simpleTokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.tokens <= 0 {
		return false
	}
	tb.tokens--
	return true
}

// RateLimiterConfig controls how the rate limiter behaves.
type RateLimiterConfig struct {
	RequestsPerMinute int
}

// NewIPRateLimiterMiddleware limits requests per IP address.
// Default: 60 req/min per IP.
func NewIPRateLimiterMiddleware(cfg RateLimiterConfig) gin.HandlerFunc {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60
	}

	var (
		bucketsMu sync.Mutex
		buckets   = make(map[string]*simpleTokenBucket)
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		bucketsMu.Lock()
		bucket, ok := buckets[ip]
		if !ok {
			bucket = newSimpleTokenBucket(cfg.RequestsPerMinute, cfg.RequestsPerMinute, time.Minute)
			buckets[ip] = bucket
		}
		bucketsMu.Unlock()

		if !bucket.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "Too many requests, please try again later.",
			})
			return
		}

		c.Next()
	}
}
