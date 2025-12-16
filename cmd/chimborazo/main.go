package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
			fmt.Println("Not implemented yet")
		},
	}
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [recipe.yaml]",
		Short: "Validate a recipe without building",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Validating %s...\n", args[0])
			fmt.Println("Not implemented yet")
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
