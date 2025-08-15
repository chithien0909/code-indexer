package server

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// registerTools registers all MCP tools
func (s *MCPServer) registerTools() error {
	s.logger.Info("🔧 Starting tool registration process...")

	// Register core indexing tools
	s.logger.Info("📦 Registering core tools...")
	if err := s.registerCoreTools(); err != nil {
		s.logger.Error("❌ Failed to register core tools", zap.Error(err))
		return fmt.Errorf("failed to register core tools: %w", err)
	}
	s.logger.Info("✅ Core tools registered successfully", zap.Int("count", 5))

	// Register utility tools
	s.logger.Info("🛠️ Registering utility tools...")
	if err := s.registerUtilityTools(); err != nil {
		s.logger.Error("❌ Failed to register utility tools", zap.Error(err))
		return fmt.Errorf("failed to register utility tools: %w", err)
	}
	s.logger.Info("✅ Utility tools registered successfully", zap.Int("count", 11))

	// Register project management tools
	s.logger.Info("📋 Registering project management tools...")
	if err := s.registerProjectTools(); err != nil {
		s.logger.Error("❌ Failed to register project tools", zap.Error(err))
		return fmt.Errorf("failed to register project tools: %w", err)
	}
	s.logger.Info("✅ Project management tools registered successfully", zap.Int("count", 5))

	// Register session management tools if multi-session is enabled
	if s.config.Server.MultiSession.Enabled {
		s.logger.Info("👥 Registering session management tools...")
		if err := s.registerSessionTools(); err != nil {
			s.logger.Error("❌ Failed to register session tools", zap.Error(err))
			return fmt.Errorf("failed to register session tools: %w", err)
		}
		s.logger.Info("✅ Session management tools registered successfully", zap.Int("count", 3))
	} else {
		s.logger.Info("👥 Session management tools disabled")
	}

	// Register AI model tools if enabled
	if s.config.Models.Enabled {
		s.logger.Info("🤖 Registering AI model tools...")
		if err := s.registerModelTools(); err != nil {
			s.logger.Error("❌ Failed to register AI model tools", zap.Error(err))
			return fmt.Errorf("failed to register model tools: %w", err)
		}
		s.logger.Info("✅ AI model tools registered successfully", zap.Int("count", 3))
	} else {
		s.logger.Info("🤖 AI model tools disabled")
		if err := s.registerModelTools(); err != nil {
			return fmt.Errorf("failed to register model tools: %w", err)
		}
	}

	// Log detailed tool summary like daemon mode
	s.logToolsSummary()
	s.logger.Info("🎉 All MCP tools registered successfully")
	return nil
}

// logToolsSummary logs a detailed summary of all registered tools
func (s *MCPServer) logToolsSummary() {
	// Count tools by category
	categories := map[string]int{
		"core":    5,
		"utility": 11,
		"project": 5,
		"ai":      0, // Will be 3 if models enabled
		"session": 0, // Will be 3 if multi-session enabled
	}

	// Adjust counts based on enabled features
	if s.config.Models.Enabled {
		categories["ai"] = 3
	}
	if s.config.Server.MultiSession.Enabled {
		categories["session"] = 3
	}

	// Calculate total
	total := 0
	for _, count := range categories {
		total += count
	}

	// Create tools list for detailed logging
	tools := []map[string]string{
		// Core tools
		{"category": "core", "name": "index_repository", "description": "Index a Git repository for searching"},
		{"category": "core", "name": "search_code", "description": "Search across all indexed repositories"},
		{"category": "core", "name": "get_metadata", "description": "Get detailed metadata for specific files"},
		{"category": "core", "name": "list_repositories", "description": "List all indexed repositories with statistics"},
		{"category": "core", "name": "get_index_stats", "description": "Get indexing statistics and information"},

		// Utility tools
		{"category": "utility", "name": "find_files", "description": "Find files matching patterns with wildcards"},
		{"category": "utility", "name": "find_symbols", "description": "Find symbols (functions, classes, variables) by name"},
		{"category": "utility", "name": "get_file_content", "description": "Get full content of specific files with line ranges"},
		{"category": "utility", "name": "list_directory", "description": "List files and directories in specific paths"},
		{"category": "utility", "name": "delete_lines", "description": "Delete a range of lines within a file"},
		{"category": "utility", "name": "insert_at_line", "description": "Insert content at a given line in a file"},
		{"category": "utility", "name": "replace_lines", "description": "Replace a range of lines with new content"},
		{"category": "utility", "name": "get_file_snippet", "description": "Extract a specific code snippet from a file"},
		{"category": "utility", "name": "find_references", "description": "Find all references to a symbol across indexed repositories"},
		{"category": "utility", "name": "refresh_index", "description": "Refresh the search index for specific repositories or all repositories"},
		{"category": "utility", "name": "git_blame", "description": "Get Git blame information for a specific file or file range"},

		// Project tools
		{"category": "project", "name": "get_current_config", "description": "Get the current configuration of the agent"},
		{"category": "project", "name": "initial_instructions", "description": "Get the initial instructions for the current project"},
		{"category": "project", "name": "remove_project", "description": "Remove a project from the configuration"},
		{"category": "project", "name": "restart_language_server", "description": "Restart the language server"},
		{"category": "project", "name": "summarize_changes", "description": "Provide instructions for summarizing codebase changes"},
	}

	// Add AI tools if enabled
	if s.config.Models.Enabled {
		aiTools := []map[string]string{
			{"category": "ai", "name": "generate_code", "description": "Generate code from natural language descriptions using AI"},
			{"category": "ai", "name": "analyze_code", "description": "Analyze code quality and get AI suggestions"},
			{"category": "ai", "name": "explain_code", "description": "Get AI explanations of code functionality"},
		}
		tools = append(tools, aiTools...)
	}

	// Add session tools if enabled
	if s.config.Server.MultiSession.Enabled {
		sessionTools := []map[string]string{
			{"category": "session", "name": "list_sessions", "description": "List all active VSCode IDE sessions"},
			{"category": "session", "name": "create_session", "description": "Create a new VSCode IDE session"},
			{"category": "session", "name": "get_session_info", "description": "Get information about the current session"},
		}
		tools = append(tools, sessionTools...)
	}

	// Log the summary in detailed format like Serena
	s.logger.Info("📊 MCP Tools Summary",
		zap.Any("categories", categories),
		zap.Any("server_info", map[string]interface{}{
			"name":          s.config.Server.Name,
			"version":       s.config.Server.Version,
			"multi_session": s.config.Server.MultiSession.Enabled,
			"models":        s.config.Models.Enabled,
		}),
		zap.Any("tools", tools),
		zap.Int("total", total))

	// Log individual tool categories for better visibility
	s.logger.Info("📦 Core Tools Available", zap.Int("count", categories["core"]))
	s.logger.Info("🛠️ Utility Tools Available", zap.Int("count", categories["utility"]))
	s.logger.Info("📋 Project Tools Available", zap.Int("count", categories["project"]))
	if categories["ai"] > 0 {
		s.logger.Info("🤖 AI Tools Available", zap.Int("count", categories["ai"]))
	}
	if categories["session"] > 0 {
		s.logger.Info("👥 Session Tools Available", zap.Int("count", categories["session"]))
	}

	s.logger.Info("🎯 Total Tools Available", zap.Int("total", total))
}

// registerCoreTools registers core indexing and search tools
func (s *MCPServer) registerCoreTools() error {
	s.logger.Info("Registering core tools...")

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
	// Use session-aware handler if multi-session is enabled
	if s.config.Server.MultiSession.Enabled {
		s.server.AddTool(indexRepoTool, s.wrapWithSession(s.handleIndexRepositorySession))
	} else {
		s.server.AddTool(indexRepoTool, s.handleIndexRepository)
	}
	s.logger.Debug("Registered tool: index_repository")

	// Search Code Tool
	searchCodeTool := mcp.NewTool("search_code",
		mcp.WithDescription("Search across all indexed repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithString("type",
			mcp.Description("Search type: function, class, variable, content, file, comment"),
		),
		mcp.WithString("language",
			mcp.Description("Filter by programming language"),
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

	s.logger.Info("Core tools registered successfully", zap.Int("tool_count", 5))
	return nil
}

// registerUtilityTools registers utility tools for file operations
func (s *MCPServer) registerUtilityTools() error {
	s.logger.Info("Registering utility tools...")

	// Find Files Tool
	findFilesTool := mcp.NewTool("find_files",
		mcp.WithDescription("Find files matching patterns in indexed repositories"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("File name pattern (supports wildcards like *.go, *test*, etc.)"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name to search in (optional, searches all if not specified)"),
		),
		mcp.WithBoolean("include_content",
			mcp.Description("Include file content preview in results"),
		),
	)
	s.server.AddTool(findFilesTool, s.handleFindFiles)

	// Find Symbols Tool
	findSymbolsTool := mcp.NewTool("find_symbols",
		mcp.WithDescription("Find symbols (functions, classes, variables) by name"),
		mcp.WithString("symbol_name",
			mcp.Required(),
			mcp.Description("Symbol name or pattern to search for"),
		),
		mcp.WithString("symbol_type",
			mcp.Description("Type of symbol: function, class, variable, constant, interface"),
		),
		mcp.WithString("language",
			mcp.Description("Programming language to filter by"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name to search in (optional)"),
		),
	)
	s.server.AddTool(findSymbolsTool, s.handleFindSymbols)

	// Get File Content Tool
	getFileContentTool := mcp.NewTool("get_file_content",
		mcp.WithDescription("Get the full content of a specific file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name (optional)"),
		),
		mcp.WithNumber("start_line",
			mcp.Description("Start line number (optional, 1-based)"),
		),
		mcp.WithNumber("end_line",
			mcp.Description("End line number (optional, 1-based)"),
		),
	)
	s.server.AddTool(getFileContentTool, s.handleGetFileContent)

	// List Directory Tool
	listDirectoryTool := mcp.NewTool("list_directory",
		mcp.WithDescription("List files and directories in a specific path"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Directory path to list"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name (optional)"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("List recursively (default: false)"),
		),
		mcp.WithString("file_filter",
			mcp.Description("File extension filter (e.g., '.go', '.py')"),
		),
	)
	s.server.AddTool(listDirectoryTool, s.handleListDirectory)

	// File Manipulation Tools

	// Delete Lines Tool
	deleteLinesTool := mcp.NewTool("delete_lines",
		mcp.WithDescription("Delete a range of lines within a file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("start_line",
			mcp.Required(),
			mcp.Description("Start line number (1-based, inclusive)"),
		),
		mcp.WithNumber("end_line",
			mcp.Required(),
			mcp.Description("End line number (1-based, inclusive)"),
		),
	)
	s.server.AddTool(deleteLinesTool, s.handleDeleteLines)

	// Insert At Line Tool
	insertAtLineTool := mcp.NewTool("insert_at_line",
		mcp.WithDescription("Insert content at a given line in a file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("line_number",
			mcp.Required(),
			mcp.Description("Line number where to insert content (1-based)"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to insert (supports multi-line content)"),
		),
	)
	s.server.AddTool(insertAtLineTool, s.handleInsertAtLine)

	// Replace Lines Tool
	replaceLinesTool := mcp.NewTool("replace_lines",
		mcp.WithDescription("Replace a range of lines within a file with new content"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("start_line",
			mcp.Required(),
			mcp.Description("Start line number (1-based, inclusive)"),
		),
		mcp.WithNumber("end_line",
			mcp.Required(),
			mcp.Description("End line number (1-based, inclusive)"),
		),
		mcp.WithString("new_content",
			mcp.Required(),
			mcp.Description("New content to replace the lines (supports multi-line content)"),
		),
	)
	s.server.AddTool(replaceLinesTool, s.handleReplaceLines)

	// Advanced Utility Tools

	// Get File Snippet Tool
	getFileSnippetTool := mcp.NewTool("get_file_snippet",
		mcp.WithDescription("Extract a specific code snippet from a file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("start_line",
			mcp.Required(),
			mcp.Description("Start line number (1-based, inclusive)"),
		),
		mcp.WithNumber("end_line",
			mcp.Required(),
			mcp.Description("End line number (1-based, inclusive)"),
		),
		mcp.WithBoolean("include_context",
			mcp.Description("Include surrounding context lines"),
		),
	)
	s.server.AddTool(getFileSnippetTool, s.handleGetFileSnippet)

	// Find References Tool
	findReferencesTool := mcp.NewTool("find_references",
		mcp.WithDescription("Find all references to a symbol across indexed repositories"),
		mcp.WithString("symbol_name",
			mcp.Required(),
			mcp.Description("Symbol name to search for"),
		),
		mcp.WithString("symbol_type",
			mcp.Description("Type of symbol: function, class, variable, constant, interface"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name to search in (optional)"),
		),
		mcp.WithBoolean("include_definitions",
			mcp.Description("Include symbol definitions in results (default: true)"),
		),
	)
	s.server.AddTool(findReferencesTool, s.handleFindReferences)

	// Refresh Index Tool
	refreshIndexTool := mcp.NewTool("refresh_index",
		mcp.WithDescription("Refresh the search index for specific repositories or all repositories"),
		mcp.WithString("repository",
			mcp.Description("Repository name to refresh (optional - if not provided, refresh all)"),
		),
		mcp.WithBoolean("force_rebuild",
			mcp.Description("Force complete rebuild of the index"),
		),
	)
	s.server.AddTool(refreshIndexTool, s.handleRefreshIndex)

	// Git Blame Tool
	gitBlameTool := mcp.NewTool("git_blame",
		mcp.WithDescription("Get Git blame information for a specific file or file range"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("start_line",
			mcp.Description("Start line number (optional, 1-based)"),
		),
		mcp.WithNumber("end_line",
			mcp.Description("End line number (optional, 1-based)"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository name (optional)"),
		),
	)
	s.server.AddTool(gitBlameTool, s.handleGitBlame)

	s.logger.Info("Utility tools registered successfully", zap.Int("tool_count", 11))
	return nil
}

// registerProjectTools registers project management tools with the MCP server
func (s *MCPServer) registerProjectTools() error {
	s.logger.Info("Registering project management tools...")

	// Get Current Config Tool
	getCurrentConfigTool := mcp.NewTool("get_current_config",
		mcp.WithDescription("Get the current configuration of the agent, including active projects, tools, contexts, and modes"),
	)
	s.server.AddTool(getCurrentConfigTool, s.handleGetCurrentConfig)

	// Initial Instructions Tool
	initialInstructionsTool := mcp.NewTool("initial_instructions",
		mcp.WithDescription("Get the initial instructions for the current project (for environments where system prompt cannot be set)"),
	)
	s.server.AddTool(initialInstructionsTool, s.handleInitialInstructions)

	// Remove Project Tool
	removeProjectTool := mcp.NewTool("remove_project",
		mcp.WithDescription("Remove a project from the configuration"),
		mcp.WithString("project_name",
			mcp.Required(),
			mcp.Description("Name of the project to remove"),
		),
	)
	s.server.AddTool(removeProjectTool, s.handleRemoveProject)

	// Restart Language Server Tool
	restartLanguageServerTool := mcp.NewTool("restart_language_server",
		mcp.WithDescription("Restart the language server (useful when external edits occur)"),
	)
	s.server.AddTool(restartLanguageServerTool, s.handleRestartLanguageServer)

	// Summarize Changes Tool
	summarizeChangesTool := mcp.NewTool("summarize_changes",
		mcp.WithDescription("Provide instructions for summarizing codebase changes"),
	)
	s.server.AddTool(summarizeChangesTool, s.handleSummarizeChanges)

	s.logger.Info("Project management tools registered successfully", zap.Int("tool_count", 5))
	return nil
}

// registerSessionTools registers session management tools with the MCP server
func (s *MCPServer) registerSessionTools() error {
	s.logger.Info("Registering session management tools...")

	// List Sessions Tool
	listSessionsTool := mcp.NewTool("list_sessions",
		mcp.WithDescription("List all active VSCode IDE sessions"),
	)
	s.server.AddTool(listSessionsTool, s.wrapWithSession(s.handleListSessions))

	// Create Session Tool
	createSessionTool := mcp.NewTool("create_session",
		mcp.WithDescription("Create a new VSCode IDE session"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name for the new session"),
		),
		mcp.WithString("workspace_dir",
			mcp.Description("Workspace directory for the session (optional)"),
		),
	)
	s.server.AddTool(createSessionTool, s.wrapWithSession(s.handleCreateSession))

	// Get Session Info Tool
	getSessionInfoTool := mcp.NewTool("get_session_info",
		mcp.WithDescription("Get information about the current session and multi-session configuration"),
	)
	s.server.AddTool(getSessionInfoTool, s.wrapWithSession(s.handleGetSessionInfo))

	s.logger.Info("Session management tools registered successfully", zap.Int("tool_count", 3))
	return nil
}

// registerModelTools registers AI model tools with the MCP server
func (s *MCPServer) registerModelTools() error {
	if !s.modelsEngine.IsEnabled() {
		s.logger.Info("Models engine disabled, skipping model tool registration")
		return nil
	}

	s.logger.Info("Registering AI model tools...")

	// Register generate_code tool
	generateCodeTool := mcp.NewTool("generate_code",
		mcp.WithDescription("Generate code from natural language description using AI"),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Natural language description of what the code should do"),
		),
		mcp.WithString("language",
			mcp.Required(),
			mcp.Description("Programming language (go, python, javascript, etc.)"),
		),
	)
	s.server.AddTool(generateCodeTool, s.handleGenerateCode)

	// Register analyze_code tool
	analyzeCodeTool := mcp.NewTool("analyze_code",
		mcp.WithDescription("Analyze code quality and get suggestions using AI"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Code to analyze"),
		),
		mcp.WithString("language",
			mcp.Required(),
			mcp.Description("Programming language"),
		),
	)
	s.server.AddTool(analyzeCodeTool, s.handleAnalyzeCode)

	// Register explain_code tool
	explainCodeTool := mcp.NewTool("explain_code",
		mcp.WithDescription("Get AI explanation of code functionality"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Code to explain"),
		),
		mcp.WithString("language",
			mcp.Required(),
			mcp.Description("Programming language"),
		),
	)
	s.server.AddTool(explainCodeTool, s.handleExplainCode)

	s.logger.Info("AI model tools registered successfully", zap.Int("tool_count", 3))
	return nil
}
