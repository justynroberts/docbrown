package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/orchestrator"
)

var (
	generateProvider string
	generateTemplate string
	genNoCache       bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation",
	Long: `Generate documentation from the analyzed codebase using LLMs
and the configured template.`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateProvider, "provider", "", "LLM provider (anthropic/ollama/auto)")
	generateCmd.Flags().StringVar(&generateTemplate, "template", "", "documentation template")
	generateCmd.Flags().BoolVar(&genNoCache, "no-cache", false, "disable cache, regenerate all")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply CLI overrides
	if generateProvider != "" {
		cfg.LLM.Provider = generateProvider
	}
	if generateTemplate != "" {
		cfg.Documentation.Template = generateTemplate
	}
	if genNoCache {
		cfg.Cache.Enabled = false
	}

	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Execute generation
	ctx := context.Background()
	if err := orch.ExecuteGenerate(ctx); err != nil {
		return err
	}

	return nil
}
