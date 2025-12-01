// Package paths provides known configuration paths across different platforms.
package paths

import (
	"os"
	"path/filepath"
)

// Known returns a list of well-known configuration paths for Linux systems.
// These paths are relative to the user's home directory.
func Known() []string {
	return []string{
		".bashrc",
		".bash_profile",
		".profile",
		".zshrc",
		".zprofile",

		".config/nvim",
		".config/helix",
		".config/emacs",
		".config/doom",
		".config/vim",

		".config/alacritty",
		".config/kitty",
		".config/wezterm",
		".config/foot",

		".config/tmux",
		".tmux.conf",

		".config/i3",
		".config/sway",
		".config/hypr",
		".config/waybar",
		".config/polybar",
		".config/rofi",
		".config/wofi",
		".config/dunst",
		".config/picom",

		".config/fish",
		".config/starship.toml",

		".config/git",
		".gitconfig",
		".gitignore_global",

		".ssh/config",

		".config/Code/User/settings.json",
		".config/Code/User/keybindings.json",

		".config/gtk-3.0",
		".config/gtk-4.0",
		".config/qt5ct",
		".config/qt6ct",

		".config/fontconfig",
		".config/mimeapps.list",

		".config/mpv",
		".config/ranger",
		".config/lf",
		".config/yazi",
		".config/zathura",

		".config/docker",
		".config/k9s",
		".kube/config",

		".config/npm",
		".npmrc",
		".cargo/config.toml",
		".rustup/settings.toml",

		".gnupg/gpg.conf",
		".gnupg/gpg-agent.conf",
	}
}

// Existing filters the known paths to only those that exist on the filesystem.
func Existing() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var existing []string
	for _, p := range Known() {
		full := filepath.Join(home, p)
		if _, err := os.Stat(full); err == nil {
			existing = append(existing, p)
		}
	}
	return existing, nil
}
