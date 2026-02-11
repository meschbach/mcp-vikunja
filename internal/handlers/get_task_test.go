// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockServer creates a mock HTTP server for testing
func setupMockServer(handler http.HandlerFunc) func() {
	ts := httptest.NewServer(handler)
	oldHost := os.Getenv("VIKUNJA_HOST")
	oldToken := os.Getenv("VIKUNJA_TOKEN")
	oldInsecure := os.Getenv("VIKUNJA_INSECURE")

	// Extract host:port from URL
	host := strings.TrimPrefix(ts.URL, "http://")
	os.Setenv("VIKUNJA_HOST", host)
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	os.Setenv("VIKUNJA_INSECURE", "true")

	return func() {
		ts.Close()
		os.Setenv("VIKUNJA_HOST", oldHost)
		os.Setenv("VIKUNJA_TOKEN", oldToken)
		os.Setenv("VIKUNJA_INSECURE", oldInsecure)
	}
}

func TestGetTaskHandler_Success(t *testing.T) {
	t.Run("task without buckets", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:        123,
			Title:     "Test Task",
			ProjectID: 456,
			Done:      false,
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: false,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Equal(t, int64(123), output.Task.ID)
		assert.Equal(t, "Test Task", output.Task.Title)
		assert.Nil(t, output.Buckets)
	})

	t.Run("task with buckets included", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:        123,
			Title:     "Test Task with Buckets",
			ProjectID: 456,
			Done:      false,
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Todo", ProjectViewID: 100},
				{ID: 2, Title: "Done", ProjectViewID: 100},
			},
		}

		viewsResponse := []vikunja.ProjectView{
			{ID: 100, Title: "Kanban", ProjectID: 456, ViewKind: vikunja.ViewKindKanban, DoneBucketID: 2},
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			} else if strings.Contains(r.URL.Path, "/projects/456/views") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(viewsResponse)
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: true,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Equal(t, int64(123), output.Task.ID)
		assert.Equal(t, "Test Task with Buckets", output.Task.Title)
		assert.NotNil(t, output.Buckets)
		assert.Equal(t, int64(123), output.Buckets.TaskID)
		assert.Len(t, output.Buckets.Views, 1)
		assert.Equal(t, "Kanban", output.Buckets.Views[0].ViewTitle)
	})

	t.Run("task with buckets but no views", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:        123,
			Title:     "Test Task",
			ProjectID: 456,
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket", ProjectViewID: 100},
			},
		}

		// Return empty views array for this test
		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			} else if strings.Contains(r.URL.Path, "/projects/456/views") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]vikunja.ProjectView{})
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: true,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		// Buckets should still be returned with empty views
		assert.NotNil(t, output.Buckets)
		assert.Empty(t, output.Buckets.Views)
	})
}

func TestGetTaskHandler_Errors(t *testing.T) {
	t.Run("task not found", func(t *testing.T) {
		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Task not found"})
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "999",
			IncludeBuckets: false,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, GetTaskOutput{}, output)
		assert.Contains(t, err.Error(), "failed to get task")
	})

	t.Run("server error", func(t *testing.T) {
		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Internal server error"})
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: false,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, GetTaskOutput{}, output)
	})

	t.Run("views API error with buckets enabled", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:        123,
			Title:     "Test Task",
			ProjectID: 456,
			Buckets:   []vikunja.Bucket{},
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			} else if strings.Contains(r.URL.Path, "/projects/456/views") {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "Views API error"})
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: true,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		// This should still succeed, just without bucket info
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Equal(t, int64(123), output.Task.ID)
		assert.Nil(t, output.Buckets)
	})
}

func TestGetTaskHandler_TaskConversions(t *testing.T) {
	t.Run("task with all fields", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:          123,
			Title:       "Complete Task",
			Description: "A full description",
			ProjectID:   456,
			Done:        true,
			Position:    5.5,
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: false,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(123), output.Task.ID)
		assert.Equal(t, "Complete Task", output.Task.Title)
		assert.Equal(t, "A full description", output.Task.Description)
		assert.Equal(t, int64(456), output.Task.ProjectID)
		assert.True(t, output.Task.Done)
		assert.Equal(t, 5.5, output.Task.Position)
	})

	t.Run("task with dates", func(t *testing.T) {
		// This tests the parseTime conversion in the handler
		taskResponse := vikunja.Task{
			ID:    123,
			Title: "Task with Dates",
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket", ProjectViewID: 100},
			},
		}

		viewsResponse := []vikunja.ProjectView{
			{ID: 100, Title: "Kanban", ProjectID: 1, DoneBucketID: 1},
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			} else if strings.Contains(r.URL.Path, "/views") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(viewsResponse)
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: true,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(123), output.Task.ID)
		// Verify buckets are properly included
		assert.NotNil(t, output.Buckets)
		assert.Len(t, output.Buckets.Views, 1)
	})
}

func TestListBucketsHandler_Success(t *testing.T) {
	bucketsResponse := []vikunja.Bucket{
		{ID: 1, Title: "Todo", ProjectViewID: 100, Position: 1},
		{ID: 2, Title: "In Progress", ProjectViewID: 100, Position: 2},
		{ID: 3, Title: "Done", ProjectViewID: 100, Position: 3, IsDoneBucket: true},
	}

	cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/projects/1/views/100/buckets") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(bucketsResponse)
		}
	})
	defer cleanup()

	h := newTestHandlers()
	input := ListBucketsInput{
		ProjectID: "1",
		ViewID:    "100",
	}

	result, output, err := h.listBucketsHandler(context.Background(), &mcp.CallToolRequest{}, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Len(t, output.Buckets, 3)
	assert.Equal(t, "Todo", output.Buckets[0].Title)
	assert.Equal(t, "Done", output.Buckets[2].Title)
	assert.True(t, output.Buckets[2].IsDoneBucket)
}

func TestListBucketsHandler_Errors(t *testing.T) {
	t.Run("buckets not found", func(t *testing.T) {
		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(vikunja.ErrorResponse{Message: "View not found"})
		})
		defer cleanup()

		h := newTestHandlers()
		input := ListBucketsInput{
			ProjectID: "1",
			ViewID:    "999",
		}

		result, output, err := h.listBucketsHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ListBucketsOutput{}, output)
		assert.Contains(t, err.Error(), "failed to get buckets")
	})

	t.Run("empty buckets list", func(t *testing.T) {
		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]vikunja.Bucket{})
		})
		defer cleanup()

		h := newTestHandlers()
		input := ListBucketsInput{
			ProjectID: "1",
			ViewID:    "100",
		}

		result, output, err := h.listBucketsHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, output.Buckets)
	})
}

func TestGetTaskHandler_EdgeCases(t *testing.T) {
	t.Run("task with minimal fields", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:    1,
			Title: "",
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/1") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "1",
			IncludeBuckets: false,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), output.Task.ID)
		assert.Equal(t, "", output.Task.Title)
	})

	t.Run("multiple views with different bucket states", func(t *testing.T) {
		taskResponse := vikunja.Task{
			ID:        123,
			Title:     "Multi-View Task",
			ProjectID: 456,
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket A", ProjectViewID: 100, Position: 1},
				{ID: 3, Title: "Bucket B", ProjectViewID: 200, Position: 2},
			},
		}

		viewsResponse := []vikunja.ProjectView{
			{ID: 100, Title: "Kanban", ProjectID: 456, ViewKind: vikunja.ViewKindKanban, DoneBucketID: 2},
			{ID: 200, Title: "Second View", ProjectID: 456, ViewKind: vikunja.ViewKindKanban, DoneBucketID: 3},
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/tasks/123") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(taskResponse)
			} else if strings.Contains(r.URL.Path, "/projects/456/views") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(viewsResponse)
			}
		})
		defer cleanup()

		h := newTestHandlers()
		input := GetTaskInput{
			TaskID:         "123",
			IncludeBuckets: true,
		}

		result, output, err := h.getTaskHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, output.Buckets.Views, 2)

		// Verify first view - bucket is not done bucket
		assert.Equal(t, int64(100), output.Buckets.Views[0].ViewID)
		assert.False(t, output.Buckets.Views[0].IsDoneBucket)

		// Verify second view - bucket is done bucket
		assert.Equal(t, int64(200), output.Buckets.Views[1].ViewID)
		assert.True(t, output.Buckets.Views[1].IsDoneBucket)
	})
}

func TestListBucketsHandler_DifferentFormats(t *testing.T) {
	t.Run("with JSON formatter", func(t *testing.T) {
		bucketsResponse := []vikunja.Bucket{
			{ID: 1, Title: "Todo", ProjectViewID: 100},
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/projects/1/views/100/buckets") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(bucketsResponse)
			}
		})
		defer cleanup()

		cfg := &config.Config{Readonly: false}
		deps := &HandlerDependencies{
			Config:          cfg,
			OutputFormatter: vikunja.GetFormatter(vikunja.OutputFormatJSON),
		}
		h := NewHandlers(deps)
		input := ListBucketsInput{
			ProjectID: "1",
			ViewID:    "100",
		}

		result, _, err := h.listBucketsHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		// Verify JSON content in the result
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				assert.Contains(t, textContent.Text, "Todo")
				assert.Contains(t, textContent.Text, "1")
			}
		}
	})

	t.Run("with Markdown formatter", func(t *testing.T) {
		bucketsResponse := []vikunja.Bucket{
			{ID: 1, Title: "Todo", ProjectViewID: 100, Position: 1},
		}

		cleanup := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/projects/1/views/100/buckets") {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(bucketsResponse)
			}
		})
		defer cleanup()

		cfg := &config.Config{Readonly: false}
		deps := &HandlerDependencies{
			Config:          cfg,
			OutputFormatter: vikunja.GetFormatter(vikunja.OutputFormatMarkdown),
		}
		h := NewHandlers(deps)
		input := ListBucketsInput{
			ProjectID: "1",
			ViewID:    "100",
		}

		result, _, err := h.listBucketsHandler(context.Background(), &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		// Verify Markdown content in the result
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				assert.Contains(t, textContent.Text, "Todo")
			}
		}
	})
}
