package delivery

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/The-honoured1/gostream/core"
)

// Engine handles the HTTP delivery of media streams.
type Engine struct {
	provider   core.Provider
	storage    core.Storage
	cache      core.Cache
	maxConns   int
	semaphore  chan struct{}
	bufferPool *sync.Pool
	logger     *log.Logger
}

type Options struct {
	MaxConcurrentRequests int
	CacheSize            int64
}

func NewEngine(p core.Provider, s core.Storage, c core.Cache, opts Options) *Engine {
	return &Engine{
		provider:  p,
		storage:   s,
		cache:     c,
		maxConns:  opts.MaxConcurrentRequests,
		semaphore: make(chan struct{}, opts.MaxConcurrentRequests),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 32*1024) // 32KB buffer
			},
		},
		logger: log.New(log.Writer(), "[DELIVERY] ", log.LstdFlags),
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Backpressure: Limit concurrent active requests
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	default:
		http.Error(w, "Too many concurrent requests", http.StatusServiceUnavailable)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id := pathParts[1]
	stream, ok := e.provider.Get(id)
	if !ok {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	// Route based on request type
	action := ""
	if len(pathParts) > 2 {
		action = pathParts[2]
	}

	switch action {
	case "stream":
		e.handleProgressive(w, r, stream)
	case "playlist.m3u8":
		e.handleHLSPlaylist(w, r, id)
	case "segment":
		if len(pathParts) > 3 {
			e.handleHLSSegment(w, r, id, pathParts[3])
		} else {
			http.Error(w, "Segment index missing", http.StatusBadRequest)
		}
	default:
		// Dynamic decision: if HLS exists, we might want to tell the client
		// but since this is a GET request, we'll serve the raw stream
		// unless we want to redirect. Let's check for HLS presence.
		playlistPath := fmt.Sprintf("hls/%s/playlist.m3u8", id)
		if exists, _ := e.storage.Exists(r.Context(), playlistPath); exists {
			// If it's a browser-like request or common player, maybe redirect?
			// For now, let's just serve the raw stream as requested or provide a header.
			w.Header().Set("X-HLS-Available", "true")
			w.Header().Set("X-HLS-Playlist", fmt.Sprintf("/media/%s/playlist.m3u8", id))
		}
		e.handleProgressive(w, r, stream)
	}
}

func (e *Engine) handleProgressive(w http.ResponseWriter, r *http.Request, s core.MediaStream) {
	ctx := r.Context()
	file, err := s.Open(ctx)
	if err != nil {
		e.logger.Printf("Failed to open stream: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := s.Stat(ctx)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", s.ContentType())
	
	// http.ServeContent handles Range requests efficiently
	http.ServeContent(w, r, stat.Name, time.Time{}, file)
}

func (e *Engine) handleHLSPlaylist(w http.ResponseWriter, r *http.Request, id string) {
	playlistPath := fmt.Sprintf("hls/%s/playlist.m3u8", id)
	e.serveFromStorage(w, r, playlistPath, "application/vnd.apple.mpegurl")
}

func (e *Engine) handleHLSSegment(w http.ResponseWriter, r *http.Request, id, segmentName string) {
	segmentPath := fmt.Sprintf("hls/%s/%s", id, segmentName)
	e.serveFromStorage(w, r, segmentPath, "video/MP2T")
}

func (e *Engine) serveFromStorage(w http.ResponseWriter, r *http.Request, path, contentType string) {
	ctx := r.Context()
	
	// Cache lookup
	if e.cache != nil {
		if data, hit := e.cache.Get(path); hit {
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("X-Cache", "HIT")
			w.Write(data)
			return
		}
	}

	exists, _ := e.storage.Exists(ctx, path)
	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	file, err := e.storage.Open(ctx, path)
	if err != nil {
		http.Error(w, "Failed to open storage file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Cache", "MISS")

	// Use pool for memory-efficient transfer
	buf := e.bufferPool.Get().([]byte)
	defer e.bufferPool.Put(buf)

	// In a production engine, we would also populate the cache here if the file is small (e.g. segments)
	if strings.HasSuffix(path, ".m3u8") || strings.HasSuffix(path, ".ts") {
		data, _ := io.ReadAll(file)
		if e.cache != nil {
			e.cache.Set(path, data)
		}
		w.Write(data)
		return
	}

	io.CopyBuffer(w, file, buf)
}
