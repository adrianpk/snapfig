# Snapfig Documentation

Snapfig backs up and restores your dotfiles straight from their original locations.

## Quick Links

- [Getting Started](getting-started.md) - Installation and your first backup
- [CLI Reference](cli-reference.md) - All commands and flags
- [Background Runner](daemon.md) - Automated backups with the daemon
- [Workflows](workflows.md) - Common use cases and recipes
- [Architecture](architecture.md) - How Snapfig works internally

## Overview

Snapfig copies your configuration files to a local vault (`~/.snapfig/vault/`) that's automatically versioned with git. Unlike symlink-based tools, you keep working with files in their original locations. Snapfig mirrors them on demand or automatically via the background runner.

### Key Features

- **Real copies, not symlinks** - Actual redundancy, not just pointers
- **Git versioning** - Every backup is a commit, full history preserved
- **Background sync** - Set intervals for copy/push/pull operations
- **Smart copy** - Only changed files are copied
- **Selective restore** - Choose exactly which files to restore
- **Nested git handling** - Preserves or removes `.git` dirs in config repos

## Getting Help

```bash
snapfig --help
snapfig <command> --help
```

---

**Next:** [Getting Started](getting-started.md)

**All docs:** [Getting Started](getting-started.md) · [CLI Reference](cli-reference.md) · [Background Runner](daemon.md) · [Workflows](workflows.md) · [Architecture](architecture.md)
