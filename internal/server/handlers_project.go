package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/my-mcp/code-indexer/pkg/types"
	"go.uber.org/zap"
)

// Project management tool handlers for configuration and project operations

// handleGetCurrentConfig handles current configuration requests
func (s *MCPServer) handleGetCurrentConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get current config", zap.String("tool", request.Params.Name))

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	// Get repository statistics
	repoStats, err := s.searcher.GetIndexStats(ctx)
	var statsInterface interface{}
	if err != nil {
		s.logger.Warn("Failed to get repository stats", zap.Error(err))
		statsInterface = map[string]interface{}{"error": "Failed to retrieve stats"}
	} else {
		statsInterface = repoStats
	}

	// Get available repositories
	repositories, err := s.searcher.ListRepositories(ctx)
	if err != nil {
		s.logger.Warn("Failed to list repositories", zap.Error(err))
		repositories = []types.Repository{}
	}

	config := map[string]interface{}{
		"server": map[string]interface{}{
			"name":    s.config.Server.Name,
			"version": s.config.Server.Version,
			"status":  "running",
		},
		"project": map[string]interface{}{
			"working_directory": cwd,
			"repositories":      repositories,
			"repository_count":  len(repositories),
		},
		"tools": map[string]interface{}{
			"total_count": 20,
			"categories": map[string]interface{}{
				"core":    5,
				"utility": 7, // Updated to include file manipulation tools
				"ai":      3,
				"project": 5, // New category
			},
		},
		"indexer": map[string]interface{}{
			"enabled": true,
			"stats":   statsInterface,
		},
		"models": map[string]interface{}{
			"enabled":       s.modelsEngine.IsEnabled(),
			"default_model": s.config.Models.DefaultModel,
		},
		"system": map[string]interface{}{
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format configuration"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleInitialInstructions handles initial instructions requests
func (s *MCPServer) handleInitialInstructions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling initial instructions", zap.String("tool", request.Params.Name))

	instructions := map[string]interface{}{
		"title": "MCP Code Indexer - Initial Instructions",
		"description": "Welcome to the MCP Code Indexer! This tool provides intelligent code analysis and assistance.",
		"instructions": []string{
			"1. Start by indexing your repositories using 'index_repository' tool",
			"2. Use 'search_code' to find specific code patterns across your indexed repositories",
			"3. Explore files and symbols using 'find_files' and 'find_symbols' tools",
			"4. Get file content and metadata using 'get_file_content' and 'get_metadata' tools",
			"5. Use AI tools for code generation, analysis, and explanation",
			"6. Manage your project using project management tools",
			"7. Edit files directly using file manipulation tools",
		},
		"available_tools": map[string]interface{}{
			"core_tools": []string{
				"index_repository - Index Git repositories for searching",
				"search_code - Search across all indexed repositories",
				"get_metadata - Get detailed metadata for specific files",
				"list_repositories - List all indexed repositories",
				"get_index_stats - Get indexing statistics",
			},
			"utility_tools": []string{
				"find_files - Find files matching patterns",
				"find_symbols - Find symbols (functions, classes, variables)",
				"get_file_content - Get full content of specific files",
				"list_directory - List files and directories",
				"delete_lines - Delete a range of lines from a file",
				"insert_at_line - Insert content at a specific line",
				"replace_lines - Replace a range of lines with new content",
			},
			"ai_tools": []string{
				"generate_code - Generate code from natural language",
				"analyze_code - Analyze code quality and get suggestions",
				"explain_code - Get AI explanation of code functionality",
			},
			"project_tools": []string{
				"get_current_config - Get current configuration and status",
				"initial_instructions - Get these initial instructions",
				"remove_project - Remove a project from configuration",
				"restart_language_server - Restart the language server",
				"summarize_changes - Get instructions for summarizing changes",
			},
		},
		"tips": []string{
			"Use natural language to describe what you want to accomplish",
			"The AI tools can help you understand and improve your code",
			"File manipulation tools allow direct editing of your codebase",
			"Project tools help manage your development environment",
		},
		"examples": []string{
			"'Index my repository at /path/to/project'",
			"'Find all Go test files'",
			"'Generate a HTTP handler function in Go'",
			"'Analyze this function for performance issues'",
			"'Delete lines 10-20 from main.go'",
			"'Insert a new function at line 50 in utils.go'",
		},
	}

	content, err := json.MarshalIndent(instructions, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format instructions"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleRemoveProject handles project removal requests
func (s *MCPServer) handleRemoveProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling remove project", zap.String("tool", request.Params.Name))

	projectName, err := request.RequireString("project_name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_name parameter: %v", err)), nil
	}

	// Check if project exists in repositories
	repositories, err := s.searcher.ListRepositories(ctx)
	if err != nil {
		s.logger.Error("Failed to list repositories", zap.Error(err))
		return mcp.NewToolResultError("Failed to access repository list"), nil
	}

	projectFound := false
	for _, repo := range repositories {
		if repo.Name == projectName {
			projectFound = true
			break
		}
	}

	if !projectFound {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found in indexed repositories", projectName)), nil
	}

	// Note: In a real implementation, you would remove the project from the index
	// For now, we'll simulate the removal
	result := map[string]interface{}{
		"success":      true,
		"project_name": projectName,
		"message":      fmt.Sprintf("Project '%s' would be removed from configuration", projectName),
		"note":         "This is a simulated removal. In production, this would remove the project from the search index and configuration.",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	s.logger.Info("Project removal requested", zap.String("project", projectName))

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleRestartLanguageServer handles language server restart requests
func (s *MCPServer) handleRestartLanguageServer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling restart language server", zap.String("tool", request.Params.Name))

	// In a real implementation, this would restart the language server
	// For now, we'll simulate the restart and provide useful information
	result := map[string]interface{}{
		"success": true,
		"message": "Language server restart initiated",
		"details": map[string]interface{}{
			"reason":     "External file changes detected or manual restart requested",
			"action":     "Simulated restart - in production this would restart the Go language server",
			"timestamp":  time.Now().Format(time.RFC3339),
			"suggestion": "If you're experiencing issues with code completion or analysis, this restart should resolve them",
		},
		"next_steps": []string{
			"Wait a few seconds for the language server to fully restart",
			"Try using code completion or analysis features",
			"If issues persist, check the language server logs",
		},
	}

	s.logger.Info("Language server restart simulated")

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleSummarizeChanges handles change summarization requests
func (s *MCPServer) handleSummarizeChanges(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling summarize changes", zap.String("tool", request.Params.Name))

	instructions := map[string]interface{}{
		"title": "Codebase Change Summarization Instructions",
		"description": "Guidelines for effectively summarizing changes made to the codebase",
		"summarization_framework": map[string]interface{}{
			"structure": []string{
				"1. **Overview** - Brief description of what was changed and why",
				"2. **Files Modified** - List of files that were added, modified, or deleted",
				"3. **Key Changes** - Detailed breakdown of significant modifications",
				"4. **Impact Analysis** - How these changes affect the system",
				"5. **Testing** - What testing was done or is recommended",
				"6. **Next Steps** - Any follow-up actions required",
			},
			"categories": []string{
				"🆕 **New Features** - Added functionality",
				"🐛 **Bug Fixes** - Resolved issues",
				"♻️ **Refactoring** - Code structure improvements",
				"📚 **Documentation** - Updated docs or comments",
				"🔧 **Configuration** - Settings or build changes",
				"🧪 **Tests** - Added or modified tests",
				"🚀 **Performance** - Optimization improvements",
				"🔒 **Security** - Security-related changes",
			},
		},
		"best_practices": []string{
			"Use clear, concise language that non-technical stakeholders can understand",
			"Include specific file names and line numbers when relevant",
			"Explain the business value or technical benefit of each change",
			"Mention any breaking changes or migration requirements",
			"Include before/after code snippets for complex changes",
			"Reference any related issues, tickets, or requirements",
			"Highlight any new dependencies or external changes",
		},
		"example_summary": map[string]interface{}{
			"overview": "Added new file manipulation tools to the MCP Code Indexer to enable direct file editing capabilities",
			"files_modified": []string{
				"internal/server/handlers_utility.go - Added 3 new file manipulation handlers",
				"internal/server/handlers_project.go - Created new file with 5 project management tools",
				"internal/server/tools.go - Updated tool registration for 8 new tools",
			},
			"key_changes": []string{
				"Implemented delete_lines, insert_at_line, and replace_lines tools for direct file editing",
				"Added project management tools for configuration and environment management",
				"Expanded total tool count from 12 to 20 tools",
				"Maintained modular architecture and error handling patterns",
			},
			"impact": "Users can now directly edit files and manage projects through the MCP interface, significantly expanding the tool's capabilities",
		},
		"tools_for_analysis": []string{
			"Use 'search_code' to find recent changes in the codebase",
			"Use 'get_file_content' to examine specific files that were modified",
			"Use 'list_repositories' to see which projects were affected",
			"Use 'get_index_stats' to understand the scope of changes",
		},
	}

	content, err := json.MarshalIndent(instructions, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format instructions"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}
