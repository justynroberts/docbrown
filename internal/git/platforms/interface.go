package platforms

// Platform defines the interface for Git platform integrations
type Platform interface {
	// Name returns the platform name
	Name() string

	// CreatePR creates a pull request
	CreatePR(opts PROptions) (string, error)
}

// PROptions contains options for creating a pull request
type PROptions struct {
	Title      string
	Body       string
	Branch     string
	BaseBranch string
	Labels     []string
}
