// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
)

// RestoreResult contains the result of a restore operation.
type RestoreResult struct {
	Restored     []string
	Skipped      []string
	Backups      []string // paths that were backed up before overwrite
	FilesUpdated int      // files actually copied (new or changed)
	FilesSkipped int      // files skipped (unchanged)
}

// Restorer handles restoring paths from the vault.
type Restorer struct {
	cfg        *config.Config
	home       string
	vaultDir   string
	backupTime string
}

// NewRestorer creates a new Restorer instance.
func NewRestorer(cfg *config.Config) (*Restorer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	vaultDir, err := config.DefaultVaultDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault directory: %w", err)
	}

	return &Restorer{
		cfg:        cfg,
		home:       home,
		vaultDir:   vaultDir,
		backupTime: time.Now().Format("200601021504"),
	}, nil
}

// Restore copies all enabled watched paths from vault to their original locations.
// Uses smart restore: only copies files that have changed (no full backup needed).
func (r *Restorer) Restore() (*RestoreResult, error) {
	result := &RestoreResult{}

	for _, w := range r.cfg.Watching {
		if !w.Enabled {
			continue
		}

		srcPath := filepath.Join(r.vaultDir, w.Path)
		dstPath := filepath.Join(r.home, w.Path)

		// Check if source exists in vault
		srcInfo, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			result.Skipped = append(result.Skipped, w.Path)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to stat vault path %s: %w", w.Path, err)
		}

		// Copy from vault to destination (smart restore - only changed files)
		gitMode := w.EffectiveGitMode(r.cfg.Git)
		if srcInfo.IsDir() {
			if err := r.restoreDir(srcPath, dstPath, gitMode, result); err != nil {
				return nil, fmt.Errorf("failed to restore %s: %w", w.Path, err)
			}
		} else {
			if err := r.restoreFile(srcPath, dstPath, srcInfo.Mode(), result); err != nil {
				return nil, fmt.Errorf("failed to restore %s: %w", w.Path, err)
			}
		}

		result.Restored = append(result.Restored, w.Path)
	}

	return result, nil
}

// restoreDir recursively copies a directory, reverting .git_disabled to .git.
// Uses smart restore: only copies files that have changed.
func (r *Restorer) restoreDir(src, dst string, gitMode config.GitMode, result *RestoreResult) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstName := entry.Name()

		// Revert .git_disabled back to .git
		if entry.Name() == ".git_disabled" && entry.IsDir() && gitMode == config.GitModeDisable {
			dstName = ".git"
		}

		dstPath := filepath.Join(dst, dstName)

		if entry.IsDir() {
			if err := r.restoreDir(srcPath, dstPath, gitMode, result); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := r.restoreFile(srcPath, dstPath, info.Mode(), result); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldRestore checks if a file needs to be restored by comparing Size and ModTime.
// Unlike shouldCopy (for backup), this copies if files differ in ANY way,
// since the vault is the source of truth during restore.
func shouldRestore(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	dstInfo, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// Restore if size differs
	if srcInfo.Size() != dstInfo.Size() {
		return true, nil
	}
	// Restore if ModTime differs (in either direction)
	if !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
		return true, nil
	}

	return false, nil
}

// restoreFile copies a single file preserving permissions and ModTime.
// Uses smart restore: skips if file hasn't changed (same ModTime and Size).
func (r *Restorer) restoreFile(src, dst string, mode os.FileMode, result *RestoreResult) error {
	needsCopy, err := shouldRestore(src, dst)
	if err != nil {
		return err
	}
	if !needsCopy {
		result.FilesSkipped++
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Preserve ModTime from source so smart restore can detect changes
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return err
	}

	result.FilesUpdated++
	return nil
}

// VaultEntry represents a file or directory in the vault that matches config.
type VaultEntry struct {
	Path     string   // relative path from vault root
	IsDir    bool     // whether it's a directory
	Children []string // for directories: child paths (relative to vault root)
}

// ListVaultEntries returns all entries in the vault that match the config watching list.
// It only returns entries that actually exist in the vault.
func (r *Restorer) ListVaultEntries() ([]VaultEntry, error) {
	var entries []VaultEntry

	for _, w := range r.cfg.Watching {
		if !w.Enabled {
			continue
		}

		vaultPath := filepath.Join(r.vaultDir, w.Path)
		info, err := os.Stat(vaultPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to stat vault path %s: %w", w.Path, err)
		}

		entry := VaultEntry{
			Path:  w.Path,
			IsDir: info.IsDir(),
		}

		if info.IsDir() {
			children, err := r.collectChildren(vaultPath, w.Path)
			if err != nil {
				return nil, err
			}
			entry.Children = children
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// collectChildren recursively collects all file paths within a directory.
func (r *Restorer) collectChildren(dirPath, relBase string) ([]string, error) {
	var children []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dirPath {
			return nil
		}

		rel, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		fullRel := filepath.Join(relBase, rel)
		children = append(children, fullRel)
		return nil
	})

	return children, err
}

// RestoreSelective restores only the specified paths from vault.
// paths should be relative paths (as they appear in config).
func (r *Restorer) RestoreSelective(paths []string) (*RestoreResult, error) {
	result := &RestoreResult{}

	// Create a map for quick lookup
	pathSet := make(map[string]bool)
	for _, p := range paths {
		pathSet[p] = true
	}

	for _, w := range r.cfg.Watching {
		if !w.Enabled {
			continue
		}

		// Check if this watched path or any of its children should be restored
		srcPath := filepath.Join(r.vaultDir, w.Path)
		dstPath := filepath.Join(r.home, w.Path)

		srcInfo, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to stat vault path %s: %w", w.Path, err)
		}

		gitMode := w.EffectiveGitMode(r.cfg.Git)

		if srcInfo.IsDir() {
			// For directories, check if whole dir or specific files should be restored
			if pathSet[w.Path] {
				// Restore entire directory
				if err := r.smartRestore(srcPath, dstPath, srcInfo, gitMode, w.Path, result); err != nil {
					return nil, err
				}
			} else {
				// Check for individual files within this directory
				restored, err := r.restoreSelectiveDir(srcPath, dstPath, w.Path, pathSet, gitMode, result)
				if err != nil {
					return nil, err
				}
				if !restored {
					result.Skipped = append(result.Skipped, w.Path)
				}
			}
		} else {
			// Single file
			if pathSet[w.Path] {
				if err := r.smartRestore(srcPath, dstPath, srcInfo, gitMode, w.Path, result); err != nil {
					return nil, err
				}
			}
		}
	}

	return result, nil
}

// smartRestore restores from vault using smart copy (only changed files).
func (r *Restorer) smartRestore(srcPath, dstPath string, srcInfo os.FileInfo, gitMode config.GitMode, relPath string, result *RestoreResult) error {
	if srcInfo.IsDir() {
		if err := r.restoreDir(srcPath, dstPath, gitMode, result); err != nil {
			return fmt.Errorf("failed to restore %s: %w", relPath, err)
		}
	} else {
		if err := r.restoreFile(srcPath, dstPath, srcInfo.Mode(), result); err != nil {
			return fmt.Errorf("failed to restore %s: %w", relPath, err)
		}
	}

	result.Restored = append(result.Restored, relPath)
	return nil
}

// restoreSelectiveDir restores only selected files within a directory.
func (r *Restorer) restoreSelectiveDir(srcDir, dstDir, baseRel string, pathSet map[string]bool, gitMode config.GitMode, result *RestoreResult) (bool, error) {
	anyRestored := false

	err := filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if srcPath == srcDir {
			return nil
		}

		rel, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		fullRel := filepath.Join(baseRel, rel)

		if pathSet[fullRel] {
			dstPath := filepath.Join(dstDir, rel)

			// Handle .git_disabled -> .git renaming
			if info.Name() == ".git_disabled" && info.IsDir() && gitMode == config.GitModeDisable {
				dstPath = filepath.Join(filepath.Dir(dstPath), ".git")
			}

			if info.IsDir() {
				// Skip walking into this directory, restore it completely
				if err := r.smartRestore(srcPath, dstPath, info, gitMode, fullRel, result); err != nil {
					return err
				}
				anyRestored = true
				return filepath.SkipDir
			} else {
				// Restore single file
				if err := r.smartRestore(srcPath, dstPath, info, gitMode, fullRel, result); err != nil {
					return err
				}
				anyRestored = true
			}
		}

		return nil
	})

	return anyRestored, err
}
