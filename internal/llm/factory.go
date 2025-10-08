package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/docbrown/cli/internal/config"
)

// NewProvider creates a new LLM provider based on configuration
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.LLM.Provider {
	case "anthropic":
		return newAnthropicFromConfig(cfg)
	case "ollama":
		return newOllamaFromConfig(cfg)
	case "auto":
		return detectProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.LLM.Provider)
	}
}

// newAnthropicFromConfig creates an Anthropic provider from config
func newAnthropicFromConfig(cfg *config.Config) (Provider, error) {
	if cfg.LLM.Anthropic.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured (set ANTHROPIC_API_KEY)")
	}

	return NewAnthropicProvider(
		cfg.LLM.Anthropic.APIKey,
		cfg.LLM.Anthropic.Model,
		cfg.LLM.Anthropic.MaxTokens,
	), nil
}

// newOllamaFromConfig creates an Ollama provider from config
func newOllamaFromConfig(cfg *config.Config) (Provider, error) {
	provider := NewOllamaProvider(
		cfg.LLM.Ollama.Endpoint,
		cfg.LLM.Ollama.Model,
		cfg.LLM.Ollama.ContextSize,
		cfg.LLM.Ollama.Timeout,
	)

	// Check if Ollama is actually available
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := provider.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Ollama not available: %w", err)
	}

	return provider, nil
}

// detectProvider auto-detects the best available provider
func detectProvider(cfg *config.Config) (Provider, error) {
	// Try Ollama first (free and local)
	ollama := NewOllamaProvider(
		cfg.LLM.Ollama.Endpoint,
		cfg.LLM.Ollama.Model,
		cfg.LLM.Ollama.ContextSize,
		cfg.LLM.Ollama.Timeout,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := ollama.Ping(ctx); err == nil {
		fmt.Println("Using Ollama (local, free)")
		return ollama, nil
	}

	// Fall back to Anthropic
	if cfg.LLM.Anthropic.APIKey != "" {
		fmt.Println("Using Anthropic Claude")
		return NewAnthropicProvider(
			cfg.LLM.Anthropic.APIKey,
			cfg.LLM.Anthropic.Model,
			cfg.LLM.Anthropic.MaxTokens,
		), nil
	}

	return nil, fmt.Errorf("no LLM provider available (tried Ollama and Anthropic)")
}

// CheckProviderStatus checks the status of all configured providers
func CheckProviderStatus(cfg *config.Config) map[string]string {
	status := make(map[string]string)

	// Check Ollama
	ollama := NewOllamaProvider(
		cfg.LLM.Ollama.Endpoint,
		cfg.LLM.Ollama.Model,
		cfg.LLM.Ollama.ContextSize,
		cfg.LLM.Ollama.Timeout,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := ollama.Ping(ctx); err == nil {
		status["ollama"] = "available"
	} else {
		status["ollama"] = fmt.Sprintf("unavailable: %v", err)
	}

	// Check Anthropic
	if cfg.LLM.Anthropic.APIKey != "" {
		status["anthropic"] = "configured"
	} else {
		status["anthropic"] = "not configured (missing API key)"
	}

	return status
}
