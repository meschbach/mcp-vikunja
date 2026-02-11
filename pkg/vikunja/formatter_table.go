package vikunja

import (
	"fmt"
	"text/tabwriter"

	"github.com/fatih/color"
)

// FormatProjects formats a list of projects as a table
func (f *Formatter) FormatProjects(projects []Project) error {
	if f.useColor {
		headerColor := color.New(color.FgCyan, color.Bold)
		_, _ = fmt.Fprintln(f.output, headerColor.Sprint("PROJECTS"))
		_, _ = fmt.Fprintln(f.output)
	}

	w := tabwriter.NewWriter(f.output, 0, 0, 3, ' ', 0)

	if f.useColor {
		headerColor := color.New(color.FgYellow, color.Bold)
		_, _ = fmt.Fprintln(w, headerColor.Sprint("NAME")+"\t"+headerColor.Sprint("ID")+"\t"+headerColor.Sprint("URI"))
	} else {
		_, _ = fmt.Fprintln(w, "NAME\tID\tURI")
	}

	for _, p := range projects {
		uri := fmt.Sprintf("vikunja://projects/%d", p.ID)
		_, _ = fmt.Fprintf(w, "%s\t%d\t%s\n", p.Title, p.ID, uri)
	}

	return w.Flush()
}

// FormatProject formats a single project with full details
func (f *Formatter) FormatProject(project *Project) error {
	if f.useColor {
		titleColor := color.New(color.FgCyan, color.Bold)
		labelColor := color.New(color.FgYellow)
		_, _ = fmt.Fprintf(f.output, "%s\n\n", titleColor.Sprint(project.Title))
		_, _ = fmt.Fprintf(f.output, "%s %d\n", labelColor.Sprint("ID:"), project.ID)
		_, _ = fmt.Fprintf(f.output, "%s %s\n", labelColor.Sprint("URI:"), fmt.Sprintf("vikunja://projects/%d", project.ID))
		if project.Description != "" {
			_, _ = fmt.Fprintf(f.output, "\n%s\n%s\n", labelColor.Sprint("Description:"), project.Description)
		}
	} else {
		_, _ = fmt.Fprintf(f.output, "%s\n\n", project.Title)
		_, _ = fmt.Fprintf(f.output, "ID: %d\n", project.ID)
		_, _ = fmt.Fprintf(f.output, "URI: %s\n", fmt.Sprintf("vikunja://projects/%d", project.ID))
		if project.Description != "" {
			_, _ = fmt.Fprintf(f.output, "\nDescription:\n%s\n", project.Description)
		}
	}
	return nil
}

// FormatTasks formats a list of tasks as a table
func (f *Formatter) FormatTasks(tasks []Task) error {
	if f.useColor {
		headerColor := color.New(color.FgCyan, color.Bold)
		_, _ = fmt.Fprintln(f.output, headerColor.Sprint("TASKS"))
		_, _ = fmt.Fprintln(f.output)
	}

	w := tabwriter.NewWriter(f.output, 0, 0, 3, ' ', 0)

	if f.useColor {
		headerColor := color.New(color.FgYellow, color.Bold)
		_, _ = fmt.Fprintln(w, headerColor.Sprint("TITLE")+"\t"+headerColor.Sprint("ID")+"\t"+headerColor.Sprint("BUCKET")+"\t"+headerColor.Sprint("URI"))
	} else {
		_, _ = fmt.Fprintln(w, "TITLE\tID\tBUCKET\tURI")
	}

	for _, t := range tasks {
		uri := fmt.Sprintf("vikunja://tasks/%d", t.ID)
		bucket := "-"
		if len(t.Buckets) > 0 {
			bucket = t.Buckets[0].Title
		}
		_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", t.Title, t.ID, bucket, uri)
	}

	return w.Flush()
}

// FormatTask formats a single task with full details
func (f *Formatter) FormatTask(task *Task) error {
	if f.useColor {
		titleColor := color.New(color.FgCyan, color.Bold)
		labelColor := color.New(color.FgYellow)
		_, _ = fmt.Fprintf(f.output, "%s\n\n", titleColor.Sprint(task.Title))
		_, _ = fmt.Fprintf(f.output, "%s %d\n", labelColor.Sprint("ID:"), task.ID)
		_, _ = fmt.Fprintf(f.output, "%s %s\n", labelColor.Sprint("URI:"), fmt.Sprintf("vikunja://tasks/%d", task.ID))
		if task.ProjectID > 0 {
			_, _ = fmt.Fprintf(f.output, "%s %d\n", labelColor.Sprint("Project ID:"), task.ProjectID)
		}
		if task.Description != "" {
			_, _ = fmt.Fprintf(f.output, "\n%s\n%s\n", labelColor.Sprint("Description:"), task.Description)
		}
	} else {
		_, _ = fmt.Fprintf(f.output, "%s\n\n", task.Title)
		_, _ = fmt.Fprintf(f.output, "ID: %d\n", task.ID)
		_, _ = fmt.Fprintf(f.output, "URI: %s\n", fmt.Sprintf("vikunja://tasks/%d", task.ID))
		if task.ProjectID > 0 {
			_, _ = fmt.Fprintf(f.output, "Project ID: %d\n", task.ProjectID)
		}
		if task.Description != "" {
			_, _ = fmt.Fprintf(f.output, "\nDescription:\n%s\n", task.Description)
		}
	}
	return nil
}

// FormatBuckets formats a list of buckets
func (f *Formatter) FormatBuckets(buckets []Bucket) error {
	w := tabwriter.NewWriter(f.output, 0, 0, 2, ' ', 0)

	headerColor := color.New(color.FgCyan, color.Bold)
	if !f.useColor {
		headerColor = color.New()
	}

	_, _ = fmt.Fprintln(w, headerColor.Sprint("TITLE")+"\t"+headerColor.Sprint("ID")+"\t"+headerColor.Sprint("DONE")+"\t"+headerColor.Sprint("POSITION"))

	for _, b := range buckets {
		done := "No"
		if b.IsDoneBucket {
			done = "Yes"
		}
		_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%.2f\n", b.Title, b.ID, done, b.Position)
	}

	return w.Flush()
}

// FormatProjectViews formats a list of project views
func (f *Formatter) FormatProjectViews(views []ProjectView) error {
	w := tabwriter.NewWriter(f.output, 0, 0, 2, ' ', 0)

	headerColor := color.New(color.FgCyan, color.Bold)
	if !f.useColor {
		headerColor = color.New()
	}

	_, _ = fmt.Fprintln(w, headerColor.Sprint("TITLE")+"\t"+headerColor.Sprint("ID")+"\t"+headerColor.Sprint("KIND")+"\t"+headerColor.Sprint("POSITION"))

	for _, v := range views {
		_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%.2f\n", v.Title, v.ID, v.ViewKind, v.Position)
	}

	return w.Flush()
}

// FormatTaskWithBuckets formats a single task with bucket information
func (f *Formatter) FormatTaskWithBuckets(task *Task, bucketInfo *TaskBucketInfo) error {
	if f.useColor {
		titleColor := color.New(color.FgCyan, color.Bold)
		labelColor := color.New(color.FgYellow)
		_, _ = fmt.Fprintf(f.output, "%s\n\n", titleColor.Sprint(task.Title))
		_, _ = fmt.Fprintf(f.output, "%s %d\n", labelColor.Sprint("ID:"), task.ID)
		_, _ = fmt.Fprintf(f.output, "%s %s\n", labelColor.Sprint("URI:"), fmt.Sprintf("vikunja://tasks/%d", task.ID))
		if task.ProjectID > 0 {
			_, _ = fmt.Fprintf(f.output, "%s %d\n", labelColor.Sprint("Project ID:"), task.ProjectID)
		}
		if task.Description != "" {
			_, _ = fmt.Fprintf(f.output, "\n%s\n%s\n", labelColor.Sprint("Description:"), task.Description)
		}

		if bucketInfo != nil && len(bucketInfo.Views) > 0 {
			_, _ = fmt.Fprintf(f.output, "\n%s\n", labelColor.Sprint("Bucket Information:"))
			for _, view := range bucketInfo.Views {
				if view.ViewKind == ViewKindKanban && view.BucketTitle != nil {
					doneMark := ""
					if view.IsDoneBucket {
						doneMark = " [DONE]"
					}
					_, _ = fmt.Fprintf(f.output, "  %s (%s): %s%s\n",
						view.ViewTitle, view.ViewKind, *view.BucketTitle, doneMark)
				}
			}
		}
	} else {
		_, _ = fmt.Fprintf(f.output, "%s\n\n", task.Title)
		_, _ = fmt.Fprintf(f.output, "ID: %d\n", task.ID)
		_, _ = fmt.Fprintf(f.output, "URI: %s\n", fmt.Sprintf("vikunja://tasks/%d", task.ID))
		if task.ProjectID > 0 {
			_, _ = fmt.Fprintf(f.output, "Project ID: %d\n", task.ProjectID)
		}
		if task.Description != "" {
			_, _ = fmt.Fprintf(f.output, "\nDescription:\n%s\n", task.Description)
		}

		if bucketInfo != nil && len(bucketInfo.Views) > 0 {
			_, _ = fmt.Fprintf(f.output, "\nBucket Information:\n")
			for _, view := range bucketInfo.Views {
				if view.ViewKind == ViewKindKanban && view.BucketTitle != nil {
					doneMark := ""
					if view.IsDoneBucket {
						doneMark = " [DONE]"
					}
					_, _ = fmt.Fprintf(f.output, "  %s (%s): %s%s\n",
						view.ViewTitle, view.ViewKind, *view.BucketTitle, doneMark)
				}
			}
		}
	}
	return nil
}

// FormatViewTasks formats a view with its buckets and tasks
func (f *Formatter) FormatViewTasks(vt *ViewTasks) error {
	titleColor := color.New(color.FgCyan, color.Bold)
	bucketColor := color.New(color.FgYellow, color.Bold)
	taskColor := color.New(color.FgWhite)
	doneColor := color.New(color.FgHiBlack)

	if !f.useColor {
		titleColor = color.New()
		bucketColor = color.New()
		taskColor = color.New()
		doneColor = color.New()
	}

	_, _ = fmt.Fprintf(f.output, "%s (ID: %d)\n\n", titleColor.Sprint(vt.ViewTitle), vt.ViewID)

	for _, bt := range vt.Buckets {
		doneMark := ""
		if bt.Bucket.IsDoneBucket {
			doneMark = " [DONE]"
		}
		_, _ = fmt.Fprintf(f.output, "%s (ID: %d)%s\n", bucketColor.Sprint(bt.Bucket.Title), bt.Bucket.ID, doneMark)

		if len(bt.Tasks) == 0 {
			_, _ = fmt.Fprintf(f.output, "  (no tasks)\n")
		} else {
			for _, t := range bt.Tasks {
				tc := taskColor
				if t.Done {
					tc = doneColor
				}
				_, _ = fmt.Fprintf(f.output, "  - [Task %d] %s\n", t.ID, tc.Sprint(t.Title))
			}
		}
		_, _ = fmt.Fprintln(f.output)
	}

	return nil
}
