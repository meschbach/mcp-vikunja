// Package logging provides structured logging configuration for the MCP Vikunja server.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level represents log levels.
type Level string

// LevelDebug represents debug level.
// LevelInfo represents info level.
// LevelWarn represents warn level.
// LevelError represents error level.
const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Format represents log output formats.
type Format string

// FormatJSON represents JSON output format.
// FormatText represents text output format.
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
	output, err := openOutput(cfg.Output)
	if err != nil {
		return nil, err
	}

	level := parseLevel(cfg.Level)
	handler := newHandler(output, level, cfg.Format)

	return slog.New(handler), nil
}

func openOutput(path string) (io.Writer, error) {
	switch path {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		// #nosec G304 - path is configured by operator via LOG_OUTPUT env var
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		return file, nil
	}
}

func parseLevel(lvl Level) slog.Level {
	switch lvl {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newHandler(output io.Writer, level slog.Level, format Format) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}

	switch format {
	case FormatJSON:
		return slog.NewJSONHandler(output, opts)
	case FormatText:
		return slog.NewTextHandler(output, opts)
	default:
		return slog.NewJSONHandler(output, opts)
	}
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
func RedactSensitive(key, value string) slog.Attr {
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
