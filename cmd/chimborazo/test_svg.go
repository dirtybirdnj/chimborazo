// +build ignore

// Test script to generate a sample SVG
// Run with: go run test_svg.go

package main

import (
	"fmt"
	"os"

	"github.com/paulmach/orb"

	"github.com/dirtybirdnj/chimborazo/internal/geometry"
	"github.com/dirtybirdnj/chimborazo/internal/output"
)

func main() {
	// Vermont-ish bounds
	bounds := orb.Bound{
		Min: orb.Point{-73.5, 42.7},
		Max: orb.Point{-71.5, 45.2},
	}

	// Create SVG writer (12x18 inch page)
	writer := output.NewSVGWriter(12, 18, 0.5, bounds)

	// Create some sample features

	// Lake Champlain (simplified polygon)
	lakeChamplain := orb.Polygon{
		orb.Ring{
			{-73.4, 44.0},
			{-73.3, 44.5},
			{-73.2, 45.0},
			{-73.3, 45.0},
			{-73.4, 44.8},
			{-73.45, 44.3},
			{-73.4, 44.0},
		},
	}

	// Vermont border (very simplified)
	vermontBorder := orb.Polygon{
		orb.Ring{
			{-73.3, 42.73}, // SW corner
			{-72.5, 42.73}, // SE-ish
			{-71.5, 42.73}, // SE corner
			{-71.5, 45.0},  // NE corner
			{-72.0, 45.0},  // border jog
			{-73.3, 45.0},  // NW-ish
			{-73.3, 44.0},  // Lake Champlain area
			{-73.3, 42.73}, // back to start
		},
	}

	// Some towns (as points)
	burlington := orb.Point{-73.21, 44.48}
	montpelier := orb.Point{-72.58, 44.26}
	stowe := orb.Point{-72.69, 44.46}
	killington := orb.Point{-72.82, 43.62}

	// Create layers
	layers := []output.Layer{
		{
			Name: "border",
			Features: geometry.FeatureCollection{
				geometry.NewFeature(vermontBorder),
			},
			Style: output.Style{
				Stroke:      "#333333",
				StrokeWidth: 2.0,
				Fill:        "#f5f5dc", // beige
				Opacity:     1.0,
			},
			Order: 1,
		},
		{
			Name: "lake",
			Features: geometry.FeatureCollection{
				geometry.NewFeature(lakeChamplain),
			},
			Style: output.Style{
				Stroke:      "#1976d2",
				StrokeWidth: 1.5,
				Fill:        "#bbdefb",
				Opacity:     1.0,
			},
			Order: 2,
		},
		{
			Name: "towns",
			Features: geometry.FeatureCollection{
				geometry.NewFeature(burlington),
				geometry.NewFeature(montpelier),
				geometry.NewFeature(stowe),
				geometry.NewFeature(killington),
			},
			Style: output.Style{
				Stroke:      "#c62828",
				StrokeWidth: 3.0,
				Fill:        "#c62828",
				Opacity:     1.0,
			},
			Order: 3,
		},
	}

	// Render SVG
	svg := writer.Render(layers)

	// Write to file
	outputPath := "output/vermont_test.svg"
	os.MkdirAll("output", 0755)

	err := os.WriteFile(outputPath, []byte(svg), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ SVG generated: %s\n", outputPath)
	fmt.Printf("  Page size: 12x18 inches\n")
	fmt.Printf("  Layers: %d\n", len(layers))
	fmt.Printf("  Features: border, lake, 4 towns\n")
}
