package resolution

import (
	"context"
	"errors"
	"testing"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient is a test double for the Client interface.
type mockClient struct {
	projects         []vikunja.Project
	views            []vikunja.ProjectView
	buckets          []vikunja.Bucket
	getProjectErr    error
	getViewsErr      error
	getBucketsErr    error
	getProjectResult *vikunja.Project
}

func (m *mockClient) GetProjects(ctx context.Context) ([]vikunja.Project, error) {
	return m.projects, m.getProjectErr
}

func (m *mockClient) GetProjectViews(ctx context.Context, projectID int64) ([]vikunja.ProjectView, error) {
	return m.views, m.getViewsErr
}

func (m *mockClient) GetViewBuckets(ctx context.Context, projectID int64, viewID int64) ([]vikunja.Bucket, error) {
	return m.buckets, m.getBucketsErr
}

func (m *mockClient) GetProject(ctx context.Context, projectID int64) (*vikunja.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	if m.getProjectResult != nil {
		return m.getProjectResult, nil
	}
	// If not set, return a generic project
	return &vikunja.Project{ID: projectID, Title: "Project"}, nil
}

func TestResolveProject(t *testing.T) {
	ctx := context.Background()

	t.Run("numeric ID returns project without calling client", func(t *testing.T) {
		client := &mockClient{}
		proj, err := ResolveProject(ctx, client, "123")
		require.NoError(t, err)
		assert.Equal(t, int64(123), proj.ID)
		assert.Contains(t, proj.Title, "Project 123")
	})

	t.Run("title exact match returns project", func(t *testing.T) {
		client := &mockClient{
			projects: []vikunja.Project{
				{ID: 1, Title: "Inbox"},
				{ID: 2, Title: "Work"},
			},
		}
		proj, err := ResolveProject(ctx, client, "Inbox")
		require.NoError(t, err)
		assert.Equal(t, int64(1), proj.ID)
		assert.Equal(t, "Inbox", proj.Title)
	})

	t.Run("title not found returns error with suggestions", func(t *testing.T) {
		client := &mockClient{
			projects: []vikunja.Project{
				{ID: 1, Title: "Inbox"},
				{ID: 2, Title: "Work"},
			},
		}
		_, err := ResolveProject(ctx, client, "Nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project with title \"Nonexistent\" not found")
		assert.Contains(t, err.Error(), "Available projects: [Inbox Work]")
	})

	t.Run("multiple projects with same title returns error", func(t *testing.T) {
		client := &mockClient{
			projects: []vikunja.Project{
				{ID: 1, Title: "Inbox"},
				{ID: 2, Title: "Inbox"},
			},
		}
		_, err := ResolveProject(ctx, client, "Inbox")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple projects found with title \"Inbox\"")
	})

	t.Run("empty identifier returns error", func(t *testing.T) {
		client := &mockClient{}
		_, err := ResolveProject(ctx, client, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project identifier is required")
	})

	t.Run("non-numeric identifier treated as title returns not found", func(t *testing.T) {
		client := &mockClient{}
		_, err := ResolveProject(ctx, client, "12.3")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `project with title "12.3" not found`)
	})
}

func TestFindKanbanView(t *testing.T) {
	ctx := context.Background()

	kanbanView := vikunja.ProjectView{
		ID:        5,
		ProjectID: 1,
		Title:     "Kanban",
		ViewKind:  vikunja.ViewKindKanban,
	}

	t.Run("finds kanban view", func(t *testing.T) {
		client := &mockClient{
			views: []vikunja.ProjectView{
				{ID: 1, ViewKind: vikunja.ViewKindList},
				kanbanView,
			},
		}
		view, err := FindKanbanView(ctx, client, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(5), view.ID)
	})

	t.Run("returns error when no kanban view", func(t *testing.T) {
		client := &mockClient{
			views: []vikunja.ProjectView{
				{ID: 1, ViewKind: vikunja.ViewKindList},
			},
		}
		_, err := FindKanbanView(ctx, client, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kanban view not found in project 1")
	})

	t.Run("propagates GetProjectViews error", func(t *testing.T) {
		client := &mockClient{
			getViewsErr: errors.New("failed to fetch views"),
		}
		_, err := FindKanbanView(ctx, client, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get project views")
	})
}

func TestFindBucketByIDOrTitle(t *testing.T) {
	ctx := context.Background()

	kanbanView := vikunja.ProjectView{
		ID:        5,
		ProjectID: 1,
		Title:     "Kanban",
		ViewKind:  vikunja.ViewKindKanban,
	}

	bucket1 := vikunja.Bucket{ID: 10, Title: "Todo"}
	bucket2 := vikunja.Bucket{ID: 20, Title: "In Progress"}
	bucket3 := vikunja.Bucket{ID: 30, Title: "Done"}

	t.Run("numeric ID returns without validation", func(t *testing.T) {
		client := &mockClient{}
		bucketID, err := FindBucketByIDOrTitle(ctx, client, 1, "15")
		require.NoError(t, err)
		assert.Equal(t, int64(15), *bucketID)
	})

	t.Run("title exact match returns bucket ID", func(t *testing.T) {
		client := &mockClient{
			views:   []vikunja.ProjectView{kanbanView},
			buckets: []vikunja.Bucket{bucket1, bucket2, bucket3},
		}
		bucketID, err := FindBucketByIDOrTitle(ctx, client, 1, "In Progress")
		require.NoError(t, err)
		assert.Equal(t, int64(20), *bucketID)
	})

	t.Run("title not found returns error with suggestions", func(t *testing.T) {
		client := &mockClient{
			views:   []vikunja.ProjectView{kanbanView},
			buckets: []vikunja.Bucket{bucket1, bucket2, bucket3},
			getProjectResult: &vikunja.Project{
				ID:    1,
				Title: "My Project",
			},
		}
		_, err := FindBucketByIDOrTitle(ctx, client, 1, "Unknown")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bucket \"Unknown\" not found")
		assert.Contains(t, err.Error(), "Available buckets in project 'My Project'")
	})

	t.Run("multiple buckets with same title returns error", func(t *testing.T) {
		client := &mockClient{
			views: []vikunja.ProjectView{kanbanView},
			buckets: []vikunja.Bucket{
				{ID: 10, Title: "Todo"},
				{ID: 11, Title: "Todo"},
			},
		}
		_, err := FindBucketByIDOrTitle(ctx, client, 1, "Todo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple buckets found with title \"Todo\"")
	})

	t.Run("non-numeric bucket when no kanban view returns kanban error", func(t *testing.T) {
		client := &mockClient{
			views: []vikunja.ProjectView{
				{ID: 1, ViewKind: vikunja.ViewKindList},
			},
		}
		_, err := FindBucketByIDOrTitle(ctx, client, 1, "Inbox")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kanban view not found in project 1")
	})

	t.Run("get view buckets error propagates", func(t *testing.T) {
		client := &mockClient{
			views:         []vikunja.ProjectView{kanbanView},
			getBucketsErr: errors.New("bucket fetch failed"),
		}
		_, err := FindBucketByIDOrTitle(ctx, client, 1, "Todo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get buckets for kanban view")
	})
}
