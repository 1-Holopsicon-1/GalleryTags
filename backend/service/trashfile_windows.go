//go:build windows

package service

import (
	"fmt"
	"os"
	"os/exec"
)

func trashFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("empty path")
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	// Use PowerShell to move file to Recycle Bin via Shell.Application COM.
	ps := `
Add-Type -AssemblyName Microsoft.VisualBasic
[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile(
  $args[0],
  [Microsoft.VisualBasic.FileIO.UIOption]::OnlyErrorDialogs,
  [Microsoft.VisualBasic.FileIO.RecycleOption]::SendToRecycleBin
)
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps, "-", filePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("recycle: %w: %s", err, string(out))
	}
	return nil
}
