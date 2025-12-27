// Package sources provides Quebec administrative boundary data.
// Data from MERN (Ministère de l'Énergie et des Ressources naturelles du Québec).
// License: CC-BY 4.0
package sources

import (
	"fmt"
	"strings"
)

// QuebecURLs maps Quebec data types to their download URLs.
var QuebecURLs = map[string]string{
	// Administrative boundaries at 1:100,000 scale (47 MB)
	"sda_100k": "https://diffusion.mern.gouv.qc.ca/diffusion/RGQ/Vectoriel/Theme/Regional/SDA_100k/SHP/BDAT(adm)_SHP.zip",
	// Administrative boundaries at 1:20,000 scale (88 MB)
	"sda_20k": "https://diffusion.mern.gouv.qc.ca/Diffusion/RGQ/Vectoriel/Theme/Local/SDA_20k/SHP/SHP.zip",
}

// QuebecLayers maps layer names to shapefile prefixes within the archives.
var QuebecLayers = map[string]string{
	"municipalities": "munic_s", // Municipal boundaries
	"mrc":            "mrc_s",   // Regional county municipalities
	"regions":        "regio_s", // Administrative regions
	"metropolitan":   "comet_s", // Metropolitan communities
}

// QuebecURI represents a parsed quebec:// URI.
type QuebecURI struct {
	Layer           string
	Source          string // sda_100k or sda_20k
	ShapefilePrefix string
	URL             string
}

// ParseQuebecURI parses a quebec URI.
// Supported formats:
//   - quebec://municipalities
//   - quebec://mrc
//   - quebec://regions
//   - quebec://sda_100k
func ParseQuebecURI(uri string) (*QuebecURI, error) {
	var path string
	if strings.HasPrefix(uri, "quebec://") {
		path = uri[9:]
	} else if strings.HasPrefix(uri, "quebec:") {
		path = uri[7:]
	} else {
		return nil, fmt.Errorf("not a quebec URI: %s", uri)
	}

	layer := strings.ToLower(path)

	// Determine source and shapefile prefix
	var source, shapefilePrefix string

	switch layer {
	case "municipalities", "mrc", "regions", "metropolitan":
		source = "sda_100k"
		shapefilePrefix = QuebecLayers[layer]
	case "sda_100k", "sda_20k":
		source = layer
		shapefilePrefix = "munic_s" // Default to municipalities
	default:
		return nil, fmt.Errorf("unknown quebec layer: %s (valid: municipalities, mrc, regions, metropolitan, sda_100k, sda_20k)", layer)
	}

	url, ok := QuebecURLs[source]
	if !ok {
		return nil, fmt.Errorf("no URL for quebec source: %s", source)
	}

	return &QuebecURI{
		Layer:           layer,
		Source:          source,
		ShapefilePrefix: shapefilePrefix,
		URL:             url,
	}, nil
}

// CacheKey returns a unique key for caching this URI's data.
func (q *QuebecURI) CacheKey() string {
	return fmt.Sprintf("quebec/%s", q.Source)
}
