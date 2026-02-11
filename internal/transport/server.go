// Package transport provides transport implementations for the MCP Vikunja server.
package transport

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/internal/health"
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
	server        *mcp.Server
	config        *config.Config
	healthChecker *health.HealthChecker
}

// Run starts the MCP server with HTTP transport.
func (s *HTTPServer) Run(ctx context.Context) error {
	// Create the streamable HTTP handler
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return s.server
		},
		&mcp.StreamableHTTPOptions{
			SessionTimeout: s.config.HTTP.SessionTimeout,
			Stateless:      s.config.HTTP.Stateless,
		},
	)

	// Create mux and register handlers
	mux := http.NewServeMux()

	// Register MCP handler
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)

	// Register health check handlers if health checker is configured
	if s.healthChecker != nil {
		mux.HandleFunc("/health", s.healthChecker.HTTPHandler(""))
		mux.HandleFunc("/health/live", s.healthChecker.HTTPHandler(health.CheckTypeLiveness))
		mux.HandleFunc("/health/ready", s.healthChecker.HTTPHandler(health.CheckTypeReadiness))
	}

	// Create HTTP server with proper timeouts, defaulting to port 8080
	addr := s.config.HTTP.Address()
	if addr == "" || addr == ":0" {
		addr = ":8080"
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
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
		// Graceful shutdown - this is an expected condition, not an error
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			// Log shutdown errors but don't treat them as failures since shutdown is expected
			return nil
		}

		// Wait for the server goroutine to finish
		<-errChan // Ignore any error from the goroutine since shutdown is expected
		return nil

	case err := <-errChan:
		// Server failed to start or unexpected error occurred
		return err
	}
}

// SetHealthChecker sets the health checker for the HTTP server
func (s *HTTPServer) SetHealthChecker(hc *health.HealthChecker) {
	s.healthChecker = hc
}
