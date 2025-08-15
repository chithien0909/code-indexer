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

        # For serve command, use daemon mode which we know works
        if args[0] == "serve":
            # Use daemon mode for better compatibility
            import socket
            import time
            import requests

            # Find available port
            port = 9991
            for p in range(9991, 10000):
                try:
                    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                    sock.bind(('localhost', p))
                    sock.close()
                    port = p
                    break
                except OSError:
                    continue

            # Start daemon
            daemon_args = ["daemon", "--port", str(port)]
            daemon_process = subprocess.Popen(
                [get_binary_path()] + daemon_args,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL
            )

            # Wait for daemon to start
            daemon_url = f"http://localhost:{port}"
            for _ in range(30):
                try:
                    response = requests.get(f"{daemon_url}/api/health", timeout=1)
                    if response.status_code == 200:
                        break
                except:
                    time.sleep(1)
            else:
                daemon_process.terminate()
                print("Failed to start daemon", file=sys.stderr)
                return 1

            # Handle MCP stdio protocol
            try:
                import json
                for line in sys.stdin:
                    line = line.strip()
                    if not line:
                        continue

                    try:
                        request = json.loads(line)
                        method = request.get("method")

                        if method == "initialize":
                            response = {
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
                                tools_resp = requests.get(f"{daemon_url}/api/tools", timeout=5)
                                tools_data = tools_resp.json() if tools_resp.status_code == 200 else {"tools": []}
                                response = {
                                    "jsonrpc": "2.0",
                                    "id": request.get("id"),
                                    "result": {"tools": tools_data.get("tools", [])}
                                }
                            except:
                                response = {
                                    "jsonrpc": "2.0",
                                    "id": request.get("id"),
                                    "result": {"tools": []}
                                }
                        else:
                            response = {
                                "jsonrpc": "2.0",
                                "id": request.get("id"),
                                "error": {"code": -32601, "message": f"Method not found: {method}"}
                            }

                        print(json.dumps(response))
                        sys.stdout.flush()

                    except json.JSONDecodeError:
                        continue

            finally:
                daemon_process.terminate()
                daemon_process.wait()

            return 0
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
