package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
)

func TestFindViewHandler_MissingEnvironment(t *testing.T) {
	// Test with missing environment variables
	os.Unsetenv("VIKUNJA_HOST")
	os.Unsetenv("VIKUNJA_TOKEN")

	ctx := context.Background()
	input := FindViewInput{
		ProjectTitle: "Test Project",
		ViewName:     "Kanban View",
	}

	result, output, err := findViewHandler(ctx, nil, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	assert.Nil(t, result)
	assert.Equal(t, FindViewOutput{}, output)
}

func TestFindViewHandler_MissingProjectIdentifier(t *testing.T) {
	// Set environment variables
	os.Setenv("VIKUNJA_HOST", "test.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
	}()

	ctx := context.Background()
	input := FindViewInput{
		ViewName: "Kanban View",
	}

	result, output, err := findViewHandler(ctx, nil, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either project_id or project_title must be specified")
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	assert.Equal(t, FindViewOutput{}, output)
}

func TestListViewsHandler_MissingEnvironment(t *testing.T) {
	// Test with missing environment variables
	os.Unsetenv("VIKUNJA_HOST")
	os.Unsetenv("VIKUNJA_TOKEN")

	ctx := context.Background()
	input := ListViewsInput{
		ProjectTitle: "Test Project",
	}

	result, output, err := listViewsHandler(ctx, nil, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	assert.Nil(t, result)
	assert.Equal(t, ListViewsOutput{}, output)
}

func TestListViewsHandler_MissingProjectIdentifier(t *testing.T) {
	// Set environment variables
	os.Setenv("VIKUNJA_HOST", "test.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
	}()

	ctx := context.Background()
	input := ListViewsInput{}

	result, output, err := listViewsHandler(ctx, nil, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either project_id or project_title must be specified")
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	assert.Equal(t, ListViewsOutput{}, output)
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "Kanban View",
			substr:   "Kanban",
			expected: true,
		},
		{
			name:     "case insensitive match",
			s:        "Kanban View",
			substr:   "kanban",
			expected: true,
		},
		{
			name:     "partial match",
			s:        "Kanban View",
			substr:   "View",
			expected: true,
		},
		{
			name:     "no match",
			s:        "Kanban View",
			substr:   "List",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "Kanban View",
			substr:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsIgnoreCase(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToView(t *testing.T) {
	input := vikunja.ProjectView{
		ID:                      1,
		ProjectID:               42,
		Title:                   "Test View",
		ViewKind:                vikunja.ViewKindKanban,
		Position:                1.5,
		BucketConfigurationMode: vikunja.BucketConfigurationModeManual,
		DefaultBucketID:         5,
		DoneBucketID:            6,
	}

	result := toView(input)

	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, int64(42), result.ProjectID)
	assert.Equal(t, "Test View", result.Title)
	assert.Equal(t, vikunja.ViewKindKanban, result.ViewKind)
	assert.Equal(t, 1.5, result.Position)
	assert.Equal(t, vikunja.BucketConfigurationModeManual, result.BucketConfigurationMode)
	assert.Equal(t, int64(5), result.DefaultBucketID)
	assert.Equal(t, int64(6), result.DoneBucketID)
	assert.Equal(t, "vikunja://project/42/view/1", result.URI)
}

func TestToViews(t *testing.T) {
	input := []vikunja.ProjectView{
		{ID: 1, Title: "View 1", ViewKind: vikunja.ViewKindKanban},
		{ID: 2, Title: "View 2", ViewKind: vikunja.ViewKindList},
	}

	result := toViews(input)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "View 1", result[0].Title)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "View 2", result[1].Title)
}

func TestToViewsWithNil(t *testing.T) {
	result := toViews(nil)
	assert.Nil(t, result)
}
