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

// OutputFormat represents the desired output format.
type OutputFormat string

// OutputFormat constants define the available output formats.
const (
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatMarkdown OutputFormat = "markdown"
	OutputFormatBoth     OutputFormat = "both"
)

// MarkdownFormatter formats data as markdown using the Formatter
type MarkdownFormatter struct {
	formatter *Formatter
}

func (f *MarkdownFormatter) formatSliceAsMarkdown(v interface{}) (string, error) {
	switch data := v.(type) {
	case []*Task:
		return f.formatter.FormatTasksAsMarkdown(data), nil
	case []*Project:
		return f.formatter.FormatProjectsAsMarkdown(data), nil
	case []*Bucket:
		return f.formatter.FormatBucketsAsMarkdown(data), nil
	case []*ProjectView:
		var result string
		for i, view := range data {
			if i > 0 {
				result += "\n---\n\n"
			}
			result += f.formatter.FormatViewAsMarkdown(view)
		}
		return result, nil
	default:
		return "", fmt.Errorf("unsupported slice type for markdown")
	}
}

func (f *MarkdownFormatter) formatPointerAsMarkdown(v interface{}) (string, error) {
	switch data := v.(type) {
	case *Task, *Project, *Bucket, *ProjectView, *ViewTasks, *ViewTasksSummary:
		return f.formatViaReflect(data)
	case TaskOutput:
		return f.formatter.FormatTaskWithBucketsMarkdown(&data.Task, data.Buckets), nil
	case ViewOutput:
		return f.formatter.FormatProjectAndViewMarkdown(&data.Project, &data.View), nil
	default:
		return "", fmt.Errorf("unsupported pointer type for markdown")
	}
}

func (f *MarkdownFormatter) formatViaReflect(v interface{}) (string, error) {
	switch data := v.(type) {
	case *Task:
		return f.formatter.FormatTaskAsMarkdown(data), nil
	case *Project:
		return f.formatter.FormatProjectAsMarkdown(data), nil
	case *Bucket:
		return f.formatter.FormatBucketsAsMarkdown([]*Bucket{data}), nil
	case *ProjectView:
		return f.formatter.FormatViewAsMarkdown(data), nil
	case *ViewTasks:
		return f.formatter.FormatViewTasksAsMarkdown(data), nil
	case *ViewTasksSummary:
		return f.formatter.FormatViewTasksSummaryAsMarkdown(data), nil
	default:
		return "", fmt.Errorf("unsupported type")
	}
}

func (f *MarkdownFormatter) formatValueAsMarkdown(v interface{}) (string, error) {
	switch data := v.(type) {
	case ViewTasksSummary:
		return f.formatter.FormatViewTasksSummaryAsMarkdown(&data), nil
	case ViewsOutput:
		return f.formatter.FormatProjectAndViewListMarkdown(&data.Project, data.Views), nil
	default:
		if f.isHandlersProject(data) {
			return f.formatHandlersProject(data), nil
		}
		return f.formatFallbackMarkdown(data)
	}
}

func (f *MarkdownFormatter) formatFallbackMarkdown(v interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal fallback JSON: %w", err)
	}
	return fmt.Sprintf("<!-- Unsupported type for markdown, falling back to JSON -->\n```json\n%s\n```", string(jsonData)), nil
}

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

// NewMarkdownFormatter creates a new markdown formatter
func NewMarkdownFormatter() *MarkdownFormatter {
	formatter := NewFormatter(false, nil)
	return &MarkdownFormatter{
		formatter: formatter,
	}
}

// Format formats data as markdown based on the data type.
func (f *MarkdownFormatter) Format(data interface{}) (string, error) {
	switch v := data.(type) {
	case []*Task, []*Project, []*Bucket, []*ProjectView:
		return f.formatSliceAsMarkdown(v)
	case *Task, *Project, *Bucket, *ProjectView, *ViewTasks, *ViewTasksSummary, TaskOutput, ViewOutput:
		return f.formatPointerAsMarkdown(v)
	case ViewTasksSummary, ViewsOutput:
		return f.formatValueAsMarkdown(v)
	default:
		if f.isHandlersProject(v) {
			return f.formatHandlersProject(v), nil
		}
		return f.formatFallbackMarkdown(v)
	}
}

func (f *MarkdownFormatter) isHandlersProject(v interface{}) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}

	idField := val.FieldByName("ID")
	titleField := val.FieldByName("Title")
	uriField := val.FieldByName("URI")

	return idField.IsValid() && titleField.IsValid() && uriField.IsValid()
}

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
