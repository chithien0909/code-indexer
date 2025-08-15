"""MCP Code Indexer - Python Package Wrapper"""

import os
import sys
import subprocess
import platform
from pathlib import Path
from typing import List

__version__ = "1.1.0"
__author__ = "MCP Code Indexer Team"

def get_binary_path() -> Path:
    """Get the path to the code-indexer binary."""
    package_dir = Path(__file__).parent
    system = platform.system().lower()
    binary_name = "code-indexer"
    if system == "windows":
        binary_name += ".exe"

    # Look for the binary in the package
    possible_paths = [
        package_dir / "bin" / binary_name,
        package_dir / binary_name,
        package_dir.parent.parent / "bin" / binary_name,
        Path.cwd() / "bin" / binary_name,
    ]

    for path in possible_paths:
        if path.exists() and path.is_file():
            # Ensure it's executable
            if not os.access(path, os.X_OK):
                try:
                    path.chmod(0o755)
                except (OSError, PermissionError):
                    continue
            return path

    # Try system PATH
    import shutil
    binary_in_path = shutil.which(binary_name)
    if binary_in_path:
        return Path(binary_in_path)

    raise FileNotFoundError(
        f"Could not find {binary_name} binary. "
        f"Searched in: {[str(p) for p in possible_paths]}"
    )

def run_code_indexer(args: List[str], **kwargs) -> subprocess.CompletedProcess:
    """Run the code-indexer binary with the given arguments."""
    binary_path = get_binary_path()
    cmd = [str(binary_path)] + args
    defaults = {"check": False, "capture_output": False, "text": True}
    defaults.update(kwargs)
    return subprocess.run(cmd, **defaults)

def main() -> int:
    """Main entry point for the uvx-installed package."""
    try:
        args = sys.argv[1:]
        if not args:
            args = ["serve"]

        # For serve command, use stdio bridge for better IDE compatibility
        if args[0] == "serve":
            # Use stdio bridge that converts MCP stdio to daemon HTTP calls
            import json
            import urllib.request
            import urllib.parse
            import urllib.error
            import signal
            import time

            class MCPStdioBridge:
                def __init__(self):
                    self.daemon_process = None
                    self.daemon_port = 9991
                    self.daemon_url = f"http://localhost:{self.daemon_port}"

                def start_daemon(self):
                    try:
                        binary_path = str(get_binary_path())  # Convert PosixPath to string
                        cmd = [binary_path, "daemon", "--port", str(self.daemon_port)]
                        self.daemon_process = subprocess.Popen(
                            cmd,
                            stdout=subprocess.DEVNULL,
                            stderr=subprocess.DEVNULL
                        )

                        # Wait for daemon to start
                        for i in range(60):
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
                    if self.daemon_process:
                        self.daemon_process.terminate()
                        try:
                            self.daemon_process.wait(timeout=5)
                        except subprocess.TimeoutExpired:
                            self.daemon_process.kill()

                def http_request(self, url, data=None):
                    try:
                        req = urllib.request.Request(url)
                        if data:
                            req.add_header("Content-Type", "application/json")
                            req.data = data
                        with urllib.request.urlopen(req, timeout=30) as response:
                            return json.loads(response.read().decode())
                    except Exception as e:
                        return {"error": str(e)}

                def handle_request(self, request):
                    method = request.get("method")
                    if method == "initialize":
                        return {
                            "jsonrpc": "2.0",
                            "id": request.get("id"),
                            "result": {
                                "protocolVersion": "2024-11-05",
                                "capabilities": {"tools": {}},
                                "serverInfo": {"name": "Code Indexer", "version": "1.1.0"}
                            }
                        }
                    elif method == "tools/list":
                        try:
                            tools_data = self.http_request(f"{self.daemon_url}/api/tools")
                            tools = tools_data.get("tools", [])
                            return {
                                "jsonrpc": "2.0",
                                "id": request.get("id"),
                                "result": {"tools": tools}
                            }
                        except Exception as e:
                            return {
                                "jsonrpc": "2.0",
                                "id": request.get("id"),
                                "result": {"tools": []}
                            }
                    elif method == "tools/call":
                        try:
                            params = request.get("params", {})
                            call_data = {
                                "tool": params.get("name"),
                                "arguments": params.get("arguments", {})
                            }
                            result_data = self.http_request(
                                f"{self.daemon_url}/api/call",
                                json.dumps(call_data).encode()
                            )
                            return {
                                "jsonrpc": "2.0",
                                "id": request.get("id"),
                                "result": {
                                    "content": [{
                                        "type": "text",
                                        "text": json.dumps(result_data.get("result", {}), indent=2)
                                    }]
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
                    # Setup signal handlers
                    signal.signal(signal.SIGINT, lambda s, f: self.stop_daemon())
                    signal.signal(signal.SIGTERM, lambda s, f: self.stop_daemon())

                    if not self.start_daemon():
                        return 1

                    try:
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
                                continue
                            except Exception as e:
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

            # Run the bridge
            bridge = MCPStdioBridge()
            return bridge.run()
        else:
            # For other commands, run directly
            result = run_code_indexer(args)
            return result.returncode

    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        print("Make sure the code-indexer binary is available.", file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        print("\nInterrupted by user", file=sys.stderr)
        return 130
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        return 1

if __name__ == "__main__":
    sys.exit(main())
