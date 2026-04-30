// Package server provides a localhost HTTP file server used by WebKit2GTK
// for streaming video and serving local images outside Wails' AssetServer.
package server

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileServer serves local files from disk on a random localhost port.
type FileServer struct {
	port int
}

// New starts a localhost HTTP server on a random port and returns a *FileServer.
// Panics if the listener cannot be created (called once at startup).
func New() *FileServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("localfile: failed to listen: %v", err))
	}

	fs := &FileServer{
		port: listener.Addr().(*net.TCPAddr).Port,
	}

	mux := http.NewServeMux()
	mux.Handle("/local/", http.HandlerFunc(fs.handleFile))
	go func() {
		if err := http.Serve(listener, mux); err != nil {
			slog.Error("file server error", "error", err)
		}
	}()

	slog.Info("file server started", "port", fs.port)
	return fs
}

// Port returns the dynamically assigned port number.
func (fs *FileServer) Port() int {
	return fs.port
}

// BaseURL returns the base URL for file requests.
func (fs *FileServer) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/local/", fs.port)
}

// handleFile is the HTTP handler for /local/ requests.
// Implements full Range request support required by WebKit2GTK for <video>.
func (fs *FileServer) handleFile(w http.ResponseWriter, r *http.Request) {
	encodedPath := strings.TrimPrefix(r.URL.Path, "/local/")
	filePath, err := url.PathUnescape(encodedPath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	if stat.IsDir() {
		http.Error(w, "is a directory", http.StatusBadRequest)
		return
	}

	contentType := contentTypeFor(filePath)
	size := stat.Size()
	rangeHeader := r.Header.Get("Range")

	// Always advertise Range support.
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", contentType)

	if rangeHeader == "" {
		// Full file response.
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		w.WriteHeader(http.StatusOK)
		sendFile(w, filePath)
		return
	}

	// Parse "Range: bytes=start-end"
	start, end, ok := parseRange(rangeHeader, size)
	if !ok {
		http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	contentLength := end - start + 1
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(size, 10))
	w.WriteHeader(http.StatusPartialContent)
	sendFileRange(w, filePath, start, contentLength)
}

func contentTypeFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mkv", ".webm":
		return "video/webm"
	case ".mp4", ".mov":
		return "video/mp4"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func parseRange(header string, size int64) (start, end int64, ok bool) {
	// Format: "bytes=start-end" or "bytes=start-"
	if !strings.HasPrefix(header, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(header, "bytes=")
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}

	if parts[1] == "" {
		end = size - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, false
		}
	}

	if start < 0 || end >= size || start > end {
		return 0, 0, false
	}

	return start, end, true
}

func sendFile(w io.Writer, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func sendFileRange(w io.Writer, path string, offset, length int64) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	f.Seek(offset, io.SeekStart)
	io.CopyN(w, f, length)
}
