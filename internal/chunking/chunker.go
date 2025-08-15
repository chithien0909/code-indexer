package chunking

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// ChunkingStrategy defines how code should be chunked
type ChunkingStrategy string

const (
	// SemanticChunking creates chunks based on code structure (functions, classes)
	SemanticChunking ChunkingStrategy = "semantic"
	// LineBasedChunking creates chunks based on line count
	LineBasedChunking ChunkingStrategy = "line_based"
	// HybridChunking combines semantic and line-based approaches
	HybridChunking ChunkingStrategy = "hybrid"
)

// ChunkingConfig defines configuration for code chunking
type ChunkingConfig struct {
	Strategy         ChunkingStrategy `yaml:"strategy"`
	MaxChunkLines    int              `yaml:"max_chunk_lines"`
	MinChunkLines    int              `yaml:"min_chunk_lines"`
	OverlapLines     int              `yaml:"overlap_lines"`
	PreserveContext  bool             `yaml:"preserve_context"`
	IncludeComments  bool             `yaml:"include_comments"`
	IncludeImports   bool             `yaml:"include_imports"`
}

// DefaultChunkingConfig returns default chunking configuration
func DefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		Strategy:         SemanticChunking,
		MaxChunkLines:    100,
		MinChunkLines:    5,
		OverlapLines:     5,
		PreserveContext:  true,
		IncludeComments:  true,
		IncludeImports:   true,
	}
}

// Chunker handles intelligent code chunking
type Chunker struct {
	config ChunkingConfig
}

// NewChunker creates a new code chunker
func NewChunker(config ChunkingConfig) *Chunker {
	return &Chunker{
		config: config,
	}
}

// ChunkFile creates semantic chunks from a code file
func (c *Chunker) ChunkFile(file *types.CodeFile) []types.CodeChunk {
	switch c.config.Strategy {
	case SemanticChunking:
		return c.semanticChunking(file)
	case LineBasedChunking:
		return c.lineBasedChunking(file)
	case HybridChunking:
		return c.hybridChunking(file)
	default:
		return c.semanticChunking(file)
	}
}

// semanticChunking creates chunks based on code structure
func (c *Chunker) semanticChunking(file *types.CodeFile) []types.CodeChunk {
	var chunks []types.CodeChunk
	lines := strings.Split(file.Content, "\n")

	// Create chunks for each function
	for _, function := range file.Functions {
		chunk := c.createFunctionChunk(file, function, lines)
		chunks = append(chunks, chunk)
	}

	// Create chunks for each class
	for _, class := range file.Classes {
		chunk := c.createClassChunk(file, class, lines)
		chunks = append(chunks, chunk)
	}

	// Create chunks for standalone code blocks (not in functions or classes)
	standaloneChunks := c.createStandaloneChunks(file, lines, chunks)
	chunks = append(chunks, standaloneChunks...)

	// Add file-level chunk with imports and top-level comments
	if c.config.IncludeImports || c.config.IncludeComments {
		fileChunk := c.createFileHeaderChunk(file, lines)
		if fileChunk.Content != "" {
			chunks = append([]types.CodeChunk{fileChunk}, chunks...)
		}
	}

	return chunks
}

// createFunctionChunk creates a chunk for a function
func (c *Chunker) createFunctionChunk(file *types.CodeFile, function types.Function, lines []string) types.CodeChunk {
	startLine := function.StartLine - 1 // Convert to 0-based
	endLine := function.EndLine - 1

	// Extend to include context if enabled
	if c.config.PreserveContext {
		startLine = max(0, startLine-c.config.OverlapLines)
		endLine = min(len(lines)-1, endLine+c.config.OverlapLines)
	}

	// Extract content
	content := strings.Join(lines[startLine:endLine+1], "\n")

	// Create chunk ID
	chunkID := c.generateChunkID(file.ID, "function", function.Name, startLine)

	// Build context information
	context := map[string]interface{}{
		"function_name":   function.Name,
		"signature":       function.Signature,
		"parameters":      function.Parameters,
		"return_type":     function.ReturnType,
		"visibility":      function.Visibility,
		"is_method":       function.IsMethod,
		"language":        file.Language,
		"file_path":       file.Path,
	}

	return types.CodeChunk{
		ID:        chunkID,
		FileID:    file.ID,
		Type:      "function",
		Name:      function.Name,
		StartLine: startLine + 1, // Convert back to 1-based
		EndLine:   endLine + 1,
		Content:   content,
		Context:   context,
	}
}

// createClassChunk creates a chunk for a class
func (c *Chunker) createClassChunk(file *types.CodeFile, class types.Class, lines []string) types.CodeChunk {
	startLine := class.StartLine - 1 // Convert to 0-based
	endLine := class.EndLine - 1

	// For large classes, we might want to split them further
	if endLine-startLine > c.config.MaxChunkLines {
		// Create a header chunk for the class definition
		classHeaderEnd := min(endLine, startLine+c.config.MaxChunkLines)
		content := strings.Join(lines[startLine:classHeaderEnd+1], "\n")
		
		chunkID := c.generateChunkID(file.ID, "class", class.Name, startLine)
		
		context := map[string]interface{}{
			"class_name":    class.Name,
			"super_class":   class.SuperClass,
			"interfaces":    class.Interfaces,
			"visibility":    class.Visibility,
			"language":      file.Language,
			"file_path":     file.Path,
			"is_partial":    endLine > classHeaderEnd,
		}

		return types.CodeChunk{
			ID:        chunkID,
			FileID:    file.ID,
			Type:      "class",
			Name:      class.Name,
			StartLine: startLine + 1,
			EndLine:   classHeaderEnd + 1,
			Content:   content,
			Context:   context,
		}
	}

	// Extend to include context if enabled
	if c.config.PreserveContext {
		startLine = max(0, startLine-c.config.OverlapLines)
		endLine = min(len(lines)-1, endLine+c.config.OverlapLines)
	}

	content := strings.Join(lines[startLine:endLine+1], "\n")
	chunkID := c.generateChunkID(file.ID, "class", class.Name, startLine)

	context := map[string]interface{}{
		"class_name":   class.Name,
		"super_class":  class.SuperClass,
		"interfaces":   class.Interfaces,
		"visibility":   class.Visibility,
		"language":     file.Language,
		"file_path":    file.Path,
	}

	return types.CodeChunk{
		ID:        chunkID,
		FileID:    file.ID,
		Type:      "class",
		Name:      class.Name,
		StartLine: startLine + 1,
		EndLine:   endLine + 1,
		Content:   content,
		Context:   context,
	}
}

// createStandaloneChunks creates chunks for code not in functions or classes
func (c *Chunker) createStandaloneChunks(file *types.CodeFile, lines []string, existingChunks []types.CodeChunk) []types.CodeChunk {
	var chunks []types.CodeChunk
	
	// Create a map of covered lines
	coveredLines := make(map[int]bool)
	for _, chunk := range existingChunks {
		for i := chunk.StartLine - 1; i < chunk.EndLine; i++ {
			coveredLines[i] = true
		}
	}

	// Find uncovered regions
	var currentChunkStart = -1
	for i, line := range lines {
		if !coveredLines[i] && strings.TrimSpace(line) != "" {
			if currentChunkStart == -1 {
				currentChunkStart = i
			}
		} else if currentChunkStart != -1 {
			// End of uncovered region
			if i-currentChunkStart >= c.config.MinChunkLines {
				chunk := c.createStandaloneChunk(file, lines, currentChunkStart, i-1)
				chunks = append(chunks, chunk)
			}
			currentChunkStart = -1
		}
	}

	// Handle final uncovered region
	if currentChunkStart != -1 && len(lines)-currentChunkStart >= c.config.MinChunkLines {
		chunk := c.createStandaloneChunk(file, lines, currentChunkStart, len(lines)-1)
		chunks = append(chunks, chunk)
	}

	return chunks
}

// createStandaloneChunk creates a chunk for standalone code
func (c *Chunker) createStandaloneChunk(file *types.CodeFile, lines []string, startLine, endLine int) types.CodeChunk {
	content := strings.Join(lines[startLine:endLine+1], "\n")
	chunkID := c.generateChunkID(file.ID, "block", "", startLine)

	context := map[string]interface{}{
		"language":   file.Language,
		"file_path":  file.Path,
		"chunk_type": "standalone",
	}

	return types.CodeChunk{
		ID:        chunkID,
		FileID:    file.ID,
		Type:      "block",
		StartLine: startLine + 1,
		EndLine:   endLine + 1,
		Content:   content,
		Context:   context,
	}
}

// createFileHeaderChunk creates a chunk for file-level imports and comments
func (c *Chunker) createFileHeaderChunk(file *types.CodeFile, lines []string) types.CodeChunk {
	var headerLines []string
	var endLine int

	// Include imports
	if c.config.IncludeImports {
		for _, imp := range file.Imports {
			if imp.StartLine-1 < len(lines) {
				headerLines = append(headerLines, lines[imp.StartLine-1])
				endLine = max(endLine, imp.StartLine-1)
			}
		}
	}

	// Include top-level comments
	if c.config.IncludeComments {
		for _, comment := range file.Comments {
			if comment.StartLine <= 20 { // Only include comments in first 20 lines
				for i := comment.StartLine - 1; i < comment.EndLine && i < len(lines); i++ {
					headerLines = append(headerLines, lines[i])
					endLine = max(endLine, i)
				}
			}
		}
	}

	if len(headerLines) == 0 {
		return types.CodeChunk{}
	}

	content := strings.Join(headerLines, "\n")
	chunkID := c.generateChunkID(file.ID, "header", "", 0)

	context := map[string]interface{}{
		"language":     file.Language,
		"file_path":    file.Path,
		"chunk_type":   "file_header",
		"import_count": len(file.Imports),
	}

	return types.CodeChunk{
		ID:        chunkID,
		FileID:    file.ID,
		Type:      "header",
		StartLine: 1,
		EndLine:   endLine + 1,
		Content:   content,
		Context:   context,
	}
}

// lineBasedChunking creates chunks based on line count
func (c *Chunker) lineBasedChunking(file *types.CodeFile) []types.CodeChunk {
	var chunks []types.CodeChunk
	lines := strings.Split(file.Content, "\n")

	for i := 0; i < len(lines); i += c.config.MaxChunkLines - c.config.OverlapLines {
		endLine := min(i+c.config.MaxChunkLines, len(lines))
		
		content := strings.Join(lines[i:endLine], "\n")
		chunkID := c.generateChunkID(file.ID, "block", "", i)

		context := map[string]interface{}{
			"language":   file.Language,
			"file_path":  file.Path,
			"chunk_type": "line_based",
		}

		chunk := types.CodeChunk{
			ID:        chunkID,
			FileID:    file.ID,
			Type:      "block",
			StartLine: i + 1,
			EndLine:   endLine,
			Content:   content,
			Context:   context,
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// hybridChunking combines semantic and line-based approaches
func (c *Chunker) hybridChunking(file *types.CodeFile) []types.CodeChunk {
	// Start with semantic chunking
	chunks := c.semanticChunking(file)

	// Split large chunks using line-based approach
	var finalChunks []types.CodeChunk
	for _, chunk := range chunks {
		if chunk.EndLine-chunk.StartLine > c.config.MaxChunkLines {
			subChunks := c.splitLargeChunk(chunk)
			finalChunks = append(finalChunks, subChunks...)
		} else {
			finalChunks = append(finalChunks, chunk)
		}
	}

	return finalChunks
}

// splitLargeChunk splits a large chunk into smaller ones
func (c *Chunker) splitLargeChunk(chunk types.CodeChunk) []types.CodeChunk {
	var subChunks []types.CodeChunk
	lines := strings.Split(chunk.Content, "\n")

	for i := 0; i < len(lines); i += c.config.MaxChunkLines - c.config.OverlapLines {
		endLine := min(i+c.config.MaxChunkLines, len(lines))
		
		content := strings.Join(lines[i:endLine], "\n")
		chunkID := c.generateChunkID(chunk.FileID, chunk.Type+"_part", chunk.Name, chunk.StartLine+i)

		// Copy and update context
		context := make(map[string]interface{})
		for k, v := range chunk.Context {
			context[k] = v
		}
		context["is_partial"] = true
		context["part_number"] = len(subChunks) + 1

		subChunk := types.CodeChunk{
			ID:        chunkID,
			FileID:    chunk.FileID,
			Type:      chunk.Type,
			Name:      chunk.Name,
			StartLine: chunk.StartLine + i,
			EndLine:   chunk.StartLine + endLine - 1,
			Content:   content,
			Context:   context,
		}

		subChunks = append(subChunks, subChunk)
	}

	return subChunks
}

// generateChunkID generates a unique ID for a chunk
func (c *Chunker) generateChunkID(fileID, chunkType, name string, startLine int) string {
	data := fmt.Sprintf("%s:%s:%s:%d", fileID, chunkType, name, startLine)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:8])
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
