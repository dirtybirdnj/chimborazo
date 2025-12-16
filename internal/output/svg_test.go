package output

import (
	"strings"
	"testing"
)

func TestNewSVGWriter(t *testing.T) {
	w := NewSVGWriter(800, 600)

	if w.Width != 800 {
		t.Errorf("Width = %d, want 800", w.Width)
	}
	if w.Height != 600 {
		t.Errorf("Height = %d, want 600", w.Height)
	}
}

func TestSVGWriter_Header(t *testing.T) {
	w := NewSVGWriter(800, 600)
	header := w.Header()

	if !strings.Contains(header, "xmlns") {
		t.Error("Header missing xmlns attribute")
	}
	if !strings.Contains(header, "viewBox") {
		t.Error("Header missing viewBox attribute")
	}
	if !strings.Contains(header, "800") || !strings.Contains(header, "600") {
		t.Error("Header missing dimensions")
	}
	if !strings.HasPrefix(header, "<svg") {
		t.Error("Header should start with <svg")
	}
}

func TestSVGWriter_Footer(t *testing.T) {
	w := NewSVGWriter(100, 100)
	footer := w.Footer()

	if footer != "</svg>" {
		t.Errorf("Footer = %q, want </svg>", footer)
	}
}

func TestSVGWriter_PolygonToPath(t *testing.T) {
	w := NewSVGWriter(100, 100)

	tests := []struct {
		name   string
		points [][2]float64
		want   string
	}{
		{
			name:   "square",
			points: [][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
			want:   "M 0.00 0.00 L 100.00 0.00 L 100.00 100.00 L 0.00 100.00 Z",
		},
		{
			name:   "triangle",
			points: [][2]float64{{50, 0}, {100, 100}, {0, 100}},
			want:   "M 50.00 0.00 L 100.00 100.00 L 0.00 100.00 Z",
		},
		{
			name:   "empty",
			points: [][2]float64{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := w.PolygonToPath(tt.points)
			if got != tt.want {
				t.Errorf("PolygonToPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSVGWriter_RoundTrip(t *testing.T) {
	w := NewSVGWriter(200, 150)

	// Build a complete SVG
	svg := w.Header() + "\n"
	svg += `<path d="` + w.PolygonToPath([][2]float64{{10, 10}, {190, 10}, {190, 140}, {10, 140}}) + `"/>` + "\n"
	svg += w.Footer()

	// Verify structure
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("SVG should start with <svg")
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("SVG should end with </svg>")
	}
	if !strings.Contains(svg, "<path") {
		t.Error("SVG should contain path element")
	}
}
