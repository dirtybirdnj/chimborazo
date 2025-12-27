/*
Package geometry provides geometric transformations for spatial data.

# The Transformation Stage

Between the acquisition of specimens and the publication of understanding lies
the hardest work: transformation. I sat for months in Quito, comparing
measurements, correlating observations, seeking patterns that would only
emerge through systematic comparison.

A single temperature reading tells you nothing. But temperature readings at
fifty altitudes, correlated with vegetation observations, with barometric
pressure, with latitude—suddenly you see the system. You see that plants
arrange themselves by temperature bands, not by arbitrary boundaries.

This is what I mean by "unity in diversity." The raw observations are diverse—
chaotic, even. The unity emerges only through transformation.

# The Operations

This package implements geometric operations that transform raw coordinates
into meaningful shapes:

	subtract   - Remove one geometry from another (water from land)
	clip       - Limit geometry to bounding box
	merge      - Combine all features into a single geometry
	dissolve   - Merge features that share an attribute value
	simplify   - Reduce vertex count while preserving shape
	buffer     - Expand or contract geometry boundaries
	clean      - Fix topology issues (self-intersections, slivers)

Each operation takes a FeatureCollection and returns a transformed
FeatureCollection. Operations are applied sequentially as specified
in the recipe.

# Subtract: Revealing Relationships

When I mapped the shores of Lake Valencia in Venezuela, I had to distinguish
land from water. The lake did not merely occupy space—it carved that space
from the surrounding terrain. The relationship between lake and shore became
visible only when I rendered them as figure and ground.

	subtract(land_polygons, water_polygons)
	// Land now has water-shaped holes

The subtract operation performs Boolean difference: for each feature in the
input, it removes the intersection with the target geometry. Features
entirely contained by the target are removed. Features with no intersection
are returned unchanged.

# Clip: Bounding the Investigation

My maps could not extend infinitely. Each study had bounds—the Orinoco basin,
the Quito highlands, the Mexican plateau. Features extending beyond those
bounds were clipped at the boundary.

	clip(features, bounds)
	// Only geometry within bounds remains

# Merge and Dissolve: Finding Unity

Administrative boundaries are often fragmented—census blocks, municipalities,
provinces split along arbitrary lines. The merge operation combines all
features into a single geometry. The dissolve operation groups features by
an attribute value and merges each group:

	merge(town_polygons)           // One unified polygon
	dissolve(blocks, "COUNTYFP")   // One polygon per county

# Simplify: Clarity Through Reduction

For publication, some detail must be sacrificed. A coastline with ten thousand
vertices becomes unmanageable. The simplify operation applies Douglas-Peucker
simplification, reducing vertex count while preserving essential shape:

	simplify(features, tolerance: 0.001)

Larger tolerance means more simplification. The trick is finding the tolerance
that removes noise without losing meaningful detail.

# Usage

	// Create an operation context
	ctx := &geometry.OperationContext{
	    Bounds:  [4]float64{-73.5, 42.7, -71.5, 45.1},
	    Sources: map[string]FeatureCollection{
	        "water": waterFeatures,
	    },
	}

	// Apply operations in sequence
	features := loadedFeatures
	for _, op := range layer.Operations {
	    var err error
	    features, err = geometry.Apply(features, op, ctx)
	    if err != nil {
	        // Operation failed
	    }
	}

# Types

The primary types are:

  - [Feature]: An orb.Geometry with properties and computed metadata
  - [FeatureCollection]: A slice of Features
  - [OperationContext]: Bounds and source references for operations
  - [GeometryOperator]: Interface implemented by each operation type

# Dependencies

This package uses github.com/paulmach/orb for geometry primitives:

  - orb.Polygon, orb.MultiPolygon for area features
  - orb.LineString, orb.MultiLineString for linear features
  - orb.Point for point features
  - orb/clip for clipping operations
  - orb/simplify for Douglas-Peucker simplification
*/
package geometry
