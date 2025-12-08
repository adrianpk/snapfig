// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"fmt"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
)

// Service defines the interface for all snapfig operations.
// This allows for dependency injection and easy testing.
type Service interface {
	// Copy copies all enabled watched paths to the vault.
	Copy() (*CopyResult, error)

	// Restore restores all enabled watched paths from vault.
	Restore() (*RestoreResult, error)

	// RestoreSelective restores only the specified paths from vault.
	RestoreSelective(paths []string) (*RestoreResult, error)

	// ListVaultEntries returns all entries in the vault that match the config.
	ListVaultEntries() ([]VaultEntry, error)

	// Push pushes the vault to the configured remote.
	Push() error

	// Pull pulls the vault from remote, cloning if needed.
	Pull() (*PullResult, error)

	// SetRemote configures the git remote for the vault.
	SetRemote(url string) error

	// SaveConfig saves the configuration to the given path.
	SaveConfig(path string) error

	// Config returns the current configuration.
	Config() *config.Config

	// VaultDir returns the vault directory path.
	VaultDir() string

	// UpdateWatching updates the watching list in config.
	UpdateWatching(watching []config.Watched)
}

// DefaultService is the production implementation of Service.
type DefaultService struct {
	cfg        *config.Config
	vaultDir   string
	configPath string
}

// NewService creates a new DefaultService.
func NewService(cfg *config.Config, configPath string) (*DefaultService, error) {
	vaultDir, err := cfg.VaultDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault directory: %w", err)
	}

	return &DefaultService{
		cfg:        cfg,
		vaultDir:   vaultDir,
		configPath: configPath,
	}, nil
}

// Copy copies all enabled watched paths to the vault.
func (s *DefaultService) Copy() (*CopyResult, error) {
	copier, err := NewCopier(s.cfg)
	if err != nil {
		return nil, err
	}
	return copier.Copy()
}

// Restore restores all enabled watched paths from vault.
func (s *DefaultService) Restore() (*RestoreResult, error) {
	restorer, err := NewRestorer(s.cfg)
	if err != nil {
		return nil, err
	}
	return restorer.Restore()
}

// RestoreSelective restores only the specified paths from vault.
func (s *DefaultService) RestoreSelective(paths []string) (*RestoreResult, error) {
	restorer, err := NewRestorer(s.cfg)
	if err != nil {
		return nil, err
	}
	return restorer.RestoreSelective(paths)
}

// ListVaultEntries returns all entries in the vault that match the config.
func (s *DefaultService) ListVaultEntries() ([]VaultEntry, error) {
	restorer, err := NewRestorer(s.cfg)
	if err != nil {
		return nil, err
	}
	return restorer.ListVaultEntries()
}

// Push pushes the vault to the configured remote.
func (s *DefaultService) Push() error {
	return PushVaultWithToken(s.vaultDir, s.cfg.GitToken)
}

// Pull pulls the vault from remote, cloning if needed.
func (s *DefaultService) Pull() (*PullResult, error) {
	return PullVaultWithToken(s.vaultDir, s.cfg.Remote, s.cfg.GitToken)
}

// SetRemote configures the git remote for the vault.
func (s *DefaultService) SetRemote(url string) error {
	return SetRemote(s.vaultDir, url)
}

// SaveConfig saves the configuration to the configured path.
func (s *DefaultService) SaveConfig(path string) error {
	if path == "" {
		path = s.configPath
	}
	return s.cfg.Save(path)
}

// Config returns the current configuration.
func (s *DefaultService) Config() *config.Config {
	return s.cfg
}

// VaultDir returns the vault directory path.
func (s *DefaultService) VaultDir() string {
	return s.vaultDir
}

// UpdateWatching updates the watching list in config.
func (s *DefaultService) UpdateWatching(watching []config.Watched) {
	s.cfg.Watching = watching
}

// DefaultConfigPath returns the default config path.
func DefaultConfigPath() (string, error) {
	configDir, err := config.DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yml"), nil
}
