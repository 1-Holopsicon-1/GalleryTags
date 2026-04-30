package service

import "GalleryTags/backend/repository"

// SettingsService handles application settings.
type SettingsService struct {
	repo repository.Store
}

// NewSettingsService creates a new SettingsService.
func NewSettingsService(repo repository.Store) *SettingsService {
	return &SettingsService{repo: repo}
}

// GetInboxPath returns the configured inbox folder path.
func (s *SettingsService) GetInboxPath() (string, error) {
	return s.repo.GetSetting("inbox_path")
}

// SetInboxPath persists the inbox folder path.
func (s *SettingsService) SetInboxPath(path string) error {
	return s.repo.SetSetting("inbox_path", path)
}
