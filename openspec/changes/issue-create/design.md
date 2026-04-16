## Context

The MCP Vikunja server currently provides read-only access to Vikunja tasks, projects, and buckets. The existing codebase has:
- A Vikunja client in `pkg/vikunja/` with methods for fetching data
- Handlers in `internal/handlers/` for MCP tools like `list_tasks`, `get_task`, `list_projects`, `list_buckets`
- A handler pattern using the MCP Go SDK with tool registration
- Support for both stdio and HTTP transports

The missing piece is task creation, which is listed in the README as "coming soon."

## Goals / Non-Goals

**Goals:**
- Add a `create_task` MCP tool that allows creating new tasks in Vikunja projects
- Support specifying task title, description, project assignment (by name or ID), optional bucket assignment (by name or ID with Kanban view default), and due dates
- Follow existing handler patterns in `internal/handlers/`
- Maintain consistency with existing code style and conventions
- Include comprehensive unit tests

**Non-Goals:**
- Modifying existing task-related tools or their behavior
- Changing the Vikunja client API
- Implementing task update or delete operations (future work)
- Changing output formatting or transport mechanisms

## Decisions

### 1. Follow existing handler pattern

**Decision**: Implement `create_task` following the same pattern as existing handlers (e.g., `get_task.go`, `move_task.go`).

**Rationale**: 
- Maintains code consistency and reduces cognitive overhead
- Existing handlers are well-structured and testable
- Uses established patterns for parameter parsing, error handling, and response formatting

**Alternatives considered**:
- Creating a separate service layer: rejected as over-engineering for a single operation
- Integrating with a different client method: not applicable since client already supports task creation via API

### 2. Generic field naming

**Decision**: Use `project` and `bucket` field names in CreateTaskInput (not `project_id`/`project_name` or `bucket_id`/`bucket_name`).

**Rationale**:
- Matches pattern in `ListTasksInput` which uses `Project`, `View`, `Bucket` fields
- Reduces model confusion; each field accepts either ID (numeric string) or name (string)
- Clearer API surface for MCP clients

**Alternatives considered**:
- Separate `project_id` and `project_title` fields: rejected as overly verbose; validation logic is same

### 3. Use existing Vikunja client method

**Decision**: Call `client.CreateTask()` from `pkg/vikunja/client_tasks.go` from the handler after resolving names to IDs.

**Rationale**:
- The client already has methods for other task operations; `CreateTask` follows the same interface
- Keeps HTTP/API communication encapsulated in the client package
- Allows reusability outside MCP context (e.g., CLI tool)
- Resolution logic stays in handler, keeping client simple and focused

### 4. Input validation and resolution in handler

**Decision**: The handler will accept either project ID or project name (and similarly for bucket). It will:
1. Validate required fields are present (title non-empty)
2. Resolve project name to ID using `client.GetProjects()` if needed
3. If bucket is specified, resolve bucket name to ID using `client.GetProjectViews()` and `client.GetViewBuckets()` with the "Kanban" view as default
4. Call `client.CreateTask()` with resolved IDs
5. Immediately return validation errors before any API calls

**Resolution Rules**:
- **Project resolution**: When `project` is a numeric string, attempt to resolve as ID first before trying title match. If both a numeric ID and a title match exist, prefer the ID. If multiple projects share the same title, return an error (duplicate).
- **Bucket resolution**: When `bucket` is a numeric string, attempt to resolve as bucket ID first before trying title match. If both a numeric ID and a title match exist, prefer the ID.
- **Bucket view**: Always use the "Kanban" view (exact case-sensitive match). If the "Kanban" view does not exist, return an error.
- **Case sensitivity**: All name resolution (project, bucket, view) is case-sensitive.
- **Independent resolution**: Project and bucket are resolved independently; any combination is allowed (project ID + bucket name, project name + bucket ID, etc.)

**Rationale**:
- Provides flexible, user-friendly interface for MCP clients
- Leverages existing list operations for name resolution
- Validation errors can clearly indicate if a provided name doesn't match any accessible resource
- Centralizes resolution logic in handler rather than spreading across client layers
- Default to "Kanban" view for bucket resolution as it's the most common and other views may not have buckets
- Accumulate as much validation as possible before returning errors
- ID-first resolution prevents ambiguity with numeric titles

**Alternatives considered**:
- Adding resolution methods to the Vikunja client: rejected to keep client simple and focused on API operations
- Accepting only IDs: rejected as requiring extra lookup work by the MCP client before calling create_task
- Case-insensitive matching: rejected to maintain consistency with existing handler patterns

## Risks / Trade-offs

- **Risk**: Vikunja API for task creation may require additional parameters not yet identified.
  - **Mitigation**: Review Vikunja API documentation and existing client code; add optional parameters as needed.

- **Risk**: Inconsistent behavior compared to other handlers if output formatting is not handled properly.
  - **Mitigation**: Follow the pattern in `get_task.go` and `move_task.go` for formatting success responses and errors.

- **Trade-off**: Simpler implementation may lack advanced features (labels, assignees, etc.) in first version.
  - This is acceptable as a minimum viable product; can be extended later.

## Error Handling Patterns

- Use existing `enhancedProjectNotFoundError` pattern from `handlers/utils.go` for project name resolution failures
- Create similar `enhancedBucketNotFoundError` for bucket resolution failures, suggesting available bucket names
- Duplicate project names: error with message "multiple projects found with title {title}, please use project ID"
- Missing Kanban view: error with message "project {project} does not have a 'Kanban' view, which is required for bucket resolution"
- All errors should include actionable suggestions (e.g., "use list_projects() to see all projects")
