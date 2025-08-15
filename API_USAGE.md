# MCP Code Indexer - API Usage Guide

## ✅ **Multi-Session API Successfully Implemented**

The MCP Code Indexer now provides HTTP API endpoints that allow multiple VSCode instances to connect to a single daemon server, enabling true multi-session support.

## 🚀 **Starting the Daemon Server**

### **1. Start the Daemon**
```bash
# Start the daemon on localhost:8080
./bin/code-indexer daemon --port 8080 --host localhost

# Or with custom configuration
./bin/code-indexer daemon --port 8080 --host localhost --config config.yaml
```

### **2. Verify Server is Running**
```bash
curl http://localhost:8080/api/health
```

## 📡 **API Endpoints**

### **Base URL:** `http://localhost:8080`

### **1. Health Check - `/api/health`**
**Method:** GET  
**Description:** Check server health and status

```bash
curl http://localhost:8080/api/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-15T15:15:40+07:00",
  "version": "1.0.0",
  "uptime": "694ns",
  "sessions": {
    "active_sessions": 0,
    "inactive_sessions": 0,
    "total_sessions": 0
  }
}
```

### **2. List Tools - `/api/tools`**
**Method:** GET  
**Description:** Get all available tools and server information

```bash
curl http://localhost:8080/api/tools
```

**Response:**
```json
{
  "tools": [...],
  "total": 27,
  "categories": {
    "core": 5,
    "utility": 11,
    "project": 5,
    "session": 3,
    "ai": 3
  },
  "server_info": {
    "name": "Code Indexer",
    "version": "1.0.0",
    "multi_session": true
  }
}
```

### **3. Execute Tool - `/api/call`**
**Method:** POST  
**Description:** Execute any MCP tool

```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "tool_name",
    "arguments": {...},
    "session_id": "optional_session_id"
  }'
```

**Request Body:**
- `tool` (string, required): Name of the tool to execute
- `arguments` (object, required): Tool-specific arguments
- `session_id` (string, optional): Session identifier for multi-session support

**Response:**
```json
{
  "success": true,
  "tool": "tool_name",
  "result": {...}
}
```

### **4. Session Management - `/api/sessions`**
**Method:** GET, POST  
**Description:** Manage VSCode IDE sessions

#### **List Sessions (GET)**
```bash
curl http://localhost:8080/api/sessions
```

**Response:**
```json
{
  "sessions": [],
  "stats": {
    "active_sessions": 0,
    "inactive_sessions": 0,
    "total_sessions": 0
  }
}
```

#### **Create Session (POST)**
```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-session",
    "workspace_dir": "/path/to/workspace"
  }'
```

**Response:**
```json
{
  "success": true,
  "session": {
    "id": "session-uuid",
    "name": "my-session",
    "workspace_dir": "/path/to/workspace",
    "created_at": "2025-08-15T15:15:40+07:00",
    "active": true
  },
  "message": "Session 'my-session' created successfully"
}
```

## 🛠️ **Tool Examples**

### **1. Session Management Tools**

#### **List Sessions**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "list_sessions", "arguments": {}}'
```

#### **Get Session Info**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "get_session_info", "arguments": {}}'
```

### **2. Core Tools**

#### **Index Repository**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "index_repository",
    "arguments": {
      "path": "/path/to/repository",
      "name": "my-repo"
    }
  }'
```

#### **Search Code**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "search_code",
    "arguments": {
      "query": "function main",
      "limit": 10
    }
  }'
```

#### **Get Index Stats**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "get_index_stats", "arguments": {}}'
```

### **3. Utility Tools**

#### **Find Files**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "find_files",
    "arguments": {
      "pattern": "*.go",
      "repository": "my-repo"
    }
  }'
```

#### **Get File Content**
```bash
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "get_file_content",
    "arguments": {
      "file_path": "main.go",
      "start_line": 1,
      "end_line": 50
    }
  }'
```

## 🔧 **VSCode Integration**

### **Option 1: Direct API Calls**
You can create a VSCode extension that makes HTTP calls to the daemon:

```javascript
// VSCode extension example
const response = await fetch('http://localhost:8080/api/call', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    tool: 'search_code',
    arguments: { query: 'function main' },
    session_id: vscode.workspace.name
  })
});
const result = await response.json();
```

### **Option 2: MCP Client Wrapper**
Create a wrapper that translates MCP calls to HTTP API calls:

```json
{
  "mcpServers": {
    "CodeIndexer": {
      "command": "node",
      "args": ["mcp-http-client.js", "http://localhost:8080"]
    }
  }
}
```

## 🎯 **Multi-Session Workflow**

### **1. Start Daemon Server**
```bash
# Terminal 1: Start the daemon
./bin/code-indexer daemon --port 8080
```

### **2. Multiple VSCode Instances**
Each VSCode instance can:
- Create its own session
- Make tool calls with session context
- Share indexed data efficiently

### **3. Session-Aware Operations**
```bash
# Create session for frontend project
curl -X POST http://localhost:8080/api/sessions \
  -d '{"name": "frontend", "workspace_dir": "/path/to/frontend"}'

# Create session for backend project  
curl -X POST http://localhost:8080/api/sessions \
  -d '{"name": "backend", "workspace_dir": "/path/to/backend"}'

# Make session-specific tool calls
curl -X POST http://localhost:8080/api/call \
  -d '{
    "tool": "search_code",
    "arguments": {"query": "React component"},
    "session_id": "frontend-session-id"
  }'
```

## 📊 **Benefits**

### **✅ Single Server Instance**
- One daemon serves multiple VSCode instances
- Efficient resource usage
- Shared indexed data across sessions

### **✅ HTTP API Interface**
- Standard REST API endpoints
- Easy integration with any client
- Language-agnostic access

### **✅ Session Isolation**
- Each VSCode instance has its own session
- Workspace-specific contexts
- Independent configurations

### **✅ Real-time Operations**
- Immediate tool execution
- Live session management
- Dynamic session creation

## 🚀 **Production Deployment**

### **1. Systemd Service**
```ini
[Unit]
Description=MCP Code Indexer Daemon
After=network.target

[Service]
Type=simple
User=codeindexer
ExecStart=/usr/local/bin/code-indexer daemon --port 8080 --config /etc/code-indexer/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### **2. Docker Container**
```dockerfile
FROM golang:1.21-alpine AS builder
COPY . /app
WORKDIR /app
RUN go build -o code-indexer ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/code-indexer /usr/local/bin/
EXPOSE 8080
CMD ["code-indexer", "daemon", "--port", "8080", "--host", "0.0.0.0"]
```

### **3. Load Balancer**
For high availability, you can run multiple daemon instances behind a load balancer.

## 🎉 **Ready for Multi-Session Development!**

The MCP Code Indexer daemon with HTTP API endpoints is now ready to serve multiple VSCode instances simultaneously, providing efficient multi-session development support with complete session isolation and shared resource optimization.
