// Package main is the entry point for the MCP Vikunja server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/meschbach/mcp-vikunja/internal/handlers"
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
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-vikunja",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{},
	)

	// Register Vikunja tool handlers
	handlers.Register(s)

	logger.Info("starting MCP server")

	return s.Run(ctx, &mcp.StdioTransport{})
}
