// Package snapfig implements core operations for configuration management.
package snapfig

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// sshURLRegex matches SSH-style git URLs like git@github.com:user/repo.git
var sshURLRegex = regexp.MustCompile(`^git@([^:]+):(.+)$`)

// urlWithToken returns a URL with the token embedded for HTTPS auth.
// If no token is provided, returns the original URL unchanged.
// If the URL is SSH format and token is provided, converts to HTTPS with token.
func urlWithToken(remoteURL, token string) string {
	if token == "" {
		return remoteURL
	}

	// Check if it's an SSH URL (git@host:path)
	if matches := sshURLRegex.FindStringSubmatch(remoteURL); matches != nil {
		host := matches[1]
		path := matches[2]
		return fmt.Sprintf("https://x-access-token:%s@%s/%s", token, host, path)
	}

	// Check if it's already an HTTPS URL
	if strings.HasPrefix(remoteURL, "https://") {
		parsed, err := url.Parse(remoteURL)
		if err != nil {
			return remoteURL
		}
		parsed.User = url.UserPassword("x-access-token", token)
		return parsed.String()
	}

	// Unknown format, return as-is
	return remoteURL
}

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
	return PushVaultWithToken(vaultDir, "")
}

// PushVaultWithToken pushes the vault to the configured remote using token auth if provided.
// If token is empty, uses the configured remote URL directly (SSH or other).
func PushVaultWithToken(vaultDir, token string) error {
	hasRemote, remoteURL, err := HasRemote(vaultDir)
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

	// Push - use token-embedded URL if token provided
	var pushCmd *exec.Cmd
	if token != "" {
		authURL := urlWithToken(remoteURL, token)
		pushCmd = exec.Command("git", "push", "-u", authURL, branch)
	} else {
		pushCmd = exec.Command("git", "push", "-u", "origin", branch)
	}
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
	return PullVaultWithToken(vaultDir, "", "")
}

// PullVaultWithRemote pulls from remote, cloning first if vault doesn't exist.
func PullVaultWithRemote(vaultDir, remoteURL string) (*PullResult, error) {
	return PullVaultWithToken(vaultDir, remoteURL, "")
}

// PullVaultWithToken pulls from remote using token auth if provided.
// If token is empty, uses SSH or configured credentials.
func PullVaultWithToken(vaultDir, remoteURL, token string) (*PullResult, error) {
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

		// Clone - use token-embedded URL if token provided
		cloneURL := urlWithToken(remoteURL, token)
		cloneCmd := exec.Command("git", "clone", cloneURL, vaultDir)
		if output, err := cloneCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("clone failed: %s", strings.TrimSpace(string(output)))
		}

		result.Cloned = true
		return result, nil
	}

	// Vault exists, do normal pull
	hasRemote, currentRemoteURL, err := HasRemote(vaultDir)
	if err != nil {
		return nil, err
	}
	if !hasRemote {
		return nil, fmt.Errorf("no remote configured")
	}

	// Pull - use token-embedded URL if token provided
	var pullCmd *exec.Cmd
	if token != "" {
		authURL := urlWithToken(currentRemoteURL, token)
		pullCmd = exec.Command("git", "pull", authURL)
	} else {
		pullCmd = exec.Command("git", "pull")
	}
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
