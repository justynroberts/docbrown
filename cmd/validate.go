package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/validator"
)

var (
	validateStrict bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate generated documentation",
	Long: `Validate the generated documentation for quality, correctness,
and completeness. Checks markdown syntax, links, and Backstage catalog.`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "fail on warnings")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply CLI overrides
	if validateStrict {
		cfg.Quality.StrictMode = true
	}

	fmt.Println("Validating documentation...")
	fmt.Println()

	// Create validator
	v := validator.NewValidator(cfg.Documentation.OutputDir, cfg.Quality.StrictMode)

	// Validate
	results, err := v.Validate()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display results
	fmt.Println(v.FormatResults(results))

	// Check minimum score
	if results.QualityScore < cfg.Quality.MinScore {
		fmt.Printf("\n⚠ Quality score %.1f is below minimum %.1f\n",
			results.QualityScore, cfg.Quality.MinScore)

		if cfg.Quality.StrictMode {
			os.Exit(1)
		}
	}

	// Fail in strict mode if there are errors
	if cfg.Quality.StrictMode {
		if len(results.MarkdownErrors) > 0 || len(results.BrokenLinks) > 0 || !results.CatalogValid {
			fmt.Println("\n✗ Validation failed (strict mode)")
			os.Exit(1)
		}
	}

	if results.QualityScore >= cfg.Quality.MinScore {
		fmt.Println("\n✅ Documentation quality meets requirements")
	}

	return nil
}
