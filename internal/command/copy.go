package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	SnapfigDir      = ".snapfig"
	DefaultVaultDir = "vault"
)

var CopyCommand = &cobra.Command{
	Use:   "copy",
	Short: "Copy selected locations into snapfig's vault dir",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := config.Config{}

		w := NewWorker(&cfg)

		return w.Copy()
	},
}

func (w *Worker) Copy() error {
	err := w.Setup()
	if err != nil {
		return err
	}

	var snapfig *Snapfig
	
	configFile := filepath.Join(w.configDir, ConfigFile)  
	data, err := afero.ReadFile(w.fs, configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &snapfig)
	if err != nil {
		return err
	}

	if snapfig.Git == "" {
		snapfig.Git = "disable"
	}

	for _, watched := range snapfig.Watching {
		if watched.Git == "" {
			watched.Git = snapfig.Git
		}

		err := afero.Walk(w.fs, watched.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			destPath := filepath.Join(SnapfigDir, DefaultVaultDir, path)
			if !info.IsDir() {
				return nil
			}

			if info.Name() == ".git" && watched.Git == "disable" {
				return afero.NewOsFs().Rename(path, path+"_disabled")
			}

			if info.Name() == ".git" && watched.Git == "remove" {
				return afero.NewOsFs().RemoveAll(path)
			}

			return afero.NewOsFs().MkdirAll(destPath, info.Mode())
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) ensureVaultDir() (string, error) {
	vaultDir := filepath.Join(w.fsRoot, SnapfigDir, DefaultVaultDir)
	if err := w.fs.MkdirAll(vaultDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create vault directory: %w", err)
	}
	return vaultDir, nil
}

func (w *Worker) CopyFile(srcFile, dstFile string) error {
	content, err := afero.ReadFile(w.fs, srcFile)
	if err != nil {
		return err
	}
	err = afero.WriteFile(w.fs, dstFile, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
