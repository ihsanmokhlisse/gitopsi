package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/progress"
)

var openCmd = &cobra.Command{
	Use:   "open [target]",
	Short: "Open a GitOps resource in the browser",
	Long: `Open various GitOps resources in your default web browser.

Targets:
  argocd  - Open ArgoCD UI
  git     - Open Git repository
  docs    - Open generated documentation

Examples:
  gitopsi open argocd   # Open ArgoCD UI
  gitopsi open git      # Open Git repository in browser
  gitopsi open docs     # Open local documentation`,
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	target := args[0]

	summary, err := progress.LoadSummary(".")
	if err != nil {
		return fmt.Errorf("no gitopsi setup found. Run 'gitopsi init' first: %w", err)
	}

	var url string

	switch target {
	case "argocd", "argo":
		url = summary.GitOpsTool.URL
		if url == "" {
			return fmt.Errorf("ArgoCD URL not found in setup summary")
		}

	case "git", "repo", "repository":
		url = summary.Git.WebURL
		if url == "" {
			url = summary.Git.URL
		}
		if url == "" {
			return fmt.Errorf("Git URL not found in setup summary")
		}

	case "docs", "documentation":
		url = fmt.Sprintf("file://%s/docs/README.md", summary.Setup.Version)
		pterm.Info.Printf("Opening local documentation: %s/docs/\n", summary.Setup.Version)

	case "flux":
		return fmt.Errorf("Flux does not have a web UI. Use 'flux' CLI instead")

	default:
		return fmt.Errorf("unknown target: %s. Valid targets: argocd, git, docs", target)
	}

	if verbose {
		pterm.Info.Printf("Opening: %s\n", url)
	}

	return openBrowser(url)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

