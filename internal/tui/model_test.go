package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adrianpk/snapfig/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)

	if model.current != screenPicker {
		t.Errorf("New() current screen = %d, want screenPicker", model.current)
	}
	if model.cfg != cfg {
		t.Error("New() cfg not set correctly")
	}
}

func TestNewDemoMode(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", true)

	if !model.demoMode {
		t.Error("New() demoMode should be true")
	}
}

func TestInit(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	cmd := model.Init()

	// Init should return a command
	if cmd == nil {
		t.Error("Init() returned nil command")
	}
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.width != 80 {
		t.Errorf("Update() width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("Update() height = %d, want 24", m.height)
	}
}

func TestUpdateCopyDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	tests := []struct {
		name       string
		msg        CopyDoneMsg
		wantBusy   bool
		wantStatus string
	}{
		{
			name:       "success",
			msg:        CopyDoneMsg{filesUpdated: 5, filesSkipped: 2, filesRemoved: 1},
			wantBusy:   false,
			wantStatus: "Copied: 5 updated, 2 unchanged, 1 removed",
		},
		{
			name:       "error",
			msg:        CopyDoneMsg{err: errNoPathsSelected},
			wantBusy:   false,
			wantStatus: "Error: no paths selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.busy = true
			updated, _ := model.Update(tt.msg)
			m := updated.(Model)

			if m.busy != tt.wantBusy {
				t.Errorf("busy = %v, want %v", m.busy, tt.wantBusy)
			}
			if m.status != tt.wantStatus {
				t.Errorf("status = %q, want %q", m.status, tt.wantStatus)
			}
		})
	}
}

func TestUpdateRestoreDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := RestoreDoneMsg{filesUpdated: 3, filesSkipped: 1}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after RestoreDoneMsg")
	}
	expected := "Restored: 3 updated, 1 unchanged"
	if m.status != expected {
		t.Errorf("status = %q, want %q", m.status, expected)
	}
}

func TestUpdatePushDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := PushDoneMsg{}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after PushDoneMsg")
	}
	if m.status != "Pushed to remote" {
		t.Errorf("status = %q, want 'Pushed to remote'", m.status)
	}
}

func TestUpdatePullDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	tests := []struct {
		name       string
		msg        PullDoneMsg
		wantStatus string
	}{
		{
			name:       "pulled",
			msg:        PullDoneMsg{cloned: false},
			wantStatus: "Pulled from remote",
		},
		{
			name:       "cloned",
			msg:        PullDoneMsg{cloned: true},
			wantStatus: "Cloned from remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.busy = true
			updated, _ := model.Update(tt.msg)
			m := updated.(Model)

			if m.status != tt.wantStatus {
				t.Errorf("status = %q, want %q", m.status, tt.wantStatus)
			}
		})
	}
}

func TestUpdateBackupDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := BackupDoneMsg{filesUpdated: 2, filesSkipped: 1, filesRemoved: 0}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after BackupDoneMsg")
	}
	expected := "Backup: 2 updated, 1 unchanged, 0 removed, pushed"
	if m.status != expected {
		t.Errorf("status = %q, want %q", m.status, expected)
	}
}

func TestUpdateSyncDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	tests := []struct {
		name       string
		msg        SyncDoneMsg
		wantStatus string
	}{
		{
			name:       "pulled and restored",
			msg:        SyncDoneMsg{cloned: false, filesUpdated: 3, filesSkipped: 2},
			wantStatus: "Sync: pulled, 3 updated, 2 unchanged",
		},
		{
			name:       "cloned and restored",
			msg:        SyncDoneMsg{cloned: true, filesUpdated: 5, filesSkipped: 0},
			wantStatus: "Sync: cloned, 5 updated, 0 unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.busy = true
			updated, _ := model.Update(tt.msg)
			m := updated.(Model)

			if m.status != tt.wantStatus {
				t.Errorf("status = %q, want %q", m.status, tt.wantStatus)
			}
		})
	}
}

func TestUpdateSelectiveRestoreDoneMsg(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true
	model.current = screenRestorePicker

	msg := SelectiveRestoreDoneMsg{filesUpdated: 2, filesSkipped: 0}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false")
	}
	if m.current != screenPicker {
		t.Error("should return to screenPicker after selective restore")
	}
	expected := "Restored: 2 updated, 0 unchanged"
	if m.status != expected {
		t.Errorf("status = %q, want %q", m.status, expected)
	}
}

func TestUpdateQuitKeys(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)

	tests := []string{"ctrl+c", "f10"}

	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			if key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			} else if key == "f10" {
				msg = tea.KeyMsg{Type: tea.KeyF10}
			}

			_, cmd := model.Update(msg)
			// Quit command should be returned
			if cmd == nil {
				t.Errorf("Update(%s) should return quit command", key)
			}
		})
	}
}

func TestView(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.width = 80
	model.height = 24

	view := model.View()

	if len(view) == 0 {
		t.Error("View() returned empty string")
	}

	// Should contain action bar elements
	if !containsAll(view, []string{"F2", "F3", "F4", "F5", "F10"}) {
		t.Error("View() missing action bar elements")
	}
}

func TestViewWithStatus(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.width = 80
	model.height = 24
	model.status = "Test status"

	view := model.View()

	if !containsAll(view, []string{"Test status"}) {
		t.Error("View() should contain status")
	}
}

func TestViewWithError(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.width = 80
	model.height = 24
	model.status = "Error: something went wrong"

	view := model.View()

	if !containsAll(view, []string{"Error:"}) {
		t.Error("View() should contain error status")
	}
}

func TestRenderActionBar(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)

	actionBar := model.renderActionBar()

	expected := []string{"F2", "Copy", "F3", "Push", "F4", "Pull", "F5", "Restore", "F10", "Quit"}
	if !containsAll(actionBar, expected) {
		t.Errorf("renderActionBar() missing expected elements, got: %s", actionBar)
	}
}

func TestBusyPreventsActions(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	// When busy, F2 should not trigger copy
	msg := tea.KeyMsg{Type: tea.KeyF2}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("F2 should not trigger command when busy")
	}
}

// Helper to check if string contains all substrings
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !containsString(s, sub) {
			return false
		}
	}
	return true
}

func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStringHelper(s, sub))
}

func containsStringHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Error variable for testing
var errNoPathsSelected = error(nil)

func init() {
	errNoPathsSelected = fmt.Errorf("no paths selected")
}
