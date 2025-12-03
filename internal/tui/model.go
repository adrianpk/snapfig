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
	screenRestorePicker
)

// Model is the root TUI model that manages screen navigation.
type Model struct {
	current       screen
	picker        screens.PickerModel
	settings      screens.SettingsModel
	restorePicker screens.RestorePickerModel
	cfg           *config.Config
	configPath    string
	width         int
	height        int
	status        string
	busy          bool
	demoMode      bool
}

// CopyDoneMsg is sent when copy operation completes.
type CopyDoneMsg struct {
	err          error
	copied       int    // paths processed
	skipped      int    // paths skipped (not found)
	filesUpdated int    // files actually copied
	filesSkipped int    // files unchanged
	filesRemoved int    // stale files removed
}

// RestoreDoneMsg is sent when restore operation completes.
type RestoreDoneMsg struct {
	err          error
	restored     int
	backups      int
	skipped      int
	filesUpdated int
	filesSkipped int
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
	err          error
	copied       int
	skipped      int
	filesUpdated int
	filesSkipped int
	filesRemoved int
}

// SyncDoneMsg is sent when sync (pull+restore) completes.
type SyncDoneMsg struct {
	err          error
	cloned       bool
	restored     int
	backups      int
	skipped      int
	filesUpdated int
	filesSkipped int
}

// SelectiveRestoreDoneMsg is sent when selective restore completes.
type SelectiveRestoreDoneMsg struct {
	err          error
	restored     int
	backups      int
	skipped      int
	filesUpdated int
	filesSkipped int
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
			m.status = fmt.Sprintf("Copied: %d updated, %d unchanged, %d removed",
				msg.filesUpdated, msg.filesSkipped, msg.filesRemoved)
		}
		return m, nil

	case RestoreDoneMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Restored: %d updated, %d unchanged",
				msg.filesUpdated, msg.filesSkipped)
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
			m.status = fmt.Sprintf("Backup: %d updated, %d unchanged, %d removed, pushed",
				msg.filesUpdated, msg.filesSkipped, msg.filesRemoved)
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
			m.status = fmt.Sprintf("Sync: %s, %d updated, %d unchanged",
				action, msg.filesUpdated, msg.filesSkipped)
		}
		return m, nil

	case SelectiveRestoreDoneMsg:
		m.busy = false
		m.current = screenPicker
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Restored: %d updated, %d unchanged",
				msg.filesUpdated, msg.filesSkipped)
		}
		return m, nil

	case screens.RestorePickerInitMsg:
		// Pass to restore picker
		updated, cmd := m.restorePicker.Update(msg)
		m.restorePicker = updated.(screens.RestorePickerModel)
		return m, cmd

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
		case "f6":
			if !m.busy && m.current == screenPicker {
				if len(m.cfg.Watching) == 0 {
					m.status = "No paths configured"
					return m, nil
				}
				m.restorePicker = screens.NewRestorePicker()
				m.current = screenRestorePicker
				m.status = "Loading vault contents..."
				return m, m.initRestorePicker()
			}
			return m, nil
		case "f9":
			if !m.busy && m.current == screenPicker {
				m.settings = screens.NewSettings(m.cfg.Remote, m.cfg.Daemon)
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

	case screenRestorePicker:
		updated, cmd := m.restorePicker.Update(msg)
		m.restorePicker = updated.(screens.RestorePickerModel)

		// Check for Esc
		if m.restorePicker.WasCanceled() {
			m.current = screenPicker
			m.status = ""
			return m, nil
		}

		// Check for Enter (confirm restore)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && m.restorePicker.Loaded() {
			selected := m.restorePicker.Selected()
			if len(selected) == 0 {
				m.status = "No files selected"
				return m, nil
			}
			m.busy = true
			m.status = "Restoring selected files..."
			return m, m.doSelectiveRestore(selected)
		}

		return m, cmd

	case screenSettings:
		updated, cmd := m.settings.Update(msg)
		m.settings = updated.(screens.SettingsModel)

		// Check if user pressed Enter or Esc
		if m.settings.WasSaved() {
			m.cfg.Remote = m.settings.Remote()
			m.cfg.Daemon = m.settings.DaemonConfig()
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
	case screenRestorePicker:
		b.WriteString(m.restorePicker.View())
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
		{"F6", "Selective"},
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

		return CopyDoneMsg{
			copied:       len(result.Copied),
			skipped:      len(result.Skipped),
			filesUpdated: result.FilesUpdated,
			filesSkipped: result.FilesSkipped,
			filesRemoved: result.FilesRemoved,
		}
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
			restored:     len(result.Restored),
			backups:      len(result.Backups),
			skipped:      len(result.Skipped),
			filesUpdated: result.FilesUpdated,
			filesSkipped: result.FilesSkipped,
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

		return BackupDoneMsg{
			copied:       len(result.Copied),
			skipped:      len(result.Skipped),
			filesUpdated: result.FilesUpdated,
			filesSkipped: result.FilesSkipped,
			filesRemoved: result.FilesRemoved,
		}
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
			cloned:       pullResult.Cloned,
			restored:     len(restoreResult.Restored),
			backups:      len(restoreResult.Backups),
			skipped:      len(restoreResult.Skipped),
			filesUpdated: restoreResult.FilesUpdated,
			filesSkipped: restoreResult.FilesSkipped,
		}
	}
}

// initRestorePicker initializes the restore picker with vault entries.
func (m *Model) initRestorePicker() tea.Cmd {
	return func() tea.Msg {
		restorer, err := snapfig.NewRestorer(m.cfg)
		if err != nil {
			return screens.RestorePickerInitMsg{Err: err}
		}

		entries, err := restorer.ListVaultEntries()
		return screens.RestorePickerInitMsg{Entries: entries, Err: err}
	}
}

// doSelectiveRestore restores only the selected paths.
func (m *Model) doSelectiveRestore(paths []string) tea.Cmd {
	return func() tea.Msg {
		restorer, err := snapfig.NewRestorer(m.cfg)
		if err != nil {
			return SelectiveRestoreDoneMsg{err: err}
		}

		result, err := restorer.RestoreSelective(paths)
		if err != nil {
			return SelectiveRestoreDoneMsg{err: err}
		}

		return SelectiveRestoreDoneMsg{
			restored:     len(result.Restored),
			backups:      len(result.Backups),
			skipped:      len(result.Skipped),
			filesUpdated: result.FilesUpdated,
			filesSkipped: result.FilesSkipped,
		}
	}
}
