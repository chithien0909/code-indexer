package ml

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/pkg/types"
)

// Engine represents the ML engine for code analysis
type Engine struct {
	config    *config.MLConfig
	logger    *zap.Logger
	models    map[string]interface{}
	cache     *EmbeddingCache
	mu        sync.RWMutex
	enabled   bool
}

// EmbeddingCache represents a cache for code embeddings
type EmbeddingCache struct {
	embeddings map[string]*types.CodeEmbedding
	maxSize    int
	mu         sync.RWMutex
}

// NewEngine creates a new ML engine instance
func NewEngine(cfg *config.MLConfig, logger *zap.Logger) (*Engine, error) {
	if !cfg.Enabled {
		logger.Info("ML engine disabled in configuration")
		return &Engine{
			config:  cfg,
			logger:  logger,
			enabled: false,
		}, nil
	}

	logger.Info("Initializing ML engine", zap.String("models_dir", cfg.ModelsDir))

	cache := &EmbeddingCache{
		embeddings: make(map[string]*types.CodeEmbedding),
		maxSize:    cfg.MaxEmbeddingCache,
	}

	engine := &Engine{
		config:  cfg,
		logger:  logger,
		models:  make(map[string]interface{}),
		cache:   cache,
		enabled: true,
	}

	// Initialize models
	if err := engine.initializeModels(); err != nil {
		return nil, fmt.Errorf("failed to initialize ML models: %w", err)
	}

	logger.Info("ML engine initialized successfully")
	return engine, nil
}

// IsEnabled returns whether the ML engine is enabled
func (e *Engine) IsEnabled() bool {
	return e.enabled
}

// initializeModels initializes the Spago models
func (e *Engine) initializeModels() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// TODO: Load actual Spago models from disk
	// For now, we'll create placeholder models
	e.models["embedding"] = &MockEmbeddingModel{}
	e.models["classification"] = &MockClassificationModel{}
	e.models["quality"] = &MockQualityModel{}
	e.models["similarity"] = &MockSimilarityModel{}

	e.logger.Info("ML models initialized", zap.Int("model_count", len(e.models)))
	return nil
}

// GenerateEmbedding generates a vector embedding for code
func (e *Engine) GenerateEmbedding(ctx context.Context, code string, fileID string) (*types.CodeEmbedding, error) {
	if !e.enabled {
		return nil, fmt.Errorf("ML engine is disabled")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", fileID, hashString(code))
	if cached := e.cache.Get(cacheKey); cached != nil {
		e.logger.Debug("Using cached embedding", zap.String("cache_key", cacheKey))
		return cached, nil
	}

	e.mu.RLock()
	model, exists := e.models["embedding"]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("embedding model not found")
	}

	// Generate embedding using Spago model
	embeddingModel := model.(*MockEmbeddingModel)
	vector, err := embeddingModel.Encode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	embedding := &types.CodeEmbedding{
		ID:         cacheKey,
		FileID:     fileID,
		Vector:     vector,
		Dimensions: len(vector),
		Model:      e.config.EmbeddingModel,
	}

	// Cache the embedding
	if e.config.CacheEmbeddings {
		e.cache.Set(cacheKey, embedding)
	}

	return embedding, nil
}

// AnalyzeSimilarity analyzes similarity between code snippets
func (e *Engine) AnalyzeSimilarity(ctx context.Context, code1, code2 string) (*types.SimilarityResult, error) {
	if !e.enabled {
		return nil, fmt.Errorf("ML engine is disabled")
	}

	// Generate embeddings for both code snippets
	emb1, err := e.GenerateEmbedding(ctx, code1, "temp1")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for code1: %w", err)
	}

	emb2, err := e.GenerateEmbedding(ctx, code2, "temp2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for code2: %w", err)
	}

	// Calculate cosine similarity
	similarity := cosineSimilarity(emb1.Vector, emb2.Vector)

	result := &types.SimilarityResult{
		SourceID:      "temp1",
		TargetID:      "temp2",
		Score:         similarity,
		Type:          "code_snippet",
		SourceSnippet: truncateString(code1, 200),
		TargetSnippet: truncateString(code2, 200),
	}

	if similarity > e.config.SimilarityThreshold {
		result.Explanation = "High similarity detected - potential code duplication"
	}

	return result, nil
}

// PredictQuality predicts code quality metrics
func (e *Engine) PredictQuality(ctx context.Context, file *types.CodeFile) (*types.QualityMetrics, error) {
	if !e.enabled {
		return nil, fmt.Errorf("ML engine is disabled")
	}

	e.mu.RLock()
	model, exists := e.models["quality"]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("quality model not found")
	}

	qualityModel := model.(*MockQualityModel)
	metrics, err := qualityModel.Predict(file)
	if err != nil {
		return nil, fmt.Errorf("failed to predict quality: %w", err)
	}

	return metrics, nil
}

// ClassifyIntent classifies the intent of code
func (e *Engine) ClassifyIntent(ctx context.Context, code string) (*types.IntentClassification, error) {
	if !e.enabled {
		return nil, fmt.Errorf("ML engine is disabled")
	}

	e.mu.RLock()
	model, exists := e.models["classification"]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("classification model not found")
	}

	classificationModel := model.(*MockClassificationModel)
	result, err := classificationModel.Classify(code)
	if err != nil {
		return nil, fmt.Errorf("failed to classify intent: %w", err)
	}

	return result, nil
}

// Close gracefully shuts down the ML engine
func (e *Engine) Close() error {
	if !e.enabled {
		return nil
	}

	e.logger.Info("Shutting down ML engine")
	
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear models
	e.models = nil
	
	// Clear cache
	e.cache.Clear()

	return nil
}

// EmbeddingCache methods

// Get retrieves an embedding from cache
func (c *EmbeddingCache) Get(key string) *types.CodeEmbedding {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.embeddings[key]
}

// Set stores an embedding in cache
func (c *EmbeddingCache) Set(key string, embedding *types.CodeEmbedding) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple LRU eviction if cache is full
	if len(c.embeddings) >= c.maxSize {
		// Remove first item (simple eviction strategy)
		for k := range c.embeddings {
			delete(c.embeddings, k)
			break
		}
	}

	c.embeddings[key] = embedding
}

// Clear clears the cache
func (c *EmbeddingCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.embeddings = make(map[string]*types.CodeEmbedding)
}

// Size returns the current cache size
func (c *EmbeddingCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.embeddings)
}
