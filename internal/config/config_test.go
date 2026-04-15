package config

import (
	"os"
	"testing"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setEnv(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv(key); err != nil {
			t.Logf("warning: failed to unset env %s: %v", key, err)
		}
	})
}

func TestLoad_Defaults(t *testing.T) {
	t.Cleanup(func() {
		for _, key := range []string{"MCP_TRANSPORT", "MCP_HTTP_HOST", "MCP_HTTP_PORT", "VIKUNJA_HOST", "VIKUNJA_TOKEN"} {
			if err := os.Unsetenv(key); err != nil {
				t.Logf("warning: failed to unset env %s: %v", key, err)
			}
		}
	})

	cfg, err := Load(nil, nil)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, TransportStdio, cfg.Transport)
	assert.Equal(t, "localhost", cfg.HTTP.Host)
	assert.Equal(t, 8080, cfg.HTTP.Port)
	assert.Equal(t, 30*time.Minute, cfg.HTTP.SessionTimeout)
	assert.False(t, cfg.HTTP.Stateless)
	assert.Equal(t, 30*time.Second, cfg.HTTP.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.HTTP.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.HTTP.IdleTimeout)
}

func TestLoad_HTTPTransport(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "http")

	cfg, err := Load(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, TransportHTTP, cfg.Transport)
}

func TestLoad_InvalidTransport(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "websocket")

	_, err := Load(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transport type")
}

func TestLoad_HTTPConfig(t *testing.T) {
	setEnv(t, "MCP_HTTP_HOST", "0.0.0.0")
	setEnv(t, "MCP_HTTP_PORT", "9000")
	setEnv(t, "MCP_HTTP_SESSION_TIMEOUT", "1h")
	setEnv(t, "MCP_HTTP_STATELESS", "true")
	setEnv(t, "MCP_HTTP_READ_TIMEOUT", "60s")
	setEnv(t, "MCP_HTTP_WRITE_TIMEOUT", "45s")
	setEnv(t, "MCP_HTTP_IDLE_TIMEOUT", "300s")

	cfg, err := Load(nil, nil)
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.HTTP.Host)
	assert.Equal(t, 9000, cfg.HTTP.Port)
	assert.Equal(t, time.Hour, cfg.HTTP.SessionTimeout)
	assert.True(t, cfg.HTTP.Stateless)
	assert.Equal(t, 60*time.Second, cfg.HTTP.ReadTimeout)
	assert.Equal(t, 45*time.Second, cfg.HTTP.WriteTimeout)
	assert.Equal(t, 300*time.Second, cfg.HTTP.IdleTimeout)
}

func TestLoad_VikunjaConfig(t *testing.T) {
	setEnv(t, "VIKUNJA_HOST", "https://vikunja.example.com")
	setEnv(t, "VIKUNJA_TOKEN", "test-token-123")
	setEnv(t, "VIKUNJA_INSECURE", "true")

	cfg, err := Load(nil, nil)
	require.NoError(t, err)

	assert.Equal(t, "https://vikunja.example.com", cfg.Vikunja.Host)
	assert.Equal(t, "test-token-123", cfg.Vikunja.Token)
	assert.True(t, cfg.Vikunja.Insecure)
}

func TestLoad_InvalidHTTPPort(t *testing.T) {
	setEnv(t, "MCP_HTTP_PORT", "invalid")

	_, err := Load(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HTTP port")
}

func TestLoad_InvalidDuration(t *testing.T) {
	setEnv(t, "MCP_HTTP_SESSION_TIMEOUT", "invalid")

	_, err := Load(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session timeout")
}

func TestLoad_InvalidBool(t *testing.T) {
	setEnv(t, "MCP_HTTP_STATELESS", "invalid")

	_, err := Load(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid stateless flag")
}

func TestValidate_ValidStdioConfig(t *testing.T) {
	cfg := &Config{
		Transport: TransportStdio,
		Vikunja: VikunjaConfig{
			Host:  "https://vikunja.example.com",
			Token: "test-token",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_ValidHTTPConfig(t *testing.T) {
	cfg := &Config{
		Transport: TransportHTTP,
		HTTP: HTTPConfig{
			Host: "localhost",
			Port: 19876,
		},
		Vikunja: VikunjaConfig{
			Host:  "https://vikunja.example.com",
			Token: "test-token",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidTransport(t *testing.T) {
	cfg := &Config{
		Transport: "invalid",
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transport type")
}

func TestValidate_InvalidHTTPPort(t *testing.T) {
	cfg := &Config{
		Transport: TransportHTTP,
		HTTP: HTTPConfig{
			Host: "localhost",
			Port: 0,
		},
		Vikunja: VikunjaConfig{
			Host:  "https://vikunja.example.com",
			Token: "test-token",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HTTP port")
}

func TestValidate_EmptyHost(t *testing.T) {
	cfg := &Config{
		Transport: TransportHTTP,
		HTTP: HTTPConfig{
			Host: "",
			Port: 8080,
		},
		Vikunja: VikunjaConfig{
			Host:  "https://vikunja.example.com",
			Token: "test-token",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP host cannot be empty")
}

func TestValidate_MissingVikunjaHost(t *testing.T) {
	cfg := &Config{
		Transport: TransportStdio,
		Vikunja: VikunjaConfig{
			Host:  "",
			Token: "test-token",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "VIKUNJA_HOST is required")
}

func TestValidate_MissingVikunjaToken(t *testing.T) {
	cfg := &Config{
		Transport: TransportStdio,
		Vikunja: VikunjaConfig{
			Host:  "https://vikunja.example.com",
			Token: "",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "VIKUNJA_TOKEN is required")
}

func TestHTTPConfig_Address(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{
			name:     "localhost default port",
			host:     "localhost",
			port:     8080,
			expected: "localhost:8080",
		},
		{
			name:     "IPv4 address",
			host:     "127.0.0.1",
			port:     9000,
			expected: "127.0.0.1:9000",
		},
		{
			name:     "IPv6 address",
			host:     "::1",
			port:     8080,
			expected: "[::1]:8080",
		},
		{
			name:     "empty host",
			host:     "",
			port:     8080,
			expected: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := HTTPConfig{
				Host: tt.host,
				Port: tt.port,
			}
			assert.Equal(t, tt.expected, cfg.Address())
		})
	}
}

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected vikunja.OutputFormat
		hasError bool
	}{
		{"json", vikunja.OutputFormatJSON, false},
		{"JSON", vikunja.OutputFormatJSON, false},
		{"markdown", vikunja.OutputFormatMarkdown, false},
		{"md", vikunja.OutputFormatMarkdown, false},
		{"MARKDOWN", vikunja.OutputFormatMarkdown, false},
		{"both", vikunja.OutputFormatBoth, false},
		{"invalid", vikunja.OutputFormatJSON, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseOutputFormat(tt.input)
			if tt.hasError {
				require.Error(t, err)
				assert.Equal(t, vikunja.OutputFormatJSON, result) // Should return default on error
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoad_CLIFormatPrecedence(t *testing.T) {
	t.Cleanup(func() {
		if err := os.Unsetenv("VIKUNJA_OUTPUT_FORMAT"); err != nil {
			t.Logf("warning: failed to unset env: %v", err)
		}
	})

	json := "json"
	cfg, err := Load(&json, nil)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatJSON, cfg.OutputFormat)

	markdown := "markdown"
	cfg, err = Load(&markdown, nil)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatMarkdown, cfg.OutputFormat)
}

func TestLoad_EnvironmentVariableFallback(t *testing.T) {
	t.Cleanup(func() {
		if err := os.Unsetenv("VIKUNJA_OUTPUT_FORMAT"); err != nil {
			t.Logf("warning: failed to unset env: %v", err)
		}
	})

	cfg, err := Load(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatMarkdown, cfg.OutputFormat)

	setEnv(t, "VIKUNJA_OUTPUT_FORMAT", "markdown")
	cfg, err = Load(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatMarkdown, cfg.OutputFormat)
}
