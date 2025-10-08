package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Engine handles template loading and rendering
type Engine struct {
	templatePath string
	templates    map[string]*template.Template
}

// NewEngine creates a new template engine
func NewEngine(templatePath string) *Engine {
	return &Engine{
		templatePath: templatePath,
		templates:    make(map[string]*template.Template),
	}
}

// LoadTemplate loads a template by name
func (e *Engine) LoadTemplate(name string) (*Template, error) {
	templateDir := filepath.Join(e.templatePath, name)

	// Check if template exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	// Load template.yaml
	configPath := filepath.Join(templateDir, "template.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template config: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template config: %w", err)
	}

	tmpl.Path = templateDir

	// Load template files
	for i := range tmpl.Files {
		file := &tmpl.Files[i]
		tmplPath := filepath.Join(templateDir, file.Template)

		t, err := template.ParseFiles(tmplPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template file %s: %w", file.Template, err)
		}

		e.templates[file.Name] = t
	}

	return &tmpl, nil
}

// Render renders a template with the given data
func (e *Engine) Render(templateName string, data interface{}) (string, error) {
	tmpl, ok := e.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not loaded: %s", templateName)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// RenderToFile renders a template to a file
func (e *Engine) RenderToFile(templateName string, data interface{}, outputPath string) error {
	content, err := e.Render(templateName, data)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// RenderAll renders all files in a template
func (e *Engine) RenderAll(tmpl *Template, data TemplateData, outputDir string) ([]string, error) {
	var generatedFiles []string

	for _, file := range tmpl.Files {
		// Determine output path
		outputPath := file.Output

		// Handle template variables in output path
		if strings.Contains(outputPath, "{{") {
			outputPath = e.expandPath(outputPath, data)
		}

		fullPath := filepath.Join(outputDir, outputPath)

		// Check if this is a foreach template
		if file.Foreach != "" {
			// Render multiple times for each item
			items := e.getForEachItems(file.Foreach, data)
			for _, item := range items {
				itemPath := e.expandPath(outputPath, item)
				fullItemPath := filepath.Join(outputDir, itemPath)

				if err := e.RenderToFile(file.Name, item, fullItemPath); err != nil {
					return generatedFiles, err
				}

				generatedFiles = append(generatedFiles, fullItemPath)
			}
		} else {
			// Render once
			if err := e.RenderToFile(file.Name, data, fullPath); err != nil {
				return generatedFiles, err
			}

			generatedFiles = append(generatedFiles, fullPath)
		}
	}

	return generatedFiles, nil
}

// expandPath expands template variables in a path
func (e *Engine) expandPath(path string, data interface{}) string {
	// Simple replacement for now
	result := path

	// Handle ComponentData directly
	if comp, ok := data.(ComponentData); ok {
		result = strings.ReplaceAll(result, "{{.Name}}", comp.Name)
		result = strings.ReplaceAll(result, "{{.ComponentName}}", comp.Name)
		return result
	}

	// Handle ServiceData directly
	if svc, ok := data.(ServiceData); ok {
		result = strings.ReplaceAll(result, "{{.Name}}", svc.Name)
		result = strings.ReplaceAll(result, "{{.ServiceName}}", svc.Name)
		return result
	}

	// Extract data as map (fallback)
	if m, ok := data.(map[string]interface{}); ok {
		for key, value := range m {
			placeholder := "{{." + key + "}}"
			if str, ok := value.(string); ok {
				result = strings.ReplaceAll(result, placeholder, str)
			}
		}
	}

	return result
}

// getForEachItems gets items for a foreach loop
func (e *Engine) getForEachItems(foreach string, data TemplateData) []interface{} {
	// Simplified implementation
	switch foreach {
	case "components":
		result := make([]interface{}, len(data.Components))
		for i, comp := range data.Components {
			// Pass the component data directly so templates can access fields
			result[i] = comp
		}
		return result
	case "services":
		result := make([]interface{}, len(data.Services))
		for i, svc := range data.Services {
			// Pass the service data directly
			result[i] = svc
		}
		return result
	}

	return []interface{}{}
}

// ListTemplates lists available templates
func (e *Engine) ListTemplates() ([]string, error) {
	var templates []string

	entries, err := os.ReadDir(e.templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has a template.yaml
			configPath := filepath.Join(e.templatePath, entry.Name(), "template.yaml")
			if _, err := os.Stat(configPath); err == nil {
				templates = append(templates, entry.Name())
			}
		}
	}

	return templates, nil
}
