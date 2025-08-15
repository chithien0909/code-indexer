package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Indexer IndexerConfig `mapstructure:"indexer"`
	Search  SearchConfig  `mapstructure:"search"`
	Server  ServerConfig  `mapstructure:"server"`
	Logging LoggingConfig `mapstructure:"logging"`
	Models  ModelsConfig  `mapstructure:"models"`
}

// IndexerConfig represents indexer-specific configuration
type IndexerConfig struct {
	SupportedExtensions []string `mapstructure:"supported_extensions"`
	MaxFileSize         int64    `mapstructure:"max_file_size"`
	ExcludePatterns     []string `mapstructure:"exclude_patterns"`
	IndexDir            string   `mapstructure:"index_dir"`
	RepoDir             string   `mapstructure:"repo_dir"`
}

// SearchConfig represents search-specific configuration
type SearchConfig struct {
	MaxResults       int     `mapstructure:"max_results"`
	HighlightSnippets bool   `mapstructure:"highlight_snippets"`
	SnippetLength    int     `mapstructure:"snippet_length"`
	FuzzyTolerance   float64 `mapstructure:"fuzzy_tolerance"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Name           string `mapstructure:"name"`
	Version        string `mapstructure:"version"`
	EnableRecovery bool   `mapstructure:"enable_recovery"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
	File       string `mapstructure:"file"`
	JSONFormat bool   `mapstructure:"json_format"`
}

// ModelsConfig represents AI models configuration
type ModelsConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	DefaultModel string `mapstructure:"default_model"`
	ModelsDir    string `mapstructure:"models_dir"`
	MaxTokens    int    `mapstructure:"max_tokens"`
	Temperature  float64 `mapstructure:"temperature"`
}



// PatternSearchConfig represents pattern search configuration
type PatternSearchConfig struct {
	MaxResults      int      `mapstructure:"max_results"`
	TimeoutSeconds  int      `mapstructure:"timeout_seconds"`
	SupportedTypes  []string `mapstructure:"supported_types"`
}

// CodeSmellsConfig represents code smells detection configuration
type CodeSmellsConfig struct {
	DefaultSeverity     string   `mapstructure:"default_severity"`
	EnabledSmells       []string `mapstructure:"enabled_smells"`
	SeverityThresholds  map[string]float64 `mapstructure:"severity_thresholds"`
}

// SecurityConfig represents security analysis configuration
type SecurityConfig struct {
	EnabledChecks       []string `mapstructure:"enabled_checks"`
	ConfidenceThreshold float64  `mapstructure:"confidence_threshold"`
	ExcludePatterns     []string `mapstructure:"exclude_patterns"`
}

// ComplexityConfig represents complexity analysis configuration
type ComplexityConfig struct {
	CyclomaticThreshold int     `mapstructure:"cyclomatic_threshold"`
	CognitiveThreshold  int     `mapstructure:"cognitive_threshold"`
	HalsteadThreshold   float64 `mapstructure:"halstead_threshold"`
}

// TestCoverageConfig represents test coverage configuration
type TestCoverageConfig struct {
	MinCoverageThreshold float64  `mapstructure:"min_coverage_threshold"`
	TestDirectories      []string `mapstructure:"test_directories"`
	CoverageFormats      []string `mapstructure:"coverage_formats"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	DefaultMetrics      []string `mapstructure:"default_metrics"`
	OutputFormats       []string `mapstructure:"output_formats"`
	IncludeTrends       bool     `mapstructure:"include_trends"`
}

// EvolutionConfig represents evolution analysis configuration
type EvolutionConfig struct {
	DefaultTimeRange    int      `mapstructure:"default_time_range"`
	MaxCommits          int      `mapstructure:"max_commits"`
	IncludeAuthors      bool     `mapstructure:"include_authors"`
}

// PatternExtractionConfig represents pattern extraction configuration
type PatternExtractionConfig struct {
	MinOccurrences      int     `mapstructure:"min_occurrences"`
	MinPatternSize      int     `mapstructure:"min_pattern_size"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold"`
}



// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Indexer: IndexerConfig{
			SupportedExtensions: []string{
				".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".hpp",
				".rs", ".rb", ".php", ".cs", ".kt", ".swift", ".scala", ".clj",
				".hs", ".ml", ".sh", ".bash", ".zsh", ".fish", ".ps1", ".sql",
				".r", ".m", ".dart", ".lua", ".perl", ".pl",
			},
			MaxFileSize: 1048576, // 1MB
			ExcludePatterns: []string{
				"*/node_modules/*", "*/vendor/*", "*/.git/*", "*/build/*",
				"*/dist/*", "*/target/*", "*/__pycache__/*", "*.pyc",
				"*.class", "*.jar", "*.war", "*.ear", "*.exe", "*.dll",
				"*.so", "*.dylib", "*.a", "*.lib", "*.o", "*.obj",
				"*.min.js", "*.min.css",
			},
			IndexDir: "./index",
			RepoDir:  "./repositories",
		},
		Search: SearchConfig{
			MaxResults:       100,
			HighlightSnippets: true,
			SnippetLength:    200,
			FuzzyTolerance:   0.2,
		},
		Server: ServerConfig{
			Name:           "Code Indexer",
			Version:        "1.0.0",
			EnableRecovery: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
			File:       "",
			JSONFormat: true,
		},
		Models: ModelsConfig{
			Enabled:      true,
			DefaultModel: "code-assistant-v1",
			ModelsDir:    "./models",
			MaxTokens:    2048,
			Temperature:  0.7,
		},

	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Look for config file in current directory and common locations
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.code-indexer")
		viper.AddConfigPath("/etc/code-indexer")
	}

	// Environment variable support
	viper.SetEnvPrefix("INDEXER")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate and normalize paths
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration and normalizes paths
func (c *Config) Validate() error {
	// Validate indexer configuration
	if c.Indexer.IndexDir != "" {
		absDir, err := filepath.Abs(c.Indexer.IndexDir)
		if err != nil {
			return fmt.Errorf("invalid indexer index directory path %s: %w", c.Indexer.IndexDir, err)
		}
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("failed to create indexer index directory %s: %w", absDir, err)
		}
		c.Indexer.IndexDir = absDir
	}

	if c.Indexer.RepoDir != "" {
		absDir, err := filepath.Abs(c.Indexer.RepoDir)
		if err != nil {
			return fmt.Errorf("invalid indexer repo directory path %s: %w", c.Indexer.RepoDir, err)
		}
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("failed to create indexer repo directory %s: %w", absDir, err)
		}
		c.Indexer.RepoDir = absDir
	}

	if c.Indexer.MaxFileSize <= 0 {
		c.Indexer.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	// Validate Models configuration
	if c.Models.Enabled {
		if c.Models.ModelsDir != "" {
			absDir, err := filepath.Abs(c.Models.ModelsDir)
			if err != nil {
				return fmt.Errorf("invalid models directory path %s: %w", c.Models.ModelsDir, err)
			}
			if err := os.MkdirAll(absDir, 0755); err != nil {
				return fmt.Errorf("failed to create models directory %s: %w", absDir, err)
			}
			c.Models.ModelsDir = absDir
		}

		if c.Models.MaxTokens <= 0 {
			c.Models.MaxTokens = 2048
		}

		if c.Models.Temperature < 0 || c.Models.Temperature > 2 {
			c.Models.Temperature = 0.7
		}
	}

	// Validate numeric values
	if c.Indexer.MaxFileSize <= 0 {
		c.Indexer.MaxFileSize = 1048576 // 1MB default
	}
	
	if c.Search.MaxResults <= 0 {
		c.Search.MaxResults = 100
	}
	
	if c.Search.SnippetLength <= 0 {
		c.Search.SnippetLength = 200
	}
	
	if c.Search.FuzzyTolerance < 0 || c.Search.FuzzyTolerance > 1 {
		c.Search.FuzzyTolerance = 0.2
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[c.Logging.Level] {
		c.Logging.Level = "info"
	}

	// Validate Models configuration
	if c.Models.Enabled {
		if c.Models.ModelsDir != "" {
			absDir, err := filepath.Abs(c.Models.ModelsDir)
			if err != nil {
				return fmt.Errorf("invalid models directory path %s: %w", c.Models.ModelsDir, err)
			}
			if err := os.MkdirAll(absDir, 0755); err != nil {
				return fmt.Errorf("failed to create models directory %s: %w", absDir, err)
			}
			c.Models.ModelsDir = absDir
		}

		if c.Models.MaxTokens <= 0 {
			c.Models.MaxTokens = 2048
		}

		if c.Models.Temperature < 0 || c.Models.Temperature > 2 {
			c.Models.Temperature = 0.7
		}
	}

	return nil
}

// IsFileSupported checks if a file extension is supported for indexing
func (c *Config) IsFileSupported(filename string) bool {
	ext := filepath.Ext(filename)
	for _, supportedExt := range c.Indexer.SupportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// ShouldExcludeFile checks if a file should be excluded based on patterns
func (c *Config) ShouldExcludeFile(filePath string) bool {
	for _, pattern := range c.Indexer.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
		// Also check if any parent directory matches the pattern
		dir := filepath.Dir(filePath)
		for dir != "." && dir != "/" {
			if matched, _ := filepath.Match(pattern, dir); matched {
				return true
			}
			dir = filepath.Dir(dir)
		}
	}
	return false
}


