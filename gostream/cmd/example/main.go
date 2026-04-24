package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/The-honoured1/gostream"
	"github.com/The-honoured1/gostream/hls"
	"github.com/The-honoured1/gostream/storage"
)

func main() {
	// 1. Prepare sample files
	prepareDummyFiles()

	// 2. Initialize Gostream
	server := gostream.New()

	// 3. Register different media types
	server.Register("song", "media/sample.mp3")
	server.Register("movie", "media/sample.mp4")
	server.Register("archive", "media/sample.zip")

	// 4. Setup HLS for the movie (optional module)
	setupHLS(server, "movie", "media/sample.mp4")

	// 5. Start the engine
	fmt.Println("==================================================")
	fmt.Println("🚀 Gostream Advanced Engine")
	fmt.Println("==================================================")
	fmt.Println("🎵 Audio (Progressive): http://localhost:8080/media/song/stream")
	fmt.Println("🎥 Video (Progressive): http://localhost:8080/media/movie/stream")
	fmt.Println("📦 Archive (Binary):    http://localhost:8080/media/archive/stream")
	fmt.Println("📺 Video (HLS):         http://localhost:8080/media/movie/playlist.m3u8")
	fmt.Println("==================================================")

	log.Fatal(server.Start(":8080"))
}

func setupHLS(s *gostream.Server, id, path string) {
	// In a real system, this would be integrated into the server.Register process
	// but here we show the modularity.
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
		} else {
			log.Printf("✅ HLS processing completed for %s", id)
		}
	}()
}

func prepareDummyFiles() {
	os.MkdirAll("media", 0755)
	
	// Create a dummy MP3
	os.WriteFile("media/sample.mp3", make([]byte, 1024*1024), 0644)
	
	// Create a dummy ZIP
	os.WriteFile("media/sample.zip", make([]byte, 5*1024*1024), 0644)

	// Create a dummy MP4 (if ffmpeg is available)
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		fmt.Println("Generating sample MP4...")
		exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=10:size=640x360:rate=30", "-c:v", "libx264", "media/sample.mp4").Run()
	} else {
		fmt.Println("ffmpeg not found, creating empty sample.mp4")
		os.WriteFile("media/sample.mp4", make([]byte, 2*1024*1024), 0644)
	}
}
