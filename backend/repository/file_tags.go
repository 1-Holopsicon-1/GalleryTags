package repository

import (
	"GalleryTags/backend/model"
	"fmt"
	"strings"
)

// GetFileTags returns all tags applied to the given file path.
func (s *SQLiteStore) GetFileTags(filePath string) ([]model.Tag, error) {
	rows, err := s.Query(
		`SELECT t.id, t.name, t.type, COALESCE(t.folder, ''), COALESCE(t.color, '#4a9eff'), COALESCE(t.hotkey, '')
		 FROM tags t
		 JOIN file_tags ft ON t.id = ft.tag_id
		 WHERE ft.file_path = ?`,
		filePath,
	)
	if err != nil {
		return nil, fmt.Errorf("query file tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Folder, &t.Color, &t.Hotkey); err != nil {
			return nil, fmt.Errorf("scan file tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file tags: %w", err)
	}
	if tags == nil {
		tags = []model.Tag{}
	}
	return tags, nil
}

// SaveFileTags replaces all tag associations for a file.
func (s *SQLiteStore) SaveFileTags(filePath string, tagIDs []int) error {
	// Delete existing associations.
	if _, err := s.Exec("DELETE FROM file_tags WHERE file_path = ?", filePath); err != nil {
		return fmt.Errorf("delete file tags: %w", err)
	}

	// Insert new associations.
	for _, tagID := range tagIDs {
		if _, err := s.Exec(
			"INSERT INTO file_tags (file_path, tag_id) VALUES (?, ?)",
			filePath, tagID,
		); err != nil {
			return fmt.Errorf("insert file tag: %w", err)
		}
	}
	return nil
}

// UpdateFilePath updates all file_tags records from oldPath to newPath.
func (s *SQLiteStore) UpdateFilePath(oldPath, newPath string) error {
	_, err := s.Exec(
		"UPDATE file_tags SET file_path = ? WHERE file_path = ?",
		newPath, oldPath,
	)
	if err != nil {
		return fmt.Errorf("update file path: %w", err)
	}
	return nil
}

// FindFolderTags looks up folder-type tags among the given IDs.
// Returns the folder path and count of matching folder tags.
func (s *SQLiteStore) FindFolderTags(tagIDs []int) (folderPath string, count int, err error) {
	if len(tagIDs) == 0 {
		return "", 0, nil
	}

	placeholders := make([]string, len(tagIDs))
	args := make([]any, len(tagIDs))
	for i, id := range tagIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT folder FROM tags WHERE id IN (%s) AND type = 'folder'",
		strings.Join(placeholders, ","),
	)

	rows, err := s.Query(query, args...)
	if err != nil {
		return "", 0, fmt.Errorf("query folder tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			return "", 0, fmt.Errorf("scan folder tag: %w", err)
		}
		count++
		folderPath = folder
	}
	if err := rows.Err(); err != nil {
		return "", 0, fmt.Errorf("iterate folder tags: %w", err)
	}

	return folderPath, count, nil
}

// GetFilesWithTagIDs returns file paths that have ALL of the specified tagIDs (intersection).
// Empty tagIDs returns empty slice.
func (s *SQLiteStore) GetFilesWithTagIDs(tagIDs []int) ([]string, error) {
	if len(tagIDs) == 0 {
		return []string{}, nil
	}

	placeholders := make([]string, len(tagIDs))
	args := make([]any, len(tagIDs)+1) // +1 for the HAVING COUNT argument
	for i, id := range tagIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(tagIDs)] = len(tagIDs)

	query := fmt.Sprintf(
		`SELECT file_path FROM file_tags
		 WHERE tag_id IN (%s)
		 GROUP BY file_path
		 HAVING COUNT(DISTINCT tag_id) = ?`,
		strings.Join(placeholders, ","),
	)

	rows, err := s.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query files with tag IDs: %w", err)
	}
	defer rows.Close()

	var filePaths []string
	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return nil, fmt.Errorf("scan file path: %w", err)
		}
		filePaths = append(filePaths, filePath)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file paths: %w", err)
	}
	if filePaths == nil {
		filePaths = []string{}
	}
	return filePaths, nil
}
