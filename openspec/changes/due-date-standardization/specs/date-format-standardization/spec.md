## ADDED Requirements

### Requirement: Provide consistent date formatting utilities
The system SHALL provide centralized functions for formatting and parsing dates to ensure consistency across all MCP tools.

#### Scenario: Format function produces ISO 8601 UTC
- **WHEN** any tool formats a time.Time value
- **THEN** the system SHALL return string in "2006-01-02T15:04:05Z" format (Go's reference time layout)

#### Scenario: Parse function accepts ISO 8601 strings
- **WHEN** any tool parses a date string
- **THEN** the system SHALL accept any valid ISO 8601 variant (with or without timezone, with different separators)

#### Scenario: Format function ensures UTC
- **WHEN** tool passes local time to Format function
- **THEN** the system SHALL convert to UTC before formatting

#### Scenario: All task-related tools use the same formatting
- **WHEN** different tools (create_task, update_task, list_tasks) display due dates
- **THEN** all SHALL use identical format and timezone representation
