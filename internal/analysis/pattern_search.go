package analysis

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// PatternSearchAnalyzer implements pattern-based code search
type PatternSearchAnalyzer struct {
	logger  *zap.Logger
	enabled bool
}

// NewPatternSearchAnalyzer creates a new pattern search analyzer
func NewPatternSearchAnalyzer(logger *zap.Logger) *PatternSearchAnalyzer {
	return &PatternSearchAnalyzer{
		logger:  logger,
		enabled: true,
	}
}

// Name returns the analyzer name
func (p *PatternSearchAnalyzer) Name() string {
	return "pattern_search"
}

// IsEnabled returns whether the analyzer is enabled
func (p *PatternSearchAnalyzer) IsEnabled() bool {
	return p.enabled
}

// Analyze performs pattern search analysis
func (p *PatternSearchAnalyzer) Analyze(ctx context.Context, target interface{}) (interface{}, error) {
	request, ok := target.(*types.PatternSearchRequest)
	if !ok {
		return nil, fmt.Errorf("invalid target type for pattern search analyzer")
	}

	startTime := time.Now()
	p.logger.Info("Starting pattern search analysis",
		zap.String("pattern", request.Pattern),
		zap.String("language", request.Language),
		zap.Bool("include_tests", request.IncludeTests))

	// Compile regex pattern
	regex, err := regexp.Compile(request.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Perform search (mock implementation)
	matches := p.searchPattern(regex, request)

	searchTime := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds

	result := &types.PatternSearchResult{
		Pattern:      request.Pattern,
		Language:     request.Language,
		TotalMatches: len(matches),
		Matches:      matches,
		SearchTime:   searchTime,
	}

	p.logger.Info("Pattern search completed",
		zap.Int("matches", len(matches)),
		zap.Float64("search_time_ms", searchTime))

	return result, nil
}

// searchPattern performs the actual pattern search
func (p *PatternSearchAnalyzer) searchPattern(regex *regexp.Regexp, request *types.PatternSearchRequest) []types.PatternMatch {
	var results []types.PatternMatch

	// Mock implementation - in a real implementation, this would:
	// 1. Search through indexed files
	// 2. Apply language filters
	// 3. Include/exclude test files based on request
	// 4. Use the compiled regex to find matches

	// Sample mock data
	mockFiles := []struct {
		path     string
		content  string
		language string
		isTest   bool
	}{
		{
			path:     "src/main.go",
			content:  "func main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			language: "go",
			isTest:   false,
		},
		{
			path:     "src/utils.go",
			content:  "func processData(data string) error {\n\treturn nil\n}",
			language: "go",
			isTest:   false,
		},
		{
			path:     "src/main_test.go",
			content:  "func TestMain(t *testing.T) {\n\t// test code\n}",
			language: "go",
			isTest:   true,
		},
	}

	for _, file := range mockFiles {
		// Skip test files if not requested
		if file.isTest && !request.IncludeTests {
			continue
		}

		// Apply language filter
		if request.Language != "" && file.language != request.Language {
			continue
		}

		// Search for pattern in file content
		lines := strings.Split(file.content, "\n")
		for lineNum, line := range lines {
			if matches := regex.FindAllStringIndex(line, -1); matches != nil {
				for _, match := range matches {
					patternMatch := types.PatternMatch{
						FileID:      fmt.Sprintf("file_%d", len(matches)+1),
						FilePath:    file.path,
						LineNumber:  lineNum + 1,
						ColumnStart: match[0] + 1,
						ColumnEnd:   match[1] + 1,
						MatchText:   line[match[0]:match[1]],
						Context: map[string]string{
							"line":     line,
							"language": file.language,
						},
					}
					results = append(results, patternMatch)
				}
			}
		}
	}

	return results
}

// searchByAST performs AST-based pattern search (placeholder)
func (p *PatternSearchAnalyzer) searchByAST(pattern string, language string) []types.PatternMatch {
	// This would implement AST-based pattern matching
	// For now, return empty results
	p.logger.Debug("AST-based search not yet implemented", zap.String("pattern", pattern))
	return []types.PatternMatch{}
}

// searchBySemantic performs semantic pattern search (placeholder)
func (p *PatternSearchAnalyzer) searchBySemantic(pattern string, language string) []types.PatternMatch {
	// This would implement semantic pattern matching using ML
	// For now, return empty results
	p.logger.Debug("Semantic search not yet implemented", zap.String("pattern", pattern))
	return []types.PatternMatch{}
}

// validatePattern validates the search pattern
func (p *PatternSearchAnalyzer) validatePattern(pattern string, patternType string) error {
	switch patternType {
	case "regex":
		_, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	case "ast":
		// Validate AST pattern syntax
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("AST pattern cannot be empty")
		}
	case "semantic":
		// Validate semantic pattern
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("semantic pattern cannot be empty")
		}
	default:
		return fmt.Errorf("unsupported pattern type: %s", patternType)
	}

	return nil
}

// getLanguageExtensions returns file extensions for a language
func (p *PatternSearchAnalyzer) getLanguageExtensions(language string) []string {
	extensions := map[string][]string{
		"go":         {".go"},
		"python":     {".py"},
		"javascript": {".js", ".jsx", ".ts", ".tsx"},
		"java":       {".java"},
		"c":          {".c", ".h"},
		"cpp":        {".cpp", ".hpp", ".cc", ".cxx"},
		"rust":       {".rs"},
		"ruby":       {".rb"},
		"php":        {".php"},
		"csharp":     {".cs"},
		"kotlin":     {".kt"},
		"swift":      {".swift"},
		"scala":      {".scala"},
	}

	if exts, exists := extensions[language]; exists {
		return exts
	}

	return []string{} // Return empty slice for unknown languages
}

// isTestFile determines if a file is a test file
func (p *PatternSearchAnalyzer) isTestFile(filePath string) bool {
	testPatterns := []string{
		"_test.",
		"test_",
		"/test/",
		"/tests/",
		".test.",
		".spec.",
		"spec_",
	}

	lowerPath := strings.ToLower(filePath)
	for _, pattern := range testPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// formatSearchResults formats search results for display
func (p *PatternSearchAnalyzer) formatSearchResults(matches []types.PatternMatch) map[string]interface{} {
	// Group matches by file
	fileGroups := make(map[string][]types.PatternMatch)
	for _, match := range matches {
		fileGroups[match.FilePath] = append(fileGroups[match.FilePath], match)
	}

	// Create summary statistics
	languageStats := make(map[string]int)
	for _, match := range matches {
		if lang, exists := match.Context["language"]; exists {
			languageStats[lang]++
		}
	}

	return map[string]interface{}{
		"total_matches":    len(matches),
		"files_with_matches": len(fileGroups),
		"language_distribution": languageStats,
		"grouped_matches": fileGroups,
	}
}
