## ADDED Requirements

### Requirement: Normalize timezone to UTC
The system SHALL convert all provided due dates to UTC timezone before sending to Vikunja API.

#### Scenario: Date with timezone offset converts to UTC
- **WHEN** user provides due date "2025-12-25T14:30:00+05:30"
- **THEN** the system SHALL convert it to "2025-12-25T09:00:00Z" before API call

#### Scenario: Date already in UTC remains unchanged
- **WHEN** user provides due date "2025-12-25T14:30:00Z"
- **THEN** the system SHALL use it as-is without modification

#### Scenario: Date without timezone defaults to UTC
- **WHEN** user provides due date "2025-12-25T14:30:00"
- **THEN** the system SHALL treat it as UTC "2025-12-25T14:30:00Z"

#### Scenario: Response from Vikunja API returns in UTC
- **WHEN** system receives task from Vikunja with due date in UTC
- **THEN** the system SHALL preserve the UTC format in response to user
