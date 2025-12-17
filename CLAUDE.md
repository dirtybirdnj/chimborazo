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

Chimborazo is a Go port of the Strata Python geospatial toolkit. It transforms YAML recipes into SVG maps.

**Key constraint**: This project uses a hybrid workflow where local LLMs (via clood) do implementation and Claude provides architecture, specs, and review.

## Architecture Overview

```
Recipe (YAML) → Parser → Pipeline → Sources → Geometry Ops → SVG Output
```

| Module | Go Package | Python Equivalent | Purpose |
|--------|------------|-------------------|---------|
| Config | `internal/config` | `maury` | Recipe parsing |
| Sources | `internal/sources` | `thoreau` | Data fetching/caching |
| Geometry | `internal/geometry` | `humboldt` | Geo operations |
| Output | `internal/output` | `kelley` | SVG generation |
| Pipeline | `pkg/pipeline` | `maury.pipeline` | Orchestration |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/paulmach/orb` | Geometry primitives |
| `github.com/paulmach/orb/geojson` | GeoJSON parsing |
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

## Local LLM Integration

This repo works with [clood](https://github.com/dirtybirdnj/clood) for local LLM orchestration.

### Context Files for Local LLMs

The `llm-context/` directory contains optimized documentation:
- `TYPES.md` - All type definitions
- `INTERFACES.md` - Interface contracts
- `PATTERNS.md` - Code patterns to follow
- `OPERATIONS.md` - Geometry operation specs

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
