package handler

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/The-honoured1/gostream/storage"
)

// StreamHandler handles raw video streaming using HTTP Range requests.
func StreamHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract video ID/path from URL (simplified for now)
		path := r.URL.Path
		// Assuming the request is /stream/videoID
		videoPath := strings.TrimPrefix(path, "/stream/")
		if videoPath == "" {
			http.Error(w, "missing video path", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		exists, err := s.Exists(ctx, videoPath)
		if err != nil || !exists {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}

		file, err := s.Open(ctx, videoPath)
		if err != nil {
			http.Error(w, "failed to open video", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		info, err := s.Stat(ctx, videoPath)
		if err != nil {
			http.Error(w, "failed to stat video", http.StatusInternalServerError)
			return
		}

		// http.ServeContent handles Range requests automatically!
		// It requires a ReadSeeker, which our storage.Open returns.
		http.ServeContent(w, r, info.Name,  /* lastModified time could be added here if storage.Stat provided it */  nil, file)
	}
}

// HLSHandler handles serving of .m3u8 playlists and .ts segments.
type HLSHandler struct {
	storage storage.Storage
	baseDir string // Directory where HLS files are stored
}

func NewHLSHandler(s storage.Storage, baseDir string) *HLSHandler {
	return &HLSHandler{
		storage: s,
		baseDir: baseDir,
	}
}

func (h *HLSHandler) ServePlaylist(w http.ResponseWriter, r *http.Request) {
	// Path example: /video/{id}/playlist.m3u8
	path := strings.TrimPrefix(r.URL.Path, "/")
	h.serveFile(w, r, path, "application/vnd.apple.mpegurl")
}

func (h *HLSHandler) ServeSegment(w http.ResponseWriter, r *http.Request) {
	// Path example: /video/{id}/segment/0.ts
	path := strings.TrimPrefix(r.URL.Path, "/")
	h.serveFile(w, r, path, "video/MP2T")
}

func (h *HLSHandler) serveFile(w http.ResponseWriter, r *http.Request, path string, contentType string) {
	ctx := r.Context()
	
	exists, err := h.storage.Exists(ctx, path)
	if err != nil || !exists {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	file, err := h.storage.Open(ctx, path)
	if err != nil {
		http.Error(w, "failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow CORS for streaming
	
	// For HLS segments, we don't usually need Range requests, but for large ones we could use ServeContent
	// Here we use io.Copy for simple streaming
	io.Copy(w, file)
}
