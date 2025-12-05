# Getting Started

## Installation

### Using Go

```bash
go install github.com/adrianpk/snapfig@latest
```

### From Source

```bash
git clone https://github.com/adrianpk/snapfig
cd snapfig
go build -o snapfig .
```

## Your First Backup

1. Launch the TUI:
   ```bash
   snapfig
   ```

2. Navigate with arrow keys or `j/k`

3. Press `Space` on directories you want to backup (cycles through modes):
   - `[ ]` - Not selected
   - `[x]` - Selected, remove nested `.git` directories
   - `[g]` - Selected, preserve `.git` as `.git_disabled`

4. Press `F9` to open Settings and enter your git remote URL

5. Press `F7` to backup (copies to vault and pushes to remote)

## Restoring on a New Machine

1. Install snapfig

2. Launch and configure your remote:
   ```bash
   snapfig
   # Press F9, enter your remote URL
   ```

3. Press `F8` to sync (pulls from remote and restores files)

## Fire-and-Forget Setup

For automation or headless setup:

```bash
snapfig setup \
  --paths=".config/nvim:g,.zshrc:x,.bashrc:x" \
  --remote="git@github.com:user/dotfiles.git" \
  --copy-interval="1h" \
  --push-interval="24h"
```

This creates the config, runs initial copy, and starts the daemon.

---

**Next:** [CLI Reference](cli-reference.md)

**All docs:** [Index](index.md) · [CLI Reference](cli-reference.md) · [Background Runner](daemon.md) · [Workflows](workflows.md) · [Architecture](architecture.md)
