/*
Package output renders processed geometry to SVG and other formats.

# The Revelation Stage

The Naturgemälde was not merely a diagram. It was a revelation.

I drew Chimborazo in cross-section, showing the vegetation zones arranged by
altitude. At the base: tropical palms. Higher: oaks and shrubs of the temperate
zone. Higher still: alpine grasses. At the heights I reached: bare rock and
eternal snow. But I did not stop there. I added columns showing temperature,
barometric pressure, humidity, the boiling point of water at each altitude.
Everything on a single page.

A viewer could see at a glance what no table of numbers could convey: that
these phenomena are connected. That altitude, temperature, pressure, and life
form a single system. The diagram revealed what raw data could not.

Florence Kelley understood this principle when she mapped Chicago's 19th Ward
in 1895. She did not merely list wages and nationalities in tables. She
mapped them—color-coded by ethnicity, shaded by income level. "By looking at
just eight pages," she wrote, "the reader could easily see the clustering of
nationalities and the distribution of families' weekly wages."

A map is not decoration. A map is evidence.

# SVG Output

This package produces Scalable Vector Graphics—the format chosen for its
precision, portability, and compatibility with pen plotters and laser cutters.
Every geometry becomes an SVG path; every layer becomes an SVG group; every
style becomes explicit attributes.

	<svg width="12in" height="18in" viewBox="-73.5 42.7 2 2.4">
	  <g id="layer-boundary" stroke="#333" stroke-width="2">
	    <path d="M-73.2 44.5 L-72.8 44.6 L..." />
	  </g>
	  <g id="layer-water" stroke="#2196F3" fill="none">
	    <path d="M-72.5 43.8 L-72.4 43.9 L..." />
	  </g>
	</svg>

# Coordinate Transformation

Geographic coordinates (longitude/latitude in degrees) must be transformed
to page coordinates (inches or millimeters). This package handles:

  - Aspect ratio preservation
  - Latitude correction (longitude degrees are narrower at higher latitudes)
  - Page margin calculation
  - ViewBox generation for geographic bounds

# Layer Ordering

William Smith understood that strata always appear in the same order. In SVG,
the layer order determines what renders on top of what. Layers are sorted by
their Order field before rendering:

  - Higher Order values render first (at the bottom of the SVG)
  - Lower Order values render last (on top)

This ensures that roads appear above water, boundaries above backgrounds,
labels above everything.

# Style Application

Every style choice serves comprehension:

	Stroke width  - Visual hierarchy (state > county > town)
	Stroke color  - Semantic distinction (water blue, roads gray)
	Fill          - Area emphasis (or "none" for outline-only)
	Opacity       - Layering effects

Styles can be specified per-layer or per-feature (via color_map and fill_by).

# Plotter Optimization

For pen plotters, SVG output is optimized:

  - Stroke-only rendering (fill: none)
  - Path simplification to reduce plot time
  - Layer separation for multi-pen plotting
  - Optional plotter_fill conversion (colors to hatch patterns)

# Usage

	// Create an SVG document
	doc := output.NewSVGDocument(
	    output.WithPageSize("12x18"),
	    output.WithBounds(bounds),
	)

	// Add layers in order
	for _, layer := range sortedLayers {
	    doc.AddLayer(layer.Name, features, layer.Style)
	}

	// Render to file
	err := doc.WriteToFile("output/map.svg")

	// Or render to io.Writer
	err := doc.WriteTo(os.Stdout)

# Types

The primary types are:

  - [SVGDocument]: The document being constructed
  - [SVGLayer]: A group of paths with shared styling
  - [SVGPath]: A single path element with optional style overrides
  - [PageSize]: Named page dimensions (letter, tabloid, A4, custom)
*/
package output
