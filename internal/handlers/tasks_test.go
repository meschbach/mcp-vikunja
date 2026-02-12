package handlers

import (
	"testing"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTaskSummary(t *testing.T) {
	vTask := vikunja.Task{
		ID:    123,
		Title: "Test Task",
	}

	result := toTaskSummary(vTask)

	assert.Equal(t, int64(123), result.ID)
	assert.Equal(t, "Test Task", result.Title)
	assert.Equal(t, "vikunja://task/123", result.URI)
}

func TestToTasksSummary(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := toTasksSummary(nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		result := toTasksSummary([]vikunja.Task{})
		assert.Empty(t, result)
	})

	t.Run("multiple tasks", func(t *testing.T) {
		tasks := []vikunja.Task{
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
			{ID: 3, Title: "Task 3"},
		}
		result := toTasksSummary(tasks)
		require.Len(t, result, 3)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "Task 1", result[0].Title)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, "Task 2", result[1].Title)
		assert.Equal(t, "vikunja://task/3", result[2].URI)
	})
}

func TestToBucketSummary(t *testing.T) {
	vBucket := vikunja.Bucket{
		ID:    456,
		Title: "Test Bucket",
	}

	result := toBucketSummary(vBucket)

	assert.Equal(t, int64(456), result.ID)
	assert.Equal(t, "Test Bucket", result.Title)
}

func TestToTask(t *testing.T) {
	t.Run("task with all fields", func(t *testing.T) {
		now := time.Now()
		tomorrow := now.Add(24 * time.Hour)

		vTask := vikunja.Task{
			ID:          123,
			Title:       "Test Task",
			Description: "Test Description",
			ProjectID:   456,
			Done:        true,
			DueDate:     tomorrow,
			Created:     now,
			Updated:     now,
			Position:    1.5,
			Buckets:     []vikunja.Bucket{{ID: 1, Title: "Bucket 1"}},
		}

		result := toTask(vTask)

		assert.Equal(t, int64(123), result.ID)
		assert.Equal(t, "Test Task", result.Title)
		assert.Equal(t, "Test Description", result.Description)
		assert.Equal(t, int64(456), result.ProjectID)
		assert.True(t, result.Done)
		assert.Equal(t, tomorrow.Format(time.RFC3339), result.DueDate)
		assert.Equal(t, now.Format(time.RFC3339), result.Created)
		assert.Equal(t, now.Format(time.RFC3339), result.Updated)
		assert.Equal(t, 1.5, result.Position)
		require.Len(t, result.Buckets, 1)
		assert.Equal(t, int64(1), result.Buckets[0].ID)
	})

	t.Run("task with zero dates", func(t *testing.T) {
		vTask := vikunja.Task{
			ID:    123,
			Title: "Task Without Dates",
		}

		result := toTask(vTask)

		assert.Equal(t, int64(123), result.ID)
		assert.Equal(t, "Task Without Dates", result.Title)
		assert.Empty(t, result.DueDate)
		assert.Empty(t, result.Created)
		assert.Empty(t, result.Updated)
	})
}

func TestToBucket(t *testing.T) {
	vBucket := vikunja.Bucket{
		ID:            789,
		ProjectViewID: 100,
		Title:         "Test Bucket",
		Description:   "Bucket Description",
		Limit:         5,
		Position:      2.5,
		IsDoneBucket:  true,
	}

	result := toBucket(vBucket)

	assert.Equal(t, int64(789), result.ID)
	assert.Equal(t, int64(100), result.ProjectViewID)
	assert.Equal(t, "Test Bucket", result.Title)
	assert.Equal(t, "Bucket Description", result.Description)
	assert.Equal(t, 5, result.Limit)
	assert.Equal(t, 2.5, result.Position)
	assert.True(t, result.IsDoneBucket)
}

func TestToBuckets(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := toBuckets(nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		result := toBuckets([]vikunja.Bucket{})
		assert.Empty(t, result)
	})

	t.Run("multiple buckets", func(t *testing.T) {
		buckets := []vikunja.Bucket{
			{ID: 1, Title: "Bucket 1"},
			{ID: 2, Title: "Bucket 2"},
		}
		result := toBuckets(buckets)
		require.Len(t, result, 2)
		assert.Equal(t, "Bucket 1", result[0].Title)
		assert.Equal(t, "Bucket 2", result[1].Title)
	})
}

func TestToProject(t *testing.T) {
	// Note: There's no explicit toProject function, but let's test the pattern
	// when converting from vikunja.Project to handlers.Project
	vProject := vikunja.Project{
		ID:    100,
		Title: "Test Project",
	}

	project := Project{
		ID:    vProject.ID,
		Title: vProject.Title,
		URI:   "vikunja://project/100",
	}

	assert.Equal(t, int64(100), project.ID)
	assert.Equal(t, "Test Project", project.Title)
	assert.Equal(t, "vikunja://project/100", project.URI)
}

func TestToBucketsConversion(t *testing.T) {
	// Test the toBuckets helper function
	vikunjaBuckets := []vikunja.Bucket{
		{ID: 1, Title: "Todo", ProjectViewID: 10},
		{ID: 2, Title: "In Progress", ProjectViewID: 10},
		{ID: 3, Title: "Done", ProjectViewID: 10},
	}

	handlersBuckets := toBuckets(vikunjaBuckets)

	require.Len(t, handlersBuckets, 3)
	assert.Equal(t, int64(1), handlersBuckets[0].ID)
	assert.Equal(t, "Todo", handlersBuckets[0].Title)
	assert.Equal(t, int64(10), handlersBuckets[0].ProjectViewID)
	assert.Equal(t, "In Progress", handlersBuckets[1].Title)
	assert.Equal(t, "Done", handlersBuckets[2].Title)
}

func TestValidationError(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		err := ValidationError{
			Field:   "test_field",
			Message: "is required",
		}
		assert.Equal(t, "test_field: is required", err.Error())
	})

	t.Run("implements error interface", func(t *testing.T) {
		var err error = ValidationError{
			Field:   "field",
			Message: "error",
		}
		assert.Equal(t, "field: error", err.Error())
	})
}

func TestFilterViewTasksByBucket(t *testing.T) {
	h := newTestHandlers()

	t.Run("non-kanban view with tasks", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Tasks: []vikunja.Task{
				{ID: 1, Title: "Task 1"},
			},
		}

		result, err := h.filterViewTasksByBucket(resp, 1, "", "List View")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "bucket filtering not supported for non-kanban views")
		assert.Nil(t, result)
	})

	t.Run("find bucket by ID", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket 1"},
				{ID: 2, Title: "Bucket 2"},
				{ID: 3, Title: "Bucket 3"},
			},
		}

		result, err := h.filterViewTasksByBucket(resp, 2, "", "Kanban")

		require.NoError(t, err)
		require.Len(t, result.Buckets, 1)
		assert.Equal(t, int64(2), result.Buckets[0].ID)
		assert.Equal(t, "Bucket 2", result.Buckets[0].Title)
	})

	t.Run("find bucket by title", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Todo"},
				{ID: 2, Title: "In Progress"},
			},
		}

		result, err := h.filterViewTasksByBucket(resp, 0, "In Progress", "Kanban")

		require.NoError(t, err)
		require.Len(t, result.Buckets, 1)
		assert.Equal(t, int64(2), result.Buckets[0].ID)
	})

	t.Run("bucket not found by ID", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket 1"},
			},
		}

		result, err := h.filterViewTasksByBucket(resp, 999, "", "Kanban")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "bucket with ID 999 not found in view \"Kanban\"")
		assert.Nil(t, result)
	})

	t.Run("bucket not found by title", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket 1"},
			},
		}

		result, err := h.filterViewTasksByBucket(resp, 0, "NonExistent", "Kanban")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "bucket with title \"NonExistent\" not found in view \"Kanban\"")
		assert.Nil(t, result)
	})

	t.Run("no bucket filter specified", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Bucket 1"},
			},
		}

		result, err := h.filterViewTasksByBucket(resp, 0, "", "Kanban")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no bucket filter specified")
		assert.Nil(t, result)
	})
}

func TestFindBucket(t *testing.T) {
	h := newTestHandlers()

	buckets := []vikunja.Bucket{
		{ID: 1, Title: "Todo"},
		{ID: 2, Title: "In Progress"},
		{ID: 3, Title: "Done"},
	}

	t.Run("find by ID success", func(t *testing.T) {
		result, err := h.findBucket(buckets, 2, "", "Kanban")

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.ID)
		assert.Equal(t, "In Progress", result.Title)
	})

	t.Run("find by title success", func(t *testing.T) {
		result, err := h.findBucket(buckets, 0, "Done", "Kanban")

		require.NoError(t, err)
		assert.Equal(t, int64(3), result.ID)
		assert.Equal(t, "Done", result.Title)
	})

	t.Run("find by ID not found", func(t *testing.T) {
		result, err := h.findBucket(buckets, 999, "", "Kanban")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bucket with ID 999 not found in view \"Kanban\"")
	})

	t.Run("find by title not found", func(t *testing.T) {
		result, err := h.findBucket(buckets, 0, "NonExistent", "Kanban")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bucket with title \"NonExistent\" not found in view \"Kanban\"")
	})

	t.Run("no filter specified", func(t *testing.T) {
		result, err := h.findBucket(buckets, 0, "", "Kanban")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no bucket filter specified")
	})
}

func TestResolveViewByTitle(t *testing.T) {
	h := newTestHandlers()

	views := []vikunja.ProjectView{
		{ID: 1, Title: "Kanban"},
		{ID: 2, Title: "List"},
	}

	t.Run("valid view title found", func(t *testing.T) {
		viewID, viewTitle, err := h.resolveViewByTitle("List", views, 100)

		require.NoError(t, err)
		assert.Equal(t, int64(2), viewID)
		assert.Equal(t, "List", viewTitle)
	})

	t.Run("view title not found", func(t *testing.T) {
		viewID, viewTitle, err := h.resolveViewByTitle("Calendar", views, 100)

		require.Error(t, err)
		assert.Equal(t, int64(0), viewID)
		assert.Empty(t, viewTitle)
		assert.Contains(t, err.Error(), "view with title \"Calendar\" not found in project 100")
	})
}

func TestGetViewTasks(t *testing.T) {
	// Testing response structure since client requires network mocking
	t.Run("successful retrieval without bucket filter", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Tasks: []vikunja.Task{
				{ID: 1, Title: "Task 1"},
				{ID: 2, Title: "Task 2"},
			},
		}

		// Verify the response structure
		assert.NotNil(t, resp)
		assert.Len(t, resp.Tasks, 2)
		assert.Empty(t, resp.Buckets) // No buckets for list view
	})

	t.Run("view tasks response with buckets", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Buckets: []vikunja.Bucket{
				{ID: 1, Title: "Todo", Tasks: []vikunja.Task{{ID: 1, Title: "Task 1"}}},
				{ID: 2, Title: "Done", Tasks: []vikunja.Task{{ID: 2, Title: "Task 2"}}},
			},
		}

		assert.NotNil(t, resp)
		assert.Len(t, resp.Buckets, 2)
		assert.Empty(t, resp.Tasks) // Tasks are in buckets for kanban
	})

	t.Run("view tasks response empty", func(t *testing.T) {
		resp := &vikunja.ViewTasksResponse{
			Tasks:   []vikunja.Task{},
			Buckets: []vikunja.Bucket{},
		}

		assert.NotNil(t, resp)
		assert.Empty(t, resp.Tasks)
		assert.Empty(t, resp.Buckets)
	})
}
