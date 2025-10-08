package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "docbrown",
	Short: "Automated documentation generator for Backstage",
	Long: `DocBrown automatically generates high-quality, Backstage-compatible
documentation from any codebase using LLMs (Claude or local Ollama models).

It analyzes code structure, generates comprehensive markdown documentation,
and seamlessly integrates with Git workflows through automatic PR creation
or direct push.`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .docbrown.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName(".docbrown")
		viper.SetConfigType("yaml")

		// Also look for global config
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home + "/.docbrown")
		}
	}

	viper.SetEnvPrefix("DOCBROWN")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
