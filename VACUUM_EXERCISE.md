# Vacuum Exercise: Implement Translate Operation

This is a test exercise for the Claude Vacuum workflow. Complete this task using only clood tools and local LLMs.

## The Goal

Add a `translate` geometry operation that shifts all coordinates by X/Y offset.

## Prerequisites

```bash
# Verify you have what you need
clood preflight
clood hosts
```

## Step 1: Understand the Codebase

```bash
# Find where operations are handled
clood grep "applyOperation" ~/Code/chimborazo

# See existing operations
clood grep "case \"" ~/Code/chimborazo/pkg/pipeline/builder.go

# Look at geometry package
clood tree ~/Code/chimborazo/internal/geometry
clood symbols ~/Code/chimborazo/internal/geometry/
```

## Step 2: Study the Pattern

Look at how existing operations work:

```bash
# Read the builder
cat ~/Code/chimborazo/pkg/pipeline/builder.go

# Read existing geometry operations
cat ~/Code/chimborazo/internal/geometry/operations.go
```

Existing operations follow this pattern:
1. Function in `internal/geometry/` that processes `FeatureCollection`
2. Case in `applyOperation` that calls the function with params

## Step 3: Ask for Help (Local LLM)

```bash
# Build context
clood context ~/Code/chimborazo/internal/geometry > /tmp/geom_ctx.txt
clood context ~/Code/chimborazo/pkg/pipeline >> /tmp/geom_ctx.txt

# Ask for implementation guidance
clood ask "Based on this code, write a TranslateCollection function that shifts all points by X,Y offset. Follow the existing patterns." --context /tmp/geom_ctx.txt
```

## Step 4: Implement

### In `internal/geometry/operations.go`, add:

```go
// TranslateCollection shifts all geometry by the given X/Y offset.
func TranslateCollection(fc FeatureCollection, offsetX, offsetY float64) FeatureCollection {
    result := make(FeatureCollection, len(fc))
    for i, f := range fc {
        result[i] = &Feature{
            Geometry:   translateGeometry(f.Geometry, offsetX, offsetY),
            Properties: f.Properties,
        }
    }
    return result
}

func translateGeometry(g orb.Geometry, dx, dy float64) orb.Geometry {
    // Implement for each geometry type...
    // See SimplifyGeometry for the pattern
}
```

### In `pkg/pipeline/builder.go`, add case:

```go
case "translate":
    offsetX := 0.0
    offsetY := 0.0
    if x, ok := op.Params["x"].(float64); ok {
        offsetX = x
    }
    if y, ok := op.Params["y"].(float64); ok {
        offsetY = y
    }
    return geometry.TranslateCollection(features, offsetX, offsetY), nil
```

## Step 5: Build & Test

```bash
cd ~/Code/chimborazo

# Build
go build ./...

# Run tests
go test ./...

# Validate the test recipe
./chimborazo validate recipes/vacuum_test.yaml

# Build it (after uncommenting the translate operation)
./chimborazo build recipes/vacuum_test.yaml

# Check output
open output/vacuum_test.svg
```

## Success Criteria

- [ ] `TranslateCollection` function exists in geometry package
- [ ] Function handles Point, LineString, Polygon, MultiPolygon
- [ ] `translate` case added to `applyOperation`
- [ ] `recipes/vacuum_test.yaml` builds successfully
- [ ] Output SVG shows translated geometry (blue) offset from original (gray)

## If You Get Stuck

1. Read the existing `SimplifyCollection` implementation
2. Use catfight pattern: ask both qwen2.5-coder and deepseek-r1
3. Check the orb library docs: `go doc github.com/paulmach/orb`
4. Look at strata for reference: `clood grep "translate" ~/Code/strata`

## When Complete

Create an issue closure comment:

```bash
gh issue close 12 --repo dirtybirdnj/chimborazo \
  --comment "Implemented translate operation. Vacuum exercise completed successfully."
```

---

*This exercise validates that the vacuum workflow is functional. The iteration is the victory.*
