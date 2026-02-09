// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
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
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("vikunja-host", "", "Vikunja instance URL (env: VIKUNJA_HOST)")
	rootCmd.PersistentFlags().String("vikunja-token", "", "Vikunja API token (env: VIKUNJA_TOKEN)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
}
