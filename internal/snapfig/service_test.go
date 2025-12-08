package snapfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestNewService(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.Config() != cfg {
		t.Error("Config() should return the same config")
	}
	if svc.VaultDir() == "" {
		t.Error("VaultDir() should not be empty")
	}
}

func TestDefaultServiceCopy(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	// Set HOME for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	// Create test file to copy
	testDir := filepath.Join(homeDir, ".config", "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	testFile := filepath.Join(testDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true, Git: config.GitModeDisable},
		},
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	result, err := svc.Copy()
	if err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}

	if len(result.Copied) == 0 {
		t.Error("Copy should have copied paths")
	}
}

func TestDefaultServiceRestore(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	// Set HOME for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	// Create vault structure
	vaultDir := filepath.Join(tmpDir, "vault")
	vaultTestDir := filepath.Join(vaultDir, ".config", "test")
	if err := os.MkdirAll(vaultTestDir, 0755); err != nil {
		t.Fatalf("failed to create vault test dir: %v", err)
	}
	vaultFile := filepath.Join(vaultTestDir, "file.txt")
	if err := os.WriteFile(vaultFile, []byte("vault content"), 0644); err != nil {
		t.Fatalf("failed to create vault file: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true, Git: config.GitModeDisable},
		},
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	result, err := svc.Restore()
	if err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	if len(result.Restored) == 0 {
		t.Error("Restore should have restored paths")
	}
}

func TestDefaultServiceUpdateWatching(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	newWatching := []config.Watched{
		{Path: ".config/new", Enabled: true},
	}
	svc.UpdateWatching(newWatching)

	if len(cfg.Watching) != 1 {
		t.Errorf("Watching length = %d, want 1", len(cfg.Watching))
	}
	if cfg.Watching[0].Path != ".config/new" {
		t.Errorf("Watching[0].Path = %q, want '.config/new'", cfg.Watching[0].Path)
	}
}

func TestDefaultServiceSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
	}

	svc, err := NewService(cfg, configPath)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	err = svc.SaveConfig("")
	if err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should have been created")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath returned error: %v", err)
	}

	if path == "" {
		t.Error("DefaultConfigPath should return a non-empty path")
	}
	if !filepath.IsAbs(path) {
		t.Error("DefaultConfigPath should return an absolute path")
	}
}

func TestMockServiceReset(t *testing.T) {
	mockSvc := NewMockService(nil)

	// Make some calls
	mockSvc.Copy()
	mockSvc.Restore()
	mockSvc.Push()

	if !mockSvc.CopyCalled {
		t.Error("CopyCalled should be true")
	}
	if !mockSvc.RestoreCalled {
		t.Error("RestoreCalled should be true")
	}
	if !mockSvc.PushCalled {
		t.Error("PushCalled should be true")
	}

	// Reset
	mockSvc.Reset()

	if mockSvc.CopyCalled {
		t.Error("CopyCalled should be false after reset")
	}
	if mockSvc.RestoreCalled {
		t.Error("RestoreCalled should be false after reset")
	}
	if mockSvc.PushCalled {
		t.Error("PushCalled should be false after reset")
	}
}

func TestMockServiceCustomFunctions(t *testing.T) {
	mockSvc := NewMockService(nil)

	customCalled := false
	mockSvc.CopyFunc = func() (*CopyResult, error) {
		customCalled = true
		return &CopyResult{FilesUpdated: 99}, nil
	}

	result, _ := mockSvc.Copy()

	if !customCalled {
		t.Error("custom CopyFunc should have been called")
	}
	if result.FilesUpdated != 99 {
		t.Errorf("FilesUpdated = %d, want 99", result.FilesUpdated)
	}
}

func TestMockServiceVaultDir(t *testing.T) {
	mockSvc := NewMockService(nil)
	if mockSvc.VaultDir() != "/tmp/test-vault" {
		t.Errorf("VaultDir() = %q, want '/tmp/test-vault'", mockSvc.VaultDir())
	}
}

func TestMockServiceSetRemote(t *testing.T) {
	mockSvc := NewMockService(nil)
	err := mockSvc.SetRemote("https://github.com/test/repo.git")

	if err != nil {
		t.Errorf("SetRemote returned error: %v", err)
	}
	if !mockSvc.SetRemoteCalled {
		t.Error("SetRemoteCalled should be true")
	}
	if mockSvc.SetRemoteURL != "https://github.com/test/repo.git" {
		t.Errorf("SetRemoteURL = %q, want 'https://github.com/test/repo.git'", mockSvc.SetRemoteURL)
	}
}

func TestMockServiceListVaultEntries(t *testing.T) {
	mockSvc := NewMockService(nil)
	mockSvc.ListVaultEntriesFunc = func() ([]VaultEntry, error) {
		return []VaultEntry{
			{Path: ".config/test", IsDir: true},
			{Path: ".zshrc", IsDir: false},
		}, nil
	}

	entries, err := mockSvc.ListVaultEntries()
	if err != nil {
		t.Errorf("ListVaultEntries returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("len(entries) = %d, want 2", len(entries))
	}
	if !mockSvc.ListVaultEntriesCalled {
		t.Error("ListVaultEntriesCalled should be true")
	}
}

func TestMockServiceRestoreSelective(t *testing.T) {
	mockSvc := NewMockService(nil)
	paths := []string{".config/test/file1", ".config/test/file2"}

	result, err := mockSvc.RestoreSelective(paths)
	if err != nil {
		t.Errorf("RestoreSelective returned error: %v", err)
	}
	if !mockSvc.RestoreSelectiveCalled {
		t.Error("RestoreSelectiveCalled should be true")
	}
	if len(mockSvc.RestoreSelectivePaths) != 2 {
		t.Errorf("RestoreSelectivePaths length = %d, want 2", len(mockSvc.RestoreSelectivePaths))
	}
	if result.FilesUpdated != 2 {
		t.Errorf("FilesUpdated = %d, want 2", result.FilesUpdated)
	}
}

func TestMockServiceSaveConfig(t *testing.T) {
	mockSvc := NewMockService(nil)
	err := mockSvc.SaveConfig("/tmp/test.yml")

	if err != nil {
		t.Errorf("SaveConfig returned error: %v", err)
	}
	if !mockSvc.SaveConfigCalled {
		t.Error("SaveConfigCalled should be true")
	}
	if mockSvc.SaveConfigPath != "/tmp/test.yml" {
		t.Errorf("SaveConfigPath = %q, want '/tmp/test.yml'", mockSvc.SaveConfigPath)
	}
}

func TestMockServicePull(t *testing.T) {
	mockSvc := NewMockService(nil)
	mockSvc.PullFunc = func() (*PullResult, error) {
		return &PullResult{Cloned: true}, nil
	}

	result, err := mockSvc.Pull()
	if err != nil {
		t.Errorf("Pull returned error: %v", err)
	}
	if !result.Cloned {
		t.Error("Cloned should be true")
	}
	if !mockSvc.PullCalled {
		t.Error("PullCalled should be true")
	}
}

func TestMockServiceConfig(t *testing.T) {
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: "/test/vault",
	}
	mockSvc := NewMockService(cfg)

	result := mockSvc.Config()
	if result != cfg {
		t.Error("Config() should return the same config")
	}
}

func TestMockServiceUpdateWatching(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}
	mockSvc := NewMockService(cfg)

	watching := []config.Watched{
		{Path: ".config/test", Enabled: true},
	}
	mockSvc.UpdateWatching(watching)

	if len(cfg.Watching) != 1 {
		t.Errorf("Watching length = %d, want 1", len(cfg.Watching))
	}
}

func TestDefaultServiceRestoreSelective(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	// Create vault structure with files
	vaultDir := filepath.Join(tmpDir, "vault")
	vaultTestDir := filepath.Join(vaultDir, ".config", "test")
	if err := os.MkdirAll(vaultTestDir, 0755); err != nil {
		t.Fatalf("failed to create vault test dir: %v", err)
	}
	vaultFile := filepath.Join(vaultTestDir, "file.txt")
	if err := os.WriteFile(vaultFile, []byte("selective content"), 0644); err != nil {
		t.Fatalf("failed to create vault file: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true, Git: config.GitModeDisable},
		},
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	result, err := svc.RestoreSelective([]string{".config/test"})
	if err != nil {
		t.Fatalf("RestoreSelective returned error: %v", err)
	}

	if result == nil {
		t.Fatal("RestoreSelective returned nil result")
	}
}

func TestDefaultServiceListVaultEntries(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	// Create vault structure
	vaultDir := filepath.Join(tmpDir, "vault")
	vaultTestDir := filepath.Join(vaultDir, ".config", "test")
	if err := os.MkdirAll(vaultTestDir, 0755); err != nil {
		t.Fatalf("failed to create vault test dir: %v", err)
	}
	vaultFile := filepath.Join(vaultTestDir, "file.txt")
	if err := os.WriteFile(vaultFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create vault file: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true, Git: config.GitModeDisable},
		},
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	entries, err := svc.ListVaultEntries()
	if err != nil {
		t.Fatalf("ListVaultEntries returned error: %v", err)
	}

	if len(entries) == 0 {
		t.Error("ListVaultEntries should return entries")
	}
}

func TestDefaultServicePushNoRepo(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	// Push should fail - no git repo
	err = svc.Push()
	if err == nil {
		t.Error("Push should fail when no git repo exists")
	}
}

func TestDefaultServicePullNoRemote(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
		Remote:    "", // No remote
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	// Pull should fail - no remote configured
	_, err = svc.Pull()
	if err == nil {
		t.Error("Pull should fail when no remote configured")
	}
}

func TestDefaultServiceSetRemote(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: tmpDir,
	}

	// Initialize git repo first
	if err := InitVaultRepo(tmpDir); err != nil {
		t.Fatalf("InitVaultRepo returned error: %v", err)
	}

	svc, err := NewService(cfg, filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	err = svc.SetRemote("https://github.com/test/repo.git")
	if err != nil {
		t.Fatalf("SetRemote returned error: %v", err)
	}

	// Verify remote was set
	hasRemote, url, _ := HasRemote(tmpDir)
	if !hasRemote {
		t.Error("remote should be set")
	}
	if url != "https://github.com/test/repo.git" {
		t.Errorf("remote url = %q, want 'https://github.com/test/repo.git'", url)
	}
}
