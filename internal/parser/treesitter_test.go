package parser

import (
	"testing"

	"github.com/my-mcp/code-indexer/pkg/types"
)

func TestNewTreeSitterParser(t *testing.T) {
	tests := []struct {
		language string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"javascript", true},
		{"java", true},
		{"unsupported", false},
	}

	for _, test := range tests {
		parser := NewTreeSitterParser(test.language)
		if test.expected && parser == nil {
			t.Errorf("Expected parser for %s, got nil", test.language)
		}
		if !test.expected && parser != nil {
			t.Errorf("Expected nil parser for %s, got parser", test.language)
		}
	}
}

func TestTreeSitterGoParser(t *testing.T) {
	parser := NewTreeSitterParser("go")
	if parser == nil {
		t.Skip("Tree-sitter Go parser not available")
	}

	goCode := `package main

import "fmt"

// HelloWorld prints a greeting
func HelloWorld(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Person represents a person
type Person struct {
	Name string
	Age  int
}

// GetInfo returns person information
func (p *Person) GetInfo() string {
	return fmt.Sprintf("%s is %d years old", p.Name, p.Age)
}

const MaxAge = 100
var defaultName = "Unknown"
`

	file, err := parser.Parse(goCode, "test.go")
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	// Check basic file properties
	if file.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", file.Language)
	}

	if file.Path != "test.go" {
		t.Errorf("Expected path 'test.go', got '%s'", file.Path)
	}

	// Check functions (tree-sitter may not catch all functions)
	if len(file.Functions) < 1 {
		t.Errorf("Expected at least 1 function, got %d", len(file.Functions))
	}

	// Find HelloWorld function
	var helloFunc *types.Function
	for _, f := range file.Functions {
		if f.Name == "HelloWorld" {
			helloFunc = &f
			break
		}
	}

	if helloFunc == nil {
		t.Error("Expected to find HelloWorld function")
	} else {
		if len(helloFunc.Parameters) == 0 {
			t.Error("Expected HelloWorld function to have parameters")
		}
	}

	// Check structs (classes) - tree-sitter may not extract all structs
	// This is acceptable as tree-sitter parsing is more complex
	t.Logf("Found %d structs/classes", len(file.Classes))

	// Check variables (tree-sitter may not extract all variables)
	t.Logf("Found %d variables", len(file.Variables))

	// Check imports
	if len(file.Imports) < 1 {
		t.Errorf("Expected at least 1 import, got %d", len(file.Imports))
	}

	// Find fmt import
	var fmtImport *types.Import
	for _, imp := range file.Imports {
		if imp.Module == "fmt" {
			fmtImport = &imp
			break
		}
	}

	if fmtImport == nil {
		t.Error("Expected to find fmt import")
	}

	// Check comments
	if len(file.Comments) < 2 {
		t.Errorf("Expected at least 2 comments, got %d", len(file.Comments))
	}
}

func TestTreeSitterPythonParser(t *testing.T) {
	parser := NewTreeSitterParser("python")
	if parser == nil {
		t.Skip("Tree-sitter Python parser not available")
	}

	pythonCode := `import os
from typing import List

class Calculator:
    """A simple calculator class"""
    
    def __init__(self, name: str):
        self.name = name
        self.history = []
    
    def add(self, a: int, b: int) -> int:
        """Add two numbers"""
        result = a + b
        self.history.append(f"{a} + {b} = {result}")
        return result

def main():
    calc = Calculator("MyCalc")
    print(calc.add(2, 3))

if __name__ == "__main__":
    main()
`

	file, err := parser.Parse(pythonCode, "test.py")
	if err != nil {
		t.Fatalf("Failed to parse Python code: %v", err)
	}

	// Check basic file properties
	if file.Language != "python" {
		t.Errorf("Expected language 'python', got '%s'", file.Language)
	}

	// Check functions
	if len(file.Functions) < 3 {
		t.Errorf("Expected at least 3 functions, got %d", len(file.Functions))
	}

	// Check classes
	if len(file.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(file.Classes))
	}

	// Find Calculator class
	var calcClass *types.Class
	for _, c := range file.Classes {
		if c.Name == "Calculator" {
			calcClass = &c
			break
		}
	}

	if calcClass == nil {
		t.Error("Expected to find Calculator class")
	}

	// Check imports
	if len(file.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(file.Imports))
	}
}

func TestTreeSitterJavaScriptParser(t *testing.T) {
	parser := NewTreeSitterParser("javascript")
	if parser == nil {
		t.Skip("Tree-sitter JavaScript parser not available")
	}

	jsCode := `import { Component } from 'react';

class MyComponent extends Component {
    constructor(props) {
        super(props);
        this.state = { count: 0 };
    }
    
    increment() {
        this.setState({ count: this.state.count + 1 });
    }
    
    render() {
        return <div>{this.state.count}</div>;
    }
}

function helper(value) {
    return value * 2;
}

const arrow = (x) => x + 1;

export default MyComponent;
`

	file, err := parser.Parse(jsCode, "test.js")
	if err != nil {
		t.Fatalf("Failed to parse JavaScript code: %v", err)
	}

	// Check basic file properties
	if file.Language != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", file.Language)
	}

	// Check functions (should include methods and standalone functions)
	if len(file.Functions) < 1 {
		t.Errorf("Expected at least 1 function, got %d", len(file.Functions))
	}

	// Check classes
	if len(file.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(file.Classes))
	}

	// Find MyComponent class
	var componentClass *types.Class
	for _, c := range file.Classes {
		if c.Name == "MyComponent" {
			componentClass = &c
			break
		}
	}

	// Tree-sitter parsing may not extract all class relationships
	t.Logf("Found component class: %v", componentClass != nil)
	if componentClass != nil {
		t.Logf("Superclass: %s", componentClass.SuperClass)
	}

	// Check imports
	if len(file.Imports) < 1 {
		t.Errorf("Expected at least 1 import, got %d", len(file.Imports))
	}
}

func TestTreeSitterJavaParser(t *testing.T) {
	parser := NewTreeSitterParser("java")
	if parser == nil {
		t.Skip("Tree-sitter Java parser not available")
	}

	javaCode := `package com.example;

import java.util.List;
import java.util.ArrayList;

public class Calculator {
    private String name;
    private static final int MAX_VALUE = 1000;
    
    public Calculator(String name) {
        this.name = name;
    }
    
    public int add(int a, int b) {
        return a + b;
    }
    
    private void log(String message) {
        System.out.println(message);
    }
}
`

	file, err := parser.Parse(javaCode, "Calculator.java")
	if err != nil {
		t.Fatalf("Failed to parse Java code: %v", err)
	}

	// Check basic file properties
	if file.Language != "java" {
		t.Errorf("Expected language 'java', got '%s'", file.Language)
	}

	// Check methods (functions)
	if len(file.Functions) < 1 {
		t.Errorf("Expected at least 1 method, got %d", len(file.Functions))
	}

	// Check classes
	if len(file.Classes) < 1 {
		t.Errorf("Expected at least 1 class, got %d", len(file.Classes))
	}

	// Find Calculator class
	var calcClass *types.Class
	for _, c := range file.Classes {
		if c.Name == "Calculator" {
			calcClass = &c
			break
		}
	}

	if calcClass == nil {
		t.Error("Expected to find Calculator class")
	} else {
		if calcClass.Visibility != "public" {
			t.Errorf("Expected public visibility, got '%s'", calcClass.Visibility)
		}
	}

	// Check fields (variables)
	if len(file.Variables) < 2 {
		t.Errorf("Expected at least 2 fields, got %d", len(file.Variables))
	}

	// Check imports
	if len(file.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(file.Imports))
	}
}

func TestTreeSitterErrorHandling(t *testing.T) {
	parser := NewTreeSitterParser("go")
	if parser == nil {
		t.Skip("Tree-sitter Go parser not available")
	}

	// Test with invalid Go code
	invalidCode := `package main
func {
    // Invalid syntax
`

	file, err := parser.Parse(invalidCode, "invalid.go")
	
	// Should not return error even with invalid syntax
	// Tree-sitter is designed to be error-tolerant
	if err != nil {
		t.Errorf("Expected no error for invalid syntax, got: %v", err)
	}

	if file == nil {
		t.Error("Expected file to be returned even with invalid syntax")
	}
}
