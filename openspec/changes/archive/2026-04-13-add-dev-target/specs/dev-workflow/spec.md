## ADDED Requirements

### Requirement: make dev orchestrates full local development workflow
The `make dev` command SHALL build binaries, start docker-compose services, setup Vikunja user/token, and run tests in sequence without manual intervention.

#### Scenario: Fresh start
- **WHEN** developer runs `make dev` on a machine with no existing setup
- **THEN** binaries are built, containers start, user is created, token is generated, and tests pass

#### Scenario: Idempotent re-run
- **WHEN** developer runs `make dev` again on an already configured machine
- **THEN** existing user and token are reused, services restart, tests run

### Requirement: make dev-up builds and starts services
The `make dev-up` command SHALL build binaries via release.sh and start all docker-compose services.

#### Scenario: Successful start
- **WHEN** developer runs `make dev-up`
- **THEN** release/ directory contains built binaries and all containers are running

### Requirement: make setup-user creates/updates Vikunja user and token
The `make setup-user` command SHALL create a Vikunja user if needed and generate an API token, saving it to .env file.

#### Scenario: First-time setup
- **WHEN** developer runs `make setup-user` with no existing user
- **THEN** new user is registered via Vikunja API, token is generated, variables saved to .env

#### Scenario: User already exists
- **WHEN** developer runs `make setup-user` when user already exists
- **THEN** existing token is retrieved or new one generated, .env is updated

### Requirement: make dev-down stops services
The `make dev-down` command SHALL stop all docker-compose services.

#### Scenario: Clean shutdown
- **WHEN** developer runs `make dev-down`
- **THEN** all containers are stopped

### Requirement: make dev-clean removes all data
The `make dev-clean` command SHALL remove all containers and volumes for a completely fresh start.

#### Scenario: Full reset
- **WHEN** developer runs `make dev-clean`
- **THEN** all containers, volumes, and state are destroyed

### Requirement: Health checks confirm Vikunja is ready
The setup process SHALL wait for Vikunja to be healthy before attempting user registration.

#### Scenario: Service not ready
- **WHEN** setup runs before Vikunja API is ready
- **THEN** script retries with exponential backoff until health check passes