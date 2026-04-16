## Why

Developing and testing the MCP Vikunja server requires multiple manual steps: building binaries via release.sh, starting docker-compose, waiting for services, creating a Vikunja user via the web UI, generating an API token, and manually configuring environment variables. This creates friction for local development and makes integration testing cumbersome.

## What Changes

- Add `make dev` target to Makefile that orchestrates: build → start services → setup user/token → run tests
- Add `make dev-down` target to tear down services
- Add `make setup-user` target that auto-creates Vikunja user and generates API token via Vikunja REST API
- Update docker-compose.yml to ensure release binaries exist before building container
- Add an init script (scripts/setup-vikunja.sh) that handles user registration and token generation

## Capabilities

### New Capabilities
- `dev-workflow`: Single command `make dev` to build, start, test against local Vikunja

### Modified Capabilities
None - this is a workflow improvement without spec-level behavior changes.

## Impact

- Makefile: New targets
- docker-compose.yml: Build order fix
- scripts/setup-vikunja.sh: New init script
- .env.example: Update with new variables