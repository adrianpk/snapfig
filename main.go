package main

import (
	"fmt"
	"os"

	"github.com/adrianpk/snapfig/internal/command"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "snapfig",
		Short: "A tool for versioning system configuration files",
	}

	rootCmd.AddCommand(command.ScanCmd.Command)
	rootCmd.AddCommand(command.CopyCmd.Command)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
