# Docker Compose Configuration for MCP Vikunja Local Testing

This docker-compose setup provides a complete local development environment for the MCP Vikunja server with:

- 🚀 **Vikunja API Server** - Official Vikunja container with PostgreSQL database
- 🔧 **MCP Vikunja Server** - The MCP server connected to local Vikunja
- 🛠️ **Development Mode** - Hot-reloading stdio server for development
- 🧪 **Integration Testing** - Dedicated test runner with proper configuration

## Quick Start

### 1. Start All Services

```bash
# Start all services in detached mode
docker-compose up -d

# Or start with logs visible
docker-compose up
```

### 2. Verify Services Are Running

```bash
# Check all services
docker-compose ps

# View logs
docker-compose logs -f
```

### 3. Access Services

- 📡 **Vikunja Web Interface**: http://localhost:3456
- 🔗 **MCP HTTP Server**: http://localhost:8080
- 🔧 **MCP Test Service**: http://localhost:8081

### 4. Create Initial Admin User

Open http://localhost:8000 in your browser and create an admin account to access the Vikunja web interface.

## Development Workflow

### Development Mode (Hot Reloading)

The `mcp-vikunja-dev` service provides a development environment with:

- Source code mounted from host
- Stdio transport for local CLI tools
- Real-time file watching
- Markdown output format for AI/LLM compatibility

```bash
# Start only the development service
docker-compose up mcp-vikunja-dev

# Or rebuild and start
docker-compose up --build mcp-vikunja-dev
```

### HTTP Transport Development

For testing HTTP transport functionality:

```bash
# Start HTTP server
docker-compose up mcp-vikunja-server

# Test with curl
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }'
```

## Integration Testing

### Running Tests

The `mcp-vikunja-test` service is configured for integration testing:

```bash
# Run integration tests
docker-compose up mcp-vikunja-test

# Or run tests manually inside the container
docker-compose exec mcp-vikunja-test bash
```

### Test Configuration

- Uses HTTP transport with stateless sessions
- Runs on port 8081 to avoid conflicts
- Includes insecure mode for testing
- Mounts source code for test execution

## Environment Variables

All services use these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VIKUNJA_HOST` | `http://vikunja:3456` | Vikunja API endpoint |
| `VIKUNJA_TOKEN` | `dev-token-for-local-testing` | API authentication token |
| `VIKUNJA_INSECURE` | `false` | Skip TLS verification |
| `MCP_TRANSPORT` | `http` | Transport mode (stdio/http) |
| `MCP_HTTP_HOST` | `0.0.0.0` | HTTP server bind address |
| `MCP_HTTP_PORT` | `8080` | HTTP server port |
| `MCP_HTTP_SESSION_TIMEOUT` | `1h` | Session timeout duration |
| `MCP_HTTP_STATELESS` | `false` | Disable session tracking |
| `VIKUNJA_OUTPUT_FORMAT` | `markdown` | Output format (json/markdown/both) |

## Custom Configuration

### Override Environment Variables

Create a `.env` file in the project root:

```env
# Custom Vikunja configuration
VIKUNJA_HOST=http://vikunja:3456
VIKUNJA_TOKEN=your-custom-token
VIKUNJA_INSECURE=true

# Custom MCP server configuration
MCP_TRANSPORT=http
MCP_HTTP_HOST=0.0.0.0
MCP_HTTP_PORT=9000
MCP_HTTP_SESSION_TIMEOUT=2h
MCP_HTTP_STATELESS=false
```

### Custom Services

Add custom services to the `docker-compose.yml` file:

```yaml
# Example custom service
custom-tool:
  image: custom-image
  environment:
    - VIKUNJA_HOST=http://vikunja:8000
    - CUSTOM_VAR=value
  networks:
    - mcp-vikunja-network
```

## Cleanup

### Stop All Services

```bash
docker-compose down
```

### Remove All Data

```bash
docker-compose down -v
```

### Rebuild Services

```bash
docker-compose build --no-cache
```

## Troubleshooting

### Common Issues

1. **Vikunja Health Check Fails**
   - Wait for PostgreSQL to initialize completely
   - Check database logs: `docker-compose logs vikunja-db`

2. **MCP Server Fails to Start**
   - Verify Vikunja is healthy first
   - Check MCP server logs: `docker-compose logs mcp-vikunja-server`

3. **Port Conflicts**
   - Stop conflicting services: `docker-compose down`
   - Change port mappings in `docker-compose.yml`

### Debug Commands

```bash
# Access running container
docker-compose exec vikunja bash

# View service status
docker-compose ps

# View detailed service information
docker-compose config

# Restart specific service
docker-compose restart mcp-vikunja-server
```

## Integration with Existing Tests

This setup works seamlessly with the existing test suite:

- **Unit Tests**: Run directly with `go test ./...`
- **Integration Tests**: Use the test service configuration
- **HTTP Tests**: Test against `http://localhost:8080`
- **Stdio Tests**: Use the development service

### Example Test Commands

```bash
# Run all tests
go test ./...

# Run integration tests with Docker
go test -tags=integration ./test/integration/...

# Test HTTP transport
go test -run TestHTTPTransport ./...

# Test stdio transport
go test -run TestStdioTransport ./...
```

## Production Considerations

For production deployment:

- Use persistent volumes for data storage
- Configure proper TLS certificates
- Set up reverse proxy (nginx/traefik)
- Implement proper authentication
- Configure health checks and monitoring
- Use environment-specific configuration

This development setup provides a solid foundation for local testing while being easily adaptable for production deployment.