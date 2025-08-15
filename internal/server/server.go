package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/internal/repository"
	"github.com/my-mcp/code-indexer/internal/search"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// MCPServer wraps the MCP server with our application logic
type MCPServer struct {
	server     *server.MCPServer
	config     *config.Config
	logger     *zap.Logger
	indexer    *indexer.Indexer
	repoMgr    *repository.Manager
	searcher   *search.Engine
}

// New creates a new MCP server instance
func New(cfg *config.Config, logger *zap.Logger) (*MCPServer, error) {
	// Create MCP server with configuration
	opts := []server.ServerOption{
		server.WithToolCapabilities(true),
	}

	if cfg.Server.EnableRecovery {
		opts = append(opts, server.WithRecovery())
	}

	mcpServer := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		opts...,
	)

	// Initialize components
	repoMgr, err := repository.NewManager(cfg.Indexer.RepoDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository manager: %w", err)
	}

	searcher, err := search.NewEngine(cfg.Indexer.IndexDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}

	idx, err := indexer.New(cfg, repoMgr, searcher, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	s := &MCPServer{
		server:   mcpServer,
		config:   cfg,
		logger:   logger,
		indexer:  idx,
		repoMgr:  repoMgr,
		searcher: searcher,
	}

	// Register MCP tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// registerTools registers all MCP tools
func (s *MCPServer) registerTools() error {
	// Index Repository Tool
	indexRepoTool := mcp.NewTool("index_repository",
		mcp.WithDescription("Index a Git repository for searching"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Local path or Git URL to repository"),
		),
		mcp.WithString("name",
			mcp.Description("Custom name for the repository (optional)"),
		),
	)
	s.server.AddTool(indexRepoTool, s.handleIndexRepository)

	// Search Code Tool
	searchCodeTool := mcp.NewTool("search_code",
		mcp.WithDescription("Search across all indexed repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithString("type",
			mcp.Description("Search type: function, class, variable, content, file, comment"),
			mcp.Enum("function", "class", "variable", "content", "file", "comment"),
		),
		mcp.WithString("language",
			mcp.Description("Filter by programming language (e.g., go, python, javascript)"),
		),
		mcp.WithString("repository",
			mcp.Description("Filter by repository name"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return (default: 100)"),
		),
	)
	s.server.AddTool(searchCodeTool, s.handleSearchCode)

	// Get Metadata Tool
	getMetadataTool := mcp.NewTool("get_metadata",
		mcp.WithDescription("Get detailed metadata for a specific file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name (optional)"),
		),
	)
	s.server.AddTool(getMetadataTool, s.handleGetMetadata)

	// List Repositories Tool
	listReposTool := mcp.NewTool("list_repositories",
		mcp.WithDescription("List all indexed repositories with statistics"),
	)
	s.server.AddTool(listReposTool, s.handleListRepositories)

	// Get Index Stats Tool
	getStatsTool := mcp.NewTool("get_index_stats",
		mcp.WithDescription("Get indexing statistics and information"),
	)
	s.server.AddTool(getStatsTool, s.handleGetIndexStats)

	return nil
}

// handleIndexRepository handles repository indexing requests
func (s *MCPServer) handleIndexRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path parameter: %v", err)), nil
	}

	name := request.GetString("name", "")

	s.logger.Info("Starting repository indexing", zap.String("path", path), zap.String("name", name))

	// Start indexing (this could be made async with progress reporting)
	repo, err := s.indexer.IndexRepository(ctx, path, name)
	if err != nil {
		s.logger.Error("Failed to index repository", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to index repository: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"repository":  repo,
		"message":     fmt.Sprintf("Successfully indexed repository '%s'", repo.Name),
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleSearchCode handles code search requests
func (s *MCPServer) handleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid query parameter: %v", err)), nil
	}

	searchQuery := types.SearchQuery{
		Query:      query,
		MaxResults: s.config.Search.MaxResults,
	}

	searchQuery.Type = request.GetString("type", "")
	searchQuery.Language = request.GetString("language", "")
	searchQuery.Repository = request.GetString("repository", "")

	maxResults := request.GetFloat("max_results", 0)
	if maxResults > 0 {
		searchQuery.MaxResults = int(maxResults)
	}

	s.logger.Info("Performing code search", zap.String("query", query), zap.String("type", searchQuery.Type))

	results, err := s.searcher.Search(ctx, searchQuery)
	if err != nil {
		s.logger.Error("Search failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	result := map[string]any{
		"success":      true,
		"query":        searchQuery,
		"results":      results,
		"total_found":  len(results),
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleGetMetadata handles file metadata requests
func (s *MCPServer) handleGetMetadata(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")

	s.logger.Info("Getting file metadata", zap.String("file_path", filePath), zap.String("repository", repository))

	metadata, err := s.searcher.GetFileMetadata(ctx, filePath, repository)
	if err != nil {
		s.logger.Error("Failed to get file metadata", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file metadata: %v", err)), nil
	}

	result := map[string]any{
		"success":  true,
		"metadata": metadata,
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleListRepositories handles repository listing requests
func (s *MCPServer) handleListRepositories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Listing repositories")

	repositories, err := s.searcher.ListRepositories(ctx)
	if err != nil {
		s.logger.Error("Failed to list repositories", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
	}

	result := map[string]any{
		"success":      true,
		"repositories": repositories,
		"total_count":  len(repositories),
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleGetIndexStats handles index statistics requests
func (s *MCPServer) handleGetIndexStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Getting index statistics")

	stats, err := s.searcher.GetIndexStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get index statistics", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get index statistics: %v", err)), nil
	}

	result := map[string]any{
		"success": true,
		"stats":   stats,
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// Serve starts the MCP server using stdio transport
func (s *MCPServer) Serve() error {
	s.logger.Info("Starting MCP server", zap.String("name", s.config.Server.Name), zap.String("version", s.config.Server.Version))
	return server.ServeStdio(s.server)
}

// Close gracefully shuts down the server
func (s *MCPServer) Close() error {
	s.logger.Info("Shutting down MCP server")
	
	if err := s.searcher.Close(); err != nil {
		s.logger.Error("Failed to close search engine", zap.Error(err))
	}
	
	return nil
}
