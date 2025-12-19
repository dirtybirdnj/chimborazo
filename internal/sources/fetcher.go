// Package sources provides data acquisition for chimborazo.
// It replaces the thoreau module from the Python strata project.
package sources

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Fetcher downloads and caches remote files.
type Fetcher struct {
	CacheDir string
	Client   *http.Client
}

// FetchResult contains fetch operation metadata.
type FetchResult struct {
	Path      string    // Local file path
	FromCache bool      // True if served from cache
	Size      int64     // File size in bytes
	FetchedAt time.Time // Download or cache hit timestamp
}

// FetchOption configures fetch behavior.
type FetchOption func(*fetchOptions)

type fetchOptions struct {
	force   bool
	timeout time.Duration
}

// NewFetcher creates a fetcher with the given cache directory.
// Creates the directory if it doesn't exist.
func NewFetcher(cacheDir string) (*Fetcher, error) {
	// Expand ~ if present
	if len(cacheDir) > 0 && cacheDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("expanding home dir: %w", err)
		}
		cacheDir = filepath.Join(home, cacheDir[1:])
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	return &Fetcher{
		CacheDir: cacheDir,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CachePath returns the cache path for a URL without fetching.
// Uses SHA256 hash of URL and preserves the file extension.
func (f *Fetcher) CachePath(url string) string {
	hash := sha256.Sum256([]byte(url))
	ext := filepath.Ext(url)
	// Handle URLs with query strings: /file.zip?token=xxx -> .zip
	if idx := len(ext) - 1; idx > 0 {
		for i := 0; i < len(ext); i++ {
			if ext[i] == '?' || ext[i] == '&' || ext[i] == '#' {
				ext = ext[:i]
				break
			}
		}
	}
	filename := hex.EncodeToString(hash[:]) + ext
	return filepath.Join(f.CacheDir, filename)
}

// IsCached returns true if the URL is in cache.
func (f *Fetcher) IsCached(url string) bool {
	_, err := os.Stat(f.CachePath(url))
	return err == nil
}

// Fetch downloads a URL and returns the local path.
// Returns cached file if available, downloads otherwise.
func (f *Fetcher) Fetch(url string, opts ...FetchOption) (*FetchResult, error) {
	// Apply options
	options := &fetchOptions{
		timeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Override client timeout if specified
	client := f.Client
	if options.timeout != 30*time.Second {
		client = &http.Client{Timeout: options.timeout}
	}

	cachePath := f.CachePath(url)

	// Check cache unless force is set
	if !options.force {
		if info, err := os.Stat(cachePath); err == nil {
			return &FetchResult{
				Path:      cachePath,
				FromCache: true,
				Size:      info.Size(),
				FetchedAt: info.ModTime(),
			}, nil
		}
	}

	// Download
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "chimborazo/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}

	// Write to temp file first (atomic write)
	tmpPath := cachePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}

	size, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("writing cache file: %w", err)
	}
	tmpFile.Close()

	// Rename to final path (atomic on POSIX)
	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("finalizing cache file: %w", err)
	}

	return &FetchResult{
		Path:      cachePath,
		FromCache: false,
		Size:      size,
		FetchedAt: time.Now(),
	}, nil
}

// WithForce bypasses the cache and re-downloads.
func WithForce() FetchOption {
	return func(o *fetchOptions) {
		o.force = true
	}
}

// WithTimeout sets a custom timeout for this fetch.
func WithTimeout(d time.Duration) FetchOption {
	return func(o *fetchOptions) {
		o.timeout = d
	}
}

// Clear removes a URL from cache.
func (f *Fetcher) Clear(url string) error {
	path := f.CachePath(url)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing cache for %s: %w", url, err)
	}
	return nil
}

// ClearAll removes all cached files.
func (f *Fetcher) ClearAll() error {
	entries, err := os.ReadDir(f.CacheDir)
	if err != nil {
		return fmt.Errorf("reading cache dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(f.CacheDir, entry.Name()))
		}
	}
	return nil
}
