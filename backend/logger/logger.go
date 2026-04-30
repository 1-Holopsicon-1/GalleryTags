package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"GalleryTags/backend/paths"
)

var once sync.Once
var log *slog.Logger

// Init sets up the global logger writing to both stderr and a log file.
func Init() {
	once.Do(func() {
		logDir := paths.MkdirAll(paths.LogDir())
		logPath := filepath.Join(logDir, "app.log")

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "logger: cannot open %s: %v\n", logPath, err)
			log = slog.Default()
			return
		}

		multi := io.MultiWriter(os.Stderr, f)
		log = slog.New(slog.NewTextHandler(multi, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(log)

		log.Info("logger initialized", "path", logPath)
	})
}

// L returns the global logger.
func L() *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}
