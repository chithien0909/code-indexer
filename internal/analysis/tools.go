package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// ToolHandler handles analysis-related MCP tools
type ToolHandler struct {
	engine *Engine
	logger *zap.Logger
}

// NewToolHandler creates a new analysis tool handler
func NewToolHandler(engine *Engine, logger *zap.Logger) *ToolHandler {
	return &ToolHandler{
		engine: engine,
		logger: logger,
	}
}

// RegisterTools registers all analysis tools with the MCP server
func (h *ToolHandler) RegisterTools(mcpServer *server.MCPServer) error {
	if !h.engine.IsEnabled() {
		h.logger.Info("Analysis engine disabled, skipping analysis tool registration")
		return nil
	}

	// Register search_by_pattern tool
	patternSearchTool := mcp.NewTool("search_by_pattern",
		mcp.WithDescription("Search code using regex patterns and AST queries"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Regex or AST pattern to search"),
		),
		mcp.WithString("language",
			mcp.Description("Target programming language"),
		),
		mcp.WithBoolean("include_tests",
			mcp.Description("Include test files in search"),
		),
	)
	mcpServer.AddTool(patternSearchTool, h.handleSearchByPattern)

	// Register find_dependencies tool
	dependencyTool := mcp.NewTool("find_dependencies",
		mcp.WithDescription("Analyze and map code dependencies"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("File to analyze dependencies for"),
		),
		mcp.WithNumber("depth",
			mcp.Description("Dependency depth (1-5)"),
		),
		mcp.WithBoolean("include_external",
			mcp.Description("Include external dependencies"),
		),
	)
	mcpServer.AddTool(dependencyTool, h.handleFindDependencies)

	// Register detect_code_smells tool
	codeSmellsTool := mcp.NewTool("detect_code_smells",
		mcp.WithDescription("Identify code smells and anti-patterns"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("File to analyze"),
		),
		mcp.WithString("severity_threshold",
			mcp.Description("Minimum severity: low, medium, high, critical"),
		),
		mcp.WithArray("smell_types",
			mcp.Description("Specific smells to check"),
		),
	)
	mcpServer.AddTool(codeSmellsTool, h.handleDetectCodeSmells)

	// Register analyze_complexity tool
	complexityTool := mcp.NewTool("analyze_complexity",
		mcp.WithDescription("Calculate cyclomatic and cognitive complexity"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("File path or function name"),
		),
		mcp.WithArray("complexity_types",
			mcp.Description("Types: cyclomatic, cognitive, halstead"),
		),
		mcp.WithNumber("threshold",
			mcp.Description("Complexity threshold for warnings"),
		),
	)
	mcpServer.AddTool(complexityTool, h.handleAnalyzeComplexity)

	// Register detect_security_issues tool
	securityTool := mcp.NewTool("detect_security_issues",
		mcp.WithDescription("Scan for potential security vulnerabilities"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("File to scan"),
		),
		mcp.WithArray("vulnerability_types",
			mcp.Description("Types: sql_injection, xss, hardcoded_secrets"),
		),
		mcp.WithNumber("confidence_threshold",
			mcp.Description("Minimum confidence level"),
		),
	)
	mcpServer.AddTool(securityTool, h.handleDetectSecurityIssues)

	// Register analyze_test_coverage tool
	testCoverageTool := mcp.NewTool("analyze_test_coverage",
		mcp.WithDescription("Analyze test coverage and suggest improvements"),
		mcp.WithString("source_file",
			mcp.Required(),
			mcp.Description("Source file to analyze"),
		),
		mcp.WithString("test_directory",
			mcp.Description("Directory containing tests"),
		),
		mcp.WithString("coverage_type",
			mcp.Description("Coverage type: line, branch, function"),
		),
	)
	mcpServer.AddTool(testCoverageTool, h.handleAnalyzeTestCoverage)

	// Register generate_metrics_report tool
	metricsReportTool := mcp.NewTool("generate_metrics_report",
		mcp.WithDescription("Generate comprehensive code metrics report"),
		mcp.WithString("repository",
			mcp.Required(),
			mcp.Description("Repository to analyze"),
		),
		mcp.WithArray("metrics",
			mcp.Description("Metrics: loc, complexity, maintainability, debt"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: json, markdown, html"),
		),
	)
	mcpServer.AddTool(metricsReportTool, h.handleGenerateMetricsReport)

	// Register analyze_code_evolution tool
	evolutionTool := mcp.NewTool("analyze_code_evolution",
		mcp.WithDescription("Track how code has evolved over time"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("File to track"),
		),
		mcp.WithNumber("time_range",
			mcp.Description("Number of months to analyze"),
		),
		mcp.WithArray("metrics",
			mcp.Description("Metrics: changes, complexity_trend, author_activity"),
		),
	)
	mcpServer.AddTool(evolutionTool, h.handleAnalyzeCodeEvolution)

	// Register extract_common_patterns tool
	patternExtractionTool := mcp.NewTool("extract_common_patterns",
		mcp.WithDescription("Find common code patterns for extraction"),
		mcp.WithString("repository",
			mcp.Required(),
			mcp.Description("Repository to analyze"),
		),
		mcp.WithNumber("min_occurrences",
			mcp.Description("Minimum pattern occurrences"),
		),
		mcp.WithNumber("pattern_size",
			mcp.Description("Minimum lines for pattern"),
		),
	)
	mcpServer.AddTool(patternExtractionTool, h.handleExtractCommonPatterns)

	// Register optimize_imports tool
	importOptimizationTool := mcp.NewTool("optimize_imports",
		mcp.WithDescription("Analyze and optimize import statements"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("File to optimize"),
		),
		mcp.WithBoolean("remove_unused",
			mcp.Description("Remove unused imports"),
		),
		mcp.WithBoolean("sort_imports",
			mcp.Description("Sort import statements"),
		),
	)
	mcpServer.AddTool(importOptimizationTool, h.handleOptimizeImports)

	h.logger.Info("Analysis tools registered successfully", zap.Int("tool_count", 10))
	return nil
}

// Tool handlers

func (h *ToolHandler) handleSearchByPattern(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling pattern search", zap.String("tool", request.Params.Name))

	pattern, err := request.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid pattern parameter: %v", err)), nil
	}

	language := request.GetString("language", "")
	includeTests := h.getBooleanValue(request, "include_tests", false)

	result, err := h.engine.SearchByPattern(ctx, pattern, language, includeTests)
	if err != nil {
		h.logger.Error("Failed to search by pattern", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search by pattern: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleFindDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling dependency analysis", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	depth := int(request.GetFloat("depth", 2))
	includeExternal := h.getBooleanValue(request, "include_external", true)

	result, err := h.engine.FindDependencies(ctx, filePath, depth, includeExternal)
	if err != nil {
		h.logger.Error("Failed to analyze dependencies", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze dependencies: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleDetectCodeSmells(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling code smells detection", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	severityThreshold := request.GetString("severity_threshold", "medium")
	smellTypes := h.getStringArray(request, "smell_types")

	result, err := h.engine.DetectCodeSmells(ctx, filePath, severityThreshold, smellTypes)
	if err != nil {
		h.logger.Error("Failed to detect code smells", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to detect code smells: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleAnalyzeComplexity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling complexity analysis", zap.String("tool", request.Params.Name))

	target, err := request.RequireString("target")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid target parameter: %v", err)), nil
	}

	complexityTypes := h.getStringArray(request, "complexity_types")
	if len(complexityTypes) == 0 {
		complexityTypes = []string{"cyclomatic", "cognitive"}
	}

	threshold := int(request.GetFloat("threshold", 10))

	result, err := h.engine.AnalyzeComplexity(ctx, target, complexityTypes, threshold)
	if err != nil {
		h.logger.Error("Failed to analyze complexity", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze complexity: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleDetectSecurityIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling security issues detection", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	vulnerabilityTypes := h.getStringArray(request, "vulnerability_types")
	confidenceThreshold := request.GetFloat("confidence_threshold", 0.7)

	result, err := h.engine.DetectSecurityIssues(ctx, filePath, vulnerabilityTypes, confidenceThreshold)
	if err != nil {
		h.logger.Error("Failed to detect security issues", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to detect security issues: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleAnalyzeTestCoverage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling test coverage analysis", zap.String("tool", request.Params.Name))

	sourceFile, err := request.RequireString("source_file")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid source_file parameter: %v", err)), nil
	}

	testDirectory := request.GetString("test_directory", "")
	coverageType := request.GetString("coverage_type", "line")

	result, err := h.engine.AnalyzeTestCoverage(ctx, sourceFile, testDirectory, coverageType)
	if err != nil {
		h.logger.Error("Failed to analyze test coverage", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze test coverage: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleGenerateMetricsReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling metrics report generation", zap.String("tool", request.Params.Name))

	repository, err := request.RequireString("repository")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid repository parameter: %v", err)), nil
	}

	metrics := h.getStringArray(request, "metrics")
	if len(metrics) == 0 {
		metrics = []string{"loc", "complexity", "maintainability"}
	}

	format := request.GetString("format", "json")

	result, err := h.engine.GenerateMetricsReport(ctx, repository, metrics, format)
	if err != nil {
		h.logger.Error("Failed to generate metrics report", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate metrics report: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleAnalyzeCodeEvolution(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling code evolution analysis", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	timeRange := int(request.GetFloat("time_range", 6))
	metrics := h.getStringArray(request, "metrics")
	if len(metrics) == 0 {
		metrics = []string{"changes", "complexity_trend"}
	}

	result, err := h.engine.AnalyzeCodeEvolution(ctx, filePath, timeRange, metrics)
	if err != nil {
		h.logger.Error("Failed to analyze code evolution", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze code evolution: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleExtractCommonPatterns(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling pattern extraction", zap.String("tool", request.Params.Name))

	repository, err := request.RequireString("repository")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid repository parameter: %v", err)), nil
	}

	minOccurrences := int(request.GetFloat("min_occurrences", 3))
	patternSize := int(request.GetFloat("pattern_size", 5))

	result, err := h.engine.ExtractCommonPatterns(ctx, repository, minOccurrences, patternSize)
	if err != nil {
		h.logger.Error("Failed to extract common patterns", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to extract common patterns: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (h *ToolHandler) handleOptimizeImports(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling import optimization", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	removeUnused := h.getBooleanValue(request, "remove_unused", true)
	sortImports := h.getBooleanValue(request, "sort_imports", true)

	result, err := h.engine.OptimizeImports(ctx, filePath, removeUnused, sortImports)
	if err != nil {
		h.logger.Error("Failed to optimize imports", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to optimize imports: %v", err)), nil
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// Helper methods for extracting values from request

// getBooleanValue extracts a boolean value from request arguments
func (h *ToolHandler) getBooleanValue(request mcp.CallToolRequest, key string, defaultValue bool) bool {
	args := h.getArguments(request)
	if value, exists := args[key]; exists {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// getStringArray extracts a string array from request arguments
func (h *ToolHandler) getStringArray(request mcp.CallToolRequest, key string) []string {
	args := h.getArguments(request)
	if value, exists := args[key]; exists {
		if arr, ok := value.([]interface{}); ok {
			var result []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

// getArguments safely extracts arguments from request
func (h *ToolHandler) getArguments(request mcp.CallToolRequest) map[string]interface{} {
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}
	return make(map[string]interface{})
}
