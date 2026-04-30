package main

import (
	"context"
	"path/filepath"
	"sync"

	"GalleryTags/backend/logger"
	"GalleryTags/backend/model"
	"GalleryTags/backend/paths"
	"GalleryTags/backend/repository"
	"GalleryTags/backend/server"
	"GalleryTags/backend/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App exposes backend methods to the Svelte frontend via Wails bindings.
type App struct {
	ctx        context.Context
	fileServer *server.FileServer
	tags       *service.TagService
	files      *service.FileService
	settings   *service.SettingsService

	scanMu     sync.Mutex
	scanCancel context.CancelFunc
}

// NewApp creates an App with its FileServer dependency.
func NewApp(fs *server.FileServer) *App {
	return &App{fileServer: fs}
}

// startup is called by Wails when the app starts. Initializes DB and services.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dataDir := paths.MkdirAll(paths.DataDir())
	configDir := paths.MkdirAll(paths.ConfigDir())

	logger.Init()
	log := logger.L()

	log.Info("starting GalleryTags", "dataDir", dataDir, "configDir", configDir)

	store, err := repository.InitDB(filepath.Join(dataDir, "gallery.db"))
	if err != nil {
		log.Error("db init failed", "error", err)
		panic("db init: " + err.Error())
	}
	log.Info("database initialized")

	a.tags = service.NewTagService(store)
	a.files = service.NewFileService(store)
	a.settings = service.NewSettingsService(store)

	runtime.WindowMaximise(a.ctx)
	log.Info("startup complete")
}

// GetMediaBaseURL returns the base URL for local file requests.
func (a *App) GetMediaBaseURL() string {
	return a.fileServer.BaseURL()
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func (a *App) GetInboxPath() string {
	path, err := a.settings.GetInboxPath()
	if err != nil {
		logger.L().Error("get inbox path failed", "error", err)
		return ""
	}
	return path
}

func (a *App) SetInboxPath(path string) error {
	logger.L().Info("set inbox path", "path", path)
	return a.settings.SetInboxPath(path)
}

// ---------------------------------------------------------------------------
// File scanning
// ---------------------------------------------------------------------------

func (a *App) ScanInbox(recursive bool, sortBy string) ([]model.FileInfo, error) {
	log := logger.L()
	a.scanMu.Lock()
	if a.scanCancel != nil {
		a.scanCancel()
		a.scanCancel = nil
		log.Info("cancelled previous scan")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.scanCancel = cancel
	a.scanMu.Unlock()

	log.Info("scan started", "recursive", recursive, "sortBy", sortBy)
	result, err := a.files.ScanInboxContext(ctx, recursive, sortBy)

	a.scanMu.Lock()
	if a.scanCancel != nil {
		a.scanCancel = nil
	}
	a.scanMu.Unlock()

	if err != nil {
		log.Error("scan failed", "error", err, "recursive", recursive, "sortBy", sortBy)
	} else {
		log.Info("scan complete", "files", len(result))
	}

	return result, err
}

// ---------------------------------------------------------------------------
// Tag management
// ---------------------------------------------------------------------------

func (a *App) GetTags() ([]model.Tag, error) {
	return a.tags.GetTags()
}

func (a *App) CreateTag(name, tagType, folder, color, hotkey string) (model.Tag, error) {
	logger.L().Info("create tag", "name", name, "type", tagType, "folder", folder)
	return a.tags.CreateTag(name, tagType, folder, color, hotkey)
}

func (a *App) UpdateTag(id int, name, tagType, folder, color, hotkey string) error {
	logger.L().Info("update tag", "id", id, "name", name, "type", tagType)
	return a.tags.UpdateTag(id, name, tagType, folder, color, hotkey)
}

func (a *App) DeleteTag(id int) error {
	logger.L().Info("delete tag", "id", id)
	return a.tags.DeleteTag(id)
}

// ---------------------------------------------------------------------------
// Tag application
// ---------------------------------------------------------------------------

func (a *App) ApplyTags(filePath string, tagIDs []int) (model.ApplyResult, error) {
	logger.L().Info("apply tags", "file", filePath, "tagIDs", tagIDs)
	result, err := a.files.ApplyTags(filePath, tagIDs)
	if err != nil {
		logger.L().Error("apply tags failed", "file", filePath, "error", err)
	} else if result.Moved {
		logger.L().Info("file moved", "from", filePath, "to", result.NewPath)
	}
	return result, err
}

// ---------------------------------------------------------------------------
// File deletion
// ---------------------------------------------------------------------------

func (a *App) TrashFile(filePath string) error {
	logger.L().Info("trash file", "file", filePath)
	err := a.files.TrashFile(filePath)
	if err != nil {
		logger.L().Error("trash file failed", "file", filePath, "error", err)
	}
	return err
}

// ---------------------------------------------------------------------------
// File tags
// ---------------------------------------------------------------------------

func (a *App) GetFileTags(filePath string) ([]model.Tag, error) {
	return a.tags.GetFileTags(filePath)
}

// GetFilteredFiles returns file paths matching all given tag IDs.
func (a *App) GetFilteredFiles(tagIDs []int) ([]string, error) {
	return a.tags.GetFilesWithTagIDs(tagIDs)
}

// ---------------------------------------------------------------------------
// Dialogs
// ---------------------------------------------------------------------------

func (a *App) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder",
	})
}
