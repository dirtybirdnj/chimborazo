# Spec 001: HTTP Fetcher

**Status**: Ready for Implementation
**Module**: `internal/sources`
**Assignee**: Local LLM via clood

## Overview

Implement an HTTP fetcher that downloads remote files and caches them locally. This is the foundation for all data acquisition in Chimborazo.

## Requirements

### Functional
1. Download files from HTTP/HTTPS URLs
2. Cache downloaded files locally
3. Return cached file if it exists (skip download)
4. Support forced re-download (bypass cache)
5. Handle HTTP errors gracefully

### Non-Functional
1. Configurable timeout (default 30s)
2. Configurable cache directory
3. Thread-safe for concurrent downloads

## Type Definitions

```go
package sources

import (
    "net/http"
    "time"
)

// Fetcher downloads and caches remote files
type Fetcher struct {
    CacheDir string
    Client   *http.Client
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
```

## Function Signatures

```go
// NewFetcher creates a fetcher with the given cache directory
func NewFetcher(cacheDir string) (*Fetcher, error)

// Fetch downloads a URL and returns the local path
// Returns cached file if available, downloads otherwise
func (f *Fetcher) Fetch(url string, opts ...FetchOption) (*FetchResult, error)

// WithForce bypasses the cache and re-downloads
func WithForce() FetchOption

// WithTimeout sets a custom timeout for this fetch
func WithTimeout(d time.Duration) FetchOption

// CachePath returns the cache path for a URL (without fetching)
func (f *Fetcher) CachePath(url string) string

// IsCached returns true if the URL is in cache
func (f *Fetcher) IsCached(url string) bool

// Clear removes a URL from cache
func (f *Fetcher) Clear(url string) error

// ClearAll removes all cached files
func (f *Fetcher) ClearAll() error
```

## Cache Path Strategy

Convert URL to cache path using SHA256 hash:

```go
func (f *Fetcher) CachePath(url string) string {
    hash := sha256.Sum256([]byte(url))
    filename := hex.EncodeToString(hash[:]) + filepath.Ext(url)
    return filepath.Join(f.CacheDir, filename)
}
```

Example:
- URL: `https://www2.census.gov/geo/tiger/TIGER2023/STATE/tl_2023_us_state.zip`
- Hash: `a3f2b1c4...` (truncated)
- Path: `~/.cache/chimborazo/a3f2b1c4...d5e6f7.zip`

## Behavior

### Fetch Flow

```
Fetch(url) called
    │
    ├─► Check cache
    │       │
    │       ├─► Exists? Return FetchResult{FromCache: true}
    │       │
    │       └─► Missing? Continue to download
    │
    ├─► HTTP GET request
    │       │
    │       ├─► Success (200)? Write to cache, return result
    │       │
    │       └─► Error? Return error (don't cache)
    │
    └─► Return FetchResult
```

### Error Cases

| Condition | Behavior |
|-----------|----------|
| Invalid URL | Return error immediately |
| Network error | Return error, don't cache |
| HTTP 4xx/5xx | Return error with status code |
| Disk full | Return error, clean up partial file |
| Timeout | Return error |

## Example Usage

```go
fetcher, err := sources.NewFetcher("~/.cache/chimborazo")
if err != nil {
    log.Fatal(err)
}

// First call downloads
result, err := fetcher.Fetch("https://example.com/data.zip")
// result.FromCache == false

// Second call uses cache
result, err = fetcher.Fetch("https://example.com/data.zip")
// result.FromCache == true

// Force re-download
result, err = fetcher.Fetch("https://example.com/data.zip", sources.WithForce())
// result.FromCache == false
```

## Tests Required

```go
func TestFetcher_Fetch(t *testing.T)           // Happy path
func TestFetcher_FetchCached(t *testing.T)     // Cache hit
func TestFetcher_FetchForce(t *testing.T)      // Force bypass
func TestFetcher_FetchNotFound(t *testing.T)   // 404 handling
func TestFetcher_FetchTimeout(t *testing.T)    // Timeout handling
func TestFetcher_CachePath(t *testing.T)       // Path generation
func TestFetcher_Clear(t *testing.T)           // Cache removal
```

Use `httptest.NewServer` for testing HTTP behavior.

## Constraints

- DO NOT use external HTTP libraries (use `net/http`)
- DO NOT add logging (return errors instead)
- DO NOT modify files outside CacheDir
- DO use `filepath` for path operations
- DO use `crypto/sha256` for hashing
- DO handle partial downloads (clean up on error)

## Files to Create

```
internal/sources/
├── doc.go           # Package documentation
├── fetcher.go       # Implementation
└── fetcher_test.go  # Tests
```

## Acceptance Criteria

1. All tests pass
2. `go build ./...` succeeds
3. `go vet ./...` reports no issues
4. Concurrent fetches don't corrupt cache
5. Errors include context (URL, status code)

## Implementation Notes

For the local LLM:

1. Start with `NewFetcher` and `CachePath`
2. Then implement `Fetch` without options
3. Add `WithForce` and `WithTimeout` options
4. Finally add `Clear` and `ClearAll`
5. Write tests alongside each function

Reference `llm-context/PATTERNS.md` for code style.
