// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

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
	"github.com/spf13/cobra"
)

// stdioCmd represents the stdio command
var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "Start MCP server with stdio transport",
	Long: `Start the MCP server using stdio transport protocol.

This command starts the MCP server that communicates through standard input/output.
This is the standard transport mode for MCP clients and is ideal for command-line
integrations or when the server is launched directly by the MCP client.

The stdio transport uses the Model Context Protocol over stdin/stdout for
bidirectional communication with the client.`,
	Example: `  mcp-vikunja stdio
  mcp-vikunja stdio --vikunja-host https://vikunja.example.com
  mcp-vikunja stdio --verbose`,
	RunE: runStdio,
}

func init() {
	rootCmd.AddCommand(stdioCmd)
}

func runStdio(cmd *cobra.Command, args []string) error {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	// Get output format from CLI flag
	var cliFormat *string
	if format := cmd.Flag("output-format").Value.String(); format != "" {
		cliFormat = &format
	}

	// Create configuration
	cfg, err := config.Load(cliFormat)
	if err != nil {
		return fmt.Errorf("failed to create stdio configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	// Log configuration
	logger.Info("starting MCP server with stdio transport",
		"vikunja_host", cfg.Vikunja.Host,
		"version", "0.1.0",
	)

	// Create MCP server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-vikunja",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{},
	)

	// Register Vikunja tool handlers
	handlers.Register(s, cfg)

	// Create transport server
	transportServer, err := transport.CreateTransportServer(s, cfg)
	if err != nil {
		return fmt.Errorf("failed to create transport server: %w", err)
	}

	// Start the server
	return transportServer.Run(ctx)
}
