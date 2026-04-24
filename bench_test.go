package gostream_test

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/The-honoured1/gostream"
)

func BenchmarkConcurrent1000(b *testing.B) {
	// Setup server
	server := gostream.New()
	
	// Create a 1MB memory stream for testing
	data := make([]byte, 1024*1024)
	server.RegisterMemory("bench", data)

	// Setup httptest server using our engine
	ts := httptest.NewServer(server.GetEngine()) // We need to expose GetEngine() or similar
	defer ts.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		concurrency := 1000
		wg.Add(concurrency)

		for c := 0; c < concurrency; c++ {
			go func() {
				defer wg.Done()
				// Simulate a request for the stream
				res, err := ts.Client().Get(ts.URL + "/media/bench/stream")
				if err != nil {
					return
				}
				defer res.Body.Close()
				// Read only some data to keep it fast but valid
				buf := make([]byte, 4096)
				res.Body.Read(buf)
			}()
		}
		wg.Wait()
	}
}
