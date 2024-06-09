package command

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

type Snapfig struct {
	Git      string    `yaml:"git"`
	Watching []Watched `yaml:"watching"`
}

type Watched struct {
	Path string `yaml:"path"`
	Git  string `yaml:"git"`
}

var ScanCommand = &cobra.Command{
	Use:   "scan",
	Short: "Scan common locations in the user's home directory",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := config.Config{
			Git: viper.GetString("git"),
		}

		err := cfg.Validate()
		if err != nil {
			return err
		}

		w := NewWorker(&cfg)

		return w.Scan()
	},
}

func (w *Worker) Scan() error {
	if w.fsRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		w.fsRoot = home
	}

	w.fs = afero.NewBasePathFs(afero.NewOsFs(), w.fsRoot)

	toolDir, err := w.ensureSnapfigDir()
	if err != nil {
		return err
	}

	w.toolDir = toolDir

	err = w.doScan(commonLocations())
	if err != nil {
		return err
	}

	err = w.createSnapfigFile()
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) ensureSnapfigDir() (string, error) {
	toolDir := filepath.Join(".config", "snapfig")
	if err := w.fs.MkdirAll(toolDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return toolDir, nil
}

func (w *Worker) doScan(dirs []string) error {
	w.userMsg("Scanning directories:")
	for _, dir := range dirs {
		if _, err := w.fs.Stat(dir); os.IsNotExist(err) {
			w.userMsg(fmt.Sprintf("Discarding non-existent directory: %s", dir))
			continue
		}

		err := afero.Walk(w.fs, dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			w.userMsg(fmt.Sprintf("Error walking directory: %s", dir))
			continue
		}

		w.userMsg(fmt.Sprintf("Processing directory: %s", dir))
		w.dirs = append(w.dirs, dir)
	}

	return nil
}

func (w *Worker) createSnapfigFile() error {
	// snapfigFile := filepath.Join(wfsRoot, w.toolDir, "config.yml")
	snapfigFile := filepath.Join(w.toolDir, "config.yml")

	err := w.moveOld(snapfigFile)
	if err != nil {
		return err
	}

	file, err := w.fs.Create(snapfigFile)
	if err != nil {
		w.userMsg(fmt.Sprintf("Error creating config file %s: %v\n", snapfigFile, err))
		return err
	}
	defer file.Close()

	watched := make([]Watched, len(w.dirs))
	for i, path := range w.dirs {
		watched[i] = Watched{Path: path, Git: w.Cfg().Git}
	}

	snapfig := &Snapfig{
		Git:      w.Cfg().Git,
		Watching: watched,
	}

	data, err := yaml.Marshal(snapfig)
	if err != nil {
		w.userMsg(fmt.Sprintf("Error marshaling data: %v\n", err))
		return err
	}

	fmt.Printf("Writing to file: %s\n", snapfigFile)
	_, err = file.Write(data)
	if err != nil {
		w.userMsg(fmt.Sprintf("Error writing file: %v\n", err))
		return err
	}

	w.userMsg(fmt.Sprintf("Config file created at %s\n", snapfigFile))
	w.userMsg("Common locations have been included in the config file.")
	w.userMsg("Feel free to modify it to suit your needs.")

	return nil
}

func (w *Worker) moveOld(fullPath string) error {
	fmt.Printf("Checking if file exists: %s\n", fullPath)
	if _, err := w.fs.Stat(fullPath); err == nil {
		dateSuffix := time.Now().Format("20060102150405")
		newName := filepath.Join(w.toolDir, "config.yml."+dateSuffix)

		err := w.fs.Rename(fullPath, newName)
		if err != nil {
			return fmt.Errorf("error moving old config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("Error checking if file exists: %v\n", err)
		return fmt.Errorf("error checking if config file exists: %w", err)
	}

	return nil
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
