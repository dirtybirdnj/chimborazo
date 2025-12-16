// Package config handles YAML recipe parsing for Chimborazo.
//
// It loads recipe files that define map builds:
//
//	recipe, _ := config.Load("examples/vermont.yaml")
//	// recipe.Name, recipe.Layers, etc.
//
// Equivalent to the maury module in Strata Python.
package config
