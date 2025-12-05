# Snapfig

[![Tests](https://github.com/adrianpk/snapfig/actions/workflows/test.yml/badge.svg)](https://github.com/adrianpk/snapfig/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/adrianpk/snapfig/branch/main/graph/badge.svg)](https://codecov.io/gh/adrianpk/snapfig)

Backup and restore your dotfiles straight from their original locations.

<p align="center">
  <img src="docs/img/snapfig.png" alt="Snapfig TUI" width="700">
</p>

## What It Is

Snapfig copies your configuration files to a local vault (`~/.snapfig/vault/`) versioned with git. Unlike symlink-based tools, files stay in their original locations—Snapfig mirrors them on demand.

**Why not symlinks?** Real copies mean real redundancy. If originals break, you have actual backups.

## Two Ways to Use It

### Interactive (TUI)

```bash
snapfig
```

Navigate with arrows, `Space` to select, `F7` to backup, `F8` to sync.

### Command Line

The TUI is optional. All operations work from the command line:

```bash
snapfig copy              # Copy to vault
snapfig push              # Push to remote
snapfig pull              # Pull from remote
snapfig restore           # Restore from vault
```

Or fire-and-forget setup for scripting:

```bash
snapfig setup \
  --paths=".config/nvim:g,.zshrc:x,.bashrc:x" \
  --remote="git@github.com:user/dotfiles.git"
```

## Automation

### Background Runner

```bash
snapfig daemon start
```

Configurable intervals for copy, push, pull. See [daemon docs](docs/daemon.md).

### Or Just Cron

The daemon is optional. Prefer cron? Use it:

```cron
0 * * * * snapfig copy
0 3 * * * snapfig push
```

## Git Handling

Config directories that are git repos (nvim, doom emacs, etc.):

- `[x]` mode: Removes `.git` in vault (clean copy)
- `[g]` mode: Renames `.git` to `.git_disabled` (preserves history)

Originals are never modified.

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

## Documentation

- [Getting Started](docs/getting-started.md)
- [CLI Reference](docs/cli-reference.md)
- [Background Runner](docs/daemon.md)
- [Workflows](docs/workflows.md)
- [Architecture](docs/architecture.md)

## Planned Improvements

- ~~Smart copy~~
- ~~Selective restore~~
- ~~Background runner~~
- ~~Fire-and-forget setup command~~
- ~~Alternative vault location~~
- ~~Automated tests~~
- ~~Token-based authentication for git cloud services~~

## License

MIT
