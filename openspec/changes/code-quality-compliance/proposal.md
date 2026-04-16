## Why

The mcp-vikunja codebase must consistently meet golangci-lint standards and have all tests passing to ensure code quality, maintainability, and reliability. This is critical for production readiness, developer confidence, and maintaining the project's health. Currently, the codebase has linting issues and test failures that need to be addressed.

## What Changes

- Fix all golangci-lint violations across the codebase
- Ensure all unit tests pass (including fixing any failing tests)
- Verify all integration tests pass where applicable
- Update CI/CD configuration to enforce linting and test requirements
- Add missing tests to achieve adequate coverage
- Apply code formatting standards (gofmt, goimports)
- Address any static analysis warnings from `go vet`

## Capabilities

### New Capabilities

- `code-quality-compliance`: Establishes and maintains code quality standards including linting, testing, and static analysis for the entire mcp-vikunja project.

### Modified Capabilities

*(none)*

## Impact

- All source files in the repository will be reviewed and fixed to meet linting standards
- Test suites will be updated to ensure comprehensive coverage and passing status
- CI/CD pipeline (if exists) may need updates to enforce quality gates
- Developer workflows will be updated to include linting and testing checks
- Code formatting will be standardized across the codebase
