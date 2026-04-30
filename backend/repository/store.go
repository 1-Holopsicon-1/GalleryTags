package repository

import (
	"GalleryTags/backend/model"
	"database/sql"

	_ "modernc.org/sqlite"
)

// Store defines the data access interface.
type Store interface {
	// Tags
	GetTags() ([]model.Tag, error)
	CreateTag(name, tagType, folder, color, hotkey string) (model.Tag, error)
	UpdateTag(id int, name, tagType, folder, color, hotkey string) error
	DeleteTag(id int) error

	// File tags
	GetFileTags(filePath string) ([]model.Tag, error)
	SaveFileTags(filePath string, tagIDs []int) error
	UpdateFilePath(oldPath, newPath string) error
	FindFolderTags(tagIDs []int) (folderPath string, count int, err error)
	GetFilesWithTagIDs(tagIDs []int) ([]string, error)

	// Settings
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error

	// Lifecycle
	Close() error
}

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	*sql.DB
}

// InitDB opens a SQLite database, configures it, and runs migrations.
func InitDB(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

	// Enable foreign key constraints.
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, err
	}

	s := &SQLiteStore{db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close shuts down the database connection.
func (s *SQLiteStore) Close() error {
	return s.DB.Close()
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tags (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		name      TEXT NOT NULL UNIQUE,
		type      TEXT NOT NULL CHECK(type IN ('label', 'folder')),
		folder    TEXT,
		color     TEXT DEFAULT '#4a9eff'
	);

	CREATE TABLE IF NOT EXISTS file_tags (
		file_path   TEXT NOT NULL,
		tag_id      INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		applied_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (file_path, tag_id)
	);

	CREATE TABLE IF NOT EXISTS settings (
		key   TEXT PRIMARY KEY,
		value TEXT
	);
	`
	if _, err := s.Exec(schema); err != nil {
		return err
	}
	// Add hotkey column if not exists (migration for existing DBs).
	s.Exec("ALTER TABLE tags ADD COLUMN hotkey TEXT NOT NULL DEFAULT ''")
	return nil
}
