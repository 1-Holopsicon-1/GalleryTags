package repository

import "database/sql"

// GetSetting returns the value for a setting key, or empty string if not found.
func (s *SQLiteStore) GetSetting(key string) (string, error) {
	var value string
	err := s.QueryRow(
		"SELECT value FROM settings WHERE key = ?", key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting persists a key-value setting.
func (s *SQLiteStore) SetSetting(key, value string) error {
	_, err := s.Exec(
		"INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}
