// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InitVaultRepo initializes the vault as a git repository if not already.
func InitVaultRepo(vaultDir string) error {
	gitDir := filepath.Join(vaultDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Already a git repo
		return nil
	}

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = vaultDir
	return cmd.Run()
}

// CommitVault commits all changes in the vault with the given message.
func CommitVault(vaultDir, message string) error {
	// Add all
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = vaultDir
	if err := addCmd.Run(); err != nil {
		return err
	}

	// Check if there are changes to commit
	diffCmd := exec.Command("git", "diff", "--cached", "--quiet")
	diffCmd.Dir = vaultDir
	if err := diffCmd.Run(); err == nil {
		// No changes to commit
		return nil
	}

	// Commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = vaultDir
	return commitCmd.Run()
}

// HasRemote checks if the vault repo has a remote configured.
func HasRemote(vaultDir string) (bool, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = vaultDir
	output, err := cmd.Output()
	if err != nil {
		return false, "", nil // No remote configured
	}

	return true, strings.TrimSpace(string(output)), nil
}

// PushVault pushes the vault to the configured remote.
func PushVault(vaultDir string) error {
	hasRemote, _, err := HasRemote(vaultDir)
	if err != nil {
		return err
	}
	if !hasRemote {
		return fmt.Errorf("no remote configured. Run: cd %s && git remote add origin <url>", vaultDir)
	}

	// Get current branch
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = vaultDir
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))
	if branch == "" {
		branch = "main"
	}

	// Push
	pushCmd := exec.Command("git", "push", "-u", "origin", branch)
	pushCmd.Dir = vaultDir
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("push failed: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// PullResult contains the result of a pull operation.
type PullResult struct {
	Cloned bool
}

// PullVault pulls from the configured remote.
// If vault doesn't exist but remoteURL is provided, it clones first.
func PullVault(vaultDir string) (*PullResult, error) {
	return PullVaultWithRemote(vaultDir, "")
}

// PullVaultWithRemote pulls from remote, cloning first if vault doesn't exist.
func PullVaultWithRemote(vaultDir, remoteURL string) (*PullResult, error) {
	result := &PullResult{}

	// Check if vault exists
	gitDir := filepath.Join(vaultDir, ".git")
	vaultExists := false
	if _, err := os.Stat(gitDir); err == nil {
		vaultExists = true
	}

	if !vaultExists {
		// Need to clone
		if remoteURL == "" {
			return nil, fmt.Errorf("vault doesn't exist. Configure remote in Settings (F9) first")
		}

		// Ensure parent directory exists
		snapfigDir := filepath.Dir(vaultDir)
		if err := os.MkdirAll(snapfigDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Clone
		cloneCmd := exec.Command("git", "clone", remoteURL, vaultDir)
		if output, err := cloneCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("clone failed: %s", strings.TrimSpace(string(output)))
		}

		result.Cloned = true
		return result, nil
	}

	// Vault exists, do normal pull
	hasRemote, _, err := HasRemote(vaultDir)
	if err != nil {
		return nil, err
	}
	if !hasRemote {
		return nil, fmt.Errorf("no remote configured")
	}

	pullCmd := exec.Command("git", "pull")
	pullCmd.Dir = vaultDir
	if output, err := pullCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pull failed: %s", strings.TrimSpace(string(output)))
	}

	return result, nil
}

// SetRemote configures the remote origin for the vault.
func SetRemote(vaultDir, url string) error {
	// Ensure vault is a git repo
	if err := InitVaultRepo(vaultDir); err != nil {
		return err
	}

	// Check if origin already exists
	hasRemote, currentURL, err := HasRemote(vaultDir)
	if err != nil {
		return err
	}

	if hasRemote {
		if currentURL == url {
			return nil // Already set to this URL
		}
		// Update existing remote
		cmd := exec.Command("git", "remote", "set-url", "origin", url)
		cmd.Dir = vaultDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to update remote: %s", strings.TrimSpace(string(output)))
		}
	} else {
		// Add new remote
		cmd := exec.Command("git", "remote", "add", "origin", url)
		cmd.Dir = vaultDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add remote: %s", strings.TrimSpace(string(output)))
		}
	}

	return nil
}
