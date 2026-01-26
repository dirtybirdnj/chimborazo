# Chimborazo

> *"If instead of geographic maps, we only possessed tables covering latitude, longitude, and altitude, a great number of curious connections that continents manifest in their forms would have stayed forever lost."*
>
> — Alexander von Humboldt, 1807

---

## The Expedition

In June of 1802, I stood at 19,286 feet on the slopes of Chimborazo—higher than any European had ever climbed. The summit remained hidden in cloud, an impassable crevasse blocking the final thousand feet. Yet what I beheld from that height changed how humanity understands the world: not isolated facts, but *connections*. The way vegetation zones rise with altitude. The way temperature, pressure, and life itself form patterns that no table of numbers could reveal.

I returned to Paris with sixty thousand specimens and ten thousand measurements. But specimens are not understanding. Measurements are not maps. The work of *transformation*—of rendering raw observation into forms that reveal hidden truths—that work consumed the next three decades of my life.

**Chimborazo** is a toolkit for that same transformation: from raw geospatial data to maps that reveal what tables cannot show.

---

## What Chimborazo Does

```
Recipe (YAML) → Fetch Data → Transform Geometry → Render SVG → The Map
```

You declare what you want in a recipe. Chimborazo acquires the data, transforms it through geometric operations, and renders SVG maps suitable for screen, print, or pen plotter.

```yaml
name: Vermont Waters
bounds: [-73.5, 42.7, -71.5, 45.1]

sources:
  state: { uri: "census://state/VT" }
  water: { uri: "census://water/VT" }

layers:
  - name: boundary
    source: state
    style: { stroke: "#333", stroke_width: 2 }

  - name: lakes
    source: water
    operations:
      - type: clip
      - type: simplify
        tolerance: 0.0001
    style: { stroke: "#2196F3", fill: "none" }

output:
  formats:
    - type: svg
      page_size: "12x18"
```

---

## Installation

```bash
go install github.com/dirtybirdnj/chimborazo/cmd/chimborazo@latest
```

Or build from source:

```bash
git clone https://github.com/dirtybirdnj/chimborazo.git
cd chimborazo
go build -o chimborazo ./cmd/chimborazo
```

---

## The Stages of the Expedition

Every map follows the path I walked from Madrid to Chimborazo and back to Paris:

### 1. Preparation (`internal/config/`)

Before departure, one must study the route, gather instruments, secure permissions. The recipe is your expedition plan—validated before a single HTTP request is made.

### 2. Acquisition (`internal/sources/`)

I learned in Cumaná what the surveyor Thoreau knew: *"It is remarkable how long men will believe in the bottomlessness of a pond without taking the trouble to sound it."* One must fetch data from authoritative sources—Census TIGER files, Natural Earth, OpenStreetMap—not speculation or hearsay.

### 3. Transformation (`internal/geometry/`)

Raw coordinates are not understanding. The unity emerges only through comparison and combination: clipping to bounds, subtracting water from land, merging fragmented features, simplifying for output. This is where hidden connections reveal themselves.

### 4. Revelation (`internal/output/`)

Florence Kelley understood that *"a map is not decoration—a map is evidence."* The SVG output must serve comprehension. Every stroke width, every layer order, every color choice serves the goal of making patterns visible that would otherwise stay forever lost.

### 5. Return (`pkg/pipeline/`)

Matthew Fontaine Maury collected ship logs from captains worldwide, aggregating thousands of observations into charts that no single voyage could produce. The pipeline coordinates all stages, all sources, producing unified output from disparate inputs.

---

## Usage

```bash
# Validate a recipe without building
chimborazo validate examples/vermont.yaml

# Build a map
chimborazo build examples/vermont.yaml

# Build with verbose output
chimborazo build -v examples/vermont.yaml

# List cached data sources
chimborazo cache list

# Clear the cache
chimborazo cache clear
```

---

## Architecture

```
chimborazo/
├── cmd/chimborazo/      # CLI entrypoint
├── internal/
│   ├── config/          # Preparation: recipe parsing & validation
│   ├── sources/         # Acquisition: data fetching & caching
│   ├── geometry/        # Transformation: geometric operations
│   └── output/          # Revelation: SVG rendering
├── pkg/pipeline/        # Return: build orchestration
├── examples/            # Sample recipes
└── testdata/            # Test fixtures
```

---

## The Cartographic Ensemble

This project stands on the shoulders of mapmakers who came before. Their methods inform every stage:

| Figure | Era | Contribution | Principle |
|--------|-----|--------------|-----------|
| **William Smith** | 1815 | First geological map of a nation | Layer order matters |
| **Henry David Thoreau** | 1854 | Surveyor of Walden Pond | Go to the source |
| **Alexander von Humboldt** | 1807 | Invented isotherms, biogeography | Unity in diversity |
| **Matthew Fontaine Maury** | 1855 | Aggregated global ocean data | Coordinate many sources |
| **Florence Kelley** | 1895 | Hull House social maps | Maps are evidence |

Their voices echo through the documentation. Their principles guide the code.

---

## Project Lineage

This is the third summit attempt:

1. **vt-geodata** — Initial experiments, Vermont-focused
2. **Strata** — Python implementation with full feature set
3. **Chimborazo** — Go port for performance and portability

The name honors Humboldt's 1802 ascent—the highest point reached by any European until the Himalayan surveys thirty years later. We climb toward the same summit: maps that reveal what raw data cannot show.

---

## Development

This project uses a hybrid workflow:
- **Claude**: Architecture, specifications, code review
- **Local LLMs**: Implementation via [clood](https://github.com/dirtybirdnj/clood)

See `llm-context/` for documentation optimized for local LLM consumption.

---

## Status

**Phase: Proof of Concept**

| Stage | Package | Status |
|-------|---------|--------|
| Preparation | `config/` | ✅ Recipe parsing |
| Acquisition | `sources/` | ✅ HTTP fetcher with cache |
| Transformation | `geometry/` | 🔲 Operations pipeline |
| Revelation | `output/` | ✅ Basic SVG |
| Return | `pipeline/` | ✅ Build orchestration |

---

## License

MIT

---

> *"The most important aim of all physical science is this: to recognize unity in diversity."*
>
> — Humboldt
