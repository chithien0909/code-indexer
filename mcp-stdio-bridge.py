#!/usr/bin/env python3
"""
MCP stdio bridge that converts stdio MCP protocol to HTTP daemon calls.
This bridges the gap between IDE stdio expectations and our working daemon mode.
"""

import json
import sys
import subprocess
import time
import urllib.request
import urllib.parse
import urllib.error
import signal
import os
import threading
from typing import Dict, Any, Optional

class MCPStdioBridge:
    def __init__(self):
        self.daemon_process: Optional[subprocess.Popen] = None
        self.daemon_port = 9991
        self.daemon_url = f"http://localhost:{self.daemon_port}"
        self.initialized = False
        
    def start_daemon(self) -> bool:
        """Start the MCP daemon in the background."""
        try:
            # Start daemon process
            cmd = [
                "uvx", "--from", "git+https://github.com/chithien0909/code-indexer.git",
                "code-indexer", "daemon", "--port", str(self.daemon_port)
            ]
            
            self.daemon_process = subprocess.Popen(
                cmd,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                stdin=subprocess.DEVNULL
            )
            
            # Wait for daemon to start (up to 30 seconds)
            for i in range(60):  # 60 * 0.5 = 30 seconds
                try:
                    req = urllib.request.Request(f"{self.daemon_url}/api/health")
                    with urllib.request.urlopen(req, timeout=1) as response:
                        if response.status == 200:
                            return True
                except:
                    time.sleep(0.5)
                    
            return False
        except Exception as e:
            print(f"Failed to start daemon: {e}", file=sys.stderr)
            return False
    
    def stop_daemon(self):
        """Stop the daemon process."""
        if self.daemon_process:
            self.daemon_process.terminate()
            try:
                self.daemon_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.daemon_process.kill()
                self.daemon_process.wait()
    
    def http_request(self, url: str, data: Optional[bytes] = None) -> Dict[str, Any]:
        """Make HTTP request to daemon."""
        try:
            req = urllib.request.Request(url)
            if data:
                req.add_header('Content-Type', 'application/json')
                req.data = data
            
            with urllib.request.urlopen(req, timeout=30) as response:
                return json.loads(response.read().decode())
        except Exception as e:
            return {"error": str(e)}
    
    def handle_initialize(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle MCP initialize request."""
        self.initialized = True
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
            tools_data = self.http_request(f"{self.daemon_url}/api/tools")
            
            if "error" in tools_data:
                raise Exception(tools_data["error"])
            
            # Convert daemon tools format to MCP format
            tools = tools_data.get("tools", [])
            mcp_tools = []
            
            for tool in tools:
                mcp_tool = {
                    "name": tool.get("name", ""),
                    "description": tool.get("description", ""),
                    "inputSchema": tool.get("input_schema", {
                        "type": "object",
                        "properties": {},
                        "required": []
                    })
                }
                mcp_tools.append(mcp_tool)
            
            return {
                "jsonrpc": "2.0",
                "id": request.get("id"),
                "result": {
                    "tools": mcp_tools
                }
            }
            
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
            
            result_data = self.http_request(
                f"{self.daemon_url}/api/call",
                json.dumps(call_data).encode()
            )
            
            if "error" in result_data:
                raise Exception(result_data["error"])
            
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
    
    def run(self) -> int:
        """Run the stdio bridge."""
        # Setup signal handlers
        def signal_handler(signum, frame):
            self.stop_daemon()
            sys.exit(0)
        
        signal.signal(signal.SIGINT, signal_handler)
        signal.signal(signal.SIGTERM, signal_handler)
        
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
    bridge = MCPStdioBridge()
    sys.exit(bridge.run())
