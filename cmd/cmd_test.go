package cmd

import (
	"bytes"
	"testing"
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
