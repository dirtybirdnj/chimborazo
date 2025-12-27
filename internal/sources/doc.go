/*
Package sources provides data acquisition from authoritative geospatial sources.

# The Acquisition Stage

I learned in Cumaná what the surveyor Thoreau knew: one must go to the source.
When I arrived in Venezuela in July 1799, local tales abounded of bottomless
lakes, of rivers that connected impossible watersheds, of mountains taller
than any in Europe. I could have repeated these tales. Instead, I measured.
I sounded the lakes. I traced the rivers. I climbed the mountains with
barometer in hand.

Thoreau, surveying Walden Pond fifty years later, put it perfectly: "It is
remarkable how long men will believe in the bottomlessness of a pond without
taking the trouble to sound it." He made over a hundred soundings through
the winter ice, because he would not trust hearsay.

This package embodies that principle. Every coordinate, every boundary, every
water feature comes directly from the agencies that surveyed them—not from
speculation, not from outdated caches, not from third-party aggregators.

# Supported URI Schemes

The fetcher resolves URIs with the following schemes:

	census://  - U.S. Census Bureau TIGER/Line shapefiles
	           Example: census://water/VT (Vermont water features)
	           Example: census://county/VT (Vermont counties)

	nhn://     - Canadian National Hydro Network
	           Example: nhn://02OJ (watershed 02OJ)

	file://    - Local filesystem paths
	           Example: file:///path/to/data.geojson

	http://    - Remote URLs (GeoJSON or shapefile)
	https://   Example: https://example.com/boundaries.geojson

# Caching

I could not carry sixty thousand specimens back to Paris without organization.
Everything was catalogued, preserved, referenced by collection number. Years
later, I could locate any specimen instantly.

Downloaded files are cached in ~/.cache/chimborazo/ by default. The cache
structure mirrors the URI path:

	~/.cache/chimborazo/
	├── census/
	│   ├── water/
	│   │   └── VT/
	│   └── county/
	│       └── VT/
	└── nhn/
	    └── 02OJ/

Subsequent requests for the same URI return the cached file without network
access. Cache entries include metadata: fetch time, size, checksum.

# Usage

	// Create a fetcher with default cache location
	fetcher, err := sources.NewFetcher()

	// Or specify cache directory
	fetcher, err := sources.NewFetcher(
	    sources.WithCacheDir("/custom/cache/path"),
	    sources.WithTimeout(60 * time.Second),
	)

	// Fetch a single source
	result, err := fetcher.Fetch("census://water/VT")
	if err != nil {
	    // Network error, invalid URI, etc.
	}
	// result.Path contains local file path
	// result.Fresh indicates whether newly downloaded

	// Fetch multiple sources concurrently
	results := fetcher.FetchAll([]string{
	    "census://water/VT",
	    "census://county/VT",
	    "census://state/VT",
	})

# Types

The primary types are:

  - [Fetcher]: Downloads and caches remote data files
  - [FetchResult]: The outcome of a fetch operation (path, freshness, error)
  - [CacheEntry]: Metadata about a cached file
  - [FetcherOption]: Configuration options for the fetcher
*/
package sources
