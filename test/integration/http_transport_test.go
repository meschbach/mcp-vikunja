//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPort = 18080

func TestHTTPTransport_Integration(t *testing.T) {
	// Set environment variables for HTTP transport
	os.Setenv("MCP_TRANSPORT", "http")
	os.Setenv("MCP_HTTP_HOST", "localhost")
	os.Setenv("MCP_HTTP_PORT", fmt.Sprintf("%d", testPort))
	os.Setenv("VIKUNJA_HOST", "https://vikunja.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token-123")
	defer func() {
		os.Unsetenv("MCP_TRANSPORT")
		os.Unsetenv("MCP_HTTP_HOST")
		os.Setenv("MCP_HTTP_PORT", "")
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
	}()

	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err)
	err = cfg.Validate()
	require.NoError(t, err)

	assert.Equal(t, config.TransportHTTP, cfg.Transport)
	assert.Equal(t, fmt.Sprintf("localhost:%d", testPort), cfg.HTTP.Address())

	// Create MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-mcp-vikunja",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	// Create HTTP handler
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return server
		},
		&mcp.StreamableHTTPOptions{
			SessionTimeout: 30 * time.Second,
			Stateless:      false,
		},
	)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	serverErrChan := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	defer func() {
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		<-serverErrChan
	}()

	// Test health check - simple request to verify server is running
	baseURL := fmt.Sprintf("http://localhost:%d", testPort)

	// Test MCP protocol communication
	client := &http.Client{Timeout: 10 * time.Second}

	// Send initialize request
	initRequest := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	reqBody, err := json.Marshal(initRequest)
	require.NoError(t, err)

	resp, err := client.Post(baseURL, "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var initResponse mcp.JSONRPCResponse
	err = json.Unmarshal(body, &initResponse)
	require.NoError(t, err)

	assert.Equal(t, "2.0", initResponse.JSONRPC)
	assert.Equal(t, float64(1), initResponse.ID)
	assert.NotNil(t, initResponse.Result)

	// Test tools/list request
	toolsRequest := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	reqBody, err = json.Marshal(toolsRequest)
	require.NoError(t, err)

	resp2, err := client.Post(baseURL, "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	var toolsResponse mcp.JSONRPCResponse
	err = json.Unmarshal(body2, &toolsResponse)
	require.NoError(t, err)

	assert.Equal(t, "2.0", toolsResponse.JSONRPC)
	assert.Equal(t, float64(2), toolsResponse.ID)

	// Verify the response contains tools
	result, ok := toolsResponse.Result.(map[string]interface{})
	require.True(t, ok)
	tools, exists := result["tools"]
	assert.True(t, exists, "Response should contain tools")

	toolsList, ok := tools.([]interface{})
	assert.True(t, ok, "Tools should be an array")
	assert.Greater(t, len(toolsList), 0, "Should have at least one tool")

	select {
	case err := <-serverErrChan:
		t.Fatalf("Server error: %v", err)
	default:
		// Server is running normally
	}
}
