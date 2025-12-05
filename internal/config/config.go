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

// DaemonConfig holds settings for the background runner.
type DaemonConfig struct {
	CopyInterval string `yaml:"copy_interval,omitempty"` // e.g. "1h", "30m"
	PushInterval string `yaml:"push_interval,omitempty"` // e.g. "24h", "12h"
	PullInterval string `yaml:"pull_interval,omitempty"` // disabled by default
	AutoRestore  bool   `yaml:"auto_restore,omitempty"`  // restore after pull
}

// Config represents the main Snapfig configuration.
type Config struct {
	Git       GitMode      `yaml:"git"`
	Remote    string       `yaml:"remote,omitempty"`
	GitToken  string       `yaml:"git_token,omitempty"` // app token for HTTPS auth
	VaultPath string       `yaml:"vault_path,omitempty"` // custom vault location
	Watching  []Watched    `yaml:"watching"`
	Daemon    DaemonConfig `yaml:"daemon,omitempty"`
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

// DefaultSnapfigDir returns the base ~/.snapfig directory path.
func DefaultSnapfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".snapfig"), nil
}

// PidFilePath returns the path to the daemon PID file.
func PidFilePath() (string, error) {
	dir, err := DefaultSnapfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "daemon.pid"), nil
}

// LogFilePath returns the path to the daemon log file.
func LogFilePath() (string, error) {
	dir, err := DefaultSnapfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "daemon.log"), nil
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

// VaultDir returns the vault directory path, using custom path if set.
func (c *Config) VaultDir() (string, error) {
	if c.VaultPath != "" {
		// Expand ~ if present
		if len(c.VaultPath) > 0 && c.VaultPath[0] == '~' {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, c.VaultPath[1:]), nil
		}
		return c.VaultPath, nil
	}
	return DefaultVaultDir()
}

// SnapfigDir returns the base snapfig directory (parent of vault).
func (c *Config) SnapfigDir() (string, error) {
	vaultDir, err := c.VaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Dir(vaultDir), nil
}
