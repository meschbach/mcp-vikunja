package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/fatih/color"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

type itemFetcher func(ctx context.Context, id int64) (interface{}, error)

func newGetCmd(itemType string, fetch itemFetcher, isProject bool) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: fmt.Sprintf("Get a %s by ID", itemType),
		Long:  fmt.Sprintf("Retrieve detailed information about a specific %s.", itemType),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getByID(cmd.Context(), args[0], itemType, fetch, isProject)
		},
	}
}

func getByID(ctx context.Context, arg, itemType string, fetch itemFetcher, isProject bool) error {
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s ID: %s (must be a number)", itemType, arg)
	}

	logger.Debug("getting "+itemType, "id", id)
	item, err := fetch(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get %s %d: %w", itemType, id, err)
	}

	if item == nil {
		return printNotFound(itemType, id)
	}

	formatter := vikunja.NewFormatter(!noColor, outputWriter)
	return formatItem(formatter, item, isProject)
}

func printNotFound(itemType string, id int64) error {
	msg := fmt.Sprintf("%s not found: %d\n", capitalize(itemType), id)
	if !noColor {
		color.Red(msg)
		return nil
	}
	return writeAll(outputWriter, msg)
}

func formatItem(formatter *vikunja.Formatter, item interface{}, isProject bool) error {
	switch {
	case jsonFmt:
		return formatJSON(formatter, item, isProject)
	case markdown:
		return formatMarkdown(formatter, item, isProject)
	default:
		return formatDefault(formatter, item, isProject)
	}
}

func formatJSON(formatter *vikunja.Formatter, item interface{}, isProject bool) error {
	if isProject {
		project, ok := item.(*vikunja.Project)
		if !ok {
			return fmt.Errorf("expected project type")
		}
		return formatter.FormatProjectAsJSON(project)
	}
	task, ok := item.(*vikunja.Task)
	if !ok {
		return fmt.Errorf("expected task type")
	}
	return formatter.FormatTaskAsJSON(task)
}

func formatMarkdown(formatter *vikunja.Formatter, item interface{}, isProject bool) error {
	var output string
	if isProject {
		project, ok := item.(*vikunja.Project)
		if !ok {
			return fmt.Errorf("expected project type")
		}
		output = formatter.FormatProjectAsMarkdown(project)
	} else {
		task, ok := item.(*vikunja.Task)
		if !ok {
			return fmt.Errorf("expected task type")
		}
		output = formatter.FormatTaskAsMarkdown(task)
	}
	return writeAll(outputWriter, output)
}

func formatDefault(formatter *vikunja.Formatter, item interface{}, isProject bool) error {
	if isProject {
		project, ok := item.(*vikunja.Project)
		if !ok {
			return fmt.Errorf("expected project type")
		}
		return formatter.FormatProject(project)
	}
	task, ok := item.(*vikunja.Task)
	if !ok {
		return fmt.Errorf("expected task type")
	}
	return formatter.FormatTaskAsJSON(task)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] &^= 32
	return string(runes)
}

func writeAll(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}
