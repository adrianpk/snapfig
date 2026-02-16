package snapfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestNewCopier(t *testing.T) {
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
			copier, err := NewCopier(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("NewCopier() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewCopier() unexpected error: %v", err)
			}
			if copier == nil {
				t.Error("NewCopier() returned nil copier")
			}
		})
	}
}

func TestCopySingleFile(t *testing.T) {
	setupTestGitConfig(t)

	// Create temp directories for home and vault
	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	// Create source file
	srcFile := filepath.Join(homeDir, ".testrc")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
		},
	}

	// Create copier with custom home
	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if len(result.Copied) != 1 {
		t.Errorf("Copy() copied %d paths, want 1", len(result.Copied))
	}
	if result.FilesUpdated != 1 {
		t.Errorf("Copy() updated %d files, want 1", result.FilesUpdated)
	}

	// Verify file was copied
	dstFile := filepath.Join(vaultDir, ".testrc")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("copied file content = %q, want %q", string(content), "test content")
	}
}

func TestCopyDirectory(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create source directory with files
	srcDir := filepath.Join(homeDir, ".config", "myapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "config.yml"), []byte("key: value"), 0644)
	os.WriteFile(filepath.Join(srcDir, "settings.json"), []byte("{}"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myapp", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if result.FilesUpdated != 2 {
		t.Errorf("Copy() updated %d files, want 2", result.FilesUpdated)
	}

	// Verify files were copied
	dstConfig := filepath.Join(vaultDir, ".config", "myapp", "config.yml")
	if _, err := os.Stat(dstConfig); os.IsNotExist(err) {
		t.Error("Copy() did not copy config.yml")
	}

	dstSettings := filepath.Join(vaultDir, ".config", "myapp", "settings.json")
	if _, err := os.Stat(dstSettings); os.IsNotExist(err) {
		t.Error("Copy() did not copy settings.json")
	}
}

func TestCopySkipsNonExistent(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".nonexistent", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if len(result.Skipped) != 1 {
		t.Errorf("Copy() skipped %d paths, want 1", len(result.Skipped))
	}
}

func TestCopySkipsDisabled(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create source file
	os.MkdirAll(homeDir, 0755)
	os.WriteFile(filepath.Join(homeDir, ".testrc"), []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: false}, // Disabled
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if len(result.Copied) != 0 {
		t.Errorf("Copy() copied %d paths, want 0 (disabled)", len(result.Copied))
	}
}

func TestCopyGitModeRemove(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create source directory with .git
	srcDir := filepath.Join(homeDir, ".config", "myrepo")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0644)

	// Create .git directory
	gitDir := filepath.Join(srcDir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeRemove,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myrepo", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	// .git should NOT be copied
	dstGit := filepath.Join(vaultDir, ".config", "myrepo", ".git")
	if _, err := os.Stat(dstGit); !os.IsNotExist(err) {
		t.Error("Copy() with GitModeRemove should not copy .git directory")
	}

	// But file.txt should be copied
	dstFile := filepath.Join(vaultDir, ".config", "myrepo", "file.txt")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Copy() did not copy file.txt")
	}

	if result.FilesUpdated != 1 {
		t.Errorf("Copy() updated %d files, want 1", result.FilesUpdated)
	}
}

func TestCopyGitModeDisable(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create source directory with .git
	srcDir := filepath.Join(homeDir, ".config", "myrepo")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0644)

	gitDir := filepath.Join(srcDir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myrepo", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	_, err = copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	// .git should be renamed to .git_disabled
	dstGit := filepath.Join(vaultDir, ".config", "myrepo", ".git")
	if _, err := os.Stat(dstGit); !os.IsNotExist(err) {
		t.Error("Copy() with GitModeDisable should rename .git")
	}

	dstGitDisabled := filepath.Join(vaultDir, ".config", "myrepo", ".git_disabled")
	if _, err := os.Stat(dstGitDisabled); os.IsNotExist(err) {
		t.Error("Copy() with GitModeDisable should create .git_disabled")
	}
}

func TestShouldCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shouldcopy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	tests := []struct {
		name       string
		setup      func()
		wantCopy   bool
		wantErr    bool
	}{
		{
			name: "destination does not exist",
			setup: func() {
				os.WriteFile(src, []byte("content"), 0644)
				os.Remove(dst)
			},
			wantCopy: true,
		},
		{
			name: "files are identical",
			setup: func() {
				content := []byte("same content")
				os.WriteFile(src, content, 0644)
				os.WriteFile(dst, content, 0644)
				// Set same mod time
				now := time.Now()
				os.Chtimes(src, now, now)
				os.Chtimes(dst, now, now)
			},
			wantCopy: false,
		},
		{
			name: "source is newer",
			setup: func() {
				os.WriteFile(dst, []byte("old content"), 0644)
				time.Sleep(10 * time.Millisecond)
				os.WriteFile(src, []byte("new content"), 0644)
			},
			wantCopy: true,
		},
		{
			name: "size differs",
			setup: func() {
				os.WriteFile(src, []byte("longer content here"), 0644)
				os.WriteFile(dst, []byte("short"), 0644)
			},
			wantCopy: true,
		},
		{
			name: "source does not exist",
			setup: func() {
				os.Remove(src)
				os.WriteFile(dst, []byte("content"), 0644)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			needsCopy, err := shouldCopy(src, dst)
			if tt.wantErr {
				if err == nil {
					t.Error("shouldCopy() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("shouldCopy() unexpected error: %v", err)
			}
			if needsCopy != tt.wantCopy {
				t.Errorf("shouldCopy() = %v, want %v", needsCopy, tt.wantCopy)
			}
		})
	}
}

func TestCopySmartSkipsUnchanged(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)

	// Create source file
	srcFile := filepath.Join(homeDir, ".testrc")
	os.WriteFile(srcFile, []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	// First copy
	result1, err := copier.Copy()
	if err != nil {
		t.Fatalf("First Copy() error: %v", err)
	}
	if result1.FilesUpdated != 1 {
		t.Errorf("First Copy() updated %d files, want 1", result1.FilesUpdated)
	}

	// Second copy without changes should skip
	result2, err := copier.Copy()
	if err != nil {
		t.Fatalf("Second Copy() error: %v", err)
	}
	if result2.FilesUpdated != 0 {
		t.Errorf("Second Copy() updated %d files, want 0 (unchanged)", result2.FilesUpdated)
	}
	if result2.FilesSkipped != 1 {
		t.Errorf("Second Copy() skipped %d files, want 1", result2.FilesSkipped)
	}
}

func TestCopyRemovesStaleFiles(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	// Create source directory with initial files
	srcDir := filepath.Join(homeDir, ".config", "myapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "keep.txt"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(srcDir, "delete.txt"), []byte("delete"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/myapp", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	// First copy
	_, err = copier.Copy()
	if err != nil {
		t.Fatalf("First Copy() error: %v", err)
	}

	// Remove a file from source
	os.Remove(filepath.Join(srcDir, "delete.txt"))

	// Second copy should remove stale file
	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Second Copy() error: %v", err)
	}

	if result.FilesRemoved != 1 {
		t.Errorf("Copy() removed %d files, want 1", result.FilesRemoved)
	}

	// Verify stale file was removed
	staleFile := filepath.Join(vaultDir, ".config", "myapp", "delete.txt")
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Error("Copy() did not remove stale file")
	}
}

func TestCopyCreatesManifest(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	snapfigDir := filepath.Dir(vaultDir)
	os.MkdirAll(homeDir, 0755)

	os.WriteFile(filepath.Join(homeDir, ".testrc"), []byte("content"), 0644)

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: snapfigDir,
	}

	_, err = copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	// Verify manifest was created inside vault (YAML format)
	manifestPath := filepath.Join(vaultDir, "manifest.yml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Copy() did not create manifest.yml inside vault")
	}

	// Verify manifest can be loaded and contains correct data
	manifest, err := LoadManifest(vaultDir)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if len(manifest.Entries) != 1 {
		t.Errorf("expected 1 manifest entry, got %d", len(manifest.Entries))
	}

	if manifest.Entries[0].Path != ".testrc" {
		t.Errorf("expected entry path .testrc, got %s", manifest.Entries[0].Path)
	}
}

func TestCopySymlinkCreatesMarker(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "copier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")

	srcDir := filepath.Join(homeDir, ".config", "themes")
	os.MkdirAll(srcDir, 0755)

	targetDir := filepath.Join(tmpDir, "external", "catppuccin")
	os.MkdirAll(targetDir, 0755)
	os.WriteFile(filepath.Join(targetDir, "theme.yml"), []byte("colors: mocha"), 0644)

	symlinkPath := filepath.Join(srcDir, "current")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	cfg := &config.Config{
		Git:       config.GitModeDisable,
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/themes", Enabled: true},
		},
	}

	copier := &Copier{
		cfg:        cfg,
		home:       homeDir,
		vaultDir:   vaultDir,
		snapfigDir: filepath.Dir(vaultDir),
	}

	result, err := copier.Copy()
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	markerPath := filepath.Join(vaultDir, ".config", "themes", "current.snapfig-symlink")
	content, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}

	expected := "ln -s " + targetDir + " current\n"
	if string(content) != expected {
		t.Errorf("marker content = %q, want %q", string(content), expected)
	}

	targetCopied := filepath.Join(vaultDir, ".config", "themes", "current", "theme.yml")
	if _, err := os.Stat(targetCopied); !os.IsNotExist(err) {
		t.Error("symlink target should not be copied")
	}

	if result.FilesUpdated != 1 {
		t.Errorf("Copy() updated %d files, want 1 (marker only)", result.FilesUpdated)
	}
}
