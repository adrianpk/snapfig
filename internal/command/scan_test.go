package command_test

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/adrianpk/snapfig/internal/command"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestScanCommand(t *testing.T) {
	fs := afero.NewMemMapFs()

	homeDir := "home/user"
	actualDirs := []string{
		homeDir,
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, ".config/nvim"),
		filepath.Join(homeDir, ".config/emacs"),
	}

	for _, dir := range actualDirs {
		err := fs.MkdirAll(dir, 0755)
		assert.NoError(t, err)
	}

	cmd := command.ScanCmd
	cmd.SetFS(fs)

	err := cmd.Flags().Set("path", homeDir)
	if err != nil {
		t.Fatal(err)
	}

	err = command.ScanCmd.RunE()
	if err != nil {
		t.Fatal(err)
	}

	configFilePath := filepath.Join(homeDir, ".config/snapfig/config.yml")
	configFile, err := fs.Open(configFilePath)
	assert.NoError(t, err)

	configData, err := io.ReadAll(configFile)
	assert.NoError(t, err)

	var config map[string][]string
	err = yaml.Unmarshal(configData, &config)
	assert.NoError(t, err)

	snapfigDir := filepath.Join(homeDir, ".config/snapfig")
	// Next one is created by snapfig and is a common location, so it should be in the config
	actualDirs = append(actualDirs, snapfigDir)

	assert.Equal(t, actualDirs, config["watching"])
}
