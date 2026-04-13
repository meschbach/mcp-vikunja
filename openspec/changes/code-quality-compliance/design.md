## Context

The mcp-vikunja project currently has golangci-lint violations and test failures that prevent the codebase from meeting quality standards. The project uses Go with the MCP (Model Context Protocol) server implementation for Vikunja integration. The codebase follows specific guidelines in AGENTS.md covering linting, testing, formatting, and code quality standards.

Current issues include:
- golangci-lint violations (exceeds project thresholds)
- Potential failing unit tests
- Code formatting inconsistencies (gofmt/goimports)
- Static analysis warnings from `go vet`

The goal is to establish a baseline where all quality checks pass consistently.

## Goals / Non-Goals

**Goals:**
- Fix all golangci-lint violations to meet the project's `.golangci.yml` configuration
- Ensure 100% of unit tests pass (`go test ./...`)
- Apply consistent code formatting (gofmt, goimports)
- Address all `go vet` warnings
- Achieve minimum 80% test coverage for production code (excluding `cmd/` and `main`)
- Ensure code adheres to file size limits (200 lines preferred, 400 max) and function size limits (50 lines preferred, 100 max)
- Verify no global variables exist
- Ensure all dependencies are properly managed and `go mod tidy` produces no changes

**Non-Goals:**
- Redesign architecture or refactor for performance (separate effort)
- Change existing linting rules or thresholds (use project defaults)
- Add new functionality or features
- Modify the MCP protocol implementation
- Change project structure or guidelines in AGENTS.md

## Decisions

**Decision 1: Sequential Fix Approach**
We'll address quality issues in a systematic order:
1. Run `go mod tidy` and `go mod verify` to ensure dependencies are clean
2. Apply `gofmt` and `goimports` to fix formatting
3. Run `go vet` and fix any warnings
4. Run `golangci-lint run` and categorize violations
5. Fix linting issues one by one, prioritizing:
   - Errors that could cause bugs
   - Code maintainability issues (cyclomatic complexity, function length)
   - Style violations
6. Run `go test ./...` and fix any failing tests
7. Re-run lint to ensure fixes didn't introduce new issues
8. Verify coverage meets 80% threshold

**Decision 2: Linter Configuration**
Use the existing `.golangci.yml` configuration from the project root. Do not modify linting rules; instead, fix code to comply. The configuration includes:
- errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports
- gosec (security), gocritic, gocyclo (complexity), dupl (duplication), misspell
- forbidigo (bans fmt.Print* in favor of structured logging)
- paralleltest, tenv, testifylint for testing
- exhaustivestruct, makezero, prealloc for code quality

**Decision 3: Test Strategy**
- Run all tests with `go test ./...` and ensure 100% pass rate
- Fix failing unit tests first, then integration tests (if any with `//go:build integration`)
- Add missing tests for critical paths where coverage is low
- Use `testify/assert` and `testify/require` as per existing patterns
- Maintain table-driven test structure
- Run benchmarks for performance-critical code if affected

**Decision 4: Code Formatting**
- Use `go fmt ./...` for basic formatting
- Use `goimports -w .` to fix import organization (standard library, third-party, internal)
- Ensure 4-space indentation (Go tabs)
- One blank line between functions, no blank line before closing brace
- Respect 120 character line limit (soft)

**Decision 5: File and Function Size**
- Refactor files exceeding 400 lines (absolute maximum) or approaching 200 lines (preferred)
- Split large functions (>100 lines absolute, >50 preferred) into smaller, focused functions
- Reduce cyclomatic complexity >10 using helper functions or early returns

**Decision 6: Zero Global Variables**
- Search for and eliminate any global variables
- Convert to dependency injection or package-level initialization where appropriate
- Ensure testability by avoiding global state

**Decision 7: Dependency Management**
- Run `go mod tidy` until no changes
- Run `go mod verify` to ensure checksums match
- Vendor dependencies only if required (use `go mod vendor`)
- Review `go.mod` for unused dependencies

## Risks / Trade-offs

**Risk:** Fixing lint violations might introduce subtle behavior changes.
**Mitigation:** Run comprehensive test suite after each fix; manually review logic changes, especially around error handling and nil checks.

**Risk:** Some lint rules (e.g., `gosec`) may flag code that is actually safe but requires careful review.
**Mitigation:** Use `//nolinte` comments sparingly and only with justification; document exceptions.

**Risk:** Enforcing 80% coverage may require significant test writing for legacy code.
**Mitigation:** Prioritize critical paths; use integration tests where unit tests are difficult; accept temporarily lower coverage if justified, but aim for full compliance.

**Risk:** Large refactoring to reduce file/function size could introduce bugs.
**Mitigation:** Make incremental improvements; test thoroughly after each refactor; keep changes focused on quality, not behavior.

**Risk:** Breaking CI/CD pipeline if configuration changes are needed.
**Mitigation:** Verify all commands locally before committing; update CI configuration only if absolutely necessary and document changes.

## Migration Plan

1. **Preparation**: Clean up dependencies (`go mod tidy`, `go mod verify`)
2. **Formatting**: Apply code formatting (`go fmt`, `goimports`)
3. **Static Analysis**: Fix `go vet` warnings
4. **Linting**: Iteratively fix golangci-lint violations; categorize and prioritize
5. **Testing**: Ensure all tests pass; add missing tests to reach coverage target
6. **Validation**: Re-run all quality checks in sequence to confirm compliance
7. **Final Verification**: Run `golangci-lint run`, `go test -cover ./...`, `go vet`, `go fmt` one last time
8. **Documentation**: Update any necessary documentation if patterns change

**Rollback Strategy**: Each fix is a separate git commit; if a change introduces bugs, revert that specific commit. Maintain a branch for the quality work to isolate changes.

## Open Questions

- Should we enforce the 80% coverage threshold immediately, or allow temporary exceptions for legacy code that is stable but untested?
- Are there specific golangci-lint rules that conflict with the project's coding patterns and require exceptions? If so, what is the justification process?
- Should we automatically format all code with `goimports` even if it changes import ordering slightly from current manual organization?
- What is the threshold for "adequate coverage" - is 80% strict or a guideline?
- Are there any integration tests that require external services (Vikunja instance) that should be skipped or mocked during this quality sweep?
- Should we run `golangci-lint run --fix` where possible to automate fixes, or manually address each violation for better control?
