package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/orchestrator"
)

var (
	autoProvider string
)

var autoCmd = &cobra.Command{
	Use:   "auto",
	Short: "Run complete documentation workflow",
	Long: `Run the complete documentation workflow:
  1. Analyze codebase
  2. Generate documentation
  3. Validate quality

This is the recommended way to use DocBrown.`,
	RunE: runAuto,
}

func init() {
	rootCmd.AddCommand(autoCmd)

	autoCmd.Flags().StringVar(&autoProvider, "provider", "", "LLM provider (anthropic/ollama/auto)")
}

func runAuto(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply CLI overrides
	if autoProvider != "" {
		cfg.LLM.Provider = autoProvider
	}

	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Execute auto workflow
	ctx := context.Background()
	if err := orch.ExecuteAuto(ctx); err != nil {
		return err
	}

	return nil
}
