// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Test Vikunja connection",
	Long: `Test the connection to the Vikunja instance and verify API access.

This command performs a health check by connecting to the configured Vikunja
instance and testing API access. It validates the host, token, and basic
functionality to ensure the MCP server will be able to communicate properly.`,
	Example: `  mcp-vikunja health
  mcp-vikunja health --vikunja-host https://vikunja.example.com --vikunja-token your-token`,
	RunE: runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	// Get configuration from flags or environment
	host := cmd.Flag("vikunja-host").Value.String()
	if host == "" {
		host = os.Getenv("VIKUNJA_HOST")
	}
	if host == "" {
		return fmt.Errorf("vikunja host is required (use --vikunja-host or VIKUNJA_HOST)")
	}

	token := cmd.Flag("vikunja-token").Value.String()
	if token == "" {
		token = os.Getenv("VIKUNJA_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("vikunja token is required (use --vikunja-token or VIKUNJA_TOKEN)")
	}

	cmd.Printf("Testing connection to Vikunja at: %s\n", host)

	// Create Vikunja client
	vikunjaClient, err := vikunja.NewClient(host, token, false)
	if err != nil {
		return fmt.Errorf("failed to create Vikunja client: %w", err)
	}

	// Test connection by fetching projects
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd.Printf("Fetching projects...\n")
	projects, err := vikunjaClient.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Vikunja: %w", err)
	}

	cmd.Printf("✓ Successfully connected to Vikunja\n")
	cmd.Printf("✓ Found %d projects\n", len(projects))

	// Test fetching tasks from first project (if any)
	if len(projects) > 0 {
		cmd.Printf("Testing task access from project '%s'...\n", projects[0].Title)
		tasks, err := vikunjaClient.GetTasks(ctx, projects[0].ID)
		if err != nil {
			return fmt.Errorf("failed to fetch tasks: %w", err)
		}
		cmd.Printf("✓ Successfully accessed %d tasks\n", len(tasks))
	}

	cmd.Printf("✓ All health checks passed - MCP server should work correctly\n")
	return nil
}
