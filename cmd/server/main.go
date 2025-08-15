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
	rootCmd.AddCommand(mcpServerCmd())
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

func mcpServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp-server",
		Short: "Start the MCP server (optimized for uvx)",
		Long: `Start the MCP server optimized for direct uvx execution.
This command is designed to be invoked directly by uvx without requiring
a separate daemon process. It provides the same functionality as 'serve'
but with optimizations for process spawning and uvx integration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPServer()
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

func runMCPServer() error {
	// Load configuration with uvx-optimized defaults
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override log level if specified
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	// For uvx execution, optimize logging for stdio
	// Disable file logging to avoid conflicts with stdio communication
	if cfg.Logging.File != "" && configPath == "" {
		// Only disable file logging if using default config
		cfg.Logging.File = ""
	}

	// Initialize logger with uvx-optimized settings
	logger, err := initLoggerForUVX(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Debug("Starting MCP Code Indexer (uvx mode)",
		zap.String("version", "1.1.0"),
		zap.String("log_level", cfg.Logging.Level),
		zap.String("mode", "uvx-optimized"))

	// Create MCP server with uvx optimizations
	mcpServer, err := server.NewForUVX(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Setup graceful shutdown (simplified for uvx)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Debug("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- mcpServer.ServeStdio()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Debug("Shutting down MCP server...")
		if err := mcpServer.Close(); err != nil {
			logger.Error("Error during server shutdown", zap.Error(err))
		}
		return nil
	case err := <-serverErr:
		if err != nil {
			logger.Error("MCP server error", zap.Error(err))
			return err
		}
		return nil
	}
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

// initLoggerForUVX initializes a logger optimized for uvx execution
func initLoggerForUVX(cfg config.LoggingConfig) (*zap.Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config optimized for uvx
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// For uvx mode, we want minimal logging to stderr to avoid interfering with stdio
	// Only log errors and warnings to stderr, debug/info to file if specified
	var cores []zapcore.Core

	// Always add a stderr core for errors and warnings
	stderrLevel := zapcore.WarnLevel
	if level == zapcore.DebugLevel {
		stderrLevel = zapcore.DebugLevel
	}

	stderrCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stderr),
		stderrLevel,
	)
	cores = append(cores, stderrCore)

	// Add file core if file logging is enabled
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		var encoder zapcore.Encoder
		if cfg.JSONFormat {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(file),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Create logger with multiple cores
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
