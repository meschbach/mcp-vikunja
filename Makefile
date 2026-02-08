.PHONY: build build-cli build-mcp build-all test clean lint fmt run download-spec

VIKUNJA_SPEC_URL := https://raw.githubusercontent.com/go-vikunja/vikunja/main/pkg/swagger/swagger.yaml

# Build the application
build: build-cli build-mcp

# Build CLI tool
build-cli:
	go build -o bin/vikunja-cli ./cmd/vikunja-cli

# Build MCP server
build-mcp:
	go build -o bin/mcp-vikunja ./cmd/mcp-vikunja

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
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

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
