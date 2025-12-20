// Package geometry provides geometric types and operations for Chimborazo.
// It wraps github.com/paulmach/orb with additional functionality for
// cartographic processing.
package geometry

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// Feature wraps a geometry with properties, similar to GeoJSON Feature.
type Feature struct {
	Geometry   orb.Geometry           `json:"geometry"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// NewFeature creates a Feature from an orb geometry.
func NewFeature(g orb.Geometry) *Feature {
	return &Feature{
		Geometry:   g,
		Properties: make(map[string]interface{}),
	}
}

// FromGeoJSON converts a geojson.Feature to our Feature type.
func FromGeoJSON(f *geojson.Feature) *Feature {
	return &Feature{
		Geometry:   f.Geometry,
		Properties: f.Properties,
	}
}

// ToGeoJSON converts our Feature to a geojson.Feature.
func (f *Feature) ToGeoJSON() *geojson.Feature {
	gf := geojson.NewFeature(f.Geometry)
	gf.Properties = f.Properties
	return gf
}

// Bound returns the bounding box of the feature's geometry.
func (f *Feature) Bound() orb.Bound {
	return f.Geometry.Bound()
}

// FeatureCollection is a slice of features.
type FeatureCollection []*Feature

// Bound returns the bounding box containing all features.
func (fc FeatureCollection) Bound() orb.Bound {
	if len(fc) == 0 {
		return orb.Bound{}
	}
	bound := fc[0].Bound()
	for _, f := range fc[1:] {
		bound = bound.Union(f.Bound())
	}
	return bound
}

// ToGeoJSON converts the collection to a geojson.FeatureCollection.
func (fc FeatureCollection) ToGeoJSON() *geojson.FeatureCollection {
	gjfc := geojson.NewFeatureCollection()
	for _, f := range fc {
		gjfc.Append(f.ToGeoJSON())
	}
	return gjfc
}
