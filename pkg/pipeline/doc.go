// Package pipeline orchestrates the map build process.
//
// It coordinates config loading, data fetching, geometry operations,
// and SVG output into a single build workflow.
//
//	builder := pipeline.NewBuilder(recipe)
//	result, _ := builder.Build()
//	// result.OutputPath contains the generated SVG
package pipeline
