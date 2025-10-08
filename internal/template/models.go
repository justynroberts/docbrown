package template

import "time"

// Template represents a documentation template
type Template struct {
	Name        string         `yaml:"name"`
	Version     string         `yaml:"version"`
	Description string         `yaml:"description"`
	Path        string         `yaml:"-"`
	Files       []TemplateFile `yaml:"files"`
	Prompts     map[string]string `yaml:"prompts,omitempty"`
}

// TemplateFile represents a file to be generated from a template
type TemplateFile struct {
	Name        string `yaml:"name"`
	Template    string `yaml:"template"`
	Output      string `yaml:"output"`
	Description string `yaml:"description"`
	Foreach     string `yaml:"foreach,omitempty"`
	Condition   string `yaml:"condition,omitempty"`
}

// TemplateData contains data passed to templates
type TemplateData struct {
	// Repository metadata
	RepoName      string
	RepoURL       string
	DefaultBranch string
	Description   string

	// Components
	Components []ComponentData
	Services   []ServiceData
	Libraries  []LibraryData
	Frontends  []FrontendData

	// Architecture
	Architecture ArchitectureData

	// Generated content
	Overview       string
	GettingStarted string

	// Metadata
	Timestamp   time.Time
	GeneratedBy string
	Version     string
}

// ComponentData represents component data for templates
type ComponentData struct {
	Name          string
	Type          string
	Language      string
	Path          string
	Description   string
	Overview      string
	APIs          []APIData
	Dependencies  []DependencyData
	UsageExample  string
	Architecture  string
	Configuration map[string]string
	HasTests      bool
	TestCoverage  float64
}

// ServiceData represents service data for templates
type ServiceData struct {
	Name         string
	Type         string
	Description  string
	Endpoints    []APIData
	Port         int
	Dependencies []DependencyData
}

// LibraryData represents library data for templates
type LibraryData struct {
	Name        string
	Language    string
	Description string
	Functions   []FunctionData
}

// FrontendData represents frontend data for templates
type FrontendData struct {
	Name        string
	Framework   string
	Description string
	Routes      []RouteData
}

// APIData represents API endpoint data
type APIData struct {
	Method          string
	Path            string
	Description     string
	Parameters      []ParameterData
	RequestExample  string
	ResponseExample string
	ErrorCodes      []ErrorCodeData
	Authentication  string
}

// ParameterData represents parameter data
type ParameterData struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Example     string
}

// ErrorCodeData represents error code data
type ErrorCodeData struct {
	Code        int
	Message     string
	Description string
}

// DependencyData represents dependency data
type DependencyData struct {
	Name    string
	Version string
	Purpose string
	Type    string
}

// ArchitectureData represents architecture information
type ArchitectureData struct {
	Overview     string
	Components   []string
	Patterns     []string
	Technologies []string
	Diagram      string
}

// FunctionData represents function documentation
type FunctionData struct {
	Name        string
	Signature   string
	Description string
	Parameters  []ParameterData
	Returns     string
	Example     string
}

// RouteData represents frontend route data
type RouteData struct {
	Path        string
	Component   string
	Description string
}
