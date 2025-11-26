// rate_limiter.go - Rate limiting for provider requests
// Implements token bucket algorithm for rate limiting
package provider

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
// This ensures we don't exceed the provider's rate limit
type RateLimiter struct {
	rate       int        // Requests per minute
	tokens     int        // Current available tokens
	maxTokens  int        // Maximum tokens (same as rate)
	lastUpdate time.Time  // Last time tokens were refilled
	mu         sync.Mutex // Protects token bucket
}

// NewRateLimiter creates a new rate limiter
// rate: maximum requests per minute
func NewRateLimiter(rate int) *RateLimiter {
	if rate < 1 {
		rate = 1 // Minimum 1 request per minute
	}
	return &RateLimiter{
		rate:       rate,
		tokens:     rate,
		maxTokens:  rate,
		lastUpdate: time.Now(),
	}
}

// Wait blocks until a token is available
// This implements the token bucket algorithm
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate)

	// Calculate how many tokens to add
	// Rate is per minute, so we add tokens proportionally
	tokensToAdd := int(elapsed.Minutes() * float64(rl.rate))
	if tokensToAdd > 0 {
		rl.tokens = rl.tokens + tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastUpdate = now
	}

	// If we have tokens, use one immediately
	if rl.tokens > 0 {
		rl.tokens--
		return
	}

	// No tokens available, calculate wait time
	// Wait until next token is available
	timePerToken := time.Minute / time.Duration(rl.rate)
	waitTime := timePerToken - elapsed
	if waitTime > 0 {
		rl.mu.Unlock()
		time.Sleep(waitTime)
		rl.mu.Lock()
		rl.tokens--
		rl.lastUpdate = time.Now()
	} else {
		rl.tokens--
		rl.lastUpdate = now
	}
}

// SetRate updates the rate limit
// Useful when provider rate limit changes
func (rl *RateLimiter) SetRate(rate int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rate < 1 {
		rate = 1
	}
	rl.rate = rate
	rl.maxTokens = rate
	if rl.tokens > rate {
		rl.tokens = rate
	}
}
