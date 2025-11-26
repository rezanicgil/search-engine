// redis_cache.go - Cache implementations (in-memory + Redis)
// Provides caching layer for frequently accessed data
package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a simple interface for key-value caching with TTL support.
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

type item struct {
	value      interface{}
	expiration time.Time
}

// InMemoryCache is a threadsafe in-memory implementation of Cache.
// It is good enough for this case study and can be replaced with Redis later.
type InMemoryCache struct {
	mu         sync.RWMutex
	items      map[string]item
	defaultTTL time.Duration
}

// NewInMemoryCache creates a new in-memory cache with a default TTL.
func NewInMemoryCache(defaultTTL time.Duration) *InMemoryCache {
	if defaultTTL <= 0 {
		defaultTTL = time.Minute
	}
	c := &InMemoryCache{
		items:      make(map[string]item),
		defaultTTL: defaultTTL,
	}

	// Background cleanup goroutine.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			c.cleanup()
		}
	}()

	return c
}

// Get returns a value if present and not expired.
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(it.expiration) {
		// Lazy delete expired item.
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	return it.value, true
}

// Set stores a value with an optional TTL (0 = use default TTL).
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	c.mu.Lock()
	c.items[key] = item{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// cleanup removes expired items.
func (c *InMemoryCache) cleanup() {
	now := time.Now()
	c.mu.Lock()
	for k, it := range c.items {
		if now.After(it.expiration) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

// RedisCache is a Redis-backed implementation of Cache.
// It stores values as gob-encoded bytes under the given key.
type RedisCache struct {
	client *redis.Client
}

// RedisCacheWrapper wraps redis.Client to implement Cache interface
// This allows sharing the same Redis client for cache and rate limiting
type RedisCacheWrapper struct {
	Client *redis.Client
}

// NewRedisCache creates a new Redis cache client.
// addr: "host:port", db: Redis DB index.
func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Basic connectivity check (non-fatal: return nil on error).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil
	}

	return &RedisCache{client: rdb}
}

// Get returns a value if present. The caller must type-assert it back.
func (r *RedisCache) Get(key string) (interface{}, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	// For simplicity, we just store JSON-encoded bytes or gob-encoded bytes.
	// The caller is expected to store pointers to structs and cast them back.
	return val, true
}

// Set stores a value with TTL. Value is expected to be JSON-serializable []byte.
func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	default:
		// Fallback: ignore non-byte values in this simple implementation.
		return
	}

	if ttl <= 0 {
		ttl = time.Minute
	}
	_ = r.client.Set(ctx, key, b, ttl).Err()
}

// Get implements Cache interface for RedisCacheWrapper
func (r *RedisCacheWrapper) Get(key string) (interface{}, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	val, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return val, true
}

// Set implements Cache interface for RedisCacheWrapper
func (r *RedisCacheWrapper) Set(key string, value interface{}, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	default:
		return
	}

	if ttl <= 0 {
		ttl = time.Minute
	}
	_ = r.Client.Set(ctx, key, b, ttl).Err()
}
