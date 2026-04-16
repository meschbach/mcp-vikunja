## 1. Preparation & Dependency Management

- [ ] 1.1 Run `go mod tidy` until no changes
- [ ] 1.2 Run `go mod verify` to ensure checksums match
- [ ] 1.3 Review go.mod for any unused dependencies
- [ ] 1.4 Commit changes to go.mod and go.sum

## 2. Code Formatting

- [ ] 2.1 Run `go fmt ./...` to apply standard formatting
- [ ] 2.2 Run `goimports -w .` to organize imports
- [ ] 2.3 Verify no formatting changes remain
- [ ] 2.4 Commit formatting changes

## 3. Static Analysis

- [ ] 3.1 Run `go vet ./...` and capture all warnings
- [ ] 3.2 Fix each `go vet` warning
- [ ] 3.3 Re-run `go vet` to confirm all warnings resolved
- [ ] 3.4 Commit static analysis fixes

## 4. Linting

- [ ] 4.1 Run `golangci-lint run` and capture all violations
- [ ] 4.2 Categorize violations by type and severity
- [ ] 4.3 Fix bug-risk errors first (errcheck, gosec, staticcheck)
- [ ] 4.4 Address cyclomatic complexity violations (gocyclo) by refactoring complex functions
- [ ] 4.5 Fix file/function size violations by splitting large files and functions
- [ ] 4.6 Eliminate any global variables found
- [ ] 4.7 Replace `fmt.Print*` logging with `log/slog` structured logging
- [ ] 4.8 Fix style violations (gocritic, misspell, etc.)
- [ ] 4.9 Address test-related lints (paralleltest, tenv, testifylint)
- [ ] 4.10 Fix remaining violations (ineffassign, unused, gosimple, etc.)
- [ ] 4.11 Re-run `golangci-lint run` after each fix or set of fixes
- [ ] 4.12 Commit linting fixes

## 5. Testing

- [ ] 5.1 Run `go test ./...` and capture any failing tests
- [ ] 5.2 Fix all failing unit tests
- [ ] 5.3 Run `go test -cover ./...` to measure coverage
- [ ] 5.4 Identify packages below 80% coverage
- [ ] 5.5 Add unit tests for critical uncovered paths in low-coverage packages
- [ ] 5.6 Re-run `go test ./...` to ensure all tests pass
- [ ] 5.7 Re-run coverage check to verify 80% threshold met
- [ ] 5.8 Commit test fixes and additions

## 6. Validation

- [ ] 6.1 Re-run all quality checks in sequence: go vet, golangci-lint, go test, coverage
- [ ] 6.2 Verify all checks pass with zero violations
- [ ] 6.3 Verify coverage meets 80% minimum
- [ ] 6.4 Verify no global variables exist
- [ ] 6.5 Verify file/function size limits respected
- [ ] 6.6 Document any exceptions or justified deviations

## 7. Final Verification

- [ ] 7.1 Perform final comprehensive check (all commands one last time)
- [ ] 7.2 Ensure commit message clearly describes quality improvements
- [ ] 7.3 Push changes to branch for review

## 8. Documentation

- [ ] 8.1 Update AGENTS.md if coding patterns changed (optional)
- [ ] 8.2 Document any linter exceptions with justification in code comments
- [ ] 8.3 Update README with quality standards if needed
