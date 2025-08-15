"""
MCP Code Indexer - Python Package Wrapper

This package provides a Python wrapper around the Go-based MCP Code Indexer,
allowing it to be installed and executed via uvx.
"""

import os
import sys
import subprocess
import platform
from pathlib import Path
from typing import List, Optional

__version__ = "1.1.0"
__author__ = "MCP Code Indexer Team"
__email__ = "team@mcp-code-indexer.dev"

# Package metadata
__all__ = [
    "__version__",
    "__author__", 
    "__email__",
    "main",
    "get_binary_path",
    "run_code_indexer"
]


def get_binary_path() -> Path:
    """
    Get the path to the code-indexer binary.
    
    Returns:
        Path to the code-indexer binary
        
    Raises:
        FileNotFoundError: If the binary is not found
    """
    # Get the package directory
    package_dir = Path(__file__).parent
    
    # Determine the binary name based on the platform
    system = platform.system().lower()
    arch = platform.machine().lower()
    
    # Map platform names to binary names
    binary_name = "code-indexer"
    if system == "windows":
        binary_name += ".exe"
    
    # Look for the binary in several locations
    possible_paths = [
        # In the package directory
        package_dir / "bin" / binary_name,
        # In the package directory (flat structure)
        package_dir / binary_name,
        # In the parent directory (development setup)
        package_dir.parent.parent / "bin" / binary_name,
        # In the current working directory
        Path.cwd() / "bin" / binary_name,
        Path.cwd() / binary_name,
    ]
    
    for path in possible_paths:
        if path.exists() and path.is_file():
            # Make sure it's executable
            if not os.access(path, os.X_OK):
                try:
                    path.chmod(0o755)
                except (OSError, PermissionError):
                    continue
            return path
    
    # If not found, try to find it in PATH
    import shutil
    binary_in_path = shutil.which(binary_name)
    if binary_in_path:
        return Path(binary_in_path)
    
    raise FileNotFoundError(
        f"Could not find {binary_name} binary. "
        f"Searched in: {[str(p) for p in possible_paths]}"
    )


def run_code_indexer(args: List[str], **kwargs) -> subprocess.CompletedProcess:
    """
    Run the code-indexer binary with the given arguments.
    
    Args:
        args: Command line arguments to pass to code-indexer
        **kwargs: Additional keyword arguments to pass to subprocess.run
        
    Returns:
        CompletedProcess instance
        
    Raises:
        FileNotFoundError: If the binary is not found
        subprocess.CalledProcessError: If the command fails
    """
    binary_path = get_binary_path()
    
    # Prepare the command
    cmd = [str(binary_path)] + args
    
    # Set default values for subprocess.run
    defaults = {
        "check": False,
        "capture_output": False,
        "text": True,
    }
    defaults.update(kwargs)
    
    # Run the command
    return subprocess.run(cmd, **defaults)


def main() -> int:
    """
    Main entry point for the uvx-installed package.
    
    This function is called when the package is executed via uvx.
    It forwards all command line arguments to the code-indexer binary.
    
    Returns:
        Exit code from the code-indexer binary
    """
    try:
        # Get command line arguments (excluding the script name)
        args = sys.argv[1:]
        
        # If no arguments provided, default to mcp-server command
        if not args:
            args = ["mcp-server"]
        
        # Run the code-indexer binary
        result = run_code_indexer(args)
        
        return result.returncode
        
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        print(
            "Make sure the code-indexer binary is available in your PATH "
            "or in the package directory.",
            file=sys.stderr
        )
        return 1
    except KeyboardInterrupt:
        print("\nInterrupted by user", file=sys.stderr)
        return 130
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        return 1


# For backwards compatibility
if __name__ == "__main__":
    sys.exit(main())
