package cmd

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/tui"
)

var demoMode bool

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal interface",
	RunE:  runTUI,
}

func init() {
	tuiCmd.Flags().BoolVar(&demoMode, "demo", false, "Demo mode: only show safe paths for screenshots")
	rootCmd.AddCommand(tuiCmd)
}

// runTUI launches the interactive terminal interface.
// Requires interactive terminal; TUI logic tested via model_test.go.
func runTUI(cmd *cobra.Command, args []string) error {
	configDir, err := config.DefaultConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	configPath := filepath.Join(configDir, "config.yml")

	// Load existing config or create new one
	cfg, err := config.Load(configPath)
	if err != nil {
		cfg = &config.Config{
			Git: config.GitModeDisable,
		}
	}

	m := tui.New(cfg, configPath, demoMode)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
