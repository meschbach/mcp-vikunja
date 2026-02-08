package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

var (
	host     string
	token    string
	jsonFmt  bool
	insecure bool
	output   string
	verbose  bool
	noColor  bool
)

var (
	client       *vikunja.Client
	outputWriter io.Writer = os.Stdout
	logger       *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "vikunja-cli",
	Short: "CLI tool for Vikunja task management",
	Long:  `A command-line interface for interacting with Vikunja task management API.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Setup logging
		logLevel := slog.LevelWarn
		if verbose {
			logLevel = slog.LevelDebug
		}
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))

		// Setup color output
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}

		// Validate required flags
		if host == "" {
			return fmt.Errorf("host is required (use --host/-h flag or VIKUNJA_HOST environment variable)")
		}
		if token == "" {
			return fmt.Errorf("token is required (use --token/-t flag or VIKUNJA_TOKEN environment variable)")
		}

		// Setup output writer
		if output != "" {
			file, err := os.Create(filepath.Clean(output))
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			outputWriter = file
		}

		// Initialize client
		var err error
		client, err = vikunja.NewClient(host, token, insecure)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		logger.Debug("client initialized", "host", host)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func GetClient() *vikunja.Client {
	return client
}

func GetOutputWriter() io.Writer {
	return outputWriter
}

func IsJSON() bool {
	return jsonFmt
}

func GetLogger() *slog.Logger {
	return logger
}

func init() {
	// Get defaults from environment
	host = os.Getenv("VIKUNJA_HOST")
	token = os.Getenv("VIKUNJA_TOKEN")

	// Define flags
	rootCmd.PersistentFlags().StringVar(&host, "host", host, "Vikunja instance hostname (or VIKUNJA_HOST env)")
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", token, "API token for authentication (or VIKUNJA_TOKEN env)")
	rootCmd.PersistentFlags().BoolVarP(&jsonFmt, "json", "j", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Write output to file instead of stdout")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging to stderr")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}
