package vikunja

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// OutputFormatter defines the interface for formatting output
type OutputFormatter interface {
	Format(data interface{}) (string, error)
}

// OutputFormat represents the desired output format
type OutputFormat string

const (
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatMarkdown OutputFormat = "markdown"
	OutputFormatBoth     OutputFormat = "both"
)

// JSONFormatter formats data as JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format formats data as JSON
func (f *JSONFormatter) Format(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}

// isHandlersProject checks if the given interface is the handlers.Project type (has URI field)
func (f *MarkdownFormatter) isHandlersProject(v interface{}) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}

	// Check for ID, Title, and URI fields
	idField := val.FieldByName("ID")
	titleField := val.FieldByName("Title")
	uriField := val.FieldByName("URI")

	return idField.IsValid() && titleField.IsValid() && uriField.IsValid()
}

// formatHandlersProject formats a handlers.Project type (with URI field) as markdown
func (f *MarkdownFormatter) formatHandlersProject(v interface{}) string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	id := val.FieldByName("ID").Int()
	title := val.FieldByName("Title").String()
	uri := val.FieldByName("URI").String()

	return fmt.Sprintf("# %s\n\n- **ID**: %d\n- **URI**: %s\n", title, id, uri)
}

// MarkdownFormatter formats data as markdown using the Formatter
type MarkdownFormatter struct {
	formatter *Formatter
}

// NewMarkdownFormatter creates a new markdown formatter
func NewMarkdownFormatter() *MarkdownFormatter {
	// Create a formatter with capture output capability
	formatter := NewFormatter(false, nil) // No color for markdown output
	return &MarkdownFormatter{
		formatter: formatter,
	}
}

// Format formats data as markdown based on the data type
func (f *MarkdownFormatter) Format(data interface{}) (string, error) {
	switch v := data.(type) {
	case []Task:
		return f.formatter.FormatTasksAsMarkdown(v), nil
	case Task:
		return f.formatter.FormatTaskAsMarkdown(v), nil
	case []Project:
		return f.formatter.FormatProjectsAsMarkdown(v), nil
	case Project:
		return f.formatter.FormatProjectAsMarkdown(v), nil
	case []Bucket:
		return f.formatter.FormatBucketsAsMarkdown(v), nil
	case Bucket:
		return f.formatter.FormatBucketsAsMarkdown([]Bucket{v}), nil
	case ProjectView:
		return f.formatter.FormatViewAsMarkdown(v), nil
	case []ProjectView:
		// Handle multiple views
		var result string
		for i, view := range v {
			if i > 0 {
				result += "\n---\n\n"
			}
			result += f.formatter.FormatViewAsMarkdown(view)
		}
		return result, nil
	case *ViewTasks:
		return f.formatter.FormatViewTasksAsMarkdown(v), nil
	case ViewTasksSummary:
		return f.formatter.FormatViewTasksSummaryAsMarkdown(&v), nil
	case *ViewTasksSummary:
		return f.formatter.FormatViewTasksSummaryAsMarkdown(v), nil
	case TaskOutput:
		return f.formatter.FormatTaskWithBucketsMarkdown(v.Task, v.Buckets), nil
	case ViewOutput:
		return f.formatter.FormatProjectAndViewMarkdown(v.Project, v.View), nil
	case ViewsOutput:
		return f.formatter.FormatProjectAndViewListMarkdown(v.Project, v.Views), nil
	default:
		// Check if this is the handlers.Project type (with URI field)
		if f.isHandlersProject(v) {
			return f.formatHandlersProject(v), nil
		}
		// Fallback to JSON for unknown types
		return fmt.Sprintf("<!-- Unsupported type for markdown, falling back to JSON -->\n```json\n%s\n```",
			func() string {
				jsonData, _ := json.MarshalIndent(v, "", "  ")
				return string(jsonData)
			}()), nil
	}
}

// BothFormatter returns both JSON and markdown formats
type BothFormatter struct {
	jsonFormatter     *JSONFormatter
	markdownFormatter *MarkdownFormatter
}

// NewBothFormatter creates a new formatter that returns both formats
func NewBothFormatter() *BothFormatter {
	return &BothFormatter{
		jsonFormatter:     NewJSONFormatter(),
		markdownFormatter: NewMarkdownFormatter(),
	}
}

// Format formats data as both JSON and markdown
func (f *BothFormatter) Format(data interface{}) (string, error) {
	jsonOutput, err := f.jsonFormatter.Format(data)
	if err != nil {
		return "", err
	}

	markdownOutput, err := f.markdownFormatter.Format(data)
	if err != nil {
		return "", err
	}

	// Combine both formats with clear separation
	return fmt.Sprintf("# Both JSON and Markdown Output\n\n## Markdown Output\n\n%s\n\n---\n\n## JSON Output\n\n```json\n%s\n```",
		markdownOutput, jsonOutput), nil
}

// GetFormatter returns the appropriate formatter based on the output format
func GetFormatter(format OutputFormat) OutputFormatter {
	switch format {
	case OutputFormatJSON:
		return NewJSONFormatter()
	case OutputFormatMarkdown:
		return NewMarkdownFormatter()
	case OutputFormatBoth:
		return NewBothFormatter()
	default:
		return NewJSONFormatter() // Default to JSON
	}
}
