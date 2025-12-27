# Chimborazo Session - Dec 27, 2024

## Current State
Working on Vermont towns + hydro map with water subtraction.

## The Problem We're Stuck On
Small lakes are not appearing on the map despite adjusting filter thresholds from 3mm to 7mm. We've swung between two extremes:
1. **Too many features**: Thousands of tiny black specks (holes in towns from subtracted water)
2. **Too few features**: No small lakes visible at all

### Root Cause (suspected)
There are TWO separate filters that need to work together:
1. `min_feature_size` in output config - filters the **water display layer** (blue water shown)
2. `min_size` in subtract operation params - filters what gets **subtracted from towns**

The water layer at 3mm shows only ~7 features (big lakes like Champlain, Memphremagog).
The subtract at 3mm processes ~300+ water features per county.

**Mismatch**: Many lakes get cut OUT of towns but aren't DISPLAYED as blue because the water layer filter is too aggressive.

### What was tried
- min_feature_size: 4mm, 5mm, 6mm, 7mm
- min_size on subtract: 3mm, 5mm, 6mm, 7mm
- Both set to same value (3mm currently)

## Key Files

### Recipe: `examples/vermont-towns-hydro.yaml`
```yaml
output:
  min_feature_size: 3  # Filter water features smaller than 3mm

layers:
  - name: water
    source: census://areawater/VT
    # This layer gets filtered by min_feature_size

  - name: towns
    operations:
      - type: subtract
        params:
          source: census://areawater/VT
          by_county: true
          min_size: 3  # Separate filter for subtraction
```

### Code changes made this session

1. **`pkg/pipeline/builder.go`**:
   - Added `min_size` parameter to subtract operation
   - Modified `subtractByCounty()` to accept bounds and minSizeMM parameters
   - Filters water features BEFORE subtracting them from towns

2. **`internal/output/svg.go`**:
   - Changed stroke_width interpretation from pixels to mm
   - Added `mmToPixels()` helper function
   - Rulers now use 0.3mm line width

3. **`internal/config/recipe.go`**:
   - Already had min_feature_size and rulers options

## Recipe Settings (current)
- Output: 12x18 inches
- min_feature_size: 3mm
- subtract min_size: 3mm
- County stroke: 0.5mm
- Town stroke: 0.3mm
- State boundary: REMOVED (was too thick at 2.0mm)

## Things That Work
- County-by-county water subtraction (avoids polygon queue limits)
- Stroke width in mm (converts to pixels at render time)
- Rulers on left and top edges
- Color mapping by county FIPS code

## Next Steps to Try
1. **Debug the water layer**: Check why only ~7 features pass the 3mm filter when statewide there should be hundreds of lakes >= 3mm
2. **Check the merged water data**: The statewide water layer loads merged data - maybe the merge or cache is stale
3. **Compare per-county vs statewide counts**: The subtract sees ~300+ features per county at 3mm but the display layer only sees ~7 total?
4. **Try removing min_feature_size entirely** from output config - let ALL water display as blue, only filter the subtraction

## Pending Tasks
- Create full Lake Champlain map with NY
- Add river/linearwater layer

## Commands to rebuild
```bash
go build -o chimborazo ./cmd/chimborazo
./chimborazo build --verbose examples/vermont-towns-hydro.yaml
open output/vermont-towns-hydro.svg

# Generate PNG for analysis
qlmanage -t -s 2000 -o /tmp output/vermont-towns-hydro.svg
open /tmp/vermont-towns-hydro.svg.png
```
