# MCP Code Indexer Tools Documentation

This document describes all the available tools in the MCP Code Indexer server. All tools are now fully integrated with the MCP framework and use real indexing, search, and file system operations.

## 🛠️ **Available Tools (27 Total) - Fully MCP Integrated**

## ✨ **MCP Integration Features**

All tools have been updated to use real MCP functionality:

- **Real Search Integration**: Tools use the Bleve search engine for actual file and symbol searching
- **File System Operations**: Direct file reading and directory listing from indexed repositories
- **Repository Management**: Integration with Git repository manager for file resolution
- **Language Detection**: Automatic programming language detection from file extensions
- **Error Handling**: Robust error handling with meaningful error messages
- **Performance**: Optimized search queries with configurable result limits
- **Fuzzy Matching**: Support for fuzzy symbol name matching
- **Content Filtering**: Line range support for file content retrieval

### **Core Indexing Tools (5)**

#### 1. `index_repository`
**Description:** Index a Git repository for searching
**Parameters:**
- `path` (required): Local path or Git URL to repository
- `name` (optional): Custom name for the repository

**Example Usage:**
```
Index the repository at /path/to/repo with name "my-project"
```

#### 2. `search_code`
**Description:** Search across all indexed repositories
**Parameters:**
- `query` (required): Search query
- `type` (optional): Search type (function, class, variable, content, file, comment)
- `language` (optional): Filter by programming language
- `repository` (optional): Filter by repository name
- `max_results` (optional): Maximum number of results (default: 100)

**Example Usage:**
```
Search for "handleRequest" functions in Go files
```

#### 3. `get_metadata`
**Description:** Get detailed metadata for a specific file
**Parameters:**
- `file_path` (required): Path to the file
- `repository` (optional): Repository name

**Example Usage:**
```
Get metadata for src/main.go in my-project repository
```

#### 4. `list_repositories`
**Description:** List all indexed repositories with statistics
**Parameters:** None

**Example Usage:**
```
Show all indexed repositories and their stats
```

#### 5. `get_index_stats`
**Description:** Get indexing statistics and information
**Parameters:** None

**Example Usage:**
```
Show indexing statistics and system information
```

### **Utility Tools (11)**

#### 6. `find_files`
**Description:** Find files matching patterns in indexed repositories
**Parameters:**
- `pattern` (required): File name pattern (supports wildcards like *.go, *test*, etc.)
- `repository` (optional): Repository name to search in
- `include_content` (optional): Include file content preview in results

**Example Usage:**
```
Find all Go test files: pattern="*_test.go"
Find all Python files in specific repo: pattern="*.py", repository="my-python-project"
```

#### 7. `find_symbols`
**Description:** Find symbols (functions, classes, variables) by name
**Parameters:**
- `symbol_name` (required): Symbol name or pattern to search for
- `symbol_type` (optional): Type of symbol (function, class, variable, constant, interface)
- `language` (optional): Programming language to filter by
- `repository` (optional): Repository name to search in

**Example Usage:**
```
Find all functions named "processData"
Find all classes in Python files
Find variables containing "config" in Go files
```

#### 8. `get_file_content`
**Description:** Get the full content of a specific file
**Parameters:**
- `file_path` (required): Path to the file
- `repository` (optional): Repository name
- `start_line` (optional): Start line number (1-based)
- `end_line` (optional): End line number (1-based)

**Example Usage:**
```
Get full content of src/main.go
Get lines 10-50 of utils.py
```

#### 9. `list_directory`
**Description:** List files and directories in a specific path
**Parameters:**
- `directory_path` (required): Directory path to list
- `repository` (optional): Repository name
- `recursive` (optional): List recursively (default: false)
- `file_filter` (optional): File extension filter (e.g., '.go', '.py')

**Example Usage:**
```
List all files in src/ directory
List all Go files recursively in project
```

#### 10. `delete_lines`
**Description:** Delete a range of lines within a file
**Parameters:**
- `file_path` (required): Path to the file
- `start_line` (required): Start line number (1-based, inclusive)
- `end_line` (required): End line number (1-based, inclusive)

**Example Usage:**
```
Delete lines 10-20 from main.go
Remove function definition from lines 45-60
```

#### 11. `insert_at_line`
**Description:** Insert content at a given line in a file
**Parameters:**
- `file_path` (required): Path to the file
- `line_number` (required): Line number where to insert content (1-based)
- `content` (required): Content to insert (supports multi-line content)

**Example Usage:**
```
Insert a new function at line 50 in utils.go
Add import statement at line 5
```

#### 12. `replace_lines`
**Description:** Replace a range of lines within a file with new content
**Parameters:**
- `file_path` (required): Path to the file
- `start_line` (required): Start line number (1-based, inclusive)
- `end_line` (required): End line number (1-based, inclusive)
- `new_content` (required): New content to replace the lines (supports multi-line content)

**Example Usage:**
```
Replace function implementation in lines 25-40
Update configuration block from lines 10-15
```

#### 21. `get_file_snippet`
**Description:** Extract a specific code snippet from a file
**Parameters:**
- `file_path` (required): Path to the file
- `start_line` (required): Start line number (1-based, inclusive)
- `end_line` (required): End line number (1-based, inclusive)
- `include_context` (optional): Include surrounding context lines

**Example Usage:**
```
Extract lines 25-40 from main.go with context
Get function definition from lines 100-120
```

#### 22. `find_references`
**Description:** Find all references to a symbol across indexed repositories
**Parameters:**
- `symbol_name` (required): Symbol name to search for
- `symbol_type` (optional): Type of symbol (function, class, variable, etc.)
- `repository` (optional): Repository name to search in
- `include_definitions` (optional): Include symbol definitions in results

**Example Usage:**
```
Find all references to function "handleRequest"
Search for variable usage across all repositories
```

#### 23. `refresh_index`
**Description:** Refresh the search index for specific repositories or all repositories
**Parameters:**
- `repository` (optional): Repository name to refresh (if not provided, refresh all)
- `force_rebuild` (optional): Force complete rebuild of the index

**Example Usage:**
```
Refresh index for specific repository after changes
Force rebuild of entire search index
```

#### 24. `git_blame`
**Description:** Get Git blame information for a specific file or file range
**Parameters:**
- `file_path` (required): Path to the file
- `start_line` (optional): Start line number (1-based)
- `end_line` (optional): End line number (1-based)
- `repository` (optional): Repository name

**Example Usage:**
```
Get blame info for entire file
Show commit history for lines 50-100
```

### **Project Management Tools (5)**

#### 13. `get_current_config`
**Description:** Get the current configuration of the agent, including active projects, tools, contexts, and modes
**Parameters:** None

**Example Usage:**
```
Show current server configuration and status
Display available tools and project information
```

#### 14. `initial_instructions`
**Description:** Get the initial instructions for the current project (for environments where system prompt cannot be set)
**Parameters:** None

**Example Usage:**
```
Show getting started guide
Display available tools and usage examples
```

#### 15. `remove_project`
**Description:** Remove a project from the configuration
**Parameters:**
- `project_name` (required): Name of the project to remove

**Example Usage:**
```
Remove project "old-project" from configuration
Clean up unused project references
```

#### 16. `restart_language_server`
**Description:** Restart the language server (useful when external edits occur)
**Parameters:** None

**Example Usage:**
```
Restart language server after external file changes
Refresh code completion and analysis
```

#### 17. `summarize_changes`
**Description:** Provide instructions for summarizing codebase changes
**Parameters:** None

**Example Usage:**
```
Get guidelines for change summarization
Learn best practices for documenting modifications
```

### **Session Management Tools (3)**

#### 25. `list_sessions`
**Description:** List all active VSCode IDE sessions
**Parameters:** None

**Example Usage:**
```
Show all active VSCode sessions
List current IDE instances
```

#### 26. `create_session`
**Description:** Create a new VSCode IDE session
**Parameters:**
- `name` (required): Name for the new session
- `workspace_dir` (optional): Workspace directory for the session

**Example Usage:**
```
Create a new session called 'frontend-dev'
Start a new IDE session for /path/to/project
```

#### 27. `get_session_info`
**Description:** Get information about the current session and multi-session configuration
**Parameters:** None

**Example Usage:**
```
Show current session information
Get multi-session configuration details
```

### **AI Model Tools (3)**

#### 25. `generate_code`
**Description:** Generate code from natural language description using AI
**Parameters:**
- `prompt` (required): Natural language description of what the code should do
- `language` (required): Programming language (go, python, javascript, etc.)

**Example Usage:**
```
Generate a Go HTTP server function
Create a Python class for data processing
Write a JavaScript async function for API calls
```

#### 26. `analyze_code`
**Description:** Analyze code quality and get suggestions using AI
**Parameters:**
- `code` (required): Code to analyze
- `language` (required): Programming language

**Example Usage:**
```
Analyze this Go function for quality issues
Check this Python class for improvements
Review this JavaScript code for best practices
```

#### 27. `explain_code`
**Description:** Get AI explanation of code functionality
**Parameters:**
- `code` (required): Code to explain
- `language` (required): Programming language

**Example Usage:**
```
Explain what this function does
Describe the purpose of this class
Break down this algorithm step by step
```

## 🚀 **Usage Examples**

### **Finding Code**
```
"Find all test files in the project"
→ Uses find_files with pattern="*test*"

"Show me all HTTP handler functions"
→ Uses find_symbols with symbol_name="*handler*", symbol_type="function"

"List all Go files in the src directory"
→ Uses list_directory with directory_path="src", file_filter=".go"
```

### **Code Analysis**
```
"Search for error handling patterns"
→ Uses search_code with query="error handling", type="content"

"Get the content of the main configuration file"
→ Uses get_file_content with file_path="config/main.yaml"

"Show repository statistics"
→ Uses get_index_stats
```

### **AI-Powered Assistance**
```
"Generate a REST API endpoint in Go"
→ Uses generate_code with appropriate prompt and language="go"

"Analyze this function for performance issues"
→ Uses analyze_code with the function code

"Explain what this algorithm does"
→ Uses explain_code with the algorithm code
```

## 🔧 **Configuration**

The tools are automatically registered when the MCP server starts. Make sure your `config.yaml` has:

```yaml
models:
  enabled: true
  default_model: "code-assistant-v1"
  models_dir: "./models"
  max_tokens: 2048
  temperature: 0.7
```

## 📊 **Tool Categories Summary**

| Category | Count | Purpose |
|----------|-------|---------|
| **Core Indexing** | 5 | Repository indexing and basic search |
| **Utility** | 11 | File operations, symbol finding, file manipulation, and advanced code intelligence |
| **Project Management** | 5 | Configuration, instructions, and project management |
| **Session Management** | 3 | Multi-session support and VSCode instance management |
| **AI Models** | 3 | Code generation, analysis, and explanation |
| **Total** | **27** | **Complete multi-session code intelligence toolkit** |

## 🎯 **Next Steps**

These tools provide a solid foundation for code intelligence. You can extend them by:

1. **Adding more AI tools** (refactoring, debugging, testing)
2. **Implementing advanced search** (semantic search, code similarity)
3. **Adding project management tools** (dependency analysis, metrics)
4. **Integrating with external services** (GitHub, GitLab, CI/CD)

The MCP Code Indexer is now a powerful, extensible platform for intelligent code assistance!
