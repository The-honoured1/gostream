package segmenter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Options defines configuration for the segmenting process.
type Options struct {
	SegmentTime int    // Duration of each segment in seconds
	OutputPrefix string // Prefix for segment files
	PlaylistName string // Name of the m3u8 file
}

// DefaultOptions returns a set of sensible default options.
func DefaultOptions() Options {
	return Options{
		SegmentTime:  10,
		OutputPrefix: "segment",
		PlaylistName: "playlist.m3u8",
	}
}

// Segmenter handles the conversion of video files into HLS segments using FFmpeg.
type Segmenter struct {
	ffmpegPath string
}

// New creates a new Segmenter. It checks if ffmpeg is available in the PATH.
func New() (*Segmenter, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}
	return &Segmenter{ffmpegPath: path}, nil
}

// Process converts an input video file into HLS segments and a playlist.
// The output files are saved into the provided outputDir.
func (s *Segmenter) Process(ctx context.Context, inputPath, outputDir string, opts Options) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	playlistPath := filepath.Join(outputDir, opts.PlaylistName)
	segmentPattern := filepath.Join(outputDir, fmt.Sprintf("%s%%d.ts", opts.OutputPrefix))

	// ffmpeg -i input.mp4 -codec: copy -start_number 0 -hls_time 10 -hls_list_size 0 -f hls playlist.m3u8
	args := []string{
		"-i", inputPath,
		"-codec:", "copy", // Use copy to avoid re-encoding (efficient)
		"-start_number", "0",
		"-hls_time", fmt.Sprintf("%d", opts.SegmentTime),
		"-hls_list_size", "0", // 0 means include all segments (VOD)
		"-f", "hls",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	}

	cmd := exec.CommandContext(ctx, s.ffmpegPath, args...)
	
	// Set up logging or stdout/stderr if needed
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	return nil
}
