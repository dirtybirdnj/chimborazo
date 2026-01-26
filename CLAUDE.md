# Claude Agent Guidelines for Chimborazo

## ⚠️ NAMING CONVENTION - READ FIRST

| What | Name | Usage |
|------|------|-------|
| **Project** | Chimborazo | Repo, imports, documentation |
| **CLI Command** | `chimbo` | What users type |

```bash
# Correct:
chimbo build recipe.yaml
go install github.com/dirtybirdnj/chimborazo/cmd/chimbo@latest

# Incorrect:
chimborazo build recipe.yaml  # ❌ Use chimbo
```

---

## Project Context

Chimborazo is a Go port of the Strata Python geospatial toolkit. It transforms YAML recipes into SVG maps optimized for pen plotters and laser cutters.

**Key constraint**: This project uses a hybrid workflow where local LLMs (via clood) do implementation and Claude provides architecture, specs, and review.

## Narrative Branding

This project uses a historical narrative voice. Documentation is written from the perspective of **Alexander von Humboldt** in Paris (1805-1834), reflecting on his expedition to the Americas and drawing connections to cartographic principles.

### The Expedition Metaphor

| Stage | Package | Expedition Phase | Historical Reference |
|-------|---------|------------------|---------------------|
| Preparation | `internal/config` | Madrid: instruments & permissions | Maury's coordination |
| Acquisition | `internal/sources` | Collecting specimens | Thoreau's "go to the source" |
| Transformation | `internal/geometry` | Comparing & combining | Humboldt's unity in diversity |
| Revelation | `internal/output` | The Naturgemälde diagram | Kelley's maps-as-evidence |
| Return | `pkg/pipeline` | 30 volumes in Paris | Full orchestration |

### Voice Guidelines

When writing documentation:
- Use retrospective, reflective tone (Humboldt looking back on the journey)
- Reference historical figures as intellectual companions, not package names
- Connect technical concepts to expedition metaphors
- See `BRANDING.md` for full voice guidelines and example prose

### Key Documents

- `README.md` - Narrative introduction and quick start
- `BRANDING.md` - Voice guidelines and historical ensemble
- `EXPEDITION.md` - Full narrative journey through architecture
- Package `doc.go` files - Narrative + technical documentation

## Architecture Overview

```
Recipe (YAML) → Parser → Pipeline → Sources → Geometry Ops → SVG Output
```

| Stage | Go Package | Purpose |
|-------|------------|---------|
| Preparation | `internal/config` | Recipe parsing & validation |
| Acquisition | `internal/sources` | Data fetching & caching |
| Transformation | `internal/geometry` | Geometric operations |
| Revelation | `internal/output` | SVG generation |
| Return | `pkg/pipeline` | Build orchestration |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/paulmach/orb` | Geometry primitives |
| `github.com/paulmach/orb/geojson` | GeoJSON parsing |
| `github.com/paulmach/orb/clip` | Clipping operations |
| `github.com/paulmach/orb/simplify` | Douglas-Peucker simplification |
| `github.com/ajstarks/svgo` | SVG generation |
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | YAML parsing |

## Code Patterns

### Error Handling
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Geometry Types
```go
// Use orb types directly, wrap only when adding behavior
type Polygon = orb.Polygon
type MultiPolygon = orb.MultiPolygon
```

### Options Pattern
```go
type FetcherOption func(*Fetcher)

func WithTimeout(d time.Duration) FetcherOption {
    return func(f *Fetcher) { f.timeout = d }
}
```

## For Claude Agents

### When Writing Documentation

1. Open with a Humboldt-voice paragraph connecting to the expedition
2. Transition to technical documentation
3. Reference historical figures as wisdom sources, not namesakes
4. Use the stage names (Preparation, Acquisition, etc.) not module names

Example doc.go structure:
```go
/*
Package sources provides data acquisition from authoritative sources.

# The Acquisition Stage

I learned in Cumaná what the surveyor Thoreau knew: one must go to the source.
[narrative paragraph]

# Supported URI Schemes

[technical documentation]
*/
package sources
```

### When Reviewing Local LLM Code

1. Check error handling is complete
2. Verify orb types are used correctly
3. Ensure operations match Strata Python behavior
4. Look for edge cases (empty geometries, nil pointers)
5. Confirm tests cover happy path and errors

### When Writing Specs for Local LLMs

Include in every spec:
1. Exact function signatures
2. All relevant type definitions
3. Example inputs and outputs
4. Edge cases to handle
5. What NOT to do (anti-patterns)

### When Asked to Implement

Prefer to write a detailed spec and let the local LLM implement. Only implement directly if:
- The task is very small (< 20 lines)
- The local LLM has failed multiple times
- Time is critical

## The Cartographic Ensemble

These historical figures inform the project's principles. Reference them in documentation as intellectual companions:

| Figure | Principle | Application |
|--------|-----------|-------------|
| **William Smith** (1815) | Layer order matters | SVG z-index, rendering order |
| **Henry David Thoreau** (1854) | Go to the source | Authoritative data, no speculation |
| **Alexander von Humboldt** (1807) | Unity in diversity | Transformation reveals connections |
| **Matthew Fontaine Maury** (1855) | Coordinate many sources | Pipeline aggregation |
| **Florence Kelley** (1895) | Maps are evidence | Output serves comprehension |

## Local LLM Integration

This repo works with [clood](https://github.com/dirtybirdnj/clood) for local LLM orchestration.

### Context Files for Local LLMs

The `llm-context/` directory contains optimized documentation:
- `TYPES.md` - All type definitions
- `PATTERNS.md` - Code patterns to follow

When creating prompts for local LLMs, reference these files.

## Git Operations

- Use conventional commits: `feat(sources): add HTTP fetcher`
- Create feature branches: `feature/http-fetcher`
- PRs should reference the spec/issue they implement
- Include test files in the same commit as implementation

## Do NOT

- Install dependencies without asking
- Force push to main
- Implement large features without a spec
- Skip error handling
- Use `panic` except for truly unrecoverable errors
- Add dependencies not listed above without discussion
- Name packages after historical figures (use functional names)
- Write documentation without the narrative voice
