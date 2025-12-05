package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
)

var (
	setupPaths        string
	setupRemote       string
	setupVaultPath    string
	setupCopyInterval string
	setupPushInterval string
	setupPullInterval string
	setupAutoRestore  bool
	setupNoDaemon     bool
	setupForce        bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-shot setup: configure paths, create config, and start daemon",
	Long: `Fire-and-forget setup for automation scripts.

Examples:
  snapfig setup --paths=".config/nvim:g,.zshrc:x" --remote="git@github.com:user/vault.git"
  snapfig setup --paths=".config/nvim:g" --copy-interval="30m" --push-interval="12h"

Path format: path:mode where mode is:
  x = remove .git directories (default)
  g = preserve .git as .git_disabled`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().StringVar(&setupPaths, "paths", "", "Paths to watch (format: path1:mode,path2:mode)")
	setupCmd.Flags().StringVar(&setupRemote, "remote", "", "Git remote URL")
	setupCmd.Flags().StringVar(&setupVaultPath, "vault-path", "", "Custom vault location (default: ~/.snapfig/vault)")
	setupCmd.Flags().StringVar(&setupCopyInterval, "copy-interval", "1h", "Copy interval (e.g. 30m, 1h)")
	setupCmd.Flags().StringVar(&setupPushInterval, "push-interval", "24h", "Push interval (e.g. 12h, 24h)")
	setupCmd.Flags().StringVar(&setupPullInterval, "pull-interval", "", "Pull interval (empty = disabled)")
	setupCmd.Flags().BoolVar(&setupAutoRestore, "auto-restore", false, "Auto restore after pull")
	setupCmd.Flags().BoolVar(&setupNoDaemon, "no-daemon", false, "Don't start daemon after setup")
	setupCmd.Flags().BoolVar(&setupForce, "force", false, "Overwrite existing config")

	setupCmd.MarkFlagRequired("paths")

	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	configDir, err := config.DefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.yml")

	if _, err := os.Stat(configPath); err == nil && !setupForce {
		return fmt.Errorf("config already exists at %s\nUse --force to overwrite, or use the TUI (snapfig) to edit", configPath)
	}

	// Parse paths
	watching, err := parsePaths(setupPaths)
	if err != nil {
		return err
	}

	if len(watching) == 0 {
		return fmt.Errorf("no valid paths provided")
	}

	// Build config
	cfg := &config.Config{
		Git:       config.GitModeDisable,
		Remote:    setupRemote,
		VaultPath: setupVaultPath,
		Watching:  watching,
		Daemon: config.DaemonConfig{
			CopyInterval: setupCopyInterval,
			PushInterval: setupPushInterval,
			PullInterval: setupPullInterval,
			AutoRestore:  setupAutoRestore,
		},
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Config saved to %s\n", configPath)

	// Show what we're watching
	fmt.Println("\nWatching:")
	for _, w := range watching {
		mode := "x"
		if w.Git == config.GitModeDisable {
			mode = "g"
		}
		fmt.Printf("  %s [%s]\n", w.Path, mode)
	}

	// Run initial copy
	fmt.Println("\nRunning initial copy...")
	copier, err := snapfig.NewCopier(cfg)
	if err != nil {
		return fmt.Errorf("failed to create copier: %w", err)
	}

	result, err := copier.Copy()
	if err != nil {
		return fmt.Errorf("initial copy failed: %w", err)
	}
	fmt.Printf("Copied: %d paths, %d files updated\n", len(result.Copied), result.FilesUpdated)

	// Configure remote if provided
	if setupRemote != "" {
		vaultDir, err := cfg.VaultDir()
		if err != nil {
			fmt.Printf("Warning: failed to get vault directory: %v\n", err)
		} else if err := snapfig.SetRemote(vaultDir, setupRemote); err != nil {
			fmt.Printf("Warning: failed to set git remote: %v\n", err)
		} else {
			fmt.Printf("Remote configured: %s\n", setupRemote)
		}
	}

	// Start daemon unless --no-daemon
	if !setupNoDaemon {
		fmt.Println("\nStarting daemon...")
		if err := startDaemonFromSetup(); err != nil {
			fmt.Printf("Warning: failed to start daemon: %v\n", err)
			fmt.Println("You can start it manually with: snapfig daemon start")
		} else {
			fmt.Println("Daemon started")
		}
	}

	// Print persistence instructions
	printPersistenceInstructions()

	return nil
}

func parsePaths(pathsStr string) ([]config.Watched, error) {
	var watching []config.Watched

	parts := strings.Split(pathsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse path:mode
		var path string
		var gitMode config.GitMode = config.GitModeRemove // default: x

		if idx := strings.LastIndex(part, ":"); idx != -1 && idx < len(part)-1 {
			path = part[:idx]
			mode := part[idx+1:]
			switch strings.ToLower(mode) {
			case "g":
				gitMode = config.GitModeDisable
			case "x":
				gitMode = config.GitModeRemove
			default:
				return nil, fmt.Errorf("invalid mode '%s' for path '%s' (use 'x' or 'g')", mode, path)
			}
		} else {
			path = part
		}

		// Clean path
		path = strings.TrimSpace(path)
		path = strings.TrimPrefix(path, "~/")
		path = strings.TrimPrefix(path, "/")

		if path == "" {
			continue
		}

		watching = append(watching, config.Watched{
			Path:    path,
			Git:     gitMode,
			Enabled: true,
		})
	}

	return watching, nil
}

func startDaemonFromSetup() error {
	pid, running := getDaemonPid()
	if running {
		return fmt.Errorf("daemon already running (pid %d)", pid)
	}

	// Reuse the daemon start logic
	return daemonStartCmd.RunE(daemonStartCmd, nil)
}

func printPersistenceInstructions() {
	home, _ := os.UserHomeDir()

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("SETUP COMPLETE")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("\nTo start daemon automatically on login:")
	fmt.Println("\nOption 1 - Shell rc (simple):")
	fmt.Println("  echo 'snapfig daemon start 2>/dev/null' >> ~/.bashrc")
	fmt.Println("  # or for zsh:")
	fmt.Println("  echo 'snapfig daemon start 2>/dev/null' >> ~/.zshrc")
	fmt.Println("\nOption 2 - Systemd user service (recommended):")
	fmt.Printf("  mkdir -p %s/.config/systemd/user\n", home)
	fmt.Printf("  cat > %s/.config/systemd/user/snapfig.service << 'EOF'\n", home)
	fmt.Println("[Unit]")
	fmt.Println("Description=Snapfig background runner")
	fmt.Println("")
	fmt.Println("[Service]")
	fmt.Println("ExecStart=/usr/local/bin/snapfig daemon run")
	fmt.Println("Restart=on-failure")
	fmt.Println("")
	fmt.Println("[Install]")
	fmt.Println("WantedBy=default.target")
	fmt.Println("EOF")
	fmt.Println("\n  systemctl --user enable --now snapfig")
	fmt.Println(strings.Repeat("─", 60))
}
