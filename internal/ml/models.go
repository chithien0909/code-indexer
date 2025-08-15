package ml

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// Mock models for demonstration - replace with actual Spago implementations

// MockEmbeddingModel represents a mock embedding model
type MockEmbeddingModel struct {
	dimensions int
}

// Encode generates a mock embedding vector for code
func (m *MockEmbeddingModel) Encode(code string) ([]float32, error) {
	// Simple mock implementation - in reality, this would use Spago neural networks
	dimensions := 128
	if m.dimensions > 0 {
		dimensions = m.dimensions
	}

	// Create deterministic but varied embeddings based on code content
	hash := sha256.Sum256([]byte(code))
	seed := int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])
	rng := rand.New(rand.NewSource(seed))

	vector := make([]float32, dimensions)
	
	// Generate features based on code characteristics
	codeLength := float32(len(code))
	lineCount := float32(strings.Count(code, "\n") + 1)
	funcCount := float32(strings.Count(code, "func "))
	classCount := float32(strings.Count(code, "class ") + strings.Count(code, "type "))
	
	// Base features
	vector[0] = normalizeFeature(codeLength, 10000)
	vector[1] = normalizeFeature(lineCount, 1000)
	vector[2] = normalizeFeature(funcCount, 100)
	vector[3] = normalizeFeature(classCount, 50)
	
	// Language-specific features
	if strings.Contains(code, "package ") {
		vector[4] = 0.9 // Go indicator
	} else if strings.Contains(code, "import ") && strings.Contains(code, "def ") {
		vector[5] = 0.9 // Python indicator
	} else if strings.Contains(code, "function ") || strings.Contains(code, "const ") {
		vector[6] = 0.9 // JavaScript indicator
	}
	
	// Fill remaining dimensions with controlled random values
	for i := 7; i < dimensions; i++ {
		vector[i] = float32(rng.NormFloat64() * 0.1)
	}
	
	// Normalize vector
	return normalizeVector(vector), nil
}

// MockClassificationModel represents a mock intent classification model
type MockClassificationModel struct{}

// Classify classifies the intent of code
func (m *MockClassificationModel) Classify(code string) (*types.IntentClassification, error) {
	// Simple rule-based classification for demonstration
	categories := make(map[string]float64)
	
	// Analyze code patterns
	if strings.Contains(code, "test") || strings.Contains(code, "Test") {
		categories["testing"] = 0.8
		categories["utility"] = 0.2
	} else if strings.Contains(code, "main") || strings.Contains(code, "Main") {
		categories["entry_point"] = 0.9
		categories["application"] = 0.1
	} else if strings.Contains(code, "http") || strings.Contains(code, "HTTP") || strings.Contains(code, "server") {
		categories["web_service"] = 0.7
		categories["networking"] = 0.3
	} else if strings.Contains(code, "database") || strings.Contains(code, "sql") || strings.Contains(code, "SQL") {
		categories["data_access"] = 0.8
		categories["persistence"] = 0.2
	} else if strings.Contains(code, "config") || strings.Contains(code, "Config") {
		categories["configuration"] = 0.9
		categories["utility"] = 0.1
	} else {
		categories["business_logic"] = 0.6
		categories["utility"] = 0.4
	}
	
	// Find highest confidence category
	var intent string
	var confidence float64
	for category, score := range categories {
		if score > confidence {
			intent = category
			confidence = score
		}
	}
	
	return &types.IntentClassification{
		CodeSnippet: truncateString(code, 200),
		Intent:      intent,
		Confidence:  confidence,
		Categories:  categories,
		Description: generateIntentDescription(intent),
	}, nil
}

// MockQualityModel represents a mock code quality prediction model
type MockQualityModel struct{}

// Predict predicts code quality metrics
func (m *MockQualityModel) Predict(file *types.CodeFile) (*types.QualityMetrics, error) {
	// Simple heuristic-based quality assessment
	content := file.Content
	lines := strings.Split(content, "\n")
	
	// Calculate basic metrics
	lineCount := len(lines)
	avgLineLength := calculateAverageLineLength(lines)
	commentRatio := calculateCommentRatio(content)
	functionCount := len(file.Functions)
	
	// Maintainability (based on function count and line length)
	maintainability := 1.0 - math.Min(0.8, float64(functionCount)/50.0)
	if avgLineLength > 100 {
		maintainability *= 0.8
	}
	
	// Complexity (based on line count and nesting)
	complexity := math.Min(1.0, float64(lineCount)/1000.0)
	nestingLevel := calculateNestingLevel(content)
	if nestingLevel > 5 {
		complexity = math.Min(1.0, complexity*1.5)
	}
	
	// Readability (based on comments and naming)
	readability := commentRatio
	if hasGoodNaming(content) {
		readability = math.Min(1.0, readability*1.2)
	}
	
	// Documentation (based on comment ratio)
	documentation := commentRatio
	
	// Overall score
	overallScore := (maintainability + (1.0-complexity) + readability + documentation) / 4.0
	
	suggestions := generateQualitySuggestions(maintainability, complexity, readability, documentation)
	
	return &types.QualityMetrics{
		FileID:          file.ID,
		Maintainability: maintainability,
		Complexity:      complexity,
		Readability:     readability,
		Documentation:   documentation,
		OverallScore:    overallScore,
		Suggestions:     suggestions,
	}, nil
}

// MockSimilarityModel represents a mock similarity analysis model
type MockSimilarityModel struct{}

// Utility functions

// normalizeFeature normalizes a feature value to [0, 1] range
func normalizeFeature(value, max float32) float32 {
	normalized := value / max
	if normalized > 1.0 {
		return 1.0
	}
	return normalized
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(vector []float32) []float32 {
	var magnitude float32
	for _, v := range vector {
		magnitude += v * v
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))
	
	if magnitude == 0 {
		return vector
	}
	
	normalized := make([]float32, len(vector))
	for i, v := range vector {
		normalized[i] = v / magnitude
	}
	return normalized
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	
	if normA == 0 || normB == 0 {
		return 0.0
	}
	
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// hashString creates a hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash[:8])
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// generateIntentDescription generates a description for an intent
func generateIntentDescription(intent string) string {
	descriptions := map[string]string{
		"testing":        "Code focused on testing functionality",
		"entry_point":    "Main entry point of the application",
		"web_service":    "Web service or HTTP-related functionality",
		"data_access":    "Database or data access operations",
		"configuration":  "Configuration and setup code",
		"business_logic": "Core business logic implementation",
		"utility":        "Utility or helper functions",
	}
	
	if desc, exists := descriptions[intent]; exists {
		return desc
	}
	return "General purpose code"
}

// calculateAverageLineLength calculates the average line length
func calculateAverageLineLength(lines []string) float64 {
	if len(lines) == 0 {
		return 0
	}
	
	totalLength := 0
	for _, line := range lines {
		totalLength += len(strings.TrimSpace(line))
	}
	
	return float64(totalLength) / float64(len(lines))
}

// calculateCommentRatio calculates the ratio of comment lines to total lines
func calculateCommentRatio(content string) float64 {
	lines := strings.Split(content, "\n")
	commentLines := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || 
		   strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			commentLines++
		}
	}
	
	if len(lines) == 0 {
		return 0
	}
	
	return float64(commentLines) / float64(len(lines))
}

// calculateNestingLevel calculates the maximum nesting level
func calculateNestingLevel(content string) int {
	lines := strings.Split(content, "\n")
	maxLevel := 0
	currentLevel := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "{") {
			currentLevel++
			if currentLevel > maxLevel {
				maxLevel = currentLevel
			}
		}
		if strings.Contains(trimmed, "}") {
			currentLevel--
		}
	}
	
	return maxLevel
}

// hasGoodNaming checks if the code has good naming conventions
func hasGoodNaming(content string) bool {
	// Simple heuristic: check for descriptive variable names
	lines := strings.Split(content, "\n")
	goodNames := 0
	totalNames := 0
	
	for _, line := range lines {
		if strings.Contains(line, "var ") || strings.Contains(line, ":=") {
			totalNames++
			// Check for descriptive names (length > 3, not all caps)
			words := strings.Fields(line)
			for _, word := range words {
				if len(word) > 3 && word != strings.ToUpper(word) {
					goodNames++
					break
				}
			}
		}
	}
	
	if totalNames == 0 {
		return true
	}
	
	return float64(goodNames)/float64(totalNames) > 0.7
}

// generateQualitySuggestions generates quality improvement suggestions
func generateQualitySuggestions(maintainability, complexity, readability, documentation float64) []string {
	var suggestions []string
	
	if maintainability < 0.7 {
		suggestions = append(suggestions, "Consider breaking down large functions into smaller, more focused ones")
	}
	
	if complexity > 0.7 {
		suggestions = append(suggestions, "Reduce code complexity by simplifying conditional logic")
	}
	
	if readability < 0.6 {
		suggestions = append(suggestions, "Add more comments to explain complex logic")
		suggestions = append(suggestions, "Use more descriptive variable and function names")
	}
	
	if documentation < 0.5 {
		suggestions = append(suggestions, "Add documentation comments for public functions and types")
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Code quality looks good! Consider adding unit tests if not present")
	}
	
	return suggestions
}
