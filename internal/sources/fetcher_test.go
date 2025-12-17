package sources

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFetcher(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := NewFetcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFetcher() error = %v", err)
	}

	if f.CacheDir != tmpDir {
		t.Errorf("CacheDir = %q, want %q", f.CacheDir, tmpDir)
	}

	if f.Client == nil {
		t.Error("Client is nil")
	}
}

func TestNewFetcher_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "nested", "cache")

	f, err := NewFetcher(cacheDir)
	if err != nil {
		t.Fatalf("NewFetcher() error = %v", err)
	}

	if _, err := os.Stat(f.CacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestFetcher_CachePath(t *testing.T) {
	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	tests := []struct {
		name string
		url  string
		ext  string
	}{
		{
			name: "zip extension",
			url:  "https://example.com/data.zip",
			ext:  ".zip",
		},
		{
			name: "json extension",
			url:  "https://example.com/config.json",
			ext:  ".json",
		},
		{
			name: "no extension",
			url:  "https://example.com/data",
			ext:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := f.CachePath(tt.url)

			// Should be in cache directory
			if !strings.HasPrefix(path, tmpDir) {
				t.Errorf("Path %q not in cache dir %q", path, tmpDir)
			}

			// Should have correct extension
			if filepath.Ext(path) != tt.ext {
				t.Errorf("Extension = %q, want %q", filepath.Ext(path), tt.ext)
			}

			// Same URL should give same path
			path2 := f.CachePath(tt.url)
			if path != path2 {
				t.Errorf("Same URL gave different paths: %q vs %q", path, path2)
			}
		})
	}

	// Different URLs should give different paths
	path1 := f.CachePath("https://example.com/a")
	path2 := f.CachePath("https://example.com/b")
	if path1 == path2 {
		t.Error("Different URLs gave same path")
	}
}

func TestFetcher_Fetch(t *testing.T) {
	// Create test server
	content := "test content for fetcher"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	result, err := f.Fetch(server.URL + "/test.txt")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if result.FromCache {
		t.Error("First fetch should not be from cache")
	}

	if result.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", result.Size, len(content))
	}

	// Verify file contents
	data, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("Reading cached file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Content = %q, want %q", string(data), content)
	}
}

func TestFetcher_FetchCached(t *testing.T) {
	content := "cached content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	url := server.URL + "/cached.txt"

	// First fetch
	result1, err := f.Fetch(url)
	if err != nil {
		t.Fatalf("First Fetch() error = %v", err)
	}
	if result1.FromCache {
		t.Error("First fetch should not be from cache")
	}

	// Second fetch should be from cache
	result2, err := f.Fetch(url)
	if err != nil {
		t.Fatalf("Second Fetch() error = %v", err)
	}
	if !result2.FromCache {
		t.Error("Second fetch should be from cache")
	}
	if result2.Path != result1.Path {
		t.Errorf("Paths differ: %q vs %q", result1.Path, result2.Path)
	}
}

func TestFetcher_FetchForce(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	url := server.URL + "/force.txt"

	// First fetch
	_, err := f.Fetch(url)
	if err != nil {
		t.Fatalf("First Fetch() error = %v", err)
	}
	if callCount != 1 {
		t.Errorf("Call count = %d, want 1", callCount)
	}

	// Second fetch without force (should use cache)
	_, err = f.Fetch(url)
	if err != nil {
		t.Fatalf("Second Fetch() error = %v", err)
	}
	if callCount != 1 {
		t.Errorf("Call count = %d, want 1 (should use cache)", callCount)
	}

	// Third fetch with force (should download again)
	result, err := f.Fetch(url, WithForce())
	if err != nil {
		t.Fatalf("Third Fetch() error = %v", err)
	}
	if callCount != 2 {
		t.Errorf("Call count = %d, want 2 (force should bypass cache)", callCount)
	}
	if result.FromCache {
		t.Error("Force fetch should not report from cache")
	}
}

func TestFetcher_FetchNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	_, err := f.Fetch(server.URL + "/missing.txt")
	if err == nil {
		t.Fatal("Expected error for 404")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Error should contain 404: %v", err)
	}
}

func TestFetcher_FetchTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte("slow"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	_, err := f.Fetch(server.URL+"/slow.txt", WithTimeout(100*time.Millisecond))
	if err == nil {
		t.Fatal("Expected timeout error")
	}
}

func TestFetcher_IsCached(t *testing.T) {
	content := "test"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	url := server.URL + "/check.txt"

	// Not cached initially
	if f.IsCached(url) {
		t.Error("Should not be cached initially")
	}

	// Fetch it
	_, err := f.Fetch(url)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Now it should be cached
	if !f.IsCached(url) {
		t.Error("Should be cached after fetch")
	}
}

func TestFetcher_Clear(t *testing.T) {
	content := "to delete"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	url := server.URL + "/delete.txt"

	// Fetch it
	_, err := f.Fetch(url)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Clear it
	if err := f.Clear(url); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Should no longer be cached
	if f.IsCached(url) {
		t.Error("Should not be cached after clear")
	}

	// Clear again should not error
	if err := f.Clear(url); err != nil {
		t.Fatalf("Clear() on missing file error = %v", err)
	}
}

func TestFetcher_ClearAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	f, _ := NewFetcher(tmpDir)

	// Fetch multiple files
	urls := []string{
		server.URL + "/a.txt",
		server.URL + "/b.txt",
		server.URL + "/c.txt",
	}

	for _, url := range urls {
		if _, err := f.Fetch(url); err != nil {
			t.Fatalf("Fetch(%s) error = %v", url, err)
		}
	}

	// Verify all cached
	for _, url := range urls {
		if !f.IsCached(url) {
			t.Errorf("%s should be cached", url)
		}
	}

	// Clear all
	if err := f.ClearAll(); err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}

	// Verify all cleared
	for _, url := range urls {
		if f.IsCached(url) {
			t.Errorf("%s should not be cached after ClearAll", url)
		}
	}
}
