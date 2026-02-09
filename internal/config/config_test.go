package config

import (
	"os"
	"testing"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("MCP_TRANSPORT")
	os.Unsetenv("MCP_HTTP_HOST")
	os.Unsetenv("MCP_HTTP_PORT")
	os.Unsetenv("VIKUNJA_HOST")
	os.Unsetenv("VIKUNJA_TOKEN")

	cfg, err := Load(nil)
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
	os.Setenv("MCP_TRANSPORT", "http")
	defer os.Unsetenv("MCP_TRANSPORT")

	cfg, err := Load(nil)
	require.NoError(t, err)
	assert.Equal(t, TransportHTTP, cfg.Transport)
}

func TestLoad_InvalidTransport(t *testing.T) {
	os.Setenv("MCP_TRANSPORT", "websocket")
	defer os.Unsetenv("MCP_TRANSPORT")

	_, err := Load(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transport type")
}

func TestLoad_HTTPConfig(t *testing.T) {
	os.Setenv("MCP_HTTP_HOST", "0.0.0.0")
	os.Setenv("MCP_HTTP_PORT", "9000")
	os.Setenv("MCP_HTTP_SESSION_TIMEOUT", "1h")
	os.Setenv("MCP_HTTP_STATELESS", "true")
	os.Setenv("MCP_HTTP_READ_TIMEOUT", "60s")
	os.Setenv("MCP_HTTP_WRITE_TIMEOUT", "45s")
	os.Setenv("MCP_HTTP_IDLE_TIMEOUT", "300s")
	defer func() {
		os.Unsetenv("MCP_HTTP_HOST")
		os.Unsetenv("MCP_HTTP_PORT")
		os.Unsetenv("MCP_HTTP_SESSION_TIMEOUT")
		os.Unsetenv("MCP_HTTP_STATELESS")
		os.Unsetenv("MCP_HTTP_READ_TIMEOUT")
		os.Unsetenv("MCP_HTTP_WRITE_TIMEOUT")
		os.Unsetenv("MCP_HTTP_IDLE_TIMEOUT")
	}()

	cfg, err := Load(nil)
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
	os.Setenv("VIKUNJA_HOST", "https://vikunja.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token-123")
	os.Setenv("VIKUNJA_INSECURE", "true")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
		os.Unsetenv("VIKUNJA_INSECURE")
	}()

	cfg, err := Load(nil)
	require.NoError(t, err)

	assert.Equal(t, "https://vikunja.example.com", cfg.Vikunja.Host)
	assert.Equal(t, "test-token-123", cfg.Vikunja.Token)
	assert.True(t, cfg.Vikunja.Insecure)
}

func TestLoad_InvalidHTTPPort(t *testing.T) {
	os.Setenv("MCP_HTTP_PORT", "invalid")
	defer os.Unsetenv("MCP_HTTP_PORT")

	_, err := Load(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HTTP port")
}

func TestLoad_InvalidDuration(t *testing.T) {
	os.Setenv("MCP_HTTP_SESSION_TIMEOUT", "invalid")
	defer os.Unsetenv("MCP_HTTP_SESSION_TIMEOUT")

	_, err := Load(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session timeout")
}

func TestLoad_InvalidBool(t *testing.T) {
	os.Setenv("MCP_HTTP_STATELESS", "invalid")
	defer os.Unsetenv("MCP_HTTP_STATELESS")

	_, err := Load(nil)
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
				assert.Error(t, err)
				assert.Equal(t, vikunja.OutputFormatJSON, result) // Should return default on error
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoad_CLIFormatPrecedence(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("VIKUNJA_OUTPUT_FORMAT")

	// Test CLI flag overrides environment
	json := "json"
	cfg, err := Load(&json)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatJSON, cfg.OutputFormat)

	// Test CLI flag with value
	markdown := "markdown"
	cfg, err = Load(&markdown)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatMarkdown, cfg.OutputFormat)

	// Test CLI flag with invalid value
	invalid := "invalid"
	cfg, err = Load(&invalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --output-format value")

	// Clean up
	os.Unsetenv("VIKUNJA_OUTPUT_FORMAT")
}

func TestLoad_EnvironmentVariableFallback(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("VIKUNJA_OUTPUT_FORMAT")

	// Test with no CLI flag - should use default (Markdown for AI/LLM compatibility)
	cfg, err := Load(nil)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatMarkdown, cfg.OutputFormat)

	// Test with environment variable - no CLI flag
	os.Setenv("VIKUNJA_OUTPUT_FORMAT", "markdown")
	cfg, err = Load(nil)
	require.NoError(t, err)
	assert.Equal(t, vikunja.OutputFormatMarkdown, cfg.OutputFormat)

	// Clean up
	os.Unsetenv("VIKUNJA_OUTPUT_FORMAT")
}
