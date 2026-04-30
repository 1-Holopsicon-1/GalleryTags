//go:build linux

package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func trashFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("empty path")
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	cmd := exec.Command("gio", "trash", filePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gio trash: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
