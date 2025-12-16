# Chimborazo Types

## Core Types

### Recipe (internal/config)

```go
// Recipe defines a complete map build
type Recipe struct {
    Name   string       `yaml:"name"`
    Output OutputConfig `yaml:"output"`
    Bounds BoundsConfig `yaml:"bounds"`
    Layers []Layer      `yaml:"layers"`
}

// OutputConfig defines output file settings
type OutputConfig struct {
    Path   string  `yaml:"path"`
    Width  float64 `yaml:"width"`
    Height float64 `yaml:"height"`
    Units  string  `yaml:"units"` // "inches", "mm", "px"
}

// BoundsConfig defines the map extent
type BoundsConfig struct {
    Source  string  `yaml:"source"`  // URI like "census://state/VT"
    Padding float64 `yaml:"padding"` // Fraction to expand bounds
}

// Layer defines a single map layer
type Layer struct {
    Name       string            `yaml:"name"`
    Source     string            `yaml:"source"`
    Operations []Operation       `yaml:"operations,omitempty"`
    Style      map[string]string `yaml:"style,omitempty"`
    Filter     string            `yaml:"filter,omitempty"`
}

// Operation defines a geometry operation
type Operation struct {
    Type   string                 `yaml:"type"` // "clip", "subtract", "merge", etc.
    Params map[string]interface{} `yaml:"params,omitempty"`
}
```

### Sources (internal/sources)

```go
// Fetcher downloads and caches remote data
type Fetcher struct {
    CacheDir string
    Client   *http.Client
    Timeout  time.Duration
}

// FetchResult contains the result of a fetch operation
type FetchResult struct {
    Path      string    // Local file path
    FromCache bool      // True if served from cache
    Size      int64     // File size in bytes
    FetchedAt time.Time // When downloaded (or cache time)
}

// Source represents a data source
type Source interface {
    // Fetch retrieves data and returns local path
    Fetch(uri string) (*FetchResult, error)

    // Supports returns true if this source handles the URI scheme
    Supports(uri string) bool
}
```

### Geometry (internal/geometry)

```go
// Using paulmach/orb types directly
import "github.com/paulmach/orb"

type Point = orb.Point
type Ring = orb.Ring
type Polygon = orb.Polygon
type MultiPolygon = orb.MultiPolygon
type Bound = orb.Bound
type Geometry = orb.Geometry

// Feature wraps a geometry with properties
type Feature struct {
    Geometry   orb.Geometry
    Properties map[string]interface{}
}

// FeatureCollection is a slice of features
type FeatureCollection []Feature
```

### Output (internal/output)

```go
// SVGWriter generates SVG output
type SVGWriter struct {
    Width  float64
    Height float64
    Units  string
    Bounds orb.Bound
}

// StyleSet defines visual styling
type StyleSet struct {
    Stroke      string
    StrokeWidth float64
    Fill        string
    Opacity     float64
}
```

### Pipeline (pkg/pipeline)

```go
// Builder orchestrates the build process
type Builder struct {
    Recipe  *config.Recipe
    Fetcher *sources.Fetcher
    Output  *output.SVGWriter
}

// BuildResult contains the result of a build
type BuildResult struct {
    OutputPath string
    Duration   time.Duration
    LayerCount int
    Errors     []error
}
```
