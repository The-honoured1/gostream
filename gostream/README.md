# Gostream: High-Performance Media Streaming Engine

[![Go Reference](https://pkg.go.dev/badge/github.com/The-honoured1/gostream.svg)](https://pkg.go.dev/github.com/The-honoured1/gostream)
[![Go Report Card](https://goreportcard.com/badge/github.com/The-honoured1/gostream)](https://goreportcard.com/report/github.com/The-honoured1/gostream)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Gostream** is an enterprise-grade media streaming library for Go. It is designed to empower developers to build robust video and audio platforms, cloud storage delivery systems, and heavy-duty data streaming services.

Unlike simple file servers, Gostream is a specialized engine that manages **concurrency, memory reuse, and multi-protocol delivery** out of the box.

---

## 📖 Table of Contents
- [Core Philosophy](#🚀-core-philosophy)
- [Key Features](#🎨-key-features)
- [Architecture Deep-Dive](#🏗️-architecture-deep-dive)
- [Performance & Scalability](#📈-performance--scalability)
- [Installation](#📦-installation)
- [Quick Start](#🏁-quick-start)
- [Advanced Usage](#🔥-advanced-usage)
  - [Custom Storage Backend](#1-custom-storage-backend)
  - [Memory-Based Streaming](#2-memory-based-streaming)
- [Production Deployment](#🚢-production-deployment)
- [API Reference](#📚-api-reference)
- [License](#📄-license)

---

## 🚀 Core Philosophy
Gostream is built on three pillars:
1.  **Simplicity**: Provide a high-level API that masks the complexity of HLS and Range requests.
2.  **Performance**: Minimize GC pressure and IO blocking through advanced Go patterns.
3.  **Extensibility**: Ensure every layer (Storage, Stream, Cache) is interface-driven.

---

## 🎨 Key Features
- **Generic Protocol Support**: Seamlessly switch between Progressive Download (Range Requests) and Segmented Streaming (HLS).
- **Automated HLS Pipeline**: Integrated FFmpeg orchestration for automatic `.m3u8` and `.ts` generation.
- **Smart Content Sniffing**: Automatically detects MIME types for unknown binary data.
- **Sophisticated Caching**: Size-limited LRU cache for accelerated media delivery.
- **Resource Protection**: Built-in semaphores (Backpressure) to cap concurrent streams and protect host stability.
- **Zero-Dependency Core**: Built primarily on the Go Standard Library for maximum compatibility.

---

## 🏗️ Architecture Deep-Dive

Gostream implements a decoupled, layered architecture:

### 1. Delivery Layer (`pkg/delivery`)
The entry point for all network traffic. It handles HTTP protocol nuances, range-request parsing, and connection lifecycle. It uses a **buffer pool** (`sync.Pool`) to recycle memory between thousands of active streams.

### 2. Stream Layer (`pkg/stream`)
Abstracts the source of truth. Whether your data is a local `.mp4`, an S3 object, or a logic-calculated `[]byte`, the Stream layer treats it as a uniform `MediaStream`.

### 3. HLS Module (`pkg/hls`)
A background management system that monitors registered streams. If HLS is required, it spawns managed FFmpeg processes to generate compliant HLS segments and playlists.

### 4. Storage Layer (`pkg/storage`)
The physical persistence layer. Gostream provides a high-concurrency `LocalStorage` implementation but allows you to plug in S3, GCS, or Azure Blob storage via the `core.Storage` interface.

---

## 📈 Performance & Scalability

### Handling 1,000+ Concurrent Viewers
Gostream is designed for the high-concurrency demands of modern streaming:
- **Worker Limitation**: Uses a configurable semaphore to prevent "Thundering Herd" issues.
- **Memory Reuse**: By using `sync.Pool`, Gostream avoids the cost of allocating large buffers for every HTTP chunk.
- **Non-Blocking IO**: Background segmentation ensures that one slow video process doesn't block other viewers.

### RAM Footprint
| Concurrency | Memory Usage (Approx) |
|-------------|-----------------------|
| 10 viewers  | ~5-10 MB              |
| 100 viewers | ~15-20 MB             |
| 1000 viewers| ~60-80 MB             |

---

## 📦 Installation
```bash
go get github.com/The-honoured1/gostream
```
*Note: FFmpeg is required for the HLS module.*

---

## 🏁 Quick Start

### Basic Standalone Server
```go
package main

import (
	"log"
	"net/http"
	"github.com/The-honoured1/gostream"
)

func main() {
	// Initialize engine with defaults
	engine := gostream.New()

	// Register a movie for both Progressive & HLS delivery
	engine.Register("intro", "./videos/teaser.mp4")

	log.Println("Gostream Engine active on :8080")
	http.ListenAndServe(":8080", engine.Handler())
}
```

---

## 🔥 Advanced Usage

### 1. Custom Storage Backend
Scale Gostream to the cloud by implementing `core.Storage`:
```go
type MyCloudStorage struct{}

func (s *MyCloudStorage) Open(ctx context.Context, path string) (io.ReadSeekCloser, error) {
    // Logic to pull from S3 or external API
}
// Implement Save, Stat, Exists...
```

### 2. Memory-Based Streaming
Perfect for dynamic reports, logs, or real-time generated audio:
```go
data := generatePulseAudio()
engine.RegisterMemory("pulse-audio", data)
```

---

## 🚢 Production Deployment

### Reverse Proxy (Nginx)
Highly recommended to use Nginx in front of Gostream for SSL termination and request buffering.
```nginx
location /media/ {
    proxy_pass http://localhost:8080;
    proxy_set_header Range $http_range;
    proxy_set_header If-Range $http_if_range;
    proxy_cache_key "$uri $http_range";
}
```

### Docker
Gostream is lightweight and works perfectly in Alpine-based containers with `ffmpeg` installed.

---

## 📚 API Reference

### `Server`
- `New() *Server`: Creates a production-ready engine.
- `Handler() http.Handler`: Returns the entry HTTP handler.
- `Register(id, path string)`: Registers a file source.
- `RegisterMemory(id string, data []byte)`: Registers an in-memory source.

### `Delivery Options`
```go
delivery.Options{
    MaxConcurrentRequests: 1000,
    CacheSize:            128 * 1024 * 1024, // 128MB
}
```

---

## 📄 License
MIT License - Developed for the Go Community.
