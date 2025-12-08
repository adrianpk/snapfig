// Package screens contains the individual TUI screens.
package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/adrianpk/snapfig/internal/snapfig"
	"github.com/adrianpk/snapfig/internal/tui/styles"
)

// RestoreNode represents a file or directory available for selective restore.
type RestoreNode struct {
	Path     string
	IsDir    bool
	Depth    int
	Selected bool
	Expanded bool
	Parent   *RestoreNode
	Children []*RestoreNode
}

// RestorePickerModel handles selective restore with tree navigation.
type RestorePickerModel struct {
	root     *RestoreNode
	flat     []*RestoreNode
	cursor   int
	width    int
	height   int
	loaded   bool
	err      error
	done     bool
	canceled bool
}

// RestorePickerInitMsg is sent when vault entries are loaded.
type RestorePickerInitMsg struct {
	Entries []snapfig.VaultEntry
	Err     error
}

// NewRestorePicker creates a new restore picker screen.
func NewRestorePicker() RestorePickerModel {
	return RestorePickerModel{}
}

// InitRestorePicker loads vault entries matching config.
// Returns async tea.Cmd; RestorePickerInitMsg handling tested in Update tests.
func InitRestorePicker(restorer *snapfig.Restorer) tea.Cmd {
	return func() tea.Msg {
		entries, err := restorer.ListVaultEntries()
		return RestorePickerInitMsg{Entries: entries, Err: err}
	}
}

func (m RestorePickerModel) Init() tea.Cmd {
	return nil
}

func (m RestorePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case RestorePickerInitMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loaded = true
			return m, nil
		}
		m.buildTree(msg.Entries)
		m.loaded = true

	case tea.KeyMsg:
		if !m.loaded {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.flat)-1 {
				m.cursor++
			}
		case "l", "right":
			if len(m.flat) > 0 {
				n := m.flat[m.cursor]
				if n.IsDir && len(n.Children) > 0 {
					n.Expanded = !n.Expanded
					m.rebuildFlat()
				}
			}
		case "h", "left":
			if len(m.flat) > 0 {
				n := m.flat[m.cursor]
				if n.IsDir && n.Expanded {
					n.Expanded = false
					m.rebuildFlat()
				} else if n.Parent != nil {
					for i, fn := range m.flat {
						if fn == n.Parent {
							m.cursor = i
							break
						}
					}
				}
			}
		case " ", "r":
			if len(m.flat) > 0 {
				n := m.flat[m.cursor]
				n.Selected = !n.Selected
				m.propagateSelection(n)
				m.rebuildFlat()
			}
		case "a":
			// Select all
			for _, n := range m.flat {
				n.Selected = true
			}
		case "n":
			// Clear selection
			m.clearSelection(m.root)
			m.rebuildFlat()
		case "esc":
			m.canceled = true
			m.done = true
		}
	}

	return m, nil
}

func (m *RestorePickerModel) buildTree(entries []snapfig.VaultEntry) {
	m.root = &RestoreNode{
		Path:     "",
		IsDir:    true,
		Expanded: true,
	}

	// Build nodes for each vault entry
	for _, entry := range entries {
		node := &RestoreNode{
			Path:     entry.Path,
			IsDir:    entry.IsDir,
			Depth:    1,
			Expanded: false,
			Parent:   m.root,
		}

		if entry.IsDir && len(entry.Children) > 0 {
			// Add children as sub-nodes
			m.addChildren(node, entry.Children, entry.Path)
		}

		m.root.Children = append(m.root.Children, node)
	}

	m.rebuildFlat()
}

func (m *RestorePickerModel) addChildren(parent *RestoreNode, children []string, basePath string) {
	// Build a tree structure from flat paths
	dirs := make(map[string]*RestoreNode)

	for _, childPath := range children {
		rel := strings.TrimPrefix(childPath, basePath+"/")
		if rel == "" || rel == childPath {
			continue
		}

		parts := strings.Split(rel, "/")
		current := parent

		for i := range parts {
			isLast := i == len(parts)-1
			fullPath := basePath + "/" + strings.Join(parts[:i+1], "/")

			if !isLast {
				// This is a directory
				if existing, ok := dirs[fullPath]; ok {
					current = existing
				} else {
					dirNode := &RestoreNode{
						Path:     fullPath,
						IsDir:    true,
						Depth:    parent.Depth + i + 1,
						Parent:   current,
						Expanded: false,
					}
					current.Children = append(current.Children, dirNode)
					dirs[fullPath] = dirNode
					current = dirNode
				}
			} else {
				// This is the final item
				if _, exists := dirs[fullPath]; !exists {
					child := &RestoreNode{
						Path:   childPath,
						IsDir:  false, // Could be dir too, but simplified for now
						Depth:  parent.Depth + i + 1,
						Parent: current,
					}
					// Check if this is actually a directory by looking at other paths
					for _, other := range children {
						if strings.HasPrefix(other, childPath+"/") {
							child.IsDir = true
							break
						}
					}
					current.Children = append(current.Children, child)
					if child.IsDir {
						dirs[fullPath] = child
					}
				}
			}
		}
	}
}

func (m *RestorePickerModel) rebuildFlat() {
	m.flat = nil
	m.buildFlat(m.root)
}

func (m *RestorePickerModel) buildFlat(n *RestoreNode) {
	if n != m.root {
		m.flat = append(m.flat, n)
	}

	if n.IsDir && n.Expanded {
		for _, child := range n.Children {
			m.buildFlat(child)
		}
	}
}

func (m *RestorePickerModel) propagateSelection(n *RestoreNode) {
	// If selecting a directory, select all children
	if n.IsDir && n.Selected {
		m.selectChildren(n, true)
	} else if n.IsDir && !n.Selected {
		m.selectChildren(n, false)
	}

	// Update parent state
	m.updateParentSelection(n)
}

func (m *RestorePickerModel) selectChildren(n *RestoreNode, selected bool) {
	for _, child := range n.Children {
		child.Selected = selected
		if child.IsDir {
			m.selectChildren(child, selected)
		}
	}
}

func (m *RestorePickerModel) updateParentSelection(n *RestoreNode) {
	if n.Parent == nil || n.Parent == m.root {
		return
	}

	// Parent is selected if all children are selected
	allSelected := true
	for _, sibling := range n.Parent.Children {
		if !sibling.Selected {
			allSelected = false
			break
		}
	}
	n.Parent.Selected = allSelected
	m.updateParentSelection(n.Parent)
}

func (m *RestorePickerModel) clearSelection(n *RestoreNode) {
	n.Selected = false
	for _, child := range n.Children {
		m.clearSelection(child)
	}
}

func (m RestorePickerModel) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Selective Restore"))
	b.WriteString("\n")
	b.WriteString(styles.Subtitle.Render("Select files to restore from vault"))
	b.WriteString("\n\n")

	if !m.loaded {
		b.WriteString(styles.Dimmed.Render("Loading vault contents..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(styles.Error.Render("Error: " + m.err.Error()))
		return b.String()
	}

	if len(m.flat) == 0 {
		b.WriteString(styles.Dimmed.Render("No files in vault matching config"))
		b.WriteString("\n\n")
		b.WriteString(styles.Help.Render("Press Esc to go back"))
		return b.String()
	}

	visible := m.visibleItems()
	for i, n := range visible {
		actualIdx := m.flat[m.viewportStart()+i]
		isCursor := m.flat[m.cursor] == actualIdx

		line := m.renderNode(n, isCursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.Help.Render("↑/↓ navigate • ←/→ collapse/expand • Space/r select • a all • n none • Enter restore • Esc cancel"))

	return b.String()
}

func (m RestorePickerModel) viewportStart() int {
	maxVisible := m.maxVisible()
	if len(m.flat) <= maxVisible {
		return 0
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	if start+maxVisible > len(m.flat) {
		start = len(m.flat) - maxVisible
	}
	return start
}

func (m RestorePickerModel) maxVisible() int {
	if m.height > 10 {
		return m.height - 8
	}
	return 15
}

func (m RestorePickerModel) visibleItems() []*RestoreNode {
	start := m.viewportStart()
	end := start + m.maxVisible()
	if end > len(m.flat) {
		end = len(m.flat)
	}
	return m.flat[start:end]
}

func (m RestorePickerModel) renderNode(n *RestoreNode, isCursor bool) string {
	indent := strings.Repeat("  ", n.Depth-1)

	cursor := " "
	if isCursor {
		cursor = styles.CursorChar
	}

	checkbox := styles.UncheckedBox
	if n.Selected {
		checkbox = styles.RestoreBox
	}

	icon := ""
	if n.IsDir {
		if n.Expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	} else {
		icon = "  "
	}

	// Show just the filename for cleaner display
	name := n.Path
	if idx := strings.LastIndex(n.Path, "/"); idx >= 0 {
		name = n.Path[idx+1:]
	}
	if n.IsDir {
		name += "/"
	}

	line := cursor + " " + indent + checkbox + " " + icon + name

	if isCursor {
		return styles.Selected.Render(line)
	} else if n.Selected {
		return styles.Normal.Render(line)
	}
	return styles.Dimmed.Render(line)
}

// Selected returns the list of selected paths for restore.
func (m RestorePickerModel) Selected() []string {
	var selected []string
	m.collectSelected(m.root, &selected)
	return selected
}

func (m RestorePickerModel) collectSelected(n *RestoreNode, selected *[]string) {
	if n.Selected && n != m.root {
		*selected = append(*selected, n.Path)
		return // Don't include children if parent is selected
	}
	for _, child := range n.Children {
		m.collectSelected(child, selected)
	}
}

// IsDone returns true if the user has confirmed or canceled.
func (m RestorePickerModel) IsDone() bool {
	return m.done
}

// WasCanceled returns true if the user pressed Esc.
func (m RestorePickerModel) WasCanceled() bool {
	return m.canceled
}

// Loaded returns true if vault entries have been loaded.
func (m RestorePickerModel) Loaded() bool {
	return m.loaded
}

// HasError returns the error if loading failed.
func (m RestorePickerModel) HasError() error {
	return m.err
}

// MarkDone marks the picker as done (for external confirmation).
func (m *RestorePickerModel) MarkDone() {
	m.done = true
}
