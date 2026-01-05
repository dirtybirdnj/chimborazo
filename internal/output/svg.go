// Package output provides SVG and other output format writers for Chimborazo.
package output

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/paulmach/orb"

	"github.com/dirtybirdnj/chimborazo/internal/geometry"
)

// SVGWriter generates SVG output from geographic features.
type SVGWriter struct {
	Width        float64   // page width in inches
	Height       float64   // page height in inches
	Margin       float64   // margin in inches
	Bounds       orb.Bound // geographic bounds to render
	DPI          float64   // dots per inch (default 96)
	ShowRulers   bool      // draw inch rulers on left and top edges
	NoBackground bool      // skip white background rect (for plotter/rat-king output)
}

// Style defines the visual appearance of a feature.
type Style struct {
	Stroke      string  // stroke color (e.g., "#ff0000" or "none")
	StrokeWidth float64 // stroke width in mm (converted to pixels at render time)
	Fill        string  // fill color (e.g., "#00ff00" or "none")
	Opacity     float64 // overall opacity (0.0 - 1.0)
}

// Layer represents a named group of features with a style.
type Layer struct {
	Name          string
	Features      geometry.FeatureCollection
	Style         Style
	Order         int
	FillBy        string            // Property name to color by
	ColorMap      map[string]string // Property value → fill color
	VaryFill      bool              // Apply slight color variations
	Patterns      []string          // Pattern names to cycle through (for rat-king data attributes)
	ShowLabels    bool              // Show labels for features
	LabelProperty string            // Property to use for labels (e.g., "NAME")
}

// DefaultStyle returns a reasonable default style.
func DefaultStyle() Style {
	return Style{
		Stroke:      "#333333",
		StrokeWidth: 0.3, // 0.3mm - typical pen plotter width
		Fill:        "none",
		Opacity:     1.0,
	}
}

// NewSVGWriter creates a new SVG writer with the given dimensions and bounds.
func NewSVGWriter(width, height, margin float64, bounds orb.Bound) *SVGWriter {
	return &SVGWriter{
		Width:  width,
		Height: height,
		Margin: margin,
		Bounds: bounds,
		DPI:    96.0,
	}
}

// viewportSize returns the drawable area in pixels (after margins).
func (w *SVGWriter) viewportSize() (float64, float64) {
	marginPx := w.Margin * w.DPI
	widthPx := w.Width*w.DPI - 2*marginPx
	heightPx := w.Height*w.DPI - 2*marginPx
	return widthPx, heightPx
}

// projectPoint transforms a geographic point to SVG coordinates.
// Y is flipped because SVG origin is top-left, but geo origin is bottom-left.
// Applies latitude correction so maps don't appear horizontally stretched.
func (w *SVGWriter) projectPoint(p orb.Point) (float64, float64) {
	marginPx := w.Margin * w.DPI
	vpWidth, vpHeight := w.viewportSize()

	// Geographic extent
	geoWidth := w.Bounds.Max[0] - w.Bounds.Min[0]
	geoHeight := w.Bounds.Max[1] - w.Bounds.Min[1]

	// Apply latitude correction - at higher latitudes, degrees of longitude
	// cover less ground distance than degrees of latitude
	centerLat := (w.Bounds.Min[1] + w.Bounds.Max[1]) / 2
	latCorrection := math.Cos(centerLat * math.Pi / 180)
	effectiveWidth := geoWidth * latCorrection

	// Scale to fit (maintain aspect ratio)
	scaleX := vpWidth / effectiveWidth
	scaleY := vpHeight / geoHeight
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Center the map
	offsetX := (vpWidth - effectiveWidth*scale) / 2
	offsetY := (vpHeight - geoHeight*scale) / 2

	// Transform point - apply same latitude correction to x coordinate
	x := marginPx + offsetX + (p[0]-w.Bounds.Min[0])*latCorrection*scale
	y := marginPx + offsetY + (w.Bounds.Max[1]-p[1])*scale // flip Y

	return x, y
}

// mmToPixels converts millimeters to pixels at the current DPI.
func (w *SVGWriter) mmToPixels(mm float64) float64 {
	return (mm / 25.4) * w.DPI
}

// pointToPath converts a Point to an SVG circle element.
func (w *SVGWriter) pointToPath(p orb.Point, style Style) string {
	x, y := w.projectPoint(p)
	strokeWidthPx := w.mmToPixels(style.StrokeWidth)
	return fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.2f" stroke="%s" stroke-width="%.2f" fill="%s" opacity="%.2f"/>`,
		x, y, strokeWidthPx*2, style.Stroke, strokeWidthPx, style.Fill, style.Opacity)
}

// lineStringToPath converts a LineString to an SVG path.
func (w *SVGWriter) lineStringToPath(ls orb.LineString) string {
	if len(ls) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, p := range ls {
		x, y := w.projectPoint(p)
		if i == 0 {
			sb.WriteString(fmt.Sprintf("M%.2f %.2f", x, y))
		} else {
			sb.WriteString(fmt.Sprintf(" L%.2f %.2f", x, y))
		}
	}
	return sb.String()
}

// ringToPath converts a Ring (closed LineString) to an SVG path.
func (w *SVGWriter) ringToPath(r orb.Ring) string {
	if len(r) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, p := range r {
		x, y := w.projectPoint(p)
		if i == 0 {
			sb.WriteString(fmt.Sprintf("M%.2f %.2f", x, y))
		} else {
			sb.WriteString(fmt.Sprintf(" L%.2f %.2f", x, y))
		}
	}
	sb.WriteString(" Z") // close path
	return sb.String()
}

// polygonToPath converts a Polygon to an SVG path (with holes).
func (w *SVGWriter) polygonToPath(poly orb.Polygon) string {
	var sb strings.Builder
	for _, ring := range poly {
		sb.WriteString(w.ringToPath(ring))
		sb.WriteString(" ")
	}
	return strings.TrimSpace(sb.String())
}

// geometryToPath converts any orb.Geometry to SVG path data.
func (w *SVGWriter) geometryToPath(g orb.Geometry) string {
	switch v := g.(type) {
	case orb.Point:
		// Points are handled separately as circles
		return ""
	case orb.LineString:
		return w.lineStringToPath(v)
	case orb.Ring:
		return w.ringToPath(v)
	case orb.Polygon:
		return w.polygonToPath(v)
	case orb.MultiPoint:
		// Handle in WriteFeature
		return ""
	case orb.MultiLineString:
		var sb strings.Builder
		for _, ls := range v {
			sb.WriteString(w.lineStringToPath(ls))
			sb.WriteString(" ")
		}
		return strings.TrimSpace(sb.String())
	case orb.MultiPolygon:
		var sb strings.Builder
		for _, poly := range v {
			sb.WriteString(w.polygonToPath(poly))
			sb.WriteString(" ")
		}
		return strings.TrimSpace(sb.String())
	default:
		return ""
	}
}

// WriteFeature converts a Feature to SVG element(s).
func (w *SVGWriter) WriteFeature(f *geometry.Feature, style Style) string {
	return w.WriteFeatureWithAttrs(f, style, nil)
}

// WriteFeatureWithAttrs converts a Feature to SVG element(s) with optional data attributes.
func (w *SVGWriter) WriteFeatureWithAttrs(f *geometry.Feature, style Style, dataAttrs map[string]string) string {
	if f == nil || f.Geometry == nil {
		return ""
	}

	// Handle points specially (as circles)
	switch v := f.Geometry.(type) {
	case orb.Point:
		return w.pointToPath(v, style)
	case orb.MultiPoint:
		var sb strings.Builder
		for _, p := range v {
			sb.WriteString(w.pointToPath(p, style))
			sb.WriteString("\n")
		}
		return sb.String()
	}

	// All other geometries become paths
	pathData := w.geometryToPath(f.Geometry)
	if pathData == "" {
		return ""
	}

	strokeWidthPx := w.mmToPixels(style.StrokeWidth)

	// Build data attributes string
	var attrStr string
	if len(dataAttrs) > 0 {
		var attrs []string
		for k, v := range dataAttrs {
			attrs = append(attrs, fmt.Sprintf(`data-%s="%s"`, k, v))
		}
		// Sort for consistent output
		sort.Strings(attrs)
		attrStr = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf(`<path d="%s" stroke="%s" stroke-width="%.2f" fill="%s" opacity="%.2f" fill-rule="evenodd"%s/>`,
		pathData, style.Stroke, strokeWidthPx, style.Fill, style.Opacity, attrStr)
}

// WriteCollection converts a FeatureCollection to SVG elements.
func (w *SVGWriter) WriteCollection(fc geometry.FeatureCollection, style Style) string {
	var sb strings.Builder
	for _, f := range fc {
		element := w.WriteFeature(f, style)
		if element != "" {
			sb.WriteString("    ")
			sb.WriteString(element)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// WriteCollectionWithColors converts a FeatureCollection with per-feature coloring.
func (w *SVGWriter) WriteCollectionWithColors(fc geometry.FeatureCollection, style Style, fillBy string, colorMap map[string]string, varyFill bool) string {
	return w.WriteCollectionWithColorsAndPatterns(fc, style, fillBy, colorMap, varyFill, nil)
}

// WriteCollectionWithColorsAndPatterns converts a FeatureCollection with per-feature coloring and pattern attributes.
func (w *SVGWriter) WriteCollectionWithColorsAndPatterns(fc geometry.FeatureCollection, style Style, fillBy string, colorMap map[string]string, varyFill bool, patterns []string) string {
	var sb strings.Builder
	for i, f := range fc {
		// Determine fill color for this feature
		featureStyle := style
		colorFound := false
		if fillBy != "" && colorMap != nil {
			if propVal, ok := f.Properties[fillBy]; ok {
				key := fmt.Sprintf("%v", propVal)
				if color, exists := colorMap[key]; exists {
					featureStyle.Fill = color
					colorFound = true
				}
			}
		}

		// Calculate shade index (0-5) for this feature
		shade := i % 6

		// Apply variation to the fill color (either from color map or base style)
		if varyFill && featureStyle.Fill != "none" && featureStyle.Fill != "" {
			if colorFound {
				featureStyle.Fill = varyColor(featureStyle.Fill, i)
			} else {
				// Apply variation to base style fill
				featureStyle.Fill = varyColor(style.Fill, i)
			}
		}

		// Build data attributes for rat-king
		var dataAttrs map[string]string
		if len(patterns) > 0 {
			pattern := patterns[i%len(patterns)]
			dataAttrs = map[string]string{
				"pattern": pattern,
				"shade":   fmt.Sprintf("%d", shade),
			}
		}

		element := w.WriteFeatureWithAttrs(f, featureStyle, dataAttrs)
		if element != "" {
			sb.WriteString("    ")
			sb.WriteString(element)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// varyColor applies a variation to a hex color, cycling through 6 distinct shades.
func varyColor(hexColor string, seed int) string {
	// Parse hex color
	if len(hexColor) != 7 || hexColor[0] != '#' {
		return hexColor
	}

	var r, g, b int
	fmt.Sscanf(hexColor, "#%02x%02x%02x", &r, &g, &b)

	// Create 6 distinct shades: -3, -2, -1, 0, +1, +2 levels
	// Each level shifts the color by ~15 RGB units for visible distinction
	shade := seed % 6
	offsets := []int{-3, -2, -1, 0, 1, 2}
	variation := offsets[shade] * 15

	r = clamp(r+variation, 0, 255)
	g = clamp(g+variation, 0, 255)
	b = clamp(b+variation, 0, 255)

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// clamp restricts a value to a range.
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// calculateCentroid calculates the centroid of a geometry.
func (w *SVGWriter) calculateCentroid(g orb.Geometry) (float64, float64) {
	bound := g.Bound()
	// Simple centroid: center of bounding box
	centerLon := (bound.Min[0] + bound.Max[0]) / 2
	centerLat := (bound.Min[1] + bound.Max[1]) / 2
	return w.projectPoint(orb.Point{centerLon, centerLat})
}

// WriteLabels generates SVG text elements for feature labels.
func (w *SVGWriter) WriteLabels(fc geometry.FeatureCollection, labelProperty string) string {
	var sb strings.Builder
	for _, f := range fc {
		if f == nil || f.Geometry == nil {
			continue
		}

		// Get label text from property
		labelText := ""
		if prop, ok := f.Properties[labelProperty]; ok {
			labelText = fmt.Sprintf("%v", prop)
		}
		if labelText == "" {
			continue
		}

		// Calculate centroid
		x, y := w.calculateCentroid(f.Geometry)

		// Write text element with small font
		sb.WriteString(fmt.Sprintf(`    <text x="%.2f" y="%.2f" font-size="6" text-anchor="middle" fill="#333333" font-family="sans-serif">%s</text>
`, x, y, labelText))
	}
	return sb.String()
}

// renderRulers generates SVG elements for inch rulers on left and top edges.
func (w *SVGWriter) renderRulers() string {
	var sb strings.Builder
	inchPx := w.DPI
	marginPx := w.Margin * w.DPI
	widthPx := w.Width * w.DPI
	heightPx := w.Height * w.DPI

	// 0.3mm line width (typical pen width)
	lineWidthPx := (0.3 / 25.4) * w.DPI

	sb.WriteString(fmt.Sprintf(`  <g id="rulers" stroke="#666666" stroke-width="%.2f" fill="none">
`, lineWidthPx))

	// Top ruler (horizontal, along the top edge)
	// Draw ruler baseline
	sb.WriteString(fmt.Sprintf(`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>
`, marginPx, marginPx, widthPx-marginPx, marginPx))

	// Draw inch marks on top ruler
	for i := 0; i <= int(w.Width-2*w.Margin); i++ {
		x := marginPx + float64(i)*inchPx
		// Full inch mark (longer)
		sb.WriteString(fmt.Sprintf(`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>
`, x, marginPx, x, marginPx-10))
		// Half inch mark
		if i < int(w.Width-2*w.Margin) {
			halfX := x + inchPx/2
			sb.WriteString(fmt.Sprintf(`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>
`, halfX, marginPx, halfX, marginPx-6))
		}
		// Label
		sb.WriteString(fmt.Sprintf(`    <text x="%.2f" y="%.2f" font-size="10" text-anchor="middle" fill="#666666">%d</text>
`, x, marginPx-12, i))
	}

	// Left ruler (vertical, along the left edge)
	// Draw ruler baseline
	sb.WriteString(fmt.Sprintf(`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>
`, marginPx, marginPx, marginPx, heightPx-marginPx))

	// Draw inch marks on left ruler
	for i := 0; i <= int(w.Height-2*w.Margin); i++ {
		y := marginPx + float64(i)*inchPx
		// Full inch mark (longer)
		sb.WriteString(fmt.Sprintf(`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>
`, marginPx, y, marginPx-10, y))
		// Half inch mark
		if i < int(w.Height-2*w.Margin) {
			halfY := y + inchPx/2
			sb.WriteString(fmt.Sprintf(`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>
`, marginPx, halfY, marginPx-6, halfY))
		}
		// Label
		sb.WriteString(fmt.Sprintf(`    <text x="%.2f" y="%.2f" font-size="10" text-anchor="end" fill="#666666">%d</text>
`, marginPx-12, y+3, i))
	}

	sb.WriteString("  </g>\n")
	return sb.String()
}

// Render generates a complete SVG document from layers.
func (w *SVGWriter) Render(layers []Layer) string {
	widthPx := w.Width * w.DPI
	heightPx := w.Height * w.DPI

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg"
     width="%.2f" height="%.2f"
     viewBox="0 0 %.2f %.2f">
`, widthPx, heightPx, widthPx, heightPx))

	// Background (optional - skip for plotter output)
	if !w.NoBackground {
		sb.WriteString(fmt.Sprintf(`  <rect width="%.2f" height="%.2f" fill="white"/>
`, widthPx, heightPx))
	}

	// Render each layer in a group
	for _, layer := range layers {
		sb.WriteString(fmt.Sprintf(`  <g id="%s">
`, layer.Name))
		// Use color mapping if fill_by is specified, vary_fill for variations, or patterns for rat-king
		if (layer.FillBy != "" && layer.ColorMap != nil) || layer.VaryFill || len(layer.Patterns) > 0 {
			sb.WriteString(w.WriteCollectionWithColorsAndPatterns(layer.Features, layer.Style, layer.FillBy, layer.ColorMap, layer.VaryFill, layer.Patterns))
		} else {
			sb.WriteString(w.WriteCollection(layer.Features, layer.Style))
		}
		// Add labels if enabled
		if layer.ShowLabels && layer.LabelProperty != "" {
			sb.WriteString(w.WriteLabels(layer.Features, layer.LabelProperty))
		}
		sb.WriteString("  </g>\n")
	}

	// Rulers (if enabled)
	if w.ShowRulers {
		sb.WriteString(w.renderRulers())
	}

	sb.WriteString("</svg>\n")

	return sb.String()
}

// RenderToFile writes the SVG to a file.
func (w *SVGWriter) RenderToFile(layers []Layer, path string) error {
	content := w.Render(layers)
	return os.WriteFile(path, []byte(content), 0644)
}

// RenderSingleLayer generates an SVG document with just one layer (no background).
// Useful for per-layer output for pen plotters.
func (w *SVGWriter) RenderSingleLayer(layer Layer) string {
	widthPx := w.Width * w.DPI
	heightPx := w.Height * w.DPI

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg"
     width="%.2f" height="%.2f"
     viewBox="0 0 %.2f %.2f">
`, widthPx, heightPx, widthPx, heightPx))

	// No background for individual layers - cleaner for plotting

	// Render the single layer
	sb.WriteString(fmt.Sprintf(`  <g id="%s">
`, layer.Name))
	// Use color mapping if fill_by is specified, vary_fill for variations, or patterns for rat-king
	if (layer.FillBy != "" && layer.ColorMap != nil) || layer.VaryFill || len(layer.Patterns) > 0 {
		sb.WriteString(w.WriteCollectionWithColorsAndPatterns(layer.Features, layer.Style, layer.FillBy, layer.ColorMap, layer.VaryFill, layer.Patterns))
	} else {
		sb.WriteString(w.WriteCollection(layer.Features, layer.Style))
	}
	sb.WriteString("  </g>\n")

	sb.WriteString("</svg>\n")

	return sb.String()
}

// RenderLayerToFile writes a single layer SVG to a file.
func (w *SVGWriter) RenderLayerToFile(layer Layer, path string) error {
	content := w.RenderSingleLayer(layer)
	return os.WriteFile(path, []byte(content), 0644)
}
