package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/daemon"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the background runner",
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background runner",
	RunE:  runDaemonStart,
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background runner",
	RunE:  runDaemonStop,
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show background runner status",
	RunE:  runDaemonStatus,
}

var daemonRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run daemon in foreground (internal use)",
	Hidden: true,
	RunE:   runDaemonForeground,
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonRunCmd)
	rootCmd.AddCommand(daemonCmd)
}

func runDaemonStart(cmd *cobra.Command, args []string) error {
	pid, running := getDaemonPid()
	if running {
		return fmt.Errorf("daemon already running (pid %d)", pid)
	}

	// Get paths
	logPath, err := config.LogFilePath()
	if err != nil {
		return err
	}

	snapfigDir, err := config.DefaultSnapfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(snapfigDir, 0755); err != nil {
		return err
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Start daemon process
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	proc := exec.Command(exe, "daemon", "run")
	proc.Stdout = logFile
	proc.Stderr = logFile
	proc.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session, detach from terminal
	}

	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("Daemon started (pid %d)\n", proc.Process.Pid)
	return nil
}

func runDaemonStop(cmd *cobra.Command, args []string) error {
	pid, running := getDaemonPid()
	if !running {
		fmt.Println("Daemon is not running")
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// Remove PID file
	pidPath, _ := config.PidFilePath()
	os.Remove(pidPath)

	fmt.Printf("Daemon stopped (pid %d)\n", pid)
	return nil
}

func runDaemonStatus(cmd *cobra.Command, args []string) error {
	pid, running := getDaemonPid()
	if !running {
		fmt.Println("Daemon is not running")
		return nil
	}

	fmt.Printf("Daemon is running (pid %d)\n", pid)

	// Load config to show intervals
	configDir, _ := config.DefaultConfigDir()
	configPath := filepath.Join(configDir, "config.yml")
	cfg, err := config.Load(configPath)
	if err == nil && cfg.Daemon.CopyInterval != "" {
		fmt.Printf("  Copy interval: %s\n", cfg.Daemon.CopyInterval)
		if cfg.Daemon.PushInterval != "" {
			fmt.Printf("  Push interval: %s\n", cfg.Daemon.PushInterval)
		}
		if cfg.Daemon.PullInterval != "" {
			fmt.Printf("  Pull interval: %s\n", cfg.Daemon.PullInterval)
			fmt.Printf("  Auto restore: %v\n", cfg.Daemon.AutoRestore)
		}
	}

	return nil
}

func runDaemonForeground(cmd *cobra.Command, args []string) error {
	configDir, err := config.DefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.yml")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	d, err := daemon.New(cfg, configPath)
	if err != nil {
		return err
	}

	return d.Run()
}

func getDaemonPid() (int, bool) {
	pidPath, err := config.PidFilePath()
	if err != nil {
		return 0, false
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, false
	}

	// Check if process is actually running
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}

	// On Unix, FindProcess always succeeds. Send signal 0 to check if alive.
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return 0, false
	}

	return pid, true
}
