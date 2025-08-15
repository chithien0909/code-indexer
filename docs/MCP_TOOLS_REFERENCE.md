# MCP Code Indexer - Tools Quick Reference

This is a quick reference guide for all MCP tools provided by the Code Indexer server.

## Available Tools

### 1. `index_repository`
**Purpose:** Index a Git repository for searching

**Parameters:**
- `path` (required): Local path or Git URL to repository
- `name` (optional): Custom name for the repository

**Example Usage:**
```json
{
  "tool": "index_repository",
  "arguments": {
    "path": "/path/to/local/repo",
    "name": "my-project"
  }
}
```

**AI Prompt Examples:**
- "Please index my current project repository"
- "Index the repository at /home/user/projects/webapp"
- "Index the remote repository https://github.com/user/repo.git"

---

### 2. `search_code`
**Purpose:** Search across all indexed repositories

**Parameters:**
- `query` (required): Search query
- `type` (optional): Search type - `function`, `class`, `variable`, `content`, `file`, `comment`
- `language` (optional): Filter by programming language
- `repository` (optional): Filter by repository name
- `max_results` (optional): Maximum number of results (default: 100)

**Example Usage:**
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "authentication",
    "type": "function",
    "language": "go",
    "max_results": 20
  }
}
```

**AI Prompt Examples:**
- "Find all functions related to authentication"
- "Search for classes containing 'database' in Python files"
- "Find all variables named 'config' in the backend repository"
- "Search for comments mentioning 'TODO' or 'FIXME'"
- "Find all files with 'handler' in their name"

---

### 3. `get_metadata`
**Purpose:** Get detailed metadata for a specific file

**Parameters:**
- `file_path` (required): Path to the file
- `repository` (optional): Repository name

**Example Usage:**
```json
{
  "tool": "get_metadata",
  "arguments": {
    "file_path": "src/main.go",
    "repository": "my-project"
  }
}
```

**AI Prompt Examples:**
- "Show me the structure of the main.go file"
- "Get metadata for src/auth/login.py"
- "Analyze the functions and classes in utils/helpers.js"

---

### 4. `list_repositories`
**Purpose:** List all indexed repositories with statistics

**Parameters:** None

**Example Usage:**
```json
{
  "tool": "list_repositories",
  "arguments": {}
}
```

**AI Prompt Examples:**
- "List all indexed repositories"
- "Show me what repositories are available for searching"
- "Give me an overview of all indexed projects"

---

### 5. `get_index_stats`
**Purpose:** Get comprehensive indexing statistics

**Parameters:** None

**Example Usage:**
```json
{
  "tool": "get_index_stats",
  "arguments": {}
}
```

**AI Prompt Examples:**
- "Show me indexing statistics"
- "How many files and functions are indexed?"
- "Give me a breakdown of indexed content by language"

## Common AI Prompts and Expected Tool Usage

### Repository Management
| Prompt | Expected Tool(s) |
|--------|------------------|
| "Index my current project" | `index_repository` |
| "What repositories do I have indexed?" | `list_repositories` |
| "Show me indexing statistics" | `get_index_stats` |

### Code Search
| Prompt | Expected Tool(s) |
|--------|------------------|
| "Find all authentication functions" | `search_code` (type: function, query: authentication) |
| "Search for database classes in Python" | `search_code` (type: class, language: python, query: database) |
| "Find TODO comments" | `search_code` (type: comment, query: TODO) |
| "Search for error handling code" | `search_code` (query: error) |

### Code Analysis
| Prompt | Expected Tool(s) |
|--------|------------------|
| "Analyze the main.go file structure" | `get_metadata` (file_path: main.go) |
| "Show me all functions in auth.py" | `get_metadata` (file_path: auth.py) |
| "What's in the utils directory?" | `search_code` (query: utils, type: file) |

### Complex Workflows
| Prompt | Expected Tool(s) |
|--------|------------------|
| "Give me an overview of my codebase" | `list_repositories` + `get_index_stats` |
| "Find all API endpoints" | `search_code` (query: endpoint OR route OR handler) |
| "Show me all database models" | `search_code` (type: class, query: model) |

## Search Tips

### Effective Query Patterns
- **Function names**: Use specific function names or patterns
- **Class names**: Search for class names or inheritance patterns
- **Variable names**: Look for configuration, constants, or state variables
- **Comments**: Search for TODOs, FIXMEs, or documentation keywords
- **File paths**: Use partial paths or file name patterns

### Search Type Usage
- **`function`**: Methods, functions, procedures
- **`class`**: Classes, structs, interfaces, types
- **`variable`**: Variables, constants, fields, properties
- **`content`**: Any text content in files
- **`file`**: File names and paths
- **`comment`**: Comments and documentation

### Language Filters
Use language filters for polyglot projects:
- `go`, `python`, `javascript`, `typescript`, `java`, `cpp`, `rust`, `ruby`, `php`, `csharp`, `kotlin`, `swift`, `scala`, etc.

### Repository Filters
Use repository names to focus searches:
- Exact repository names from `list_repositories`
- Useful for multi-project workspaces

## Response Formats

### Successful Responses
All tools return JSON with:
- `success: true`
- Tool-specific data
- Optional metadata

### Error Responses
Failed operations return:
- `success: false` or error message
- Error details
- Suggested fixes (when applicable)

## Best Practices

### For AI Assistants
1. **Start with repository listing** to understand available code
2. **Use specific search types** for better results
3. **Combine tools** for comprehensive analysis
4. **Filter by language/repository** for focused searches
5. **Check metadata** for detailed file analysis

### For Users
1. **Index repositories first** before searching
2. **Use descriptive repository names** for easier filtering
3. **Re-index after major changes** to keep results current
4. **Monitor index size** and clean up when needed
5. **Use exclude patterns** to avoid indexing unnecessary files

## Troubleshooting

### No Search Results
1. Verify repository is indexed: `list_repositories`
2. Check index statistics: `get_index_stats`
3. Try broader search terms
4. Remove filters and search again

### Slow Performance
1. Reduce `max_results` parameter
2. Use more specific search types
3. Add language/repository filters
4. Check index size with `get_index_stats`

### Missing Files/Functions
1. Verify file extensions are supported
2. Check exclude patterns in configuration
3. Re-index the repository
4. Verify file permissions

For detailed troubleshooting, see [Integration Guide](CURSOR_AUGMENT_INTEGRATION.md#troubleshooting).
