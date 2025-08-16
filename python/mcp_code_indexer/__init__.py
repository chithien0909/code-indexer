"""MCP Code Indexer - Python Package Wrapper"""

import os
import sys
import subprocess
import platform
from pathlib import Path

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



def main() -> int:
    """Main entry point for the uvx-installed package."""
    try:
        args = sys.argv[1:]
        if not args:
            args = ["serve"]

        # Get the Go binary path
        binary_path = get_binary_path()
        if not binary_path:
            print("Error: code-indexer binary not found", file=sys.stderr)
            return 1

        # Execute the Go binary directly with the provided arguments
        # The Go binary handles MCP protocol natively
        result = subprocess.run([binary_path] + args)
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
