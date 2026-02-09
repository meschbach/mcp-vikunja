// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionInfo contains version and build information
var versionInfo = struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
}{
	Version:   "0.1.0",
	Commit:    "unknown",
	BuildTime: "unknown",
	GoVersion: runtime.Version(),
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display detailed version and build information for the MCP Vikunja server.

This includes the version number, git commit hash, build time, and Go runtime
version. Useful for debugging and ensuring you're running the expected build.`,
	Example: `  mcp-vikunja version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-vikunja version %s\n", versionInfo.Version)
		fmt.Printf("Commit:      %s\n", versionInfo.Commit)
		fmt.Printf("Built:       %s\n", versionInfo.BuildTime)
		fmt.Printf("Go version:  %s\n", versionInfo.GoVersion)
		fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// SetVersionInfo allows setting version information at build time
func SetVersionInfo(version, commit, buildTime string) {
	if version != "" {
		versionInfo.Version = version
	}
	if commit != "" {
		versionInfo.Commit = commit
	}
	if buildTime != "" {
		versionInfo.BuildTime = buildTime
	}
}
