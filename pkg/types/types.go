package types

import (
	"time"
)

// Repository represents a Git repository that has been indexed
type Repository struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	URL         string    `json:"url,omitempty"`
	IndexedAt   time.Time `json:"indexed_at"`
	FileCount   int       `json:"file_count"`
	TotalLines  int       `json:"total_lines"`
	Languages   []string  `json:"languages"`
	LastCommit  string    `json:"last_commit,omitempty"`
	Branch      string    `json:"branch,omitempty"`
}

// CodeFile represents a source code file with its metadata
type CodeFile struct {
	ID           string     `json:"id"`
	RepositoryID string     `json:"repository_id"`
	Path         string     `json:"path"`
	RelativePath string     `json:"relative_path"`
	Language     string     `json:"language"`
	Extension    string     `json:"extension"`
	Size         int64      `json:"size"`
	Lines        int        `json:"lines"`
	Content      string     `json:"content,omitempty"`
	Hash         string     `json:"hash"`
	ModifiedAt   time.Time  `json:"modified_at"`
	IndexedAt    time.Time  `json:"indexed_at"`
	Functions    []Function `json:"functions,omitempty"`
	Classes      []Class    `json:"classes,omitempty"`
	Variables    []Variable `json:"variables,omitempty"`
	Imports      []Import   `json:"imports,omitempty"`
	Comments     []Comment  `json:"comments,omitempty"`
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
