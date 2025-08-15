package parser

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// TreeSitterParser provides enhanced parsing using tree-sitter
type TreeSitterParser struct {
	BaseParser
	tsLanguage *sitter.Language
}

// NewTreeSitterParser creates a new tree-sitter parser for the given language
func NewTreeSitterParser(lang string) *TreeSitterParser {
	var language *sitter.Language

	switch lang {
	case "go":
		language = golang.GetLanguage()
	case "python":
		language = python.GetLanguage()
	case "javascript", "typescript":
		language = javascript.GetLanguage()
	case "java":
		language = java.GetLanguage()
	default:
		return nil // Unsupported language
	}

	return &TreeSitterParser{
		BaseParser: BaseParser{language: lang},
		tsLanguage: language,
	}
}

// Parse parses source code using tree-sitter for enhanced accuracy
func (p *TreeSitterParser) Parse(content string, filePath string) (*types.CodeFile, error) {
	file := &types.CodeFile{
		Path:     filePath,
		Language: p.language,
		Lines:    p.countLines(content),
		Content:  content,
	}

	// Create parser
	parser := sitter.NewParser()
	parser.SetLanguage(p.tsLanguage)

	// Parse the source code
	sourceCode := []byte(content)
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse with tree-sitter: %w", err)
	}
	defer tree.Close()

	// Store the AST for potential future use
	file.TreeSitterAST = tree.RootNode()

	// Extract metadata based on language
	switch p.BaseParser.language {
	case "go":
		p.parseGoCode(tree.RootNode(), sourceCode, file)
	case "python":
		p.parsePythonCode(tree.RootNode(), sourceCode, file)
	case "javascript", "typescript":
		p.parseJavaScriptCode(tree.RootNode(), sourceCode, file)
	case "java":
		p.parseJavaCode(tree.RootNode(), sourceCode, file)
	}

	return file, nil
}

// parseGoCode extracts Go-specific metadata using tree-sitter
func (p *TreeSitterParser) parseGoCode(node *sitter.Node, source []byte, file *types.CodeFile) {
	p.walkNode(node, source, func(n *sitter.Node) {
		switch n.Type() {
		case "function_declaration", "method_declaration":
			function := p.extractGoFunction(n, source)
			file.Functions = append(file.Functions, function)

		case "type_declaration":
			// Check if it's a struct
			if p.hasChildOfType(n, "struct_type") {
				class := p.extractGoStruct(n, source)
				file.Classes = append(file.Classes, class)
			}

		case "var_declaration", "const_declaration":
			variables := p.extractGoVariables(n, source)
			file.Variables = append(file.Variables, variables...)

		case "import_declaration":
			imports := p.extractGoImports(n, source)
			file.Imports = append(file.Imports, imports...)

		case "comment":
			comment := p.extractComment(n, source)
			file.Comments = append(file.Comments, comment)
		}
	})
}

// parsePythonCode extracts Python-specific metadata using tree-sitter
func (p *TreeSitterParser) parsePythonCode(node *sitter.Node, source []byte, file *types.CodeFile) {
	p.walkNode(node, source, func(n *sitter.Node) {
		switch n.Type() {
		case "function_definition":
			function := p.extractPythonFunction(n, source)
			file.Functions = append(file.Functions, function)

		case "class_definition":
			class := p.extractPythonClass(n, source)
			file.Classes = append(file.Classes, class)

		case "assignment":
			variable := p.extractPythonVariable(n, source)
			if variable.Name != "" {
				file.Variables = append(file.Variables, variable)
			}

		case "import_statement", "import_from_statement":
			imports := p.extractPythonImports(n, source)
			file.Imports = append(file.Imports, imports...)

		case "comment":
			comment := p.extractComment(n, source)
			file.Comments = append(file.Comments, comment)
		}
	})
}

// parseJavaScriptCode extracts JavaScript-specific metadata using tree-sitter
func (p *TreeSitterParser) parseJavaScriptCode(node *sitter.Node, source []byte, file *types.CodeFile) {
	p.walkNode(node, source, func(n *sitter.Node) {
		switch n.Type() {
		case "function_declaration", "function_expression", "arrow_function":
			function := p.extractJavaScriptFunction(n, source)
			file.Functions = append(file.Functions, function)

		case "class_declaration":
			class := p.extractJavaScriptClass(n, source)
			file.Classes = append(file.Classes, class)

		case "variable_declaration":
			variables := p.extractJavaScriptVariables(n, source)
			file.Variables = append(file.Variables, variables...)

		case "import_statement":
			imports := p.extractJavaScriptImports(n, source)
			file.Imports = append(file.Imports, imports...)

		case "comment":
			comment := p.extractComment(n, source)
			file.Comments = append(file.Comments, comment)
		}
	})
}

// parseJavaCode extracts Java-specific metadata using tree-sitter
func (p *TreeSitterParser) parseJavaCode(node *sitter.Node, source []byte, file *types.CodeFile) {
	p.walkNode(node, source, func(n *sitter.Node) {
		switch n.Type() {
		case "method_declaration":
			function := p.extractJavaMethod(n, source)
			file.Functions = append(file.Functions, function)

		case "class_declaration":
			class := p.extractJavaClass(n, source)
			file.Classes = append(file.Classes, class)

		case "field_declaration":
			variables := p.extractJavaFields(n, source)
			file.Variables = append(file.Variables, variables...)

		case "import_declaration":
			imports := p.extractJavaImports(n, source)
			file.Imports = append(file.Imports, imports...)

		case "line_comment", "block_comment":
			comment := p.extractComment(n, source)
			file.Comments = append(file.Comments, comment)
		}
	})
}

// walkNode recursively walks through all nodes in the AST
func (p *TreeSitterParser) walkNode(node *sitter.Node, source []byte, callback func(*sitter.Node)) {
	callback(node)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.walkNode(child, source, callback)
	}
}

// hasChildOfType checks if a node has a child of the specified type
func (p *TreeSitterParser) hasChildOfType(node *sitter.Node, nodeType string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			return true
		}
	}
	return false
}

// getNodeText extracts text content from a node
func (p *TreeSitterParser) getNodeText(node *sitter.Node, source []byte) string {
	return string(source[node.StartByte():node.EndByte()])
}

// getLineNumber converts byte position to line number
func (p *TreeSitterParser) getLineNumber(node *sitter.Node) int {
	return int(node.StartPoint().Row) + 1
}

// getEndLineNumber converts byte position to end line number
func (p *TreeSitterParser) getEndLineNumber(node *sitter.Node) int {
	return int(node.EndPoint().Row) + 1
}

// extractComment extracts comment information from a node
func (p *TreeSitterParser) extractComment(node *sitter.Node, source []byte) types.Comment {
	text := p.getNodeText(node, source)
	
	// Clean up comment markers
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "//") {
		text = strings.TrimSpace(strings.TrimPrefix(text, "//"))
	} else if strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/") {
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "/*"), "*/"))
	} else if strings.HasPrefix(text, "#") {
		text = strings.TrimSpace(strings.TrimPrefix(text, "#"))
	}

	commentType := "line"
	if strings.Contains(p.getNodeText(node, source), "/*") {
		commentType = "block"
	}

	return types.Comment{
		Text:      text,
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
		Type:      commentType,
	}
}

// extractGoFunction extracts Go function information
func (p *TreeSitterParser) extractGoFunction(node *sitter.Node, source []byte) types.Function {
	function := types.Function{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
		Signature: p.getNodeText(node, source),
	}

	// Extract function name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			function.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract parameters and return type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter_list" {
			function.Parameters = p.extractGoParameters(child, source)
		} else if child.Type() == "type_identifier" || child.Type() == "pointer_type" {
			function.ReturnType = p.getNodeText(child, source)
		}
	}

	return function
}

// extractGoStruct extracts Go struct information
func (p *TreeSitterParser) extractGoStruct(node *sitter.Node, source []byte) types.Class {
	class := types.Class{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
	}

	// Extract struct name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_spec" {
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "type_identifier" {
					class.Name = p.getNodeText(grandchild, source)
					break
				}
			}
			break
		}
	}

	return class
}

// extractGoVariables extracts Go variable declarations
func (p *TreeSitterParser) extractGoVariables(node *sitter.Node, source []byte) []types.Variable {
	var variables []types.Variable
	isConstant := strings.Contains(p.getNodeText(node, source), "const")

	// Walk through variable specifications
	p.walkNode(node, source, func(n *sitter.Node) {
		if n.Type() == "var_spec" || n.Type() == "const_spec" {
			variable := types.Variable{
				StartLine:  p.getLineNumber(n),
				EndLine:    p.getEndLineNumber(n),
				IsConstant: isConstant,
			}

			// Extract variable name and type
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "identifier" && variable.Name == "" {
					variable.Name = p.getNodeText(child, source)
				} else if child.Type() == "type_identifier" || child.Type() == "pointer_type" {
					variable.Type = p.getNodeText(child, source)
				}
			}

			if variable.Name != "" {
				variables = append(variables, variable)
			}
		}
	})

	return variables
}

// extractGoImports extracts Go import statements
func (p *TreeSitterParser) extractGoImports(node *sitter.Node, source []byte) []types.Import {
	var imports []types.Import

	p.walkNode(node, source, func(n *sitter.Node) {
		if n.Type() == "import_spec" {
			importStmt := types.Import{
				StartLine: p.getLineNumber(n),
			}

			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "interpreted_string_literal" {
					module := p.getNodeText(child, source)
					// Remove quotes
					module = strings.Trim(module, `"`)
					importStmt.Module = module
				} else if child.Type() == "package_identifier" {
					importStmt.Alias = p.getNodeText(child, source)
				}
			}

			if importStmt.Module != "" {
				imports = append(imports, importStmt)
			}
		}
	})

	return imports
}

// extractGoParameters extracts function parameters
func (p *TreeSitterParser) extractGoParameters(node *sitter.Node, source []byte) []string {
	var parameters []string

	p.walkNode(node, source, func(n *sitter.Node) {
		if n.Type() == "parameter_declaration" {
			param := p.getNodeText(n, source)
			parameters = append(parameters, param)
		}
	})

	return parameters
}

// extractPythonFunction extracts Python function information
func (p *TreeSitterParser) extractPythonFunction(node *sitter.Node, source []byte) types.Function {
	function := types.Function{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
		Signature: p.getNodeText(node, source),
	}

	// Extract function name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			function.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract parameters
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameters" {
			function.Parameters = p.extractPythonParameters(child, source)
			break
		}
	}

	return function
}

// extractPythonClass extracts Python class information
func (p *TreeSitterParser) extractPythonClass(node *sitter.Node, source []byte) types.Class {
	class := types.Class{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
	}

	// Extract class name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			class.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract superclass
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "argument_list" {
			if child.ChildCount() > 0 {
				superclass := child.Child(0)
				if superclass.Type() == "identifier" {
					class.SuperClass = p.getNodeText(superclass, source)
				}
			}
			break
		}
	}

	return class
}

// extractPythonVariable extracts Python variable assignment
func (p *TreeSitterParser) extractPythonVariable(node *sitter.Node, source []byte) types.Variable {
	variable := types.Variable{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
	}

	// Extract variable name from left side of assignment
	if node.ChildCount() > 0 {
		left := node.Child(0)
		if left.Type() == "identifier" {
			variable.Name = p.getNodeText(left, source)
		}
	}

	return variable
}

// extractPythonImports extracts Python import statements
func (p *TreeSitterParser) extractPythonImports(node *sitter.Node, source []byte) []types.Import {
	var imports []types.Import

	if node.Type() == "import_statement" {
		// Handle "import module" statements
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dotted_name" || child.Type() == "identifier" {
				importStmt := types.Import{
					Module:    p.getNodeText(child, source),
					StartLine: p.getLineNumber(node),
				}
				imports = append(imports, importStmt)
			}
		}
	} else if node.Type() == "import_from_statement" {
		// Handle "from module import ..." statements
		importStmt := types.Import{
			StartLine: p.getLineNumber(node),
		}

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dotted_name" || child.Type() == "identifier" {
				if importStmt.Module == "" {
					importStmt.Module = p.getNodeText(child, source)
				} else {
					importStmt.Alias = p.getNodeText(child, source)
				}
			}
		}

		if importStmt.Module != "" {
			imports = append(imports, importStmt)
		}
	}

	return imports
}

// extractPythonParameters extracts function parameters
func (p *TreeSitterParser) extractPythonParameters(node *sitter.Node, source []byte) []string {
	var parameters []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			param := p.getNodeText(child, source)
			if param != "self" { // Skip 'self' parameter
				parameters = append(parameters, param)
			}
		}
	}

	return parameters
}

// extractJavaScriptFunction extracts JavaScript function information
func (p *TreeSitterParser) extractJavaScriptFunction(node *sitter.Node, source []byte) types.Function {
	function := types.Function{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
		Signature: p.getNodeText(node, source),
	}

	// Extract function name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			function.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract parameters
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "formal_parameters" {
			function.Parameters = p.extractJavaScriptParameters(child, source)
			break
		}
	}

	return function
}

// extractJavaScriptClass extracts JavaScript class information
func (p *TreeSitterParser) extractJavaScriptClass(node *sitter.Node, source []byte) types.Class {
	class := types.Class{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
	}

	// Extract class name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			class.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract superclass
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "class_heritage" {
			if child.ChildCount() > 0 {
				superclass := child.Child(0)
				if superclass.Type() == "identifier" {
					class.SuperClass = p.getNodeText(superclass, source)
				}
			}
			break
		}
	}

	return class
}

// extractJavaScriptVariables extracts JavaScript variable declarations
func (p *TreeSitterParser) extractJavaScriptVariables(node *sitter.Node, source []byte) []types.Variable {
	var variables []types.Variable

	p.walkNode(node, source, func(n *sitter.Node) {
		if n.Type() == "variable_declarator" {
			variable := types.Variable{
				StartLine: p.getLineNumber(n),
				EndLine:   p.getEndLineNumber(n),
			}

			// Extract variable name
			if n.ChildCount() > 0 {
				nameNode := n.Child(0)
				if nameNode.Type() == "identifier" {
					variable.Name = p.getNodeText(nameNode, source)
				}
			}

			// Check if it's const
			parentText := p.getNodeText(node, source)
			variable.IsConstant = strings.HasPrefix(strings.TrimSpace(parentText), "const")

			if variable.Name != "" {
				variables = append(variables, variable)
			}
		}
	})

	return variables
}

// extractJavaScriptImports extracts JavaScript import statements
func (p *TreeSitterParser) extractJavaScriptImports(node *sitter.Node, source []byte) []types.Import {
	var imports []types.Import

	importStmt := types.Import{
		StartLine: p.getLineNumber(node),
	}

	// Extract module path
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "string" {
			module := p.getNodeText(child, source)
			// Remove quotes
			module = strings.Trim(module, `"'`)
			importStmt.Module = module
			break
		}
	}

	if importStmt.Module != "" {
		imports = append(imports, importStmt)
	}

	return imports
}

// extractJavaScriptParameters extracts function parameters
func (p *TreeSitterParser) extractJavaScriptParameters(node *sitter.Node, source []byte) []string {
	var parameters []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			param := p.getNodeText(child, source)
			parameters = append(parameters, param)
		}
	}

	return parameters
}

// extractJavaMethod extracts Java method information
func (p *TreeSitterParser) extractJavaMethod(node *sitter.Node, source []byte) types.Function {
	function := types.Function{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
		Signature: p.getNodeText(node, source),
		IsMethod:  true,
	}

	// Extract method name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			function.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract parameters and return type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "formal_parameters" {
			function.Parameters = p.extractJavaParameters(child, source)
		} else if child.Type() == "type_identifier" || child.Type() == "generic_type" {
			function.ReturnType = p.getNodeText(child, source)
		}
	}

	// Extract visibility
	methodText := p.getNodeText(node, source)
	if strings.Contains(methodText, "public") {
		function.Visibility = "public"
	} else if strings.Contains(methodText, "private") {
		function.Visibility = "private"
	} else if strings.Contains(methodText, "protected") {
		function.Visibility = "protected"
	}

	return function
}

// extractJavaClass extracts Java class information
func (p *TreeSitterParser) extractJavaClass(node *sitter.Node, source []byte) types.Class {
	class := types.Class{
		StartLine: p.getLineNumber(node),
		EndLine:   p.getEndLineNumber(node),
	}

	// Extract class name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			class.Name = p.getNodeText(child, source)
			break
		}
	}

	// Extract superclass and interfaces
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "superclass" {
			if child.ChildCount() > 0 {
				superclass := child.Child(0)
				if superclass.Type() == "type_identifier" {
					class.SuperClass = p.getNodeText(superclass, source)
				}
			}
		} else if child.Type() == "super_interfaces" {
			// Extract interfaces
			p.walkNode(child, source, func(n *sitter.Node) {
				if n.Type() == "type_identifier" {
					class.Interfaces = append(class.Interfaces, p.getNodeText(n, source))
				}
			})
		}
	}

	// Extract visibility
	classText := p.getNodeText(node, source)
	if strings.Contains(classText, "public") {
		class.Visibility = "public"
	} else if strings.Contains(classText, "private") {
		class.Visibility = "private"
	} else if strings.Contains(classText, "protected") {
		class.Visibility = "protected"
	}

	return class
}

// extractJavaFields extracts Java field declarations
func (p *TreeSitterParser) extractJavaFields(node *sitter.Node, source []byte) []types.Variable {
	var variables []types.Variable

	p.walkNode(node, source, func(n *sitter.Node) {
		if n.Type() == "variable_declarator" {
			variable := types.Variable{
				StartLine: p.getLineNumber(n),
				EndLine:   p.getEndLineNumber(n),
			}

			// Extract field name
			if n.ChildCount() > 0 {
				nameNode := n.Child(0)
				if nameNode.Type() == "identifier" {
					variable.Name = p.getNodeText(nameNode, source)
				}
			}

			// Extract type from parent field declaration
			parent := node
			for i := 0; i < int(parent.ChildCount()); i++ {
				child := parent.Child(i)
				if child.Type() == "type_identifier" || child.Type() == "generic_type" {
					variable.Type = p.getNodeText(child, source)
					break
				}
			}

			// Check if it's final (constant)
			fieldText := p.getNodeText(node, source)
			variable.IsConstant = strings.Contains(fieldText, "final")

			// Extract visibility
			if strings.Contains(fieldText, "public") {
				variable.Visibility = "public"
			} else if strings.Contains(fieldText, "private") {
				variable.Visibility = "private"
			} else if strings.Contains(fieldText, "protected") {
				variable.Visibility = "protected"
			}

			if variable.Name != "" {
				variables = append(variables, variable)
			}
		}
	})

	return variables
}

// extractJavaImports extracts Java import statements
func (p *TreeSitterParser) extractJavaImports(node *sitter.Node, source []byte) []types.Import {
	var imports []types.Import

	importStmt := types.Import{
		StartLine: p.getLineNumber(node),
	}

	// Extract import path
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "scoped_identifier" || child.Type() == "identifier" {
			importStmt.Module = p.getNodeText(child, source)
			break
		}
	}

	if importStmt.Module != "" {
		imports = append(imports, importStmt)
	}

	return imports
}

// extractJavaParameters extracts method parameters
func (p *TreeSitterParser) extractJavaParameters(node *sitter.Node, source []byte) []string {
	var parameters []string

	p.walkNode(node, source, func(n *sitter.Node) {
		if n.Type() == "formal_parameter" {
			param := p.getNodeText(n, source)
			parameters = append(parameters, param)
		}
	})

	return parameters
}
