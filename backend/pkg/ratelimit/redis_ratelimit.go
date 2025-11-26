// redis_ratelimit.go - Redis-based rate limiting
// Provides distributed rate limiting using Redis
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements rate limiting using Redis
// Uses sliding window algorithm for accurate rate limiting
type RedisRateLimiter struct {
	client *redis.Client
	prefix string
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
// prefix: Key prefix for rate limit keys (e.g., "ratelimit:")
func NewRedisRateLimiter(client *redis.Client, prefix string) *RedisRateLimiter {
	if prefix == "" {
		prefix = "ratelimit:"
	}
	return &RedisRateLimiter{
		client: client,
		prefix: prefix,
	}
}

// Allow checks if a request is allowed based on the rate limit
// Returns true if allowed, false if rate limit exceeded
// Also returns remaining requests and reset time
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
	redisKey := r.prefix + key
	now := time.Now()
	windowStart := now.Add(-window)

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Remove expired entries (older than window)
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart.Unix()))

	// Count current requests in the window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: fmt.Sprintf("%d", now.UnixNano()), // Unique member
	})

	// Set expiration for the key (window + 1 second buffer)
	pipe.Expire(ctx, redisKey, window+time.Second)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("redis rate limit error: %w", err)
	}

	// Get count after adding current request
	count := int(countCmd.Val())

	// Check if limit exceeded
	allowed = count <= limit
	remaining = limit - count
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (oldest entry + window)
	if count > 0 {
		oldestCmd := r.client.ZRangeWithScores(ctx, redisKey, 0, 0)
		if len(oldestCmd.Val()) > 0 {
			oldestScore := int64(oldestCmd.Val()[0].Score)
			resetTime = time.Unix(oldestScore, 0).Add(window)
		} else {
			resetTime = now.Add(window)
		}
	} else {
		resetTime = now.Add(window)
	}

	return allowed, remaining, resetTime, nil
}

// GetRemaining returns the remaining requests for a key without consuming a request
func (r *RedisRateLimiter) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (remaining int, resetTime time.Time, err error) {
	redisKey := r.prefix + key
	now := time.Now()
	windowStart := now.Add(-window)

	// Remove expired entries
	err = r.client.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart.Unix())).Err()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("redis rate limit error: %w", err)
	}

	// Count current requests
	count, err := r.client.ZCard(ctx, redisKey).Result()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("redis rate limit error: %w", err)
	}

	remaining = limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time
	if count > 0 {
		oldestCmd := r.client.ZRangeWithScores(ctx, redisKey, 0, 0)
		if err := oldestCmd.Err(); err == nil && len(oldestCmd.Val()) > 0 {
			oldestScore := int64(oldestCmd.Val()[0].Score)
			resetTime = time.Unix(oldestScore, 0).Add(window)
		} else {
			resetTime = now.Add(window)
		}
	} else {
		resetTime = now.Add(window)
	}

	return remaining, resetTime, nil
}

// Reset removes all rate limit entries for a key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	redisKey := r.prefix + key
	return r.client.Del(ctx, redisKey).Err()
}
