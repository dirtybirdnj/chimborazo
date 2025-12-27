package sources

import (
	"fmt"
	"strings"

	"github.com/jonas-p/go-shp"
	"github.com/paulmach/orb"

	"github.com/dirtybirdnj/chimborazo/internal/geometry"
)

// ReadShapefile reads a shapefile and returns a FeatureCollection.
// The shpPath should point to the .shp file; companion files (.dbf, .shx, .prj)
// are loaded automatically from the same directory.
func ReadShapefile(shpPath string) (geometry.FeatureCollection, error) {
	// Open the shapefile (this also opens .dbf automatically if present)
	shape, err := shp.Open(shpPath)
	if err != nil {
		return nil, fmt.Errorf("opening shapefile: %w", err)
	}
	defer shape.Close()

	// Get field names for attribute reading
	fields := shape.Fields()

	var features geometry.FeatureCollection

	// Read all shapes
	for shape.Next() {
		n, shpShape := shape.Shape()

		// Convert shape to orb geometry
		geom := shapeToOrb(shpShape)
		if geom == nil {
			continue // Skip unsupported geometry types
		}

		// Read attributes
		props := make(map[string]interface{})
		for i, field := range fields {
			val := shape.ReadAttribute(n, i)
			props[field.String()] = val
		}

		features = append(features, &geometry.Feature{
			Geometry:   geom,
			Properties: props,
		})
	}

	if err := shape.Err(); err != nil {
		return nil, fmt.Errorf("reading shapefile: %w", err)
	}

	return features, nil
}

// shapeToOrb converts a go-shp shape to an orb.Geometry.
func shapeToOrb(s shp.Shape) orb.Geometry {
	if s == nil {
		return nil
	}

	switch v := s.(type) {
	case *shp.Point:
		return orb.Point{v.X, v.Y}

	case *shp.PointM:
		return orb.Point{v.X, v.Y}

	case *shp.PointZ:
		return orb.Point{v.X, v.Y}

	case *shp.PolyLine:
		return polyLineToOrb(v)

	case *shp.PolyLineM:
		return polyLinePartsToOrb(v.NumParts, v.Parts, v.Points)

	case *shp.PolyLineZ:
		return polyLinePartsToOrb(v.NumParts, v.Parts, v.Points)

	case *shp.Polygon:
		return polygonToOrb(v)

	case *shp.PolygonM:
		pl := (*shp.PolyLineZ)(v)
		return polygonPartsToOrb(pl.NumParts, pl.Parts, pl.Points)

	case *shp.PolygonZ:
		pl := (*shp.PolyLineZ)(v)
		return polygonPartsToOrb(pl.NumParts, pl.Parts, pl.Points)

	case *shp.MultiPoint:
		points := make(orb.MultiPoint, len(v.Points))
		for i, p := range v.Points {
			points[i] = orb.Point{p.X, p.Y}
		}
		return points

	case *shp.MultiPointM:
		points := make(orb.MultiPoint, len(v.Points))
		for i, p := range v.Points {
			points[i] = orb.Point{p.X, p.Y}
		}
		return points

	case *shp.MultiPointZ:
		points := make(orb.MultiPoint, len(v.Points))
		for i, p := range v.Points {
			points[i] = orb.Point{p.X, p.Y}
		}
		return points

	default:
		return nil
	}
}

func polyLineToOrb(pl *shp.PolyLine) orb.Geometry {
	return polyLinePartsToOrb(pl.NumParts, pl.Parts, pl.Points)
}

func polyLinePartsToOrb(numParts int32, parts []int32, points []shp.Point) orb.Geometry {
	if numParts == 0 || len(points) == 0 {
		return nil
	}

	if numParts == 1 {
		// Single line
		line := make(orb.LineString, len(points))
		for i, p := range points {
			line[i] = orb.Point{p.X, p.Y}
		}
		return line
	}

	// Multi-line
	lines := make(orb.MultiLineString, numParts)
	for i := int32(0); i < numParts; i++ {
		start := parts[i]
		var end int32
		if i < numParts-1 {
			end = parts[i+1]
		} else {
			end = int32(len(points))
		}

		line := make(orb.LineString, end-start)
		for j := start; j < end; j++ {
			line[j-start] = orb.Point{points[j].X, points[j].Y}
		}
		lines[i] = line
	}
	return lines
}

func polygonToOrb(pg *shp.Polygon) orb.Geometry {
	// Polygon is just a type alias for PolyLine in go-shp
	pl := (*shp.PolyLine)(pg)
	return polygonPartsToOrb(pl.NumParts, pl.Parts, pl.Points)
}

func polygonPartsToOrb(numParts int32, parts []int32, points []shp.Point) orb.Geometry {
	if numParts == 0 || len(points) == 0 {
		return nil
	}

	// Extract rings
	rings := make([]orb.Ring, numParts)
	for i := int32(0); i < numParts; i++ {
		start := parts[i]
		var end int32
		if i < numParts-1 {
			end = parts[i+1]
		} else {
			end = int32(len(points))
		}

		ring := make(orb.Ring, end-start)
		for j := start; j < end; j++ {
			ring[j-start] = orb.Point{points[j].X, points[j].Y}
		}
		rings[i] = ring
	}

	// In shapefiles, the first ring is the exterior, subsequent rings are holes.
	// However, multiple exterior rings indicate a MultiPolygon.
	// We detect this by checking ring orientation (CCW = exterior, CW = hole in shapefile spec).

	// For simplicity, if there's only one ring, return a simple Polygon.
	// If there are multiple rings, assume first is exterior and rest are holes.
	// TODO: Proper multi-polygon detection based on ring orientation.

	if len(rings) == 1 {
		return orb.Polygon{rings[0]}
	}

	// Assume single polygon with holes
	// First ring is exterior, rest are holes
	poly := make(orb.Polygon, len(rings))
	for i, ring := range rings {
		poly[i] = ring
	}
	return poly
}

// ReadShapefileFiltered reads a shapefile and filters features by state FIPS.
// Used for national files that need to be filtered to a specific state.
func ReadShapefileFiltered(shpPath, stateFIPS string) (geometry.FeatureCollection, error) {
	features, err := ReadShapefile(shpPath)
	if err != nil {
		return nil, err
	}

	if stateFIPS == "" {
		return features, nil
	}

	// Filter by STATEFP or STATEFP20 or GEOID prefix
	var filtered geometry.FeatureCollection
	for _, f := range features {
		if matchesStateFIPS(f.Properties, stateFIPS) {
			filtered = append(filtered, f)
		}
	}

	return filtered, nil
}

func matchesStateFIPS(props map[string]interface{}, fips string) bool {
	// Check STATEFP
	if v, ok := props["STATEFP"]; ok {
		if str, ok := v.(string); ok && str == fips {
			return true
		}
	}

	// Check STATEFP20
	if v, ok := props["STATEFP20"]; ok {
		if str, ok := v.(string); ok && str == fips {
			return true
		}
	}

	// Check GEOID prefix
	if v, ok := props["GEOID"]; ok {
		if str, ok := v.(string); ok && strings.HasPrefix(str, fips) {
			return true
		}
	}

	return false
}
