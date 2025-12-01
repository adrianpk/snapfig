package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore paths from the vault",
	Long:  "Restores all enabled watched paths from ~/.snapfig/vault/ to their original locations. Existing files are backed up with a .YYYYMMDDHHMM.bak suffix before overwriting.",
	RunE:  runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
	configDir, err := config.DefaultConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	configPath := filepath.Join(configDir, "config.yml")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Watching) == 0 {
		fmt.Println("No paths configured. Run 'snapfig tui' to select paths.")
		return nil
	}

	restorer, err := snapfig.NewRestorer(cfg)
	if err != nil {
		return err
	}

	fmt.Println("Restoring from vault...")
	result, err := restorer.Restore()
	if err != nil {
		return err
	}

	for _, p := range result.Backups {
		fmt.Printf("  Backed up: %s\n", p)
	}
	for _, p := range result.Restored {
		fmt.Printf("  Restored: %s\n", p)
	}
	for _, p := range result.Skipped {
		fmt.Printf("  Skipped: %s (not in vault)\n", p)
	}

	fmt.Printf("\nDone. %d restored, %d backed up, %d skipped.\n",
		len(result.Restored), len(result.Backups), len(result.Skipped))
	return nil
}
