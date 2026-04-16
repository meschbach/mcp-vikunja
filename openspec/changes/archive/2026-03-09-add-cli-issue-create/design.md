## Context

The mcp-vikunja project already includes a `vikunja-cli` command with support for listing and retrieving tasks, projects, and buckets. The CLI is built using Cobra, with a root command that provides shared flags (host, token, output format, etc.) and a vikunja.Client for API interactions. Creating tasks is currently only possible through the web UI or direct API calls.

**Goals**: Add a new CLI command `vikunja-cli tasks create <title> [description]` to enable quick, scriptable task creation.

**Constraints**: 
- Follow existing CLI patterns (root command structure, formatting options, error handling)
- Use the existing `Client.CreateTask` method from `pkg/vikunja`
- Maintain consistency with output formatting (table, JSON, markdown)
- Support flexible project and bucket assignment using ID or title with automatic resolution
- Output format should match `vikunja-cli tasks get`

**Stakeholders**: CLI users wanting to automate task creation, developers using Vikunja for workflow automation.

## Goals / Non-Goals

**Goals:**
- Implement `tasks create` subcommand following existing CLI conventions
- Support optional `--project` flag accepting either numeric ID or project title (defaults to "Inbox" if omitted)
- Support optional `--bucket` flag accepting either numeric ID or bucket title
- Accept title as required positional argument and description as optional second positional argument
- Return the created task with full details including bucket information (like `tasks get`)
- Include proper validation and error messages
- Add tests for the new command

**Non-Goals:**
- Changing the root command structure or existing flags
- Modifying the underlying vikunja.Client (it already supports CreateTask)
- Supporting complex task features (labels, assignees, due date) in v1
- Adding interactive mode or prompts
- Changing output formatting infrastructure

## Decisions

**Decision 1: Command Structure**
- Extend existing `tasks` command group with `create` subcommand
- Command format: `vikunja-cli tasks create [FLAGS] <title> [description]`
- Reason: Aligns with Vikunja's terminology (tasks, not issues). Follows existing pattern: `tasks list`, `tasks get`.
- Alternatives: Could create separate `issues` command group but would create inconsistency. Integrating into `tasks` is more coherent.

**Decision 2: Argument Parsing**
- Title: required positional argument
- Description: optional positional argument (second arg) – no separate `--description` flag
- Project: optional flag `--project` (value can be numeric ID or project title). If omitted, defaults to "Inbox".
- Bucket: optional flag `--bucket` (value can be numeric ID or bucket title)
- Reason: Simplicity – description as a single optional argument keeps command concise. Using `--project` and `--bucket` (not `-id` suffix) allows flexible identification by name or ID, consistent with other MCP tools. Defaulting to "Inbox" matches common Vikunja user expectations.
- Resolution: Use same logic as MCP `create_task` handler: try numeric ID first, then treat as title; for buckets, require project's Kanban view to resolve title.

**Decision 3: Resolution & Validation**
- Project: Optional. If `--project` flag is omitted, use "Inbox" as the value to resolve. Resolve via `resolveProject()` (numeric ID first, then title search). Error if not found or ambiguous (multiple matches).
- Bucket: If provided, resolve via `findBucketByIDOrTitle()` (numeric ID first, then title search within project's Kanban view). Error if not found or ambiguous, or if project lacks Kanban view.
- Validate required arguments before API call. Follow existing error patterns: `fmt.Errorf("failed to create task: %w", err)`.
- On success, output created task with bucket information (if bucket was set) using the same formatting as `tasks get`.

**Decision 4: Output**
- Return the full created task with bucket context (like `tasks get` does via `FormatTaskWithBuckets`)
- Use `vikunja.NewFormatter` and call appropriate format method based on flags (table, JSON, markdown)
- On error: print to stderr, exit non-zero

## Decisions

### Decision 1: Command Structure
- Extend existing `tasks` command group with `create` subcommand
- Command format: `vikunja-cli tasks create [FLAGS] <title> [description]`
- Reason: Aligns with Vikunja's terminology (tasks, not issues). Follows existing pattern: `tasks list`, `tasks get`.
- Alternatives: Could create separate `issues` command group but would create inconsistency. Integrating into `tasks` is more coherent.

### Decision 2: Argument Parsing
- Title: required positional argument
- Description: optional positional argument (second arg) – no separate `--description` flag
- Project: optional flag `--project` (value can be numeric ID or project title). If omitted, defaults to "Inbox".
- Bucket: optional flag `--bucket` (value can be numeric ID or bucket title)
- Reason: Simplicity – description as a single optional argument keeps command concise. Using `--project` and `--bucket` (not `-id` suffix) allows flexible identification by name or ID, consistent with other MCP tools. Defaulting to "Inbox" matches common Vikunja user expectations.
- Resolution: Use same logic as MCP `create_task` handler: try numeric ID first, then treat as title; for buckets, require project's Kanban view to resolve title.

### Decision 3: Resolution & Validation
- Project: Optional. If `--project` flag is omitted, use "Inbox" as the value to resolve. Resolve via `resolveProject()` (numeric ID first, then title search). Error if not found or ambiguous (multiple matches).
- Bucket: If provided, resolve via `findBucketByIDOrTitle()` (numeric ID first, then title search within project's Kanban view). Error if not found or ambiguous, or if project lacks Kanban view.
- Validate required arguments before API call. Follow existing error patterns: `fmt.Errorf("failed to create task: %w", err)`.
- On success, output created task with bucket information (if bucket was set) using the same formatting as `tasks get`.

### Decision 4: Output
- Return the full created task with bucket context (like `tasks get` does via `FormatTaskWithBuckets`)
- Use `vikunja.NewFormatter` and call appropriate format method based on flags (table, JSON, markdown)
- On error: print to stderr, exit non-zero

### Decision 5: Shared Resolution Package
- Create `pkg/resolution` with exported functions `ResolveProject`, `FindBucketByIDOrTitle`, and `FindKanbanView`
- Both CLI and MCP handlers will import and use these shared functions to avoid duplication
- This refactor is explicitly scoped in tasks.md

### Decision 6: Configurable Default Project
- Not in v1. The default will be hard-coded to "Inbox"
- Future work could add config file support to override this

### Decision 7: Bucket Info Fetching
- Always fetch bucket information after task creation using `GetTaskBuckets`, regardless of whether a bucket was explicitly assigned
- Ensures output exactly matches `tasks get` and maintains consistency

### Decision 8: Validate Existence Before Creating
- No separate validation is needed – the resolution functions already check existence
- The subsequent `CreateTask` call would fail anyway if the project/bucket does not exist, but resolution provides user-friendly errors earlier

### Decision 9: Output Consistency
- Output will exactly match `tasks get` by using the same formatting functions (`FormatTaskWithBuckets` for table, `FormatTaskAsJSON` with bucket context for JSON, `FormatTaskAsMarkdown`)
- Always fetching bucket info ensures this consistency

## Risks / Trade-offs

**Risk**: Users may expect a `--description` flag in addition to positional description.
- Mitigation: Document that description is the second positional argument; quoting is supported. Simplicity outweighs flexibility for v1.

**Risk**: Defaulting to "Inbox" assumes every user has a project with that title.
- Mitigation: If "Inbox" resolution fails, provide clear error message suggesting to use `--project` to specify an existing project. Future: allow config file to override default project name.

**Trade-off**: Not supporting labels/assignees in v1 keeps implementation simple but limits functionality.
- Rationale: Focus on core use case (create task with title/description/project). Can extend later.

**Risk**: Bucket resolution by title requires a Kanban view; some projects may not have one.
- Mitigation: Clear error message explaining that a Kanban view is required to use bucket title. Users can still use numeric bucket ID.

## Migration Plan

This is a pure addition (no breaking changes):
1. Create `pkg/resolution` package with exported resolution functions and comprehensive unit tests
2. Update MCP handlers (`internal/handlers/create_task.go`, etc.) to use `pkg/resolution` instead of local utils in `internal/handlers/utils.go`
3. Verify MCP tests still pass: `go test ./internal/handlers/...`
4. Add `cmd/vikunja-cli/cmd/tasks_create.go` implementing `tasks create` subcommand, using shared `pkg/resolution` for project/bucket resolution
5. Register subcommand in `cmd/vikunja-cli/cmd/tasks.go` via `tasksCmd.AddCommand(tasksCreateCmd)`
6. Ensure `Client.CreateTask` works with resolved IDs and always fetch bucket info via `GetTaskBuckets` for output consistency
7. Update documentation (README) with `vikunja-cli tasks create` usage examples
8. Add unit tests in `cmd/vikunja-cli/cmd/tasks_create_test.go` covering flags, resolution integration, error cases, and formatting
9. Run full test suite: `go test ./...` to verify no regressions
10. Build and verify: `go build -o bin/vikunja-cli ./cmd/vikunja-cli`

**Rollback**: If issues arise, remove the `tasks_create.go` file and revert `pkg/resolution` changes. Data migration not needed.