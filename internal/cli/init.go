package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/generator"
	outputpkg "github.com/ihsanmokhlisse/gitopsi/internal/output"
	"github.com/ihsanmokhlisse/gitopsi/internal/prompt"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new GitOps repository",
	Long: `Initialize a new GitOps repository structure with all necessary
manifests, documentation, and scripts.

Can run in interactive mode (default) or with a config file.

Examples:
  gitopsi init                        # Interactive mode
  gitopsi init --config gitops.yaml   # Config file mode
  gitopsi init --dry-run              # Preview without writing`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	var cfg *config.Config
	var err error

	if cfgFile != "" {
		fmt.Printf("📄 Loading config from: %s\n", cfgFile)
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		fmt.Println("🎯 gitopsi - GitOps Repository Generator")
		fmt.Println()
		cfg, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
	}

	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	outputDir := GetOutput()
	if outputDir == "" {
		outputDir = "."
	}

	var absOutput string
	absOutput, err = filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}

	if !dryRun {
		if _, err := os.Stat(filepath.Join(absOutput, cfg.Project.Name)); err == nil {
			return fmt.Errorf("directory already exists: %s", cfg.Project.Name)
		}
	}

	writer := outputpkg.New(absOutput, dryRun, verbose)
	gen := generator.New(cfg, writer, verbose)

	if dryRun {
		fmt.Println()
		fmt.Println("🔍 DRY RUN - No files will be written")
		fmt.Println()
	}

	if err := gen.Generate(); err != nil {
		return err
	}

	if dryRun {
		fmt.Println("\n🔍 DRY RUN complete - no files were written")
	}

	return nil
}
