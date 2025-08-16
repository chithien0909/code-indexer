#!/usr/bin/env python3
"""Build script for MCP Code Indexer Python package."""

import os
import shutil
import subprocess
import sys
import platform
from pathlib import Path

def run_command(cmd, cwd=None):
    """Run a command and return the result."""
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error running command: {' '.join(cmd)}")
        print(f"stdout: {result.stdout}")
        print(f"stderr: {result.stderr}")
        sys.exit(1)
    return result

def build_rust_binary():
    """Build the Rust binary."""
    print("Building Rust binary...")
    
    # Build in release mode
    run_command(["cargo", "build", "--release"])
    
    # Get the binary name based on platform
    binary_name = "code-indexer"
    if platform.system().lower() == "windows":
        binary_name += ".exe"
    
    # Source and destination paths
    source_path = Path("target/release") / binary_name
    dest_dir = Path("python/mcp_code_indexer/bin")
    dest_path = dest_dir / binary_name
    
    # Create destination directory
    dest_dir.mkdir(parents=True, exist_ok=True)
    
    # Copy binary
    if source_path.exists():
        print(f"Copying {source_path} to {dest_path}")
        shutil.copy2(source_path, dest_path)
        
        # Make executable on Unix-like systems
        if platform.system().lower() != "windows":
            os.chmod(dest_path, 0o755)
        
        print(f"Binary copied to {dest_path}")
    else:
        print(f"Error: Binary not found at {source_path}")
        sys.exit(1)

def clean():
    """Clean build artifacts."""
    print("Cleaning build artifacts...")
    
    # Remove Rust build artifacts
    if Path("target").exists():
        shutil.rmtree("target")
    
    # Remove Python build artifacts
    for pattern in ["build", "dist", "*.egg-info"]:
        for path in Path(".").glob(pattern):
            if path.is_dir():
                shutil.rmtree(path)
            else:
                path.unlink()
    
    # Remove copied binary
    bin_dir = Path("python/mcp_code_indexer/bin")
    if bin_dir.exists():
        shutil.rmtree(bin_dir)

def build_python_package():
    """Build the Python package."""
    print("Building Python package...")
    
    # Build the package
    run_command([sys.executable, "-m", "build"])

def main():
    """Main build function."""
    if len(sys.argv) > 1 and sys.argv[1] == "clean":
        clean()
        return
    
    # Ensure we're in the right directory
    if not Path("Cargo.toml").exists():
        print("Error: Must run from project root (where Cargo.toml is located)")
        sys.exit(1)
    
    # Build Rust binary first
    build_rust_binary()
    
    # Build Python package
    build_python_package()
    
    print("Build complete!")

if __name__ == "__main__":
    main()
