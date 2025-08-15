#!/usr/bin/env python3
"""
Setup script for MCP Code Indexer Python package.

This script builds the Go binary and packages it with the Python wrapper
for distribution via PyPI and uvx.
"""

import os
import sys
import shutil
import subprocess
import platform
from pathlib import Path
from setuptools import setup, find_packages
from setuptools.command.build_py import build_py
from setuptools.command.develop import develop
from setuptools.command.install import install


class BuildGoCommand:
    """Mixin class for building the Go binary."""
    
    def build_go_binary(self):
        """Build the Go binary for the current platform."""
        print("Building Go binary...")
        
        # Get the project root directory
        project_root = Path(__file__).parent
        
        # Determine the binary name based on the platform
        system = platform.system().lower()
        binary_name = "code-indexer"
        if system == "windows":
            binary_name += ".exe"
        
        # Create the bin directory in the Python package
        bin_dir = project_root / "python" / "mcp_code_indexer" / "bin"
        bin_dir.mkdir(parents=True, exist_ok=True)
        
        # Build the Go binary
        env = os.environ.copy()
        env["CGO_ENABLED"] = "0"  # Static binary
        
        # Set GOOS and GOARCH for cross-compilation if needed
        if system == "windows":
            env["GOOS"] = "windows"
        elif system == "darwin":
            env["GOOS"] = "darwin"
        else:
            env["GOOS"] = "linux"
        
        # Determine architecture
        arch = platform.machine().lower()
        if arch in ("x86_64", "amd64"):
            env["GOARCH"] = "amd64"
        elif arch in ("aarch64", "arm64"):
            env["GOARCH"] = "arm64"
        else:
            env["GOARCH"] = "amd64"  # Default fallback
        
        # Build command
        output_path = bin_dir / binary_name
        cmd = [
            "go", "build",
            "-o", str(output_path),
            "-ldflags", "-s -w",  # Strip debug info for smaller binary
            "./cmd/server"
        ]
        
        try:
            result = subprocess.run(
                cmd,
                cwd=project_root,
                env=env,
                check=True,
                capture_output=True,
                text=True
            )
            print(f"Successfully built binary: {output_path}")
            
            # Make the binary executable
            if output_path.exists():
                output_path.chmod(0o755)
                
        except subprocess.CalledProcessError as e:
            print(f"Failed to build Go binary: {e}")
            print(f"stdout: {e.stdout}")
            print(f"stderr: {e.stderr}")
            sys.exit(1)
        except FileNotFoundError:
            print("Error: Go compiler not found. Please install Go.")
            print("Visit https://golang.org/doc/install for installation instructions.")
            sys.exit(1)


class CustomBuildPy(build_py, BuildGoCommand):
    """Custom build command that builds the Go binary first."""
    
    def run(self):
        self.build_go_binary()
        super().run()


class CustomDevelop(develop, BuildGoCommand):
    """Custom develop command that builds the Go binary first."""
    
    def run(self):
        self.build_go_binary()
        super().run()


class CustomInstall(install, BuildGoCommand):
    """Custom install command that builds the Go binary first."""
    
    def run(self):
        self.build_go_binary()
        super().run()


def read_file(filename):
    """Read a file and return its contents."""
    with open(filename, 'r', encoding='utf-8') as f:
        return f.read()


def get_version():
    """Get version from pyproject.toml or fallback to default."""
    try:
        import tomllib
    except ImportError:
        try:
            import tomli as tomllib
        except ImportError:
            return "1.1.0"  # Fallback version
    
    try:
        with open("pyproject.toml", "rb") as f:
            data = tomllib.load(f)
            return data["project"]["version"]
    except (FileNotFoundError, KeyError):
        return "1.1.0"  # Fallback version


# Read long description from README
try:
    long_description = read_file("README.md")
    long_description_content_type = "text/markdown"
except FileNotFoundError:
    long_description = "MCP Code Indexer - Index and search source code repositories"
    long_description_content_type = "text/plain"

# Setup configuration
setup(
    name="mcp-code-indexer",
    version=get_version(),
    description="MCP Code Indexer - Index and search source code repositories via Model Context Protocol",
    long_description=long_description,
    long_description_content_type=long_description_content_type,
    author="MCP Code Indexer Team",
    author_email="team@mcp-code-indexer.dev",
    url="https://github.com/my-mcp/code-indexer",
    project_urls={
        "Documentation": "https://github.com/my-mcp/code-indexer/blob/main/README.md",
        "Source": "https://github.com/my-mcp/code-indexer",
        "Tracker": "https://github.com/my-mcp/code-indexer/issues",
    },
    packages=find_packages(where="python"),
    package_dir={"": "python"},
    package_data={
        "mcp_code_indexer": [
            "bin/*",
            "config/*",
            "*.yaml",
            "*.yml", 
            "*.json"
        ],
    },
    include_package_data=True,
    entry_points={
        "console_scripts": [
            "code-indexer=mcp_code_indexer:main",
        ],
    },
    python_requires=">=3.8",
    install_requires=[
        # No Python dependencies - this is just a wrapper around the Go binary
    ],
    extras_require={
        "dev": [
            "pytest>=7.0",
            "pytest-cov>=4.0",
            "black>=23.0",
            "isort>=5.0",
            "flake8>=6.0",
            "mypy>=1.0",
        ],
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Go",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Tools",
        "Topic :: Text Processing :: Indexing",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
    ],
    keywords=[
        "mcp",
        "model-context-protocol",
        "code-indexer", 
        "search",
        "llm",
        "ai",
        "development-tools"
    ],
    cmdclass={
        "build_py": CustomBuildPy,
        "develop": CustomDevelop,
        "install": CustomInstall,
    },
    zip_safe=False,  # Binary files need to be extracted
)
