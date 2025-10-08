package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/template"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage documentation templates",
	Long:  `List and manage documentation templates.`,
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available documentation templates.`,
	RunE:  runTemplatesList,
}

var templatesShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template details",
	Long:  `Display details about a specific template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplatesShow,
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesShowCmd)
}

func runTemplatesList(cmd *cobra.Command, args []string) error {
	engine := template.NewEngine("templates")

	templates, err := engine.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	fmt.Println("Available templates:")
	fmt.Println()

	if len(templates) == 0 {
		fmt.Println("No templates found")
		return nil
	}

	for _, name := range templates {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println()
	fmt.Println("Use: docbrown generate --template <name>")

	return nil
}

func runTemplatesShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	engine := template.NewEngine("templates")

	tmpl, err := engine.LoadTemplate(name)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	fmt.Printf("Template: %s\n", tmpl.Name)
	fmt.Printf("Version: %s\n", tmpl.Version)
	fmt.Printf("Description: %s\n", tmpl.Description)
	fmt.Println()

	fmt.Println("Files:")
	for _, file := range tmpl.Files {
		fmt.Printf("  - %s → %s\n", file.Template, file.Output)
		if file.Description != "" {
			fmt.Printf("    %s\n", file.Description)
		}
	}

	return nil
}
