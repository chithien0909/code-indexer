# MCP Server Modular Architecture

This directory contains the modular MCP (Model Context Protocol) server implementation, split into focused files for better maintainability and organization.

## 📁 **File Structure**

```
internal/server/
├── server.go           # Main server struct and initialization
├── tools.go            # Tool registration and configuration
├── handlers_core.go    # Core tool handlers (indexing, search, metadata)
├── handlers_utility.go # Utility tool handlers (file operations, symbols)
├── handlers_ai.go      # AI model tool handlers (generate, analyze, explain)
├── helpers.go          # Helper methods and utilities
└── README.md          # This documentation
```

## 🏗️ **Architecture Overview**

### **server.go** - Core Server
- **MCPServer struct** - Main server structure with all dependencies
- **New()** - Server initialization and component setup
- **Serve()** - Start the MCP server with stdio transport
- **Close()** - Graceful shutdown and cleanup

### **tools.go** - Tool Registration
- **registerTools()** - Main tool registration orchestrator
- **registerCoreTools()** - Register indexing and search tools (5 tools)
- **registerUtilityTools()** - Register file operation tools (4 tools)
- **registerModelTools()** - Register AI model tools (3 tools)

### **handlers_core.go** - Core Functionality
**5 Core Tools:**
- `handleIndexRepository` - Index Git repositories
- `handleSearchCode` - Search across indexed code
- `handleGetMetadata` - Get file metadata
- `handleListRepositories` - List indexed repositories
- `handleGetIndexStats` - Get indexing statistics

### **handlers_utility.go** - File Operations
**4 Utility Tools:**
- `handleFindFiles` - Find files by pattern with wildcards
- `handleFindSymbols` - Find symbols (functions, classes, variables)
- `handleGetFileContent` - Get file content with line range support
- `handleListDirectory` - List directory contents recursively

### **handlers_ai.go** - AI Capabilities
**3 AI Tools:**
- `handleGenerateCode` - Generate code from natural language
- `handleAnalyzeCode` - Analyze code quality and get suggestions
- `handleExplainCode` - Explain code functionality

### **helpers.go** - Utilities
- `getBooleanValue()` - Extract boolean values from MCP requests
- `getArguments()` - Extract arguments from MCP requests
- `listDirectoryContents()` - Directory listing with filtering

## 🔧 **Key Features**

### **Separation of Concerns**
- **Server Logic** - Isolated in server.go
- **Tool Registration** - Centralized in tools.go
- **Handler Logic** - Grouped by functionality
- **Utilities** - Shared helpers in helpers.go

### **Maintainability**
- **Single Responsibility** - Each file has a clear purpose
- **Easy Navigation** - Find handlers by feature category
- **Modular Testing** - Test handlers independently
- **Clear Dependencies** - Explicit imports and interfaces

### **Scalability**
- **Easy Extension** - Add new tools in appropriate handler files
- **Feature Grouping** - Related functionality stays together
- **Clean Interfaces** - Well-defined method signatures
- **Consistent Patterns** - Uniform error handling and logging

## 🚀 **Adding New Tools**

### **1. Core Tools (Indexing/Search)**
Add to `handlers_core.go`:
```go
func (s *MCPServer) handleNewCoreTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Implementation
}
```

Register in `tools.go` → `registerCoreTools()`:
```go
newTool := mcp.NewTool("new_core_tool", ...)
s.server.AddTool(newTool, s.handleNewCoreTool)
```

### **2. Utility Tools (File Operations)**
Add to `handlers_utility.go`:
```go
func (s *MCPServer) handleNewUtilityTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Implementation
}
```

Register in `tools.go` → `registerUtilityTools()`:
```go
newTool := mcp.NewTool("new_utility_tool", ...)
s.server.AddTool(newTool, s.handleNewUtilityTool)
```

### **3. AI Tools (Model Operations)**
Add to `handlers_ai.go`:
```go
func (s *MCPServer) handleNewAITool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Implementation
}
```

Register in `tools.go` → `registerModelTools()`:
```go
newTool := mcp.NewTool("new_ai_tool", ...)
s.server.AddTool(newTool, s.handleNewAITool)
```

## 📊 **Current Tool Count**

| Category | File | Tools | Description |
|----------|------|-------|-------------|
| **Core** | handlers_core.go | 5 | Indexing, search, metadata |
| **Utility** | handlers_utility.go | 4 | File operations, symbols |
| **AI** | handlers_ai.go | 3 | Code generation, analysis |
| **Total** | | **12** | **Complete MCP toolkit** |

## 🎯 **Benefits of Modular Structure**

### **Development**
- ✅ **Faster Navigation** - Find code by feature
- ✅ **Easier Debugging** - Isolated functionality
- ✅ **Cleaner Diffs** - Changes affect specific files
- ✅ **Better Testing** - Test handlers independently

### **Maintenance**
- ✅ **Reduced Complexity** - Smaller, focused files
- ✅ **Clear Ownership** - Each file has specific responsibility
- ✅ **Easy Refactoring** - Modify features without affecting others
- ✅ **Documentation** - Self-documenting structure

### **Collaboration**
- ✅ **Parallel Development** - Multiple developers can work simultaneously
- ✅ **Code Reviews** - Focused reviews on specific functionality
- ✅ **Knowledge Transfer** - Clear structure for new team members
- ✅ **Feature Ownership** - Teams can own specific handler files

## 🔄 **Migration from Monolithic**

The original 907-line `server.go` has been split into:
- **server.go** (95 lines) - Core server logic
- **tools.go** (230 lines) - Tool registration
- **handlers_core.go** (150 lines) - Core handlers
- **handlers_utility.go** (290 lines) - Utility handlers
- **handlers_ai.go** (85 lines) - AI handlers
- **helpers.go** (85 lines) - Helper utilities

**Total: ~935 lines** (slightly more due to better documentation and structure)

This modular approach provides better maintainability while preserving all functionality and improving code organization! 🎉
