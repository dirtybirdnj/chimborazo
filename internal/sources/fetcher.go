package sources

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Fetcher downloads and caches remote files
type Fetcher struct {
	CacheDir string
	Client   *http.Client
	mu       sync.Mutex
}

// FetchResult contains fetch operation metadata
type FetchResult struct {
	Path      string    // Local file path
	FromCache bool      // True if served from cache
	Size      int64     // File size in bytes
	FetchedAt time.Time // Download timestamp
}

// FetchOption configures fetch behavior
type FetchOption func(*fetchOptions)

type fetchOptions struct {
	force   bool          // Bypass cache
	timeout time.Duration // Override default timeout
}

// NewFetcher creates a fetcher with the given cache directory
func NewFetcher(cacheDir string) (*Fetcher, error) {
	// Expand ~ to home directory
	if len(cacheDir) > 0 && cacheDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("expanding home directory: %w", err)
		}
		cacheDir = filepath.Join(home, cacheDir[1:])
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	return &Fetcher{
		CacheDir: cacheDir,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CachePath returns the cache path for a URL (without fetching)
func (f *Fetcher) CachePath(url string) string {
	hash := sha256.Sum256([]byte(url))
	filename := hex.EncodeToString(hash[:]) + filepath.Ext(url)
	return filepath.Join(f.CacheDir, filename)
}

// IsCached returns true if the URL is in cache
func (f *Fetcher) IsCached(url string) bool {
	path := f.CachePath(url)
	_, err := os.Stat(path)
	return err == nil
}

// Fetch downloads a URL and returns the local path
// Returns cached file if available, downloads otherwise
func (f *Fetcher) Fetch(url string, opts ...FetchOption) (*FetchResult, error) {
	// Apply options
	options := &fetchOptions{
		force:   false,
		timeout: 0, // Use client default
	}
	for _, opt := range opts {
		opt(options)
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

	// Create a client with custom timeout if specified
	client := f.Client
	if options.timeout > 0 {
		client = &http.Client{
			Timeout: options.timeout,
		}
	}

	// Download the file
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: HTTP %d %s", url, resp.StatusCode, resp.Status)
	}

	// Write to temporary file first (atomic write)
	f.mu.Lock()
	defer f.mu.Unlock()

	tmpPath := cachePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}

	size, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("writing %s: %w", url, err)
	}
	tmpFile.Close()

	// Rename temp file to final path
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

// WithForce bypasses the cache and re-downloads
func WithForce() FetchOption {
	return func(o *fetchOptions) {
		o.force = true
	}
}

// WithTimeout sets a custom timeout for this fetch
func WithTimeout(d time.Duration) FetchOption {
	return func(o *fetchOptions) {
		o.timeout = d
	}
}

// Clear removes a URL from cache
func (f *Fetcher) Clear(url string) error {
	path := f.CachePath(url)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // Already cleared
	}
	return err
}

// ClearAll removes all cached files
func (f *Fetcher) ClearAll() error {
	entries, err := os.ReadDir(f.CacheDir)
	if err != nil {
		return fmt.Errorf("reading cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(f.CacheDir, entry.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", path, err)
		}
	}
	return nil
}
