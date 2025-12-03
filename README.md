# Snapfig

Backup and restore your dotfiles straight from their original locations.

<p align="center">
  <img src="docs/img/snapfig.png" alt="Snapfig TUI" width="700">
</p>

> **Note:** This is the current version of Snapfig. The [original version](https://github.com/adrianpk/snapfig-deprecated) relied on a local webserver and browser-based interface. This version uses [Bubble Tea](https://github.com/charmbracelet/bubbletea) and runs directly in the console.

---

## What it does

Snapfig copies your configuration files to a local vault (`~/.snapfig/vault/`) that's automatically versioned with git. Unlike tools that use symlinks, you keep working with your files in their original locations. Snapfig mirrors them when you ask.

## Install

```bash
go install github.com/adrianpk/snapfig@latest
```

Or build from source:

```bash
git clone https://github.com/adrianpk/snapfig
cd snapfig
go build -o snapfig .
```

## Quick start

```bash
snapfig
```

1. Navigate with arrow keys or `j/k`
2. `Space` to select paths (cycles through modes)
3. `F9` to configure your remote repository
4. `F7` to backup (copy + push)

## TUI Controls

| Key | Action |
|-----|--------|
| `Space` | Cycle selection: `[ ]` → `[x]` → `[g]` → `[ ]` |
| `F2` | Copy to vault |
| `F3` | Push to remote |
| `F4` | Pull from remote (clones if vault doesn't exist) |
| `F5` | Restore from vault |
| `F6` | **Selective restore** (choose specific files to restore) |
| `F7` | **Backup** (Copy + Push in one step) |
| `F8` | **Sync** (Pull + Restore in one step) |
| `F9` | Settings (remote URL, background runner intervals) |
| `F10` | Quit |

### Selection modes

- `[ ]` = Not selected
- `[x]` = Selected, remove `.git` directories in backup
- `[g]` = Selected, preserve `.git` as `.git_disabled` (keeps history)

## Typical workflows

### On your main machine

1. `snapfig` or `snapfig tui`
2. Select your config directories
3. `F9` → enter your git remote URL
4. `F7` → backs up and pushes

### On a new machine

1. Copy your `~/.config/snapfig/config.yml` (or just set the remote via F9)
2. `snapfig tui`
3. `F8` → clones vault and restores everything

### Regular backups

Just press `F7`. It copies changes and pushes to remote.

## CLI Commands

```bash
snapfig                 # Interactive interface (default)
snapfig copy            # Copy to vault
snapfig push            # Push to remote
snapfig pull            # Pull from remote (clones if needed)
snapfig restore         # Restore from vault
snapfig daemon start    # Start background runner
snapfig daemon stop     # Stop background runner
snapfig daemon status   # Show background runner status
```

## How it handles git

Many config directories (nvim, doom emacs, etc.) are git repos themselves. Snapfig handles this:

- **remove** (`[x]`): Deletes `.git` in the vault copy. Clean backup, no nested repos.
- **disable** (`[g]`): Renames `.git` to `.git_disabled`. On restore, reverts back to `.git`.

Original directories are never modified.

## Background Runner

The daemon runs scheduled backups in the background:

```bash
snapfig daemon start    # Start (logs to ~/.snapfig/daemon.log)
snapfig daemon status   # Check if running
snapfig daemon stop     # Stop
```

Configure intervals in Settings (F9) or directly in `config.yml`. The daemon uses smart copy, so only changed files are copied.

## Files

```
~/.config/snapfig/
└── config.yml          # Configuration (paths to watch, remote URL, daemon settings)

~/.snapfig/
├── vault/              # Your backed up files (git repo)
├── manifest.md         # Summary of what's backed up
├── daemon.pid          # PID file when daemon is running
└── daemon.log          # Daemon activity log
```

## Config format

```yaml
git: disable
remote: git@github.com:user/dotfiles.git

watching:
  - path: .config/nvim
    git: disable
    enabled: true
  - path: .zshrc
    git: remove
    enabled: true

daemon:
  copy_interval: 1h      # Smart copy every hour
  push_interval: 24h     # Push to remote daily
  pull_interval: ""      # Disabled by default
  auto_restore: false    # Auto restore after pull
```

Paths are relative to home directory. Intervals use Go duration format (e.g. `30m`, `1h`, `24h`).

## Planned Improvements

- ~~Smart copy: copy only updated files within directories instead of replicating entire directory structures.~~
- ~~Selective restore: Allow restoring only specific dotfiles instead of restoring everything.~~
- ~~Background runner for periodic snapshots.~~
- Improve and polish the interface.
- Token-based authentication for git cloud services.
- Add automated tests.

## License

MIT
