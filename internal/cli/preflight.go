package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// PreflightResult represents the result of a preflight check
type PreflightResult struct {
	Name    string
	Status  string // ok, warn, fail
	Message string
	Details string
}

var preflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Run pre-flight checks on the target cluster",
	Long: `Validates cluster readiness for GitOps deployment.

Checks include:
- Cluster connectivity and API access
- Required permissions
- GitOps tool status (ArgoCD/Flux)
- Required CRDs
- Platform detection`,
	RunE: runPreflight,
}

var (
	preflightClusterURL string
	preflightKubeconfig string
	preflightContext    string
	preflightGitopsTool string
	preflightTimeout    int
)

func init() {
	rootCmd.AddCommand(preflightCmd)

	preflightCmd.Flags().StringVar(&preflightClusterURL, "cluster", "", "Cluster API URL")
	preflightCmd.Flags().StringVar(&preflightKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	preflightCmd.Flags().StringVar(&preflightContext, "context", "", "Kubernetes context to use")
	preflightCmd.Flags().StringVar(&preflightGitopsTool, "gitops-tool", "argocd", "GitOps tool to check (argocd, flux)")
	preflightCmd.Flags().IntVar(&preflightTimeout, "timeout", 30, "Timeout in seconds for each check")
}

func runPreflight(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(preflightTimeout)*time.Second)
	defer cancel()

	pterm.DefaultHeader.WithFullWidth().Println("🔍 Pre-flight Cluster Check")
	fmt.Println()

	results := make([]PreflightResult, 0)

	// 1. Check cluster connectivity
	result := checkClusterConnectivity(ctx)
	results = append(results, result)
	printResult(result)

	if result.Status == "fail" {
		printSummary(results)
		return fmt.Errorf("cluster connectivity check failed")
	}

	// 2. Check API server version
	result = checkAPIServerVersion(ctx)
	results = append(results, result)
	printResult(result)

	// 3. Check permissions
	result = checkPermissions(ctx)
	results = append(results, result)
	printResult(result)

	// 4. Check platform (OpenShift/Kubernetes)
	result = checkPlatform(ctx)
	results = append(results, result)
	printResult(result)

	// 5. Check GitOps tool
	result = checkGitOpsTool(ctx, preflightGitopsTool)
	results = append(results, result)
	printResult(result)

	// 6. Check required CRDs
	result = checkRequiredCRDs(ctx, preflightGitopsTool)
	results = append(results, result)
	printResult(result)

	// 7. Check storage classes
	result = checkStorageClasses(ctx)
	results = append(results, result)
	printResult(result)

	fmt.Println()
	printSummary(results)

	// Return error if any critical checks failed
	for _, r := range results {
		if r.Status == "fail" {
			return fmt.Errorf("pre-flight check failed: %s", r.Name)
		}
	}

	return nil
}

func checkClusterConnectivity(ctx context.Context) PreflightResult {
	result := PreflightResult{Name: "Cluster Connectivity"}

	args := []string{"cluster-info"}
	if preflightContext != "" {
		args = append([]string{"--context", preflightContext}, args...)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Try with oc for OpenShift
		cmd = exec.CommandContext(ctx, "oc", args...)
		output, err = cmd.CombinedOutput()
	}

	if err != nil {
		result.Status = "fail"
		result.Message = "Cannot connect to cluster"
		result.Details = string(output)
		return result
	}

	result.Status = "ok"
	result.Message = "Connected"
	return result
}

func checkAPIServerVersion(ctx context.Context) PreflightResult {
	result := PreflightResult{Name: "API Server Version"}

	args := []string{"version", "--short"}
	if preflightContext != "" {
		args = append([]string{"--context", preflightContext}, args...)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		cmd = exec.CommandContext(ctx, "oc", args...)
		output, err = cmd.CombinedOutput()
	}

	if err != nil {
		result.Status = "warn"
		result.Message = "Could not determine version"
		return result
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Server") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				result.Message = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	if result.Message == "" {
		result.Message = strings.TrimSpace(string(output))
	}

	result.Status = "ok"
	return result
}

func checkPermissions(ctx context.Context) PreflightResult {
	result := PreflightResult{Name: "Cluster Permissions"}

	checks := []struct {
		resource string
		verb     string
	}{
		{"namespaces", "create"},
		{"deployments", "create"},
		{"services", "create"},
		{"configmaps", "create"},
		{"secrets", "create"},
	}

	failed := []string{}
	for _, check := range checks {
		args := []string{"auth", "can-i", check.verb, check.resource}
		if preflightContext != "" {
			args = append([]string{"--context", preflightContext}, args...)
		}

		cmd := exec.CommandContext(ctx, "kubectl", args...)
		output, _ := cmd.CombinedOutput()

		if !strings.Contains(strings.ToLower(string(output)), "yes") {
			failed = append(failed, fmt.Sprintf("%s %s", check.verb, check.resource))
		}
	}

	if len(failed) > 0 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Missing: %s", strings.Join(failed, ", "))
		return result
	}

	result.Status = "ok"
	result.Message = "All required permissions granted"
	return result
}

func checkPlatform(ctx context.Context) PreflightResult {
	result := PreflightResult{Name: "Platform Detection"}

	// Check for OpenShift
	args := []string{"api-resources", "--api-group=route.openshift.io"}
	if preflightContext != "" {
		args = append([]string{"--context", preflightContext}, args...)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, _ := cmd.CombinedOutput()

	if strings.Contains(string(output), "routes") {
		result.Status = "ok"
		result.Message = "OpenShift"
		return result
	}

	// Check for EKS
	args = []string{"get", "nodes", "-o", "jsonpath={.items[0].spec.providerID}"}
	if preflightContext != "" {
		args = append([]string{"--context", preflightContext}, args...)
	}

	cmd = exec.CommandContext(ctx, "kubectl", args...)
	output, _ = cmd.CombinedOutput()

	if strings.Contains(string(output), "aws") {
		result.Status = "ok"
		result.Message = "Amazon EKS"
		return result
	}

	if strings.Contains(string(output), "azure") {
		result.Status = "ok"
		result.Message = "Azure AKS"
		return result
	}

	if strings.Contains(string(output), "gce") {
		result.Status = "ok"
		result.Message = "Google GKE"
		return result
	}

	result.Status = "ok"
	result.Message = "Kubernetes"
	return result
}

func checkGitOpsTool(ctx context.Context, tool string) PreflightResult {
	result := PreflightResult{Name: fmt.Sprintf("GitOps Tool (%s)", tool)}

	var namespace string
	var deployments []string

	switch tool {
	case "argocd":
		// Check for OpenShift GitOps first
		namespace = "openshift-gitops"
		deployments = []string{"openshift-gitops-server", "openshift-gitops-repo-server", "openshift-gitops-applicationset-controller"}

		// Check if namespace exists
		args := []string{"get", "namespace", namespace}
		if preflightContext != "" {
			args = append([]string{"--context", preflightContext}, args...)
		}

		cmd := exec.CommandContext(ctx, "kubectl", args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			// Try standard argocd namespace
			namespace = "argocd"
			deployments = []string{"argocd-server", "argocd-repo-server", "argocd-applicationset-controller"}

			args = []string{"get", "namespace", namespace}
			if preflightContext != "" {
				args = append([]string{"--context", preflightContext}, args...)
			}

			cmd = exec.CommandContext(ctx, "kubectl", args...)
			if _, err := cmd.CombinedOutput(); err != nil {
				result.Status = "warn"
				result.Message = "Not installed"
				result.Details = "Neither openshift-gitops nor argocd namespace found"
				return result
			}
		}

	case "flux":
		namespace = "flux-system"
		deployments = []string{"source-controller", "kustomize-controller", "helm-controller"}

		args := []string{"get", "namespace", namespace}
		if preflightContext != "" {
			args = append([]string{"--context", preflightContext}, args...)
		}

		cmd := exec.CommandContext(ctx, "kubectl", args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			result.Status = "warn"
			result.Message = "Not installed"
			result.Details = "flux-system namespace not found"
			return result
		}
	}

	// Check deployments
	running := 0
	total := len(deployments)

	for _, deploy := range deployments {
		args := []string{"get", "deployment", deploy, "-n", namespace, "-o", "jsonpath={.status.availableReplicas}"}
		if preflightContext != "" {
			args = append([]string{"--context", preflightContext}, args...)
		}

		cmd := exec.CommandContext(ctx, "kubectl", args...)
		output, err := cmd.CombinedOutput()

		if err == nil && strings.TrimSpace(string(output)) != "" && strings.TrimSpace(string(output)) != "0" {
			running++
		}
	}

	if running == 0 {
		result.Status = "fail"
		result.Message = fmt.Sprintf("Not running (0/%d components)", total)
		result.Details = fmt.Sprintf("Namespace %s exists but no deployments running", namespace)
		return result
	}

	if running < total {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Partially running (%d/%d components)", running, total)
		return result
	}

	result.Status = "ok"
	result.Message = fmt.Sprintf("Running (%d/%d components) in %s", running, total, namespace)
	return result
}

func checkRequiredCRDs(ctx context.Context, tool string) PreflightResult {
	result := PreflightResult{Name: "Required CRDs"}

	var crds []string

	switch tool {
	case "argocd":
		crds = []string{
			"applications.argoproj.io",
			"applicationsets.argoproj.io",
			"appprojects.argoproj.io",
		}
	case "flux":
		crds = []string{
			"gitrepositories.source.toolkit.fluxcd.io",
			"kustomizations.kustomize.toolkit.fluxcd.io",
			"helmreleases.helm.toolkit.fluxcd.io",
		}
	}

	found := 0
	missing := []string{}

	for _, crd := range crds {
		args := []string{"get", "crd", crd}
		if preflightContext != "" {
			args = append([]string{"--context", preflightContext}, args...)
		}

		cmd := exec.CommandContext(ctx, "kubectl", args...)
		if _, err := cmd.CombinedOutput(); err == nil {
			found++
		} else {
			missing = append(missing, crd)
		}
	}

	if found == 0 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("None found (0/%d)", len(crds))
		result.Details = fmt.Sprintf("Missing: %s", strings.Join(missing, ", "))
		return result
	}

	if found < len(crds) {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Partial (%d/%d)", found, len(crds))
		result.Details = fmt.Sprintf("Missing: %s", strings.Join(missing, ", "))
		return result
	}

	result.Status = "ok"
	result.Message = fmt.Sprintf("All present (%d/%d)", found, len(crds))
	return result
}

func checkStorageClasses(ctx context.Context) PreflightResult {
	result := PreflightResult{Name: "Storage Classes"}

	args := []string{"get", "storageclass", "-o", "jsonpath={.items[*].metadata.name}"}
	if preflightContext != "" {
		args = append([]string{"--context", preflightContext}, args...)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Status = "warn"
		result.Message = "Could not check"
		return result
	}

	classes := strings.Fields(string(output))
	if len(classes) == 0 {
		result.Status = "warn"
		result.Message = "None found"
		return result
	}

	result.Status = "ok"
	result.Message = fmt.Sprintf("%d available", len(classes))
	result.Details = strings.Join(classes, ", ")
	return result
}

func printResult(result PreflightResult) {
	var icon string
	var color pterm.Color

	switch result.Status {
	case "ok":
		icon = "✅"
		color = pterm.FgGreen
	case "warn":
		icon = "⚠️ "
		color = pterm.FgYellow
	case "fail":
		icon = "❌"
		color = pterm.FgRed
	}

	pterm.Printf("%s %-25s %s\n", icon, result.Name, pterm.NewStyle(color).Sprint(result.Message))

	if result.Details != "" && (result.Status == "warn" || result.Status == "fail") {
		pterm.Printf("   └─ %s\n", pterm.FgGray.Sprint(result.Details))
	}
}

func printSummary(results []PreflightResult) {
	ok := 0
	warn := 0
	fail := 0

	for _, r := range results {
		switch r.Status {
		case "ok":
			ok++
		case "warn":
			warn++
		case "fail":
			fail++
		}
	}

	pterm.DefaultBox.WithTitle("Summary").Println(
		fmt.Sprintf("✅ Passed: %d  ⚠️  Warnings: %d  ❌ Failed: %d", ok, warn, fail),
	)

	if fail > 0 {
		pterm.Error.Println("Pre-flight check FAILED - cluster not ready for GitOps deployment")
	} else if warn > 0 {
		pterm.Warning.Println("Pre-flight check PASSED with warnings - some components may need attention")
	} else {
		pterm.Success.Println("Pre-flight check PASSED - cluster ready for GitOps deployment")
	}
}




