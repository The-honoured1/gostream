package storage

import (
	"context"
	"io"
)

// Storage defines the interface for interacting with media storage layers.
// It is designed to be pluggable, allowing for local FS, S3, or other backends.
type Storage interface {
	// Open returns an io.ReadCloser for the specified path.
	// It's the caller's responsibility to close the reader.
	Open(ctx context.Context, path string) (io.ReadSeekCloser, error)

	// Save writes the data from the reader to the specified path.
	Save(ctx context.Context, path string, data io.Reader) error

	// Exists checks if the specified path exists in the storage.
	Exists(ctx context.Context, path string) (bool, error)

	// Delete removes the specified path from the storage.
	Delete(ctx context.Context, path string) error

	// Stat returns file information for the specified path.
	Stat(ctx context.Context, path string) (FileInfo, error)
}

// FileInfo provides metadata about a file in storage.
type FileInfo struct {
	Size uint64
	Name string
}

// ReadSeekCloser is an interface that groups the basic Read, Seek, and Close methods.
type ioReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// ReadSeekCloser is a type alias for the grouped interface.
type ReadSeekCloser = ioReadSeekCloser
