package hls

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/The-honoured1/gostream/core"
)

// Manager handles the lifecycle of HLS segments and playlists.
type Manager struct {
	storage core.Storage
	ffmpeg  string
	logger  *log.Logger
}

func NewManager(s core.Storage) (*Manager, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}
	return &Manager{
		storage: s,
		ffmpeg:  path,
		logger:  log.New(log.Writer(), "[HLS] ", log.LstdFlags),
	}, nil
}

// ProcessStream converts a MediaStream into HLS segments if possible.
func (m *Manager) ProcessStream(ctx context.Context, id string, s core.MediaStream) error {
	// Check if already processed
	playlistPath := fmt.Sprintf("hls/%s/playlist.m3u8", id)
	if exists, _ := m.storage.Exists(ctx, playlistPath); exists {
		m.logger.Printf("HLS already exists for %s", id)
		return nil
	}

	// Output directory in storage
	// For LocalStorage, we can just use the path. For S3, we'd need a temp dir.
	outputDir := filepath.Join("data", "hls", id)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// If it's a FileStream, we have the path.
	// We'll use a hack here to get the path if it's a file stream.
	// In a real production system, MediaStream would have a Source() method.
	sourcePath := ""
	if fs, ok := s.(interface{ GetPath() string }); ok {
		sourcePath = fs.GetPath()
	} else {
		return fmt.Errorf("HLS only supported for file-backed streams in this version")
	}

	m.logger.Printf("Starting HLS segmentation for %s...", id)

	// ffmpeg command
	cmd := exec.CommandContext(ctx, m.ffmpeg,
		"-i", sourcePath,
		"-codec:", "copy",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		"-hls_segment_filename", filepath.Join(outputDir, "seg%d.ts"),
		filepath.Join(outputDir, "playlist.m3u8"),
	)

	return cmd.Run()
}
