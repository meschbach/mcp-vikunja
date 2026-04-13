## 1. CLI Command Implementation

- [x] 1.1 Create `cmd/vikunja-cli/cmd/tasks_create.go` with `tasks create` subcommand
- [x] 1.2 Implement subcommand with optional flag `--project` (accepts ID or title, defaults to "Inbox") and optional flag `--bucket` (accepts ID or title)
- [x] 1.3 Add argument parsing: positional title (required), optional positional description (no separate --description flag)
- [x] 1.4 In command RunE, determine project value: if flag omitted, use "Inbox"; if provided, use the given value
- [x] 1.5 Resolve project using shared `resolution.ResolveProject()` (numeric ID first, then title). Resolve bucket if provided using shared `resolution.FindBucketByIDOrTitle()`.
- [x] 1.6 Call `client.CreateTask()` with resolved numeric IDs (handle nil bucketID and zero dueDate)
- [x] 1.7 Always fetch task with expanded bucket information using `GetTaskBuckets()` after creation to match `tasks get` output consistency
- [x] 1.8 Format and output the created task with bucket context using existing formatters, respecting root flags (JSON/markdown/table)

## 2. Shared Resolution Package (Refactor)

- [x] 2.1 Create new package `pkg/resolution` with exported functions:
  - `ResolveProject(ctx context.Context, client *vikunja.Client, identifier string) (*resolution.Project, error)` – returns resolved project
  - `FindBucketByIDOrTitle(ctx context.Context, client *vikunja.Client, projectID int64, bucketInput string) (*int64, error)` – returns bucket ID
  - `FindKanbanView(ctx context.Context, client *vikunja.Client, projectID int64) (*vikunja.ProjectView, error)`
- [x] 2.2 Define `resolution.Project` type (minimal: ID, Title) to avoid circular dependencies
- [x] 2.3 Move or copy helper functions from `internal/handlers/utils.go`: `findProjectByIDOrTitle()`, `findViewByName()`, error helpers (`enhancedProjectNotFoundError`, `enhancedBucketNotFoundError`)
- [x] 2.4 Write unit tests for `pkg/resolution` covering:
  - project resolution by ID, by title (exact match, no match, multiple matches)
  - bucket resolution by ID, by title (including Kanban view requirement)
  - default "Inbox" handling
- [x] 2.5 Update MCP handler `internal/handlers/create_task.go` to use `pkg/resolution` instead of local utils
- [x] 2.6 Verify MCP tests still pass: `go test ./internal/handlers/...`
- [x] 2.7 Ensure resolution package provides consistent error messages (e.g., "project not found", "multiple projects match", "kanban view not found") matching MCP patterns; document error cases in package comments

## 3. Registration and Integration

- [x] 3.1 Register `tasksCreateCmd` in `cmd/vikunja-cli/cmd/tasks.go` via `tasksCmd.AddCommand(tasksCreateCmd)`
- [x] 3.2 Verify that the new command inherits root-level flags (host, token, json, markdown, output, verbose, no-color, insecure)

## 5. Testing

- [x] 5.1 Add unit tests in `cmd/vikunja-cli/cmd/tasks_create_test.go` for command initialization and flag parsing
- [x] 5.2 Test resolution integration with shared `pkg/resolution` – project by ID, by title, default "Inbox"
- [x] 5.3 Test bucket resolution with `pkg/resolution` – by ID, by title, Kanban view errors
- [x] 5.4 Test default project behavior: when no --project flag, verify command uses "Inbox" and calls resolution
- [x] 5.5 Test that bucket info is always fetched after creation (`GetTaskBuckets` called)
- [x] 5.6 Test creation success case with mocked client (verify CreateTask called with correct resolved IDs)
- [x] 5.7 Test error cases: missing title, resolution failures (not found, ambiguous, Inbox not found), API errors
- [x] 5.8 Test output formatting: table, JSON, markdown (using test helper from other commands)
- [x] 5.9 Run full test suite: `go test ./...` to ensure no regressions

## 6. Documentation and Verification

- [x] 4.1 Update `README.md` (or CLI docs) with `vikunja-cli tasks create` usage examples
- [x] 4.2 Run `go fmt ./...` and `goimports -w .` to format code
- [x] 4.3 Build CLI: `go build -o bin/vikunja-cli ./cmd/vikunja-cli` to verify compilation
- [x] 4.4 Optional: manual test with a dev Vikunja instance to confirm end-to-end flow