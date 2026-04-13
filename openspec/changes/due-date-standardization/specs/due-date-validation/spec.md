## ADDED Requirements

### Requirement: Validate ISO 8601 date strings
The system SHALL validate that provided due date strings conform to ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ).

#### Scenario: Valid ISO 8601 date with UTC timezone
- **WHEN** user provides "2025-12-25T14:30:00Z"
- **THEN** the system SHALL accept the date as valid

#### Scenario: Valid ISO 8601 date with timezone offset
- **WHEN** user provides "2025-12-25T14:30:00+05:30"
- **THEN** the system SHALL accept the date and convert it to UTC

#### Scenario: Invalid date format
- **WHEN** user provides "12/25/2025"
- **THEN** the system SHALL reject with an error stating "due date must be in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)"

#### Scenario: Invalid date value (February 30)
- **WHEN** user provides "2025-02-30T12:00:00Z"
- **THEN** the system SHALL reject with an error stating "invalid due date"
