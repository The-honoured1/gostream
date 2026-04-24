package core

import (
	"context"
	"io"
)

// MediaStream is the core abstraction for any streamable content.
type MediaStream interface {
	// Open returns an io.ReadSeekCloser for the stream content.
	Open(ctx context.Context) (io.ReadSeekCloser, error)
	// Stat returns metadata about the stream.
	Stat(ctx context.Context) (FileInfo, error)
	// ContentType returns the MIME type of the stream.
	ContentType() string
}

// FileInfo provides metadata about a stream or file.
type FileInfo struct {
	ID   string
	Size int64
	Name string
}

// Storage handles the persistence of media and segments.
type Storage interface {
	Save(ctx context.Context, path string, data io.Reader) error
	Open(ctx context.Context, path string) (io.ReadSeekCloser, error)
	Stat(ctx context.Context, path string) (FileInfo, error)
	Exists(ctx context.Context, path string) (bool, error)
}

// Cache defines the interface for the hot-data acceleration layer.
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte)
	Delete(key string)
}

// Provider resolves a unique ID into a MediaStream.
type Provider interface {
	Get(id string) (MediaStream, bool)
}
