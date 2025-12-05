package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKnown(t *testing.T) {
	paths := Known()

	// Verify we get a non-empty list
	if len(paths) == 0 {
		t.Error("Known() returned empty list")
	}

	// Verify some expected paths are present
	expectedPaths := []string{
		".bashrc",
		".zshrc",
		".config/nvim",
		".gitconfig",
		".ssh/config",
	}

	pathSet := make(map[string]bool)
	for _, p := range paths {
		pathSet[p] = true
	}

	for _, expected := range expectedPaths {
		if !pathSet[expected] {
			t.Errorf("Known() missing expected path %q", expected)
		}
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, p := range paths {
		if seen[p] {
			t.Errorf("Known() has duplicate path %q", p)
		}
		seen[p] = true
	}

	// Verify paths are relative (not absolute)
	for _, p := range paths {
		if filepath.IsAbs(p) {
			t.Errorf("Known() path %q is absolute, expected relative", p)
		}
	}
}

func TestExisting(t *testing.T) {
	// Call Existing() - should work with real home dir
	existing, err := Existing()
	if err != nil {
		t.Fatalf("Existing() error: %v", err)
	}

	// Result should be a subset of Known()
	knownSet := make(map[string]bool)
	for _, p := range Known() {
		knownSet[p] = true
	}

	for _, p := range existing {
		if !knownSet[p] {
			t.Errorf("Existing() returned path %q not in Known()", p)
		}
	}
}

func TestExistingVerifiesActualExistence(t *testing.T) {
	existing, err := Existing()
	if err != nil {
		t.Fatalf("Existing() error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Verify each returned path actually exists
	for _, p := range existing {
		fullPath := filepath.Join(home, p)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Existing() returned %q but it doesn't exist at %q", p, fullPath)
		}
	}
}

func TestExistingFiltersNonExistent(t *testing.T) {
	existing, err := Existing()
	if err != nil {
		t.Fatalf("Existing() error: %v", err)
	}

	known := Known()

	// Existing should be <= Known in count
	if len(existing) > len(known) {
		t.Errorf("Existing() returned %d paths, but Known() only has %d",
			len(existing), len(known))
	}
}

func TestKnownCategories(t *testing.T) {
	paths := Known()
	pathSet := make(map[string]bool)
	for _, p := range paths {
		pathSet[p] = true
	}

	// Test shell configs exist
	shellConfigs := []string{".bashrc", ".zshrc", ".config/fish"}
	for _, p := range shellConfigs {
		if !pathSet[p] {
			t.Errorf("Known() missing shell config %q", p)
		}
	}

	// Test editor configs exist
	editorConfigs := []string{".config/nvim", ".config/helix"}
	for _, p := range editorConfigs {
		if !pathSet[p] {
			t.Errorf("Known() missing editor config %q", p)
		}
	}

	// Test terminal configs exist
	terminalConfigs := []string{".config/alacritty", ".config/kitty"}
	for _, p := range terminalConfigs {
		if !pathSet[p] {
			t.Errorf("Known() missing terminal config %q", p)
		}
	}

	// Test git configs exist
	gitConfigs := []string{".gitconfig", ".config/git"}
	for _, p := range gitConfigs {
		if !pathSet[p] {
			t.Errorf("Known() missing git config %q", p)
		}
	}
}
