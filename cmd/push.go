package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/adrianpk/snapfig/internal/snapfig"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push vault to remote",
	Long:  "Pushes the vault git repository to the configured remote origin.",
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	vaultDir, err := cfg.VaultDir()
	if err != nil {
		return err
	}

	hasRemote, url, err := snapfig.HasRemote(vaultDir)
	if err != nil {
		return err
	}
	if !hasRemote {
		return fmt.Errorf("no remote configured. Run: cd %s && git remote add origin <url>", vaultDir)
	}

	fmt.Printf("Pushing to %s...\n", url)
	if err := snapfig.PushVault(vaultDir); err != nil {
		return err
	}

	fmt.Println("Done.")
	return nil
}
