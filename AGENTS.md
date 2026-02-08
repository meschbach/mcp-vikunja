# MCP Vikunja Agent Guidelines

## Project Overview
A Model Context Protocol (MCP) server for Vikunja task management integration.

## Build/Test Commands

```bash
# Build
go build -o bin/mcp-vikunja ./cmd/mcp-vikunja

# Run
go run ./cmd/mcp-vikunja

# Test all
go test ./...

# Test with verbose output
go test -v ./...

# Run single test
go test -run TestFunctionName ./path/to/package

# Test specific package
go test ./internal/vikunja

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Lint (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...
goimports -w .

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Vendor dependencies
go mod vendor

# Static analysis
go vet ./...
```

## Code Style Guidelines

### Imports
- Standard library first, separated by blank line
- Third-party imports next, separated by blank line  
- Internal/project imports last
- Use `goimports` for automatic formatting
- Group imports logically within sections

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/meschbach/mcp-vikunja/internal/vikunja"
)
```

### Formatting
- Use `gofmt` for consistent formatting
- Max line length: 120 characters (soft limit)
- 4-space indentation (tabs in Go)
- One blank line between functions
- No blank line before closing brace

### Naming Conventions
- **Packages**: Short, lowercase, no underscores (e.g., `vikunja`, `handlers`)
- **Exported identifiers**: PascalCase (e.g., `NewServer`, `TaskHandler`)
- **Unexported identifiers**: camelCase (e.g., `newServer`, `taskHandler`)
- **Interfaces**: End in `-er` when possible (e.g., `Reader`, `Writer`)
- **Error variables**: Start with `Err` (e.g., `ErrNotFound`)
- **Constants**: PascalCase for exported, camelCase for unexported

### Types
- Use strong typing; avoid `interface{}` unless necessary
- Define interfaces where you use them, not where you implement them
- Prefer concrete types for simple cases
- Use struct tags for JSON serialization: `json:"field_name,omitempty"`

### Error Handling
- Always check errors explicitly
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Return errors rather than logging and continuing
- Use sentinel errors for common cases (define as package vars)
- Don't panic in production code; return errors instead

```go
if err != nil {
    return fmt.Errorf("failed to fetch task: %w", err)
}
```

### Logging
- Use structured logging via `log/slog` (Go 1.21+)
- Include relevant context in log messages
- Use appropriate log levels: Debug, Info, Warn, Error
- Never log sensitive information

### MCP Patterns
- Each tool should have a dedicated handler function
- Use descriptive tool names with clear descriptions
- Validate all inputs before processing
- Return structured results as maps or structs

### Testing
- Use `testify/assert` and `testify/require`
- Table-driven tests for multiple test cases
- Test file naming: `*_test.go`
- Mock external dependencies using interfaces
- Keep tests in same package (use `package foo_test` for black-box)

```go
func TestHandler(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"valid", "input", "output"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := handler(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Project Structure
```
.
├── cmd/mcp-vikunja/     # Main application entry point
├── internal/            # Private application code
│   ├── vikunja/        # Vikunja API client
│   ├── handlers/       # MCP tool handlers
│   └── config/         # Configuration management
├── pkg/                 # Public library code (if any)
└── docs/               # Documentation
```

### Configuration
- Use environment variables for configuration
- Support `.env` files for local development
- Validate configuration at startup
- Fail fast on missing required config

### Documentation
- All exported types and functions must have doc comments
- Comments start with the name being documented
- Use complete sentences
- Include examples in doc comments where helpful

### Dependencies
- Minimize external dependencies
- Prefer standard library when possible
- Pin dependency versions in go.mod
- Run `go mod tidy` before committing
- Review all new dependencies for security

## Static Analysis & Linting

### Required Tools
- **gofmt** - Standard Go formatting
- **goimports** - Import formatting and grouping
- **golangci-lint** - Comprehensive linting with presets:
  - `errcheck` - Unchecked errors
  - `govet` - Go vet checks
  - `staticcheck` - Static analysis
  - `unused` - Unused code detection
- **go vet** - Built-in static analysis

### Linting Configuration
Create `.golangci.yml`:
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
```

## Code Quality Standards

### General Practices
- **No naked returns** - Always name return values or use explicit returns
- **No init() functions** - Avoid implicit initialization; use explicit constructors
- **No global state** - Pass dependencies explicitly for testability
- **Context propagation** - Always pass `context.Context` as first parameter
- **Interface segregation** - Define interfaces at point of use, not implementation

### Error Handling
- Check all errors explicitly
- Wrap with context: `fmt.Errorf("failed to X: %w", err)`
- Define sentinel errors as package variables
- Never panic in production code

### Struct Field Tags
Always use JSON tags for API types:
```go
type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description,omitempty"`
}
```

## Architecture Patterns

### Constructor Functions
Use `NewXxx()` pattern with functional options:
```go
// Client creation with options
client := vikunja.NewClient(token,
    vikunja.WithHTTPClient(customHTTP),
    vikunja.WithBaseURL(customURL),
)
```

### Dependency Injection
Pass dependencies explicitly via constructors:
```go
// Good - explicit dependencies
type Handler struct {
    client *vikunja.Client
}

func NewHandler(client *vikunja.Client) *Handler {
    return &Handler{client: client}
}

// Avoid - global state
var globalClient *vikunja.Client
```

### Interface Naming
- Use `-er` suffix when possible: `Reader`, `Writer`, `Handler`
- Keep interfaces small and focused
- Define interfaces where you use them (consumer-side)

## Testing Standards

### Test Organization
- Test files: `*_test.go` in same package
- Table-driven tests for multiple scenarios
- Subtests with descriptive names via `t.Run()`

### Test Coverage
- **Minimum 70% coverage** for production code
- **Exclude** `cmd/` and `main` packages from coverage requirements
- Mock external dependencies using interfaces
- Integration tests use build tag: `//go:build integration`

### Example Test Structure
```go
func TestClient_ListTasks(t *testing.T) {
    tests := []struct {
        name      string
        projectID int
        mockResp  []Task
        wantErr   bool
    }{
        {
            name:      "valid project",
            projectID: 1,
            mockResp:  []Task{{ID: 1, Title: "Task 1"}},
            wantErr:   false,
        },
        {
            name:      "empty project",
            projectID: 99,
            mockResp:  []Task{},
            wantErr:   false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := setupMockClient(tt.mockResp)
            tasks, err := client.ListTasks(context.Background(), tt.projectID)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.mockResp, tasks)
        })
    }
}
```

### Integration Tests
```go
//go:build integration

func TestClientIntegration_ListTasks(t *testing.T) {
    // Tests that make actual API calls
    // Run with: go test -tags=integration ./...
}
```

## Build & CI Standards

### Pre-commit Checklist
```bash
# 1. Format code
go fmt ./...
goimports -w .

# 2. Run linters
golangci-lint run
go vet ./...

# 3. Run tests
go test ./...

# 4. Check coverage
go test -cover ./...

# 5. Verify build
go build ./...

# 6. Tidy dependencies
go mod tidy
go mod verify
```

### Git Hooks
Consider using `pre-commit` or `lefthook` to automate checks:
```yaml
# .lefthook.yml
pre-commit:
  commands:
    gofmt:
      run: go fmt ./...
    golangci-lint:
      run: golangci-lint run
    test:
      run: go test ./...
```

### CI Pipeline
Required checks for pull requests:
1. `go fmt` verification (no formatting changes needed)
2. `golangci-lint` passes
3. `go test ./...` passes
4. Coverage report meets minimum threshold
5. `go build ./...` succeeds
6. `go mod tidy` produces no changes
