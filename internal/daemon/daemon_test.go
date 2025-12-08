package daemon

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config with intervals",
			cfg: &config.Config{
				VaultPath: tmpDir,
				Daemon: config.DaemonConfig{
					CopyInterval: "1h",
					PushInterval: "2h",
					PullInterval: "30m",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with copy only",
			cfg: &config.Config{
				VaultPath: tmpDir,
				Daemon: config.DaemonConfig{
					CopyInterval: "15m",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid copy interval",
			cfg: &config.Config{
				VaultPath: tmpDir,
				Daemon: config.DaemonConfig{
					CopyInterval: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid push interval",
			cfg: &config.Config{
				VaultPath: tmpDir,
				Daemon: config.DaemonConfig{
					PushInterval: "bad",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid pull interval",
			cfg: &config.Config{
				VaultPath: tmpDir,
				Daemon: config.DaemonConfig{
					PullInterval: "wrong",
				},
			},
			wantErr: true,
		},
		{
			name: "empty intervals is valid",
			cfg: &config.Config{
				VaultPath: tmpDir,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := New(tt.cfg, "/tmp/config.toml")
			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() unexpected error: %v", err)
			}
			if d == nil {
				t.Error("New() returned nil daemon")
			}
		})
	}
}

func TestParseIntervals(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		copyInterval string
		pushInterval string
		pullInterval string
		wantErr      bool
	}{
		{
			name:         "all valid",
			copyInterval: "1h",
			pushInterval: "30m",
			pullInterval: "2h30m",
			wantErr:      false,
		},
		{
			name:         "seconds only",
			copyInterval: "60s",
			wantErr:      false,
		},
		{
			name:         "invalid format",
			copyInterval: "not-a-duration",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				VaultPath: tmpDir,
				Daemon: config.DaemonConfig{
					CopyInterval: tt.copyInterval,
					PushInterval: tt.pushInterval,
					PullInterval: tt.pullInterval,
				},
			}

			_, err := New(cfg, "/tmp/config.toml")
			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error for invalid interval")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() unexpected error: %v", err)
			}
		})
	}
}

func TestRunNoIntervals(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		VaultPath: tmpDir,
	}

	d, err := New(cfg, "/tmp/config.toml")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = d.Run()
	if err == nil {
		t.Error("Run() should error when no intervals configured")
	}
}

func TestWritePidFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		VaultPath: tmpDir,
		Daemon: config.DaemonConfig{
			CopyInterval: "1h",
		},
	}

	d, err := New(cfg, "/tmp/config.toml")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = d.writePidFile()
	if err != nil {
		t.Fatalf("writePidFile() error: %v", err)
	}

	// Verify PID file exists
	pidPath, _ := config.PidFilePath()
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Error("writePidFile() did not create PID file")
	}

	// Clean up
	d.removePidFile()
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("removePidFile() did not remove PID file")
	}
}

func TestReloadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial config file (YAML format)
	configPath := filepath.Join(tmpDir, "config.yaml")
	initialConfig := `git: disable
vault_path: "` + tmpDir + `"
daemon:
  copy_interval: "1h"
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}

	d, err := New(cfg, configPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Reload without changes
	changed := d.reloadConfig()
	if changed {
		t.Error("reloadConfig() should return false when no changes")
	}

	// Update config file
	updatedConfig := `git: disable
vault_path: "` + tmpDir + `"
daemon:
  copy_interval: "2h"
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("failed to write updated config: %v", err)
	}

	// Reload with changes
	changed = d.reloadConfig()
	if !changed {
		t.Error("reloadConfig() should return true when intervals changed")
	}
}

func TestReloadConfigInvalidFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		VaultPath: tmpDir,
		Daemon: config.DaemonConfig{
			CopyInterval: "1h",
		},
	}

	// Use non-existent config path
	d, err := New(cfg, "/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Reload should fail gracefully
	changed := d.reloadConfig()
	if changed {
		t.Error("reloadConfig() should return false when file doesn't exist")
	}
}

func TestReloadConfigInvalidInterval(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial config file (YAML format)
	configPath := filepath.Join(tmpDir, "config.yaml")
	initialConfig := `git: disable
vault_path: "` + tmpDir + `"
daemon:
  copy_interval: "1h"
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}

	d, err := New(cfg, configPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Update config with invalid interval
	invalidConfig := `git: disable
vault_path: "` + tmpDir + `"
daemon:
  copy_interval: "invalid"
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	// Reload should fail gracefully
	changed := d.reloadConfig()
	if changed {
		t.Error("reloadConfig() should return false when interval is invalid")
	}
}

func TestDaemonFields(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		VaultPath: tmpDir,
		Daemon: config.DaemonConfig{
			CopyInterval: "1h",
			PushInterval: "30m",
			PullInterval: "2h",
		},
	}

	d, err := New(cfg, "/tmp/config.toml")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Verify parsed intervals
	if d.copyInterval.String() != "1h0m0s" {
		t.Errorf("copyInterval = %v, want 1h0m0s", d.copyInterval)
	}
	if d.pushInterval.String() != "30m0s" {
		t.Errorf("pushInterval = %v, want 30m0s", d.pushInterval)
	}
	if d.pullInterval.String() != "2h0m0s" {
		t.Errorf("pullInterval = %v, want 2h0m0s", d.pullInterval)
	}
}

func TestDoCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	homeDir := filepath.Join(tmpDir, "home")
	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(vaultDir, 0755)

	// Create a file to copy
	testFile := filepath.Join(homeDir, ".testrc")
	os.WriteFile(testFile, []byte("test content"), 0644)

	cfg := &config.Config{
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".testrc", Enabled: true},
		},
		Daemon: config.DaemonConfig{
			CopyInterval: "1h",
		},
	}

	d := &Daemon{
		cfg:          cfg,
		configPath:   "/tmp/config.yaml",
		vaultDir:     vaultDir,
		copyInterval: 0,
		logger:       log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// Override home directory by modifying copier expectations
	// This test won't work perfectly without mocking, but we can test the flow
	d.doCopy()
	// doCopy should not panic and should handle errors gracefully
}

func TestDoPush(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		VaultPath: tmpDir,
		Daemon: config.DaemonConfig{
			PushInterval: "1h",
		},
	}

	d := &Daemon{
		cfg:          cfg,
		configPath:   "/tmp/config.yaml",
		vaultDir:     tmpDir,
		pushInterval: 0,
		logger:       log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doPush will fail (no git repo) but should handle gracefully
	d.doPush()
}

func TestDoPull(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		VaultPath: tmpDir,
		Remote:    "",
		Daemon: config.DaemonConfig{
			PullInterval: "1h",
		},
	}

	d := &Daemon{
		cfg:          cfg,
		configPath:   "/tmp/config.yaml",
		vaultDir:     tmpDir,
		pullInterval: 0,
		logger:       log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doPull will fail (no remote) but should handle gracefully
	d.doPull()
}

func TestDoRestore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(vaultDir, 0755)

	cfg := &config.Config{
		VaultPath: vaultDir,
		Daemon: config.DaemonConfig{
			AutoRestore: true,
		},
	}

	d := &Daemon{
		cfg:        cfg,
		configPath: "/tmp/config.yaml",
		vaultDir:   vaultDir,
		logger:     log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doRestore should handle case with no watched paths
	d.doRestore()
}

func TestDoPullWithAutoRestore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(vaultDir, 0755)

	// Initialize git repo so pull can work
	cfg := &config.Config{
		VaultPath: vaultDir,
		Remote:    "", // No remote
		Daemon: config.DaemonConfig{
			PullInterval: "1h",
			AutoRestore:  true,
		},
	}

	d := &Daemon{
		cfg:          cfg,
		configPath:   "/tmp/config.yaml",
		vaultDir:     vaultDir,
		pullInterval: 0,
		logger:       log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doPull will fail but AutoRestore code path will not be reached
	d.doPull()
}

func TestDoCopyWithResults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(vaultDir, 0755)

	cfg := &config.Config{
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".nonexistent", Enabled: true},
		},
		Daemon: config.DaemonConfig{
			CopyInterval: "1h",
		},
	}

	d := &Daemon{
		cfg:          cfg,
		configPath:   "/tmp/config.yaml",
		vaultDir:     vaultDir,
		copyInterval: 0,
		logger:       log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// This will trigger the skipped path logging
	d.doCopy()
}

func TestDoPullAutoRestoreEnabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(vaultDir, 0755)

	homeDir := filepath.Join(tmpDir, "home")
	os.MkdirAll(homeDir, 0755)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	cfg := &config.Config{
		VaultPath: vaultDir,
		Remote:    "", // No remote, so pull will fail
		Daemon: config.DaemonConfig{
			AutoRestore: true, // Enable auto restore
		},
	}

	d := &Daemon{
		cfg:          cfg,
		configPath:   filepath.Join(tmpDir, "config.yaml"),
		vaultDir:     vaultDir,
		pullInterval: 0,
		logger:       log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doPull will fail because no remote, but we test the code path exists
	d.doPull()
}

func TestDoRestoreNoWatching(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	os.MkdirAll(vaultDir, 0755)

	homeDir := filepath.Join(tmpDir, "home")
	os.MkdirAll(homeDir, 0755)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	cfg := &config.Config{
		VaultPath: vaultDir,
		Watching:  nil, // No watching paths
	}

	d := &Daemon{
		cfg:        cfg,
		configPath: filepath.Join(tmpDir, "config.yaml"),
		vaultDir:   vaultDir,
		logger:     log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doRestore with no watching paths should log nothing to restore
	d.doRestore()
}

func TestDoRestoreWithResults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultDir := filepath.Join(tmpDir, "vault")
	vaultConfigDir := filepath.Join(vaultDir, ".config", "test")
	os.MkdirAll(vaultConfigDir, 0755)
	os.WriteFile(filepath.Join(vaultConfigDir, "file.txt"), []byte("content"), 0644)

	homeDir := filepath.Join(tmpDir, "home")
	os.MkdirAll(homeDir, 0755)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	cfg := &config.Config{
		VaultPath: vaultDir,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true},
		},
	}

	d := &Daemon{
		cfg:        cfg,
		configPath: filepath.Join(tmpDir, "config.yaml"),
		vaultDir:   vaultDir,
		logger:     log.New(os.Stdout, "[test] ", log.LstdFlags),
	}

	// doRestore should restore files
	d.doRestore()
}
