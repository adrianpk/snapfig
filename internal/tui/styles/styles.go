// Package styles defines the visual styling for the TUI.
package styles

import "github.com/charmbracelet/lipgloss"

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	Subtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	Selected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	Dimmed = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("78"))

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	Cursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	ActionKey = lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("232")).
			Padding(0, 1)

	ActionLabel = lipgloss.NewStyle().
			Background(lipgloss.Color("24")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
)

const (
	CheckedBox   = "[x]"
	UncheckedBox = "[ ]"
	GitBox       = "[g]"
	CursorChar   = ">"
	NoCursor     = " "
)
