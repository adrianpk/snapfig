# Snapfig

Snapfig is a tool designed to manage and version locations within your system. While the feature set is a floating target, the current behavior is as follows:

- Snapfig scans the user's home directory (or an alternative directory specified by the user) to create a list of locations that are typically backed up and/or versioned.
- This list is editable, allowing the user to add or remove paths. Initially, this process will be text-based (using YAML), but a future CLI tool will streamline this task.
- Snapfig watches these locations for changes and replicates them at a predetermined or configurable frequency to a specific location.
- All changes are automatically versioned.
- In the future, Snapfig will be able to recover the latest or a specific version from local storage or a remote repository. This feature will be useful for setting up a complete machine from scratch.

Please note that this README should be seen as a declaration of intent and a placeholder guide for defining future features. Snapfig is currently not ready for use.

## Notes

Please note that this README should be seen as a declaration of intent and a placeholder guide for defining future features. Snapfig is currently not ready for use.


## Usage

To use Snapfig, you can run the `snapfig` command followed by a subcommand. Currently, the available subcommands are `scan`, `copy`, and others.

```bash
snapfig scan
snapfig copy
```

### Scan Command

The `scan` command is used to scan files and perform operations based on the provided flags.

#### Git Flag

The `git` flag modifies how the `scan` command handles Git repositories in the scanned files. It can have two values: `remove` and `disable`.

- `remove`: The `scan` command will remove all Git related setup (.git directories of main repo and submodules) in the copied version of the directory stored in the Snapfig vault. The original directory remains untouched, preserving its Git setup.

- `disable`: The `scan` command will preserve the Git setup in the copied version of the directory stored in the Snapfig vault, but the .git directories (main and submodules) will be renamed to `.git_disabled`. Again, the original directory remains untouched.

This gives you the freedom to preserve the Git version control in the original directory if needed, while still allowing Snapfig to operate on a version of the directory without Git interference.

Example usage:

```bash
snapfig scan --git=remove
snapfig scan --git=disable
```

In both cases, the scan command will create a configuration file in the specified or default directory at `~/.config/snapfig/config.yml`.

### Start Versioning Process

[Not implemented yet]
