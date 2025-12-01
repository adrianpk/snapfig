# Persistence Model

## Directory Structure

```
~/
├── .config/
│   └── snapfig/
│       ├── config.yml        # User configuration (what to watch)
│       └── warnings.log      # Warning logs
│
└── .snapfig/
    └── vault/                # Replicated content (versioned)
        ├── .config/
        │   └── nvim/
        ├── .zsh/
        ├── Projects/
        │   └── dotfiles/
        └── ...
```

## Paths

| Purpose | Path | Description |
|---------|------|-------------|
| Config dir | `~/.config/snapfig/` | Snapfig's own configuration |
| Config file | `~/.config/snapfig/config.yml` | What directories to watch |
| Vault root | `~/.snapfig/` | Snapfig's data directory |
| Vault content | `~/.snapfig/vault/` | Replicated files (git repo) |

## Config File Format (config.yml)

```yaml
git: disable  # Global git mode: "disable" or "remove"

watching:
  - path: .config/nvim
    enabled: true
  - path: .config/doom
    enabled: true
    git: remove        # Override global git mode
  - path: .zsh
    enabled: true
  - path: Projects/dotfiles
    enabled: false     # Disabled, not synced
  - path: dev/scripts
    enabled: true
```

## Git Modes

How nested `.git` directories are handled when copying to vault:

- **disable**: Rename `.git` to `.git_disabled` (preserves history, avoids submodule issues)
- **remove**: Delete `.git` entirely (clean copy, no git history)

## Data Flow

```
1. User selects paths (TUI or edit YAML)
           ↓
2. Selection saved to ~/.config/snapfig/config.yml
           ↓
3. snapfig copy/sync reads config
           ↓
4. Copies enabled paths to ~/.snapfig/vault/
           ↓
5. Git commits changes in vault (behind the scenes)
           ↓
6. Optional: push to remote
```

## Vault as Git Repo

The vault directory (`~/.snapfig/vault/`) is itself a git repository:

- Initialized automatically on first copy
- Each sync creates a commit
- User can configure remote for backup
- User never interacts with git directly (Snapfig handles it)

## Restoration Flow

```
1. Fresh system: install git, install snapfig
           ↓
2. Clone vault: git clone <remote> ~/.snapfig/vault
           ↓
3. Run: snapfig restore
           ↓
4. Snapfig reads vault structure
           ↓
5. Copies files back to original locations (~/.config/nvim, etc.)
```

## Path Storage

Paths in config.yml are **relative to home directory**:

- `.config/nvim` → `~/.config/nvim`
- `.zsh` → `~/.zsh`
- `Projects/dotfiles` → `~/Projects/dotfiles`

This makes configs portable across systems.
