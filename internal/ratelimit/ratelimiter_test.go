package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 100, cfg.RequestsPerMinute)
	assert.Equal(t, 10, cfg.BurstSize)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envEnabled  string
		envRPM      string
		envBurst    string
		expectedCfg Config
	}{
		{
			name:        "default values",
			expectedCfg: DefaultConfig(),
		},
		{
			name:        "disabled",
			envEnabled:  "false",
			expectedCfg: Config{Enabled: false, RequestsPerMinute: 100, BurstSize: 10},
		},
		{
			name:        "custom rpm",
			envRPM:      "200",
			expectedCfg: Config{Enabled: true, RequestsPerMinute: 200, BurstSize: 10},
		},
		{
			name:        "custom burst",
			envBurst:    "20",
			expectedCfg: Config{Enabled: true, RequestsPerMinute: 100, BurstSize: 20},
		},
		{
			name:        "all custom",
			envEnabled:  "true",
			envRPM:      "300",
			envBurst:    "30",
			expectedCfg: Config{Enabled: true, RequestsPerMinute: 300, BurstSize: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars first
			os.Unsetenv("RATE_LIMIT_ENABLED")
			os.Unsetenv("RATE_LIMIT_REQUESTS_PER_MINUTE")
			os.Unsetenv("RATE_LIMIT_BURST_SIZE")

			// Set env vars
			if tt.envEnabled != "" {
				os.Setenv("RATE_LIMIT_ENABLED", tt.envEnabled)
				defer os.Unsetenv("RATE_LIMIT_ENABLED")
			}
			if tt.envRPM != "" {
				os.Setenv("RATE_LIMIT_REQUESTS_PER_MINUTE", tt.envRPM)
				defer os.Unsetenv("RATE_LIMIT_REQUESTS_PER_MINUTE")
			}
			if tt.envBurst != "" {
				os.Setenv("RATE_LIMIT_BURST_SIZE", tt.envBurst)
				defer os.Unsetenv("RATE_LIMIT_BURST_SIZE")
			}

			cfg := LoadConfig()
			assert.Equal(t, tt.expectedCfg.Enabled, cfg.Enabled)
			assert.Equal(t, tt.expectedCfg.RequestsPerMinute, cfg.RequestsPerMinute)
			assert.Equal(t, tt.expectedCfg.BurstSize, cfg.BurstSize)
		})
	}
}

func TestTokenBucket(t *testing.T) {
	t.Run("allow within capacity", func(t *testing.T) {
		tb := NewTokenBucket(5, 1) // 5 tokens, 1 per second refill

		// Should allow first 5 requests
		for i := 0; i < 5; i++ {
			assert.True(t, tb.Allow(), "Request %d should be allowed", i)
		}

		// 6th request should be denied
		assert.False(t, tb.Allow(), "6th request should be denied")
	})

	t.Run("refill over time", func(t *testing.T) {
		tb := NewTokenBucket(2, 10) // 2 tokens, 10 per second refill

		// Use up tokens
		assert.True(t, tb.Allow())
		assert.True(t, tb.Allow())
		assert.False(t, tb.Allow())

		// Wait for refill
		time.Sleep(150 * time.Millisecond)

		// Should have at least 1 token now
		assert.True(t, tb.Allow())
	})

	t.Run("remaining tokens", func(t *testing.T) {
		tb := NewTokenBucket(5, 1)

		assert.Equal(t, 5, tb.Remaining())
		tb.Allow()
		assert.Equal(t, 4, tb.Remaining())
	})

	t.Run("reset time", func(t *testing.T) {
		tb := NewTokenBucket(5, 1)

		// Use some tokens
		tb.Allow()
		tb.Allow()

		resetTime := tb.ResetTime()
		assert.True(t, resetTime.After(time.Now()))
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("allow when enabled", func(t *testing.T) {
		cfg := Config{Enabled: true, RequestsPerMinute: 60, BurstSize: 2}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		// Should allow burst size requests
		assert.True(t, rl.Allow("client1"))
		assert.True(t, rl.Allow("client1"))

		// Should deny after burst
		assert.False(t, rl.Allow("client1"))
	})

	t.Run("always allow when disabled", func(t *testing.T) {
		cfg := Config{Enabled: false}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		// Should always allow
		for i := 0; i < 100; i++ {
			assert.True(t, rl.Allow("client1"))
		}
	})

	t.Run("per-client isolation", func(t *testing.T) {
		cfg := Config{Enabled: true, RequestsPerMinute: 60, BurstSize: 2}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		// Use up client1's burst
		rl.Allow("client1")
		rl.Allow("client1")
		assert.False(t, rl.Allow("client1"))

		// client2 should still have burst available
		assert.True(t, rl.Allow("client2"))
		assert.True(t, rl.Allow("client2"))
	})

	t.Run("get limit info", func(t *testing.T) {
		cfg := Config{Enabled: true, RequestsPerMinute: 60, BurstSize: 5}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		// Get info before any requests
		remaining, resetTime, limit := rl.GetLimitInfo("client1")
		assert.Equal(t, 5, remaining)
		assert.Equal(t, 60, limit)
		// Reset time should be reasonable (not zero, not too far in the past)
		assert.False(t, resetTime.IsZero())
		assert.True(t, resetTime.After(time.Now().Add(-time.Second)))

		// Use a token
		rl.Allow("client1")

		remaining, _, _ = rl.GetLimitInfo("client1")
		assert.Equal(t, 4, remaining)
	})
}

func TestExtractClientID(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.2, 10.0.0.3"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "192.168.1.2"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.2",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:12345",
			expected:   "10.0.0.1",
		},
		{
			name: "X-Forwarded-For takes precedence",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
				"X-Real-IP":       "192.168.1.2",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			clientID := ExtractClientID(req)
			assert.Equal(t, tt.expected, clientID)
		})
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	t.Run("allows requests under limit", func(t *testing.T) {
		cfg := Config{Enabled: true, RequestsPerMinute: 60, BurstSize: 5}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		cfg := Config{Enabled: true, RequestsPerMinute: 60, BurstSize: 1}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// First request should succeed
		req1 := httptest.NewRequest("GET", "/test", nil)
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// Second request from same client should be rate limited
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = req1.RemoteAddr // Same client
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
		assert.Contains(t, rec2.Body.String(), "rate limit exceeded")
		assert.NotEmpty(t, rec2.Header().Get("X-RateLimit-Limit"))
	})

	t.Run("passes through when disabled", func(t *testing.T) {
		cfg := Config{Enabled: false}
		rl := NewRateLimiter(cfg)
		defer rl.Stop()

		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestRateLimiter_Cleanup(t *testing.T) {
	cfg := Config{Enabled: true, RequestsPerMinute: 60, BurstSize: 5}
	rl := NewRateLimiter(cfg)

	// Create a bucket
	rl.Allow("client1")
	rl.mu.RLock()
	_, exists := rl.buckets["client1"]
	rl.mu.RUnlock()
	assert.True(t, exists)

	rl.Stop()
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	tb := NewTokenBucket(1000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Allow()
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	cfg := Config{Enabled: true, RequestsPerMinute: 60000, BurstSize: 1000}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("client1")
	}
}
