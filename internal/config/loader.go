package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadRecipe reads and parses a recipe YAML file
func LoadRecipe(path string) (*Recipe, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening recipe file: %w", err)
	}
	defer file.Close()

	var recipe Recipe
	if err := yaml.NewDecoder(file).Decode(&recipe); err != nil {
		return nil, fmt.Errorf("parsing recipe YAML: %w", err)
	}

	return &recipe, nil
}

// LoadRecipeBytes parses recipe from YAML bytes
func LoadRecipeBytes(data []byte) (*Recipe, error) {
	var recipe Recipe
	if err := yaml.Unmarshal(data, &recipe); err != nil {
		return nil, fmt.Errorf("parsing recipe YAML: %w", err)
	}
	return &recipe, nil
}

// ValidateRecipe checks that a recipe has required fields
func ValidateRecipe(r *Recipe) error {
	if r.Name == "" {
		return fmt.Errorf("recipe name is required")
	}

	if r.Output.Path == "" {
		return fmt.Errorf("output path is required")
	}

	if r.Output.Width <= 0 {
		return fmt.Errorf("output width must be positive, got %v", r.Output.Width)
	}

	if r.Output.Height <= 0 {
		return fmt.Errorf("output height must be positive, got %v", r.Output.Height)
	}

	if len(r.Layers) == 0 {
		return fmt.Errorf("at least one layer is required")
	}

	// Validate each layer
	for i, layer := range r.Layers {
		if layer.Name == "" {
			return fmt.Errorf("layer %d: name is required", i)
		}
		if layer.Source == "" {
			return fmt.Errorf("layer %d (%s): source is required", i, layer.Name)
		}
	}

	return nil
}
