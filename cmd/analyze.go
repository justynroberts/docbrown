package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/orchestrator"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze repository structure",
	Long: `Scan and analyze the codebase structure to identify components,
detect programming languages, and extract metadata.`,
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Execute analysis
	ctx := context.Background()
	if _, err := orch.ExecuteAnalyze(ctx); err != nil {
		return err
	}

	return nil
}
