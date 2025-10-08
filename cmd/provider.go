package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/llm"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage LLM providers",
	Long:  `Check status and manage LLM provider configurations.`,
}

var providerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check provider status",
	Long:  `Check the availability and status of configured LLM providers.`,
	RunE:  runProviderStatus,
}

func init() {
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerStatusCmd)
}

func runProviderStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("LLM Providers:")
	fmt.Println()

	// Check provider status
	status := llm.CheckProviderStatus(cfg)

	// Ollama
	fmt.Println("Ollama (Local):")
	fmt.Printf("  Status: %s\n", status["ollama"])
	fmt.Printf("  Endpoint: %s\n", cfg.LLM.Ollama.Endpoint)
	fmt.Printf("  Model: %s\n", cfg.LLM.Ollama.Model)
	fmt.Printf("  Cost: Free\n")
	fmt.Println()

	// Anthropic
	fmt.Println("Anthropic Claude (Cloud):")
	fmt.Printf("  Status: %s\n", status["anthropic"])
	if cfg.LLM.Anthropic.APIKey != "" {
		fmt.Println("  API Key: Configured")
	} else {
		fmt.Println("  API Key: Not configured")
	}
	fmt.Printf("  Model: %s\n", cfg.LLM.Anthropic.Model)
	fmt.Printf("  Estimated cost: ~$0.50/repo\n")
	fmt.Println()

	// Recommendation
	if status["ollama"] == "available" {
		fmt.Println("Recommendation: Using Ollama (free, available)")
	} else if status["anthropic"] == "configured" {
		fmt.Println("Recommendation: Using Anthropic (Ollama not available)")
	} else {
		fmt.Println("⚠ No provider available")
		fmt.Println()
		fmt.Println("To use Ollama:")
		fmt.Println("  1. Install: https://ollama.ai/download")
		fmt.Println("  2. Run: ollama pull qwen2.5-coder:latest")
		fmt.Println()
		fmt.Println("To use Anthropic:")
		fmt.Println("  1. Get API key: https://console.anthropic.com")
		fmt.Println("  2. Set: export ANTHROPIC_API_KEY=sk-...")
	}

	return nil
}
