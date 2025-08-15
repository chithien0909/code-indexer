package search

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/pkg/types"
)

// Engine provides search functionality using Bleve
type Engine struct {
	index  bleve.Index
	logger *zap.Logger
}

// Document represents a searchable document in the index
type Document struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // "file", "function", "class", "variable", "comment", "chunk"
	RepositoryID string                 `json:"repository_id"`
	Repository   string                 `json:"repository"`
	FilePath     string                 `json:"file_path"`
	Language     string                 `json:"language"`
	Name         string                 `json:"name,omitempty"`
	Content      string                 `json:"content"`
	StartLine    int                    `json:"start_line"`
	EndLine      int                    `json:"end_line"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IndexedAt    time.Time              `json:"indexed_at"`
}

// NewEngine creates a new search engine
func NewEngine(indexDir string, logger *zap.Logger) (*Engine, error) {
	// Create index mapping
	indexMapping := createIndexMapping()

	// Open or create the index
	index, err := bleve.Open(indexDir)
	if err != nil {
		// If index doesn't exist or has issues, create a new one
		logger.Info("Index not found or corrupted, creating new index", zap.String("path", indexDir), zap.Error(err))
		index, err = bleve.New(indexDir, indexMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create search index: %w", err)
		}
		logger.Info("Created new search index", zap.String("path", indexDir))
	} else {
		logger.Info("Opened existing search index", zap.String("path", indexDir))
	}

	return &Engine{
		index:  index,
		logger: logger,
	}, nil
}

// createIndexMapping creates the Bleve index mapping
func createIndexMapping() mapping.IndexMapping {
	// Create a mapping
	indexMapping := bleve.NewIndexMapping()

	// Create document mapping
	docMapping := bleve.NewDocumentMapping()

	// Text fields with analysis
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = true
	textFieldMapping.Index = true
	textFieldMapping.IncludeTermVectors = true

	// Keyword fields (exact match)
	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Store = true
	keywordFieldMapping.Index = true

	// Numeric fields
	numericFieldMapping := bleve.NewNumericFieldMapping()
	numericFieldMapping.Store = true
	numericFieldMapping.Index = true

	// Date fields
	dateFieldMapping := bleve.NewDateTimeFieldMapping()
	dateFieldMapping.Store = true
	dateFieldMapping.Index = true

	// Map fields
	docMapping.AddFieldMappingsAt("type", keywordFieldMapping)
	docMapping.AddFieldMappingsAt("repository_id", keywordFieldMapping)
	docMapping.AddFieldMappingsAt("repository", keywordFieldMapping)
	docMapping.AddFieldMappingsAt("file_path", textFieldMapping)
	docMapping.AddFieldMappingsAt("language", keywordFieldMapping)
	docMapping.AddFieldMappingsAt("name", textFieldMapping)
	docMapping.AddFieldMappingsAt("content", textFieldMapping)
	docMapping.AddFieldMappingsAt("start_line", numericFieldMapping)
	docMapping.AddFieldMappingsAt("end_line", numericFieldMapping)
	docMapping.AddFieldMappingsAt("indexed_at", dateFieldMapping)

	// Set default mapping
	indexMapping.DefaultMapping = docMapping

	return indexMapping
}

// IndexFile indexes a code file and all its components
func (e *Engine) IndexFile(ctx context.Context, file *types.CodeFile, repo *types.Repository) error {
	batch := e.index.NewBatch()

	// Index the file itself
	fileDoc := Document{
		ID:           fmt.Sprintf("file:%s:%s", repo.ID, file.RelativePath),
		Type:         "file",
		RepositoryID: repo.ID,
		Repository:   repo.Name,
		FilePath:     file.RelativePath,
		Language:     file.Language,
		Name:         filepath.Base(file.Path),
		Content:      file.Content,
		StartLine:    1,
		EndLine:      file.Lines,
		IndexedAt:    time.Now(),
	}
	batch.Index(fileDoc.ID, fileDoc)

	// Index functions
	for _, function := range file.Functions {
		funcDoc := Document{
			ID:           fmt.Sprintf("function:%s:%s:%s:%d", repo.ID, file.RelativePath, function.Name, function.StartLine),
			Type:         "function",
			RepositoryID: repo.ID,
			Repository:   repo.Name,
			FilePath:     file.RelativePath,
			Language:     file.Language,
			Name:         function.Name,
			Content:      function.Signature,
			StartLine:    function.StartLine,
			EndLine:      function.EndLine,
			Metadata: map[string]interface{}{
				"parameters":   function.Parameters,
				"return_type":  function.ReturnType,
				"visibility":   function.Visibility,
				"is_method":    function.IsMethod,
				"class_name":   function.ClassName,
				"doc_string":   function.DocString,
				"annotations":  function.Annotations,
			},
			IndexedAt: time.Now(),
		}
		batch.Index(funcDoc.ID, funcDoc)
	}

	// Index classes
	for _, class := range file.Classes {
		classDoc := Document{
			ID:           fmt.Sprintf("class:%s:%s:%s:%d", repo.ID, file.RelativePath, class.Name, class.StartLine),
			Type:         "class",
			RepositoryID: repo.ID,
			Repository:   repo.Name,
			FilePath:     file.RelativePath,
			Language:     file.Language,
			Name:         class.Name,
			Content:      class.Name,
			StartLine:    class.StartLine,
			EndLine:      class.EndLine,
			Metadata: map[string]interface{}{
				"visibility":   class.Visibility,
				"super_class":  class.SuperClass,
				"interfaces":   class.Interfaces,
				"doc_string":   class.DocString,
				"annotations":  class.Annotations,
			},
			IndexedAt: time.Now(),
		}
		batch.Index(classDoc.ID, classDoc)
	}

	// Index variables
	for _, variable := range file.Variables {
		varDoc := Document{
			ID:           fmt.Sprintf("variable:%s:%s:%s:%d", repo.ID, file.RelativePath, variable.Name, variable.StartLine),
			Type:         "variable",
			RepositoryID: repo.ID,
			Repository:   repo.Name,
			FilePath:     file.RelativePath,
			Language:     file.Language,
			Name:         variable.Name,
			Content:      fmt.Sprintf("%s %s", variable.Name, variable.Type),
			StartLine:    variable.StartLine,
			EndLine:      variable.EndLine,
			Metadata: map[string]interface{}{
				"type":        variable.Type,
				"value":       variable.Value,
				"visibility":  variable.Visibility,
				"is_constant": variable.IsConstant,
				"is_global":   variable.IsGlobal,
				"scope":       variable.Scope,
			},
			IndexedAt: time.Now(),
		}
		batch.Index(varDoc.ID, varDoc)
	}

	// Index comments
	for i, comment := range file.Comments {
		commentDoc := Document{
			ID:           fmt.Sprintf("comment:%s:%s:%d:%d", repo.ID, file.RelativePath, comment.StartLine, i),
			Type:         "comment",
			RepositoryID: repo.ID,
			Repository:   repo.Name,
			FilePath:     file.RelativePath,
			Language:     file.Language,
			Content:      comment.Text,
			StartLine:    comment.StartLine,
			EndLine:      comment.EndLine,
			Metadata: map[string]interface{}{
				"comment_type": comment.Type,
			},
			IndexedAt: time.Now(),
		}
		batch.Index(commentDoc.ID, commentDoc)
	}

	// Index chunks
	for _, chunk := range file.Chunks {
		chunkDoc := Document{
			ID:           fmt.Sprintf("chunk:%s:%s:%s:%d", repo.ID, file.RelativePath, chunk.ID, chunk.StartLine),
			Type:         "chunk",
			RepositoryID: repo.ID,
			Repository:   repo.Name,
			FilePath:     file.RelativePath,
			Language:     file.Language,
			Name:         chunk.Name,
			Content:      chunk.Content,
			StartLine:    chunk.StartLine,
			EndLine:      chunk.EndLine,
			Metadata: map[string]interface{}{
				"chunk_type":    chunk.Type,
				"chunk_id":      chunk.ID,
				"context":       chunk.Context,
				"dependencies":  chunk.Dependencies,
			},
			IndexedAt: time.Now(),
		}
		batch.Index(chunkDoc.ID, chunkDoc)
	}

	// Execute the batch
	return e.index.Batch(batch)
}

// Search performs a search query and returns results
func (e *Engine) Search(ctx context.Context, query types.SearchQuery) ([]types.SearchResult, error) {
	// Build the search query
	searchQuery := e.buildSearchQuery(query)

	// Create search request
	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = query.MaxResults
	if searchRequest.Size <= 0 {
		searchRequest.Size = 100
	}

	// Add highlighting
	searchRequest.Highlight = bleve.NewHighlight()
	searchRequest.Highlight.AddField("content")
	searchRequest.Highlight.AddField("name")

	// Include fields in results
	searchRequest.Fields = []string{"*"}

	// Execute search
	searchResult, err := e.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	results := make([]types.SearchResult, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		result, err := e.convertSearchHit(hit)
		if err != nil {
			e.logger.Warn("Failed to convert search hit", zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	e.logger.Info("Search completed",
		zap.String("query", query.Query),
		zap.String("type", query.Type),
		zap.Int("total_hits", int(searchResult.Total)),
		zap.Int("returned", len(results)))

	return results, nil
}

// buildSearchQuery builds a Bleve query from the search parameters
func (e *Engine) buildSearchQuery(searchQuery types.SearchQuery) query.Query {
	var queries []query.Query

	// Main content query
	if searchQuery.Query != "" {
		if searchQuery.Fuzzy {
			// Fuzzy search
			fuzzyQuery := bleve.NewFuzzyQuery(searchQuery.Query)
			queries = append(queries, fuzzyQuery)
		} else {
			// Regular text search across multiple fields
			contentMatchQuery := bleve.NewMatchQuery(searchQuery.Query)
			contentMatchQuery.SetField("content")

			nameMatchQuery := bleve.NewMatchQuery(searchQuery.Query)
			nameMatchQuery.SetField("name")

			pathMatchQuery := bleve.NewMatchQuery(searchQuery.Query)
			pathMatchQuery.SetField("file_path")

			contentQuery := bleve.NewDisjunctionQuery(
				contentMatchQuery,
				nameMatchQuery,
				pathMatchQuery,
			)
			queries = append(queries, contentQuery)
		}
	}

	// Type filter
	if searchQuery.Type != "" {
		typeQuery := bleve.NewTermQuery(searchQuery.Type)
		typeQuery.SetField("type")
		queries = append(queries, typeQuery)
	}

	// Language filter
	if searchQuery.Language != "" {
		langQuery := bleve.NewTermQuery(searchQuery.Language)
		langQuery.SetField("language")
		queries = append(queries, langQuery)
	}

	// Repository filter
	if searchQuery.Repository != "" {
		repoQuery := bleve.NewTermQuery(searchQuery.Repository)
		repoQuery.SetField("repository")
		queries = append(queries, repoQuery)
	}

	// File path filter
	if searchQuery.FilePath != "" {
		pathQuery := bleve.NewWildcardQuery("*" + searchQuery.FilePath + "*")
		pathQuery.SetField("file_path")
		queries = append(queries, pathQuery)
	}

	// Combine all queries
	if len(queries) == 0 {
		return bleve.NewMatchAllQuery()
	} else if len(queries) == 1 {
		return queries[0]
	} else {
		return bleve.NewConjunctionQuery(queries...)
	}
}

// convertSearchHit converts a Bleve search hit to our result format
func (e *Engine) convertSearchHit(hit *search.DocumentMatch) (types.SearchResult, error) {
	result := types.SearchResult{
		ID:    hit.ID,
		Score: hit.Score,
	}

	// Extract fields from the hit
	if repoID, ok := hit.Fields["repository_id"].(string); ok {
		result.RepositoryID = repoID
	}
	if repo, ok := hit.Fields["repository"].(string); ok {
		result.Repository = repo
	}
	if filePath, ok := hit.Fields["file_path"].(string); ok {
		result.FilePath = filePath
	}
	if language, ok := hit.Fields["language"].(string); ok {
		result.Language = language
	}
	if docType, ok := hit.Fields["type"].(string); ok {
		result.Type = docType
	}
	if name, ok := hit.Fields["name"].(string); ok {
		result.Name = name
	}
	if content, ok := hit.Fields["content"].(string); ok {
		result.Content = content
	}
	if startLine, ok := hit.Fields["start_line"].(float64); ok {
		result.StartLine = int(startLine)
	}
	if endLine, ok := hit.Fields["end_line"].(float64); ok {
		result.EndLine = int(endLine)
	}

	// Add highlights
	if len(hit.Fragments) > 0 {
		result.Highlights = make(map[string]string)
		for field, fragments := range hit.Fragments {
			if len(fragments) > 0 {
				result.Highlights[field] = strings.Join(fragments, "...")
			}
		}
	}

	// Create snippet from content or highlights
	if result.Highlights != nil && result.Highlights["content"] != "" {
		result.Snippet = result.Highlights["content"]
	} else if len(result.Content) > 200 {
		result.Snippet = result.Content[:200] + "..."
	} else {
		result.Snippet = result.Content
	}

	return result, nil
}

// GetFileMetadata retrieves metadata for a specific file
func (e *Engine) GetFileMetadata(ctx context.Context, filePath, repository string) (*types.CodeFile, error) {
	// Build query to find the file
	var searchQuery query.Query
	if repository != "" {
		fileQuery := bleve.NewTermQuery("file")
		fileQuery.SetField("type")

		repoQuery := bleve.NewTermQuery(repository)
		repoQuery.SetField("repository")

		pathQuery := bleve.NewWildcardQuery("*"+filePath+"*")
		pathQuery.SetField("file_path")

		searchQuery = bleve.NewConjunctionQuery(fileQuery, repoQuery, pathQuery)
	} else {
		fileQuery := bleve.NewTermQuery("file")
		fileQuery.SetField("type")

		pathQuery := bleve.NewWildcardQuery("*"+filePath+"*")
		pathQuery.SetField("file_path")

		searchQuery = bleve.NewConjunctionQuery(fileQuery, pathQuery)
	}

	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = 1
	searchRequest.Fields = []string{"*"}

	searchResult, err := e.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for file: %w", err)
	}

	if len(searchResult.Hits) == 0 {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	hit := searchResult.Hits[0]

	// Extract basic file info
	file := &types.CodeFile{}
	if path, ok := hit.Fields["file_path"].(string); ok {
		file.RelativePath = path
	}
	if lang, ok := hit.Fields["language"].(string); ok {
		file.Language = lang
	}
	if content, ok := hit.Fields["content"].(string); ok {
		file.Content = content
	}
	if startLine, ok := hit.Fields["start_line"].(float64); ok {
		file.Lines = int(startLine)
	}

	// Get repository info
	var repoID string
	if id, ok := hit.Fields["repository_id"].(string); ok {
		repoID = id
		file.RepositoryID = id
	}

	// Now get all related components (functions, classes, variables, comments)
	if repoID != "" {
		if err := e.enrichFileMetadata(ctx, file, repoID); err != nil {
			e.logger.Warn("Failed to enrich file metadata", zap.Error(err))
		}
	}

	return file, nil
}

// enrichFileMetadata adds functions, classes, variables, and comments to a file
func (e *Engine) enrichFileMetadata(ctx context.Context, file *types.CodeFile, repoID string) error {
	// Query for all components of this file
	repoQuery := bleve.NewTermQuery(repoID)
	repoQuery.SetField("repository_id")

	pathQuery := bleve.NewWildcardQuery("*"+file.RelativePath+"*")
	pathQuery.SetField("file_path")

	funcQuery := bleve.NewTermQuery("function")
	funcQuery.SetField("type")

	classQuery := bleve.NewTermQuery("class")
	classQuery.SetField("type")

	varQuery := bleve.NewTermQuery("variable")
	varQuery.SetField("type")

	commentQuery := bleve.NewTermQuery("comment")
	commentQuery.SetField("type")

	typeQuery := bleve.NewDisjunctionQuery(funcQuery, classQuery, varQuery, commentQuery)

	searchQuery := bleve.NewConjunctionQuery(repoQuery, pathQuery, typeQuery)

	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = 1000 // Large number to get all components
	searchRequest.Fields = []string{"*"}

	searchResult, err := e.index.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("failed to search for file components: %w", err)
	}

	// Process each component
	for _, hit := range searchResult.Hits {
		docType, _ := hit.Fields["type"].(string)

		switch docType {
		case "function":
			function := e.extractFunction(hit)
			file.Functions = append(file.Functions, function)
		case "class":
			class := e.extractClass(hit)
			file.Classes = append(file.Classes, class)
		case "variable":
			variable := e.extractVariable(hit)
			file.Variables = append(file.Variables, variable)
		case "comment":
			comment := e.extractComment(hit)
			file.Comments = append(file.Comments, comment)
		}
	}

	return nil
}

// extractFunction extracts function data from a search hit
func (e *Engine) extractFunction(hit *search.DocumentMatch) types.Function {
	function := types.Function{}

	if name, ok := hit.Fields["name"].(string); ok {
		function.Name = name
	}
	if content, ok := hit.Fields["content"].(string); ok {
		function.Signature = content
	}
	if startLine, ok := hit.Fields["start_line"].(float64); ok {
		function.StartLine = int(startLine)
	}
	if endLine, ok := hit.Fields["end_line"].(float64); ok {
		function.EndLine = int(endLine)
	}

	// Extract metadata if available
	if metadata, ok := hit.Fields["metadata"].(map[string]interface{}); ok {
		if params, ok := metadata["parameters"].([]interface{}); ok {
			for _, p := range params {
				if param, ok := p.(string); ok {
					function.Parameters = append(function.Parameters, param)
				}
			}
		}
		if returnType, ok := metadata["return_type"].(string); ok {
			function.ReturnType = returnType
		}
		if visibility, ok := metadata["visibility"].(string); ok {
			function.Visibility = visibility
		}
		if isMethod, ok := metadata["is_method"].(bool); ok {
			function.IsMethod = isMethod
		}
		if className, ok := metadata["class_name"].(string); ok {
			function.ClassName = className
		}
		if docString, ok := metadata["doc_string"].(string); ok {
			function.DocString = docString
		}
	}

	return function
}

// extractClass extracts class data from a search hit
func (e *Engine) extractClass(hit *search.DocumentMatch) types.Class {
	class := types.Class{}

	if name, ok := hit.Fields["name"].(string); ok {
		class.Name = name
	}
	if startLine, ok := hit.Fields["start_line"].(float64); ok {
		class.StartLine = int(startLine)
	}
	if endLine, ok := hit.Fields["end_line"].(float64); ok {
		class.EndLine = int(endLine)
	}

	// Extract metadata if available
	if metadata, ok := hit.Fields["metadata"].(map[string]interface{}); ok {
		if visibility, ok := metadata["visibility"].(string); ok {
			class.Visibility = visibility
		}
		if superClass, ok := metadata["super_class"].(string); ok {
			class.SuperClass = superClass
		}
		if docString, ok := metadata["doc_string"].(string); ok {
			class.DocString = docString
		}
	}

	return class
}

// extractVariable extracts variable data from a search hit
func (e *Engine) extractVariable(hit *search.DocumentMatch) types.Variable {
	variable := types.Variable{}

	if name, ok := hit.Fields["name"].(string); ok {
		variable.Name = name
	}
	if startLine, ok := hit.Fields["start_line"].(float64); ok {
		variable.StartLine = int(startLine)
	}
	if endLine, ok := hit.Fields["end_line"].(float64); ok {
		variable.EndLine = int(endLine)
	}

	// Extract metadata if available
	if metadata, ok := hit.Fields["metadata"].(map[string]interface{}); ok {
		if varType, ok := metadata["type"].(string); ok {
			variable.Type = varType
		}
		if value, ok := metadata["value"].(string); ok {
			variable.Value = value
		}
		if visibility, ok := metadata["visibility"].(string); ok {
			variable.Visibility = visibility
		}
		if isConstant, ok := metadata["is_constant"].(bool); ok {
			variable.IsConstant = isConstant
		}
		if isGlobal, ok := metadata["is_global"].(bool); ok {
			variable.IsGlobal = isGlobal
		}
		if scope, ok := metadata["scope"].(string); ok {
			variable.Scope = scope
		}
	}

	return variable
}

// extractComment extracts comment data from a search hit
func (e *Engine) extractComment(hit *search.DocumentMatch) types.Comment {
	comment := types.Comment{}

	if content, ok := hit.Fields["content"].(string); ok {
		comment.Text = content
	}
	if startLine, ok := hit.Fields["start_line"].(float64); ok {
		comment.StartLine = int(startLine)
	}
	if endLine, ok := hit.Fields["end_line"].(float64); ok {
		comment.EndLine = int(endLine)
	}

	// Extract metadata if available
	if metadata, ok := hit.Fields["metadata"].(map[string]interface{}); ok {
		if commentType, ok := metadata["comment_type"].(string); ok {
			comment.Type = commentType
		}
	}

	return comment
}

// ListRepositories returns all indexed repositories
func (e *Engine) ListRepositories(ctx context.Context) ([]types.Repository, error) {
	// Query for all file documents to get repository info
	fileQuery := bleve.NewTermQuery("file")
	fileQuery.SetField("type")

	searchRequest := bleve.NewSearchRequest(fileQuery)
	searchRequest.Size = 10000 // Large number to get all files
	searchRequest.Fields = []string{"repository_id", "repository", "language"}

	searchResult, err := e.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for repositories: %w", err)
	}

	// Aggregate repository information
	repoMap := make(map[string]*types.Repository)
	languageStats := make(map[string]map[string]int)

	for _, hit := range searchResult.Hits {
		repoID, _ := hit.Fields["repository_id"].(string)
		repoName, _ := hit.Fields["repository"].(string)
		language, _ := hit.Fields["language"].(string)

		if repoID == "" {
			continue
		}

		// Initialize repository if not exists
		if _, exists := repoMap[repoID]; !exists {
			repoMap[repoID] = &types.Repository{
				ID:   repoID,
				Name: repoName,
			}
			languageStats[repoID] = make(map[string]int)
		}

		// Update file count and language stats
		repoMap[repoID].FileCount++
		if language != "" {
			languageStats[repoID][language]++
		}
	}

	// Convert to slice and add language information
	repositories := make([]types.Repository, 0, len(repoMap))
	for repoID, repo := range repoMap {
		// Extract unique languages
		languages := make([]string, 0, len(languageStats[repoID]))
		for lang := range languageStats[repoID] {
			languages = append(languages, lang)
		}
		repo.Languages = languages

		repositories = append(repositories, *repo)
	}

	return repositories, nil
}

// GetIndexStats returns indexing statistics
func (e *Engine) GetIndexStats(ctx context.Context) (*types.IndexStats, error) {
	stats := &types.IndexStats{
		LanguageStats:   make(map[string]int),
		RepositoryStats: make(map[string]types.Repository),
		LastIndexed:     time.Now(),
	}

	// Get document count by type
	types := []string{"file", "function", "class", "variable", "comment"}

	for _, docType := range types {
		typeQuery := bleve.NewTermQuery(docType)
		typeQuery.SetField("type")
		searchRequest := bleve.NewSearchRequest(typeQuery)
		searchRequest.Size = 0 // We only want the count

		searchResult, err := e.index.Search(searchRequest)
		if err != nil {
			e.logger.Warn("Failed to get stats for type", zap.String("type", docType), zap.Error(err))
			continue
		}

		count := int(searchResult.Total)
		switch docType {
		case "file":
			stats.TotalFiles = count
		case "function":
			stats.TotalFunctions = count
		case "class":
			stats.TotalClasses = count
		case "variable":
			stats.TotalVariables = count
		}
	}

	// Get repositories
	repositories, err := e.ListRepositories(ctx)
	if err != nil {
		e.logger.Warn("Failed to get repositories for stats", zap.Error(err))
	} else {
		stats.TotalRepositories = len(repositories)
		for _, repo := range repositories {
			stats.RepositoryStats[repo.Name] = repo
			for _, lang := range repo.Languages {
				stats.LanguageStats[lang] += repo.FileCount
			}
		}
	}

	return stats, nil
}

// DeleteRepository removes all documents for a repository from the index
func (e *Engine) DeleteRepository(ctx context.Context, repositoryID string) error {
	// Query for all documents of this repository
	repoQuery := bleve.NewTermQuery(repositoryID)
	repoQuery.SetField("repository_id")

	searchRequest := bleve.NewSearchRequest(repoQuery)
	searchRequest.Size = 10000 // Large number to get all documents
	searchRequest.Fields = []string{"_id"}

	searchResult, err := e.index.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("failed to search for repository documents: %w", err)
	}

	// Delete documents in batches
	batch := e.index.NewBatch()
	for _, hit := range searchResult.Hits {
		batch.Delete(hit.ID)
	}

	return e.index.Batch(batch)
}

// Close closes the search engine
func (e *Engine) Close() error {
	return e.index.Close()
}
