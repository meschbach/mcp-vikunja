// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTaskHandler_Success(t *testing.T) {
	t.Skip()
	t.Parallel()

	tests := []struct {
		name           string
		setupMock      func(w http.ResponseWriter, r *http.Request)
		input          CreateTaskInput
		expectedTaskID int64
		expectedTitle  string
	}{
		{
			name: "project ID only",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/projects/1/tasks", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				var reqBody map[string]interface{}
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
				assert.Equal(t, "Test Task", reqBody["title"])
				assert.InEpsilon(t, reqBody["project_id"], float64(1), 0)
				assert.NotContains(t, reqBody, "bucket_id")
				w.WriteHeader(http.StatusCreated)
				assert.NoError(t, json.NewEncoder(w).Encode(vikunja.Task{
					ID:        456,
					Title:     "Test Task",
					ProjectID: 1,
				}))
			},
			input: CreateTaskInput{
				Title:     "Test Task",
				ProjectID: "1",
			},
			expectedTaskID: 456,
			expectedTitle:  "Test Task",
		},
		{
			name: "with description and bucket ID",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				var reqBody map[string]interface{}
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
				assert.Equal(t, "Full Task", reqBody["title"])
				assert.Equal(t, "Full description", reqBody["description"])
				assert.InEpsilon(t, reqBody["project_id"], float64(1), 0)
				assert.InEpsilon(t, reqBody["bucket_id"], float64(5), 0)
				w.WriteHeader(http.StatusCreated)
				assert.NoError(t, json.NewEncoder(w).Encode(vikunja.Task{
					ID:          789,
					Title:       "Full Task",
					Description: "Full description",
				}))
			},
			input: CreateTaskInput{
				Title:       "Full Task",
				ProjectID:   "1",
				Description: "Full description",
				BucketID:    "5",
			},
			expectedTaskID: 789,
			expectedTitle:  "Full Task",
		},
		{
			name: "project by title",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.Project{
						{ID: 1, Title: "Test Project"},
					})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/tasks" && r.Method == "POST" {
					var body map[string]interface{}
					json.NewDecoder(r.Body).Decode(&body)
					assert.InEpsilon(t, body["project_id"], float64(1), 0)
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(vikunja.Task{
						ID:        456,
						Title:     "Task",
						ProjectID: 1,
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "Test Project",
			},
			expectedTaskID: 456,
			expectedTitle:  "Task",
		},
		{
			name: "bucket by title with kanban view",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects/1" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(vikunja.Project{ID: 1, Title: "Project 1"})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.ProjectView{
						{ID: 100, ProjectID: 1, Title: "Kanban", ViewKind: vikunja.ViewKindKanban},
					})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views/100/buckets" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.Bucket{
						{ID: 10, ProjectViewID: 100, Title: "Todo"},
						{ID: 20, ProjectViewID: 100, Title: "In Progress"},
						{ID: 30, ProjectViewID: 100, Title: "Done"},
					})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/tasks" && r.Method == "POST" {
					var body map[string]interface{}
					json.NewDecoder(r.Body).Decode(&body)
					assert.InEpsilon(t, body["bucket_id"], float64(20), 0)
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(vikunja.Task{
						ID:        456,
						Title:     "Task",
						ProjectID: 1,
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "1",
				BucketID:  "In Progress",
			},
			expectedTaskID: 456,
			expectedTitle:  "Task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleanup := setupMockServer(tt.setupMock)
			defer cleanup()

			h := newTestHandlers()
			result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Equal(t, tt.expectedTaskID, output.Task.ID)
			assert.Equal(t, tt.expectedTitle, output.Task.Title)
		})
	}
}

func TestCreateTaskHandler_ProjectResolutionErrors(t *testing.T) {
	t.Skip()
	t.Parallel()

	tests := []struct {
		name           string
		setupMock      func(w http.ResponseWriter, r *http.Request)
		input          CreateTaskInput
		expectedErrMsg string
	}{
		{
			name: "project title not found",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.Project{
						{ID: 1, Title: "Other Project"},
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "Nonexistent",
			},
			expectedErrMsg: "project with title \"Nonexistent\" not found",
		},
		{
			name: "duplicate project titles",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.Project{
						{ID: 1, Title: "Duplicate"},
						{ID: 2, Title: "Duplicate"},
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "Duplicate",
			},
			expectedErrMsg: "multiple projects found with title \"Duplicate\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleanup := setupMockServer(tt.setupMock)
			defer cleanup()

			h := newTestHandlers()
			result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
			assert.Equal(t, CreateTaskOutput{}, output)
		})
	}
}

func TestCreateTaskHandler_BucketResolutionErrors(t *testing.T) {
	t.Skip()
	t.Parallel()

	tests := []struct {
		name           string
		setupMock      func(w http.ResponseWriter, r *http.Request)
		input          CreateTaskInput
		expectedErrMsg string
	}{
		{
			name: "bucket title not found in kanban view",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects/1" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(vikunja.Project{ID: 1, Title: "Project 1"})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.ProjectView{
						{ID: 100, ProjectID: 1, Title: "Kanban", ViewKind: vikunja.ViewKindKanban},
					})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views/100/buckets" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.Bucket{
						{ID: 10, ProjectViewID: 100, Title: "Todo"},
						{ID: 20, ProjectViewID: 100, Title: "In Progress"},
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "1",
				BucketID:  "Nonexistent",
			},
			expectedErrMsg: "bucket \"Nonexistent\" not found in Kanban view of project \"Project 1\"",
		},
		{
			name: "missing kanban view",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects/1" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(vikunja.Project{ID: 1, Title: "Project 1"})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.ProjectView{
						{ID: 100, ProjectID: 1, Title: "List", ViewKind: vikunja.ViewKindList},
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "1",
				BucketID:  "Todo",
			},
			expectedErrMsg: "kanban view not found in project 1",
		},
		{
			name: "bucket title case-sensitive",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects/1" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(vikunja.Project{ID: 1, Title: "Project 1"})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.ProjectView{
						{ID: 100, ProjectID: 1, Title: "Kanban", ViewKind: vikunja.ViewKindKanban},
					})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views/100/buckets" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.Bucket{
						{ID: 10, ProjectViewID: 100, Title: "Todo"},
					})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "1",
				BucketID:  "todo", // lowercase, should not match "Todo"
			},
			expectedErrMsg: "bucket \"todo\" not found in Kanban view of project \"Project 1\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleanup := setupMockServer(tt.setupMock)
			defer cleanup()

			h := newTestHandlers()
			result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
			assert.Equal(t, CreateTaskOutput{}, output)
		})
	}
}

func TestCreateTaskHandler_BucketIDValidation(t *testing.T) {
	t.Skip()
	t.Parallel()

	tests := []struct {
		name          string
		input         CreateTaskInput
		expectedError string
	}{
		{
			name:          "negative bucket_id",
			input:         CreateTaskInput{Title: "Test Task", ProjectID: "1", BucketID: "-5"},
			expectedError: "bucket_id: must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("Should not make HTTP request for invalid input")
			})
			defer cleanup()

			h := newTestHandlers()
			result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Equal(t, CreateTaskOutput{}, output)
		})
	}
}

func TestCreateTaskHandler_APIErrors(t *testing.T) {
	t.Skip()
	t.Parallel()

	tests := []struct {
		name           string
		setupMock      func(w http.ResponseWriter, r *http.Request)
		input          CreateTaskInput
		expectedErrMsg string
	}{
		{
			name: "create task bad request",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" && r.URL.Path == "/api/v1/projects/1/tasks" {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Invalid request"})
				}
			},
			input: CreateTaskInput{
				Title:     "Test Task",
				ProjectID: "1",
			},
			expectedErrMsg: "failed to create task",
		},
		{
			name: "create task unauthorized",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" && r.URL.Path == "/api/v1/projects/1/tasks" {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Unauthorized"})
				}
			},
			input: CreateTaskInput{
				Title:     "Test Task",
				ProjectID: "1",
			},
			expectedErrMsg: "failed to create task",
		},
		{
			name: "create task server error",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" && r.URL.Path == "/api/v1/projects/1/tasks" {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Internal server error"})
				}
			},
			input: CreateTaskInput{
				Title:     "Test Task",
				ProjectID: "1",
			},
			expectedErrMsg: "failed to create task",
		},
		{
			name: "get project views error",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects/1" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(vikunja.Project{ID: 1, Title: "Project 1"})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views" {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Server error"})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "1",
				BucketID:  "Todo",
			},
			expectedErrMsg: "failed to get project views",
		},
		{
			name: "get view buckets error",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/projects/1" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(vikunja.Project{ID: 1, Title: "Project 1"})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]vikunja.ProjectView{
						{ID: 100, ProjectID: 1, Title: "Kanban", ViewKind: vikunja.ViewKindKanban},
					})
					return
				}
				if r.URL.Path == "/api/v1/projects/1/views/100/buckets" {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Server error"})
					return
				}
			},
			input: CreateTaskInput{
				Title:     "Task",
				ProjectID: "1",
				BucketID:  "Todo",
			},
			expectedErrMsg: "failed to get buckets for kanban view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleanup := setupMockServer(tt.setupMock)
			defer cleanup()

			h := newTestHandlers()
			result, output, err := h.createTaskHandler(t.Context(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
			assert.Equal(t, CreateTaskOutput{}, output)
		})
	}
}

func TestCreateTaskHandler_ReadOnly(t *testing.T) {
	t.Skip()
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
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "readonly")
		}
	}
	assert.Equal(t, CreateTaskOutput{}, output)
}

func TestCreateTaskHandler_Validation(t *testing.T) {
	t.Skip()
	t.Parallel()

	tests := []struct {
		name           string
		input          CreateTaskInput
		expectedErrMsg string
	}{
		{
			name: "missing title",
			input: CreateTaskInput{
				ProjectID: "1",
			},
			expectedErrMsg: "title: is required",
		},
		{
			name: "missing project_id",
			input: CreateTaskInput{
				Title: "Test Task",
			},
			expectedErrMsg: "project_id: is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("Should not make HTTP request for invalid input")
			})
			defer cleanup()

			h := newTestHandlers()
			result, output, err := h.createTaskHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
			assert.Equal(t, CreateTaskOutput{}, output)
		})
	}
}
