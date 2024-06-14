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

		info, err := w.fs.Stat(watched.Path)
		if os.IsNotExist(err) {
			continue
		}

		destPath := filepath.Join(SnapfigDir, DefaultVaultDir, watched.Path)

		if err == nil && info.IsDir() {
			if err := w.CopyDir(watched.Path, destPath); err != nil {
				return err
			}
			continue
		}

		if err := w.CopyFile(watched.Path, destPath); err != nil {
			return err
		}
	}

	return nil
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must not exist.
func (w *Worker) CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	info, err := w.fs.Stat(src)
	if err != nil {
		return err
	}

	err = w.fs.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	entries, err := afero.ReadDir(w.fs, src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = w.CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = w.CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
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
