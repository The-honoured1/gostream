package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/The-honoured1/gostream"
	"github.com/The-honoured1/gostream/pkg/hls"
	"github.com/The-honoured1/gostream/pkg/storage"
)

func main() {
	// 1. Initialize Gostream
	engine := gostream.New()

	// 2. Register media sources (assuming these files exist)
	engine.Register("song", "media/sample.mp3")
	engine.Register("movie", "media/sample.mp4")

	// 3. Optional: Trigger HLS processing for video
	setupHLS(engine, "movie", "media/sample.mp4")

	// 4. Start the HTTP server (User-controlled)
	fmt.Println("🚀 Gostream engine ready. Serving on :8080")
	fmt.Println("🎵 Audio: http://localhost:8080/media/song/stream")
	fmt.Println("🎥 Video: http://localhost:8080/media/movie/stream")

	server := &http.Server{
		Addr:    ":8080",
		Handler: engine.Handler(),
	}

	log.Fatal(server.ListenAndServe())
}

func setupHLS(s *gostream.Server, id, path string) {
	st, _ := storage.NewLocalStorage("data")
	m, err := hls.NewManager(st)
	if err != nil {
		log.Printf("⚠️ HLS Manager cold not be initialized: %v", err)
		return
	}

	stream, _ := s.Get(id)
	go func() {
		if err := m.ProcessStream(context.Background(), id, stream); err != nil {
			log.Printf("❌ HLS processing failed for %s: %v", id, err)
		}
	}()
}
