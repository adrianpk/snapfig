# Snapfig User Guide

This is the complete guide to using Snapfig. It covers installation, backup workflows, restoration on new machines, and troubleshooting.

---

## What is Snapfig?

Snapfig copies your configuration files to a versioned vault. Unlike symlink-based dotfile managers, your files stay in their original locations. Snapfig mirrors them to `~/.snapfig/vault/`, which is a Git repository you can push to a remote for backup and sync.

| Feature | Description |
|---------|-------------|
| No symlinks | Files stay where applications expect them |
| Git versioning | Every backup is a commit with full history |
| Smart copy | Only changed files are copied |
| Selective restore | Choose exactly which files to restore |
| Background sync | Daemon runs copy/push/pull on a schedule |
| Nested git handling | Preserves or removes `.git` directories in config repos |

---

## Installation

### Using Go

```bash
go install github.com/adrianpk/snapfig@latest
```

Requires Go 1.21 or later. The binary is installed to `$GOPATH/bin/` (usually `~/go/bin/`).

### From Source

```bash
git clone https://github.com/adrianpk/snapfig
cd snapfig
go build -o snapfig .
sudo mv snapfig /usr/local/bin/
```

### Verify Installation

```bash
snapfig --help
```

---

## Directory Structure

After first use, Snapfig creates this structure:

```
~/
├── .config/
│   └── snapfig/
│       └── config.yml         # Your configuration
│
└── .snapfig/
    ├── vault/                 # Replicated files (git repo)
    │   ├── .config/
    │   │   └── nvim/
    │   ├── .zshrc
    │   └── ...
    ├── manifest.yml           # List of backed-up paths
    ├── daemon.pid             # PID when daemon is running
    └── daemon.log             # Daemon activity log
```

| Path | Purpose |
|------|---------|
| `~/.config/snapfig/config.yml` | Stores which paths to watch and daemon settings |
| `~/.snapfig/vault/` | The Git repository containing copies of your files |
| `~/.snapfig/manifest.yml` | Records which paths are backed up (used for sync on new machines) |
| `~/.snapfig/daemon.log` | Log file for background operations |

---

## Workflow 1: Creating a Backup (Main Machine)

This workflow sets up Snapfig on your primary machine and creates your first backup.

### Step 1: Launch Snapfig

```bash
snapfig
```

This opens the interactive TUI (Terminal User Interface).

### Step 2: Select Directories to Watch

Navigate the tree with arrow keys or `j/k`. Press `Space` on each directory you want to backup.

The checkbox cycles through three states:

| State | Symbol | Meaning |
|-------|--------|---------|
| Not selected | `[ ]` | Directory is not backed up |
| Remove git | `[x]` | Backup, delete `.git` directories in the vault copy (original untouched) |
| Disable git | `[g]` | Backup, rename `.git` to `.git_disabled` in the vault copy (original untouched) |

**Important:** Snapfig never modifies your original files. All `.git` handling happens only in the vault copy.

**Example selection:**

```
> [g] .config/nvim/           ★    # Preserve git (has lazy.nvim repo)
  [x] .config/alacritty/           # Remove git (no nested repo)
  [x] .zshrc                       # Remove git (it's a file)
  [ ] .config/Code/                # Not selected (too large)
```

Use `[g]` mode for directories that are themselves Git repositories (nvim with plugin managers, doom emacs, etc.). Use `[x]` for everything else.

### Step 3: Configure Git Remote

Press `F9` to open Settings.

Enter your Git remote URL:

```
Remote URL: git@github.com:username/dotfiles.git
```

Other settings:

| Setting | Default | Description |
|---------|---------|-------------|
| Git token | (empty) | For HTTPS auth. Leave empty to use SSH. |
| Vault location | `~/.snapfig/vault` | Where files are copied |
| Copy interval | `1h` | Daemon copies every hour |
| Push interval | `24h` | Daemon pushes daily |
| Pull interval | (disabled) | How often to pull (leave empty) |
| Auto restore | `false` | Restore after pull |

Press `Enter` to save, `Esc` to cancel.

### Step 4: Run Backup

Press `F7` (Backup) to:

1. Copy selected paths to the vault
2. Commit changes to Git
3. Push to the configured remote

The status bar shows progress:

```
Backup: 3 updated, 12 unchanged, 0 removed, pushed
```

### Step 5: Start the Daemon (Optional)

If you want automatic backups:

```bash
snapfig daemon start
```

Check status:

```bash
snapfig daemon status
```

Output:

```
Daemon is running (pid 12345)
  Copy interval: 1h
  Push interval: 24h
```

### What Happens During Copy

```
1. Snapfig reads ~/.config/snapfig/config.yml
           ↓
2. For each watched path:
   - Compare source file with vault copy (modtime + size)
   - Skip if identical
   - Copy if changed
           ↓
3. Handle .git directories:
   - [x] mode: delete .git
   - [g] mode: rename .git → .git_disabled
           ↓
4. Update manifest.yml with backed-up paths
           ↓
5. Git commit: "Snapfig backup YYYY-MM-DD HH:MM"
```

### CLI Alternative

If you prefer the command line:

```bash
# One-shot setup
snapfig setup \
  --paths=".config/nvim:g,.zshrc:x,.bashrc:x,.config/alacritty:x" \
  --remote="git@github.com:username/dotfiles.git" \
  --copy-interval="1h" \
  --push-interval="24h"

# Manual backup (any time)
snapfig copy && snapfig push
```

Path format: `path:mode` where mode is `x` (remove .git) or `g` (disable .git).

---

## Workflow 2: Restoring on a New Machine

This workflow sets up Snapfig on a fresh machine and restores your configuration from the remote repository.

### Prerequisites

- Git installed and configured with SSH keys (or token for HTTPS)
- Snapfig installed (see Installation above)

### Step 1: Launch Snapfig

```bash
snapfig
```

The picker shows your home directory (empty or default configs).

### Step 2: Configure Remote

Press `F9` to open Settings.

Enter the same remote URL you used on your main machine:

```
Remote URL: git@github.com:username/dotfiles.git
```

Press `Enter` to save.

### Step 3: Sync

Press `F8` (Sync) to:

1. Clone the vault from the remote (or pull if vault exists)
2. Read the manifest to discover which paths were backed up
3. Restore all paths to their original locations

The status bar shows:

```
Sync: cloned, 3 updated, 0 unchanged
```

### What Happens During Sync

```
1. Check if vault exists at ~/.snapfig/vault/
           ↓
   If NO:  Clone from remote URL
   If YES: Git pull
           ↓
2. Read manifest.yml from vault
           ↓
3. Update config.yml with paths from manifest
   (This is how new machine "learns" what to watch)
           ↓
4. For each path in manifest:
   - Check if source exists in vault
   - Backup existing local file with timestamp (.bak)
   - Copy from vault to ~/
           ↓
5. Handle .git_disabled:
   - Rename back to .git
```

### Step 4: Verify

Your files are now in place:

```bash
ls -la ~/.config/nvim/
ls -la ~/.zshrc
```

The TUI now shows your paths as selected with sync status:

```
> [g] .config/nvim/           [synced]
  [x] .zshrc                  [synced]
```

### Step 5: Start Daemon (Optional)

```bash
snapfig daemon start
```

Now this machine will also backup on schedule.

### Selective Restore

If you only want specific files:

1. Press `F6` (Selective Restore)
2. Navigate and select files with `Space`
3. Press `Enter` to restore only those files

### CLI Alternative

```bash
# Configure remote
snapfig setup \
  --paths="" \
  --remote="git@github.com:username/dotfiles.git" \
  --no-daemon

# Pull and restore
snapfig pull && snapfig restore
```

Or with sync command through TUI equivalent:

```bash
# After pull, the manifest populates config automatically
snapfig pull
snapfig restore
```

---

## Configuration Reference

### Config File Location

```
~/.config/snapfig/config.yml
```

### Full Example

```yaml
git: disable                          # Global git mode
remote: git@github.com:user/dotfiles.git
git_token: ""                         # For HTTPS auth
vault_path: ""                        # Custom vault location

watching:
  - path: .config/nvim
    git: disable                      # Override global mode
    enabled: true
  - path: .zshrc
    git: remove
    enabled: true
  - path: .config/alacritty
    git: remove
    enabled: true

daemon:
  copy_interval: 1h
  push_interval: 24h
  pull_interval: ""                   # Disabled
  auto_restore: false
```

### Git Modes

These modes control how `.git` directories are handled **in the vault copy only**. Your original files are never modified.

| Mode | Config Value | Effect on Copy (vault) | Effect on Restore |
|------|--------------|------------------------|-------------------|
| Remove | `remove` | Deletes `.git` in vault copy | No action |
| Disable | `disable` | Renames `.git` → `.git_disabled` in vault | Renames `.git_disabled` → `.git` |

**Why this exists:** The vault itself is a Git repository. Some config directories (like neovim with plugin managers) contain `.git` subdirectories. Without handling them, Git would see these as submodules, complicating the vault. Renaming to `.git_disabled` keeps the vault clean while preserving the nested repos for restore.

### Daemon Intervals

Format: Go duration (`30s`, `15m`, `1h`, `24h`)

| Setting | Description | Recommended |
|---------|-------------|-------------|
| `copy_interval` | How often to copy to vault | `1h` |
| `push_interval` | How often to push to remote | `24h` |
| `pull_interval` | How often to pull from remote | Disabled (empty) |
| `auto_restore` | Restore automatically after pull | `false` |

**Warning:** Enabling `pull_interval` and `auto_restore` on multiple machines can cause conflicts.

---

## TUI Reference

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `h` | Collapse directory / go to parent |
| `→` / `l` / `Enter` | Expand directory |
| `Space` | Cycle selection: `[ ]` → `[x]` → `[g]` |
| `a` | Select all (remove mode) |
| `n` | Deselect all |

### Actions

| Key | Action | CLI Equivalent |
|-----|--------|----------------|
| `F2` | Copy to vault | `snapfig copy` |
| `F3` | Push to remote | `snapfig push` |
| `F4` | Pull from remote | `snapfig pull` |
| `F5` | Restore all | `snapfig restore` |
| `F6` | Selective restore | (TUI only) |
| `F7` | Backup (copy + push) | `snapfig copy && snapfig push` |
| `F8` | Sync (pull + restore) | `snapfig pull && snapfig restore` |
| `F9` | Settings | (TUI only) |
| `F10` / `Ctrl+C` | Quit | - |

### Sync Status Tags

| Tag | Meaning |
|-----|---------|
| `[synced]` | Path exists locally, in vault, and in manifest |
| `[backup]` | Path exists locally and in manifest, but not in vault |
| `[restore]` | Path exists in vault and manifest, but not locally |
| `[orphan]` | Path only exists in manifest |
| (no tag) | Path is not tracked in manifest |

---

## Daemon Persistence

The daemon runs as a foreground process. To keep it running after reboot:

### Option 1: Shell RC (Simple)

Add to `~/.bashrc` or `~/.zshrc`:

```bash
snapfig daemon start 2>/dev/null
```

### Option 2: Systemd User Service (Recommended)

Create `~/.config/systemd/user/snapfig.service`:

```ini
[Unit]
Description=Snapfig background runner

[Service]
ExecStart=/usr/local/bin/snapfig daemon run
Restart=on-failure

[Install]
WantedBy=default.target
```

Enable and start:

```bash
systemctl --user enable --now snapfig
```

Check status:

```bash
systemctl --user status snapfig
```

View logs:

```bash
journalctl --user -u snapfig -f
```

---

## Troubleshooting

### "No remote configured"

```
Error: no remote configured. Run 'snapfig' and configure in Settings (F9)
```

**Solution:** Open TUI, press `F9`, enter your Git remote URL.

### Push fails with authentication error

```
Error: authentication required
```

**Solutions:**

1. **SSH:** Ensure your SSH key is added to ssh-agent:
   ```bash
   eval "$(ssh-agent -s)"
   ssh-add ~/.ssh/id_ed25519
   ```

2. **HTTPS:** Add a Git token in Settings (`F9`).

### Restore overwrites local changes

By default, restore creates backups with timestamp suffix:

```
.zshrc.202501151230.bak
```

To recover:

```bash
mv ~/.zshrc.202501151230.bak ~/.zshrc
```

### Daemon not starting

Check if already running:

```bash
snapfig daemon status
```

Check logs:

```bash
tail -50 ~/.snapfig/daemon.log
```

Kill stale process:

```bash
rm ~/.snapfig/daemon.pid
snapfig daemon start
```

### "Config already exists"

When running `snapfig setup`:

```
Error: config already exists at ~/.config/snapfig/config.yml
```

**Solution:** Use `--force` to overwrite:

```bash
snapfig setup --paths="..." --remote="..." --force
```

### Vault is not a Git repository

```
Error: not a git repository
```

**Solution:** Initialize or re-clone:

```bash
rm -rf ~/.snapfig/vault
snapfig pull
```

### Files not syncing on new machine

After `F8` (Sync), paths show `[restore]` but nothing happens.

**Cause:** Config was not updated from manifest.

**Solution:** The manifest sync happens automatically on pull when `config.yml` has no watched paths. If you manually added paths before syncing, remove them:

```bash
rm ~/.config/snapfig/config.yml
snapfig pull
snapfig restore
```

---

## Common Workflows

### Daily Backup (Main Machine)

Press `F7` or:

```bash
snapfig copy && snapfig push
```

### Update Configs on Secondary Machine

Press `F8` or:

```bash
snapfig pull && snapfig restore
```

### Add New Path to Backup

1. Navigate to path in TUI
2. Press `Space` to select
3. Press `F7` to backup

### Remove Path from Backup

1. Navigate to path in TUI
2. Press `Space` until `[ ]`
3. Press `F2` to update vault (removes from manifest)

### View What's Backed Up

```bash
cat ~/.snapfig/manifest.yml
```

Or check the vault directly:

```bash
ls -la ~/.snapfig/vault/
```

### View Backup History

```bash
cd ~/.snapfig/vault
git log --oneline
```

---

## Quick Reference Card

| Task | TUI | CLI |
|------|-----|-----|
| Select paths | `Space` | `--paths` in setup |
| Copy to vault | `F2` | `snapfig copy` |
| Push to remote | `F3` | `snapfig push` |
| Pull from remote | `F4` | `snapfig pull` |
| Restore all | `F5` | `snapfig restore` |
| Backup (copy+push) | `F7` | `snapfig copy && snapfig push` |
| Sync (pull+restore) | `F8` | `snapfig pull && snapfig restore` |
| Settings | `F9` | Edit `~/.config/snapfig/config.yml` |
| Start daemon | - | `snapfig daemon start` |
| Stop daemon | - | `snapfig daemon stop` |
