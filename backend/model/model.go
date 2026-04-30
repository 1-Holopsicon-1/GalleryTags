package model

import "time"

// Tag represents a user-defined label or folder mapping.
type Tag struct {
	ID     int
	Name   string
	Type   string // "label" or "folder"
	Folder string // only for type="folder"
	Color  string // hex color
	Hotkey string // keyboard shortcut, e.g. "ctrl+n", "1", ""
}

// FileInfo describes a media file found in the inbox.
type FileInfo struct {
	Path      string
	Name      string
	Type      string    // "image" or "video"
	ModTime   time.Time // last modification time
	BirthTime time.Time // creation time (btime on Linux)
}

// ApplyResult describes the outcome of applying tags to a file.
type ApplyResult struct {
	NewPath string
	Moved   bool
}
