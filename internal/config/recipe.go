package config

// Recipe defines a complete map build
type Recipe struct {
	Name     string                `yaml:"name"`
	Output   OutputConfig          `yaml:"output"`
	Bounds   BoundsConfig          `yaml:"bounds"`
	Sources  map[string]SourceDef  `yaml:"sources,omitempty"`
	Defaults DefaultsConfig        `yaml:"defaults,omitempty"`
	Layers   []Layer               `yaml:"layers"`
}

// SourceDef defines a named data source.
type SourceDef struct {
	URI string `yaml:"uri"`
}

// DefaultsConfig defines default values inherited by layers.
type DefaultsConfig struct {
	Style map[string]string `yaml:"style,omitempty"`
}

// OutputConfig defines output file settings
type OutputConfig struct {
	Path           string  `yaml:"path"`
	Width          float64 `yaml:"width"`
	Height         float64 `yaml:"height"`
	Units          string  `yaml:"units"`            // "inches", "mm", "px"
	PerLayer       bool    `yaml:"per_layer"`        // Write each layer as separate SVG
	Margin         float64 `yaml:"margin"`           // Margin in inches (default 0.5)
	Quality        string  `yaml:"quality"`          // "low", "medium", "high", "plotter" (default "high")
	MinFeatureSize float64 `yaml:"min_feature_size"` // Minimum feature size in mm (filters tiny features)
	Rulers         bool    `yaml:"rulers"`           // Draw inch rulers on left and top edges
}

// QualityPreset defines simplification settings for a quality level.
type QualityPreset struct {
	Tolerance   float64 // Douglas-Peucker tolerance in degrees
	Description string
}

// QualityPresets maps quality names to their settings.
var QualityPresets = map[string]QualityPreset{
	"low": {
		Tolerance:   0.001,   // Aggressive simplification for web preview
		Description: "Web preview - smaller files, less detail",
	},
	"medium": {
		Tolerance:   0.0003,  // Moderate simplification for screen display
		Description: "Screen display - balanced size and detail",
	},
	"high": {
		Tolerance:   0.0001,  // Minimal simplification for print
		Description: "Print quality - maximum detail",
	},
	"plotter": {
		Tolerance:   0.00005, // Very minimal - optimized for pen plotters
		Description: "Pen plotter - smooth curves, maximum fidelity",
	},
}

// GetQualityPreset returns the preset for a quality level, defaulting to "high".
func GetQualityPreset(quality string) QualityPreset {
	if preset, ok := QualityPresets[quality]; ok {
		return preset
	}
	return QualityPresets["high"]
}

// BoundsConfig defines the map extent
type BoundsConfig struct {
	Source   string    `yaml:"source"`   // URI like "census://state/VT"
	Padding  float64   `yaml:"padding"`  // Fraction to expand bounds
	Explicit []float64 `yaml:"explicit"` // [West, South, East, North] in degrees
}

// Layer defines a single map layer
type Layer struct {
	Name          string            `yaml:"name"`
	Source        string            `yaml:"source"`
	Operations    []Operation       `yaml:"operations,omitempty"`
	Style         map[string]string `yaml:"style,omitempty"`
	FillBy        string            `yaml:"fill_by,omitempty"`       // Property to color by
	ColorMap      map[string]string `yaml:"color_map,omitempty"`     // Property value → color
	VaryFill      bool              `yaml:"vary_fill,omitempty"`     // Slight color variations
	Patterns      []string          `yaml:"patterns,omitempty"`      // Pattern names to cycle through (for rat-king)
	Filter        string            `yaml:"filter,omitempty"`
	Labels        bool              `yaml:"labels,omitempty"`        // Show labels for features
	LabelProperty string            `yaml:"label_property,omitempty"` // Property to use for labels (default: "NAME")
}

// Operation defines a geometry operation
type Operation struct {
	Type   string                 `yaml:"type"` // "clip", "subtract", "merge", etc.
	Params map[string]interface{} `yaml:"params,omitempty"`
}
