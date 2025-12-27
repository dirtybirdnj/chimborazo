package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/dirtybirdnj/chimborazo/internal/geometry"
)

// MaxConcurrentDownloads limits parallel downloads to avoid overwhelming servers.
const MaxConcurrentDownloads = 4

// Resolver fetches and parses geospatial data from various URI schemes.
type Resolver struct {
	Fetcher  *Fetcher
	CacheDir string
	Verbose  bool
}

// NewResolver creates a new source resolver.
func NewResolver(cacheDir string) (*Resolver, error) {
	fetcher, err := NewFetcher(cacheDir)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		Fetcher:  fetcher,
		CacheDir: fetcher.CacheDir,
	}, nil
}

// Resolve fetches and parses data from a URI, returning a FeatureCollection.
// Supported schemes:
//   - census://type/state (e.g., census://state/VT, census://areawater/VT)
//   - file:///path/to/file.geojson
//   - file:///path/to/file.shp
//   - http(s)://example.com/data.geojson
//   - http(s)://example.com/data.zip
func (r *Resolver) Resolve(uri string) (geometry.FeatureCollection, error) {
	scheme := r.getScheme(uri)

	switch scheme {
	case "census":
		return r.resolveCensus(uri)
	case "quebec":
		return r.resolveQuebec(uri)
	case "canada":
		return r.resolveCanada(uri)
	case "file":
		return r.resolveFile(uri)
	case "http", "https":
		return r.resolveHTTP(uri)
	default:
		return nil, fmt.Errorf("unsupported URI scheme: %s", scheme)
	}
}

// getScheme extracts the scheme from a URI.
func (r *Resolver) getScheme(uri string) string {
	if strings.HasPrefix(uri, "census://") || strings.HasPrefix(uri, "census:") {
		return "census"
	}
	if strings.HasPrefix(uri, "quebec://") || strings.HasPrefix(uri, "quebec:") {
		return "quebec"
	}
	if strings.HasPrefix(uri, "canada://") || strings.HasPrefix(uri, "canada:") {
		return "canada"
	}
	if strings.HasPrefix(uri, "file://") || strings.HasPrefix(uri, "file:") {
		return "file"
	}
	if strings.HasPrefix(uri, "https://") {
		return "https"
	}
	if strings.HasPrefix(uri, "http://") {
		return "http"
	}
	// Assume file path if no scheme
	return "file"
}

// resolveCensus handles census:// URIs.
func (r *Resolver) resolveCensus(uri string) (geometry.FeatureCollection, error) {
	parsed, err := ParseCensusURI(uri)
	if err != nil {
		return nil, fmt.Errorf("parsing census URI: %w", err)
	}

	// Create cache directory for this source
	cacheKey := parsed.CacheKey()
	sourceDir := filepath.Join(r.CacheDir, cacheKey)

	// Check if we already have processed data
	mergedPath := filepath.Join(sourceDir, "merged.shp")
	if _, err := os.Stat(mergedPath); err == nil {
		r.logf("Using cached: %s", uri)
		return ReadShapefile(mergedPath)
	}

	filteredPath := filepath.Join(sourceDir, "filtered.shp")
	if _, err := os.Stat(filteredPath); err == nil {
		r.logf("Using cached: %s", uri)
		return ReadShapefile(filteredPath)
	}

	// Check for single shapefile
	existingShp, _ := FindShapefile(sourceDir)
	if existingShp != "" {
		r.logf("Using cached: %s", uri)
		if parsed.TypeInfo.National {
			return ReadShapefileFiltered(existingShp, parsed.StateFIPS)
		}
		return ReadShapefile(existingShp)
	}

	// Need to download
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return nil, fmt.Errorf("creating source directory: %w", err)
	}

	if parsed.TypeInfo.National {
		// Download national file and filter
		return r.downloadAndFilterNational(parsed, sourceDir)
	} else if parsed.IsMultiFile() {
		// Download multiple county files and merge
		return r.downloadAndMergeCounties(parsed, sourceDir)
	} else {
		// Download single file
		return r.downloadSingle(parsed.URLs[0], sourceDir, "")
	}
}

// downloadSingle downloads a single ZIP, extracts, and reads the shapefile.
func (r *Resolver) downloadSingle(url, destDir, stateFIPS string) (geometry.FeatureCollection, error) {
	r.logf("Downloading: %s", url)

	result, err := r.Fetcher.Fetch(url)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}

	// Extract ZIP
	if err := ExtractZip(result.Path, destDir); err != nil {
		return nil, fmt.Errorf("extracting zip: %w", err)
	}

	// Find and read shapefile
	shpPath, err := FindShapefile(destDir)
	if err != nil {
		return nil, err
	}

	if stateFIPS != "" {
		return ReadShapefileFiltered(shpPath, stateFIPS)
	}
	return ReadShapefile(shpPath)
}

// downloadAndFilterNational downloads a national file and filters to state.
func (r *Resolver) downloadAndFilterNational(parsed *CensusURI, destDir string) (geometry.FeatureCollection, error) {
	r.logf("Downloading national file for %s (will filter to state %s)...", parsed.Type, parsed.State)

	features, err := r.downloadSingle(parsed.URLs[0], destDir, parsed.StateFIPS)
	if err != nil {
		return nil, err
	}

	r.logf("Filtered to %d features for state %s", len(features), parsed.State)

	// TODO: Save filtered shapefile for future caching
	// For now, we rely on the extracted files being cached

	return features, nil
}

// countyResult holds the result of a single county download.
type countyResult struct {
	index    int
	features geometry.FeatureCollection
	err      error
}

// downloadAndMergeCounties downloads multiple county files concurrently and merges them.
func (r *Resolver) downloadAndMergeCounties(parsed *CensusURI, destDir string) (geometry.FeatureCollection, error) {
	r.logf("Downloading %d county files for %s/%s (max %d concurrent)...",
		len(parsed.URLs), parsed.State, parsed.Type, MaxConcurrentDownloads)

	// Channel for work items
	jobs := make(chan int, len(parsed.URLs))
	results := make(chan countyResult, len(parsed.URLs))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < MaxConcurrentDownloads; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				url := parsed.URLs[i]
				countyDir := filepath.Join(destDir, fmt.Sprintf("county_%03d", i))

				if err := os.MkdirAll(countyDir, 0755); err != nil {
					results <- countyResult{index: i, err: err}
					continue
				}

				features, err := r.downloadSingle(url, countyDir, "")
				results <- countyResult{index: i, features: features, err: err}
			}
		}()
	}

	// Send all jobs
	for i := range parsed.URLs {
		jobs <- i
	}
	close(jobs)

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allFeatures geometry.FeatureCollection
	successCount := 0
	for result := range results {
		if result.err != nil {
			r.logf("Skipped county %d: %v", result.index+1, result.err)
			continue
		}
		allFeatures = append(allFeatures, result.features...)
		successCount++
	}

	if len(allFeatures) == 0 {
		return nil, fmt.Errorf("no data found for %s", parsed.CacheKey())
	}

	r.logf("Merged %d features from %d/%d counties", len(allFeatures), successCount, len(parsed.URLs))

	return allFeatures, nil
}

// resolveFile handles file:// URIs and plain file paths.
func (r *Resolver) resolveFile(uri string) (geometry.FeatureCollection, error) {
	path := uri
	if strings.HasPrefix(uri, "file://") {
		path = uri[7:]
	} else if strings.HasPrefix(uri, "file:") {
		path = uri[5:]
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("expanding home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Determine file type
	lower := strings.ToLower(path)

	if strings.HasSuffix(lower, ".geojson") || strings.HasSuffix(lower, ".json") {
		return r.readGeoJSON(path)
	}

	if strings.HasSuffix(lower, ".shp") {
		return ReadShapefile(path)
	}

	if strings.HasSuffix(lower, ".zip") {
		// Extract to temp directory and read
		tmpDir, err := os.MkdirTemp("", "chimborazo-")
		if err != nil {
			return nil, fmt.Errorf("creating temp directory: %w", err)
		}
		// Note: We don't clean up tmpDir; rely on OS temp cleanup

		if err := ExtractZip(path, tmpDir); err != nil {
			return nil, fmt.Errorf("extracting zip: %w", err)
		}

		// Try shapefile first, then geojson
		if shpPath, err := FindShapefile(tmpDir); err == nil {
			return ReadShapefile(shpPath)
		}
		if jsonPath, err := FindGeoJSON(tmpDir); err == nil {
			return r.readGeoJSON(jsonPath)
		}

		return nil, fmt.Errorf("no shapefile or geojson found in zip: %s", path)
	}

	return nil, fmt.Errorf("unsupported file type: %s", path)
}

// resolveHTTP handles http:// and https:// URIs.
func (r *Resolver) resolveHTTP(uri string) (geometry.FeatureCollection, error) {
	result, err := r.Fetcher.Fetch(uri)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", uri, err)
	}

	// Determine type from URL or content
	lower := strings.ToLower(uri)

	if strings.HasSuffix(lower, ".zip") || strings.Contains(lower, ".zip?") {
		// Extract and read
		tmpDir, err := os.MkdirTemp("", "chimborazo-")
		if err != nil {
			return nil, fmt.Errorf("creating temp directory: %w", err)
		}

		if err := ExtractZip(result.Path, tmpDir); err != nil {
			return nil, fmt.Errorf("extracting zip: %w", err)
		}

		if shpPath, err := FindShapefile(tmpDir); err == nil {
			return ReadShapefile(shpPath)
		}
		if jsonPath, err := FindGeoJSON(tmpDir); err == nil {
			return r.readGeoJSON(jsonPath)
		}

		return nil, fmt.Errorf("no shapefile or geojson found in zip: %s", uri)
	}

	// Assume GeoJSON
	return r.readGeoJSON(result.Path)
}

// readGeoJSON reads a GeoJSON file and returns a FeatureCollection.
func (r *Resolver) readGeoJSON(path string) (geometry.FeatureCollection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Try FeatureCollection first
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err == nil {
		return geoJSONToFeatures(fc), nil
	}

	// Try single Feature
	f, err := geojson.UnmarshalFeature(data)
	if err == nil {
		return geometry.FeatureCollection{geometry.FromGeoJSON(f)}, nil
	}

	// Try Geometry
	g, err := geojson.UnmarshalGeometry(data)
	if err == nil {
		return geometry.FeatureCollection{
			&geometry.Feature{
				Geometry:   g.Geometry(),
				Properties: make(map[string]interface{}),
			},
		}, nil
	}

	return nil, fmt.Errorf("parsing GeoJSON: could not parse as FeatureCollection, Feature, or Geometry")
}

// geoJSONToFeatures converts a geojson.FeatureCollection to our FeatureCollection.
func geoJSONToFeatures(fc *geojson.FeatureCollection) geometry.FeatureCollection {
	result := make(geometry.FeatureCollection, len(fc.Features))
	for i, f := range fc.Features {
		result[i] = geometry.FromGeoJSON(f)
	}
	return result
}

// GetBounds returns the bounding box of a resolved source.
func (r *Resolver) GetBounds(uri string) (orb.Bound, error) {
	features, err := r.Resolve(uri)
	if err != nil {
		return orb.Bound{}, err
	}

	return features.Bound(), nil
}

func (r *Resolver) logf(format string, args ...interface{}) {
	if r.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// resolveQuebec handles quebec:// URIs for Quebec administrative boundaries.
func (r *Resolver) resolveQuebec(uri string) (geometry.FeatureCollection, error) {
	parsed, err := ParseQuebecURI(uri)
	if err != nil {
		return nil, fmt.Errorf("parsing quebec URI: %w", err)
	}

	// Create cache directory for this source
	cacheKey := parsed.CacheKey()
	sourceDir := filepath.Join(r.CacheDir, cacheKey)

	// Check if we already have the shapefile - walk directory tree
	var matches []string
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".shp") && strings.Contains(filepath.Base(path), parsed.ShapefilePrefix) {
			matches = append(matches, path)
		}
		return nil
	})
	if len(matches) > 0 {
		r.logf("Using cached: %s", uri)
		return ReadShapefile(matches[0])
	}

	// Need to download
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return nil, fmt.Errorf("creating source directory: %w", err)
	}

	r.logf("Downloading: %s", parsed.URL)

	// Quebec files can be large (47-88 MB), use 5 minute timeout
	result, err := r.Fetcher.Fetch(parsed.URL, WithTimeout(5*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", parsed.URL, err)
	}

	// Extract ZIP
	if err := ExtractZip(result.Path, sourceDir); err != nil {
		return nil, fmt.Errorf("extracting zip: %w", err)
	}

	// Find the shapefile matching our layer (walk directory tree)
	matches = nil // reset from earlier check
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".shp") && strings.Contains(filepath.Base(path), parsed.ShapefilePrefix) {
			matches = append(matches, path)
		}
		return nil
	})
	if len(matches) == 0 {
		// Fall back to any shapefile
		shpPath, err := FindShapefile(sourceDir)
		if err != nil {
			return nil, err
		}
		return ReadShapefile(shpPath)
	}

	r.logf("Found shapefile: %s", matches[0])
	return ReadShapefile(matches[0])
}

// resolveCanada handles canada:// URIs for Canadian data (CanVec, NHN).
func (r *Resolver) resolveCanada(uri string) (geometry.FeatureCollection, error) {
	parsed, err := ParseCanadaURI(uri)
	if err != nil {
		return nil, fmt.Errorf("parsing canada URI: %w", err)
	}

	// Create cache directory for this source
	cacheKey := parsed.CacheKey()
	sourceDir := filepath.Join(r.CacheDir, cacheKey)

	// Check if we already have the shapefile - walk directory tree
	pattern := parsed.ShapefilePattern()
	var matches []string
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				matches = append(matches, path)
			}
		}
		return nil
	})
	if len(matches) > 0 {
		r.logf("Using cached: %s", uri)
		return ReadShapefile(matches[0])
	}

	// Need to download
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return nil, fmt.Errorf("creating source directory: %w", err)
	}

	r.logf("Downloading: %s", parsed.URL)

	// Canadian files can be large (up to 150 MB for CanVec), use 5 minute timeout
	result, err := r.Fetcher.Fetch(parsed.URL, WithTimeout(5*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", parsed.URL, err)
	}

	// Extract ZIP
	if err := ExtractZip(result.Path, sourceDir); err != nil {
		return nil, fmt.Errorf("extracting zip: %w", err)
	}

	// Find the shapefile matching our pattern (walk directory tree)
	matches = nil // reset from earlier check
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				matches = append(matches, path)
			}
		}
		return nil
	})
	if len(matches) == 0 {
		// Fall back to any shapefile
		shpPath, err := FindShapefile(sourceDir)
		if err != nil {
			return nil, err
		}
		return ReadShapefile(shpPath)
	}

	r.logf("Found shapefile: %s", matches[0])
	return ReadShapefile(matches[0])
}
