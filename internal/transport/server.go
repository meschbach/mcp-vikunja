// Package transport provides transport implementations for the MCP Vikunja server.
package transport

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents an MCP server that can run on different transports.
type Server interface {
	Run(ctx context.Context) error
}

// TransportFactory creates a transport server based on configuration.
func CreateTransportServer(mcpServer *mcp.Server, cfg *config.Config) (Server, error) {
	switch cfg.Transport {
	case config.TransportStdio:
		return &StdioServer{
			server: mcpServer,
		}, nil

	case config.TransportHTTP:
		return &HTTPServer{
			server: mcpServer,
			config: cfg,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", cfg.Transport)
	}
}

// StdioServer implements the stdio transport.
type StdioServer struct {
	server *mcp.Server
}

// Run starts the MCP server with stdio transport.
func (s *StdioServer) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

// HTTPServer implements the HTTP transport using the streamable protocol.
type HTTPServer struct {
	server *mcp.Server
	config *config.Config
}

// Run starts the MCP server with HTTP transport.
func (s *HTTPServer) Run(ctx context.Context) error {
	// Create the streamable HTTP handler
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return s.server
		},
		&mcp.StreamableHTTPOptions{
			SessionTimeout: s.config.HTTP.SessionTimeout,
			Stateless:      s.config.HTTP.Stateless,
		},
	)

	// Create HTTP server with proper timeouts
	httpServer := &http.Server{
		Addr:         s.config.HTTP.Address(),
		Handler:      handler,
		ReadTimeout:  s.config.HTTP.ReadTimeout,
		WriteTimeout: s.config.HTTP.WriteTimeout,
		IdleTimeout:  s.config.HTTP.IdleTimeout,
	}

	// Start the HTTP server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				errChan <- fmt.Errorf("HTTP server failed: %w", err)
			} else {
				// Send nil error to signal graceful shutdown completion
				errChan <- nil
			}
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown failed: %w", err)
		}

		// Wait for the server goroutine to finish
		err := <-errChan
		if err != nil {
			return err
		}
		return nil

	case err := <-errChan:
		// Server failed to start or shutdown completed
		return err
	}
}
