package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	endpoint    string
	model       string
	contextSize int
	timeout     time.Duration
	client      *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(endpoint, model string, contextSize int, timeout time.Duration) *OllamaProvider {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen2.5-coder:latest"
	}
	if contextSize == 0 {
		contextSize = 8192
	}
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	return &OllamaProvider{
		endpoint:    endpoint,
		model:       model,
		contextSize: contextSize,
		timeout:     timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the provider name
func (o *OllamaProvider) Name() string {
	return "ollama"
}

// IsAvailable checks if the provider is available
func (o *OllamaProvider) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return o.Ping(ctx) == nil
}

// Ping checks if the provider is reachable
func (o *OllamaProvider) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", o.endpoint+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not available: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

// Analyze analyzes a codebase component
func (o *OllamaProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	prompt := o.buildAnalysisPrompt(req)

	response, err := o.generateWithFormat(ctx, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Parse JSON response
	var result AnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If JSON parsing fails, return a simple result
		return &AnalysisResult{
			Overview: response,
		}, nil
	}

	return &result, nil
}

// Generate generates documentation content
func (o *OllamaProvider) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	prompt := req.Prompt
	if prompt == "" {
		prompt = o.buildGeneratePrompt(req)
	}

	response, err := o.generateWithFormat(ctx, prompt, false)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return response, nil
}

// EstimateCost estimates the cost (Ollama is free)
func (o *OllamaProvider) EstimateCost(tokens int) float64 {
	return 0.0 // Ollama is free
}

// generateWithFormat makes a generation request to Ollama
func (o *OllamaProvider) generateWithFormat(ctx context.Context, prompt string, jsonFormat bool) (string, error) {
	reqBody := map[string]interface{}{
		"model":  o.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
			"num_ctx":     o.contextSize,
		},
	}

	// Only add format for JSON responses
	if jsonFormat {
		reqBody["format"] = "json"
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Response, nil
}

// buildAnalysisPrompt builds the analysis prompt
func (o *OllamaProvider) buildAnalysisPrompt(req AnalysisRequest) string {
	var sb strings.Builder

	sb.WriteString("Analyze this codebase and return JSON with the following structure:\n")
	sb.WriteString("{\n")
	sb.WriteString(`  "overview": "High-level description of the project",` + "\n")
	sb.WriteString(`  "components": [{"name": "...", "type": "service|library|frontend", "language": "...", "path": "...", "description": "..."}],` + "\n")
	sb.WriteString(`  "services": [{"name": "...", "type": "rest|grpc|graphql", "description": "..."}],` + "\n")
	sb.WriteString(`  "architecture": {"overview": "...", "patterns": [], "technologies": []}` + "\n")
	sb.WriteString("}\n\n")

	sb.WriteString("File Tree:\n")
	sb.WriteString(req.FileTree)
	sb.WriteString("\n\n")

	// For Ollama, limit the number of files due to smaller context window
	maxFiles := 5
	if len(req.KeyFiles) > maxFiles {
		sb.WriteString(fmt.Sprintf("Key Files (showing %d of %d):\n", maxFiles, len(req.KeyFiles)))
		for i := 0; i < maxFiles; i++ {
			file := req.KeyFiles[i]
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", file.Path))
			sb.WriteString(file.Content)
			sb.WriteString("\n")
		}
	} else if len(req.KeyFiles) > 0 {
		sb.WriteString("Key Files:\n")
		for _, file := range req.KeyFiles {
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", file.Path))
			sb.WriteString(file.Content)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// buildGeneratePrompt builds the generation prompt
func (o *OllamaProvider) buildGeneratePrompt(req GenerateRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Document the following %s component in detail.\n\n", req.ComponentType))
	sb.WriteString(fmt.Sprintf("Component: %s\n", req.ComponentName))
	sb.WriteString(fmt.Sprintf("Type: %s\n", req.ComponentType))
	sb.WriteString(fmt.Sprintf("Language: %s\n", req.Language))
	sb.WriteString(fmt.Sprintf("Path: %s\n\n", req.Path))

	// For Ollama, limit the number of files
	maxFiles := 10
	if len(req.Files) > maxFiles {
		sb.WriteString(fmt.Sprintf("Source Files (showing %d of %d):\n", maxFiles, len(req.Files)))
		for i := 0; i < maxFiles; i++ {
			file := req.Files[i]
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", file.Path))
			// Truncate very long files
			content := file.Content
			if len(content) > 5000 {
				content = content[:5000] + "\n... (truncated)"
			}
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	} else if len(req.Files) > 0 {
		sb.WriteString("Source Files:\n")
		for _, file := range req.Files {
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", file.Path))
			content := file.Content
			if len(content) > 5000 {
				content = content[:5000] + "\n... (truncated)"
			}
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Generate comprehensive documentation covering:\n")
	sb.WriteString("1. Purpose and Overview\n")
	sb.WriteString("2. Architecture\n")
	sb.WriteString("3. Public APIs and Interfaces\n")
	sb.WriteString("4. Dependencies\n")
	sb.WriteString("5. Configuration\n")
	sb.WriteString("6. Usage Examples\n\n")
	sb.WriteString("IMPORTANT: Output ONLY well-formatted markdown documentation.\n")
	sb.WriteString("Do NOT output JSON. Output plain markdown text only.\n")
	sb.WriteString("Start your response directly with markdown (no JSON wrapper).\n")

	return sb.String()
}
