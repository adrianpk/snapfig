// Package tui implements the terminal user interface using Bubble Tea.
package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/adrianpk/snapfig/internal/tui/screens"
	"github.com/adrianpk/snapfig/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenPicker screen = iota
	screenSettings
)

// Model is the root TUI model that manages screen navigation.
type Model struct {
	current    screen
	picker     screens.PickerModel
	settings   screens.SettingsModel
	cfg        *config.Config
	configPath string
	width      int
	height     int
	status     string
	busy       bool
	demoMode   bool
}

// CopyDoneMsg is sent when copy operation completes.
type CopyDoneMsg struct {
	err     error
	copied  int
	skipped int
}

// RestoreDoneMsg is sent when restore operation completes.
type RestoreDoneMsg struct {
	err      error
	restored int
	backups  int
	skipped  int
}

// PushDoneMsg is sent when push operation completes.
type PushDoneMsg struct {
	err error
}

// PullDoneMsg is sent when pull operation completes.
type PullDoneMsg struct {
	err    error
	cloned bool
}

// BackupDoneMsg is sent when backup (copy+push) completes.
type BackupDoneMsg struct {
	err     error
	copied  int
	skipped int
}

// SyncDoneMsg is sent when sync (pull+restore) completes.
type SyncDoneMsg struct {
	err      error
	cloned   bool
	restored int
	backups  int
	skipped  int
}

// New creates a new root TUI model.
func New(cfg *config.Config, configPath string, demoMode bool) Model {
	return Model{
		current:    screenPicker,
		picker:     screens.NewPicker(cfg, demoMode),
		cfg:        cfg,
		configPath: configPath,
		demoMode:   demoMode,
	}
}

func (m Model) Init() tea.Cmd {
	return m.picker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Pass to picker with reduced height for action bar
		pickerMsg := tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height - 2}
		updated, cmd := m.picker.Update(pickerMsg)
		m.picker = updated.(screens.PickerModel)
		return m, cmd

	case CopyDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Copied %d paths (%d skipped)", msg.copied, msg.skipped)
		}
		return m, nil

	case RestoreDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Restored %d paths (%d backed up, %d skipped)", msg.restored, msg.backups, msg.skipped)
		}
		return m, nil

	case PushDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = "Pushed to remote"
		}
		return m, nil

	case PullDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else if msg.cloned {
			m.status = "Cloned from remote"
		} else {
			m.status = "Pulled from remote"
		}
		return m, nil

	case BackupDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Backup complete: %d copied, pushed to remote", msg.copied)
		}
		return m, nil

	case SyncDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			action := "pulled"
			if msg.cloned {
				action = "cloned"
			}
			m.status = fmt.Sprintf("Sync complete: %s, %d restored", action, msg.restored)
		}
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c", "f10":
			return m, tea.Quit
		case "f2":
			if !m.busy {
				m.busy = true
				m.status = "Copying..."
				return m, m.doCopy()
			}
			return m, nil
		case "f3":
			if !m.busy {
				m.busy = true
				m.status = "Pushing..."
				return m, m.doPush()
			}
			return m, nil
		case "f4":
			if !m.busy {
				m.busy = true
				m.status = "Pulling..."
				return m, m.doPull()
			}
			return m, nil
		case "f5":
			if !m.busy {
				m.busy = true
				m.status = "Restoring..."
				return m, m.doRestore()
			}
			return m, nil
		case "f7":
			if !m.busy {
				m.busy = true
				m.status = "Backing up (copy + push)..."
				return m, m.doBackup()
			}
			return m, nil
		case "f8":
			if !m.busy {
				m.busy = true
				m.status = "Syncing (pull + restore)..."
				return m, m.doSync()
			}
			return m, nil
		case "f9":
			if !m.busy && m.current == screenPicker {
				m.settings = screens.NewSettings(m.cfg.Remote)
				m.current = screenSettings
				return m, m.settings.Init()
			}
			return m, nil
		}
	}

	// Route to current screen
	switch m.current {
	case screenPicker:
		updated, cmd := m.picker.Update(msg)
		m.picker = updated.(screens.PickerModel)
		return m, cmd

	case screenSettings:
		updated, cmd := m.settings.Update(msg)
		m.settings = updated.(screens.SettingsModel)

		// Check if user pressed Enter or Esc
		if m.settings.WasSaved() {
			m.cfg.Remote = m.settings.Remote()
			if err := m.cfg.Save(m.configPath); err != nil {
				m.status = fmt.Sprintf("Error saving: %v", err)
			} else {
				// Configure git remote
				if m.cfg.Remote != "" {
					if err := snapfig.SetRemote(m.cfg.Remote); err != nil {
						m.status = fmt.Sprintf("Saved config, but git remote failed: %v", err)
					} else {
						m.status = "Settings saved"
					}
				} else {
					m.status = "Settings saved"
				}
			}
			m.current = screenPicker
			return m, nil
		}

		// Check for Esc (settings not saved but we need to go back)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "esc" {
			m.current = screenPicker
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Main content
	switch m.current {
	case screenPicker:
		b.WriteString(m.picker.View())
	case screenSettings:
		b.WriteString(m.settings.View())
	}

	// Pad to fill screen before action bar
	content := b.String()
	lines := strings.Count(content, "\n")
	padding := m.height - lines - 3 // 3 = status + action bar + buffer
	if padding > 0 {
		b.WriteString(strings.Repeat("\n", padding))
	}

	// Status line
	if m.status != "" {
		b.WriteString("\n")
		if strings.HasPrefix(m.status, "Error:") {
			b.WriteString(styles.Error.Render(m.status))
		} else if m.busy {
			b.WriteString(styles.Subtitle.Render(m.status))
		} else {
			b.WriteString(styles.Success.Render(m.status))
		}
	} else {
		b.WriteString("\n")
	}

	// Action bar
	b.WriteString("\n")
	b.WriteString(m.renderActionBar())

	return b.String()
}

func (m Model) renderActionBar() string {
	items := []struct {
		key   string
		label string
	}{
		{"F2", "Copy"},
		{"F3", "Push"},
		{"F4", "Pull"},
		{"F5", "Restore"},
		{"F7", "Backup"},
		{"F8", "Sync"},
		{"F9", "Settings"},
		{"F10", "Quit"},
	}

	var parts []string
	for _, item := range items {
		key := styles.ActionKey.Render(item.key)
		label := styles.ActionLabel.Render(item.label)
		parts = append(parts, key+label)
	}

	return strings.Join(parts, " ")
}

// SelectedPaths returns the paths selected in the picker with their git modes.
func (m Model) SelectedPaths() []screens.Selection {
	return m.picker.Selected()
}

// doCopy saves config and copies to vault.
func (m *Model) doCopy() tea.Cmd {
	return func() tea.Msg {
		// Build config from selection
		selected := m.picker.Selected()
		if len(selected) == 0 {
			return CopyDoneMsg{err: fmt.Errorf("no paths selected")}
		}

		m.cfg.Watching = make([]config.Watched, 0, len(selected))
		for _, sel := range selected {
			gitMode := config.GitModeRemove
			if sel.GitMode == screens.StateDisable {
				gitMode = config.GitModeDisable
			}
			m.cfg.Watching = append(m.cfg.Watching, config.Watched{
				Path:    sel.Path,
				Git:     gitMode,
				Enabled: true,
			})
		}

		// Save config
		configDir, err := config.DefaultConfigDir()
		if err != nil {
			return CopyDoneMsg{err: err}
		}
		configPath := filepath.Join(configDir, "config.yml")
		if err := m.cfg.Save(configPath); err != nil {
			return CopyDoneMsg{err: err}
		}

		// Copy to vault
		copier, err := snapfig.NewCopier(m.cfg)
		if err != nil {
			return CopyDoneMsg{err: err}
		}

		result, err := copier.Copy()
		if err != nil {
			return CopyDoneMsg{err: err}
		}

		return CopyDoneMsg{copied: len(result.Copied), skipped: len(result.Skipped)}
	}
}

// doRestore restores from vault to original locations.
func (m *Model) doRestore() tea.Cmd {
	return func() tea.Msg {
		if len(m.cfg.Watching) == 0 {
			return RestoreDoneMsg{err: fmt.Errorf("no paths configured")}
		}

		restorer, err := snapfig.NewRestorer(m.cfg)
		if err != nil {
			return RestoreDoneMsg{err: err}
		}

		result, err := restorer.Restore()
		if err != nil {
			return RestoreDoneMsg{err: err}
		}

		return RestoreDoneMsg{
			restored: len(result.Restored),
			backups:  len(result.Backups),
			skipped:  len(result.Skipped),
		}
	}
}

// doPush pushes vault to remote.
func (m *Model) doPush() tea.Cmd {
	return func() tea.Msg {
		if err := snapfig.PushVault(); err != nil {
			return PushDoneMsg{err: err}
		}
		return PushDoneMsg{}
	}
}

// doPull pulls vault from remote, cloning if needed.
func (m *Model) doPull() tea.Cmd {
	return func() tea.Msg {
		result, err := snapfig.PullVaultWithRemote(m.cfg.Remote)
		if err != nil {
			return PullDoneMsg{err: err}
		}
		return PullDoneMsg{cloned: result.Cloned}
	}
}

// doBackup performs copy + push in one step.
func (m *Model) doBackup() tea.Cmd {
	return func() tea.Msg {
		// First, do the copy
		selected := m.picker.Selected()
		if len(selected) == 0 {
			return BackupDoneMsg{err: fmt.Errorf("no paths selected")}
		}

		m.cfg.Watching = make([]config.Watched, 0, len(selected))
		for _, sel := range selected {
			gitMode := config.GitModeRemove
			if sel.GitMode == screens.StateDisable {
				gitMode = config.GitModeDisable
			}
			m.cfg.Watching = append(m.cfg.Watching, config.Watched{
				Path:    sel.Path,
				Git:     gitMode,
				Enabled: true,
			})
		}

		// Save config
		if err := m.cfg.Save(m.configPath); err != nil {
			return BackupDoneMsg{err: err}
		}

		// Copy to vault
		copier, err := snapfig.NewCopier(m.cfg)
		if err != nil {
			return BackupDoneMsg{err: err}
		}

		result, err := copier.Copy()
		if err != nil {
			return BackupDoneMsg{err: err}
		}

		// Push
		if err := snapfig.PushVault(); err != nil {
			return BackupDoneMsg{err: fmt.Errorf("copied but push failed: %w", err)}
		}

		return BackupDoneMsg{copied: len(result.Copied), skipped: len(result.Skipped)}
	}
}

// doSync performs pull + restore in one step.
func (m *Model) doSync() tea.Cmd {
	return func() tea.Msg {
		// First, pull (or clone)
		pullResult, err := snapfig.PullVaultWithRemote(m.cfg.Remote)
		if err != nil {
			return SyncDoneMsg{err: err}
		}

		// Then restore
		if len(m.cfg.Watching) == 0 {
			return SyncDoneMsg{err: fmt.Errorf("no paths configured")}
		}

		restorer, err := snapfig.NewRestorer(m.cfg)
		if err != nil {
			return SyncDoneMsg{err: err}
		}

		restoreResult, err := restorer.Restore()
		if err != nil {
			return SyncDoneMsg{err: fmt.Errorf("pulled but restore failed: %w", err)}
		}

		return SyncDoneMsg{
			cloned:   pullResult.Cloned,
			restored: len(restoreResult.Restored),
			backups:  len(restoreResult.Backups),
			skipped:  len(restoreResult.Skipped),
		}
	}
}
