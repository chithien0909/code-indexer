package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/session"
)

// SessionAwareHandler wraps tool handlers to provide session isolation
type SessionAwareHandler func(ctx context.Context, request *session.SessionAwareRequest) (*mcp.CallToolResult, error)

// wrapWithSession wraps a session-aware handler to work with MCP
func (s *MCPServer) wrapWithSession(handler SessionAwareHandler) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// If multi-session is not enabled, use legacy behavior
		if !s.config.Server.MultiSession.Enabled || s.sessionContext == nil {
			// Convert to session-aware request with default session
			sessionRequest := &session.SessionAwareRequest{
				Request: request,
				Session: &session.Session{
					ID:           "default",
					Name:         "default",
					WorkspaceDir: "",
					Config:       s.config,
					Context:      make(map[string]interface{}),
					Active:       true,
				},
				Context: ctx,
			}
			return handler(ctx, sessionRequest)
		}

		// Create session-aware request
		sessionRequest, err := s.sessionContext.NewSessionAwareRequest(ctx, request)
		if err != nil {
			s.logger.Error("Failed to create session-aware request", zap.Error(err))
			return mcp.NewToolResultError(fmt.Sprintf("Session error: %v", err)), nil
		}

		// Log session information
		s.logger.Debug("Processing request with session",
			zap.String("tool", request.Params.Name),
			zap.String("session_id", sessionRequest.Session.ID),
			zap.String("workspace", sessionRequest.Session.WorkspaceDir))

		// Call the handler with session context
		result, err := handler(sessionRequest.Context, sessionRequest)
		if err != nil {
			return result, err
		}

		// Add session information to successful responses
		if result != nil && result.Content != nil {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				var responseData map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &responseData); err == nil {
					s.sessionContext.AddSessionInfoToResponse(responseData, sessionRequest.Session)
					if updatedJSON, err := json.MarshalIndent(responseData, "", "  "); err == nil {
						textContent.Text = string(updatedJSON)
					}
				}
			}
		}

		return result, nil
	}
}

// getSessionFromContext is a helper to extract session from context
func (s *MCPServer) getSessionFromContext(ctx context.Context) (*session.Session, error) {
	if s.sessionContext == nil {
		return nil, fmt.Errorf("session context not available")
	}
	return s.sessionContext.GetSessionFromContext(ctx)
}

// resolveSessionPath resolves a file path relative to session workspace
func (s *MCPServer) resolveSessionPath(ctx context.Context, filePath string) string {
	if s.sessionContext == nil {
		return filePath
	}

	session, err := s.sessionContext.GetSessionFromContext(ctx)
	if err != nil {
		return filePath
	}

	return s.sessionContext.ResolveSessionPath(session, filePath)
}

// getSessionConfig returns the configuration for the current session
func (s *MCPServer) getSessionConfig(ctx context.Context) *config.Config {
	if s.sessionContext == nil {
		return s.config
	}

	session, err := s.sessionContext.GetSessionFromContext(ctx)
	if err != nil {
		return s.config
	}

	return session.Config
}

// validateSessionAccess validates that the current session can access a resource
func (s *MCPServer) validateSessionAccess(ctx context.Context, resourcePath string) error {
	if s.sessionContext == nil {
		return nil // No session isolation
	}

	session, err := s.sessionContext.GetSessionFromContext(ctx)
	if err != nil {
		return err
	}

	return s.sessionContext.ValidateSessionAccess(session, resourcePath)
}

// Session management tool handlers

// handleListSessions handles session listing requests
func (s *MCPServer) handleListSessions(ctx context.Context, request *session.SessionAwareRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling list sessions", zap.String("tool", request.Request.Params.Name))

	if s.sessionManager == nil {
		return mcp.NewToolResultError("Multi-session support not enabled"), nil
	}

	sessions := s.sessionManager.ListSessions()
	stats := s.sessionManager.GetSessionStats()

	result := map[string]interface{}{
		"sessions": sessions,
		"stats":    stats,
		"total":    len(sessions),
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleCreateSession handles session creation requests
func (s *MCPServer) handleCreateSession(ctx context.Context, request *session.SessionAwareRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling create session", zap.String("tool", request.Request.Params.Name))

	if s.sessionManager == nil {
		return mcp.NewToolResultError("Multi-session support not enabled"), nil
	}

	name, err := request.Request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid name parameter: %v", err)), nil
	}

	workspaceDir := request.Request.GetString("workspace_dir", "")

	newSession, err := s.sessionManager.CreateSession(name, workspaceDir)
	if err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create session: %v", err)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"session": newSession,
		"message": fmt.Sprintf("Session '%s' created successfully", name),
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleGetSessionInfo handles session information requests
func (s *MCPServer) handleGetSessionInfo(ctx context.Context, request *session.SessionAwareRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get session info", zap.String("tool", request.Request.Params.Name))

	result := map[string]interface{}{
		"current_session": request.Session,
		"multi_session_enabled": s.config.Server.MultiSession.Enabled,
		"session_config": s.config.Server.MultiSession,
	}

	if s.sessionManager != nil {
		result["session_stats"] = s.sessionManager.GetSessionStats()
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// Helper methods for session-aware operations

// getBooleanValue extracts a boolean value from session-aware request arguments
func (s *MCPServer) getBooleanValueFromSession(request *session.SessionAwareRequest, key string, defaultValue bool) bool {
	args := s.getArgumentsFromSession(request)
	if value, exists := args[key]; exists {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// getArgumentsFromSession extracts arguments from session-aware request
func (s *MCPServer) getArgumentsFromSession(request *session.SessionAwareRequest) map[string]interface{} {
	if args, ok := request.Request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}
	return make(map[string]interface{})
}
