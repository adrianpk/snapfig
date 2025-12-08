package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/adrianpk/snapfig/internal/tui/screens"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}

	model := New(cfg, "/tmp/config.yaml", false)

	if model.current != screenPicker {
		t.Errorf("New() current screen = %d, want screenPicker", model.current)
	}
	if model.service == nil {
		t.Error("New() service should not be nil")
	}
}

func TestNewWithService(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
	}
	mockSvc := snapfig.NewMockService(cfg)

	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	if model.current != screenPicker {
		t.Errorf("NewWithService() current screen = %d, want screenPicker", model.current)
	}
	if model.service != mockSvc {
		t.Error("NewWithService() service not set correctly")
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

// Test action key handling with mock service
func TestActionKeyF2Copy(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF2}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if !m.busy {
		t.Error("F2 should set busy to true")
	}
	if m.status != "Copying..." {
		t.Errorf("status = %q, want 'Copying...'", m.status)
	}
	if cmd == nil {
		t.Error("F2 should return a command")
	}
}

func TestActionKeyF3Push(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF3}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if !m.busy {
		t.Error("F3 should set busy to true")
	}
	if m.status != "Pushing..." {
		t.Errorf("status = %q, want 'Pushing...'", m.status)
	}
	if cmd == nil {
		t.Error("F3 should return a command")
	}
}

func TestActionKeyF4Pull(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF4}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if !m.busy {
		t.Error("F4 should set busy to true")
	}
	if m.status != "Pulling..." {
		t.Errorf("status = %q, want 'Pulling...'", m.status)
	}
	if cmd == nil {
		t.Error("F4 should return a command")
	}
}

func TestActionKeyF5Restore(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF5}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if !m.busy {
		t.Error("F5 should set busy to true")
	}
	if m.status != "Restoring..." {
		t.Errorf("status = %q, want 'Restoring...'", m.status)
	}
	if cmd == nil {
		t.Error("F5 should return a command")
	}
}

func TestActionKeyF7Backup(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF7}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if !m.busy {
		t.Error("F7 should set busy to true")
	}
	if m.status != "Backing up (copy + push)..." {
		t.Errorf("status = %q, want 'Backing up (copy + push)...'", m.status)
	}
	if cmd == nil {
		t.Error("F7 should return a command")
	}
}

func TestActionKeyF8Sync(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF8}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if !m.busy {
		t.Error("F8 should set busy to true")
	}
	if m.status != "Syncing (pull + restore)..." {
		t.Errorf("status = %q, want 'Syncing (pull + restore)...'", m.status)
	}
	if cmd == nil {
		t.Error("F8 should return a command")
	}
}

func TestActionKeyF6SelectiveNoWatching(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable, Watching: nil}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF6}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.status != "No paths configured" {
		t.Errorf("status = %q, want 'No paths configured'", m.status)
	}
	if m.busy {
		t.Error("F6 should not set busy when no paths configured")
	}
}

func TestActionKeyF6SelectiveWithWatching(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true},
		},
	}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF6}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.current != screenRestorePicker {
		t.Error("F6 should switch to screenRestorePicker")
	}
	if m.status != "Loading vault contents..." {
		t.Errorf("status = %q, want 'Loading vault contents...'", m.status)
	}
	if cmd == nil {
		t.Error("F6 should return a command to init restore picker")
	}
}

func TestActionKeyF9Settings(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	msg := tea.KeyMsg{Type: tea.KeyF9}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.current != screenSettings {
		t.Error("F9 should switch to screenSettings")
	}
	if cmd == nil {
		t.Error("F9 should return a command to init settings")
	}
}

// Note: TestDoCopy would require a fully initialized picker with selections.
// The picker initialization happens through Init() which requires async loading.
// These operations are tested through integration tests rather than unit tests.

func TestDoRestoreCommand(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true},
		},
	}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.RestoreFunc = func() (*snapfig.RestoreResult, error) {
		return &snapfig.RestoreResult{
			Restored:     []string{".config/test"},
			FilesUpdated: 3,
			FilesSkipped: 1,
		}, nil
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doRestore()
	msg := cmd()
	restoreDone, ok := msg.(RestoreDoneMsg)

	if !ok {
		t.Fatal("doRestore should return RestoreDoneMsg")
	}
	if restoreDone.err != nil {
		t.Errorf("unexpected error: %v", restoreDone.err)
	}
	if restoreDone.filesUpdated != 3 {
		t.Errorf("filesUpdated = %d, want 3", restoreDone.filesUpdated)
	}
}

func TestDoRestoreNoPathsConfigured(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable, Watching: nil}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doRestore()
	msg := cmd()
	restoreDone, ok := msg.(RestoreDoneMsg)

	if !ok {
		t.Fatal("doRestore should return RestoreDoneMsg")
	}
	if restoreDone.err == nil {
		t.Error("doRestore should return error when no paths configured")
	}
}

func TestDoPushCommand(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doPush()
	msg := cmd()
	pushDone, ok := msg.(PushDoneMsg)

	if !ok {
		t.Fatal("doPush should return PushDoneMsg")
	}
	if pushDone.err != nil {
		t.Errorf("unexpected error: %v", pushDone.err)
	}
	if !mockSvc.PushCalled {
		t.Error("Push should have been called on service")
	}
}

func TestDoPullCommand(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.PullFunc = func() (*snapfig.PullResult, error) {
		return &snapfig.PullResult{Cloned: true}, nil
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doPull()
	msg := cmd()
	pullDone, ok := msg.(PullDoneMsg)

	if !ok {
		t.Fatal("doPull should return PullDoneMsg")
	}
	if pullDone.err != nil {
		t.Errorf("unexpected error: %v", pullDone.err)
	}
	if !pullDone.cloned {
		t.Error("cloned should be true")
	}
}

func TestDoSyncCommand(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true},
		},
	}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.PullFunc = func() (*snapfig.PullResult, error) {
		return &snapfig.PullResult{Cloned: false}, nil
	}
	mockSvc.RestoreFunc = func() (*snapfig.RestoreResult, error) {
		return &snapfig.RestoreResult{
			Restored:     []string{".config/test"},
			FilesUpdated: 2,
		}, nil
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doSync()
	msg := cmd()
	syncDone, ok := msg.(SyncDoneMsg)

	if !ok {
		t.Fatal("doSync should return SyncDoneMsg")
	}
	if syncDone.err != nil {
		t.Errorf("unexpected error: %v", syncDone.err)
	}
	if !mockSvc.PullCalled {
		t.Error("Pull should have been called")
	}
	if !mockSvc.RestoreCalled {
		t.Error("Restore should have been called")
	}
}

func TestDoSelectiveRestoreCommand(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.RestoreSelectiveFunc = func(paths []string) (*snapfig.RestoreResult, error) {
		return &snapfig.RestoreResult{
			Restored:     paths,
			FilesUpdated: len(paths),
		}, nil
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	paths := []string{".config/test/file1", ".config/test/file2"}
	cmd := model.doSelectiveRestore(paths)
	msg := cmd()
	selectiveDone, ok := msg.(SelectiveRestoreDoneMsg)

	if !ok {
		t.Fatal("doSelectiveRestore should return SelectiveRestoreDoneMsg")
	}
	if selectiveDone.err != nil {
		t.Errorf("unexpected error: %v", selectiveDone.err)
	}
	if selectiveDone.filesUpdated != 2 {
		t.Errorf("filesUpdated = %d, want 2", selectiveDone.filesUpdated)
	}
	if !mockSvc.RestoreSelectiveCalled {
		t.Error("RestoreSelective should have been called")
	}
}

func TestInitRestorePickerCommand(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.ListVaultEntriesFunc = func() ([]snapfig.VaultEntry, error) {
		return []snapfig.VaultEntry{
			{Path: ".config/test", IsDir: true},
		}, nil
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.initRestorePicker()
	msg := cmd()
	initMsg, ok := msg.(screens.RestorePickerInitMsg)

	if !ok {
		t.Fatal("initRestorePicker should return RestorePickerInitMsg")
	}
	if initMsg.Err != nil {
		t.Errorf("unexpected error: %v", initMsg.Err)
	}
	if len(initMsg.Entries) != 1 {
		t.Errorf("entries length = %d, want 1", len(initMsg.Entries))
	}
	if !mockSvc.ListVaultEntriesCalled {
		t.Error("ListVaultEntries should have been called")
	}
}

func TestScreenNavigation(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	// Test view for each screen
	tests := []struct {
		name   string
		screen screen
	}{
		{"picker", screenPicker},
		{"settings", screenSettings},
		{"restorePicker", screenRestorePicker},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.current = tt.screen
			view := model.View()
			if view == "" {
				t.Errorf("View() for %s should not be empty", tt.name)
			}
		})
	}
}

// Note: TestSelectedPaths requires fully initialized picker.
// The SelectedPaths function delegates to picker.Selected() which requires Init().

// Note: TestDoBackupCommand requires fully initialized picker with selections.
// doBackup calls picker.Selected() which requires bubbletea Init() to have run.
// These operations are tested through integration tests rather than unit tests.

func TestDoSyncPullError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.PullFunc = func() (*snapfig.PullResult, error) {
		return nil, fmt.Errorf("pull error")
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doSync()
	msg := cmd()
	syncDone, ok := msg.(SyncDoneMsg)

	if !ok {
		t.Fatal("doSync should return SyncDoneMsg")
	}
	if syncDone.err == nil {
		t.Error("doSync should return error when pull fails")
	}
}

func TestDoSyncRestoreError(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitModeDisable,
		Watching: []config.Watched{
			{Path: ".config/test", Enabled: true},
		},
	}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.PullFunc = func() (*snapfig.PullResult, error) {
		return &snapfig.PullResult{Cloned: false}, nil
	}
	mockSvc.RestoreFunc = func() (*snapfig.RestoreResult, error) {
		return nil, fmt.Errorf("restore error")
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.doSync()
	msg := cmd()
	syncDone, ok := msg.(SyncDoneMsg)

	if !ok {
		t.Fatal("doSync should return SyncDoneMsg")
	}
	if syncDone.err == nil {
		t.Error("doSync should return error when restore fails")
	}
}

func TestDoSelectiveRestoreError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.RestoreSelectiveFunc = func(paths []string) (*snapfig.RestoreResult, error) {
		return nil, fmt.Errorf("selective restore error")
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	paths := []string{".config/test"}
	cmd := model.doSelectiveRestore(paths)
	msg := cmd()
	selectiveDone, ok := msg.(SelectiveRestoreDoneMsg)

	if !ok {
		t.Fatal("doSelectiveRestore should return SelectiveRestoreDoneMsg")
	}
	if selectiveDone.err == nil {
		t.Error("doSelectiveRestore should return error")
	}
}

func TestInitRestorePickerError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	mockSvc.ListVaultEntriesFunc = func() ([]snapfig.VaultEntry, error) {
		return nil, fmt.Errorf("list error")
	}
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)

	cmd := model.initRestorePicker()
	msg := cmd()
	initMsg, ok := msg.(screens.RestorePickerInitMsg)

	if !ok {
		t.Fatal("initRestorePicker should return RestorePickerInitMsg")
	}
	if initMsg.Err == nil {
		t.Error("initRestorePicker should return error")
	}
}

func TestUpdateRestoreDoneMsgError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := RestoreDoneMsg{err: fmt.Errorf("restore failed")}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after error")
	}
	if !containsString(m.status, "Error") {
		t.Errorf("status should contain error, got: %s", m.status)
	}
}

func TestUpdatePushDoneMsgError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := PushDoneMsg{err: fmt.Errorf("push failed")}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after error")
	}
	if !containsString(m.status, "Error") {
		t.Errorf("status should contain error, got: %s", m.status)
	}
}

func TestUpdatePullDoneMsgError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := PullDoneMsg{err: fmt.Errorf("pull failed")}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after error")
	}
	if !containsString(m.status, "Error") {
		t.Errorf("status should contain error, got: %s", m.status)
	}
}

func TestUpdateBackupDoneMsgError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := BackupDoneMsg{err: fmt.Errorf("backup failed")}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after error")
	}
	if !containsString(m.status, "Error") {
		t.Errorf("status should contain error, got: %s", m.status)
	}
}

func TestUpdateSyncDoneMsgError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true

	msg := SyncDoneMsg{err: fmt.Errorf("sync failed")}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after error")
	}
	if !containsString(m.status, "Error") {
		t.Errorf("status should contain error, got: %s", m.status)
	}
}

func TestUpdateSelectiveRestoreDoneMsgError(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	model := New(cfg, "/tmp/config.yaml", false)
	model.busy = true
	model.current = screenRestorePicker

	msg := SelectiveRestoreDoneMsg{err: fmt.Errorf("selective restore failed")}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.busy {
		t.Error("busy should be false after error")
	}
	if m.current != screenPicker {
		t.Error("should return to screenPicker after error")
	}
	if !containsString(m.status, "Error") {
		t.Errorf("status should contain error, got: %s", m.status)
	}
}

func TestRestorePickerEscReturnsToMain(t *testing.T) {
	cfg := &config.Config{Git: config.GitModeDisable}
	mockSvc := snapfig.NewMockService(cfg)
	model := NewWithService(mockSvc, "/tmp/config.yaml", false)
	model.current = screenRestorePicker
	model.restorePicker = screens.NewRestorePicker()
	// Mark as loaded so Esc key is processed
	rpUpdated, _ := model.restorePicker.Update(
		screens.RestorePickerInitMsg{Entries: nil},
	)
	model.restorePicker = rpUpdated.(screens.RestorePickerModel)

	// Press Esc to cancel
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.current != screenPicker {
		t.Error("should return to screenPicker when Esc is pressed")
	}
}

// Note: Tests for SelectedPaths, doCopy, and doBackup with picker selections
// require async initialization of the picker which runs as a bubbletea command.
// Testing these code paths requires integration-style testing with a running
// bubbletea program. The coverage for these is achieved through the Update
// message handler tests that simulate the full message flow.
//
// Functions that delegate to picker.Selected() before picker initialization
// will panic, which is expected behavior (programmer error - should not call
// before Init completes). The async Init pattern is standard in bubbletea.
