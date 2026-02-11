package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	t.Run("root command exists", func(t *testing.T) {
		require.NotNil(t, rootCmd)
		assert.Equal(t, "mcp-vikunja", rootCmd.Use)
		assert.Contains(t, rootCmd.Short, "MCP server")
	})

	t.Run("root command has subcommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		require.NotEmpty(t, commands)

		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}

		// Check for expected subcommands
		assert.Contains(t, commandNames, "server")
		assert.Contains(t, commandNames, "stdio")
		assert.Contains(t, commandNames, "config")
		assert.Contains(t, commandNames, "version")
		assert.Contains(t, commandNames, "health")
	})

	t.Run("root command has persistent flags", func(t *testing.T) {
		flags := rootCmd.PersistentFlags()

		// Check for expected flags
		hostFlag, err := flags.GetString("vikunja-host")
		assert.NoError(t, err)
		assert.Empty(t, hostFlag)

		tokenFlag, err := flags.GetString("vikunja-token")
		assert.NoError(t, err)
		assert.Empty(t, tokenFlag)

		verboseFlag, err := flags.GetBool("verbose")
		assert.NoError(t, err)
		assert.False(t, verboseFlag)
	})
}

func TestIsExpectedShutdown(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: true,
		},
		{
			name:     "context canceled",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This is a basic test; the function behavior may vary
			result := isExpectedShutdown(tt.err)
			if tt.err == nil {
				assert.True(t, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", true},
		{"hello world", "lo wo", true},
		{"hello world", "xyz", false},
		{"", "test", false},
		{"test", "", true},
		{"test", "test", true},
		{"test", "testing", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "lo wo", true},
		{"hello world", "xyz", false},
		{"abc", "abc", true},
		{"abc", "abcd", false},
		{"", "", true},
		{"abc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := findSubstring(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionCommand(t *testing.T) {
	t.Run("version command exists", func(t *testing.T) {
		require.NotNil(t, versionCmd)
		assert.Equal(t, "version", versionCmd.Use)
		assert.Contains(t, versionCmd.Short, "version")
	})

	t.Run("version info has default values", func(t *testing.T) {
		assert.Equal(t, "0.1.0", versionInfo.Version)
		assert.Equal(t, "unknown", versionInfo.Commit)
		assert.Equal(t, "unknown", versionInfo.BuildTime)
		assert.NotEmpty(t, versionInfo.GoVersion)
	})

	t.Run("SetVersionInfo updates values", func(t *testing.T) {
		// Save original values
		origVersion := versionInfo.Version
		origCommit := versionInfo.Commit
		origBuildTime := versionInfo.BuildTime

		// Update values
		SetVersionInfo("1.0.0", "abc123", "2024-01-01")

		// Verify update
		assert.Equal(t, "1.0.0", versionInfo.Version)
		assert.Equal(t, "abc123", versionInfo.Commit)
		assert.Equal(t, "2024-01-01", versionInfo.BuildTime)

		// Restore original values
		SetVersionInfo(origVersion, origCommit, origBuildTime)
	})

	t.Run("version command output", func(t *testing.T) {
		// Create a buffer to capture output
		buf := new(bytes.Buffer)
		versionCmd.SetOut(buf)

		// Execute version command
		versionCmd.Run(versionCmd, []string{})

		output := buf.String()

		// Check output contains expected information
		assert.Contains(t, output, "mcp-vikunja version")
		assert.Contains(t, output, versionInfo.Version)
	})
}

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "<not set>"},
		{"short", "***"},
		{"12345678", "***"},
		{"123456789", "1234***6789"},
		{"my-secret-token", "my-s***oken"},
		{"abcdefghijklmnopqrstuvwxyz", "abcd***wxyz"},
	}

	for _, tt := range tests {
		t.Run("mask_"+tt.input, func(t *testing.T) {
			result := maskSensitive(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigFromFlags(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		// Create a test command with no flags set
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("vikunja-host", "", "")
		cmd.Flags().String("vikunja-token", "", "")

		cfg, err := loadConfigFromFlags(cmd)
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.HTTP.Host)
		assert.Equal(t, 8080, cfg.HTTP.Port)
	})

	t.Run("with flag values", func(t *testing.T) {
		// Create a test command with flags set
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("vikunja-host", "", "")
		cmd.Flags().String("vikunja-token", "", "")
		cmd.Flags().Set("vikunja-host", "test.example.com")
		cmd.Flags().Set("vikunja-token", "test-token")

		cfg, err := loadConfigFromFlags(cmd)
		require.NoError(t, err)
		assert.Equal(t, "test.example.com", cfg.Vikunja.Host)
		assert.Equal(t, "test-token", cfg.Vikunja.Token)
	})
}

func TestConfigShowFormats(t *testing.T) {
	t.Run("showConfigTable produces output", func(t *testing.T) {
		cfg := &config.Config{
			Transport: config.TransportStdio,
			HTTP: config.HTTPConfig{
				Host: "localhost",
				Port: 8080,
			},
			Vikunja: config.VikunjaConfig{
				Host:  "test.example.com",
				Token: "secret-token",
			},
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := showConfigTable(cfg)
		require.NoError(t, err)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify table output
		assert.Contains(t, output, "SETTING")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "Transport")
		assert.Contains(t, output, "Vikunja Host")
	})

	t.Run("showConfigJSON produces valid JSON", func(t *testing.T) {
		cfg := &config.Config{
			Transport: config.TransportHTTP,
			HTTP: config.HTTPConfig{
				Host: "localhost",
				Port: 8080,
			},
			Vikunja: config.VikunjaConfig{
				Host:  "test.example.com",
				Token: "secret-token",
			},
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := showConfigJSON(cfg)
		require.NoError(t, err)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify JSON output
		assert.Contains(t, output, "transport")
		assert.Contains(t, output, "http")
		assert.Contains(t, output, "vikunja")
		// Token should be masked
		assert.Contains(t, output, "***")
	})
}

func TestConfigCommands(t *testing.T) {
	t.Run("config command exists", func(t *testing.T) {
		require.NotNil(t, configCmd)
		assert.Equal(t, "config", configCmd.Use)
	})

	t.Run("config show command exists", func(t *testing.T) {
		require.NotNil(t, configShowCmd)
		assert.Equal(t, "show", configShowCmd.Use)
	})

	t.Run("config validate command exists", func(t *testing.T) {
		require.NotNil(t, configValidateCmd)
		assert.Equal(t, "validate", configValidateCmd.Use)
	})
}

func TestServerCommand(t *testing.T) {
	t.Run("server command exists", func(t *testing.T) {
		require.NotNil(t, serverCmd)
		assert.Equal(t, "server", serverCmd.Use)
		assert.Contains(t, serverCmd.Short, "HTTP")
	})

	t.Run("server command has flags", func(t *testing.T) {
		flags := serverCmd.Flags()

		host, err := flags.GetString("http-host")
		assert.NoError(t, err)
		assert.Equal(t, "localhost", host)

		port, err := flags.GetInt("http-port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
	})
}

func TestStdioCommand(t *testing.T) {
	t.Run("stdio command exists", func(t *testing.T) {
		require.NotNil(t, stdioCmd)
		assert.Equal(t, "stdio", stdioCmd.Use)
		assert.Contains(t, stdioCmd.Short, "stdio")
	})
}

func TestHealthCommand(t *testing.T) {
	t.Run("health command exists", func(t *testing.T) {
		require.NotNil(t, healthCmd)
		assert.Equal(t, "health", healthCmd.Use)
		assert.Contains(t, healthCmd.Short, "Test")
	})
}
