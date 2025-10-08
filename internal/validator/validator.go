package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Validator validates documentation quality
type Validator struct {
	strictMode bool
	docsDir    string
}

// NewValidator creates a new validator
func NewValidator(docsDir string, strictMode bool) *Validator {
	return &Validator{
		docsDir:    docsDir,
		strictMode: strictMode,
	}
}

// ValidationResults contains validation results
type ValidationResults struct {
	MarkdownErrors    []ValidationError
	BrokenLinks       []BrokenLink
	CatalogValid      bool
	CatalogError      string
	HasOverview       bool
	HasAPIDocs        bool
	HasArchitecture   bool
	HasGettingStarted bool
	QualityScore      float64
}

// ValidationError represents a validation error
type ValidationError struct {
	File    string
	Line    int
	Type    string
	Message string
}

// BrokenLink represents a broken link
type BrokenLink struct {
	Source string
	Target string
	Line   int
}

// Validate performs comprehensive validation
func (v *Validator) Validate() (*ValidationResults, error) {
	results := &ValidationResults{}

	// Find all markdown files
	mdFiles, err := v.findMarkdownFiles()
	if err != nil {
		return nil, err
	}

	// Validate markdown syntax
	for _, file := range mdFiles {
		errors := v.validateMarkdownFile(file)
		results.MarkdownErrors = append(results.MarkdownErrors, errors...)
	}

	// Check for broken links
	results.BrokenLinks = v.checkLinks(mdFiles)

	// Check for required files (Backstage/MkDocs uses docs/docs/ structure)
	docsSubdir := filepath.Join(v.docsDir, "docs")
	results.HasOverview = v.fileExists(filepath.Join(docsSubdir, "index.md"))
	results.HasArchitecture = v.fileExists(filepath.Join(docsSubdir, "architecture", "overview.md"))
	results.HasGettingStarted = v.fileExists(filepath.Join(docsSubdir, "guides", "getting-started.md"))
	results.HasAPIDocs = v.hasAPIFiles()

	// Validate Backstage catalog
	results.CatalogValid, results.CatalogError = v.validateCatalog()

	// Calculate quality score
	results.QualityScore = v.calculateQualityScore(results)

	return results, nil
}

// findMarkdownFiles finds all markdown files in docs directory
func (v *Validator) findMarkdownFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(v.docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// validateMarkdownFile validates a single markdown file
func (v *Validator) validateMarkdownFile(path string) []ValidationError {
	var errors []ValidationError

	content, err := os.ReadFile(path)
	if err != nil {
		errors = append(errors, ValidationError{
			File:    path,
			Type:    "read-error",
			Message: err.Error(),
		})
		return errors
	}

	lines := strings.Split(string(content), "\n")

	// Check for unclosed code blocks
	inCodeBlock := false
	codeBlockStart := 0

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inCodeBlock {
				inCodeBlock = false
			} else {
				inCodeBlock = true
				codeBlockStart = i + 1
			}
		}
	}

	if inCodeBlock {
		errors = append(errors, ValidationError{
			File:    path,
			Line:    codeBlockStart,
			Type:    "unclosed-code-block",
			Message: "Unclosed code block",
		})
	}

	// Check heading hierarchy (skip lines in code blocks)
	lastLevel := 0
	inCodeBlock2 := false
	for i, line := range lines {
		// Track code blocks for heading check
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock2 = !inCodeBlock2
			continue
		}

		// Skip headings inside code blocks
		if inCodeBlock2 {
			continue
		}

		if strings.HasPrefix(line, "#") {
			level := 0
			for _, ch := range line {
				if ch == '#' {
					level++
				} else {
					break
				}
			}

			if level > lastLevel+1 && lastLevel > 0 {
				errors = append(errors, ValidationError{
					File:    path,
					Line:    i + 1,
					Type:    "heading-skip",
					Message: fmt.Sprintf("Heading level skip from h%d to h%d", lastLevel, level),
				})
			}

			lastLevel = level
		}
	}

	return errors
}

// checkLinks checks for broken internal links
func (v *Validator) checkLinks(files []string) []BrokenLink {
	var broken []BrokenLink

	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			matches := linkRe.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 2 {
					target := match[2]

					// Skip external links
					if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
						continue
					}

					// Check if target exists
					targetPath := filepath.Join(filepath.Dir(file), target)

					// Handle anchors
					if strings.Contains(target, "#") {
						parts := strings.Split(target, "#")
						targetPath = filepath.Join(filepath.Dir(file), parts[0])
					}

					if !v.fileExists(targetPath) {
						broken = append(broken, BrokenLink{
							Source: file,
							Target: target,
							Line:   i + 1,
						})
					}
				}
			}
		}
	}

	return broken
}

// validateCatalog validates the Backstage catalog file
func (v *Validator) validateCatalog() (bool, string) {
	// Check in docs directory first (standard location)
	catalogPath := filepath.Join(v.docsDir, "catalog-info.yaml")

	// Fall back to root if not found in docs
	if !v.fileExists(catalogPath) {
		catalogPath = "catalog-info.yaml"
	}

	if !v.fileExists(catalogPath) {
		return false, "catalog-info.yaml not found in docs/ or root"
	}

	content, err := os.ReadFile(catalogPath)
	if err != nil {
		return false, fmt.Sprintf("failed to read catalog: %v", err)
	}

	var catalog map[string]interface{}
	if err := yaml.Unmarshal(content, &catalog); err != nil {
		return false, fmt.Sprintf("invalid YAML: %v", err)
	}

	// Check required fields
	if _, ok := catalog["apiVersion"]; !ok {
		return false, "missing apiVersion"
	}

	if _, ok := catalog["kind"]; !ok {
		return false, "missing kind"
	}

	metadata, ok := catalog["metadata"].(map[string]interface{})
	if !ok {
		return false, "missing metadata"
	}

	if _, ok := metadata["name"]; !ok {
		return false, "missing metadata.name"
	}

	return true, ""
}

// hasAPIFiles checks if API documentation exists
func (v *Validator) hasAPIFiles() bool {
	// Check in docs/docs/api (Backstage/MkDocs structure)
	apiDir := filepath.Join(v.docsDir, "docs", "api")
	info, err := os.Stat(apiDir)
	if err == nil && info.IsDir() {
		return true
	}

	// Fall back to docs/api
	apiDir = filepath.Join(v.docsDir, "api")
	info, err = os.Stat(apiDir)
	return err == nil && info.IsDir()
}

// fileExists checks if a file exists
func (v *Validator) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// calculateQualityScore calculates a quality score
func (v *Validator) calculateQualityScore(results *ValidationResults) float64 {
	score := 0.0
	maxScore := 10.0

	// Markdown syntax (2 points)
	if len(results.MarkdownErrors) == 0 {
		score += 2.0
	} else {
		score += 2.0 * (1.0 - float64(len(results.MarkdownErrors))/10.0)
	}

	// Links valid (1.5 points)
	if len(results.BrokenLinks) == 0 {
		score += 1.5
	} else {
		score += 1.5 * (1.0 - float64(len(results.BrokenLinks))/5.0)
	}

	// Has overview (1 point)
	if results.HasOverview {
		score += 1.0
	}

	// Has API docs (1.5 points)
	if results.HasAPIDocs {
		score += 1.5
	}

	// Has architecture (1 point)
	if results.HasArchitecture {
		score += 1.0
	}

	// Has getting started (1 point)
	if results.HasGettingStarted {
		score += 1.0
	}

	// Catalog valid (2 points)
	if results.CatalogValid {
		score += 2.0
	}

	if score < 0 {
		score = 0
	}
	if score > maxScore {
		score = maxScore
	}

	return score
}

// FormatResults formats validation results as a string
func (v *Validator) FormatResults(results *ValidationResults) string {
	var sb strings.Builder

	sb.WriteString("Validation Results:\n\n")

	// Markdown errors
	if len(results.MarkdownErrors) == 0 {
		sb.WriteString("✓ Markdown syntax valid\n")
	} else {
		sb.WriteString(fmt.Sprintf("✗ %d markdown issues:\n", len(results.MarkdownErrors)))
		for _, err := range results.MarkdownErrors {
			sb.WriteString(fmt.Sprintf("  %s:%d - %s\n", err.File, err.Line, err.Message))
		}
	}

	// Links
	if len(results.BrokenLinks) == 0 {
		sb.WriteString("✓ All links valid\n")
	} else {
		sb.WriteString(fmt.Sprintf("✗ %d broken links:\n", len(results.BrokenLinks)))
		for _, link := range results.BrokenLinks {
			sb.WriteString(fmt.Sprintf("  %s:%d - %s\n", link.Source, link.Line, link.Target))
		}
	}

	// Catalog
	if results.CatalogValid {
		sb.WriteString("✓ Backstage catalog valid\n")
	} else {
		sb.WriteString(fmt.Sprintf("✗ Backstage catalog invalid: %s\n", results.CatalogError))
	}

	// Coverage
	sb.WriteString("\nCoverage:\n")
	sb.WriteString(fmt.Sprintf("  Overview: %s\n", boolCheck(results.HasOverview)))
	sb.WriteString(fmt.Sprintf("  API Docs: %s\n", boolCheck(results.HasAPIDocs)))
	sb.WriteString(fmt.Sprintf("  Architecture: %s\n", boolCheck(results.HasArchitecture)))
	sb.WriteString(fmt.Sprintf("  Getting Started: %s\n", boolCheck(results.HasGettingStarted)))

	// Score
	sb.WriteString(fmt.Sprintf("\nQuality Score: %.1f/10.0\n", results.QualityScore))

	if results.QualityScore >= 9.0 {
		sb.WriteString("Grade: Excellent ⭐\n")
	} else if results.QualityScore >= 7.0 {
		sb.WriteString("Grade: Good ✓\n")
	} else if results.QualityScore >= 5.0 {
		sb.WriteString("Grade: Acceptable ⚠\n")
	} else {
		sb.WriteString("Grade: Needs Improvement ✗\n")
	}

	return sb.String()
}

func boolCheck(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}
