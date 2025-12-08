package cmd

import (
	"fmt"
	"io"

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

// runPush delegates to runPushWithOutput which is unit tested.
func runPush(cmd *cobra.Command, args []string) error {
	return runPushWithOutput(cmd.OutOrStdout())
}

func runPushWithOutput(w io.Writer) error {
	cfg, configPath, err := loadConfigWithPath()
	if err != nil {
		return err
	}

	svc, err := ServiceFactory(cfg, configPath)
	if err != nil {
		return err
	}

	hasRemote, url, err := HasRemoteFunc(svc.VaultDir())
	if err != nil {
		return err
	}
	if !hasRemote {
		return fmt.Errorf("no remote configured. Run: cd %s && git remote add origin <url>", svc.VaultDir())
	}

	fmt.Fprintf(w, "Pushing to %s...\n", url)
	if err := svc.Push(); err != nil {
		return err
	}

	fmt.Fprintln(w, "Done.")
	return nil
}
