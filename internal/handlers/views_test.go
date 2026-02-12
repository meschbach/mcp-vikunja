package handlers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindViewHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         FindViewInput
		expectedError string
	}{
		{
			name:          "missing view_name",
			input:         FindViewInput{ViewName: ""},
			expectedError: "view_name: is required",
		},
		{
			name: "valid view_name with no project specified",
			input: FindViewInput{
				ViewName: "Kanban",
			},
			expectedError: "either project_id or project_title must be specified",
		},
		{
			name: "valid view_name with project_id",
			input: FindViewInput{
				ViewName:  "Kanban",
				ProjectID: "1",
			},
			expectedError: "failed to fetch project views", // Will fail due to network
		},
		{
			name: "valid view_name with project_title",
			input: FindViewInput{
				ViewName:     "Kanban",
				ProjectTitle: "Inbox",
			},
			expectedError: "failed to list projects", // Will fail due to network
		},
		{
			name: "invalid project_id format",
			input: FindViewInput{
				ViewName:  "Kanban",
				ProjectID: "invalid",
			},
			expectedError: "invalid project_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.findViewHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.True(t, result.IsError)
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					assert.Contains(t, textContent.Text, tt.expectedError)
				}
			}
			assert.Equal(t, FindViewOutput{}, output)
		})
	}
}

func TestListViewsHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         ListViewsInput
		expectedError string
	}{
		{
			name:          "no project specified",
			input:         ListViewsInput{},
			expectedError: "either project_id or project_title must be specified",
		},
		{
			name: "valid with project_id",
			input: ListViewsInput{
				ProjectID: "1",
			},
			expectedError: "failed to fetch project views", // Will fail due to network
		},
		{
			name: "valid with project_title",
			input: ListViewsInput{
				ProjectTitle: "Inbox",
			},
			expectedError: "failed to list projects", // Will fail due to network
		},
		{
			name: "invalid project_id format",
			input: ListViewsInput{
				ProjectID: "invalid",
			},
			expectedError: "invalid project_id",
		},
		{
			name: "invalid view_kind",
			input: ListViewsInput{
				ProjectID: "1",
				ViewKind:  "invalid_kind",
			},
			expectedError: "view_kind: must be one of",
		},
		{
			name: "valid view_kind list",
			input: ListViewsInput{
				ProjectID: "1",
				ViewKind:  "list",
			},
			expectedError: "failed to fetch project views", // Will fail due to network
		},
		{
			name: "valid view_kind kanban",
			input: ListViewsInput{
				ProjectID: "1",
				ViewKind:  "kanban",
			},
			expectedError: "failed to fetch project views", // Will fail due to network
		},
		{
			name: "valid view_kind gantt",
			input: ListViewsInput{
				ProjectID: "1",
				ViewKind:  "gantt",
			},
			expectedError: "failed to fetch project views", // Will fail due to network
		},
		{
			name: "valid view_kind table",
			input: ListViewsInput{
				ProjectID: "1",
				ViewKind:  "table",
			},
			expectedError: "failed to fetch project views", // Will fail due to network
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.listViewsHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.True(t, result.IsError)
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					assert.Contains(t, textContent.Text, tt.expectedError)
				}
			}
			assert.Equal(t, ListViewsOutput{}, output)
		})
	}
}

func TestListTasksHandler_Validation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         ListTasksInput
		expectedError string
	}{
		{
			name: "non-numeric project is treated as title",
			input: ListTasksInput{
				Project: "invalid",
			},
			expectedError: "failed to list projects", // Will try to find "invalid" as title and fail on network
		},
		{
			name: "non-numeric view with valid project",
			input: ListTasksInput{
				Project: "1", // Valid ID to get past project resolution
				View:    "invalid-view",
			},
			expectedError: "project with ID 1 not found", // Will fail due to network when getting project
		},
		{
			name:  "default project and view search",
			input: ListTasksInput{
				// No project specified, should default to "Inbox" project
			},
			expectedError: "failed to list projects", // Will fail due to network when looking for "Inbox"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.listTasksHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			if result != nil {
				assert.True(t, result.IsError)
				if len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						assert.Contains(t, textContent.Text, tt.expectedError)
					}
				}
			}
			assert.Equal(t, ListTasksOutput{}, output)
		})
	}
}

func TestListTasksHandler_BucketFilteringValidation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	// Test that bucket filtering with a non-existent bucket fails gracefully
	input := ListTasksInput{
		Project: "1",
		View:    "1",
		Bucket:  "nonexistent-bucket",
	}

	_, output, err := h.listTasksHandler(context.Background(), &mcp.CallToolRequest{}, input)

	// Should fail because project "1" doesn't exist
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project with ID 1 not found")
	assert.Equal(t, ListTasksOutput{}, output)
}

func TestListTasksHandler_WithProjectTitle(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	// Test with project as title - should look for project by title
	input := ListTasksInput{
		Project: "Inbox",
	}

	result, output, err := h.listTasksHandler(context.Background(), &mcp.CallToolRequest{}, input)

	// Should fail because we can't connect to Vikunja
	require.Error(t, err)
	// Result might be nil or have error content depending on where it fails
	if result != nil {
		assert.True(t, result.IsError)
		assert.Contains(t, err.Error(), "failed to list projects")
	}
	assert.Equal(t, ListTasksOutput{}, output)
}

func TestFindProjectByIDOrTitle(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	// We can't easily test the success path without mocking the client,
	// but we can test the validation and error paths

	t.Run("no project_id or project_title", func(t *testing.T) {
		// Create a minimal client to test findProjectByIDOrTitle directly
		// Since we can't call it directly (it's not exported), test via listViewsHandler
		input := ListViewsInput{}
		result, _, err := h.listViewsHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.Error(t, err)
		assert.True(t, result.IsError)
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				assert.Contains(t, textContent.Text, "either project_id or project_title must be specified")
			}
		}
	})

	t.Run("invalid project_id format", func(t *testing.T) {
		input := ListViewsInput{
			ProjectID: "not-a-number",
		}
		result, _, err := h.listViewsHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.Error(t, err)
		assert.True(t, result.IsError)
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				assert.Contains(t, textContent.Text, "invalid project_id")
			}
		}
	})
}

func TestEnhancedErrorMessages(t *testing.T) {
	t.Run("enhancedProjectNotFoundError", func(t *testing.T) {
		// Test with empty available projects
		err := enhancedProjectNotFoundError("MissingProject", []string{})
		assert.Contains(t, err.Error(), "project with title \"MissingProject\" not found")
		assert.Contains(t, err.Error(), "Try: list_projects()")

		// Test with few available projects
		err = enhancedProjectNotFoundError("MissingProject", []string{"Project1", "Project2"})
		assert.Contains(t, err.Error(), "project with title \"MissingProject\" not found")
		assert.Contains(t, err.Error(), "Available projects:")
		assert.Contains(t, err.Error(), "Project1")
		assert.Contains(t, err.Error(), "Project2")

		// Test with many available projects
		manyProjects := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
		err = enhancedProjectNotFoundError("MissingProject", manyProjects)
		assert.Contains(t, err.Error(), "project with title \"MissingProject\" not found")
		assert.Contains(t, err.Error(), "Available projects include:")
		assert.Contains(t, err.Error(), "Alpha")
		assert.Contains(t, err.Error(), "Beta")
		assert.Contains(t, err.Error(), "and 3 others")
	})

	t.Run("enhancedViewNotFoundError", func(t *testing.T) {
		// Test with empty available views
		err := enhancedViewNotFoundError("MissingView", "MyProject", []string{})
		assert.Contains(t, err.Error(), "view with title \"MissingView\" not found")
		assert.Contains(t, err.Error(), "Try: list_views()")

		// Test with few available views
		err = enhancedViewNotFoundError("MissingView", "MyProject", []string{"Kanban", "List"})
		assert.Contains(t, err.Error(), "view with title \"MissingView\" not found")
		assert.Contains(t, err.Error(), "Available views in project 'MyProject':")
		assert.Contains(t, err.Error(), "Kanban")
		assert.Contains(t, err.Error(), "List")

		// Test with many available views
		manyViews := []string{"View1", "View2", "View3", "View4"}
		err = enhancedViewNotFoundError("MissingView", "MyProject", manyViews)
		assert.Contains(t, err.Error(), "view with title \"MissingView\" not found")
		assert.Contains(t, err.Error(), "Available views in project 'MyProject' include:")
		assert.Contains(t, err.Error(), "View1")
		assert.Contains(t, err.Error(), "View2")
		assert.Contains(t, err.Error(), "and 2 others")
	})
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "Hello World", true},
		{"Hello World", "xyz", false},
		{"", "test", false},
		{"test", "", true},
		{"UPPERCASE", "upper", true},
		{"lowercase", "LOWER", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := containsIgnoreCase(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindViewByName(t *testing.T) {
	views := []vikunja.ProjectView{
		{ID: 1, Title: "Kanban"},
		{ID: 2, Title: "List"},
		{ID: 3, Title: "Gantt"},
	}

	tests := []struct {
		name        string
		viewName    string
		fuzzy       bool
		expectFound bool
		expectedID  int64
	}{
		{"exact match", "Kanban", false, true, 1},
		{"case-sensitive exact match", "kanban", false, false, 0},
		{"fuzzy match", "Kan", true, true, 1},
		{"fuzzy match case insensitive", "LIST", true, true, 2},
		{"no match", "Calendar", false, false, 0},
		{"fuzzy no match", "XYZ", true, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := findViewByName(views, tt.viewName, tt.fuzzy, "TestProject")

			if tt.expectFound {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, view.ID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			}
		})
	}
}

func TestToView(t *testing.T) {
	pv := vikunja.ProjectView{
		ID:                      123,
		ProjectID:               456,
		Title:                   "Test View",
		ViewKind:                vikunja.ViewKindKanban,
		Position:                1.5,
		BucketConfigurationMode: vikunja.BucketConfigurationModeManual,
		DefaultBucketID:         789,
		DoneBucketID:            999,
	}

	view := toView(pv)

	assert.Equal(t, int64(123), view.ID)
	assert.Equal(t, int64(456), view.ProjectID)
	assert.Equal(t, "Test View", view.Title)
	assert.Equal(t, vikunja.ViewKindKanban, view.ViewKind)
	assert.Equal(t, 1.5, view.Position)
	assert.Equal(t, vikunja.BucketConfigurationModeManual, view.BucketConfigurationMode)
	assert.Equal(t, int64(789), view.DefaultBucketID)
	assert.Equal(t, int64(999), view.DoneBucketID)
	assert.Equal(t, "vikunja://project/456/view/123", view.URI)
}

func TestToViews(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := toViews(nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		result := toViews([]vikunja.ProjectView{})
		assert.Empty(t, result)
	})

	t.Run("multiple views", func(t *testing.T) {
		views := []vikunja.ProjectView{
			{ID: 1, Title: "Kanban"},
			{ID: 2, Title: "List"},
		}
		result := toViews(views)
		require.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Kanban", result[0].Title)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, "List", result[1].Title)
	})
}

func TestMoveTaskToBucketHandler_ExpandedValidation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	h := newTestHandlers()

	tests := []struct {
		name          string
		input         MoveTaskToBucketInput
		expectedError string
	}{
		{
			name:          "missing task_id",
			input:         MoveTaskToBucketInput{ProjectID: "1", ViewID: "2", BucketID: "3"},
			expectedError: "task_id: is required",
		},
		{
			name:          "missing project_id",
			input:         MoveTaskToBucketInput{TaskID: "1", ViewID: "2", BucketID: "3"},
			expectedError: "project_id: is required",
		},
		{
			name:          "missing view_id",
			input:         MoveTaskToBucketInput{TaskID: "1", ProjectID: "2", BucketID: "3"},
			expectedError: "view_id: is required",
		},
		{
			name:          "missing bucket_id",
			input:         MoveTaskToBucketInput{TaskID: "1", ProjectID: "2", ViewID: "3"},
			expectedError: "bucket_id: is required",
		},
		{
			name:          "invalid task_id",
			input:         MoveTaskToBucketInput{TaskID: "invalid", ProjectID: "1", ViewID: "2", BucketID: "3"},
			expectedError: "task_id: must be a valid integer",
		},
		{
			name:          "invalid project_id",
			input:         MoveTaskToBucketInput{TaskID: "1", ProjectID: "invalid", ViewID: "2", BucketID: "3"},
			expectedError: "project_id: must be a valid integer",
		},
		{
			name:          "invalid view_id",
			input:         MoveTaskToBucketInput{TaskID: "1", ProjectID: "2", ViewID: "invalid", BucketID: "3"},
			expectedError: "view_id: must be a valid integer",
		},
		{
			name:          "invalid bucket_id",
			input:         MoveTaskToBucketInput{TaskID: "1", ProjectID: "2", ViewID: "3", BucketID: "invalid"},
			expectedError: "bucket_id: must be a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := h.moveTaskToBucketHandler(context.Background(), &mcp.CallToolRequest{}, tt.input)

			require.Error(t, err)
			assert.True(t, result.IsError)
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					assert.Contains(t, textContent.Text, tt.expectedError)
				}
			}
			assert.Equal(t, MoveTaskToBucketOutput{}, output)
		})
	}
}

func TestCreateVikunjaClient_MissingEnvVars(t *testing.T) {
	// Save current env vars
	oldHost := os.Getenv("VIKUNJA_HOST")
	oldToken := os.Getenv("VIKUNJA_TOKEN")
	defer func() {
		os.Setenv("VIKUNJA_HOST", oldHost)
		os.Setenv("VIKUNJA_TOKEN", oldToken)
	}()

	tests := []struct {
		name        string
		host        string
		token       string
		expectError bool
	}{
		{
			name:        "missing both",
			host:        "",
			token:       "",
			expectError: true,
		},
		{
			name:        "missing host",
			host:        "",
			token:       "token",
			expectError: true,
		},
		{
			name:        "missing token",
			host:        "host",
			token:       "",
			expectError: true,
		},
		{
			name:        "both present",
			host:        "test.example.com",
			token:       "test-token",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("VIKUNJA_HOST", tt.host)
			os.Setenv("VIKUNJA_TOKEN", tt.token)

			client, err := createVikunjaClient()

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, client)
				assert.Contains(t, err.Error(), "VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		expected string // RFC3339 format or empty for zero time
	}{
		{
			name:     "empty string returns zero time",
			timeStr:  "",
			expected: "",
		},
		{
			name:     "valid RFC3339 time",
			timeStr:  "2024-01-15T10:30:00Z",
			expected: "2024-01-15T10:30:00Z",
		},
		{
			name:     "invalid format returns zero time",
			timeStr:  "not-a-valid-time",
			expected: "",
		},
		{
			name:     "RFC3339 with timezone",
			timeStr:  "2024-01-15T10:30:00-05:00",
			expected: "2024-01-15T10:30:00-05:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTime(tt.timeStr)
			if tt.expected == "" {
				assert.True(t, result.IsZero())
			} else {
				assert.Equal(t, tt.expected, result.Format(time.RFC3339))
			}
		})
	}
}

func TestToVikunjaBuckets(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := toVikunjaBuckets(nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		result := toVikunjaBuckets([]Bucket{})
		assert.Empty(t, result)
	})

	t.Run("converts buckets correctly", func(t *testing.T) {
		buckets := []Bucket{
			{ID: 1, Title: "Bucket 1", ProjectViewID: 100},
			{ID: 2, Title: "Bucket 2", ProjectViewID: 100, Description: "Test description", Limit: 5},
		}
		result := toVikunjaBuckets(buckets)
		require.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Bucket 1", result[0].Title)
		assert.Equal(t, int64(100), result[0].ProjectViewID)
		assert.Equal(t, "Bucket 2", result[1].Title)
		assert.Equal(t, "Test description", result[1].Description)
		assert.Equal(t, 5, result[1].Limit)
	})
}

func TestConvertToVikunjaViewTasksSummary(t *testing.T) {
	h := newTestHandlers()

	vt := ViewTasksSummary{
		ViewID:    123,
		ViewTitle: "Test View",
		Buckets: []BucketTasksSummary{
			{
				Bucket: BucketSummary{ID: 1, Title: "Bucket 1"},
				Tasks: []TaskSummary{
					{ID: 101, Title: "Task 1"},
					{ID: 102, Title: "Task 2"},
				},
			},
			{
				Bucket: BucketSummary{ID: 2, Title: "Bucket 2"},
				Tasks:  []TaskSummary{},
			},
		},
	}

	result := h.convertToVikunjaViewTasksSummary(vt)

	assert.Equal(t, int64(123), result.ViewID)
	assert.Equal(t, "Test View", result.ViewTitle)
	require.Len(t, result.Buckets, 2)
	assert.Equal(t, int64(1), result.Buckets[0].Bucket.ID)
	assert.Equal(t, "Bucket 1", result.Buckets[0].Bucket.Title)
	require.Len(t, result.Buckets[0].Tasks, 2)
	assert.Equal(t, int64(101), result.Buckets[0].Tasks[0].ID)
	assert.Equal(t, "Task 2", result.Buckets[0].Tasks[1].Title)
	assert.Empty(t, result.Buckets[1].Tasks)
}

func TestBuildViewTasksSummary(t *testing.T) {
	h := newTestHandlers()

	t.Run("with buckets", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket 1", Tasks: []vikunja.Task{{ID: 101, Title: "Task 1"}}},
				{ID: 2, Title: "Bucket 2", Tasks: []vikunja.Task{{ID: 102, Title: "Task 2"}}},
			},
		}

		result := h.buildViewTasksSummary(123, "Test View", resp)

		assert.Equal(t, int64(123), result.ViewID)
		assert.Equal(t, "Test View", result.ViewTitle)
		require.Len(t, result.Buckets, 2)
		assert.Equal(t, "Bucket 1", result.Buckets[0].Bucket.Title)
		assert.Equal(t, "Bucket 2", result.Buckets[1].Bucket.Title)
	})

	t.Run("without buckets (flat tasks)", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Tasks: []vikunja.Task{
				{ID: 101, Title: "Task 1"},
				{ID: 102, Title: "Task 2"},
			},
		}

		result := h.buildViewTasksSummary(456, "List View", resp)

		assert.Equal(t, int64(456), result.ViewID)
		assert.Equal(t, "List View", result.ViewTitle)
		require.Len(t, result.Buckets, 1)
		assert.Equal(t, int64(0), result.Buckets[0].Bucket.ID)
		assert.Equal(t, "All Tasks", result.Buckets[0].Bucket.Title)
		require.Len(t, result.Buckets[0].Tasks, 2)
	})
}
