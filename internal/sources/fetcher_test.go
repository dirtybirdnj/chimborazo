package sources

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFetcher(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFetcher(dir)
	if err != nil {
		t.Fatalf("NewFetcher failed: %v", err)
	}
	if f.CacheDir != dir {
		t.Errorf("CacheDir = %q, want %q", f.CacheDir, dir)
	}
	if f.Client == nil {
		t.Error("Client is nil")
	}
}

func TestNewFetcher_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cache")
	_, err := NewFetcher(dir)
	if err != nil {
		t.Fatalf("NewFetcher failed: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Cache dir was not created")
	}
}

func TestCachePath(t *testing.T) {
	f, _ := NewFetcher(t.TempDir())

	tests := []struct {
		url     string
		wantExt string
	}{
		{"https://example.com/data.zip", ".zip"},
		{"https://example.com/file.geojson", ".geojson"},
		{"https://example.com/data.zip?token=abc", ".zip"},
		{"https://example.com/noext", ""},
	}

	for _, tt := range tests {
		path := f.CachePath(tt.url)
		ext := filepath.Ext(path)
		if ext != tt.wantExt {
			t.Errorf("CachePath(%q) ext = %q, want %q", tt.url, ext, tt.wantExt)
		}
		// Should be in cache dir
		if filepath.Dir(path) != f.CacheDir {
			t.Errorf("CachePath(%q) not in cache dir", tt.url)
		}
	}
}

func TestCachePath_Deterministic(t *testing.T) {
	f, _ := NewFetcher(t.TempDir())
	url := "https://example.com/test.zip"

	path1 := f.CachePath(url)
	path2 := f.CachePath(url)

	if path1 != path2 {
		t.Errorf("CachePath not deterministic: %q != %q", path1, path2)
	}
}

func TestFetch(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())

	// First fetch - should download
	result, err := f.Fetch(server.URL + "/test.txt")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if result.FromCache {
		t.Error("First fetch should not be from cache")
	}
	if result.Size != 12 {
		t.Errorf("Size = %d, want 12", result.Size)
	}

	// Second fetch - should be cached
	result2, err := f.Fetch(server.URL + "/test.txt")
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}
	if !result2.FromCache {
		t.Error("Second fetch should be from cache")
	}
}

func TestFetch_WithForce(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("content"))
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())

	// First fetch
	f.Fetch(server.URL + "/test.txt")
	if callCount != 1 {
		t.Fatalf("Expected 1 call, got %d", callCount)
	}

	// Second fetch without force - should use cache
	f.Fetch(server.URL + "/test.txt")
	if callCount != 1 {
		t.Errorf("Expected 1 call (cached), got %d", callCount)
	}

	// Third fetch with force - should re-download
	f.Fetch(server.URL+"/test.txt", WithForce())
	if callCount != 2 {
		t.Errorf("Expected 2 calls (forced), got %d", callCount)
	}
}

func TestFetch_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())

	_, err := f.Fetch(server.URL + "/missing.txt")
	if err == nil {
		t.Error("Expected error for 404")
	}
}

func TestFetch_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte("slow"))
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())

	_, err := f.Fetch(server.URL+"/slow.txt", WithTimeout(100*time.Millisecond))
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestIsCached(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())
	url := server.URL + "/test.txt"

	if f.IsCached(url) {
		t.Error("Should not be cached initially")
	}

	f.Fetch(url)

	if !f.IsCached(url) {
		t.Error("Should be cached after fetch")
	}
}

func TestClear(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())
	url := server.URL + "/test.txt"

	f.Fetch(url)
	if !f.IsCached(url) {
		t.Fatal("Should be cached")
	}

	f.Clear(url)
	if f.IsCached(url) {
		t.Error("Should not be cached after clear")
	}
}

func TestClearAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer server.Close()

	f, _ := NewFetcher(t.TempDir())

	f.Fetch(server.URL + "/a.txt")
	f.Fetch(server.URL + "/b.txt")

	f.ClearAll()

	if f.IsCached(server.URL+"/a.txt") || f.IsCached(server.URL+"/b.txt") {
		t.Error("Cache should be empty after ClearAll")
	}
}
