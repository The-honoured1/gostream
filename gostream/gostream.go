package gostream

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/The-honoured1/gostream/core"
	"github.com/The-honoured1/gostream/internal/cache"
	"github.com/The-honoured1/gostream/pkg/delivery"
	"github.com/The-honoured1/gostream/pkg/storage"
	"github.com/The-honoured1/gostream/pkg/stream"
)

// Server is the high-level manager for the streaming engine.
type Server struct {
	mu        sync.RWMutex
	streams   map[string]core.MediaStream
	storage   core.Storage
	engine    *delivery.Engine
	cache     core.Cache
	logger    *log.Logger
}

// New creates a new Gostream server with production-grade defaults.
func New() *Server {
	s, _ := storage.NewLocalStorage("data")
	c := cache.NewLRUCache(100 * 1024 * 1024) // 100MB
	
	srv := &Server{
		streams: make(map[string]core.MediaStream),
		storage: s,
		cache:   c,
		logger:  log.New(log.Writer(), "[GOSTREAM] ", log.LstdFlags),
	}

	// Default engine with 1000 concurrent request limit
	srv.engine = delivery.NewEngine(srv, s, c, delivery.Options{
		MaxConcurrentRequests: 1000,
		CacheSize:            100 * 1024 * 1024,
	})

	return srv
}

// Register adds a file-based media source to the engine.
func (s *Server) Register(id, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	fs := stream.NewFileStream(path)
	s.streams[id] = fs
	s.logger.Printf("Registered stream '%s' from %s", id, path)
}

// RegisterMemory adds a memory-based media source to the engine.
func (s *Server) RegisterMemory(id string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	ms := stream.NewMemoryStream(id, data)
	s.streams[id] = ms
	s.logger.Printf("Registered memory stream '%s' (%d bytes)", id, len(data))
}

// AddStream adds a custom MediaStream implementation manually.
func (s *Server) AddStream(id string, stream core.MediaStream) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[id] = stream
	s.logger.Printf("Registered custom stream '%s'", id)
}

// Get implements core.Provider interface.
func (s *Server) Get(id string) (core.MediaStream, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.streams[id]
	return m, ok
}

// GetEngine returns the underlying delivery engine (primarily for testing).
func (s *Server) GetEngine() *delivery.Engine {
	return s.engine
}

// Start launches the streaming engine.
func (s *Server) Start(addr string) error {
	s.logger.Printf("Starting media engine on %s", addr)
	
	mux := http.NewServeMux()
	// All requests go through the delivery engine
	mux.Handle("/", s.engine)
	
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server.ListenAndServe()
}
