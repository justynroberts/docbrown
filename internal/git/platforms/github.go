package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GitHub implements Platform for GitHub
type GitHub struct {
	owner string
	repo  string
	token string
}

// NewGitHub creates a new GitHub platform
func NewGitHub(owner, repo, token string) *GitHub {
	return &GitHub{
		owner: owner,
		repo:  repo,
		token: token,
	}
}

// Name returns the platform name
func (gh *GitHub) Name() string {
	return "github"
}

// CreatePR creates a pull request on GitHub
func (gh *GitHub) CreatePR(opts PROptions) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", gh.owner, gh.repo)

	reqBody := map[string]interface{}{
		"title": opts.Title,
		"head":  opts.Branch,
		"base":  opts.BaseBranch,
		"body":  opts.Body,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gh.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		HTMLURL string `json:"html_url"`
		Number  int    `json:"number"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Add labels if specified
	if len(opts.Labels) > 0 {
		gh.addLabels(result.Number, opts.Labels)
	}

	return result.HTMLURL, nil
}

// addLabels adds labels to a PR
func (gh *GitHub) addLabels(prNumber int, labels []string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/labels",
		gh.owner, gh.repo, prNumber)

	reqBody := map[string]interface{}{
		"labels": labels,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+gh.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add labels (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
