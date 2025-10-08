package git

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Operations handles Git operations
type Operations struct {
	repo       *git.Repository
	remoteName string
	baseBranch string
}

// NewOperations creates a new Git operations handler
func NewOperations(remoteName, baseBranch string) (*Operations, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	if remoteName == "" {
		remoteName = "origin"
	}
	if baseBranch == "" {
		baseBranch = "main"
	}

	return &Operations{
		repo:       repo,
		remoteName: remoteName,
		baseBranch: baseBranch,
	}, nil
}

// GetCurrentBranch returns the current branch name
func (g *Operations) GetCurrentBranch() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", err
	}

	return head.Name().Short(), nil
}

// CreateBranch creates a new branch
func (g *Operations) CreateBranch(branchName string) error {
	// Get HEAD reference
	head, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())

	if err := g.repo.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// CheckoutBranch checks out a branch
func (g *Operations) CheckoutBranch(branchName string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}

	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: false,
	})
}

// CreateAndCheckoutBranch creates and checks out a new branch
func (g *Operations) CreateAndCheckoutBranch(branchName string) error {
	if err := g.CreateBranch(branchName); err != nil {
		return err
	}

	return g.CheckoutBranch(branchName)
}

// StageFiles stages files for commit
func (g *Operations) StageFiles(patterns []string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}

	for _, pattern := range patterns {
		if _, err := w.Add(pattern); err != nil {
			return fmt.Errorf("failed to stage %s: %w", pattern, err)
		}
	}

	return nil
}

// Commit creates a commit
func (g *Operations) Commit(message string) (string, error) {
	w, err := g.repo.Worktree()
	if err != nil {
		return "", err
	}

	// Check if there are changes to commit
	status, err := w.Status()
	if err != nil {
		return "", err
	}

	if status.IsClean() {
		return "", fmt.Errorf("no changes to commit")
	}

	hash, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "DocBrown",
			Email: "docbrown@example.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return hash.String(), nil
}

// Push pushes to remote
func (g *Operations) Push(branchName, token string) error {
	auth := &http.BasicAuth{
		Username: "docbrown",
		Password: token,
	}

	return g.repo.Push(&git.PushOptions{
		RemoteName: g.remoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)),
		},
		Auth: auth,
	})
}

// PushDirect pushes directly to the base branch
func (g *Operations) PushDirect(token string) error {
	currentBranch, err := g.GetCurrentBranch()
	if err != nil {
		return err
	}

	return g.Push(currentBranch, token)
}

// GetRemoteURL gets the remote URL
func (g *Operations) GetRemoteURL() (string, error) {
	remote, err := g.repo.Remote(g.remoteName)
	if err != nil {
		return "", err
	}

	cfg := remote.Config()
	if len(cfg.URLs) == 0 {
		return "", fmt.Errorf("no URL configured for remote %s", g.remoteName)
	}

	return cfg.URLs[0], nil
}

// DocsExist checks if documentation already exists
func (g *Operations) DocsExist() bool {
	_, err := os.Stat("docs")
	return err == nil
}

// DeterminePushStrategy determines whether to use PR or direct push
func (g *Operations) DeterminePushStrategy(strategy string, forcePR bool) string {
	if forcePR {
		return "pr"
	}

	if strategy == "direct" {
		return "direct"
	}

	if strategy == "pr" {
		return "pr"
	}

	// Auto strategy
	if g.DocsExist() {
		return "pr" // Docs exist, use PR for review
	}

	return "direct" // No docs, push directly
}

// DetectPlatform detects the Git platform from remote URL
func (g *Operations) DetectPlatform() (string, error) {
	url, err := g.GetRemoteURL()
	if err != nil {
		return "", err
	}

	if strings.Contains(url, "github.com") {
		return "github", nil
	}

	if strings.Contains(url, "gitlab.com") {
		return "gitlab", nil
	}

	if strings.Contains(url, "bitbucket.org") {
		return "bitbucket", nil
	}

	return "", fmt.Errorf("unsupported platform for URL: %s", url)
}

// ParseGitHubURL parses a GitHub URL to extract owner and repo
func ParseGitHubURL(url string) (owner, repo string, err error) {
	// Handle git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		parts := strings.TrimPrefix(url, "git@github.com:")
		parts = strings.TrimSuffix(parts, ".git")
		split := strings.Split(parts, "/")
		if len(split) == 2 {
			return split[0], split[1], nil
		}
	}

	// Handle https://github.com/owner/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		parts := strings.TrimPrefix(url, "https://github.com/")
		parts = strings.TrimSuffix(parts, ".git")
		split := strings.Split(parts, "/")
		if len(split) == 2 {
			return split[0], split[1], nil
		}
	}

	return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
}

// ParseGitLabURL parses a GitLab URL
func ParseGitLabURL(url string) (projectID string, err error) {
	// Simplified - would need more robust parsing
	if strings.HasPrefix(url, "git@gitlab.com:") {
		parts := strings.TrimPrefix(url, "git@gitlab.com:")
		parts = strings.TrimSuffix(parts, ".git")
		return strings.ReplaceAll(parts, "/", "%2F"), nil
	}

	if strings.HasPrefix(url, "https://gitlab.com/") {
		parts := strings.TrimPrefix(url, "https://gitlab.com/")
		parts = strings.TrimSuffix(parts, ".git")
		return strings.ReplaceAll(parts, "/", "%2F"), nil
	}

	return "", fmt.Errorf("invalid GitLab URL: %s", url)
}
