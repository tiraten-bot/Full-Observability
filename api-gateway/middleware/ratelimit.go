package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tair/full-observability/pkg/logger"
)

// RateLimiter implements rate limiting using Redis
type RateLimiter struct {
	redis       *redis.Client
	maxRequests int           // Maximum requests allowed
	window      time.Duration // Time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:       redisClient,
		maxRequests: maxRequests,
		window:      window,
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get identifier (user ID if authenticated, otherwise IP)
		identifier := c.IP()
		
		if userID := c.Locals("user_id"); userID != nil {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		// Check rate limit
		allowed, remaining, resetTime, err := rl.checkLimit(c.UserContext(), identifier)
		if err != nil {
			logger.Logger.Error().
				Err(err).
				Str("identifier", identifier).
				Msg("Rate limiter error")
			// On error, allow request but log it
			return c.Next()
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			logger.Logger.Warn().
				Str("identifier", identifier).
				Int("limit", rl.maxRequests).
				Msg("Rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Try again in %v", time.Until(resetTime).Round(time.Second)),
				"retry_after": time.Until(resetTime).Seconds(),
			})
		}

		return c.Next()
	}
}

// checkLimit checks if request is within rate limit using sliding window
func (rl *RateLimiter) checkLimit(ctx context.Context, identifier string) (bool, int, time.Time, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now()
	windowStart := now.Add(-rl.window)

	pipe := rl.redis.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count requests in current window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// Set expiration
	pipe.Expire(ctx, key, rl.window+time.Minute)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	// Get count
	count := countCmd.Val()

	// Calculate remaining and reset time
	remaining := rl.maxRequests - int(count) - 1
	if remaining < 0 {
		remaining = 0
	}

	resetTime := now.Add(rl.window)

	// Check if limit exceeded
	allowed := count < int64(rl.maxRequests)

	return allowed, remaining, resetTime, nil
}

// GlobalRateLimiter creates a rate limiter for all requests
func GlobalRateLimiter(redisClient *redis.Client) fiber.Handler {
	limiter := NewRateLimiter(redisClient, 100, time.Minute) // 100 req/min
	return limiter.Middleware()
}

// UserRateLimiter creates a stricter rate limiter for authenticated users
func UserRateLimiter(redisClient *redis.Client) fiber.Handler {
	limiter := NewRateLimiter(redisClient, 60, time.Minute) // 60 req/min per user
	return limiter.Middleware()
}

