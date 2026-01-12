package output

import (
	"fmt"
	"strings"
)

// SVGWriter generates SVG output for map geometries
type SVGWriter struct {
	Width  int
	Height int
}

// NewSVGWriter creates a writer with the given dimensions
func NewSVGWriter(width, height int) *SVGWriter {
	return &SVGWriter{
		Width:  width,
		Height: height,
	}
}

// Header returns the SVG opening tag with viewBox
func (s *SVGWriter) Header() string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`, s.Width, s.Height)
}

// Footer returns the SVG closing tag
func (s *SVGWriter) Footer() string {
	return "</svg>"
}

// PolygonToPath converts a slice of points to an SVG path d attribute
func (s *SVGWriter) PolygonToPath(points [][2]float64) string {
	if len(points) == 0 {
		return ""
	}

	var parts []string
	for i, point := range points {
		if i == 0 {
			parts = append(parts, fmt.Sprintf("M %.2f %.2f", point[0], point[1]))
		} else {
			parts = append(parts, fmt.Sprintf("L %.2f %.2f", point[0], point[1]))
		}
	}
	parts = append(parts, "Z")
	return strings.Join(parts, " ")
}
