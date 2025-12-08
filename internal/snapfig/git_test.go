package snapfig

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestGitConfig(t *testing.T) {
	t.Helper()
	// Ensure git has user config for commits
	cmd := exec.Command("git", "config", "--global", "user.email")
	if err := cmd.Run(); err != nil {
		exec.Command("git", "config", "--global", "user.email", "test@test.com").Run()
		exec.Command("git", "config", "--global", "user.name", "Test User").Run()
	}
}

func TestInitVaultRepo(t *testing.T) {
	setupTestGitConfig(t)

	tests := []struct {
		name    string
		setup   func(dir string) error
		wantErr bool
	}{
		{
			name:    "init new repo",
			setup:   nil,
			wantErr: false,
		},
		{
			name: "init existing repo is noop",
			setup: func(dir string) error {
				cmd := exec.Command("git", "init")
				cmd.Dir = dir
				return cmd.Run()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "git-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			vaultDir := filepath.Join(tmpDir, "vault")

			if tt.setup != nil {
				if err := os.MkdirAll(vaultDir, 0755); err != nil {
					t.Fatalf("failed to create vault dir: %v", err)
				}
				if err := tt.setup(vaultDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err = InitVaultRepo(vaultDir)
			if tt.wantErr {
				if err == nil {
					t.Error("InitVaultRepo() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("InitVaultRepo() unexpected error: %v", err)
			}

			// Verify .git directory exists
			gitDir := filepath.Join(vaultDir, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				t.Error("InitVaultRepo() did not create .git directory")
			}
		})
	}
}

func TestCommitVault(t *testing.T) {
	setupTestGitConfig(t)

	tests := []struct {
		name       string
		setup      func(dir string) error
		wantCommit bool
		wantErr    bool
	}{
		{
			name: "commit with changes",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "test.txt"), []byte("content"), 0644)
			},
			wantCommit: true,
			wantErr:    false,
		},
		{
			name:       "commit with no changes is noop",
			setup:      nil,
			wantCommit: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "git-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			vaultDir := filepath.Join(tmpDir, "vault")

			// Initialize repo
			if err := InitVaultRepo(vaultDir); err != nil {
				t.Fatalf("InitVaultRepo() failed: %v", err)
			}

			// Make initial commit to have a valid repo state
			os.WriteFile(filepath.Join(vaultDir, ".gitkeep"), []byte(""), 0644)
			exec.Command("git", "-C", vaultDir, "add", "-A").Run()
			exec.Command("git", "-C", vaultDir, "commit", "-m", "initial").Run()

			if tt.setup != nil {
				if err := tt.setup(vaultDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err = CommitVault(vaultDir, "test commit")
			if tt.wantErr {
				if err == nil {
					t.Error("CommitVault() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("CommitVault() unexpected error: %v", err)
			}
		})
	}
}

func TestHasRemote(t *testing.T) {
	setupTestGitConfig(t)

	tests := []struct {
		name       string
		setup      func(dir string) error
		wantRemote bool
		wantURL    string
	}{
		{
			name:       "no remote configured",
			setup:      nil,
			wantRemote: false,
			wantURL:    "",
		},
		{
			name: "remote configured",
			setup: func(dir string) error {
				cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
				cmd.Dir = dir
				return cmd.Run()
			},
			wantRemote: true,
			wantURL:    "https://github.com/test/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "git-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			vaultDir := filepath.Join(tmpDir, "vault")
			if err := InitVaultRepo(vaultDir); err != nil {
				t.Fatalf("InitVaultRepo() failed: %v", err)
			}

			if tt.setup != nil {
				if err := tt.setup(vaultDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			hasRemote, url, err := HasRemote(vaultDir)
			if err != nil {
				t.Fatalf("HasRemote() unexpected error: %v", err)
			}

			if hasRemote != tt.wantRemote {
				t.Errorf("HasRemote() = %v, want %v", hasRemote, tt.wantRemote)
			}
			if url != tt.wantURL {
				t.Errorf("HasRemote() url = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

func TestSetRemote(t *testing.T) {
	setupTestGitConfig(t)

	tests := []struct {
		name    string
		setup   func(dir string) error
		url     string
		wantErr bool
	}{
		{
			name:    "add new remote",
			setup:   nil,
			url:     "https://github.com/test/repo.git",
			wantErr: false,
		},
		{
			name: "update existing remote",
			setup: func(dir string) error {
				cmd := exec.Command("git", "remote", "add", "origin", "https://old-url.com/repo.git")
				cmd.Dir = dir
				return cmd.Run()
			},
			url:     "https://new-url.com/repo.git",
			wantErr: false,
		},
		{
			name: "same url is noop",
			setup: func(dir string) error {
				cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
				cmd.Dir = dir
				return cmd.Run()
			},
			url:     "https://github.com/test/repo.git",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "git-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			vaultDir := filepath.Join(tmpDir, "vault")
			if err := InitVaultRepo(vaultDir); err != nil {
				t.Fatalf("InitVaultRepo() failed: %v", err)
			}

			if tt.setup != nil {
				if err := tt.setup(vaultDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err = SetRemote(vaultDir, tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("SetRemote() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("SetRemote() unexpected error: %v", err)
			}

			// Verify remote is set correctly
			hasRemote, url, err := HasRemote(vaultDir)
			if err != nil {
				t.Fatalf("HasRemote() failed: %v", err)
			}
			if !hasRemote {
				t.Error("SetRemote() did not set remote")
			}
			if url != tt.url {
				t.Errorf("SetRemote() url = %q, want %q", url, tt.url)
			}
		})
	}
}

func TestPushVaultNoRemote(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	if err := InitVaultRepo(vaultDir); err != nil {
		t.Fatalf("InitVaultRepo() failed: %v", err)
	}

	err = PushVault(vaultDir)
	if err == nil {
		t.Error("PushVault() expected error when no remote configured, got nil")
	}
}

func TestPullVaultNoVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")

	// PullVault without remote URL should fail
	_, err = PullVault(vaultDir)
	if err == nil {
		t.Error("PullVault() expected error when vault doesn't exist, got nil")
	}
}

func TestPullVaultWithRemoteNoVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")

	// PullVaultWithRemote with invalid URL should fail
	_, err = PullVaultWithRemote(vaultDir, "invalid-url")
	if err == nil {
		t.Error("PullVaultWithRemote() expected error with invalid URL, got nil")
	}
}

func TestPullVaultExistingNoRemote(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	if err := InitVaultRepo(vaultDir); err != nil {
		t.Fatalf("InitVaultRepo() failed: %v", err)
	}

	// Make initial commit
	os.WriteFile(filepath.Join(vaultDir, ".gitkeep"), []byte(""), 0644)
	exec.Command("git", "-C", vaultDir, "add", "-A").Run()
	exec.Command("git", "-C", vaultDir, "commit", "-m", "initial").Run()

	// Pull should fail without remote
	_, err = PullVaultWithRemote(vaultDir, "")
	if err == nil {
		t.Error("PullVaultWithRemote() expected error when no remote, got nil")
	}
}

func TestSetRemoteCreatesRepo(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	// Don't create repo first

	err = SetRemote(vaultDir, "https://github.com/test/repo.git")
	if err != nil {
		t.Fatalf("SetRemote() unexpected error: %v", err)
	}

	// Verify repo was created
	gitDir := filepath.Join(vaultDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("SetRemote() did not create git repo")
	}
}

func TestCommitVaultNotRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Try to commit in non-git directory
	err = CommitVault(tmpDir, "test")
	if err == nil {
		t.Error("CommitVault() expected error in non-git directory, got nil")
	}
}

func TestPushVaultWithLocalRemote(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a bare repo as "remote"
	bareDir := filepath.Join(tmpDir, "remote.git")
	cmd := exec.Command("git", "init", "--bare", bareDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create bare repo: %v", err)
	}

	// Create vault repo
	vaultDir := filepath.Join(tmpDir, "vault")
	if err := InitVaultRepo(vaultDir); err != nil {
		t.Fatalf("InitVaultRepo() failed: %v", err)
	}

	// Add remote
	if err := SetRemote(vaultDir, bareDir); err != nil {
		t.Fatalf("SetRemote() failed: %v", err)
	}

	// Create a file and commit
	os.WriteFile(filepath.Join(vaultDir, "test.txt"), []byte("content"), 0644)
	if err := CommitVault(vaultDir, "initial commit"); err != nil {
		t.Fatalf("CommitVault() failed: %v", err)
	}

	// Push should succeed
	if err := PushVault(vaultDir); err != nil {
		t.Fatalf("PushVault() unexpected error: %v", err)
	}
}

func TestPushVaultGetBranch(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create bare repo
	bareDir := filepath.Join(tmpDir, "remote.git")
	exec.Command("git", "init", "--bare", bareDir).Run()

	// Create vault with custom branch name
	vaultDir := filepath.Join(tmpDir, "vault")
	InitVaultRepo(vaultDir)

	// Create initial commit on master/main
	os.WriteFile(filepath.Join(vaultDir, "test.txt"), []byte("content"), 0644)
	exec.Command("git", "-C", vaultDir, "add", "-A").Run()
	exec.Command("git", "-C", vaultDir, "commit", "-m", "initial").Run()

	// Add remote and push
	SetRemote(vaultDir, bareDir)
	err = PushVault(vaultDir)
	if err != nil {
		t.Fatalf("PushVault() error: %v", err)
	}
}

func TestPullVaultClone(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source repo with content
	sourceDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(sourceDir, 0755)
	exec.Command("git", "init", sourceDir).Run()
	os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("hello"), 0644)
	exec.Command("git", "-C", sourceDir, "add", "-A").Run()
	exec.Command("git", "-C", sourceDir, "commit", "-m", "initial").Run()

	// Clone to vault (vault doesn't exist yet)
	vaultDir := filepath.Join(tmpDir, "vault")

	result, err := PullVaultWithRemote(vaultDir, sourceDir)
	if err != nil {
		t.Fatalf("PullVaultWithRemote() error: %v", err)
	}

	if !result.Cloned {
		t.Error("result.Cloned should be true")
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(vaultDir, "test.txt")); os.IsNotExist(err) {
		t.Error("cloned file should exist")
	}
}

func TestPullVaultExisting(t *testing.T) {
	setupTestGitConfig(t)

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create bare repo as remote
	bareDir := filepath.Join(tmpDir, "remote.git")
	exec.Command("git", "init", "--bare", bareDir).Run()

	// Create vault, commit, push
	vaultDir := filepath.Join(tmpDir, "vault")
	InitVaultRepo(vaultDir)
	os.WriteFile(filepath.Join(vaultDir, "test.txt"), []byte("content"), 0644)
	exec.Command("git", "-C", vaultDir, "add", "-A").Run()
	exec.Command("git", "-C", vaultDir, "commit", "-m", "initial").Run()
	SetRemote(vaultDir, bareDir)
	exec.Command("git", "-C", vaultDir, "push", "-u", "origin", "master").Run()

	// Pull should work
	result, err := PullVaultWithRemote(vaultDir, bareDir)
	if err != nil {
		t.Fatalf("PullVaultWithRemote() error: %v", err)
	}

	if result.Cloned {
		t.Error("result.Cloned should be false for existing repo")
	}
}

func TestUrlWithToken(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		token    string
		wantURL  string
		wantSubs []string // substrings that should be in the result
	}{
		{
			name:    "no token returns original URL",
			url:     "https://github.com/user/repo.git",
			token:   "",
			wantURL: "https://github.com/user/repo.git",
		},
		{
			name:    "SSH URL with token converts to HTTPS",
			url:     "git@github.com:user/repo.git",
			token:   "ghp_test123",
			wantURL: "https://x-access-token:ghp_test123@github.com/user/repo.git",
		},
		{
			name:     "HTTPS URL with token adds auth",
			url:      "https://github.com/user/repo.git",
			token:    "ghp_test456",
			wantSubs: []string{"x-access-token", "ghp_test456", "github.com"},
		},
		{
			name:    "unknown format returns original",
			url:     "file:///local/path",
			token:   "token123",
			wantURL: "file:///local/path",
		},
		{
			name:    "empty URL with token",
			url:     "",
			token:   "token",
			wantURL: "",
		},
		{
			name:     "HTTPS URL with path components",
			url:      "https://gitlab.com/group/subgroup/repo.git",
			token:    "glpat_token",
			wantSubs: []string{"x-access-token", "glpat_token", "gitlab.com/group/subgroup/repo.git"},
		},
		{
			name:    "SSH URL with different host",
			url:     "git@gitlab.com:user/project.git",
			token:   "glpat_abc",
			wantURL: "https://x-access-token:glpat_abc@gitlab.com/user/project.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := urlWithToken(tt.url, tt.token)

			if tt.wantURL != "" {
				if result != tt.wantURL {
					t.Errorf("urlWithToken() = %q, want %q", result, tt.wantURL)
				}
			}

			for _, sub := range tt.wantSubs {
				if !strings.Contains(result, sub) {
					t.Errorf("urlWithToken() = %q, should contain %q", result, sub)
				}
			}
		})
	}
}

func TestUrlWithTokenParseError(t *testing.T) {
	// Test with malformed HTTPS URL that will fail url.Parse
	result := urlWithToken("https://[::1]:invalid-port/path", "token")
	// Should return original URL on parse error
	if result != "https://[::1]:invalid-port/path" {
		t.Errorf("urlWithToken() should return original URL on parse error, got %q", result)
	}
}
