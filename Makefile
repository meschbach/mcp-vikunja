.PHONY: build build-cli build-mcp build-all test test-cover clean lint lint-install fmt run download-spec dev dev-up dev-down dev-clean setup-user release check vet tidy deps

VIKUNJA_SPEC_URL := https://raw.githubusercontent.com/go-vikunja/vikunja/main/pkg/swagger/swagger.yaml

# Detect platform for release binaries
DETECTED_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ifeq ($(shell uname -m),x86_64)
DETECTED_ARCH := amd64
else
DETECTED_ARCH := arm64
endif
RELEASE_DIR := release/$(DETECTED_ARCH)_$(DETECTED_OS)

# Build CLI tool (from release binaries)
build-cli: release
	@mkdir -p bin
	cp $(RELEASE_DIR)/vikunja-cli bin/

# Build MCP server (from release binaries)
build-mcp: release
	@mkdir -p bin
	cp $(RELEASE_DIR)/mcp-vikunja bin/

# Build both binaries
build-all: build-cli build-mcp

# Download OpenAPI spec from go-vikunja (for reference)
download-spec:
	mkdir -p api
	curl -L -o api/vikunja.yaml $(VIKUNJA_SPEC_URL)

# Run the application
run:
	go run ./cmd/mcp-vikunja

# Run tests (requires docker-compose: make dev-up && make setup-user)
# Tests run against the configured VIKUNJA_HOST from .env
test:
	@if [ ! -f .env ]; then \
		echo "No .env file found. Run 'make setup-user' first."; \
		exit 1; \
	fi
	@echo "Running tests against $$(grep ^VIKUNJA_HOST .env | cut -d= -f2)..."
	VIKUNJA_HOST=$$(grep ^VIKUNJA_HOST .env | cut -d= -f2) VIKUNJA_TOKEN=$$(grep ^VIKUNJA_TOKEN .env | cut -d= -f2) VIKUNJA_INSECURE=true go test -tags=integration ./...

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html release/

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run go vet
vet:
	go vet ./...

# Check if dependencies are tidy
tidy-check:
	go mod tidy
	@if [ -n "$$(git diff)" ]; then \
		echo "go mod tidy produced changes. Run 'go mod tidy' and commit."; \
		git diff; \
		exit 1; \
	fi

# Run tests with coverage (requires docker-compose: make dev-up && make setup-user)
test-cover:
	@if [ ! -f .env ]; then \
		echo "No .env file found. Run 'make setup-user' first."; \
		exit 1; \
	fi
	VIKUNJA_HOST=$$(grep ^VIKUNJA_HOST .env | cut -d= -f2) VIKUNJA_TOKEN=$$(grep ^VIKUNJA_TOKEN .env | cut -d= -f2) VIKUNJA_INSECURE=true go test -race -tags=integration -coverprofile=coverage.out -coverpkg=./... ./...
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $$coverage"; \
	if [ "$$(echo "$$coverage < 27" | bc -l)" -eq 1 ]; then \
		echo "Coverage below 27%"; exit 1; \
	fi

# Run all checks (fmt, lint, vet, tidy-check, test, test-cover, build)
check: fmt vet lint tidy-check test test-cover build
	@echo ""
	@echo "✅ All checks passed!"

# Download dependencies
deps:
	go mod download

# Tidy dependencies
tidy:
	go mod tidy

# Vendor dependencies
vendor:
	go mod vendor

# Development targets
build: release
	@echo "Binaries built in release/ directory"

release:
	@echo "Building release binaries..."
	./release.sh

dev-up: release
	@echo "Starting development services..."
	docker-compose up -d

dev-down:
	@echo "Stopping development services..."
	docker-compose down

dev-clean:
	@echo "Stopping services and removing volumes..."
	docker-compose down -v

setup-user:
	@echo "Setting up Vikunja user..."
	@if [ ! -f .env ]; then cp .env.example .env; fi
	./scripts/setup-vikunja.sh

dev: release dev-up setup-user test
	@echo ""
	@echo "Development environment ready!"
	@echo "Vikunja: http://localhost:3456"
	@echo "MCP Server: http://localhost:8080"
