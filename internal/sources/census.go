package sources

import (
	"fmt"
	"strings"
)

// StateFIPS maps state abbreviations to FIPS codes.
var StateFIPS = map[string]string{
	"al": "01", "ak": "02", "az": "04", "ar": "05", "ca": "06",
	"co": "08", "ct": "09", "de": "10", "fl": "12", "ga": "13",
	"hi": "15", "id": "16", "il": "17", "in": "18", "ia": "19",
	"ks": "20", "ky": "21", "la": "22", "me": "23", "md": "24",
	"ma": "25", "mi": "26", "mn": "27", "ms": "28", "mo": "29",
	"mt": "30", "ne": "31", "nv": "32", "nh": "33", "nj": "34",
	"nm": "35", "ny": "36", "nc": "37", "nd": "38", "oh": "39",
	"ok": "40", "or": "41", "pa": "42", "ri": "44", "sc": "45",
	"sd": "46", "tn": "47", "tx": "48", "ut": "49", "vt": "50",
	"va": "51", "wa": "53", "wv": "54", "wi": "55", "wy": "56",
	"dc": "11", "pr": "72",
}

// TIGERType describes a TIGER/Line data type.
type TIGERType struct {
	Folder    string // Folder name in TIGER URL
	PerCounty bool   // Whether data is split by county
	National  bool   // Whether it's a national file filtered by state
}

// TIGERTypes maps type names to their configuration.
var TIGERTypes = map[string]TIGERType{
	"state":       {Folder: "STATE", PerCounty: false, National: true},
	"county":      {Folder: "COUNTY", PerCounty: false, National: true},
	"cousub":      {Folder: "COUSUB", PerCounty: false, National: false},
	"place":       {Folder: "PLACE", PerCounty: false, National: false},
	"areawater":   {Folder: "AREAWATER", PerCounty: true, National: false},
	"linearwater": {Folder: "LINEARWATER", PerCounty: true, National: false},
	"roads":       {Folder: "ROADS", PerCounty: true, National: false},
	"prisecroads": {Folder: "PRISECROADS", PerCounty: false, National: false},
}

// StateCounties maps state abbreviations to their county FIPS codes.
var StateCounties = map[string][]string{
	"vt": {"001", "003", "005", "007", "009", "011", "013", "015", "017", "019", "021", "023", "025", "027"},
	"nh": {"001", "003", "005", "007", "009", "011", "013", "015", "017", "019"},
	"me": {"001", "003", "005", "007", "009", "011", "013", "015", "017", "019", "021", "023", "025", "027", "029", "031"},
	"ma": {"001", "003", "005", "007", "009", "011", "013", "015", "017", "019", "021", "023", "025", "027"},
	"ny": {"001", "003", "005", "007", "009", "011", "013", "015", "017", "019", "021", "023", "025", "027", "029", "031", "033", "035", "037", "039", "041", "043", "045", "047", "049", "051", "053", "055", "057", "059", "061", "063", "065", "067", "069", "071", "073", "075", "077", "079", "081", "083", "085", "087", "089", "091", "093", "095", "097", "099", "101", "103", "105", "107", "109", "111", "113", "115", "117", "119", "121", "123"},
}

// CensusURI represents a parsed census:// URI.
type CensusURI struct {
	Year      string
	State     string
	Type      string
	StateFIPS string
	TypeInfo  TIGERType
	URLs      []string
}

// ParseCensusURI parses a census URI.
// Supports formats:
//   - census://state/VT (shorthand, uses current year)
//   - census://areawater/VT
//   - census:tiger/2023/vt/cousub (full format)
func ParseCensusURI(uri string) (*CensusURI, error) {
	// Strip scheme
	var path string
	if strings.HasPrefix(uri, "census://") {
		path = uri[9:]
	} else if strings.HasPrefix(uri, "census:") {
		path = uri[7:]
	} else {
		return nil, fmt.Errorf("not a census URI: %s", uri)
	}

	// Try shorthand format: type/state (e.g., "state/VT" or "areawater/VT")
	parts := strings.Split(path, "/")

	var year, state, layerType string

	if len(parts) == 2 {
		// Shorthand: type/state
		layerType = strings.ToLower(parts[0])
		state = strings.ToLower(parts[1])
		year = "2023" // Default to 2023
	} else if len(parts) == 4 && parts[0] == "tiger" {
		// Full format: tiger/year/state/type
		year = parts[1]
		state = strings.ToLower(parts[2])
		layerType = strings.ToLower(parts[3])
	} else {
		return nil, fmt.Errorf("invalid census URI format: %s (expected census://type/state or census:tiger/year/state/type)", uri)
	}

	// Validate state
	fips, ok := StateFIPS[state]
	if !ok {
		return nil, fmt.Errorf("unknown state: %s", state)
	}

	// Validate type
	typeInfo, ok := TIGERTypes[layerType]
	if !ok {
		validTypes := make([]string, 0, len(TIGERTypes))
		for t := range TIGERTypes {
			validTypes = append(validTypes, t)
		}
		return nil, fmt.Errorf("unknown TIGER type: %s (valid: %s)", layerType, strings.Join(validTypes, ", "))
	}

	// Build URLs
	var urls []string
	baseURL := fmt.Sprintf("https://www2.census.gov/geo/tiger/TIGER%s/%s", year, typeInfo.Folder)

	if typeInfo.National {
		// National file (e.g., state, county)
		url := fmt.Sprintf("%s/tl_%s_us_%s.zip", baseURL, year, layerType)
		urls = append(urls, url)
	} else if typeInfo.PerCounty {
		// Per-county files
		counties := StateCounties[state]
		if len(counties) == 0 {
			// If we don't have county list, use a generic approach
			return nil, fmt.Errorf("no county list for state %s (per-county data requires county list)", state)
		}
		for _, county := range counties {
			fullFIPS := fips + county
			url := fmt.Sprintf("%s/tl_%s_%s_%s.zip", baseURL, year, fullFIPS, layerType)
			urls = append(urls, url)
		}
	} else {
		// Per-state file
		url := fmt.Sprintf("%s/tl_%s_%s_%s.zip", baseURL, year, fips, layerType)
		urls = append(urls, url)
	}

	return &CensusURI{
		Year:      year,
		State:     state,
		Type:      layerType,
		StateFIPS: fips,
		TypeInfo:  typeInfo,
		URLs:      urls,
	}, nil
}

// PrimaryURL returns the first URL (for display purposes).
func (c *CensusURI) PrimaryURL() string {
	if len(c.URLs) > 0 {
		return c.URLs[0]
	}
	return ""
}

// IsMultiFile returns true if this URI requires downloading multiple files.
func (c *CensusURI) IsMultiFile() bool {
	return len(c.URLs) > 1
}

// CacheKey returns a unique key for caching this URI's data.
func (c *CensusURI) CacheKey() string {
	return fmt.Sprintf("census/%s/%s/%s", c.Year, c.State, c.Type)
}
