package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

type Command struct {
	*cobra.Command
	fs   afero.Fs
	root string
	dirs []string
}

var ScanCmd = NewScanCommand()

func NewScanCommand() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "scan",
			Short: "Scan common locations in the user's home directory",
		},
		dirs: []string{},
	}

	cmd.Flags().StringVarP(&cmd.root, "path", "p", "", "Path to start scanning from")
	cmd.Command.RunE = cmd.run

	return cmd
}

func (c *Command) RunE() error {
	return c.Command.RunE(c.Command, nil)
}

func (cmd *Command) run(c *cobra.Command, args []string) error {
	if cmd.root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cmd.root = home
	}

	cmd.fs = afero.NewBasePathFs(afero.NewOsFs(), cmd.root)

	configDir, err := cmd.ensureConfigDir()
	if err != nil {
		return err
	}

	err = cmd.doScan(commonLocations())
	if err != nil {
		return err
	}

	err = cmd.createConfigFile(configDir)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *Command) ensureConfigDir() (string, error) {
	configDir := filepath.Join(".config", "snapfig")
	if err := cmd.fs.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return configDir, nil
}

func (cmd *Command) doScan(dirs []string) error {
	cmd.userMsg("Scanning directories:")
	for _, dir := range dirs {
		if _, err := cmd.fs.Stat(dir); os.IsNotExist(err) {
			cmd.userMsg(fmt.Sprintf("Discarding non-existent directory: %s", dir))
			continue
		}

		err := afero.Walk(cmd.fs, dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			cmd.userMsg(fmt.Sprintf("Error walking directory: %s", dir))
			continue
		}

		cmd.userMsg(fmt.Sprintf("Processing directory: %s", dir))
		cmd.dirs = append(cmd.dirs, dir)
	}

	return nil
}

func (cmd *Command) createConfigFile(configDir string) error {
	configFile := filepath.Join(configDir, "config.yml")
	if _, err := cmd.fs.Stat(configFile); os.IsNotExist(err) {
		file, err := cmd.fs.Create(configFile)
		if err != nil {
			cmd.userMsg(fmt.Sprintf("Error creating config file %s: %v\n", configFile, err))
			return err
		}
		defer file.Close()

		config := map[string][]string{
			"watching": cmd.dirs,
		}

		encoder := yaml.NewEncoder(file)
		if err := encoder.Encode(&config); err != nil {
			cmd.userMsg(fmt.Sprintf("Error marshaling config data: %v\n", err))
			return err
		}

		cmd.userMsg(fmt.Sprintf("\nConfig file created at: %s\n\n", configFile))
		cmd.userMsg("Common locations have been included in the config file.")
		cmd.userMsg("Feel free to append other directories as needed.")
		return nil
	}

	return nil
}

func (cmd *Command) userMsg(message string) {
	fmt.Println(message)
}

// SetFS sets a custom filesystem.
// It's only used for tests for now.
func (c *Command) SetFS(fs afero.Fs) {
	c.fs = fs
}

func commonLocations() []string {
	return []string{
		".",
		".config/nvim",
		".config/emacs",
		".config/doom",
		".config/kde*",
		".config/gnome",
		".config/lxde",
		".config/xfce",
		".config/code",
		".config/snapfig",
		".config/vscode",
		".config/git",
		".config/ssh",
		".config/bash",
		".config/zsh",
		".config/fish",
		".config/fzf",
		".config/vim",
		".config/tmux",
		".config/i3",
		".config/xresources",
		".config/xinitrc",
		".config/xsession",
		".config/gitconfig",
		".config/gitignore",
		".config/npmrc",
		".config/cargo",
		".config/kube",
		".config/aws",
		".config/gnupg",
		".config/docker",
		".config/fluxbox",
		".config/bspwm",
		".config/herbstluftwm",
		".config/openbox",
		".config/awesome",
		".config/sway",
		".config/dwm",
		".config/qtile",
		".config/spectrwm",
		".config/wayland",
		".config/xmonad",
		".config/xmobar",
		".config/stumpwm",
		".config/lisp",
		".config/rustup",
		".config/helix",
		".config/inkscape",
		".config/gimp",
		".config/blender",
		".config/thunderbird",
		".config/filezilla",
		".config/transmission",
		".config/audacity",
		".config/vlc",
		".config/mpv",
		".config/spotify",
		".config/discord",
		".config/slack",
		".config/telegram",
		".config/zoom",
		".config/sublime-text",
		".config/atom",
		".config/gedit",
		".config/notepad++",
		"AppData",
		"Documents/Outlook Files",
	}
}
