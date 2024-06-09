package main

import (
	"log"

	"github.com/adrianpk/snapfig/internal/command"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "snapfig",
		Short: "A tool for versioning system configuration files",
	}

	rootCmd.PersistentFlags().String("git", "disable", "Git flag can be 'remove' or 'disable'")
	viper.BindPFlag("git", rootCmd.PersistentFlags().Lookup("git"))

	scanCmd := command.ScanCommand
	rootCmd.AddCommand(scanCmd)

	copyCmd := command.CopyCommand
	rootCmd.AddCommand(copyCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
