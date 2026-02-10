// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	readonly     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcp-vikunja",
	Short: "MCP server for Vikunja task management",
	Long: `mcp-vikunja provides Model Context Protocol integration with Vikunja.

This server allows MCP clients to interact with Vikunja task management
through either stdio or HTTP + Streamable transport protocols.

Available subcommands:
  server  Start MCP server with HTTP + Streamable transport
  stdio   Start MCP server with stdio transport
  config  Configuration management
  version Show version information
  health  Test Vikunja connection

Examples:
  mcp-vikunja server --host 0.0.0.0 --port 8080
  mcp-vikunja stdio --vikunja-host https://vikunja.example.com
  mcp-vikunja config show
  mcp-vikunja health`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Don't show help for expected shutdown scenarios
		if !isExpectedShutdown(err) {
			os.Exit(1)
		}
		os.Exit(0)
	}
}

// isExpectedShutdown checks if the error represents an expected shutdown scenario
func isExpectedShutdown(err error) bool {
	if err == nil {
		return true
	}
	// Check for context cancellation or other expected shutdown conditions
	errStr := err.Error()
	return contains(errStr, "context canceled") ||
		contains(errStr, "signal") ||
		contains(errStr, "shutdown")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("vikunja-host", "", "Vikunja instance URL (env: VIKUNJA_HOST)")
	rootCmd.PersistentFlags().String("vikunja-token", "", "Vikunja API token (env: VIKUNJA_TOKEN)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "", "Output format: json (legacy), markdown (default), both (CLI overrides VIKUNJA_OUTPUT_FORMAT)")
	rootCmd.PersistentFlags().BoolVar(&readonly, "readonly", false, "Enable readonly mode to prevent write operations (env: MCP_READONLY)")
}
