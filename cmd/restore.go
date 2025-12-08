package cmd

import (
	"fmt"
	"io"

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

// runRestore delegates to runRestoreWithOutput which is unit tested.
func runRestore(cmd *cobra.Command, args []string) error {
	return runRestoreWithOutput(cmd.OutOrStdout())
}

func runRestoreWithOutput(w io.Writer) error {
	cfg, configPath, err := loadConfigWithPath()
	if err != nil {
		return err
	}

	if len(cfg.Watching) == 0 {
		fmt.Fprintln(w, "No paths configured. Run 'snapfig tui' to select paths.")
		return nil
	}

	svc, err := ServiceFactory(cfg, configPath)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "Restoring from vault...")
	result, err := svc.Restore()
	if err != nil {
		return err
	}

	for _, p := range result.Backups {
		fmt.Fprintf(w, "  Backed up: %s\n", p)
	}
	for _, p := range result.Restored {
		fmt.Fprintf(w, "  Restored: %s\n", p)
	}
	for _, p := range result.Skipped {
		fmt.Fprintf(w, "  Skipped: %s (not in vault)\n", p)
	}

	fmt.Fprintf(w, "\nDone. %d restored, %d backed up, %d skipped.\n",
		len(result.Restored), len(result.Backups), len(result.Skipped))
	return nil
}
