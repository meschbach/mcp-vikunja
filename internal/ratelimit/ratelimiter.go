// Package ratelimit provides rate limiting functionality for the MCP Vikunja server.
package ratelimit

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds rate limiting configuration
type Config struct {
	Enabled           bool
	RequestsPerMinute int
	BurstSize         int
}

// DefaultConfig returns the default rate limiting configuration
func DefaultConfig() Config {
	return Config{
		Enabled:           true,
		RequestsPerMinute: 100,
		BurstSize:         10,
	}
}

// LoadConfig loads rate limiting configuration from environment
func LoadConfig() Config {
	cfg := DefaultConfig()

	if enabled := getEnv("RATE_LIMIT_ENABLED", "true"); enabled == "false" {
		cfg.Enabled = false
	}

	if rpm := getEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", "100"); rpm != "" {
		if val, err := strconv.Atoi(rpm); err == nil {
			cfg.RequestsPerMinute = val
		}
	}

	if burst := getEnv("RATE_LIMIT_BURST_SIZE", "10"); burst != "" {
		if val, err := strconv.Atoi(burst); err == nil {
			cfg.BurstSize = val
		}
	}

	return cfg
}

// TokenBucket implements the token bucket rate limiting algorithm
type TokenBucket struct {
	capacity   int
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     float64(capacity),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	tb.lastRefill = now

	// Check if we have a token available
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// Remaining returns the current number of tokens
func (tb *TokenBucket) Remaining() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	tb.lastRefill = now

	return int(tb.tokens)
}

// ResetTime returns the time when the bucket will be full
func (tb *TokenBucket) ResetTime() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tokensNeeded := float64(tb.capacity) - tb.tokens
	secondsNeeded := tokensNeeded / tb.refillRate
	return time.Now().Add(time.Duration(secondsNeeded) * time.Second)
}

// RateLimiter manages rate limits for multiple clients
type RateLimiter struct {
	config        Config
	buckets       map[string]*TokenBucket
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	done          chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config Config) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		buckets: make(map[string]*TokenBucket),
		done:    make(chan struct{}),
	}

	// Start cleanup goroutine
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given client is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	if !rl.config.Enabled {
		return true
	}

	rl.mu.RLock()
	bucket, exists := rl.buckets[clientID]
	rl.mu.RUnlock()

	if !exists {
		// Create new bucket for this client
		refillRate := float64(rl.config.RequestsPerMinute) / 60.0
		bucket = NewTokenBucket(rl.config.BurstSize, refillRate)

		rl.mu.Lock()
		rl.buckets[clientID] = bucket
		rl.mu.Unlock()
	}

	return bucket.Allow()
}

// GetLimitInfo returns rate limit information for a client
func (rl *RateLimiter) GetLimitInfo(clientID string) (remaining int, resetTime time.Time, limit int) {
	if !rl.config.Enabled {
		return rl.config.BurstSize, time.Now().Add(time.Hour), rl.config.RequestsPerMinute
	}

	rl.mu.RLock()
	bucket, exists := rl.buckets[clientID]
	rl.mu.RUnlock()

	if !exists {
		return rl.config.BurstSize, time.Now(), rl.config.RequestsPerMinute
	}

	return bucket.Remaining(), bucket.ResetTime(), rl.config.RequestsPerMinute
}

// Stop stops the rate limiter and cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.done)
	rl.cleanupTicker.Stop()
}

// cleanup removes expired buckets periodically
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.mu.Lock()
			for id, bucket := range rl.buckets {
				// Remove buckets that have been full for a while (inactive clients)
				if bucket.Remaining() == rl.config.BurstSize {
					delete(rl.buckets, id)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// ExtractClientID extracts a client identifier from an HTTP request
// It tries X-Forwarded-For header, then X-Real-IP, then RemoteAddr
func ExtractClientID(r *http.Request) string {
	// Try X-Forwarded-For header (for clients behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	remoteAddr := r.RemoteAddr
	// Strip port if present
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		remoteAddr = remoteAddr[:idx]
	}
	return remoteAddr
}

// Middleware returns an HTTP middleware function for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		clientID := ExtractClientID(r)

		// Check if request is allowed
		if !rl.Allow(clientID) {
			remaining, resetTime, limit := rl.GetLimitInfo(clientID)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			w.WriteHeader(http.StatusTooManyRequests)

			fmt.Fprintf(w, `{"error": "rate limit exceeded", "limit": %d, "remaining": %d, "reset_time": "%s"}`,
				limit, remaining, resetTime.Format(time.RFC3339))
			return
		}

		// Add rate limit headers to successful response
		remaining, resetTime, limit := rl.GetLimitInfo(clientID)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		next.ServeHTTP(w, r)
	})
}

// helper function
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
