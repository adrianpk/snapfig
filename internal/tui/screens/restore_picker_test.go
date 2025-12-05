package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/adrianpk/snapfig/internal/snapfig"
)

func TestNewRestorePicker(t *testing.T) {
	m := NewRestorePicker()

	if m.loaded {
		t.Error("loaded should be false initially")
	}
	if m.done {
		t.Error("done should be false initially")
	}
	if m.canceled {
		t.Error("canceled should be false initially")
	}
}

func TestRestorePickerInit(t *testing.T) {
	m := NewRestorePicker()
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestRestorePickerUpdateWindowSize(t *testing.T) {
	m := NewRestorePicker()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestRestorePickerUpdateInitMsgError(t *testing.T) {
	m := NewRestorePicker()

	msg := RestorePickerInitMsg{Err: &testError{}}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if m.err == nil {
		t.Error("err should be set")
	}
	if !m.loaded {
		t.Error("loaded should be true even on error")
	}
}

func TestRestorePickerUpdateInitMsgSuccess(t *testing.T) {
	m := NewRestorePicker()

	entries := []snapfig.VaultEntry{
		{Path: ".bashrc", IsDir: false},
		{Path: ".config/nvim", IsDir: true, Children: []string{".config/nvim/init.lua"}},
	}
	msg := RestorePickerInitMsg{Entries: entries}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if !m.loaded {
		t.Error("loaded should be true")
	}
	if m.root == nil {
		t.Error("root should be set")
	}
}

func TestRestorePickerNavigationNotLoaded(t *testing.T) {
	m := NewRestorePicker()

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	// Should not panic
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestRestorePickerEscCancels(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if !m.canceled {
		t.Error("canceled should be true after Esc")
	}
}

func TestRestorePickerWasCanceled(t *testing.T) {
	m := NewRestorePicker()

	if m.WasCanceled() {
		t.Error("WasCanceled() should be false initially")
	}

	m.canceled = true
	if !m.WasCanceled() {
		t.Error("WasCanceled() should be true")
	}
}

func TestRestorePickerLoaded(t *testing.T) {
	m := NewRestorePicker()

	if m.Loaded() {
		t.Error("Loaded() should be false initially")
	}

	m.loaded = true
	if !m.Loaded() {
		t.Error("Loaded() should be true")
	}
}

func TestRestorePickerHasError(t *testing.T) {
	m := NewRestorePicker()

	if m.HasError() != nil {
		t.Error("HasError() should be nil initially")
	}

	m.err = &testError{}
	if m.HasError() == nil {
		t.Error("HasError() should return error")
	}
}

func TestRestorePickerIsDone(t *testing.T) {
	m := NewRestorePicker()

	if m.IsDone() {
		t.Error("IsDone() should be false initially")
	}

	m.done = true
	if !m.IsDone() {
		t.Error("IsDone() should be true")
	}
}

func TestRestorePickerMarkDone(t *testing.T) {
	m := NewRestorePicker()

	m.MarkDone()
	if !m.done {
		t.Error("done should be true after MarkDone()")
	}
}

func TestRestorePickerView(t *testing.T) {
	m := NewRestorePicker()
	m.width = 80
	m.height = 24

	// Before loading
	view := m.View()
	if len(view) == 0 {
		t.Error("View() should return something")
	}

	// With error
	m.loaded = true
	m.err = &testError{}
	view = m.View()
	if !containsString(view, "Error") {
		t.Error("View() should show error")
	}
}

func TestRestorePickerSelected(t *testing.T) {
	m := NewRestorePicker()

	// With root but no selection
	m.root = &RestoreNode{Path: "root"}
	child := &RestoreNode{Path: ".bashrc", Selected: false, Parent: m.root}
	m.root.Children = []*RestoreNode{child}

	selected := m.Selected()
	if len(selected) != 0 {
		t.Errorf("Selected() = %d items, want 0", len(selected))
	}

	// With selection
	child.Selected = true

	selected = m.Selected()
	if len(selected) != 1 {
		t.Errorf("Selected() = %d items, want 1", len(selected))
	}
}

func TestRestorePickerNavigationDown(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	child1 := &RestoreNode{Path: ".bashrc", Parent: m.root}
	child2 := &RestoreNode{Path: ".zshrc", Parent: m.root}
	m.root.Children = []*RestoreNode{child1, child2}
	m.flat = []*RestoreNode{child1, child2}
	m.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
}

func TestRestorePickerNavigationUp(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	child1 := &RestoreNode{Path: ".bashrc", Parent: m.root}
	child2 := &RestoreNode{Path: ".zshrc", Parent: m.root}
	m.root.Children = []*RestoreNode{child1, child2}
	m.flat = []*RestoreNode{child1, child2}
	m.cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestRestorePickerSpaceToggles(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	child := &RestoreNode{Path: ".bashrc", Parent: m.root, Selected: false}
	m.root.Children = []*RestoreNode{child}
	m.flat = []*RestoreNode{child} // root is not in flat, only children
	m.cursor = 0                   // cursor at child (index 0)

	if child.Selected {
		t.Error("child should not be selected initially")
	}

	msg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	// Check the node directly since rebuildFlat recreates the flat slice
	if !child.Selected {
		t.Error("child should be selected after space")
	}
}
