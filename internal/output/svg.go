// Package output provides SVG and other output format writers for Chimborazo.
package output

import (
	"fmt"
	"strings"

	"github.com/paulmach/orb"

	"github.com/dirtybirdnj/chimborazo/internal/geometry"
)

// SVGWriter generates SVG output from geographic features.
type SVGWriter struct {
	Width  float64   // page width in inches
	Height float64   // page height in inches
	Margin float64   // margin in inches
	Bounds orb.Bound // geographic bounds to render
	DPI    float64   // dots per inch (default 96)
}

// Style defines the visual appearance of a feature.
type Style struct {
	Stroke      string  // stroke color (e.g., "#ff0000" or "none")
	StrokeWidth float64 // stroke width in SVG units
	Fill        string  // fill color (e.g., "#00ff00" or "none")
	Opacity     float64 // overall opacity (0.0 - 1.0)
}

// Layer represents a named group of features with a style.
type Layer struct {
	Name     string
	Features geometry.FeatureCollection
	Style    Style
	Order    int
}

// DefaultStyle returns a reasonable default style.
func DefaultStyle() Style {
	return Style{
		Stroke:      "#333333",
		StrokeWidth: 1.0,
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
func (w *SVGWriter) projectPoint(p orb.Point) (float64, float64) {
	marginPx := w.Margin * w.DPI
	vpWidth, vpHeight := w.viewportSize()

	// Geographic extent
	geoWidth := w.Bounds.Max[0] - w.Bounds.Min[0]
	geoHeight := w.Bounds.Max[1] - w.Bounds.Min[1]

	// Scale to fit (maintain aspect ratio)
	scaleX := vpWidth / geoWidth
	scaleY := vpHeight / geoHeight
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Center the map
	offsetX := (vpWidth - geoWidth*scale) / 2
	offsetY := (vpHeight - geoHeight*scale) / 2

	// Transform point
	x := marginPx + offsetX + (p[0]-w.Bounds.Min[0])*scale
	y := marginPx + offsetY + (w.Bounds.Max[1]-p[1])*scale // flip Y

	return x, y
}

// pointToPath converts a Point to an SVG circle element.
func (w *SVGWriter) pointToPath(p orb.Point, style Style) string {
	x, y := w.projectPoint(p)
	return fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.1f" stroke="%s" stroke-width="%.2f" fill="%s" opacity="%.2f"/>`,
		x, y, style.StrokeWidth*2, style.Stroke, style.StrokeWidth, style.Fill, style.Opacity)
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

	return fmt.Sprintf(`<path d="%s" stroke="%s" stroke-width="%.2f" fill="%s" opacity="%.2f" fill-rule="evenodd"/>`,
		pathData, style.Stroke, style.StrokeWidth, style.Fill, style.Opacity)
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

	// Background (optional)
	sb.WriteString(fmt.Sprintf(`  <rect width="%.2f" height="%.2f" fill="white"/>
`, widthPx, heightPx))

	// Render each layer in a group
	for _, layer := range layers {
		sb.WriteString(fmt.Sprintf(`  <g id="%s">
`, layer.Name))
		sb.WriteString(w.WriteCollection(layer.Features, layer.Style))
		sb.WriteString("  </g>\n")
	}

	sb.WriteString("</svg>\n")

	return sb.String()
}

// RenderToFile writes the SVG to a file.
func (w *SVGWriter) RenderToFile(layers []Layer, path string) error {
	// Import os when implementing
	// content := w.Render(layers)
	// return os.WriteFile(path, []byte(content), 0644)
	return nil // TODO: implement
}
