package parser

import (
	"regexp"
	"strings"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// Parser interface for language-specific parsers
type Parser interface {
	Parse(content string, filePath string) (*types.CodeFile, error)
	GetLanguage() string
}

// Registry holds all available parsers
type Registry struct {
	parsers map[string]Parser
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	registry := &Registry{
		parsers: make(map[string]Parser),
	}

	// Register built-in parsers
	registry.Register(NewGoParser())
	registry.Register(NewPythonParser())
	registry.Register(NewJavaScriptParser())
	registry.Register(NewJavaParser())
	registry.Register(NewGenericParser())

	return registry
}

// Register adds a parser to the registry
func (r *Registry) Register(parser Parser) {
	r.parsers[parser.GetLanguage()] = parser
}

// GetParser returns a parser for the given language
func (r *Registry) GetParser(language string) Parser {
	if parser, exists := r.parsers[language]; exists {
		return parser
	}
	// Return generic parser as fallback
	return r.parsers["generic"]
}

// ParseFile parses a file and extracts metadata
func (r *Registry) ParseFile(content string, filePath, language string) (*types.CodeFile, error) {
	parser := r.GetParser(language)
	return parser.Parse(content, filePath)
}

// BaseParser provides common functionality for all parsers
type BaseParser struct {
	language string
}

// GetLanguage returns the language this parser handles
func (p *BaseParser) GetLanguage() string {
	return p.language
}

// extractComments extracts comments from source code
func (p *BaseParser) extractComments(content string, lineCommentPrefix, blockCommentStart, blockCommentEnd string) []types.Comment {
	var comments []types.Comment
	lines := strings.Split(content, "\n")

	inBlockComment := false
	blockCommentStartLine := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Handle block comments
		if blockCommentStart != "" && blockCommentEnd != "" {
			if !inBlockComment && strings.Contains(trimmedLine, blockCommentStart) {
				inBlockComment = true
				blockCommentStartLine = lineNum
			}
			if inBlockComment && strings.Contains(trimmedLine, blockCommentEnd) {
				// Extract block comment content
				commentText := p.extractBlockCommentText(lines, blockCommentStartLine-1, lineNum-1, blockCommentStart, blockCommentEnd)
				comments = append(comments, types.Comment{
					Text:      commentText,
					StartLine: blockCommentStartLine,
					EndLine:   lineNum,
					Type:      "block",
				})
				inBlockComment = false
			}
		}

		// Handle line comments
		if lineCommentPrefix != "" && strings.HasPrefix(trimmedLine, lineCommentPrefix) {
			commentText := strings.TrimSpace(strings.TrimPrefix(trimmedLine, lineCommentPrefix))
			commentType := "line"
			if strings.HasPrefix(trimmedLine, lineCommentPrefix+lineCommentPrefix) {
				commentType = "doc"
			}
			
			comments = append(comments, types.Comment{
				Text:      commentText,
				StartLine: lineNum,
				EndLine:   lineNum,
				Type:      commentType,
			})
		}
	}

	return comments
}

// extractBlockCommentText extracts text from a block comment
func (p *BaseParser) extractBlockCommentText(lines []string, startLine, endLine int, startMarker, endMarker string) string {
	var commentLines []string
	
	for i := startLine; i <= endLine && i < len(lines); i++ {
		line := lines[i]
		
		// Remove start marker from first line
		if i == startLine {
			if idx := strings.Index(line, startMarker); idx >= 0 {
				line = line[idx+len(startMarker):]
			}
		}
		
		// Remove end marker from last line
		if i == endLine {
			if idx := strings.Index(line, endMarker); idx >= 0 {
				line = line[:idx]
			}
		}
		
		// Clean up common comment formatting
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(line[1:])
		}
		
		commentLines = append(commentLines, line)
	}
	
	return strings.Join(commentLines, " ")
}

// countLines counts the number of lines in content
func (p *BaseParser) countLines(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

// findLineNumber finds the line number of a substring in content
func (p *BaseParser) findLineNumber(content, substring string) int {
	index := strings.Index(content, substring)
	if index == -1 {
		return 1
	}
	
	return strings.Count(content[:index], "\n") + 1
}

// GenericParser provides basic parsing for any text file
type GenericParser struct {
	BaseParser
}

// NewGenericParser creates a new generic parser
func NewGenericParser() *GenericParser {
	return &GenericParser{
		BaseParser: BaseParser{language: "generic"},
	}
}

// Parse implements basic parsing for any file
func (p *GenericParser) Parse(content string, filePath string) (*types.CodeFile, error) {
	file := &types.CodeFile{
		Path:     filePath,
		Language: "generic",
		Lines:    p.countLines(content),
		Content:  content,
	}

	// Extract basic comments (try common comment styles)
	comments := []types.Comment{}
	
	// Try different comment styles
	commentStyles := []struct {
		line, blockStart, blockEnd string
	}{
		{"//", "/*", "*/"},     // C-style
		{"#", "", ""},          // Shell/Python style
		{"--", "/*", "*/"},     // SQL style
		{";", "", ""},          // Lisp style
	}

	for _, style := range commentStyles {
		styleComments := p.extractComments(content, style.line, style.blockStart, style.blockEnd)
		comments = append(comments, styleComments...)
	}

	file.Comments = comments
	return file, nil
}

// GoParser parses Go source files
type GoParser struct {
	BaseParser
}

// NewGoParser creates a new Go parser
func NewGoParser() *GoParser {
	return &GoParser{
		BaseParser: BaseParser{language: "go"},
	}
}

// Parse parses Go source code
func (p *GoParser) Parse(content string, filePath string) (*types.CodeFile, error) {
	file := &types.CodeFile{
		Path:     filePath,
		Language: "go",
		Lines:    p.countLines(content),
		Content:  content,
	}

	// Extract comments
	file.Comments = p.extractComments(content, "//", "/*", "*/")

	// Extract imports
	file.Imports = p.extractGoImports(content)

	// Extract functions
	file.Functions = p.extractGoFunctions(content)

	// Extract structs (as classes)
	file.Classes = p.extractGoStructs(content)

	// Extract variables and constants
	file.Variables = p.extractGoVariables(content)

	return file, nil
}

// extractGoImports extracts import statements from Go code
func (p *GoParser) extractGoImports(content string) []types.Import {
	var imports []types.Import
	
	// Single import pattern
	singleImportRe := regexp.MustCompile(`import\s+"([^"]+)"`)
	matches := singleImportRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		imports = append(imports, types.Import{
			Module:    match[1],
			StartLine: p.findLineNumber(content, match[0]),
		})
	}

	// Multi-line import pattern
	multiImportRe := regexp.MustCompile(`import\s*\(\s*([^)]+)\)`)
	multiMatches := multiImportRe.FindAllStringSubmatch(content, -1)
	for _, match := range multiMatches {
		importBlock := match[1]
		lines := strings.Split(importBlock, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			// Extract module name from quoted string
			if strings.Contains(line, `"`) {
				moduleRe := regexp.MustCompile(`"([^"]+)"`)
				if moduleMatch := moduleRe.FindStringSubmatch(line); len(moduleMatch) > 1 {
					imports = append(imports, types.Import{
						Module:    moduleMatch[1],
						StartLine: p.findLineNumber(content, line),
					})
				}
			}
		}
	}

	return imports
}

// extractGoFunctions extracts function definitions from Go code
func (p *GoParser) extractGoFunctions(content string) []types.Function {
	var functions []types.Function
	
	// Function pattern: func (receiver) name(params) (returns) {
	funcRe := regexp.MustCompile(`func\s*(?:\([^)]*\))?\s*(\w+)\s*\([^)]*\)(?:\s*\([^)]*\))?\s*{`)
	matches := funcRe.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		funcName := match[1]
		startLine := p.findLineNumber(content, match[0])
		
		functions = append(functions, types.Function{
			Name:      funcName,
			StartLine: startLine,
			Signature: strings.TrimSpace(match[0]),
		})
	}

	return functions
}

// extractGoStructs extracts struct definitions from Go code
func (p *GoParser) extractGoStructs(content string) []types.Class {
	var structs []types.Class
	
	// Struct pattern: type Name struct {
	structRe := regexp.MustCompile(`type\s+(\w+)\s+struct\s*{`)
	matches := structRe.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		structName := match[1]
		startLine := p.findLineNumber(content, match[0])
		
		structs = append(structs, types.Class{
			Name:      structName,
			StartLine: startLine,
		})
	}

	return structs
}

// extractGoVariables extracts variable and constant declarations from Go code
func (p *GoParser) extractGoVariables(content string) []types.Variable {
	var variables []types.Variable
	
	// Variable patterns
	patterns := []struct {
		regex     *regexp.Regexp
		isConstant bool
	}{
		{regexp.MustCompile(`var\s+(\w+)(?:\s+(\w+))?\s*=?`), false},
		{regexp.MustCompile(`const\s+(\w+)(?:\s+(\w+))?\s*=`), true},
		{regexp.MustCompile(`(\w+)\s*:=`), false}, // Short variable declaration
	}
	
	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			varName := match[1]
			varType := ""
			if len(match) > 2 {
				varType = match[2]
			}
			
			variables = append(variables, types.Variable{
				Name:       varName,
				Type:       varType,
				StartLine:  p.findLineNumber(content, match[0]),
				IsConstant: pattern.isConstant,
			})
		}
	}

	return variables
}

// PythonParser parses Python source files
type PythonParser struct {
	BaseParser
}

// NewPythonParser creates a new Python parser
func NewPythonParser() *PythonParser {
	return &PythonParser{
		BaseParser: BaseParser{language: "python"},
	}
}

// Parse parses Python source code
func (p *PythonParser) Parse(content string, filePath string) (*types.CodeFile, error) {
	file := &types.CodeFile{
		Path:     filePath,
		Language: "python",
		Lines:    p.countLines(content),
		Content:  content,
	}

	// Extract comments
	file.Comments = p.extractComments(content, "#", `"""`, `"""`)

	// Extract imports
	file.Imports = p.extractPythonImports(content)

	// Extract functions
	file.Functions = p.extractPythonFunctions(content)

	// Extract classes
	file.Classes = p.extractPythonClasses(content)

	// Extract variables
	file.Variables = p.extractPythonVariables(content)

	return file, nil
}

// extractPythonImports extracts import statements from Python code
func (p *PythonParser) extractPythonImports(content string) []types.Import {
	var imports []types.Import

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`import\s+(\w+(?:\.\w+)*)`),
		regexp.MustCompile(`from\s+(\w+(?:\.\w+)*)\s+import\s+(.+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			module := match[1]
			alias := ""
			if len(match) > 2 {
				alias = match[2]
			}

			imports = append(imports, types.Import{
				Module:    module,
				Alias:     alias,
				StartLine: p.findLineNumber(content, match[0]),
			})
		}
	}

	return imports
}

// extractPythonFunctions extracts function definitions from Python code
func (p *PythonParser) extractPythonFunctions(content string) []types.Function {
	var functions []types.Function

	funcRe := regexp.MustCompile(`def\s+(\w+)\s*\([^)]*\):`)
	matches := funcRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		funcName := match[1]
		startLine := p.findLineNumber(content, match[0])

		functions = append(functions, types.Function{
			Name:      funcName,
			StartLine: startLine,
			Signature: strings.TrimSpace(match[0]),
		})
	}

	return functions
}

// extractPythonClasses extracts class definitions from Python code
func (p *PythonParser) extractPythonClasses(content string) []types.Class {
	var classes []types.Class

	classRe := regexp.MustCompile(`class\s+(\w+)(?:\([^)]*\))?:`)
	matches := classRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		className := match[1]
		startLine := p.findLineNumber(content, match[0])

		classes = append(classes, types.Class{
			Name:      className,
			StartLine: startLine,
		})
	}

	return classes
}

// extractPythonVariables extracts variable assignments from Python code
func (p *PythonParser) extractPythonVariables(content string) []types.Variable {
	var variables []types.Variable

	varRe := regexp.MustCompile(`^(\w+)\s*=`)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if matches := varRe.FindStringSubmatch(line); len(matches) > 1 {
			varName := matches[1]

			variables = append(variables, types.Variable{
				Name:      varName,
				StartLine: i + 1,
			})
		}
	}

	return variables
}

// JavaScriptParser parses JavaScript/TypeScript source files
type JavaScriptParser struct {
	BaseParser
}

// NewJavaScriptParser creates a new JavaScript parser
func NewJavaScriptParser() *JavaScriptParser {
	return &JavaScriptParser{
		BaseParser: BaseParser{language: "javascript"},
	}
}

// Parse parses JavaScript source code
func (p *JavaScriptParser) Parse(content string, filePath string) (*types.CodeFile, error) {
	file := &types.CodeFile{
		Path:     filePath,
		Language: "javascript",
		Lines:    p.countLines(content),
		Content:  content,
	}

	// Extract comments
	file.Comments = p.extractComments(content, "//", "/*", "*/")

	// Extract imports
	file.Imports = p.extractJSImports(content)

	// Extract functions
	file.Functions = p.extractJSFunctions(content)

	// Extract classes
	file.Classes = p.extractJSClasses(content)

	// Extract variables
	file.Variables = p.extractJSVariables(content)

	return file, nil
}

// extractJSImports extracts import statements from JavaScript code
func (p *JavaScriptParser) extractJSImports(content string) []types.Import {
	var imports []types.Import

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`import\s+.*\s+from\s+['"]([^'"]+)['"]`),
		regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			module := match[1]

			imports = append(imports, types.Import{
				Module:    module,
				StartLine: p.findLineNumber(content, match[0]),
			})
		}
	}

	return imports
}

// extractJSFunctions extracts function definitions from JavaScript code
func (p *JavaScriptParser) extractJSFunctions(content string) []types.Function {
	var functions []types.Function

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`function\s+(\w+)\s*\([^)]*\)`),
		regexp.MustCompile(`(\w+)\s*:\s*function\s*\([^)]*\)`),
		regexp.MustCompile(`(\w+)\s*=>\s*{`),
		regexp.MustCompile(`const\s+(\w+)\s*=\s*\([^)]*\)\s*=>`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			funcName := match[1]
			startLine := p.findLineNumber(content, match[0])

			functions = append(functions, types.Function{
				Name:      funcName,
				StartLine: startLine,
				Signature: strings.TrimSpace(match[0]),
			})
		}
	}

	return functions
}

// extractJSClasses extracts class definitions from JavaScript code
func (p *JavaScriptParser) extractJSClasses(content string) []types.Class {
	var classes []types.Class

	classRe := regexp.MustCompile(`class\s+(\w+)(?:\s+extends\s+\w+)?\s*{`)
	matches := classRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		className := match[1]
		startLine := p.findLineNumber(content, match[0])

		classes = append(classes, types.Class{
			Name:      className,
			StartLine: startLine,
		})
	}

	return classes
}

// extractJSVariables extracts variable declarations from JavaScript code
func (p *JavaScriptParser) extractJSVariables(content string) []types.Variable {
	var variables []types.Variable

	patterns := []struct {
		regex     *regexp.Regexp
		isConstant bool
	}{
		{regexp.MustCompile(`var\s+(\w+)`), false},
		{regexp.MustCompile(`let\s+(\w+)`), false},
		{regexp.MustCompile(`const\s+(\w+)`), true},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			varName := match[1]

			variables = append(variables, types.Variable{
				Name:       varName,
				StartLine:  p.findLineNumber(content, match[0]),
				IsConstant: pattern.isConstant,
			})
		}
	}

	return variables
}

// JavaParser parses Java source files
type JavaParser struct {
	BaseParser
}

// NewJavaParser creates a new Java parser
func NewJavaParser() *JavaParser {
	return &JavaParser{
		BaseParser: BaseParser{language: "java"},
	}
}

// Parse parses Java source code
func (p *JavaParser) Parse(content string, filePath string) (*types.CodeFile, error) {
	file := &types.CodeFile{
		Path:     filePath,
		Language: "java",
		Lines:    p.countLines(content),
		Content:  content,
	}

	// Extract comments
	file.Comments = p.extractComments(content, "//", "/*", "*/")

	// Extract imports
	file.Imports = p.extractJavaImports(content)

	// Extract methods
	file.Functions = p.extractJavaMethods(content)

	// Extract classes
	file.Classes = p.extractJavaClasses(content)

	// Extract variables
	file.Variables = p.extractJavaVariables(content)

	return file, nil
}

// extractJavaImports extracts import statements from Java code
func (p *JavaParser) extractJavaImports(content string) []types.Import {
	var imports []types.Import

	importRe := regexp.MustCompile(`import\s+(?:static\s+)?([^;]+);`)
	matches := importRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		module := strings.TrimSpace(match[1])

		imports = append(imports, types.Import{
			Module:    module,
			StartLine: p.findLineNumber(content, match[0]),
		})
	}

	return imports
}

// extractJavaMethods extracts method definitions from Java code
func (p *JavaParser) extractJavaMethods(content string) []types.Function {
	var methods []types.Function

	methodRe := regexp.MustCompile(`(?:public|private|protected)?\s*(?:static)?\s*\w+\s+(\w+)\s*\([^)]*\)\s*{`)
	matches := methodRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		methodName := match[1]
		startLine := p.findLineNumber(content, match[0])

		methods = append(methods, types.Function{
			Name:      methodName,
			StartLine: startLine,
			Signature: strings.TrimSpace(match[0]),
			IsMethod:  true,
		})
	}

	return methods
}

// extractJavaClasses extracts class definitions from Java code
func (p *JavaParser) extractJavaClasses(content string) []types.Class {
	var classes []types.Class

	classRe := regexp.MustCompile(`(?:public|private|protected)?\s*(?:abstract)?\s*class\s+(\w+)(?:\s+extends\s+\w+)?(?:\s+implements\s+[^{]+)?\s*{`)
	matches := classRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		className := match[1]
		startLine := p.findLineNumber(content, match[0])

		classes = append(classes, types.Class{
			Name:      className,
			StartLine: startLine,
		})
	}

	return classes
}

// extractJavaVariables extracts field declarations from Java code
func (p *JavaParser) extractJavaVariables(content string) []types.Variable {
	var variables []types.Variable

	fieldRe := regexp.MustCompile(`(?:public|private|protected)?\s*(?:static)?\s*(?:final)?\s*\w+\s+(\w+)\s*[=;]`)
	matches := fieldRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		varName := match[1]
		startLine := p.findLineNumber(content, match[0])

		variables = append(variables, types.Variable{
			Name:      varName,
			StartLine: startLine,
		})
	}

	return variables
}
