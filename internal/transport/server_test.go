package transport

import (
	"context"
	"testing"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTransportServer_Stdio(t *testing.T) {
	cfg := &config.Config{
		Transport: config.TransportStdio,
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	server, err := CreateTransportServer(mcpServer, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Verify it's a StdioServer
	_, ok := server.(*StdioServer)
	assert.True(t, ok)
}

func TestCreateTransportServer_HTTP(t *testing.T) {
	cfg := &config.Config{
		Transport: config.TransportHTTP,
		HTTP: config.HTTPConfig{
			Host: "localhost",
			Port: 8080,
		},
		Vikunja: config.VikunjaConfig{
			Host:  "https://example.com",
			Token: "test-token",
		},
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	server, err := CreateTransportServer(mcpServer, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Verify it's an HTTPServer
	_, ok := server.(*HTTPServer)
	assert.True(t, ok)
}

func TestCreateTransportServer_Unsupported(t *testing.T) {
	cfg := &config.Config{
		Transport: "websocket",
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	server, err := CreateTransportServer(mcpServer, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported transport type")
	assert.Nil(t, server)
}

func TestStdioServer_Run(t *testing.T) {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	server := &StdioServer{
		server: mcpServer,
	}

	// Note: We can't actually test the stdio transport without complex setup,
	// but we can verify the structure is correct
	assert.NotNil(t, server.server)
}

func TestHTTPServer_Run_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Transport: config.TransportHTTP,
		HTTP: config.HTTPConfig{
			Host:           "localhost",
			Port:           0, // Use random available port
			SessionTimeout: 30 * time.Second,
			Stateless:      false,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			IdleTimeout:    10 * time.Second,
		},
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	server := &HTTPServer{
		server: mcpServer,
		config: cfg,
	}

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run should return nil after graceful shutdown
	err := server.Run(ctx)
	assert.NoError(t, err)
}

func TestHTTPServer_Run_BoundPort(t *testing.T) {
	cfg := &config.Config{
		Transport: config.TransportHTTP,
		HTTP: config.HTTPConfig{
			Host:           "localhost",
			Port:           0, // Use random available port
			SessionTimeout: 30 * time.Second,
			Stateless:      false,
		},
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	server := &HTTPServer{
		server: mcpServer,
		config: cfg,
	}

	// Test that we can create the server (actual network binding is tested in integration tests)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.config)
}
