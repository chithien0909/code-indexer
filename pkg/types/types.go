package types

import (
	"time"
)

// Repository represents a Git repository that has been indexed
type Repository struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Path            string            `json:"path"`
	URL             string            `json:"url,omitempty"`
	IndexedAt       time.Time         `json:"indexed_at"`
	FileCount       int               `json:"file_count"`
	TotalLines      int               `json:"total_lines"`
	Languages       []string          `json:"languages"`
	LastCommit      string            `json:"last_commit,omitempty"`
	Branch          string            `json:"branch,omitempty"`
	LastIndexedHash string            `json:"last_indexed_hash,omitempty"`
	Submodules      []Submodule       `json:"submodules,omitempty"`
	IndexingMode    string            `json:"indexing_mode,omitempty"` // "full", "incremental", "sparse"
	SparsePatterns  []string          `json:"sparse_patterns,omitempty"`
	CommitHistory   []CommitInfo      `json:"commit_history,omitempty"`
}

// Submodule represents a Git submodule
type Submodule struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	URL    string `json:"url"`
	Hash   string `json:"hash"`
	Branch string `json:"branch,omitempty"`
}

// CommitInfo represents information about a Git commit
type CommitInfo struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Date      time.Time `json:"date"`
	Files     []string  `json:"files,omitempty"`
}

// IncrementalIndexRequest represents a request for incremental indexing
type IncrementalIndexRequest struct {
	RepositoryID string `json:"repository_id"`
	FromCommit   string `json:"from_commit,omitempty"`
	ToCommit     string `json:"to_commit,omitempty"`
	ForceRebuild bool   `json:"force_rebuild,omitempty"`
}

// CodeChunk represents a semantic chunk of code
type CodeChunk struct {
	ID           string                 `json:"id"`
	FileID       string                 `json:"file_id"`
	Type         string                 `json:"type"` // "function", "class", "method", "block"
	Name         string                 `json:"name,omitempty"`
	StartLine    int                    `json:"start_line"`
	EndLine      int                    `json:"end_line"`
	Content      string                 `json:"content"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

// CodeFile represents a source code file with its metadata
type CodeFile struct {
	ID           string      `json:"id"`
	RepositoryID string      `json:"repository_id"`
	Path         string      `json:"path"`
	RelativePath string      `json:"relative_path"`
	Language     string      `json:"language"`
	Extension    string      `json:"extension"`
	Size         int64       `json:"size"`
	Lines        int         `json:"lines"`
	Content      string      `json:"content,omitempty"`
	Hash         string      `json:"hash"`
	ModifiedAt   time.Time   `json:"modified_at"`
	IndexedAt    time.Time   `json:"indexed_at"`
	Functions    []Function  `json:"functions,omitempty"`
	Classes      []Class     `json:"classes,omitempty"`
	Variables    []Variable  `json:"variables,omitempty"`
	Imports      []Import    `json:"imports,omitempty"`
	Comments     []Comment   `json:"comments,omitempty"`
	Chunks       []CodeChunk `json:"chunks,omitempty"`
	TreeSitterAST interface{} `json:"tree_sitter_ast,omitempty"`
}

// Function represents a function or method definition
type Function struct {
	Name        string     `json:"name"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
	Parameters  []string   `json:"parameters,omitempty"`
	ReturnType  string     `json:"return_type,omitempty"`
	Visibility  string     `json:"visibility,omitempty"`
	IsMethod    bool       `json:"is_method"`
	ClassName   string     `json:"class_name,omitempty"`
	DocString   string     `json:"doc_string,omitempty"`
	Signature   string     `json:"signature"`
	Body        string     `json:"body,omitempty"`
	Annotations []string   `json:"annotations,omitempty"`
}

// Class represents a class or struct definition
type Class struct {
	Name        string     `json:"name"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
	Visibility  string     `json:"visibility,omitempty"`
	SuperClass  string     `json:"super_class,omitempty"`
	Interfaces  []string   `json:"interfaces,omitempty"`
	DocString   string     `json:"doc_string,omitempty"`
	Methods     []Function `json:"methods,omitempty"`
	Fields      []Variable `json:"fields,omitempty"`
	Annotations []string   `json:"annotations,omitempty"`
}

// Variable represents a variable or constant declaration
type Variable struct {
	Name       string   `json:"name"`
	Type       string   `json:"type,omitempty"`
	Value      string   `json:"value,omitempty"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	Visibility string   `json:"visibility,omitempty"`
	IsConstant bool     `json:"is_constant"`
	IsGlobal   bool     `json:"is_global"`
	Scope      string   `json:"scope,omitempty"`
}

// Import represents an import or include statement
type Import struct {
	Module    string `json:"module"`
	Alias     string `json:"alias,omitempty"`
	StartLine int    `json:"start_line"`
	IsWildcard bool  `json:"is_wildcard"`
}

// Comment represents a comment in the code
type Comment struct {
	Text      string `json:"text"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Type      string `json:"type"` // "line", "block", "doc"
}

// SearchResult represents a search result
type SearchResult struct {
	ID           string            `json:"id"`
	RepositoryID string            `json:"repository_id"`
	Repository   string            `json:"repository"`
	FilePath     string            `json:"file_path"`
	Language     string            `json:"language"`
	Type         string            `json:"type"` // "function", "class", "variable", "content", "comment"
	Name         string            `json:"name,omitempty"`
	Content      string            `json:"content"`
	Snippet      string            `json:"snippet,omitempty"`
	StartLine    int               `json:"start_line"`
	EndLine      int               `json:"end_line"`
	Score        float64           `json:"score"`
	Highlights   map[string]string `json:"highlights,omitempty"`
	Context      map[string]any    `json:"context,omitempty"`
}

// SearchQuery represents a search query with filters
type SearchQuery struct {
	Query      string   `json:"query"`
	Type       string   `json:"type,omitempty"`       // "function", "class", "variable", "content", "file", "comment"
	Language   string   `json:"language,omitempty"`   // Filter by programming language
	Repository string   `json:"repository,omitempty"` // Filter by repository name
	FilePath   string   `json:"file_path,omitempty"`  // Filter by file path pattern
	MaxResults int      `json:"max_results,omitempty"`
	Fuzzy      bool     `json:"fuzzy,omitempty"`
}

// IndexStats represents indexing statistics
type IndexStats struct {
	TotalRepositories int                    `json:"total_repositories"`
	TotalFiles        int                    `json:"total_files"`
	TotalLines        int                    `json:"total_lines"`
	TotalFunctions    int                    `json:"total_functions"`
	TotalClasses      int                    `json:"total_classes"`
	TotalVariables    int                    `json:"total_variables"`
	LanguageStats     map[string]int         `json:"language_stats"`
	RepositoryStats   map[string]Repository  `json:"repository_stats"`
	LastIndexed       time.Time              `json:"last_indexed"`
}

// ParserConfig represents configuration for language parsers
type ParserConfig struct {
	Language         string   `json:"language"`
	Extensions       []string `json:"extensions"`
	CommentPrefixes  []string `json:"comment_prefixes"`
	BlockCommentStart string  `json:"block_comment_start,omitempty"`
	BlockCommentEnd   string  `json:"block_comment_end,omitempty"`
	DocCommentPrefix  string  `json:"doc_comment_prefix,omitempty"`
}

// IndexingProgress represents the progress of an indexing operation
type IndexingProgress struct {
	RepositoryID    string    `json:"repository_id"`
	Repository      string    `json:"repository"`
	Status          string    `json:"status"` // "starting", "cloning", "parsing", "indexing", "completed", "failed"
	FilesProcessed  int       `json:"files_processed"`
	TotalFiles      int       `json:"total_files"`
	CurrentFile     string    `json:"current_file,omitempty"`
	Error           string    `json:"error,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ElapsedSeconds  float64   `json:"elapsed_seconds"`
}

// ML-related types

// CodeEmbedding represents a vector embedding of code
type CodeEmbedding struct {
	ID         string    `json:"id"`
	FileID     string    `json:"file_id"`
	ChunkID    string    `json:"chunk_id,omitempty"`
	Vector     []float32 `json:"vector"`
	Dimensions int       `json:"dimensions"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
}

// SimilarityResult represents code similarity analysis result
type SimilarityResult struct {
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	Score        float64 `json:"score"`
	Type         string  `json:"type"` // "function", "class", "file", "chunk"
	SourceSnippet string `json:"source_snippet"`
	TargetSnippet string `json:"target_snippet"`
	Explanation   string `json:"explanation,omitempty"`
}

// QualityMetrics represents code quality prediction results
type QualityMetrics struct {
	FileID          string  `json:"file_id"`
	Maintainability float64 `json:"maintainability"`
	Complexity      float64 `json:"complexity"`
	Readability     float64 `json:"readability"`
	TestCoverage    float64 `json:"test_coverage,omitempty"`
	Documentation   float64 `json:"documentation"`
	OverallScore    float64 `json:"overall_score"`
	Suggestions     []string `json:"suggestions,omitempty"`
}

// IntentClassification represents code intent classification result
type IntentClassification struct {
	CodeSnippet string            `json:"code_snippet"`
	Intent      string            `json:"intent"`
	Confidence  float64           `json:"confidence"`
	Categories  map[string]float64 `json:"categories"`
	Description string            `json:"description,omitempty"`
}

// CodeSummary represents AI-generated code summary
type CodeSummary struct {
	FileID      string   `json:"file_id"`
	Summary     string   `json:"summary"`
	KeyPoints   []string `json:"key_points"`
	Functions   []string `json:"functions,omitempty"`
	Classes     []string `json:"classes,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Complexity   string   `json:"complexity"` // "low", "medium", "high"
}

// PatternDetection represents detected code patterns
type PatternDetection struct {
	Pattern     string   `json:"pattern"`
	Type        string   `json:"type"` // "design_pattern", "anti_pattern", "code_smell"
	Confidence  float64  `json:"confidence"`
	Locations   []Location `json:"locations"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // "low", "medium", "high", "critical"
}

// Location represents a location in code
type Location struct {
	FileID    string `json:"file_id"`
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Snippet   string `json:"snippet,omitempty"`
}

// RefactoringSuggestion represents ML-based refactoring suggestions
type RefactoringSuggestion struct {
	Type        string   `json:"type"` // "extract_method", "rename", "move_class", etc.
	Priority    string   `json:"priority"` // "low", "medium", "high"
	Confidence  float64  `json:"confidence"`
	Location    Location `json:"location"`
	Description string   `json:"description"`
	Before      string   `json:"before,omitempty"`
	After       string   `json:"after,omitempty"`
	Benefits    []string `json:"benefits,omitempty"`
}

// BugPrediction represents potential bug prediction
type BugPrediction struct {
	Type        string   `json:"type"` // "null_pointer", "memory_leak", "logic_error", etc.
	Probability float64  `json:"probability"`
	Severity    string   `json:"severity"` // "low", "medium", "high", "critical"
	Location    Location `json:"location"`
	Description string   `json:"description"`
	Suggestion  string   `json:"suggestion,omitempty"`
}

// Model-based AI Types

// CodeGeneration represents AI-generated code
type CodeGeneration struct {
	Prompt        string                 `json:"prompt"`
	Language      string                 `json:"language"`
	GeneratedCode string                 `json:"generated_code"`
	Confidence    float64                `json:"confidence"`
	Model         string                 `json:"model"`
	GeneratedAt   time.Time              `json:"generated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CodeAnalysis represents AI code analysis results
type CodeAnalysis struct {
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	Summary     string    `json:"summary"`
	Quality     float64   `json:"quality_score"`
	Suggestions []string  `json:"suggestions"`
	Issues      []string  `json:"issues"`
	Complexity  string    `json:"complexity"`
	Model       string    `json:"model"`
	AnalyzedAt  time.Time `json:"analyzed_at"`
}

// CodeExplanation represents AI code explanation
type CodeExplanation struct {
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	Explanation string    `json:"explanation"`
	KeyConcepts []string  `json:"key_concepts"`
	Purpose     string    `json:"purpose"`
	Complexity  string    `json:"complexity"`
	Model       string    `json:"model"`
	ExplainedAt time.Time `json:"explained_at"`
}






