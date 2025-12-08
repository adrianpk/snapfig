package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/adrianpk/snapfig/internal/config"
)

func TestNewSettings(t *testing.T) {
	daemon := config.DaemonConfig{
		CopyInterval: "1h",
		PushInterval: "24h",
		PullInterval: "",
		AutoRestore:  false,
	}

	m := NewSettings("git@github.com:user/repo.git", "", "/custom/vault", daemon)

	if m.Remote() != "git@github.com:user/repo.git" {
		t.Errorf("Remote() = %q, want git@github.com:user/repo.git", m.Remote())
	}
	if m.VaultPath() != "/custom/vault" {
		t.Errorf("VaultPath() = %q, want /custom/vault", m.VaultPath())
	}
	if m.focused != fieldRemote {
		t.Errorf("focused = %d, want fieldRemote", m.focused)
	}
}

func TestSettingsInit(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return blink command")
	}
}

func TestSettingsNavigationDown(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})

	// Initial focus is fieldRemote (0)
	if m.focused != fieldRemote {
		t.Fatalf("initial focus = %d, want fieldRemote", m.focused)
	}

	// Press tab to move to next field (gitToken)
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	if m.focused != fieldGitToken {
		t.Errorf("after tab, focused = %d, want fieldGitToken", m.focused)
	}

	// Press down to move to next field (vaultPath)
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = m.Update(msg)
	m = updated.(SettingsModel)

	if m.focused != fieldVaultPath {
		t.Errorf("after down, focused = %d, want fieldVaultPath", m.focused)
	}
}

func TestSettingsNavigationUp(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})

	// Move to vault path first
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	// Now press up/shift+tab to go back
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = m.Update(msg)
	m = updated.(SettingsModel)

	if m.focused != fieldRemote {
		t.Errorf("after up, focused = %d, want fieldRemote", m.focused)
	}
}

func TestSettingsNavigationWraps(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})

	// Press up from first field should wrap to last
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	if m.focused != fieldAutoRestore {
		t.Errorf("wrap up, focused = %d, want fieldAutoRestore", m.focused)
	}

	// Press down should wrap back to first
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = m.Update(msg)
	m = updated.(SettingsModel)

	if m.focused != fieldRemote {
		t.Errorf("wrap down, focused = %d, want fieldRemote", m.focused)
	}
}

func TestSettingsEnterSaves(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})

	if m.WasSaved() {
		t.Error("WasSaved() should be false initially")
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	if !m.WasSaved() {
		t.Error("WasSaved() should be true after Enter")
	}
}

func TestSettingsAutoRestoreToggle(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{AutoRestore: false})

	// Navigate to auto restore field
	for i := 0; i < fieldAutoRestore; i++ {
		msg := tea.KeyMsg{Type: tea.KeyTab}
		updated, _ := m.Update(msg)
		m = updated.(SettingsModel)
	}

	if m.focused != fieldAutoRestore {
		t.Fatalf("focused = %d, want fieldAutoRestore", m.focused)
	}

	// Initial state
	dc := m.DaemonConfig()
	if dc.AutoRestore {
		t.Error("AutoRestore should be false initially")
	}

	// Press space to toggle
	msg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	dc = m.DaemonConfig()
	if !dc.AutoRestore {
		t.Error("AutoRestore should be true after toggle")
	}

	// Toggle again
	updated, _ = m.Update(msg)
	m = updated.(SettingsModel)

	dc = m.DaemonConfig()
	if dc.AutoRestore {
		t.Error("AutoRestore should be false after second toggle")
	}
}

func TestSettingsDaemonConfig(t *testing.T) {
	daemon := config.DaemonConfig{
		CopyInterval: "30m",
		PushInterval: "12h",
		PullInterval: "6h",
		AutoRestore:  true,
	}

	m := NewSettings("", "", "", daemon)
	result := m.DaemonConfig()

	if result.CopyInterval != "30m" {
		t.Errorf("CopyInterval = %q, want 30m", result.CopyInterval)
	}
	if result.PushInterval != "12h" {
		t.Errorf("PushInterval = %q, want 12h", result.PushInterval)
	}
	if result.PullInterval != "6h" {
		t.Errorf("PullInterval = %q, want 6h", result.PullInterval)
	}
	if !result.AutoRestore {
		t.Error("AutoRestore should be true")
	}
}

func TestSettingsView(t *testing.T) {
	m := NewSettings("git@github.com:test/repo.git", "", "", config.DaemonConfig{})
	m.width = 80
	m.height = 24

	view := m.View()

	if len(view) == 0 {
		t.Error("View() returned empty string")
	}

	// Should contain key elements
	if !containsString(view, "Settings") {
		t.Error("View() should contain title 'Settings'")
	}
	if !containsString(view, "Remote URL") {
		t.Error("View() should contain 'Remote URL'")
	}
	if !containsString(view, "Background Runner") {
		t.Error("View() should contain 'Background Runner'")
	}
}

func TestSettingsWindowSize(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSettingsGitToken(t *testing.T) {
	m := NewSettings("", "my-secret-token", "", config.DaemonConfig{})

	token := m.GitToken()
	if token != "my-secret-token" {
		t.Errorf("GitToken() = %q, want 'my-secret-token'", token)
	}
}

func TestSettingsRemote(t *testing.T) {
	m := NewSettings("git@github.com:user/repo.git", "", "", config.DaemonConfig{})

	remote := m.Remote()
	if remote != "git@github.com:user/repo.git" {
		t.Errorf("Remote() = %q, want 'git@github.com:user/repo.git'", remote)
	}
}

func TestSettingsVaultPath(t *testing.T) {
	m := NewSettings("", "", "/custom/vault/path", config.DaemonConfig{})

	vaultPath := m.VaultPath()
	if vaultPath != "/custom/vault/path" {
		t.Errorf("VaultPath() = %q, want '/custom/vault/path'", vaultPath)
	}
}

func TestSettingsWasSaved(t *testing.T) {
	m := NewSettings("", "", "", config.DaemonConfig{})

	if m.WasSaved() {
		t.Error("WasSaved() should be false initially")
	}

	// Trigger save with Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	m = updated.(SettingsModel)

	if !m.WasSaved() {
		t.Error("WasSaved() should be true after Enter")
	}
}
