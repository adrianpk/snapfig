package command

import (
	"fmt"
	"os"

	"github.com/adrianpk/snapfig/internal/config"
	"github.com/spf13/afero"
)

// Worker is the command's companion struct that does the actual work.
type Worker struct {
	cfg       *config.Config
	fs        afero.Fs
	fsRoot    string
	configDir string
	dirs      []string
}

func NewWorker(cfg *config.Config) *Worker {
	return &Worker{
		cfg:  cfg,
		fs:   afero.NewOsFs(),
		dirs: []string{},
	}
}

func (w *Worker) Setup() error {
	if w.fsRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		w.fsRoot = home
	}

	w.fs = afero.NewBasePathFs(afero.NewOsFs(), w.fsRoot)

	toolConfigDir, err := w.ensureSnapfigDir()
	if err != nil {
		return err
	}

	w.configDir = toolConfigDir

	return nil
}

// Cfg returns the ommand's configuration
func (w *Worker) Cfg() *config.Config {
	return w.cfg
}

func (cmd *Worker) userMsg(message string) {
	fmt.Println(message)
}

// SetFS sets a ustom filesystem.
// It's only used for tests for now.
func (w *Worker) SetFS(fs afero.Fs) {
	w.fs = fs
}
