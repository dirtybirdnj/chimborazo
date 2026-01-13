package sources

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

// CachePathForURL converts a URL to a local cache file path using SHA256 hash.
// The hash ensures unique filenames while preserving the original extension.
//
// Example:
//
//	CachePathForURL("/tmp/cache", "https://example.com/data.zip")
//	// Returns: /tmp/cache/abc123...def.zip
func CachePathForURL(cacheDir, url string) string {
	ext := filepath.Ext(url)
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return filepath.Join(cacheDir, hash+ext)
}
