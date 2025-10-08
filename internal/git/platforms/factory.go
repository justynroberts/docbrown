package platforms

import (
	"fmt"

	"github.com/docbrown/cli/internal/git"
)

// NewPlatform creates a Platform based on the detected platform
func NewPlatform(platformName, remoteURL, token string) (Platform, error) {
	switch platformName {
	case "github":
		owner, repo, err := git.ParseGitHubURL(remoteURL)
		if err != nil {
			return nil, err
		}
		return NewGitHub(owner, repo, token), nil

	case "gitlab":
		projectID, err := git.ParseGitLabURL(remoteURL)
		if err != nil {
			return nil, err
		}
		return NewGitLab(projectID, token), nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platformName)
	}
}
