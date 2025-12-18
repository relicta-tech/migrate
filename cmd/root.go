// Package cmd implements the CLI commands for the migrate tool.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/relicta-tech/migrate/internal/converter"
	"github.com/relicta-tech/migrate/internal/detector"
	"github.com/relicta-tech/migrate/internal/output"
)

var (
	// Flags
	outputFile string
	dryRun     bool
	verbose    bool
	force      bool

	// Version info (set by ldflags)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "migrate [directory]",
	Short: "Migrate to Relicta from other release tools",
	Long: `Migrate converts configuration from other release management tools to Relicta.

Supported tools:
  - semantic-release (.releaserc, .releaserc.json, .releaserc.yaml, release.config.js)
  - release-it (.release-it.json, .release-it.yaml, .release-it.js, package.json)
  - standard-version (.versionrc, .versionrc.json, package.json)

Usage:
  migrate                    # Auto-detect and convert in current directory
  migrate /path/to/project   # Convert specific project
  migrate --dry-run          # Preview without writing files`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrate,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "release.config.yaml", "Output file path")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without writing files")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing release.config.yaml")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(detectCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("migrate %s (commit: %s, built: %s)\n", version, commit, date)
	},
}

var detectCmd = &cobra.Command{
	Use:   "detect [directory]",
	Short: "Detect which release tool is configured",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		result, err := detector.Detect(dir)
		if err != nil {
			return err
		}

		if result.Tool == detector.ToolNone {
			fmt.Println("No release tool configuration detected.")
			return nil
		}

		fmt.Printf("Detected: %s\n", result.Tool)
		fmt.Printf("Config file: %s\n", result.ConfigFile)
		if verbose && len(result.Details) > 0 {
			fmt.Println("\nDetails:")
			for k, v := range result.Details {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
		return nil
	},
}

func runMigrate(_ *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Check if output already exists
	outputPath := filepath.Join(dir, outputFile)
	if _, err := os.Stat(outputPath); err == nil && !force && !dryRun {
		return fmt.Errorf("%s already exists. Use --force to overwrite", outputPath)
	}

	// Detect tool
	if verbose {
		fmt.Println("Detecting release tool configuration...")
	}

	result, err := detector.Detect(dir)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	if result.Tool == detector.ToolNone {
		return fmt.Errorf("no release tool configuration found in %s", dir)
	}

	fmt.Printf("Detected: %s (%s)\n", result.Tool, result.ConfigFile)

	// Convert configuration
	if verbose {
		fmt.Println("Converting configuration...")
	}

	config, err := converter.Convert(result)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Output
	if dryRun {
		fmt.Println("\n--- Generated release.config.yaml (dry-run) ---")
		yaml, err := output.ToYAML(config)
		if err != nil {
			return err
		}
		fmt.Println(yaml)
		fmt.Println("--- End of preview ---")
		return nil
	}

	// Write file
	if err := output.WriteYAML(outputPath, config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("\nSuccessfully created %s\n", outputPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the generated configuration")
	fmt.Println("  2. Run 'relicta plan --dry-run' to test")
	fmt.Println("  3. Remove old configuration files when ready")

	return nil
}
