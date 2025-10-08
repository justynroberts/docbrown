package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/docbrown/cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and modify DocBrown configuration settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration from all sources.`,
	RunE:  runConfigShow,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get configuration value",
	Long:  `Get a specific configuration value.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long:  `Set a configuration value in the repository config file.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Configuration:")
	fmt.Println()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(data))

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfgMgr := config.NewManager()
	if _, err := cfgMgr.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	value := cfgMgr.Get(key)
	if value == nil {
		return fmt.Errorf("key not found: %s", key)
	}

	fmt.Println(value)

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfgMgr := config.NewManager()
	if _, err := cfgMgr.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfgMgr.Set(key, value); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Set %s = %s\n", key, value)

	return nil
}
