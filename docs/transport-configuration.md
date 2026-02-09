# MCP Vikunja Server Transport Configuration

The MCP Vikunja server supports multiple transport mechanisms for communication with clients. This document explains how to configure and use different transport modes.

## Transport Modes

### 1. Stdio Transport (Default)

The stdio transport is the default and most common way to run MCP servers. It communicates over standard input/output streams.

**Usage:**
```bash
# Environment variables required for all transports
export VIKUNJA_HOST="https://your-vikunja-instance.com"
export VIKUNJA_TOKEN="your-api-token"

# Run with stdio transport (default)
./mcp-vikunja
```

**Best for:**
- Local CLI tools
- Development environments
- MCP clients that launch servers as subprocesses

### 2. HTTP Transport (Streamable)

The HTTP transport enables the server to run as a web service, supporting multiple concurrent client connections and remote access.

**Usage:**
```bash
# Environment variables
export VIKUNJA_HOST="https://your-vikunja-instance.com"
export VIKUNJA_TOKEN="your-api-token"
export MCP_TRANSPORT="http"

# Optional HTTP configuration (defaults shown)
export MCP_HTTP_HOST="localhost"          # Server host
export MCP_HTTP_PORT="8080"              # Server port
export MCP_HTTP_SESSION_TIMEOUT="30m"    # Session timeout
export MCP_HTTP_STATELESS="false"       # Session tracking
export MCP_HTTP_READ_TIMEOUT="30s"      # HTTP read timeout
export MCP_HTTP_WRITE_TIMEOUT="30s"      # HTTP write timeout
export MCP_HTTP_IDLE_TIMEOUT="120s"      # HTTP idle timeout

# Start the server
./mcp-vikunja
```

**Best for:**
- Web applications
- Remote access scenarios
- Multiple concurrent clients
- Cloud deployments

## Configuration Reference

### Transport Selection

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MCP_TRANSPORT` | No | `stdio` | Transport type: `stdio` or `http` |

### HTTP Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MCP_HTTP_HOST` | No | `localhost` | Server bind address |
| `MCP_HTTP_PORT` | No | `8080` | Server port (1-65535) |
| `MCP_HTTP_SESSION_TIMEOUT` | No | `30m` | Session timeout duration |
| `MCP_HTTP_STATELESS` | No | `false` | Disable session tracking |
| `MCP_HTTP_READ_TIMEOUT` | No | `30s` | HTTP read timeout |
| `MCP_HTTP_WRITE_TIMEOUT` | No | `30s` | HTTP write timeout |
| `MCP_HTTP_IDLE_TIMEOUT` | No | `120s` | HTTP idle timeout |

### Vikunja Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `VIKUNJA_HOST` | Yes | Vikunja instance URL |
| `VIKUNJA_TOKEN` | Yes | Vikunja API token |

## Deployment Examples

### Docker Compose (HTTP Transport)

```yaml
version: '3.8'
services:
  mcp-vikunja:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MCP_TRANSPORT=http
      - MCP_HTTP_HOST=0.0.0.0
      - VIKUNJA_HOST=${VIKUNJA_HOST}
      - VIKUNJA_TOKEN=${VIKUNJA_TOKEN}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-vikunja
spec:
  replicas: 2
  selector:
    matchLabels:
      app: mcp-vikunja
  template:
    metadata:
      labels:
        app: mcp-vikunja
    spec:
      containers:
      - name: mcp-vikunja
        image: mcp-vikunja:latest
        ports:
        - containerPort: 8080
        env:
        - name: MCP_TRANSPORT
          value: "http"
        - name: MCP_HTTP_HOST
          value: "0.0.0.0"
        - name: VIKUNJA_HOST
          valueFrom:
            secretKeyRef:
              name: vikunja-config
              key: host
        - name: VIKUNJA_TOKEN
          valueFrom:
            secretKeyRef:
              name: vikunja-config
              key: token
---
apiVersion: v1
kind: Service
metadata:
  name: mcp-vikunja-service
spec:
  selector:
    app: mcp-vikunja
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Client Connection Examples

### MCP Client (HTTP)

```typescript
import { MCPServer } from '@modelcontextprotocol/sdk/server/index.js';

// Connect to HTTP transport
const response = await fetch('http://localhost:8080', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'initialize',
    params: {
      protocolVersion: '2025-03-26',
      capabilities: { tools: {} },
      clientInfo: { name: 'test-client', version: '1.0.0' }
    }
  })
});

const result = await response.json();
```

### cURL Test

```bash
# Initialize connection
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "capabilities": {"tools": {}},
      "clientInfo": {"name": "test-client", "version": "1.0.0"}
    }
  }'

# List available tools
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'
```

## Security Considerations

### HTTP Transport
- By default, the server binds to `localhost` only
- For production deployments, use proper authentication/authorization
- Consider using HTTPS in production environments
- Implement rate limiting and proper CORS headers as needed

### Token Management
- Store `VIKUNJA_TOKEN` securely (use secrets management in production)
- Rotate tokens regularly
- Use environment-specific tokens when possible

## Troubleshooting

### Port Already in Use
```
Error: HTTP address localhost:8080 is already in use
```
Solution: Change the port using `MCP_HTTP_PORT` environment variable.

### Invalid Configuration
```
Error: VIKUNJA_HOST is required
```
Solution: Ensure all required environment variables are set.

### Connection Timeout
```
Error: HTTP server failed: listen tcp :8080: bind: address already in use
```
Solution: Use a different port or stop the conflicting service.

## Migration Guide

### From Stdio to HTTP
1. Set `MCP_TRANSPORT=http`
2. Configure HTTP settings if needed
3. Update client code to use HTTP endpoints
4. Handle connection management and session persistence

The migration is backward compatible - stdio transport remains the default if no transport is specified.