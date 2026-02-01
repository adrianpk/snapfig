// Package screens contains the individual TUI screens.
package screens

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/paths"
	"github.com/adrianpk/snapfig/internal/tui/styles"
)

// SelectState represents the selection state of a node.
type SelectState int

const (
	StateNone    SelectState = iota // [ ] not selected
	StateRemove                     // [x] selected, remove .git
	StateDisable                    // [g] selected, disable .git
)

// SyncStatus represents the synchronization state of a path.
type SyncStatus int

const (
	SyncUntracked SyncStatus = iota // not in manifest
	SyncSynced                      // local + vault + manifest
	SyncNeedsBackup                 // local + manifest, missing from vault
	SyncNeedsRestore                // vault + manifest, missing from local
	SyncOrphan                      // only in manifest
)

type node struct {
	name      string
	path      string
	isDir     bool
	expanded  bool
	state     SelectState
	depth     int
	parent    *node
	children  []*node
	loaded    bool
	wellKnown bool
}

// PickerModel handles directory selection with tree navigation.
type PickerModel struct {
	root          *node
	flat          []*node
	cursor        int
	home          string
	err           error
	loaded        bool
	wellKnown     map[string]bool
	preselected   map[string]SelectState // from config
	width         int
	height        int
	demoMode      bool
	demoPaths     map[string]bool
	vaultPath     string          // relative to home, to exclude from listing
	vaultDir      string          // full path to vault directory
	manifestPaths map[string]bool // paths listed in manifest
}

type initMsg struct {
	home      string
	wellKnown map[string]bool
	err       error
}

// demoPaths returns paths safe to show in demo/screenshot mode.
func getDemoPaths() map[string]bool {
	paths := []string{
		// XDG directories (POSIX standard)
		"Documents",
		"Downloads",
		"Pictures",
		"Music",
		"Videos",
		"Desktop",
		"Public",
		"Templates",
		// Common config dirs safe for demo
		".config",
		".config/nvim",
		".config/emacs",
		".config/doom",
		".config/helix",
		".config/ghostty",
		".config/alacritty",
		".config/kitty",
		".config/wezterm",
		".config/tmux",
		".config/fish",
		".config/starship.toml",
		".config/i3",
		".config/sway",
		".config/hypr",
		".config/waybar",
		".config/rofi",
		".config/wofi",
		".config/dunst",
		".config/mpv",
		".config/ranger",
		".config/yazi",
		".config/zathura",
		// Shell configs
		".bashrc",
		".zshrc",
		".profile",
		".tmux.conf",
		// Git (generic)
		".gitconfig",
	}
	result := make(map[string]bool, len(paths))
	for _, p := range paths {
		result[p] = true
	}
	return result
}

// NewPicker creates a new tree picker screen with optional preselected paths from config.
func NewPicker(cfg *config.Config, demoMode bool) PickerModel {
	return NewPickerWithSync(cfg, demoMode, "", nil)
}

// NewPickerWithSync creates a picker with sync status information.
func NewPickerWithSync(cfg *config.Config, demoMode bool, vaultDir string, manifestPaths []string) PickerModel {
	preselected := make(map[string]SelectState)
	var vaultPath string
	if cfg != nil {
		for _, w := range cfg.Watching {
			if w.Enabled {
				state := StateRemove
				if w.Git == config.GitModeDisable {
					state = StateDisable
				}
				preselected[w.Path] = state
			}
		}
		// Get vault path relative to home to exclude from listing
		if cfgVaultDir, err := cfg.VaultDir(); err == nil {
			if vaultDir == "" {
				vaultDir = cfgVaultDir
			}
			if home, err := os.UserHomeDir(); err == nil {
				if rel, err := filepath.Rel(home, cfgVaultDir); err == nil {
					// Get the top-level directory (e.g., ".snapfig" from ".snapfig/vault")
					parts := strings.Split(rel, string(os.PathSeparator))
					if len(parts) > 0 {
						vaultPath = parts[0]
					}
				}
			}
		}
	}
	// Default to .snapfig if no config
	if vaultPath == "" {
		vaultPath = ".snapfig"
	}
	var demoPaths map[string]bool
	if demoMode {
		demoPaths = getDemoPaths()
	}

	// Build manifest paths map
	manifestMap := make(map[string]bool)
	for _, p := range manifestPaths {
		manifestMap[p] = true
	}

	return PickerModel{
		preselected:   preselected,
		demoMode:      demoMode,
		demoPaths:     demoPaths,
		vaultPath:     vaultPath,
		vaultDir:      vaultDir,
		manifestPaths: manifestMap,
	}
}

func (m PickerModel) Init() tea.Cmd {
	return initPicker
}

// initPicker is called asynchronously by bubbletea to initialize the picker.
// Async tea.Cmd; initMsg handling tested in Update tests.
func initPicker() tea.Msg {
	home, err := os.UserHomeDir()
	if err != nil {
		return initMsg{err: err}
	}

	known := paths.Known()
	wellKnown := make(map[string]bool, len(known))
	for _, p := range known {
		wellKnown[p] = true
		parts := strings.Split(p, string(os.PathSeparator))
		for i := 1; i < len(parts); i++ {
			parent := strings.Join(parts[:i], string(os.PathSeparator))
			wellKnown[parent] = true
		}
	}

	return initMsg{home: home, wellKnown: wellKnown}
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case initMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.home = msg.home
		m.wellKnown = msg.wellKnown
		m.root = &node{
			name:     "~",
			path:     "",
			isDir:    true,
			expanded: true,
			depth:    0,
		}
		m.loadChildren(m.root)
		m.rebuildFlat()
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
		case "enter", "l", "right":
			if len(m.flat) > 0 {
				n := m.flat[m.cursor]
				if n.isDir {
					if !n.loaded {
						m.loadChildren(n)
					}
					n.expanded = !n.expanded
					m.rebuildFlat()
				}
			}
		case "h", "left":
			if len(m.flat) > 0 {
				n := m.flat[m.cursor]
				if n.isDir && n.expanded {
					n.expanded = false
					m.rebuildFlat()
				} else if n.parent != nil {
					for i, fn := range m.flat {
						if fn == n.parent {
							m.cursor = i
							break
						}
					}
				}
			}
		case " ":
			if len(m.flat) > 0 {
				n := m.flat[m.cursor]
				// Cycle: None -> Remove -> Disable -> None
				switch n.state {
				case StateNone:
					n.state = StateRemove
				case StateRemove:
					n.state = StateDisable
				case StateDisable:
					n.state = StateNone
				}
				m.propagateState(n)
				m.rebuildFlat()
			}
		case "a":
			// Select all with Remove mode
			for _, n := range m.flat {
				n.state = StateRemove
			}
		case "n":
			m.clearState(m.root)
			m.rebuildFlat()
		}
	}

	return m, nil
}

func (m *PickerModel) loadChildren(n *node) {
	if n.loaded || !n.isDir {
		return
	}

	fullPath := filepath.Join(m.home, n.path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		n.loaded = true
		return
	}

	var dirs, files []*node
	for _, e := range entries {
		name := e.Name()

		// Only exclude the vault directory (e.g., .snapfig)
		if n.depth == 0 && name == m.vaultPath {
			continue
		}

		childPath := name
		if n.path != "" {
			childPath = filepath.Join(n.path, name)
		}

		// In demo mode, only show paths in the demo list
		if m.demoMode && !m.isDemoPath(childPath) {
			continue
		}

		child := &node{
			name:      name,
			path:      childPath,
			isDir:     e.IsDir(),
			depth:     n.depth + 1,
			parent:    n,
			wellKnown: m.wellKnown[childPath],
		}

		// Apply preselected state from config
		if state, ok := m.preselected[childPath]; ok {
			child.state = state
		}

		if e.IsDir() {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		iWK := m.wellKnown[dirs[i].path]
		jWK := m.wellKnown[dirs[j].path]
		if iWK != jWK {
			return iWK
		}
		return dirs[i].name < dirs[j].name
	})

	sort.Slice(files, func(i, j int) bool {
		iWK := m.wellKnown[files[i].path]
		jWK := m.wellKnown[files[j].path]
		if iWK != jWK {
			return iWK
		}
		return files[i].name < files[j].name
	})

	n.children = append(dirs, files...)
	n.loaded = true
}


// isDemoPath checks if a path should be shown in demo mode.
// Returns true if the path is in the demo list or is a parent of a demo path.
func (m *PickerModel) isDemoPath(path string) bool {
	// Direct match
	if m.demoPaths[path] {
		return true
	}
	// Check if this path is a parent of any demo path
	for demoPath := range m.demoPaths {
		if strings.HasPrefix(demoPath, path+"/") {
			return true
		}
	}
	return false
}

func (m *PickerModel) rebuildFlat() {
	m.flat = nil
	m.buildFlat(m.root)
}

func (m *PickerModel) buildFlat(n *node) {
	if n != m.root {
		m.flat = append(m.flat, n)
	}

	if n.isDir && n.expanded {
		for _, child := range n.children {
			m.buildFlat(child)
		}
	}
}

func (m *PickerModel) propagateState(n *node) {
	if n.isDir {
		m.setChildrenState(n, n.state)
	}
	m.updateParentState(n)
}

func (m *PickerModel) setChildrenState(n *node, state SelectState) {
	for _, child := range n.children {
		child.state = state
		if child.isDir {
			m.setChildrenState(child, state)
		}
	}
}

func (m *PickerModel) updateParentState(n *node) {
	if n.parent == nil {
		return
	}
	// Parent takes the state of children if all match, otherwise None
	firstState := n.parent.children[0].state
	allSame := true
	for _, sibling := range n.parent.children {
		if sibling.state != firstState {
			allSame = false
			break
		}
	}
	if allSame {
		n.parent.state = firstState
	} else {
		n.parent.state = StateNone
	}
	m.updateParentState(n.parent)
}

func (m *PickerModel) clearState(n *node) {
	n.state = StateNone
	for _, child := range n.children {
		m.clearState(child)
	}
}

func (m PickerModel) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Snapfig"))
	b.WriteString("\n")
	b.WriteString(styles.Subtitle.Render("Select directories to watch"))
	b.WriteString("\n\n")

	if !m.loaded {
		b.WriteString(styles.Dimmed.Render("Loading..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(styles.Error.Render("Error: " + m.err.Error()))
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
	b.WriteString(styles.Help.Render("↑/↓ navigate • ←/→ collapse/expand • space [x]remove/[g]disable • a all • n none • q quit"))

	return b.String()
}

func (m PickerModel) viewportStart() int {
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

func (m PickerModel) maxVisible() int {
	if m.height > 10 {
		return m.height - 8
	}
	return 15
}

func (m PickerModel) visibleItems() []*node {
	start := m.viewportStart()
	end := start + m.maxVisible()
	if end > len(m.flat) {
		end = len(m.flat)
	}
	return m.flat[start:end]
}

func (m PickerModel) renderNode(n *node, isCursor bool) string {
	indent := strings.Repeat("  ", n.depth)

	cursor := " "
	if isCursor {
		cursor = styles.CursorChar
	}

	var checkbox string
	switch n.state {
	case StateNone:
		checkbox = styles.UncheckedBox
	case StateRemove:
		checkbox = styles.CheckedBox
	case StateDisable:
		checkbox = styles.GitBox
	}

	icon := ""
	if n.isDir {
		if n.expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	} else {
		icon = "  "
	}

	name := n.name
	if n.isDir {
		name += "/"
	}

	line := cursor + " " + indent + checkbox + " " + icon + name

	if n.wellKnown {
		line += " ★"
	}

	// Add sync status tag
	status := m.getSyncStatus(n.path)
	tag := syncStatusTag(status)
	if tag != "" {
		line += " " + tag
	}

	if isCursor {
		return styles.Selected.Render(line)
	} else if n.state != StateNone {
		return styles.Normal.Render(line)
	} else if n.wellKnown {
		return styles.Subtitle.Render(line)
	}
	return styles.Dimmed.Render(line)
}

// getSyncStatus calculates the sync status for a path.
func (m PickerModel) getSyncStatus(path string) SyncStatus {
	if len(m.manifestPaths) == 0 {
		return SyncUntracked
	}

	inManifest := m.manifestPaths[path]
	if !inManifest {
		return SyncUntracked
	}

	// Check if exists in vault
	inVault := false
	if m.vaultDir != "" {
		vaultPath := filepath.Join(m.vaultDir, path)
		if _, err := os.Stat(vaultPath); err == nil {
			inVault = true
		}
	}

	// Check if exists locally (we know it does if it's in the picker,
	// but for completeness we check anyway)
	inLocal := false
	if m.home != "" {
		localPath := filepath.Join(m.home, path)
		if _, err := os.Stat(localPath); err == nil {
			inLocal = true
		}
	}

	// Determine status
	if inLocal && inVault && inManifest {
		return SyncSynced
	}
	if inLocal && inManifest && !inVault {
		return SyncNeedsBackup
	}
	if !inLocal && inVault && inManifest {
		return SyncNeedsRestore
	}
	if !inLocal && !inVault && inManifest {
		return SyncOrphan
	}

	return SyncUntracked
}

// syncStatusTag returns the display tag for a sync status.
func syncStatusTag(status SyncStatus) string {
	switch status {
	case SyncSynced:
		return "[synced]"
	case SyncNeedsBackup:
		return "[backup]"
	case SyncNeedsRestore:
		return "[restore]"
	case SyncOrphan:
		return "[orphan]"
	default:
		return ""
	}
}

// Selection represents a selected path with its git mode.
type Selection struct {
	Path    string
	GitMode SelectState
}

// Selected returns the list of selected paths with their git modes.
func (m PickerModel) Selected() []Selection {
	var selected []Selection
	m.collectSelected(m.root, &selected)
	return selected
}

func (m PickerModel) collectSelected(n *node, selected *[]Selection) {
	if n.state != StateNone && n != m.root {
		*selected = append(*selected, Selection{
			Path:    n.path,
			GitMode: n.state,
		})
		return
	}
	for _, child := range n.children {
		m.collectSelected(child, selected)
	}
}
