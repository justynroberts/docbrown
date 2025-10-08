package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MetadataExtractor extracts metadata from project files
type MetadataExtractor struct {
	rootPath string
}

// NewMetadataExtractor creates a new metadata extractor
func NewMetadataExtractor(rootPath string) *MetadataExtractor {
	return &MetadataExtractor{
		rootPath: rootPath,
	}
}

// ExtractDependencies extracts dependencies for a component
func (m *MetadataExtractor) ExtractDependencies(comp *Component) []Dependency {
	var deps []Dependency

	switch comp.Language {
	case "go":
		deps = m.extractGoDependencies(comp.Path)
	case "python":
		deps = m.extractPythonDependencies(comp.Path)
	case "javascript", "typescript":
		deps = m.extractNodeDependencies(comp.Path)
	case "rust":
		deps = m.extractRustDependencies(comp.Path)
	case "java":
		deps = m.extractJavaDependencies(comp.Path)
	case "ruby":
		deps = m.extractRubyDependencies(comp.Path)
	case "csharp":
		deps = m.extractCSharpDependencies(comp.Path)
	}

	return deps
}

// extractGoDependencies extracts dependencies from go.mod
func (m *MetadataExtractor) extractGoDependencies(path string) []Dependency {
	var deps []Dependency

	goModPath := filepath.Join(path, "go.mod")
	if path == "." {
		goModPath = "go.mod"
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return deps
	}

	lines := strings.Split(string(content), "\n")
	inRequire := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}

		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		if strings.HasPrefix(line, "require ") || inRequire {
			// Parse require line: "github.com/foo/bar v1.2.3"
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]

				// Include all dependencies (both direct and indirect)
				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
					Type:    "external",
				})
			}
		}
	}

	return deps
}

// extractPythonDependencies extracts dependencies from requirements.txt or pyproject.toml
func (m *MetadataExtractor) extractPythonDependencies(path string) []Dependency {
	var deps []Dependency

	// Try requirements.txt first
	reqPath := filepath.Join(path, "requirements.txt")
	if path == "." {
		reqPath = "requirements.txt"
	}

	content, err := os.ReadFile(reqPath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Skip comments and empty lines
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse line: "package==1.2.3" or "package>=1.2.3"
			re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)([>=<~!]+)?(.*)$`)
			matches := re.FindStringSubmatch(line)

			if len(matches) >= 2 {
				name := matches[1]
				version := ""
				if len(matches) >= 4 {
					version = matches[3]
				}

				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
					Type:    "external",
				})
			}
		}
		return deps
	}

	// Try pyproject.toml
	pyprojectPath := filepath.Join(path, "pyproject.toml")
	if path == "." {
		pyprojectPath = "pyproject.toml"
	}

	content, err = os.ReadFile(pyprojectPath)
	if err != nil {
		return deps
	}

	// Simple parsing for dependencies section
	lines := strings.Split(string(content), "\n")
	inDeps := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "[tool.poetry.dependencies]") || strings.Contains(line, "[project.dependencies]") {
			inDeps = true
			continue
		}

		if inDeps && strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}

		if inDeps && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
					Type:    "external",
				})
			}
		}
	}

	return deps
}

// extractNodeDependencies extracts dependencies from package.json
func (m *MetadataExtractor) extractNodeDependencies(path string) []Dependency {
	var deps []Dependency

	pkgPath := filepath.Join(path, "package.json")
	if path == "." {
		pkgPath = "package.json"
	}

	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return deps
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return deps
	}

	// Add regular dependencies
	for name, version := range pkg.Dependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: version,
			Type:    "external",
		})
	}

	// Don't include dev dependencies by default
	// Could make this configurable

	return deps
}

// ExtractMetadata extracts additional metadata like ports, endpoints, etc.
func (m *MetadataExtractor) ExtractMetadata(comp *Component) {
	// Try to find port configurations
	comp.Endpoints = m.extractEndpoints(comp)
}

// extractEndpoints attempts to find API endpoints in the code
func (m *MetadataExtractor) extractEndpoints(comp *Component) []Endpoint {
	var endpoints []Endpoint

	// This is a simplified version - could be much more sophisticated
	for _, file := range comp.Files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		text := string(content)

		// Look for common route patterns
		patterns := []string{
			// Go patterns
			`\.Handle\("([^"]+)"`,
			`\.HandleFunc\("([^"]+)"`,
			`router\.([A-Z]+)\("([^"]+)"`,
			// Express.js patterns
			`app\.get\("([^"]+)"`,
			`app\.post\("([^"]+)"`,
			`app\.put\("([^"]+)"`,
			`app\.delete\("([^"]+)"`,
			// FastAPI patterns
			`@app\.get\("([^"]+)"\)`,
			`@app\.post\("([^"]+)"\)`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatch(text, -1)

			for _, match := range matches {
				if len(match) >= 2 {
					path := match[1]
					method := "GET"

					// Try to determine method
					if strings.Contains(match[0], "post") || strings.Contains(match[0], "POST") {
						method = "POST"
					} else if strings.Contains(match[0], "put") || strings.Contains(match[0], "PUT") {
						method = "PUT"
					} else if strings.Contains(match[0], "delete") || strings.Contains(match[0], "DELETE") {
						method = "DELETE"
					}

					endpoints = append(endpoints, Endpoint{
						Method:      method,
						Path:        path,
						Description: fmt.Sprintf("%s endpoint", method),
					})
				}
			}
		}
	}

	return endpoints
}

// extractRustDependencies extracts dependencies from Cargo.toml
func (m *MetadataExtractor) extractRustDependencies(path string) []Dependency {
	var deps []Dependency

	cargoPath := filepath.Join(path, "Cargo.toml")
	if path == "." {
		cargoPath = "Cargo.toml"
	}

	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return deps
	}

	lines := strings.Split(string(content), "\n")
	inDeps := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "[dependencies]") {
			inDeps = true
			continue
		}

		if inDeps && strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}

		if inDeps && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
					Type:    "external",
				})
			}
		}
	}

	return deps
}

// extractJavaDependencies extracts dependencies from pom.xml or build.gradle
func (m *MetadataExtractor) extractJavaDependencies(path string) []Dependency {
	var deps []Dependency

	// Try Maven pom.xml first
	pomPath := filepath.Join(path, "pom.xml")
	if path == "." {
		pomPath = "pom.xml"
	}

	content, err := os.ReadFile(pomPath)
	if err == nil {
		// Simple XML parsing for <dependency> tags
		text := string(content)
		groupRe := regexp.MustCompile(`<groupId>([^<]+)</groupId>`)
		artifactRe := regexp.MustCompile(`<artifactId>([^<]+)</artifactId>`)
		versionRe := regexp.MustCompile(`<version>([^<]+)</version>`)

		groups := groupRe.FindAllStringSubmatch(text, -1)
		artifacts := artifactRe.FindAllStringSubmatch(text, -1)
		versions := versionRe.FindAllStringSubmatch(text, -1)

		// Match up dependencies (simplified - assumes order matches)
		for i := 0; i < len(groups) && i < len(artifacts); i++ {
			version := ""
			if i < len(versions) {
				version = versions[i][1]
			}
			deps = append(deps, Dependency{
				Name:    groups[i][1] + ":" + artifacts[i][1],
				Version: version,
				Type:    "external",
			})
		}
		return deps
	}

	// Try Gradle build.gradle
	gradlePath := filepath.Join(path, "build.gradle")
	if path == "." {
		gradlePath = "build.gradle"
	}

	content, err = os.ReadFile(gradlePath)
	if err != nil {
		return deps
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Match: implementation 'group:artifact:version'
		if strings.Contains(line, "implementation") || strings.Contains(line, "compile") {
			re := regexp.MustCompile(`['"]([^:]+):([^:]+):([^'"]+)['"]`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 4 {
				deps = append(deps, Dependency{
					Name:    matches[1] + ":" + matches[2],
					Version: matches[3],
					Type:    "external",
				})
			}
		}
	}

	return deps
}

// extractRubyDependencies extracts dependencies from Gemfile
func (m *MetadataExtractor) extractRubyDependencies(path string) []Dependency {
	var deps []Dependency

	gemfilePath := filepath.Join(path, "Gemfile")
	if path == "." {
		gemfilePath = "Gemfile"
	}

	content, err := os.ReadFile(gemfilePath)
	if err != nil {
		return deps
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Match: gem 'name', 'version'
		if strings.HasPrefix(line, "gem ") {
			re := regexp.MustCompile(`gem\s+['"]([^'"]+)['"]\s*,?\s*['"]?([^'"]*)['"]?`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				name := matches[1]
				version := ""
				if len(matches) >= 3 {
					version = matches[2]
				}

				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
					Type:    "external",
				})
			}
		}
	}

	return deps
}

// extractCSharpDependencies extracts dependencies from .csproj files
func (m *MetadataExtractor) extractCSharpDependencies(path string) []Dependency {
	var deps []Dependency

	// Find .csproj files
	files, err := os.ReadDir(path)
	if err != nil {
		return deps
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csproj") {
			csprojPath := filepath.Join(path, file.Name())
			content, err := os.ReadFile(csprojPath)
			if err != nil {
				continue
			}

			// Parse <PackageReference Include="Name" Version="1.0.0" />
			text := string(content)
			re := regexp.MustCompile(`<PackageReference\s+Include="([^"]+)"\s+Version="([^"]+)"`)
			matches := re.FindAllStringSubmatch(text, -1)

			for _, match := range matches {
				if len(match) >= 3 {
					deps = append(deps, Dependency{
						Name:    match[1],
						Version: match[2],
						Type:    "external",
					})
				}
			}
		}
	}

	return deps
}
