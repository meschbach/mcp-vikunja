// Package vikunja provides a client for the Vikunja API.
package vikunja

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()
	assert.NotNil(t, formatter)
}

func TestJSONFormatter_Format(t *testing.T) {
	formatter := NewJSONFormatter()

	t.Run("format simple struct", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		result, err := formatter.Format(data)

		require.NoError(t, err)
		assert.Contains(t, result, "key")
		assert.Contains(t, result, "value")
		assert.Contains(t, result, "{")
		assert.Contains(t, result, "}")
	})

	t.Run("format nested struct", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
			"nested": map[string]string{
				"key": "value",
			},
		}
		result, err := formatter.Format(data)

		require.NoError(t, err)
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "nested")
	})

	t.Run("format empty struct", func(t *testing.T) {
		data := struct{}{}
		result, err := formatter.Format(data)

		require.NoError(t, err)
		assert.Contains(t, result, "{}")
	})

	t.Run("format slice", func(t *testing.T) {
		data := []string{"item1", "item2"}
		result, err := formatter.Format(data)

		require.NoError(t, err)
		assert.Contains(t, result, "item1")
		assert.Contains(t, result, "item2")
		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
	})

	t.Run("format with invalid data", func(t *testing.T) {
		// Channel cannot be marshaled to JSON
		data := make(chan int)
		result, err := formatter.Format(data)

		require.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to marshal JSON")
	})
}

func TestNewMarkdownFormatter(t *testing.T) {
	formatter := NewMarkdownFormatter()
	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.formatter)
}

func TestMarkdownFormatter_Format_Tasks(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format single task", func(t *testing.T) {
		task := Task{
			ID:    1,
			Title: "Test Task",
		}
		result, err := formatter.Format(task)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Test Task")
	})

	t.Run("format task slice", func(t *testing.T) {
		tasks := []Task{
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
		}
		result, err := formatter.Format(tasks)

		require.NoError(t, err)
		assert.Contains(t, result, "Task 1")
		assert.Contains(t, result, "Task 2")
	})
}

func TestMarkdownFormatter_Format_Projects(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format single project", func(t *testing.T) {
		project := Project{
			ID:    1,
			Title: "Test Project",
		}
		result, err := formatter.Format(project)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Test Project")
	})

	t.Run("format project slice", func(t *testing.T) {
		projects := []Project{
			{ID: 1, Title: "Project 1"},
			{ID: 2, Title: "Project 2"},
		}
		result, err := formatter.Format(projects)

		require.NoError(t, err)
		assert.Contains(t, result, "Project 1")
		assert.Contains(t, result, "Project 2")
	})
}

func TestMarkdownFormatter_Format_Buckets(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format single bucket", func(t *testing.T) {
		bucket := Bucket{
			ID:    1,
			Title: "Test Bucket",
		}
		result, err := formatter.Format(bucket)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Test Bucket")
	})

	t.Run("format bucket slice", func(t *testing.T) {
		buckets := []Bucket{
			{ID: 1, Title: "Bucket 1"},
			{ID: 2, Title: "Bucket 2"},
		}
		result, err := formatter.Format(buckets)

		require.NoError(t, err)
		assert.Contains(t, result, "Bucket 1")
		assert.Contains(t, result, "Bucket 2")
	})
}

func TestMarkdownFormatter_Format_Views(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format single view", func(t *testing.T) {
		view := ProjectView{
			ID:    1,
			Title: "Kanban",
		}
		result, err := formatter.Format(view)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Kanban")
	})

	t.Run("format view slice", func(t *testing.T) {
		views := []ProjectView{
			{ID: 1, Title: "Kanban"},
			{ID: 2, Title: "List"},
		}
		result, err := formatter.Format(views)

		require.NoError(t, err)
		assert.Contains(t, result, "Kanban")
		assert.Contains(t, result, "List")
	})
}

func TestMarkdownFormatter_Format_ViewTasks(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format ViewTasks", func(t *testing.T) {
		viewTasks := &ViewTasks{
			ViewID:    1,
			ViewTitle: "Kanban",
		}
		result, err := formatter.Format(viewTasks)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("format ViewTasksSummary", func(t *testing.T) {
		summary := ViewTasksSummary{
			ViewID:    1,
			ViewTitle: "Kanban",
		}
		result, err := formatter.Format(summary)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Kanban")
	})

	t.Run("format ViewTasksSummary pointer", func(t *testing.T) {
		summary := &ViewTasksSummary{
			ViewID:    1,
			ViewTitle: "Kanban",
		}
		result, err := formatter.Format(summary)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Kanban")
	})
}

func TestMarkdownFormatter_Format_TaskOutput(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format TaskOutput", func(t *testing.T) {
		output := TaskOutput{
			Task: Task{
				ID:    1,
				Title: "Test Task",
			},
		}
		result, err := formatter.Format(output)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Test Task")
	})
}

func TestMarkdownFormatter_Format_ViewOutput(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format ViewOutput", func(t *testing.T) {
		output := ViewOutput{
			Project: Project{ID: 1, Title: "Project"},
			View:    ProjectView{ID: 1, Title: "Kanban"},
		}
		result, err := formatter.Format(output)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestMarkdownFormatter_Format_ViewsOutput(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format ViewsOutput", func(t *testing.T) {
		output := ViewsOutput{
			Project: Project{ID: 1, Title: "Project"},
			Views: []ProjectView{
				{ID: 1, Title: "Kanban"},
				{ID: 2, Title: "List"},
			},
		}
		result, err := formatter.Format(output)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Kanban")
		assert.Contains(t, result, "List")
	})
}

func TestMarkdownFormatter_isHandlersProject(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("is handlers project", func(t *testing.T) {
		// This type has ID, Title, and URI fields
		type HandlersProject struct {
			ID    int64
			Title string
			URI   string
		}
		project := HandlersProject{ID: 1, Title: "Test", URI: "test://1"}
		assert.True(t, formatter.isHandlersProject(project))
	})

	t.Run("is not handlers project - missing URI", func(t *testing.T) {
		type OtherProject struct {
			ID    int64
			Title string
		}
		project := OtherProject{ID: 1, Title: "Test"}
		assert.False(t, formatter.isHandlersProject(project))
	})

	t.Run("is not handlers project - pointer", func(t *testing.T) {
		type HandlersProject struct {
			ID    int64
			Title string
			URI   string
		}
		project := &HandlersProject{ID: 1, Title: "Test", URI: "test://1"}
		assert.True(t, formatter.isHandlersProject(project))
	})

	t.Run("is not handlers project - non-struct", func(t *testing.T) {
		assert.False(t, formatter.isHandlersProject("string"))
		assert.False(t, formatter.isHandlersProject(123))
		assert.False(t, formatter.isHandlersProject([]string{"test"}))
	})
}

func TestMarkdownFormatter_formatHandlersProject(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format handlers project", func(t *testing.T) {
		type HandlersProject struct {
			ID    int64
			Title string
			URI   string
		}
		project := HandlersProject{ID: 1, Title: "Test Project", URI: "vikunja://project/1"}
		result := formatter.formatHandlersProject(project)

		assert.Contains(t, result, "Test Project")
		assert.Contains(t, result, "ID")
		assert.Contains(t, result, "1")
		assert.Contains(t, result, "URI")
		assert.Contains(t, result, "vikunja://project/1")
	})
}

func TestMarkdownFormatter_Format_UnknownType(t *testing.T) {
	formatter := NewMarkdownFormatter()

	t.Run("format unknown type falls back to JSON", func(t *testing.T) {
		type CustomType struct {
			Name  string
			Value int
		}
		data := CustomType{Name: "Test", Value: 42}
		result, err := formatter.Format(data)

		require.NoError(t, err)
		assert.Contains(t, result, "<!-- Unsupported type for markdown")
		assert.Contains(t, result, "Name")
		assert.Contains(t, result, "Test")
		assert.Contains(t, result, "42")
	})
}

func TestNewBothFormatter(t *testing.T) {
	formatter := NewBothFormatter()
	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.jsonFormatter)
	assert.NotNil(t, formatter.markdownFormatter)
}

func TestBothFormatter_Format(t *testing.T) {
	formatter := NewBothFormatter()

	t.Run("format simple data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		result, err := formatter.Format(data)

		require.NoError(t, err)
		assert.Contains(t, result, "JSON Output")
		assert.Contains(t, result, "Markdown Output")
		assert.Contains(t, result, "key")
		assert.Contains(t, result, "value")
	})

	t.Run("format struct", func(t *testing.T) {
		task := Task{ID: 1, Title: "Test Task"}
		result, err := formatter.Format(task)

		require.NoError(t, err)
		assert.Contains(t, result, "JSON Output")
		assert.Contains(t, result, "Markdown Output")
		assert.Contains(t, result, "Test Task")
	})
}

func TestGetFormatter(t *testing.T) {
	t.Run("get JSON formatter", func(t *testing.T) {
		formatter := GetFormatter(OutputFormatJSON)
		assert.NotNil(t, formatter)
		_, ok := formatter.(*JSONFormatter)
		assert.True(t, ok)
	})

	t.Run("get Markdown formatter", func(t *testing.T) {
		formatter := GetFormatter(OutputFormatMarkdown)
		assert.NotNil(t, formatter)
		_, ok := formatter.(*MarkdownFormatter)
		assert.True(t, ok)
	})

	t.Run("get Both formatter", func(t *testing.T) {
		formatter := GetFormatter(OutputFormatBoth)
		assert.NotNil(t, formatter)
		_, ok := formatter.(*BothFormatter)
		assert.True(t, ok)
	})

	t.Run("default to JSON for unknown format", func(t *testing.T) {
		formatter := GetFormatter(OutputFormat("unknown"))
		assert.NotNil(t, formatter)
		_, ok := formatter.(*JSONFormatter)
		assert.True(t, ok)
	})

	t.Run("default to JSON for empty format", func(t *testing.T) {
		formatter := GetFormatter(OutputFormat(""))
		assert.NotNil(t, formatter)
		_, ok := formatter.(*JSONFormatter)
		assert.True(t, ok)
	})
}

func TestOutputFormatConstants(t *testing.T) {
	assert.Equal(t, OutputFormat("json"), OutputFormatJSON)
	assert.Equal(t, OutputFormat("markdown"), OutputFormatMarkdown)
	assert.Equal(t, OutputFormat("both"), OutputFormatBoth)
}
