# Multi-IDE Architecture Design

## Overview

This document outlines the architecture changes needed to support multiple IDE instances running concurrently with the MCP Code Indexer without conflicts.

## Current Limitations

1. **Single Connection Model**: Current `serve` mode only supports one stdio connection
2. **Shared Resource Conflicts**: Multiple IDEs can conflict when accessing index files or repositories
3. **No Connection Management**: Limited connection pooling and session isolation
4. **Resource Locking Issues**: Insufficient locking mechanisms for concurrent operations

## Proposed Architecture

### 1. Connection Management Layer

#### Multi-Connection Server
- **Enhanced Daemon Mode**: Upgrade daemon mode to handle multiple concurrent MCP connections
- **Connection Pool**: Implement connection pooling with per-connection session isolation
- **WebSocket Support**: Add WebSocket transport for better real-time communication
- **Connection Lifecycle**: Proper connection tracking and cleanup

#### Connection Types
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Cursor IDE    │    │   VS Code IDE   │    │  Other MCP IDE  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ├───────────────────────┼───────────────────────┤
         │              MCP Protocol Layer               │
         └─────────────────────────────────────────────────┘
                                │
         ┌─────────────────────────────────────────────────┐
         │           Connection Manager                    │
         │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
         │  │ Connection 1│ │ Connection 2│ │ Connection N││
         │  │ (Session A) │ │ (Session B) │ │ (Session C) ││
         │  └─────────────┘ └─────────────┘ └─────────────┘│
         └─────────────────────────────────────────────────┘
```

### 2. Resource Management Layer

#### Shared Resource Manager
- **Index Lock Manager**: Fine-grained locking for search index operations
- **Repository Lock Manager**: Coordinate repository access and cloning
- **File System Locks**: Prevent concurrent file operations conflicts
- **Operation Queue**: Queue conflicting operations for sequential execution

#### Resource Isolation Levels
1. **Shared Index Mode**: All IDEs share the same search index (default)
2. **Workspace Isolation**: Each workspace gets its own index partition
3. **Full Isolation**: Each IDE session gets completely separate resources

### 3. Session Management Enhancement

#### Enhanced Session Manager
- **Connection-Session Mapping**: Map each connection to a unique session
- **Session Persistence**: Maintain session state across connection drops
- **Workspace Detection**: Automatically detect and isolate workspaces
- **Session Cleanup**: Automatic cleanup of inactive sessions

#### Session Isolation Strategies
```yaml
isolation_mode: "workspace"  # Options: shared, workspace, full
workspace_detection: "auto"  # Auto-detect workspace boundaries
session_timeout: "2h"        # Session timeout
cleanup_interval: "30m"      # Cleanup check interval
```

### 4. Concurrency Control

#### Operation Coordination
- **Read-Write Locks**: Separate read/write access to shared resources
- **Operation Priorities**: Priority system for different operation types
- **Deadlock Prevention**: Ordered lock acquisition to prevent deadlocks
- **Timeout Handling**: Operation timeouts to prevent hanging

#### Lock Hierarchy
```
Global Server Lock
├── Index Manager Lock
│   ├── Repository Index Locks
│   └── Search Operation Locks
├── Repository Manager Lock
│   ├── Clone Operation Locks
│   └── File Access Locks
└── Session Manager Lock
    ├── Session Creation Locks
    └── Session Cleanup Locks
```

## Implementation Plan

### Phase 1: Enhanced Connection Management
1. Upgrade daemon mode to handle multiple concurrent connections
2. Implement connection pooling and session mapping
3. Add WebSocket transport support
4. Enhance connection lifecycle management

### Phase 2: Resource Locking System
1. Implement fine-grained locking for search index
2. Add repository operation coordination
3. Create operation queue system
4. Add deadlock prevention mechanisms

### Phase 3: Session Isolation
1. Enhance session manager for better isolation
2. Implement workspace detection and isolation
3. Add session persistence and recovery
4. Create session-aware resource allocation

### Phase 4: Configuration and Monitoring
1. Add multi-IDE configuration options
2. Implement connection and session monitoring
3. Add performance metrics and logging
4. Create troubleshooting tools

## Configuration Schema

```yaml
server:
  multi_ide:
    enabled: true
    max_connections: 50
    connection_timeout: "30s"
    transport_types: ["http", "websocket"]
    
  resource_management:
    isolation_mode: "workspace"  # shared, workspace, full
    max_concurrent_operations: 10
    operation_timeout: "5m"
    enable_operation_queue: true
    
  locking:
    enable_fine_grained_locks: true
    lock_timeout: "30s"
    deadlock_detection: true
    
  monitoring:
    enable_metrics: true
    log_connections: true
    performance_tracking: true
```

## Benefits

1. **True Concurrent Support**: Multiple IDEs can work simultaneously without conflicts
2. **Resource Safety**: Proper locking prevents data corruption and conflicts
3. **Performance**: Optimized resource sharing and operation coordination
4. **Scalability**: Support for many concurrent IDE connections
5. **Reliability**: Robust error handling and recovery mechanisms
6. **Flexibility**: Configurable isolation levels based on needs

## Migration Path

1. **Backward Compatibility**: Existing single-IDE setups continue to work
2. **Gradual Adoption**: Teams can migrate one IDE at a time
3. **Configuration Migration**: Automatic migration of existing configurations
4. **Testing Support**: Tools to test multi-IDE scenarios

## Quick Start Guide

### 1. Enable Multi-IDE Support

Update your `config.yaml`:

```yaml
server:
  multi_ide:
    enabled: true
    max_connections: 50
    connection_timeout_seconds: 300
    transport_types: ["http", "websocket"]

    resource_management:
      isolation_mode: "workspace"  # shared, workspace, full
      max_concurrent_operations: 10
      operation_timeout_minutes: 5
      enable_operation_queue: true

    locking:
      enable_fine_grained_locks: true
      lock_timeout_seconds: 30
      enable_deadlock_detection: true

    monitoring:
      enable_metrics: true
      log_connections: true
      performance_tracking: true
```

### 2. Start the Enhanced Server

```bash
# Start with multi-IDE support
./bin/code-indexer daemon --port 8080

# Or with custom config
./bin/code-indexer daemon --port 8080 --config multi-ide-config.yaml
```

### 3. Configure IDEs

#### Cursor IDE Configuration
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:8080/api/call",
        "-H", "Content-Type: application/json",
        "-d", "@-"
      ],
      "env": {
        "MCP_SESSION_ID": "cursor-session-1"
      }
    }
  }
}
```

#### VS Code Configuration
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:8080/api/call",
        "-H", "Content-Type: application/json",
        "-d", "@-"
      ],
      "env": {
        "MCP_SESSION_ID": "vscode-session-1"
      }
    }
  }
}
```

### 4. Verify Multi-IDE Operation

```bash
# Check server status
curl http://localhost:8080/api/health

# List active connections
curl http://localhost:8080/api/sessions

# Monitor logs
tail -f indexer.log | grep -E "(connection|session|lock)"
```

## Troubleshooting

### Common Issues

1. **Connection Limit Reached**
   - Increase `max_connections` in config
   - Check for zombie connections
   - Verify cleanup intervals

2. **Resource Lock Timeouts**
   - Increase `lock_timeout_seconds`
   - Check for deadlocks in logs
   - Verify operation timeouts

3. **Session Isolation Issues**
   - Verify `isolation_mode` setting
   - Check workspace detection
   - Review session logs

### Monitoring Commands

```bash
# Connection statistics
curl http://localhost:8080/api/stats/connections

# Lock statistics
curl http://localhost:8080/api/stats/locks

# Performance metrics
curl http://localhost:8080/api/stats/performance
```
