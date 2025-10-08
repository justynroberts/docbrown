package analyzer

import (
	"fmt"
)

// Analyzer is the main analyzer that orchestrates scanning and detection
type Analyzer struct {
	rootPath        string
	excludePatterns []string
	scanner         *Scanner
	detector        *Detector
	metadata        *MetadataExtractor
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(rootPath string, excludePatterns []string) *Analyzer {
	return &Analyzer{
		rootPath:        rootPath,
		excludePatterns: excludePatterns,
		scanner:         NewScanner(rootPath, excludePatterns),
		detector:        NewDetector(rootPath),
		metadata:        NewMetadataExtractor(rootPath),
	}
}

// Analyze performs a full analysis of the repository
func (a *Analyzer) Analyze() (*RepoStructure, error) {
	// Step 1: Scan the repository
	fmt.Println("Scanning repository...")
	structure, err := a.scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Step 2: Detect components
	fmt.Println("Detecting components...")
	components, err := a.detector.DetectComponents(structure)
	if err != nil {
		return nil, fmt.Errorf("component detection failed: %w", err)
	}

	// Step 3: Extract metadata for each component
	fmt.Println("Extracting metadata...")
	for i := range components {
		// Extract dependencies
		components[i].Dependencies = a.metadata.ExtractDependencies(&components[i])

		// Extract additional metadata (endpoints, etc.)
		a.metadata.ExtractMetadata(&components[i])
	}

	structure.Components = components

	fmt.Printf("Found %d components\n", len(components))
	for _, comp := range components {
		fmt.Printf("  - %s (%s, %s) - %d dependencies, %d endpoints\n",
			comp.Name, comp.Type, comp.Language, len(comp.Dependencies), len(comp.Endpoints))
	}

	return structure, nil
}
