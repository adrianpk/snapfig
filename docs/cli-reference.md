# CLI Reference

## Commands

### `snapfig` (default)

Launches the interactive TUI.

```bash
snapfig
```

### `snapfig copy`

Copies watched paths to the vault. Only changed files are copied (smart copy).

```bash
snapfig copy
```

### `snapfig push`

Pushes the vault to the configured remote.

```bash
snapfig push
```

### `snapfig pull`

Pulls from remote. Clones the repository if the vault doesn't exist.

```bash
snapfig pull
```

### `snapfig restore`

Restores all files from vault to their original locations.

```bash
snapfig restore
```

### `snapfig daemon`

Manages the background runner.

```bash
snapfig daemon start   # Start the daemon
snapfig daemon stop    # Stop the daemon
snapfig daemon status  # Show status and configuration
```

See [Background Runner](daemon.md) for configuration details.

### `snapfig setup`

Fire-and-forget setup for automation.

```bash
snapfig setup \
  --paths=".config/nvim:g,.zshrc:x" \
  --remote="git@github.com:user/dotfiles.git"
```

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--paths` | Paths to watch (required). Format: `path:mode` | - |
| `--remote` | Git remote URL | - |
| `--vault-path` | Custom vault location | `~/.snapfig/vault` |
| `--copy-interval` | Copy interval | `1h` |
| `--push-interval` | Push interval | `24h` |
| `--pull-interval` | Pull interval | disabled |
| `--auto-restore` | Auto restore after pull | `false` |
| `--no-daemon` | Don't start daemon | `false` |
| `--force` | Overwrite existing config | `false` |

#### Path Format

Paths use `path:mode` format:
- `path:x` - Remove `.git` directories (default)
- `path:g` - Preserve `.git` as `.git_disabled`

Example: `.config/nvim:g,.zshrc:x,.bashrc`

## TUI Controls

| Key | Action |
|-----|--------|
| `Space` | Cycle selection: `[ ]` → `[x]` → `[g]` → `[ ]` |
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Expand/collapse directory |
| `F2` | Copy to vault |
| `F3` | Push to remote |
| `F4` | Pull from remote |
| `F5` | Restore from vault |
| `F6` | Selective restore |
| `F7` | Backup (copy + push) |
| `F8` | Sync (pull + restore) |
| `F9` | Settings |
| `F10` | Quit |

### Selection Modes

- `[ ]` - Not selected
- `[x]` - Selected, remove `.git` directories in backup
- `[g]` - Selected, preserve `.git` as `.git_disabled` (keeps history)

---

**Next:** [Background Runner](daemon.md)

**All docs:** [Index](index.md) · [Getting Started](getting-started.md) · [Background Runner](daemon.md) · [Workflows](workflows.md) · [Architecture](architecture.md)
