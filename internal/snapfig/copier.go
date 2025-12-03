// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
)

// CopyResult contains the result of a copy operation.
type CopyResult struct {
	Copied       []string
	Skipped      []string
	FilesUpdated int   // files actually copied (new or changed)
	FilesSkipped int   // files skipped (unchanged)
	FilesRemoved int   // stale files removed from vault
	GitError     error // non-fatal git error
}

// CopiedItem represents an item that was copied with its git mode.
type CopiedItem struct {
	Path    string
	GitMode config.GitMode
	IsDir   bool
}

// Copier handles copying watched paths to the vault.
type Copier struct {
	cfg         *config.Config
	home        string
	vaultDir    string
	snapfigDir  string
	copiedItems []CopiedItem
}

// NewCopier creates a new Copier instance.
func NewCopier(cfg *config.Config) (*Copier, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	vaultDir, err := config.DefaultVaultDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault directory: %w", err)
	}

	snapfigDir := filepath.Dir(vaultDir) // ~/.snapfig

	return &Copier{
		cfg:        cfg,
		home:       home,
		vaultDir:   vaultDir,
		snapfigDir: snapfigDir,
	}, nil
}

// Copy copies all enabled watched paths to the vault.
func (c *Copier) Copy() (*CopyResult, error) {
	result := &CopyResult{}
	c.copiedItems = nil

	if err := os.MkdirAll(c.vaultDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create vault directory: %w", err)
	}

	for _, w := range c.cfg.Watching {
		if !w.Enabled {
			continue
		}

		srcPath := filepath.Join(c.home, w.Path)
		dstPath := filepath.Join(c.vaultDir, w.Path)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			result.Skipped = append(result.Skipped, w.Path)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", w.Path, err)
		}

		// Smart copy: no RemoveAll, copyPath handles incremental updates
		gitMode := w.EffectiveGitMode(c.cfg.Git)
		if err := c.copyPath(srcPath, dstPath, gitMode, result); err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", w.Path, err)
		}

		c.copiedItems = append(c.copiedItems, CopiedItem{
			Path:    w.Path,
			GitMode: gitMode,
			IsDir:   info.IsDir(),
		})
		result.Copied = append(result.Copied, w.Path)
	}

	// Write manifest
	if err := c.writeManifest(); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Initialize git repo if needed and commit
	if err := InitVaultRepo(); err != nil {
		// Non-fatal: git might not be installed
		result.GitError = err
	} else {
		msg := fmt.Sprintf("snapfig: backup %d paths", len(result.Copied))
		if err := CommitVault(msg); err != nil {
			result.GitError = err
		}
	}

	return result, nil
}

// writeManifest creates a markdown manifest of copied items.
func (c *Copier) writeManifest() error {
	manifestPath := filepath.Join(c.snapfigDir, "manifest.md")

	var content string
	content += "# Snapfig Manifest\n\n"
	content += fmt.Sprintf("Last updated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	content += "## Backed Up Paths\n\n"
	content += "| Path | Type | Git Mode |\n"
	content += "|------|------|----------|\n"

	for _, item := range c.copiedItems {
		itemType := "file"
		if item.IsDir {
			itemType = "dir"
		}

		gitModeStr := string(item.GitMode)
		if item.GitMode == config.GitModeDisable {
			gitModeStr = "disable (.git → .git_disabled)"
		} else if item.GitMode == config.GitModeRemove {
			gitModeStr = "remove (.git deleted)"
		}

		content += fmt.Sprintf("| `%s` | %s | %s |\n", item.Path, itemType, gitModeStr)
	}

	content += fmt.Sprintf("\n## Summary\n\n- **Total items**: %d\n", len(c.copiedItems))
	content += fmt.Sprintf("- **Vault location**: `%s`\n", c.vaultDir)

	return os.WriteFile(manifestPath, []byte(content), 0644)
}
