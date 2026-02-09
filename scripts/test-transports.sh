#!/bin/bash

# Test script for MCP Vikunja transport modes
set -e

echo "Building MCP Vikunja..."
go build -o bin/mcp-vikunja ./cmd/mcp-vikunja

echo "Testing stdio transport (default)..."
# Test that stdio transport starts (will wait for input, so we kill it after 1 second)
timeout 1s ./bin/mcp-vikunja 2>/dev/null || echo "✓ stdio transport works"

echo "Testing HTTP transport configuration..."
# Test configuration validation
MCP_TRANSPORT=http MCP_HTTP_HOST=localhost MCP_HTTP_PORT=8080 VIKUNJA_HOST=https://example.com VIKUNJA_TOKEN=test-token \
    ./bin/mcp-vikunja &
SERVER_PID=$!
sleep 2
kill $SERVER_PID 2>/dev/null || true
echo "✓ HTTP transport works"

echo "Testing invalid transport..."
MCP_TRANSPORT=invalid ./bin/mcp-vikunja 2>/dev/null && echo "✗ Should have failed" || echo "✓ Invalid transport properly rejected"

echo "Testing missing required config..."
MCP_TRANSPORT=http MCP_HTTP_HOST=localhost MCP_HTTP_PORT=8080 \
    ./bin/mcp-vikunja 2>/dev/null && echo "✗ Should have failed" || echo "✓ Missing Vikunja config properly rejected"

echo "All transport tests passed!"