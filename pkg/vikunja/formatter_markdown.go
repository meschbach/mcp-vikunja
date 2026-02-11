package vikunja

import (
	"fmt"
	"strings"
	"time"
)

// FormatTasksAsMarkdown formats tasks as markdown
func (f *Formatter) FormatTasksAsMarkdown(tasks []Task) string {
	if len(tasks) == 0 {
		return "## No tasks found\n"
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("## Tasks (%d)\n\n", len(tasks)))

	// Main task table
	buf.WriteString("| ID | Title | Done | Due Date | Project |\n")
	buf.WriteString("|---|---|---|---|---|\n")

	for _, task := range tasks {
		done := "❌"
		if task.Done {
			done = "✅"
		}

		dueDate := "-"
		if !task.DueDate.IsZero() {
			dueDate = task.DueDate.Format("2006-01-02")
		}

		project := "-"
		if task.ProjectID > 0 {
			project = fmt.Sprintf("[%d](vikunja://projects/%d)", task.ProjectID, task.ProjectID)
		}

		title := strings.ReplaceAll(task.Title, "|", "\\|") // Escape pipe characters
		buf.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			task.ID, title, done, dueDate, project))
	}

	buf.WriteString("\n<details>\n<summary>Task Details</summary>\n\n")
	for _, task := range tasks {
		buf.WriteString(f.formatTaskDetailsMarkdown(task))
	}
	buf.WriteString("</details>\n")

	return buf.String()
}

// FormatTaskAsMarkdown formats a single task as markdown
func (f *Formatter) FormatTaskAsMarkdown(task Task) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# %s\n\n", task.Title))
	buf.WriteString(f.formatTaskDetailsMarkdown(task))

	return buf.String()
}

// formatTaskDetailsMarkdown formats detailed task information
func (f *Formatter) formatTaskDetailsMarkdown(task Task) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("### %s\n\n", task.Title))
	buf.WriteString(fmt.Sprintf("- **ID**: %d\n", task.ID))
	buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://tasks/%d](vikunja://tasks/%d)\n", task.ID, task.ID))

	if task.ProjectID > 0 {
		buf.WriteString(fmt.Sprintf("- **Project**: [%d](vikunja://projects/%d)\n", task.ProjectID, task.ProjectID))
	}

	if !task.Created.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Created**: %s\n", task.Created.Format(time.RFC3339)))
	}

	if !task.Updated.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Updated**: %s\n", task.Updated.Format(time.RFC3339)))
	}

	if !task.DueDate.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Due Date**: %s\n", task.DueDate.Format("2006-01-02")))
	}

	if task.Done {
		buf.WriteString("- **Status**: ✅ Completed\n")
	} else {
		buf.WriteString("- **Status**: ❌ Pending\n")
	}

	if task.Description != "" {
		buf.WriteString(fmt.Sprintf("\n**Description**:\n%s\n", task.Description))
	}

	buf.WriteString("\n---\n\n")
	return buf.String()
}

// FormatProjectsAsMarkdown formats projects as markdown
func (f *Formatter) FormatProjectsAsMarkdown(projects []Project) string {
	if len(projects) == 0 {
		return "# No projects found\n"
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("# Projects (%d)\n\n", len(projects)))

	for _, project := range projects {
		buf.WriteString(fmt.Sprintf("## 📁 %s\n\n", project.Title))
		buf.WriteString(fmt.Sprintf("- **ID**: %d\n", project.ID))
		buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID))

		if project.Identifier != "" && strings.TrimSpace(project.Identifier) != "" {
			buf.WriteString(fmt.Sprintf("- **Identifier**: `%s`\n", project.Identifier))
		}

		if !project.Created.IsZero() && project.Created.Year() > 1970 {
			buf.WriteString(fmt.Sprintf("- **Created**: %s\n", project.Created.Format("2006-01-02")))
		}

		if project.Description != "" && strings.TrimSpace(project.Description) != "" {
			buf.WriteString(fmt.Sprintf("\n**Description**:\n%s\n", project.Description))
		}

		buf.WriteString("\n---\n\n")
	}

	return buf.String()
}

// FormatProjectAsMarkdown formats a single project as markdown
func (f *Formatter) FormatProjectAsMarkdown(project Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# %s\n\n", project.Title))
	buf.WriteString(fmt.Sprintf("- **ID**: %d\n", project.ID))
	buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID))

	if project.Identifier != "" {
		buf.WriteString(fmt.Sprintf("- **Identifier**: `%s`\n", project.Identifier))
	}

	if project.OwnerID > 0 {
		buf.WriteString(fmt.Sprintf("- **Owner ID**: %d\n", project.OwnerID))
	}

	if !project.Created.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Created**: %s\n", project.Created.Format(time.RFC3339)))
	}

	if !project.Updated.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Updated**: %s\n", project.Updated.Format(time.RFC3339)))
	}

	if project.Description != "" {
		buf.WriteString(fmt.Sprintf("\n**Description**:\n%s\n", project.Description))
	}

	return buf.String()
}

// FormatBucketsAsMarkdown formats buckets as markdown
func (f *Formatter) FormatBucketsAsMarkdown(buckets []Bucket) string {
	if len(buckets) == 0 {
		return "## No buckets found\n"
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("## 📋 Buckets (%d)\n\n", len(buckets)))

	buf.WriteString("| 📁 Bucket | ID | Tasks | Limit | Done |\n")
	buf.WriteString("|---|---|---|---|---|\n")

	for _, bucket := range buckets {
		taskCount := len(bucket.Tasks)
		limit := "-"
		if bucket.Limit > 0 {
			limit = fmt.Sprintf("%d", bucket.Limit)
		}

		done := "❌"
		if bucket.IsDoneBucket {
			done = "✅"
		}

		title := strings.ReplaceAll(bucket.Title, "|", "\\|") // Escape pipe characters
		buf.WriteString(fmt.Sprintf("| %s | %d | %d | %s | %s |\n",
			title, bucket.ID, taskCount, limit, done))
	}

	return buf.String()
}

// FormatViewAsMarkdown formats a project view as markdown
func (f *Formatter) FormatViewAsMarkdown(view ProjectView) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# %s\n\n", view.Title))
	buf.WriteString(fmt.Sprintf("- **ID**: %d\n", view.ID))
	buf.WriteString(fmt.Sprintf("- **Project ID**: %d\n", view.ProjectID))
	buf.WriteString(fmt.Sprintf("- **Type**: %s\n", view.ViewKind))
	buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://views/%d](vikunja://views/%d)\n", view.ID, view.ID))
	buf.WriteString(fmt.Sprintf("- **Position**: %.2f\n", view.Position))

	if view.DefaultBucketID > 0 {
		buf.WriteString(fmt.Sprintf("- **Default Bucket**: %d\n", view.DefaultBucketID))
	}

	if view.DoneBucketID > 0 {
		buf.WriteString(fmt.Sprintf("- **Done Bucket**: %d\n", view.DoneBucketID))
	}

	return buf.String()
}

// FormatViewTasksAsMarkdown formats a view with buckets and tasks as markdown
func (f *Formatter) FormatViewTasksAsMarkdown(vt *ViewTasks) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# %s (ID: %d)\n\n", vt.ViewTitle, vt.ViewID))

	for _, bt := range vt.Buckets {
		doneMark := ""
		if bt.Bucket.IsDoneBucket {
			doneMark = " ✅"
		}

		buf.WriteString(fmt.Sprintf("## %s (ID: %d)%s\n\n", bt.Bucket.Title, bt.Bucket.ID, doneMark))

		if len(bt.Tasks) == 0 {
			buf.WriteString("(no tasks)\n\n")
		} else {
			for _, task := range bt.Tasks {
				status := "❌"
				if task.Done {
					status = "✅"
				}

				title := strings.ReplaceAll(task.Title, "|", "\\|") // Escape pipe characters
				buf.WriteString(fmt.Sprintf("- %s [Task %d] %s\n", status, task.ID, title))
			}
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// FormatViewTasksSummaryAsMarkdown formats a view with buckets and tasks summary as markdown
func (f *Formatter) FormatViewTasksSummaryAsMarkdown(vt *ViewTasksSummary) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# 📋 %s (ID: %d)\n\n", vt.ViewTitle, vt.ViewID))

	for _, bt := range vt.Buckets {
		doneMark := ""
		// Note: BucketSummary doesn't have IsDoneBucket field, so we can't check it here

		buf.WriteString(fmt.Sprintf("## 📁 %s (ID: %d)%s\n\n", bt.Bucket.Title, bt.Bucket.ID, doneMark))

		if len(bt.Tasks) == 0 {
			buf.WriteString("(no tasks)\n\n")
		} else {
			for _, task := range bt.Tasks {
				// Note: TaskSummary doesn't have Done field, so we can't check completion status

				title := strings.ReplaceAll(task.Title, "|", "\\|") // Escape pipe characters
				buf.WriteString(fmt.Sprintf("- [Task %d] %s\n", task.ID, title))
			}
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// FormatTaskWithBucketsMarkdown formats a task with bucket information as markdown
func (f *Formatter) FormatTaskWithBucketsMarkdown(task Task, bucketInfo *TaskBucketInfo) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# %s\n\n", task.Title))
	buf.WriteString(fmt.Sprintf("- **ID**: %d\n", task.ID))
	buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://tasks/%d](vikunja://tasks/%d)\n", task.ID, task.ID))

	if task.ProjectID > 0 {
		buf.WriteString(fmt.Sprintf("- **Project**: [%d](vikunja://projects/%d)\n", task.ProjectID, task.ProjectID))
	}

	if !task.Created.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Created**: %s\n", task.Created.Format(time.RFC3339)))
	}

	if !task.Updated.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Updated**: %s\n", task.Updated.Format(time.RFC3339)))
	}

	if !task.DueDate.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Due Date**: %s\n", task.DueDate.Format("2006-01-02")))
	}

	if task.Done {
		buf.WriteString("- **Status**: ✅ Completed\n")
	} else {
		buf.WriteString("- **Status**: ❌ Pending\n")
	}

	if task.Description != "" {
		buf.WriteString(fmt.Sprintf("\n**Description**:\n%s\n", task.Description))
	}

	if bucketInfo != nil && len(bucketInfo.Views) > 0 {
		buf.WriteString("\n**Bucket Information**:\n")
		for _, view := range bucketInfo.Views {
			if view.ViewKind == ViewKindKanban && view.BucketTitle != nil {
				doneMark := ""
				if view.IsDoneBucket {
					doneMark = " ✅"
				}
				buf.WriteString(fmt.Sprintf("- %s (%s): %s%s\n",
					view.ViewTitle, view.ViewKind, *view.BucketTitle, doneMark))
			}
		}
	}

	return buf.String()
}

// FormatProjectAndViewMarkdown formats a project and view as markdown
func (f *Formatter) FormatProjectAndViewMarkdown(project Project, view ProjectView) string {
	var buf strings.Builder

	// Project section
	buf.WriteString(fmt.Sprintf("# 📁 %s\n\n", project.Title))
	buf.WriteString(fmt.Sprintf("- **ID**: %d\n", project.ID))
	buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID))

	if project.Identifier != "" {
		buf.WriteString(fmt.Sprintf("- **Identifier**: `%s`\n", project.Identifier))
	}

	if project.OwnerID > 0 {
		buf.WriteString(fmt.Sprintf("- **Owner ID**: %d\n", project.OwnerID))
	}

	if !project.Created.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Created**: %s\n", project.Created.Format(time.RFC3339)))
	}

	if !project.Updated.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Updated**: %s\n", project.Updated.Format(time.RFC3339)))
	}

	if project.Description != "" {
		buf.WriteString(fmt.Sprintf("\n**Description**:\n%s\n", project.Description))
	}

	// View section
	buf.WriteString("\n---\n\n")
	buf.WriteString(f.FormatViewAsMarkdown(view))

	return buf.String()
}

// FormatProjectAndViewListMarkdown formats a project and multiple views as markdown
func (f *Formatter) FormatProjectAndViewListMarkdown(project Project, views []ProjectView) string {
	var buf strings.Builder

	// Project section
	buf.WriteString(fmt.Sprintf("# 📁 %s\n\n", project.Title))
	buf.WriteString(fmt.Sprintf("- **ID**: %d\n", project.ID))
	buf.WriteString(fmt.Sprintf("- **URI**: [vikunja://projects/%d](vikunja://projects/%d)\n", project.ID, project.ID))

	if project.Identifier != "" {
		buf.WriteString(fmt.Sprintf("- **Identifier**: `%s`\n", project.Identifier))
	}

	if project.OwnerID > 0 {
		buf.WriteString(fmt.Sprintf("- **Owner ID**: %d\n", project.OwnerID))
	}

	if !project.Created.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Created**: %s\n", project.Created.Format(time.RFC3339)))
	}

	if !project.Updated.IsZero() {
		buf.WriteString(fmt.Sprintf("- **Updated**: %s\n", project.Updated.Format(time.RFC3339)))
	}

	if project.Description != "" {
		buf.WriteString(fmt.Sprintf("\n**Description**:\n%s\n", project.Description))
	}

	// Views section
	buf.WriteString(fmt.Sprintf("\n## Views (%d)\n\n", len(views)))
	buf.WriteString("| 📋 View | ID | Type | Position |\n")
	buf.WriteString("|---|---|---|---|\n")

	for _, view := range views {
		title := strings.ReplaceAll(view.Title, "|", "\\|") // Escape pipe characters

		// Add emoji based on view type
		viewEmoji := ""
		switch view.ViewKind {
		case ViewKindKanban:
			viewEmoji = "📋 "
		case ViewKindList:
			viewEmoji = "📝 "
		case ViewKindGantt:
			viewEmoji = "📊 "
		case ViewKindTable:
			viewEmoji = "📋 "
		}

		buf.WriteString(fmt.Sprintf("| %s%s | %d | %s | %.2f |\n", viewEmoji, title, view.ID, view.ViewKind, view.Position))
	}

	return buf.String()
}
