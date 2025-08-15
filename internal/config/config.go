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
	Level       string `mapstructure:"level"`
	File        string `mapstructure:"file"`
	JSONFormat  bool   `mapstructure:"json_format"`
	MaxSize     int    `mapstructure:"max_size"`
	MaxBackups  int    `mapstructure:"max_backups"`
	MaxAge      int    `mapstructure:"max_age"`
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
			File:       "indexer.log",
			JSONFormat: false,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     30,
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
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate validates the configuration and normalizes paths
func (c *Config) validate() error {
	// Ensure directories exist or can be created
	dirs := []string{c.Indexer.IndexDir, c.Indexer.RepoDir}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("invalid directory path %s: %w", dir, err)
		}
		
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", absDir, err)
		}
	}

	// Normalize paths
	if c.Indexer.IndexDir != "" {
		abs, _ := filepath.Abs(c.Indexer.IndexDir)
		c.Indexer.IndexDir = abs
	}
	
	if c.Indexer.RepoDir != "" {
		abs, _ := filepath.Abs(c.Indexer.RepoDir)
		c.Indexer.RepoDir = abs
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
