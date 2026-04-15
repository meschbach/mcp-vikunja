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

func runConfigShow(cmd *cobra.Command, _ []string) error {
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

	lines := buildConfigLines(cfg)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			flushErr := w.Flush()
			if flushErr != nil {
				return flushErr
			}
			return err
		}
	}

	return w.Flush()
}

func buildConfigLines(cfg *config.Config) []string {
	lines := []string{
		"SETTING\tVALUE",
		"-------\t-----",
		"Transport\t" + string(cfg.Transport),
		"Vikunja Host\t" + maskSensitive(cfg.Vikunja.Host),
		"Vikunja Token\t" + maskSensitive(cfg.Vikunja.Token),
	}

	if cfg.Transport == config.TransportHTTP {
		lines = append(lines,
			"HTTP Host\t"+cfg.HTTP.Host,
			"HTTP Port\t"+fmt.Sprintf("%d", cfg.HTTP.Port),
			"Session Timeout\t"+cfg.HTTP.SessionTimeout.String(),
			"Stateless\t"+fmt.Sprintf("%t", cfg.HTTP.Stateless),
			"Read Timeout\t"+cfg.HTTP.ReadTimeout.String(),
			"Write Timeout\t"+cfg.HTTP.WriteTimeout.String(),
			"Idle Timeout\t"+cfg.HTTP.IdleTimeout.String(),
		)
	}

	return lines
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

func runConfigValidate(cmd *cobra.Command, _ []string) error {
	cmd.Printf("Loading configuration...\n")

	// Load configuration
	cfg, err := loadConfigFromFlags(cmd)
	if err != nil {
		cmd.Printf("❌ Failed to load configuration: %v\n", err)
		return err
	}

	cmd.Printf("Validating configuration...\n")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		cmd.Printf("❌ Configuration validation failed:\n")
		cmd.Printf("   %s\n", err.Error())
		return err
	}

	cmd.Printf("✅ Configuration is valid\n")

	// Show summary
	cmd.Printf("\nConfiguration Summary:\n")
	cmd.Printf("  Transport: %s\n", cfg.Transport)
	cmd.Printf("  Vikunja Host: %s\n", cfg.Vikunja.Host)

	if cfg.Transport == config.TransportHTTP {
		cmd.Printf("  HTTP Address: %s\n", cfg.HTTP.Address())
		cmd.Printf("  Session Timeout: %s\n", cfg.HTTP.SessionTimeout)
		cmd.Printf("  Stateless: %t\n", cfg.HTTP.Stateless)
	}

	cmd.Printf("\n✅ MCP server should start successfully\n")
	return nil
}

func loadConfigFromFlags(cmd *cobra.Command) (*config.Config, error) {
	cfg := newDefaultConfig()

	applyVikunjaFlags(cmd, cfg)

	envCfg, err := config.Load(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	mergeEnvConfig(cfg, envCfg)

	return cfg, nil
}

func newDefaultConfig() *config.Config {
	return &config.Config{
		Transport: config.TransportStdio,
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
}

func applyVikunjaFlags(cmd *cobra.Command, cfg *config.Config) {
	if host := cmd.Flag("vikunja-host").Value.String(); host != "" {
		cfg.Vikunja.Host = host
	}
	if token := cmd.Flag("vikunja-token").Value.String(); token != "" {
		cfg.Vikunja.Token = token
	}
}

func mergeEnvConfig(cfg, envCfg *config.Config) {
	if cfg.Vikunja.Host == "" && envCfg.Vikunja.Host != "" {
		cfg.Vikunja.Host = envCfg.Vikunja.Host
	}
	if cfg.Vikunja.Token == "" && envCfg.Vikunja.Token != "" {
		cfg.Vikunja.Token = envCfg.Vikunja.Token
	}
	if envCfg.Transport != "" {
		cfg.Transport = envCfg.Transport
		cfg.HTTP = envCfg.HTTP
	}
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
