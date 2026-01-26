# Chimborazo Development Roadmap

> A guide for future agents working on this codebase

## Current Status: Phase 1 Complete

**Chimborazo can now build real maps from Census data.**

```bash
go run ./cmd/chimborazo build -v examples/vermont.yaml
# Outputs: output/vermont.svg (261KB, 15 features)
```

---

## Phase 1: Make It Work (COMPLETE)

The foundation is in place:

| Component | Status | Files |
|-----------|--------|-------|
| Census URI resolver | Done | `internal/sources/census.go` |
| ZIP extraction | Done | `internal/sources/zip.go` |
| Shapefile reading | Done | `internal/sources/shapefile.go` |
| Source resolver | Done | `internal/sources/resolver.go` |
| Bounds from geometry | Done | `pkg/pipeline/builder.go` |
| SVG output | Done | `internal/output/svg.go` |
| CLI with verbose flag | Done | `cmd/chimborazo/main.go` |

---

## Phase 2: Core Operations (IN PROGRESS)

Add the remaining geometry operations needed for real cartographic work.

### 2.1 Subtract Operation (DONE - with caveats)

The "cut water from land" operation using `github.com/engelsjk/polygol`.

**Status:** Working for moderate-sized datasets. Large datasets (14,000+ features)
may hit polygol's queue size limit.

**Usage:**
```yaml
operations:
  - type: subtract
    params:
      source: census://areawater/VT  # or file:///path/to/water.shp
```

**Known limitations:**
- Very large feature counts may trigger "queue size too big" error
- Workaround: Use per-county water files instead of merged state file
- Future: Implement batched processing for large datasets

**Test recipes:**
- `examples/vermont-water-simple.yaml` - single county, works
- `examples/vermont-water.yaml` - full state, may fail on complex data

### 2.2 Filter Operation (DONE)

Filter features by property values.

**Status:** Complete. Supports =, !=, >, <, >=, <=, contains, starts_with, ends_with.

**Usage:**
```yaml
operations:
  - type: filter
    params:
      field: NAME
      operator: "="
      value: Chittenden
```

**Test recipe:** `examples/vermont-filter.yaml`

### 2.3 Dissolve Operation (DONE)

Merge features that share an attribute value using polygon union.

**Status:** Complete. Uses polygol union to combine geometries sharing the same
property value.

**Usage:**
```yaml
operations:
  - type: dissolve
    params:
      field: COUNTYFP  # Group by this field
```

**Test recipe:** `examples/vermont-dissolve.yaml` - dissolves 256 towns into 14 counties.

### 2.3 Buffer Operation

Expand or contract geometry boundaries.

**Complexity:** High - requires arc approximation for corners.

**Options:**
1. `github.com/ctessum/geom` with CGO/GEOS
2. Pure Go approximation (good enough for cartographic display)

---

## Phase 3: Output Polish (IN PROGRESS)

### 3.1 Per-Layer SVG Files (DONE)

Output each layer as a separate SVG file for multi-pen plotting.

**Status:** Complete. Each layer written as `{basename}_{layername}.svg`.

**Usage:**
```yaml
output:
  path: output/vermont.svg
  per_layer: true   # Creates vermont_state.svg, vermont_counties.svg, etc.
  margin: 0.5       # Optional margin in inches
```

**Test recipe:** `examples/vermont-plotter.yaml`

### 3.2 Quality Levels (DONE)

Simplification presets for different output uses.

**Status:** Complete. Auto-applies simplification based on quality setting.

**Usage:**
```yaml
output:
  quality: high  # low, medium, high, plotter
```

| Level | Tolerance | File Size | Use Case |
|-------|-----------|-----------|----------|
| low | 0.001 | ~33KB | Web preview |
| medium | 0.0003 | ~65KB | Screen display |
| high | 0.0001 | ~127KB | Print |
| plotter | 0.00005 | ~200KB | Pen plotter |

**Test recipe:** `examples/vermont-quality-test.yaml`

### 3.3 Plotter Fill

Convert fill colors to hatch patterns for pen plotters.

```yaml
style:
  fill: "#336699"
  plotter_fill: hatch  # none, hatch, crosshatch, dots
```

---

## Phase 4: Recipe Schema Alignment (COMPLETE)

Align recipe format with Strata Python for compatibility.

### 4.1 Sources Block (DONE)

Named sources for reuse across layers and bounds.

**Status:** Complete. Sources defined at top level, referenced by name.

**Usage:**
```yaml
sources:
  state:
    uri: census://state/VT
  water:
    uri: census://areawater/VT

bounds:
  source: state  # Reference by name

layers:
  - name: boundary
    source: state  # Reference by name
```

**Test recipe:** `examples/vermont-sources.yaml`

### 4.2 Explicit Bounds (DONE)

Support explicit bounds as alternative to source-derived.

**Status:** Complete. Bounds can be specified as [West, South, East, North].

**Usage:**
```yaml
bounds:
  explicit: [-73.5, 44.0, -73.0, 45.0]  # [W, S, E, N]
  padding: 0.02  # Optional padding
```

**Test recipe:** `examples/lake-champlain.yaml`

### 4.3 Style Inheritance (DONE)

Layer styles inherit from defaults.

**Status:** Complete. Layers inherit from `defaults.style` and can override
individual properties.

**Usage:**
```yaml
defaults:
  style:
    stroke: "#333333"
    stroke_width: "1.0"
    fill: "none"

layers:
  - name: water
    style:
      stroke: "#2196F3"  # Override just stroke color
```

**Test recipe:** `examples/vermont-defaults.yaml`

---

## Phase 5: Additional Sources

### 5.1 Natural Earth

High-quality global data for context maps.

```
naturalearth://cultural/admin_0_countries
naturalearth://physical/lakes
```

### 5.2 OpenStreetMap

Via Overpass API or pre-extracted files.

```
osm://way[highway=primary]
osm://relation[boundary=administrative]
```

### 5.3 Canadian NHN

National Hydro Network for cross-border maps.

```
nhn://02OJ  # Watershed ID
```

---

## Phase 6: Performance & Polish (IN PROGRESS)

### 6.1 Concurrent Downloads (DONE)

Fetch multiple county files in parallel using a worker pool.

**Status:** Complete. Uses 4 concurrent workers for county downloads.

**Implementation:** `internal/sources/resolver.go` - `downloadAndMergeCounties()`

### 6.2 Progress Reporting

Real-time progress for long builds.

### 6.3 Cache Management

Expiration, size limits, integrity checking.

### 6.4 Error Recovery

Resume interrupted downloads, handle partial failures gracefully.

---

## Architecture Notes for Agents

### Package Structure

```
internal/
├── config/       # Recipe parsing (Preparation stage)
├── sources/      # Data fetching (Acquisition stage)
│   ├── census.go    # Census URI resolver
│   ├── fetcher.go   # HTTP download + cache
│   ├── resolver.go  # Unified source resolution
│   ├── shapefile.go # Shapefile → FeatureCollection
│   └── zip.go       # ZIP extraction
├── geometry/     # Operations (Transformation stage)
└── output/       # SVG generation (Revelation stage)

pkg/
└── pipeline/     # Build orchestration (Return stage)
```

### Key Types

```go
// Feature with geometry and properties
type Feature struct {
    Geometry   orb.Geometry
    Properties map[string]interface{}
}

// Collection of features
type FeatureCollection []*Feature

// Census URI parser
type CensusURI struct {
    Year, State, Type string
    URLs              []string  // May be multiple for per-county data
}

// Build orchestrator
type Builder struct {
    Recipe   *config.Recipe
    Resolver *sources.Resolver
}
```

### Testing a Build

```bash
# Build with verbose output
go run ./cmd/chimborazo build -v examples/vermont.yaml

# Check the output
ls -la output/vermont.svg

# Clear cache to force re-download
go run ./cmd/chimborazo cache clear
```

### Adding a New Operation

1. Add case to `Builder.applyOperation()` in `pkg/pipeline/builder.go`
2. Implement the operation in `internal/geometry/operations.go`
3. Add tests in `internal/geometry/operations_test.go`
4. Document in `internal/geometry/doc.go`

### Adding a New Source Type

1. Create resolver in `internal/sources/` (e.g., `naturalearth.go`)
2. Add URI prefix handling to `Resolver.Resolve()` in `resolver.go`
3. Update URI scheme docs in `internal/sources/doc.go`

---

## Dependencies

| Package | Purpose | Pure Go? |
|---------|---------|----------|
| `github.com/paulmach/orb` | Geometry primitives | Yes |
| `github.com/paulmach/orb/clip` | Clipping | Yes |
| `github.com/paulmach/orb/simplify` | Simplification | Yes |
| `github.com/jonas-p/go-shp` | Shapefile reading | Yes |
| `github.com/spf13/cobra` | CLI framework | Yes |
| `gopkg.in/yaml.v3` | YAML parsing | Yes |

**For Phase 2 operations, consider:**
- `github.com/engelsjk/polygol` - polygon boolean ops (pure Go)
- `github.com/ctessum/geom` - geometry operations (may need CGO)

---

## Reference Materials

- **Strata Python:** `~/Code/strata/` - reference implementation
- **LLM Context:** `~/Code/strata/llm-context/` - operation specs
- **Historical Narrative:** See `BRANDING.md` and `EXPEDITION.md`

---

> *"The view from 19,286 feet revealed what no table of numbers could convey."*
