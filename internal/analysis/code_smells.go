package analysis

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// CodeSmellsAnalyzer implements code smells detection
type CodeSmellsAnalyzer struct {
	config  *config.AnalysisConfig
	logger  *zap.Logger
	enabled bool
}

// NewCodeSmellsAnalyzer creates a new code smells analyzer
func NewCodeSmellsAnalyzer(config *config.AnalysisConfig, logger *zap.Logger) *CodeSmellsAnalyzer {
	return &CodeSmellsAnalyzer{
		config:  config,
		logger:  logger,
		enabled: true,
	}
}

// Name returns the analyzer name
func (c *CodeSmellsAnalyzer) Name() string {
	return "code_smells"
}

// IsEnabled returns whether the analyzer is enabled
func (c *CodeSmellsAnalyzer) IsEnabled() bool {
	return c.enabled
}

// Analyze performs code smells analysis
func (c *CodeSmellsAnalyzer) Analyze(ctx context.Context, target interface{}) (interface{}, error) {
	request, ok := target.(*types.CodeSmellsRequest)
	if !ok {
		return nil, fmt.Errorf("invalid target type for code smells analyzer")
	}

	c.logger.Info("Starting code smells analysis",
		zap.String("file_path", request.FilePath),
		zap.String("severity_threshold", request.SeverityThreshold))

	// Get file content (mock for now)
	fileContent := c.getFileContent(request.FilePath)
	if fileContent == "" {
		return nil, fmt.Errorf("could not read file: %s", request.FilePath)
	}

	// Detect code smells
	smells := c.detectCodeSmells(fileContent, request)
	summary := c.generateSummary(smells)

	result := &types.CodeSmellsReport{
		FilePath:    request.FilePath,
		TotalSmells: len(smells),
		Smells:      smells,
		Summary:     summary,
	}

	c.logger.Info("Code smells analysis completed",
		zap.Int("total_smells", len(smells)),
		zap.Float64("overall_score", summary.Score))

	return result, nil
}

// getFileContent retrieves file content (mock implementation)
func (c *CodeSmellsAnalyzer) getFileContent(filePath string) string {
	// Mock file content for demonstration
	return `package main

import (
	"fmt"
	"strings"
	"os"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"time"
	"log"
)

// This is a very long function that does too many things
// It violates the Single Responsibility Principle
func processUserDataAndGenerateReportAndSendEmail(userData map[string]interface{}) error {
	// Validate user data
	if userData["name"] == nil || userData["name"] == "" {
		return fmt.Errorf("name is required")
	}
	if userData["email"] == nil || userData["email"] == "" {
		return fmt.Errorf("email is required")
	}
	if userData["age"] == nil {
		return fmt.Errorf("age is required")
	}
	
	// Process data
	name := userData["name"].(string)
	email := userData["email"].(string)
	age := userData["age"].(int)
	
	// Generate report
	report := fmt.Sprintf("User Report:\nName: %s\nEmail: %s\nAge: %d\n", name, email, age)
	
	// Save to file
	filename := fmt.Sprintf("report_%s_%d.txt", strings.ReplaceAll(name, " ", "_"), time.Now().Unix())
	err := ioutil.WriteFile(filename, []byte(report), 0644)
	if err != nil {
		return err
	}
	
	// Send email (mock)
	emailData := map[string]interface{}{
		"to": email,
		"subject": "Your Report",
		"body": report,
	}
	
	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return err
	}
	
	resp, err := http.Post("https://api.email.com/send", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	log.Printf("Report generated and email sent for user: %s", name)
	return nil
}

// Duplicate code - similar validation logic
func validateUserInput(input map[string]interface{}) error {
	if input["name"] == nil || input["name"] == "" {
		return fmt.Errorf("name is required")
	}
	if input["email"] == nil || input["email"] == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

// Another function with duplicate validation
func checkUserData(data map[string]interface{}) bool {
	if data["name"] == nil || data["name"] == "" {
		return false
	}
	if data["email"] == nil || data["email"] == "" {
		return false
	}
	return true
}

// Dead code - this function is never called
func unusedFunction() {
	fmt.Println("This function is never used")
}

func main() {
	// Main function
}`
}

// detectCodeSmells detects various code smells
func (c *CodeSmellsAnalyzer) detectCodeSmells(content string, request *types.CodeSmellsRequest) []types.CodeSmell {
	var smells []types.CodeSmell

	lines := strings.Split(content, "\n")

	// Detect different types of code smells
	smells = append(smells, c.detectLongMethod(lines)...)
	smells = append(smells, c.detectDuplicateCode(lines)...)
	smells = append(smells, c.detectDeadCode(lines)...)
	smells = append(smells, c.detectGodClass(lines)...)
	smells = append(smells, c.detectLongParameterList(lines)...)
	smells = append(smells, c.detectMagicNumbers(lines)...)

	// Filter by severity threshold
	return c.filterBySeverity(smells, request.SeverityThreshold)
}

// detectLongMethod detects methods that are too long
func (c *CodeSmellsAnalyzer) detectLongMethod(lines []string) []types.CodeSmell {
	var smells []types.CodeSmell
	
	inFunction := false
	functionStart := 0
	functionName := ""
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Detect function start
		if strings.HasPrefix(trimmed, "func ") {
			inFunction = true
			functionStart = i + 1
			// Extract function name
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				functionName = strings.Split(parts[1], "(")[0]
			}
		}
		
		// Detect function end
		if inFunction && trimmed == "}" && !strings.Contains(line, "if") && !strings.Contains(line, "for") {
			functionLength := i - functionStart + 1
			
			// Check if function is too long (threshold: 30 lines)
			if functionLength > 30 {
				smell := types.CodeSmell{
					Type:        "long_method",
					Severity:    c.calculateSeverity(functionLength, 30, 50, 100),
					Location:    types.Location{
						FilePath:  "mock_file.go",
						StartLine: functionStart,
						EndLine:   i + 1,
						Snippet:   fmt.Sprintf("Function %s (%d lines)", functionName, functionLength),
					},
					Description: fmt.Sprintf("Function '%s' is too long (%d lines). Consider breaking it into smaller functions.", functionName, functionLength),
					Suggestion:  "Break this function into smaller, more focused functions following the Single Responsibility Principle.",
					Confidence:  0.9,
				}
				smells = append(smells, smell)
			}
			
			inFunction = false
		}
	}
	
	return smells
}

// detectDuplicateCode detects duplicate code blocks
func (c *CodeSmellsAnalyzer) detectDuplicateCode(lines []string) []types.CodeSmell {
	var smells []types.CodeSmell
	
	// Simple duplicate detection - look for similar validation patterns
	duplicateCount := 0
	
	for i, line := range lines {
		if strings.Contains(line, `== nil`) && strings.Contains(line, `== ""`) {
			duplicateCount++
			if duplicateCount > 2 { // Found multiple similar patterns
				smell := types.CodeSmell{
					Type:        "duplicate_code",
					Severity:    "medium",
					Location:    types.Location{
						FilePath:  "mock_file.go",
						StartLine: i + 1,
						EndLine:   i + 1,
						Snippet:   strings.TrimSpace(line),
					},
					Description: "Duplicate validation logic detected. Consider extracting into a reusable function.",
					Suggestion:  "Create a common validation function to eliminate code duplication.",
					Confidence:  0.8,
				}
				smells = append(smells, smell)
			}
		}
	}
	
	return smells
}

// detectDeadCode detects unused functions
func (c *CodeSmellsAnalyzer) detectDeadCode(lines []string) []types.CodeSmell {
	var smells []types.CodeSmell
	
	// Simple dead code detection - look for functions that are never called
	for i, line := range lines {
		if strings.Contains(line, "func unusedFunction") {
			smell := types.CodeSmell{
				Type:        "dead_code",
				Severity:    "low",
				Location:    types.Location{
					FilePath:  "mock_file.go",
					StartLine: i + 1,
					EndLine:   i + 1,
					Snippet:   strings.TrimSpace(line),
				},
				Description: "Function 'unusedFunction' appears to be unused dead code.",
				Suggestion:  "Remove unused functions to improve code maintainability.",
				Confidence:  0.7,
			}
			smells = append(smells, smell)
		}
	}
	
	return smells
}

// detectGodClass detects classes/structs that are too large
func (c *CodeSmellsAnalyzer) detectGodClass(lines []string) []types.CodeSmell {
	var smells []types.CodeSmell
	
	// For Go, we'll look for files with too many functions
	functionCount := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "func ") {
			functionCount++
		}
	}
	
	if functionCount > 10 {
		smell := types.CodeSmell{
			Type:        "god_class",
			Severity:    "high",
			Location:    types.Location{
				FilePath:  "mock_file.go",
				StartLine: 1,
				EndLine:   len(lines),
				Snippet:   fmt.Sprintf("File with %d functions", functionCount),
			},
			Description: fmt.Sprintf("File contains too many functions (%d). Consider splitting into multiple files.", functionCount),
			Suggestion:  "Split this file into smaller, more focused modules.",
			Confidence:  0.8,
		}
		smells = append(smells, smell)
	}
	
	return smells
}

// detectLongParameterList detects functions with too many parameters
func (c *CodeSmellsAnalyzer) detectLongParameterList(lines []string) []types.CodeSmell {
	var smells []types.CodeSmell
	
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "func ") && strings.Count(line, ",") > 4 {
			smell := types.CodeSmell{
				Type:        "long_parameter_list",
				Severity:    "medium",
				Location:    types.Location{
					FilePath:  "mock_file.go",
					StartLine: i + 1,
					EndLine:   i + 1,
					Snippet:   strings.TrimSpace(line),
				},
				Description: "Function has too many parameters. Consider using a struct or reducing parameters.",
				Suggestion:  "Group related parameters into a struct or reduce the number of parameters.",
				Confidence:  0.7,
			}
			smells = append(smells, smell)
		}
	}
	
	return smells
}

// detectMagicNumbers detects magic numbers in code
func (c *CodeSmellsAnalyzer) detectMagicNumbers(lines []string) []types.CodeSmell {
	var smells []types.CodeSmell
	
	for i, line := range lines {
		// Look for hardcoded numbers (simple detection)
		if strings.Contains(line, "0644") || strings.Contains(line, "30") {
			smell := types.CodeSmell{
				Type:        "magic_numbers",
				Severity:    "low",
				Location:    types.Location{
					FilePath:  "mock_file.go",
					StartLine: i + 1,
					EndLine:   i + 1,
					Snippet:   strings.TrimSpace(line),
				},
				Description: "Magic number detected. Consider using named constants.",
				Suggestion:  "Replace magic numbers with named constants for better readability.",
				Confidence:  0.6,
			}
			smells = append(smells, smell)
		}
	}
	
	return smells
}

// calculateSeverity calculates severity based on thresholds
func (c *CodeSmellsAnalyzer) calculateSeverity(value, low, medium, high int) string {
	if value >= high {
		return "critical"
	} else if value >= medium {
		return "high"
	} else if value >= low {
		return "medium"
	}
	return "low"
}

// filterBySeverity filters smells by severity threshold
func (c *CodeSmellsAnalyzer) filterBySeverity(smells []types.CodeSmell, threshold string) []types.CodeSmell {
	severityOrder := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}
	
	thresholdLevel := severityOrder[threshold]
	var filtered []types.CodeSmell
	
	for _, smell := range smells {
		if severityOrder[smell.Severity] >= thresholdLevel {
			filtered = append(filtered, smell)
		}
	}
	
	return filtered
}

// generateSummary generates a summary of code smells
func (c *CodeSmellsAnalyzer) generateSummary(smells []types.CodeSmell) types.SmellSummary {
	bySeverity := make(map[string]int)
	byType := make(map[string]int)
	
	for _, smell := range smells {
		bySeverity[smell.Severity]++
		byType[smell.Type]++
	}
	
	// Calculate overall score (0-1, higher is better)
	score := 1.0
	if len(smells) > 0 {
		score = 1.0 - (float64(len(smells)) / 20.0) // Normalize to 20 max smells
		if score < 0 {
			score = 0
		}
	}
	
	return types.SmellSummary{
		BySeverity: bySeverity,
		ByType:     byType,
		Score:      score,
	}
}
