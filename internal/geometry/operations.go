package geometry

import (
	"errors"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/simplify"
)

// ErrNotImplemented is returned for operations not yet implemented.
var ErrNotImplemented = errors.New("operation not implemented")

// Clip keeps only the portion of geometry within the given bounds.
// Uses the Sutherland-Hodgman algorithm via orb/clip.
func Clip(geom orb.Geometry, bounds orb.Bound) orb.Geometry {
	return clip.Geometry(bounds, geom)
}

// ClipFeature clips a feature's geometry to the given bounds.
func ClipFeature(f *Feature, bounds orb.Bound) *Feature {
	clipped := Clip(f.Geometry, bounds)
	if clipped == nil {
		return nil
	}
	return &Feature{
		Geometry:   clipped,
		Properties: f.Properties,
	}
}

// ClipCollection clips all features in a collection to the given bounds.
// Features that fall entirely outside the bounds are excluded.
func ClipCollection(fc FeatureCollection, bounds orb.Bound) FeatureCollection {
	var result FeatureCollection
	for _, f := range fc {
		if clipped := ClipFeature(f, bounds); clipped != nil {
			result = append(result, clipped)
		}
	}
	return result
}

// Simplify reduces the complexity of a geometry using Douglas-Peucker algorithm.
// Tolerance is in the same units as the geometry coordinates.
func Simplify(geom orb.Geometry, tolerance float64) orb.Geometry {
	return simplify.DouglasPeucker(tolerance).Simplify(geom)
}

// SimplifyFeature simplifies a feature's geometry.
func SimplifyFeature(f *Feature, tolerance float64) *Feature {
	return &Feature{
		Geometry:   Simplify(f.Geometry, tolerance),
		Properties: f.Properties,
	}
}

// SimplifyCollection simplifies all features in a collection.
func SimplifyCollection(fc FeatureCollection, tolerance float64) FeatureCollection {
	result := make(FeatureCollection, len(fc))
	for i, f := range fc {
		result[i] = SimplifyFeature(f, tolerance)
	}
	return result
}

// Merge combines multiple geometries into a single geometry collection.
// For polygons, this creates a MultiPolygon. For lines, a MultiLineString.
func Merge(geometries ...orb.Geometry) orb.Geometry {
	if len(geometries) == 0 {
		return nil
	}
	if len(geometries) == 1 {
		return geometries[0]
	}

	// Try to create typed multi-geometries
	var polygons []orb.Polygon
	var lines []orb.LineString
	var points []orb.Point

	for _, g := range geometries {
		switch v := g.(type) {
		case orb.Polygon:
			polygons = append(polygons, v)
		case orb.MultiPolygon:
			polygons = append(polygons, v...)
		case orb.LineString:
			lines = append(lines, v)
		case orb.MultiLineString:
			lines = append(lines, v...)
		case orb.Point:
			points = append(points, v)
		case orb.MultiPoint:
			points = append(points, v...)
		}
	}

	// Return the appropriate multi-type
	if len(polygons) > 0 && len(lines) == 0 && len(points) == 0 {
		return orb.MultiPolygon(polygons)
	}
	if len(lines) > 0 && len(polygons) == 0 && len(points) == 0 {
		return orb.MultiLineString(lines)
	}
	if len(points) > 0 && len(polygons) == 0 && len(lines) == 0 {
		return orb.MultiPoint(points)
	}

	// Mixed types - return as collection (orb doesn't have GeometryCollection,
	// so we'll just return the first geometry for now)
	// TODO: Handle mixed geometry types properly
	return geometries[0]
}

// MergeFeatures combines multiple features into a single feature with merged geometry.
// Properties from the first feature are preserved.
func MergeFeatures(features ...*Feature) *Feature {
	if len(features) == 0 {
		return nil
	}

	geoms := make([]orb.Geometry, len(features))
	for i, f := range features {
		geoms[i] = f.Geometry
	}

	return &Feature{
		Geometry:   Merge(geoms...),
		Properties: features[0].Properties,
	}
}

// Subtract removes the clip geometry from the subject geometry.
// This is a boolean difference operation (cookie cutter).
//
// NOTE: This operation requires GEOS or a pure-Go polygon clipping library.
// Currently returns an error as it's not yet implemented.
//
// Candidate libraries:
// - github.com/ctessum/geom (pure Go, has Difference)
// - github.com/twpayne/go-geom (may need CGO)
// - github.com/engelsjk/polygol (pure Go polygon boolean ops)
func Subtract(subject, clip orb.Geometry) (orb.Geometry, error) {
	// TODO: Implement using a polygon boolean library
	// For MVP, we might skip subtract and use layer ordering instead
	return nil, ErrNotImplemented
}

// Buffer expands (positive distance) or contracts (negative distance) a geometry.
//
// NOTE: This operation requires GEOS or significant pure-Go implementation.
// Currently returns an error as it's not yet implemented.
//
// Candidate libraries:
// - github.com/twpayne/go-geom with GEOS
// - Custom implementation for simple cases
func Buffer(geom orb.Geometry, distance float64) (orb.Geometry, error) {
	// TODO: Implement buffering
	// For MVP, we might skip buffer operations
	return nil, ErrNotImplemented
}

// IdentifyIslands finds enclosed regions (holes that are actually islands)
// within a polygon.
//
// NOTE: Complex operation, not yet implemented.
func IdentifyIslands(poly orb.Polygon, minArea float64) ([]orb.Polygon, error) {
	// TODO: Implement island identification
	// This involves analyzing polygon rings and their relationships
	return nil, ErrNotImplemented
}
