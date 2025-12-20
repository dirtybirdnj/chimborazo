// Package pipeline orchestrates the map building process.
// It connects recipe parsing, data fetching, geometry operations, and output generation.
package pipeline

import (
	"fmt"
	"sort"
	"time"

	"github.com/paulmach/orb"

	"github.com/dirtybirdnj/chimborazo/internal/config"
	"github.com/dirtybirdnj/chimborazo/internal/geometry"
	"github.com/dirtybirdnj/chimborazo/internal/output"
	"github.com/dirtybirdnj/chimborazo/internal/sources"
)

// Builder orchestrates the map building process.
type Builder struct {
	Recipe  *config.Recipe
	Fetcher *sources.Fetcher
	Verbose bool
}

// BuildResult contains the outcome of a build.
type BuildResult struct {
	OutputPath  string
	Duration    time.Duration
	LayerCount  int
	FeatureCount int
	Errors      []error
	Warnings    []string
}

// NewBuilder creates a pipeline builder for a recipe.
func NewBuilder(recipe *config.Recipe, fetcher *sources.Fetcher) *Builder {
	return &Builder{
		Recipe:  recipe,
		Fetcher: fetcher,
	}
}

// Build executes the full pipeline: fetch → process → render.
func (b *Builder) Build() (*BuildResult, error) {
	start := time.Now()
	result := &BuildResult{}

	// Step 1: Validate recipe
	if err := config.ValidateRecipe(b.Recipe); err != nil {
		return nil, fmt.Errorf("invalid recipe: %w", err)
	}

	// Step 2: Determine bounds
	bounds, err := b.resolveBounds()
	if err != nil {
		return nil, fmt.Errorf("resolving bounds: %w", err)
	}

	// Step 3: Process each layer
	var layers []output.Layer
	for _, layerDef := range b.Recipe.Layers {
		layer, err := b.processLayer(layerDef, bounds)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("layer %s: %w", layerDef.Name, err))
			continue
		}
		if layer != nil {
			layers = append(layers, *layer)
			result.FeatureCount += len(layer.Features)
		}
	}

	// Sort layers by order
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].Order < layers[j].Order
	})

	result.LayerCount = len(layers)

	// Step 4: Render output
	writer := output.NewSVGWriter(
		b.Recipe.Output.Width,
		b.Recipe.Output.Height,
		0.5, // default margin
		bounds,
	)

	svgContent := writer.Render(layers)

	// Step 5: Write to file
	// TODO: Actually write to file
	_ = svgContent
	result.OutputPath = b.Recipe.Output.Path

	result.Duration = time.Since(start)
	return result, nil
}

// resolveBounds determines the geographic extent for the map.
func (b *Builder) resolveBounds() (orb.Bound, error) {
	// For now, use a hardcoded Vermont-area bounds
	// TODO: Parse from recipe.Bounds.Source or calculate from data

	// If bounds source is specified, we'd fetch that source and get its bounds
	// For MVP, return a default
	return orb.Bound{
		Min: orb.Point{-73.5, 42.7},
		Max: orb.Point{-71.5, 45.0},
	}, nil
}

// processLayer fetches data and applies operations for a single layer.
func (b *Builder) processLayer(layerDef config.Layer, bounds orb.Bound) (*output.Layer, error) {
	// Step 1: Fetch source data
	features, err := b.fetchSource(layerDef.Source)
	if err != nil {
		return nil, fmt.Errorf("fetching source: %w", err)
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

	// Step 3: Convert style
	style := b.parseStyle(layerDef.Style)

	return &output.Layer{
		Name:     layerDef.Name,
		Features: features,
		Style:    style,
		Order:    0, // TODO: parse from recipe
	}, nil
}

// fetchSource retrieves data from a source URI.
func (b *Builder) fetchSource(uri string) (geometry.FeatureCollection, error) {
	// TODO: Implement source resolution
	// For now, return empty collection
	//
	// The source URI format is: scheme:path
	// Examples:
	//   census:tiger/2023/vt/cousub
	//   file:./data/towns.geojson
	//   url:https://example.com/data.json
	//
	// Each scheme needs a resolver that:
	// 1. Constructs the actual URL
	// 2. Uses the Fetcher to download/cache
	// 3. Parses the response (GeoJSON, Shapefile, etc.)
	// 4. Returns a FeatureCollection

	return nil, fmt.Errorf("source fetching not yet implemented: %s", uri)
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

	case "subtract":
		// TODO: Implement when geometry.Subtract is ready
		return features, nil

	case "merge":
		if len(features) == 0 {
			return features, nil
		}
		merged := geometry.MergeFeatures(features...)
		if merged != nil {
			return geometry.FeatureCollection{merged}, nil
		}
		return features, nil

	case "buffer":
		// TODO: Implement when geometry.Buffer is ready
		return features, nil

	default:
		return features, fmt.Errorf("unknown operation: %s", op.Type)
	}
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
		// Parse string to float
		fmt.Sscanf(sw, "%f", &style.StrokeWidth)
	}
	if op, ok := styleMap["opacity"]; ok {
		fmt.Sscanf(op, "%f", &style.Opacity)
	}

	return style
}

// Logf prints a message if verbose mode is enabled.
func (b *Builder) Logf(format string, args ...interface{}) {
	if b.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}
