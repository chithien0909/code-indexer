package parser

import (
	"testing"
)

func TestGoParser(t *testing.T) {
	parser := NewGoParser()
	
	goCode := `package main

import (
	"fmt"
	"net/http"
)

// main is the entry point
func main() {
	fmt.Println("Hello, World!")
}

// handleRequest handles HTTP requests
func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %s", r.URL.Path)
}

type Server struct {
	port int
}

func (s *Server) Start() error {
	return nil
}
`

	file, err := parser.Parse(goCode, "test.go")
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	// Check basic file info
	if file.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", file.Language)
	}

	if file.Lines == 0 {
		t.Error("Expected lines to be counted")
	}

	// Check functions
	if len(file.Functions) < 2 {
		t.Errorf("Expected at least 2 functions, got %d", len(file.Functions))
	}

	// Check for main function
	foundMain := false
	for _, fn := range file.Functions {
		if fn.Name == "main" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Error("Expected to find main function")
	}

	// Check imports
	if len(file.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(file.Imports))
	}

	// Check comments
	if len(file.Comments) < 2 {
		t.Errorf("Expected at least 2 comments, got %d", len(file.Comments))
	}

	// Check structs (as classes)
	if len(file.Classes) < 1 {
		t.Errorf("Expected at least 1 struct, got %d", len(file.Classes))
	}
}

func TestPythonParser(t *testing.T) {
	parser := NewPythonParser()
	
	pythonCode := `"""
Module for calculations
"""

import math
from typing import List

def calculate_sum(a: int, b: int) -> int:
    """Calculate the sum of two numbers"""
    return a + b

class Calculator:
    """A simple calculator class"""
    
    def __init__(self):
        self.history = []
    
    def add(self, x: float, y: float) -> float:
        """Add two numbers"""
        result = x + y
        self.history.append(f"{x} + {y} = {result}")
        return result

# Global variable
PI = 3.14159
`

	file, err := parser.Parse(pythonCode, "test.py")
	if err != nil {
		t.Fatalf("Failed to parse Python code: %v", err)
	}

	// Check basic file info
	if file.Language != "python" {
		t.Errorf("Expected language 'python', got '%s'", file.Language)
	}

	// Check functions
	if len(file.Functions) < 1 {
		t.Errorf("Expected at least 1 function, got %d", len(file.Functions))
	}

	// Check classes
	if len(file.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(file.Classes))
	}

	// Check for Calculator class
	foundCalculator := false
	for _, cls := range file.Classes {
		if cls.Name == "Calculator" {
			foundCalculator = true
			break
		}
	}
	if !foundCalculator {
		t.Error("Expected to find Calculator class")
	}

	// Check imports
	if len(file.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(file.Imports))
	}

	// Check variables
	if len(file.Variables) < 1 {
		t.Errorf("Expected at least 1 variable, got %d", len(file.Variables))
	}
}

func TestJavaScriptParser(t *testing.T) {
	parser := NewJavaScriptParser()
	
	jsCode := `// Configuration module
const config = {
    port: 3000,
    host: 'localhost'
};

function getConfig() {
    return config;
}

const handleRequest = (req, res) => {
    res.json({ message: 'Hello World' });
};

class Server {
    constructor(port) {
        this.port = port;
    }
    
    start() {
        console.log('Server starting...');
    }
}

module.exports = { config, getConfig, Server };
`

	file, err := parser.Parse(jsCode, "test.js")
	if err != nil {
		t.Fatalf("Failed to parse JavaScript code: %v", err)
	}

	// Check basic file info
	if file.Language != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", file.Language)
	}

	// Check functions
	if len(file.Functions) < 2 {
		t.Errorf("Expected at least 2 functions, got %d", len(file.Functions))
	}

	// Check classes
	if len(file.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(file.Classes))
	}

	// Check variables
	if len(file.Variables) < 2 {
		t.Errorf("Expected at least 2 variables, got %d", len(file.Variables))
	}
}

func TestGenericParser(t *testing.T) {
	parser := NewGenericParser()
	
	textContent := `# This is a shell script
echo "Hello World"

# Another comment
ls -la
`

	file, err := parser.Parse(textContent, "test.sh")
	if err != nil {
		t.Fatalf("Failed to parse generic content: %v", err)
	}

	// Check basic file info
	if file.Language != "generic" {
		t.Errorf("Expected language 'generic', got '%s'", file.Language)
	}

	// Check that lines are counted
	if file.Lines == 0 {
		t.Error("Expected lines to be counted")
	}

	// Check that comments are extracted
	if len(file.Comments) < 2 {
		t.Errorf("Expected at least 2 comments, got %d", len(file.Comments))
	}
}

func TestParserRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test Go parser
	goParser := registry.GetParser("go")
	if goParser.GetLanguage() != "go" {
		t.Errorf("Expected Go parser, got %s", goParser.GetLanguage())
	}

	// Test Python parser
	pythonParser := registry.GetParser("python")
	if pythonParser.GetLanguage() != "python" {
		t.Errorf("Expected Python parser, got %s", pythonParser.GetLanguage())
	}

	// Test fallback to generic parser
	unknownParser := registry.GetParser("unknown")
	if unknownParser.GetLanguage() != "generic" {
		t.Errorf("Expected generic parser for unknown language, got %s", unknownParser.GetLanguage())
	}
}

func TestBaseParserHelpers(t *testing.T) {
	parser := &BaseParser{language: "test"}

	// Test line counting
	content := "line1\nline2\nline3"
	lines := parser.countLines(content)
	if lines != 3 {
		t.Errorf("Expected 3 lines, got %d", lines)
	}

	// Test empty content
	emptyLines := parser.countLines("")
	if emptyLines != 0 {
		t.Errorf("Expected 0 lines for empty content, got %d", emptyLines)
	}

	// Test line number finding
	lineNum := parser.findLineNumber(content, "line2")
	if lineNum != 2 {
		t.Errorf("Expected line 2, got %d", lineNum)
	}

	// Test comment extraction
	codeWithComments := `// This is a line comment
/* This is a 
   block comment */
code here`

	comments := parser.extractComments(codeWithComments, "//", "/*", "*/")
	if len(comments) < 2 {
		t.Errorf("Expected at least 2 comments, got %d", len(comments))
	}
}
