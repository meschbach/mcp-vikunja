package vikunja

import (
	"fmt"
	"strings"
	"time"
)

// parseDate parses a date string in various formats
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

// FormatTasksAsMarkdown formats tasks as markdown
func (f *Formatter) FormatTasksAsMarkdown(tasks []*Task) string {
	if len(tasks) == 0 {
		return "## No tasks found\n"
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "## Tasks (%d)\n\n", len(tasks))

	buf.WriteString("| ID | Title | Done | Due Date | Project |\n")
	buf.WriteString("|---|---|---|---|---|\n")

	for _, task := range tasks {
		done := "❌"
		if task.Done {
			done = "✅"
		}

		dueDate := "-"
		if task.DueDate != "" {
			if t := parseDate(task.DueDate); !t.IsZero() {
				dueDate = t.Format("2006-01-02")
			}
		}

		project := "-"
		if task.ProjectID > 0 {
			project = fmt.Sprintf("[%d](vikunja://projects/%d)", task.ProjectID, task.ProjectID)
		}

		title := strings.ReplaceAll(task.Title, "|", "\\|")
		fmt.Fprintf(&buf, "| %d | %s | %s | %s | %s |\n",
			task.ID, title, done, dueDate, project)
	}

	buf.WriteString("\n<details>\n<summary>Task Details</summary>\n\n")
	for _, task := range tasks {
		buf.WriteString(f.formatTaskDetailsMarkdown(task))
	}
	buf.WriteString("</details>\n")

	return buf.String()
}

// FormatTaskAsMarkdown formats a single task as markdown
func (f *Formatter) FormatTaskAsMarkdown(task *Task) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s\n\n", task.Title)
	buf.WriteString(f.formatTaskDetailsMarkdown(task))

	return buf.String()
}

func formatDateField(dateStr, layout, label string, buf *strings.Builder) {
	if dateStr == "" {
		return
	}
	t := parseDate(dateStr)
	if t.IsZero() {
		return
	}
	fmt.Fprintf(buf, "- **%s**: %s\n", label, t.Format(layout))
}

// formatTaskDetailsMarkdown formats detailed task information
func (f *Formatter) formatTaskDetailsMarkdown(task *Task) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "### %s\n\n", task.Title)
	fmt.Fprintf(&buf, "- **ID**: %d\n", task.ID)
	fmt.Fprintf(&buf, "- **URI**: [vikunja://tasks/%d](vikunja://tasks/%d)\n", task.ID, task.ID)

	if task.ProjectID > 0 {
		fmt.Fprintf(&buf, "- **Project**: [%d](vikunja://projects/%d)\n", task.ProjectID, task.ProjectID)
	}

	formatDateField(task.Created, time.RFC3339, "Created", &buf)
	formatDateField(task.Updated, time.RFC3339, "Updated", &buf)
	formatDateField(task.DueDate, "2006-01-02", "Due Date", &buf)

	if task.Done {
		buf.WriteString("- **Status**: ✅ Completed\n")
	} else {
		buf.WriteString("- **Status**: ❌ Pending\n")
	}

	if task.Description != "" {
		fmt.Fprintf(&buf, "\n**Description**:\n%s\n", task.Description)
	}

	buf.WriteString("\n---\n\n")
	return buf.String()
}

func formatProjectField(project *Project, buf *strings.Builder) {
	if project.Identifier != nil && strings.TrimSpace(*project.Identifier) != "" {
		fmt.Fprintf(buf, "- **Identifier**: `%s`\n", *project.Identifier)
	}

	if project.Created != "" {
		t := parseDate(project.Created)
		if !t.IsZero() && t.Year() > 1970 {
			fmt.Fprintf(buf, "- **Created**: %s\n", t.Format("2006-01-02"))
		}
	}

	if project.Description != "" && strings.TrimSpace(project.Description) != "" {
		fmt.Fprintf(buf, "\n**Description**:\n%s\n", project.Description)
	}
}

// FormatProjectsAsMarkdown formats projects as markdown
func (f *Formatter) FormatProjectsAsMarkdown(projects []*Project) string {
	if len(projects) == 0 {
		return "# No projects found\n"
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "# Projects (%d)\n\n", len(projects))

	for _, project := range projects {
		fmt.Fprintf(&buf, "## 📁 %s\n\n", project.Title)
		fmt.Fprintf(&buf, "- **ID**: %d\n", project.ID)
		fmt.Fprintf(&buf, "- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID)

		formatProjectField(project, &buf)

		buf.WriteString("\n---\n\n")
	}

	return buf.String()
}

// FormatProjectAsMarkdown formats a single project as markdown
func (f *Formatter) FormatProjectAsMarkdown(project *Project) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s\n\n", project.Title)
	fmt.Fprintf(&buf, "- **ID**: %d\n", project.ID)
	fmt.Fprintf(&buf, "- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID)

	if project.Identifier != nil && *project.Identifier != "" {
		fmt.Fprintf(&buf, "- **Identifier**: `%s`\n", *project.Identifier)
	}

	if project.Created != "" {
		if t := parseDate(project.Created); !t.IsZero() {
			fmt.Fprintf(&buf, "- **Created**: %s\n", t.Format(time.RFC3339))
		}
	}

	if project.Updated != "" {
		if t := parseDate(project.Updated); !t.IsZero() {
			fmt.Fprintf(&buf, "- **Updated**: %s\n", t.Format(time.RFC3339))
		}
	}

	if project.Description != "" {
		fmt.Fprintf(&buf, "\n**Description**:\n%s\n", project.Description)
	}

	return buf.String()
}

// FormatBucketsAsMarkdown formats buckets as markdown
func (f *Formatter) FormatBucketsAsMarkdown(buckets []*Bucket) string {
	if len(buckets) == 0 {
		return "## No buckets found\n"
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "## 📋 Buckets (%d)\n\n", len(buckets))

	buf.WriteString("| 📁 Bucket | ID | Tasks | Limit |\n")
	buf.WriteString("|---|---|---|---|\n")

	for _, bucket := range buckets {
		taskCount := len(bucket.Tasks)
		limit := "-"
		if bucket.Limit != nil && *bucket.Limit > 0 {
			limit = fmt.Sprintf("%d", *bucket.Limit)
		}

		title := strings.ReplaceAll(bucket.Title, "|", "\\|")
		fmt.Fprintf(&buf, "| %s | %d | %d | %s |\n",
			title, bucket.ID, taskCount, limit)
	}

	return buf.String()
}

// FormatViewAsMarkdown formats a project view as markdown
func (f *Formatter) FormatViewAsMarkdown(view *ProjectView) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s\n\n", view.Title)
	fmt.Fprintf(&buf, "- **ID**: %d\n", view.ID)
	fmt.Fprintf(&buf, "- **Project ID**: %d\n", view.ProjectID)
	fmt.Fprintf(&buf, "- **Type**: %s\n", view.ViewKind)
	fmt.Fprintf(&buf, "- **URI**: [vikunja://views/%d](vikunja://views/%d)\n", view.ID, view.ID)
	fmt.Fprintf(&buf, "- **Position**: %.2f\n", view.Position)

	if view.DefaultBucketID > 0 {
		fmt.Fprintf(&buf, "- **Default Bucket**: %d\n", view.DefaultBucketID)
	}

	if view.DoneBucketID > 0 {
		fmt.Fprintf(&buf, "- **Done Bucket**: %d\n", view.DoneBucketID)
	}

	return buf.String()
}

// FormatViewTasksAsMarkdown formats a view with buckets and tasks as markdown
func (f *Formatter) FormatViewTasksAsMarkdown(vt *ViewTasks) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s (ID: %d)\n\n", vt.ViewTitle, vt.ViewID)

	for i := range vt.Buckets {
		bt := vt.Buckets[i]
		fmt.Fprintf(&buf, "## %s (ID: %d)\n\n", bt.Bucket.Title, bt.Bucket.ID)

		if len(bt.Tasks) == 0 {
			buf.WriteString("(no tasks)\n\n")
		} else {
			for _, task := range bt.Tasks {
				status := "❌"
				if task.Done {
					status = "✅"
				}

				title := strings.ReplaceAll(task.Title, "|", "\\|")
				fmt.Fprintf(&buf, "- %s [Task %d] %s\n", status, task.ID, title)
			}
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// FormatViewTasksSummaryAsMarkdown formats a view with buckets and tasks summary as markdown
func (f *Formatter) FormatViewTasksSummaryAsMarkdown(vt *ViewTasksSummary) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# 📋 %s (ID: %d)\n\n", vt.ViewTitle, vt.ViewID)

	for _, bt := range vt.Buckets {
		doneMark := ""
		// Note: BucketSummary doesn't have IsDoneBucket field, so we can't check it here

		fmt.Fprintf(&buf, "## 📁 %s (ID: %d)%s\n\n", bt.Bucket.Title, bt.Bucket.ID, doneMark)

		if len(bt.Tasks) == 0 {
			buf.WriteString("(no tasks)\n\n")
		} else {
			for _, task := range bt.Tasks {
				// Note: TaskSummary doesn't have Done field, so we can't check completion status

				title := strings.ReplaceAll(task.Title, "|", "\\|") // Escape pipe characters
				fmt.Fprintf(&buf, "- [Task %d] %s\n", task.ID, title)
			}
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

func formatTaskStatus(task *Task, buf *strings.Builder) {
	if task.Done {
		buf.WriteString("- **Status**: ✅ Completed\n")
	} else {
		buf.WriteString("- **Status**: ❌ Pending\n")
	}
}

func formatBucketInfo(bucketInfo *TaskBucketInfo, buf *strings.Builder) {
	if bucketInfo == nil || len(bucketInfo.Views) == 0 {
		return
	}

	buf.WriteString("\n**Bucket Information**:\n")
	for _, view := range bucketInfo.Views {
		if view.ViewKind == ViewKindKanban && view.BucketTitle != nil {
			doneMark := ""
			if view.IsDoneBucket {
				doneMark = " ✅"
			}
			fmt.Fprintf(buf, "- %s (%s): %s%s\n",
				view.ViewTitle, view.ViewKind, *view.BucketTitle, doneMark)
		}
	}
}

// FormatTaskWithBucketsMarkdown formats a task with bucket information as markdown
func (f *Formatter) FormatTaskWithBucketsMarkdown(task *Task, bucketInfo *TaskBucketInfo) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s\n\n", task.Title)
	fmt.Fprintf(&buf, "- **ID**: %d\n", task.ID)
	fmt.Fprintf(&buf, "- **URI**: [vikunja://tasks/%d](vikunja://tasks/%d)\n", task.ID, task.ID)

	if task.ProjectID > 0 {
		fmt.Fprintf(&buf, "- **Project**: [%d](vikunja://projects/%d)\n", task.ProjectID, task.ProjectID)
	}

	formatDateField(task.Created, time.RFC3339, "Created", &buf)
	formatDateField(task.Updated, time.RFC3339, "Updated", &buf)
	formatDateField(task.DueDate, "2006-01-02", "Due Date", &buf)

	formatTaskStatus(task, &buf)

	if task.Description != "" {
		fmt.Fprintf(&buf, "\n**Description**:\n%s\n", task.Description)
	}

	formatBucketInfo(bucketInfo, &buf)

	return buf.String()
}

// FormatProjectAndViewMarkdown formats a project and view as markdown
func (f *Formatter) FormatProjectAndViewMarkdown(project *Project, view *ProjectView) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# 📁 %s\n\n", project.Title)
	fmt.Fprintf(&buf, "- **ID**: %d\n", project.ID)
	fmt.Fprintf(&buf, "- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID)

	if project.Identifier != nil && *project.Identifier != "" {
		fmt.Fprintf(&buf, "- **Identifier**: `%s`\n", *project.Identifier)
	}

	if project.Created != "" {
		if t := parseDate(project.Created); !t.IsZero() {
			fmt.Fprintf(&buf, "- **Created**: %s\n", t.Format(time.RFC3339))
		}
	}

	if project.Updated != "" {
		if t := parseDate(project.Updated); !t.IsZero() {
			fmt.Fprintf(&buf, "- **Updated**: %s\n", t.Format(time.RFC3339))
		}
	}

	if project.Description != "" {
		fmt.Fprintf(&buf, "\n**Description**:\n%s\n", project.Description)
	}

	buf.WriteString("\n---\n\n")
	buf.WriteString(f.FormatViewAsMarkdown(view))

	return buf.String()
}

var viewEmojiMap = map[string]string{
	"kanban": "📋 ",
	"list":   "📝 ",
	"gantt":  "📊 ",
	"table":  "📋 ",
}

func getViewEmoji(viewKind string) string {
	if emoji, ok := viewEmojiMap[viewKind]; ok {
		return emoji
	}
	return ""
}

func formatProjectDetails(project *Project, buf *strings.Builder) {
	if project.Identifier != nil && strings.TrimSpace(*project.Identifier) != "" {
		fmt.Fprintf(buf, "- **Identifier**: `%s`\n", *project.Identifier)
	}

	if project.Created != "" {
		t := parseDate(project.Created)
		if !t.IsZero() {
			fmt.Fprintf(buf, "- **Created**: %s\n", t.Format(time.RFC3339))
		}
	}

	if project.Updated != "" {
		t := parseDate(project.Updated)
		if !t.IsZero() {
			fmt.Fprintf(buf, "- **Updated**: %s\n", t.Format(time.RFC3339))
		}
	}

	if project.Description != "" {
		fmt.Fprintf(buf, "\n**Description**:\n%s\n", project.Description)
	}
}

// FormatProjectAndViewListMarkdown formats a project and multiple views as markdown
func (f *Formatter) FormatProjectAndViewListMarkdown(project *Project, views []*ProjectView) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# 📁 %s\n\n", project.Title)
	fmt.Fprintf(&buf, "- **ID**: %d\n", project.ID)
	fmt.Fprintf(&buf, "- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID)

	formatProjectDetails(project, &buf)

	fmt.Fprintf(&buf, "\n## Views (%d)\n\n", len(views))
	buf.WriteString("| 📋 View | ID | Type | Position |\n")
	buf.WriteString("|---|---|---|---|\n")

	for _, view := range views {
		title := strings.ReplaceAll(view.Title, "|", "\\|")
		viewEmoji := getViewEmoji(string(view.ViewKind))

		fmt.Fprintf(&buf, "| %s%s | %d | %s | %.2f |\n", viewEmoji, title, view.ID, view.ViewKind, view.Position)
	}

	return buf.String()
}
