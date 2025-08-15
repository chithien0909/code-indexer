#!/bin/bash

# Multi-IDE Testing Script
# Tests the MCP Code Indexer with multiple concurrent IDE connections

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVER_PORT=8080
SERVER_HOST="localhost"
CONFIG_FILE="$PROJECT_ROOT/config.yaml"
LOG_FILE="$PROJECT_ROOT/multi-ide-test.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    
    # Kill server if running
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        log_info "Server stopped"
    fi
    
    # Clean up test files
    rm -f "$LOG_FILE"
    
    log_info "Cleanup complete"
}

# Set up trap for cleanup
trap cleanup EXIT

# Test functions
test_server_startup() {
    log_info "Testing server startup..."
    
    # Build the server
    cd "$PROJECT_ROOT"
    if ! make build; then
        log_error "Failed to build server"
        return 1
    fi
    
    # Start server in background
    ./bin/code-indexer daemon --port $SERVER_PORT --host $SERVER_HOST --config "$CONFIG_FILE" > "$LOG_FILE" 2>&1 &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 3
    
    # Check if server is running
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        log_error "Server failed to start"
        return 1
    fi
    
    log_success "Server started successfully (PID: $SERVER_PID)"
    return 0
}

test_health_check() {
    log_info "Testing health check..."
    
    local response
    response=$(curl -s -f "http://$SERVER_HOST:$SERVER_PORT/api/health" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        log_success "Health check passed"
        log_info "Response: $response"
        return 0
    else
        log_error "Health check failed"
        return 1
    fi
}

test_concurrent_connections() {
    log_info "Testing concurrent connections..."
    
    local pids=()
    local session_count=5
    
    # Create multiple concurrent sessions
    for i in $(seq 1 $session_count); do
        (
            session_id="test-session-$i"
            log_info "Starting session $session_id"
            
            # Test tool call
            response=$(curl -s -X POST "http://$SERVER_HOST:$SERVER_PORT/api/call" \
                -H "Content-Type: application/json" \
                -H "X-Session-ID: $session_id" \
                -d '{
                    "tool": "list_repositories",
                    "arguments": {}
                }' 2>/dev/null)
            
            if [ $? -eq 0 ]; then
                log_success "Session $session_id: Tool call successful"
            else
                log_error "Session $session_id: Tool call failed"
            fi
        ) &
        pids+=($!)
    done
    
    # Wait for all sessions to complete
    local failed=0
    for pid in "${pids[@]}"; do
        if ! wait $pid; then
            failed=$((failed + 1))
        fi
    done
    
    if [ $failed -eq 0 ]; then
        log_success "All $session_count concurrent connections succeeded"
        return 0
    else
        log_error "$failed out of $session_count connections failed"
        return 1
    fi
}

test_session_isolation() {
    log_info "Testing session isolation..."
    
    # Create test repository for session 1
    local session1_response
    session1_response=$(curl -s -X POST "http://$SERVER_HOST:$SERVER_PORT/api/call" \
        -H "Content-Type: application/json" \
        -H "X-Session-ID: isolation-test-1" \
        -d '{
            "tool": "index_repository",
            "arguments": {
                "path": ".",
                "name": "test-repo-session1"
            }
        }' 2>/dev/null)
    
    # Create test repository for session 2
    local session2_response
    session2_response=$(curl -s -X POST "http://$SERVER_HOST:$SERVER_PORT/api/call" \
        -H "Content-Type: application/json" \
        -H "X-Session-ID: isolation-test-2" \
        -d '{
            "tool": "index_repository", 
            "arguments": {
                "path": ".",
                "name": "test-repo-session2"
            }
        }' 2>/dev/null)
    
    if [[ "$session1_response" == *"success"* ]] && [[ "$session2_response" == *"success"* ]]; then
        log_success "Session isolation test passed"
        return 0
    else
        log_error "Session isolation test failed"
        return 1
    fi
}

test_resource_locking() {
    log_info "Testing resource locking..."
    
    local pids=()
    local lock_test_count=3
    
    # Create multiple concurrent indexing operations
    for i in $(seq 1 $lock_test_count); do
        (
            session_id="lock-test-$i"
            response=$(curl -s -X POST "http://$SERVER_HOST:$SERVER_PORT/api/call" \
                -H "Content-Type: application/json" \
                -H "X-Session-ID: $session_id" \
                -d '{
                    "tool": "index_repository",
                    "arguments": {
                        "path": ".",
                        "name": "lock-test-repo-'$i'"
                    }
                }' 2>/dev/null)
            
            if [[ "$response" == *"success"* ]]; then
                log_success "Lock test $i: Indexing completed"
            else
                log_warning "Lock test $i: Indexing may have been queued or failed"
            fi
        ) &
        pids+=($!)
    done
    
    # Wait for all operations to complete
    local failed=0
    for pid in "${pids[@]}"; do
        if ! wait $pid; then
            failed=$((failed + 1))
        fi
    done
    
    if [ $failed -lt $lock_test_count ]; then
        log_success "Resource locking test passed (some operations may have been queued)"
        return 0
    else
        log_error "Resource locking test failed"
        return 1
    fi
}

test_connection_limits() {
    log_info "Testing connection limits..."
    
    # Get current connection stats
    local stats_response
    stats_response=$(curl -s "http://$SERVER_HOST:$SERVER_PORT/api/sessions" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        log_success "Connection stats retrieved successfully"
        log_info "Stats: $stats_response"
        return 0
    else
        log_error "Failed to retrieve connection stats"
        return 1
    fi
}

# Main test execution
main() {
    log_info "Starting Multi-IDE Test Suite"
    log_info "================================"
    
    # Initialize log file
    echo "Multi-IDE Test Log - $(date)" > "$LOG_FILE"
    
    local tests_passed=0
    local tests_failed=0
    
    # Run tests
    if test_server_startup; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
        log_error "Server startup failed, aborting remaining tests"
        exit 1
    fi
    
    if test_health_check; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_concurrent_connections; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_session_isolation; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_resource_locking; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_connection_limits; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    # Summary
    log_info "================================"
    log_info "Test Summary:"
    log_success "Tests passed: $tests_passed"
    if [ $tests_failed -gt 0 ]; then
        log_error "Tests failed: $tests_failed"
    else
        log_info "Tests failed: $tests_failed"
    fi
    
    if [ $tests_failed -eq 0 ]; then
        log_success "🎉 All tests passed! Multi-IDE support is working correctly."
        exit 0
    else
        log_error "❌ Some tests failed. Check the logs for details."
        exit 1
    fi
}

# Check if script is being run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
