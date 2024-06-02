package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan common locations in the user's home directory",
	Run: func(cmd *cobra.Command, args []string) {
		startingPath, _ := cmd.Flags().GetString("path")
		if startingPath == "" {
			startingPath, _ = os.UserHomeDir()
		}

		configDir := ensureConfigDir(startingPath)

		// Well known locations (TODO: Add more later)
		commonLocations := []string{
			startingPath,
			filepath.Join(startingPath, "Documents"),
			filepath.Join(startingPath, ".config/nvim"),
			filepath.Join(startingPath, ".config/emacs"),
			filepath.Join(startingPath, ".config/doom"),
			filepath.Join(startingPath, ".config/kde*"),
			filepath.Join(startingPath, ".config/gnome"),
			filepath.Join(startingPath, ".config/lxde"),
			filepath.Join(startingPath, ".config/xfce"),
			filepath.Join(startingPath, ".config/code"),
			filepath.Join(startingPath, ".config/snapfig"),
		}

		doScan(commonLocations)
		createConfigFile(configDir, commonLocations)
	},
}

func userMsg(message string) {
	fmt.Println(message)
}

// Function to check and create snapfig's config directory
func ensureConfigDir(startingPath string) string {
	configDir := filepath.Join(startingPath, ".config/snapfig")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			userMsg(fmt.Sprintf("Error creating config directory %s: %v\n", configDir, err))
			os.Exit(1)
		}
	}
	return configDir
}

func doScan(dirs []string) {
	userMsg("Scanned directories:")
	for _, loc := range dirs {
		matches, err := filepath.Glob(loc)
		if err != nil {
			userMsg(fmt.Sprintf("Error scanning directory %s: %v\n", loc, err))
			continue
		}
		for _, match := range matches {
			userMsg(match)
		}
	}
}

func createConfigFile(configDir string, dirs []string) {
	configFilePath := filepath.Join(configDir, "config.yml")
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		f, err := os.Create(configFilePath)
		if err != nil {
			userMsg(fmt.Sprintf("Error creating config file %s: %v\n", configFilePath, err))
			os.Exit(1)
		}
		defer f.Close()

		config := map[string][]string{
			"watching": dirs,
		}

		data, err := yaml.Marshal(&config)
		if err != nil {
			userMsg(fmt.Sprintf("Error marshaling config data: %v\n", err))
			os.Exit(1)
		}

		_, err = f.Write(data)
		if err != nil {
			userMsg(fmt.Sprintf("Error writing to config file %s: %v\n", configFilePath, err))
			os.Exit(1)
		}

		userMsg(fmt.Sprintf("\nConfig file created at: %s\n\n", configFilePath))
		userMsg("Common locations have been included in the config file.")
		userMsg("Feel free to append other directories as needed.")
		return
	}
	userMsg(fmt.Sprintf("\nConfig file already exists at: %s\n\n", configFilePath))
}
