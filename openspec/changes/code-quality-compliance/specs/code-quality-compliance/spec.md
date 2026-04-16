## ADDED Requirements

### Requirement: Zero golangci-lint violations
The codebase SHALL have zero violations when analyzed with golangci-lint using the project's `.golangci.yml` configuration.

#### Scenario: Running golangci-lint produces no errors
- **WHEN** `golangci-lint run` is executed
- **THEN** the command exits with code 0 and produces no output indicating violations

### Requirement: All tests pass
The system SHALL have all unit and integration tests passing.

#### Scenario: Running all tests succeeds
- **WHEN** `go test ./...` is executed
- **THEN** all tests pass and the command exits with code 0

### Requirement: Minimum test coverage
The production code (excluding `cmd/` and `main` packages) SHALL achieve at least 80% test coverage.

#### Scenario: Coverage meets threshold
- **WHEN** `go test -cover ./...` is executed
- **THEN** the coverage percentage for all packages meets or exceeds 80%

### Requirement: Code formatting compliance
All Go source files SHALL be formatted according to Go standards (gofmt) with proper import organization (goimports).

#### Scenario: Code passes formatting checks
- **WHEN** `go fmt ./...` and `goimports -w .` are applied
- **THEN** no formatting changes are required and the code is properly formatted

### Requirement: No static analysis warnings
The codebase SHALL produce no warnings from `go vet`.

#### Scenario: Static analysis passes
- **WHEN** `go vet ./...` is executed
- **THEN** no warnings are produced and the command exits with code 0

### Requirement: No global variables
The codebase SHALL contain no global variables that could introduce mutable global state.

#### Scenario: Code contains only local state
- **WHEN** the codebase is scanned for global variables
- **THEN** no package-level mutable variables are found

### Requirement: File and function size limits
All source files SHALL not exceed 400 lines (absolute maximum) with a preferred maximum of 200 lines. All functions SHALL not exceed 100 lines (absolute maximum) with a preferred maximum of 50 lines and cyclomatic complexity not exceeding 10.

#### Scenario: Code structure meets size limits
- **WHEN** source files and functions are analyzed
- **THEN** all files are ≤400 lines (≤200 preferred) and all functions are ≤100 lines (≤50 preferred) with complexity ≤10

### Requirement: Dependency management hygiene
The Go module dependencies SHALL be clean and verified.

#### Scenario: go mod tidy produces no changes
- **WHEN** `go mod tidy` is executed
- **THEN** no changes are made to go.mod or go.sum

#### Scenario: go mod verify succeeds
- **WHEN** `go mod verify` is executed
- **THEN** all module checksums are verified successfully and no errors occur

### Requirement: Structured logging compliance
The codebase SHALL use structured logging via `log/slog` and not use `fmt.Printf` for logging.

#### Scenario: No forbidden logging patterns
- **WHEN** the codebase is scanned for `fmt.Print*` calls
- **THEN** no logging statements using fmt are found (except in tests or debug code)
