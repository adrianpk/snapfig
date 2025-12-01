package cmd

import (
	"fmt"

	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/spf13/cobra"
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
	hasRemote, url, err := snapfig.HasRemote()
	if err != nil {
		return err
	}
	if !hasRemote {
		return fmt.Errorf("no remote configured. Run: cd ~/.snapfig/vault && git remote add origin <url>")
	}

	fmt.Printf("Pushing to %s...\n", url)
	if err := snapfig.PushVault(); err != nil {
		return err
	}

	fmt.Println("Done.")
	return nil
}
