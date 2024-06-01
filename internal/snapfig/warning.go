package snapfig

import (
	"fmt"
	"log"
	"os"
	"time"
)

const warningFile = "~/.config/snapfig/warnings.log"

func LogWarning(message string) error {
	f, err := os.OpenFile(warningFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening warning file: %v", err)
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	warningMessage := fmt.Sprintf("[%s] WARNING: %s\n", timestamp, message)

	if _, err := f.WriteString(warningMessage); err != nil {
		log.Printf("Error writing to warning file: %v", err)
		return err
	}
	return nil
}
