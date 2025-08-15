package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// AI model tool handlers for code generation, analysis, and explanation

// handleGenerateCode handles code generation requests
func (s *MCPServer) handleGenerateCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling code generation", zap.String("tool", request.Params.Name))

	prompt, err := request.RequireString("prompt")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid prompt parameter: %v", err)), nil
	}

	language, err := request.RequireString("language")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid language parameter: %v", err)), nil
	}

	result, err := s.modelsEngine.GenerateCode(ctx, prompt, language)
	if err != nil {
		s.logger.Error("Failed to generate code", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate code: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleAnalyzeCode handles code analysis requests
func (s *MCPServer) handleAnalyzeCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling code analysis", zap.String("tool", request.Params.Name))

	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid code parameter: %v", err)), nil
	}

	language, err := request.RequireString("language")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid language parameter: %v", err)), nil
	}

	result, err := s.modelsEngine.AnalyzeCode(ctx, code, language)
	if err != nil {
		s.logger.Error("Failed to analyze code", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze code: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleExplainCode handles code explanation requests
func (s *MCPServer) handleExplainCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling code explanation", zap.String("tool", request.Params.Name))

	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid code parameter: %v", err)), nil
	}

	language, err := request.RequireString("language")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid language parameter: %v", err)), nil
	}

	result, err := s.modelsEngine.ExplainCode(ctx, code, language)
	if err != nil {
		s.logger.Error("Failed to explain code", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to explain code: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}
