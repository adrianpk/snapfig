package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy watched paths to the vault",
	Long:  "Copies all enabled watched paths from the config to the vault, handling .git directories according to the configured mode.",
	RunE:  runCopy,
}

func init() {
	rootCmd.AddCommand(copyCmd)
}

// runCopy delegates to runCopyWithOutput which is unit tested.
func runCopy(cmd *cobra.Command, args []string) error {
	return runCopyWithOutput(cmd.OutOrStdout())
}

func runCopyWithOutput(w io.Writer) error {
	cfg, configPath, err := loadConfigWithPath()
	if err != nil {
		return err
	}

	if len(cfg.Watching) == 0 {
		fmt.Fprintln(w, "No paths configured. Run 'snapfig' to select paths.")
		return nil
	}

	svc, err := ServiceFactory(cfg, configPath)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "Copying to vault...")
	result, err := svc.Copy()
	if err != nil {
		return err
	}

	for _, p := range result.Copied {
		fmt.Fprintf(w, "  Copied: %s\n", p)
	}
	for _, p := range result.Skipped {
		fmt.Fprintf(w, "  Skipped: %s (not found)\n", p)
	}

	fmt.Fprintf(w, "\nDone. %d copied, %d skipped. Vault: %s\n", len(result.Copied), len(result.Skipped), svc.VaultDir())
	return nil
}
