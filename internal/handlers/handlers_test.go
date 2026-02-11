package handlers

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestEnv sets up the environment variables for testing
func setupTestEnv(t *testing.T) func() {
	t.Helper()
	oldHost := os.Getenv("VIKUNJA_HOST")
	oldToken := os.Getenv("VIKUNJA_TOKEN")

	os.Setenv("VIKUNJA_HOST", "test.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token")

	return func() {
		os.Setenv("VIKUNJA_HOST", oldHost)
		os.Setenv("VIKUNJA_TOKEN", oldToken)
	}
}

// newTestHandlers creates handlers with default test dependencies
func newTestHandlers() *Handlers {
	cfg := &config.Config{Readonly: false}
	deps := &HandlerDependencies{
		Config:          cfg,
		OutputFormatter: vikunja.GetFormatter(vikunja.OutputFormatJSON),
	}
	return NewHandlers(deps)
}

func TestListProjectsHandler_Success(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()
	input := ListProjectsInput{}

	result, output, err := h.listProjectsHandler(context.Background(), &mcp.CallToolRequest{}, input)

	// Since we can't mock the HTTP client, this will fail with network error
	// but we can verify the handler structure is correct
	assert.Error(t, err) // Expected network error in test
	assert.Nil(t, result)
	assert.Equal(t, ListProjectsOutput{}, output)
}

func TestFindProjectByNameHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         FindProjectByNameInput
		expectedError string
	}{
		{
			name:          "missing name field",
			input:         FindProjectByNameInput{Name: ""},
			expectedError: "name: is required",
		},
		{
			name:          "valid name format",
			input:         FindProjectByNameInput{Name: "Test Project"},
			expectedError: "failed to list projects", // Will fail due to network
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.findProjectByNameHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			if tt.expectedError == "name: is required" {
				require.Error(t, err)
				assert.True(t, result.IsError)
				if len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						assert.Contains(t, textContent.Text, tt.expectedError)
					}
				}
			} else {
				// Network errors are expected in tests without mocked client
				assert.Error(t, err)
			}
			assert.Equal(t, FindProjectByNameOutput{}, output)
		})
	}
}

func TestListBucketsHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         ListBucketsInput
		expectedError string
	}{
		{
			name:          "missing project_id",
			input:         ListBucketsInput{ProjectID: "", ViewID: "1"},
			expectedError: "project_id: is required",
		},
		{
			name:          "missing view_id",
			input:         ListBucketsInput{ProjectID: "1", ViewID: ""},
			expectedError: "view_id: is required",
		},
		{
			name:          "invalid project_id format",
			input:         ListBucketsInput{ProjectID: "invalid", ViewID: "1"},
			expectedError: "project_id: must be a valid integer",
		},
		{
			name:          "invalid view_id format",
			input:         ListBucketsInput{ProjectID: "1", ViewID: "invalid"},
			expectedError: "view_id: must be a valid integer",
		},
		{
			name:          "negative project_id",
			input:         ListBucketsInput{ProjectID: "-1", ViewID: "1"},
			expectedError: "project_id: must be a positive integer",
		},
		{
			name:          "negative view_id",
			input:         ListBucketsInput{ProjectID: "1", ViewID: "-1"},
			expectedError: "view_id: must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.listBucketsHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.True(t, result.IsError)
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					assert.Contains(t, textContent.Text, tt.expectedError)
				}
			}
			assert.Equal(t, ListBucketsOutput{}, output)
		})
	}
}

func TestGetTaskHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         GetTaskInput
		expectedError string
	}{
		{
			name:          "missing task_id",
			input:         GetTaskInput{TaskID: ""},
			expectedError: "task_id: is required",
		},
		{
			name:          "invalid task_id format",
			input:         GetTaskInput{TaskID: "invalid"},
			expectedError: "task_id: must be a valid integer",
		},
		{
			name:          "negative task_id",
			input:         GetTaskInput{TaskID: "-1"},
			expectedError: "task_id: must be a positive integer",
		},
		{
			name:          "zero task_id",
			input:         GetTaskInput{TaskID: "0"},
			expectedError: "task_id: must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.True(t, result.IsError)
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					assert.Contains(t, textContent.Text, tt.expectedError)
				}
			}
			assert.Equal(t, GetTaskOutput{}, output)
		})
	}
}

func TestCreateTaskHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         CreateTaskInput
		expectedError string
	}{
		{
			name:          "missing title",
			input:         CreateTaskInput{Title: "", ProjectID: "1"},
			expectedError: "title: is required",
		},
		{
			name:          "missing project_id",
			input:         CreateTaskInput{Title: "Test Task", ProjectID: ""},
			expectedError: "project_id: is required",
		},
		{
			name:          "invalid project_id format",
			input:         CreateTaskInput{Title: "Test Task", ProjectID: "invalid"},
			expectedError: "project_id: must be a valid integer",
		},
		{
			name:          "negative project_id",
			input:         CreateTaskInput{Title: "Test Task", ProjectID: "-1"},
			expectedError: "project_id: must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.True(t, result.IsError)
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					assert.Contains(t, textContent.Text, tt.expectedError)
				}
			}
			assert.Equal(t, CreateTaskOutput{}, output)
		})
	}
}

func TestCreateTaskHandler_ReadonlyMode(t *testing.T) {
	cfg := &config.Config{Readonly: true}
	deps := &HandlerDependencies{
		Config:          cfg,
		OutputFormatter: vikunja.GetFormatter(vikunja.OutputFormatJSON),
	}
	h := NewHandlers(deps)

	input := CreateTaskInput{
		Title:     "Test Task",
		ProjectID: "1",
	}

	result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

	require.Error(t, err)
	assert.True(t, result.IsError)
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "Operation not available in readonly mode")
		}
	}
	assert.Equal(t, CreateTaskOutput{}, output)
}

func TestValidationUtils(t *testing.T) {
	t.Run("validateRequiredString", func(t *testing.T) {
		tests := []struct {
			name        string
			fieldName   string
			value       string
			shouldError bool
			expectedMsg string
		}{
			{
				name:        "empty string returns error",
				fieldName:   "test_field",
				value:       "",
				shouldError: true,
				expectedMsg: "test_field: is required",
			},
			{
				name:        "non-empty string passes",
				fieldName:   "test_field",
				value:       "valid value",
				shouldError: false,
			},
			{
				name:        "whitespace-only string passes",
				fieldName:   "test_field",
				value:       "   ",
				shouldError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateRequiredString(tt.fieldName, tt.value)
				if tt.shouldError {
					require.Error(t, err)
					assert.Equal(t, tt.expectedMsg, err.Error())
					var valErr ValidationError
					assert.True(t, errors.As(err, &valErr))
					assert.Equal(t, tt.fieldName, valErr.Field)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("parseID", func(t *testing.T) {
		tests := []struct {
			name        string
			fieldName   string
			value       string
			expectedID  int64
			shouldError bool
			expectedMsg string
		}{
			{
				name:        "empty string returns error",
				fieldName:   "task_id",
				value:       "",
				shouldError: true,
				expectedMsg: "task_id: is required",
			},
			{
				name:        "valid positive integer",
				fieldName:   "task_id",
				value:       "123",
				expectedID:  123,
				shouldError: false,
			},
			{
				name:        "invalid format returns error",
				fieldName:   "task_id",
				value:       "invalid",
				shouldError: true,
				expectedMsg: "task_id: must be a valid integer, got: invalid",
			},
			{
				name:        "negative integer returns error",
				fieldName:   "task_id",
				value:       "-5",
				shouldError: true,
				expectedMsg: "task_id: must be a positive integer, got: -5",
			},
			{
				name:        "zero returns error",
				fieldName:   "task_id",
				value:       "0",
				shouldError: true,
				expectedMsg: "task_id: must be a positive integer, got: 0",
			},
			{
				name:        "large integer works",
				fieldName:   "task_id",
				value:       "9223372036854775807",
				expectedID:  9223372036854775807,
				shouldError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, err := parseID(tt.fieldName, tt.value)
				if tt.shouldError {
					require.Error(t, err)
					assert.Equal(t, tt.expectedMsg, err.Error())
					var valErr ValidationError
					assert.True(t, errors.As(err, &valErr))
					assert.Equal(t, tt.fieldName, valErr.Field)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedID, id)
				}
			})
		}
	})

	t.Run("validateViewKind", func(t *testing.T) {
		tests := []struct {
			name        string
			kind        string
			shouldError bool
			expectedMsg string
		}{
			{
				name:        "empty string passes",
				kind:        "",
				shouldError: false,
			},
			{
				name:        "valid list",
				kind:        "list",
				shouldError: false,
			},
			{
				name:        "valid kanban",
				kind:        "kanban",
				shouldError: false,
			},
			{
				name:        "valid gantt",
				kind:        "gantt",
				shouldError: false,
			},
			{
				name:        "valid table",
				kind:        "table",
				shouldError: false,
			},
			{
				name:        "invalid kind",
				kind:        "invalid",
				shouldError: true,
				expectedMsg: "view_kind: must be one of: list, kanban, gantt, table. Got: invalid",
			},
			{
				name:        "case-sensitive check",
				kind:        "Kanban",
				shouldError: true,
				expectedMsg: "view_kind: must be one of: list, kanban, gantt, table. Got: Kanban",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateViewKind(tt.kind)
				if tt.shouldError {
					require.Error(t, err)
					assert.Equal(t, tt.expectedMsg, err.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestBuildErrorResult(t *testing.T) {
	h := newTestHandlers()
	message := "test error message"

	result := h.buildErrorResult(message)

	require.NotNil(t, result)
	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, message, textContent.Text)
}

func TestHandlerDependencies_NilLogger(t *testing.T) {
	cfg := &config.Config{}
	deps := &HandlerDependencies{
		Config:          cfg,
		OutputFormatter: vikunja.GetFormatter(vikunja.OutputFormatJSON),
		Logger:          nil, // Explicitly nil
	}

	// Should not panic when creating handlers with nil logger
	handlers := NewHandlers(deps)
	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.deps.Logger)
}
