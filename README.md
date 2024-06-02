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

Snapfig provides various commands for managing your configuration files. Here are some common usage examples:

### Initialize Configuration

To initialize a configuration file, use the `scan` command:

```sh
snapfig scan
```

This command will scan the user's home directory by default. If you wish to specify an alternative directory as the root for the scan, use the --path option:

```sh
snapfig scan --path /alt/path
```

In both cases, the scan command will create a configuration file in the specified or default directory at `~/.config/snapfig/config.yml`.

### Start Versioning Process

[Not implemented yet]
