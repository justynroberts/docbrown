package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/cache"
	"github.com/docbrown/cli/internal/config"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage documentation cache",
	Long:  `View and manage the documentation generation cache.`,
}

var cacheShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show cache status",
	Long:  `Display information about the current cache.`,
	RunE:  runCacheShow,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cache",
	Long:  `Clear the documentation cache. Next run will regenerate all components.`,
	RunE:  runCacheClear,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheShowCmd)
	cacheCmd.AddCommand(cacheClearCmd)
}

func runCacheShow(cmd *cobra.Command, args []string) error {
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cacheMgr := cache.NewManager(
		cfg.Cache.Dir+"/cache.yaml",
		cfg.Cache.Enabled,
		cfg.Cache.TTL,
	)

	if err := cacheMgr.Load(); err != nil {
		fmt.Println("Cache: Not initialized")
		return nil
	}

	stats := cacheMgr.GetStats()

	fmt.Println("Cache Status:")
	fmt.Println()

	if enabled, ok := stats["enabled"].(bool); ok && !enabled {
		fmt.Println("Status: Disabled")
		return nil
	}

	fmt.Println("Status: Enabled")

	if lastRun, ok := stats["last_run"].(interface{}); ok {
		fmt.Printf("Last run: %v\n", lastRun)
	}

	if components, ok := stats["components"].(int); ok {
		fmt.Printf("Components cached: %d\n", components)
	}

	if unchanged, ok := stats["unchanged"].(int); ok {
		fmt.Printf("Unchanged: %d\n", unchanged)
	}

	if stale, ok := stats["stale"].(int); ok {
		fmt.Printf("Stale: %d\n", stale)
	}

	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cacheMgr := cache.NewManager(
		cfg.Cache.Dir+"/cache.yaml",
		cfg.Cache.Enabled,
		cfg.Cache.TTL,
	)

	if err := cacheMgr.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("✓ Cache cleared")
	fmt.Println()
	fmt.Println("Next run will regenerate all components.")

	return nil
}
