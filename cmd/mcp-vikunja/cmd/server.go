// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/internal/handlers"
	"github.com/meschbach/mcp-vikunja/internal/transport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	httpHost       string
	httpPort       int
	sessionTimeout string
	stateless      bool
	readTimeout    string
	writeTimeout   string
	idleTimeout    string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start MCP server with HTTP + Streamable transport",
	Long: `Start the MCP server using HTTP + Streamable transport protocol.

This command starts an HTTP server that provides the Model Context Protocol
interface through a streamable HTTP endpoint. This is ideal for web applications
or when you need to serve multiple clients through HTTP.

The server supports session management, configurable timeouts, and can run
in either stateful or stateless mode.`,
	Example: `  mcp-vikunja server
  mcp-vikunja server --http-host 0.0.0.0 --http-port 8080
  mcp-vikunja server --session-timeout 1h --stateless`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Server-specific flags
	serverCmd.Flags().StringVar(&httpHost, "http-host", "localhost", "HTTP server host (env: MCP_HTTP_HOST)")
	serverCmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP server port (env: MCP_HTTP_PORT)")
	serverCmd.Flags().StringVar(&sessionTimeout, "session-timeout", "30m", "Session timeout duration (env: MCP_HTTP_SESSION_TIMEOUT)")
	serverCmd.Flags().BoolVar(&stateless, "stateless", false, "Enable stateless mode (env: MCP_HTTP_STATELESS)")
	serverCmd.Flags().StringVar(&readTimeout, "read-timeout", "30s", "HTTP read timeout (env: MCP_HTTP_READ_TIMEOUT)")
	serverCmd.Flags().StringVar(&writeTimeout, "write-timeout", "30s", "HTTP write timeout (env: MCP_HTTP_WRITE_TIMEOUT)")
	serverCmd.Flags().StringVar(&idleTimeout, "idle-timeout", "2m", "HTTP idle timeout (env: MCP_HTTP_IDLE_TIMEOUT)")
}

func runServer(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("failed to create server configuration: %w", err)
	}

	// Force HTTP transport for server command
	cfg.Transport = config.TransportHTTP

	// Override server-specific settings from flags
	if httpHost != "" {
		cfg.HTTP.Host = httpHost
	}
	if httpPort > 0 {
		cfg.HTTP.Port = httpPort
	}
	if sessionTimeout != "" {
		cfg.HTTP.SessionTimeout, err = time.ParseDuration(sessionTimeout)
		if err != nil {
			return fmt.Errorf("invalid session timeout: %w", err)
		}
	}
	if stateless {
		cfg.HTTP.Stateless = stateless
	}
	if readTimeout != "" {
		cfg.HTTP.ReadTimeout, err = time.ParseDuration(readTimeout)
		if err != nil {
			return fmt.Errorf("invalid read timeout: %w", err)
		}
	}
	if writeTimeout != "" {
		cfg.HTTP.WriteTimeout, err = time.ParseDuration(writeTimeout)
		if err != nil {
			return fmt.Errorf("invalid write timeout: %w", err)
		}
	}
	if idleTimeout != "" {
		cfg.HTTP.IdleTimeout, err = time.ParseDuration(idleTimeout)
		if err != nil {
			return fmt.Errorf("invalid idle timeout: %w", err)
		}
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
	logger.Info("starting MCP server with HTTP transport",
		"address", cfg.HTTP.Address(),
		"session_timeout", cfg.HTTP.SessionTimeout,
		"stateless", cfg.HTTP.Stateless,
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
	err = transportServer.Run(ctx)
	if err != nil {
		// Only log error if it's not context cancellation (which is expected)
		if ctx.Err() != context.Canceled {
			logger.Error("server error", "error", err)
			return err
		}
	}
	logger.Info("server shutdown completed")
	return nil
}
