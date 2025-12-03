// Package cmd implements the CLI commands.
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "snapfig",
	Short: "A tool for managing and versioning configuration files",
	Long: `Snapfig observes directories and replicates their contents
into a versioned store without requiring symlinks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to TUI when no subcommand is specified
		return tuiCmd.RunE(cmd, args)
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/snapfig/config.yml)")
	rootCmd.PersistentFlags().String("git", "disable", "git mode: 'disable' or 'remove'")

	viper.BindPFlag("git", rootCmd.PersistentFlags().Lookup("git"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configPath := filepath.Join(home, ".config", "snapfig")
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.ReadInConfig()
}
