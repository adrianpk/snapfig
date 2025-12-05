package snapfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestNewRestorer(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Git: config.GitModeDisable,
			},
			wantErr: false,
		},
		{
			name: "config with custom vault path",
			cfg: &config.Config{
				Git:       config.GitModeRemove,
				VaultPath: "/tmp/test-vault",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restorer, err := NewRestorer(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("NewRestorer() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewRestorer() unexpected error: %v", err)
			}
			if restorer == nil {
				t.Error("NewRestorer() returned nil restorer")
			}
		})
	}
}

func TestRestoreSingleFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	// Create source file in vault
	vaultFile := filepath.Join(vaultDir, ".testrc")
	if err := os.WriteFile(vaultFile, []byte("restored content"), 0644); err != nil {
		t.Fatalf("failed to create vault file: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	result, err := restorer.Restore()
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	if len(result.Restored) != 1 {
		t.Errorf("Restore() restored %d paths, want 1", len(result.Restored))
	}
	if result.FilesUpdated != 1 {
		t.Errorf("Restore() updated %d files, want 1", result.FilesUpdated)
	}

	// Verify file was restored
	dstFile := filepath.Join(homeDir, ".testrc")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(content) != "restored content" {
		t.Errorf("restored file content = %q, want %q", string(content), "restored content")
	}
}

func TestRestoreDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create source directory in vault
	vaultSubDir := filepath.Join(vaultDir, ".config", "myapp")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "config.yml"), []byte("key: value"), 0644)
	os.WriteFile(filepath.Join(vaultSubDir, "settings.json"), []byte("{}"), 0644)

	// Create destination home dir
	os.MkdirAll(homeDir, 0755)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myapp", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	result, err := restorer.Restore()
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	if result.FilesUpdated != 2 {
		t.Errorf("Restore() updated %d files, want 2", result.FilesUpdated)
	}

	// Verify files were restored
	dstConfig := filepath.Join(homeDir, ".config", "myapp", "config.yml")
	if _, err := os.Stat(dstConfig); os.IsNotExist(err) {
		t.Error("Restore() did not restore config.yml")
	}
}

func TestRestoreSkipsNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".nonexistent", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	result, err := restorer.Restore()
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	if len(result.Skipped) != 1 {
		t.Errorf("Restore() skipped %d paths, want 1", len(result.Skipped))
	}
}

func TestRestoreSkipsDisabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	os.WriteFile(filepath.Join(vaultDir, ".testrc"), []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: false}, // Disabled
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	result, err := restorer.Restore()
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	if len(result.Restored) != 0 {
		t.Errorf("Restore() restored %d paths, want 0 (disabled)", len(result.Restored))
	}
}

func TestRestoreRevertsGitDisabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create vault structure with .git_disabled
	vaultSubDir := filepath.Join(vaultDir, ".config", "myrepo")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "file.txt"), []byte("content"), 0644)

	gitDisabled := filepath.Join(vaultSubDir, ".git_disabled")
	os.MkdirAll(gitDisabled, 0755)
	os.WriteFile(filepath.Join(gitDisabled, "config"), []byte("git config"), 0644)

	os.MkdirAll(homeDir, 0755)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myrepo", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	_, err = restorer.Restore()
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	// .git_disabled should be restored as .git
	dstGit := filepath.Join(homeDir, ".config", "myrepo", ".git")
	if _, err := os.Stat(dstGit); os.IsNotExist(err) {
		t.Error("Restore() should revert .git_disabled to .git")
	}

	dstGitDisabled := filepath.Join(homeDir, ".config", "myrepo", ".git_disabled")
	if _, err := os.Stat(dstGitDisabled); !os.IsNotExist(err) {
		t.Error("Restore() should not create .git_disabled in destination")
	}
}

func TestShouldRestore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shouldrestore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	tests := []struct {
		name        string
		setup       func()
		wantRestore bool
		wantErr     bool
	}{
		{
			name: "destination does not exist",
			setup: func() {
				os.WriteFile(src, []byte("content"), 0644)
				os.Remove(dst)
			},
			wantRestore: true,
		},
		{
			name: "files are identical",
			setup: func() {
				content := []byte("same content")
				os.WriteFile(src, content, 0644)
				os.WriteFile(dst, content, 0644)
				now := time.Now()
				os.Chtimes(src, now, now)
				os.Chtimes(dst, now, now)
			},
			wantRestore: false,
		},
		{
			name: "mod time differs",
			setup: func() {
				content := []byte("same content")
				os.WriteFile(src, content, 0644)
				os.WriteFile(dst, content, 0644)
				// Different mod times
				os.Chtimes(src, time.Now(), time.Now())
				os.Chtimes(dst, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))
			},
			wantRestore: true,
		},
		{
			name: "size differs",
			setup: func() {
				os.WriteFile(src, []byte("longer content"), 0644)
				os.WriteFile(dst, []byte("short"), 0644)
			},
			wantRestore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			needsRestore, err := shouldRestore(src, dst)
			if tt.wantErr {
				if err == nil {
					t.Error("shouldRestore() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("shouldRestore() unexpected error: %v", err)
			}
			if needsRestore != tt.wantRestore {
				t.Errorf("shouldRestore() = %v, want %v", needsRestore, tt.wantRestore)
			}
		})
	}
}

func TestRestoreSmartSkipsUnchanged(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	vaultFile := filepath.Join(vaultDir, ".testrc")
	os.WriteFile(vaultFile, []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	// First restore
	result1, err := restorer.Restore()
	if err != nil {
		t.Fatalf("First Restore() error: %v", err)
	}
	if result1.FilesUpdated != 1 {
		t.Errorf("First Restore() updated %d files, want 1", result1.FilesUpdated)
	}

	// Second restore should skip unchanged
	result2, err := restorer.Restore()
	if err != nil {
		t.Fatalf("Second Restore() error: %v", err)
	}
	if result2.FilesUpdated != 0 {
		t.Errorf("Second Restore() updated %d files, want 0", result2.FilesUpdated)
	}
	if result2.FilesSkipped != 1 {
		t.Errorf("Second Restore() skipped %d files, want 1", result2.FilesSkipped)
	}
}

func TestListVaultEntries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	// Create vault entries
	if err := os.WriteFile(filepath.Join(vaultDir, ".testrc"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create .testrc: %v", err)
	}
	vaultSubDir := filepath.Join(vaultDir, ".config", "myapp")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "config.yml"), []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
			{Path: ".config/myapp", Enabled: true},
			{Path: ".nonexistent", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	entries, err := restorer.ListVaultEntries()
	if err != nil {
		t.Fatalf("ListVaultEntries() error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("ListVaultEntries() returned %d entries, want 2", len(entries))
	}

	// Verify entries
	foundFile := false
	foundDir := false
	for _, e := range entries {
		if e.Path == ".testrc" && !e.IsDir {
			foundFile = true
		}
		if e.Path == ".config/myapp" && e.IsDir {
			foundDir = true
			if len(e.Children) == 0 {
				t.Error("ListVaultEntries() directory should have children")
			}
		}
	}

	if !foundFile {
		t.Error("ListVaultEntries() missing .testrc file")
	}
	if !foundDir {
		t.Error("ListVaultEntries() missing .config/myapp directory")
	}
}

func TestRestoreSelective(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	// Create vault entries
	if err := os.WriteFile(filepath.Join(vaultDir, ".testrc"), []byte("testrc"), 0644); err != nil {
		t.Fatalf("failed to create .testrc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, ".bashrc"), []byte("bashrc"), 0644); err != nil {
		t.Fatalf("failed to create .bashrc: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
			{Path: ".bashrc", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	// Only restore .testrc
	result, err := restorer.RestoreSelective([]string{".testrc"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	if len(result.Restored) != 1 {
		t.Errorf("RestoreSelective() restored %d paths, want 1", len(result.Restored))
	}

	// .testrc should exist
	if _, err := os.Stat(filepath.Join(homeDir, ".testrc")); os.IsNotExist(err) {
		t.Error("RestoreSelective() should restore .testrc")
	}

	// .bashrc should NOT exist (not selected)
	if _, err := os.Stat(filepath.Join(homeDir, ".bashrc")); !os.IsNotExist(err) {
		t.Error("RestoreSelective() should not restore .bashrc")
	}
}

func TestRestoreSelectiveDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	// Create vault directory with files
	vaultSubDir := filepath.Join(vaultDir, ".config", "myapp")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "config.yml"), []byte("config"), 0644)
	os.WriteFile(filepath.Join(vaultSubDir, "other.txt"), []byte("other"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myapp", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	// Only restore one file from the directory
	result, err := restorer.RestoreSelective([]string{".config/myapp/config.yml"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	if len(result.Restored) != 1 {
		t.Errorf("RestoreSelective() restored %d paths, want 1", len(result.Restored))
	}

	// config.yml should exist
	if _, err := os.Stat(filepath.Join(homeDir, ".config", "myapp", "config.yml")); os.IsNotExist(err) {
		t.Error("RestoreSelective() should restore config.yml")
	}

	// other.txt should NOT exist
	if _, err := os.Stat(filepath.Join(homeDir, ".config", "myapp", "other.txt")); !os.IsNotExist(err) {
		t.Error("RestoreSelective() should not restore other.txt")
	}
}

func TestRestoreSelectiveEntireDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	// Create vault directory with files
	vaultSubDir := filepath.Join(vaultDir, ".config", "myapp")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "config.yml"), []byte("config"), 0644)
	os.WriteFile(filepath.Join(vaultSubDir, "other.txt"), []byte("other"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myapp", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	// Restore entire directory
	result, err := restorer.RestoreSelective([]string{".config/myapp"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	if len(result.Restored) != 1 {
		t.Errorf("RestoreSelective() restored %d paths, want 1", len(result.Restored))
	}

	// Both files should exist
	if _, err := os.Stat(filepath.Join(homeDir, ".config", "myapp", "config.yml")); os.IsNotExist(err) {
		t.Error("RestoreSelective() should restore config.yml")
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".config", "myapp", "other.txt")); os.IsNotExist(err) {
		t.Error("RestoreSelective() should restore other.txt")
	}
}

func TestRestoreSelectiveSkipsNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".nonexistent", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	result, err := restorer.RestoreSelective([]string{".nonexistent"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	if len(result.Restored) != 0 {
		t.Errorf("RestoreSelective() restored %d paths, want 0", len(result.Restored))
	}
}

func TestRestoreSelectiveSkipsDisabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	os.WriteFile(filepath.Join(vaultDir, ".testrc"), []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: false},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	result, err := restorer.RestoreSelective([]string{".testrc"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	if len(result.Restored) != 0 {
		t.Errorf("RestoreSelective() should skip disabled, restored %d", len(result.Restored))
	}
}

func TestRestoreSelectiveDirWithGitDisabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	// Create vault directory with .git_disabled subdirectory
	vaultSubDir := filepath.Join(vaultDir, ".config", "myrepo")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "file.txt"), []byte("content"), 0644)

	gitDisabled := filepath.Join(vaultSubDir, ".git_disabled")
	os.MkdirAll(gitDisabled, 0755)
	os.WriteFile(filepath.Join(gitDisabled, "config"), []byte("git config"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myrepo", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	// Restore just the .git_disabled directory
	result, err := restorer.RestoreSelective([]string{".config/myrepo/.git_disabled"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	if len(result.Restored) != 1 {
		t.Errorf("RestoreSelective() restored %d paths, want 1", len(result.Restored))
	}

	// .git_disabled should be restored as .git
	dstGit := filepath.Join(homeDir, ".config", "myrepo", ".git")
	if _, err := os.Stat(dstGit); os.IsNotExist(err) {
		t.Error("RestoreSelective() should restore .git_disabled as .git")
	}
}

func TestRestoreSelectiveNoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	// Create vault directory with files
	vaultSubDir := filepath.Join(vaultDir, ".config", "myapp")
	os.MkdirAll(vaultSubDir, 0755)
	os.WriteFile(filepath.Join(vaultSubDir, "config.yml"), []byte("config"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myapp", Enabled: true},
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	// Restore a file that doesn't match any in the directory
	result, err := restorer.RestoreSelective([]string{".config/myapp/nonexistent.yml"})
	if err != nil {
		t.Fatalf("RestoreSelective() error: %v", err)
	}

	// Should be marked as skipped since nothing was restored
	if len(result.Skipped) != 1 {
		t.Errorf("RestoreSelective() skipped %d paths, want 1", len(result.Skipped))
	}
}

func TestShouldRestoreSourceStatError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Non-existent source
	_, err = shouldRestore(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dst"))
	if err == nil {
		t.Error("shouldRestore() should return error for non-existent source")
	}
}

func TestListVaultEntriesDisabledItems(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	os.WriteFile(filepath.Join(vaultDir, ".testrc"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(vaultDir, ".bashrc"), []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
			{Path: ".bashrc", Enabled: false}, // Disabled
		},
	}

	restorer := &Restorer{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}

	entries, err := restorer.ListVaultEntries()
	if err != nil {
		t.Fatalf("ListVaultEntries() error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("ListVaultEntries() returned %d entries, want 1 (disabled skipped)", len(entries))
	}
	if len(entries) > 0 && entries[0].Path != ".testrc" {
		t.Errorf("ListVaultEntries() entry = %q, want .testrc", entries[0].Path)
	}
}
