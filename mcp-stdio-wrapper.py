#!/usr/bin/env python3
"""
MCP stdio wrapper that bridges stdio MCP protocol to daemon mode.
This is a workaround for stdio communication issues.
"""

import json
import sys
import subprocess
import time
import requests
import signal
import os
from typing import Dict, Any

class MCPStdioWrapper:
    def __init__(self):
        self.daemon_process = None
        self.daemon_port = 9991
        self.daemon_url = f"http://localhost:{self.daemon_port}"
        
    def start_daemon(self):
        """Start the MCP daemon in the background."""
        try:
            # Start daemon
            cmd = [
                "uvx", "--from", "git+https://github.com/chithien0909/code-indexer.git",
                "code-indexer", "daemon", "--port", str(self.daemon_port)
            ]
            
            self.daemon_process = subprocess.Popen(
                cmd,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL
            )
            
            # Wait for daemon to start
            for _ in range(30):  # Wait up to 30 seconds
                try:
                    response = requests.get(f"{self.daemon_url}/api/health", timeout=1)
                    if response.status_code == 200:
                        return True
                except:
                    time.sleep(1)
                    
            return False
        except Exception as e:
            print(f"Failed to start daemon: {e}", file=sys.stderr)
            return False
    
    def stop_daemon(self):
        """Stop the daemon process."""
        if self.daemon_process:
            self.daemon_process.terminate()
            self.daemon_process.wait()
    
    def handle_initialize(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle MCP initialize request."""
        return {
            "jsonrpc": "2.0",
            "id": request.get("id"),
            "result": {
                "protocolVersion": "2024-11-05",
                "capabilities": {
                    "tools": {}
                },
                "serverInfo": {
                    "name": "Code Indexer",
                    "version": "1.1.0"
                }
            }
        }
    
    def handle_tools_list(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle tools/list request."""
        try:
            response = requests.get(f"{self.daemon_url}/api/tools", timeout=5)
            if response.status_code == 200:
                tools_data = response.json()
                tools = tools_data.get("tools", [])
                
                return {
                    "jsonrpc": "2.0",
                    "id": request.get("id"),
                    "result": {
                        "tools": tools
                    }
                }
            else:
                raise Exception(f"API returned {response.status_code}")
                
        except Exception as e:
            return {
                "jsonrpc": "2.0",
                "id": request.get("id"),
                "error": {
                    "code": -32603,
                    "message": f"Failed to get tools: {e}"
                }
            }
    
    def handle_tool_call(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle tools/call request."""
        try:
            params = request.get("params", {})
            tool_name = params.get("name")
            arguments = params.get("arguments", {})
            
            call_data = {
                "tool": tool_name,
                "arguments": arguments
            }
            
            response = requests.post(
                f"{self.daemon_url}/api/call",
                json=call_data,
                timeout=30
            )
            
            if response.status_code == 200:
                result_data = response.json()
                return {
                    "jsonrpc": "2.0",
                    "id": request.get("id"),
                    "result": {
                        "content": [
                            {
                                "type": "text",
                                "text": json.dumps(result_data.get("result", {}), indent=2)
                            }
                        ]
                    }
                }
            else:
                raise Exception(f"API returned {response.status_code}")
                
        except Exception as e:
            return {
                "jsonrpc": "2.0",
                "id": request.get("id"),
                "error": {
                    "code": -32603,
                    "message": f"Tool call failed: {e}"
                }
            }
    
    def handle_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle incoming MCP request."""
        method = request.get("method")
        
        if method == "initialize":
            return self.handle_initialize(request)
        elif method == "tools/list":
            return self.handle_tools_list(request)
        elif method == "tools/call":
            return self.handle_tool_call(request)
        else:
            return {
                "jsonrpc": "2.0",
                "id": request.get("id"),
                "error": {
                    "code": -32601,
                    "message": f"Method not found: {method}"
                }
            }
    
    def run(self):
        """Run the stdio wrapper."""
        # Setup signal handlers
        signal.signal(signal.SIGINT, lambda s, f: self.stop_daemon())
        signal.signal(signal.SIGTERM, lambda s, f: self.stop_daemon())
        
        # Start daemon
        if not self.start_daemon():
            print("Failed to start daemon", file=sys.stderr)
            return 1
        
        try:
            # Process stdin line by line
            for line in sys.stdin:
                line = line.strip()
                if not line:
                    continue
                
                try:
                    request = json.loads(line)
                    response = self.handle_request(request)
                    print(json.dumps(response))
                    sys.stdout.flush()
                except json.JSONDecodeError:
                    # Invalid JSON, ignore
                    continue
                except Exception as e:
                    # Send error response
                    error_response = {
                        "jsonrpc": "2.0",
                        "id": None,
                        "error": {
                            "code": -32603,
                            "message": f"Internal error: {e}"
                        }
                    }
                    print(json.dumps(error_response))
                    sys.stdout.flush()
        
        finally:
            self.stop_daemon()
        
        return 0

if __name__ == "__main__":
    wrapper = MCPStdioWrapper()
    sys.exit(wrapper.run())
