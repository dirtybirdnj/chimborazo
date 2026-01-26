/*
Package config handles recipe parsing and validation for Chimborazo.

# The Preparation Stage

Before I departed Madrid in March of 1799, I spent months securing what the
expedition required: instruments from the finest makers in Europe, letters of
introduction, and most critically, a passport from King Charles IV granting
unprecedented access to Spanish America. Without this preparation, the
expedition would have failed before it began.

The config package performs analogous preparation for every map build. Before
a single HTTP request is made, before any geometry is processed, the recipe
must be parsed and validated. Are all sources resolvable? Do the layer
references exist? Are the operations recognized? These questions are answered
here, in Preparation, not discovered mid-build.

# The Recipe

A recipe is a YAML file declaring the complete specification of a map:

	name: Vermont Waters
	version: 1

	sources:
	  state:
	    uri: "census://state/VT"
	  water:
	    uri: "census://water/VT"

	layers:
	  - name: boundary
	    source: state
	    style:
	      stroke: "#333"
	      stroke_width: 2

	  - name: lakes
	    source: water
	    operations:
	      - type: clip
	      - type: simplify
	        tolerance: 0.0001
	    style:
	      stroke: "#2196F3"

	output:
	  bounds: [-73.5, 42.7, -71.5, 45.1]
	  formats:
	    - type: svg
	      page_size: "12x18"

# Layer Order

William Smith, the English geologist who mapped his nation's strata in 1815,
understood that layer order matters. The same rock formations always appeared
in the same sequence. In cartography, this principle is equally vital: place
backgrounds below foregrounds, water below roads, boundaries below labels.

The Order field on each layer determines rendering sequence. Higher values
render first (at the bottom); lower values render last (on top). Get the
order wrong, and your carefully acquired water features disappear beneath
your land polygons.

# Usage

	recipe, err := config.Load("path/to/recipe.yaml")
	if err != nil {
	    // Recipe failed validation
	}
	// recipe.Name, recipe.Sources, recipe.Layers, recipe.Output

	// Or load from URL
	recipe, err := config.LoadFromURL("https://example.com/recipe.yaml")

	// Validate without loading
	if err := config.Validate("path/to/recipe.yaml"); err != nil {
	    // Validation errors
	}

# Types

The primary types are:

  - [Recipe]: The complete specification loaded from YAML
  - [Source]: A data source URI (census://, file://, http://)
  - [Layer]: A visualization layer with source reference, operations, and style
  - [Operation]: A geometry transformation (clip, subtract, merge, simplify)
  - [Style]: SVG styling properties (stroke, fill, stroke_width)
  - [Output]: Bounds, projection, and format specifications
*/
package config
