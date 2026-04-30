package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the directory for app configuration.
//   - Linux: XDG_CONFIG_HOME/GalleryTags (default ~/.config/GalleryTags)
//   - Windows: %LOCALAPPDATA%/GalleryTags
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "GalleryTags")
	}
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "GalleryTags")
}

// DataDir returns the directory for app data (database).
//   - Linux: XDG_DATA_HOME/GalleryTags (default ~/.local/share/GalleryTags)
//   - Windows: %LOCALAPPDATA%/GalleryTags
func DataDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "GalleryTags")
	}
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataDir, "GalleryTags")
}

// LogDir returns the directory for log files.
//   - Linux: XDG_STATE_HOME/GalleryTags (default ~/.local/state/GalleryTags)
//   - Windows: %LOCALAPPDATA%/GalleryTags
func LogDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "GalleryTags")
	}
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, _ := os.UserHomeDir()
		stateDir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateDir, "GalleryTags")
}

// MkdirAll creates a directory and all parents, panics on failure.
func MkdirAll(dir string) string {
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("mkdir " + dir + ": " + err.Error())
	}
	return dir
}
