// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
	"gopkg.in/yaml.v3"
)

// ManifestEntry represents a single entry in the manifest.
type ManifestEntry struct {
	Path    string          `yaml:"path"`
	Git     config.GitMode  `yaml:"git"`
	Enabled bool            `yaml:"enabled"`
	IsDir   bool            `yaml:"is_dir"`
}

// Manifest represents the vault manifest with all backed up paths.
type Manifest struct {
	Version       int             `yaml:"version"`
	LastUpdated   string          `yaml:"last_updated"`
	Entries       []ManifestEntry `yaml:"entries"`
	VaultLocation string          `yaml:"vault_location"`
}

const manifestFilename = "manifest.yml"
const manifestVersion = 1

// ManifestPath returns the path to the manifest file inside the vault.
func ManifestPath(vaultDir string) string {
	return filepath.Join(vaultDir, manifestFilename)
}

// WriteManifest writes the manifest to the vault directory.
func WriteManifest(vaultDir string, entries []ManifestEntry) error {
	manifest := Manifest{
		Version:       manifestVersion,
		LastUpdated:   time.Now().Format("2006-01-02 15:04:05"),
		Entries:       entries,
		VaultLocation: vaultDir,
	}

	data, err := yaml.Marshal(&manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestPath := ManifestPath(vaultDir)
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// LoadManifest reads and parses the manifest from the vault directory.
func LoadManifest(vaultDir string) (*Manifest, error) {
	manifestPath := ManifestPath(vaultDir)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest not found in vault: %w", err)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// ManifestExists checks if a manifest exists in the vault.
func ManifestExists(vaultDir string) bool {
	manifestPath := ManifestPath(vaultDir)
	_, err := os.Stat(manifestPath)
	return err == nil
}

// ToWatching converts manifest entries to config.Watched slice.
// This enables reconstructing the config from the manifest on a new machine.
func (m *Manifest) ToWatching() []config.Watched {
	watching := make([]config.Watched, 0, len(m.Entries))
	for _, entry := range m.Entries {
		watching = append(watching, config.Watched{
			Path:    entry.Path,
			Git:     entry.Git,
			Enabled: entry.Enabled,
		})
	}
	return watching
}

// FromWatching creates manifest entries from a config.Watched slice and CopiedItems.
func FromWatching(watching []config.Watched, copiedItems []CopiedItem) []ManifestEntry {
	// Create a map of copied items for quick lookup
	copiedMap := make(map[string]CopiedItem)
	for _, item := range copiedItems {
		copiedMap[item.Path] = item
	}

	entries := make([]ManifestEntry, 0, len(watching))
	for _, w := range watching {
		if !w.Enabled {
			continue
		}

		entry := ManifestEntry{
			Path:    w.Path,
			Git:     w.Git,
			Enabled: w.Enabled,
			IsDir:   false, // default
		}

		// Get IsDir from copiedItems if available
		if item, ok := copiedMap[w.Path]; ok {
			entry.IsDir = item.IsDir
			// Use the git mode from copied item (effective mode)
			entry.Git = item.GitMode
		}

		entries = append(entries, entry)
	}

	return entries
}
