// Package ratelimit provides Redis-backed rate limiting for the betting platform.
//
// It implements:
// - Per-user rate limiting
// - Per-IP rate limiting
// - Sliding window algorithm
// - Distributed rate limiting across multiple instances
// - Configurable limits and windows
package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// Config holds rate limiting configuration
type Config struct {
	// Default limits
	DefaultRequestsPerWindow int
	DefaultWindow            time.Duration

	// User-specific limits
	UserRequestsPerWindow int
	UserWindow            time.Duration

	// IP-specific limits
	IPRequestsPerWindow int
	IPWindow            time.Duration

	// Global limits (across all users/IPs)
	GlobalRequestsPerWindow int
	GlobalWindow            time.Duration

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Key prefixes
	UserPrefix   string
	IPPrefix     string
	GlobalPrefix string
}

// DefaultConfig returns a default rate limiting configuration
func DefaultConfig() Config {
	return Config{
		DefaultRequestsPerWindow: 100,
		DefaultWindow:            time.Minute,

		UserRequestsPerWindow: 50,
		UserWindow:            time.Minute,

		IPRequestsPerWindow: 30,
		IPWindow:            time.Minute,

		GlobalRequestsPerWindow: 1000,
		GlobalWindow:            time.Minute,

		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,

		UserPrefix:   "ratelimit:user:",
		IPPrefix:     "ratelimit:ip:",
		GlobalPrefix: "ratelimit:global:",
	}
}

// RedisLimiter implements rate limiting using Redis
type RedisLimiter struct {
	client *redis.Client
	config Config
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(ctx context.Context, config Config) (*RedisLimiter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisLimiter{
		client: rdb,
		config: config,
	}, nil
}

// LimitResult represents the result of a rate limit check
type LimitResult struct {
	Allowed     bool
	Remaining   int
	ResetTime   time.Time
	LimitType   string // "user", "ip", "global"
	Key         string
	WindowStart time.Time
	WindowEnd   time.Time
}

// CheckUserLimit checks if a user has exceeded their rate limit
func (rl *RedisLimiter) CheckUserLimit(ctx context.Context, userID uuid.UUID) (*LimitResult, error) {
	key := rl.config.UserPrefix + userID.String()
	return rl.checkLimit(ctx, key, rl.config.UserRequestsPerWindow, rl.config.UserWindow, "user")
}

// CheckIPLimit checks if an IP has exceeded their rate limit
func (rl *RedisLimiter) CheckIPLimit(ctx context.Context, ip string) (*LimitResult, error) {
	// Normalize IP address
	normalizedIP := net.ParseIP(ip)
	if normalizedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Use IPv4 or IPv6 representation consistently
	var key string
	if normalizedIP.To4() != nil {
		key = rl.config.IPPrefix + normalizedIP.String()
	} else {
		key = rl.config.IPPrefix + normalizedIP.String()
	}

	return rl.checkLimit(ctx, key, rl.config.IPRequestsPerWindow, rl.config.IPWindow, "ip")
}

// CheckGlobalLimit checks if the global limit has been exceeded
func (rl *RedisLimiter) CheckGlobalLimit(ctx context.Context) (*LimitResult, error) {
	key := rl.config.GlobalPrefix + "all"
	return rl.checkLimit(ctx, key, rl.config.GlobalRequestsPerWindow, rl.config.GlobalWindow, "global")
}

// checkLimit performs the actual rate limit check using sliding window
func (rl *RedisLimiter) checkLimit(ctx context.Context, key string, limit int, window time.Duration, limitType string) (*LimitResult, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	windowEnd := windowStart.Add(window)

	// Use Lua script for atomic sliding window implementation
	luaScript := `
local key = KEYS[1]
local window_start = ARGV[1]
local window_end = ARGV[2]
local limit = tonumber(ARGV[3])
local now = ARGV[4]

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, window_start - 1)

-- Count current requests
local current = redis.call('ZCARD', key)

-- Check if limit exceeded
if current >= limit then
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    return {0, limit, current, oldest[2]}
end

-- Add current request
redis.call('ZADD', key, now, now)
redis.call('EXPIRE', key, math.ceil(tonumber(window_end - window_start) / 1000000000))

return {1, limit - current, limit - current, now}
`

	result, err := rl.client.Eval(ctx, luaScript, []string{key},
		windowStart.UnixNano(),
		windowEnd.UnixNano(),
		limit,
		now.UnixNano()).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// Parse Lua script result
	resultSlice, ok := result.([]any)
	if !ok || len(resultSlice) != 4 {
		return nil, fmt.Errorf("unexpected result from Redis")
	}

	allowed, _ := resultSlice[0].(int64)
	remaining, _ := resultSlice[1].(int64)
	resetTimeNano, _ := resultSlice[3].(int64)

	return &LimitResult{
		Allowed:     allowed == 1,
		Remaining:   int(remaining),
		ResetTime:   time.Unix(0, resetTimeNano),
		LimitType:   limitType,
		Key:         key,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
	}, nil
}

// CheckLimits checks multiple limits (user, IP, global) and returns the most restrictive result
func (rl *RedisLimiter) CheckLimits(ctx context.Context, userID uuid.UUID, ip string) (*LimitResult, error) {
	// Check all limits in parallel
	userChan := make(chan *LimitResult, 1)
	ipChan := make(chan *LimitResult, 1)
	globalChan := make(chan *LimitResult, 1)
	errChan := make(chan error, 3)

	// Check user limit
	go func() {
		result, err := rl.CheckUserLimit(ctx, userID)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- result
	}()

	// Check IP limit
	go func() {
		result, err := rl.CheckIPLimit(ctx, ip)
		if err != nil {
			errChan <- err
			return
		}
		ipChan <- result
	}()

	// Check global limit
	go func() {
		result, err := rl.CheckGlobalLimit(ctx)
		if err != nil {
			errChan <- err
			return
		}
		globalChan <- result
	}()

	// Collect results
	var results []*LimitResult
	var errors []error

	for range 3 {
		select {
		case result := <-userChan:
			results = append(results, result)
		case result := <-ipChan:
			results = append(results, result)
		case result := <-globalChan:
			results = append(results, result)
		case err := <-errChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("rate limit check errors: %v", errors)
	}

	// Return the most restrictive result (lowest remaining or first disallowed)
	var mostRestrictive *LimitResult
	for _, result := range results {
		if mostRestrictive == nil {
			mostRestrictive = result
			continue
		}

		// If current result is not allowed, it's more restrictive
		if !result.Allowed && mostRestrictive.Allowed {
			mostRestrictive = result
			continue
		}

		// If both are allowed, choose the one with lower remaining
		if result.Allowed && mostRestrictive.Allowed && result.Remaining < mostRestrictive.Remaining {
			mostRestrictive = result
		}
	}

	return mostRestrictive, nil
}

// GetUserStats returns detailed statistics for a user
func (rl *RedisLimiter) GetUserStats(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	key := rl.config.UserPrefix + userID.String()
	return rl.getStats(ctx, key, rl.config.UserWindow)
}

// GetIPStats returns detailed statistics for an IP
func (rl *RedisLimiter) GetIPStats(ctx context.Context, ip string) (map[string]any, error) {
	normalizedIP := net.ParseIP(ip)
	if normalizedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	var key string
	if normalizedIP.To4() != nil {
		key = rl.config.IPPrefix + normalizedIP.String()
	} else {
		key = rl.config.IPPrefix + normalizedIP.String()
	}

	return rl.getStats(ctx, key, rl.config.IPWindow)
}

// getStats returns statistics for a given key
func (rl *RedisLimiter) getStats(ctx context.Context, key string, window time.Duration) (map[string]any, error) {
	now := time.Now()
	windowStart := now.Truncate(window)

	// Get current count and oldest request
	luaScript := `
local key = KEYS[1]
local window_start = ARGV[1]

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, window_start - 1)

-- Get statistics
local current = redis.call('ZCARD', key)
local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
local newest = redis.call('ZRANGE', key, -1, -1, 'WITHSCORES')

local stats = {
    current = current,
    oldest_time = oldest[2] or 0,
    newest_time = newest[2] or 0
}

return cjson.encode(stats)
`

	result, err := rl.client.Eval(ctx, luaScript, []string{key}, windowStart.UnixNano()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	var stats map[string]any
	if err := json.Unmarshal([]byte(result.(string)), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return stats, nil
}

// ResetUserLimit resets the rate limit for a specific user
func (rl *RedisLimiter) ResetUserLimit(ctx context.Context, userID uuid.UUID) error {
	key := rl.config.UserPrefix + userID.String()
	return rl.client.Del(ctx, key).Err()
}

// ResetIPLimit resets the rate limit for a specific IP
func (rl *RedisLimiter) ResetIPLimit(ctx context.Context, ip string) error {
	normalizedIP := net.ParseIP(ip)
	if normalizedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	var key string
	if normalizedIP.To4() != nil {
		key = rl.config.IPPrefix + normalizedIP.String()
	} else {
		key = rl.config.IPPrefix + normalizedIP.String()
	}

	return rl.client.Del(ctx, key).Err()
}

// Close closes the Redis connection
func (rl *RedisLimiter) Close() error {
	return rl.client.Close()
}
