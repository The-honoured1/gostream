package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements the Storage interface using the local filesystem.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new LocalStorage instance with a specified base directory.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Ensure the base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) getFullPath(path string) string {
	return filepath.Join(s.basePath, path)
}

func (s *LocalStorage) Open(ctx context.Context, path string) (io.ReadSeekCloser, error) {
	fullPath := s.getFullPath(path)
	return os.Open(fullPath)
}

func (s *LocalStorage) Save(ctx context.Context, path string, data io.Reader) error {
	fullPath := s.getFullPath(path)
	
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	return err
}

func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := s.getFullPath(path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := s.getFullPath(path)
	return os.Remove(fullPath)
}

func (s *LocalStorage) Stat(ctx context.Context, path string) (FileInfo, error) {
	fullPath := s.getFullPath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Size: uint64(info.Size()),
		Name: info.Name(),
	}, nil
}
