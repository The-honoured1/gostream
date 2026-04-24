package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/The-honoured1/gostream/handler"
	"github.com/The-honoured1/gostream/internal/cache"
	"github.com/The-honoured1/gostream/segmenter"
	"github.com/The-honoured1/gostream/storage"
)

// Server represents the Gostream media server.
type Server struct {
	storage   storage.Storage
	segmenter *segmenter.Segmenter
	cache     *cache.MemoryCache
	
	videos    map[string]string // id -> original source path
	hlsDirs   map[string]string // id -> directory containing HLS segments
	mu        sync.RWMutex

	mux       *http.ServeMux
	logger    *log.Logger
}

// New creates a new instance of the Gostream server with default settings.
func New() *Server {
	// Use local storage by default in the 'data' directory
	s, _ := storage.NewLocalStorage("data")
	seg, _ := segmenter.New()
	
	srv := &Server{
		storage:   s,
		segmenter: seg,
		cache:     cache.NewMemoryCache(5 * time.Minute),
		videos:    make(map[string]string),
		hlsDirs:   make(map[string]string),
		mux:       http.NewServeMux(),
		logger:    log.New(os.Stdout, "[GOSTREAM] ", log.LstdFlags),
	}

	srv.setupRoutes()
	return srv
}

func (s *Server) setupRoutes() {
	// VOD Range-based streaming: /video/{id}/stream
	s.mux.HandleFunc("/video/", s.handleRawStream)
	
	// HLS Playlists & Segments: /hls/{id}/playlist.m3u8, /hls/{id}/segment/{n}.ts
	s.mux.HandleFunc("/hls/", s.handleHLS)
}

func (s *Server) handleRawStream(w http.ResponseWriter, r *http.Request) {
	pathParts := filterEmpty(strings.Split(r.URL.Path, "/"))
	if len(pathParts) < 3 || pathParts[2] != "stream" {
		http.Error(w, "invalid stream URL. Expected /video/{id}/stream", http.StatusBadRequest)
		return
	}

	id := pathParts[1]
	s.mu.RLock()
	originalPath, exists := s.videos[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("video %s not found", id), http.StatusNotFound)
		return
	}

	// For raw streaming, we use the storage handler but we need to pass a storage that's pointing to the root
	// or make a specialized handler. For simplicity here, we'll open the file directly.
	// But let's stay idiomatic with our Storage interface.
	
	// We need to 'Save' the original video into our managed storage if it's not already there?
	// No, the library user provides a path. Let's assume we can Open it.
	// Actually, let's create a temporary Storage for the original videos or just allow absolute paths.
	
	// Better: Use a simple ServeFile for the original path if it's on disk.
	http.ServeFile(w, r, originalPath)
}

func (s *Server) handleHLS(w http.ResponseWriter, r *http.Request) {
	pathParts := filterEmpty(strings.Split(r.URL.Path, "/"))
	if len(pathParts) < 2 {
		http.Error(w, "invalid HLS URL", http.StatusBadRequest)
		return
	}

	id := pathParts[1]
	filename := pathParts[len(pathParts)-1]

	// In data/hls/{id}/ we have the files
	hlsFilePath := filepath.Join("hls", id, filename)
	
	ctx := r.Context()
	exists, err := s.storage.Exists(ctx, hlsFilePath)
	if err != nil || !exists {
		s.logger.Printf("HLS file not found: %s", hlsFilePath)
		http.Error(w, "HLS file not found", http.StatusNotFound)
		return
	}

	file, err := s.storage.Open(ctx, hlsFilePath)
	if err != nil {
		http.Error(w, "failed to open HLS file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if strings.HasSuffix(filename, ".m3u8") {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	} else if strings.HasSuffix(filename, ".ts") {
		w.Header().Set("Content-Type", "video/MP2T")
	}
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.Copy(w, file)
}

func (s *Server) RegisterVideo(id, path string) error {
	s.mu.Lock()
	s.videos[id] = path
	s.mu.Unlock()

	s.logger.Printf("Registered video: %s -> %s", id, path)

	// Trigger HLS generation in background
	go func() {
		// Output to storage-relative path
		outputDir := filepath.Join("data", "hls", id) 
		s.logger.Printf("Starting HLS segmentation for %s...", id)
		
		opts := segmenter.DefaultOptions()
		err := s.segmenter.Process(context.Background(), path, outputDir, opts)
		if err != nil {
			s.logger.Printf("HLS segmentation failed for %s: %v", id, err)
		} else {
			s.logger.Printf("HLS segmentation completed for %s", id)
		}
	}()

	return nil
}

func (s *Server) Start(addr string) error {
	s.logger.Printf("Gostream engine starting on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func filterEmpty(ss []string) []string {
	var r []string
	for _, s := range ss {
		if s != "" {
			r = append(r, s)
		}
	}
	return r
}
