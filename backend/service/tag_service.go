package service

import (
	"GalleryTags/backend/model"
	"GalleryTags/backend/repository"
	"fmt"
	"strings"
)

// TagService handles tag business logic.
type TagService struct {
	repo repository.Store
}

// NewTagService creates a new TagService.
func NewTagService(repo repository.Store) *TagService {
	return &TagService{repo: repo}
}

// GetTags returns all tags.
func (s *TagService) GetTags() ([]model.Tag, error) {
	return s.repo.GetTags()
}

// CreateTag validates and creates a new tag.
func (s *TagService) CreateTag(name, tagType, folder, color, hotkey string) (model.Tag, error) {
	if err := validateTag(name, tagType, folder); err != nil {
		return model.Tag{}, err
	}
	if err := validateHotkey(s.repo, hotkey, 0); err != nil {
		return model.Tag{}, err
	}
	if tagType == "label" {
		folder = ""
	}
	if color == "" {
		color = "#4a9eff"
	}
	return s.repo.CreateTag(name, tagType, folder, color, hotkey)
}

// UpdateTag validates and updates an existing tag.
func (s *TagService) UpdateTag(id int, name, tagType, folder, color, hotkey string) error {
	if err := validateTag(name, tagType, folder); err != nil {
		return err
	}
	if err := validateHotkey(s.repo, hotkey, id); err != nil {
		return err
	}
	if tagType == "label" {
		folder = ""
	}
	if color == "" {
		color = "#4a9eff"
	}
	return s.repo.UpdateTag(id, name, tagType, folder, color, hotkey)
}

// DeleteTag removes a tag.
func (s *TagService) DeleteTag(id int) error {
	return s.repo.DeleteTag(id)
}

// GetFileTags returns tags applied to a file.
func (s *TagService) GetFileTags(filePath string) ([]model.Tag, error) {
	return s.repo.GetFileTags(filePath)
}

// GetFilesWithTagIDs returns file paths that have all specified tag IDs.
func (s *TagService) GetFilesWithTagIDs(tagIDs []int) ([]string, error) {
	return s.repo.GetFilesWithTagIDs(tagIDs)
}

func validateTag(name, tagType, folder string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("tag name must not be empty")
	}
	if tagType != "label" && tagType != "folder" {
		return fmt.Errorf("tag type must be 'label' or 'folder'")
	}
	if tagType == "folder" && strings.TrimSpace(folder) == "" {
		return fmt.Errorf("folder tag must have a non-empty path")
	}
	return nil
}

// Reserved hotkeys that conflict with app shortcuts.
var reservedHotkeys = map[string]bool{
	"arrowleft": true, "arrowright": true, "arrowup": true, "arrowdown": true,
	"f": true, " ": true, "delete": true, "escape": true,
}

func validateHotkey(repo repository.Store, hotkey string, excludeID int) error {
	hotkey = strings.TrimSpace(hotkey)
	if hotkey == "" {
		return nil
	}
	hotkey = strings.ToLower(hotkey)
	if reservedHotkeys[hotkey] {
		return fmt.Errorf("hotkey '%s' is reserved", hotkey)
	}
	tags, err := repo.GetTags()
	if err != nil {
		return nil // can't validate, let it through
	}
	for _, t := range tags {
		if t.ID != excludeID && strings.EqualFold(t.Hotkey, hotkey) {
			return fmt.Errorf("hotkey '%s' already used by tag '%s'", hotkey, t.Name)
		}
	}
	return nil
}
