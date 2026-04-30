package service

import (
	"GalleryTags/backend/model"
	"GalleryTags/backend/repository"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}
var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".webm": true, ".mov": true,
}

// FileService handles file scanning and tag application.
type FileService struct {
	repo repository.Store
}

// NewFileService creates a new FileService.
func NewFileService(repo repository.Store) *FileService {
	return &FileService{repo: repo}
}

// ScanInbox scans the configured inbox folder for media files.
// If recursive is true, scans subdirectories as well.
// sortBy controls ordering: "name", "name-desc", "mtime", "mtime-desc", "btime", "btime-desc".
func (s *FileService) ScanInbox(recursive bool, sortBy string) ([]model.FileInfo, error) {
	return s.ScanInboxContext(context.Background(), recursive, sortBy)
}

// ScanInboxContext is like ScanInbox but respects context cancellation.
func (s *FileService) ScanInboxContext(ctx context.Context, recursive bool, sortBy string) ([]model.FileInfo, error) {
	inboxPath, err := s.repo.GetSetting("inbox_path")
	if err != nil {
		return nil, fmt.Errorf("get inbox path: %w", err)
	}
	if inboxPath == "" {
		return nil, fmt.Errorf("inbox path not configured")
	}

	var files []model.FileInfo

	if recursive {
		err = filepath.WalkDir(inboxPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if strings.HasPrefix(d.Name(), ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				return nil
			}
			appendIfMedia(&files, path, d.Name())
			return nil
		})
	} else {
		entries, readErr := os.ReadDir(inboxPath)
		if readErr != nil {
			return nil, fmt.Errorf("scan inbox: %w", readErr)
		}
		for _, d := range entries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
				continue
			}
			path := filepath.Join(inboxPath, d.Name())
			appendIfMedia(&files, path, d.Name())
		}
	}

	if err != nil {
		return nil, fmt.Errorf("scan inbox: %w", err)
	}

	// Fill file times.
	for i := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		info, statErr := os.Stat(files[i].Path)
		if statErr != nil {
			continue
		}
		files[i].ModTime = info.ModTime()
		files[i].BirthTime = birthTime(info)
	}

	sort.Slice(files, func(i, j int) bool {
		switch sortBy {
		case "name-desc":
			return files[i].Name > files[j].Name
		case "mtime":
			if files[i].ModTime.Equal(files[j].ModTime) {
				return files[i].Name < files[j].Name
			}
			return files[i].ModTime.After(files[j].ModTime)
		case "mtime-desc":
			if files[i].ModTime.Equal(files[j].ModTime) {
				return files[i].Name < files[j].Name
			}
			return files[i].ModTime.Before(files[j].ModTime)
		case "btime":
			if files[i].BirthTime.Equal(files[j].BirthTime) {
				return files[i].Name < files[j].Name
			}
			return files[i].BirthTime.After(files[j].BirthTime)
		case "btime-desc":
			if files[i].BirthTime.Equal(files[j].BirthTime) {
				return files[i].Name < files[j].Name
			}
			return files[i].BirthTime.Before(files[j].BirthTime)
		default: // "name" ascending
			return files[i].Name < files[j].Name
		}
	})

	if files == nil {
		files = []model.FileInfo{}
	}
	return files, nil
}

// ApplyTags saves tag associations for a file and optionally moves it.
func (s *FileService) ApplyTags(filePath string, tagIDs []int) (model.ApplyResult, error) {
	if len(tagIDs) == 0 {
		return model.ApplyResult{}, nil
	}

	// Validate folder tags before writing.
	folderPath, folderCount, err := s.repo.FindFolderTags(tagIDs)
	if err != nil {
		return model.ApplyResult{}, fmt.Errorf("check folder tags: %w", err)
	}
	if folderCount > 1 {
		return model.ApplyResult{}, fmt.Errorf("only one folder tag allowed, found %d", folderCount)
	}

	// Persist tag associations.
	if err := s.repo.SaveFileTags(filePath, tagIDs); err != nil {
		return model.ApplyResult{}, fmt.Errorf("save tags: %w", err)
	}

	// Move file if a folder tag was applied.
	if folderCount == 1 {
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			return model.ApplyResult{}, fmt.Errorf("create folder: %w", err)
		}

		dst := filepath.Join(folderPath, filepath.Base(filePath))
		if err := moveFile(filePath, dst); err != nil {
			return model.ApplyResult{}, fmt.Errorf("move file: %w", err)
		}
		if err := s.repo.UpdateFilePath(filePath, dst); err != nil {
			return model.ApplyResult{}, fmt.Errorf("update path: %w", err)
		}
		return model.ApplyResult{NewPath: dst, Moved: true}, nil
	}

	return model.ApplyResult{}, nil
}

func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	if isCrossDeviceError(err) {
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copy: %w", err)
		}
		return os.Remove(src)
	}
	return err
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Close()
}

// TrashFile moves a file to the system trash.
// Platform-specific implementation: see trashfile_linux.go / trashfile_windows.go.
func (s *FileService) TrashFile(filePath string) error {
	return trashFile(filePath)
}

func appendIfMedia(files *[]model.FileInfo, path, name string) {
	ext := strings.ToLower(filepath.Ext(path))
	var fileType string
	if imageExts[ext] {
		fileType = "image"
	} else if videoExts[ext] {
		fileType = "video"
	} else {
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	*files = append(*files, model.FileInfo{
		Path: absPath,
		Name: name,
		Type: fileType,
	})
}
