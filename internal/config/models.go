package config

import "time"

// Config represents the complete DocBrown configuration
type Config struct {
	LLM           LLMConfig           `yaml:"llm" mapstructure:"llm"`
	Documentation DocumentationConfig `yaml:"documentation" mapstructure:"documentation"`
	Git           GitConfig           `yaml:"git" mapstructure:"git"`
	Backstage     BackstageConfig     `yaml:"backstage" mapstructure:"backstage"`
	Quality       QualityConfig       `yaml:"quality" mapstructure:"quality"`
	Cache         CacheConfig         `yaml:"cache" mapstructure:"cache"`
	Performance   PerformanceConfig   `yaml:"performance" mapstructure:"performance"`
}

// LLMConfig contains LLM provider settings
type LLMConfig struct {
	Provider  string          `yaml:"provider" mapstructure:"provider"`
	Anthropic AnthropicConfig `yaml:"anthropic" mapstructure:"anthropic"`
	Ollama    OllamaConfig    `yaml:"ollama" mapstructure:"ollama"`
}

// AnthropicConfig contains Anthropic-specific settings
type AnthropicConfig struct {
	APIKey    string `yaml:"api_key" mapstructure:"api_key"`
	Model     string `yaml:"model" mapstructure:"model"`
	MaxTokens int    `yaml:"max_tokens" mapstructure:"max_tokens"`
}

// OllamaConfig contains Ollama-specific settings
type OllamaConfig struct {
	Endpoint    string        `yaml:"endpoint" mapstructure:"endpoint"`
	Model       string        `yaml:"model" mapstructure:"model"`
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`
	ContextSize int           `yaml:"context_size" mapstructure:"context_size"`
}

// DocumentationConfig contains documentation generation settings
type DocumentationConfig struct {
	Template          string   `yaml:"template" mapstructure:"template"`
	OutputDir         string   `yaml:"output_dir" mapstructure:"output_dir"`
	IncludePatterns   []string `yaml:"include_patterns" mapstructure:"include_patterns"`
	ExcludePatterns   []string `yaml:"exclude_patterns" mapstructure:"exclude_patterns"`
	ExcludeSensitive  []string `yaml:"exclude_sensitive" mapstructure:"exclude_sensitive"`
}

// GitConfig contains Git-related settings
type GitConfig struct {
	Remote       string   `yaml:"remote" mapstructure:"remote"`
	BaseBranch   string   `yaml:"base_branch" mapstructure:"base_branch"`
	BranchPrefix string   `yaml:"branch_prefix" mapstructure:"branch_prefix"`
	PushStrategy string   `yaml:"push_strategy" mapstructure:"push_strategy"`
	PAT          string   `yaml:"pat" mapstructure:"pat"`
	EncryptedPAT string   `yaml:"encrypted_pat" mapstructure:"encrypted_pat"`
	PRTemplate   string   `yaml:"pr_template" mapstructure:"pr_template"`
	AutoMerge    bool     `yaml:"auto_merge" mapstructure:"auto_merge"`
	PRLabels     []string `yaml:"pr_labels" mapstructure:"pr_labels"`
}

// BackstageConfig contains Backstage-specific settings
type BackstageConfig struct {
	CatalogFile string                 `yaml:"catalog_file" mapstructure:"catalog_file"`
	Owner       string                 `yaml:"owner" mapstructure:"owner"`
	System      string                 `yaml:"system" mapstructure:"system"`
	Lifecycle   string                 `yaml:"lifecycle" mapstructure:"lifecycle"`
	Metadata    map[string]interface{} `yaml:"metadata" mapstructure:"metadata"`
}

// QualityConfig contains quality validation settings
type QualityConfig struct {
	MinScore              float64 `yaml:"min_score" mapstructure:"min_score"`
	RequireAPIDocs        bool    `yaml:"require_api_docs" mapstructure:"require_api_docs"`
	RequireArchitecture   bool    `yaml:"require_architecture" mapstructure:"require_architecture"`
	RequireGettingStarted bool    `yaml:"require_getting_started" mapstructure:"require_getting_started"`
	StrictMode            bool    `yaml:"strict_mode" mapstructure:"strict_mode"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	Enabled bool          `yaml:"enabled" mapstructure:"enabled"`
	Dir     string        `yaml:"dir" mapstructure:"dir"`
	TTL     time.Duration `yaml:"ttl" mapstructure:"ttl"`
}

// PerformanceConfig contains performance tuning settings
type PerformanceConfig struct {
	MaxConcurrent          int `yaml:"max_concurrent" mapstructure:"max_concurrent"`
	MaxFilesPerComponent   int `yaml:"max_files_per_component" mapstructure:"max_files_per_component"`
	MaxContextTokens       int `yaml:"max_context_tokens" mapstructure:"max_context_tokens"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: "auto",
			Anthropic: AnthropicConfig{
				Model:     "claude-sonnet-4-20250514",
				MaxTokens: 4096,
			},
			Ollama: OllamaConfig{
				Endpoint:    "http://localhost:11434",
				Model:       "qwen2.5-coder:latest",
				Timeout:     300 * time.Second,
				ContextSize: 8192,
			},
		},
		Documentation: DocumentationConfig{
			Template:  "backstage",
			OutputDir: "docs",
			IncludePatterns: []string{
				"**/*.go",
				"**/*.py",
				"**/*.ts",
				"**/*.js",
				"**/*.java",
				"**/*.rs",
			},
			ExcludePatterns: []string{
				"**/test/**",
				"**/tests/**",
				"**/*_test.go",
				"**/*_test.py",
				"**/vendor/**",
				"**/node_modules/**",
				"**/.venv/**",
				"**/dist/**",
				"**/build/**",
			},
			ExcludeSensitive: []string{
				"**/*secret*",
				"**/*key*",
				"**/*password*",
				"**/*.pem",
				"**/*.key",
				"**/.env*",
				"**/credentials*",
			},
		},
		Git: GitConfig{
			Remote:       "origin",
			BaseBranch:   "main",
			BranchPrefix: "docs/auto-gen",
			PushStrategy: "auto",
			PRLabels:     []string{"documentation", "automated"},
		},
		Backstage: BackstageConfig{
			CatalogFile: "catalog-info.yaml",
			Owner:       "team-platform",
			System:      "core",
			Lifecycle:   "production",
		},
		Quality: QualityConfig{
			MinScore:              7.0,
			RequireAPIDocs:        true,
			RequireArchitecture:   true,
			RequireGettingStarted: true,
			StrictMode:            false,
		},
		Cache: CacheConfig{
			Enabled: true,
			Dir:     ".docbrown/cache",
			TTL:     168 * time.Hour, // 7 days
		},
		Performance: PerformanceConfig{
			MaxConcurrent:        5,
			MaxFilesPerComponent: 100,
			MaxContextTokens:     8000,
		},
	}
}
