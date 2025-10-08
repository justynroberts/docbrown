package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Cache represents the cached analysis and generation state
type Cache struct {
	Version    string                    `yaml:"version"`
	LastRun    time.Time                 `yaml:"last_run"`
	Components map[string]ComponentCache `yaml:"components"`
}

// ComponentCache represents cached information for a component
type ComponentCache struct {
	Hash          string    `yaml:"hash"`
	FilesHash     string    `yaml:"files_hash"`
	LastGenerated time.Time `yaml:"last_generated"`
	Files         []string  `yaml:"files"`
}

// Manager manages the cache
type Manager struct {
	cachePath string
	cache     *Cache
	enabled   bool
	ttl       time.Duration
}

// NewManager creates a new cache manager
func NewManager(cachePath string, enabled bool, ttl time.Duration) *Manager {
	return &Manager{
		cachePath: cachePath,
		enabled:   enabled,
		ttl:       ttl,
		cache: &Cache{
			Version:    "1.0",
			Components: make(map[string]ComponentCache),
		},
	}
}

// Load loads the cache from disk
func (m *Manager) Load() error {
	if !m.enabled {
		return nil
	}

	data, err := os.ReadFile(m.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache doesn't exist yet, that's fine
			return nil
		}
		return fmt.Errorf("failed to read cache: %w", err)
	}

	if err := yaml.Unmarshal(data, m.cache); err != nil {
		return fmt.Errorf("failed to parse cache: %w", err)
	}

	return nil
}

// Save saves the cache to disk
func (m *Manager) Save() error {
	if !m.enabled {
		return nil
	}

	// Update last run time
	m.cache.LastRun = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(m.cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := yaml.Marshal(m.cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(m.cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// IsStale checks if a component's cache is stale
func (m *Manager) IsStale(componentName string, files []string) bool {
	if !m.enabled {
		return true // Always regenerate if cache disabled
	}

	cached, exists := m.cache.Components[componentName]
	if !exists {
		return true // Not cached
	}

	// Check TTL
	if time.Since(cached.LastGenerated) > m.ttl {
		return true // Expired
	}

	// Check if files changed
	currentHash := m.hashFiles(files)
	if currentHash != cached.FilesHash {
		return true // Modified
	}

	return false // Cache valid
}

// Update updates the cache for a component
func (m *Manager) Update(componentName string, files []string) {
	if !m.enabled {
		return
	}

	m.cache.Components[componentName] = ComponentCache{
		Hash:          m.hashComponent(componentName, files),
		FilesHash:     m.hashFiles(files),
		LastGenerated: time.Now(),
		Files:         files,
	}
}

// GetChangedComponents returns components that have changed since last run
func (m *Manager) GetChangedComponents(components map[string][]string) []string {
	if !m.enabled {
		// Return all if cache disabled
		result := make([]string, 0, len(components))
		for name := range components {
			result = append(result, name)
		}
		return result
	}

	var changed []string

	for name, files := range components {
		if m.IsStale(name, files) {
			changed = append(changed, name)
		}
	}

	return changed
}

// Clear clears the cache
func (m *Manager) Clear() error {
	if !m.enabled {
		return nil
	}

	m.cache = &Cache{
		Version:    "1.0",
		Components: make(map[string]ComponentCache),
	}

	if err := os.Remove(m.cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache: %w", err)
	}

	return nil
}

// GetStats returns cache statistics
func (m *Manager) GetStats() map[string]interface{} {
	if !m.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	unchanged := 0
	stale := 0

	for name, cached := range m.cache.Components {
		if m.IsStale(name, cached.Files) {
			stale++
		} else {
			unchanged++
		}
	}

	return map[string]interface{}{
		"enabled":    true,
		"last_run":   m.cache.LastRun,
		"components": len(m.cache.Components),
		"unchanged":  unchanged,
		"stale":      stale,
	}
}

// hashFiles creates a hash of all files
func (m *Manager) hashFiles(files []string) string {
	h := sha256.New()

	for _, file := range files {
		// Hash file path
		h.Write([]byte(file))

		// Hash file content
		if f, err := os.Open(file); err == nil {
			io.Copy(h, f)
			f.Close()
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

// hashComponent creates a hash of a component
func (m *Manager) hashComponent(name string, files []string) string {
	h := sha256.New()
	h.Write([]byte(name))
	h.Write([]byte(m.hashFiles(files)))
	return hex.EncodeToString(h.Sum(nil))
}

// GetCache returns the cache
func (m *Manager) GetCache() *Cache {
	return m.cache
}
