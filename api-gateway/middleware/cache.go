package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tair/full-observability/pkg/logger"
)

// CacheConfig holds cache configuration
type CacheConfig struct {
	DefaultTTL       time.Duration // Default cache TTL
	CacheableMethods []string      // HTTP methods to cache
	CacheableStatus  []int         // HTTP status codes to cache
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		DefaultTTL:       5 * time.Minute,
		CacheableMethods: []string{"GET", "HEAD"},
		CacheableStatus:  []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501},
	}
}

// CacheMiddleware implements response caching with Redis
func CacheMiddleware(redisClient *redis.Client, config CacheConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip caching if Redis is not available
		if redisClient == nil {
			return c.Next()
		}

		// Only cache specific methods
		if !isMethodCacheable(c.Method(), config.CacheableMethods) {
			return c.Next()
		}

		// Generate cache key
		cacheKey := generateCacheKey(c)

		// Try to get from cache
		ctx := context.Background()
		cachedResponse, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil && len(cachedResponse) > 0 {
			// Cache hit
			logger.Logger.Debug().
				Str("path", c.Path()).
				Str("cache_key", cacheKey).
				Msg("Cache hit")

			c.Set("X-Cache", "HIT")
			c.Set("Content-Type", "application/json")
			return c.Send(cachedResponse)
		}

		// Cache miss - execute request
		logger.Logger.Debug().
			Str("path", c.Path()).
			Str("cache_key", cacheKey).
			Msg("Cache miss")

		// Capture response
		err = c.Next()

		// Check if response should be cached
		statusCode := c.Response().StatusCode()
		if isStatusCacheable(statusCode, config.CacheableStatus) {
			responseBody := c.Response().Body()

			// Cache the response
			ttl := config.DefaultTTL
			if err := redisClient.Set(ctx, cacheKey, responseBody, ttl).Err(); err != nil {
				logger.Logger.Warn().
					Err(err).
					Str("cache_key", cacheKey).
					Msg("Failed to cache response")
			} else {
				logger.Logger.Debug().
					Str("path", c.Path()).
					Str("cache_key", cacheKey).
					Dur("ttl", ttl).
					Int("size", len(responseBody)).
					Msg("Response cached")
			}

			c.Set("X-Cache", "MISS")
		}

		return err
	}
}

// generateCacheKey generates a unique cache key for the request
func generateCacheKey(c *fiber.Ctx) string {
	// Include: method, path, query params, and auth header
	keyComponents := fmt.Sprintf("%s:%s:%s:%s",
		c.Method(),
		c.Path(),
		string(c.Request().URI().QueryString()),
		c.Get("Authorization"),
	)

	// Hash the key
	hash := sha256.Sum256([]byte(keyComponents))
	return fmt.Sprintf("cache:%s", hex.EncodeToString(hash[:]))
}

// isMethodCacheable checks if HTTP method is cacheable
func isMethodCacheable(method string, cacheableMethods []string) bool {
	for _, m := range cacheableMethods {
		if m == method {
			return true
		}
	}
	return false
}

// isStatusCacheable checks if status code is cacheable
func isStatusCacheable(status int, cacheableStatus []int) bool {
	for _, s := range cacheableStatus {
		if s == status {
			return true
		}
	}
	return false
}

// InvalidateCache invalidates cache for a specific pattern
func InvalidateCache(redisClient *redis.Client, pattern string) error {
	ctx := context.Background()

	// Find all keys matching pattern
	iter := redisClient.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	// Delete all matching keys
	if len(keys) > 0 {
		if err := redisClient.Del(ctx, keys...).Err(); err != nil {
			return err
		}

		logger.Logger.Info().
			Int("count", len(keys)).
			Str("pattern", pattern).
			Msg("Cache invalidated")
	}

	return nil
}
