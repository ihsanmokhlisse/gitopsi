package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/progress"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show GitOps setup status",
	Long: `Display the current status of the GitOps setup including
Git repository, cluster connection, and application sync status.

Examples:
  gitopsi status                # Show full status
  gitopsi status --json         # Output as JSON
  gitopsi status --quiet        # Minimal output`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	summary, err := progress.LoadSummary(projectPath)
	if err != nil {
		return fmt.Errorf("no gitopsi setup found in current directory: %w", err)
	}

	p := progress.New("Status", summary.Setup.Version)
	p.SetQuiet(quiet)
	p.SetJSON(jsonOutput)

	p.ShowSummary(summary)

	return nil
}

var jsonOutput bool
var quiet bool

func init() {
	statusCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON/YAML")
	statusCmd.Flags().BoolVar(&quiet, "quiet", false, "Minimal output")
}

