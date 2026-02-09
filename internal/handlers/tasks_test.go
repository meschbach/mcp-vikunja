package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTasksHandler_BucketFiltering(t *testing.T) {
	// Initialize global output formatter for testing
	outputFormatter = vikunja.GetFormatter("text")

	// Mock Vikunja API
	mux := http.NewServeMux()

	// 1. Projects
	mux.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		projects := []vikunja.Project{
			{ID: 1, Title: "Test Project"},
		}
		json.NewEncoder(w).Encode(projects)
	})

	// 2. Views
	mux.HandleFunc("/api/v1/projects/1/views", func(w http.ResponseWriter, r *http.Request) {
		views := []vikunja.ProjectView{
			{ID: 10, ProjectID: 1, Title: "Kanban", ViewKind: vikunja.ViewKindKanban},
		}
		json.NewEncoder(w).Encode(views)
	})

	// 3. Tasks in View
	mux.HandleFunc("/api/v1/projects/1/views/10/tasks", func(w http.ResponseWriter, r *http.Request) {
		// Return buckets with tasks
		buckets := []vikunja.Bucket{
			{
				ID:    100,
				Title: "Todo",
				Tasks: []vikunja.Task{
					{ID: 1001, Title: "Task 1"},
				},
			},
			{
				ID:    200,
				Title: "In Progress",
				Tasks: []vikunja.Task{
					{ID: 2001, Title: "Task 2"},
				},
			},
		}
		json.NewEncoder(w).Encode(buckets)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Parse server URL to get host
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	// Set environment variables to point to mock server
	os.Setenv("VIKUNJA_HOST", u.Host)
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	os.Setenv("VIKUNJA_INSECURE", "true")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
		os.Unsetenv("VIKUNJA_INSECURE")
	}()

	ctx := context.Background()

	t.Run("filter by bucket_id", func(t *testing.T) {
		input := ListTasksInput{
			ProjectID: "1",
			ViewID:    "10",
			BucketID:  "200",
		}

		_, output, err := listTasksHandler(ctx, nil, input)
		require.NoError(t, err)

		// Verify only one bucket is returned
		assert.Equal(t, 1, len(output.View.Buckets))
		assert.Equal(t, int64(200), output.View.Buckets[0].Bucket.ID)
		assert.Equal(t, "In Progress", output.View.Buckets[0].Bucket.Title)

		// Verify tasks in that bucket
		assert.Equal(t, 1, len(output.View.Buckets[0].Tasks))
		assert.Equal(t, int64(2001), output.View.Buckets[0].Tasks[0].ID)
		assert.Equal(t, "Task 2", output.View.Buckets[0].Tasks[0].Title)
	})

	t.Run("filter by bucket_title", func(t *testing.T) {
		input := ListTasksInput{
			ProjectTitle: "Test Project",
			ViewTitle:    "Kanban",
			BucketTitle:  "Todo",
		}

		_, output, err := listTasksHandler(ctx, nil, input)
		require.NoError(t, err)

		// Verify only one bucket is returned
		assert.Equal(t, 1, len(output.View.Buckets))
		assert.Equal(t, int64(100), output.View.Buckets[0].Bucket.ID)
		assert.Equal(t, "Todo", output.View.Buckets[0].Bucket.Title)

		// Verify tasks in that bucket
		assert.Equal(t, 1, len(output.View.Buckets[0].Tasks))
		assert.Equal(t, int64(1001), output.View.Buckets[0].Tasks[0].ID)
		assert.Equal(t, "Task 1", output.View.Buckets[0].Tasks[0].Title)
	})

	t.Run("bucket not found", func(t *testing.T) {
		input := ListTasksInput{
			ProjectID:   "1",
			ViewID:      "10",
			BucketTitle: "Non Existent",
		}

		_, _, err := listTasksHandler(ctx, nil, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket with title \"Non Existent\" not found in view \"Kanban\"")
	})
}

func TestListTasksHandler_BucketFilteringValidation(t *testing.T) {
	// Set environment variables
	os.Setenv("VIKUNJA_HOST", "test.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
	}()

	ctx := context.Background()

	tests := []struct {
		name        string
		input       ListTasksInput
		expectError string
	}{
		{
			name: "both bucket_id and bucket_title specified",
			input: ListTasksInput{
				ProjectTitle: "Test Project",
				ViewTitle:    "Kanban",
				BucketID:     "1",
				BucketTitle:  "Test Bucket",
			},
			expectError: "cannot specify both bucket_id and bucket_title",
		},
		{
			name: "invalid bucket_id format",
			input: ListTasksInput{
				ProjectTitle: "Test Project",
				ViewTitle:    "Kanban",
				BucketID:     "invalid",
			},
			expectError: "invalid bucket_id: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := listTasksHandler(ctx, nil, tt.input)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Equal(t, ListTasksOutput{}, output)
		})
	}
}

func TestListTasksHandler_MissingEnvironment(t *testing.T) {
	// Test with missing environment variables
	os.Unsetenv("VIKUNJA_HOST")
	os.Unsetenv("VIKUNJA_TOKEN")

	ctx := context.Background()
	input := ListTasksInput{
		ProjectTitle: "Test Project",
	}

	result, output, err := listTasksHandler(ctx, nil, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	assert.Nil(t, result)
	assert.Equal(t, ListTasksOutput{}, output)
}

// Mock client for testing bucket filtering logic
type mockVikunjaClient struct {
	buckets []vikunja.Bucket
	tasks   []vikunja.Task
}

func (m *mockVikunjaClient) GetProjects(ctx context.Context) ([]vikunja.Project, error) {
	return []vikunja.Project{
		{ID: 1, Title: "Test Project"},
	}, nil
}

func (m *mockVikunjaClient) GetProjectViews(ctx context.Context, projectID int64) ([]vikunja.ProjectView, error) {
	return []vikunja.ProjectView{
		{ID: 1, Title: "Kanban"},
	}, nil
}

func (m *mockVikunjaClient) GetViewTasks(ctx context.Context, projectID, viewID int64) (*vikunja.ViewTasksResponse, error) {
	return &vikunja.ViewTasksResponse{
		Buckets: m.buckets,
		Tasks:   m.tasks,
	}, nil
}

// Test for bucket filtering with mocked client would require refactoring the handler
// to accept client injection. For now, we test validation logic which is the main
// new functionality we've added.

func TestListTasksHandler_NoBucketFilter_BackwardCompatibility(t *testing.T) {
	// Test that the function works without bucket filtering (backward compatibility)
	os.Setenv("VIKUNJA_HOST", "test.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
	}()

	ctx := context.Background()
	input := ListTasksInput{
		ProjectTitle: "Test Project", // Will fail at API call but validates our input parsing
	}

	result, output, err := listTasksHandler(ctx, nil, input)

	// Should fail at API call since we're using invalid host, but not at input validation
	assert.Error(t, err)
	// The error will be about listing projects, not creating client or input validation
	assert.Contains(t, err.Error(), "failed to list projects")
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	assert.Equal(t, ListTasksOutput{}, output)
}
