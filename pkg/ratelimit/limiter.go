package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config defines rate limit configuration
type Config struct {
	// Requests is the maximum number of requests allowed in the window
	Requests int
	// Window is the time window for rate limiting
	Window time.Duration
}

// Preset rate limit configurations
var (
	// AuthLimit - strict limit for auth endpoints (prevent brute force)
	AuthLimit = Config{Requests: 5, Window: time.Minute}

	// PublicLimit - limit for public/unauthenticated endpoints
	PublicLimit = Config{Requests: 60, Window: time.Minute}

	// StandardLimit - limit for authenticated endpoints
	StandardLimit = Config{Requests: 120, Window: time.Minute}

	// StrictLimit - limit for sensitive operations
	StrictLimit = Config{Requests: 10, Window: time.Minute}

	// RelaxedLimit - limit for read-heavy endpoints
	RelaxedLimit = Config{Requests: 300, Window: time.Minute}
)

// Result contains rate limit check result
type Result struct {
	// Allowed indicates if the request is allowed
	Allowed bool
	// Remaining is the number of requests remaining in the window
	Remaining int
	// ResetAt is when the rate limit window resets
	ResetAt time.Time
	// RetryAfter is how long to wait before retrying (if not allowed)
	RetryAfter time.Duration
}

// Limiter defines the rate limiter interface
type Limiter interface {
	// Allow checks if a request is allowed for the given key
	Allow(ctx context.Context, key string, cfg Config) (*Result, error)
}

// RedisLimiter implements sliding window rate limiting using Redis
type RedisLimiter struct {
	client *redis.Client
	prefix string
}

// NewRedisLimiter creates a new Redis-based rate limiter
func NewRedisLimiter(client *redis.Client, prefix string) *RedisLimiter {
	if prefix == "" {
		prefix = "ratelimit"
	}
	return &RedisLimiter{
		client: client,
		prefix: prefix,
	}
}

// Allow implements sliding window rate limiting
// Uses Redis sorted sets for accurate sliding window counting
func (l *RedisLimiter) Allow(ctx context.Context, key string, cfg Config) (*Result, error) {
	now := time.Now()
	windowStart := now.Add(-cfg.Window)
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)

	// Use a pipeline for atomic operations
	pipe := l.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	currentCount := int(countCmd.Val())

	result := &Result{
		Remaining: cfg.Requests - currentCount - 1,
		ResetAt:   now.Add(cfg.Window),
	}

	if currentCount >= cfg.Requests {
		// Rate limit exceeded
		result.Allowed = false
		result.Remaining = 0

		// Calculate retry after based on oldest entry
		oldestCmd := l.client.ZRange(ctx, redisKey, 0, 0)
		oldest, err := oldestCmd.Result()
		if err == nil && len(oldest) > 0 {
			// Parse the oldest timestamp to calculate retry time
			var oldestTime int64
			fmt.Sscanf(oldest[0], "%d", &oldestTime)
			oldestEntry := time.Unix(0, oldestTime)
			result.RetryAfter = oldestEntry.Add(cfg.Window).Sub(now)
			if result.RetryAfter < 0 {
				result.RetryAfter = time.Second
			}
		} else {
			result.RetryAfter = cfg.Window
		}

		return result, nil
	}

	// Add current request to the window
	member := fmt.Sprintf("%d", now.UnixNano())
	err = l.client.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: member,
	}).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to record request: %w", err)
	}

	// Set expiry on the key
	l.client.Expire(ctx, redisKey, cfg.Window+time.Second)

	result.Allowed = true
	if result.Remaining < 0 {
		result.Remaining = 0
	}

	return result, nil
}

// KeyFromIP generates a rate limit key from IP address
func KeyFromIP(ip string, endpoint string) string {
	return fmt.Sprintf("ip:%s:%s", ip, endpoint)
}

// KeyFromUserID generates a rate limit key from user ID
func KeyFromUserID(userID int64, endpoint string) string {
	return fmt.Sprintf("user:%d:%s", userID, endpoint)
}

// KeyGlobal generates a global rate limit key for an endpoint
func KeyGlobal(endpoint string) string {
	return fmt.Sprintf("global:%s", endpoint)
}
