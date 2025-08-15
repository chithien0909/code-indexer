package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/connection"
	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/internal/locking"
	"github.com/my-mcp/code-indexer/internal/models"
	"github.com/my-mcp/code-indexer/internal/repository"
	"github.com/my-mcp/code-indexer/internal/search"
	"github.com/my-mcp/code-indexer/internal/session"
)

// MCPServer wraps the MCP server with our application logic
type MCPServer struct {
	server            *server.MCPServer
	config            *config.Config
	logger            *zap.Logger
	indexer           *indexer.Indexer
	repoMgr           *repository.Manager
	searcher          *search.Engine
	modelsEngine      *models.Engine
	sessionManager    *session.Manager
	sessionContext    *session.SessionContext
	connectionManager *connection.Manager
	lockManager       *locking.Manager
	mutex             sync.RWMutex
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

	// Create session manager if multi-session is enabled
	var sessionManager *session.Manager
	var sessionContext *session.SessionContext

	if cfg.Server.MultiSession.Enabled {
		sessionManager = session.NewManager(cfg, logger)
		sessionContext = session.NewSessionContext(sessionManager)
		logger.Info("Multi-session support enabled",
			zap.Int("max_sessions", cfg.Server.MultiSession.MaxSessions),
			zap.Bool("isolate_workspaces", cfg.Server.MultiSession.IsolateWorkspaces))
	}

	// Create connection manager if multi-IDE is enabled
	var connectionManager *connection.Manager
	if cfg.Server.MultiIDE.Enabled {
		connectionManager = connection.NewManager(cfg, sessionManager, logger)
		logger.Info("Multi-IDE support enabled",
			zap.Int("max_connections", cfg.Server.MultiIDE.MaxConnections),
			zap.String("isolation_mode", cfg.Server.MultiIDE.ResourceManagement.IsolationMode))
	}

	// Create lock manager if fine-grained locking is enabled
	var lockManager *locking.Manager
	if cfg.Server.MultiIDE.Enabled && cfg.Server.MultiIDE.Locking.EnableFineGrainedLocks {
		lockConfig := &locking.LockConfig{
			DefaultTimeout:      time.Duration(cfg.Server.MultiIDE.Locking.LockTimeoutSeconds) * time.Second,
			MaxLockDuration:     5 * time.Minute,
			CleanupInterval:     1 * time.Minute,
			EnableDeadlockCheck: cfg.Server.MultiIDE.Locking.EnableDeadlockDetection,
			MaxWaitQueueSize:    100,
		}
		lockManager = locking.NewManager(lockConfig, logger)
		logger.Info("Resource locking enabled",
			zap.Duration("default_timeout", lockConfig.DefaultTimeout),
			zap.Bool("deadlock_detection", lockConfig.EnableDeadlockCheck))
	}

	s := &MCPServer{
		server:            mcpServer,
		config:            cfg,
		logger:            logger,
		indexer:           idx,
		repoMgr:           repoMgr,
		searcher:          searcher,
		modelsEngine:      modelsEngine,
		sessionManager:    sessionManager,
		sessionContext:    sessionContext,
		connectionManager: connectionManager,
		lockManager:       lockManager,
	}

	// Register MCP tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// NewForUVX creates a new MCP server instance optimized for uvx execution
func NewForUVX(cfg *config.Config, logger *zap.Logger) (*MCPServer, error) {
	// Create MCP server with uvx-optimized configuration
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

	// Use relative paths that work better with uvx execution
	repoDir := cfg.Indexer.RepoDir
	if repoDir == "" {
		repoDir = "./repositories"
	}

	indexDir := cfg.Indexer.IndexDir
	if indexDir == "" {
		indexDir = "./index"
	}

	// Initialize components with uvx-friendly paths
	repoMgr, err := repository.NewManager(repoDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository manager: %w", err)
	}

	searcher, err := search.NewEngine(indexDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}

	idx, err := indexer.New(cfg, repoMgr, searcher, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	// Initialize models engine with safe defaults for uvx mode
	// Force disable models for uvx to avoid initialization issues
	cfg.Models.Enabled = false
	modelsEngine, err := models.NewEngine(&cfg.Models, idx, logger)
	if err != nil {
		// In uvx mode, if models engine fails to initialize, create a disabled one
		logger.Warn("Failed to create models engine, creating disabled instance", zap.Error(err))
		disabledConfig := &config.ModelsConfig{Enabled: false}
		modelsEngine, _ = models.NewEngine(disabledConfig, idx, logger)
	}

	// For uvx mode, disable multi-session and multi-IDE features for simplicity
	// Each uvx process is isolated anyway
	var sessionManager *session.Manager
	var sessionContext *session.SessionContext
	var connectionManager *connection.Manager
	var lockManager *locking.Manager

	logger.Debug("UVX mode: Multi-session and multi-IDE features disabled for process isolation")

	s := &MCPServer{
		server:            mcpServer,
		config:            cfg,
		logger:            logger,
		indexer:           idx,
		repoMgr:           repoMgr,
		searcher:          searcher,
		modelsEngine:      modelsEngine,
		sessionManager:    sessionManager,
		sessionContext:    sessionContext,
		connectionManager: connectionManager,
		lockManager:       lockManager,
	}

	// Register MCP tools
	logger.Debug("Registering MCP tools...")
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}
	logger.Debug("MCP tools registered successfully")

	return s, nil
}

// ServeStdio starts the MCP server using stdio transport (uvx-optimized)
func (s *MCPServer) ServeStdio() error {
	s.logger.Debug("Starting MCP server (stdio mode)",
		zap.String("name", s.config.Server.Name),
		zap.String("version", s.config.Server.Version))
	return server.ServeStdio(s.server)
}

// Serve starts the MCP server using stdio transport
func (s *MCPServer) Serve() error {
	s.logger.Info("Starting MCP server",
		zap.String("name", s.config.Server.Name),
		zap.String("version", s.config.Server.Version))
	return server.ServeStdio(s.server)
}

// ServeDaemon starts the MCP server as a daemon listening on TCP port
func (s *MCPServer) ServeDaemon(host string, port int) error {
	s.logger.Info("Starting MCP daemon server",
		zap.String("name", s.config.Server.Name),
		zap.String("version", s.config.Server.Version),
		zap.String("host", host),
		zap.Int("port", port))

	// Create HTTP server for handling MCP connections
	mux := http.NewServeMux()

	// Handle MCP API endpoints
	mux.HandleFunc("/api/tools", s.handleToolsAPI)
	mux.HandleFunc("/api/call", s.handleToolCall)
	mux.HandleFunc("/api/health", s.handleHealthCheck)
	mux.HandleFunc("/api/sessions", s.handleSessionsAPI)

	// Create HTTP server
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.logger.Info("MCP daemon listening", zap.String("address", addr))

	return httpServer.ListenAndServe()
}

// Close gracefully shuts down the server
func (s *MCPServer) Close() error {
	s.logger.Info("Shutting down MCP server")

	// Close connection manager if enabled
	if s.connectionManager != nil {
		if err := s.connectionManager.Close(); err != nil {
			s.logger.Error("Failed to close connection manager", zap.Error(err))
		}
	}

	// Close lock manager if enabled
	if s.lockManager != nil {
		if err := s.lockManager.Close(); err != nil {
			s.logger.Error("Failed to close lock manager", zap.Error(err))
		}
	}

	// Close session manager if enabled
	if s.sessionManager != nil {
		s.sessionManager.Close()
	}

	if err := s.searcher.Close(); err != nil {
		s.logger.Error("Failed to close search engine", zap.Error(err))
	}

	if err := s.modelsEngine.Close(); err != nil {
		s.logger.Error("Failed to close models engine", zap.Error(err))
	}

	return nil
}

// HTTP API handlers for daemon mode

// handleToolsAPI handles the /api/tools endpoint - lists all available tools
func (s *MCPServer) handleToolsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all available tools
	tools := []map[string]interface{}{
		// Core tools
		{"name": "index_repository", "category": "core", "description": "Index a Git repository for searching"},
		{"name": "search_code", "category": "core", "description": "Search across all indexed repositories"},
		{"name": "get_metadata", "category": "core", "description": "Get detailed metadata for specific files"},
		{"name": "list_repositories", "category": "core", "description": "List all indexed repositories with statistics"},
		{"name": "get_index_stats", "category": "core", "description": "Get indexing statistics and information"},

		// Utility tools
		{"name": "find_files", "category": "utility", "description": "Find files matching patterns with wildcards"},
		{"name": "find_symbols", "category": "utility", "description": "Find symbols (functions, classes, variables) by name"},
		{"name": "get_file_content", "category": "utility", "description": "Get full content of specific files with line ranges"},
		{"name": "list_directory", "category": "utility", "description": "List files and directories in specific paths"},
		{"name": "delete_lines", "category": "utility", "description": "Delete a range of lines within a file"},
		{"name": "insert_at_line", "category": "utility", "description": "Insert content at a given line in a file"},
		{"name": "replace_lines", "category": "utility", "description": "Replace a range of lines with new content"},
		{"name": "get_file_snippet", "category": "utility", "description": "Extract a specific code snippet from a file"},
		{"name": "find_references", "category": "utility", "description": "Find all references to a symbol across indexed repositories"},
		{"name": "refresh_index", "category": "utility", "description": "Refresh the search index for specific repositories or all repositories"},
		{"name": "git_blame", "category": "utility", "description": "Get Git blame information for a specific file or file range"},

		// Project management tools
		{"name": "get_current_config", "category": "project", "description": "Get the current configuration of the agent"},
		{"name": "initial_instructions", "category": "project", "description": "Get the initial instructions for the current project"},
		{"name": "remove_project", "category": "project", "description": "Remove a project from the configuration"},
		{"name": "restart_language_server", "category": "project", "description": "Restart the language server"},
		{"name": "summarize_changes", "category": "project", "description": "Provide instructions for summarizing codebase changes"},

		// AI tools
		{"name": "generate_code", "category": "ai", "description": "Generate code from natural language descriptions using AI"},
		{"name": "analyze_code", "category": "ai", "description": "Analyze code quality and get AI suggestions"},
		{"name": "explain_code", "category": "ai", "description": "Get AI explanations of code functionality"},
	}

	// Add session management tools if enabled
	if s.config.Server.MultiSession.Enabled {
		sessionTools := []map[string]interface{}{
			{"name": "list_sessions", "category": "session", "description": "List all active VSCode IDE sessions"},
			{"name": "create_session", "category": "session", "description": "Create a new VSCode IDE session"},
			{"name": "get_session_info", "category": "session", "description": "Get information about the current session"},
		}
		tools = append(tools, sessionTools...)
	}

	response := map[string]interface{}{
		"tools": tools,
		"total": len(tools),
		"categories": map[string]int{
			"core":    5,
			"utility": 11,
			"project": 5,
			"session": func() int {
				if s.config.Server.MultiSession.Enabled {
					return 3
				} else {
					return 0
				}
			}(),
			"ai": 3,
		},
		"server_info": map[string]interface{}{
			"name":          s.config.Server.Name,
			"version":       s.config.Server.Version,
			"multi_session": s.config.Server.MultiSession.Enabled,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode tools response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleToolCall handles the /api/call endpoint - executes MCP tool calls
func (s *MCPServer) handleToolCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var requestBody struct {
		Tool      string                 `json:"tool"`
		Arguments map[string]interface{} `json:"arguments"`
		SessionID string                 `json:"session_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create MCP request
	mcpRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      requestBody.Tool,
			Arguments: requestBody.Arguments,
		},
	}

	// Add session information if provided
	if requestBody.SessionID != "" {
		if mcpRequest.Params.Arguments == nil {
			mcpRequest.Params.Arguments = make(map[string]interface{})
		}
		mcpRequest.Params.Arguments.(map[string]interface{})["session_id"] = requestBody.SessionID
	}

	s.logger.Info("API tool call",
		zap.String("tool", requestBody.Tool),
		zap.String("session_id", requestBody.SessionID),
		zap.String("remote_addr", r.RemoteAddr))

	// Execute the tool call
	ctx := context.Background()
	result, err := s.executeToolCall(ctx, mcpRequest)
	if err != nil {
		s.logger.Error("Tool call failed", zap.Error(err))
		http.Error(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert MCP result to API response
	response := map[string]interface{}{
		"success": true,
		"tool":    requestBody.Tool,
		"result":  result,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode tool call response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleHealthCheck handles the /api/health endpoint
func (s *MCPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   s.config.Server.Version,
		"uptime":    time.Since(time.Now()).String(), // This would be calculated from server start time
	}

	if s.sessionManager != nil {
		health["sessions"] = s.sessionManager.GetSessionStats()
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		s.logger.Error("Failed to encode health response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleSessionsAPI handles the /api/sessions endpoint
func (s *MCPServer) handleSessionsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.sessionManager == nil {
		http.Error(w, "Multi-session support not enabled", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case "GET":
		// List sessions
		sessions := s.sessionManager.ListSessions()
		stats := s.sessionManager.GetSessionStats()

		response := map[string]interface{}{
			"sessions": sessions,
			"stats":    stats,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode sessions response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	case "POST":
		// Create new session
		var requestBody struct {
			Name         string `json:"name"`
			WorkspaceDir string `json:"workspace_dir,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		session, err := s.sessionManager.CreateSession(requestBody.Name, requestBody.WorkspaceDir)
		if err != nil {
			s.logger.Error("Failed to create session", zap.Error(err))
			http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"session": session,
			"message": fmt.Sprintf("Session '%s' created successfully", requestBody.Name),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode create session response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// executeToolCall executes an MCP tool call and returns the result
func (s *MCPServer) executeToolCall(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
	// This is a simplified version - in a real implementation, you'd route to the appropriate handler
	// For now, we'll handle a few key tools directly

	switch request.Params.Name {
	case "list_repositories":
		return s.handleListRepositories(ctx, request)
	case "get_index_stats":
		return s.handleGetIndexStats(ctx, request)
	case "search_code":
		return s.handleSearchCode(ctx, request)
	case "find_files":
		return s.handleFindFiles(ctx, request)
	case "get_file_content":
		return s.handleGetFileContent(ctx, request)
	case "list_sessions":
		if s.sessionManager != nil {
			sessions := s.sessionManager.ListSessions()
			stats := s.sessionManager.GetSessionStats()
			return map[string]interface{}{
				"sessions": sessions,
				"stats":    stats,
			}, nil
		}
		return nil, fmt.Errorf("multi-session support not enabled")
	case "get_session_info":
		return map[string]interface{}{
			"multi_session_enabled": s.config.Server.MultiSession.Enabled,
			"session_config":        s.config.Server.MultiSession,
		}, nil
	default:
		return nil, fmt.Errorf("tool not supported in API mode: %s", request.Params.Name)
	}
}
