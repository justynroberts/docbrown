package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GitLab implements Platform for GitLab
type GitLab struct {
	projectID string
	token     string
}

// NewGitLab creates a new GitLab platform
func NewGitLab(projectID, token string) *GitLab {
	return &GitLab{
		projectID: projectID,
		token:     token,
	}
}

// Name returns the platform name
func (gl *GitLab) Name() string {
	return "gitlab"
}

// CreatePR creates a merge request on GitLab
func (gl *GitLab) CreatePR(opts PROptions) (string, error) {
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/merge_requests", gl.projectID)

	reqBody := map[string]interface{}{
		"source_branch": opts.Branch,
		"target_branch": opts.BaseBranch,
		"title":         opts.Title,
		"description":   opts.Body,
	}

	if len(opts.Labels) > 0 {
		reqBody["labels"] = opts.Labels
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gl.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitLab API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		WebURL string `json:"web_url"`
		IID    int    `json:"iid"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.WebURL, nil
}
