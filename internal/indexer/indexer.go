package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/chunking"
	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/parser"
	"github.com/my-mcp/code-indexer/internal/repository"
	"github.com/my-mcp/code-indexer/internal/search"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// Indexer handles the indexing of repositories and files
type Indexer struct {
	config     *config.Config
	repoMgr    *repository.Manager
	searcher   *search.Engine
	parser     *parser.Registry
	chunker    *chunking.Chunker
	logger     *zap.Logger
}

// New creates a new indexer instance
func New(cfg *config.Config, repoMgr *repository.Manager, searcher *search.Engine, logger *zap.Logger) (*Indexer, error) {
	// Initialize chunker with default config for now
	chunkingConfig := chunking.DefaultChunkingConfig()

	return &Indexer{
		config:   cfg,
		repoMgr:  repoMgr,
		searcher: searcher,
		parser:   parser.NewRegistry(),
		chunker:  chunking.NewChunker(chunkingConfig),
		logger:   logger,
	}, nil
}

// IndexRepository indexes a complete repository
func (i *Indexer) IndexRepository(ctx context.Context, path, name string) (*types.Repository, error) {
	i.logger.Info("Starting repository indexing", zap.String("path", path), zap.String("name", name))

	// Prepare the repository (clone if remote, validate if local)
	repo, err := i.repoMgr.PrepareRepository(ctx, path, name)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare repository: %w", err)
	}

	// Start indexing process
	startTime := time.Now()
	progress := &types.IndexingProgress{
		RepositoryID: repo.ID,
		Repository:   repo.Name,
		Status:       "starting",
		StartedAt:    startTime,
	}

	i.logger.Info("Repository prepared, starting file discovery", zap.String("repo_id", repo.ID))

	// Discover files to index
	var filesToIndex []string
	err = i.repoMgr.WalkFiles(ctx, repo.Path, func(filePath string, info fs.FileInfo) error {
		// Check if file should be indexed
		if i.shouldIndexFile(filePath, info) {
			filesToIndex = append(filesToIndex, filePath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover files: %w", err)
	}

	progress.TotalFiles = len(filesToIndex)
	progress.Status = "indexing"

	i.logger.Info("File discovery completed", 
		zap.String("repo_id", repo.ID),
		zap.Int("total_files", len(filesToIndex)))

	// Index each file
	var totalLines int
	languages := make(map[string]bool)

	for _, filePath := range filesToIndex {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		progress.FilesProcessed++
		progress.CurrentFile = filePath

		// Index the file
		lines, err := i.indexFile(ctx, filePath, repo)
		if err != nil {
			i.logger.Warn("Failed to index file", 
				zap.String("file", filePath), 
				zap.Error(err))
			continue
		}

		totalLines += lines
		
		// Track language
		language := i.repoMgr.GetFileLanguage(filePath)
		if language != "unknown" {
			languages[language] = true
		}

		// Log progress periodically
		if progress.FilesProcessed%100 == 0 {
			i.logger.Info("Indexing progress", 
				zap.String("repo_id", repo.ID),
				zap.Int("processed", progress.FilesProcessed),
				zap.Int("total", progress.TotalFiles))
		}
	}

	// Update repository statistics
	repo.FileCount = len(filesToIndex)
	repo.TotalLines = totalLines
	repo.Languages = make([]string, 0, len(languages))
	for lang := range languages {
		repo.Languages = append(repo.Languages, lang)
	}
	repo.IndexedAt = time.Now()

	// Complete indexing
	progress.Status = "completed"
	completedAt := time.Now()
	progress.CompletedAt = &completedAt
	progress.ElapsedSeconds = completedAt.Sub(startTime).Seconds()

	i.logger.Info("Repository indexing completed", 
		zap.String("repo_id", repo.ID),
		zap.String("repo_name", repo.Name),
		zap.Int("files_indexed", repo.FileCount),
		zap.Int("total_lines", repo.TotalLines),
		zap.Strings("languages", repo.Languages),
		zap.Duration("elapsed", completedAt.Sub(startTime)))

	return repo, nil
}

// indexFile indexes a single file
func (i *Indexer) indexFile(ctx context.Context, filePath string, repo *types.Repository) (int, error) {
	// Read file content
	content, err := i.repoMgr.GetFileContent(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file content: %w", err)
	}

	// Get relative path
	relativePath, err := i.repoMgr.GetRelativePath(filePath, repo.Path)
	if err != nil {
		return 0, fmt.Errorf("failed to get relative path: %w", err)
	}

	// Determine language
	language := i.repoMgr.GetFileLanguage(filePath)

	// Create file hash for change detection
	hasher := sha256.New()
	hasher.Write(content)
	fileHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Create code file structure
	codeFile := &types.CodeFile{
		ID:           fmt.Sprintf("%s:%s", repo.ID, relativePath),
		RepositoryID: repo.ID,
		Path:         filePath,
		RelativePath: relativePath,
		Language:     language,
		Extension:    filepath.Ext(filePath),
		Size:         int64(len(content)),
		Content:      string(content),
		Hash:         fileHash,
		IndexedAt:    time.Now(),
	}

	// Parse the file to extract metadata
	parsedFile, err := i.parser.ParseFile(string(content), filePath, language)
	if err != nil {
		i.logger.Warn("Failed to parse file", 
			zap.String("file", filePath), 
			zap.String("language", language),
			zap.Error(err))
		// Continue with basic file info even if parsing fails
	} else {
		// Copy parsed metadata
		codeFile.Lines = parsedFile.Lines
		codeFile.Functions = parsedFile.Functions
		codeFile.Classes = parsedFile.Classes
		codeFile.Variables = parsedFile.Variables
		codeFile.Imports = parsedFile.Imports
		codeFile.Comments = parsedFile.Comments
	}

	// If parsing failed, at least count lines
	if codeFile.Lines == 0 {
		codeFile.Lines = strings.Count(string(content), "\n") + 1
	}

	// Create semantic chunks for the file
	chunks := i.chunker.ChunkFile(codeFile)
	codeFile.Chunks = chunks

	// Index the file in the search engine
	if err := i.searcher.IndexFile(ctx, codeFile, repo); err != nil {
		return 0, fmt.Errorf("failed to index file in search engine: %w", err)
	}

	return codeFile.Lines, nil
}

// shouldIndexFile determines if a file should be indexed
func (i *Indexer) shouldIndexFile(filePath string, info fs.FileInfo) bool {
	// Skip directories
	if info.IsDir() {
		return false
	}

	// Check file size limit
	if info.Size() > i.config.Indexer.MaxFileSize {
		return false
	}

	// Check if file extension is supported
	if !i.config.IsFileSupported(filePath) {
		return false
	}

	// Check exclude patterns
	if i.config.ShouldExcludeFile(filePath) {
		return false
	}

	return true
}

// ReindexRepository removes and re-indexes a repository
func (i *Indexer) ReindexRepository(ctx context.Context, repositoryID string) error {
	i.logger.Info("Starting repository re-indexing", zap.String("repo_id", repositoryID))

	// Delete existing index data for this repository
	if err := i.searcher.DeleteRepository(ctx, repositoryID); err != nil {
		return fmt.Errorf("failed to delete existing repository data: %w", err)
	}

	// TODO: Re-index the repository
	// This would require storing repository paths/URLs in a persistent store
	// For now, return an error indicating this feature needs implementation
	return fmt.Errorf("re-indexing requires repository path information - not yet implemented")
}

// GetIndexingProgress returns the current indexing progress (if any)
// This is a placeholder for future implementation of async indexing with progress tracking
func (i *Indexer) GetIndexingProgress(repositoryID string) (*types.IndexingProgress, error) {
	// TODO: Implement progress tracking for async indexing
	return nil, fmt.Errorf("progress tracking not yet implemented")
}
