## Why

Currently there's no command-line interface for quickly creating Vikunja tasks. Users need to use the web UI or API directly. Adding a CLI tool allows for scriptable task creation, integration with development workflows, and faster task reporting from terminal-based environments.

## What Changes

- Add new CLI subcommand: `vikunja-cli tasks create <title> [description]`
- Reuse existing MCP `create_task` tool (already present)
- Add CLI argument parsing and validation for `--project` and `--bucket` flags with ID/title resolution
- Integrate CLI with existing Vikunja client to call task creation
- Output created task with full details (including bucket info) consistent with `tasks get`

## Capabilities

### New Capabilities

- `tasks create`: CLI command to create Vikunja tasks with title and optional description, supporting project/bucket assignment via ID or title

### Modified Capabilities

*(none)*

## Impact

- New CLI subcommand under existing `vikunja-cli`
- Reuse of existing MCP `create_task` handler
- Client library already supports CreateTask
- Test coverage for CLI parsing and resolution logic
- Documentation updates for usage examples