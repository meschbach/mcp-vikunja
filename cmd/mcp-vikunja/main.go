// Package main is the entry point for the MCP Vikunja server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/internal/handlers"
	"github.com/meschbach/mcp-vikunja/internal/transport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	if err := run(ctx, logger); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log configuration (without sensitive data)
	logger.Info("starting MCP server",
		"transport", cfg.Transport,
		"version", "0.1.0",
	)

	if cfg.Transport == config.TransportHTTP {
		logger.Info("HTTP transport enabled",
			"address", cfg.HTTP.Address(),
			"session_timeout", cfg.HTTP.SessionTimeout,
			"stateless", cfg.HTTP.Stateless,
		)
	}

	// Create MCP server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-vikunja",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{},
	)

	// Register Vikunja tool handlers
	handlers.Register(s)

	// Create transport server
	transportServer, err := transport.CreateTransportServer(s, cfg)
	if err != nil {
		return fmt.Errorf("failed to create transport server: %w", err)
	}

	// Start the server
	return transportServer.Run(ctx)
}
