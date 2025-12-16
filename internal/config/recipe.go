package config

// Recipe defines a complete map build
type Recipe struct {
	Name   string       `yaml:"name"`
	Output OutputConfig `yaml:"output"`
	Bounds BoundsConfig `yaml:"bounds"`
	Layers []Layer      `yaml:"layers"`
}

// OutputConfig defines output file settings
type OutputConfig struct {
	Path   string  `yaml:"path"`
	Width  float64 `yaml:"width"`
	Height float64 `yaml:"height"`
	Units  string  `yaml:"units"` // "inches", "mm", "px"
}

// BoundsConfig defines the map extent
type BoundsConfig struct {
	Source  string  `yaml:"source"`  // URI like "census://state/VT"
	Padding float64 `yaml:"padding"` // Fraction to expand bounds
}

// Layer defines a single map layer
type Layer struct {
	Name       string            `yaml:"name"`
	Source     string            `yaml:"source"`
	Operations []Operation       `yaml:"operations,omitempty"`
	Style      map[string]string `yaml:"style,omitempty"`
	Filter     string            `yaml:"filter,omitempty"`
}

// Operation defines a geometry operation
type Operation struct {
	Type   string                 `yaml:"type"` // "clip", "subtract", "merge", etc.
	Params map[string]interface{} `yaml:"params,omitempty"`
}
