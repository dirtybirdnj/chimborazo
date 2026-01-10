// Package sources provides Canadian geospatial data from open data sources.
// Supported sources:
//   - CanVec (Natural Resources Canada) - Hydro at 1:1,000,000 scale
//   - NHN (National Hydro Network) - Detailed hydro by watershed workunit
//
// License: Open Government Licence - Canada
package sources

import (
	"fmt"
	"strings"
)

// CanVecURLs maps CanVec data types to their download URLs.
var CanVecURLs = map[string]string{
	// CanVec 1:1,000,000 scale - national hydro coverage (~150 MB)
	"hydro": "https://ftp.maps.canada.ca/pub/nrcan_rncan/vector/canvec/shp/Hydro/canvec_1M_CA_Hydro_shp.zip",
	// CanVec 1:1,000,000 scale - national transport coverage (~6.5 MB)
	"transport": "https://ftp.maps.canada.ca/pub/nrcan_rncan/vector/canvec/shp/Transport/canvec_1M_CA_Transport_shp.zip",
}

// NHNWorkunits maps workunit codes to their descriptions.
// Workunits are geographic areas, typically watersheds.
var NHNWorkunits = map[string]string{
	"02OJ000": "Richelieu River watershed (Lake Champlain to St. Lawrence)",
	"02OHA00": "Lake Champlain (Quebec portion, western)",
	"02OHB00": "Lake Champlain (Quebec portion, eastern - includes Missisquoi Bay)",
	"02OG000": "Yamaska River watershed",
	"02OA000": "St. Lawrence River (Montreal area)",
	"02OB000": "Ottawa River (lower)",
	"02OC000": "Ottawa River (middle)",
	"02OE000": "St-François River",
}

// CanadaURI represents a parsed canada:// URI.
type CanadaURI struct {
	Source   string // "canvec" or "nhn"
	Type     string // "hydro" for canvec, workunit code for nhn
	SubType  string // "rivers" for linear water, empty for area water
	URL      string
	CacheDir string
}

// ParseCanadaURI parses a canada URI.
// Supported formats:
//   - canada://canvec/hydro
//   - canada://canvec/roads (national roads from transport dataset)
//   - canada://canvec/rails (national railways from transport dataset)
//   - canada://nhn/02OJ000 (area water)
//   - canada://nhn/02OJ000/rivers (linear water)
func ParseCanadaURI(uri string) (*CanadaURI, error) {
	var path string
	if strings.HasPrefix(uri, "canada://") {
		path = uri[9:]
	} else if strings.HasPrefix(uri, "canada:") {
		path = uri[7:]
	} else {
		return nil, fmt.Errorf("not a canada URI: %s", uri)
	}

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid canada URI format: %s (expected canada://source/type)", uri)
	}

	source := strings.ToLower(parts[0])
	dataType := parts[1]
	subType := ""
	if len(parts) > 2 {
		subType = strings.ToLower(parts[2])
	}

	var url, cacheDir string

	switch source {
	case "canvec":
		// CanVec data
		dataType = strings.ToLower(dataType)

		// Map roads/rails to transport dataset
		dataset := dataType
		if dataType == "roads" || dataType == "rails" {
			dataset = "transport"
		}

		url, ok := CanVecURLs[dataset]
		if !ok {
			return nil, fmt.Errorf("unknown canvec type: %s (valid: hydro, roads, rails)", dataType)
		}
		return &CanadaURI{
			Source:   "canvec",
			Type:     dataType, // Keep original type (roads, rails) for shapefile selection
			URL:      url,
			CacheDir: fmt.Sprintf("canada/canvec/%s", dataset),
		}, nil

	case "nhn":
		// National Hydro Network - workunit based
		workunit := strings.ToUpper(dataType)
		if _, ok := NHNWorkunits[workunit]; !ok {
			// Allow unknown workunits but warn
			// Many more workunits exist than we've documented
		}

		// NHN URL pattern: /shp_en/{region}/nhn_rhn_{workunit}_shp_en.zip
		// Region is first 2 characters of workunit
		region := workunit[:2]
		workunitLower := strings.ToLower(workunit)
		url = fmt.Sprintf("https://ftp.maps.canada.ca/pub/nrcan_rncan/vector/geobase_nhn_rhn/shp_en/%s/nhn_rhn_%s_shp_en.zip",
			region, workunitLower)
		cacheDir = fmt.Sprintf("canada/nhn/%s", workunit)

		return &CanadaURI{
			Source:   "nhn",
			Type:     workunit,
			SubType:  subType, // "rivers" for linear, empty for area
			URL:      url,
			CacheDir: cacheDir,
		}, nil

	default:
		return nil, fmt.Errorf("unknown canada source: %s (valid: canvec, nhn)", source)
	}
}

// CacheKey returns a unique key for caching this URI's data.
func (c *CanadaURI) CacheKey() string {
	return c.CacheDir
}

// ShapefilePattern returns the glob pattern to find the relevant shapefile.
// NHN archives contain multiple shapefiles:
//   - *_HD_WATERBODY_2.shp - Area water (lakes, ponds)
//   - *_HD_SLWATER_1.shp - Single-line water (rivers, streams as lines)
// CanVec hydro contains:
//   - waterbody_2.shp - Area water
//   - water_linear_flow_1.shp - Linear water (rivers)
// CanVec transport contains:
//   - road_segment_1.shp - Roads (line features)
//   - track_segment_1.shp - Railways (line features)
func (c *CanadaURI) ShapefilePattern() string {
	if c.Source == "nhn" {
		if c.SubType == "rivers" {
			return "*SLWATER*.shp" // Linear water (uppercase in NHN)
		}
		return "*WATERBODY*.shp" // Area water (uppercase in NHN)
	}
	// CanVec - select appropriate shapefile based on type
	switch c.Type {
	case "roads":
		return "*road_segment*.shp"
	case "rails":
		return "*track_segment*.shp"
	case "hydro":
		if c.SubType == "rivers" {
			return "*water_linear_flow*.shp"
		}
		return "*waterbody*.shp"
	default:
		return "*waterbody*.shp"
	}
}
