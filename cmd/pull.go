package cmd

import (
	"fmt"

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
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	vaultDir, err := cfg.VaultDir()
	if err != nil {
		return err
	}

	remoteURL := cfg.Remote
	if remoteURL == "" {
		// Try to get from git
		hasRemote, url, err := snapfig.HasRemote(vaultDir)
		if err != nil {
			return err
		}
		if !hasRemote {
			return fmt.Errorf("no remote configured. Run 'snapfig' and configure in Settings (F9)")
		}
		remoteURL = url
	}

	fmt.Printf("Pulling from %s...\n", remoteURL)
	result, err := snapfig.PullVaultWithToken(vaultDir, remoteURL, cfg.GitToken)
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
