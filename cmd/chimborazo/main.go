package main

import (
	"fmt"
	"os"
	"os/exec"
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
	rootCmd.AddCommand(analyzeCmd())
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
	var optimize bool
	var showStats bool

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

			// Show stats if requested or always in verbose mode
			if showStats || verbose {
				fmt.Printf("  SVG stats: %d paths, %d groups, %.1f MB\n",
					result.Stats.Paths, result.Stats.Groups,
					float64(result.Stats.FileSize)/(1024*1024))
			}

			// Run svgo optimization if requested
			if optimize {
				if err := runSVGO(result.OutputPath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: svgo optimization failed: %v\n", err)
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress")
	cmd.Flags().BoolVarP(&optimize, "optimize", "O", false, "Optimize SVG with svgo (requires svgo installed)")
	cmd.Flags().BoolVarP(&showStats, "stats", "s", false, "Show SVG element statistics")

	return cmd
}

// runSVGO runs svgo to optimize the SVG file in place.
func runSVGO(path string) error {
	// Check if svgo is available
	if _, err := exec.LookPath("svgo"); err != nil {
		return fmt.Errorf("svgo not found in PATH (install with: npm install -g svgo)")
	}

	// Get original size
	origInfo, _ := os.Stat(path)
	origSize := origInfo.Size()

	// Run svgo
	cmd := exec.Command("svgo", path, "-o", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("svgo failed: %w\n%s", err, output)
	}

	// Get new size
	newInfo, _ := os.Stat(path)
	newSize := newInfo.Size()

	// Report savings
	savings := float64(origSize-newSize) / float64(origSize) * 100
	fmt.Printf("✓ Optimized with svgo: %.1f MB → %.1f MB (%.1f%% reduction)\n",
		float64(origSize)/(1024*1024), float64(newSize)/(1024*1024), savings)

	return nil
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

func analyzeCmd() *cobra.Command {
	var verbose bool
	var debugOutput bool
	var generatePNG bool
	var openResult bool

	cmd := &cobra.Command{
		Use:   "analyze [recipe.yaml]",
		Short: "Build and analyze a map for visual errors",
		Long: `Builds a map and analyzes it for potential visual errors such as:
- Water/cutout mismatches (holes in polygons without corresponding water fill)
- Overlapping features that may cause rendering issues
- Features that exceed bounds

Use --debug to generate a debug overlay SVG highlighting issues.
Use --png to generate a PNG preview for visual inspection.
Use --open to open the result in default viewer.`,
		Args: cobra.ExactArgs(1),
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

			// Build with analysis enabled
			result, analysis, err := builder.BuildWithAnalysis()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Built %s (%d layers, %d features) in %v\n",
				result.OutputPath, result.LayerCount, result.FeatureCount, result.Duration)

			// Report analysis results
			fmt.Println("\n=== Analysis Results ===")
			if len(analysis.Warnings) == 0 {
				fmt.Println("✓ No issues detected")
			} else {
				for _, w := range analysis.Warnings {
					fmt.Printf("⚠ %s\n", w)
				}
			}

			// Generate PNG preview for visual inspection
			if generatePNG {
				pngPath := "/tmp/" + filepath.Base(result.OutputPath) + ".png"
				fmt.Printf("\nGenerating PNG preview: %s\n", pngPath)

				// Use qlmanage on macOS to generate PNG
				execCmd := exec.Command("qlmanage", "-t", "-s", "2000", "-o", "/tmp", result.OutputPath)
				if err := execCmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not generate PNG (qlmanage failed): %v\n", err)
				} else {
					fmt.Printf("✓ PNG preview: %s\n", pngPath)
					fmt.Println("\n=== Visual Inspection Checklist ===")
					fmt.Println("□ Check water bodies have blue fill (no green/town color showing)")
					fmt.Println("□ Check town boundaries align with water edges")
					fmt.Println("□ Check for orphan features (small dots or specks)")
					fmt.Println("□ Check state/province boundaries are continuous")

					if openResult {
						exec.Command("open", pngPath).Run()
					}
				}
			}

			if debugOutput && len(analysis.Warnings) > 0 {
				debugPath := result.OutputPath[:len(result.OutputPath)-4] + "_debug.svg"
				if err := builder.WriteDebugOverlay(debugPath, analysis); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not write debug overlay: %v\n", err)
				} else {
					fmt.Printf("\n✓ Debug overlay written to %s\n", debugPath)
				}
			}

			// Open SVG if requested
			if openResult && !generatePNG {
				exec.Command("open", result.OutputPath).Run()
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress")
	cmd.Flags().BoolVarP(&debugOutput, "debug", "d", false, "Generate debug overlay SVG")
	cmd.Flags().BoolVarP(&generatePNG, "png", "p", false, "Generate PNG preview for visual inspection")
	cmd.Flags().BoolVarP(&openResult, "open", "o", false, "Open result in default viewer")

	return cmd
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
