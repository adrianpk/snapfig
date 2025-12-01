// Package screens contains the individual TUI screens.
package screens

import (
	"strings"

	"github.com/adrianpk/snapfig/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SettingsModel handles the settings screen.
type SettingsModel struct {
	remoteInput textinput.Model
	width       int
	height      int
	saved       bool
}

// NewSettings creates a new settings screen.
func NewSettings(currentRemote string) SettingsModel {
	ti := textinput.New()
	ti.Placeholder = "git@github.com:user/dotfiles.git"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(currentRemote)

	return SettingsModel{
		remoteInput: ti,
	}
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
		}
	}

	var cmd tea.Cmd
	m.remoteInput, cmd = m.remoteInput.Update(msg)
	return m, cmd
}

func (m SettingsModel) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Settings"))
	b.WriteString("\n\n")

	b.WriteString(styles.Normal.Render("Remote URL:"))
	b.WriteString("\n")
	b.WriteString(m.remoteInput.View())
	b.WriteString("\n\n")

	b.WriteString(styles.Help.Render("Enter to save • Esc to cancel"))

	return b.String()
}

// Remote returns the current remote URL value.
func (m SettingsModel) Remote() string {
	return strings.TrimSpace(m.remoteInput.Value())
}

// WasSaved returns true if user pressed Enter to save.
func (m SettingsModel) WasSaved() bool {
	return m.saved
}
