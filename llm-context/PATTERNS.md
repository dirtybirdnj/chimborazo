# Chimborazo Code Patterns

## Error Handling

Always wrap errors with context:

```go
// Good
if err != nil {
    return fmt.Errorf("fetch %s: %w", url, err)
}

// Bad
if err != nil {
    return err
}
```

## Options Pattern

Use functional options for configurable constructors:

```go
type Option func(*Fetcher)

func WithTimeout(d time.Duration) Option {
    return func(f *Fetcher) {
        f.Timeout = d
    }
}

func WithCacheDir(dir string) Option {
    return func(f *Fetcher) {
        f.CacheDir = dir
    }
}

func NewFetcher(opts ...Option) *Fetcher {
    f := &Fetcher{
        CacheDir: defaultCacheDir(),
        Timeout:  30 * time.Second,
    }
    for _, opt := range opts {
        opt(f)
    }
    return f
}
```

## Interface Design

Keep interfaces small and focused:

```go
// Good - single responsibility
type Fetcher interface {
    Fetch(url string) (string, error)
}

// Bad - too many responsibilities
type DataManager interface {
    Fetch(url string) (string, error)
    Cache(path string) error
    Validate(data []byte) error
    Transform(data []byte) ([]byte, error)
}
```

## Geometry Operations

Always handle empty/nil geometries:

```go
func Clip(geom orb.Geometry, bounds orb.Bound) orb.Geometry {
    if geom == nil {
        return nil
    }
    // ... operation
}
```

## File Paths

Use filepath for cross-platform compatibility:

```go
import "path/filepath"

// Good
path := filepath.Join(cacheDir, "data", filename)

// Bad
path := cacheDir + "/data/" + filename
```

## HTTP Requests

Always set timeouts and close bodies:

```go
client := &http.Client{
    Timeout: 30 * time.Second,
}

resp, err := client.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
}
```

## Testing

Test files go alongside source files:

```
internal/sources/
├── fetcher.go
├── fetcher_test.go
├── cache.go
└── cache_test.go
```

Use table-driven tests:

```go
func TestFetcher_Fetch(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr bool
    }{
        {"valid URL", "https://example.com/data.json", false},
        {"invalid URL", "not-a-url", true},
        {"404", "https://example.com/missing", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Logging

Use structured logging (when we add it):

```go
// Future pattern
log.Info("fetching data",
    "url", url,
    "cache_dir", f.CacheDir,
)
```

For now, prefer returning errors over logging.

## Package Comments

Each package should have a doc.go:

```go
// Package sources provides data acquisition for Chimborazo.
//
// It handles fetching data from various sources (Census, Natural Earth,
// local files) and caching them locally for reuse.
package sources
```
