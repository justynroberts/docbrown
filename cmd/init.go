package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/docbrown/cli/internal/config"
)

var (
	initForce    bool
	initTemplate string
	initProvider string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize DocBrown in current repository",
	Long: `Initialize DocBrown by creating a .docbrown.yaml configuration file
with sensible defaults based on the detected repository structure.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing .docbrown.yaml")
	initCmd.Flags().StringVar(&initTemplate, "template", "backstage", "template to use")
	initCmd.Flags().StringVar(&initProvider, "provider", "auto", "LLM provider (auto/anthropic/ollama)")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing DocBrown...")
	fmt.Println()

	// Check if git repository
	if !isGitRepo() {
		return fmt.Errorf("not a git repository (run 'git init' first)")
	}
	fmt.Println("✓ Git repository detected")

	// Check if config already exists
	configPath := ".docbrown.yaml"
	if _, err := os.Stat(configPath); err == nil && !initForce {
		return fmt.Errorf(".docbrown.yaml already exists (use --force to overwrite)")
	}

	// Detect repository type
	repoType := detectRepoType()
	fmt.Printf("✓ Detected: %s\n", repoType)

	// Create config with defaults
	cfg := config.DefaultConfig()
	cfg.Documentation.Template = initTemplate
	cfg.LLM.Provider = initProvider

	// Write config
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Created .docbrown.yaml with recommended settings:")
	fmt.Printf("  - Template: %s\n", initTemplate)
	fmt.Printf("  - Provider: %s\n", initProvider)
	fmt.Printf("  - Output: %s\n", cfg.Documentation.OutputDir)
	fmt.Println()

	fmt.Println("Next steps:")
	fmt.Println("  1. Set API key: export ANTHROPIC_API_KEY=sk-...")
	fmt.Println("     (or use local Ollama: no key needed)")
	fmt.Println("  2. Run: docbrown auto")

	return nil
}

func isGitRepo() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func detectRepoType() string {
	// Simple detection logic
	if _, err := os.Stat("go.mod"); err == nil {
		return "Go project"
	}
	if _, err := os.Stat("package.json"); err == nil {
		return "Node.js project"
	}
	if _, err := os.Stat("setup.py"); err == nil {
		return "Python project"
	}
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "Rust project"
	}
	return "Unknown project type"
}
