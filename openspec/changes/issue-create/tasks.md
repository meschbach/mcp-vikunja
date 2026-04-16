## 1. Vikunja Client Extension

- [x] 1.1 Verify CreateTask exists and works correctly in `pkg/vikunja/client_tasks.go`
- [x] 1.2 Verify unit tests for CreateTask in `pkg/vikunja/client_test.go` exist; if missing, add them

## 2. MCP Handler Implementation

- [x] 2.1 Already completed: `CreateTaskInput` accepts both ID/title for `ProjectID` and `BucketID`, `DueDate` removed
- [x] 2.2 Implement/update `internal/handlers/create_task.go` with resolution logic
- [x] 2.3 Input validation already implemented: title non-empty, ID parsing (already exists from other handlers)
- [x] 2.4 Implement project resolution with ID-first approach, duplicate name detection (part of 2.2)
- [x] 2.5 Implement bucket resolution via "Kanban" view (case-sensitive), with ID-first fallback (part of 2.2)
- [x] 2.6 Implement error for missing "Kanban" view when bucket is specified (part of 2.2)
- [x] 2.7 Implement comprehensive error handling: project not found, duplicate project, view not found, bucket not found, API errors (part of 2.2)
- [x] 2.8 Add missing validation: bucket_id must be validated before API calls (already partly covered by ID parsing)
- [x] 2.9 Format success response following existing pattern (already implemented in handler)

## 3. MCP Server Integration

- [x] 3.1 Register create_task tool in `internal/handlers/handlers.go`
- [x] 3.2 Verify tool appears in tools/list response

## 4. Testing

- [x] 4.1 Create/update unit tests for create_task handler in `internal/handlers/create_task_test.go`
- [x] 4.2 Add test cases covering:
  - [x] 4.2.1 Successful creation: project ID only, project name only, bucket ID, bucket name (with Kanban)
  - [x] 4.2.2 Mixed combinations: project ID + bucket name, project name + bucket ID
  - [x] 4.2.3 Missing title validation (fail before any API calls)
  - [x] 4.2.4 Project resolution: project name not found, duplicate project names (same title multiple projects)
  - [x] 4.2.5 Bucket resolution: bucket not found (with suggestions), bucket name with numeric fallback (ID-first)
  - [x] 4.2.6 Missing "Kanban" view when bucket is specified
  - [x] 4.2.7 API failures: GetProjects error, GetProjectViews error, GetViewBuckets error, CreateTask error
  - [x] 4.2.8 Case-sensitivity: bucket names "todo" vs "Todo" should be different
- [x] 4.3 Verify client CreateTask tests and add missing ones
- [x] 4.4 Remove `t.Skip()` from tests to enable them (no t.Skip() found in codebase)
- [x] 4.5 Run full test suite: `go test ./...` (per AGENTS.md line 79) - PASSED
- [x] 4.6 Check coverage: `go test -cover ./...` (per AGENTS.md line 91, target: 60%+) - PASSED (70.8%)
- [ ] 4.7 Ensure lint passes: `golangci-lint run` (per AGENTS.md line 95) - BLOCKED (lint issues exist)

## 5. Documentation

- [x] 5.1 Update README.md to list create_task as available tool (remove "coming soon")
- [x] 5.2 Add usage examples for create_task in README

## 6. Verification (INCOMPLETE)

- [x] 6.1 Build the server: `go build -o bin/mcp-vikunja ./cmd/mcp-vikunja` (per AGENTS.md line 73)
- [ ] 6.2 Run lint: `golangci-lint run` (per AGENTS.md line 95) - BLOCKED (lint issues exist)
- [x] 6.3 Run tests with coverage: `go test -cover ./...` (per AGENTS.md line 91) - PASSED (70.8%)
- [x] 6.4 Verify no formatting: `go fmt ./...` (per AGENTS.md line 98)

## 7. Issues Found During Exploration

- [x] 7.1 Tests have `t.Skip()` that must be removed to enable them (see `internal/handlers/create_task_test.go:18`) - No t.Skip() found in codebase
- [x] 7.2 Coverage for handlers package is 56.8% (target: 60%+) - now 61.8%
- [ ] 7.3 Multiple lint issues: dup, errcheck, gocognit, gocyclo, forbidigo need fixes - BLOCKED (see section 8)

## 8. Pre-commit Checklist (per AGENTS.md line 633-657)

- [x] 8.1 Format code: `go fmt ./...` then `goimports -w .`
- [x] 8.2 Run linters: `golangci-lint run` then `go vet ./...` (issues exist, need fixing)
- [x] 8.3 Run tests: `go test ./...` - PASSED
- [x] 8.4 Check coverage: `go test -cover ./...` (must meet 60% threshold) - PASSED (70.8%)
- [x] 8.5 Verify build: `go build ./...`
- [x] 8.6 Tidy dependencies: `go mod tidy` then `go mod verify`
