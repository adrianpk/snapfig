package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy watched paths to the vault",
	Long:  "Copies all enabled watched paths from the config to ~/.snapfig/vault/, handling .git directories according to the configured mode.",
	RunE:  runCopy,
}

func init() {
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
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

	copier, err := snapfig.NewCopier(cfg)
	if err != nil {
		return err
	}

	fmt.Println("Copying to vault...")
	result, err := copier.Copy()
	if err != nil {
		return err
	}

	for _, p := range result.Copied {
		fmt.Printf("  Copied: %s\n", p)
	}
	for _, p := range result.Skipped {
		fmt.Printf("  Skipped: %s (not found)\n", p)
	}

	vaultDir, _ := config.DefaultVaultDir()
	fmt.Printf("\nDone. %d copied, %d skipped. Vault: %s\n", len(result.Copied), len(result.Skipped), vaultDir)
	return nil
}
