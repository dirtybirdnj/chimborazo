package geometry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/engelsjk/polygol"
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

// FilterCollection filters features by a property condition.
// Supported operators: "=", "!=", ">", "<", ">=", "<=", "contains", "starts_with"
func FilterCollection(fc FeatureCollection, field, operator, value string) FeatureCollection {
	var result FeatureCollection
	for _, f := range fc {
		if matchesFilter(f.Properties, field, operator, value) {
			result = append(result, f)
		}
	}
	return result
}

// matchesFilter checks if a feature's properties match the filter condition.
func matchesFilter(props map[string]interface{}, field, operator, value string) bool {
	propVal, ok := props[field]
	if !ok {
		return operator == "!=" // Missing field only matches "!="
	}

	propStr := fmt.Sprintf("%v", propVal)

	switch operator {
	case "=", "==":
		return propStr == value
	case "!=":
		return propStr != value
	case "contains":
		return strings.Contains(propStr, value)
	case "starts_with":
		return strings.HasPrefix(propStr, value)
	case "ends_with":
		return strings.HasSuffix(propStr, value)
	case ">", "<", ">=", "<=":
		return compareNumeric(propStr, operator, value)
	default:
		return propStr == value // Default to equality
	}
}

// compareNumeric compares numeric values.
func compareNumeric(propStr, operator, value string) bool {
	var propNum, valNum float64
	if _, err := fmt.Sscanf(propStr, "%f", &propNum); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(value, "%f", &valNum); err != nil {
		return false
	}

	switch operator {
	case ">":
		return propNum > valNum
	case "<":
		return propNum < valNum
	case ">=":
		return propNum >= valNum
	case "<=":
		return propNum <= valNum
	default:
		return false
	}
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

// orbToPolygol converts an orb.Geometry to polygol.Geom format.
// polygol.Geom is [][][][]float64 representing a MultiPolygon.
func orbToPolygol(g orb.Geometry) polygol.Geom {
	switch v := g.(type) {
	case orb.Polygon:
		return polygonToPolygol(v)
	case orb.MultiPolygon:
		var result polygol.Geom
		for _, poly := range v {
			result = append(result, polygonToPolygol(poly)...)
		}
		return result
	case orb.Ring:
		// Treat ring as a simple polygon
		return polygonToPolygol(orb.Polygon{v})
	default:
		return nil
	}
}

// polygonToPolygol converts an orb.Polygon to polygol.Geom.
func polygonToPolygol(poly orb.Polygon) polygol.Geom {
	if len(poly) == 0 {
		return nil
	}

	// Each polygon in polygol is [][][]float64 (array of rings)
	rings := make([][][]float64, len(poly))
	for i, ring := range poly {
		rings[i] = make([][]float64, len(ring))
		for j, pt := range ring {
			rings[i][j] = []float64{pt[0], pt[1]}
		}
	}

	// Return as a single-polygon Geom
	return polygol.Geom{rings}
}

// polygolToOrb converts a polygol.Geom back to orb.Geometry.
func polygolToOrb(g polygol.Geom) orb.Geometry {
	if len(g) == 0 {
		return nil
	}

	if len(g) == 1 {
		// Single polygon
		return polygolPolygonToOrb(g[0])
	}

	// MultiPolygon
	result := make(orb.MultiPolygon, len(g))
	for i, polyData := range g {
		result[i] = polygolPolygonToOrb(polyData)
	}
	return result
}

// polygolPolygonToOrb converts a single polygol polygon ([][][]float64) to orb.Polygon.
func polygolPolygonToOrb(polyData [][][]float64) orb.Polygon {
	if len(polyData) == 0 {
		return nil
	}

	poly := make(orb.Polygon, len(polyData))
	for i, ringData := range polyData {
		ring := make(orb.Ring, len(ringData))
		for j, pt := range ringData {
			if len(pt) >= 2 {
				ring[j] = orb.Point{pt[0], pt[1]}
			}
		}
		poly[i] = ring
	}
	return poly
}

// Subtract removes the clip geometry from the subject geometry.
// This is a boolean difference operation (cookie cutter).
// Uses the polygol library for pure-Go polygon boolean operations.
func Subtract(subject, clipGeom orb.Geometry) (orb.Geometry, error) {
	subjectPG := orbToPolygol(subject)
	if subjectPG == nil {
		return subject, nil // Non-polygon geometry, return unchanged
	}

	clipPG := orbToPolygol(clipGeom)
	if clipPG == nil {
		return subject, nil // Nothing to subtract
	}

	result, err := polygol.Difference(subjectPG, clipPG)
	if err != nil {
		return nil, err
	}

	return polygolToOrb(result), nil
}

// SubtractAll removes multiple clip geometries from the subject.
func SubtractAll(subject orb.Geometry, clips ...orb.Geometry) (orb.Geometry, error) {
	result := subject
	for _, c := range clips {
		var err error
		result, err = Subtract(result, c)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, nil // Subject completely consumed
		}
	}
	return result, nil
}

// SubtractFeature subtracts a clip geometry from a feature.
func SubtractFeature(f *Feature, clipGeom orb.Geometry) (*Feature, error) {
	result, err := Subtract(f.Geometry, clipGeom)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &Feature{
		Geometry:   result,
		Properties: f.Properties,
	}, nil
}

// SubtractCollection subtracts all geometries from a clip collection from each feature
// in the subject collection. This is the "cut water from land" operation.
func SubtractCollection(subjects, clips FeatureCollection) (FeatureCollection, error) {
	if len(clips) == 0 {
		return subjects, nil
	}

	// Merge all clip geometries into one for efficiency
	clipGeoms := make([]orb.Geometry, len(clips))
	for i, c := range clips {
		clipGeoms[i] = c.Geometry
	}
	mergedClip := Merge(clipGeoms...)

	var result FeatureCollection
	for _, f := range subjects {
		subtracted, err := SubtractFeature(f, mergedClip)
		if err != nil {
			return nil, err
		}
		if subtracted != nil {
			result = append(result, subtracted)
		}
	}
	return result, nil
}

// Union combines multiple geometries into a single geometry.
// Uses polygol for polygon boolean union operations.
func Union(geoms ...orb.Geometry) (orb.Geometry, error) {
	if len(geoms) == 0 {
		return nil, nil
	}
	if len(geoms) == 1 {
		return geoms[0], nil
	}

	// Convert first geometry
	result := orbToPolygol(geoms[0])
	if result == nil {
		// If first geometry isn't a polygon, just use Merge
		return Merge(geoms...), nil
	}

	// Union remaining geometries
	for i := 1; i < len(geoms); i++ {
		pg := orbToPolygol(geoms[i])
		if pg == nil {
			continue
		}

		var err error
		result, err = polygol.Union(result, pg)
		if err != nil {
			return nil, err
		}
	}

	return polygolToOrb(result), nil
}

// Dissolve groups features by a property value and unions each group.
// Returns a FeatureCollection with one feature per unique property value.
func Dissolve(fc FeatureCollection, field string) (FeatureCollection, error) {
	if len(fc) == 0 {
		return fc, nil
	}

	// Group features by field value
	groups := make(map[string][]*Feature)
	for _, f := range fc {
		key := ""
		if val, ok := f.Properties[field]; ok {
			key = toString(val)
		}
		groups[key] = append(groups[key], f)
	}

	// Union each group
	var result FeatureCollection
	for key, features := range groups {
		if len(features) == 1 {
			// Single feature, just copy it
			result = append(result, features[0])
			continue
		}

		// Collect geometries for union
		geoms := make([]orb.Geometry, len(features))
		for i, f := range features {
			geoms[i] = f.Geometry
		}

		// Union all geometries in this group
		unioned, err := Union(geoms...)
		if err != nil {
			// If union fails, fall back to merge
			unioned = Merge(geoms...)
		}

		// Create new feature with unioned geometry
		// Copy properties from first feature, update the dissolve field
		props := make(map[string]interface{})
		for k, v := range features[0].Properties {
			props[k] = v
		}
		props[field] = key
		props["_dissolved_count"] = len(features)

		result = append(result, &Feature{
			Geometry:   unioned,
			Properties: props,
		})
	}

	return result, nil
}

// toString converts an interface{} to string for grouping.
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
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
