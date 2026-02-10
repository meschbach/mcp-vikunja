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

### File Size Limits
- **Preferred maximum**: 200 lines per file
- **Absolute maximum**: 400 lines per file  
- **Exception**: Generated code (protobuf, etc.)
- **Enforcement**: Split files when exceeding limits
- **Rationale**: Improves maintainability, reduces cognitive load, encourages focused responsibilities

### Function Size Limits
- **Preferred maximum**: 50 lines per function
- **Absolute maximum**: 100 lines per function
- **Cyclomatic complexity**: Maximum 10
- **Enforcement**: `gocyclo` linter and code review

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
- **Required**: Use structured logging via `log/slog` (Go 1.21+)
- Include relevant context in log messages
- Use appropriate log levels: Debug, Info, Warn, Error
- Never log sensitive information
- **Forbidden**: Do not use `fmt.Printf` for logging

```go
// Good - structured logging
logger.Warn("failed to get project views",
    slog.Int("task_id", taskID),
    slog.Any("error", err),
)

// Bad - forbidden pattern
fmt.Printf("Warning: failed to get project views for task %d: %v\n", taskID, err)
```

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
- **Benchmark tests**: `*_benchmark_test.go` for performance-critical code
- **Integration tests**: Use build tag `//go:build integration`

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
  timeout: 10m
  tests: true
  skip-dirs:
    - vendor

linters:
  disable-all: true
  enable:
    # Core
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    
    # Security
    - gosec
    
    # Code quality
    - gocritic
    - gocyclo
    - dupl
    - misspell
    - dogsled
    - makezero
    - prealloc
    
    # Testing
    - paralleltest
    - tenv
    - testifylint
    
    # Modern practices
    - exhaustivestruct
    - forbidigo

linters-settings:
  gocyclo:
    min-complexity: 10
  
  gosec:
    excludes:
      - G204  # Subprocess launching may be allowed in this context
  
  dupl:
    threshold: 100
  
  misspell:
    locale: US
  
  forbidigo:
    forbid:
      - 'fmt\.Print.*'  # Use structured logging instead
  
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
    disabled-checks:
      - dupImport  # https://github.com/go-critic/go-critic/issues/845
      - ifElseChain
      - octalLiteral
      - whyNoLint
      - wrapperFunc
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

## Performance Standards

### Memory Management
- Pre-allocate slices when size is known: `make([]Type, 0, capacity)`
- Use object pools for frequently allocated structs
- Avoid allocations in hot paths
- Use `strings.Builder` for string concatenation in loops

### Concurrency Patterns
```go
// Good: Proper context cancellation
func (c *Client) GetTasks(ctx context.Context) ([]Task, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    // Use context for request
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // ...
}

// Good: Goroutine lifecycle management
func processTasks(tasks []Task) <-chan Result {
    results := make(chan Result, len(tasks))
    var wg sync.WaitGroup
    
    for _, task := range tasks {
        wg.Add(1)
        go func(t Task) {
            defer wg.Done()
            // Process task
        }(task)  // Pass by value to avoid race conditions
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    return results
}
```

### Resource Management
- Set HTTP client timeouts: `&http.Client{Timeout: 30*time.Second}`
- Use `defer` for cleanup consistently
- Implement connection pooling for HTTP clients
- Always close idle connections: `defer client.CloseIdleConnections()`

## Modern Go Practices (Go 1.21+)

### Error Handling
```go
// Use errors.Is and errors.As for error inspection
if errors.Is(err, context.Canceled) {
    // Handle cancellation
}

var apiErr *APIError
if errors.As(err, &apiErr) {
    // Handle specific API error
}
```

### Structured Logging
```go
import "log/slog"

// Use structured logging
logger := slog.With("component", "vikunja-client", "operation", "get-tasks")

logger.Info("fetching tasks", 
    "project_id", projectID,
    "view_id", viewID,
)

logger.Error("failed to fetch tasks", 
    "error", err,
    "project_id", projectID,
)
```

### Generic Types (Go 1.18+)
```go
// Use generics where appropriate
type Result[T any] struct {
    Data  T
    Error error
}

func Fetch[T any](ctx context.Context, endpoint string) (Result[T], error) {
    // Generic implementation
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

### Enhanced Dependency Injection
```go
// Define interfaces at point of use
type TaskLister interface {
    ListTasks(ctx context.Context, projectID int64) ([]vikunja.Task, error)
}

type TaskPresenter interface {
    FormatTasks(tasks []vikunja.Task) (string, error)
}

// Handler depends on abstractions
type TaskHandler struct {
    lister    TaskLister
    presenter TaskPresenter
    logger    *slog.Logger
}

func NewTaskHandler(
    lister TaskLister,
    presenter TaskPresenter,
    logger *slog.Logger,
) *TaskHandler {
    return &TaskHandler{
        lister:    lister,
        presenter: presenter,
        logger:    logger,
    }
}
```

### File Organization
```
internal/
├── handlers/
│   ├── tasks.go          # Task-related MCP tools
│   ├── projects.go       # Project-related MCP tools
│   ├── views.go          # View-related MCP tools
│   ├── discovery.go      # Tool discovery handlers
│   ├── handlers.go       # Common handler utilities
│   └── handlers_test.go  # Shared test utilities
├── transport/
│   ├── server.go         # MCP server implementation
│   ├── middleware.go     # HTTP middleware
│   └── server_test.go
└── config/
    ├── config.go         # Configuration management
    └── config_test.go
```

## Testing Standards

### Test Organization
- Test files: `*_test.go` in same package
- Table-driven tests for multiple scenarios
- Subtests with descriptive names via `t.Run()`

### Test Coverage
- **Minimum 80% coverage** for production code (increased from 70%)
- **Exclude** `cmd/` and `main` packages from coverage requirements
- Mock external dependencies using interfaces
- Integration tests use build tag: `//go:build integration`
- **Benchmark tests**: Required for all performance-critical paths

### Benchmark Testing
```go
func BenchmarkClient_GetTasks(b *testing.B) {
    client := setupTestClient()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = client.GetTasks(ctx)
    }
}
```

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

# 7. Run benchmarks (performance-critical changes)
go test -bench=. ./...
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
2. `golangci-lint` passes (including new enhanced rules)
3. `go test ./...` passes
4. Coverage report meets minimum threshold (80%)
5. `go build ./...` succeeds
6. `go mod tidy` produces no changes
7. All files under 400 lines (automated check)
8. Zero global variables (automated check)
9. Benchmark tests pass for performance-critical paths
