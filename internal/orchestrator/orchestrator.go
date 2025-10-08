package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docbrown/cli/internal/analyzer"
	"github.com/docbrown/cli/internal/cache"
	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/llm"
	"github.com/docbrown/cli/internal/template"
	"github.com/docbrown/cli/internal/validator"
)

// Orchestrator coordinates the documentation generation workflow
type Orchestrator struct {
	config      *config.Config
	analyzer    *analyzer.Analyzer
	llmPool     *llm.Pool
	templateEng *template.Engine
	cacheManager *cache.Manager
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(cfg *config.Config) (*Orchestrator, error) {
	// Create LLM provider
	provider, err := llm.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Create LLM pool
	llmPool := llm.NewPool(provider, cfg.Performance.MaxConcurrent)

	// Create analyzer
	analyzer := analyzer.NewAnalyzer(".", cfg.Documentation.ExcludePatterns)

	// Create template engine
	templatePath := cfg.Documentation.TemplatePath
	if templatePath == "" {
		templatePath = "templates"
	}
	templateEng := template.NewEngine(templatePath)

	// Create cache manager
	cacheManager := cache.NewManager(
		cfg.Cache.Dir+"/cache.yaml",
		cfg.Cache.Enabled,
		cfg.Cache.TTL,
	)

	return &Orchestrator{
		config:       cfg,
		analyzer:     analyzer,
		llmPool:      llmPool,
		templateEng:  templateEng,
		cacheManager: cacheManager,
	}, nil
}

// ExecuteAnalyze performs repository analysis
func (o *Orchestrator) ExecuteAnalyze(ctx context.Context) (*analyzer.RepoStructure, error) {
	fmt.Println("🔍 Analyzing repository...")

	structure, err := o.analyzer.Analyze()
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	fmt.Printf("✓ Analysis complete\n")
	fmt.Printf("  - Files: %d\n", structure.TotalFiles)
	fmt.Printf("  - Components: %d\n", len(structure.Components))

	for lang, count := range structure.Languages {
		fmt.Printf("  - %s: %d files\n", lang, count)
	}

	return structure, nil
}

// ExecuteGenerate performs documentation generation
func (o *Orchestrator) ExecuteGenerate(ctx context.Context) error {
	fmt.Println("🤖 Generating documentation...")

	// Step 1: Analyze
	structure, err := o.ExecuteAnalyze(ctx)
	if err != nil {
		return err
	}

	// Step 2: Load cache
	if err := o.cacheManager.Load(); err != nil {
		fmt.Printf("⚠ Failed to load cache: %v\n", err)
	}

	// Step 3: Determine what needs to be regenerated
	componentsToGen := o.getComponentsToGenerate(structure)

	if len(componentsToGen) == 0 {
		fmt.Println("✓ All components up to date (using cache)")
		return nil
	}

	fmt.Printf("Generating %d components (skipping %d cached)\n",
		len(componentsToGen),
		len(structure.Components)-len(componentsToGen))

	// Step 4: Load template
	tmpl, err := o.templateEng.LoadTemplate(o.config.Documentation.Template)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Step 5: Use LLM to generate content for each component
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🤖 Calling LLM to generate content for %d components...\n", len(componentsToGen))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	enrichedComponents, err := o.generateWithLLM(ctx, structure, componentsToGen)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ LLM content generation complete")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Step 6: Build template data with LLM-generated content
	templateData := o.buildTemplateData(structure, enrichedComponents)

	// Step 7: Render templates
	generatedFiles, err := o.templateEng.RenderAll(tmpl, templateData, o.config.Documentation.OutputDir)
	if err != nil {
		return fmt.Errorf("template rendering failed: %w", err)
	}

	// Step 8: Update cache
	for _, comp := range componentsToGen {
		o.cacheManager.Update(comp.Name, comp.Files)
	}

	if err := o.cacheManager.Save(); err != nil {
		fmt.Printf("⚠ Failed to save cache: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("✓ Generated %d files\n", len(generatedFiles))
	for _, file := range generatedFiles {
		fmt.Printf("  - %s\n", file)
	}

	return nil
}

// ExecuteAuto performs the complete workflow
func (o *Orchestrator) ExecuteAuto(ctx context.Context) error {
	startTime := time.Now()

	fmt.Println("DocBrown - Automated Documentation")
	fmt.Println()

	// Step 1: Analyze
	fmt.Println("🔍 Step 1/4: Analyzing codebase...")
	structure, err := o.ExecuteAnalyze(ctx)
	if err != nil {
		return err
	}
	fmt.Println()

	// Step 2: Generate
	fmt.Println("🤖 Step 2/4: Generating documentation...")
	if err := o.ExecuteGenerate(ctx); err != nil {
		return err
	}
	fmt.Println()

	// Step 3: Validate
	fmt.Println("✅ Step 3/4: Validating quality...")
	score, err := o.ExecuteValidate()
	if err != nil {
		return err
	}
	fmt.Println()

	// Step 4: Summary
	fmt.Println("🎉 Step 4/4: Complete")
	fmt.Println()

	duration := time.Since(startTime)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Summary:")
	fmt.Printf("  Components processed: %d\n", len(structure.Components))
	fmt.Printf("  Quality score: %.1f/10.0\n", score)
	fmt.Printf("  Time: %s\n", duration.Round(time.Second))

	// Show cost if using paid provider
	if o.llmPool.GetProvider().Name() == "anthropic" {
		cost := o.llmPool.GetTotalCost()
		fmt.Printf("  Cost: $%.2f\n", cost)
	} else {
		fmt.Println("  Cost: $0.00 (Ollama)")
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Review generated documentation in docs/")
	fmt.Println("  - Run: docbrown pr (to create pull request)")
	fmt.Println("  - Or: docbrown pr --push-direct (to push directly)")

	return nil
}

// ExecuteValidate performs validation
func (o *Orchestrator) ExecuteValidate() (float64, error) {
	v := validator.NewValidator(o.config.Documentation.OutputDir, o.config.Quality.StrictMode)

	results, err := v.Validate()
	if err != nil {
		return 0.0, err
	}

	// Check minimum score
	if results.QualityScore < o.config.Quality.MinScore {
		fmt.Printf("⚠ Quality score %.1f below minimum %.1f\n",
			results.QualityScore, o.config.Quality.MinScore)
	} else {
		fmt.Printf("✓ Quality score: %.1f/10.0\n", results.QualityScore)
	}

	return results.QualityScore, nil
}

// generateWithLLM uses the LLM to generate content for components
func (o *Orchestrator) generateWithLLM(ctx context.Context, structure *analyzer.RepoStructure, components []analyzer.Component) ([]EnrichedComponent, error) {
	enriched := make([]EnrichedComponent, len(components))

	for i, comp := range components {
		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(components), comp.Name)
		fmt.Printf("  Type: %s | Language: %s | Files: %d\n", comp.Type, comp.Language, len(comp.Files))

		// Prepare context for LLM
		keyFiles := o.selectKeyFiles(comp)
		fmt.Printf("  📄 Selected %d key files for analysis\n", len(keyFiles))

		// Call LLM to analyze and generate overview
		fmt.Printf("  🤖 Analyzing component structure...\n")
		analysisReq := llm.AnalysisRequest{
			ComponentName: comp.Name,
			ComponentType: comp.Type,
			Language:      comp.Language,
			Path:          comp.Path,
			FileTree:      structure.FileTree,
			KeyFiles:      keyFiles,
		}

		result, err := o.llmPool.Analyze(ctx, analysisReq)
		if err != nil {
			fmt.Printf("  ⚠ LLM analysis failed: %v\n", comp.Name, err)
			// Continue with basic info
			enriched[i] = EnrichedComponent{
				Component: comp,
				Overview:  "Documentation for " + comp.Name,
			}
			continue
		}
		fmt.Printf("  ✓ Analysis complete (%d chars)\n", len(result.Overview))

		// Generate detailed documentation
		fmt.Printf("  🤖 Generating detailed documentation...\n")
		generateReq := llm.GenerateRequest{
			ComponentName: comp.Name,
			ComponentType: comp.Type,
			Language:      comp.Language,
			Path:          comp.Path,
			Files:         keyFiles,
		}

		detailedDocs, err := o.llmPool.Generate(ctx, generateReq)
		if err != nil {
			fmt.Printf("  ⚠ Documentation generation failed: %v\n", comp.Name, err)
			detailedDocs = "## " + comp.Name + "\n\n" + result.Overview
		}
		fmt.Printf("  ✓ Documentation complete (%d chars)\n", len(detailedDocs))

		enriched[i] = EnrichedComponent{
			Component:    comp,
			Overview:     result.Overview,
			DetailedDocs: detailedDocs,
			Architecture: detailedDocs, // Use the LLM-generated detailed docs as architecture
		}

		fmt.Printf("  ✅ Component processing complete\n")
	}

	return enriched, nil
}

// selectKeyFiles selects the most important files for a component
func (o *Orchestrator) selectKeyFiles(comp analyzer.Component) []llm.FileContent {
	var keyFiles []llm.FileContent
	maxFiles := 20 // Limit files sent to LLM

	// Priority files
	priorityPatterns := []string{
		"main.", "README", "index.", "app.", "server.",
		"routes", "handler", "controller", "api",
	}

	priority := []string{}
	others := []string{}

	for _, file := range comp.Files {
		isPriority := false
		for _, pattern := range priorityPatterns {
			if contains(file, pattern) {
				priority = append(priority, file)
				isPriority = true
				break
			}
		}
		if !isPriority {
			others = append(others, file)
		}
	}

	// Take all priority files
	for _, file := range priority {
		if len(keyFiles) >= maxFiles {
			break
		}
		content, err := readFileContent(file)
		if err == nil {
			keyFiles = append(keyFiles, llm.FileContent{
				Path:    file,
				Content: content,
			})
		}
	}

	// Fill remaining with other files
	for _, file := range others {
		if len(keyFiles) >= maxFiles {
			break
		}
		content, err := readFileContent(file)
		if err == nil {
			keyFiles = append(keyFiles, llm.FileContent{
				Path:    file,
				Content: content,
			})
		}
	}

	return keyFiles
}

// EnrichedComponent contains component with LLM-generated content
type EnrichedComponent struct {
	Component    analyzer.Component
	Overview     string
	DetailedDocs string
	Architecture string
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

func readFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Limit file size
	if len(content) > 50000 {
		content = content[:50000]
	}
	return string(content), nil
}

// getComponentsToGenerate determines which components need regeneration
func (o *Orchestrator) getComponentsToGenerate(structure *analyzer.RepoStructure) []analyzer.Component {
	var components []analyzer.Component

	for _, comp := range structure.Components {
		if o.cacheManager.IsStale(comp.Name, comp.Files) {
			components = append(components, comp)
		}
	}

	return components
}

// buildTemplateData builds the data structure for templates using LLM-generated content
func (o *Orchestrator) buildTemplateData(structure *analyzer.RepoStructure, enriched []EnrichedComponent) template.TemplateData {
	// Use configured attribution or default
	generatedBy := o.config.Documentation.GeneratedBy
	if generatedBy == "" {
		generatedBy = "Generated by DocBrown v1.0.0"
	}

	data := template.TemplateData{
		RepoName:      getRepoName(),
		Description:   "Automatically generated documentation",
		Timestamp:     time.Now(),
		GeneratedBy:   generatedBy,
		Version:       "1.0.0",
		DefaultBranch: "main",
	}

	// Build overview from LLM-generated content
	if len(enriched) > 0 {
		data.Overview = enriched[0].Overview
	}

	// Convert enriched components with LLM-generated content
	for _, ec := range enriched {
		comp := ec.Component

		compData := template.ComponentData{
			Name:         comp.Name,
			Type:         comp.Type,
			Language:     comp.Language,
			Path:         comp.Path,
			Description:  comp.Description,
			Overview:     ec.Overview,
			Architecture: ec.Architecture,
			HasTests:     comp.HasTests,
		}

		data.Components = append(data.Components, compData)

		// Add to services if applicable
		if comp.Type == "service" {
			data.Services = append(data.Services, template.ServiceData{
				Name:        comp.Name,
				Type:        "rest",
				Description: ec.Overview,
			})
		}
	}

	// Build architecture data
	data.Architecture = template.ArchitectureData{
		Overview: "This repository contains " + fmt.Sprintf("%d", len(enriched)) + " components",
	}

	for lang := range structure.Languages {
		data.Architecture.Technologies = append(data.Architecture.Technologies, lang)
	}

	// Generate getting started content
	data.GettingStarted = "Follow the steps below to set up and run this project."

	return data
}

func getRepoName() string {
	// Try to get from git config or directory name
	if dir, err := os.Getwd(); err == nil {
		return filepath.Base(dir)
	}
	return "repository"
}
