package progress

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Validator performs validation checks on the setup.
type Validator struct {
	checks  []ValidationCheck
	verbose bool
}

// NewValidator creates a new Validator instance.
func NewValidator(verbose bool) *Validator {
	return &Validator{
		checks:  make([]ValidationCheck, 0),
		verbose: verbose,
	}
}

// ValidateGit performs Git-related validation checks.
func (v *Validator) ValidateGit(ctx context.Context, repoURL, branch string) []ValidationCheck {
	checks := []ValidationCheck{
		{
			Name:   "Repository accessible",
			Check:  "Can connect to Git repository",
			Status: "pending",
		},
		{
			Name:   "Branch exists",
			Check:  fmt.Sprintf("Branch '%s' exists", branch),
			Status: "pending",
		},
	}

	// Check repo accessibility
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "ls-remote", repoURL)
	if err := cmd.Run(); err != nil {
		checks[0].Status = "failed"
		checks[0].Message = fmt.Sprintf("Cannot access repository: %v", err)
	} else {
		checks[0].Status = "passed"
	}

	// Check branch exists
	cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", repoURL, branch)
	output, err := cmd.CombinedOutput()
	if err != nil || len(output) == 0 {
		checks[1].Status = "warning"
		checks[1].Message = "Branch will be created on first push"
	} else {
		checks[1].Status = "passed"
	}

	v.checks = append(v.checks, checks...)
	return checks
}

// ValidateCluster performs cluster-related validation checks.
func (v *Validator) ValidateCluster(ctx context.Context, clusterURL string) []ValidationCheck {
	checks := []ValidationCheck{
		{
			Name:   "API accessible",
			Check:  "Can reach cluster API",
			Status: "pending",
		},
		{
			Name:   "RBAC permissions",
			Check:  "Can create namespaces and deployments",
			Status: "pending",
		},
		{
			Name:   "CRD support",
			Check:  "Can create custom resources",
			Status: "pending",
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check API accessibility
	cmd := exec.CommandContext(ctx, "kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		checks[0].Status = "failed"
		checks[0].Message = fmt.Sprintf("Cannot reach cluster API: %v", err)
	} else {
		checks[0].Status = "passed"
	}

	// Check RBAC - can create namespace
	cmd = exec.CommandContext(ctx, "kubectl", "auth", "can-i", "create", "namespaces")
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "yes") {
		checks[1].Status = "warning"
		checks[1].Message = "Limited permissions - may need cluster-admin"
	} else {
		checks[1].Status = "passed"
	}

	// Check CRD support
	cmd = exec.CommandContext(ctx, "kubectl", "auth", "can-i", "create", "customresourcedefinitions")
	output, err = cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "yes") {
		checks[2].Status = "warning"
		checks[2].Message = "Cannot create CRDs - GitOps tool may need pre-installed CRDs"
	} else {
		checks[2].Status = "passed"
	}

	v.checks = append(v.checks, checks...)
	return checks
}

// ValidateArgoCD performs ArgoCD-related validation checks.
func (v *Validator) ValidateArgoCD(ctx context.Context, namespace string) []ValidationCheck {
	checks := []ValidationCheck{
		{
			Name:   "Pods healthy",
			Check:  "All ArgoCD pods are running",
			Status: "pending",
		},
		{
			Name:   "Server accessible",
			Check:  "ArgoCD UI is reachable",
			Status: "pending",
		},
		{
			Name:   "Repository connected",
			Check:  "Can sync from Git repository",
			Status: "pending",
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Check pods healthy
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-n", namespace, "-o", "jsonpath={.items[*].status.phase}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		checks[0].Status = "failed"
		checks[0].Message = fmt.Sprintf("Cannot get pod status: %v", err)
	} else {
		phases := strings.Fields(string(output))
		allRunning := true
		for _, phase := range phases {
			if phase != "Running" {
				allRunning = false
				break
			}
		}
		if allRunning && len(phases) > 0 {
			checks[0].Status = "passed"
		} else {
			checks[0].Status = "warning"
			checks[0].Message = "Some pods are not running yet"
		}
	}

	// Check server accessible
	cmd = exec.CommandContext(ctx, "kubectl", "get", "svc", "argocd-server", "-n", namespace, "-o", "jsonpath={.spec.type}")
	_, err = cmd.CombinedOutput()
	if err != nil {
		checks[1].Status = "warning"
		checks[1].Message = "Could not verify ArgoCD service"
	} else {
		checks[1].Status = "passed"
	}

	// Check repository connected - try to list repos
	cmd = exec.CommandContext(ctx, "kubectl", "get", "secret", "-n", namespace, "-l", "argocd.argoproj.io/secret-type=repository", "-o", "name")
	repoOutput, repoErr := cmd.CombinedOutput()
	if repoErr != nil || strings.TrimSpace(string(repoOutput)) == "" {
		checks[2].Status = "warning"
		checks[2].Message = "No repository secrets found"
	} else {
		checks[2].Status = "passed"
	}

	v.checks = append(v.checks, checks...)
	return checks
}

// ValidateSync performs sync validation checks.
func (v *Validator) ValidateSync(ctx context.Context, namespace string) []ValidationCheck {
	checks := []ValidationCheck{
		{
			Name:   "Initial sync",
			Check:  "Applications are synced",
			Status: "pending",
		},
		{
			Name:   "Resources healthy",
			Check:  "No degraded resources",
			Status: "pending",
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Check application sync status
	cmd := exec.CommandContext(ctx, "kubectl", "get", "applications.argoproj.io", "-n", namespace, "-o", "jsonpath={.items[*].status.sync.status}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		checks[0].Status = "warning"
		checks[0].Message = "Could not verify sync status"
	} else {
		statuses := strings.Fields(string(output))
		allSynced := true
		for _, status := range statuses {
			if status != "Synced" {
				allSynced = false
				break
			}
		}
		switch {
		case allSynced && len(statuses) > 0:
			checks[0].Status = "passed"
		case len(statuses) == 0:
			checks[0].Status = "warning"
			checks[0].Message = "No applications found"
		default:
			checks[0].Status = "warning"
			checks[0].Message = "Some applications are not synced"
		}
	}

	// Check health status
	cmd = exec.CommandContext(ctx, "kubectl", "get", "applications.argoproj.io", "-n", namespace, "-o", "jsonpath={.items[*].status.health.status}")
	output, err = cmd.CombinedOutput()
	if err != nil {
		checks[1].Status = "warning"
		checks[1].Message = "Could not verify health status"
	} else {
		statuses := strings.Fields(string(output))
		allHealthy := true
		for _, status := range statuses {
			if status != "Healthy" {
				allHealthy = false
				break
			}
		}
		switch {
		case allHealthy && len(statuses) > 0:
			checks[1].Status = "passed"
		case len(statuses) == 0:
			checks[1].Status = "passed"
			checks[1].Message = "No applications to check"
		default:
			checks[1].Status = "warning"
			checks[1].Message = "Some resources are degraded"
		}
	}

	v.checks = append(v.checks, checks...)
	return checks
}

// GetAllChecks returns all validation checks.
func (v *Validator) GetAllChecks() []ValidationCheck {
	return v.checks
}

// HasFailures returns true if any checks failed.
func (v *Validator) HasFailures() bool {
	for _, check := range v.checks {
		if check.Status == "failed" {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any checks have warnings.
func (v *Validator) HasWarnings() bool {
	for _, check := range v.checks {
		if check.Status == "warning" {
			return true
		}
	}
	return false
}
