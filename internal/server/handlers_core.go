package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// Core tool handlers for indexing, search, and metadata operations

// handleIndexRepository handles repository indexing requests
func (s *MCPServer) handleIndexRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path parameter: %v", err)), nil
	}

	name := request.GetString("name", "")

	s.logger.Info("Indexing repository", zap.String("path", path), zap.String("name", name))

	// Index the repository
	repo, err := s.indexer.IndexRepository(ctx, path, name)
	if err != nil {
		s.logger.Error("Failed to index repository", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to index repository: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":    true,
		"repository": repo,
		"message":    "Repository indexed successfully",
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

	searchType := request.GetString("type", "")
	language := request.GetString("language", "")
	repository := request.GetString("repository", "")
	maxResults := int(request.GetFloat("max_results", 100))

	s.logger.Info("Searching code", 
		zap.String("query", query), 
		zap.String("type", searchType),
		zap.String("language", language),
		zap.String("repository", repository),
		zap.Int("max_results", maxResults))

	// Perform the search
	searchQuery := types.SearchQuery{
		Query:      query,
		Type:       searchType,
		Language:   language,
		Repository: repository,
		MaxResults: maxResults,
	}

	results, err := s.searcher.Search(ctx, searchQuery)
	if err != nil {
		s.logger.Error("Failed to search code", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
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

	// Get file metadata (this would be implemented based on your search engine capabilities)
	result := map[string]interface{}{
		"file_path":  filePath,
		"repository": repository,
		"metadata": map[string]interface{}{
			"language": s.repoMgr.GetFileLanguage(filePath),
			"exists":   true, // This would be determined by actual file check
		},
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

	result := map[string]interface{}{
		"repositories": repositories,
		"count":        len(repositories),
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

	result := map[string]interface{}{
		"stats": stats,
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}
