//go:build windows

package service

import "strings"

// On Windows, os.Rename across drives returns an error mentioning "different" or the volume serial.
func isCrossDeviceError(err error) bool {
	return strings.Contains(err.Error(), "different") ||
		strings.Contains(err.Error(), "cross-device") ||
		strings.Contains(err.Error(), "not same")
}
