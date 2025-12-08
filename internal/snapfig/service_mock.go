// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"github.com/adrianpk/snapfig/internal/config"
)

// MockService is a test implementation of Service.
type MockService struct {
	cfg      *config.Config
	vaultDir string

	// Function hooks for mocking behavior
	CopyFunc              func() (*CopyResult, error)
	RestoreFunc           func() (*RestoreResult, error)
	RestoreSelectiveFunc  func(paths []string) (*RestoreResult, error)
	ListVaultEntriesFunc  func() ([]VaultEntry, error)
	PushFunc              func() error
	PullFunc              func() (*PullResult, error)
	SetRemoteFunc         func(url string) error
	SaveConfigFunc        func(path string) error
	UpdateWatchingFunc    func(watching []config.Watched)

	// Call tracking
	CopyCalled             bool
	RestoreCalled          bool
	RestoreSelectiveCalled bool
	RestoreSelectivePaths  []string
	ListVaultEntriesCalled bool
	PushCalled             bool
	PullCalled             bool
	SetRemoteCalled        bool
	SetRemoteURL           string
	SaveConfigCalled       bool
	SaveConfigPath         string
	UpdateWatchingCalled   bool
	UpdateWatchingValue    []config.Watched
}

// NewMockService creates a new MockService with default behavior.
func NewMockService(cfg *config.Config) *MockService {
	if cfg == nil {
		cfg = &config.Config{Git: config.GitModeDisable}
	}
	return &MockService{
		cfg:      cfg,
		vaultDir: "/tmp/test-vault",
	}
}

// Copy mocks the Copy operation.
func (m *MockService) Copy() (*CopyResult, error) {
	m.CopyCalled = true
	if m.CopyFunc != nil {
		return m.CopyFunc()
	}
	return &CopyResult{
		Copied:       []string{},
		Skipped:      []string{},
		FilesUpdated: 0,
		FilesSkipped: 0,
		FilesRemoved: 0,
	}, nil
}

// Restore mocks the Restore operation.
func (m *MockService) Restore() (*RestoreResult, error) {
	m.RestoreCalled = true
	if m.RestoreFunc != nil {
		return m.RestoreFunc()
	}
	return &RestoreResult{
		Restored:     []string{},
		Skipped:      []string{},
		Backups:      []string{},
		FilesUpdated: 0,
		FilesSkipped: 0,
	}, nil
}

// RestoreSelective mocks the RestoreSelective operation.
func (m *MockService) RestoreSelective(paths []string) (*RestoreResult, error) {
	m.RestoreSelectiveCalled = true
	m.RestoreSelectivePaths = paths
	if m.RestoreSelectiveFunc != nil {
		return m.RestoreSelectiveFunc(paths)
	}
	return &RestoreResult{
		Restored:     paths,
		Skipped:      []string{},
		Backups:      []string{},
		FilesUpdated: len(paths),
		FilesSkipped: 0,
	}, nil
}

// ListVaultEntries mocks the ListVaultEntries operation.
func (m *MockService) ListVaultEntries() ([]VaultEntry, error) {
	m.ListVaultEntriesCalled = true
	if m.ListVaultEntriesFunc != nil {
		return m.ListVaultEntriesFunc()
	}
	return []VaultEntry{}, nil
}

// Push mocks the Push operation.
func (m *MockService) Push() error {
	m.PushCalled = true
	if m.PushFunc != nil {
		return m.PushFunc()
	}
	return nil
}

// Pull mocks the Pull operation.
func (m *MockService) Pull() (*PullResult, error) {
	m.PullCalled = true
	if m.PullFunc != nil {
		return m.PullFunc()
	}
	return &PullResult{Cloned: false}, nil
}

// SetRemote mocks the SetRemote operation.
func (m *MockService) SetRemote(url string) error {
	m.SetRemoteCalled = true
	m.SetRemoteURL = url
	if m.SetRemoteFunc != nil {
		return m.SetRemoteFunc(url)
	}
	return nil
}

// SaveConfig mocks the SaveConfig operation.
func (m *MockService) SaveConfig(path string) error {
	m.SaveConfigCalled = true
	m.SaveConfigPath = path
	if m.SaveConfigFunc != nil {
		return m.SaveConfigFunc(path)
	}
	return nil
}

// Config returns the configuration.
func (m *MockService) Config() *config.Config {
	return m.cfg
}

// VaultDir returns the vault directory.
func (m *MockService) VaultDir() string {
	return m.vaultDir
}

// UpdateWatching updates the watching list.
func (m *MockService) UpdateWatching(watching []config.Watched) {
	m.UpdateWatchingCalled = true
	m.UpdateWatchingValue = watching
	m.cfg.Watching = watching
	if m.UpdateWatchingFunc != nil {
		m.UpdateWatchingFunc(watching)
	}
}

// Reset clears all call tracking state.
func (m *MockService) Reset() {
	m.CopyCalled = false
	m.RestoreCalled = false
	m.RestoreSelectiveCalled = false
	m.RestoreSelectivePaths = nil
	m.ListVaultEntriesCalled = false
	m.PushCalled = false
	m.PullCalled = false
	m.SetRemoteCalled = false
	m.SetRemoteURL = ""
	m.SaveConfigCalled = false
	m.SaveConfigPath = ""
	m.UpdateWatchingCalled = false
	m.UpdateWatchingValue = nil
}
