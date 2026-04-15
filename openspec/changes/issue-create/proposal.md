## Why

The MCP Vikunja server currently provides read-only access to Vikunja tasks, projects, and buckets, but lacks the
ability to create new tasks. Task creation is a fundamental operation for task management, essential for complete
workflow integration with LLMs and AI assistants. Adding the `create_task` tool will make the MCP server fully functional for task management operations.

## What Changes

- Add new `create_task` MCP tool that creates tasks in Vikunja projects
- Implement handler in `internal/handlers/create_task.go`
- Update tool registration in the MCP server to expose the new tool
- Add corresponding tests for the create functionality
- Update documentation to reflect the new capability

## Capabilities

### New Capabilities
- `task-creation`: Ability to create new tasks in Vikunja projects with title, description, bucket assignment, and other metadata

### Modified Capabilities
*(none)*

## Impact

- **Code**: New handler and test files in `internal/handlers/`, updates to MCP server registration
- **API**: New MCP tool `create_task` added to the available tools list
- **Documentation**: README and tool list updated
- **Dependencies**: Uses existing Vikunja client from `pkg/vikunja/`
- **Tests**: New unit tests for the create_task handler
