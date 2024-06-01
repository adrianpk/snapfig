# Snapfig

The tool manages the concept of locations inside your system. Each one will have an associated repository and a frequency of update. Even the home directory will be able to be versioned, with a semi-flat structure including only the dotfiles and some user-specified directories (to avoid versioning the entire directory in a repository).

## Notes

Please note that this README should be seen as a declaration of intent and a placeholder guide for defining future features. Snapfig is currently not ready for use.

## Usage

Snapfig provides various commands for managing your configuration files. Here are some common usage examples:

### Initialize Configuration

To initialize a configuration file, use the `init` command:

```sh
snapfig init --config /path/to/config.yaml
```

This command will create a sample configuration file at the specified path.

### Run Versioning Process

To start the versioning process, use the `run` command:

```sh
snapfig run --config /path/to/config.yaml
```

This command will version the specified directories and create snapshots according to the configuration.
At some point, a daemon will keep the repositories updated.
