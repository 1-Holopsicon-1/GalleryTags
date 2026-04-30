//go:build windows

package service

import (
	"os"
	"time"
)

func birthTime(info os.FileInfo) time.Time {
	// Windows: use ModTime as fallback; birth time not easily accessible via syscall.
	return info.ModTime()
}
