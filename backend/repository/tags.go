package repository

import (
	"GalleryTags/backend/model"
	"fmt"
)

// GetTags returns all tags ordered alphabetically by name.
func (s *SQLiteStore) GetTags() ([]model.Tag, error) {
	rows, err := s.Query(
		"SELECT id, name, type, COALESCE(folder, ''), COALESCE(color, '#4a9eff'), COALESCE(hotkey, '') FROM tags ORDER BY name",
	)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Folder, &t.Color, &t.Hotkey); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}
	if tags == nil {
		tags = []model.Tag{}
	}
	return tags, nil
}

// CreateTag inserts a new tag and returns it with the generated ID.
func (s *SQLiteStore) CreateTag(name, tagType, folder, color, hotkey string) (model.Tag, error) {
	result, err := s.Exec(
		"INSERT INTO tags (name, type, folder, color, hotkey) VALUES (?, ?, ?, ?, ?)",
		name, tagType, folder, color, hotkey,
	)
	if err != nil {
		return model.Tag{}, fmt.Errorf("insert tag: %w", err)
	}

	id, _ := result.LastInsertId()
	return model.Tag{
		ID:     int(id),
		Name:   name,
		Type:   tagType,
		Folder: folder,
		Color:  color,
		Hotkey: hotkey,
	}, nil
}

// UpdateTag modifies an existing tag.
func (s *SQLiteStore) UpdateTag(id int, name, tagType, folder, color, hotkey string) error {
	result, err := s.Exec(
		"UPDATE tags SET name = ?, type = ?, folder = ?, color = ?, hotkey = ? WHERE id = ?",
		name, tagType, folder, color, hotkey, id,
	)
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("tag not found: %d", id)
	}
	return nil
}

// DeleteTag removes a tag by ID.
func (s *SQLiteStore) DeleteTag(id int) error {
	result, err := s.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("tag not found: %d", id)
	}
	return nil
}
