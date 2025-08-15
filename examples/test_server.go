package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/indexer"
	"github.com/my-mcp/code-indexer/internal/repository"
	"github.com/my-mcp/code-indexer/internal/search"
	"github.com/my-mcp/code-indexer/pkg/types"
)

func main() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "code-indexer-test")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Testing MCP Code Indexer in: %s\n", tempDir)

	// Create test configuration
	cfg := config.DefaultConfig()
	cfg.Indexer.IndexDir = filepath.Join(tempDir, "index")
	cfg.Indexer.RepoDir = filepath.Join(tempDir, "repos")
	cfg.Logging.Level = "info"

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create components
	repoMgr, err := repository.NewManager(cfg.Indexer.RepoDir, logger)
	if err != nil {
		log.Fatalf("Failed to create repository manager: %v", err)
	}

	searcher, err := search.NewEngine(cfg.Indexer.IndexDir, logger)
	if err != nil {
		log.Fatalf("Failed to create search engine: %v", err)
	}
	defer searcher.Close()

	idx, err := indexer.New(cfg, repoMgr, searcher, logger)
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}

	// Create a test repository
	testRepoPath := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(testRepoPath, 0755); err != nil {
		log.Fatalf("Failed to create test repo: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"main.go": `package main

import (
	"fmt"
	"net/http"
)

// main is the entry point of the application
func main() {
	fmt.Println("Hello, World!")
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}

// handleRequest handles HTTP requests
func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %s", r.URL.Path)
}
`,
		"utils.py": `"""
Utility functions for the application
"""

def calculate_sum(a, b):
    """Calculate the sum of two numbers"""
    return a + b

class Calculator:
    """A simple calculator class"""
    
    def __init__(self):
        self.history = []
    
    def add(self, x, y):
        """Add two numbers"""
        result = x + y
        self.history.append(f"{x} + {y} = {result}")
        return result
    
    def get_history(self):
        """Get calculation history"""
        return self.history
`,
		"config.js": `// Configuration for the application
const config = {
    port: 3000,
    database: {
        host: 'localhost',
        port: 5432,
        name: 'myapp'
    }
};

function getConfig() {
    return config;
}

module.exports = { config, getConfig };
`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(testRepoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	fmt.Println("Created test repository with sample files")

	// Test indexing
	fmt.Println("\n=== Testing Repository Indexing ===")
	ctx := context.Background()
	
	start := time.Now()
	repo, err := idx.IndexRepository(ctx, testRepoPath, "test-repo")
	if err != nil {
		log.Fatalf("Failed to index repository: %v", err)
	}
	
	fmt.Printf("✅ Repository indexed successfully in %v\n", time.Since(start))
	fmt.Printf("   - Repository ID: %s\n", repo.ID)
	fmt.Printf("   - Name: %s\n", repo.Name)
	fmt.Printf("   - Files: %d\n", repo.FileCount)
	fmt.Printf("   - Lines: %d\n", repo.TotalLines)
	fmt.Printf("   - Languages: %v\n", repo.Languages)

	// Test searching
	fmt.Println("\n=== Testing Search Functionality ===")
	
	// Test 1: Search for functions
	fmt.Println("\n1. Searching for functions containing 'handle':")
	results, err := searcher.Search(ctx, types.SearchQuery{
		Query:      "handle",
		Type:       "function",
		MaxResults: 10,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	
	fmt.Printf("   Found %d results:\n", len(results))
	for _, result := range results {
		fmt.Printf("   - %s:%d: %s\n", result.FilePath, result.StartLine, result.Name)
	}

	// Test 2: Search for classes
	fmt.Println("\n2. Searching for classes:")
	results, err = searcher.Search(ctx, types.SearchQuery{
		Query:      "Calculator",
		Type:       "class",
		MaxResults: 10,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	
	fmt.Printf("   Found %d results:\n", len(results))
	for _, result := range results {
		fmt.Printf("   - %s:%d: %s\n", result.FilePath, result.StartLine, result.Name)
	}

	// Test 3: Content search
	fmt.Println("\n3. Searching for content containing 'configuration':")
	results, err = searcher.Search(ctx, types.SearchQuery{
		Query:      "configuration",
		MaxResults: 10,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	
	fmt.Printf("   Found %d results:\n", len(results))
	for _, result := range results {
		fmt.Printf("   - %s:%d (%s): %s\n", result.FilePath, result.StartLine, result.Type, result.Snippet)
	}

	// Test 4: Language-specific search
	fmt.Println("\n4. Searching for Python functions:")
	results, err = searcher.Search(ctx, types.SearchQuery{
		Query:      "def",
		Language:   "python",
		MaxResults: 10,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	
	fmt.Printf("   Found %d results:\n", len(results))
	for _, result := range results {
		fmt.Printf("   - %s:%d: %s\n", result.FilePath, result.StartLine, result.Content)
	}

	// Test file metadata
	fmt.Println("\n=== Testing File Metadata ===")
	metadata, err := searcher.GetFileMetadata(ctx, "main.go", "test-repo")
	if err != nil {
		log.Fatalf("Failed to get file metadata: %v", err)
	}
	
	fmt.Printf("✅ File metadata retrieved:\n")
	fmt.Printf("   - File: %s\n", metadata.RelativePath)
	fmt.Printf("   - Language: %s\n", metadata.Language)
	fmt.Printf("   - Lines: %d\n", metadata.Lines)
	fmt.Printf("   - Functions: %d\n", len(metadata.Functions))
	fmt.Printf("   - Variables: %d\n", len(metadata.Variables))
	fmt.Printf("   - Comments: %d\n", len(metadata.Comments))
	fmt.Printf("   - Imports: %d\n", len(metadata.Imports))

	// Test repository listing
	fmt.Println("\n=== Testing Repository Listing ===")
	repositories, err := searcher.ListRepositories(ctx)
	if err != nil {
		log.Fatalf("Failed to list repositories: %v", err)
	}
	
	fmt.Printf("✅ Found %d repositories:\n", len(repositories))
	for _, repo := range repositories {
		fmt.Printf("   - %s (%s): %d files, languages: %v\n", 
			repo.Name, repo.ID, repo.FileCount, repo.Languages)
	}

	// Test index statistics
	fmt.Println("\n=== Testing Index Statistics ===")
	stats, err := searcher.GetIndexStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get index stats: %v", err)
	}
	
	fmt.Printf("✅ Index statistics:\n")
	fmt.Printf("   - Total repositories: %d\n", stats.TotalRepositories)
	fmt.Printf("   - Total files: %d\n", stats.TotalFiles)
	fmt.Printf("   - Total functions: %d\n", stats.TotalFunctions)
	fmt.Printf("   - Total classes: %d\n", stats.TotalClasses)
	fmt.Printf("   - Total variables: %d\n", stats.TotalVariables)
	fmt.Printf("   - Language stats: %v\n", stats.LanguageStats)

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("\nThe MCP Code Indexer is working correctly and ready to use.")
}
