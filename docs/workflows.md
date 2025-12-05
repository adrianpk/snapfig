# Workflows

## Daily Use

### Main Machine Setup

1. Launch: `snapfig`
2. Select directories with `Space`
3. Configure remote: `F9`
4. Backup: `F7`

### New Machine Setup

1. Install snapfig
2. Launch and set remote: `F9`
3. Sync everything: `F8`

### Regular Backups

Just press `F7` (or run `snapfig copy && snapfig push`).

## CLI-Only Workflow

Don't want the TUI? Use commands directly:

```bash
# Initial setup
snapfig setup \
  --paths=".config/nvim:g,.zshrc:x,.bashrc:x" \
  --remote="git@github.com:user/dotfiles.git" \
  --no-daemon

# Manual backup
snapfig copy && snapfig push

# Restore on new machine
snapfig pull && snapfig restore
```

## Cron Instead of Daemon

If you prefer cron over the daemon:

```bash
# Edit crontab
crontab -e
```

```cron
# Copy every hour
0 * * * * /usr/local/bin/snapfig copy

# Push every day at 3am
0 3 * * * /usr/local/bin/snapfig push
```

## Multi-Machine Sync

### Recommended: One Primary

- **Primary machine**: Daemon with copy + push
- **Secondary machines**: Manual pull when needed

### Advanced: Bidirectional

Configure all machines with pull interval:

```yaml
daemon:
  copy_interval: 1h
  push_interval: 6h
  pull_interval: 6h    # Enable pull
  auto_restore: true   # Auto-apply changes
```

**Warning**: Can cause conflicts if same files change on multiple machines.

## Selective Restore

Restore only specific files:

1. Press `F6` in TUI
2. Navigate and select files to restore
3. Confirm

## Preserving Git History

For config directories that are git repos (nvim, doom emacs):

1. Select with `[g]` mode (not `[x]`)
2. `.git` is renamed to `.git_disabled` in vault
3. On restore, `.git_disabled` becomes `.git` again

This preserves your config repo's history.

## External Backup Drive

Configure vault on external drive:

```yaml
vault_path: /mnt/backup/dotfiles
```

Or via setup:

```bash
snapfig setup \
  --paths=".config/nvim:g,.zshrc:x" \
  --vault-path="/mnt/backup/dotfiles"
```

The daemon will copy to the external location automatically.

---

**Next:** [Architecture](architecture.md)

**All docs:** [Index](index.md) · [Getting Started](getting-started.md) · [CLI Reference](cli-reference.md) · [Background Runner](daemon.md) · [Architecture](architecture.md)
