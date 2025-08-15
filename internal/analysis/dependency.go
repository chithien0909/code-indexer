package analysis

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// DependencyAnalyzer implements dependency analysis
type DependencyAnalyzer struct {
	indexer *indexer.Indexer
	logger  *zap.Logger
	enabled bool
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(indexer *indexer.Indexer, logger *zap.Logger) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		indexer: indexer,
		logger:  logger,
		enabled: true,
	}
}

// Name returns the analyzer name
func (d *DependencyAnalyzer) Name() string {
	return "dependency"
}

// IsEnabled returns whether the analyzer is enabled
func (d *DependencyAnalyzer) IsEnabled() bool {
	return d.enabled
}

// Analyze performs dependency analysis
func (d *DependencyAnalyzer) Analyze(ctx context.Context, target interface{}) (interface{}, error) {
	request, ok := target.(*types.DependencyRequest)
	if !ok {
		return nil, fmt.Errorf("invalid target type for dependency analyzer")
	}

	d.logger.Info("Starting dependency analysis",
		zap.String("file_path", request.FilePath),
		zap.Int("depth", request.Depth),
		zap.Bool("include_external", request.IncludeExternal))

	// Get file information from indexer (mock for now)
	file := d.getFileInfo(request.FilePath)
	if file == nil {
		return nil, fmt.Errorf("file not found: %s", request.FilePath)
	}

	// Analyze dependencies
	dependencies := d.analyzeDependencies(file, request.Depth, request.IncludeExternal)
	dependents := d.analyzeDependents(file, request.Depth)
	graph := d.buildDependencyGraph(file, dependencies, dependents)
	metrics := d.calculateDependencyMetrics(dependencies, dependents)

	result := &types.DependencyAnalysis{
		FilePath:     request.FilePath,
		Dependencies: dependencies,
		Dependents:   dependents,
		Graph:        graph,
		Metrics:      metrics,
	}

	d.logger.Info("Dependency analysis completed",
		zap.Int("dependencies", len(dependencies)),
		zap.Int("dependents", len(dependents)))

	return result, nil
}

// getFileInfo retrieves file information (mock implementation)
func (d *DependencyAnalyzer) getFileInfo(filePath string) *types.CodeFile {
	// In a real implementation, this would query the indexer
	// For now, return mock data
	return &types.CodeFile{
		ID:       "file_1",
		Path:     filePath,
		Language: d.detectLanguage(filePath),
		Content:  "// Mock file content",
		Imports: []types.Import{
			{Module: "fmt", StartLine: 1},
			{Module: "strings", StartLine: 2},
			{Module: "github.com/example/pkg", StartLine: 3},
		},
	}
}

// analyzeDependencies analyzes what the file depends on
func (d *DependencyAnalyzer) analyzeDependencies(file *types.CodeFile, depth int, includeExternal bool) []types.Dependency {
	var dependencies []types.Dependency

	// Analyze imports
	for _, imp := range file.Imports {
		depType := d.classifyDependency(imp.Module)
		
		// Skip external dependencies if not requested
		if !includeExternal && depType == "external" {
			continue
		}

		dependency := types.Dependency{
			Name:        imp.Module,
			Type:        depType,
			FilePath:    d.resolveImportPath(imp.Module, file.Language),
			UsageCount:  d.countUsages(file, imp.Module),
			ImportLines: []int{imp.StartLine},
			UsageLines:  d.findUsageLines(file, imp.Module),
		}

		dependencies = append(dependencies, dependency)
	}

	// Recursively analyze dependencies if depth > 1
	if depth > 1 {
		for _, dep := range dependencies {
			if dep.Type == "internal" && dep.FilePath != "" {
				subFile := d.getFileInfo(dep.FilePath)
				if subFile != nil {
					subDeps := d.analyzeDependencies(subFile, depth-1, includeExternal)
					dependencies = append(dependencies, subDeps...)
				}
			}
		}
	}

	return d.deduplicateDependencies(dependencies)
}

// analyzeDependents analyzes what depends on this file
func (d *DependencyAnalyzer) analyzeDependents(file *types.CodeFile, depth int) []types.Dependency {
	var dependents []types.Dependency

	// Mock implementation - in reality, this would search the index
	// for files that import this file
	mockDependents := []types.Dependency{
		{
			Name:       "main.go",
			Type:       "internal",
			FilePath:   "cmd/main.go",
			UsageCount: 3,
			ImportLines: []int{5},
			UsageLines: []int{10, 15, 20},
		},
		{
			Name:       "utils_test.go",
			Type:       "internal",
			FilePath:   "pkg/utils_test.go",
			UsageCount: 1,
			ImportLines: []int{3},
			UsageLines: []int{25},
		},
	}

	dependents = append(dependents, mockDependents...)

	return dependents
}

// buildDependencyGraph creates a dependency graph
func (d *DependencyAnalyzer) buildDependencyGraph(file *types.CodeFile, dependencies, dependents []types.Dependency) types.DependencyGraph {
	var nodes []types.DependencyNode
	var edges []types.DependencyEdge

	// Add current file as central node
	centralNode := types.DependencyNode{
		ID:       file.ID,
		Name:     filepath.Base(file.Path),
		Type:     "current",
		FilePath: file.Path,
	}
	nodes = append(nodes, centralNode)

	// Add dependency nodes and edges
	for _, dep := range dependencies {
		node := types.DependencyNode{
			ID:       fmt.Sprintf("dep_%s", dep.Name),
			Name:     dep.Name,
			Type:     dep.Type,
			FilePath: dep.FilePath,
		}
		nodes = append(nodes, node)

		edge := types.DependencyEdge{
			From:   file.ID,
			To:     node.ID,
			Type:   "imports",
			Weight: dep.UsageCount,
		}
		edges = append(edges, edge)
	}

	// Add dependent nodes and edges
	for _, dep := range dependents {
		node := types.DependencyNode{
			ID:       fmt.Sprintf("dependent_%s", dep.Name),
			Name:     dep.Name,
			Type:     dep.Type,
			FilePath: dep.FilePath,
		}
		nodes = append(nodes, node)

		edge := types.DependencyEdge{
			From:   node.ID,
			To:     file.ID,
			Type:   "imports",
			Weight: dep.UsageCount,
		}
		edges = append(edges, edge)
	}

	return types.DependencyGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

// calculateDependencyMetrics calculates dependency metrics
func (d *DependencyAnalyzer) calculateDependencyMetrics(dependencies, dependents []types.Dependency) types.DependencyMetrics {
	totalDeps := len(dependencies)
	externalDeps := 0
	internalDeps := 0

	for _, dep := range dependencies {
		if dep.Type == "external" {
			externalDeps++
		} else {
			internalDeps++
		}
	}

	// Calculate coupling score (0-1, lower is better)
	couplingScore := float64(totalDeps) / 20.0 // Normalize to 20 max dependencies
	if couplingScore > 1.0 {
		couplingScore = 1.0
	}

	// Calculate cohesion score (0-1, higher is better)
	cohesionScore := 1.0 - (float64(externalDeps) / float64(totalDeps+1))

	return types.DependencyMetrics{
		TotalDependencies:    totalDeps,
		ExternalDependencies: externalDeps,
		InternalDependencies: internalDeps,
		CouplingScore:        couplingScore,
		CohesionScore:        cohesionScore,
	}
}

// Helper methods

func (d *DependencyAnalyzer) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	languageMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".cs":   "csharp",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	return "unknown"
}

func (d *DependencyAnalyzer) classifyDependency(module string) string {
	// Standard library patterns
	standardLibs := map[string][]string{
		"go":         {"fmt", "strings", "os", "io", "net", "http"},
		"python":     {"os", "sys", "json", "re", "datetime"},
		"javascript": {"fs", "path", "util", "crypto"},
	}

	// Check if it's a standard library
	for _, libs := range standardLibs {
		for _, lib := range libs {
			if module == lib {
				return "standard"
			}
		}
	}

	// Check if it's external (contains domain or starts with github, etc.)
	if strings.Contains(module, ".") || 
	   strings.HasPrefix(module, "github.com") ||
	   strings.HasPrefix(module, "gitlab.com") ||
	   strings.HasPrefix(module, "bitbucket.org") {
		return "external"
	}

	return "internal"
}

func (d *DependencyAnalyzer) resolveImportPath(module, language string) string {
	// Mock implementation - in reality, this would resolve the actual file path
	if d.classifyDependency(module) == "internal" {
		return fmt.Sprintf("pkg/%s/%s.%s", module, module, d.getLanguageExtension(language))
	}
	return ""
}

func (d *DependencyAnalyzer) getLanguageExtension(language string) string {
	extensions := map[string]string{
		"go":         "go",
		"python":     "py",
		"javascript": "js",
		"typescript": "ts",
		"java":       "java",
		"cpp":        "cpp",
		"c":          "c",
		"rust":       "rs",
		"ruby":       "rb",
		"php":        "php",
		"csharp":     "cs",
	}

	if ext, exists := extensions[language]; exists {
		return ext
	}
	return "txt"
}

func (d *DependencyAnalyzer) countUsages(file *types.CodeFile, module string) int {
	// Mock implementation - count occurrences in content
	return strings.Count(file.Content, module)
}

func (d *DependencyAnalyzer) findUsageLines(file *types.CodeFile, module string) []int {
	// Mock implementation - find line numbers where module is used
	lines := strings.Split(file.Content, "\n")
	var usageLines []int

	for i, line := range lines {
		if strings.Contains(line, module) {
			usageLines = append(usageLines, i+1)
		}
	}

	return usageLines
}

func (d *DependencyAnalyzer) deduplicateDependencies(dependencies []types.Dependency) []types.Dependency {
	seen := make(map[string]bool)
	var result []types.Dependency

	for _, dep := range dependencies {
		if !seen[dep.Name] {
			seen[dep.Name] = true
			result = append(result, dep)
		}
	}

	return result
}
