package logging

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, LevelInfo, cfg.Level)
	assert.Equal(t, FormatJSON, cfg.Format)
	assert.Equal(t, "stdout", cfg.Output)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envLevel    string
		envFormat   string
		envOutput   string
		expectedCfg Config
	}{
		{
			name:        "default values",
			expectedCfg: DefaultConfig(),
		},
		{
			name:        "debug level",
			envLevel:    "debug",
			expectedCfg: Config{Level: LevelDebug, Format: FormatJSON, Output: "stdout"},
		},
		{
			name:        "text format",
			envFormat:   "text",
			expectedCfg: Config{Level: LevelInfo, Format: FormatText, Output: "stdout"},
		},
		{
			name:        "custom output",
			envOutput:   "/var/log/app.log",
			expectedCfg: Config{Level: LevelInfo, Format: FormatJSON, Output: "/var/log/app.log"},
		},
		{
			name:        "all custom",
			envLevel:    "error",
			envFormat:   "text",
			envOutput:   "stderr",
			expectedCfg: Config{Level: LevelError, Format: FormatText, Output: "stderr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			if tt.envLevel != "" {
				os.Setenv("LOG_LEVEL", tt.envLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}
			if tt.envFormat != "" {
				os.Setenv("LOG_FORMAT", tt.envFormat)
				defer os.Unsetenv("LOG_FORMAT")
			}
			if tt.envOutput != "" {
				os.Setenv("LOG_OUTPUT", tt.envOutput)
				defer os.Unsetenv("LOG_OUTPUT")
			}

			cfg := LoadConfig()
			assert.Equal(t, tt.expectedCfg, cfg)
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name:   "json stdout",
			config: Config{Level: LevelInfo, Format: FormatJSON, Output: "stdout"},
		},
		{
			name:   "text stderr",
			config: Config{Level: LevelDebug, Format: FormatText, Output: "stderr"},
		},
		{
			name:   "debug level",
			config: Config{Level: LevelDebug, Format: FormatJSON, Output: "stdout"},
		},
		{
			name:   "warn level",
			config: Config{Level: LevelWarn, Format: FormatJSON, Output: "stdout"},
		},
		{
			name:   "error level",
			config: Config{Level: LevelError, Format: FormatJSON, Output: "stdout"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestNewLogger_InvalidFile(t *testing.T) {
	// Try to create a logger with an invalid file path
	config := Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: "/nonexistent/path/to/file.log",
	}

	logger, err := NewLogger(config)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestMustNewLogger(t *testing.T) {
	config := Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: "stdout",
	}

	// Should not panic
	logger := MustNewLogger(config)
	assert.NotNil(t, logger)
}

func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		expected string
	}{
		{
			key:      "token",
			value:    "secret-token-123",
			expected: "***REDACTED***",
		},
		{
			key:      "api_key",
			value:    "my-api-key",
			expected: "***REDACTED***",
		},
		{
			key:      "password",
			value:    "my-password",
			expected: "***REDACTED***",
		},
		{
			key:      "auth_token",
			value:    "bearer-token",
			expected: "***REDACTED***",
		},
		{
			key:      "normal_field",
			value:    "normal-value",
			expected: "normal-value",
		},
		{
			key:      "title",
			value:    "My Task",
			expected: "My Task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			attr := RedactSensitive(tt.key, tt.value)
			assert.Equal(t, tt.key, attr.Key)
			assert.Equal(t, tt.expected, attr.Value.String())
		})
	}
}

func TestWithComponent(t *testing.T) {
	config := Config{Level: LevelInfo, Format: FormatJSON, Output: "stdout"}
	logger, err := NewLogger(config)
	require.NoError(t, err)

	componentLogger := WithComponent(logger, "test-component")
	assert.NotNil(t, componentLogger)
}

func TestWithRequestID(t *testing.T) {
	config := Config{Level: LevelInfo, Format: FormatJSON, Output: "stdout"}
	logger, err := NewLogger(config)
	require.NoError(t, err)

	requestLogger := WithRequestID(logger, "req-123")
	assert.NotNil(t, requestLogger)
}

func TestLevelConstants(t *testing.T) {
	assert.Equal(t, Level("debug"), LevelDebug)
	assert.Equal(t, Level("info"), LevelInfo)
	assert.Equal(t, Level("warn"), LevelWarn)
	assert.Equal(t, Level("error"), LevelError)
}

func TestFormatConstants(t *testing.T) {
	assert.Equal(t, Format("json"), FormatJSON)
	assert.Equal(t, Format("text"), FormatText)
}

func TestNewLogger_UnsupportedLevel(t *testing.T) {
	config := Config{
		Level:  Level("unsupported"),
		Format: FormatJSON,
		Output: "stdout",
	}

	// Should default to Info level
	logger, err := NewLogger(config)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_UnsupportedFormat(t *testing.T) {
	config := Config{
		Level:  LevelInfo,
		Format: Format("unsupported"),
		Output: "stdout",
	}

	// Should default to JSON format
	logger, err := NewLogger(config)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}
