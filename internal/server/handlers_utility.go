package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// Utility tool handlers for file operations and symbol finding

// handleFindFiles handles file finding requests
func (s *MCPServer) handleFindFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling find files", zap.String("tool", request.Params.Name))

	pattern, err := request.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid pattern parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")
	includeContent := s.getBooleanValue(request, "include_content", false)

	// Use the search engine to find files matching the pattern
	searchQuery := types.SearchQuery{
		Query:      pattern,
		Type:       "file",
		Repository: repository,
		MaxResults: 100,
	}

	searchResults, err := s.searcher.Search(ctx, searchQuery)
	if err != nil {
		s.logger.Error("Failed to search files", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	files := make([]map[string]interface{}, 0, len(searchResults))
	for _, result := range searchResults {
		fileInfo := map[string]interface{}{
			"path":       result.FilePath,
			"repository": result.Repository,
			"score":      result.Score,
			"language":   result.Language,
			"type":       result.Type,
			"start_line": result.StartLine,
			"end_line":   result.EndLine,
		}

		// Add highlights if available
		if result.Highlights != nil {
			fileInfo["highlights"] = result.Highlights
		}

		// Include content preview if requested
		if includeContent && result.Content != "" {
			// Limit content preview to first 500 characters
			content := result.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			fileInfo["content"] = content
		}

		// Add snippet if available
		if result.Snippet != "" {
			fileInfo["snippet"] = result.Snippet
		}

		files = append(files, fileInfo)
	}

	response := map[string]interface{}{
		"pattern":       pattern,
		"repository":    repository,
		"files":         files,
		"total_matches": len(files),
	}

	content, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleFindSymbols handles symbol finding requests
func (s *MCPServer) handleFindSymbols(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling find symbols", zap.String("tool", request.Params.Name))

	symbolName, err := request.RequireString("symbol_name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid symbol_name parameter: %v", err)), nil
	}

	symbolType := request.GetString("symbol_type", "")
	language := request.GetString("language", "")
	repository := request.GetString("repository", "")

	// Use the search engine to find symbols
	searchQuery := types.SearchQuery{
		Query:      symbolName,
		Type:       symbolType, // If empty, will search all symbol types
		Language:   language,
		Repository: repository,
		MaxResults: 100,
		Fuzzy:      true, // Enable fuzzy matching for symbol names
	}

	searchResults, err := s.searcher.Search(ctx, searchQuery)
	if err != nil {
		s.logger.Error("Failed to search symbols", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	symbols := make([]map[string]interface{}, 0, len(searchResults))
	for _, result := range searchResults {
		// Only include actual symbols (not file content)
		if result.Type == "file" || result.Type == "content" {
			continue
		}

		symbolInfo := map[string]interface{}{
			"name":       result.Name,
			"type":       result.Type,
			"file_path":  result.FilePath,
			"repository": result.Repository,
			"language":   result.Language,
			"start_line": result.StartLine,
			"end_line":   result.EndLine,
			"score":      result.Score,
		}

		// Add content/signature if available
		if result.Content != "" {
			symbolInfo["signature"] = result.Content
		}

		// Add snippet for context
		if result.Snippet != "" {
			symbolInfo["context"] = result.Snippet
		}

		// Add highlights if available
		if result.Highlights != nil {
			symbolInfo["highlights"] = result.Highlights
		}

		symbols = append(symbols, symbolInfo)
	}

	response := map[string]interface{}{
		"symbol_name":   symbolName,
		"symbol_type":   symbolType,
		"language":      language,
		"repository":    repository,
		"symbols":       symbols,
		"total_matches": len(symbols),
	}

	content, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleGetFileContent handles file content retrieval requests
func (s *MCPServer) handleGetFileContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get file content", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")
	startLine := int(request.GetFloat("start_line", 0))
	endLine := int(request.GetFloat("end_line", 0))

	// Try to resolve the full file path
	var fullPath string
	if repository != "" {
		// If repository is specified, look for the file in that repository
		// For now, we'll search in the repositories directory
		repoPath := filepath.Join("./repositories", repository)
		fullPath = filepath.Join(repoPath, filePath)
	} else {
		// Try the file path as-is first
		fullPath = filePath
	}

	// Read the file content
	contentBytes, err := s.repoMgr.GetFileContent(fullPath)
	if err != nil {
		// If that fails and no repository was specified, try searching for the file
		if repository == "" {
			// Search for the file in indexed repositories
			searchQuery := types.SearchQuery{
				Query:      filepath.Base(filePath),
				Type:       "file",
				MaxResults: 1,
			}

			searchResults, searchErr := s.searcher.Search(ctx, searchQuery)
			if searchErr == nil && len(searchResults) > 0 {
				// Try to read from the first match
				fullPath = searchResults[0].FilePath
				contentBytes, err = s.repoMgr.GetFileContent(fullPath)
			}
		}

		if err != nil {
			s.logger.Error("Failed to read file content", zap.String("path", fullPath), zap.Error(err))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
		}
	}

	content := string(contentBytes)
	lines := strings.Split(content, "\n")

	// Apply line range if specified
	if startLine > 0 && endLine > 0 && startLine <= len(lines) && endLine <= len(lines) && startLine <= endLine {
		lines = lines[startLine-1 : endLine]
		content = strings.Join(lines, "\n")
	}

	// Detect language from file extension
	language := s.repoMgr.GetFileLanguage(filePath)

	result := map[string]interface{}{
		"file_path":   filePath,
		"full_path":   fullPath,
		"repository":  repository,
		"content":     content,
		"total_lines": len(strings.Split(string(contentBytes), "\n")),
		"start_line":  startLine,
		"end_line":    endLine,
		"language":    language,
		"size":        len(contentBytes),
	}

	responseContent, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(responseContent)), nil
}

// handleListDirectory handles directory listing requests
func (s *MCPServer) handleListDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling list directory", zap.String("tool", request.Params.Name))

	directoryPath, err := request.RequireString("directory_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid directory_path parameter: %v", err)), nil
	}

	repository := request.GetString("repository", "")
	recursive := s.getBooleanValue(request, "recursive", false)
	fileFilter := request.GetString("file_filter", "")

	// Resolve the full directory path
	var fullPath string
	if repository != "" {
		repoPath := filepath.Join("./repositories", repository)
		fullPath = filepath.Join(repoPath, directoryPath)
	} else {
		fullPath = directoryPath
	}

	// List directory contents
	entries, err := s.listDirectoryContents(fullPath, recursive, fileFilter)
	if err != nil {
		s.logger.Error("Failed to list directory", zap.String("path", fullPath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list directory: %v", err)), nil
	}

	result := map[string]interface{}{
		"directory_path": directoryPath,
		"full_path":      fullPath,
		"repository":     repository,
		"recursive":      recursive,
		"file_filter":    fileFilter,
		"entries":        entries,
		"total_entries":  len(entries),
	}

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// File manipulation tool handlers for direct file editing

// handleDeleteLines handles line deletion requests
func (s *MCPServer) handleDeleteLines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling delete lines", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	startLine := int(request.GetFloat("start_line", 0))
	endLine := int(request.GetFloat("end_line", 0))

	if startLine <= 0 || endLine <= 0 {
		return mcp.NewToolResultError("start_line and end_line must be positive integers"), nil
	}

	if startLine > endLine {
		return mcp.NewToolResultError("start_line must be less than or equal to end_line"), nil
	}

	// Read the file content
	contentBytes, err := s.repoMgr.GetFileContent(filePath)
	if err != nil {
		s.logger.Error("Failed to read file for line deletion", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	lines := strings.Split(string(contentBytes), "\n")
	totalLines := len(lines)

	if startLine > totalLines || endLine > totalLines {
		return mcp.NewToolResultError(fmt.Sprintf("Line numbers exceed file length (%d lines)", totalLines)), nil
	}

	// Delete the specified lines (convert to 0-based indexing)
	newLines := append(lines[:startLine-1], lines[endLine:]...)
	newContent := strings.Join(newLines, "\n")

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		s.logger.Error("Failed to write file after line deletion", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write file: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":       true,
		"file_path":     filePath,
		"start_line":    startLine,
		"end_line":      endLine,
		"lines_deleted": endLine - startLine + 1,
		"original_lines": totalLines,
		"new_lines":     len(newLines),
		"message":       fmt.Sprintf("Successfully deleted lines %d-%d from %s", startLine, endLine, filePath),
	}

	s.logger.Info("Lines deleted successfully",
		zap.String("file", filePath),
		zap.Int("start", startLine),
		zap.Int("end", endLine))

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleInsertAtLine handles line insertion requests
func (s *MCPServer) handleInsertAtLine(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling insert at line", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	lineNumber := int(request.GetFloat("line_number", 0))
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid content parameter: %v", err)), nil
	}

	if lineNumber <= 0 {
		return mcp.NewToolResultError("line_number must be a positive integer"), nil
	}

	// Read the file content
	contentBytes, err := s.repoMgr.GetFileContent(filePath)
	if err != nil {
		s.logger.Error("Failed to read file for line insertion", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	lines := strings.Split(string(contentBytes), "\n")
	totalLines := len(lines)

	if lineNumber > totalLines+1 {
		return mcp.NewToolResultError(fmt.Sprintf("Line number %d exceeds file length (%d lines)", lineNumber, totalLines)), nil
	}

	// Insert the content at the specified line (convert to 0-based indexing)
	insertIndex := lineNumber - 1
	if insertIndex > len(lines) {
		insertIndex = len(lines)
	}

	// Split content by newlines to handle multi-line insertions
	contentLines := strings.Split(content, "\n")

	// Insert the new lines
	newLines := make([]string, 0, len(lines)+len(contentLines))
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, contentLines...)
	newLines = append(newLines, lines[insertIndex:]...)

	newContent := strings.Join(newLines, "\n")

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		s.logger.Error("Failed to write file after line insertion", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write file: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":        true,
		"file_path":      filePath,
		"line_number":    lineNumber,
		"lines_inserted": len(contentLines),
		"original_lines": totalLines,
		"new_lines":      len(newLines),
		"content":        content,
		"message":        fmt.Sprintf("Successfully inserted %d lines at line %d in %s", len(contentLines), lineNumber, filePath),
	}

	s.logger.Info("Lines inserted successfully",
		zap.String("file", filePath),
		zap.Int("line", lineNumber),
		zap.Int("inserted", len(contentLines)))

	responseContent, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(responseContent)), nil
}

// handleReplaceLines handles line replacement requests
func (s *MCPServer) handleReplaceLines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling replace lines", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	startLine := int(request.GetFloat("start_line", 0))
	endLine := int(request.GetFloat("end_line", 0))
	newContent, err := request.RequireString("new_content")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid new_content parameter: %v", err)), nil
	}

	if startLine <= 0 || endLine <= 0 {
		return mcp.NewToolResultError("start_line and end_line must be positive integers"), nil
	}

	if startLine > endLine {
		return mcp.NewToolResultError("start_line must be less than or equal to end_line"), nil
	}

	// Read the file content
	contentBytes, err := s.repoMgr.GetFileContent(filePath)
	if err != nil {
		s.logger.Error("Failed to read file for line replacement", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	lines := strings.Split(string(contentBytes), "\n")
	totalLines := len(lines)

	if startLine > totalLines || endLine > totalLines {
		return mcp.NewToolResultError(fmt.Sprintf("Line numbers exceed file length (%d lines)", totalLines)), nil
	}

	// Split new content by newlines to handle multi-line replacements
	newContentLines := strings.Split(newContent, "\n")

	// Replace the specified lines (convert to 0-based indexing)
	newLines := make([]string, 0, len(lines)-((endLine-startLine)+1)+len(newContentLines))
	newLines = append(newLines, lines[:startLine-1]...)
	newLines = append(newLines, newContentLines...)
	newLines = append(newLines, lines[endLine:]...)

	finalContent := strings.Join(newLines, "\n")

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(finalContent), 0644)
	if err != nil {
		s.logger.Error("Failed to write file after line replacement", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write file: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":         true,
		"file_path":       filePath,
		"start_line":      startLine,
		"end_line":        endLine,
		"lines_replaced":  endLine - startLine + 1,
		"new_lines_count": len(newContentLines),
		"original_lines":  totalLines,
		"final_lines":     len(newLines),
		"new_content":     newContent,
		"message":         fmt.Sprintf("Successfully replaced lines %d-%d in %s with %d new lines", startLine, endLine, filePath, len(newContentLines)),
	}

	s.logger.Info("Lines replaced successfully",
		zap.String("file", filePath),
		zap.Int("start", startLine),
		zap.Int("end", endLine),
		zap.Int("new_lines", len(newContentLines)))

	responseContent, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(responseContent)), nil
}

// Advanced utility tool handlers for enhanced code intelligence

// handleGetFileSnippet handles file snippet extraction requests
func (s *MCPServer) handleGetFileSnippet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get file snippet", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	startLine := int(request.GetFloat("start_line", 0))
	endLine := int(request.GetFloat("end_line", 0))
	includeContext := s.getBooleanValue(request, "include_context", false)

	if startLine <= 0 || endLine <= 0 {
		return mcp.NewToolResultError("start_line and end_line must be positive integers"), nil
	}

	if startLine > endLine {
		return mcp.NewToolResultError("start_line must be less than or equal to end_line"), nil
	}

	// Read the file content
	contentBytes, err := s.repoMgr.GetFileContent(filePath)
	if err != nil {
		s.logger.Error("Failed to read file for snippet extraction", zap.String("path", filePath), zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	lines := strings.Split(string(contentBytes), "\n")
	totalLines := len(lines)

	if startLine > totalLines || endLine > totalLines {
		return mcp.NewToolResultError(fmt.Sprintf("Line numbers exceed file length (%d lines)", totalLines)), nil
	}

	// Extract the snippet
	snippetLines := lines[startLine-1 : endLine]
	snippet := strings.Join(snippetLines, "\n")

	// Add context if requested
	var contextBefore, contextAfter []string
	contextSize := 3 // Number of context lines to include

	if includeContext {
		// Get context before
		contextStart := startLine - contextSize - 1
		if contextStart < 0 {
			contextStart = 0
		}
		if contextStart < startLine-1 {
			contextBefore = lines[contextStart : startLine-1]
		}

		// Get context after
		contextEnd := endLine + contextSize
		if contextEnd > totalLines {
			contextEnd = totalLines
		}
		if contextEnd > endLine {
			contextAfter = lines[endLine:contextEnd]
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"file_path":     filePath,
		"start_line":    startLine,
		"end_line":      endLine,
		"snippet":       snippet,
		"snippet_lines": len(snippetLines),
		"total_lines":   totalLines,
		"language":      s.repoMgr.GetFileLanguage(filePath),
	}

	if includeContext {
		result["context_before"] = strings.Join(contextBefore, "\n")
		result["context_after"] = strings.Join(contextAfter, "\n")
		result["context_before_lines"] = len(contextBefore)
		result["context_after_lines"] = len(contextAfter)
	}

	s.logger.Info("File snippet extracted successfully",
		zap.String("file", filePath),
		zap.Int("start", startLine),
		zap.Int("end", endLine),
		zap.Bool("context", includeContext))

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleFindReferences handles symbol reference finding requests
func (s *MCPServer) handleFindReferences(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling find references", zap.String("tool", request.Params.Name))

	symbolName, err := request.RequireString("symbol_name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid symbol_name parameter: %v", err)), nil
	}

	symbolType := request.GetString("symbol_type", "")
	repository := request.GetString("repository", "")
	includeDefinitions := s.getBooleanValue(request, "include_definitions", true)

	// Search for the symbol in code content
	searchQuery := types.SearchQuery{
		Query:      symbolName,
		Type:       "content", // Search in file content for references
		Language:   "",
		Repository: repository,
		MaxResults: 200, // Higher limit for references
		Fuzzy:      false, // Exact matches for references
	}

	searchResults, err := s.searcher.Search(ctx, searchQuery)
	if err != nil {
		s.logger.Error("Failed to search for references", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Reference search failed: %v", err)), nil
	}

	// Also search for symbol definitions if requested
	var definitionResults []types.SearchResult
	if includeDefinitions {
		defQuery := types.SearchQuery{
			Query:      symbolName,
			Type:       symbolType, // Search for actual symbol definitions
			Language:   "",
			Repository: repository,
			MaxResults: 50,
			Fuzzy:      false,
		}

		definitionResults, err = s.searcher.Search(ctx, defQuery)
		if err != nil {
			s.logger.Warn("Failed to search for definitions", zap.Error(err))
			// Continue without definitions
		}
	}

	references := make([]map[string]interface{}, 0)
	definitions := make([]map[string]interface{}, 0)

	// Process content references
	for _, result := range searchResults {
		// Skip if this looks like a definition rather than a reference
		if !includeDefinitions && (strings.Contains(result.Content, "func "+symbolName) ||
			strings.Contains(result.Content, "class "+symbolName) ||
			strings.Contains(result.Content, "var "+symbolName) ||
			strings.Contains(result.Content, "const "+symbolName)) {
			continue
		}

		refInfo := map[string]interface{}{
			"file_path":    result.FilePath,
			"repository":   result.Repository,
			"language":     result.Language,
			"line_number":  result.StartLine,
			"context":      result.Snippet,
			"content":      result.Content,
			"score":        result.Score,
			"type":         "reference",
		}

		if result.Highlights != nil {
			refInfo["highlights"] = result.Highlights
		}

		references = append(references, refInfo)
	}

	// Process definitions
	for _, result := range definitionResults {
		defInfo := map[string]interface{}{
			"file_path":    result.FilePath,
			"repository":   result.Repository,
			"language":     result.Language,
			"line_number":  result.StartLine,
			"end_line":     result.EndLine,
			"context":      result.Snippet,
			"content":      result.Content,
			"symbol_type":  result.Type,
			"score":        result.Score,
			"type":         "definition",
		}

		if result.Highlights != nil {
			defInfo["highlights"] = result.Highlights
		}

		definitions = append(definitions, defInfo)
	}

	result := map[string]interface{}{
		"symbol_name":         symbolName,
		"symbol_type":         symbolType,
		"repository":          repository,
		"include_definitions": includeDefinitions,
		"references":          references,
		"definitions":         definitions,
		"reference_count":     len(references),
		"definition_count":    len(definitions),
		"total_matches":       len(references) + len(definitions),
	}

	s.logger.Info("References found successfully",
		zap.String("symbol", symbolName),
		zap.Int("references", len(references)),
		zap.Int("definitions", len(definitions)))

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleGitBlame handles Git blame requests
func (s *MCPServer) handleGitBlame(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling git blame", zap.String("tool", request.Params.Name))

	filePath, err := request.RequireString("file_path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file_path parameter: %v", err)), nil
	}

	startLine := int(request.GetFloat("start_line", 0))
	endLine := int(request.GetFloat("end_line", 0))
	repository := request.GetString("repository", "")

	// Resolve the full file path
	var fullPath string
	var repoPath string

	if repository != "" {
		// If repository is specified, look for it in indexed repositories
		repositories, err := s.searcher.ListRepositories(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
		}

		repoFound := false
		for _, repo := range repositories {
			if repo.Name == repository {
				repoFound = true
				repoPath = repo.Path
				fullPath = filepath.Join(repo.Path, filePath)
				break
			}
		}

		if !repoFound {
			return mcp.NewToolResultError(fmt.Sprintf("Repository '%s' not found", repository)), nil
		}
	} else {
		// Try to find the file in any repository
		fullPath = filePath
		// For now, we'll use the current directory as repo path
		repoPath = "."
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("File not found: %s", fullPath)), nil
	}

	// Execute git blame command
	var gitArgs []string
	if startLine > 0 && endLine > 0 {
		gitArgs = []string{"blame", "-L", fmt.Sprintf("%d,%d", startLine, endLine), "--porcelain", filePath}
	} else {
		gitArgs = []string{"blame", "--porcelain", filePath}
	}

	// Change to repository directory for git command
	originalDir, err := os.Getwd()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get current directory: %v", err)), nil
	}

	if repoPath != "." {
		err = os.Chdir(repoPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to change to repository directory: %v", err)), nil
		}
		defer os.Chdir(originalDir)
	}

	// Execute git blame
	cmd := exec.Command("git", gitArgs...)
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error("Git blame command failed", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Git blame failed: %v", err)), nil
	}

	// Parse git blame output
	blameLines := s.parseGitBlameOutput(string(output))

	result := map[string]interface{}{
		"success":     true,
		"file_path":   filePath,
		"full_path":   fullPath,
		"repository":  repository,
		"start_line":  startLine,
		"end_line":    endLine,
		"blame_info":  blameLines,
		"total_lines": len(blameLines),
	}

	s.logger.Info("Git blame completed successfully",
		zap.String("file", filePath),
		zap.Int("lines", len(blameLines)))

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleRefreshIndex handles index refresh requests
func (s *MCPServer) handleRefreshIndex(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling refresh index", zap.String("tool", request.Params.Name))

	repository := request.GetString("repository", "")
	forceRebuild := s.getBooleanValue(request, "force_rebuild", false)

	var refreshedRepos []string
	var errors []string

	if repository != "" {
		// Refresh specific repository
		s.logger.Info("Refreshing specific repository", zap.String("repository", repository))

		// Check if repository exists
		repositories, err := s.searcher.ListRepositories(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
		}

		repoFound := false
		var repoPath string
		for _, repo := range repositories {
			if repo.Name == repository {
				repoFound = true
				repoPath = repo.Path
				break
			}
		}

		if !repoFound {
			return mcp.NewToolResultError(fmt.Sprintf("Repository '%s' not found in indexed repositories", repository)), nil
		}

		// Re-index the specific repository
		_, err = s.indexer.IndexRepository(ctx, repoPath, repository)
		if err != nil {
			s.logger.Error("Failed to refresh repository", zap.String("repository", repository), zap.Error(err))
			errors = append(errors, fmt.Sprintf("Failed to refresh %s: %v", repository, err))
		} else {
			refreshedRepos = append(refreshedRepos, repository)
		}
	} else {
		// Refresh all repositories
		s.logger.Info("Refreshing all repositories", zap.Bool("force_rebuild", forceRebuild))

		repositories, err := s.searcher.ListRepositories(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
		}

		for _, repo := range repositories {
			s.logger.Info("Refreshing repository", zap.String("name", repo.Name), zap.String("path", repo.Path))

			_, err := s.indexer.IndexRepository(ctx, repo.Path, repo.Name)
			if err != nil {
				s.logger.Error("Failed to refresh repository", zap.String("repository", repo.Name), zap.Error(err))
				errors = append(errors, fmt.Sprintf("Failed to refresh %s: %v", repo.Name, err))
			} else {
				refreshedRepos = append(refreshedRepos, repo.Name)
			}
		}
	}

	// Get updated index statistics
	stats, err := s.searcher.GetIndexStats(ctx)
	var statsInterface interface{}
	if err != nil {
		s.logger.Warn("Failed to get updated index stats", zap.Error(err))
		statsInterface = map[string]interface{}{"error": "Failed to retrieve updated stats"}
	} else {
		statsInterface = stats
	}

	result := map[string]interface{}{
		"success":           len(errors) == 0,
		"repository":        repository,
		"force_rebuild":     forceRebuild,
		"refreshed_repos":   refreshedRepos,
		"refreshed_count":   len(refreshedRepos),
		"errors":            errors,
		"error_count":       len(errors),
		"updated_stats":     statsInterface,
		"message":           fmt.Sprintf("Refreshed %d repositories", len(refreshedRepos)),
	}

	if len(errors) > 0 {
		result["message"] = fmt.Sprintf("Refreshed %d repositories with %d errors", len(refreshedRepos), len(errors))
	}

	s.logger.Info("Index refresh completed",
		zap.Int("refreshed", len(refreshedRepos)),
		zap.Int("errors", len(errors)))

	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to format response"), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// parseGitBlameOutput parses the porcelain output from git blame
func (s *MCPServer) parseGitBlameOutput(output string) []map[string]interface{} {
	lines := strings.Split(output, "\n")
	var blameLines []map[string]interface{}

	var currentCommit map[string]interface{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Check if this is a commit hash line (starts with 40-character hex)
		if len(line) >= 40 && strings.Contains(line, " ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				commitHash := parts[0]
				lineNum, err := strconv.Atoi(parts[2])
				if err != nil {
					continue
				}

				currentCommit = map[string]interface{}{
					"commit_hash": commitHash,
					"line_number": lineNum,
				}
			}
		} else if currentCommit != nil {
			// Parse metadata lines
			if strings.HasPrefix(line, "author ") {
				currentCommit["author"] = strings.TrimPrefix(line, "author ")
			} else if strings.HasPrefix(line, "author-mail ") {
				email := strings.TrimPrefix(line, "author-mail ")
				email = strings.Trim(email, "<>")
				currentCommit["author_email"] = email
			} else if strings.HasPrefix(line, "author-time ") {
				timeStr := strings.TrimPrefix(line, "author-time ")
				if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
					currentCommit["author_time"] = time.Unix(timestamp, 0).Format(time.RFC3339)
				}
			} else if strings.HasPrefix(line, "summary ") {
				currentCommit["summary"] = strings.TrimPrefix(line, "summary ")
			} else if strings.HasPrefix(line, "\t") {
				// This is the actual code line
				currentCommit["code"] = strings.TrimPrefix(line, "\t")
				blameLines = append(blameLines, currentCommit)
				currentCommit = nil
			}
		}
	}

	return blameLines
}
