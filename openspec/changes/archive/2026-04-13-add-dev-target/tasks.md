## 1. Build Release Binaries

- [x] 1.1 Update release.sh if needed (verify it builds mcp-vikunja correctly)
- [x] 1.2 Test `./release.sh` locally to ensure binaries are created in release/linux_amd64/

## 2. Docker Compose Configuration

- [x] 2.1 Add build pre-step to Makefile (run release.sh before docker-compose up)
- [x] 2.2 Update docker-compose.yml build context to use release/ directory
- [x] 2.3 Fix mcp-vikunja-server service to build before running

## 3. Create Init Script

- [x] 3.1 Create scripts/setup-vikunja.sh
- [x] 3.2 Implement user registration via POST /register
- [x] 3.3 Implement token generation via POST /login (Vikunja v2 uses /login for JWT)
- [x] 3.4 Add idempotent detection (check if user exists first)
- [x] 3.5 Save token to .env file

## 4. Update Makefile

- [x] 4.1 Add `make build` target (calls release.sh)
- [x] 4.2 Add `make dev-up` target (build + docker-compose up)
- [x] 4.3 Add `make setup-user` target
- [x] 4.4 Add `make dev` target (orchestrates all: build + up + setup + test)
- [x] 4.5 Add `make dev-down` target (docker-compose down)
- [x] 4.6 Add `make dev-clean` target (down + remove volumes)

## 5. Configuration

- [x] 5.1 Update .env.example with all required variables
- [x] 5.2 Document environment variables in comments

## 6. Testing

- [x] 6.1 Test `make dev` from clean state
- [x] 6.2 Test `make dev-down`
- [x] 6.3 Test idempotency (run make dev again)
- [x] 6.4 Verify MCP server can connect to Vikunja

> **Note:** Testing completed successfully. All docker-compose services start, Vikunja runs on port 3456, setup script creates user and generates JWT token.