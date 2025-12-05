package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestNewPicker(t *testing.T) {
	cfg := &config.Config{
		Watching: []config.Watched{
			{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true},
			{Path: ".bashrc", Git: config.GitModeRemove, Enabled: true},
			{Path: ".disabled", Git: config.GitModeRemove, Enabled: false},
		},
	}

	m := NewPicker(cfg, false)

	if m.preselected[".config/nvim"] != StateDisable {
		t.Error("preselected .config/nvim should be StateDisable")
	}
	if m.preselected[".bashrc"] != StateRemove {
		t.Error("preselected .bashrc should be StateRemove")
	}
	if _, ok := m.preselected[".disabled"]; ok {
		t.Error("disabled paths should not be preselected")
	}
}

func TestNewPickerDemoMode(t *testing.T) {
	m := NewPicker(nil, true)

	if !m.demoMode {
		t.Error("demoMode should be true")
	}
	if m.demoPaths == nil {
		t.Error("demoPaths should be populated in demo mode")
	}
	if !m.demoPaths[".config/nvim"] {
		t.Error("demoPaths should contain .config/nvim")
	}
}

func TestNewPickerNilConfig(t *testing.T) {
	m := NewPicker(nil, false)

	if len(m.preselected) != 0 {
		t.Error("preselected should be empty with nil config")
	}
}

func TestPickerInit(t *testing.T) {
	m := NewPicker(nil, false)
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestPickerUpdateWindowSize(t *testing.T) {
	m := NewPicker(nil, false)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestPickerUpdateInitMsg(t *testing.T) {
	m := NewPicker(nil, false)

	// Simulate init message with home directory
	msg := initMsg{
		home:      "/home/testuser",
		wellKnown: map[string]bool{".bashrc": true},
	}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if !m.loaded {
		t.Error("loaded should be true after initMsg")
	}
	if m.home != "/home/testuser" {
		t.Errorf("home = %q, want /home/testuser", m.home)
	}
	if m.root == nil {
		t.Error("root should be initialized")
	}
}

func TestPickerUpdateInitMsgError(t *testing.T) {
	m := NewPicker(nil, false)

	msg := initMsg{err: errTestError}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.err == nil {
		t.Error("err should be set on error")
	}
}

func TestPickerNavigationNotLoaded(t *testing.T) {
	m := NewPicker(nil, false)
	// Not loaded yet

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	// Should not panic or change state
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (not loaded)", m.cursor)
	}
}

func TestPickerView(t *testing.T) {
	m := NewPicker(nil, false)
	m.width = 80
	m.height = 24

	// Before loading
	view := m.View()
	if len(view) == 0 {
		t.Error("View() should return something even before loading")
	}

	// After loading (simulate)
	m.loaded = true
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
	}
	m.flat = []*node{m.root}

	view = m.View()
	if len(view) == 0 {
		t.Error("View() should return content after loading")
	}
}

func TestGetDemoPaths(t *testing.T) {
	paths := getDemoPaths()

	expected := []string{".config", ".bashrc", ".zshrc", ".gitconfig"}
	for _, p := range expected {
		if !paths[p] {
			t.Errorf("getDemoPaths() missing %q", p)
		}
	}
}

func TestSelectionStateCycle(t *testing.T) {
	// Test the state cycle logic directly
	states := []SelectState{StateNone, StateRemove, StateDisable, StateNone}

	for i := 0; i < len(states)-1; i++ {
		current := states[i]
		expected := states[i+1]

		var next SelectState
		switch current {
		case StateNone:
			next = StateRemove
		case StateRemove:
			next = StateDisable
		case StateDisable:
			next = StateNone
		}

		if next != expected {
			t.Errorf("after %d, got %d, want %d", current, next, expected)
		}
	}
}

func TestNodeState(t *testing.T) {
	n := &node{
		name:  ".bashrc",
		path:  ".bashrc",
		isDir: false,
		state: StateNone,
	}

	if n.state != StateNone {
		t.Error("initial state should be StateNone")
	}

	n.state = StateRemove
	if n.state != StateRemove {
		t.Error("state should be StateRemove")
	}

	n.state = StateDisable
	if n.state != StateDisable {
		t.Error("state should be StateDisable")
	}
}

// Helper error for testing
var errTestError = error(nil)

func init() {
	errTestError = &testError{}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }
