package screens

import (
	"os"
	"path/filepath"
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

func TestIsDemoPath(t *testing.T) {
	tests := []struct {
		path      string
		demoPaths map[string]bool
		want      bool
	}{
		{
			path:      ".config",
			demoPaths: map[string]bool{".config": true, ".bashrc": true},
			want:      true,
		},
		{
			path:      ".config/nvim",
			demoPaths: map[string]bool{".config/nvim": true},
			want:      true,
		},
		{
			path:      ".bashrc",
			demoPaths: map[string]bool{".config": true},
			want:      false,
		},
		{
			path:      ".bashrc",
			demoPaths: map[string]bool{".bashrc": true},
			want:      true,
		},
		{
			path:      ".config",
			demoPaths: nil,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			m := &PickerModel{demoPaths: tt.demoPaths}
			got := m.isDemoPath(tt.path)
			if got != tt.want {
				t.Errorf("isDemoPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestPickerSelected(t *testing.T) {
	m := NewPicker(nil, false)
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		children: []*node{
			{
				name:  ".config",
				path:  ".config",
				isDir: true,
				state: StateRemove,
				children: []*node{
					{name: "nvim", path: ".config/nvim", isDir: true, state: StateRemove},
				},
			},
			{name: ".bashrc", path: ".bashrc", isDir: false, state: StateDisable},
			{name: ".zshrc", path: ".zshrc", isDir: false, state: StateNone},
		},
	}
	m.loaded = true

	selected := m.Selected()
	if len(selected) != 2 {
		t.Errorf("len(selected) = %d, want 2", len(selected))
	}

	// Check .config has the right state (StateRemove)
	found := false
	for _, s := range selected {
		if s.Path == ".config" && s.GitMode == StateRemove {
			found = true
			break
		}
	}
	if !found {
		t.Error(".config should be in selected with StateRemove")
	}
}

func TestPickerPropagateState(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create a tree structure
	child1 := &node{name: "nvim", path: ".config/nvim", isDir: true, state: StateNone}
	child2 := &node{name: "zsh", path: ".config/zsh", isDir: true, state: StateNone}
	parent := &node{
		name:     ".config",
		path:     ".config",
		isDir:    true,
		state:    StateRemove, // Set the state before propagating
		expanded: true,
		children: []*node{child1, child2},
	}
	child1.parent = parent
	child2.parent = parent

	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		children: []*node{parent},
	}
	parent.parent = m.root

	m.rebuildFlat()

	// Test propagate state to children (propagateState uses the node's current state)
	m.propagateState(parent)

	if child1.state != StateRemove {
		t.Errorf("child1.state = %d, want StateRemove", child1.state)
	}
	if child2.state != StateRemove {
		t.Errorf("child2.state = %d, want StateRemove", child2.state)
	}
}

func TestPickerClearState(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	child := &node{name: "nvim", path: ".config/nvim", isDir: true, state: StateRemove}
	parent := &node{
		name:     ".config",
		path:     ".config",
		isDir:    true,
		state:    StateRemove,
		expanded: true,
		children: []*node{child},
	}
	child.parent = parent

	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		children: []*node{parent},
	}
	parent.parent = m.root

	m.clearState(m.root)

	if parent.state != StateNone {
		t.Errorf("parent.state = %d, want StateNone after clearState", parent.state)
	}
	if child.state != StateNone {
		t.Errorf("child.state = %d, want StateNone after clearState", child.state)
	}
}

func TestPickerNavigationKeys(t *testing.T) {
	tests := []struct {
		name       string
		keyMsg     tea.KeyMsg
		wantCursor int
	}{
		{
			name:       "down key",
			keyMsg:     tea.KeyMsg{Type: tea.KeyDown},
			wantCursor: 1,
		},
		{
			name:       "j key",
			keyMsg:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantCursor: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPicker(nil, false)
			m.loaded = true
			m.width = 80
			m.height = 24

			// Create a simple tree
			m.root = &node{
				name:     "~",
				path:     "",
				isDir:    true,
				expanded: true,
				children: []*node{
					{name: ".config", path: ".config", isDir: true},
					{name: ".bashrc", path: ".bashrc", isDir: false},
				},
			}
			m.rebuildFlat()
			m.cursor = 0

			updated, _ := m.Update(tt.keyMsg)
			m = updated.(PickerModel)

			if m.cursor != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", m.cursor, tt.wantCursor)
			}
		})
	}
}

func TestPickerSpaceKeyToggle(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true
	m.width = 80
	m.height = 24

	child := &node{name: ".config", path: ".config", isDir: true, state: StateNone}
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		children: []*node{child},
	}
	child.parent = m.root
	m.rebuildFlat()
	m.cursor = 0

	// First space: None -> Remove
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.flat[0].state != StateRemove {
		t.Errorf("state = %d, want StateRemove after first space", m.flat[0].state)
	}

	// Second space: Remove -> Disable
	updated, _ = m.Update(msg)
	m = updated.(PickerModel)

	if m.flat[0].state != StateDisable {
		t.Errorf("state = %d, want StateDisable after second space", m.flat[0].state)
	}

	// Third space: Disable -> None
	updated, _ = m.Update(msg)
	m = updated.(PickerModel)

	if m.flat[0].state != StateNone {
		t.Errorf("state = %d, want StateNone after third space", m.flat[0].state)
	}
}

func TestPickerSelectAllKey(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true
	m.width = 80
	m.height = 24

	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		children: []*node{
			{name: ".config", path: ".config", isDir: true, state: StateNone},
			{name: ".bashrc", path: ".bashrc", isDir: false, state: StateNone},
		},
	}
	m.rebuildFlat()

	// Press 'a' to select all
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	for _, n := range m.flat {
		if n.state != StateRemove {
			t.Errorf("node %q state = %d, want StateRemove", n.name, n.state)
		}
	}
}

func TestPickerNoneKey(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true
	m.width = 80
	m.height = 24

	child := &node{name: ".config", path: ".config", isDir: true, state: StateRemove}
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		children: []*node{child},
	}
	child.parent = m.root
	m.rebuildFlat()

	// Press 'n' to clear all
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if child.state != StateNone {
		t.Errorf("child state = %d, want StateNone", child.state)
	}
}

func TestLoadChildrenWithTempDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	dirs := []string{
		".config/nvim",
		".config/emacs",
		"Documents",
		".hidden_dir",
	}
	files := []string{
		".bashrc",
		".zshrc",
		"readme.txt",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}

	m := NewPicker(nil, false)
	m.home = tmpDir
	m.wellKnown = map[string]bool{
		".config/nvim": true,
	}
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		depth:    0,
	}

	// Load children
	m.loadChildren(m.root)

	if !m.root.loaded {
		t.Error("root should be marked as loaded")
	}
	if len(m.root.children) == 0 {
		t.Error("root should have children")
	}

	// Check that .config is present
	var hasConfig bool
	for _, child := range m.root.children {
		if child.name == ".config" {
			hasConfig = true
			break
		}
	}
	if !hasConfig {
		t.Error(".config should be in children")
	}
}

func TestLoadChildrenDemoMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	dirs := []string{
		".config",
		".config/nvim",
		"Documents",
		"secret_dir",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	m := NewPicker(nil, true) // Demo mode
	m.home = tmpDir
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		depth:    0,
	}

	m.loadChildren(m.root)

	// In demo mode, only demo paths should be shown
	for _, child := range m.root.children {
		if child.name == "secret_dir" {
			t.Error("secret_dir should not be shown in demo mode")
		}
	}
}

func TestLoadChildrenWithPreselected(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .config/nvim
	nvimDir := filepath.Join(tmpDir, ".config", "nvim")
	if err := os.MkdirAll(nvimDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	cfg := &config.Config{
		Watching: []config.Watched{
			{Path: ".config/nvim", Git: config.GitModeDisable, Enabled: true},
		},
	}

	m := NewPicker(cfg, false)
	m.home = tmpDir
	m.wellKnown = map[string]bool{".config/nvim": true}
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		depth:    0,
	}

	m.loadChildren(m.root)

	// Find .config and load its children
	var configNode *node
	for _, child := range m.root.children {
		if child.name == ".config" {
			configNode = child
			break
		}
	}

	if configNode == nil {
		t.Fatal(".config should be in children")
	}

	m.loadChildren(configNode)

	// Find nvim
	var nvimNode *node
	for _, child := range configNode.children {
		if child.name == "nvim" {
			nvimNode = child
			break
		}
	}

	if nvimNode == nil {
		t.Fatal("nvim should be in .config children")
	}

	if nvimNode.state != StateDisable {
		t.Errorf("nvim state = %d, want StateDisable", nvimNode.state)
	}
}

func TestLoadChildrenReadDirError(t *testing.T) {
	m := NewPicker(nil, false)
	m.home = "/nonexistent/path/that/doesnt/exist"
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		depth:    0,
	}

	// Should not panic, just mark as loaded with no children
	m.loadChildren(m.root)

	if !m.root.loaded {
		t.Error("root should be marked as loaded even on error")
	}
}

func TestLoadChildrenAlreadyLoaded(t *testing.T) {
	m := NewPicker(nil, false)
	m.root = &node{
		name:     "~",
		path:     "",
		isDir:    true,
		expanded: true,
		loaded:   true, // Already loaded
	}

	// Should return immediately without doing anything
	m.loadChildren(m.root)

	if len(m.root.children) != 0 {
		t.Error("should not load children for already loaded node")
	}
}

func TestLoadChildrenNotDir(t *testing.T) {
	m := NewPicker(nil, false)
	fileNode := &node{
		name:  "file.txt",
		path:  "file.txt",
		isDir: false,
	}

	// Should return immediately without doing anything
	m.loadChildren(fileNode)

	if fileNode.loaded {
		t.Error("should not mark file as loaded")
	}
}

func TestViewportCalculations(t *testing.T) {
	tests := []struct {
		name        string
		height      int
		wantVisible int
	}{
		{name: "small window", height: 10, wantVisible: 15},  // height <= 10 returns 15
		{name: "medium window", height: 20, wantVisible: 12}, // height - 8
		{name: "large window", height: 40, wantVisible: 32},  // height - 8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPicker(nil, false)
			m.height = tt.height
			m.loaded = true

			maxVis := m.maxVisible()
			if maxVis != tt.wantVisible {
				t.Errorf("maxVisible() = %d, want %d", maxVis, tt.wantVisible)
			}
		})
	}
}

func TestPickerViewportStart(t *testing.T) {
	tests := []struct {
		name      string
		height    int
		flatLen   int
		cursor    int
		wantStart int
	}{
		{
			name:      "few items returns 0",
			height:    20,
			flatLen:   5, // Less than maxVisible (12)
			cursor:    0,
			wantStart: 0,
		},
		{
			name:      "cursor at start",
			height:    20,
			flatLen:   30,
			cursor:    0,
			wantStart: 0,
		},
		{
			name:      "cursor in middle centers viewport",
			height:    20,
			flatLen:   30,
			cursor:    15,
			wantStart: 9, // cursor - maxVisible/2 = 15 - 6 = 9
		},
		{
			name:      "cursor near end clamps to end",
			height:    20,
			flatLen:   30,
			cursor:    29,
			wantStart: 18, // flatLen - maxVisible = 30 - 12 = 18
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPicker(nil, false)
			m.height = tt.height
			m.loaded = true
			m.cursor = tt.cursor

			// Build flat list with nodes
			m.flat = make([]*node, tt.flatLen)
			for i := 0; i < tt.flatLen; i++ {
				m.flat[i] = &node{path: "item", depth: 1}
			}

			start := m.viewportStart()
			if start != tt.wantStart {
				t.Errorf("viewportStart() = %d, want %d", start, tt.wantStart)
			}
		})
	}
}

func TestPickerVisibleItems(t *testing.T) {
	m := NewPicker(nil, false)
	m.height = 20
	m.loaded = true

	// Build flat list with 5 items
	m.flat = make([]*node, 5)
	for i := 0; i < 5; i++ {
		m.flat[i] = &node{path: "item", depth: 1}
	}
	m.cursor = 0

	visible := m.visibleItems()
	if len(visible) != 5 {
		t.Errorf("visibleItems() returned %d items, want 5", len(visible))
	}
}

func TestPickerRenderNode(t *testing.T) {
	m := NewPicker(nil, false)
	m.width = 80

	tests := []struct {
		name     string
		node     *node
		isCursor bool
		wantSub  string
	}{
		{
			name:     "file not selected",
			node:     &node{name: ".bashrc", path: ".bashrc", depth: 1, isDir: false, state: StateNone},
			isCursor: false,
			wantSub:  ".bashrc",
		},
		{
			name:     "file with remove state",
			node:     &node{name: ".bashrc", path: ".bashrc", depth: 1, isDir: false, state: StateRemove},
			isCursor: false,
			wantSub:  ".bashrc",
		},
		{
			name:     "dir expanded",
			node:     &node{name: ".config", path: ".config", depth: 1, isDir: true, expanded: true},
			isCursor: false,
			wantSub:  "▼",
		},
		{
			name:     "dir collapsed",
			node:     &node{name: ".config", path: ".config", depth: 1, isDir: true, expanded: false},
			isCursor: false,
			wantSub:  "▶",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderNode(tt.node, tt.isCursor)
			if !containsSubstring(result, tt.wantSub) {
				t.Errorf("renderNode() = %q, should contain %q", result, tt.wantSub)
			}
		})
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestPickerLeftKeyCollapseDir(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create expanded directory (root is skipped in flat)
	dir := &node{name: ".config", path: ".config", isDir: true, expanded: true}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{dir}}
	dir.parent = root

	m.root = root
	m.rebuildFlat()
	m.cursor = 0 // On the directory (root is skipped)

	// Press 'h' to collapse
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if dir.expanded {
		t.Error("directory should be collapsed after 'h' key")
	}
}

func TestPickerLeftKeyJumpToParent(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create file inside a directory (so parent is visible in flat list)
	dir := &node{name: ".config", path: ".config", isDir: true, expanded: true}
	file := &node{name: "file.txt", path: ".config/file.txt", isDir: false}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{dir}}
	dir.parent = root
	dir.children = []*node{file}
	file.parent = dir

	m.root = root
	m.rebuildFlat() // flat: [dir, file]
	m.cursor = 1    // On the file

	// Press 'left' to jump to parent dir
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (parent dir)", m.cursor)
	}
}

func TestPickerSpaceStateCycle(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create node - set file as direct child so flat list is stable
	// Note: rebuildFlat skips root, so flat only contains children
	file := &node{name: ".bashrc", path: ".bashrc", isDir: false, state: StateNone}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{file}}
	file.parent = root

	m.root = root
	m.rebuildFlat() // Use rebuildFlat to ensure proper flat list
	m.cursor = 0    // Cursor on file (root is skipped in flat)

	// Verify we have the right setup
	if len(m.flat) != 1 {
		t.Fatalf("flat len = %d, want 1", len(m.flat))
	}

	// Space once: StateNone -> StateRemove
	msg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)
	if file.state != StateRemove {
		t.Errorf("after first space: state = %d, want StateRemove", file.state)
	}

	// Space again: StateRemove -> StateDisable
	updated, _ = m.Update(msg)
	m = updated.(PickerModel)
	if file.state != StateDisable {
		t.Errorf("after second space: state = %d, want StateDisable", file.state)
	}

	// Space again: StateDisable -> StateNone
	updated, _ = m.Update(msg)
	m = updated.(PickerModel)
	if file.state != StateNone {
		t.Errorf("after third space: state = %d, want StateNone", file.state)
	}
}

func TestPickerEnterExpandDir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, ".config")
	os.MkdirAll(testDir, 0755)

	m := NewPicker(nil, false)
	m.loaded = true
	m.home = tmpDir

	// Create collapsed directory (root is skipped in flat)
	dir := &node{name: ".config", path: ".config", isDir: true, expanded: false, loaded: false}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{dir}}
	dir.parent = root

	m.root = root
	m.rebuildFlat()
	m.cursor = 0 // On the directory (root is skipped)

	// Press 'enter' to expand
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if !dir.expanded {
		t.Error("directory should be expanded after 'enter' key")
	}
	if !dir.loaded {
		t.Error("directory should be loaded after expansion")
	}
}

func TestPickerRightKey(t *testing.T) {
	tmpDir := t.TempDir()

	m := NewPicker(nil, false)
	m.loaded = true
	m.home = tmpDir

	// Create collapsed directory (root is skipped in flat)
	dir := &node{name: ".config", path: ".config", isDir: true, expanded: false, loaded: true}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{dir}}
	dir.parent = root

	m.root = root
	m.rebuildFlat()
	m.cursor = 0 // On the directory (root is skipped)

	// Press 'right' to expand
	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if !dir.expanded {
		t.Error("directory should be expanded after 'right' key")
	}
}

func TestPickerLKey(t *testing.T) {
	tmpDir := t.TempDir()

	m := NewPicker(nil, false)
	m.loaded = true
	m.home = tmpDir

	// Create collapsed directory (root is skipped in flat)
	dir := &node{name: ".config", path: ".config", isDir: true, expanded: false, loaded: true}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{dir}}
	dir.parent = root

	m.root = root
	m.rebuildFlat()
	m.cursor = 0 // On the directory (root is skipped)

	// Press 'l' to expand
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if !dir.expanded {
		t.Error("directory should be expanded after 'l' key")
	}
}

func TestPickerUpKeyBoundary(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true
	m.cursor = 0

	// Create a single file (root is skipped in flat)
	file := &node{name: ".bashrc", path: ".bashrc"}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{file}}
	file.parent = root
	m.root = root
	m.rebuildFlat() // flat: [file]

	// Press 'up' at cursor 0
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should not go below 0)", m.cursor)
	}
}

func TestPickerDownKeyBoundary(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create a single file (root is skipped in flat)
	file := &node{name: ".bashrc", path: ".bashrc"}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{file}}
	file.parent = root
	m.root = root
	m.rebuildFlat() // flat: [file]
	m.cursor = 0    // Last (and only) item

	// Press 'down' at last item
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should not exceed len-1)", m.cursor)
	}
}

func TestPickerKKey(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create two files (root is skipped in flat)
	file1 := &node{name: ".bashrc", path: ".bashrc"}
	file2 := &node{name: ".zshrc", path: ".zshrc"}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{file1, file2}}
	file1.parent = root
	file2.parent = root
	m.root = root
	m.rebuildFlat() // flat: [file1, file2]
	m.cursor = 1    // On second file

	// Press 'k' to go up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestPickerJKey(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create two files (root is skipped in flat)
	file1 := &node{name: ".bashrc", path: ".bashrc"}
	file2 := &node{name: ".zshrc", path: ".zshrc"}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{file1, file2}}
	file1.parent = root
	file2.parent = root
	m.root = root
	m.rebuildFlat() // flat: [file1, file2]
	m.cursor = 0    // On first file

	// Press 'j' to go down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
}

func TestPickerEmptyFlatHandling(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Create empty root with no children (flat will be empty)
	root := &node{name: "~", path: "", isDir: true, expanded: true}
	m.root = root
	m.rebuildFlat() // flat: []

	// These should not panic
	msgs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyEnter},
		tea.KeyMsg{Type: tea.KeyLeft},
		tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}},
	}

	for _, msg := range msgs {
		updated, _ := m.Update(msg)
		m = updated.(PickerModel)
	}
	// If we got here without panic, test passes
}

func TestPickerLeftKeyCollapsedDirNoParent(t *testing.T) {
	m := NewPicker(nil, false)
	m.loaded = true

	// Collapsed directory at top level (parent is skipped root)
	dir := &node{name: ".config", path: ".config", isDir: true, expanded: false}
	root := &node{name: "~", path: "", isDir: true, expanded: true, children: []*node{dir}}
	dir.parent = root // parent is root, which is skipped in flat

	m.root = root
	m.rebuildFlat() // flat: [dir]
	m.cursor = 0

	// Press 'h' on collapsed dir with parent not in flat
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updated, _ := m.Update(msg)
	m = updated.(PickerModel)

	// Should not panic, cursor stays same (parent is root, not in flat)
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}
