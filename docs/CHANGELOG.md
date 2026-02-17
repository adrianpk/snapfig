# Changelog

All notable changes to snapfig will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3] - 2026-02-17

### Added

- Restore now creates real symlinks when the target exists, falling back to marker files when it doesn't
- Changelog documentation

## [0.1.2] - 2026-02-16

### Added

- Symlink handling: symlinks are preserved as marker files during backup
- User guide documentation

## [0.1.1] - 2026-02-15

### Added

- Sync status tags in the TUI picker showing which items need syncing
- Manifest-based sync detection to track what has been backed up

## [0.1.0] - 2026-02-14

Initial release.

### Added

- Smart copy: only backs up files that have changed (by size/mtime)
- Smart restore: only restores files that differ from the vault
- Selective restore: choose specific files or directories to restore
- TUI mode as the default interface with intuitive file picker
- Background daemon for automatic periodic backups
- Fire-and-forget setup command for quick initialization
- Alternative vault location support (custom paths)
- Git authentication with token support and SSH fallback
- Git mode options: remove or disable `.git` directories in backups
