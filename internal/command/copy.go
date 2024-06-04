package command

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const SnapfigDir = ".snapfig"

type CopyCommand struct {
	*cobra.Command
	fs   afero.Fs
	dirs []string
}

var CopyCmd = NewCopyCommand()

func NewCopyCommand() *CopyCommand {
	cmd := &CopyCommand{
		Command: &cobra.Command{
			Use:   "copy",
			Short: "Copy the content of each collected dir into a new dir",
		},
		dirs: []string{},
	}

	cmd.Command.RunE = cmd.run

	return cmd
}

func (cmd *CopyCommand) run(c *cobra.Command, args []string) error {
	cmd.fs = afero.NewBasePathFs(afero.NewOsFs(), SnapfigDir)

	for _, dir := range cmd.dirs {
		err := afero.Walk(cmd.fs, dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			destPath := filepath.Join(SnapfigDir, path)
			if info.IsDir() {
				return afero.NewOsFs().MkdirAll(destPath, info.Mode())
			}
			return cmd.CopyFile(path, destPath)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *CopyCommand) CopyFile(srcFile, dstFile string) error {
	content, err := afero.ReadFile(cmd.fs, srcFile)
	if err != nil {
		return err
	}
	err = afero.WriteFile(cmd.fs, dstFile, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *CopyCommand) SetDirs(dirs []string) {
	c.dirs = dirs
}
