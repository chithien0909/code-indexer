package server

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/internal/models"
	"github.com/my-mcp/code-indexer/internal/repository"
	"github.com/my-mcp/code-indexer/internal/search"
)

// MCPServer wraps the MCP server with our application logic
type MCPServer struct {
	server       *server.MCPServer
	config       *config.Config
	logger       *zap.Logger
	indexer      *indexer.Indexer
	repoMgr      *repository.Manager
	searcher     *search.Engine
	modelsEngine *models.Engine
}

// New creates a new MCP server instance
func New(cfg *config.Config, logger *zap.Logger) (*MCPServer, error) {
	// Create MCP server with configuration
	opts := []server.ServerOption{
		server.WithToolCapabilities(true),
	}

	// Always enable recovery for stability
	opts = append(opts, server.WithRecovery())

	mcpServer := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		opts...,
	)

	// Initialize components
	repoMgr, err := repository.NewManager("./repositories", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository manager: %w", err)
	}

	searcher, err := search.NewEngine("./index", logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}

	idx, err := indexer.New(cfg, repoMgr, searcher, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	modelsEngine, err := models.NewEngine(&cfg.Models, idx, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create models engine: %w", err)
	}

	s := &MCPServer{
		server:       mcpServer,
		config:       cfg,
		logger:       logger,
		indexer:      idx,
		repoMgr:      repoMgr,
		searcher:     searcher,
		modelsEngine: modelsEngine,
	}

	// Register MCP tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// Serve starts the MCP server using stdio transport
func (s *MCPServer) Serve() error {
	s.logger.Info("Starting MCP server", 
		zap.String("name", s.config.Server.Name), 
		zap.String("version", s.config.Server.Version))
	return server.ServeStdio(s.server)
}

// Close gracefully shuts down the server
func (s *MCPServer) Close() error {
	s.logger.Info("Shutting down MCP server")

	if err := s.searcher.Close(); err != nil {
		s.logger.Error("Failed to close search engine", zap.Error(err))
	}

	if err := s.modelsEngine.Close(); err != nil {
		s.logger.Error("Failed to close models engine", zap.Error(err))
	}

	return nil
}
