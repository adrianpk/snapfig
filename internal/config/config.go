// Package config handles configuration loading and validation for Snapfig.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GitMode defines how nested git repositories are handled during replication.
type GitMode string

const (
	GitModeDisable GitMode = "disable"
	GitModeRemove  GitMode = "remove"
)

// Config represents the main Snapfig configuration.
type Config struct {
	Git      GitMode   `yaml:"git"`
	Remote   string    `yaml:"remote,omitempty"`
	Watching []Watched `yaml:"watching"`
}

// Watched represents a directory being observed by Snapfig.
type Watched struct {
	Path    string  `yaml:"path"`
	Git     GitMode `yaml:"git,omitempty"`
	Enabled bool    `yaml:"enabled"`
}

// DefaultConfigDir returns the default configuration directory path.
func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "snapfig"), nil
}

// DefaultVaultDir returns the default vault directory path.
func DefaultVaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".snapfig", "vault"), nil
}

// Load reads and parses the configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Git == "" {
		cfg.Git = GitModeDisable
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Git != GitModeDisable && c.Git != GitModeRemove {
		return errors.New("git mode must be 'disable' or 'remove'")
	}
	return nil
}

// EffectiveGitMode returns the git mode for a watched path,
// falling back to the global setting if not specified.
func (w *Watched) EffectiveGitMode(global GitMode) GitMode {
	if w.Git != "" {
		return w.Git
	}
	return global
}
