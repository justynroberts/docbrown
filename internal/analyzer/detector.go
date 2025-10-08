package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// Detector detects components in a repository
type Detector struct {
	rootPath string
}

// NewDetector creates a new detector
func NewDetector(rootPath string) *Detector {
	return &Detector{
		rootPath: rootPath,
	}
}

// DetectComponents detects all components in the repository
func (d *Detector) DetectComponents(structure *RepoStructure) ([]Component, error) {
	var components []Component

	// Strategy 1: Check for monorepo structure
	components = append(components, d.detectMonorepo()...)

	// Strategy 2: Check for single service/library
	if len(components) == 0 {
		if comp := d.detectSingleComponent(); comp != nil {
			components = append(components, *comp)
		}
	}

	// Strategy 3: Detect by directory structure
	if len(components) == 0 {
		components = append(components, d.detectByDirectory()...)
	}

	// Enrich components with metadata
	for i := range components {
		d.enrichComponent(&components[i], structure)
	}

	return components, nil
}

// detectMonorepo detects components in a monorepo structure
func (d *Detector) detectMonorepo() []Component {
	var components []Component

	// Common monorepo patterns
	patterns := []string{
		"services/*",
		"apps/*",
		"packages/*",
		"libs/*",
	}

	for _, pattern := range patterns {
		dirs, _ := filepath.Glob(filepath.Join(d.rootPath, pattern))
		for _, dir := range dirs {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				if comp := d.analyzeDirectory(dir); comp != nil {
					components = append(components, *comp)
				}
			}
		}
	}

	return components
}

// detectSingleComponent detects a single-component repository
func (d *Detector) detectSingleComponent() *Component {
	comp := &Component{
		Path: ".",
	}

	// Check for Go module
	if d.fileExists("go.mod") {
		comp.Language = "go"
		comp.Type = d.detectGoComponentType()
		comp.Name = d.extractGoModuleName()
		return comp
	}

	// Check for Python package
	if d.fileExists("setup.py") || d.fileExists("pyproject.toml") {
		comp.Language = "python"
		comp.Type = d.detectPythonComponentType()
		comp.Name = d.extractPythonPackageName()
		return comp
	}

	// Check for Node.js package
	if d.fileExists("package.json") {
		comp.Language = "javascript"
		comp.Type = d.detectNodeComponentType()
		comp.Name = d.extractNodePackageName()
		return comp
	}

	// Check for Rust crate
	if d.fileExists("Cargo.toml") {
		comp.Language = "rust"
		comp.Type = "service"
		comp.Name = filepath.Base(d.rootPath)
		return comp
	}

	// Check for Java project
	if d.fileExists("pom.xml") || d.fileExists("build.gradle") || d.fileExists("build.gradle.kts") {
		comp.Language = "java"
		comp.Type = "service"
		comp.Name = filepath.Base(d.rootPath)
		return comp
	}

	// Check for Ruby gem/Rails app
	if d.fileExists("Gemfile") {
		comp.Language = "ruby"
		if d.fileExists("config/routes.rb") {
			comp.Type = "service" // Rails app
		} else {
			comp.Type = "library" // Ruby gem
		}
		comp.Name = filepath.Base(d.rootPath)
		return comp
	}

	// Check for C# project
	if d.dirHasFile(".", "*.csproj") || d.fileExists("*.sln") {
		comp.Language = "csharp"
		comp.Type = "service"
		comp.Name = filepath.Base(d.rootPath)
		return comp
	}

	return nil
}

// detectByDirectory detects components by analyzing directory structure
func (d *Detector) detectByDirectory() []Component {
	var components []Component

	// Look for main entry points
	entryPoints := []string{
		"cmd/*",
		"src/main.*",
		"main.*",
	}

	for _, pattern := range entryPoints {
		matches, _ := filepath.Glob(filepath.Join(d.rootPath, pattern))
		for _, match := range matches {
			if comp := d.analyzeDirectory(filepath.Dir(match)); comp != nil {
				components = append(components, *comp)
			}
		}
	}

	return components
}

// analyzeDirectory analyzes a directory to determine component type
func (d *Detector) analyzeDirectory(dir string) *Component {
	comp := &Component{
		Path: dir,
		Name: filepath.Base(dir),
	}

	// Detect language
	if d.dirHasFile(dir, "*.go") {
		comp.Language = "go"
	} else if d.dirHasFile(dir, "*.py") {
		comp.Language = "python"
	} else if d.dirHasFile(dir, "*.ts") || d.dirHasFile(dir, "*.js") {
		comp.Language = "typescript"
	} else if d.dirHasFile(dir, "*.rs") {
		comp.Language = "rust"
	} else if d.dirHasFile(dir, "*.java") {
		comp.Language = "java"
	} else if d.dirHasFile(dir, "*.rb") {
		comp.Language = "ruby"
	} else if d.dirHasFile(dir, "*.cs") {
		comp.Language = "csharp"
	}

	// Detect type
	if d.dirHasFile(dir, "Dockerfile") {
		comp.Type = "service"
	} else if d.dirHasFile(dir, "package.json") {
		// Check if it's a frontend
		if d.dirContains(dir, "react") || d.dirContains(dir, "vue") || d.dirContains(dir, "angular") {
			comp.Type = "frontend"
		} else {
			comp.Type = "service"
		}
	} else {
		comp.Type = "library"
	}

	return comp
}

// enrichComponent enriches a component with additional metadata
func (d *Detector) enrichComponent(comp *Component, structure *RepoStructure) {
	// Collect files
	comp.Files = d.collectComponentFiles(comp.Path)

	// Check for tests
	comp.HasTests = d.hasTests(comp.Path)

	// Extract dependencies
	comp.Dependencies = d.extractDependencies(comp)

	// Generate description
	if comp.Description == "" {
		comp.Description = d.generateDescription(comp)
	}
}

// collectComponentFiles collects all relevant files for a component
func (d *Detector) collectComponentFiles(path string) []string {
	var files []string

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			// Skip common directories
			base := filepath.Base(filePath)
			if base == "node_modules" || base == "vendor" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include source files
		if isSourceFile(filePath) {
			relPath, _ := filepath.Rel(d.rootPath, filePath)
			files = append(files, relPath)
		}

		return nil
	})

	return files
}

// hasTests checks if a component has tests
func (d *Detector) hasTests(path string) bool {
	hasTests := false

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if isTestFile(filePath) {
			hasTests = true
			return filepath.SkipDir
		}
		return nil
	})

	return hasTests
}

// extractDependencies extracts dependencies for a component
func (d *Detector) extractDependencies(comp *Component) []Dependency {
	var deps []Dependency

	switch comp.Language {
	case "go":
		deps = d.extractGoDependencies(comp.Path)
	case "python":
		deps = d.extractPythonDependencies(comp.Path)
	case "javascript", "typescript":
		deps = d.extractNodeDependencies(comp.Path)
	}

	return deps
}

// Helper functions

func (d *Detector) fileExists(path string) bool {
	fullPath := filepath.Join(d.rootPath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

func (d *Detector) dirHasFile(dir, pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))
	return len(matches) > 0
}

func (d *Detector) dirContains(dir, substring string) bool {
	// Simple check - could be improved
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Name()), substring) {
			return true
		}
	}
	return false
}

func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	sourceExts := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".java", ".rs", ".rb", ".php", ".c", ".cpp",
	}

	for _, sourceExt := range sourceExts {
		if ext == sourceExt {
			return true
		}
	}

	return false
}

// Placeholder detection methods (to be implemented)

func (d *Detector) detectGoComponentType() string {
	if d.fileExists("cmd") || d.fileExists("main.go") {
		return "service"
	}
	return "library"
}

func (d *Detector) detectPythonComponentType() string {
	if d.fileExists("Dockerfile") {
		return "service"
	}
	return "library"
}

func (d *Detector) detectNodeComponentType() string {
	if d.fileExists("Dockerfile") {
		return "service"
	}
	return "library"
}

func (d *Detector) extractGoModuleName() string {
	return filepath.Base(d.rootPath)
}

func (d *Detector) extractPythonPackageName() string {
	return filepath.Base(d.rootPath)
}

func (d *Detector) extractNodePackageName() string {
	return filepath.Base(d.rootPath)
}

func (d *Detector) extractGoDependencies(path string) []Dependency {
	// TODO: Parse go.mod
	return []Dependency{}
}

func (d *Detector) extractPythonDependencies(path string) []Dependency {
	// TODO: Parse requirements.txt or pyproject.toml
	return []Dependency{}
}

func (d *Detector) extractNodeDependencies(path string) []Dependency {
	// TODO: Parse package.json
	return []Dependency{}
}

func (d *Detector) generateDescription(comp *Component) string {
	return "A " + comp.Language + " " + comp.Type + " component"
}
