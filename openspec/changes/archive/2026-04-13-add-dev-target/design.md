## Context

The MCP Vikunja server currently requires manual setup for local development. Developers must:
1. Run `./release.sh` to build binaries
2. Run `docker-compose up -d` to start services
3. Wait for Vikunja to be ready
4. Open browser at localhost:8000, create admin account
5. Generate API token in Vikunja settings
6. Configure environment variables manually
7. Run tests

This makes onboarding slow and integration testing inconsistent.

## Goals / Non-Goals

**Goals:**
- Single `make dev` command builds, starts services, creates user/token, runs tests
- Reuses existing release.sh for binary builds (CI/CD source of truth)
- Auto-detects if user already exists (idempotent)
- Works on fresh machine

**Non-Goals:**
- Not building MCP server inside Docker container (simpler to debug)
- Not adding Redis (in-memory rate limiting is sufficient for dev)
- Not modifying core MCP server behavior

## Decisions

1. **Use release.sh instead of inline build**
   - Reuses CI/CD pipeline's build script
   - Consistent between local and CI environments
   - Creates release/linux_amd64/mcp-vikunja for docker-compose

2. **Auto-provision user via Vikunja REST API**
   - POST /register creates user (if signups enabled)
   - POST /token generates API token
   - Requires VIKUNJA_SIGNUP_ENABLED=true in compose

3. **Idempotent setup**
   - Script checks if user exists before creating
   - Re-uses existing token if available
   - Safe to re-run

4. **Environment via .env file**
   - Makefile creates .env from template
   - Includes all required variables
   - Easy to customize

## Risks / Trade-offs

- [Risk] Vikunja version changesbreaking /register API → [Mitigation] Pin to vikunja/api:latest and test
- [Risk] Database migrations on fresh start → [Mitigation] Health check waits for ready
- [Risk] Port conflicts (8000, 5432, 8080) → [Mitigation] Document in help output