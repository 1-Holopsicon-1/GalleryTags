//go:build linux

package main

import (
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

func platformOptions() *linux.Options {
	return &linux.Options{
		WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
	}
}

func applyPlatformOptions(app *options.App) {
	app.Linux = platformOptions()
}
