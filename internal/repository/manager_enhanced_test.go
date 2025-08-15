package repository

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestGitignoreSupport(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .gitignore file
	gitignoreContent := `*.log
node_modules/
.env
build/
*.tmp
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create test files
	testFiles := []string{
		"main.go",
		"app.log",           // Should be ignored
		"config.env",        // Should NOT be ignored (.env pattern is exact)
		".env",              // Should be ignored
		"src/helper.go",
		"src/debug.log",     // Should be ignored
		"build/output.js",   // Should be ignored
		"temp.tmp",          // Should be ignored
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create node_modules directory (should be ignored)
	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeModulesDir, "package.js"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file in node_modules: %v", err)
	}

	// Create manager
	logger := zap.NewNop()
	manager, err := NewManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test file walking with gitignore
	var discoveredFiles []string
	err = manager.WalkFiles(context.Background(), tempDir, func(filePath string, info fs.FileInfo) error {
		relPath, _ := filepath.Rel(tempDir, filePath)
		discoveredFiles = append(discoveredFiles, relPath)
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk files: %v", err)
	}

	// Check that ignored files are not discovered
	ignoredFiles := []string{"app.log", ".env", "src/debug.log", "build/output.js", "temp.tmp", "node_modules/package.js"}
	for _, ignored := range ignoredFiles {
		for _, discovered := range discoveredFiles {
			if discovered == ignored {
				t.Errorf("File %s should have been ignored but was discovered", ignored)
			}
		}
	}

	// Check that non-ignored files are discovered
	expectedFiles := []string{"main.go", "config.env", "src/helper.go"}
	for _, expected := range expectedFiles {
		found := false
		for _, discovered := range discoveredFiles {
			if discovered == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("File %s should have been discovered but was not", expected)
		}
	}
}

func TestSubmoduleDetection(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .gitmodules file
	gitmodulesContent := `[submodule "vendor/lib1"]
	path = vendor/lib1
	url = https://github.com/example/lib1.git
	branch = main

[submodule "vendor/lib2"]
	path = vendor/lib2
	url = https://github.com/example/lib2.git
`
	gitmodulesPath := filepath.Join(tempDir, ".gitmodules")
	err = os.WriteFile(gitmodulesPath, []byte(gitmodulesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitmodules: %v", err)
	}

	// Create manager
	logger := zap.NewNop()
	manager, err := NewManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test submodule detection
	submodules, err := manager.GetSubmodules(tempDir)
	if err != nil {
		t.Fatalf("Failed to get submodules: %v", err)
	}

	if len(submodules) != 2 {
		t.Errorf("Expected 2 submodules, got %d", len(submodules))
		// Debug: print what we got
		for i, sub := range submodules {
			t.Logf("Submodule %d: Name=%s, Path=%s, URL=%s, Branch=%s", i, sub.Name, sub.Path, sub.URL, sub.Branch)
		}
	}

	// Check first submodule
	if len(submodules) > 0 {
		sub1 := submodules[0]
		if sub1.Name != "vendor/lib1" {
			t.Errorf("Expected submodule name 'vendor/lib1', got '%s'", sub1.Name)
		}
		if sub1.Path != "vendor/lib1" {
			t.Errorf("Expected submodule path 'vendor/lib1', got '%s'", sub1.Path)
		}
		if sub1.URL != "https://github.com/example/lib1.git" {
			t.Errorf("Expected submodule URL 'https://github.com/example/lib1.git', got '%s'", sub1.URL)
		}
		if sub1.Branch != "main" {
			t.Errorf("Expected submodule branch 'main', got '%s'", sub1.Branch)
		}
	}

	// Check second submodule
	if len(submodules) > 1 {
		sub2 := submodules[1]
		if sub2.Name != "vendor/lib2" {
			t.Errorf("Expected submodule name 'vendor/lib2', got '%s'", sub2.Name)
		}
		if sub2.Path != "vendor/lib2" {
			t.Errorf("Expected submodule path 'vendor/lib2', got '%s'", sub2.Path)
		}
		if sub2.URL != "https://github.com/example/lib2.git" {
			t.Errorf("Expected submodule URL 'https://github.com/example/lib2.git', got '%s'", sub2.URL)
		}
	}
}

func TestSubmoduleDetectionNoGitmodules(t *testing.T) {
	// Create temporary directory without .gitmodules
	tempDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager
	logger := zap.NewNop()
	manager, err := NewManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test submodule detection with no .gitmodules
	submodules, err := manager.GetSubmodules(tempDir)
	if err != nil {
		t.Fatalf("Failed to get submodules: %v", err)
	}

	if len(submodules) != 0 {
		t.Errorf("Expected 0 submodules, got %d", len(submodules))
	}
}

func TestCommitHistoryNonGitRepo(t *testing.T) {
	// Create temporary directory that's not a git repo
	tempDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager
	logger := zap.NewNop()
	manager, err := NewManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test commit history on non-git repo
	commits, err := manager.GetCommitHistory(tempDir, "", 10)
	if err == nil {
		t.Error("Expected error for non-git repository")
	}

	if len(commits) != 0 {
		t.Errorf("Expected 0 commits for non-git repo, got %d", len(commits))
	}
}

func TestGitignoreCache(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .gitignore file
	gitignoreContent := `*.log`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create manager
	logger := zap.NewNop()
	manager, err := NewManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Load gitignore first time
	gi1 := manager.loadGitignore(tempDir)
	if gi1 == nil {
		t.Error("Expected gitignore to be loaded")
	}

	// Load gitignore second time (should use cache)
	gi2 := manager.loadGitignore(tempDir)
	if gi2 == nil {
		t.Error("Expected gitignore to be loaded from cache")
	}

	// Should be the same instance (cached)
	if gi1 != gi2 {
		t.Error("Expected same gitignore instance from cache")
	}

	// Test that cache contains the entry
	if len(manager.gitignores) != 1 {
		t.Errorf("Expected 1 cached gitignore, got %d", len(manager.gitignores))
	}
}

func TestIsIgnoredByGit(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .gitignore file
	gitignoreContent := `*.log
build/
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create manager
	logger := zap.NewNop()
	manager, err := NewManager(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test ignored files
	testCases := []struct {
		file     string
		ignored  bool
	}{
		{"main.go", false},
		{"app.log", true},
		{"src/debug.log", true},
		{"build/output.js", true},
		{"src/main.go", false},
	}

	for _, tc := range testCases {
		filePath := filepath.Join(tempDir, tc.file)
		ignored := manager.isIgnoredByGit(filePath, tempDir)
		if ignored != tc.ignored {
			t.Errorf("File %s: expected ignored=%v, got %v", tc.file, tc.ignored, ignored)
		}
	}
}
