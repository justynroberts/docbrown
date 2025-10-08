# 🚀 DocBrown - AI-Powered Documentation Generator

> Automatically generate high-quality, Backstage-compatible documentation from any codebase using LLMs (Claude or local Ollama models).

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

---

## ✨ Features

- 🤖 **AI-Powered** - Uses Claude (Anthropic) or Ollama (local) to generate intelligent documentation
- 🌍 **7 Languages** - Go, Python, JavaScript/TypeScript, Rust, Java, Ruby, C#
- 📦 **Dependency Extraction** - Automatically extracts and documents all dependencies
- 🔍 **Component Detection** - Identifies services, libraries, and frontends
- ✅ **Quality Validation** - Built-in validation with 10-point scoring system
- 🎯 **Backstage Compatible** - Generates Backstage TechDocs ready files
- 🔄 **Incremental Updates** - Smart caching only regenerates what changed
- 🌳 **Git Integration** - Automatic PR creation for GitHub, GitLab, Bitbucket
- 💰 **Cost Tracking** - Monitors LLM API costs (or use Ollama for free!)
- ⚡ **Fast** - Complete docs in under 30 seconds

---

## 📦 Installation

### Option 1: Download Binary (Recommended)

```bash
# macOS/Linux
curl -L https://github.com/justynroberts/docbrown/releases/latest/download/docbrown-$(uname -s)-$(uname -m).tar.gz | tar xz
sudo mv docbrown /usr/local/bin/

# Verify installation
docbrown --version
```

### Option 2: Build from Source

```bash
# Requires Go 1.21+
git clone https://github.com/justynroberts/docbrown.git
cd docbrown
make build

# Or simply
go install github.com/justynroberts/docbrown@latest
```

### Option 3: Using Docker

```bash
docker pull justynroberts/docbrown:latest
docker run -v $(pwd):/workspace justynroberts/docbrown auto
```

---

## 🚀 Quick Start

```bash
# 1. Navigate to your project
cd /path/to/your/project

# 2. Initialize DocBrown
docbrown init

# 3. Generate documentation (complete workflow)
docbrown auto
```

**That's it!** Documentation is now in `docs/` directory.

---

## 📖 Usage

### Basic Commands

```bash
# Initialize configuration
docbrown init

# Analyze repository structure
docbrown analyze

# Generate documentation
docbrown generate

# Complete workflow (analyze + generate + validate)
docbrown auto

# Validate documentation quality
docbrown validate

# Check LLM provider status
docbrown provider status
```

### Advanced Commands

```bash
# Create pull request with docs
docbrown pr

# Push directly to main branch
docbrown pr --push-direct

# Manage configuration
docbrown config show
docbrown config set llm.provider anthropic

# Manage cache
docbrown cache show
docbrown cache clear

# List templates
docbrown templates list
```

---

## ⚙️ Configuration

Create `.docbrown.yaml` in your project root:

```yaml
llm:
  provider: auto  # auto, ollama, anthropic

  ollama:
    endpoint: http://localhost:11434
    model: llama3.2:latest

  anthropic:
    api_key: ""  # or set ANTHROPIC_API_KEY env var
    model: claude-sonnet-4-20250514

documentation:
  output_dir: docs
  template: backstage
  template_path: ""  # optional: custom template directory

cache:
  enabled: true
  ttl: 168h  # 7 days

quality:
  min_score: 7.0
  strict_mode: false
```

---

## 🌍 Supported Languages

| Language | Dependencies | Endpoints | Status |
|----------|--------------|-----------|--------|
| **Go** | `go.mod` | HTTP handlers | ✅ Stable |
| **Python** | `requirements.txt`, `pyproject.toml` | FastAPI, Flask | ✅ Stable |
| **JavaScript/TypeScript** | `package.json` | Express | ✅ Stable |
| **Rust** | `Cargo.toml` | - | ✅ New |
| **Java** | `pom.xml`, `build.gradle` | - | ✅ New |
| **Ruby** | `Gemfile` | Rails routes | ✅ New |
| **C#** | `.csproj` | - | ✅ New |

[See full language support documentation →](SUPPORTED_LANGUAGES.md)

---

## 📊 Example Output

### Console Output

```
DocBrown - Automated Documentation

🔍 Step 1/4: Analyzing codebase...
Found 3 components
  - api-gateway (service, go) - 43 dependencies
  - user-service (service, python) - 28 dependencies
  - frontend (frontend, typescript) - 156 dependencies

🤖 Step 2/4: Generating documentation...
[1/3] Processing: api-gateway
  ✓ Analysis complete (2,500 chars)
  ✓ Documentation complete (5,200 chars)

✅ Step 3/4: Validating quality...
Quality Score: 10.0/10.0

🎉 Step 4/4: Complete

Summary:
  Components processed: 3
  Quality score: 10.0/10.0
  Time: 45s
  Cost: $0.00 (Ollama)
```

### Generated Documentation Structure

```
docs/
├── catalog-info.yaml          # Backstage catalog
├── mkdocs.yml                 # MkDocs configuration
└── docs/
    ├── index.md               # Overview
    ├── api/
    │   └── index.md          # API documentation
    ├── architecture/
    │   └── overview.md       # Architecture docs
    ├── components/
    │   ├── api-gateway.md
    │   ├── user-service.md
    │   └── frontend.md
    └── guides/
        └── getting-started.md
```

[See complete sample output →](SAMPLE_OUTPUT.md)

---

## 🎯 Real-World Examples

### Go Microservice

```bash
cd my-go-service
docbrown auto
```

**Output:**
- ✅ 43 dependencies from `go.mod`
- ✅ 12 HTTP endpoints detected
- ✅ 2,500+ characters of AI-generated docs
- ✅ Quality score: 10.0/10.0

### Python FastAPI Project

```bash
cd my-fastapi-app
docbrown auto
```

**Output:**
- ✅ 28 dependencies from `requirements.txt`
- ✅ 18 API endpoints detected
- ✅ Complete API documentation
- ✅ Quality score: 9.5/10.0

### Multi-Language Monorepo

```bash
my-monorepo/
├── services/
│   ├── api-gateway/        # Go
│   ├── user-service/       # Java
│   └── notification/       # Python
└── libraries/
    ├── shared-models/      # TypeScript
    └── crypto-utils/       # Rust

docbrown auto
# Detects all 5 components across 5 languages!
```

---

## 🤖 LLM Providers

### Ollama (Local, Free)

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull llama3.2

# DocBrown will auto-detect and use it
docbrown auto
```

**Pros:**
- ✅ Completely free
- ✅ Works offline
- ✅ No API keys needed
- ✅ Privacy-first

**Cons:**
- ⚠️ Slower than cloud APIs
- ⚠️ Requires local resources

### Anthropic Claude (Cloud, Paid)

```bash
# Set API key
export ANTHROPIC_API_KEY=sk-ant-...

# DocBrown will use Claude
docbrown auto
```

**Pros:**
- ✅ High-quality output
- ✅ Fast response times
- ✅ Large context windows

**Cons:**
- ⚠️ Costs money (~$0.50/repo)
- ⚠️ Requires internet

---

## ✅ Quality Validation

DocBrown includes comprehensive quality validation:

```bash
docbrown validate
```

**Checks:**
- ✓ Markdown syntax
- ✓ Link validity (no broken links)
- ✓ Backstage catalog schema
- ✓ Coverage (overview, architecture, getting started, API docs)

**Scoring:**
- 9.0-10.0: Excellent ⭐
- 7.0-8.9: Good ✓
- 5.0-6.9: Acceptable ⚠
- 0.0-4.9: Needs Improvement ✗

---

## 🔄 Git Integration

### Create Pull Request

```bash
# Auto-detect platform (GitHub/GitLab/Bitbucket)
export GITHUB_TOKEN=ghp_...
docbrown pr
```

**Output:**
```
✓ Created branch: docs/auto-gen-2025-10-08
✓ Committed changes
✓ Pushed to remote
✓ Created PR #123: "Update documentation"

PR URL: https://github.com/user/repo/pull/123
```

### Push Directly

```bash
# Skip PR, push to main
docbrown pr --push-direct
```

---

## 📈 Performance

| Metric | Value |
|--------|-------|
| **Typical Generation Time** | 20-60 seconds |
| **Cache Hit Rate** | ~90% on subsequent runs |
| **Cost (Ollama)** | $0.00 |
| **Cost (Claude)** | $0.30-$0.80 per repo |
| **Quality Score Average** | 8.5-10.0 |

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Make (optional)

### Build

```bash
# Clone repository
git clone https://github.com/justynroberts/docbrown.git
cd docbrown

# Build
make build

# Test
make test

# Run
./bin/docbrown --help
```

### Project Structure

```
docbrown/
├── cmd/              # CLI commands
├── internal/
│   ├── analyzer/    # Code analysis
│   ├── llm/         # LLM providers
│   ├── template/    # Template engine
│   ├── validator/   # Quality validation
│   ├── git/         # Git operations
│   └── cache/       # Caching system
├── templates/        # Documentation templates
└── docs/            # Project documentation
```

### Custom Templates

DocBrown supports custom documentation templates. You can create your own templates to match your organization's documentation standards.

#### Using Custom Templates

1. **Create template directory:**
```bash
mkdir -p my-templates/custom
```

2. **Create template.yaml:**
```yaml
name: custom
version: 1.0.0
description: My custom documentation template

files:
  - name: index
    template: index.md.tmpl
    output: docs/README.md
    description: Main documentation

  - name: component
    template: component.md.tmpl
    output: docs/components/{{.ComponentName}}.md
    foreach: components
    description: Component documentation

prompts:
  analysis: |
    Analyze this codebase and provide detailed information...

  component: |
    Document this component including architecture and APIs...
```

3. **Create template files** (e.g., `index.md.tmpl`):
```markdown
# {{.RepoName}}

{{.Overview}}

## Components
{{range .Components}}
- **{{.Name}}** ({{.Type}}) - {{.Description}}
{{end}}
```

4. **Configure DocBrown:**
```yaml
documentation:
  template: custom
  template_path: ./my-templates
```

5. **Run DocBrown:**
```bash
docbrown auto
```

#### Template Variables

Available variables in templates:
- `{{.RepoName}}` - Repository name
- `{{.Overview}}` - LLM-generated overview
- `{{.Components}}` - List of detected components
- `{{.Services}}` - List of services
- `{{.Architecture.Overview}}` - Architecture overview
- `{{.Architecture.Technologies}}` - Technology stack
- `{{.RepoURL}}` - Repository URL
- `{{.DefaultBranch}}` - Default branch name
- `{{.Timestamp}}` - Generation timestamp

For components (in `foreach: components`):
- `{{.Name}}` - Component name
- `{{.Type}}` - Component type (service/library/frontend)
- `{{.Language}}` - Programming language
- `{{.Description}}` - Component description
- `{{.Path}}` - Component path
- `{{.Dependencies}}` - Component dependencies

#### Built-in Templates

DocBrown includes three built-in templates:
- **backstage** - Backstage TechDocs compatible (default)
- **mkdocs** - Standard MkDocs format
- **minimal** - Minimal documentation structure

View built-in templates:
```bash
docbrown templates list
docbrown templates show backstage
```

---

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Adding a New Language

1. Add detection logic in `internal/analyzer/detector.go`
2. Add dependency extraction in `internal/analyzer/metadata.go`
3. Add file patterns to config
4. Submit PR!

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **Anthropic** - For Claude API
- **Ollama** - For local LLM support
- **Backstage** - For TechDocs format
- **MkDocs** - For documentation framework

---

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/justynroberts/docbrown/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/justynroberts/docbrown/discussions)
- 📖 **Documentation**: [Full Docs](https://justynroberts.github.io/docbrown)

---

## ⭐ Star History

If you find DocBrown useful, please star the repository!

---

**Built with ❤️ using Go and AI**

*"Where we're going, we don't need manual documentation!"* - Doc Brown (probably)
