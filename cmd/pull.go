package cmd

import (
	"fmt"
	"io"

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

// runPull delegates to runPullWithOutput which is unit tested.
func runPull(cmd *cobra.Command, args []string) error {
	return runPullWithOutput(cmd.OutOrStdout())
}

func runPullWithOutput(w io.Writer) error {
	cfg, configPath, err := loadConfigWithPath()
	if err != nil {
		return err
	}

	svc, err := ServiceFactory(cfg, configPath)
	if err != nil {
		return err
	}

	remoteURL := cfg.Remote
	if remoteURL == "" {
		// Try to get from git
		hasRemote, url, err := HasRemoteFunc(svc.VaultDir())
		if err != nil {
			return err
		}
		if !hasRemote {
			return fmt.Errorf("no remote configured. Run 'snapfig' and configure in Settings (F9)")
		}
		remoteURL = url
	}

	fmt.Fprintf(w, "Pulling from %s...\n", remoteURL)
	result, err := svc.Pull()
	if err != nil {
		return err
	}

	if result.Cloned {
		fmt.Fprintln(w, "Cloned successfully.")
	} else {
		fmt.Fprintln(w, "Pulled successfully.")
	}
	return nil
}
