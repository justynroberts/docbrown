package analyzer

// RepoStructure represents the analyzed repository structure
type RepoStructure struct {
	RootPath   string
	Components []Component
	FileTree   string
	Languages  map[string]int // language -> file count
	TotalFiles int
}

// Component represents a detected component in the repository
type Component struct {
	Name         string
	Type         string   // service, library, frontend, cli
	Language     string
	Path         string
	Files        []string
	Description  string
	HasTests     bool
	Dependencies []Dependency
	Endpoints    []Endpoint
	EntryPoint   string
}

// Dependency represents a dependency
type Dependency struct {
	Name    string
	Version string
	Type    string // internal, external
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Method      string
	Path        string
	Description string
}

// FileInfo contains information about a file
type FileInfo struct {
	Path     string
	Size     int64
	Language string
	IsTest   bool
}
