// Package screens contains the individual TUI screens.
package screens

import (
	"strings"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	fieldRemote = iota
	fieldVaultPath
	fieldCopyInterval
	fieldPushInterval
	fieldPullInterval
	fieldAutoRestore
	fieldCount
)

// SettingsModel handles the settings screen.
type SettingsModel struct {
	remoteInput       textinput.Model
	vaultPathInput    textinput.Model
	copyIntervalInput textinput.Model
	pushIntervalInput textinput.Model
	pullIntervalInput textinput.Model
	autoRestore       bool
	focused           int
	width             int
	height            int
	saved             bool
}

// NewSettings creates a new settings screen.
func NewSettings(currentRemote, currentVaultPath string, daemon config.DaemonConfig) SettingsModel {
	remote := textinput.New()
	remote.Placeholder = "git@github.com:user/dotfiles.git"
	remote.CharLimit = 256
	remote.Width = 50
	remote.SetValue(currentRemote)

	vaultPath := textinput.New()
	vaultPath.Placeholder = "~/.snapfig/vault (default)"
	vaultPath.CharLimit = 256
	vaultPath.Width = 50
	vaultPath.SetValue(currentVaultPath)

	copyInt := textinput.New()
	copyInt.Placeholder = "1h"
	copyInt.CharLimit = 16
	copyInt.Width = 20
	copyInt.SetValue(daemon.CopyInterval)

	pushInt := textinput.New()
	pushInt.Placeholder = "24h"
	pushInt.CharLimit = 16
	pushInt.Width = 20
	pushInt.SetValue(daemon.PushInterval)

	pullInt := textinput.New()
	pullInt.Placeholder = "disabled"
	pullInt.CharLimit = 16
	pullInt.Width = 20
	pullInt.SetValue(daemon.PullInterval)

	m := SettingsModel{
		remoteInput:       remote,
		vaultPathInput:    vaultPath,
		copyIntervalInput: copyInt,
		pushIntervalInput: pushInt,
		pullIntervalInput: pullInt,
		autoRestore:       daemon.AutoRestore,
		focused:           fieldRemote,
	}
	m.updateFocus()
	return m
}

func (m SettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.saved = true
			return m, nil
		case "esc":
			return m, nil
		case "tab", "down":
			m.focused = (m.focused + 1) % fieldCount
			m.updateFocus()
			return m, nil
		case "shift+tab", "up":
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			m.updateFocus()
			return m, nil
		case " ":
			if m.focused == fieldAutoRestore {
				m.autoRestore = !m.autoRestore
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case fieldRemote:
		m.remoteInput, cmd = m.remoteInput.Update(msg)
	case fieldVaultPath:
		m.vaultPathInput, cmd = m.vaultPathInput.Update(msg)
	case fieldCopyInterval:
		m.copyIntervalInput, cmd = m.copyIntervalInput.Update(msg)
	case fieldPushInterval:
		m.pushIntervalInput, cmd = m.pushIntervalInput.Update(msg)
	case fieldPullInterval:
		m.pullIntervalInput, cmd = m.pullIntervalInput.Update(msg)
	}
	return m, cmd
}

func (m *SettingsModel) updateFocus() {
	m.remoteInput.Blur()
	m.vaultPathInput.Blur()
	m.copyIntervalInput.Blur()
	m.pushIntervalInput.Blur()
	m.pullIntervalInput.Blur()

	switch m.focused {
	case fieldRemote:
		m.remoteInput.Focus()
	case fieldVaultPath:
		m.vaultPathInput.Focus()
	case fieldCopyInterval:
		m.copyIntervalInput.Focus()
	case fieldPushInterval:
		m.pushIntervalInput.Focus()
	case fieldPullInterval:
		m.pullIntervalInput.Focus()
	}
}

func (m SettingsModel) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Settings"))
	b.WriteString("\n\n")

	// Remote
	b.WriteString(styles.Normal.Render("Remote URL:"))
	b.WriteString("\n")
	b.WriteString(m.remoteInput.View())
	b.WriteString("\n\n")

	// Vault path
	b.WriteString(styles.Normal.Render("Vault location (leave empty for default):"))
	b.WriteString("\n")
	b.WriteString(m.vaultPathInput.View())
	b.WriteString("\n\n")

	// Daemon section
	b.WriteString(styles.Subtitle.Render("Background Runner"))
	b.WriteString("\n\n")

	b.WriteString(styles.Normal.Render("Copy interval (e.g. 1h, 30m):"))
	b.WriteString("\n")
	b.WriteString(m.copyIntervalInput.View())
	b.WriteString("\n\n")

	b.WriteString(styles.Normal.Render("Push interval (e.g. 24h):"))
	b.WriteString("\n")
	b.WriteString(m.pushIntervalInput.View())
	b.WriteString("\n\n")

	b.WriteString(styles.Normal.Render("Pull interval (leave empty to disable):"))
	b.WriteString("\n")
	b.WriteString(m.pullIntervalInput.View())
	b.WriteString("\n\n")

	// Auto restore checkbox
	checkbox := "[ ]"
	if m.autoRestore {
		checkbox = "[x]"
	}
	label := styles.Normal.Render("Auto restore after pull:")
	if m.focused == fieldAutoRestore {
		checkbox = styles.Selected.Render(checkbox)
	}
	b.WriteString(label + " " + checkbox)
	b.WriteString("\n\n")

	b.WriteString(styles.Help.Render("Tab/↑↓ navigate • Space toggle • Enter save • Esc cancel"))

	return b.String()
}

// Remote returns the current remote URL value.
func (m SettingsModel) Remote() string {
	return strings.TrimSpace(m.remoteInput.Value())
}

// VaultPath returns the current vault path value.
func (m SettingsModel) VaultPath() string {
	return strings.TrimSpace(m.vaultPathInput.Value())
}

// DaemonConfig returns the daemon configuration values.
func (m SettingsModel) DaemonConfig() config.DaemonConfig {
	return config.DaemonConfig{
		CopyInterval: strings.TrimSpace(m.copyIntervalInput.Value()),
		PushInterval: strings.TrimSpace(m.pushIntervalInput.Value()),
		PullInterval: strings.TrimSpace(m.pullIntervalInput.Value()),
		AutoRestore:  m.autoRestore,
	}
}

// WasSaved returns true if user pressed Enter to save.
func (m SettingsModel) WasSaved() bool {
	return m.saved
}
