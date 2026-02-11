package vikunja

import (
	"encoding/json"
)

// FormatAsJSON formats data as JSON
func (f *Formatter) FormatAsJSON(v interface{}) error {
	encoder := json.NewEncoder(f.output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// FormatProjectsAsJSON formats projects as JSON
func (f *Formatter) FormatProjectsAsJSON(projects []Project) error {
	return f.FormatAsJSON(projects)
}

// FormatProjectAsJSON formats a single project as JSON
func (f *Formatter) FormatProjectAsJSON(project *Project) error {
	return f.FormatAsJSON(project)
}

// FormatTasksAsJSON formats tasks as JSON
func (f *Formatter) FormatTasksAsJSON(tasks []Task) error {
	return f.FormatAsJSON(tasks)
}

// FormatTaskAsJSON formats a single task as JSON
func (f *Formatter) FormatTaskAsJSON(task *Task) error {
	return f.FormatAsJSON(task)
}
