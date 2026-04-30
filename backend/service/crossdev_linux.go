//go:build linux

package service

import (
	"errors"
	"os"
	"syscall"
)

func isCrossDeviceError(err error) bool {
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) {
		return linkErr.Err == syscall.EXDEV
	}
	return false
}
