package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/The-honoured1/gostream/core"
)

// LocalStorage implements the core.Storage interface for the local filesystem.
type LocalStorage struct {
	basePath string
	mu       sync.RWMutex
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) getPath(path string) string {
	return filepath.Join(s.basePath, path)
}

func (s *LocalStorage) Save(ctx context.Context, path string, data io.Reader) error {
	fullPath := s.getPath(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	return err
}

func (s *LocalStorage) Open(ctx context.Context, path string) (io.ReadSeekCloser, error) {
	fullPath := s.getPath(path)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return os.Open(fullPath)
}

func (s *LocalStorage) Stat(ctx context.Context, path string) (core.FileInfo, error) {
	fullPath := s.getPath(path)
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, err := os.Stat(fullPath)
	if err != nil {
		return core.FileInfo{}, err
	}

	return core.FileInfo{
		ID:   path,
		Size: info.Size(),
		Name: info.Name(),
	}, nil
}

func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := s.getPath(path)
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
