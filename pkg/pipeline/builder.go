// Package pipeline orchestrates the map building process.
// It connects recipe parsing, data fetching, geometry operations, and output generation.
package pipeline

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/paulmach/orb"

	"github.com/dirtybirdnj/chimborazo/internal/config"
	"github.com/dirtybirdnj/chimborazo/internal/geometry"
	"github.com/dirtybirdnj/chimborazo/internal/output"
	"github.com/dirtybirdnj/chimborazo/internal/sources"
)

// Builder orchestrates the map building process.
type Builder struct {
	Recipe        *config.Recipe
	Resolver      *sources.Resolver
	Verbose       bool
	emittedLayers map[string]*emittedLayer // Layers created by operations (e.g., subtract with emit_as)
}

// emittedLayer holds features emitted by an operation for later rendering.
type emittedLayer struct {
	Features geometry.FeatureCollection
	Style    map[string]string
	Order    int
}

// BuildResult contains the outcome of a build.
type BuildResult struct {
	OutputPath   string
	Duration     time.Duration
	LayerCount   int
	FeatureCount int
	Bounds       orb.Bound
	Errors       []error
	Warnings     []string
}

// AnalysisResult contains findings from post-build analysis.
type AnalysisResult struct {
	Warnings      []string
	HolesWithoutFill []orb.Bound // Polygon holes with no corresponding water
	FillWithoutHole  []orb.Bound // Water features with no corresponding hole
}

// NewBuilder creates a pipeline builder for a recipe.
func NewBuilder(recipe *config.Recipe, cacheDir string) (*Builder, error) {
	resolver, err := sources.NewResolver(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("creating resolver: %w", err)
	}

	return &Builder{
		Recipe:        recipe,
		Resolver:      resolver,
		emittedLayers: make(map[string]*emittedLayer),
	}, nil
}

// Build executes the full pipeline: fetch → process → render.
func (b *Builder) Build() (*BuildResult, error) {
	start := time.Now()
	result := &BuildResult{}

	b.Resolver.Verbose = b.Verbose

	// Step 1: Validate recipe
	if err := config.ValidateRecipe(b.Recipe); err != nil {
		return nil, fmt.Errorf("invalid recipe: %w", err)
	}
	b.logf("Recipe validated: %s", b.Recipe.Name)

	// Step 2: Determine bounds
	bounds, err := b.resolveBounds()
	if err != nil {
		return nil, fmt.Errorf("resolving bounds: %w", err)
	}
	result.Bounds = bounds
	b.logf("Bounds: [%.4f, %.4f, %.4f, %.4f]",
		bounds.Min[0], bounds.Min[1], bounds.Max[0], bounds.Max[1])

	// Step 3: Process each layer
	var layers []output.Layer
	for _, layerDef := range b.Recipe.Layers {
		b.logf("Processing layer: %s", layerDef.Name)

		layer, err := b.processLayer(layerDef, bounds)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("layer %s: %w", layerDef.Name, err))
			b.logf("  Error: %v", err)
			continue
		}
		if layer != nil {
			layers = append(layers, *layer)
			result.FeatureCount += len(layer.Features)
			b.logf("  %d features", len(layer.Features))
		}
	}

	// Step 3.5: Add emitted layers (from subtract with emit_as, etc.)
	for name, emitted := range b.emittedLayers {
		b.logf("Adding emitted layer: %s (%d features)", name, len(emitted.Features))

		// Clip emitted features to bounds (they come from subtract which doesn't clip)
		features := geometry.ClipCollection(emitted.Features, bounds)
		if len(features) != len(emitted.Features) {
			b.logf("  Clipped to bounds: %d → %d features", len(emitted.Features), len(features))
		}

		// Filter by min_feature_size if set
		if b.Recipe.Output.MinFeatureSize > 0 {
			before := len(features)
			features = b.filterByOutputSize(features, bounds, b.Recipe.Output.MinFeatureSize)
			if len(features) != before {
				b.logf("  Filtered by size (>%.1fmm): %d → %d features", b.Recipe.Output.MinFeatureSize, before, len(features))
			}
		}

		style := b.parseStyle(emitted.Style)
		layers = append(layers, output.Layer{
			Name:     name,
			Features: features,
			Style:    style,
			Order:    emitted.Order,
		})
		result.FeatureCount += len(features)
	}

	// Sort layers by order (higher order = render first = bottom of stack)
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].Order > layers[j].Order
	})

	result.LayerCount = len(layers)

	// Step 4: Ensure output directory exists
	outPath := b.Recipe.Output.Path
	if outPath == "" {
		outPath = "output/map.svg"
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	// Step 5: Render output
	margin := b.Recipe.Output.Margin
	if margin == 0 {
		margin = 0.5 // default margin in inches
	}
	writer := output.NewSVGWriter(
		b.Recipe.Output.Width,
		b.Recipe.Output.Height,
		margin,
		bounds,
	)
	writer.ShowRulers = b.Recipe.Output.Rulers
	// For plotter quality, skip background rect (for rat-king post-processing)
	if b.Recipe.Output.Quality == "plotter" {
		writer.NoBackground = true
	}

	// Step 6: Write to file(s)
	if b.Recipe.Output.PerLayer {
		// Write each layer as a separate file
		outDir := filepath.Dir(outPath)
		baseName := filepath.Base(outPath)
		ext := filepath.Ext(baseName)
		nameWithoutExt := baseName[:len(baseName)-len(ext)]

		for _, layer := range layers {
			layerPath := filepath.Join(outDir, fmt.Sprintf("%s_%s%s", nameWithoutExt, layer.Name, ext))
			if err := writer.RenderLayerToFile(layer, layerPath); err != nil {
				return nil, fmt.Errorf("writing layer %s: %w", layer.Name, err)
			}
			b.logf("Wrote: %s", layerPath)
		}

		// Also write the combined file
		if err := writer.RenderToFile(layers, outPath); err != nil {
			return nil, fmt.Errorf("writing output: %w", err)
		}
		b.logf("Wrote: %s (combined)", outPath)
		result.OutputPath = outPath
	} else {
		if err := writer.RenderToFile(layers, outPath); err != nil {
			return nil, fmt.Errorf("writing output: %w", err)
		}
		result.OutputPath = outPath
		b.logf("Wrote: %s", outPath)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// BuildWithAnalysis builds the map and performs validation analysis.
func (b *Builder) BuildWithAnalysis() (*BuildResult, *AnalysisResult, error) {
	result, err := b.Build()
	if err != nil {
		return nil, nil, err
	}

	analysis := &AnalysisResult{}

	// Check for emit_as usage (ideal case - no mismatches possible)
	hasEmitAs := false
	for _, layer := range b.Recipe.Layers {
		for _, op := range layer.Operations {
			if op.Type == "subtract" {
				if _, ok := op.Params["emit_as"]; ok {
					hasEmitAs = true
				}
			}
		}
	}

	if hasEmitAs {
		analysis.Warnings = append(analysis.Warnings,
			"Using emit_as for water layers - alignment guaranteed")
	}

	// Check for separate water layers without emit_as (potential mismatch)
	waterLayers := 0
	subtractOps := 0
	for _, layer := range b.Recipe.Layers {
		if strings.Contains(strings.ToLower(layer.Name), "water") {
			waterLayers++
		}
		for _, op := range layer.Operations {
			if op.Type == "subtract" {
				if _, ok := op.Params["emit_as"]; !ok {
					subtractOps++
				}
			}
		}
	}

	if waterLayers > 0 && subtractOps > 0 && !hasEmitAs {
		analysis.Warnings = append(analysis.Warnings,
			fmt.Sprintf("Found %d separate water layers and %d subtract operations without emit_as - potential alignment issues", waterLayers, subtractOps))
	}

	// Check for min_feature_size mismatch
	outputMinSize := b.Recipe.Output.MinFeatureSize
	for _, layer := range b.Recipe.Layers {
		for _, op := range layer.Operations {
			if op.Type == "subtract" {
				if minSize, ok := op.Params["min_size"].(float64); ok {
					if minSize != outputMinSize && outputMinSize > 0 {
						analysis.Warnings = append(analysis.Warnings,
							fmt.Sprintf("Layer %s: subtract min_size (%.1f) differs from output min_feature_size (%.1f) - may cause display mismatches",
								layer.Name, minSize, outputMinSize))
					}
				}
			}
		}
	}

	return result, analysis, nil
}

// WriteDebugOverlay writes an SVG showing analysis findings.
func (b *Builder) WriteDebugOverlay(path string, analysis *AnalysisResult) error {
	// For now, just write warnings as text
	// TODO: Implement visual overlay with problem areas highlighted
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600">
<style>
  text { font-family: monospace; font-size: 14px; }
  .warning { fill: #ff6600; }
  .header { font-size: 18px; font-weight: bold; }
</style>
<text x="20" y="30" class="header">Analysis Warnings</text>
`)

	y := 60
	for _, w := range analysis.Warnings {
		fmt.Fprintf(f, `<text x="20" y="%d" class="warning">⚠ %s</text>`+"\n", y, w)
		y += 25
	}

	f.WriteString("</svg>")
	return nil
}

// resolveBounds determines the geographic extent for the map.
func (b *Builder) resolveBounds() (orb.Bound, error) {
	cfg := b.Recipe.Bounds

	// Check for explicit bounds first [West, South, East, North]
	if len(cfg.Explicit) == 4 {
		b.logf("Using explicit bounds: [%.4f, %.4f, %.4f, %.4f]",
			cfg.Explicit[0], cfg.Explicit[1], cfg.Explicit[2], cfg.Explicit[3])

		bound := orb.Bound{
			Min: orb.Point{cfg.Explicit[0], cfg.Explicit[1]}, // West, South
			Max: orb.Point{cfg.Explicit[2], cfg.Explicit[3]}, // East, North
		}

		// Apply padding
		if cfg.Padding > 0 {
			bound = padBounds(bound, cfg.Padding)
		}

		return bound, nil
	}

	// If a bounds source is specified, get bounds from that source
	if cfg.Source != "" {
		sourceURI := b.resolveSourceURI(cfg.Source)
		b.logf("Getting bounds from source: %s", sourceURI)

		bound, err := b.Resolver.GetBounds(sourceURI)
		if err != nil {
			return orb.Bound{}, fmt.Errorf("getting bounds from %s: %w", sourceURI, err)
		}

		// Apply padding
		if cfg.Padding > 0 {
			bound = padBounds(bound, cfg.Padding)
		}

		return bound, nil
	}

	// No bounds specified - this is an error
	return orb.Bound{}, fmt.Errorf("no bounds specified: use 'source' or 'explicit' in bounds config")
}

// padBounds expands a bound by a fractional padding.
// padding of 0.1 means 10% expansion on each side.
func padBounds(b orb.Bound, padding float64) orb.Bound {
	width := b.Max[0] - b.Min[0]
	height := b.Max[1] - b.Min[1]
	dx := width * padding
	dy := height * padding

	return orb.Bound{
		Min: orb.Point{b.Min[0] - dx, b.Min[1] - dy},
		Max: orb.Point{b.Max[0] + dx, b.Max[1] + dy},
	}
}

// resolveSourceURI converts a source reference to a URI.
// If the source looks like a named reference (no "://" or "/"), look it up in Sources.
func (b *Builder) resolveSourceURI(source string) string {
	// Check if it's a named source (no scheme or path separator)
	if !strings.Contains(source, "://") && !strings.HasPrefix(source, "/") && !strings.HasPrefix(source, "~") {
		if def, ok := b.Recipe.Sources[source]; ok {
			return def.URI
		}
	}
	return source
}

// processLayer fetches data and applies operations for a single layer.
func (b *Builder) processLayer(layerDef config.Layer, bounds orb.Bound) (*output.Layer, error) {
	// Step 1: Resolve source reference to URI
	sourceURI := b.resolveSourceURI(layerDef.Source)
	features, err := b.Resolver.Resolve(sourceURI)
	if err != nil {
		return nil, fmt.Errorf("resolving source: %w", err)
	}

	if len(features) == 0 {
		return nil, nil // Empty layer, skip
	}

	// Step 2: Apply operations
	for _, op := range layerDef.Operations {
		features, err = b.applyOperation(op, features, bounds)
		if err != nil {
			return nil, fmt.Errorf("operation %s: %w", op.Type, err)
		}
	}

	// Step 3: Apply quality-based simplification (if quality is set)
	if b.Recipe.Output.Quality != "" {
		preset := config.GetQualityPreset(b.Recipe.Output.Quality)
		features = geometry.SimplifyCollection(features, preset.Tolerance)
	}

	// Step 3.5: Filter out features smaller than min_feature_size (in mm)
	if b.Recipe.Output.MinFeatureSize > 0 {
		before := len(features)
		features = b.filterByOutputSize(features, bounds, b.Recipe.Output.MinFeatureSize)
		if len(features) != before {
			b.logf("  Filtered by size (>%.1fmm): %d → %d features", b.Recipe.Output.MinFeatureSize, before, len(features))
		}
	}

	// Step 4: Merge default styles with layer styles
	mergedStyleMap := b.mergeStyles(b.Recipe.Defaults.Style, layerDef.Style)
	style := b.parseStyle(mergedStyleMap)

	// Step 5: Parse order (default to 0)
	order := 0
	if orderStr, ok := mergedStyleMap["order"]; ok {
		fmt.Sscanf(orderStr, "%d", &order)
	}

	// Default label property to "NAME" if labels enabled but no property specified
	labelProp := layerDef.LabelProperty
	if layerDef.Labels && labelProp == "" {
		labelProp = "NAME"
	}

	// Convert pattern configs from config to output type
	var patterns []output.PatternConfig
	for _, p := range layerDef.Patterns {
		patterns = append(patterns, output.PatternConfig{
			Pattern: p.Pattern,
			Spacing: p.Spacing,
			Angle:   p.Angle,
		})
	}

	return &output.Layer{
		Name:          layerDef.Name,
		Features:      features,
		Style:         style,
		Order:         order,
		FillBy:        layerDef.FillBy,
		ColorMap:      layerDef.ColorMap,
		VaryFill:      layerDef.VaryFill,
		Patterns:      patterns,
		ShowLabels:    layerDef.Labels,
		LabelProperty: labelProp,
	}, nil
}

// applyOperation executes a geometry operation on features.
func (b *Builder) applyOperation(op config.Operation, features geometry.FeatureCollection, bounds orb.Bound) (geometry.FeatureCollection, error) {
	switch op.Type {
	case "clip":
		result := geometry.ClipCollection(features, bounds)
		// Debug: log dropped features
		if b.Verbose && len(result) < len(features) {
			for _, f := range features {
				found := false
				name := ""
				if n, ok := f.Properties["NAME"].(string); ok {
					name = n
				}
				for _, r := range result {
					if rn, ok := r.Properties["NAME"].(string); ok && rn == name {
						found = true
						break
					}
				}
				if !found && name != "" {
					b.logf("    Clip dropped: %s", name)
				}
			}
		}
		return result, nil

	case "simplify":
		tolerance := 0.0001 // default
		if t, ok := op.Params["tolerance"].(float64); ok {
			tolerance = t
		}
		return geometry.SimplifyCollection(features, tolerance), nil

	case "filter":
		field, _ := op.Params["field"].(string)
		operator, _ := op.Params["operator"].(string)
		value, _ := op.Params["value"].(string)

		if field == "" || value == "" {
			return nil, fmt.Errorf("filter requires 'field' and 'value' parameters")
		}
		if operator == "" {
			operator = "=" // default to equality
		}

		before := len(features)
		result := geometry.FilterCollection(features, field, operator, value)
		b.logf("  Filtered by %s %s %s: %d → %d features", field, operator, value, before, len(result))
		return result, nil

	case "filter_by_size":
		// Filter features by their rendered size in mm
		minSize := 0.0
		if ms, ok := op.Params["min_size"].(float64); ok {
			minSize = ms
		} else if ms, ok := op.Params["min_size"].(int); ok {
			minSize = float64(ms)
		}
		if minSize <= 0 {
			return features, nil
		}
		before := len(features)
		result := b.filterByOutputSize(features, bounds, minSize)
		b.logf("  Filtered by size (>%.1fmm): %d → %d features", minSize, before, len(result))
		return result, nil

	case "subtract":
		// Get the source to subtract
		sourceURI, ok := op.Params["source"].(string)
		if !ok || sourceURI == "" {
			return nil, fmt.Errorf("subtract requires 'source' parameter")
		}

		// Check for by_county mode
		byCounty, _ := op.Params["by_county"].(bool)

		// Check for min_size filter (in mm) - filters subtract source before subtracting
		minSize := 0.0
		if ms, ok := op.Params["min_size"].(float64); ok {
			minSize = ms
		} else if ms, ok := op.Params["min_size"].(int); ok {
			minSize = float64(ms)
		}

		// Check for emit_as - emits the subtracted features as a separate layer
		// This ensures the water display layer uses EXACTLY the same features as subtraction
		emitAs, _ := op.Params["emit_as"].(string)
		emitStyle := make(map[string]string)
		if style, ok := op.Params["emit_style"].(map[string]interface{}); ok {
			for k, v := range style {
				if s, ok := v.(string); ok {
					emitStyle[k] = s
				}
			}
		}
		emitOrder := 10 // Default to low order (rendered first/bottom)
		if order, ok := op.Params["emit_order"].(int); ok {
			emitOrder = order
		} else if order, ok := op.Params["emit_order"].(float64); ok {
			emitOrder = int(order)
		}

		if byCounty {
			result, waterUsed, err := b.subtractByCountyWithEmit(features, sourceURI, bounds, minSize)
			if err != nil {
				return nil, err
			}
			// Emit water layer if requested
			if emitAs != "" && len(waterUsed) > 0 {
				b.emittedLayers[emitAs] = &emittedLayer{
					Features: waterUsed,
					Style:    emitStyle,
					Order:    emitOrder,
				}
				b.logf("  Emitting %d water features as layer '%s'", len(waterUsed), emitAs)
			}
			return result, nil
		}

		// Standard subtraction - resolve all at once
		clipFeatures, err := b.Resolver.Resolve(sourceURI)
		if err != nil {
			return nil, fmt.Errorf("resolving subtract source: %w", err)
		}

		if len(clipFeatures) == 0 {
			return features, nil // Nothing to subtract
		}

		// Filter by min_size if specified
		if minSize > 0 {
			clipFeatures = b.filterByOutputSize(clipFeatures, bounds, minSize)
		}

		b.logf("  Subtracting %d features from %s", len(clipFeatures), sourceURI)

		// Emit water layer if requested
		if emitAs != "" && len(clipFeatures) > 0 {
			b.emittedLayers[emitAs] = &emittedLayer{
				Features: clipFeatures,
				Style:    emitStyle,
				Order:    emitOrder,
			}
			b.logf("  Emitting %d water features as layer '%s'", len(clipFeatures), emitAs)
		}

		// Perform the subtraction
		result, err := geometry.SubtractCollection(features, clipFeatures)
		if err != nil {
			return nil, fmt.Errorf("subtract operation failed: %w", err)
		}
		return result, nil

	case "merge":
		if len(features) == 0 {
			return features, nil
		}
		merged := geometry.MergeFeatures(features...)
		if merged != nil {
			return geometry.FeatureCollection{merged}, nil
		}
		return features, nil

	case "dissolve":
		// Get the field to dissolve on
		field, ok := op.Params["field"].(string)
		if !ok || field == "" {
			return nil, fmt.Errorf("dissolve requires 'field' parameter")
		}

		b.logf("  Dissolving %d features by %s", len(features), field)

		result, err := geometry.Dissolve(features, field)
		if err != nil {
			return nil, fmt.Errorf("dissolve operation failed: %w", err)
		}

		b.logf("  Dissolved to %d features", len(result))
		return result, nil

	case "buffer":
		// TODO: Implement when geometry.Buffer is ready
		b.logf("  Warning: buffer operation not yet implemented, skipping")
		return features, nil

	case "tag_output_size":
		// Tag features with their rendered output size category
		// Uses output dimensions from recipe to calculate mm size
		return b.tagOutputSize(features, bounds)

	case "filter_output_size":
		// Filter features by minimum rendered output size in mm
		minMM := 2.0 // default 2mm
		if m, ok := op.Params["min_mm"].(float64); ok {
			minMM = m
		}
		return b.filterOutputSize(features, bounds, minMM)

	default:
		return features, fmt.Errorf("unknown operation: %s", op.Type)
	}
}

// mergeStyles combines default styles with layer-specific styles.
// Layer styles override defaults.
func (b *Builder) mergeStyles(defaults, layer map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy defaults first
	for k, v := range defaults {
		result[k] = v
	}

	// Layer styles override defaults
	for k, v := range layer {
		result[k] = v
	}

	return result
}

// parseStyle converts recipe style map to output.Style.
func (b *Builder) parseStyle(styleMap map[string]string) output.Style {
	style := output.DefaultStyle()

	if stroke, ok := styleMap["stroke"]; ok {
		style.Stroke = stroke
	}
	if fill, ok := styleMap["fill"]; ok {
		style.Fill = fill
	}
	if sw, ok := styleMap["stroke_width"]; ok {
		fmt.Sscanf(sw, "%f", &style.StrokeWidth)
	}
	if op, ok := styleMap["opacity"]; ok {
		fmt.Sscanf(op, "%f", &style.Opacity)
	}

	return style
}

// calculateOutputSizeMM calculates the rendered size of a feature's bounding box in mm.
// Returns (widthMM, heightMM, maxDimensionMM).
func (b *Builder) calculateOutputSizeMM(geomBounds orb.Bound, mapBounds orb.Bound) (float64, float64, float64) {
	// Output dimensions in mm (convert from inches)
	outputWidthMM := b.Recipe.Output.Width * 25.4
	outputHeightMM := b.Recipe.Output.Height * 25.4
	marginMM := b.Recipe.Output.Margin * 25.4
	if marginMM == 0 {
		marginMM = 0.5 * 25.4 // default 0.5 inch margin
	}

	// Drawable area
	drawableWidthMM := outputWidthMM - 2*marginMM
	drawableHeightMM := outputHeightMM - 2*marginMM

	// Geographic extent
	geoWidth := mapBounds.Max[0] - mapBounds.Min[0]
	geoHeight := mapBounds.Max[1] - mapBounds.Min[1]

	// Apply latitude correction (same as SVG writer)
	centerLat := (mapBounds.Min[1] + mapBounds.Max[1]) / 2
	latCorrection := math.Cos(centerLat * math.Pi / 180)
	effectiveGeoWidth := geoWidth * latCorrection

	// Scale to fit (maintain aspect ratio)
	scaleX := drawableWidthMM / effectiveGeoWidth
	scaleY := drawableHeightMM / geoHeight
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Feature's geographic size
	featureGeoWidth := (geomBounds.Max[0] - geomBounds.Min[0]) * latCorrection
	featureGeoHeight := geomBounds.Max[1] - geomBounds.Min[1]

	// Convert to output mm
	widthMM := featureGeoWidth * scale
	heightMM := featureGeoHeight * scale
	maxMM := widthMM
	if heightMM > maxMM {
		maxMM = heightMM
	}

	return widthMM, heightMM, maxMM
}

// tagOutputSize tags each feature with its rendered output size category.
// Categories: "<1mm", "1-2mm", "2-3mm", "3-5mm", ">5mm"
func (b *Builder) tagOutputSize(features geometry.FeatureCollection, mapBounds orb.Bound) (geometry.FeatureCollection, error) {
	counts := make(map[string]int)

	for _, f := range features {
		if f.Geometry == nil {
			continue
		}

		geomBounds := f.Geometry.Bound()
		_, _, maxMM := b.calculateOutputSizeMM(geomBounds, mapBounds)

		var category string
		switch {
		case maxMM < 1.0:
			category = "<1mm"
		case maxMM < 2.0:
			category = "1-2mm"
		case maxMM < 3.0:
			category = "2-3mm"
		case maxMM < 5.0:
			category = "3-5mm"
		default:
			category = ">5mm"
		}

		f.Properties["output_size"] = category
		counts[category]++
	}

	b.logf("  Tagged features by output size: <1mm=%d, 1-2mm=%d, 2-3mm=%d, 3-5mm=%d, >5mm=%d",
		counts["<1mm"], counts["1-2mm"], counts["2-3mm"], counts["3-5mm"], counts[">5mm"])

	return features, nil
}

// filterOutputSize removes features smaller than minMM in their largest dimension.
func (b *Builder) filterOutputSize(features geometry.FeatureCollection, mapBounds orb.Bound, minMM float64) (geometry.FeatureCollection, error) {
	var result geometry.FeatureCollection
	filtered := 0

	for _, f := range features {
		if f.Geometry == nil {
			continue
		}

		geomBounds := f.Geometry.Bound()
		_, _, maxMM := b.calculateOutputSizeMM(geomBounds, mapBounds)

		if maxMM >= minMM {
			result = append(result, f)
		} else {
			filtered++
		}
	}

	b.logf("  Filtered by output size (min %.1fmm): %d → %d features (%d removed)",
		minMM, len(features), len(result), filtered)

	return result, nil
}

// filterByOutputSize removes features that would render smaller than minMM in the output.
// A feature is kept if EITHER its width OR height exceeds the threshold (preserves thin features like bays).
func (b *Builder) filterByOutputSize(features geometry.FeatureCollection, bounds orb.Bound, minMM float64) geometry.FeatureCollection {
	// Calculate mm per degree for this output
	// Output dimensions in mm (assuming inches, convert to mm)
	outputWidthMM := b.Recipe.Output.Width * 25.4
	outputHeightMM := b.Recipe.Output.Height * 25.4

	// Account for margins
	margin := b.Recipe.Output.Margin
	if margin == 0 {
		margin = 0.5
	}
	marginMM := margin * 25.4
	drawableWidthMM := outputWidthMM - 2*marginMM
	drawableHeightMM := outputHeightMM - 2*marginMM

	// Geographic extent
	geoWidth := bounds.Max[0] - bounds.Min[0]
	geoHeight := bounds.Max[1] - bounds.Min[1]

	// Scale factors (mm per degree)
	scaleX := drawableWidthMM / geoWidth
	scaleY := drawableHeightMM / geoHeight
	// Use the smaller scale to maintain aspect ratio
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	var result geometry.FeatureCollection
	for _, f := range features {
		if f == nil || f.Geometry == nil {
			continue
		}

		// Get feature's bounding box
		featureBounds := f.Geometry.Bound()
		featureWidth := featureBounds.Max[0] - featureBounds.Min[0]
		featureHeight := featureBounds.Max[1] - featureBounds.Min[1]

		// Convert to mm
		widthMM := featureWidth * scale
		heightMM := featureHeight * scale

		// Keep if EITHER dimension exceeds threshold (preserves thin features)
		if widthMM >= minMM || heightMM >= minMM {
			result = append(result, f)
		}
	}

	return result
}

// subtractByCounty performs water subtraction county-by-county to avoid polygon limits.
// It groups input features by COUNTYFP, fetches water for each county, and subtracts.
func (b *Builder) subtractByCounty(features geometry.FeatureCollection, sourceURI string, bounds orb.Bound, minSizeMM float64) (geometry.FeatureCollection, error) {
	// Parse the source URI to get state info
	parsed, err := sources.ParseCensusURI(sourceURI)
	if err != nil {
		return nil, fmt.Errorf("by_county requires a census:// URI: %w", err)
	}

	// Group features by COUNTYFP
	countyGroups := make(map[string]geometry.FeatureCollection)
	for _, f := range features {
		countyFP := ""
		if fp, ok := f.Properties["COUNTYFP"].(string); ok {
			countyFP = fp
		} else if fp, ok := f.Properties["COUNTYFP20"].(string); ok {
			countyFP = fp
		}
		if countyFP == "" {
			// Features without COUNTYFP go into a special group
			countyFP = "_unknown"
		}
		countyGroups[countyFP] = append(countyGroups[countyFP], f)
	}

	b.logf("  Grouped %d features into %d counties for by_county subtract", len(features), len(countyGroups))

	// Process each county
	var result geometry.FeatureCollection
	for countyFP, countyFeatures := range countyGroups {
		if countyFP == "_unknown" {
			// Can't subtract water from features without county info
			result = append(result, countyFeatures...)
			continue
		}

		// Construct per-county water URI
		// e.g., census://areawater/VT becomes a URL like tl_2023_50007_areawater.zip for county 007
		fullFIPS := parsed.StateFIPS + countyFP
		waterURL := fmt.Sprintf("https://www2.census.gov/geo/tiger/TIGER%s/%s/tl_%s_%s_%s.zip",
			parsed.Year, parsed.TypeInfo.Folder, parsed.Year, fullFIPS, parsed.Type)

		b.logf("    County %s: %d features, fetching water...", countyFP, len(countyFeatures))

		// Fetch water for this county
		waterFeatures, err := b.Resolver.Resolve(waterURL)
		if err != nil {
			b.logf("    Warning: could not fetch water for county %s: %v", countyFP, err)
			result = append(result, countyFeatures...)
			continue
		}

		if len(waterFeatures) == 0 {
			result = append(result, countyFeatures...)
			continue
		}

		// Filter water by min_size if specified
		if minSizeMM > 0 {
			before := len(waterFeatures)
			waterFeatures = b.filterByOutputSize(waterFeatures, bounds, minSizeMM)
			if len(waterFeatures) != before {
				b.logf("    County %s: filtered water by size (>%.1fmm): %d → %d", countyFP, minSizeMM, before, len(waterFeatures))
			}
		}

		if len(waterFeatures) == 0 {
			result = append(result, countyFeatures...)
			continue
		}

		b.logf("    County %s: subtracting %d water features", countyFP, len(waterFeatures))

		// Subtract water from this county's features
		subtracted, err := geometry.SubtractCollection(countyFeatures, waterFeatures)
		if err != nil {
			b.logf("    Warning: subtract failed for county %s: %v", countyFP, err)
			result = append(result, countyFeatures...)
			continue
		}

		result = append(result, subtracted...)
	}

	b.logf("  By-county subtract complete: %d → %d features", len(features), len(result))
	return result, nil
}

// subtractByCountyWithEmit performs water subtraction county-by-county and returns
// both the result AND the water features that were used for subtraction.
// This enables the emit_as feature to create a water layer with exactly the same features.
func (b *Builder) subtractByCountyWithEmit(features geometry.FeatureCollection, sourceURI string, bounds orb.Bound, minSizeMM float64) (geometry.FeatureCollection, geometry.FeatureCollection, error) {
	// Parse the source URI to get state info
	parsed, err := sources.ParseCensusURI(sourceURI)
	if err != nil {
		return nil, nil, fmt.Errorf("by_county requires a census:// URI: %w", err)
	}

	// Group features by COUNTYFP
	countyGroups := make(map[string]geometry.FeatureCollection)
	for _, f := range features {
		countyFP := ""
		if fp, ok := f.Properties["COUNTYFP"].(string); ok {
			countyFP = fp
		} else if fp, ok := f.Properties["COUNTYFP20"].(string); ok {
			countyFP = fp
		}
		if countyFP == "" {
			countyFP = "_unknown"
		}
		countyGroups[countyFP] = append(countyGroups[countyFP], f)
	}

	b.logf("  Grouped %d features into %d counties for by_county subtract", len(features), len(countyGroups))

	// Process each county
	var result geometry.FeatureCollection
	var allWaterUsed geometry.FeatureCollection

	for countyFP, countyFeatures := range countyGroups {
		if countyFP == "_unknown" {
			result = append(result, countyFeatures...)
			continue
		}

		// Construct per-county water URI
		fullFIPS := parsed.StateFIPS + countyFP
		waterURL := fmt.Sprintf("https://www2.census.gov/geo/tiger/TIGER%s/%s/tl_%s_%s_%s.zip",
			parsed.Year, parsed.TypeInfo.Folder, parsed.Year, fullFIPS, parsed.Type)

		b.logf("    County %s: %d features, fetching water...", countyFP, len(countyFeatures))

		// Fetch water for this county
		waterFeatures, err := b.Resolver.Resolve(waterURL)
		if err != nil {
			b.logf("    Warning: could not fetch water for county %s: %v", countyFP, err)
			result = append(result, countyFeatures...)
			continue
		}

		if len(waterFeatures) == 0 {
			result = append(result, countyFeatures...)
			continue
		}

		// Filter water by min_size if specified
		if minSizeMM > 0 {
			before := len(waterFeatures)
			waterFeatures = b.filterByOutputSize(waterFeatures, bounds, minSizeMM)
			if len(waterFeatures) != before {
				b.logf("    County %s: filtered water by size (>%.1fmm): %d → %d", countyFP, minSizeMM, before, len(waterFeatures))
			}
		}

		if len(waterFeatures) == 0 {
			result = append(result, countyFeatures...)
			continue
		}

		// Collect water features for emission
		allWaterUsed = append(allWaterUsed, waterFeatures...)

		b.logf("    County %s: subtracting %d water features", countyFP, len(waterFeatures))

		// Subtract water from this county's features
		subtracted, err := geometry.SubtractCollection(countyFeatures, waterFeatures)
		if err != nil {
			b.logf("    Warning: subtract failed for county %s: %v", countyFP, err)
			result = append(result, countyFeatures...)
			continue
		}

		result = append(result, subtracted...)
	}

	b.logf("  By-county subtract complete: %d → %d features, collected %d water features", len(features), len(result), len(allWaterUsed))
	return result, allWaterUsed, nil
}

// logf prints a message if verbose mode is enabled.
func (b *Builder) logf(format string, args ...interface{}) {
	if b.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}
