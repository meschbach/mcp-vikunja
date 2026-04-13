# CLI Task Create Specification

## Purpose

This capability defines the requirements for creating tasks via the Vikunja CLI. It covers command syntax, argument handling, project/bucket resolution, output formatting, error handling, and authentication.

**Status**: TBD - Implementation pending

## Requirements

### Requirement: CLI command accepts title and optional description
The `vikunja-cli tasks create` command SHALL accept a title as a required positional argument and an optional description as a second positional argument. No separate `--description` flag is provided.

#### Scenario: Create task with title only
- **WHEN** user runs `vikunja-cli tasks create "My Task Title"`
- **THEN** the command creates a new task with the given title and an empty description
- **AND** the command outputs the created task in the configured format (table, JSON, or markdown)

#### Scenario: Create task with title and description as positional arguments
- **WHEN** user runs `vikunja-cli tasks create "My Task Title" "Task description here"`
- **THEN** the command creates a new task with the given title and description
- **AND** the command outputs the created task in the configured format

### Requirement: Command supports project and bucket assignment with flexible identification
The command SHALL support an optional `--project` flag and an optional `--bucket` flag. Both flags SHALL accept either a numeric ID or a title string. If `--project` is omitted, the command SHALL default to "Inbox" and attempt to resolve the project with that title. If `--bucket` is provided, it SHALL be resolved within the context of the chosen project (default or explicit). The command SHALL perform resolution using the same logic as the MCP `create_task` handler: try parsing as numeric ID first, then treat as a title for lookup.

#### Scenario: Create task in default project (Inbox) without explicit --project
- **WHEN** user runs `vikunja-cli tasks create "My Task"`
- **THEN** the command attempts to resolve the project with title "Inbox"
- **AND** creates the task in that project
- **AND** outputs the created task

#### Scenario: Create task in project by ID
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123`
- **THEN** the command creates a new task in project with ID 123
- **AND** the command outputs the created task

#### Scenario: Create task in project by title
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project "Work Projects"`
- **THEN** the command resolves the project by title (exact match required, or error if multiple/no matches)
- **AND** creates the task in that project
- **AND** outputs the created task

#### Scenario: Create task in specific bucket by ID
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123 --bucket 456`
- **THEN** the command creates a new task in bucket 456 within project 123's Kanban view
- **AND** the task's bucket assignment is reflected in the output

#### Scenario: Create task in specific bucket by title
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123 --bucket "In Progress"`
- **THEN** the command resolves the bucket by title within the project's Kanban view
- **AND** creates the task in that bucket
- **AND** outputs the task with bucket information

#### Scenario: Create task in default project with bucket by title
- **WHEN** user runs `vikunja-cli tasks create "My Task" --bucket "In Progress"`
- **THEN** the command defaults project to "Inbox", resolves that project
- **AND** resolves the bucket "In Progress" within the Inbox project's Kanban view
- **AND** creates the task in that bucket
- **AND** outputs the task with bucket information

#### Scenario: Resolution failure – default project "Inbox" not found
- **WHEN** user runs `vikunja-cli tasks create "My Task"` and the server has no project with title "Inbox"
- **THEN** the command returns an error: "default project 'Inbox' not found" (or similar)
- **AND** exit code is non-zero

#### Scenario: Resolution failure – project not found or ambiguous
- **WHEN** user provides a non-existent project title or a title matching multiple projects
- **THEN** the command returns an error with details (e.g., "project not found" or "multiple projects match")
- **AND** exit code is non-zero

#### Scenario: Resolution failure – bucket not found or ambiguous
- **WHEN** user provides a bucket title that doesn't exist, matches multiple buckets, or the project has no Kanban view
- **THEN** the command returns an appropriate error (e.g., "bucket not found", "multiple buckets match", or "Kanban view required")
- **AND** exit code is non-zero

#### Scenario: Bucket resolution by title requires Kanban view
- **WHEN** user provides a project that lacks a Kanban view and also provides a `--bucket` value that is not a numeric ID
- **THEN** the command returns an error: "kanban view not found in project <id>"
- **AND** exit code is non-zero

### Requirement: Respects global output format flags
The command SHALL inherit and respect the root-level output format flags (`--json`, `--markdown`, `--output`) and output the created task accordingly. The output format SHALL be consistent with `vikunja-cli tasks get`.

#### Scenario: Output as table (default)
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123`
- **THEN** the command outputs the created task in table format, showing title, ID, project, description, and bucket information (if applicable)

#### Scenario: Output as JSON
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123 --json`
- **THEN** the command outputs the created task as a JSON object (including bucket context if bucket was specified)

#### Scenario: Output as Markdown
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123 --markdown`
- **THEN** the command outputs the created task in Markdown format

#### Scenario: Write output to file
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123 --output task.json`
- **THEN** the command writes the output to the specified file in the requested format

### Requirement: Expands bucket information to match tasks get
After creating a task, the command SHALL fetch expanded bucket information using `GetTaskBuckets` to present output that matches the richness of `vikunja-cli tasks get`, including bucket view information. This fetch SHALL occur regardless of whether a bucket was explicitly assigned during creation, to ensure consistent output format.

#### Scenario: Task creation includes bucket information in output
- **WHEN** user creates a task (with or without `--bucket`)
- **THEN** the command calls `GetTaskBuckets` after creation
- **AND** the output includes bucket name and view (Kanban) information when applicable, consistent with `tasks get`

### Requirement: Authentication and connection
The command SHALL use the host and token provided via root command flags or environment variables (`VIKUNJA_HOST`, `VIKUNJA_TOKEN`). If these are missing, the root command's PersistentPreRunE SHALL return an error before executing.

#### Scenario: Missing authentication
- **WHEN** user runs `vikunja-cli tasks create "My Task" --project 123` without setting host or token
- **THEN** the command exits with error: "host is required" or "token is required"

### Requirement: Error handling and logging
The command SHALL log debug information when `--verbose` is set. On API or resolution errors, the command SHALL wrap the error message (e.g., "failed to create task: <details>") and exit with non-zero status.

#### Scenario: API error propagation
- **WHEN** the Vikunja API returns an error (e.g., 400 Bad Request, 401 Unauthorized)
- **THEN** the command prints an error message to stderr
- **AND** the exit code is non-zero

#### Scenario: Resolution failure
- **WHEN** project resolution fails (no match, multiple matches) or bucket resolution fails
- **THEN** the command prints a clear error message
- **AND** the exit code is non-zero
