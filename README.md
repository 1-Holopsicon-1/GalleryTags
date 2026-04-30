# GalleryTags

Desktop gallery app for sorting images and videos using tags.
Built with [Wails](https://wails.io) (Go + Svelte + SQLite).

## Features

- Slideshow view — one file at a time, keyboard navigation
- Tag system — labels and folder tags with hotkeys
- Auto-move — assigning a folder tag moves the file on disk
- Recursive scanning with sort by name, mtime, btime
- Thumbnail sidebar with lazy loading
- Dark, minimal UI

## Requirements

- Go 1.21+
- Node.js + Yarn 4
- `webkit2gtk-4.1-dev` (Linux)
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build & Run

```bash
# Dev mode (hot reload)
wails dev -tags webkit2_41

# Production build
wails build -tags webkit2_41

# Frontend only
cd frontend && yarn build
cd frontend && yarn svelte-check
```

### Wayland Support

By default the app runs through XWayland (`GDK_BACKEND=x11`) to avoid GPU conflicts with other Wayland apps (Firefox, Steam, etc.).

```bash
# Native Wayland (may affect other GPU-accelerated windows)
./GalleryTags --wayland
```

## Keyboard Shortcuts

| Key | Action |
|-----|---------|
| `←` / `→` | Previous / Next file |
| `f` / `F` | Toggle slideshow mode |
| `Escape` | Exit slideshow |
| `1`–`9`, letters | Hotkeys (if assigned to tags) |

## Architecture

```
app.go                → Wails bindings, delegates to services
main.go               → Entry point, GDK_BACKEND flag
backend/
  model/              → Tag, FileInfo, ApplyResult types
  repository/         → SQLite queries (store, tags, file_tags, settings)
  server/             → HTTP file server (Range support for video)
  service/            → Business logic (tag, file, settings)
frontend/src/
  api/                → Wails bindings wrapper (PascalCase → camelCase)
  stores/             → Svelte stores (files, tags, preload)
  components/         → UI components
```

Clean architecture: `model/ → repository/ → service/` on backend, `types/ → api/ → stores/ → components/` on frontend.

## Database

SQLite via `modernc.org/sqlite` (pure Go, no CGO). Stored at `~/.config/GalleryTags/gallery.db`.

| Table | Purpose |
|-------|---------|
| `tags` | Tag definitions (name, type, folder path, color, hotkey) |
| `file_tags` | File ↔ tag associations |
| `settings` | Key-value config (inbox path, etc.) |

## Linux Notes

- Build tag `webkit2_41` is required (for `webkit2gtk-4.1`)
- Vendor patches in `window.c` must be re-applied after `go mod vendor`:
  - Removed Wayland decorator hack in `SetMinMaxSize`
  - Removed `SetMinMaxSize` call from `Fullscreen()`
  - No `GDK_HINT_MAX_SIZE` when max dimensions are unset
- GPU policy: `WebviewGpuPolicyNever` required on Hyprland
- `position: fixed` overlays must be outside `overflow: hidden` containers (webkit2gtk bug)

## License

MIT