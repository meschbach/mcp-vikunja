## Why

The MCP Vikunja server currently lacks consistent due date handling across its tools and interactions with the Vikunja API. This leads to potential bugs, timezone confusion, and poor user experience when creating, updating, or querying tasks with due dates. Standardizing due date formats, validation, and timezone handling will improve reliability and usability.

## What Changes

- **BREAKING**: Standardize all due date input/output to ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)
- Add validation for due date values to prevent invalid dates
- Implement consistent timezone handling (convert all to UTC before API calls)
- Update all MCP tools that accept or return due dates to use the new standard
- Add utility functions for date parsing and formatting

## Capabilities

### New Capabilities
- `due-date-validation`: Validates due date strings and ensures they are proper ISO 8601 dates
- `timezone-normalization`: Converts user-provided dates to UTC and normalizes Vikunja responses to consistent format
- `date-format-standardization`: Provides single source of truth for date formatting across all MCP tools

### Modified Capabilities
- (none)

## Impact

- Affects all task creation/update tools (`create_task`, `update_task`, etc.)
- Impacts task querying and filtering operations
- Changes date handling in project views and dashboard tools
- Requires updates to tests and documentation
- May affect integration with external date-dependent systems
