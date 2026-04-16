## 1. Package Setup

- [ ] 1.1 Create internal/dates/ package directory
- [ ] 1.2 Create dates.go with package declaration and initial structure

## 2. Core Date Utilities Implementation

- [ ] 2.1 Implement Parse function to handle ISO 8601 strings
- [ ] 2.2 Implement Format function to output ISO 8601 UTC strings
- [ ] 2.3 Implement Validate function for date string validation
- [ ] 2.4 Implement ToUTC helper for timezone conversion
- [ ] 2.5 Write unit tests for all date utility functions

## 3. Update Task-Related MCP Tools

- [ ] 3.1 Audit all task creation/update tools for date handling
- [ ] 3.2 Refactor create_task tool to use new date utilities
- [ ] 3.3 Refactor update_task tool to use new date utilities
- [ ] 3.4 Refactor get_task/list_tasks tools to format dates consistently
- [ ] 3.5 Update task querying/filtering if date-based

## 4. Update Project View and Dashboard Tools

- [ ] 4.1 Review all tools that display task due dates
- [ ] 4.2 Update formatting calls to use centralized Format function
- [ ] 4.3 Ensure all date inputs in these tools use Parse/Validate

## 5. API Client Integration

- [ ] 5.1 Verify Vikunja client expects ISO 8601 UTC dates
- [ ] 5.2 Add date normalization before API calls if needed
- [ ] 5.3 Ensure API responses are consistently interpreted as UTC

## 6. Testing and Validation

- [ ] 6.1 Add integration tests for task operations with dates
- [ ] 6.2 Test timezone edge cases (offset conversions, DST)
- [ ] 6.3 Test error handling for invalid date formats
- [ ] 6.4 Run full test suite to verify no regressions

## 7. Documentation and Cleanup

- [ ] 7.1 Update README or internal docs about date format standards
- [ ] 7.2 Add godoc comments to date utility functions
- [ ] 7.3 Run golangci-lint and fix any issues
- [ ] 7.4 Update any existing tests that assumed old date formats
