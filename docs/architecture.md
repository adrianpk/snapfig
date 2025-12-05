# Architecture

## Directory Structure

```
~/
├── .config/
│   └── snapfig/
│       └── config.yml      # User configuration
│
└── .snapfig/
    ├── vault/              # Replicated content (git repo)
    │   ├── .config/
    │   │   └── nvim/
    │   ├── .zsh/
    │   └── ...
    ├── manifest.md         # Summary of backed up files
    ├── daemon.pid          # PID when daemon is running
    └── daemon.log          # Daemon activity log
```

## Paths

| Purpose | Path | Description |
|---------|------|-------------|
| Config dir | `~/.config/snapfig/` | Snapfig's own configuration |
| Config file | `~/.config/snapfig/config.yml` | What directories to watch |
| Vault root | `~/.snapfig/` | Snapfig's data directory |
| Vault content | `~/.snapfig/vault/` | Replicated files (git repo) |

## Config File Format

```yaml
git: disable  # Global git mode: "disable" or "remove"
remote: git@github.com:user/dotfiles.git
vault_path: ""  # Optional custom vault location

watching:
  - path: .config/nvim
    enabled: true
    git: disable       # Override global mode
  - path: .zshrc
    enabled: true
    git: remove

daemon:
  copy_interval: 1h
  push_interval: 24h
  pull_interval: ""    # Disabled
  auto_restore: false
```

Paths are **relative to home directory**: `.config/nvim` → `~/.config/nvim`

## Git Modes

How nested `.git` directories are handled when copying to vault:

| Mode | What it does | Use case |
|------|--------------|----------|
| `disable` | Renames `.git` to `.git_disabled` | Preserves history, avoids submodule issues |
| `remove` | Deletes `.git` entirely | Clean copy, no git history |

On restore, `.git_disabled` is reverted back to `.git`.

## Data Flow

### Backup Flow

```
1. User selects paths (TUI or config.yml)
           ↓
2. Selection saved to ~/.config/snapfig/config.yml
           ↓
3. Copy reads config, compares files (smart copy)
           ↓
4. Changed files copied to ~/.snapfig/vault/
           ↓
5. Git commits changes automatically
           ↓
6. Push sends to remote (if configured)
```

### Restore Flow

```
1. Pull clones/updates vault from remote
           ↓
2. Restore reads vault structure
           ↓
3. Files copied back to original locations
           ↓
4. .git_disabled reverted to .git
```

## Smart Copy

Files are only copied when changed, determined by comparing:

1. **ModTime** - Last modification timestamp
2. **Size** - File size in bytes

If both match, the file is skipped. This approach is fast (no content hashing) and covers 99% of real-world cases.

## Vault as Git Repository

The vault (`~/.snapfig/vault/`) is itself a git repository:

- Initialized automatically on first copy
- Each copy creates a commit
- User can configure remote for backup
- User never interacts with git directly

## Alternative Vault Location

The vault can be placed anywhere:

```yaml
vault_path: /mnt/external/dotfiles-vault
```

Use cases:
- External drives for additional redundancy
- Network shares for centralized backups
- Any location you prefer

The path can be absolute or start with `~`.

---

**All docs:** [Index](index.md) · [Getting Started](getting-started.md) · [CLI Reference](cli-reference.md) · [Background Runner](daemon.md) · [Workflows](workflows.md)
