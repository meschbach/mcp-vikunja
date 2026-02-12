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

The MCP server now uses explicit subcommands for different transport modes:

**Option A: Stdio Transport (for local CLI tools)**
```bash
# Default Markdown output (recommended for AI/LLMs)
./bin/mcp-vikunja stdio

# JSON output (for legacy compatibility)
./bin/mcp-vikunja stdio --output-format json

# Both formats together  
./bin/mcp-vikunja stdio -o both

# Environment variable setting
VIKUNJA_OUTPUT_FORMAT=json ./bin/mcp-vikunja stdio
```

**Option B: HTTP Transport (for web applications and remote access)**
```bash
# Default Markdown output (recommended for AI/LLMs)
./bin/mcp-vikunja server

# JSON output (for legacy compatibility)
./bin/mcp-vikunja server --output-format json
```

**Option C: Show Help**
```bash
./bin/mcp-vikunja
# Shows available commands and usage examples
```

## Transport Modes

### Stdio Transport
- **Best for**: Local CLI tools, development environments
- **Usage**: `./bin/mcp-vikunja stdio`
- **Features**: Communicates over standard I/O
- **Example**: `VIKUNJA_HOST=https://example.com VIKUNJA_TOKEN=token ./bin/mcp-vikunja stdio`

### HTTP Transport
- **Best for**: Web applications, remote access, multiple clients
- **Usage**: `./bin/mcp-vikunja server`
- **Features**: Session management, concurrent connections, streamable HTTP
- **Example**: `VIKUNJA_HOST=https://example.com VIKUNJA_TOKEN=token ./bin/mcp-vikunja server --http-port 9000`

## CLI Commands

The MCP server provides a rich command-line interface:

### Server Commands
- `mcp-vikunja server` - Start HTTP + Streamable transport server
- `mcp-vikunja stdio` - Start stdio transport server

### Configuration Commands
- `mcp-vikunja config show` - Display current configuration
- `mcp-vikunja config validate` - Validate configuration and test connectivity

### Utility Commands
- `mcp-vikunja health` - Test Vikunja connection
- `mcp-vikunja version` - Show version and build information
- `mcp-vikunja help` - Show comprehensive help

### Examples
```bash
# Start HTTP server with custom settings
./bin/mcp-vikunja server --http-host 0.0.0.0 --http-port 9000 --stateless

# Start stdio server with verbose logging
./bin/mcp-vikunja stdio --verbose --vikunja-host https://example.com

# Show current configuration
./bin/mcp-vikunja config show --format json

# Validate configuration
./bin/mcp-vikunja config validate

# Test Vikunja connectivity
./bin/mcp-vikunja health --vikunja-host https://example.com --vikunja-token token
```

## Configuration

### Required Environment Variables
- `VIKUNJA_HOST` - Your Vikunja instance URL
- `VIKUNJA_TOKEN` - Your Vikunja API token

### Optional Output Format Configuration
| Variable/Flag | Default | Description |
|---------------|---------|-------------|
| `VIKUNJA_OUTPUT_FORMAT` | `markdown` | Output format: json, markdown, both |
| `--output-format` / `-o` | `markdown` | CLI flag that overrides VIKUNJA_OUTPUT_FORMAT |

**Output Format Precedence**: CLI flag > Environment variable > Default (markdown)

**Output Format Options**:
- `json` - Original JSON output (for legacy compatibility)
- `markdown` - Human-readable Markdown output with tables and formatting (recommended for AI/LLMs)
- `both` - Combined JSON and Markdown output

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
- `list_buckets` - List all buckets in a project view (defaults to Inbox project and Kanban view)
- `list_projects` - List all available projects
- `create_task` - Create new tasks (coming soon)

## Example Usage

### Testing with cURL (HTTP Transport)

```bash
# Start server with HTTP transport
VIKUNJA_HOST=https://example.com VIKUNJA_TOKEN=your-token ./bin/mcp-vikunja server &

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
  -e VIKUNJA_HOST=https://your-vikunja.com \
  -e VIKUNJA_TOKEN=your-token \
  mcp-vikunja server --http-host 0.0.0.0
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