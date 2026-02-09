# MCP Vikunja

A Model Context Protocol (MCP) server for Vikunja task management integration with support for both stdio and HTTP transports.

## Overview

This project provides an MCP server that allows LLMs and AI assistants to interact with Vikunja task management through standardized tool calls. It supports both local CLI usage (stdio) and remote web deployment (HTTP).

## Features

- 📋 List tasks from Vikunja projects and views
- 🔍 Get detailed task information including bucket placement
- 📦 List project buckets and organize tasks
- 🏗️ List all available projects
- 🚀 Multiple transport modes: stdio (CLI) and HTTP (web)
- 🔧 Configurable HTTP server with session management
- ✅ Built with the official MCP Go SDK

## Installation

```bash
go get github.com/meschbach/mcp-vikunja
```

## Quick Start

### 1. Build the Server

```bash
go build -o bin/mcp-vikunja ./cmd/mcp-vikunja
```

### 2. Configure Vikunja Connection

```bash
export VIKUNJA_HOST="https://your-vikunja-instance.com"
export VIKUNJA_TOKEN="your-api-token"
```

### 3. Run the Server

**Option A: Stdio Transport (default - for local CLI tools)**
```bash
./bin/mcp-vikunja
```

**Option B: HTTP Transport (for web applications and remote access)**
```bash
export MCP_TRANSPORT="http"
./bin/mcp-vikunja
# Server will start on http://localhost:8080
```

## Transport Modes

### Stdio Transport (Default)
- **Best for**: Local CLI tools, development environments
- **Usage**: Default mode, communicates over standard I/O
- **Example**: `./bin/mcp-vikunja`

### HTTP Transport
- **Best for**: Web applications, remote access, multiple clients
- **Usage**: Set `MCP_TRANSPORT=http`
- **Features**: Session management, concurrent connections
- **Example**: `MCP_TRANSPORT=http MCP_HTTP_PORT=9000 ./bin/mcp-vikunja`

## Configuration

### Required Environment Variables
- `VIKUNJA_HOST` - Your Vikunja instance URL
- `VIKUNJA_TOKEN` - Your Vikunja API token

### Optional HTTP Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT` | `stdio` | Transport type: `stdio` or `http` |
| `MCP_HTTP_HOST` | `localhost` | Server bind address |
| `MCP_HTTP_PORT` | `8080` | Server port |
| `MCP_HTTP_SESSION_TIMEOUT` | `30m` | Session timeout |
| `MCP_HTTP_STATELESS` | `false` | Disable session tracking |

## Available Tools

The server provides the following MCP tools:

- `list_tasks` - List tasks from projects with filtering options
- `get_task` - Get detailed task information including bucket placement
- `list_buckets` - List all buckets in a project view
- `list_projects` - List all available projects
- `create_task` - Create new tasks (coming soon)

## Example Usage

### Testing with cURL (HTTP Transport)

```bash
# Start server with HTTP transport
MCP_TRANSPORT=http VIKUNJA_HOST=https://example.com VIKUNJA_TOKEN=your-token ./bin/mcp-vikunja &

# Test the server
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }'
```

### Docker Deployment

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-vikunja ./cmd/mcp-vikunja

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-vikunja .
CMD ["./mcp-vikunja"]
```

```bash
# Build and run with Docker
docker build -t mcp-vikunja .
docker run -p 8080:8080 \
  -e MCP_TRANSPORT=http \
  -e MCP_HTTP_HOST=0.0.0.0 \
  -e VIKUNJA_HOST=https://your-vikunja.com \
  -e VIKUNJA_TOKEN=your-token \
  mcp-vikunja
```

## Development

See [AGENTS.md](AGENTS.md) for coding guidelines and commands.

### Testing

```bash
# Run all tests
go test ./...

# Run integration tests (requires Vikunja instance)
go test -tags=integration ./test/integration/...

# Test both transport modes
./scripts/test-transports.sh
```

## Documentation

- [Transport Configuration Guide](docs/transport-configuration.md) - Detailed transport setup and deployment options
- [AGENTS.md](AGENTS.md) - Development guidelines and project standards

## License

MIT