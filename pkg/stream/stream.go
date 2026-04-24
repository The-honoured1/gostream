package stream

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/The-honoured1/gostream/core"
)

// FileStream implements core.MediaStream for file-backed resources.
type FileStream struct {
	path        string
	contentType string
}

func NewFileStream(path string) *FileStream {
	return &FileStream{
		path: path,
	}
}

func (f *FileStream) GetPath() string {
	return f.path
}

func (f *FileStream) Open(ctx context.Context) (io.ReadSeekCloser, error) {
	return os.Open(f.path)
}

func (f *FileStream) Stat(ctx context.Context) (core.FileInfo, error) {
	info, err := os.Stat(f.path)
	if err != nil {
		return core.FileInfo{}, err
	}
	return core.FileInfo{
		ID:   f.path,
		Size: info.Size(),
		Name: info.Name(),
	}, nil
}

func (f *FileStream) ContentType() string {
	if f.contentType != "" {
		return f.contentType
	}

	// Dynamic detection
	file, err := os.Open(f.path)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()

	// Read first 512 bytes for sniffing
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	f.contentType = http.DetectContentType(buffer[:n])
	
	return f.contentType
}

// MemoryStream implements core.MediaStream for in-memory data.
type MemoryStream struct {
	data        []byte
	name        string
	contentType string
}

func NewMemoryStream(name string, data []byte) *MemoryStream {
	return &MemoryStream{
		name: name,
		data: data,
	}
}

func (m *MemoryStream) Open(ctx context.Context) (io.ReadSeekCloser, error) {
	return &nopSeekCloser{io.NewSectionReader(memoryReader(m.data), 0, int64(len(m.data)))}, nil
}

func (m *MemoryStream) Stat(ctx context.Context) (core.FileInfo, error) {
	return core.FileInfo{
		ID:   m.name,
		Size: int64(len(m.data)),
		Name: m.name,
	}, nil
}

func (m *MemoryStream) ContentType() string {
	if m.contentType != "" {
		return m.contentType
	}
	m.contentType = http.DetectContentType(m.data)
	return m.contentType
}

// Helpers for MemoryStream
type memoryReader []byte
func (m memoryReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(m)) {
		return 0, io.EOF
	}
	n := copy(p, m[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

type nopSeekCloser struct {
	*io.SectionReader
}
func (n *nopSeekCloser) Close() error { return nil }
