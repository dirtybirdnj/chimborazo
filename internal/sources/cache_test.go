package sources

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCachePathForURL(t *testing.T) {
	tests := []struct {
		name     string
		cacheDir string
		url      string
		wantExt  string
	}{
		{
			name:     "zip extension",
			cacheDir: "/tmp/cache",
			url:      "https://example.com/data.zip",
			wantExt:  ".zip",
		},
		{
			name:     "json extension",
			cacheDir: "/tmp/cache",
			url:      "https://api.example.com/data.json",
			wantExt:  ".json",
		},
		{
			name:     "no extension",
			cacheDir: "/tmp/cache",
			url:      "https://example.com/data",
			wantExt:  "",
		},
		{
			name:     "geojson extension",
			cacheDir: "/home/user/.cache",
			url:      "https://geo.example.com/file.geojson",
			wantExt:  ".geojson",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CachePathForURL(tt.cacheDir, tt.url)

			// Check it starts with cache dir
			if !strings.HasPrefix(got, tt.cacheDir) {
				t.Errorf("CachePathForURL() = %q, want prefix %q", got, tt.cacheDir)
			}

			// Check extension is preserved
			if filepath.Ext(got) != tt.wantExt {
				t.Errorf("CachePathForURL() ext = %q, want %q", filepath.Ext(got), tt.wantExt)
			}

			// Check hash is 64 chars (SHA256 hex = 64 chars)
			base := filepath.Base(got)
			hashPart := strings.TrimSuffix(base, tt.wantExt)
			if len(hashPart) != 64 {
				t.Errorf("CachePathForURL() hash length = %d, want 64", len(hashPart))
			}
		})
	}
}

func TestCachePathForURL_DifferentURLs(t *testing.T) {
	cacheDir := "/tmp/cache"
	url1 := "https://example.com/file1.zip"
	url2 := "https://example.com/file2.zip"

	path1 := CachePathForURL(cacheDir, url1)
	path2 := CachePathForURL(cacheDir, url2)

	if path1 == path2 {
		t.Errorf("Different URLs should produce different paths: %q == %q", path1, path2)
	}
}

func TestCachePathForURL_Deterministic(t *testing.T) {
	cacheDir := "/tmp/cache"
	url := "https://example.com/data.zip"

	path1 := CachePathForURL(cacheDir, url)
	path2 := CachePathForURL(cacheDir, url)

	if path1 != path2 {
		t.Errorf("Same URL should produce same path: %q != %q", path1, path2)
	}
}
