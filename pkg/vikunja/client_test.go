package vikunja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test HTTP server with the given handler
func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	// Extract host:port from URL (strip http:// prefix)
	host := strings.TrimPrefix(ts.URL, "http://")
	client, _ := NewClient(host, "test-token", true)
	return ts, client
}

// mockProjectsResponse returns a sample projects response
func mockProjectsResponse() []Project {
	return []Project{
		{ID: 1, Title: "Project 1", Description: "Description 1"},
		{ID: 2, Title: "Project 2", Description: "Description 2"},
	}
}

// mockProjectResponse returns a sample project response
func mockProjectResponse() Project {
	return Project{ID: 1, Title: "Test Project", Description: "Test Description"}
}

// mockTasksResponse returns a sample tasks response
func mockTasksResponse() []Task {
	return []Task{
		{ID: 1, Title: "Task 1", ProjectID: 1},
		{ID: 2, Title: "Task 2", ProjectID: 1},
	}
}

// mockTaskResponse returns a sample task response
func mockTaskResponse() Task {
	return Task{ID: 1, Title: "Test Task", ProjectID: 1}
}

// mockViewsResponse returns a sample views response
func mockViewsResponse() []ProjectView {
	return []ProjectView{
		{ID: 1, Title: "Kanban", ProjectID: 1, ViewKind: ViewKindKanban},
		{ID: 2, Title: "List", ProjectID: 1, ViewKind: ViewKindList},
	}
}

// mockBucketsResponse returns a sample buckets response
func mockBucketsResponse() []Bucket {
	return []Bucket{
		{ID: 1, Title: "Todo", ProjectViewID: 1, IsDoneBucket: false},
		{ID: 2, Title: "Done", ProjectViewID: 1, IsDoneBucket: true},
	}
}

// mockViewTasksResponse returns a sample view tasks response (bucket shape)
func mockViewTasksResponse() []Bucket {
	return []Bucket{
		{
			ID:            1,
			Title:         "Todo",
			ProjectViewID: 1,
			Tasks:         []Task{{ID: 1, Title: "Task 1"}},
		},
	}
}

// mockErrorResponse returns a sample error response
func mockErrorResponse() ErrorResponse {
	return ErrorResponse{Message: "Test error"}
}

// TestNewClient tests the client initialization
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		token    string
		insecure bool
	}{
		{
			name:     "secure connection",
			host:     "vikunja.example.com",
			token:    "test-token",
			insecure: false,
		},
		{
			name:     "insecure connection",
			host:     "localhost:3456",
			token:    "test-token",
			insecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.host, tt.token, tt.insecure)
			require.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, tt.token, client.token)
			assert.NotNil(t, client.httpClient)
		})
	}
}

// TestGetProjects tests the GetProjects method
func TestGetProjects(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedProjects := mockProjectsResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedProjects)
		})
		defer ts.Close()

		projects, err := client.GetProjects(context.Background())
		require.NoError(t, err)
		assert.Len(t, projects, 2)
		assert.Equal(t, "Project 1", projects[0].Title)
	})

	t.Run("empty response", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		})
		defer ts.Close()

		projects, err := client.GetProjects(context.Background())
		require.NoError(t, err)
		assert.Empty(t, projects)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(mockErrorResponse())
		})
		defer ts.Close()

		projects, err := client.GetProjects(context.Background())
		assert.Error(t, err)
		assert.Nil(t, projects)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("server error", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(mockErrorResponse())
		})
		defer ts.Close()

		projects, err := client.GetProjects(context.Background())
		assert.Error(t, err)
		assert.Nil(t, projects)
	})
}

// TestGetProject tests the GetProject method
func TestGetProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedProject := mockProjectResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects/1", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedProject)
		})
		defer ts.Close()

		project, err := client.GetProject(context.Background(), 1)
		require.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, int64(1), project.ID)
	})

	t.Run("not found", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(mockErrorResponse())
		})
		defer ts.Close()

		project, err := client.GetProject(context.Background(), 999)
		assert.Error(t, err)
		assert.Nil(t, project)
	})
}

// TestGetTasks tests the GetTasks method
func TestGetTasks(t *testing.T) {
	t.Run("with project ID", func(t *testing.T) {
		expectedTasks := mockTasksResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects/1/tasks", r.URL.Path)
			// The URL query should contain expand=buckets
			assert.Contains(t, r.URL.RawQuery, "expand=buckets")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTasks)
		})
		defer ts.Close()

		tasks, err := client.GetTasks(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("without project ID", func(t *testing.T) {
		expectedTasks := mockTasksResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/tasks", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTasks)
		})
		defer ts.Close()

		tasks, err := client.GetTasks(context.Background(), 0)
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		})
		defer ts.Close()

		tasks, err := client.GetTasks(context.Background(), 1)
		assert.Error(t, err)
		assert.Nil(t, tasks)
	})
}

// TestGetTask tests the GetTask method
func TestGetTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedTask := mockTaskResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/tasks/1", r.URL.Path)
			assert.Equal(t, "buckets", r.URL.Query().Get("expand"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTask)
		})
		defer ts.Close()

		task, err := client.GetTask(context.Background(), 1)
		require.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "Test Task", task.Title)
	})
}

// TestGetProjectViews tests the GetProjectViews method
func TestGetProjectViews(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedViews := mockViewsResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects/1/views", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedViews)
		})
		defer ts.Close()

		views, err := client.GetProjectViews(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, views, 2)
		assert.Equal(t, "Kanban", views[0].Title)
	})

	t.Run("empty views", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		})
		defer ts.Close()

		views, err := client.GetProjectViews(context.Background(), 1)
		require.NoError(t, err)
		assert.Empty(t, views)
	})
}

// TestGetViewBuckets tests the GetViewBuckets method
func TestGetViewBuckets(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedBuckets := mockBucketsResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects/1/views/1/buckets", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedBuckets)
		})
		defer ts.Close()

		buckets, err := client.GetViewBuckets(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.Len(t, buckets, 2)
		assert.Equal(t, "Todo", buckets[0].Title)
	})
}

// TestGetViewTasks tests the GetViewTasks method
func TestGetViewTasks(t *testing.T) {
	t.Run("bucket shape response", func(t *testing.T) {
		expectedBuckets := mockViewTasksResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects/1/views/1/tasks", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedBuckets)
		})
		defer ts.Close()

		resp, err := client.GetViewTasks(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Buckets, 1)
		assert.Empty(t, resp.Tasks)
	})

	t.Run("flat tasks response", func(t *testing.T) {
		expectedTasks := mockTasksResponse()
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedTasks)
		})
		defer ts.Close()

		resp, err := client.GetViewTasks(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Tasks, 2)
		assert.Empty(t, resp.Buckets)
	})

	t.Run("malformed response", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"invalid": json}`))
		})
		defer ts.Close()

		resp, err := client.GetViewTasks(context.Background(), 1, 1)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("not a slice response", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key": "value"}`))
		})
		defer ts.Close()

		resp, err := client.GetViewTasks(context.Background(), 1, 1)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestMoveTaskToBucket tests the MoveTaskToBucket method
func TestMoveTaskToBucket(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/projects/1/views/2/buckets/3/tasks", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TaskBucket{TaskID: 10, BucketID: 3})
		})
		defer ts.Close()

		result, err := client.MoveTaskToBucket(context.Background(), 1, 2, 3, 10)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(10), result.TaskID)
	})

	t.Run("invalid request body", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(mockErrorResponse())
		})
		defer ts.Close()

		result, err := client.MoveTaskToBucket(context.Background(), 1, 2, 3, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestHandleErrorResponse tests the handleErrorResponse method
func TestHandleErrorResponse(t *testing.T) {
	t.Run("valid error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()

		client, _ := NewClient(ts.URL, "token", true)

		// Create a mock response with error body
		w := httptest.NewRecorder()
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Test error message"})
		resp := w.Result()
		resp.StatusCode = http.StatusBadRequest

		err := client.handleErrorResponse(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Test error message")
	})

	t.Run("invalid error response body", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()

		client, _ := NewClient(ts.URL, "token", true)

		w := httptest.NewRecorder()
		w.WriteString("invalid json")
		resp := w.Result()
		resp.StatusCode = http.StatusInternalServerError

		err := client.handleErrorResponse(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode error response")
	})
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	t.Run("network error", func(t *testing.T) {
		// Create a client pointing to a non-existent server
		client, _ := NewClient("invalid-host:99999", "token", true)

		_, err := client.GetProjects(context.Background())
		assert.Error(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		// This would require context with timeout
		// For now, just verify it doesn't panic
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockProjectsResponse())
		})
		defer ts.Close()

		_, err := client.GetProjects(context.Background())
		require.NoError(t, err)
	})
}

// TestGetSingleResource tests the generic getSingleResource function
func TestGetSingleResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockProjectResponse())
		})
		defer ts.Close()

		result, err := getSingleResource[Project](context.Background(), client, ts.URL+"/test", "resource")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Project", result.Title)
	})

	t.Run("http error", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(mockErrorResponse())
		})
		defer ts.Close()

		result, err := getSingleResource[Project](context.Background(), client, ts.URL+"/test", "resource")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid"))
		})
		defer ts.Close()

		result, err := getSingleResource[Project](context.Background(), client, ts.URL+"/test", "resource")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestHTTPHelpers tests various HTTP helper scenarios
func TestHTTPHelpers(t *testing.T) {
	t.Run("request headers are set correctly", func(t *testing.T) {
		var receivedHeaders http.Header
		ts, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockProjectsResponse())
		})
		defer ts.Close()

		_, err := client.GetProjects(context.Background())
		require.NoError(t, err)

		assert.Equal(t, "Bearer test-token", receivedHeaders.Get("Authorization"))
		assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
	})
}

// BenchmarkGetProjects benchmarks the GetProjects method
func BenchmarkGetProjects(b *testing.B) {
	expectedProjects := mockProjectsResponse()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedProjects)
	}))
	defer ts.Close()

	client, _ := NewClient(ts.URL, "test-token", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetProjects(context.Background())
	}
}
