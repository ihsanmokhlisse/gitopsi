package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/progress"
)

var getPasswordCmd = &cobra.Command{
	Use:   "get-password [tool]",
	Short: "Get the admin password for a GitOps tool",
	Long: `Retrieve the initial admin password for ArgoCD or other GitOps tools.

Examples:
  gitopsi get-password argocd   # Get ArgoCD admin password
  gitopsi get-password          # Auto-detect from setup summary`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGetPassword,
}

func init() {
	rootCmd.AddCommand(getPasswordCmd)
}

func runGetPassword(cmd *cobra.Command, args []string) error {
	tool := "argocd"
	if len(args) > 0 {
		tool = args[0]
	}

	// Try to get namespace from setup summary
	namespace := tool
	summary, err := progress.LoadSummary(".")
	if err == nil && summary.GitOpsTool.Namespace != "" {
		namespace = summary.GitOpsTool.Namespace
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var password string

	switch tool {
	case "argocd":
		password, err = getArgoCDPassword(ctx, namespace)
	case "flux":
		return fmt.Errorf("flux does not have an admin password")
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}

	if err != nil {
		return err
	}

	if verbose {
		pterm.Success.Printf("Password for %s:\n", tool)
	}
	fmt.Println(password)

	return nil
}

func getArgoCDPassword(ctx context.Context, namespace string) (string, error) {
	// Try to get password from kubectl
	cmdExec := exec.CommandContext(ctx, "kubectl", "get", "secret",
		"argocd-initial-admin-secret",
		"-n", namespace,
		"-o", "jsonpath={.data.password}")

	output, err := cmdExec.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get ArgoCD password: %w\nMake sure you have access to the cluster and ArgoCD is installed", err)
	}

	// Decode base64
	decodeCmd := exec.CommandContext(ctx, "bash", "-c",
		fmt.Sprintf("echo %s | base64 -d", strings.TrimSpace(string(output))))
	decoded, err := decodeCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to decode password: %w", err)
	}

	return strings.TrimSpace(string(decoded)), nil
}

