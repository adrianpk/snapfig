package snapfig

import (
	"log"
	"os/exec"
)

func InitRepo(path string) error {
	cmd := exec.Command("git", "init", path)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		log.Printf("Error initializing repository: %v", err)
		return err
	}
	return nil
}
