package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Helper methods and utilities for MCP server operations

// getBooleanValue extracts a boolean value from MCP request arguments
func (s *MCPServer) getBooleanValue(request mcp.CallToolRequest, key string, defaultValue bool) bool {
	args := s.getArguments(request)
	if value, exists := args[key]; exists {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// getArguments extracts arguments from MCP request
func (s *MCPServer) getArguments(request mcp.CallToolRequest) map[string]interface{} {
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}
	return make(map[string]interface{})
}

// listDirectoryContents lists the contents of a directory with optional filtering
func (s *MCPServer) listDirectoryContents(dirPath string, recursive bool, fileFilter string) ([]map[string]interface{}, error) {
	var entries []map[string]interface{}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("directory not found: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Walk the directory
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == dirPath {
			return nil
		}

		// If not recursive, skip subdirectories
		if !recursive && info.IsDir() {
			relPath, _ := filepath.Rel(dirPath, path)
			if strings.Contains(relPath, string(filepath.Separator)) {
				return filepath.SkipDir
			}
		}

		// Apply file filter if specified
		if fileFilter != "" && !info.IsDir() {
			if !strings.HasSuffix(info.Name(), fileFilter) {
				return nil
			}
		}

		// Create entry
		entry := map[string]interface{}{
			"name":     info.Name(),
			"path":     path,
			"size":     info.Size(),
			"modified": info.ModTime().Format("2006-01-02T15:04:05Z"),
		}

		if info.IsDir() {
			entry["type"] = "directory"
		} else {
			entry["type"] = "file"
			entry["language"] = s.repoMgr.GetFileLanguage(info.Name())
		}

		entries = append(entries, entry)
		return nil
	}

	err = filepath.Walk(dirPath, walkFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return entries, nil
}
