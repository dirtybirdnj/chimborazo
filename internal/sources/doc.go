// Package sources provides data acquisition for Chimborazo.
//
// It handles fetching data from various sources (Census TIGER, Natural Earth,
// OpenStreetMap, local files) and caching them locally for reuse.
//
// The primary type is Fetcher, which downloads and caches remote files:
//
//	fetcher, _ := NewFetcher("~/.cache/chimborazo")
//	result, _ := fetcher.Fetch("https://example.com/data.zip")
//	// result.Path contains local file path
//
// See specs/001-http-fetcher.md for implementation details.
package sources
