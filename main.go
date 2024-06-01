package main

import (
	"log"

	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "snapfig",
		Short: "A tool for versioning system configuration files",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			configFile := cmd.Flag("config").Value.String()
			if configFile == "" {
				log.Fatal("Config file path is required")
			}

			config, err := snapfig.LoadConfig(configFile)
			if err != nil {
				log.Fatalf("Error initializing configuration: %v", err)
			}
			log.Printf("Config: %+v", config)
		},
	}

	initCmd.Flags().String("config", "", "Path to configuration file")
	initCmd.MarkFlagRequired("config")

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
