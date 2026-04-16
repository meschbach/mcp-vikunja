package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/internal/handlers"
	"github.com/meschbach/mcp-vikunja/internal/health"
	"github.com/meschbach/mcp-vikunja/internal/logging"
	"github.com/meschbach/mcp-vikunja/internal/transport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	httpHost       string
	httpPort       int
	sessionTimeout string
	stateless      bool
	readTimeout    string
	writeTimeout   string
	idleTimeout    string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start MCP server with HTTP + Streamable transport",
	Long: `Start the MCP server using HTTP + Streamable transport protocol.

This command starts an HTTP server that provides the Model Context Protocol
interface through a streamable HTTP endpoint. This is ideal for web applications
or when you need to serve multiple clients through HTTP.

The server supports session management, configurable timeouts, and can run
in either stateful or stateless mode.`,
	Example: `  mcp-vikunja server
  mcp-vikunja server --http-host 0.0.0.0 --http-port 8080
  mcp-vikunja server --session-timeout 1h --stateless`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Server-specific flags
	serverCmd.Flags().StringVar(&httpHost, "http-host", "localhost", "HTTP server host (env: MCP_HTTP_HOST)")
	serverCmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP server port (env: MCP_HTTP_PORT)")
	serverCmd.Flags().StringVar(&sessionTimeout, "session-timeout", "30m", "Session timeout duration (env: MCP_HTTP_SESSION_TIMEOUT)")
	serverCmd.Flags().BoolVar(&stateless, "stateless", false, "Enable stateless mode (env: MCP_HTTP_STATELESS)")
	serverCmd.Flags().StringVar(&readTimeout, "read-timeout", "30s", "HTTP read timeout (env: MCP_HTTP_READ_TIMEOUT)")
	serverCmd.Flags().StringVar(&writeTimeout, "write-timeout", "30s", "HTTP write timeout (env: MCP_HTTP_WRITE_TIMEOUT)")
	serverCmd.Flags().StringVar(&idleTimeout, "idle-timeout", "2m", "HTTP idle timeout (env: MCP_HTTP_IDLE_TIMEOUT)")
}

func runServer(cmd *cobra.Command, _ []string) error {
	logger, cfg, err := initServer(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := setupShutdownHandler(logger)
	defer cancel()

	return runServerTransport(ctx, cfg, logger)
}

func initServer(cmd *cobra.Command) (*slog.Logger, *config.Config, error) {
	logger, err := initServerLogging(cmd)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := initServerConfig()
	if err != nil {
		return nil, nil, err
	}

	if err := applyServerFlags(cfg); err != nil {
		return nil, nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return logger, cfg, nil
}

func setupShutdownHandler(logger *slog.Logger) (context.Context, context.CancelFunc) {
	//nolint
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	return ctx, cancel
}

func runServerTransport(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {

	// Log configuration
	logger.Info("starting MCP server with HTTP transport",
		"address", cfg.HTTP.Address(),
		"session_timeout", cfg.HTTP.SessionTimeout,
		"stateless", cfg.HTTP.Stateless,
		"vikunja_host", cfg.Vikunja.Host,
		"version", "0.1.0",
	)

	// Create MCP server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-vikunja",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{
			Instructions: `Provides access to Vikunja, a project management tool providing various views
to a set of tasks.  An instance of Vikunja contains a number of projects with a name and an ID.  Each project may contain
a number of views of various types, such as a list, Gantt, a table, or Kanban view.  Views may have multiple buckets
describing tasks in various life cycles.

If you are searching for tasks your primary tool is 'list_tasks' which will search a Vikunja instance for the project,
view, and buckets as you need.
`,
		},
	)

	// Register Vikunja tool handlers
	handlers.Register(s, cfg)

	// Create transport server
	transportServer, err := transport.CreateTransportServer(s, cfg)
	if err != nil {
		return fmt.Errorf("failed to create transport server: %w", err)
	}

	// Setup health checks for HTTP transport
	if cfg.Transport == config.TransportHTTP {
		hc := health.New()
		hc.Register(&health.ServerCheck{})

		// Try to set health checker on HTTP server
		if httpServer, ok := transportServer.(*transport.HTTPServer); ok {
			httpServer.SetHealthChecker(hc)
			logger.Info("health check endpoints registered",
				"endpoints", []string{"/health", "/health/live", "/health/ready"},
			)
		}
	}

	// Start the server
	err = transportServer.Run(ctx)
	if err != nil {
		// Only log error if it's not context cancellation (which is expected)
		if ctx.Err() != context.Canceled {
			logger.Error("server error", "error", err)
			return err
		}
	}
	logger.Info("server shutdown completed")
	return nil
}

func initServerLogging(cmd *cobra.Command) (*slog.Logger, error) {
	logConfig := logging.LoadConfig()
	if verbose, err := cmd.Flags().GetBool("verbose"); err == nil && verbose {
		logConfig.Level = logging.LevelDebug
	}
	logger, err := logging.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return logging.WithComponent(logger, "server"), nil
}

func initServerConfig() (*config.Config, error) {
	var cliFormat *string
	if format := rootCmd.Flag("output-format").Value.String(); format != "" {
		cliFormat = &format
	}
	cliReadonly := &readonly

	cfg, err := config.Load(cliFormat, cliReadonly)
	if err != nil {
		return nil, fmt.Errorf("failed to create server configuration: %w", err)
	}
	cfg.Transport = config.TransportHTTP
	return cfg, nil
}

func applyServerFlags(cfg *config.Config) error {
	if httpHost != "" {
		cfg.HTTP.Host = httpHost
	}
	if httpPort > 0 {
		cfg.HTTP.Port = httpPort
	}

	if err := parseTimeouts(cfg); err != nil {
		return err
	}

	if stateless {
		cfg.HTTP.Stateless = stateless
	}
	return nil
}

type timeoutConfig struct {
	value     string
	field     *time.Duration
	fieldName string
}

func parseTimeouts(cfg *config.Config) error {
	timeouts := []timeoutConfig{
		{sessionTimeout, &cfg.HTTP.SessionTimeout, "session timeout"},
		{readTimeout, &cfg.HTTP.ReadTimeout, "read timeout"},
		{writeTimeout, &cfg.HTTP.WriteTimeout, "write timeout"},
		{idleTimeout, &cfg.HTTP.IdleTimeout, "idle timeout"},
	}

	for _, t := range timeouts {
		if t.value == "" {
			continue
		}
		d, err := time.ParseDuration(t.value)
		if err != nil {
			return fmt.Errorf("invalid %s: %w", t.fieldName, err)
		}
		*t.field = d
	}
	return nil
}
