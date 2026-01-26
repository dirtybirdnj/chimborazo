package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRecipeBytes(t *testing.T) {
	yaml := `
name: test-map
output:
  path: output.svg
  width: 800
  height: 600
  units: px
bounds:
  source: census://state/VT
  padding: 0.1
layers:
  - name: states
    source: census://states
    style:
      fill: "#ccc"
      stroke: "#000"
`
	recipe, err := LoadRecipeBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadRecipeBytes() error = %v", err)
	}

	if recipe.Name != "test-map" {
		t.Errorf("Name = %q, want %q", recipe.Name, "test-map")
	}

	if recipe.Output.Path != "output.svg" {
		t.Errorf("Output.Path = %q, want %q", recipe.Output.Path, "output.svg")
	}

	if recipe.Output.Width != 800 {
		t.Errorf("Output.Width = %v, want 800", recipe.Output.Width)
	}

	if recipe.Output.Height != 600 {
		t.Errorf("Output.Height = %v, want 600", recipe.Output.Height)
	}

	if recipe.Bounds.Source != "census://state/VT" {
		t.Errorf("Bounds.Source = %q, want %q", recipe.Bounds.Source, "census://state/VT")
	}

	if len(recipe.Layers) != 1 {
		t.Fatalf("len(Layers) = %d, want 1", len(recipe.Layers))
	}

	if recipe.Layers[0].Name != "states" {
		t.Errorf("Layers[0].Name = %q, want %q", recipe.Layers[0].Name, "states")
	}

	if recipe.Layers[0].Style["fill"] != "#ccc" {
		t.Errorf("Layers[0].Style[fill] = %q, want %q", recipe.Layers[0].Style["fill"], "#ccc")
	}
}

func TestLoadRecipe(t *testing.T) {
	yaml := `
name: file-test
output:
  path: test.svg
  width: 100
  height: 100
layers:
  - name: base
    source: file://test.geojson
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "recipe.yaml")

	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("Writing temp file: %v", err)
	}

	recipe, err := LoadRecipe(path)
	if err != nil {
		t.Fatalf("LoadRecipe() error = %v", err)
	}

	if recipe.Name != "file-test" {
		t.Errorf("Name = %q, want %q", recipe.Name, "file-test")
	}
}

func TestLoadRecipe_NotFound(t *testing.T) {
	_, err := LoadRecipe("/nonexistent/path/recipe.yaml")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestLoadRecipeBytes_InvalidYAML(t *testing.T) {
	invalid := `
name: test
output:
  - this is wrong
  - should be object not array
`
	_, err := LoadRecipeBytes([]byte(invalid))
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}
}

func TestValidateRecipe(t *testing.T) {
	tests := []struct {
		name    string
		recipe  Recipe
		wantErr string
	}{
		{
			name:    "empty name",
			recipe:  Recipe{},
			wantErr: "name is required",
		},
		{
			name: "empty output path",
			recipe: Recipe{
				Name: "test",
			},
			wantErr: "output path is required",
		},
		{
			name: "zero width",
			recipe: Recipe{
				Name:   "test",
				Output: OutputConfig{Path: "out.svg", Width: 0, Height: 100},
			},
			wantErr: "width must be positive",
		},
		{
			name: "negative height",
			recipe: Recipe{
				Name:   "test",
				Output: OutputConfig{Path: "out.svg", Width: 100, Height: -50},
			},
			wantErr: "height must be positive",
		},
		{
			name: "no layers",
			recipe: Recipe{
				Name:   "test",
				Output: OutputConfig{Path: "out.svg", Width: 100, Height: 100},
			},
			wantErr: "at least one layer",
		},
		{
			name: "layer missing name",
			recipe: Recipe{
				Name:   "test",
				Output: OutputConfig{Path: "out.svg", Width: 100, Height: 100},
				Layers: []Layer{{Source: "file://test.json"}},
			},
			wantErr: "name is required",
		},
		{
			name: "layer missing source",
			recipe: Recipe{
				Name:   "test",
				Output: OutputConfig{Path: "out.svg", Width: 100, Height: 100},
				Layers: []Layer{{Name: "test"}},
			},
			wantErr: "source is required",
		},
		{
			name: "valid recipe",
			recipe: Recipe{
				Name:   "test",
				Output: OutputConfig{Path: "out.svg", Width: 100, Height: 100},
				Layers: []Layer{{Name: "base", Source: "file://test.json"}},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecipe(&tt.recipe)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateRecipe() unexpected error = %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("ValidateRecipe() expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Error = %q, should contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRecipeWithOperations(t *testing.T) {
	yaml := `
name: complex-map
output:
  path: output.svg
  width: 800
  height: 600
layers:
  - name: base
    source: census://states
    operations:
      - type: clip
        params:
          bounds: census://state/VT
      - type: simplify
        params:
          tolerance: 0.001
    style:
      fill: "#eee"
`
	recipe, err := LoadRecipeBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadRecipeBytes() error = %v", err)
	}

	if len(recipe.Layers[0].Operations) != 2 {
		t.Fatalf("len(Operations) = %d, want 2", len(recipe.Layers[0].Operations))
	}

	if recipe.Layers[0].Operations[0].Type != "clip" {
		t.Errorf("Operations[0].Type = %q, want %q", recipe.Layers[0].Operations[0].Type, "clip")
	}

	if recipe.Layers[0].Operations[1].Type != "simplify" {
		t.Errorf("Operations[1].Type = %q, want %q", recipe.Layers[0].Operations[1].Type, "simplify")
	}

	tol, ok := recipe.Layers[0].Operations[1].Params["tolerance"].(float64)
	if !ok || tol != 0.001 {
		t.Errorf("Operations[1].Params[tolerance] = %v, want 0.001", recipe.Layers[0].Operations[1].Params["tolerance"])
	}
}
