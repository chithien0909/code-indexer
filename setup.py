#!/usr/bin/env python3
"""Setup script for MCP Code Indexer Python package."""

from setuptools import setup, find_packages
from pathlib import Path

def ensure_binary_executable():
    """Ensure binary is executable during build."""
    binary_path = Path(__file__).parent / "python" / "mcp_code_indexer" / "bin" / "code-indexer"
    if binary_path.exists():
        binary_path.chmod(0o755)
        print(f"✅ Binary is executable: {binary_path}")
    else:
        print(f"⚠️  Binary not found: {binary_path}")

ensure_binary_executable()

setup(
    name="mcp-code-indexer",
    version="1.1.0",
    description="MCP Code Indexer - Index and search source code repositories via Model Context Protocol",
    long_description=open("README.md").read() if Path("README.md").exists() else "",
    long_description_content_type="text/markdown",
    author="MCP Code Indexer Team",
    url="https://github.com/chithien0909/code-indexer",
    packages=find_packages(where="python"),
    package_dir={"": "python"},
    package_data={
        "mcp_code_indexer": [
            "bin/*",
            "bin/code-indexer",
            "config/*",
            "*.yaml",
            "*.yml",
            "*.json"
        ]
    },
    include_package_data=True,
    entry_points={
        "console_scripts": [
            "code-indexer=mcp_code_indexer:main",
        ]
    },
    python_requires=">=3.8",
    install_requires=[],
    zip_safe=False,
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Operating System :: OS Independent",
        "Programming Language :: Go",
        "Programming Language :: Python :: 3",
        "Topic :: Software Development :: Tools",
        "Topic :: Text Processing :: Indexing",
    ],
)
