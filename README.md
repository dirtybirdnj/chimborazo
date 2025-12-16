# Chimborazo

> **Cartography at the speed of code**

A high-performance geospatial visualization toolkit written in Go. Port of the [Strata](https://github.com/dirtybirdnj/strata) Python project.

## Overview

Chimborazo transforms geospatial data into beautiful SVG maps through a declarative YAML recipe system. Named after the Ecuadorian volcano whose peak is the farthest point from Earth's center—because your maps deserve to reach new heights.

## Features

- **Recipe-driven**: Define maps in YAML, render with one command
- **Multi-source**: Census TIGER, Natural Earth, OpenStreetMap, custom GeoJSON
- **Operations**: Clip, subtract, merge, dissolve, simplify, buffer
- **Output**: SVG optimized for pen plotters and laser cutters
- **Fast**: Go performance for large datasets

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

## Quick Start

```bash
# Build a map from a recipe
chimborazo build examples/vermont.yaml

# Preview without rendering
chimborazo validate examples/vermont.yaml

# List cached data sources
chimborazo cache list
```

## Recipe Example

```yaml
name: Vermont State Map
output:
  path: output/vermont.svg
  width: 12
  height: 18
  units: inches

bounds:
  source: census://state/VT
  padding: 0.1

layers:
  - name: state_boundary
    source: census://state/VT
    style:
      stroke: black
      stroke_width: 2

  - name: counties
    source: census://county/VT
    style:
      stroke: gray
      stroke_width: 0.5

  - name: water
    source: census://water/VT
    operations:
      - clip: bounds
    style:
      fill: none
      stroke: blue
```

## Architecture

```
chimborazo/
├── cmd/chimborazo/     # CLI entrypoint
├── internal/
│   ├── config/         # Recipe parsing
│   ├── sources/        # Data acquisition (thoreau equivalent)
│   ├── geometry/       # Geo operations (humboldt equivalent)
│   └── output/         # SVG generation (kelley equivalent)
├── pkg/pipeline/       # Build orchestration
└── testdata/           # Sample recipes
```

## Development

This project uses a hybrid LLM workflow:
- **Claude**: Architecture, specs, code review
- **Local LLMs**: Implementation via [clood](https://github.com/dirtybirdnj/clood)

See `llm-context/` for documentation optimized for local LLM consumption.

## Status

**Phase: Proof of Concept**

| Feature | Status |
|---------|--------|
| Recipe parsing | 🔲 |
| HTTP fetcher | 🔲 |
| File cache | 🔲 |
| Clip operation | 🔲 |
| SVG output | 🔲 |

## Project Lineage

This is the third iteration of the project:
1. **vt-geodata** - Original Vermont-focused experiments
2. **Strata** - Python implementation with full feature set
3. **Chimborazo** - Go port for performance

## Related Projects

- [Strata](https://github.com/dirtybirdnj/strata) - Python implementation
- [clood](https://github.com/dirtybirdnj/clood) - Local LLM infrastructure
- [vt-geodata](https://github.com/dirtybirdnj/vt-geodata) - Original experiments

## License

MIT
