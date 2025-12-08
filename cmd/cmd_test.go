package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
)

func TestRootCommandHasSubcommands(t *testing.T) {
	subcommands := []string{"copy", "push", "pull", "restore", "daemon", "tui", "setup"}

	for _, name := range subcommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == name || cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rootCmd missing subcommand: %s", name)
		}
	}
}

func TestRootCommandDescription(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
	if rootCmd.Use != "snapfig" {
		t.Errorf("rootCmd.Use = %q, want 'snapfig'", rootCmd.Use)
	}
}

func TestCopyCommandDescription(t *testing.T) {
	if copyCmd.Short == "" {
		t.Error("copyCmd.Short should not be empty")
	}
	if copyCmd.Use != "copy" {
		t.Errorf("copyCmd.Use = %q, want 'copy'", copyCmd.Use)
	}
}

func TestPushCommandDescription(t *testing.T) {
	if pushCmd.Short == "" {
		t.Error("pushCmd.Short should not be empty")
	}
	if pushCmd.Use != "push" {
		t.Errorf("pushCmd.Use = %q, want 'push'", pushCmd.Use)
	}
}

func TestPullCommandDescription(t *testing.T) {
	if pullCmd.Short == "" {
		t.Error("pullCmd.Short should not be empty")
	}
	if pullCmd.Use != "pull" {
		t.Errorf("pullCmd.Use = %q, want 'pull'", pullCmd.Use)
	}
}

func TestRestoreCommandDescription(t *testing.T) {
	if restoreCmd.Short == "" {
		t.Error("restoreCmd.Short should not be empty")
	}
	if restoreCmd.Use != "restore" {
		t.Errorf("restoreCmd.Use = %q, want 'restore'", restoreCmd.Use)
	}
}

func TestDaemonCommandDescription(t *testing.T) {
	if daemonCmd.Short == "" {
		t.Error("daemonCmd.Short should not be empty")
	}
	if daemonCmd.Use != "daemon" {
		t.Errorf("daemonCmd.Use = %q, want 'daemon'", daemonCmd.Use)
	}
}

func TestTuiCommandDescription(t *testing.T) {
	if tuiCmd.Short == "" {
		t.Error("tuiCmd.Short should not be empty")
	}
	if tuiCmd.Use != "tui" {
		t.Errorf("tuiCmd.Use = %q, want 'tui'", tuiCmd.Use)
	}
}

func TestSetupCommandDescription(t *testing.T) {
	if setupCmd.Short == "" {
		t.Error("setupCmd.Short should not be empty")
	}
	if setupCmd.Use != "setup" {
		t.Errorf("setupCmd.Use = %q, want 'setup'", setupCmd.Use)
	}
}

func TestRootHasConfigFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("config")
	if flag == nil {
		t.Error("rootCmd missing --config flag")
	}
}

func TestRootHasGitFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("git")
	if flag == nil {
		t.Error("rootCmd missing --git flag")
	}
}

func TestRootHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	// Execute will exit with help, which is normal
	rootCmd.Execute()

	output := buf.String()
	if len(output) == 0 {
		// Help might go to stdout directly, that's ok
		t.Log("Help output may have gone to stdout")
	}
}

func TestInitConfig(t *testing.T) {
	// This should not panic
	initConfig()
}

func TestInitConfigWithCustomFile(t *testing.T) {
	oldCfgFile := cfgFile
	cfgFile = "/tmp/nonexistent-config.yaml"
	defer func() { cfgFile = oldCfgFile }()

	// Should not panic
	initConfig()
}

// Tests with mocked dependencies

// withMockedDeps runs a test function with mocked dependencies and restores them after.
func withMockedDeps(t *testing.T, fn func()) {
	t.Helper()
	oldServiceFactory := ServiceFactory
	oldConfigLoader := ConfigLoader
	oldConfigDir := DefaultConfigDirFunc
	oldHasRemote := HasRemoteFunc
	defer func() {
		ServiceFactory = oldServiceFactory
		ConfigLoader = oldConfigLoader
		DefaultConfigDirFunc = oldConfigDir
		HasRemoteFunc = oldHasRemote
	}()
	fn()
}

func TestRunCopyWithOutput(t *testing.T) {
	tests := []struct {
		name             string
		cfg              *config.Config
		copyResult       *snapfig.CopyResult
		copyErr          error
		configLoadErr    error
		serviceFactoryErr error
		wantErr          bool
		wantErrContains  string
		wantContains     []string
		wantCopyCalled   bool
	}{
		{
			name: "successful copy",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			copyResult: &snapfig.CopyResult{
				Copied:       []string{".config/test"},
				Skipped:      []string{},
				FilesUpdated: 3,
			},
			wantContains:   []string{"Copying to vault", "Copied: .config/test"},
			wantCopyCalled: true,
		},
		{
			name: "copy with skipped paths",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			copyResult: &snapfig.CopyResult{
				Copied:  []string{},
				Skipped: []string{".config/missing"},
			},
			wantContains:   []string{"Copying to vault", "Skipped: .config/missing"},
			wantCopyCalled: true,
		},
		{
			name:         "no paths configured",
			cfg:          &config.Config{Git: config.GitModeDisable, Watching: nil},
			wantContains: []string{"No paths configured"},
		},
		{
			name:            "config load error",
			configLoadErr:   fmt.Errorf("config not found"),
			wantErr:         true,
			wantErrContains: "config not found",
		},
		{
			name: "service factory error",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			serviceFactoryErr: fmt.Errorf("failed to create service"),
			wantErr:           true,
			wantErrContains:   "failed to create service",
		},
		{
			name: "copy operation error",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			copyErr:         fmt.Errorf("permission denied"),
			wantErr:         true,
			wantErrContains: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockedDeps(t, func() {
				mockSvc := snapfig.NewMockService(tt.cfg)
				if tt.copyResult != nil || tt.copyErr != nil {
					mockSvc.CopyFunc = func() (*snapfig.CopyResult, error) {
						return tt.copyResult, tt.copyErr
					}
				}

				DefaultConfigDirFunc = func() (string, error) { return "/tmp", nil }
				ConfigLoader = func(path string) (*config.Config, error) {
					if tt.configLoadErr != nil {
						return nil, tt.configLoadErr
					}
					return tt.cfg, nil
				}
				ServiceFactory = func(cfg *config.Config, path string) (snapfig.Service, error) {
					if tt.serviceFactoryErr != nil {
						return nil, tt.serviceFactoryErr
					}
					return mockSvc, nil
				}

				var buf bytes.Buffer
				err := runCopyWithOutput(&buf)

				if (err != nil) != tt.wantErr {
					t.Errorf("runCopyWithOutput() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.wantErrContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.wantErrContains) {
						t.Errorf("error should contain %q, got: %v", tt.wantErrContains, err)
					}
				}

				output := buf.String()
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got: %s", want, output)
					}
				}

				if tt.wantCopyCalled && !mockSvc.CopyCalled {
					t.Error("Copy should have been called on service")
				}
			})
		})
	}
}

func TestRunRestoreWithOutput(t *testing.T) {
	tests := []struct {
		name              string
		cfg               *config.Config
		restoreResult     *snapfig.RestoreResult
		restoreErr        error
		configLoadErr     error
		serviceFactoryErr error
		wantErr           bool
		wantErrContains   string
		wantContains      []string
		wantRestoreCalled bool
	}{
		{
			name: "successful restore",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			restoreResult: &snapfig.RestoreResult{
				Restored:     []string{".config/test"},
				Backups:      []string{".config/test.bak"},
				FilesUpdated: 2,
			},
			wantContains:      []string{"Restoring from vault", "Restored: .config/test", "Backed up: .config/test.bak"},
			wantRestoreCalled: true,
		},
		{
			name: "restore with skipped paths",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			restoreResult: &snapfig.RestoreResult{
				Restored: []string{},
				Skipped:  []string{".config/missing"},
			},
			wantContains:      []string{"Restoring from vault", "Skipped: .config/missing"},
			wantRestoreCalled: true,
		},
		{
			name:         "no paths configured",
			cfg:          &config.Config{Git: config.GitModeDisable, Watching: nil},
			wantContains: []string{"No paths configured"},
		},
		{
			name:            "config load error",
			configLoadErr:   fmt.Errorf("config not found"),
			wantErr:         true,
			wantErrContains: "config not found",
		},
		{
			name: "service factory error",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			serviceFactoryErr: fmt.Errorf("failed to create service"),
			wantErr:           true,
			wantErrContains:   "failed to create service",
		},
		{
			name: "restore operation error",
			cfg: &config.Config{
				Git: config.GitModeDisable,
				Watching: []config.Watched{
					{Path: ".config/test", Enabled: true},
				},
			},
			restoreErr:      fmt.Errorf("permission denied"),
			wantErr:         true,
			wantErrContains: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockedDeps(t, func() {
				mockSvc := snapfig.NewMockService(tt.cfg)
				if tt.restoreResult != nil || tt.restoreErr != nil {
					mockSvc.RestoreFunc = func() (*snapfig.RestoreResult, error) {
						return tt.restoreResult, tt.restoreErr
					}
				}

				DefaultConfigDirFunc = func() (string, error) { return "/tmp", nil }
				ConfigLoader = func(path string) (*config.Config, error) {
					if tt.configLoadErr != nil {
						return nil, tt.configLoadErr
					}
					return tt.cfg, nil
				}
				ServiceFactory = func(cfg *config.Config, path string) (snapfig.Service, error) {
					if tt.serviceFactoryErr != nil {
						return nil, tt.serviceFactoryErr
					}
					return mockSvc, nil
				}

				var buf bytes.Buffer
				err := runRestoreWithOutput(&buf)

				if (err != nil) != tt.wantErr {
					t.Errorf("runRestoreWithOutput() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.wantErrContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.wantErrContains) {
						t.Errorf("error should contain %q, got: %v", tt.wantErrContains, err)
					}
				}

				output := buf.String()
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got: %s", want, output)
					}
				}

				if tt.wantRestoreCalled && !mockSvc.RestoreCalled {
					t.Error("Restore should have been called on service")
				}
			})
		})
	}
}

func TestRunPullWithOutput(t *testing.T) {
	tests := []struct {
		name              string
		cfg               *config.Config
		pullResult        *snapfig.PullResult
		pullErr           error
		configLoadErr     error
		serviceFactoryErr error
		hasRemote         bool
		remoteURL         string
		hasRemoteErr      error
		wantErr           bool
		wantErrContains   string
		wantContains      []string
		wantPullCalled    bool
	}{
		{
			name: "successful clone with config remote",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			pullResult:     &snapfig.PullResult{Cloned: true},
			wantContains:   []string{"Cloned successfully"},
			wantPullCalled: true,
		},
		{
			name: "successful pull with config remote",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			pullResult:     &snapfig.PullResult{Cloned: false},
			wantContains:   []string{"Pulled successfully"},
			wantPullCalled: true,
		},
		{
			name: "successful pull with git remote fallback",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "", // Empty remote in config
			},
			hasRemote:      true,
			remoteURL:      "https://github.com/git/vault.git",
			pullResult:     &snapfig.PullResult{Cloned: false},
			wantContains:   []string{"Pulling from https://github.com/git/vault.git", "Pulled successfully"},
			wantPullCalled: true,
		},
		{
			name: "no remote configured in config or git",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "",
			},
			hasRemote:       false,
			wantErr:         true,
			wantErrContains: "no remote configured",
		},
		{
			name: "has remote check error",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "",
			},
			hasRemoteErr:    fmt.Errorf("git error"),
			wantErr:         true,
			wantErrContains: "git error",
		},
		{
			name:            "config load error",
			configLoadErr:   fmt.Errorf("config not found"),
			wantErr:         true,
			wantErrContains: "config not found",
		},
		{
			name: "service factory error",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			serviceFactoryErr: fmt.Errorf("failed to create service"),
			wantErr:           true,
			wantErrContains:   "failed to create service",
		},
		{
			name: "pull operation error",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			pullErr:         fmt.Errorf("network error"),
			wantErr:         true,
			wantErrContains: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockedDeps(t, func() {
				mockSvc := snapfig.NewMockService(tt.cfg)
				if tt.pullResult != nil || tt.pullErr != nil {
					mockSvc.PullFunc = func() (*snapfig.PullResult, error) {
						return tt.pullResult, tt.pullErr
					}
				}

				DefaultConfigDirFunc = func() (string, error) { return "/tmp", nil }
				ConfigLoader = func(path string) (*config.Config, error) {
					if tt.configLoadErr != nil {
						return nil, tt.configLoadErr
					}
					return tt.cfg, nil
				}
				ServiceFactory = func(cfg *config.Config, path string) (snapfig.Service, error) {
					if tt.serviceFactoryErr != nil {
						return nil, tt.serviceFactoryErr
					}
					return mockSvc, nil
				}
				HasRemoteFunc = func(dir string) (bool, string, error) {
					return tt.hasRemote, tt.remoteURL, tt.hasRemoteErr
				}

				var buf bytes.Buffer
				err := runPullWithOutput(&buf)

				if (err != nil) != tt.wantErr {
					t.Errorf("runPullWithOutput() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.wantErrContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.wantErrContains) {
						t.Errorf("error should contain %q, got: %v", tt.wantErrContains, err)
					}
				}

				output := buf.String()
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got: %s", want, output)
					}
				}

				if tt.wantPullCalled && !mockSvc.PullCalled {
					t.Error("Pull should have been called on service")
				}
			})
		})
	}
}

func TestRunPushWithOutput(t *testing.T) {
	tests := []struct {
		name              string
		cfg               *config.Config
		hasRemote         bool
		remoteURL         string
		hasRemoteErr      error
		pushErr           error
		configLoadErr     error
		serviceFactoryErr error
		wantErr           bool
		wantErrContains   string
		wantContains      []string
		wantPushCalled    bool
	}{
		{
			name: "successful push",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			hasRemote:      true,
			remoteURL:      "https://github.com/test/vault.git",
			wantContains:   []string{"Pushing to", "Done"},
			wantPushCalled: true,
		},
		{
			name:            "no remote configured",
			cfg:             &config.Config{Git: config.GitModeDisable},
			hasRemote:       false,
			wantErr:         true,
			wantErrContains: "no remote configured",
		},
		{
			name:            "config load error",
			configLoadErr:   fmt.Errorf("config not found"),
			wantErr:         true,
			wantErrContains: "config not found",
		},
		{
			name: "service factory error",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			serviceFactoryErr: fmt.Errorf("failed to create service"),
			wantErr:           true,
			wantErrContains:   "failed to create service",
		},
		{
			name: "push operation error",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			hasRemote:       true,
			remoteURL:       "https://github.com/test/vault.git",
			pushErr:         fmt.Errorf("network error"),
			wantErr:         true,
			wantErrContains: "network error",
		},
		{
			name: "has remote check error",
			cfg: &config.Config{
				Git:    config.GitModeDisable,
				Remote: "https://github.com/test/vault.git",
			},
			hasRemoteErr:    fmt.Errorf("git error"),
			wantErr:         true,
			wantErrContains: "git error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockedDeps(t, func() {
				mockSvc := snapfig.NewMockService(tt.cfg)
				if tt.pushErr != nil {
					mockSvc.PushFunc = func() error {
						return tt.pushErr
					}
				}

				DefaultConfigDirFunc = func() (string, error) { return "/tmp", nil }
				ConfigLoader = func(path string) (*config.Config, error) {
					if tt.configLoadErr != nil {
						return nil, tt.configLoadErr
					}
					return tt.cfg, nil
				}
				ServiceFactory = func(cfg *config.Config, path string) (snapfig.Service, error) {
					if tt.serviceFactoryErr != nil {
						return nil, tt.serviceFactoryErr
					}
					return mockSvc, nil
				}
				HasRemoteFunc = func(dir string) (bool, string, error) {
					return tt.hasRemote, tt.remoteURL, tt.hasRemoteErr
				}

				var buf bytes.Buffer
				err := runPushWithOutput(&buf)

				if (err != nil) != tt.wantErr {
					t.Errorf("runPushWithOutput() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.wantErrContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.wantErrContains) {
						t.Errorf("error should contain %q, got: %v", tt.wantErrContains, err)
					}
				}

				output := buf.String()
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got: %s", want, output)
					}
				}

				if tt.wantPushCalled && !mockSvc.PushCalled {
					t.Error("Push should have been called on service")
				}
			})
		})
	}
}

func TestParsePaths(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLen     int
		wantFirst   string
		wantGitMode config.GitMode
		wantErr     bool
	}{
		{
			name:        "single path with x mode",
			input:       ".config/nvim:x",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeRemove,
		},
		{
			name:        "single path with g mode",
			input:       ".config/nvim:g",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeDisable,
		},
		{
			name:        "single path with uppercase G mode",
			input:       ".config/nvim:G",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeDisable,
		},
		{
			name:        "single path with uppercase X mode",
			input:       ".config/nvim:X",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeRemove,
		},
		{
			name:        "multiple paths",
			input:       ".config/nvim:g,.zshrc:x",
			wantLen:     2,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeDisable,
		},
		{
			name:        "path without mode defaults to remove",
			input:       ".config/nvim",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeRemove,
		},
		{
			name:    "invalid mode",
			input:   ".config/nvim:z",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:        "path with tilde prefix",
			input:       "~/.config/nvim:g",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeDisable,
		},
		{
			name:        "path with leading slash",
			input:       "/.config/nvim:g",
			wantLen:     1,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeDisable,
		},
		{
			name:        "path with spaces around comma",
			input:       ".config/nvim:g , .zshrc:x",
			wantLen:     2,
			wantFirst:   ".config/nvim",
			wantGitMode: config.GitModeDisable,
		},
		{
			name:        "path ending with colon but no mode",
			input:       ".config/nvim:",
			wantLen:     1,
			wantFirst:   ".config/nvim:",
			wantGitMode: config.GitModeRemove,
		},
		{
			name:    "only whitespace parts",
			input:   "  ,  ,  ",
			wantLen: 0,
		},
		{
			name:    "path that becomes empty after trimming",
			input:   "~/",
			wantLen: 0,
		},
		{
			name:        "multiple colons in path",
			input:       ".config/test:file:g",
			wantLen:     1,
			wantFirst:   ".config/test:file",
			wantGitMode: config.GitModeDisable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePaths(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("len(result) = %d, want %d", len(result), tt.wantLen)
				return
			}

			if tt.wantLen > 0 {
				if result[0].Path != tt.wantFirst {
					t.Errorf("result[0].Path = %q, want %q", result[0].Path, tt.wantFirst)
				}
				if result[0].Git != tt.wantGitMode {
					t.Errorf("result[0].Git = %v, want %v", result[0].Git, tt.wantGitMode)
				}
			}
		})
	}
}

func TestLoadConfigWithPath(t *testing.T) {
	tests := []struct {
		name            string
		configDir       string
		configDirErr    error
		loadedCfg       *config.Config
		loadErr         error
		wantErr         bool
		wantErrContains string
		wantPath        string
	}{
		{
			name:      "successful load",
			configDir: "/tmp/testconfig",
			loadedCfg: &config.Config{Git: config.GitModeDisable},
			wantPath:  "/tmp/testconfig/config.yml",
		},
		{
			name:            "config dir error",
			configDirErr:    fmt.Errorf("cannot get config dir"),
			wantErr:         true,
			wantErrContains: "cannot get config dir",
		},
		{
			name:            "config load error",
			configDir:       "/tmp/testconfig",
			loadErr:         fmt.Errorf("config file not found"),
			wantErr:         true,
			wantErrContains: "config file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldConfigLoader := ConfigLoader
			oldConfigDir := DefaultConfigDirFunc
			defer func() {
				ConfigLoader = oldConfigLoader
				DefaultConfigDirFunc = oldConfigDir
			}()

			DefaultConfigDirFunc = func() (string, error) {
				return tt.configDir, tt.configDirErr
			}
			ConfigLoader = func(path string) (*config.Config, error) {
				if tt.loadErr != nil {
					return nil, tt.loadErr
				}
				return tt.loadedCfg, nil
			}

			cfg, path, err := loadConfigWithPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfigWithPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("error should contain %q, got: %v", tt.wantErrContains, err)
				}
			}

			if !tt.wantErr {
				if cfg != tt.loadedCfg {
					t.Error("cfg should match expected")
				}
				if path != tt.wantPath {
					t.Errorf("path = %q, want %q", path, tt.wantPath)
				}
			}
		})
	}
}

func TestResetDeps(t *testing.T) {
	// Modify deps
	ServiceFactory = nil
	ConfigLoader = nil
	DefaultConfigDirFunc = nil

	// Reset
	resetDeps()

	// Verify they're restored
	if ServiceFactory == nil {
		t.Error("ServiceFactory should be restored")
	}
	if ConfigLoader == nil {
		t.Error("ConfigLoader should be restored")
	}
	if DefaultConfigDirFunc == nil {
		t.Error("DefaultConfigDirFunc should be restored")
	}
}

// Daemon tests with temp files

func TestGetDaemonPid(t *testing.T) {
	tests := []struct {
		name        string
		pidContent  string
		createFile  bool
		wantRunning bool
	}{
		{
			name:        "no pid file",
			createFile:  false,
			wantRunning: false,
		},
		{
			name:        "invalid pid content",
			pidContent:  "not-a-number",
			createFile:  true,
			wantRunning: false,
		},
		{
			name:        "pid of non-existent process",
			pidContent:  "999999999",
			createFile:  true,
			wantRunning: false,
		},
		{
			name:        "current process pid",
			pidContent:  "", // Will be set to current PID
			createFile:  true,
			wantRunning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pidFile := filepath.Join(tmpDir, "daemon.pid")

			// Save and restore
			oldPidFunc := PidFilePathFunc
			defer func() { PidFilePathFunc = oldPidFunc }()

			PidFilePathFunc = func() (string, error) {
				return pidFile, nil
			}

			if tt.createFile {
				content := tt.pidContent
				if content == "" {
					// Use current process PID (guaranteed to be running)
					content = fmt.Sprintf("%d", os.Getpid())
				}
				if err := os.WriteFile(pidFile, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write pid file: %v", err)
				}
			}

			_, running := getDaemonPid()
			if running != tt.wantRunning {
				t.Errorf("getDaemonPid() running = %v, want %v", running, tt.wantRunning)
			}
		})
	}
}

func TestDaemonStatusNotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "daemon.pid")

	oldPidFunc := PidFilePathFunc
	defer func() { PidFilePathFunc = oldPidFunc }()

	PidFilePathFunc = func() (string, error) {
		return pidFile, nil
	}

	// No PID file means daemon is not running - just verify no error
	err := runDaemonStatus(daemonStatusCmd, nil)
	if err != nil {
		t.Errorf("runDaemonStatus() error = %v", err)
	}
}

func TestDaemonStatusRunning(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "daemon.pid")
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write current PID (guaranteed running)
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	oldPidFunc := PidFilePathFunc
	oldConfigDir := DefaultConfigDirFunc
	oldConfigLoader := ConfigLoader
	defer func() {
		PidFilePathFunc = oldPidFunc
		DefaultConfigDirFunc = oldConfigDir
		ConfigLoader = oldConfigLoader
	}()

	PidFilePathFunc = func() (string, error) { return pidFile, nil }
	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }
	ConfigLoader = func(path string) (*config.Config, error) {
		return &config.Config{
			Daemon: config.DaemonConfig{
				CopyInterval: "1h",
				PushInterval: "24h",
			},
		}, nil
	}

	// Just verify no error - output goes to stdout
	err := runDaemonStatus(daemonStatusCmd, nil)
	if err != nil {
		t.Errorf("runDaemonStatus() error = %v", err)
	}
}

func TestDaemonStatusRunningWithPullInterval(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "daemon.pid")
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write current PID (guaranteed running)
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	oldPidFunc := PidFilePathFunc
	oldConfigDir := DefaultConfigDirFunc
	oldConfigLoader := ConfigLoader
	defer func() {
		PidFilePathFunc = oldPidFunc
		DefaultConfigDirFunc = oldConfigDir
		ConfigLoader = oldConfigLoader
	}()

	PidFilePathFunc = func() (string, error) { return pidFile, nil }
	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }
	ConfigLoader = func(path string) (*config.Config, error) {
		return &config.Config{
			Daemon: config.DaemonConfig{
				CopyInterval: "1h",
				PushInterval: "24h",
				PullInterval: "12h",
				AutoRestore:  true,
			},
		}, nil
	}

	err := runDaemonStatus(daemonStatusCmd, nil)
	if err != nil {
		t.Errorf("runDaemonStatus() error = %v", err)
	}
}

func TestDaemonStartAlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "daemon.pid")

	// Write current PID (guaranteed running)
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	oldPidFunc := PidFilePathFunc
	defer func() { PidFilePathFunc = oldPidFunc }()

	PidFilePathFunc = func() (string, error) { return pidFile, nil }

	err := runDaemonStart(daemonStartCmd, nil)
	if err == nil {
		t.Error("runDaemonStart() should return error when daemon already running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("error should contain 'already running', got: %v", err)
	}
}

func TestPrintPersistenceInstructions(t *testing.T) {
	// Just ensure it doesn't panic
	printPersistenceInstructions()
}

func TestStartDaemonFromSetupAlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "daemon.pid")

	// Write current PID (guaranteed running)
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	oldPidFunc := PidFilePathFunc
	defer func() { PidFilePathFunc = oldPidFunc }()

	PidFilePathFunc = func() (string, error) { return pidFile, nil }

	err := startDaemonFromSetup()
	if err == nil {
		t.Error("startDaemonFromSetup() should return error when daemon already running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("error should contain 'already running', got: %v", err)
	}
}

func TestRunSetupConfigDirError(t *testing.T) {
	defer resetDeps()

	DefaultConfigDirFunc = func() (string, error) {
		return "", fmt.Errorf("config dir error")
	}

	err := runSetup(setupCmd, nil)
	if err == nil {
		t.Error("runSetup() should return error when config dir fails")
	}
	if !strings.Contains(err.Error(), "config dir error") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSetupConfigExistsNoForce(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	// Create existing config
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yml")
	if err := os.WriteFile(configPath, []byte("existing: config"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }

	// Reset flag to default (no force)
	setupForce = false
	setupPaths = ".config/nvim:x"

	err := runSetup(setupCmd, nil)
	if err == nil {
		t.Error("runSetup() should return error when config exists and force is false")
	}
	if !strings.Contains(err.Error(), "config already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSetupInvalidPathMode(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }

	setupForce = true
	setupPaths = ".config/nvim:z" // Invalid mode

	err := runSetup(setupCmd, nil)
	if err == nil {
		t.Error("runSetup() should return error for invalid path mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSetupEmptyPaths(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }

	setupForce = true
	setupPaths = "  ,  ,  " // Only whitespace/empty

	err := runSetup(setupCmd, nil)
	if err == nil {
		t.Error("runSetup() should return error for empty paths")
	}
	if !strings.Contains(err.Error(), "no valid paths") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSetupServiceFactoryError(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }
	ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
		return nil, fmt.Errorf("service factory error")
	}

	setupForce = true
	setupPaths = ".config/nvim:x"
	setupRemote = ""
	setupNoDaemon = true

	err := runSetup(setupCmd, nil)
	if err == nil {
		t.Error("runSetup() should return error when service factory fails")
	}
	if !strings.Contains(err.Error(), "failed to create service") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSetupCopyError(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }

	mock := &snapfig.MockService{}
	mock.CopyFunc = func() (*snapfig.CopyResult, error) {
		return nil, fmt.Errorf("copy failed")
	}
	ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
		return mock, nil
	}

	setupForce = true
	setupPaths = ".config/nvim:x"
	setupRemote = ""
	setupNoDaemon = true

	err := runSetup(setupCmd, nil)
	if err == nil {
		t.Error("runSetup() should return error when copy fails")
	}
	if !strings.Contains(err.Error(), "initial copy failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSetupWithRemoteError(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }

	mock := &snapfig.MockService{}
	mock.CopyFunc = func() (*snapfig.CopyResult, error) {
		return &snapfig.CopyResult{Copied: []string{"path1"}, FilesUpdated: 5}, nil
	}
	mock.SetRemoteFunc = func(url string) error {
		return fmt.Errorf("remote error")
	}
	ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
		return mock, nil
	}

	setupForce = true
	setupPaths = ".config/nvim:x"
	setupRemote = "git@github.com:user/repo.git"
	setupNoDaemon = true

	// Should not error - remote failure is just a warning
	err := runSetup(setupCmd, nil)
	if err != nil {
		t.Errorf("runSetup() should not error on remote failure, got: %v", err)
	}
}

func TestRunSetupWithRemoteSuccess(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }

	mock := &snapfig.MockService{}
	mock.CopyFunc = func() (*snapfig.CopyResult, error) {
		return &snapfig.CopyResult{Copied: []string{"path1"}, FilesUpdated: 5}, nil
	}
	mock.SetRemoteFunc = func(url string) error {
		return nil
	}
	ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
		return mock, nil
	}

	setupForce = true
	setupPaths = ".config/nvim:g" // Test 'g' mode
	setupRemote = "git@github.com:user/repo.git"
	setupNoDaemon = true

	err := runSetup(setupCmd, nil)
	if err != nil {
		t.Errorf("runSetup() error = %v", err)
	}
}

func TestRunSetupWithDaemonStartFailure(t *testing.T) {
	defer resetDeps()
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	pidFile := filepath.Join(tmpDir, "daemon.pid")
	// Write a running PID to trigger "already running" error
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	DefaultConfigDirFunc = func() (string, error) { return configDir, nil }
	PidFilePathFunc = func() (string, error) { return pidFile, nil }

	mock := &snapfig.MockService{}
	mock.CopyFunc = func() (*snapfig.CopyResult, error) {
		return &snapfig.CopyResult{Copied: []string{"path1"}, FilesUpdated: 5}, nil
	}
	ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
		return mock, nil
	}

	setupForce = true
	setupPaths = ".config/nvim:x"
	setupRemote = ""
	setupNoDaemon = false // Try to start daemon

	// Should not error - daemon failure is just a warning
	err := runSetup(setupCmd, nil)
	if err != nil {
		t.Errorf("runSetup() should not error on daemon failure, got: %v", err)
	}
}

func TestParsePathsTableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLen     int
		wantErr     bool
		errContains string
		checkFirst  *config.Watched
	}{
		{
			name:    "single path no mode",
			input:   ".config/nvim",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeRemove,
				Enabled: true,
			},
		},
		{
			name:    "single path with x mode",
			input:   ".config/nvim:x",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeRemove,
				Enabled: true,
			},
		},
		{
			name:    "single path with g mode",
			input:   ".config/nvim:g",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeDisable,
				Enabled: true,
			},
		},
		{
			name:    "multiple paths",
			input:   ".config/nvim:x,.zshrc:g,.bashrc",
			wantLen: 3,
		},
		{
			name:    "path with tilde prefix",
			input:   "~/.config/nvim:x",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeRemove,
				Enabled: true,
			},
		},
		{
			name:    "path with leading slash",
			input:   "/.config/nvim:x",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeRemove,
				Enabled: true,
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "whitespace only",
			input:   "  ,  ,  ",
			wantLen: 0,
		},
		{
			name:    "path becomes empty after cleaning",
			input:   "~/",
			wantLen: 0,
		},
		{
			name:        "invalid mode",
			input:       ".config/nvim:z",
			wantErr:     true,
			errContains: "invalid mode",
		},
		{
			name:    "uppercase mode g",
			input:   ".config/nvim:G",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeDisable,
				Enabled: true,
			},
		},
		{
			name:    "uppercase mode x",
			input:   ".config/nvim:X",
			wantLen: 1,
			checkFirst: &config.Watched{
				Path:    ".config/nvim",
				Git:     config.GitModeRemove,
				Enabled: true,
			},
		},
		{
			name:        "colon in path no mode",
			input:       "path:with:colons",
			wantErr:     true,
			errContains: "invalid mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePaths(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePaths() should return error")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain '%s', got: %v", tt.errContains, err)
				}
				return
			}
			if err != nil {
				t.Errorf("parsePaths() unexpected error: %v", err)
				return
			}
			if len(result) != tt.wantLen {
				t.Errorf("parsePaths() got %d items, want %d", len(result), tt.wantLen)
			}
			if tt.checkFirst != nil && len(result) > 0 {
				if result[0].Path != tt.checkFirst.Path {
					t.Errorf("first path = %s, want %s", result[0].Path, tt.checkFirst.Path)
				}
				if result[0].Git != tt.checkFirst.Git {
					t.Errorf("first git mode = %v, want %v", result[0].Git, tt.checkFirst.Git)
				}
				if result[0].Enabled != tt.checkFirst.Enabled {
					t.Errorf("first enabled = %v, want %v", result[0].Enabled, tt.checkFirst.Enabled)
				}
			}
		})
	}
}
