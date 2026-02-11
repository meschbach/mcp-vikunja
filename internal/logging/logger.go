// Package logging provides structured logging configuration for the MCP Vikunja server.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level represents log levels
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Format represents log output formats
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config holds logging configuration
type Config struct {
	Level  Level
	Format Format
	Output string // stdout, stderr, or file path
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: "stdout",
	}
}

// LoadConfig loads logging configuration from environment variables
func LoadConfig() Config {
	cfg := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Level = Level(strings.ToLower(level))
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Format = Format(strings.ToLower(format))
	}

	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		cfg.Output = output
	}

	return cfg
}

// NewLogger creates a new slog.Logger based on configuration
func NewLogger(cfg Config) (*slog.Logger, error) {
	// Determine output writer
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	// Convert our Level to slog.Level
	var level slog.Level
	switch cfg.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	case FormatText:
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	return slog.New(handler), nil
}

// MustNewLogger creates a new logger or panics on error
func MustNewLogger(cfg Config) *slog.Logger {
	logger, err := NewLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// RedactSensitive redacts sensitive values from log entries
func RedactSensitive(key string, value string) slog.Attr {
	// List of sensitive fields
	sensitiveFields := []string{
		"token", "password", "secret", "api_key", "auth",
		"authorization", "credential", "private_key",
	}

	lowerKey := strings.ToLower(key)
	for _, field := range sensitiveFields {
		if strings.Contains(lowerKey, field) {
			return slog.String(key, "***REDACTED***")
		}
	}

	return slog.String(key, value)
}

// WithComponent adds a component field to log entries
func WithComponent(logger *slog.Logger, component string) *slog.Logger {
	return logger.With("component", component)
}

// WithRequestID adds a request ID field to log entries
func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With("request_id", requestID)
}
