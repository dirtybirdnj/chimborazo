package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dirtybirdnj/chimborazo/internal/config"
	"github.com/dirtybirdnj/chimborazo/pkg/pipeline"
)

var version = "0.1.0-dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "chimborazo",
		Short: "Cartography at the speed of code",
		Long: `Chimborazo - High-performance geospatial visualization toolkit

Transform YAML recipes into beautiful SVG maps.

Named after the Ecuadorian volcano whose peak Humboldt climbed in 1802,
reaching 19,286 feet - higher than any European had ever stood.`,
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

func getCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".chimborazo/cache"
	}
	return filepath.Join(home, ".chimborazo", "cache")
}

func buildCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "build [recipe.yaml]",
		Short: "Build a map from a recipe",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			recipePath := args[0]

			if verbose {
				fmt.Printf("Loading recipe: %s\n", recipePath)
			}

			recipe, err := config.LoadRecipe(recipePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading recipe: %v\n", err)
				os.Exit(1)
			}

			builder, err := pipeline.NewBuilder(recipe, getCacheDir())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating builder: %v\n", err)
				os.Exit(1)
			}
			builder.Verbose = verbose

			result, err := builder.Build()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
				os.Exit(1)
			}

			// Report any non-fatal errors
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
			}

			fmt.Printf("✓ Built %s (%d layers, %d features) in %v\n",
				result.OutputPath, result.LayerCount, result.FeatureCount, result.Duration)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress")

	return cmd
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
			cacheDir := getCacheDir()
			fmt.Printf("Cache directory: %s\n", cacheDir)

			// Walk and list files
			filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() {
					rel, _ := filepath.Rel(cacheDir, path)
					fmt.Printf("  %s (%.1f KB)\n", rel, float64(info.Size())/1024)
				}
				return nil
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear the cache",
		Run: func(cmd *cobra.Command, args []string) {
			cacheDir := getCacheDir()
			fmt.Printf("Clearing cache: %s\n", cacheDir)

			if err := os.RemoveAll(cacheDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("✓ Cache cleared")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show cache directory path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(getCacheDir())
		},
	})

	return cmd
}
