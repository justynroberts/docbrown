package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// AnthropicProvider implements the Provider interface for Anthropic Claude
type AnthropicProvider struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
	usage     TokenUsage
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string, maxTokens int) *AnthropicProvider {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &AnthropicProvider{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{},
	}
}

// Name returns the provider name
func (a *AnthropicProvider) Name() string {
	return "anthropic"
}

// IsAvailable checks if the provider is available
func (a *AnthropicProvider) IsAvailable() bool {
	return a.apiKey != ""
}

// Ping checks if the provider is reachable
func (a *AnthropicProvider) Ping(ctx context.Context) error {
	// Simple test call with minimal tokens
	_, err := a.callAPI(ctx, "Hello", 10)
	return err
}

// Analyze analyzes a codebase component
func (a *AnthropicProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	prompt := a.buildAnalysisPrompt(req)

	response, err := a.callAPI(ctx, prompt, a.maxTokens)
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
func (a *AnthropicProvider) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	prompt := req.Prompt
	if prompt == "" {
		prompt = a.buildGeneratePrompt(req)
	}

	response, err := a.callAPI(ctx, prompt, a.maxTokens)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return response, nil
}

// EstimateCost estimates the cost for a given number of tokens
func (a *AnthropicProvider) EstimateCost(tokens int) float64 {
	// Assuming 50/50 split between input and output
	inputTokens := tokens / 2
	outputTokens := tokens / 2

	inputCost := float64(inputTokens) * 0.003 / 1000
	outputCost := float64(outputTokens) * 0.015 / 1000

	return inputCost + outputCost
}

// GetUsage returns the total token usage
func (a *AnthropicProvider) GetUsage() TokenUsage {
	return a.usage
}

// callAPI makes a call to the Anthropic API
func (a *AnthropicProvider) callAPI(ctx context.Context, prompt string, maxTokens int) (string, error) {
	reqBody := map[string]interface{}{
		"model":      a.model,
		"max_tokens": maxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Track usage
	a.usage.InputTokens += response.Usage.InputTokens
	a.usage.OutputTokens += response.Usage.OutputTokens

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return response.Content[0].Text, nil
}

// buildAnalysisPrompt builds the analysis prompt
func (a *AnthropicProvider) buildAnalysisPrompt(req AnalysisRequest) string {
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

	if len(req.KeyFiles) > 0 {
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
func (a *AnthropicProvider) buildGeneratePrompt(req GenerateRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Document the following %s component in detail.\n\n", req.ComponentType))
	sb.WriteString(fmt.Sprintf("Component: %s\n", req.ComponentName))
	sb.WriteString(fmt.Sprintf("Type: %s\n", req.ComponentType))
	sb.WriteString(fmt.Sprintf("Language: %s\n", req.Language))
	sb.WriteString(fmt.Sprintf("Path: %s\n\n", req.Path))

	if len(req.Files) > 0 {
		sb.WriteString("Source Files:\n")
		for _, file := range req.Files {
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", file.Path))
			sb.WriteString(file.Content)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nGenerate comprehensive documentation covering:\n")
	sb.WriteString("1. Purpose and Overview\n")
	sb.WriteString("2. Architecture\n")
	sb.WriteString("3. Public APIs and Interfaces\n")
	sb.WriteString("4. Dependencies\n")
	sb.WriteString("5. Configuration\n")
	sb.WriteString("6. Usage Examples\n\n")
	sb.WriteString("Output as well-formatted markdown.\n")

	return sb.String()
}
