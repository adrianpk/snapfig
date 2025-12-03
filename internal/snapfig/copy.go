package snapfig

import (
	"io"
	"os"
	"path/filepath"

	"github.com/adrianpk/snapfig/internal/config"
)

// copyPath copies a file or directory from src to dst, handling .git according to mode.
// Uses smart copy: only copies files that have changed (by ModTime/Size).
func (c *Copier) copyPath(src, dst string, gitMode config.GitMode, result *CopyResult) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return c.copyFile(src, dst, info.Mode(), result)
	}

	return c.copyDir(src, dst, gitMode, result)
}

// copyDir recursively copies a directory, handling .git according to mode.
// Removes files from dst that no longer exist in src.
func (c *Copier) copyDir(src, dst string, gitMode config.GitMode, result *CopyResult) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Build set of source entries for stale detection
	srcEntries := make(map[string]bool)
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()

		// Track the destination name (for .git -> .git_disabled mapping)
		dstName := name
		if name == ".git" && entry.IsDir() {
			switch gitMode {
			case config.GitModeRemove:
				// .git is skipped, don't track it
				continue
			case config.GitModeDisable:
				dstName = ".git_disabled"
			}
		}
		srcEntries[dstName] = true
	}

	// Remove stale entries from destination
	if err := c.removeStale(dst, srcEntries, result); err != nil {
		return err
	}

	// Copy entries
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Handle .git directories
		if entry.Name() == ".git" && entry.IsDir() {
			switch gitMode {
			case config.GitModeRemove:
				continue
			case config.GitModeDisable:
				dstPath = filepath.Join(dst, ".git_disabled")
			}
		}

		if entry.IsDir() {
			if err := c.copyDir(srcPath, dstPath, gitMode, result); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := c.copyFile(srcPath, dstPath, info.Mode(), result); err != nil {
				return err
			}
		}
	}

	return nil
}

// removeStale removes files/dirs from dst that are not in srcEntries.
func (c *Copier) removeStale(dst string, srcEntries map[string]bool, result *CopyResult) error {
	dstEntries, err := os.ReadDir(dst)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, entry := range dstEntries {
		if !srcEntries[entry.Name()] {
			stalePath := filepath.Join(dst, entry.Name())
			if err := os.RemoveAll(stalePath); err != nil {
				return err
			}
			result.FilesRemoved++
		}
	}

	return nil
}

// shouldCopy checks if a file needs to be copied by comparing ModTime and Size.
// Returns true if the file should be copied (destination doesn't exist or differs).
func shouldCopy(src, dst string) (bool, error) {
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

	// Copy if size differs or source is newer
	if srcInfo.Size() != dstInfo.Size() {
		return true, nil
	}
	if srcInfo.ModTime().After(dstInfo.ModTime()) {
		return true, nil
	}

	return false, nil
}

// copyFile copies a single file preserving permissions.
// Skips copy if file hasn't changed (same ModTime and Size).
func (c *Copier) copyFile(src, dst string, mode os.FileMode, result *CopyResult) error {
	needsCopy, err := shouldCopy(src, dst)
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

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err == nil {
		result.FilesUpdated++
	}
	return err
}
