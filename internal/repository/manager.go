package repository

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitignore "github.com/sabhiram/go-gitignore"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// Manager handles Git repository operations and file discovery
type Manager struct {
	repoDir     string
	logger      *zap.Logger
	gitignores  map[string]*gitignore.GitIgnore // Cache gitignore patterns per repository
}

// NewManager creates a new repository manager
func NewManager(repoDir string, logger *zap.Logger) (*Manager, error) {
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository directory: %w", err)
	}

	return &Manager{
		repoDir:    repoDir,
		logger:     logger,
		gitignores: make(map[string]*gitignore.GitIgnore),
	}, nil
}

// PrepareRepository prepares a repository for indexing (clone if URL, validate if local path)
func (m *Manager) PrepareRepository(ctx context.Context, path, name string) (*types.Repository, error) {
	var repoPath string
	var repoURL string
	var isRemote bool

	// Check if path is a URL
	if u, err := url.Parse(path); err == nil && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "git") {
		isRemote = true
		repoURL = path
		
		// Generate a directory name for the cloned repo
		repoName := name
		if repoName == "" {
			repoName = m.generateRepoName(path)
		}
		repoPath = filepath.Join(m.repoDir, repoName)
		
		// Clone or update the repository
		if err := m.cloneOrUpdateRepo(ctx, repoURL, repoPath); err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
	} else {
		// Local path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("invalid local path: %w", err)
		}
		
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("local repository path does not exist: %s", absPath)
		}
		
		repoPath = absPath
	}

	// Get repository information
	repo, err := m.getRepositoryInfo(repoPath, repoURL, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	m.logger.Info("Repository prepared", 
		zap.String("name", repo.Name),
		zap.String("path", repo.Path),
		zap.Bool("is_remote", isRemote))

	return repo, nil
}

// cloneOrUpdateRepo clones a repository or updates it if it already exists
func (m *Manager) cloneOrUpdateRepo(ctx context.Context, repoURL, repoPath string) error {
	// Check if repository already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Repository exists, try to update it
		m.logger.Info("Updating existing repository", zap.String("path", repoPath))
		
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			return fmt.Errorf("failed to open existing repository: %w", err)
		}
		
		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %w", err)
		}
		
		err = worktree.Pull(&git.PullOptions{})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			m.logger.Warn("Failed to pull updates, continuing with existing version", zap.Error(err))
		}
		
		return nil
	}

	// Clone the repository
	m.logger.Info("Cloning repository", zap.String("url", repoURL), zap.String("path", repoPath))
	
	_, err := git.PlainCloneContext(ctx, repoPath, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// getRepositoryInfo extracts information about a Git repository
func (m *Manager) getRepositoryInfo(repoPath, repoURL, customName string) (*types.Repository, error) {
	repo := &types.Repository{
		Path:      repoPath,
		URL:       repoURL,
		IndexedAt: time.Now(),
	}

	// Generate repository ID
	hasher := sha256.New()
	hasher.Write([]byte(repoPath))
	repo.ID = fmt.Sprintf("%x", hasher.Sum(nil))[:16]

	// Set repository name
	if customName != "" {
		repo.Name = customName
	} else if repoURL != "" {
		repo.Name = m.generateRepoName(repoURL)
	} else {
		repo.Name = filepath.Base(repoPath)
	}

	// Try to get Git information
	if gitRepo, err := git.PlainOpen(repoPath); err == nil {
		// Get current branch
		if head, err := gitRepo.Head(); err == nil {
			repo.Branch = head.Name().Short()
			repo.LastIndexedHash = head.Hash().String()
		}

		// Get latest commit
		if commits, err := gitRepo.Log(&git.LogOptions{}); err == nil {
			if commit, err := commits.Next(); err == nil {
				repo.LastCommit = commit.Hash.String()[:8]
			}
		}

		// Get submodules
		if submodules, err := m.GetSubmodules(repoPath); err == nil {
			repo.Submodules = submodules
		}

		// Set indexing mode
		repo.IndexingMode = "full"
	}

	return repo, nil
}

// generateRepoName generates a repository name from a URL
func (m *Manager) generateRepoName(repoURL string) string {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "unknown-repo"
	}

	path := strings.TrimSuffix(u.Path, ".git")
	parts := strings.Split(path, "/")
	
	if len(parts) >= 2 {
		return fmt.Sprintf("%s-%s", parts[len(parts)-2], parts[len(parts)-1])
	} else if len(parts) == 1 && parts[0] != "" {
		return parts[0]
	}

	return "unknown-repo"
}

// WalkFiles walks through all files in a repository and calls the callback for each file
func (m *Manager) WalkFiles(ctx context.Context, repoPath string, callback func(filePath string, info fs.FileInfo) error) error {
	return filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if d.IsDir() {
			// Check if directory should be ignored by gitignore
			if m.isIgnoredByGit(path, repoPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be ignored by gitignore
		if m.isIgnoredByGit(path, repoPath) {
			return nil // Skip this file
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			m.logger.Warn("Failed to get file info", zap.String("path", path), zap.Error(err))
			return nil // Continue walking
		}

		// Call the callback
		return callback(path, info)
	})
}

// GetFileContent reads the content of a file
func (m *Manager) GetFileContent(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// GetRelativePath returns the relative path of a file within a repository
func (m *Manager) GetRelativePath(filePath, repoPath string) (string, error) {
	return filepath.Rel(repoPath, filePath)
}

// GetFileLanguage determines the programming language of a file based on its extension
func (m *Manager) GetFileLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	languageMap := map[string]string{
		".go":     "go",
		".py":     "python",
		".js":     "javascript",
		".ts":     "typescript",
		".java":   "java",
		".cpp":    "cpp",
		".c":      "c",
		".h":      "c",
		".hpp":    "cpp",
		".rs":     "rust",
		".rb":     "ruby",
		".php":    "php",
		".cs":     "csharp",
		".kt":     "kotlin",
		".swift":  "swift",
		".scala":  "scala",
		".clj":    "clojure",
		".hs":     "haskell",
		".ml":     "ocaml",
		".sh":     "shell",
		".bash":   "shell",
		".zsh":    "shell",
		".fish":   "shell",
		".ps1":    "powershell",
		".sql":    "sql",
		".r":      "r",
		".m":      "matlab",
		".dart":   "dart",
		".lua":    "lua",
		".perl":   "perl",
		".pl":     "perl",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}

	return "unknown"
}

// ValidateRepository checks if a path contains a valid repository
func (m *Manager) ValidateRepository(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("repository path does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("repository path is not a directory")
	}

	// Check if it's a Git repository (optional, but helpful)
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil // Valid Git repository
	}

	// Even if it's not a Git repository, we can still index it
	return nil
}

// loadGitignore loads and caches gitignore patterns for a repository
func (m *Manager) loadGitignore(repoPath string) *gitignore.GitIgnore {
	if gi, exists := m.gitignores[repoPath]; exists {
		return gi
	}

	gitignorePath := filepath.Join(repoPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// No .gitignore file, create empty gitignore
		m.gitignores[repoPath] = gitignore.CompileIgnoreLines()
		return m.gitignores[repoPath]
	}

	gi, err := gitignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		m.logger.Warn("Failed to load .gitignore file", zap.String("path", gitignorePath), zap.Error(err))
		gi = gitignore.CompileIgnoreLines()
	}

	m.gitignores[repoPath] = gi
	return gi
}

// isIgnoredByGit checks if a file should be ignored according to .gitignore rules
func (m *Manager) isIgnoredByGit(filePath, repoPath string) bool {
	gi := m.loadGitignore(repoPath)

	// Get relative path from repository root
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return false
	}

	return gi.MatchesPath(relPath)
}

// GetSubmodules returns information about Git submodules in a repository
func (m *Manager) GetSubmodules(repoPath string) ([]types.Submodule, error) {
	var submodules []types.Submodule

	// Check for .gitmodules file first (don't require git repo for testing)
	gitmodulesPath := filepath.Join(repoPath, ".gitmodules")
	if _, err := os.Stat(gitmodulesPath); os.IsNotExist(err) {
		return submodules, nil // No submodules
	}

	// Read .gitmodules file
	content, err := os.ReadFile(gitmodulesPath)
	if err != nil {
		return submodules, fmt.Errorf("failed to read .gitmodules: %w", err)
	}

	m.logger.Debug("Reading .gitmodules content", zap.String("content", string(content)))

	// Parse .gitmodules content (simplified parsing)
	lines := strings.Split(string(content), "\n")
	var currentSubmodule *types.Submodule

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[submodule ") {
			if currentSubmodule != nil {
				submodules = append(submodules, *currentSubmodule)
			}
			// Extract name from [submodule "name"]
			start := strings.Index(line, `"`)
			end := strings.LastIndex(line, `"`)
			if start != -1 && end != -1 && start < end {
				name := line[start+1 : end]
				currentSubmodule = &types.Submodule{Name: name}
			}
		} else if currentSubmodule != nil && strings.Contains(line, " = ") {
			parts := strings.SplitN(line, " = ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				switch key {
				case "path":
					currentSubmodule.Path = value
				case "url":
					currentSubmodule.URL = value
				case "branch":
					currentSubmodule.Branch = value
				}
			}
		}
	}

	if currentSubmodule != nil {
		submodules = append(submodules, *currentSubmodule)
	}

	return submodules, nil
}

// GetCommitHistory returns recent commit history for incremental indexing
func (m *Manager) GetCommitHistory(repoPath string, fromCommit string, limit int) ([]types.CommitInfo, error) {
	var commits []types.CommitInfo

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return commits, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return commits, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get commit iterator
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return commits, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	count := 0
	foundStart := fromCommit == ""

	err = commitIter.ForEach(func(c *object.Commit) error {
		if !foundStart {
			if c.Hash.String() == fromCommit {
				foundStart = true
			}
			return nil
		}

		if count >= limit {
			return fmt.Errorf("limit reached") // Use error to break iteration
		}

		commitInfo := types.CommitInfo{
			Hash:    c.Hash.String(),
			Message: c.Message,
			Author:  c.Author.Name,
			Email:   c.Author.Email,
			Date:    c.Author.When,
		}

		// Get files changed in this commit
		if c.NumParents() > 0 {
			parent, err := c.Parent(0)
			if err == nil {
				parentTree, err := parent.Tree()
				if err == nil {
					currentTree, err := c.Tree()
					if err == nil {
						changes, err := parentTree.Diff(currentTree)
						if err == nil {
							for _, change := range changes {
								from, to := change.From, change.To
								if from.Name != "" {
									commitInfo.Files = append(commitInfo.Files, from.Name)
								}
								if to.Name != "" && from.Name != to.Name {
									commitInfo.Files = append(commitInfo.Files, to.Name)
								}
							}
						}
					}
				}
			}
		}

		commits = append(commits, commitInfo)
		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return commits, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}
