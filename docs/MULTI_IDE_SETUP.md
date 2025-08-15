# Multi-IDE Setup Guide

This guide explains how to configure and use the MCP Code Indexer with multiple IDE instances running simultaneously.

## Overview

The MCP Code Indexer now supports multiple IDE instances (Cursor, VS Code, etc.) connecting concurrently without conflicts. This enables teams to:

- Share the same indexed codebase across multiple developers
- Use different IDEs simultaneously on the same project
- Maintain session isolation while sharing resources efficiently
- Scale to support large development teams

## Prerequisites

- MCP Code Indexer v1.1.0 or later
- IDEs with MCP protocol support (Cursor, VS Code with MCP extension, etc.)
- Network connectivity between IDEs and the indexer server

## Configuration

### 1. Server Configuration

Create or update your `config.yaml` file:

```yaml
server:
  name: "Code Indexer"
  version: "1.1.0"
  enable_recovery: true
  
  # Multi-IDE Configuration
  multi_ide:
    enabled: true
    max_connections: 50
    connection_timeout_seconds: 300  # 5 minutes
    cleanup_interval_minutes: 5
    transport_types: ["http", "websocket"]
    
    # Resource Management
    resource_management:
      isolation_mode: "workspace"  # Options: shared, workspace, full
      max_concurrent_operations: 10
      operation_timeout_minutes: 5
      enable_operation_queue: true
    
    # Locking Configuration
    locking:
      enable_fine_grained_locks: true
      lock_timeout_seconds: 30
      enable_deadlock_detection: true
    
    # Monitoring
    monitoring:
      enable_metrics: true
      log_connections: true
      performance_tracking: true

# Enhanced logging for multi-IDE debugging
logging:
  level: info
  file: "indexer.log"
  json_format: true

# Search configuration optimized for concurrent access
search:
  max_results: 100
  highlight_snippets: true
  snippet_length: 200
  fuzzy_tolerance: 0.2
```

### 2. Isolation Modes

Choose the appropriate isolation mode based on your needs:

#### Shared Mode (`isolation_mode: "shared"`)
- All IDEs share the same index and resources
- Best performance, minimal resource usage
- Suitable for small teams working on the same codebase

#### Workspace Mode (`isolation_mode: "workspace"`)
- Each workspace gets its own index partition
- Good balance of isolation and resource sharing
- **Recommended for most use cases**

#### Full Mode (`isolation_mode: "full"`)
- Complete isolation between IDE sessions
- Highest resource usage but maximum isolation
- Suitable for multi-tenant environments

## IDE Configuration

### Cursor IDE

Add to your Cursor settings (`.cursor/settings.json`):

```json
{
  "mcp": {
    "servers": {
      "code-indexer": {
        "transport": "http",
        "endpoint": "http://localhost:8080/api/call",
        "headers": {
          "Content-Type": "application/json",
          "X-Session-ID": "cursor-${workspaceFolder}"
        },
        "timeout": 30000
      }
    }
  }
}
```

### VS Code

Install the MCP extension and configure in `settings.json`:

```json
{
  "mcp.servers": {
    "code-indexer": {
      "transport": "http",
      "endpoint": "http://localhost:8080/api/call",
      "headers": {
        "Content-Type": "application/json",
        "X-Session-ID": "vscode-${workspaceFolder}"
      },
      "timeout": 30000
    }
  }
}
```

### Other MCP-Compatible IDEs

For other IDEs, use the HTTP transport with these parameters:

- **Endpoint**: `http://localhost:8080/api/call`
- **Method**: `POST`
- **Headers**: 
  - `Content-Type: application/json`
  - `X-Session-ID: <unique-session-id>`
- **Timeout**: 30 seconds

## Starting the Server

### Daemon Mode (Recommended for Multi-IDE)

```bash
# Start the daemon server
./bin/code-indexer daemon --port 8080 --host 0.0.0.0

# With custom configuration
./bin/code-indexer daemon --port 8080 --config multi-ide-config.yaml

# With debug logging
./bin/code-indexer daemon --port 8080 --log-level debug
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/code-indexer .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["./code-indexer", "daemon", "--port", "8080", "--host", "0.0.0.0"]
```

```bash
# Build and run
docker build -t code-indexer-multi .
docker run -p 8080:8080 -v $(pwd)/repositories:/root/repositories code-indexer-multi
```

## Verification and Testing

### 1. Health Check

```bash
curl http://localhost:8080/api/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.1.0",
  "multi_ide_enabled": true,
  "active_connections": 0,
  "uptime": "5m30s"
}
```

### 2. Connection Testing

```bash
# Test tool call
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: test-session" \
  -d '{
    "tool": "list_repositories",
    "arguments": {}
  }'
```

### 3. Multi-IDE Test

1. Open the same project in multiple IDEs
2. Index a repository from one IDE
3. Search for code from another IDE
4. Verify both IDEs can access the indexed data

## Monitoring and Troubleshooting

### Connection Monitoring

```bash
# List active connections
curl http://localhost:8080/api/sessions

# Connection statistics
curl http://localhost:8080/api/stats/connections
```

### Performance Monitoring

```bash
# Lock statistics
curl http://localhost:8080/api/stats/locks

# Performance metrics
curl http://localhost:8080/api/stats/performance
```

### Log Analysis

```bash
# Monitor connections
tail -f indexer.log | jq 'select(.msg | contains("connection"))'

# Monitor locks
tail -f indexer.log | jq 'select(.msg | contains("lock"))'

# Monitor performance
tail -f indexer.log | jq 'select(.level == "warn" or .level == "error")'
```

## Best Practices

### 1. Session Management
- Use unique session IDs for each IDE instance
- Include workspace information in session IDs
- Clean up inactive sessions regularly

### 2. Resource Optimization
- Use workspace isolation mode for most scenarios
- Configure appropriate connection limits
- Monitor resource usage regularly

### 3. Network Configuration
- Use localhost for single-machine setups
- Configure firewall rules for network access
- Consider using HTTPS in production

### 4. Troubleshooting
- Check logs for connection and lock issues
- Monitor resource usage and performance
- Use health checks to verify server status

## Security Considerations

### 1. Network Security
- Bind to localhost for local-only access
- Use firewall rules to restrict network access
- Consider VPN for remote access

### 2. Authentication (Future Enhancement)
- Session-based authentication
- API key validation
- Role-based access control

### 3. Data Protection
- Secure index file storage
- Encrypted communication (HTTPS)
- Access logging and auditing

## Scaling and Performance

### 1. Connection Limits
- Default: 50 concurrent connections
- Adjust based on available resources
- Monitor connection usage patterns

### 2. Resource Management
- Configure operation timeouts appropriately
- Use operation queuing for high load
- Monitor lock contention

### 3. Hardware Requirements
- CPU: 2+ cores recommended for concurrent access
- RAM: 4GB+ for large codebases
- Storage: SSD recommended for index files
- Network: Gigabit for multiple remote connections
