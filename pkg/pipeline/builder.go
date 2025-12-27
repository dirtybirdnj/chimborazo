// Package pipeline orchestrates the map building process.
// It connects recipe parsing, data fetching, geometry operations, and output generation.
package pipeline

import (
	"fmt"
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
	Recipe   *config.Recipe
	Resolver *sources.Resolver
	Verbose  bool
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

// NewBuilder creates a pipeline builder for a recipe.
func NewBuilder(recipe *config.Recipe, cacheDir string) (*Builder, error) {
	resolver, err := sources.NewResolver(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("creating resolver: %w", err)
	}

	return &Builder{
		Recipe:   recipe,
		Resolver: resolver,
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

	// Step 4: Merge default styles with layer styles
	mergedStyleMap := b.mergeStyles(b.Recipe.Defaults.Style, layerDef.Style)
	style := b.parseStyle(mergedStyleMap)

	// Step 5: Parse order (default to 0)
	order := 0
	if orderStr, ok := mergedStyleMap["order"]; ok {
		fmt.Sscanf(orderStr, "%d", &order)
	}

	return &output.Layer{
		Name:     layerDef.Name,
		Features: features,
		Style:    style,
		Order:    order,
		FillBy:   layerDef.FillBy,
		ColorMap: layerDef.ColorMap,
		VaryFill: layerDef.VaryFill,
	}, nil
}

// applyOperation executes a geometry operation on features.
func (b *Builder) applyOperation(op config.Operation, features geometry.FeatureCollection, bounds orb.Bound) (geometry.FeatureCollection, error) {
	switch op.Type {
	case "clip":
		return geometry.ClipCollection(features, bounds), nil

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

	case "subtract":
		// Get the source to subtract
		sourceURI, ok := op.Params["source"].(string)
		if !ok || sourceURI == "" {
			return nil, fmt.Errorf("subtract requires 'source' parameter")
		}

		// Resolve the clip source
		clipFeatures, err := b.Resolver.Resolve(sourceURI)
		if err != nil {
			return nil, fmt.Errorf("resolving subtract source: %w", err)
		}

		if len(clipFeatures) == 0 {
			return features, nil // Nothing to subtract
		}

		b.logf("  Subtracting %d features from %s", len(clipFeatures), sourceURI)

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

// logf prints a message if verbose mode is enabled.
func (b *Builder) logf(format string, args ...interface{}) {
	if b.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}
