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
