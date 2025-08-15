package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// ToolHandler handles ML-related MCP tools
type ToolHandler struct {
	engine  *Engine
	indexer *indexer.Indexer
	logger  *zap.Logger
}

// NewToolHandler creates a new ML tool handler
func NewToolHandler(engine *Engine, indexer *indexer.Indexer, logger *zap.Logger) *ToolHandler {
	return &ToolHandler{
		engine:  engine,
		indexer: indexer,
		logger:  logger,
	}
}

// RegisterTools registers all ML tools with the MCP server
func (h *ToolHandler) RegisterTools(mcpServer *server.MCPServer) error {
	if !h.engine.IsEnabled() {
		h.logger.Info("ML engine disabled, skipping ML tool registration")
		return nil
	}

	// Register analyze_code_similarity tool
	similarityTool := mcp.NewTool("analyze_code_similarity",
		mcp.WithDescription("Analyze similarity between code snippets using ML embeddings"),
		mcp.WithString("code_snippet",
			mcp.Required(),
			mcp.Description("Code snippet to analyze"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository to search for similar code"),
		),
		mcp.WithNumber("similarity_threshold",
			mcp.Description("Minimum similarity threshold (0.0-1.0)"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return"),
		),
	)
	mcpServer.AddTool(similarityTool, h.handleCodeSimilarity)

	// Register predict_code_quality tool
	qualityTool := mcp.NewTool("predict_code_quality",
		mcp.WithDescription("Predict code quality metrics using ML models"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file to analyze"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository containing the file"),
		),
		mcp.WithArray("metrics",
			mcp.Description("Specific metrics to predict"),
		),
	)
	mcpServer.AddTool(qualityTool, h.handleCodeQuality)

	// Register classify_code_intent tool
	intentTool := mcp.NewTool("classify_code_intent",
		mcp.WithDescription("Classify the intent/purpose of code using ML"),
		mcp.WithString("code_snippet",
			mcp.Required(),
			mcp.Description("Code snippet to classify"),
		),
		mcp.WithArray("classification_types",
			mcp.Description("Types of classifications to consider"),
		),
		mcp.WithNumber("confidence_threshold",
			mcp.Description("Minimum confidence threshold"),
		),
	)
	mcpServer.AddTool(intentTool, h.handleCodeIntent)

	// Register generate_code_summary tool
	summaryTool := mcp.NewTool("generate_code_summary",
		mcp.WithDescription("Generate AI-powered summary of code files"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file to summarize"),
		),
		mcp.WithString("repository",
			mcp.Description("Repository containing the file"),
		),
		mcp.WithNumber("max_length",
			mcp.Description("Maximum summary length"),
		),
		mcp.WithArray("focus_areas",
			mcp.Description("Areas to focus on in summary"),
		),
	)
	mcpServer.AddTool(summaryTool, h.handleCodeSummary)

	h.logger.Info("ML tools registered successfully", zap.Int("tool_count", 4))
	return nil
}

// handleCodeSimilarity handles the analyze_code_similarity tool
func (h *ToolHandler) handleCodeSimilarity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling code similarity analysis", zap.String("tool", request.Params.Name))

	// Parse arguments
	codeSnippet, err := request.RequireString("code_snippet")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid code_snippet parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")
	similarityThreshold := request.GetFloat("similarity_threshold", 0.7)
	maxResults := int(request.GetFloat("max_results", 10))

	// Find similar code
	results, err := h.findSimilarCode(ctx, codeSnippet, repository, similarityThreshold, maxResults)
	if err != nil {
		h.logger.Error("Failed to find similar code", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze similarity: %v", err)), nil
	}

	// Format response
	response := map[string]interface{}{
		"query_snippet":        truncateString(codeSnippet, 200),
		"similarity_threshold": similarityThreshold,
		"results_count":        len(results),
		"similar_code":         results,
	}

	content, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleCodeQuality handles the predict_code_quality tool
func (h *ToolHandler) handleCodeQuality(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling code quality prediction", zap.String("tool", request.Params.Name))

	// Parse arguments
	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")

	// Get file from indexer
	file, err := h.getFileFromIndexer(filePath, repository)
	if err != nil {
		h.logger.Error("Failed to get file", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
	}

	// Predict quality
	metrics, err := h.engine.PredictQuality(ctx, file)
	if err != nil {
		h.logger.Error("Failed to predict quality", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to predict quality: %v", err)), nil
	}

	content, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleCodeIntent handles the classify_code_intent tool
func (h *ToolHandler) handleCodeIntent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling code intent classification", zap.String("tool", request.Params.Name))

	// Parse arguments
	codeSnippet, err := request.RequireString("code_snippet")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid code_snippet parameter: %v", err)), nil
	}

	confidenceThreshold := request.GetFloat("confidence_threshold", 0.5)

	// Classify intent
	classification, err := h.engine.ClassifyIntent(ctx, codeSnippet)
	if err != nil {
		h.logger.Error("Failed to classify intent", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to classify intent: %v", err)), nil
	}

	// Filter by confidence threshold
	if classification.Confidence < confidenceThreshold {
		classification.Intent = "uncertain"
		classification.Description = "Classification confidence below threshold"
	}

	content, err := json.MarshalIndent(classification, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleCodeSummary handles the generate_code_summary tool
func (h *ToolHandler) handleCodeSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling code summary generation", zap.String("tool", request.Params.Name))

	// Parse arguments
	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")
	maxLength := int(request.GetFloat("max_length", 500))

	// Get file from indexer
	file, err := h.getFileFromIndexer(filePath, repository)
	if err != nil {
		h.logger.Error("Failed to get file", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
	}

	// Generate summary
	summary, err := h.generateCodeSummary(ctx, file, maxLength)
	if err != nil {
		h.logger.Error("Failed to generate summary", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate summary: %v", err)), nil
	}

	content, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// Helper methods

// findSimilarCode finds code similar to the given snippet
func (h *ToolHandler) findSimilarCode(ctx context.Context, codeSnippet, repository string, threshold float64, maxResults int) ([]*types.SimilarityResult, error) {
	// This is a simplified implementation
	// In a real implementation, you would:
	// 1. Generate embedding for the input code
	// 2. Search through indexed embeddings
	// 3. Calculate similarities
	// 4. Return top matches

	var results []*types.SimilarityResult

	// For demonstration, create a mock result
	mockResult := &types.SimilarityResult{
		SourceID:      "input",
		TargetID:      "mock_file_1",
		Score:         0.85,
		Type:          "function",
		SourceSnippet: truncateString(codeSnippet, 100),
		TargetSnippet: "func mockFunction() { /* similar code */ }",
		Explanation:   "Similar function structure and naming patterns detected",
	}

	if mockResult.Score >= threshold {
		results = append(results, mockResult)
	}

	return results, nil
}

// getFileFromIndexer retrieves a file from the indexer
func (h *ToolHandler) getFileFromIndexer(filePath, repository string) (*types.CodeFile, error) {
	// This would integrate with your existing indexer
	// For now, return a mock file
	return &types.CodeFile{
		ID:       "mock_file_id",
		Path:     filePath,
		Language: "go",
		Content:  "// Mock file content\npackage main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
		Functions: []types.Function{
			{
				Name:      "main",
				StartLine: 4,
				EndLine:   6,
				Body:      "func main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			},
		},
	}, nil
}

// generateCodeSummary generates a summary for a code file
func (h *ToolHandler) generateCodeSummary(ctx context.Context, file *types.CodeFile, maxLength int) (*types.CodeSummary, error) {
	// Simple rule-based summary generation
	// In a real implementation, this would use NLP models

	var keyPoints []string
	var functions []string
	var classes []string

	// Extract functions
	for _, fn := range file.Functions {
		functions = append(functions, fn.Name)
		if strings.Contains(fn.Name, "main") {
			keyPoints = append(keyPoints, "Contains main entry point")
		}
	}

	// Extract classes/types
	for _, class := range file.Classes {
		classes = append(classes, class.Name)
		keyPoints = append(keyPoints, fmt.Sprintf("Defines %s class/type", class.Name))
	}

	// Analyze imports for dependencies
	var dependencies []string
	for _, imp := range file.Imports {
		dependencies = append(dependencies, imp.Module)
	}

	// Determine complexity
	complexity := "low"
	if len(file.Functions) > 10 {
		complexity = "medium"
	}
	if len(file.Functions) > 20 || len(strings.Split(file.Content, "\n")) > 500 {
		complexity = "high"
	}

	// Generate summary text
	summary := fmt.Sprintf("This %s file contains %d functions and %d classes/types. ", 
		file.Language, len(file.Functions), len(file.Classes))
	
	if len(functions) > 0 {
		summary += fmt.Sprintf("Main functions include: %s. ", strings.Join(functions[:min(3, len(functions))], ", "))
	}
	
	if len(dependencies) > 0 {
		summary += fmt.Sprintf("Dependencies: %s. ", strings.Join(dependencies[:min(3, len(dependencies))], ", "))
	}

	// Truncate if necessary
	if len(summary) > maxLength {
		summary = summary[:maxLength-3] + "..."
	}

	return &types.CodeSummary{
		FileID:       file.ID,
		Summary:      summary,
		KeyPoints:    keyPoints,
		Functions:    functions,
		Classes:      classes,
		Dependencies: dependencies,
		Complexity:   complexity,
	}, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
