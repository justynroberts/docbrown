package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Manager handles configuration loading and merging
type Manager struct {
	config *Config
	v      *viper.Viper
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config: DefaultConfig(),
		v:      viper.New(),
	}
}

// Load loads configuration from all sources and merges them
func (m *Manager) Load() (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Try to load global config
	if err := m.loadGlobalConfig(); err == nil {
		if err := m.v.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal global config: %w", err)
		}
	}

	// Try to load repository config
	m.v.SetConfigName(".docbrown")
	m.v.SetConfigType("yaml")
	m.v.AddConfigPath(".")

	if err := m.v.MergeInConfig(); err == nil {
		if err := m.v.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal repo config: %w", err)
		}
	}

	// Apply environment variable overrides
	m.applyEnvOverrides(config)

	m.config = config
	return config, nil
}

// loadGlobalConfig loads the global configuration file
func (m *Manager) loadGlobalConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	globalConfigPath := filepath.Join(home, ".docbrown", "config.yaml")
	m.v.SetConfigFile(globalConfigPath)

	return m.v.ReadInConfig()
}

// applyEnvOverrides applies environment variable overrides
func (m *Manager) applyEnvOverrides(config *Config) {
	// LLM provider
	if provider := os.Getenv("DOCBROWN_PROVIDER"); provider != "" {
		config.LLM.Provider = provider
	}

	// Anthropic API key
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.LLM.Anthropic.APIKey = apiKey
	}

	// GitHub/GitLab/Bitbucket tokens
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.Git.PAT = token
	} else if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		config.Git.PAT = token
	} else if token := os.Getenv("BITBUCKET_TOKEN"); token != "" {
		config.Git.PAT = token
	}
}

// Get retrieves a configuration value by key
func (m *Manager) Get(key string) interface{} {
	return m.v.Get(key)
}

// Set sets a configuration value
func (m *Manager) Set(key string, value interface{}) error {
	m.v.Set(key, value)
	return nil
}

// Save saves the current configuration to the repository config file
func (m *Manager) Save() error {
	return m.v.WriteConfig()
}

// SaveAs saves the current configuration to a specific file
func (m *Manager) SaveAs(path string) error {
	return m.v.WriteConfigAs(path)
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// Validate validates the configuration
func (m *Manager) Validate() error {
	config := m.config

	// Validate provider
	validProviders := []string{"auto", "anthropic", "ollama"}
	if !contains(validProviders, config.LLM.Provider) {
		return fmt.Errorf("invalid provider: %s (must be one of: auto, anthropic, ollama)", config.LLM.Provider)
	}

	// Validate output directory
	if config.Documentation.OutputDir == "" {
		return fmt.Errorf("output_dir cannot be empty")
	}

	// Validate push strategy
	validStrategies := []string{"auto", "direct", "pr"}
	if !contains(validStrategies, config.Git.PushStrategy) {
		return fmt.Errorf("invalid push_strategy: %s (must be one of: auto, direct, pr)", config.Git.PushStrategy)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
