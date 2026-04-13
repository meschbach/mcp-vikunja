.PHONY: build build-cli build-mcp build-all test clean lint fmt run download-spec dev dev-up dev-down dev-clean setup-user release

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

# Run all tests
test:
	go test -v -count 1 --timeout 1s ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html release/

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

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

dev: release dev-up setup-user
	@echo ""
	@echo "Development environment ready!"
	@echo "Vikunja: http://localhost:3456"
	@echo "MCP Server: http://localhost:8080"
