/*
Package pipeline orchestrates the complete map build process.

# The Return Stage

I spent twenty-three years in Paris after my return, publishing the results
of my expedition. Thirty volumes. Hundreds of maps and diagrams. Thousands
of pages. I spent my entire fortune on these publications.

Why Paris? Because only there could I find the engravers, the printers, the
scientific community, the libraries I needed. No single person, no single
resource could have produced what I produced. It required coordination—the
aggregation of many capabilities toward a single goal.

Matthew Fontaine Maury understood this better than anyone. From his office
in Washington, he collected ship logs from captains worldwide. No single
captain could map the ocean's currents. But thousands of observations,
aggregated systematically, revealed patterns invisible to any individual
voyage: the Gulf Stream, the trade wind routes, the optimal paths across
the Atlantic.

"There is a river in the ocean," he wrote. "In the severest droughts it
never fails, and in the mightiest floods it never overflows."

That river only became visible through coordination.

# The Build Pipeline

This package coordinates all stages of the map build:

	1. Preparation  - Parse and validate the recipe
	2. Acquisition  - Fetch all required sources
	3. Transformation - Process each layer through its operations
	4. Revelation   - Render the final SVG output

Each stage depends on the previous. You cannot transform what you have not
acquired. You cannot render what you have not transformed. The pipeline
ensures correct ordering, handles errors gracefully, and reports progress.

# Usage

	// Load recipe and create pipeline
	recipe, _ := config.Load("recipe.yaml")
	p := pipeline.New(recipe)

	// Execute the complete build
	result, err := p.Build()
	if err != nil {
	    // Build failed at some stage
	}

	// result.OutputFiles contains paths to generated SVGs
	// result.Stats contains build metrics

	// Or execute with options
	result, err := p.Build(
	    pipeline.WithOutputDir("./output"),
	    pipeline.WithVerbose(true),
	    pipeline.WithCacheDir("~/.cache/chimborazo"),
	)

# Build Context

As the pipeline executes, it maintains a BuildContext—the accumulated state
of the expedition:

	type BuildContext struct {
	    Recipe          *config.Recipe
	    Cache           *sources.Cache
	    SourceData      map[string]geometry.FeatureCollection
	    ProcessedLayers map[string]geometry.FeatureCollection
	}

Sources are loaded into SourceData. Layers are processed into ProcessedLayers.
The context grows as the build progresses, carrying forward what each stage
needs from previous stages.

# Error Handling

The pipeline halts at the first error, reporting which stage failed and why.
Common failure modes:

	Preparation:    Invalid recipe, unknown operation types
	Acquisition:    Network errors, invalid URIs, corrupted cache
	Transformation: Invalid geometry, operation failures
	Revelation:     File I/O errors, invalid styles

Partial results are not produced. Either the build succeeds completely, or
it fails with a clear error message indicating what went wrong.

# Progress Reporting

For long builds, the pipeline reports progress:

	p := pipeline.New(recipe)
	p.OnProgress(func(stage string, current, total int) {
	    fmt.Printf("[%s] %d/%d\n", stage, current, total)
	})
	result, err := p.Build()

# Validation Mode

To check a recipe without executing the build:

	p := pipeline.New(recipe)
	if err := p.Validate(); err != nil {
	    // Recipe has errors
	}

Validation checks:
  - All source URIs are resolvable
  - All layer source references exist
  - All operation types are recognized
  - Output directory is writable

# Types

The primary types are:

  - [Pipeline]: The build orchestrator
  - [BuildContext]: Accumulated state during build
  - [BuildResult]: Output files and statistics
  - [BuildStats]: Metrics (sources fetched, features processed, etc.)
  - [BuildOption]: Configuration for the build process
*/
package pipeline
