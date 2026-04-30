//go:build linux

package service

import (
	"os"
	"syscall"
	"time"
)

func birthTime(info os.FileInfo) time.Time {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
	}
	return info.ModTime()
}
