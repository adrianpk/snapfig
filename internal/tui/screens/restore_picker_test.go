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

func TestRestorePickerViewWithContent(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.width = 80
	m.height = 30 // Need height > 10 to exercise maxVisible branch

	// Set up tree with content
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	child1 := &RestoreNode{Path: ".bashrc", Parent: m.root, Depth: 1, Selected: true}
	child2 := &RestoreNode{Path: ".config", Parent: m.root, Depth: 1, IsDir: true, Expanded: true}
	child3 := &RestoreNode{Path: ".config/nvim", Parent: child2, Depth: 2, IsDir: true, Expanded: false}
	child2.Children = []*RestoreNode{child3}
	m.root.Children = []*RestoreNode{child1, child2}
	m.flat = []*RestoreNode{child1, child2, child3}
	m.cursor = 1

	view := m.View()

	if len(view) == 0 {
		t.Error("View() should return content")
	}
	// Verify it contains expected UI elements
	if !containsString(view, "Selective Restore") {
		t.Error("View should contain title")
	}
}

func TestRestorePickerViewEmptyFlat(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.width = 80
	m.height = 24
	m.root = &RestoreNode{Path: "root", IsDir: true}
	m.flat = []*RestoreNode{} // Empty

	view := m.View()
	if !containsString(view, "No files") {
		t.Error("View should show 'No files' for empty vault")
	}
}

func TestRestorePickerSelectChildren(t *testing.T) {
	m := NewRestorePicker()
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}

	// Parent dir with children
	dir := &RestoreNode{Path: ".config", IsDir: true, Parent: m.root, Selected: true}
	child1 := &RestoreNode{Path: ".config/nvim", Parent: dir}
	child2 := &RestoreNode{Path: ".config/emacs", Parent: dir}
	dir.Children = []*RestoreNode{child1, child2}
	m.root.Children = []*RestoreNode{dir}
	m.flat = []*RestoreNode{dir, child1, child2}
	m.cursor = 0

	// Propagate selection should select all children
	m.propagateSelection(dir)

	if !child1.Selected {
		t.Error("child1 should be selected")
	}
	if !child2.Selected {
		t.Error("child2 should be selected")
	}
}

func TestRestorePickerSelectChildrenDeselect(t *testing.T) {
	m := NewRestorePicker()
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}

	// Parent dir with children - start selected
	dir := &RestoreNode{Path: ".config", IsDir: true, Parent: m.root, Selected: false}
	child1 := &RestoreNode{Path: ".config/nvim", Parent: dir, Selected: true}
	child2 := &RestoreNode{Path: ".config/emacs", Parent: dir, Selected: true}
	dir.Children = []*RestoreNode{child1, child2}
	m.root.Children = []*RestoreNode{dir}

	// Propagate deselection
	m.propagateSelection(dir)

	if child1.Selected {
		t.Error("child1 should be deselected")
	}
	if child2.Selected {
		t.Error("child2 should be deselected")
	}
}

func TestRestorePickerClearSelection(t *testing.T) {
	m := NewRestorePicker()
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}

	// Create tree with selections
	dir := &RestoreNode{Path: ".config", IsDir: true, Parent: m.root, Selected: true}
	child := &RestoreNode{Path: ".config/nvim", Parent: dir, Selected: true}
	dir.Children = []*RestoreNode{child}
	m.root.Children = []*RestoreNode{dir}

	m.clearSelection(m.root)

	if m.root.Selected {
		t.Error("root should be deselected")
	}
	if dir.Selected {
		t.Error("dir should be deselected")
	}
	if child.Selected {
		t.Error("child should be deselected")
	}
}

func TestRestorePickerUpdateParentSelection(t *testing.T) {
	m := NewRestorePicker()
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}

	// Parent dir with children
	dir := &RestoreNode{Path: ".config", IsDir: true, Parent: m.root, Selected: false}
	child1 := &RestoreNode{Path: ".config/nvim", Parent: dir, Selected: true}
	child2 := &RestoreNode{Path: ".config/emacs", Parent: dir, Selected: true}
	dir.Children = []*RestoreNode{child1, child2}
	m.root.Children = []*RestoreNode{dir}

	// Update parent selection - should mark parent as selected since all children are
	m.updateParentSelection(child1)

	if !dir.Selected {
		t.Error("dir should be selected when all children are selected")
	}
}

func TestRestorePickerUpdateParentSelectionPartial(t *testing.T) {
	m := NewRestorePicker()
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}

	// Parent dir with children - only one selected
	dir := &RestoreNode{Path: ".config", IsDir: true, Parent: m.root, Selected: true}
	child1 := &RestoreNode{Path: ".config/nvim", Parent: dir, Selected: true}
	child2 := &RestoreNode{Path: ".config/emacs", Parent: dir, Selected: false}
	dir.Children = []*RestoreNode{child1, child2}
	m.root.Children = []*RestoreNode{dir}

	// Update parent selection - should deselect parent since not all children selected
	m.updateParentSelection(child1)

	if dir.Selected {
		t.Error("dir should not be selected when not all children are selected")
	}
}

func TestRestorePickerViewportCalculations(t *testing.T) {
	tests := []struct {
		name          string
		height        int
		flatLen       int
		cursor        int
		wantMaxVis    int
		wantStart     int
		wantVisibleCt int
	}{
		{
			name:          "small height uses default",
			height:        10,
			flatLen:       5,
			cursor:        0,
			wantMaxVis:    15,
			wantStart:     0,
			wantVisibleCt: 5,
		},
		{
			name:          "large height calculates max",
			height:        30,
			flatLen:       5,
			cursor:        0,
			wantMaxVis:    22, // 30-8
			wantStart:     0,
			wantVisibleCt: 5,
		},
		{
			name:          "cursor in middle with overflow",
			height:        20,
			flatLen:       30,
			cursor:        15,
			wantMaxVis:    12, // 20-8
			wantStart:     9,  // cursor - maxVisible/2
			wantVisibleCt: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRestorePicker()
			m.height = tt.height
			m.loaded = true

			// Build flat list
			m.flat = make([]*RestoreNode, tt.flatLen)
			for i := 0; i < tt.flatLen; i++ {
				m.flat[i] = &RestoreNode{Path: "item", Depth: 1}
			}
			m.cursor = tt.cursor

			maxVis := m.maxVisible()
			if maxVis != tt.wantMaxVis {
				t.Errorf("maxVisible() = %d, want %d", maxVis, tt.wantMaxVis)
			}

			start := m.viewportStart()
			if start != tt.wantStart {
				t.Errorf("viewportStart() = %d, want %d", start, tt.wantStart)
			}

			visible := m.visibleItems()
			if len(visible) != tt.wantVisibleCt {
				t.Errorf("visibleItems() = %d items, want %d", len(visible), tt.wantVisibleCt)
			}
		})
	}
}

func TestRestorePickerRenderNode(t *testing.T) {
	m := NewRestorePicker()
	m.width = 80

	tests := []struct {
		name     string
		node     *RestoreNode
		isCursor bool
		wantSub  string
	}{
		{
			name:     "file not selected not cursor",
			node:     &RestoreNode{Path: ".bashrc", Depth: 1, IsDir: false, Selected: false},
			isCursor: false,
			wantSub:  ".bashrc",
		},
		{
			name:     "file selected",
			node:     &RestoreNode{Path: ".bashrc", Depth: 1, IsDir: false, Selected: true},
			isCursor: false,
			wantSub:  ".bashrc",
		},
		{
			name:     "dir expanded",
			node:     &RestoreNode{Path: ".config", Depth: 1, IsDir: true, Expanded: true},
			isCursor: false,
			wantSub:  "▼",
		},
		{
			name:     "dir collapsed",
			node:     &RestoreNode{Path: ".config", Depth: 1, IsDir: true, Expanded: false},
			isCursor: false,
			wantSub:  "▶",
		},
		{
			name:     "nested path shows basename",
			node:     &RestoreNode{Path: ".config/nvim/init.lua", Depth: 3, IsDir: false},
			isCursor: false,
			wantSub:  "init.lua",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderNode(tt.node, tt.isCursor)
			if !containsString(result, tt.wantSub) {
				t.Errorf("renderNode() = %q, should contain %q", result, tt.wantSub)
			}
		})
	}
}

func TestRestorePickerSelectAll(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	child1 := &RestoreNode{Path: ".bashrc", Parent: m.root, Selected: false}
	child2 := &RestoreNode{Path: ".zshrc", Parent: m.root, Selected: false}
	m.root.Children = []*RestoreNode{child1, child2}
	m.flat = []*RestoreNode{child1, child2}

	// Press 'a' to select all
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	m.Update(msg)

	if !child1.Selected {
		t.Error("child1 should be selected after 'a'")
	}
	if !child2.Selected {
		t.Error("child2 should be selected after 'a'")
	}
}

func TestRestorePickerSelectNone(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true, Selected: true}
	child1 := &RestoreNode{Path: ".bashrc", Parent: m.root, Selected: true}
	child2 := &RestoreNode{Path: ".zshrc", Parent: m.root, Selected: true}
	m.root.Children = []*RestoreNode{child1, child2}
	m.flat = []*RestoreNode{child1, child2}

	// Press 'n' to select none
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	m.Update(msg)

	if child1.Selected {
		t.Error("child1 should be deselected after 'n'")
	}
	if child2.Selected {
		t.Error("child2 should be deselected after 'n'")
	}
}

func TestRestorePickerExpandCollapse(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	dir := &RestoreNode{Path: ".config", IsDir: true, Parent: m.root, Expanded: true}
	child := &RestoreNode{Path: ".config/nvim", Parent: dir}
	dir.Children = []*RestoreNode{child}
	m.root.Children = []*RestoreNode{dir}
	m.flat = []*RestoreNode{dir, child}
	m.cursor = 0

	// Press left to collapse
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	if dir.Expanded {
		t.Error("dir should be collapsed after left arrow")
	}

	// Press right to expand
	msg = tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = m.Update(msg)
	m = updated.(RestorePickerModel)

	if !dir.Expanded {
		t.Error("dir should be expanded after right arrow")
	}
}

func TestRestorePickerEnterNotHandled(t *testing.T) {
	// Note: Enter key is documented in help but not implemented
	// This test verifies current behavior
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true}
	child := &RestoreNode{Path: ".bashrc", Parent: m.root, Selected: true}
	m.root.Children = []*RestoreNode{child}
	m.flat = []*RestoreNode{child}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	m = updated.(RestorePickerModel)

	// Enter doesn't complete - only Esc does
	if m.done {
		t.Error("done should be false after Enter (not implemented)")
	}
}

func TestRestorePickerRKey(t *testing.T) {
	m := NewRestorePicker()
	m.loaded = true
	m.root = &RestoreNode{Path: "root", IsDir: true, Expanded: true}
	child := &RestoreNode{Path: ".bashrc", Parent: m.root, Selected: false}
	m.root.Children = []*RestoreNode{child}
	m.flat = []*RestoreNode{child}
	m.cursor = 0

	// Press 'r' to toggle (same as space)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	m.Update(msg)

	if !child.Selected {
		t.Error("child should be selected after 'r'")
	}
}

func TestRestorePickerAddChildrenSimple(t *testing.T) {
	m := NewRestorePicker()
	parent := &RestoreNode{Path: ".config", Depth: 1}

	children := []string{
		".config/nvim",
		".config/zsh",
	}

	m.addChildren(parent, children, ".config")

	if len(parent.Children) != 2 {
		t.Errorf("parent.Children len = %d, want 2", len(parent.Children))
	}
}

func TestRestorePickerAddChildrenNested(t *testing.T) {
	m := NewRestorePicker()
	parent := &RestoreNode{Path: ".config", Depth: 1}

	// Nested paths that should create directory structure
	children := []string{
		".config/nvim/init.lua",
		".config/nvim/lua/plugins.lua",
	}

	m.addChildren(parent, children, ".config")

	// Should have one child: nvim (directory)
	if len(parent.Children) != 1 {
		t.Fatalf("parent.Children len = %d, want 1", len(parent.Children))
	}

	nvimNode := parent.Children[0]
	if nvimNode.Path != ".config/nvim" {
		t.Errorf("nvim path = %q, want '.config/nvim'", nvimNode.Path)
	}
	if !nvimNode.IsDir {
		t.Error("nvim should be a directory")
	}
}

func TestRestorePickerAddChildrenDeepNesting(t *testing.T) {
	m := NewRestorePicker()
	parent := &RestoreNode{Path: ".config", Depth: 1}

	children := []string{
		".config/a/b/c/file.txt",
	}

	m.addChildren(parent, children, ".config")

	// Should create: a -> b -> c -> file.txt
	if len(parent.Children) != 1 {
		t.Fatalf("parent.Children len = %d, want 1", len(parent.Children))
	}

	// Traverse the tree
	aNode := parent.Children[0]
	if aNode.Path != ".config/a" {
		t.Errorf("a path = %q, want '.config/a'", aNode.Path)
	}
}

func TestRestorePickerAddChildrenSkipsInvalidPaths(t *testing.T) {
	m := NewRestorePicker()
	parent := &RestoreNode{Path: ".config", Depth: 1}

	children := []string{
		"",                   // Empty
		".config",            // Same as base (no relative part)
		".other/file",        // Different base (doesn't start with .config/)
		".config/valid/file", // Valid
	}

	m.addChildren(parent, children, ".config")

	// Should only have the valid child
	if len(parent.Children) != 1 {
		t.Errorf("parent.Children len = %d, want 1", len(parent.Children))
	}
}

func TestRestorePickerAddChildrenDirectoryDetection(t *testing.T) {
	m := NewRestorePicker()
	parent := &RestoreNode{Path: ".config", Depth: 1}

	// When a path has children, it should be detected as a directory
	children := []string{
		".config/dir",
		".config/dir/file.txt", // This makes dir a directory
	}

	m.addChildren(parent, children, ".config")

	if len(parent.Children) != 1 {
		t.Fatalf("parent.Children len = %d, want 1", len(parent.Children))
	}

	dirNode := parent.Children[0]
	if !dirNode.IsDir {
		t.Error("dir should be detected as directory because it has children")
	}
}

func TestRestorePickerAddChildrenExistingDirectory(t *testing.T) {
	m := NewRestorePicker()
	parent := &RestoreNode{Path: ".config", Depth: 1}

	// Multiple paths in the same directory
	children := []string{
		".config/nvim/init.lua",
		".config/nvim/plugins.lua",
	}

	m.addChildren(parent, children, ".config")

	// Should only create one nvim directory
	if len(parent.Children) != 1 {
		t.Fatalf("parent.Children len = %d, want 1 (nvim)", len(parent.Children))
	}

	nvimNode := parent.Children[0]
	// nvim should have 2 children (init.lua and plugins.lua)
	if len(nvimNode.Children) != 2 {
		t.Errorf("nvim.Children len = %d, want 2", len(nvimNode.Children))
	}
}
