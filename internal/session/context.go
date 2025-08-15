package session

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/my-mcp/code-indexer/internal/config"
)

// ContextKey is the type for session context keys
type ContextKey string

const (
	// SessionIDKey is the context key for session ID
	SessionIDKey ContextKey = "session_id"
	// SessionKey is the context key for session object
	SessionKey ContextKey = "session"
	// WorkspaceKey is the context key for workspace directory
	WorkspaceKey ContextKey = "workspace"
)

// SessionContext provides session-aware context operations
type SessionContext struct {
	manager *Manager
}

// NewSessionContext creates a new session context handler
func NewSessionContext(manager *Manager) *SessionContext {
	return &SessionContext{
		manager: manager,
	}
}

// ExtractSessionFromRequest extracts session information from MCP request
func (sc *SessionContext) ExtractSessionFromRequest(request mcp.CallToolRequest) (string, string, string, error) {
	args := sc.getArguments(request)
	
	// Try to extract session information from various sources
	sessionID := sc.getStringValue(args, "session_id", "")
	workspaceDir := sc.getStringValue(args, "workspace_dir", "")
	sessionName := sc.getStringValue(args, "session_name", "default")

	// If no explicit session info, try to infer from file paths
	if sessionID == "" && workspaceDir == "" {
		if filePath := sc.getStringValue(args, "file_path", ""); filePath != "" {
			workspaceDir = sc.inferWorkspaceFromPath(filePath)
		}
	}

	return sessionID, sessionName, workspaceDir, nil
}

// CreateSessionAwareContext creates a context with session information
func (sc *SessionContext) CreateSessionAwareContext(ctx context.Context, request mcp.CallToolRequest) (context.Context, *Session, error) {
	sessionID, sessionName, workspaceDir, err := sc.ExtractSessionFromRequest(request)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to extract session info: %w", err)
	}

	// Get or create session
	session, err := sc.manager.GetOrCreateSession(sessionID, sessionName, workspaceDir)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to get/create session: %w", err)
	}

	// Add session information to context
	ctx = context.WithValue(ctx, SessionIDKey, session.ID)
	ctx = context.WithValue(ctx, SessionKey, session)
	ctx = context.WithValue(ctx, WorkspaceKey, session.WorkspaceDir)

	return ctx, session, nil
}

// GetSessionFromContext retrieves session from context
func (sc *SessionContext) GetSessionFromContext(ctx context.Context) (*Session, error) {
	session, ok := ctx.Value(SessionKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("session not found in context")
	}
	return session, nil
}

// GetSessionIDFromContext retrieves session ID from context
func (sc *SessionContext) GetSessionIDFromContext(ctx context.Context) (string, error) {
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	if !ok {
		return "", fmt.Errorf("session ID not found in context")
	}
	return sessionID, nil
}

// GetWorkspaceFromContext retrieves workspace directory from context
func (sc *SessionContext) GetWorkspaceFromContext(ctx context.Context) (string, error) {
	workspace, ok := ctx.Value(WorkspaceKey).(string)
	if !ok {
		return "", fmt.Errorf("workspace not found in context")
	}
	return workspace, nil
}

// AddSessionInfoToResponse adds session information to tool response
func (sc *SessionContext) AddSessionInfoToResponse(response map[string]interface{}, session *Session) {
	response["session_info"] = map[string]interface{}{
		"session_id":    session.ID,
		"session_name":  session.Name,
		"workspace_dir": session.WorkspaceDir,
		"created_at":    session.CreatedAt,
		"last_access":   session.LastAccess,
	}
}

// ResolveSessionPath resolves a file path relative to session workspace
func (sc *SessionContext) ResolveSessionPath(session *Session, filePath string) string {
	if session.WorkspaceDir == "" {
		return filePath
	}

	// If path is already absolute, return as-is
	if len(filePath) > 0 && (filePath[0] == '/' || (len(filePath) > 1 && filePath[1] == ':')) {
		return filePath
	}

	// Resolve relative to workspace
	return fmt.Sprintf("%s/%s", session.WorkspaceDir, filePath)
}

// ValidateSessionAccess validates that a session can access a resource
func (sc *SessionContext) ValidateSessionAccess(session *Session, resourcePath string) error {
	// For now, allow all access within workspace
	// In the future, this could implement more sophisticated access controls
	if session.WorkspaceDir != "" {
		// Could add path validation here
	}
	return nil
}

// Helper methods

func (sc *SessionContext) getArguments(request mcp.CallToolRequest) map[string]interface{} {
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}
	return make(map[string]interface{})
}

func (sc *SessionContext) getStringValue(args map[string]interface{}, key, defaultValue string) string {
	if value, exists := args[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (sc *SessionContext) inferWorkspaceFromPath(filePath string) string {
	// Simple heuristic: try to find common workspace indicators
	// This could be enhanced with more sophisticated detection
	
	// Look for common project root indicators
	_ = []string{
		".git", ".vscode", "package.json", "go.mod", "Cargo.toml",
		"pom.xml", "build.gradle", "requirements.txt", "setup.py",
	}
	
	// For now, return the directory containing the file
	// In a real implementation, you'd walk up the directory tree
	// looking for workspace indicators
	
	if filePath == "" {
		return ""
	}
	
	// Extract directory from file path
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			return filePath[:i]
		}
	}
	
	return "."
}

// SessionAwareRequest wraps an MCP request with session context
type SessionAwareRequest struct {
	Request   mcp.CallToolRequest
	Session   *Session
	Context   context.Context
}

// NewSessionAwareRequest creates a new session-aware request
func (sc *SessionContext) NewSessionAwareRequest(ctx context.Context, request mcp.CallToolRequest) (*SessionAwareRequest, error) {
	sessionCtx, session, err := sc.CreateSessionAwareContext(ctx, request)
	if err != nil {
		return nil, err
	}

	return &SessionAwareRequest{
		Request: request,
		Session: session,
		Context: sessionCtx,
	}, nil
}

// GetSessionConfig returns the configuration for the session
func (sar *SessionAwareRequest) GetSessionConfig() *config.Config {
	return sar.Session.Config
}

// GetSessionWorkspace returns the workspace directory for the session
func (sar *SessionAwareRequest) GetSessionWorkspace() string {
	return sar.Session.WorkspaceDir
}

// ResolvePath resolves a file path relative to the session workspace
func (sar *SessionAwareRequest) ResolvePath(filePath string) string {
	if sar.Session.WorkspaceDir == "" {
		return filePath
	}

	// If path is already absolute, return as-is
	if len(filePath) > 0 && (filePath[0] == '/' || (len(filePath) > 1 && filePath[1] == ':')) {
		return filePath
	}

	// Resolve relative to workspace
	return fmt.Sprintf("%s/%s", sar.Session.WorkspaceDir, filePath)
}
