package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scanner scans the repository file structure
type Scanner struct {
	rootPath        string
	excludePatterns []string
}

// NewScanner creates a new scanner
func NewScanner(rootPath string, excludePatterns []string) *Scanner {
	return &Scanner{
		rootPath:        rootPath,
		excludePatterns: excludePatterns,
	}
}

// Scan scans the repository and returns the structure
func (s *Scanner) Scan() (*RepoStructure, error) {
	structure := &RepoStructure{
		RootPath:  s.rootPath,
		Languages: make(map[string]int),
	}

	var files []FileInfo

	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded paths
		if s.shouldExclude(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process file
		relPath, _ := filepath.Rel(s.rootPath, path)
		language := detectLanguage(path)
		isTest := isTestFile(path)

		fileInfo := FileInfo{
			Path:     relPath,
			Size:     info.Size(),
			Language: language,
			IsTest:   isTest,
		}

		files = append(files, fileInfo)

		// Count languages
		if language != "" {
			structure.Languages[language]++
		}

		structure.TotalFiles++

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}

	// Generate file tree
	structure.FileTree = s.generateFileTree(files)

	return structure, nil
}

// shouldExclude checks if a path should be excluded
func (s *Scanner) shouldExclude(path string) bool {
	relPath, _ := filepath.Rel(s.rootPath, path)

	// Always exclude
	alwaysExclude := []string{
		".git",
		".docbrown",
		"node_modules",
		"vendor",
		".venv",
		"venv",
		"__pycache__",
		"dist",
		"build",
		"target",
		".next",
		".cache",
	}

	for _, pattern := range alwaysExclude {
		if strings.Contains(relPath, pattern) {
			return true
		}
	}

	// Check configured patterns
	for _, pattern := range s.excludePatterns {
		matched, _ := filepath.Match(pattern, relPath)
		if matched {
			return true
		}
	}

	return false
}

// generateFileTree generates a textual representation of the file tree
func (s *Scanner) generateFileTree(files []FileInfo) string {
	var sb strings.Builder

	// Group files by directory
	dirs := make(map[string][]string)

	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if dir == "." {
			dir = ""
		}
		dirs[dir] = append(dirs[dir], filepath.Base(file.Path))
	}

	// Build tree structure
	sb.WriteString(".\n")

	var printDir func(dir string, indent string)
	printDir = func(dir string, indent string) {
		if files, ok := dirs[dir]; ok {
			for _, file := range files {
				sb.WriteString(fmt.Sprintf("%s├── %s\n", indent, file))
			}
		}

		// Print subdirectories
		for d := range dirs {
			if strings.HasPrefix(d, dir) && d != dir {
				parts := strings.Split(strings.TrimPrefix(d, dir), string(filepath.Separator))
				if len(parts) > 0 && parts[0] != "" {
					sb.WriteString(fmt.Sprintf("%s├── %s/\n", indent, parts[0]))
					printDir(d, indent+"│   ")
				}
			}
		}
	}

	printDir("", "")

	return sb.String()
}

// detectLanguage detects the programming language from file extension
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	languageMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascript",
		".tsx":  "typescript",
		".java": "java",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".c":    "c",
		".cpp":  "cpp",
		".cs":   "csharp",
		".kt":   "kotlin",
		".swift": "swift",
		".sh":   "shell",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	return ""
}

// isTestFile checks if a file is a test file
func isTestFile(path string) bool {
	base := filepath.Base(path)

	testPatterns := []string{
		"_test.go",
		"_test.py",
		".test.js",
		".test.ts",
		".spec.js",
		".spec.ts",
		"Test.java",
	}

	for _, pattern := range testPatterns {
		if strings.HasSuffix(base, pattern) {
			return true
		}
	}

	// Check if in test directory
	return strings.Contains(path, "/test/") || strings.Contains(path, "/tests/")
}
