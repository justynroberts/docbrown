package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/docbrown/cli/internal/config"
	"github.com/docbrown/cli/internal/git"
	"github.com/docbrown/cli/internal/git/platforms"
)

var (
	prBranch     string
	prTitle      string
	prBody       string
	prPAT        string
	pushDirect   bool
	forcePR      bool
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create pull request with generated documentation",
	Long: `Create a pull request (or push directly) with the generated documentation.
Uses the configured push strategy (auto, pr, direct) to determine behavior.`,
	RunE: runPR,
}

func init() {
	rootCmd.AddCommand(prCmd)

	prCmd.Flags().StringVar(&prBranch, "branch", "", "branch name (default: auto-generated)")
	prCmd.Flags().StringVar(&prTitle, "title", "docs: Update documentation", "PR title")
	prCmd.Flags().StringVar(&prBody, "body", "", "PR body")
	prCmd.Flags().StringVar(&prPAT, "pat", "", "personal access token")
	prCmd.Flags().BoolVar(&pushDirect, "push-direct", false, "push directly to base branch (skip PR)")
	prCmd.Flags().BoolVar(&forcePR, "force-pr", false, "always create PR even if no docs exist")
}

func runPR(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	cfg, err := cfgMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get PAT
	token := prPAT
	if token == "" {
		token = cfg.Git.PAT
	}
	if token == "" {
		return fmt.Errorf("no PAT configured (use --pat flag or set GITHUB_TOKEN)")
	}

	// Create Git operations
	gitOps, err := git.NewOperations(cfg.Git.Remote, cfg.Git.BaseBranch)
	if err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	// Determine strategy
	strategy := "pr"
	if pushDirect {
		strategy = "direct"
	}
	strategy = gitOps.DeterminePushStrategy(strategy, forcePR)

	fmt.Println("Creating pull request...")
	fmt.Println()

	// Detect platform
	platformName, err := gitOps.DetectPlatform()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	fmt.Printf("✓ Platform: %s\n", platformName)

	remoteURL, err := gitOps.GetRemoteURL()
	if err != nil {
		return err
	}

	fmt.Printf("✓ Remote: %s\n", remoteURL)
	fmt.Printf("✓ Strategy: %s\n", strategy)
	fmt.Println()

	if strategy == "direct" {
		return runDirectPush(gitOps, token)
	}

	return runPRCreation(gitOps, platformName, remoteURL, token, cfg)
}

func runDirectPush(gitOps *git.Operations, token string) error {
	fmt.Println("📝 Pushing directly to base branch...")

	// Stage files
	filesToStage := []string{
		"docs/",
		"mkdocs.yml",
		"catalog-info.yaml",
	}

	if err := gitOps.StageFiles(filesToStage); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	fmt.Println("✓ Staged files")

	// Commit
	commitMsg := `docs: Add generated documentation

🤖 Generated with DocBrown`

	hash, err := gitOps.Commit(commitMsg)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Printf("✓ Committed: %s\n", hash[:7])

	// Push
	if err := gitOps.PushDirect(token); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	fmt.Println("✓ Pushed to remote")
	fmt.Println()
	fmt.Println("✅ Documentation pushed successfully")

	return nil
}

func runPRCreation(gitOps *git.Operations, platformName, remoteURL, token string, cfg *config.Config) error {
	// Generate branch name if not specified
	branchName := prBranch
	if branchName == "" {
		branchName = fmt.Sprintf("%s-%s", cfg.Git.BranchPrefix, time.Now().Format("20060102"))
	}

	fmt.Printf("Creating branch: %s\n", branchName)

	// Create and checkout branch
	if err := gitOps.CreateAndCheckoutBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Println("✓ Created branch")

	// Stage files
	filesToStage := []string{
		"docs/",
		"mkdocs.yml",
		"catalog-info.yaml",
	}

	if err := gitOps.StageFiles(filesToStage); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	fmt.Println("✓ Staged files")

	// Commit
	commitMsg := `docs: Update documentation

🤖 Generated with DocBrown`

	hash, err := gitOps.Commit(commitMsg)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Printf("✓ Committed: %s\n", hash[:7])

	// Push
	if err := gitOps.Push(branchName, token); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	fmt.Println("✓ Pushed to remote")

	// Create PR
	fmt.Println()
	fmt.Println("Creating pull request...")

	platform, err := platforms.NewPlatform(platformName, remoteURL, token)
	if err != nil {
		return fmt.Errorf("failed to create platform client: %w", err)
	}

	body := prBody
	if body == "" {
		body = `## Summary

This PR updates the documentation using DocBrown.

## Changes

- Updated component documentation
- Refreshed architecture overview
- Updated getting started guide

🤖 Generated with [DocBrown](https://github.com/docbrown/cli)`
	}

	prURL, err := platform.CreatePR(platforms.PROptions{
		Title:      prTitle,
		Body:       body,
		Branch:     branchName,
		BaseBranch: cfg.Git.BaseBranch,
		Labels:     cfg.Git.PRLabels,
	})

	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	fmt.Printf("✓ PR created: %s\n", prURL)
	fmt.Println()
	fmt.Println("✅ Pull request created successfully")
	fmt.Println()
	fmt.Println("Next: Review and merge the PR")

	return nil
}
