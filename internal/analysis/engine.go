package analysis

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// Engine represents the advanced code analysis engine
type Engine struct {
	config   *config.AnalysisConfig
	logger   *zap.Logger
	indexer  *indexer.Indexer
	analyzers map[string]Analyzer
	mu       sync.RWMutex
	enabled  bool
}

// Analyzer interface for different types of code analysis
type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, target interface{}) (interface{}, error)
	IsEnabled() bool
}

// NewEngine creates a new analysis engine
func NewEngine(cfg *config.AnalysisConfig, indexer *indexer.Indexer, logger *zap.Logger) (*Engine, error) {
	if !cfg.Enabled {
		logger.Info("Analysis engine disabled in configuration")
		return &Engine{
			config:  cfg,
			logger:  logger,
			indexer: indexer,
			enabled: false,
		}, nil
	}

	logger.Info("Initializing analysis engine")

	engine := &Engine{
		config:    cfg,
		logger:    logger,
		indexer:   indexer,
		analyzers: make(map[string]Analyzer),
		enabled:   true,
	}

	// Initialize analyzers
	if err := engine.initializeAnalyzers(); err != nil {
		return nil, fmt.Errorf("failed to initialize analyzers: %w", err)
	}

	logger.Info("Analysis engine initialized successfully")
	return engine, nil
}

// IsEnabled returns whether the analysis engine is enabled
func (e *Engine) IsEnabled() bool {
	return e.enabled
}

// initializeAnalyzers initializes all code analyzers
func (e *Engine) initializeAnalyzers() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Pattern Search Analyzer
	e.analyzers["pattern_search"] = NewPatternSearchAnalyzer(e.logger)
	
	// Dependency Analyzer
	e.analyzers["dependency"] = NewDependencyAnalyzer(e.indexer, e.logger)
	
	// Code Smells Analyzer
	e.analyzers["code_smells"] = NewCodeSmellsAnalyzer(e.config, e.logger)
	
	// Complexity Analyzer
	e.analyzers["complexity"] = NewComplexityAnalyzer(e.logger)
	
	// Security Analyzer
	e.analyzers["security"] = NewSecurityAnalyzer(e.config, e.logger)
	
	// Test Coverage Analyzer
	e.analyzers["test_coverage"] = NewTestCoverageAnalyzer(e.logger)
	
	// Metrics Analyzer
	e.analyzers["metrics"] = NewMetricsAnalyzer(e.indexer, e.logger)
	
	// Evolution Analyzer
	e.analyzers["evolution"] = NewEvolutionAnalyzer(e.logger)
	
	// Pattern Extraction Analyzer
	e.analyzers["pattern_extraction"] = NewPatternExtractionAnalyzer(e.logger)
	
	// Import Optimizer
	e.analyzers["import_optimizer"] = NewImportOptimizerAnalyzer(e.logger)

	e.logger.Info("Code analyzers initialized", zap.Int("analyzer_count", len(e.analyzers)))
	return nil
}

// GetAnalyzer returns a specific analyzer by name
func (e *Engine) GetAnalyzer(name string) (Analyzer, error) {
	if !e.enabled {
		return nil, fmt.Errorf("analysis engine is disabled")
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	analyzer, exists := e.analyzers[name]
	if !exists {
		return nil, fmt.Errorf("analyzer '%s' not found", name)
	}

	return analyzer, nil
}

// SearchByPattern performs pattern-based code search
func (e *Engine) SearchByPattern(ctx context.Context, pattern, language string, includeTests bool) (*types.PatternSearchResult, error) {
	analyzer, err := e.GetAnalyzer("pattern_search")
	if err != nil {
		return nil, err
	}

	request := &types.PatternSearchRequest{
		Pattern:      pattern,
		Language:     language,
		IncludeTests: includeTests,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.PatternSearchResult), nil
}

// FindDependencies analyzes code dependencies
func (e *Engine) FindDependencies(ctx context.Context, filePath string, depth int, includeExternal bool) (*types.DependencyAnalysis, error) {
	analyzer, err := e.GetAnalyzer("dependency")
	if err != nil {
		return nil, err
	}

	request := &types.DependencyRequest{
		FilePath:        filePath,
		Depth:           depth,
		IncludeExternal: includeExternal,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.DependencyAnalysis), nil
}

// DetectCodeSmells identifies code smells and anti-patterns
func (e *Engine) DetectCodeSmells(ctx context.Context, filePath string, severityThreshold string, smellTypes []string) (*types.CodeSmellsReport, error) {
	analyzer, err := e.GetAnalyzer("code_smells")
	if err != nil {
		return nil, err
	}

	request := &types.CodeSmellsRequest{
		FilePath:          filePath,
		SeverityThreshold: severityThreshold,
		SmellTypes:        smellTypes,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.CodeSmellsReport), nil
}

// AnalyzeComplexity calculates various complexity metrics
func (e *Engine) AnalyzeComplexity(ctx context.Context, target string, complexityTypes []string, threshold int) (*types.ComplexityAnalysis, error) {
	analyzer, err := e.GetAnalyzer("complexity")
	if err != nil {
		return nil, err
	}

	request := &types.ComplexityRequest{
		Target:          target,
		ComplexityTypes: complexityTypes,
		Threshold:       threshold,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.ComplexityAnalysis), nil
}

// DetectSecurityIssues scans for security vulnerabilities
func (e *Engine) DetectSecurityIssues(ctx context.Context, filePath string, vulnerabilityTypes []string, confidenceThreshold float64) (*types.SecurityReport, error) {
	analyzer, err := e.GetAnalyzer("security")
	if err != nil {
		return nil, err
	}

	request := &types.SecurityRequest{
		FilePath:            filePath,
		VulnerabilityTypes:  vulnerabilityTypes,
		ConfidenceThreshold: confidenceThreshold,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.SecurityReport), nil
}

// AnalyzeTestCoverage analyzes test coverage
func (e *Engine) AnalyzeTestCoverage(ctx context.Context, sourceFile, testDirectory, coverageType string) (*types.TestCoverageReport, error) {
	analyzer, err := e.GetAnalyzer("test_coverage")
	if err != nil {
		return nil, err
	}

	request := &types.TestCoverageRequest{
		SourceFile:    sourceFile,
		TestDirectory: testDirectory,
		CoverageType:  coverageType,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.TestCoverageReport), nil
}

// GenerateMetricsReport creates comprehensive metrics report
func (e *Engine) GenerateMetricsReport(ctx context.Context, repository string, metrics []string, format string) (*types.MetricsReport, error) {
	analyzer, err := e.GetAnalyzer("metrics")
	if err != nil {
		return nil, err
	}

	request := &types.MetricsRequest{
		Repository: repository,
		Metrics:    metrics,
		Format:     format,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.MetricsReport), nil
}

// AnalyzeCodeEvolution tracks code evolution over time
func (e *Engine) AnalyzeCodeEvolution(ctx context.Context, filePath string, timeRange int, metrics []string) (*types.EvolutionAnalysis, error) {
	analyzer, err := e.GetAnalyzer("evolution")
	if err != nil {
		return nil, err
	}

	request := &types.EvolutionRequest{
		FilePath:  filePath,
		TimeRange: timeRange,
		Metrics:   metrics,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.EvolutionAnalysis), nil
}

// ExtractCommonPatterns finds common code patterns
func (e *Engine) ExtractCommonPatterns(ctx context.Context, repository string, minOccurrences, patternSize int) (*types.PatternExtractionResult, error) {
	analyzer, err := e.GetAnalyzer("pattern_extraction")
	if err != nil {
		return nil, err
	}

	request := &types.PatternExtractionRequest{
		Repository:      repository,
		MinOccurrences:  minOccurrences,
		PatternSize:     patternSize,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.PatternExtractionResult), nil
}

// OptimizeImports analyzes and optimizes import statements
func (e *Engine) OptimizeImports(ctx context.Context, filePath string, removeUnused, sortImports bool) (*types.ImportOptimizationResult, error) {
	analyzer, err := e.GetAnalyzer("import_optimizer")
	if err != nil {
		return nil, err
	}

	request := &types.ImportOptimizationRequest{
		FilePath:      filePath,
		RemoveUnused:  removeUnused,
		SortImports:   sortImports,
	}

	result, err := analyzer.Analyze(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.(*types.ImportOptimizationResult), nil
}

// Close gracefully shuts down the analysis engine
func (e *Engine) Close() error {
	if !e.enabled {
		return nil
	}

	e.logger.Info("Shutting down analysis engine")
	
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear analyzers
	e.analyzers = nil

	return nil
}
