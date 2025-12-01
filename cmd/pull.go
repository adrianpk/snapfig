package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull vault from remote",
	Long:  "Pulls the vault git repository from the configured remote origin. If vault doesn't exist, clones it.",
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	// Load config to get remote URL
	configDir, err := config.DefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.yml")

	cfg, err := config.Load(configPath)
	if err != nil {
		cfg = &config.Config{}
	}

	remoteURL := cfg.Remote
	if remoteURL == "" {
		// Try to get from git
		hasRemote, url, err := snapfig.HasRemote()
		if err != nil {
			return err
		}
		if !hasRemote {
			return fmt.Errorf("no remote configured. Run 'snapfig tui' and configure in Settings (F6)")
		}
		remoteURL = url
	}

	fmt.Printf("Pulling from %s...\n", remoteURL)
	result, err := snapfig.PullVaultWithRemote(remoteURL)
	if err != nil {
		return err
	}

	if result.Cloned {
		fmt.Println("Cloned successfully.")
	} else {
		fmt.Println("Pulled successfully.")
	}
	return nil
}
