package llm

import (
	"context"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// IsAvailable checks if the provider is available
	IsAvailable() bool

	// Ping checks if the provider is reachable
	Ping(ctx context.Context) error

	// Analyze analyzes a codebase component and returns structured insights
	Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error)

	// Generate generates documentation content
	Generate(ctx context.Context, req GenerateRequest) (string, error)

	// EstimateCost estimates the cost for a given number of tokens
	EstimateCost(tokens int) float64
}

// AnalysisRequest represents a request to analyze a codebase component
type AnalysisRequest struct {
	ComponentName string
	ComponentType string
	Language      string
	Path          string
	FileTree      string
	KeyFiles      []FileContent
}

// FileContent represents a file's content
type FileContent struct {
	Path    string
	Content string
}

// AnalysisResult contains the structured analysis results
type AnalysisResult struct {
	Overview     string        `json:"overview"`
	Components   []Component   `json:"components"`
	Services     []Service     `json:"services"`
	Architecture Architecture  `json:"architecture"`
}

// Component represents a codebase component
type Component struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Language     string   `json:"language"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// Service represents a service component
type Service struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Endpoints   []Endpoint `json:"endpoints,omitempty"`
	Port        int        `json:"port,omitempty"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Architecture represents architecture information
type Architecture struct {
	Overview     string   `json:"overview"`
	Components   []string `json:"components"`
	Patterns     []string `json:"patterns"`
	Technologies []string `json:"technologies"`
	Diagram      string   `json:"diagram,omitempty"`
}

// GenerateRequest represents a request to generate documentation
type GenerateRequest struct {
	Prompt        string
	ComponentName string
	ComponentType string
	Language      string
	Path          string
	Files         []FileContent
	Context       string
}

// TokenUsage tracks token usage for cost calculation
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}
