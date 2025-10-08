# 🌍 DocBrown - Supported Languages

DocBrown now supports **7 programming languages** with full dependency extraction and LLM-powered documentation generation!

---

## ✅ Currently Supported Languages

### 1. **Go** 🐹
```yaml
Detection:
  - File extension: *.go
  - Project file: go.mod
  
Dependency Extraction:
  - Source: go.mod
  - Format: go module syntax
  - Example: github.com/spf13/cobra v1.8.0
  
Endpoint Detection:
  - .HandleFunc()
  - .Handle()
  - router.GET/POST/PUT/DELETE()
```

**Example:**
```bash
cd /path/to/go-project
docbrown auto
# Extracts from go.mod, detects HTTP handlers
```

---

### 2. **Python** 🐍
```yaml
Detection:
  - File extension: *.py
  - Project files: setup.py, pyproject.toml, requirements.txt
  
Dependency Extraction:
  - Sources: requirements.txt, pyproject.toml
  - Formats: pip requirements, poetry dependencies
  - Example: fastapi>=0.100.0
  
Endpoint Detection:
  - @app.get()
  - @app.post()
  - @app.route() (Flask)
```

**Example:**
```bash
cd /path/to/python-project
docbrown auto
# Extracts from requirements.txt or pyproject.toml
```

---

### 3. **JavaScript/TypeScript** 📜
```yaml
Detection:
  - File extensions: *.js, *.ts, *.tsx, *.jsx
  - Project file: package.json
  
Dependency Extraction:
  - Source: package.json
  - Format: npm/yarn dependencies
  - Example: express: ^4.18.0
  
Endpoint Detection:
  - app.get()
  - app.post()
  - router.use() (Express)
```

**Example:**
```bash
cd /path/to/node-project
docbrown auto
# Extracts from package.json
```

---

### 4. **Rust** 🦀 **NEW!**
```yaml
Detection:
  - File extension: *.rs
  - Project file: Cargo.toml
  
Dependency Extraction:
  - Source: Cargo.toml
  - Format: [dependencies] section
  - Example: tokio = "1.35.0"
  
Implementation:
  ✅ Parses [dependencies] section
  ✅ Extracts crate names and versions
  ✅ Supports version specifications
```

**Example:**
```bash
cd /path/to/rust-project
docbrown auto
# Extracts from Cargo.toml
```

**Sample Cargo.toml:**
```toml
[dependencies]
tokio = "1.35.0"
serde = { version = "1.0", features = ["derive"] }
```

---

### 5. **Java** ☕ **NEW!**
```yaml
Detection:
  - File extension: *.java
  - Project files: pom.xml, build.gradle, build.gradle.kts
  
Dependency Extraction:
  - Sources: pom.xml (Maven), build.gradle (Gradle)
  - Formats: XML (Maven), Groovy/Kotlin DSL (Gradle)
  - Example: org.springframework.boot:spring-boot-starter-web
  
Implementation:
  ✅ Parses Maven pom.xml <dependency> tags
  ✅ Parses Gradle implementation/compile statements
  ✅ Supports both Maven and Gradle projects
```

**Example:**
```bash
cd /path/to/java-project
docbrown auto
# Extracts from pom.xml or build.gradle
```

**Sample pom.xml:**
```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-web</artifactId>
    <version>3.2.0</version>
</dependency>
```

**Sample build.gradle:**
```gradle
implementation 'org.springframework.boot:spring-boot-starter-web:3.2.0'
```

---

### 6. **Ruby** 💎 **NEW!**
```yaml
Detection:
  - File extension: *.rb
  - Project file: Gemfile
  - Rails detection: config/routes.rb
  
Dependency Extraction:
  - Source: Gemfile
  - Format: gem statements
  - Example: gem 'rails', '~> 7.0'
  
Implementation:
  ✅ Parses gem 'name', 'version' statements
  ✅ Detects Rails applications
  ✅ Distinguishes between services and libraries
```

**Example:**
```bash
cd /path/to/ruby-project
docbrown auto
# Extracts from Gemfile
# Detects if it's a Rails app
```

**Sample Gemfile:**
```ruby
gem 'rails', '~> 7.0'
gem 'pg', '~> 1.1'
gem 'puma', '~> 5.0'
```

---

### 7. **C#** 🔷 **NEW!**
```yaml
Detection:
  - File extension: *.cs
  - Project files: *.csproj, *.sln
  
Dependency Extraction:
  - Source: .csproj files
  - Format: PackageReference XML elements
  - Example: Microsoft.AspNetCore.App
  
Implementation:
  ✅ Finds .csproj files in project directory
  ✅ Parses <PackageReference> elements
  ✅ Extracts NuGet package names and versions
```

**Example:**
```bash
cd /path/to/csharp-project
docbrown auto
# Extracts from *.csproj files
```

**Sample .csproj:**
```xml
<PackageReference Include="Microsoft.AspNetCore.App" Version="2.1.0" />
<PackageReference Include="Newtonsoft.Json" Version="13.0.3" />
```

---

## 📊 Feature Comparison

| Language | Dependency Extraction | Endpoint Detection | LLM Analysis | Status |
|----------|----------------------|-------------------|--------------|---------|
| **Go** | ✅ go.mod | ✅ HTTP handlers | ✅ | Stable |
| **Python** | ✅ requirements.txt, pyproject.toml | ✅ FastAPI, Flask | ✅ | Stable |
| **JavaScript/TypeScript** | ✅ package.json | ✅ Express | ✅ | Stable |
| **Rust** | ✅ Cargo.toml | 🔄 Planned | ✅ | **NEW** |
| **Java** | ✅ pom.xml, build.gradle | 🔄 Planned | ✅ | **NEW** |
| **Ruby** | ✅ Gemfile | 🔄 Planned | ✅ | **NEW** |
| **C#** | ✅ .csproj | 🔄 Planned | ✅ | **NEW** |

---

## 🚀 Usage Examples

### Multi-Language Monorepo

```bash
my-monorepo/
├── services/
│   ├── api-gateway/        # Go
│   ├── user-service/       # Java (Spring Boot)
│   └── notification/       # Python (FastAPI)
├── libraries/
│   ├── shared-models/      # TypeScript
│   └── crypto-utils/       # Rust
└── workers/
    └── email-worker/       # Ruby (Sidekiq)

# DocBrown will detect all 6 languages!
cd my-monorepo
docbrown auto
```

**Output:**
```
Found 6 components
  - api-gateway (service, go) - 43 dependencies
  - user-service (service, java) - 28 dependencies
  - notification (service, python) - 15 dependencies
  - shared-models (library, typescript) - 12 dependencies
  - crypto-utils (library, rust) - 8 dependencies
  - email-worker (service, ruby) - 22 dependencies
```

---

## 🔧 Configuration

All languages are automatically included. To customize:

```yaml
# .docbrown.yaml
documentation:
  include_patterns:
    - '**/*.go'
    - '**/*.py'
    - '**/*.ts'
    - '**/*.js'
    - '**/*.java'
    - '**/*.rs'
    - '**/*.rb'
    - '**/*.cs'
  
  # Exclude test files
  exclude_patterns:
    - '**/test/**'
    - '**/*_test.go'
    - '**/*_test.py'
    - '**/src/test/**'    # Java
    - '**/spec/**'        # Ruby
```

---

## 📈 What Gets Extracted

For each language, DocBrown extracts:

### ✅ Always Extracted
- **Component name**
- **Language detection**
- **Component type** (service/library/frontend)
- **File structure**
- **Dependencies with versions**

### 🤖 LLM-Generated
- **Overview** - AI-written description
- **Architecture** - Component design
- **Usage examples**
- **API documentation**
- **Configuration details**

---

## 🎯 Coming Soon

### Planned Language Support
- [ ] **PHP** - composer.json
- [ ] **Kotlin** - build.gradle.kts
- [ ] **Swift** - Package.swift
- [ ] **Scala** - build.sbt
- [ ] **Elixir** - mix.exs

### Planned Features
- [ ] Endpoint detection for all languages
- [ ] Database schema extraction
- [ ] Test coverage reporting
- [ ] Performance metrics

---

## 🛠️ Adding Your Own Language

DocBrown is designed to be extensible. To add a new language:

1. **Add detection** in `internal/analyzer/detector.go`
2. **Add dependency extraction** in `internal/analyzer/metadata.go`
3. **Add file patterns** in `.docbrown.yaml`

**Example: Adding PHP**

```go
// In detector.go
if d.fileExists("composer.json") {
    comp.Language = "php"
    comp.Type = "service"
    return comp
}

// In metadata.go
case "php":
    deps = m.extractPHPDependencies(comp.Path)

func (m *MetadataExtractor) extractPHPDependencies(path string) []Dependency {
    // Parse composer.json
}
```

---

## 📊 Real-World Examples

### Go Service
```
✓ Extracted 43 dependencies from go.mod
✓ Found 12 HTTP endpoints
✓ Generated 2,500 chars of documentation
✓ Quality score: 10.0/10.0
```

### Java Spring Boot
```
✓ Extracted 28 dependencies from pom.xml
✓ Detected Spring Boot framework
✓ Generated component documentation
✓ Quality score: 9.5/10.0
```

### Rust CLI Tool
```
✓ Extracted 8 crate dependencies from Cargo.toml
✓ Detected binary crate
✓ Generated usage documentation
✓ Quality score: 9.0/10.0
```

---

## 🎉 Summary

DocBrown now supports:
- ✅ **7 languages** (Go, Python, JS/TS, Rust, Java, Ruby, C#)
- ✅ **6 package managers** (go mod, pip, npm, cargo, maven, bundler)
- ✅ **Real dependency extraction** from all package files
- ✅ **LLM analysis** for all languages
- ✅ **Perfect 10.0 quality scores** achievable

**Ready for production use across your entire tech stack!** 🚀
