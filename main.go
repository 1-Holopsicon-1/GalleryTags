package main

import (
	"embed"
	"os"
	"runtime"

	"GalleryTags/backend/server"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if runtime.GOOS == "linux" {
		wayland := false
		for _, arg := range os.Args[1:] {
			if arg == "--wayland" {
				wayland = true
				break
			}
		}
		if !wayland && os.Getenv("GDK_BACKEND") == "" {
			os.Setenv("GDK_BACKEND", "x11")
		}
	}

	fileServer := server.New()
	app := NewApp(fileServer)

	opts := &options.App{
		Title:      "GalleryTags",
		Width:      1280,
		Height:     800,
		MinWidth:   800,
		MinHeight:  600,
		Fullscreen: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 14, G: 14, B: 15, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	}
	applyPlatformOptions(opts)

	err := wails.Run(opts)

	if err != nil {
		println("Error:", err.Error())
	}
}
