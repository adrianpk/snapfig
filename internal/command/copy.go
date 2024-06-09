package command

import (
	"os"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const SnapfigDir = ".snapfig"

var CopyCommand = &cobra.Command{
	Use:   "copy",
	Short: "Copy selected locations into snapfig's vault dir",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := config.Config{} // No values to set for now

		w := NewWorker(&cfg)

		return w.Copy()
	},
}

func (w *Worker) Copy() error {
	w.fs = afero.NewBasePathFs(afero.NewOsFs(), SnapfigDir)

	for _, dir := range w.dirs {
		err := afero.Walk(w.fs, dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			destPath := filepath.Join(SnapfigDir, path)
			if info.IsDir() {
				return afero.NewOsFs().MkdirAll(destPath, info.Mode())
			}
			return w.CopyFile(path, destPath)
		})

		if err != nil {
			return err
		}
	}

	return nil
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

func (w *Worker) SetDirs(dirs []string) {
	w.dirs = dirs
}
