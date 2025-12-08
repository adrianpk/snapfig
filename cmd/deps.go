// Package cmd implements the CLI commands.
package cmd

import (
	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
)

// ServiceFactory creates a Service from config and path.
// This can be replaced in tests for mocking.
var ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
	return snapfig.NewService(cfg, configPath)
}

// ConfigLoader loads config from a path.
// This can be replaced in tests for mocking.
var ConfigLoader = config.Load

// DefaultConfigDirFunc returns the default config directory.
// This can be replaced in tests for mocking.
var DefaultConfigDirFunc = config.DefaultConfigDir

// HasRemoteFunc checks if a vault directory has a git remote configured.
// This can be replaced in tests for mocking.
var HasRemoteFunc = snapfig.HasRemote

// PidFilePathFunc returns the path to the daemon PID file.
// This can be replaced in tests for mocking.
var PidFilePathFunc = config.PidFilePath

// LogFilePathFunc returns the path to the daemon log file.
// This can be replaced in tests for mocking.
var LogFilePathFunc = config.LogFilePath

// DefaultSnapfigDirFunc returns the default snapfig directory.
// This can be replaced in tests for mocking.
var DefaultSnapfigDirFunc = config.DefaultSnapfigDir

// resetDeps resets all dependencies to their defaults.
// Used in tests to ensure clean state.
func resetDeps() {
	ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
		return snapfig.NewService(cfg, configPath)
	}
	ConfigLoader = config.Load
	DefaultConfigDirFunc = config.DefaultConfigDir
	HasRemoteFunc = snapfig.HasRemote
	PidFilePathFunc = config.PidFilePath
	LogFilePathFunc = config.LogFilePath
	DefaultSnapfigDirFunc = config.DefaultSnapfigDir
}
