package cli

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/progress"
)

var infoCmd = &cobra.Command{
	Use:   "info [type]",
	Short: "Get information about GitOps setup",
	Long: `Display detailed information about specific parts of the GitOps setup.

Types:
  argocd  - ArgoCD connection and credentials
  git     - Git repository information
  cluster - Cluster connection details
  apps    - Application status

Examples:
  gitopsi info argocd   # Show ArgoCD info
  gitopsi info git      # Show Git repository info
  gitopsi info cluster  # Show cluster info
  gitopsi info apps     # Show application status`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	infoType := args[0]

	summary, err := progress.LoadSummary(".")
	if err != nil {
		return fmt.Errorf("no gitopsi setup found. Run 'gitopsi init' first: %w", err)
	}

	switch infoType {
	case "argocd", "argo":
		showArgoCDInfo(summary)
	case "flux":
		showFluxInfo(summary)
	case "git", "repo":
		showGitInfo(summary)
	case "cluster":
		showClusterInfo(summary)
	case "apps", "applications":
		showAppsInfo(summary)
	default:
		return fmt.Errorf("unknown info type: %s. Valid types: argocd, git, cluster, apps", infoType)
	}

	return nil
}

func showArgoCDInfo(summary *progress.SetupSummary) {
	if summary.GitOpsTool.Name != "argocd" {
		pterm.Warning.Println("ArgoCD is not configured in this setup")
		return
	}

	pterm.DefaultSection.Println("ArgoCD Information")

	data := pterm.TableData{
		{"URL", summary.GitOpsTool.URL},
		{"Username", summary.GitOpsTool.Username},
		{"Namespace", summary.GitOpsTool.Namespace},
		{"Version", summary.GitOpsTool.Version},
		{"Status", summary.GitOpsTool.Status},
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()

	fmt.Println()
	pterm.Info.Println("Get password: gitopsi get-password argocd")
	pterm.Info.Println("Open UI: gitopsi open argocd")
}

func showFluxInfo(summary *progress.SetupSummary) {
	if summary.GitOpsTool.Name != "flux" {
		pterm.Warning.Println("Flux is not configured in this setup")
		return
	}

	pterm.DefaultSection.Println("Flux Information")

	data := pterm.TableData{
		{"Namespace", summary.GitOpsTool.Namespace},
		{"Version", summary.GitOpsTool.Version},
		{"Status", summary.GitOpsTool.Status},
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()

	fmt.Println()
	pterm.Info.Println("Check status: flux check")
	pterm.Info.Println("View resources: flux get all")
}

func showGitInfo(summary *progress.SetupSummary) {
	pterm.DefaultSection.Println("Git Repository Information")

	data := pterm.TableData{
		{"URL", summary.Git.URL},
		{"Branch", summary.Git.Branch},
		{"Provider", summary.Git.Provider},
		{"Web URL", summary.Git.WebURL},
		{"Status", summary.Git.Status},
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()

	fmt.Println()
	pterm.Info.Println("Open repository: gitopsi open git")
}

func showClusterInfo(summary *progress.SetupSummary) {
	pterm.DefaultSection.Println("Cluster Information")

	data := pterm.TableData{
		{"Name", summary.Cluster.Name},
		{"URL", summary.Cluster.URL},
		{"Platform", summary.Cluster.Platform},
		{"Version", summary.Cluster.Version},
		{"Status", summary.Cluster.Status},
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()

	if len(summary.Cluster.Namespaces) > 0 {
		fmt.Println()
		pterm.Info.Println("Namespaces:")
		for _, ns := range summary.Cluster.Namespaces {
			pterm.Printf("  • %s\n", ns)
		}
	}
}

func showAppsInfo(summary *progress.SetupSummary) {
	pterm.DefaultSection.Println("Applications")

	if len(summary.Applications) == 0 {
		pterm.Info.Println("No applications found")
		return
	}

	for _, app := range summary.Applications {
		icon := "✓"
		if app.Status != "synced" && app.Status != "healthy" {
			icon = "⚠"
		}
		pterm.Printf("%s %s (%s) - %s\n", icon, app.Name, app.Type, app.Status)

		for _, child := range app.Children {
			pterm.Printf("   └── %s\n", child)
		}
	}
}
