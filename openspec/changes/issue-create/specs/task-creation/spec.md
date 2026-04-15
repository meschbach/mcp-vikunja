## ADDED Requirements

### Requirement: Create task with minimal required fields
The system SHALL allow creation of a new task with only a title and project identifier (either project ID or project name).

#### Scenario: Successful task creation with title only
- **WHEN** an MCP client calls the create_task tool with a valid project identifier (ID or name) and a non-empty title
- **THEN** the system SHALL create a new task in the specified project
- **AND** the system SHALL return a confirmation with the new task's ID, title, and project ID

### Requirement: Support optional task description
The system SHALL support an optional description field for tasks.

#### Scenario: Task created with title and description
- **WHEN** an MCP client provides both a title and a description
- **THEN** the created task SHALL include the provided description
- **AND** the response SHALL include the description in the task details

### Requirement: Support optional bucket assignment
The system SHALL allow assigning a task to a specific bucket within a project, using either bucket ID or bucket name. Bucket resolution SHALL use the "Kanban" view by default.

#### Scenario: Task assigned to a bucket by ID
- **WHEN** an MCP client specifies a valid bucket ID
- **THEN** the created task SHALL be placed in that bucket
- **AND** the response SHALL include the bucket ID in the task details

#### Scenario: Task assigned to a bucket by name
- **WHEN** an MCP client specifies a valid bucket name (within the "Kanban" view of the project)
- **THEN** the system SHALL resolve the bucket name to a bucket ID
- **AND** the created task SHALL be placed in that bucket
- **AND** the response SHALL include the bucket ID in the task details

#### Scenario: Bucket name resolution with ID priority
- **WHEN** the bucket field contains a numeric string that could be both a valid bucket ID AND a matching bucket title
- **THEN** the system SHALL treat it as a bucket ID first
- **AND** if no bucket with that ID exists, fall back to matching by title

#### Scenario: Bucket name not found
- **WHEN** the specified bucket name does not exist in the project's "Kanban" view
- **THEN** the system SHALL return an error with a list of available bucket names
- **AND** the task SHALL NOT be created

#### Scenario: Kanban view missing
- **WHEN** a bucket name is provided but the project does not have a view named "Kanban"
- **THEN** the system SHALL return an error indicating that the "Kanban" view is required for bucket resolution
- **AND** the task SHALL NOT be created

### Requirement: Validate required fields
The system SHALL validate required input parameters before attempting to create a task. Validation SHALL occur before any API calls to Vikunja.

#### Scenario: Missing title
- **WHEN** the create_task tool is called without a title or with an empty title
- **THEN** the system SHALL return an error indicating that title is required
- **AND** the task SHALL NOT be created in Vikunja

#### Scenario: Invalid title (if any constraints apply)
- **WHEN** the provided title violates any constraints (e.g., too long, invalid characters)
- **THEN** the system SHALL return an error describing the constraint violation
- **AND** the task SHALL NOT be created

#### Scenario: Missing project identifier
- **WHEN** the create_task tool is called without a project identifier (neither project ID nor project name provided)
- **THEN** the system SHALL return an error indicating that project identifier is required
- **AND** the task SHALL NOT be created in Vikunja

#### Scenario: Duplicate project names
- **WHEN** multiple projects with the same title exist
- **THEN** the system SHALL return an error indicating the project name is ambiguous
- **AND** the task SHALL NOT be created

### Requirement: Handle Vikunja API errors
The system SHALL handle errors from the Vikunja API gracefully.

#### Scenario: Invalid project ID
- **WHEN** the provided project ID does not exist or the user lacks access
- **THEN** the system SHALL return an error indicating the project is invalid or inaccessible
- **AND** the task SHALL NOT be created
- **AND** the error SHALL be presented in a clear, user-friendly format

#### Scenario: Vikunja service unavailable
- **WHEN** the Vikunja API returns a 5xx error or the client cannot connect
- **THEN** the system SHALL return an appropriate error indicating service unavailability
- **AND** the task SHALL NOT be created
