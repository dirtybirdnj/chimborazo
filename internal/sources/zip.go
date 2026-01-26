package sources

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts a ZIP file to the specified directory.
// Returns the path to the extracted directory.
func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// Sanitize the path to prevent zip slip attacks
		destPath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("creating parent directory: %w", err)
		}

		// Extract file
		if err := extractFile(f, destPath); err != nil {
			return fmt.Errorf("extracting %s: %w", f.Name, err)
		}
	}

	return nil
}

func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// FindShapefile finds the .shp file in a directory.
// Returns the path to the shapefile or an error if not found.
func FindShapefile(dir string) (string, error) {
	var shpPath string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".shp") {
			shpPath = path
			return filepath.SkipAll // Stop after finding first .shp
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("searching for shapefile: %w", err)
	}

	if shpPath == "" {
		return "", fmt.Errorf("no shapefile (.shp) found in %s", dir)
	}

	return shpPath, nil
}

// FindGeoJSON finds a .geojson or .json file in a directory.
func FindGeoJSON(dir string) (string, error) {
	var jsonPath string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			lower := strings.ToLower(path)
			if strings.HasSuffix(lower, ".geojson") || strings.HasSuffix(lower, ".json") {
				jsonPath = path
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("searching for geojson: %w", err)
	}

	if jsonPath == "" {
		return "", fmt.Errorf("no GeoJSON file found in %s", dir)
	}

	return jsonPath, nil
}
