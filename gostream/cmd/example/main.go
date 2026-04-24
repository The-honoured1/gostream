package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/The-honoured1/gostream/server"
)

func main() {
	// 1. Setup sample video
	videoPath := "videos/sample.mp4"
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		fmt.Println("Sample video not found. Attempting to generate a dummy video using ffmpeg...")
		err := generateDummyVideo(videoPath)
		if err != nil {
			log.Fatalf("Failed to generate dummy video: %v. Please provide a sample.mp4 in the videos/ directory.", err)
		}
	}

	// 2. Initialize Gostream server
	srv := server.New()

	// 3. Register video
	err := srv.RegisterVideo("intro", videoPath)
	if err != nil {
		log.Fatalf("Failed to register video: %v", err)
	}

	// 4. Start the server
	fmt.Println("--------------------------------------------------")
	fmt.Println("Gostream Demo Server")
	fmt.Println("VOD Stream: http://localhost:8080/video/intro/stream")
	fmt.Println("HLS Playlist: http://localhost:8080/hls/intro/playlist.m3u8")
	fmt.Println("--------------------------------------------------")
	
	log.Fatal(srv.Start(":8080"))
}

func generateDummyVideo(path string) error {
	// Create videos directory if not exists
	os.MkdirAll("videos", 0755)

	// Generate a 5-second dummy video with counters
	// ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 -c:v libx264 -pix_fmt yuv420p sample.mp4
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=30:size=640x360:rate=30", "-c:v", "libx264", "-pix_fmt", "yuv420p", path)
	return cmd.Run()
}
