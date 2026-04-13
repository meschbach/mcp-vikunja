## Context

The MCP Vikunja server currently handles due dates inconsistently across different tools and API interactions. Some tools accept date strings in various formats, others use time.Time without clear serialization, and timezone handling is implicit and error-prone. This creates confusion for users and potential bugs when integrating with Vikunja's API, which expects specific date formats.

## Goals / Non-Goals

**Goals:**
- Establish a single, consistent date format (ISO 8601) for all MCP tool inputs and outputs
- Centralize date validation and parsing logic to prevent invalid dates
- Normalize timezone handling by converting all dates to UTC before API calls
- Provide backward compatibility where possible, but enforce standards for new code
- Update all affected MCP tools to use the new date handling utilities

**Non-Goals:**
- Replacing Vikunja's API behavior - we adapt to its expectations
- Supporting arbitrary date format inputs from users (must use ISO 8601)
- Changing the internal Vikunja API client's date serialization beyond consistency fixes
- Implementing timezone conversion for past dates (only normalize to UTC)

## Decisions

### 1. Use ISO 8601 with UTC for all date representations

**Decision**: All dates will be represented as ISO 8601 strings (YYYY-MM-DDTHH:MM:SSZ) in UTC timezone. No other formats will be accepted.

**Alternatives**:
- Accept multiple formats and auto-convert → Rejected: increases complexity, edge cases
- Use timestamps (Unix time) → Rejected: less human-readable, Vikunja expects ISO strings
- Keep local time with timezone offset → Rejected: inconsistent, error-prone

**Rationale**: ISO 8601 is the standard for APIs, unambiguous, and Vikunja's API expects it. UTC eliminates timezone confusion.

### 2. Centralize date handling in a dedicated package

**Decision**: Create `internal/dates/` package with:
- `Parse(input string) (time.Time, error)` - parses ISO 8601, returns UTC time
- `Format(t time.Time) string` - formats to ISO 8601 UTC string
- `Validate(input string) error` - checks if valid ISO 8601 date
- `ToUTC(t time.Time) time.Time` - ensures time is in UTC

**Alternatives**:
- Scatter date logic across tools → Rejected: inconsistency guaranteed
- Extend existing client package → Rejected: mixing concerns
- Use third-party library → Rejected: stdlib sufficient

**Rationale**: Single source of truth, easy to update, testable, reusable.

### 3. Validation happens at tool boundaries

**Decision**: MCP tools will validate date inputs using the new utilities before processing. Invalid dates return clear error messages.

**Alternatives**:
- Lazy validation (only when calling Vikunja API) → Rejected: late failure, harder to debug
- Validation only on output → Rejected: doesn't prevent bad inputs

**Rationale**: Fail fast with clear feedback to user.

### 4. No automatic timezone conversion for ambiguous inputs

**Decision**: Users must provide dates in ISO 8601 format. If no timezone specified, assume UTC. Do not attempt to guess user's local timezone.

**Alternatives**:
- Assume local timezone → Rejected: implicit behavior, error-prone across distributed teams
- Prompt for timezone → Rejected: too interactive for MCP tools

**Rationale**: Explicit beats implicit; UTC is unambiguous.

## Risks / Trade-offs

**Risk**: Existing code may have implicit date handling that will break when force-converted to UTC.
- **Mitigation**: Comprehensive test coverage, gradual rollout if deployed to production, audit all date usages.

**Risk**: Users accustomed to custom date formats may find the change restrictive.
- **Mitigation**: Clear documentation, error messages that show expected format, migration guide.

**Risk**: Timezone handling edge cases (e.g., daylight saving, historical timezone data).
- **Mitigation**: Rely on Go's time package which handles these correctly; only use UTC conversions.

**Trade-off**: Breaking change for existing integrations that send non-ISO dates. This is intentional for long-term consistency.

**Trade-off**: Slight performance overhead from additional parsing/formatting calls. Negligible in practice (date ops are not hot path).
