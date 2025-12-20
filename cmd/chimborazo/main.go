package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dirtybirdnj/chimborazo/internal/config"
	"github.com/dirtybirdnj/chimborazo/internal/sources"
	"github.com/dirtybirdnj/chimborazo/pkg/pipeline"
)

var version = "0.1.0-dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "chimborazo",
		Short: "Cartography at the speed of code",
		Long: `Chimborazo - High-performance geospatial visualization toolkit

Transform YAML recipes into beautiful SVG maps.`,
	}

	rootCmd.Version = version

	// Subcommands
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(cacheCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build [recipe.yaml]",
		Short: "Build a map from a recipe",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Building from %s...\n", args[0])

			recipe, err := config.LoadRecipe(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading recipe: %v\n", err)
				os.Exit(1)
			}

			fetcher, _ := sources.NewFetcher("~/.chimborazo/cache")
			builder := pipeline.NewBuilder(recipe, fetcher)
			builder.Verbose = true

			result, err := builder.Build()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Built %s (%d layers, %d features) in %v\n",
				result.OutputPath, result.LayerCount, result.FeatureCount, result.Duration)
		},
	}
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [recipe.yaml]",
		Short: "Validate a recipe without building",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			recipe, err := config.LoadRecipe(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
				os.Exit(1)
			}

			if err := config.ValidateRecipe(recipe); err != nil {
				fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Recipe valid: %s (%d layers)\n", recipe.Name, len(recipe.Layers))
		},
	}
}

func cacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the data cache",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List cached files",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Cache contents:")
			fmt.Println("Not implemented yet")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear the cache",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Clearing cache...")
			fmt.Println("Not implemented yet")
		},
	})

	return cmd
}
