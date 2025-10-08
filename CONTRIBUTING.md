# Contributing to DocBrown

Thank you for considering contributing to DocBrown! We welcome contributions from the community.

## 🤝 How to Contribute

### Reporting Bugs

- Check [existing issues](https://github.com/justynroberts/docbrown/issues) first
- Use the bug report template
- Include:
  - DocBrown version (`docbrown --version`)
  - Operating system and version
  - Go version (`go version`)
  - Steps to reproduce
  - Expected vs actual behavior
  - Error messages and logs

### Suggesting Features

- Check [existing feature requests](https://github.com/justynroberts/docbrown/issues?q=is%3Aissue+label%3Aenhancement)
- Use the feature request template
- Describe:
  - The problem you're trying to solve
  - Your proposed solution
  - Alternative approaches considered
  - Examples of how it would work

### Submitting Pull Requests

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/justynroberts/docbrown.git
   cd docbrown
   ```

3. **Create a branch**:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

4. **Make your changes**
5. **Test your changes** (see Testing section below)
6. **Commit with clear messages**:
   ```bash
   git commit -m "feat: add support for PHP language"
   # or
   git commit -m "fix: resolve dependency extraction for go.mod"
   ```

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request** from your fork to the main repository

## 🛠️ Development Setup

### Prerequisites

- **Go 1.21 or higher**
- **Git**
- **Make** (optional but recommended)
- **Ollama** (optional, for testing LLM integration)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/justynroberts/docbrown.git
cd docbrown

# Install dependencies
go mod download

# Build the project
make build
# or
go build -o bin/docbrown ./cmd/docbrown

# Run tests
make test
# or
go test ./...
```

### Project Structure

```
docbrown/
├── cmd/
│   └── docbrown/          # CLI entry point
│       └── commands/      # Cobra commands
├── internal/
│   ├── analyzer/          # Code analysis engine
│   ├── llm/              # LLM provider integrations
│   ├── template/         # Template rendering
│   ├── validator/        # Quality validation
│   ├── git/              # Git operations
│   ├── cache/            # Caching system
│   ├── config/           # Configuration management
│   └── orchestrator/     # Workflow coordination
├── templates/            # Documentation templates
├── docs/                 # Project documentation
└── tests/               # Integration tests
```

## 📝 Code Style

### Go Code

- Follow standard Go conventions
- Run `gofmt` and `goimports`
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

**Example:**

```go
// ExtractDependencies extracts dependencies from a component based on language
func (m *MetadataExtractor) ExtractDependencies(comp *Component) []Dependency {
    switch comp.Language {
    case "go":
        return m.extractGoDependencies(comp.Path)
    case "python":
        return m.extractPythonDependencies(comp.Path)
    default:
        return []Dependency{}
    }
}
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions or changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Build process or auxiliary tool changes

**Examples:**
```
feat: add PHP language support
fix: resolve panic when go.mod is empty
docs: update README with new installation methods
test: add unit tests for dependency extraction
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/analyzer/...
```

### Writing Tests

- Add tests for new features
- Maintain or improve test coverage
- Use table-driven tests where appropriate

**Example:**

```go
func TestExtractGoDependencies(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected int
    }{
        {
            name:     "simple go.mod",
            content:  "require github.com/spf13/cobra v1.8.0",
            expected: 1,
        },
        {
            name:     "multi-line require",
            content:  "require (\n\tgithub.com/spf13/cobra v1.8.0\n)",
            expected: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Testing

Test the complete workflow:

```bash
# Initialize test project
cd /tmp/test-project
docbrown init

# Run analysis
docbrown analyze

# Generate documentation
docbrown auto

# Validate output
docbrown validate
```

## 🌍 Adding a New Language

To add support for a new programming language:

### 1. Update Language Detection

Edit `internal/analyzer/detector.go`:

```go
// Check for PHP project
if d.fileExists("composer.json") {
    comp.Language = "php"
    comp.Type = "service"
    comp.Name = filepath.Base(d.rootPath)
    return comp
}
```

### 2. Add Dependency Extraction

Edit `internal/analyzer/metadata.go`:

```go
case "php":
    deps = m.extractPHPDependencies(comp.Path)

// extractPHPDependencies extracts dependencies from composer.json
func (m *MetadataExtractor) extractPHPDependencies(path string) []Dependency {
    var deps []Dependency

    composerPath := filepath.Join(path, "composer.json")
    if path == "." {
        composerPath = "composer.json"
    }

    content, err := os.ReadFile(composerPath)
    if err != nil {
        return deps
    }

    var composer struct {
        Require map[string]string `json:"require"`
    }

    if err := json.Unmarshal(content, &composer); err != nil {
        return deps
    }

    for name, version := range composer.Require {
        deps = append(deps, Dependency{
            Name:    name,
            Version: version,
            Type:    "external",
        })
    }

    return deps
}
```

### 3. Update Configuration

Edit `.docbrown.yaml.example`:

```yaml
documentation:
  include_patterns:
    - '**/*.php'  # Add PHP files
```

### 4. Add Tests

Create tests in `internal/analyzer/metadata_test.go`:

```go
func TestExtractPHPDependencies(t *testing.T) {
    // Test implementation
}
```

### 5. Update Documentation

- Add language to `SUPPORTED_LANGUAGES.md`
- Update `README.md` language table
- Add example in documentation

### 6. Submit PR

- Test with real PHP projects
- Update CHANGELOG
- Submit pull request with description

## 🎯 Good First Issues

Look for issues labeled `good first issue`:
- Documentation improvements
- Test coverage additions
- Minor bug fixes
- Code cleanup and refactoring

## 📋 Pull Request Checklist

Before submitting your PR, ensure:

- [ ] Code follows project style guidelines
- [ ] All tests pass (`go test ./...`)
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] PR description clearly explains changes
- [ ] No breaking changes (or clearly documented)
- [ ] Branch is up to date with main

## 🔍 Code Review Process

1. **Automated Checks**: CI runs tests and linting
2. **Maintainer Review**: A maintainer reviews code
3. **Feedback**: Address any requested changes
4. **Approval**: Maintainer approves PR
5. **Merge**: PR is merged to main

## 📚 Resources

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Backstage TechDocs](https://backstage.io/docs/features/techdocs/)
- [MkDocs Documentation](https://www.mkdocs.org/)

## 💬 Communication

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Requests**: Code contributions

## 📜 License

By contributing to DocBrown, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to DocBrown!** 🚀
