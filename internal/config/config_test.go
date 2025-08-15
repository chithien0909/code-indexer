package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test default values
	if cfg.Indexer.MaxFileSize != 1048576 {
		t.Errorf("Expected default max file size 1048576, got %d", cfg.Indexer.MaxFileSize)
	}

	if cfg.Search.MaxResults != 100 {
		t.Errorf("Expected default max results 100, got %d", cfg.Search.MaxResults)
	}

	if cfg.Server.Name != "Code Indexer" {
		t.Errorf("Expected default server name 'Code Indexer', got '%s'", cfg.Server.Name)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.Logging.Level)
	}

	// Test supported extensions
	if len(cfg.Indexer.SupportedExtensions) == 0 {
		t.Error("Expected default supported extensions to be populated")
	}

	// Check for common extensions
	hasGo := false
	hasPython := false
	for _, ext := range cfg.Indexer.SupportedExtensions {
		if ext == ".go" {
			hasGo = true
		}
		if ext == ".py" {
			hasPython = true
		}
	}

	if !hasGo {
		t.Error("Expected .go extension in default supported extensions")
	}
	if !hasPython {
		t.Error("Expected .py extension in default supported extensions")
	}
}

func TestIsFileSupported(t *testing.T) {
	cfg := DefaultConfig()

	// Test supported files
	if !cfg.IsFileSupported("main.go") {
		t.Error("Expected main.go to be supported")
	}

	if !cfg.IsFileSupported("script.py") {
		t.Error("Expected script.py to be supported")
	}

	if !cfg.IsFileSupported("app.js") {
		t.Error("Expected app.js to be supported")
	}

	// Test unsupported files
	if cfg.IsFileSupported("image.png") {
		t.Error("Expected image.png to not be supported")
	}

	if cfg.IsFileSupported("document.pdf") {
		t.Error("Expected document.pdf to not be supported")
	}

	if cfg.IsFileSupported("archive.zip") {
		t.Error("Expected archive.zip to not be supported")
	}
}

func TestShouldExcludeFile(t *testing.T) {
	cfg := DefaultConfig()

	// Test excluded patterns - need to match the actual patterns in config
	if !cfg.ShouldExcludeFile("project/node_modules/package/file.js") {
		t.Error("Expected node_modules files to be excluded")
	}

	if !cfg.ShouldExcludeFile("project/vendor/package/file.go") {
		t.Error("Expected vendor files to be excluded")
	}

	if !cfg.ShouldExcludeFile("project/.git/config") {
		t.Error("Expected .git files to be excluded")
	}

	if !cfg.ShouldExcludeFile("project/build/output.exe") {
		t.Error("Expected build files to be excluded")
	}

	if !cfg.ShouldExcludeFile("file.pyc") {
		t.Error("Expected .pyc files to be excluded")
	}

	// Test included files
	if cfg.ShouldExcludeFile("src/main.go") {
		t.Error("Expected src/main.go to not be excluded")
	}

	if cfg.ShouldExcludeFile("lib/utils.py") {
		t.Error("Expected lib/utils.py to not be excluded")
	}

	if cfg.ShouldExcludeFile("components/app.js") {
		t.Error("Expected components/app.js to not be excluded")
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := DefaultConfig()

	// Test with invalid values
	cfg.Indexer.MaxFileSize = -1
	cfg.Search.MaxResults = -1
	cfg.Search.FuzzyTolerance = 2.0
	cfg.Logging.Level = "invalid"

	err := cfg.validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Check that invalid values were corrected
	if cfg.Indexer.MaxFileSize <= 0 {
		t.Error("Expected max file size to be corrected to positive value")
	}

	if cfg.Search.MaxResults <= 0 {
		t.Error("Expected max results to be corrected to positive value")
	}

	if cfg.Search.FuzzyTolerance < 0 || cfg.Search.FuzzyTolerance > 1 {
		t.Error("Expected fuzzy tolerance to be corrected to valid range")
	}

	if cfg.Logging.Level != "info" {
		t.Error("Expected invalid log level to be corrected to 'info'")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
indexer:
  supported_extensions:
    - .go
    - .py
    - .js
  max_file_size: 2097152
  exclude_patterns:
    - "*/test/*"
    - "*.tmp"

search:
  max_results: 50
  highlight_snippets: false

server:
  name: "Test Server"
  version: "0.1.0"

logging:
  level: debug
  file: "test.log"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if cfg.Indexer.MaxFileSize != 2097152 {
		t.Errorf("Expected max file size 2097152, got %d", cfg.Indexer.MaxFileSize)
	}

	if cfg.Search.MaxResults != 50 {
		t.Errorf("Expected max results 50, got %d", cfg.Search.MaxResults)
	}

	if cfg.Search.HighlightSnippets != false {
		t.Error("Expected highlight snippets to be false")
	}

	if cfg.Server.Name != "Test Server" {
		t.Errorf("Expected server name 'Test Server', got '%s'", cfg.Server.Name)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.File != "test.log" {
		t.Errorf("Expected log file 'test.log', got '%s'", cfg.Logging.File)
	}

	// Test custom exclude patterns
	if !cfg.ShouldExcludeFile("src/test/file.go") {
		t.Error("Expected test files to be excluded")
	}

	if !cfg.ShouldExcludeFile("temp.tmp") {
		t.Error("Expected .tmp files to be excluded")
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	// Reset viper to clear any previous state
	viper.Reset()

	// Create a temporary directory to ensure no config file exists
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	os.Chdir(tempDir)

	// Try to load a non-existent config file - should use defaults
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Expected no error when config file not found, got: %v", err)
	}

	// Should return default config
	if cfg.Server.Name != "Code Indexer" {
		t.Errorf("Expected default config when file not found, got server name: '%s'", cfg.Server.Name)
	}
}

func TestConfigDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	
	cfg := DefaultConfig()
	cfg.Indexer.IndexDir = filepath.Join(tempDir, "index")
	cfg.Indexer.RepoDir = filepath.Join(tempDir, "repos")

	err := cfg.validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Check that directories were created
	if _, err := os.Stat(cfg.Indexer.IndexDir); os.IsNotExist(err) {
		t.Error("Expected index directory to be created")
	}

	if _, err := os.Stat(cfg.Indexer.RepoDir); os.IsNotExist(err) {
		t.Error("Expected repo directory to be created")
	}
}
