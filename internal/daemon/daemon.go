// Package daemon implements the background runner for scheduled copy/push/pull.
package daemon

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/adrianpk/snapfig/internal/snapfig"
)

// Daemon manages scheduled backup operations.
type Daemon struct {
	cfg          *config.Config
	configPath   string
	vaultDir     string
	copyInterval time.Duration
	pushInterval time.Duration
	pullInterval time.Duration
	logger       *log.Logger
}

// New creates a new Daemon instance.
func New(cfg *config.Config, configPath string) (*Daemon, error) {
	vaultDir, err := cfg.VaultDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault directory: %w", err)
	}

	d := &Daemon{
		cfg:        cfg,
		configPath: configPath,
		vaultDir:   vaultDir,
		logger:     log.New(os.Stdout, "[snapfig] ", log.LstdFlags),
	}

	if err := d.parseIntervals(); err != nil {
		return nil, err
	}

	return d, nil
}

// parseIntervals parses duration strings from config.
func (d *Daemon) parseIntervals() error {
	d.copyInterval = 0
	d.pushInterval = 0
	d.pullInterval = 0

	if d.cfg.Daemon.CopyInterval != "" {
		dur, err := time.ParseDuration(d.cfg.Daemon.CopyInterval)
		if err != nil {
			return fmt.Errorf("invalid copy_interval: %w", err)
		}
		d.copyInterval = dur
	}

	if d.cfg.Daemon.PushInterval != "" {
		dur, err := time.ParseDuration(d.cfg.Daemon.PushInterval)
		if err != nil {
			return fmt.Errorf("invalid push_interval: %w", err)
		}
		d.pushInterval = dur
	}

	if d.cfg.Daemon.PullInterval != "" {
		dur, err := time.ParseDuration(d.cfg.Daemon.PullInterval)
		if err != nil {
			return fmt.Errorf("invalid pull_interval: %w", err)
		}
		d.pullInterval = dur
	}

	return nil
}

// reloadConfig reloads configuration from file.
func (d *Daemon) reloadConfig() bool {
	newCfg, err := config.Load(d.configPath)
	if err != nil {
		d.logger.Printf("Config reload error: %v", err)
		return false
	}

	// Check if intervals changed
	oldCopy := d.cfg.Daemon.CopyInterval
	oldPush := d.cfg.Daemon.PushInterval
	oldPull := d.cfg.Daemon.PullInterval

	d.cfg = newCfg
	if err := d.parseIntervals(); err != nil {
		d.logger.Printf("Config reload error: %v", err)
		return false
	}

	changed := oldCopy != newCfg.Daemon.CopyInterval ||
		oldPush != newCfg.Daemon.PushInterval ||
		oldPull != newCfg.Daemon.PullInterval

	if changed {
		d.logger.Println("Config reloaded, intervals updated")
		d.logger.Printf("  Copy interval: %v", d.copyInterval)
		d.logger.Printf("  Push interval: %v", d.pushInterval)
		d.logger.Printf("  Pull interval: %v", d.pullInterval)
	}

	return changed
}

// Run starts the daemon loop.
func (d *Daemon) Run() error {
	if d.copyInterval == 0 && d.pushInterval == 0 && d.pullInterval == 0 {
		return fmt.Errorf("no intervals configured in daemon settings")
	}

	// Write PID file
	if err := d.writePidFile(); err != nil {
		return err
	}
	defer d.removePidFile()

	d.logger.Println("Daemon started")
	d.logger.Printf("  Copy interval: %v", d.copyInterval)
	d.logger.Printf("  Push interval: %v", d.pushInterval)
	d.logger.Printf("  Pull interval: %v", d.pullInterval)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create tickers
	var copyTicker, pushTicker, pullTicker *time.Ticker
	var copyChan, pushChan, pullChan <-chan time.Time

	if d.copyInterval > 0 {
		copyTicker = time.NewTicker(d.copyInterval)
		copyChan = copyTicker.C
		defer copyTicker.Stop()
	}

	if d.pushInterval > 0 {
		pushTicker = time.NewTicker(d.pushInterval)
		pushChan = pushTicker.C
		defer pushTicker.Stop()
	}

	if d.pullInterval > 0 {
		pullTicker = time.NewTicker(d.pullInterval)
		pullChan = pullTicker.C
		defer pullTicker.Stop()
	}

	// Main loop
	for {
		select {
		case <-sigChan:
			d.logger.Println("Daemon stopped")
			return nil

		case <-copyChan:
			d.doCopy()

		case <-pushChan:
			d.doPush()

		case <-pullChan:
			d.doPull()
		}
	}
}

func (d *Daemon) doCopy() {
	// Reload config to pick up any changes
	d.reloadConfig()

	d.logger.Println("Copy started")

	copier, err := snapfig.NewCopier(d.cfg)
	if err != nil {
		d.logger.Printf("Copy error: %v", err)
		return
	}

	result, err := copier.Copy()
	if err != nil {
		d.logger.Printf("Copy error: %v", err)
		return
	}

	d.logger.Printf("Copy done: %d paths, %d updated, %d unchanged, %d removed",
		len(result.Copied), result.FilesUpdated, result.FilesSkipped, result.FilesRemoved)

	for _, p := range result.Copied {
		d.logger.Printf("  copied: %s", p)
	}
	for _, p := range result.Skipped {
		d.logger.Printf("  skipped (not found): %s", p)
	}
}

func (d *Daemon) doPush() {
	d.logger.Println("Push started")

	if err := snapfig.PushVaultWithToken(d.vaultDir, d.cfg.GitToken); err != nil {
		d.logger.Printf("Push error: %v", err)
		return
	}

	d.logger.Println("Push done")
}

func (d *Daemon) doPull() {
	d.logger.Println("Pull started")

	result, err := snapfig.PullVaultWithToken(d.vaultDir, d.cfg.Remote, d.cfg.GitToken)
	if err != nil {
		d.logger.Printf("Pull error: %v", err)
		return
	}

	if result.Cloned {
		d.logger.Println("Pull done (cloned)")
	} else {
		d.logger.Println("Pull done")
	}

	if d.cfg.Daemon.AutoRestore {
		d.doRestore()
	}
}

func (d *Daemon) doRestore() {
	d.logger.Println("Restore started (auto)")

	restorer, err := snapfig.NewRestorer(d.cfg)
	if err != nil {
		d.logger.Printf("Restore error: %v", err)
		return
	}

	result, err := restorer.Restore()
	if err != nil {
		d.logger.Printf("Restore error: %v", err)
		return
	}

	d.logger.Printf("Restore done: %d updated, %d unchanged",
		result.FilesUpdated, result.FilesSkipped)
}

func (d *Daemon) writePidFile() error {
	pidPath, err := config.PidFilePath()
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

	return os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func (d *Daemon) removePidFile() {
	pidPath, _ := config.PidFilePath()
	os.Remove(pidPath)
}
