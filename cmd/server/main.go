package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/server"
)

var (
	configPath string
	logLevel   string
	port       int
	host       string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "code-indexer",
		Short: "MCP Code Indexer - Index and search source code repositories",
		Long: `A Model Context Protocol (MCP) server that indexes source code from multiple 
repositories and provides powerful search capabilities for LLM applications.`,
	}

	// Add flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")

	// Add commands
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(daemonCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long:  "Start the MCP server and listen for connections via stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}
}

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Start the MCP server as a daemon",
		Long:  "Start the MCP server as a background daemon listening on TCP port for multiple VSCode instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon()
		},
	}

	// Add daemon-specific flags
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to bind to")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("MCP Code Indexer v1.0.0")
		},
	}
}

func runServer() error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override log level if specified
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting MCP Code Indexer",
		zap.String("version", "1.0.0"),
		zap.String("log_level", cfg.Logging.Level))

	// Create MCP server
	mcpServer, err := server.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- mcpServer.Serve()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down server...")
		if err := mcpServer.Close(); err != nil {
			logger.Error("Error during server shutdown", zap.Error(err))
		}
		return nil
	case err := <-serverErr:
		if err != nil {
			logger.Error("Server error", zap.Error(err))
			return err
		}
		return nil
	}
}

func runDaemon() error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override log level if specified
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting MCP Code Indexer Daemon",
		zap.String("version", "1.0.0"),
		zap.String("host", host),
		zap.Int("port", port),
		zap.String("log_level", cfg.Logging.Level))

	// Create MCP server
	mcpServer, err := server.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Start daemon server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- mcpServer.ServeDaemon(host, port)
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down daemon...")
		if err := mcpServer.Close(); err != nil {
			logger.Error("Error during daemon shutdown", zap.Error(err))
		}
		return nil
	case err := <-serverErr:
		if err != nil {
			logger.Error("Daemon error", zap.Error(err))
			return err
		}
		return nil
	}
}

func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create writer syncer
	var writeSyncer zapcore.WriteSyncer
	if cfg.OutputPath != "" && cfg.OutputPath != "stdout" {
		// TODO: Add file rotation support using lumberjack
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
