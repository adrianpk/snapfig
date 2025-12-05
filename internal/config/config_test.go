package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	got, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("DefaultConfigDir() error = %v", err)
	}

	want := filepath.Join(home, ".config", "snapfig")
	if got != want {
		t.Errorf("DefaultConfigDir() = %q, want %q", got, want)
	}
}

func TestDefaultVaultDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	got, err := DefaultVaultDir()
	if err != nil {
		t.Fatalf("DefaultVaultDir() error = %v", err)
	}

	want := filepath.Join(home, ".snapfig", "vault")
	if got != want {
		t.Errorf("DefaultVaultDir() = %q, want %q", got, want)
	}
}

func TestDefaultSnapfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	got, err := DefaultSnapfigDir()
	if err != nil {
		t.Fatalf("DefaultSnapfigDir() error = %v", err)
	}

	want := filepath.Join(home, ".snapfig")
	if got != want {
		t.Errorf("DefaultSnapfigDir() = %q, want %q", got, want)
	}
}

func TestPidFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	got, err := PidFilePath()
	if err != nil {
		t.Fatalf("PidFilePath() error = %v", err)
	}

	want := filepath.Join(home, ".snapfig", "daemon.pid")
	if got != want {
		t.Errorf("PidFilePath() = %q, want %q", got, want)
	}
}

func TestLogFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	got, err := LogFilePath()
	if err != nil {
		t.Fatalf("LogFilePath() error = %v", err)
	}

	want := filepath.Join(home, ".snapfig", "daemon.log")
	if got != want {
		t.Errorf("LogFilePath() = %q, want %q", got, want)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantGit     GitMode
		wantRemote  string
		wantErr     bool
		wantWatched int
	}{
		{
			name: "valid config with all fields",
			content: `git: remove
remote: git@github.com:user/repo.git
watching:
  - path: .config/nvim
    enabled: true
  - path: .bashrc
    enabled: false
`,
			wantGit:     GitModeRemove,
			wantRemote:  "git@github.com:user/repo.git",
			wantWatched: 2,
			wantErr:     false,
		},
		{
			name: "default git mode when empty",
			content: `watching:
  - path: .config
    enabled: true
`,
			wantGit:     GitModeDisable,
			wantWatched: 1,
			wantErr:     false,
		},
		{
			name: "config with daemon settings",
			content: `git: disable
daemon:
  copy_interval: 1h
  push_interval: 24h
  auto_restore: true
watching: []
`,
			wantGit:     GitModeDisable,
			wantWatched: 0,
			wantErr:     false,
		},
		{
			name:    "invalid yaml",
			content: `git: [invalid`,
			wantErr: true,
		},
		{
			name:    "empty config",
			content: ``,
			wantGit: GitModeDisable,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp("", "config-*.yml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}
			tmpFile.Close()

			cfg, err := Load(tmpFile.Name())
			if tt.wantErr {
				if err == nil {
					t.Error("Load() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}

			if cfg.Git != tt.wantGit {
				t.Errorf("Load() Git = %q, want %q", cfg.Git, tt.wantGit)
			}
			if cfg.Remote != tt.wantRemote {
				t.Errorf("Load() Remote = %q, want %q", cfg.Remote, tt.wantRemote)
			}
			if len(cfg.Watching) != tt.wantWatched {
				t.Errorf("Load() Watching count = %d, want %d", len(cfg.Watching), tt.wantWatched)
			}
		})
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "save full config",
			config: Config{
				Git:    GitModeRemove,
				Remote: "git@github.com:user/repo.git",
				Watching: []Watched{
					{Path: ".config/nvim", Enabled: true},
					{Path: ".bashrc", Git: GitModeDisable, Enabled: true},
				},
				Daemon: DaemonConfig{
					CopyInterval: "1h",
					PushInterval: "24h",
				},
			},
			wantErr: false,
		},
		{
			name: "save minimal config",
			config: Config{
				Git: GitModeDisable,
			},
			wantErr: false,
		},
		{
			name: "save config with vault path",
			config: Config{
				Git:       GitModeDisable,
				VaultPath: "/custom/vault",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "config-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			configPath := filepath.Join(tmpDir, "subdir", "config.yml")

			err = tt.config.Save(configPath)
			if tt.wantErr {
				if err == nil {
					t.Error("Save() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Save() unexpected error: %v", err)
			}

			// Verify file was created
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				t.Error("Save() did not create file")
			}

			// Load and verify content
			loaded, err := Load(configPath)
			if err != nil {
				t.Fatalf("failed to load saved config: %v", err)
			}

			if loaded.Git != tt.config.Git {
				t.Errorf("Save/Load roundtrip: Git = %q, want %q", loaded.Git, tt.config.Git)
			}
			if loaded.Remote != tt.config.Remote {
				t.Errorf("Save/Load roundtrip: Remote = %q, want %q", loaded.Remote, tt.config.Remote)
			}
			if len(loaded.Watching) != len(tt.config.Watching) {
				t.Errorf("Save/Load roundtrip: Watching count = %d, want %d", len(loaded.Watching), len(tt.config.Watching))
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid disable mode",
			config:  Config{Git: GitModeDisable},
			wantErr: false,
		},
		{
			name:    "valid remove mode",
			config:  Config{Git: GitModeRemove},
			wantErr: false,
		},
		{
			name:    "invalid git mode",
			config:  Config{Git: "invalid"},
			wantErr: true,
		},
		{
			name:    "empty git mode (invalid)",
			config:  Config{Git: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEffectiveGitMode(t *testing.T) {
	tests := []struct {
		name       string
		watched    Watched
		globalMode GitMode
		want       GitMode
	}{
		{
			name:       "use watched mode when set",
			watched:    Watched{Path: ".config", Git: GitModeRemove},
			globalMode: GitModeDisable,
			want:       GitModeRemove,
		},
		{
			name:       "fallback to global when watched empty",
			watched:    Watched{Path: ".config", Git: ""},
			globalMode: GitModeRemove,
			want:       GitModeRemove,
		},
		{
			name:       "watched disable overrides global remove",
			watched:    Watched{Path: ".config", Git: GitModeDisable},
			globalMode: GitModeRemove,
			want:       GitModeDisable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.watched.EffectiveGitMode(tt.globalMode)
			if got != tt.want {
				t.Errorf("EffectiveGitMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVaultDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name      string
		vaultPath string
		want      string
	}{
		{
			name:      "default vault path",
			vaultPath: "",
			want:      filepath.Join(home, ".snapfig", "vault"),
		},
		{
			name:      "custom absolute path",
			vaultPath: "/custom/vault/path",
			want:      "/custom/vault/path",
		},
		{
			name:      "tilde expansion",
			vaultPath: "~/my-vault",
			want:      filepath.Join(home, "my-vault"),
		},
		{
			name:      "tilde with subdir",
			vaultPath: "~/.dotfiles/vault",
			want:      filepath.Join(home, ".dotfiles/vault"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{VaultPath: tt.vaultPath}
			got, err := cfg.VaultDir()
			if err != nil {
				t.Fatalf("VaultDir() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("VaultDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSnapfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name      string
		vaultPath string
		want      string
	}{
		{
			name:      "default snapfig dir",
			vaultPath: "",
			want:      filepath.Join(home, ".snapfig"),
		},
		{
			name:      "custom vault path",
			vaultPath: "/custom/path/vault",
			want:      "/custom/path",
		},
		{
			name:      "tilde vault path",
			vaultPath: "~/backups/vault",
			want:      filepath.Join(home, "backups"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{VaultPath: tt.vaultPath}
			got, err := cfg.SnapfigDir()
			if err != nil {
				t.Fatalf("SnapfigDir() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SnapfigDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDaemonConfig(t *testing.T) {
	content := `git: disable
daemon:
  copy_interval: 30m
  push_interval: 6h
  pull_interval: 12h
  auto_restore: true
watching: []
`
	tmpFile, err := os.CreateTemp("", "config-daemon-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Daemon.CopyInterval != "30m" {
		t.Errorf("Daemon.CopyInterval = %q, want %q", cfg.Daemon.CopyInterval, "30m")
	}
	if cfg.Daemon.PushInterval != "6h" {
		t.Errorf("Daemon.PushInterval = %q, want %q", cfg.Daemon.PushInterval, "6h")
	}
	if cfg.Daemon.PullInterval != "12h" {
		t.Errorf("Daemon.PullInterval = %q, want %q", cfg.Daemon.PullInterval, "12h")
	}
	if !cfg.Daemon.AutoRestore {
		t.Error("Daemon.AutoRestore = false, want true")
	}
}
