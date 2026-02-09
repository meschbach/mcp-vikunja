package vikunja

import (
	"bytes"
	"testing"
	"time"

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

func TestFormatter_FormatTasksAsMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	tests := []struct {
		name     string
		tasks    []Task
		expected string
	}{
		{
			name:     "empty tasks",
			tasks:    []Task{},
			expected: "## No tasks found\n",
		},
		{
			name: "single task",
			tasks: []Task{
				{
					ID:        1,
					Title:     "Test Task",
					Done:      false,
					ProjectID: 100,
				},
			},
			expected: `## Tasks (1)

| ID | Title | Done | Due Date | Project |
|---|---|---|---|---|
| 1 | Test Task | ❌ | - | [100](vikunja://projects/100) |

<details>
<summary>Task Details</summary>

### Test Task

- **ID**: 1
- **URI**: [vikunja://tasks/1](vikunja://tasks/1)
- **Project**: [100](vikunja://projects/100)
- **Status**: ❌ Pending

---

</details>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatTasksAsMarkdown(tt.tasks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_FormatProjectsAsMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	tests := []struct {
		name     string
		projects []Project
		expected string
	}{
		{
			name:     "empty projects",
			projects: []Project{},
			expected: "# No projects found\n",
		},
		{
			name: "single project",
			projects: []Project{
				{
					ID:          1,
					Title:       "Test Project",
					Description: "A test project description",
					Identifier:  "test-proj",
				},
			},
			expected: `# Projects (1)

## 📁 Test Project

- **ID**: 1
- **URI**: [vikunja://projects/1](vikunja://projects/1)
- **Identifier**: ` + "`" + `test-proj` + "`" + `

**Description**:
A test project description

---

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatProjectsAsMarkdown(tt.projects)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_FormatBucketsAsMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	tests := []struct {
		name     string
		buckets  []Bucket
		expected string
	}{
		{
			name:     "empty buckets",
			buckets:  []Bucket{},
			expected: "## No buckets found\n",
		},
		{
			name: "buckets with tasks",
			buckets: []Bucket{
				{
					ID:           1,
					Title:        "To Do",
					Limit:        5,
					IsDoneBucket: false,
					Tasks: []Task{
						{ID: 1, Title: "Task 1"},
						{ID: 2, Title: "Task 2"},
					},
				},
				{
					ID:           2,
					Title:        "Done",
					Limit:        0,
					IsDoneBucket: true,
					Tasks:        []Task{},
				},
			},
			expected: `## 📋 Buckets (2)

| 📁 Bucket | ID | Tasks | Limit | Done |
|---|---|---|---|---|
| To Do | 1 | 2 | 5 | ❌ |
| Done | 2 | 0 | - | ✅ |
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatBucketsAsMarkdown(tt.buckets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_FormatViewTasksSummaryAsMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	tests := []struct {
		name      string
		viewTasks ViewTasksSummary
		expected  string
	}{
		{
			name: "empty view",
			viewTasks: ViewTasksSummary{
				ViewID:    1,
				ViewTitle: "Test View",
				Buckets:   []BucketTasksSummary{},
			},
			expected: "# 📋 Test View (ID: 1)\n\n",
		},
		{
			name: "view with buckets and tasks",
			viewTasks: ViewTasksSummary{
				ViewID:    1,
				ViewTitle: "Kanban Board",
				Buckets: []BucketTasksSummary{
					{
						Bucket: BucketSummary{ID: 1, Title: "To Do"},
						Tasks: []TaskSummary{
							{ID: 1, Title: "Task 1"},
							{ID: 2, Title: "Task 2"},
						},
					},
					{
						Bucket: BucketSummary{ID: 2, Title: "Done"},
						Tasks:  []TaskSummary{},
					},
				},
			},
			expected: `# 📋 Kanban Board (ID: 1)

## 📁 To Do (ID: 1)

- [1] Task 1
- [2] Task 2

## 📁 Done (ID: 2)

(no tasks)

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatViewTasksSummaryAsMarkdown(&tt.viewTasks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_FormatTaskWithBucketsMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	tests := []struct {
		name       string
		task       Task
		bucketInfo *TaskBucketInfo
		expected   string
	}{
		{
			name: "task without buckets",
			task: Task{
				ID:          1,
				Title:       "Test Task",
				Description: "Test description",
				Done:        false,
				ProjectID:   42,
			},
			bucketInfo: nil,
			expected:   "# Test Task\n\n- **ID**: 1\n- **URI**: [vikunja://tasks/1](vikunja://tasks/1)\n- **Project**: [42](vikunja://projects/42)\n- **Status**: ❌ Pending\n\n**Description**:\nTest description\n",
		},
		{
			name: "task with buckets",
			task: Task{
				ID:        2,
				Title:     "Task with Buckets",
				Done:      true,
				ProjectID: 10,
				Created:   time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				DueDate:   time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC),
			},
			bucketInfo: &TaskBucketInfo{
				Views: []TaskViewInfo{
					{
						ViewID:       1,
						ViewTitle:    "Kanban Board",
						ViewKind:     ViewKindKanban,
						BucketID:     func() *int64 { i := int64(5); return &i }(),
						BucketTitle:  func() *string { s := "In Progress"; return &s }(),
						IsDoneBucket: false,
					},
				},
			},
			expected: "# Task with Buckets\n\n- **ID**: 2\n- **URI**: [vikunja://tasks/2](vikunja://tasks/2)\n- **Project**: [10](vikunja://projects/10)\n- **Created**: 2023-01-01T00:00:00Z\n- **Due Date**: 2023-12-31\n- **Status**: ✅ Completed\n\n**Bucket Information**:\n- Kanban Board (kanban): In Progress\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatTaskWithBucketsMarkdown(tt.task, tt.bucketInfo)
			if tt.name == "task with buckets" {
				t.Logf("Actual result:\n%s", result)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_FormatProjectAndViewMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	project := Project{
		ID:          1,
		Title:       "Test Project",
		Description: "A test project for markdown formatting",
		Identifier:  "TEST-PROJ",
		OwnerID:     42,
		Created:     time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
		Updated:     time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC),
	}

	view := ProjectView{
		ID:                      1,
		ProjectID:               1,
		Title:                   "Kanban Board",
		ViewKind:                ViewKindKanban,
		Position:                1.0,
		BucketConfigurationMode: BucketConfigurationModeManual,
		DefaultBucketID:         1,
		DoneBucketID:            2,
	}

	result := formatter.FormatProjectAndViewMarkdown(project, view)
	t.Logf("Actual project and view result:\n%s", result)

	assert.Contains(t, result, "# 📁 Test Project")
	assert.Contains(t, result, "- **ID**: 1")
	assert.Contains(t, result, "- **Identifier**: `TEST-PROJ`")
	assert.Contains(t, result, "- **Owner ID**: 42")
	assert.Contains(t, result, "- **Created**: 2023-01-01T00:00:00Z")
	assert.Contains(t, result, "- **Updated**: 2023-12-31T00:00:00Z")
	assert.Contains(t, result, "**Description**:\nA test project for markdown formatting")
	assert.Contains(t, result, "Kanban Board")
	assert.Contains(t, result, "- **ID**: 1")
	assert.Contains(t, result, "- **Type**: kanban")
	assert.Contains(t, result, "- **Position**: 1.00")
	assert.Contains(t, result, "- **Position**: 1.0")
	assert.Contains(t, result, "- **Default Bucket**: 1")
	assert.Contains(t, result, "- **Done Bucket**: 2")
}

func TestFormatter_FormatProjectAndViewListMarkdown(t *testing.T) {
	formatter := NewFormatter(false, nil)

	project := Project{
		ID:          1,
		Title:       "Multi-View Project",
		Description: "A project with multiple views",
	}

	views := []ProjectView{
		{
			ID:       1,
			Title:    "Kanban",
			ViewKind: ViewKindKanban,
			Position: 1.0,
		},
		{
			ID:       2,
			Title:    "List View",
			ViewKind: ViewKindList,
			Position: 2.0,
		},
		{
			ID:       3,
			Title:    "Gantt Chart",
			ViewKind: ViewKindGantt,
			Position: 3.0,
		},
	}

	result := formatter.FormatProjectAndViewListMarkdown(project, views)

	assert.Contains(t, result, "# 📁 Multi-View Project")
	assert.Contains(t, result, "## Views (3)")
	assert.Contains(t, result, "| 📋 View | ID | Type | Position |")
	assert.Contains(t, result, "| 📋 Kanban | 1 | kanban | 1.00 |")
	assert.Contains(t, result, "| 📝 List View | 2 | list | 2.00 |")
	assert.Contains(t, result, "| 📊 Gantt Chart | 3 | gantt | 3.00 |")
}

func TestMarkdownFormatter_AllNewTypes(t *testing.T) {
	formatter := NewMarkdownFormatter()

	// Test TaskOutput
	taskOutput := TaskOutput{
		Task: Task{
			ID:    1,
			Title: "Test Task",
			Done:  false,
		},
		Buckets: &TaskBucketInfo{
			Views: []TaskViewInfo{
				{
					ViewTitle: "Test View",
					ViewKind:  ViewKindKanban,
				},
			},
		},
	}

	result, err := formatter.Format(taskOutput)
	assert.NoError(t, err)
	assert.Contains(t, result, "# Test Task")
	assert.Contains(t, result, "- **ID**: 1")
	assert.Contains(t, result, "**Bucket Information**:")
	assert.NotContains(t, result, "<!-- Unsupported type for markdown")
}

func TestMarkdownFormatter_ViewTasksSummaryIntegration(t *testing.T) {
	// Test that the MarkdownFormatter can handle ViewTasksSummary correctly
	formatter := NewMarkdownFormatter()

	viewTasks := ViewTasksSummary{
		ViewID:    1,
		ViewTitle: "Test Board",
		Buckets: []BucketTasksSummary{
			{
				Bucket: BucketSummary{ID: 1, Title: "To Do"},
				Tasks: []TaskSummary{
					{ID: 1, Title: "Test Task 1"},
					{ID: 2, Title: "Test Task 2"},
				},
			},
		},
	}

	result, err := formatter.Format(viewTasks)
	assert.NoError(t, err)
	assert.Contains(t, result, "# 📋 Test Board (ID: 1)")
	assert.Contains(t, result, "## 📁 To Do (ID: 1)")
	assert.Contains(t, result, "[1] Test Task 1")
	assert.Contains(t, result, "[2] Test Task 2")
	assert.NotContains(t, result, "<!-- Unsupported type for markdown")
}
