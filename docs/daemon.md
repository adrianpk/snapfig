# Background Runner

The daemon runs scheduled backups in the background without manual intervention.

## Commands

```bash
snapfig daemon start    # Start the daemon
snapfig daemon stop     # Stop the daemon
snapfig daemon status   # Show status and configuration
```

## Configuration

Configure via **Settings (F9)** in the TUI, or edit `~/.config/snapfig/config.yml`:

```yaml
daemon:
  copy_interval: 1h      # How often to copy to vault
  push_interval: 24h     # How often to push to remote
  pull_interval: ""      # How often to pull (empty = disabled)
  auto_restore: false    # Restore after pull
```

## Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `copy_interval` | Runs smart copy at this interval. Only changed files are copied. | `30m`, `1h`, `2h` |
| `push_interval` | Pushes vault to remote. Requires `remote` configured. | `12h`, `24h` |
| `pull_interval` | Pulls from remote. **Disabled by default.** | `24h` |
| `auto_restore` | Automatically restores after pull. **Use carefully.** | `true`, `false` |

Intervals use Go duration format: `30s`, `15m`, `1h`, `24h`.

## Logs

Activity is logged to `~/.snapfig/daemon.log`:

```bash
tail -f ~/.snapfig/daemon.log
```

Example output:
```
[snapfig] 2025/12/03 11:33:40 Copy started
[snapfig] 2025/12/03 11:33:40 Copy done: 1 paths, 2 updated, 3 unchanged, 0 removed
[snapfig] 2025/12/03 11:33:40   copied: .config/nvim
```

## Persistence

The daemon runs as a foreground process. To keep it running:

### Shell RC (simple)

Add to `.bashrc` or `.zshrc`:

```bash
pgrep -f "snapfig daemon" || snapfig daemon start &
```

### Systemd (recommended)

Create `~/.config/systemd/user/snapfig.service`:

```ini
[Unit]
Description=Snapfig background runner

[Service]
ExecStart=/usr/local/bin/snapfig daemon start
Restart=on-failure

[Install]
WantedBy=default.target
```

Enable:

```bash
systemctl --user enable snapfig
systemctl --user start snapfig
```

## Caution with Pull/Auto-Restore

- `pull_interval` is disabled by default for safety
- On multi-machine setups, pulling can overwrite local changes
- `auto_restore: true` restores immediately after pull
- Consider your workflow before enabling these options

---

**Next:** [Workflows](workflows.md)

**All docs:** [Index](index.md) · [Getting Started](getting-started.md) · [CLI Reference](cli-reference.md) · [Workflows](workflows.md) · [Architecture](architecture.md)
