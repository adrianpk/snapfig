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
snapfig tui
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
| `F7` | **Backup** (Copy + Push in one step) |
| `F8` | **Sync** (Pull + Restore in one step) |
| `F9` | Settings (configure remote URL) |
| `F10` | Quit |

### Selection modes

- `[ ]` = Not selected
- `[x]` = Selected, remove `.git` directories in backup
- `[g]` = Selected, preserve `.git` as `.git_disabled` (keeps history)

## Typical workflows

### On your main machine

1. `snapfig tui`
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
snapfig tui       # Interactive interface
snapfig copy      # Copy to vault
snapfig push      # Push to remote
snapfig pull      # Pull from remote (clones if needed)
snapfig restore   # Restore from vault
```

## How it handles git

Many config directories (nvim, doom emacs, etc.) are git repos themselves. Snapfig handles this:

- **remove** (`[x]`): Deletes `.git` in the vault copy. Clean backup, no nested repos.
- **disable** (`[g]`): Renames `.git` to `.git_disabled`. On restore, reverts back to `.git`.

Original directories are never modified.

## Files

```
~/.config/snapfig/
└── config.yml          # Configuration (paths to watch, remote URL)

~/.snapfig/
├── vault/              # Your backed up files (git repo)
└── manifest.md         # Summary of what's backed up
```

## Restore safety

Before overwriting existing files, Snapfig backs them up:

```
~/.config/nvim → ~/.config/nvim.202412011530.bak
```

If something goes wrong, delete the restored copy and rename the `.bak` back.

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
```

Paths are relative to home directory.

## Planned Improvements

- Smart copy: copy only updated files within directories instead of replicating entire directory structures.
- Selective restore: Allow restoring only specific dotfiles instead of restoring everything.
- Improve and polish the interface.
- Background job for periodic snapshots.
- Token-based authentication for git cloud services.
- Add automated tests.

## License

MIT
