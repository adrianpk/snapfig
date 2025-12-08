// Package cmd implements the CLI commands.
package cmd

import (
	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
)

// ServiceFactory creates a Service from config and path.
var ServiceFactory = func(cfg *config.Config, configPath string) (snapfig.Service, error) {
	return snapfig.NewService(cfg, configPath)
}

// ConfigLoader loads config from a path.
var ConfigLoader = config.Load

// DefaultConfigDirFunc returns the default config directory.
var DefaultConfigDirFunc = config.DefaultConfigDir

// HasRemoteFunc checks if a vault directory has a git remote configured.
var HasRemoteFunc = snapfig.HasRemote

// PidFilePathFunc returns the path to the daemon PID file.
var PidFilePathFunc = config.PidFilePath

// LogFilePathFunc returns the path to the daemon log file.
var LogFilePathFunc = config.LogFilePath

// DefaultSnapfigDirFunc returns the default snapfig directory.
var DefaultSnapfigDirFunc = config.DefaultSnapfigDir

// resetDeps resets all dependencies to their defaults.
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
