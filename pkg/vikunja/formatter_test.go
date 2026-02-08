package vikunja

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_FormatProjects(t *testing.T) {
	projects := []Project{
		{ID: 1, Title: "Project A", Description: "Test project A"},
		{ID: 2, Title: "Project B", Description: "Test project B"},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatProjects(projects)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Project A")
	assert.Contains(t, output, "Project B")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "vikunja://projects/1")
	assert.Contains(t, output, "vikunja://projects/2")
}

func TestFormatter_FormatProject(t *testing.T) {
	project := &Project{
		ID:          1,
		Title:       "Test Project",
		Description: "A test project description",
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatProject(project)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Test Project")
	assert.Contains(t, output, "ID: 1")
	assert.Contains(t, output, "vikunja://projects/1")
	assert.Contains(t, output, "A test project description")
}

func TestFormatter_FormatTasks(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Task A", Description: "Test task A", ProjectID: 1},
		{ID: 2, Title: "Task B", Description: "Test task B", ProjectID: 1},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatTasks(tasks)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task A")
	assert.Contains(t, output, "Task B")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "vikunja://tasks/1")
	assert.Contains(t, output, "vikunja://tasks/2")
}

func TestFormatter_FormatTask(t *testing.T) {
	task := &Task{
		ID:          1,
		Title:       "Test Task",
		Description: "A test task description",
		ProjectID:   1,
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatTask(task)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Test Task")
	assert.Contains(t, output, "ID: 1")
	assert.Contains(t, output, "vikunja://tasks/1")
	assert.Contains(t, output, "Project ID: 1")
	assert.Contains(t, output, "A test task description")
}

func TestFormatter_FormatProjectsAsJSON(t *testing.T) {
	projects := []Project{
		{ID: 1, Title: "Project A", Description: "Test project A"},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatProjectsAsJSON(projects)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"id": 1`)
	assert.Contains(t, output, `"title": "Project A"`)
}

func TestFormatter_FormatTasksAsJSON(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Task A", Description: "Test task A", ProjectID: 1},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatTasksAsJSON(tasks)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"id": 1`)
	assert.Contains(t, output, `"title": "Task A"`)
	assert.Contains(t, output, `"project_id": 1`)
}

func TestFormatter_EmptyProjects(t *testing.T) {
	projects := []Project{}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatProjects(projects)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "URI")
}

func TestFormatter_EmptyTasks(t *testing.T) {
	tasks := []Task{}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatTasks(tasks)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "TITLE")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "URI")
}

func TestFormatter_FormatBuckets(t *testing.T) {
	buckets := []Bucket{
		{ID: 1, Title: "To Do", Position: 1.5, IsDoneBucket: false},
		{ID: 2, Title: "Done", Position: 2.0, IsDoneBucket: true},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatBuckets(buckets)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "TITLE")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "DONE")
	assert.Contains(t, output, "To Do")
	assert.Contains(t, output, "Done")
	assert.Contains(t, output, "1.50")
	assert.Contains(t, output, "2.00")
	assert.Contains(t, output, "No")
	assert.Contains(t, output, "Yes")
}

func TestFormatter_FormatProjectViews(t *testing.T) {
	views := []ProjectView{
		{ID: 1, Title: "List View", ViewKind: ViewKindList, Position: 1.1},
		{ID: 2, Title: "Kanban View", ViewKind: ViewKindKanban, Position: 2.2},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatProjectViews(views)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "TITLE")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "KIND")
	assert.Contains(t, output, "List View")
	assert.Contains(t, output, "Kanban View")
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "kanban")
	assert.Contains(t, output, "1.10")
	assert.Contains(t, output, "2.20")
}

func TestFormatter_FormatTaskWithBuckets(t *testing.T) {
	task := &Task{
		ID:    1,
		Title: "Test Task",
	}

	bucketTitle1 := "To Do"
	bucketTitle2 := "Done"

	bucketInfo := &TaskBucketInfo{
		TaskID: 1,
		Views: []TaskViewInfo{
			{
				ViewID:       1,
				ViewTitle:    "Kanban 1",
				ViewKind:     ViewKindKanban,
				BucketID:     func(i int64) *int64 { return &i }(10),
				BucketTitle:  &bucketTitle1,
				IsDoneBucket: false,
			},
			{
				ViewID:       2,
				ViewTitle:    "Kanban 2",
				ViewKind:     ViewKindKanban,
				BucketID:     func(i int64) *int64 { return &i }(20),
				BucketTitle:  &bucketTitle2,
				IsDoneBucket: true,
			},
		},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(false, buf)

	err := formatter.FormatTaskWithBuckets(task, bucketInfo)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Test Task")
	assert.Contains(t, output, "Kanban 1 (kanban): To Do")
	assert.Contains(t, output, "Kanban 2 (kanban): Done [DONE]")
}

func TestFormatter_FormatViewTasks(t *testing.T) {
	formatter := NewFormatter(false, nil)
	vt := &ViewTasks{
		ViewID:    1,
		ViewTitle: "Test View",
		Buckets: []BucketTasks{
			{
				Bucket: Bucket{ID: 10, Title: "Bucket 1", IsDoneBucket: false},
				Tasks: []Task{
					{ID: 100, Title: "Task 1", Done: false},
					{ID: 101, Title: "Task 2", Done: true},
				},
			},
			{
				Bucket: Bucket{ID: 20, Title: "Bucket 2", IsDoneBucket: true},
				Tasks:  []Task{},
			},
		},
	}

	output, err := formatter.CaptureOutput(func() error {
		return formatter.FormatViewTasks(vt)
	})

	assert.NoError(t, err)

	expectedParts := []string{
		"Test View (ID: 1)",
		"Bucket 1 (ID: 10)",
		"- [100] Task 1",
		"- [101] Task 2",
		"Bucket 2 (ID: 20) [DONE]",
		"(no tasks)",
	}

	for _, part := range expectedParts {
		assert.Contains(t, output, part)
	}
}
