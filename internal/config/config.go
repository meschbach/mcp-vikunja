// Package config provides configuration management for the MCP Vikunja server.
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// TransportType defines the available transport mechanisms.
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

// Config represents the complete configuration for the MCP Vikunja server.
type Config struct {
	Transport TransportType `json:"transport"`
	HTTP      HTTPConfig    `json:"http"`
	Vikunja   VikunjaConfig `json:"vikunja"`
}

// HTTPConfig contains HTTP server specific configuration.
type HTTPConfig struct {
	Host           string        `json:"host"`
	Port           int           `json:"port"`
	SessionTimeout time.Duration `json:"session_timeout"`
	Stateless      bool          `json:"stateless"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
}

// VikunjaConfig contains Vikunja client specific configuration.
type VikunjaConfig struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

// Load loads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Transport: TransportStdio, // Default to stdio for backward compatibility
		HTTP: HTTPConfig{
			Host:           "localhost",
			Port:           8080,
			SessionTimeout: 30 * time.Minute,
			Stateless:      false,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			IdleTimeout:    120 * time.Second,
		},
	}

	// Load transport type
	if transport := os.Getenv("MCP_TRANSPORT"); transport != "" {
		switch transport {
		case string(TransportStdio):
			cfg.Transport = TransportStdio
		case string(TransportHTTP):
			cfg.Transport = TransportHTTP
		default:
			return nil, fmt.Errorf("invalid transport type: %s (must be 'stdio' or 'http')", transport)
		}
	}

	// Load HTTP configuration
	if err := loadHTTPConfig(&cfg.HTTP); err != nil {
		return nil, fmt.Errorf("failed to load HTTP config: %w", err)
	}

	// Load Vikunja configuration
	if err := loadVikunjaConfig(&cfg.Vikunja); err != nil {
		return nil, fmt.Errorf("failed to load Vikunja config: %w", err)
	}

	return cfg, nil
}

// loadHTTPConfig loads HTTP-specific configuration from environment variables.
func loadHTTPConfig(cfg *HTTPConfig) error {
	if host := os.Getenv("MCP_HTTP_HOST"); host != "" {
		cfg.Host = host
	}

	if port := os.Getenv("MCP_HTTP_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid HTTP port: %s", port)
		}
		cfg.Port = p
	}

	if timeout := os.Getenv("MCP_HTTP_SESSION_TIMEOUT"); timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid session timeout: %s", timeout)
		}
		cfg.SessionTimeout = d
	}

	if stateless := os.Getenv("MCP_HTTP_STATELESS"); stateless != "" {
		s, err := strconv.ParseBool(stateless)
		if err != nil {
			return fmt.Errorf("invalid stateless flag: %s", stateless)
		}
		cfg.Stateless = s
	}

	if timeout := os.Getenv("MCP_HTTP_READ_TIMEOUT"); timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid read timeout: %s", timeout)
		}
		cfg.ReadTimeout = d
	}

	if timeout := os.Getenv("MCP_HTTP_WRITE_TIMEOUT"); timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid write timeout: %s", timeout)
		}
		cfg.WriteTimeout = d
	}

	if timeout := os.Getenv("MCP_HTTP_IDLE_TIMEOUT"); timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid idle timeout: %s", timeout)
		}
		cfg.IdleTimeout = d
	}

	return nil
}

// loadVikunjaConfig loads Vikunja-specific configuration from environment variables.
func loadVikunjaConfig(cfg *VikunjaConfig) error {
	if host := os.Getenv("VIKUNJA_HOST"); host != "" {
		cfg.Host = host
	}

	if token := os.Getenv("VIKUNJA_TOKEN"); token != "" {
		cfg.Token = token
	}

	return nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	switch c.Transport {
	case TransportStdio, TransportHTTP:
		// Valid transport types
	default:
		return fmt.Errorf("invalid transport type: %s", c.Transport)
	}

	if c.Transport == TransportHTTP {
		if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
			return fmt.Errorf("invalid HTTP port: %d (must be 1-65535)", c.HTTP.Port)
		}

		if c.HTTP.Host == "" {
			return fmt.Errorf("HTTP host cannot be empty")
		}

		// Validate that host:port is not already in use
		address := net.JoinHostPort(c.HTTP.Host, strconv.Itoa(c.HTTP.Port))
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			conn.Close()
			return fmt.Errorf("HTTP address %s is already in use", address)
		}
	}

	// Validate Vikunja configuration
	if c.Vikunja.Host == "" {
		return fmt.Errorf("VIKUNJA_HOST is required")
	}

	if c.Vikunja.Token == "" {
		return fmt.Errorf("VIKUNJA_TOKEN is required")
	}

	return nil
}

// Address returns the full HTTP address in host:port format.
func (c *HTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
