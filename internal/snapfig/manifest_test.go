package snapfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestWriteManifest(t *testing.T) {
	// Create temp vault directory
	vaultDir := t.TempDir()

	entries := []ManifestEntry{
		{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true, IsDir: true},
		{Path: ".zshrc", Git: config.GitModeRemove, Enabled: true, IsDir: false},
	}

	err := WriteManifest(vaultDir, entries)
	if err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Verify file exists inside vault
	manifestPath := filepath.Join(vaultDir, "manifest.yml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest.yml was not created inside vault")
	}

	// Verify content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	content := string(data)
	if !contains(content, "version: 1") {
		t.Error("manifest missing version")
	}
	if !contains(content, ".config/nvim") {
		t.Error("manifest missing .config/nvim entry")
	}
	if !contains(content, ".zshrc") {
		t.Error("manifest missing .zshrc entry")
	}
}

func TestLoadManifest(t *testing.T) {
	vaultDir := t.TempDir()

	// Write a manifest first
	entries := []ManifestEntry{
		{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true, IsDir: true},
		{Path: ".bashrc", Git: config.GitModeRemove, Enabled: true, IsDir: false},
	}
	if err := WriteManifest(vaultDir, entries); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Load it back
	manifest, err := LoadManifest(vaultDir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if manifest.Version != 1 {
		t.Errorf("expected version 1, got %d", manifest.Version)
	}

	if len(manifest.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(manifest.Entries))
	}

	// Verify first entry
	if manifest.Entries[0].Path != ".config/nvim" {
		t.Errorf("expected path .config/nvim, got %s", manifest.Entries[0].Path)
	}
	if manifest.Entries[0].Git != config.GitModeDisable {
		t.Errorf("expected git mode disable, got %s", manifest.Entries[0].Git)
	}
	if !manifest.Entries[0].IsDir {
		t.Error("expected .config/nvim to be a directory")
	}

	// Verify second entry
	if manifest.Entries[1].Path != ".bashrc" {
		t.Errorf("expected path .bashrc, got %s", manifest.Entries[1].Path)
	}
	if manifest.Entries[1].Git != config.GitModeRemove {
		t.Errorf("expected git mode remove, got %s", manifest.Entries[1].Git)
	}
}

func TestLoadManifest_NotFound(t *testing.T) {
	vaultDir := t.TempDir()

	_, err := LoadManifest(vaultDir)
	if err == nil {
		t.Error("expected error when manifest doesn't exist")
	}
}

func TestManifestExists(t *testing.T) {
	vaultDir := t.TempDir()

	// Should not exist initially
	if ManifestExists(vaultDir) {
		t.Error("ManifestExists should return false for empty vault")
	}

	// Write manifest
	entries := []ManifestEntry{{Path: ".zshrc", Git: config.GitModeDisable, Enabled: true}}
	if err := WriteManifest(vaultDir, entries); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Should exist now
	if !ManifestExists(vaultDir) {
		t.Error("ManifestExists should return true after writing manifest")
	}
}

func TestManifest_ToWatching(t *testing.T) {
	manifest := &Manifest{
		Version: 1,
		Entries: []ManifestEntry{
			{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true, IsDir: true},
			{Path: ".zshrc", Git: config.GitModeRemove, Enabled: true, IsDir: false},
			{Path: ".bashrc", Git: config.GitModeDisable, Enabled: false, IsDir: false},
		},
	}

	watching := manifest.ToWatching()

	if len(watching) != 3 {
		t.Fatalf("expected 3 watching entries, got %d", len(watching))
	}

	// Verify conversion
	if watching[0].Path != ".config/nvim" {
		t.Errorf("expected path .config/nvim, got %s", watching[0].Path)
	}
	if watching[0].Git != config.GitModeDisable {
		t.Errorf("expected git mode disable, got %s", watching[0].Git)
	}
	if !watching[0].Enabled {
		t.Error("expected .config/nvim to be enabled")
	}

	if watching[2].Enabled {
		t.Error("expected .bashrc to be disabled")
	}
}

func TestFromWatching(t *testing.T) {
	watching := []config.Watched{
		{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true},
		{Path: ".zshrc", Git: config.GitModeRemove, Enabled: true},
		{Path: ".disabled", Git: config.GitModeDisable, Enabled: false}, // disabled, should be excluded
	}

	copiedItems := []CopiedItem{
		{Path: ".config/nvim", GitMode: config.GitModeDisable, IsDir: true},
		{Path: ".zshrc", GitMode: config.GitModeRemove, IsDir: false},
	}

	entries := FromWatching(watching, copiedItems)

	// Only enabled entries should be included
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (disabled excluded), got %d", len(entries))
	}

	// Verify IsDir is populated from copiedItems
	if !entries[0].IsDir {
		t.Error("expected .config/nvim to be marked as directory")
	}
	if entries[1].IsDir {
		t.Error("expected .zshrc to be marked as file")
	}
}

func TestManifestPath(t *testing.T) {
	vaultDir := "/home/user/.snapfig/vault"
	expected := "/home/user/.snapfig/vault/manifest.yml"

	if ManifestPath(vaultDir) != expected {
		t.Errorf("expected %s, got %s", expected, ManifestPath(vaultDir))
	}
}

// TestFreshMachineScenario simulates setting up on a new machine
// where config.Watching is empty but vault has a manifest.
func TestFreshMachineScenario(t *testing.T) {
	// Setup: create a vault with manifest (simulating after pull/clone)
	vaultDir := t.TempDir()

	// Create vault structure with some files
	nvimDir := filepath.Join(vaultDir, ".config", "nvim")
	if err := os.MkdirAll(nvimDir, 0755); err != nil {
		t.Fatalf("failed to create nvim dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nvimDir, "init.lua"), []byte("-- nvim config"), 0644); err != nil {
		t.Fatalf("failed to write init.lua: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, ".zshrc"), []byte("# zsh config"), 0644); err != nil {
		t.Fatalf("failed to write .zshrc: %v", err)
	}

	// Write manifest (as if it was synced from remote)
	manifestEntries := []ManifestEntry{
		{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true, IsDir: true},
		{Path: ".zshrc", Git: config.GitModeRemove, Enabled: true, IsDir: false},
	}
	if err := WriteManifest(vaultDir, manifestEntries); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Simulate new machine: empty config
	cfg := &config.Config{
		Git:      config.GitModeDisable,
		Watching: []config.Watched{}, // Empty! This is the key scenario
	}

	// Load manifest and reconstruct config
	manifest, err := LoadManifest(vaultDir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// Reconstruct watching from manifest
	cfg.Watching = manifest.ToWatching()

	// Verify config is now populated
	if len(cfg.Watching) != 2 {
		t.Fatalf("expected 2 watching entries after reconstruction, got %d", len(cfg.Watching))
	}

	if cfg.Watching[0].Path != ".config/nvim" {
		t.Errorf("expected first path to be .config/nvim, got %s", cfg.Watching[0].Path)
	}
	if cfg.Watching[0].Git != config.GitModeDisable {
		t.Errorf("expected git mode disable for nvim, got %s", cfg.Watching[0].Git)
	}

	if cfg.Watching[1].Path != ".zshrc" {
		t.Errorf("expected second path to be .zshrc, got %s", cfg.Watching[1].Path)
	}
	if cfg.Watching[1].Git != config.GitModeRemove {
		t.Errorf("expected git mode remove for zshrc, got %s", cfg.Watching[1].Git)
	}
}

// TestManifestRoundTrip tests write -> load -> verify cycle
func TestManifestRoundTrip(t *testing.T) {
	vaultDir := t.TempDir()

	original := []ManifestEntry{
		{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true, IsDir: true},
		{Path: ".config/fish", Git: config.GitModeRemove, Enabled: true, IsDir: true},
		{Path: ".zshrc", Git: config.GitModeDisable, Enabled: true, IsDir: false},
		{Path: ".gitconfig", Git: config.GitModeRemove, Enabled: false, IsDir: false},
	}

	// Write
	if err := WriteManifest(vaultDir, original); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Load
	manifest, err := LoadManifest(vaultDir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// Verify all entries match
	if len(manifest.Entries) != len(original) {
		t.Fatalf("entry count mismatch: expected %d, got %d", len(original), len(manifest.Entries))
	}

	for i, orig := range original {
		loaded := manifest.Entries[i]
		if loaded.Path != orig.Path {
			t.Errorf("entry %d path mismatch: expected %s, got %s", i, orig.Path, loaded.Path)
		}
		if loaded.Git != orig.Git {
			t.Errorf("entry %d git mismatch: expected %s, got %s", i, orig.Git, loaded.Git)
		}
		if loaded.Enabled != orig.Enabled {
			t.Errorf("entry %d enabled mismatch: expected %t, got %t", i, orig.Enabled, loaded.Enabled)
		}
		if loaded.IsDir != orig.IsDir {
			t.Errorf("entry %d isDir mismatch: expected %t, got %t", i, orig.IsDir, loaded.IsDir)
		}
	}
}

// TestServiceSyncConfigFromManifest tests the service method
func TestServiceSyncConfigFromManifest(t *testing.T) {
	vaultDir := t.TempDir()

	// Write manifest to vault
	entries := []ManifestEntry{
		{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true, IsDir: true},
		{Path: ".zshrc", Git: config.GitModeRemove, Enabled: true, IsDir: false},
	}
	if err := WriteManifest(vaultDir, entries); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Create service with empty config (new machine scenario)
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching:  []config.Watched{},
	}

	svc, err := NewService(cfg, "")
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Verify HasManifest works
	if !svc.HasManifest() {
		t.Error("HasManifest should return true")
	}

	// Verify config is empty before sync
	if len(svc.Config().Watching) != 0 {
		t.Error("config.Watching should be empty before sync")
	}

	// Sync from manifest
	if err := svc.SyncConfigFromManifest(); err != nil {
		t.Fatalf("SyncConfigFromManifest failed: %v", err)
	}

	// Verify config is populated after sync
	if len(svc.Config().Watching) != 2 {
		t.Fatalf("expected 2 watching entries after sync, got %d", len(svc.Config().Watching))
	}

	// Verify entries
	if svc.Config().Watching[0].Path != ".config/nvim" {
		t.Errorf("expected first path .config/nvim, got %s", svc.Config().Watching[0].Path)
	}
	if svc.Config().Watching[1].Path != ".zshrc" {
		t.Errorf("expected second path .zshrc, got %s", svc.Config().Watching[1].Path)
	}
}

// helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
