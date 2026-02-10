// Package cmd provides cobra commands for the MCP Vikunja server.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Manage configuration for the MCP Vikunja server.

This command group provides utilities for viewing, validating, and managing
the server configuration. Configuration is primarily handled through environment
variables, but these commands help with debugging and setup.`,
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current configuration by loading from environment variables
and command-line flags. This is useful for debugging configuration issues
and verifying that the server will start with the expected settings.`,
	Example: `  mcp-vikunja config show
  mcp-vikunja config show --format json`,
	RunE: runConfigShow,
}

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long: `Validate the current configuration by loading it and checking for
common issues. This includes verifying required fields, checking network
availability, and testing Vikunja connectivity.`,
	Example: `  mcp-vikunja config validate`,
	RunE:    runConfigValidate,
}

var (
	configFormat string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)

	// Config show flags
	configShowCmd.Flags().StringVar(&configFormat, "format", "table", "Output format: table|json")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfigFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	switch configFormat {
	case "json":
		return showConfigJSON(cfg)
	case "table":
		return showConfigTable(cfg)
	default:
		return fmt.Errorf("unsupported format: %s", configFormat)
	}
}

func showConfigTable(cfg *config.Config) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SETTING\tVALUE")
	fmt.Fprintln(w, "-------\t-----")

	// Transport settings
	fmt.Fprintf(w, "Transport\t%s\n", cfg.Transport)

	// Vikunja settings
	fmt.Fprintf(w, "Vikunja Host\t%s\n", maskSensitive(cfg.Vikunja.Host))
	fmt.Fprintf(w, "Vikunja Token\t%s\n", maskSensitive(cfg.Vikunja.Token))

	// HTTP settings (only if HTTP transport)
	if cfg.Transport == config.TransportHTTP {
		fmt.Fprintf(w, "HTTP Host\t%s\n", cfg.HTTP.Host)
		fmt.Fprintf(w, "HTTP Port\t%d\n", cfg.HTTP.Port)
		fmt.Fprintf(w, "Session Timeout\t%s\n", cfg.HTTP.SessionTimeout)
		fmt.Fprintf(w, "Stateless\t%t\n", cfg.HTTP.Stateless)
		fmt.Fprintf(w, "Read Timeout\t%s\n", cfg.HTTP.ReadTimeout)
		fmt.Fprintf(w, "Write Timeout\t%s\n", cfg.HTTP.WriteTimeout)
		fmt.Fprintf(w, "Idle Timeout\t%s\n", cfg.HTTP.IdleTimeout)
	}

	return nil
}

func showConfigJSON(cfg *config.Config) error {
	// Create a copy for JSON output with masked sensitive data
	jsonCfg := struct {
		Transport config.TransportType `json:"transport"`
		HTTP      config.HTTPConfig    `json:"http,omitempty"`
		Vikunja   struct {
			Host  string `json:"host"`
			Token string `json:"token"`
		} `json:"vikunja"`
	}{
		Transport: cfg.Transport,
		HTTP:      cfg.HTTP,
		Vikunja: struct {
			Host  string `json:"host"`
			Token string `json:"token"`
		}{
			Host:  cfg.Vikunja.Host,
			Token: maskSensitive(cfg.Vikunja.Token),
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonCfg)
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	fmt.Printf("Loading configuration...\n")

	// Load configuration
	cfg, err := loadConfigFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Validating configuration...\n")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ Configuration validation failed:\n")
		fmt.Printf("   %s\n", err.Error())
		return err
	}

	fmt.Printf("✅ Configuration is valid\n")

	// Show summary
	fmt.Printf("\nConfiguration Summary:\n")
	fmt.Printf("  Transport: %s\n", cfg.Transport)
	fmt.Printf("  Vikunja Host: %s\n", cfg.Vikunja.Host)

	if cfg.Transport == config.TransportHTTP {
		fmt.Printf("  HTTP Address: %s\n", cfg.HTTP.Address())
		fmt.Printf("  Session Timeout: %s\n", cfg.HTTP.SessionTimeout)
		fmt.Printf("  Stateless: %t\n", cfg.HTTP.Stateless)
	}

	fmt.Printf("\n✅ MCP server should start successfully\n")
	return nil
}

func loadConfigFromFlags(cmd *cobra.Command) (*config.Config, error) {
	// Start with default configuration
	cfg := &config.Config{
		Transport: config.TransportStdio, // Default
		HTTP: config.HTTPConfig{
			Host:           "localhost",
			Port:           8080,
			SessionTimeout: 30 * time.Minute,
			Stateless:      false,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			IdleTimeout:    120 * time.Second,
		},
		Vikunja: config.VikunjaConfig{},
	}

	// Override with flags if provided
	if host := cmd.Flag("vikunja-host").Value.String(); host != "" {
		cfg.Vikunja.Host = host
	}
	if token := cmd.Flag("vikunja-token").Value.String(); token != "" {
		cfg.Vikunja.Token = token
	}

	// Load from environment variables (this will override flags if env vars are set)
	envCfg, err := config.Load(nil, nil) // No CLI parameters for config commands
	if err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	// Merge environment config (only if not explicitly set via flags)
	if cfg.Vikunja.Host == "" && envCfg.Vikunja.Host != "" {
		cfg.Vikunja.Host = envCfg.Vikunja.Host
	}
	if cfg.Vikunja.Token == "" && envCfg.Vikunja.Token != "" {
		cfg.Vikunja.Token = envCfg.Vikunja.Token
	}

	// Use environment transport if not explicitly set
	if envCfg.Transport != "" {
		cfg.Transport = envCfg.Transport
		cfg.HTTP = envCfg.HTTP
	}

	return cfg, nil
}

func maskSensitive(value string) string {
	if value == "" {
		return "<not set>"
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}
