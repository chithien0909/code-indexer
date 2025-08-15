package chunking

import (
	"fmt"
	"testing"

	"github.com/my-mcp/code-indexer/pkg/types"
)

func TestNewChunker(t *testing.T) {
	config := DefaultChunkingConfig()
	chunker := NewChunker(config)

	if chunker == nil {
		t.Fatal("Expected chunker to be created, got nil")
	}

	if chunker.config.Strategy != SemanticChunking {
		t.Errorf("Expected strategy %s, got %s", SemanticChunking, chunker.config.Strategy)
	}
}

func TestSemanticChunking(t *testing.T) {
	config := DefaultChunkingConfig()
	chunker := NewChunker(config)

	// Create a test file with functions and classes
	file := &types.CodeFile{
		ID:       "test-file",
		Language: "go",
		Content: `package main

import "fmt"

// Main function
func main() {
    fmt.Println("Hello, World!")
}

// Helper function
func helper() string {
    return "helper"
}

// TestStruct represents a test structure
type TestStruct struct {
    Name string
    Age  int
}

// Method for TestStruct
func (t *TestStruct) GetName() string {
    return t.Name
}`,
		Functions: []types.Function{
			{
				Name:      "main",
				StartLine: 6,
				EndLine:   8,
				Signature: "func main()",
			},
			{
				Name:      "helper",
				StartLine: 11,
				EndLine:   13,
				Signature: "func helper() string",
			},
			{
				Name:      "GetName",
				StartLine: 21,
				EndLine:   23,
				Signature: "func (t *TestStruct) GetName() string",
				IsMethod:  true,
			},
		},
		Classes: []types.Class{
			{
				Name:      "TestStruct",
				StartLine: 16,
				EndLine:   19,
			},
		},
		Comments: []types.Comment{
			{
				Text:      "Main function",
				StartLine: 5,
				EndLine:   5,
				Type:      "line",
			},
			{
				Text:      "Helper function",
				StartLine: 10,
				EndLine:   10,
				Type:      "line",
			},
			{
				Text:      "TestStruct represents a test structure",
				StartLine: 15,
				EndLine:   15,
				Type:      "line",
			},
			{
				Text:      "Method for TestStruct",
				StartLine: 20,
				EndLine:   20,
				Type:      "line",
			},
		},
		Imports: []types.Import{
			{
				Module:    "fmt",
				StartLine: 3,
			},
		},
	}

	chunks := chunker.ChunkFile(file)

	// Should have chunks for: header, main function, helper function, TestStruct class, GetName method
	expectedMinChunks := 4 // At least function and class chunks
	if len(chunks) < expectedMinChunks {
		t.Errorf("Expected at least %d chunks, got %d", expectedMinChunks, len(chunks))
	}

	// Check that we have different chunk types
	chunkTypes := make(map[string]bool)
	for _, chunk := range chunks {
		chunkTypes[chunk.Type] = true
	}

	expectedTypes := []string{"function", "class"}
	for _, expectedType := range expectedTypes {
		if !chunkTypes[expectedType] {
			t.Errorf("Expected chunk type %s not found", expectedType)
		}
	}

	// Verify chunk content is not empty
	for _, chunk := range chunks {
		if chunk.Content == "" {
			t.Errorf("Chunk %s has empty content", chunk.ID)
		}
		if chunk.StartLine <= 0 || chunk.EndLine <= 0 {
			t.Errorf("Chunk %s has invalid line numbers: start=%d, end=%d", chunk.ID, chunk.StartLine, chunk.EndLine)
		}
		if chunk.StartLine > chunk.EndLine {
			t.Errorf("Chunk %s has start line (%d) greater than end line (%d)", chunk.ID, chunk.StartLine, chunk.EndLine)
		}
	}
}

func TestLineBasedChunking(t *testing.T) {
	config := ChunkingConfig{
		Strategy:      LineBasedChunking,
		MaxChunkLines: 10,
		OverlapLines:  2,
	}
	chunker := NewChunker(config)

	file := &types.CodeFile{
		ID:       "test-file",
		Language: "go",
		Content:  generateLongContent(50), // 50 lines
	}

	chunks := chunker.ChunkFile(file)

	// Should have multiple chunks due to line limit
	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks for long content, got %d", len(chunks))
	}

	// Check chunk sizes
	for _, chunk := range chunks {
		lineCount := chunk.EndLine - chunk.StartLine + 1
		if lineCount > config.MaxChunkLines {
			t.Errorf("Chunk exceeds max lines: %d > %d", lineCount, config.MaxChunkLines)
		}
	}
}

func TestHybridChunking(t *testing.T) {
	config := ChunkingConfig{
		Strategy:      HybridChunking,
		MaxChunkLines: 5, // Small limit to force splitting
		OverlapLines:  1,
	}
	chunker := NewChunker(config)

	file := &types.CodeFile{
		ID:       "test-file",
		Language: "go",
		Content:  generateLongContent(30),
		Functions: []types.Function{
			{
				Name:      "longFunction",
				StartLine: 1,
				EndLine:   20, // Large function that should be split
				Signature: "func longFunction()",
			},
		},
	}

	chunks := chunker.ChunkFile(file)

	// Should have multiple chunks due to large function being split
	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks for large function, got %d", len(chunks))
	}
}

func TestChunkIDGeneration(t *testing.T) {
	config := DefaultChunkingConfig()
	chunker := NewChunker(config)

	// Test that same inputs generate same ID
	id1 := chunker.generateChunkID("file1", "function", "test", 10)
	id2 := chunker.generateChunkID("file1", "function", "test", 10)

	if id1 != id2 {
		t.Errorf("Expected same IDs for same inputs, got %s and %s", id1, id2)
	}

	// Test that different inputs generate different IDs
	id3 := chunker.generateChunkID("file1", "function", "test", 11)
	if id1 == id3 {
		t.Errorf("Expected different IDs for different inputs, got same ID %s", id1)
	}
}

func TestChunkContext(t *testing.T) {
	config := DefaultChunkingConfig()
	chunker := NewChunker(config)

	file := &types.CodeFile{
		ID:       "test-file",
		Language: "python",
		Path:     "/test/file.py",
		Content: `def test_function(param1, param2):
    """Test function docstring"""
    return param1 + param2`,
		Functions: []types.Function{
			{
				Name:       "test_function",
				StartLine:  1,
				EndLine:    3,
				Signature:  "def test_function(param1, param2):",
				Parameters: []string{"param1", "param2"},
				ReturnType: "int",
			},
		},
	}

	chunks := chunker.ChunkFile(file)

	// Find function chunk
	var functionChunk *types.CodeChunk
	for _, chunk := range chunks {
		if chunk.Type == "function" {
			functionChunk = &chunk
			break
		}
	}

	if functionChunk == nil {
		t.Fatal("Expected to find function chunk")
	}

	// Check context information
	if functionChunk.Context == nil {
		t.Fatal("Expected chunk to have context")
	}

	if functionChunk.Context["function_name"] != "test_function" {
		t.Errorf("Expected function_name in context, got %v", functionChunk.Context["function_name"])
	}

	if functionChunk.Context["language"] != "python" {
		t.Errorf("Expected language in context, got %v", functionChunk.Context["language"])
	}

	if functionChunk.Context["file_path"] != "/test/file.py" {
		t.Errorf("Expected file_path in context, got %v", functionChunk.Context["file_path"])
	}
}

// Helper function to generate content with specified number of lines
func generateLongContent(lines int) string {
	content := ""
	for i := 1; i <= lines; i++ {
		content += fmt.Sprintf("// Line %d\n", i)
	}
	return content
}

func TestDefaultChunkingConfig(t *testing.T) {
	config := DefaultChunkingConfig()

	if config.Strategy != SemanticChunking {
		t.Errorf("Expected default strategy %s, got %s", SemanticChunking, config.Strategy)
	}

	if config.MaxChunkLines <= 0 {
		t.Errorf("Expected positive MaxChunkLines, got %d", config.MaxChunkLines)
	}

	if config.MinChunkLines <= 0 {
		t.Errorf("Expected positive MinChunkLines, got %d", config.MinChunkLines)
	}

	if config.OverlapLines < 0 {
		t.Errorf("Expected non-negative OverlapLines, got %d", config.OverlapLines)
	}
}
