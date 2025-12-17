// Package bootstrap provides GitOps tool installation and configuration.
package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ihsanmokhlisse/gitopsi/internal/cluster"
)

// Tool represents the GitOps tool to bootstrap.
type Tool string

const (
	ToolArgoCD Tool = "argocd"
	ToolFlux   Tool = "flux"
)

// Mode represents the installation mode.
type Mode string

const (
	ModeHelm     Mode = "helm"
	ModeOLM      Mode = "olm"
	ModeManifest Mode = "manifest"
)

// Options holds bootstrap configuration.
type Options struct {
	Tool            Tool
	Mode            Mode
	Namespace       string
	Version         string
	Wait            bool
	Timeout         int
	ConfigureRepo   bool
	RepoURL         string
	RepoBranch      string
	RepoPath        string
	CreateAppOfApps bool
	SyncInitial     bool
	ProjectName     string
}

// Result holds the bootstrap result.
type Result struct {
	Tool      Tool
	URL       string
	Username  string
	Password  string
	Namespace string
	Ready     bool
	Message   string
}

// Bootstrapper handles GitOps tool installation.
type Bootstrapper struct {
	cluster *cluster.Cluster
	options *Options
}

// New creates a new Bootstrapper instance.
func New(c *cluster.Cluster, opts Options) *Bootstrapper {
	if opts.Namespace == "" {
		if opts.Tool == ToolArgoCD {
			opts.Namespace = "argocd"
		} else {
			opts.Namespace = "flux-system"
		}
	}
	if opts.Timeout == 0 {
		opts.Timeout = 300
	}
	return &Bootstrapper{
		cluster: c,
		options: &opts,
	}
}

// Bootstrap installs and configures the GitOps tool.
func (b *Bootstrapper) Bootstrap(ctx context.Context) (*Result, error) {
	result := &Result{
		Tool:      b.options.Tool,
		Namespace: b.options.Namespace,
	}

	// Create namespace
	if err := b.cluster.CreateNamespace(ctx, b.options.Namespace); err != nil {
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	// Install GitOps tool
	switch b.options.Tool {
	case ToolArgoCD:
		if err := b.installArgoCD(ctx); err != nil {
			return nil, fmt.Errorf("failed to install ArgoCD: %w", err)
		}
	case ToolFlux:
		if err := b.installFlux(ctx); err != nil {
			return nil, fmt.Errorf("failed to install Flux: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported GitOps tool: %s", b.options.Tool)
	}

	// Wait for GitOps tool to be ready
	if b.options.Wait {
		if err := b.waitForReady(ctx); err != nil {
			return nil, fmt.Errorf("GitOps tool not ready: %w", err)
		}
	}

	result.Ready = true

	// Configure repository
	if b.options.ConfigureRepo && b.options.RepoURL != "" {
		if err := b.configureRepository(ctx); err != nil {
			return nil, fmt.Errorf("failed to configure repository: %w", err)
		}
	}

	// Create App-of-Apps
	if b.options.CreateAppOfApps {
		if err := b.createAppOfApps(ctx); err != nil {
			return nil, fmt.Errorf("failed to create App-of-Apps: %w", err)
		}
	}

	// Get access info
	if b.options.Tool == ToolArgoCD {
		url, password, err := b.getArgoCDAccess(ctx)
		if err == nil {
			result.URL = url
			result.Username = "admin"
			result.Password = password
		}
	}

	result.Message = fmt.Sprintf("%s installed successfully in namespace %s", b.options.Tool, b.options.Namespace)
	return result, nil
}

// installArgoCD installs ArgoCD using the specified mode.
func (b *Bootstrapper) installArgoCD(ctx context.Context) error {
	switch b.options.Mode {
	case ModeHelm:
		return b.installArgoCDHelm(ctx)
	case ModeManifest:
		return b.installArgoCDManifest(ctx)
	case ModeOLM:
		return b.installArgoCDOLM(ctx)
	default:
		return fmt.Errorf("unsupported installation mode: %s", b.options.Mode)
	}
}

// installArgoCDHelm installs ArgoCD using Helm.
func (b *Bootstrapper) installArgoCDHelm(ctx context.Context) error {
	// Add ArgoCD Helm repo
	cmd := exec.CommandContext(ctx, "helm", "repo", "add", "argo", "https://argoproj.github.io/argo-helm")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore if repo already exists
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("failed to add Helm repo: %w: %s", err, string(output))
		}
	}

	// Update Helm repos
	cmd = exec.CommandContext(ctx, "helm", "repo", "update")
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update Helm repos: %w", err)
	}

	// Install ArgoCD
	args := []string{
		"upgrade", "--install", "argocd", "argo/argo-cd",
		"--namespace", b.options.Namespace,
		"--create-namespace",
		"--wait",
	}

	if b.options.Version != "" {
		args = append(args, "--version", b.options.Version)
	}

	cmd = exec.CommandContext(ctx, "helm", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w: %s", err, string(output))
	}

	return nil
}

// installArgoCDManifest installs ArgoCD using manifests.
func (b *Bootstrapper) installArgoCDManifest(ctx context.Context) error {
	version := b.options.Version
	if version == "" {
		version = "stable"
	}

	manifestURL := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", version)

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-n", b.options.Namespace, "-f", manifestURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply ArgoCD manifests: %w: %s", err, string(output))
	}

	return nil
}

// installArgoCDOLM installs ArgoCD using OLM.
func (b *Bootstrapper) installArgoCDOLM(ctx context.Context) error {
	// Check if OLM is installed
	cmd := exec.CommandContext(ctx, "kubectl", "get", "crd", "subscriptions.operators.coreos.com")
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("OLM not installed on cluster")
	}

	// Create ArgoCD subscription
	subscription := fmt.Sprintf(`apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: argocd-operator
  namespace: %s
spec:
  channel: alpha
  name: argocd-operator
  source: community-operators
  sourceNamespace: openshift-marketplace`, b.options.Namespace)

	return b.cluster.Apply(ctx, subscription)
}

// installFlux installs Flux using the specified mode.
func (b *Bootstrapper) installFlux(ctx context.Context) error {
	switch b.options.Mode {
	case ModeManifest:
		return b.installFluxManifest(ctx)
	case ModeHelm:
		return b.installFluxHelm(ctx)
	default:
		return fmt.Errorf("unsupported installation mode for Flux: %s", b.options.Mode)
	}
}

// installFluxManifest installs Flux using manifests.
func (b *Bootstrapper) installFluxManifest(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "flux", "install", "--namespace", b.options.Namespace)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install Flux: %w: %s", err, string(output))
	}
	return nil
}

// installFluxHelm installs Flux using Helm.
func (b *Bootstrapper) installFluxHelm(ctx context.Context) error {
	// Add Flux Helm repo
	cmd := exec.CommandContext(ctx, "helm", "repo", "add", "fluxcd", "https://fluxcd-community.github.io/helm-charts")
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("failed to add Flux Helm repo: %w: %s", err, string(output))
		}
	}

	// Update Helm repos
	cmd = exec.CommandContext(ctx, "helm", "repo", "update")
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update Helm repos: %w", err)
	}

	// Install Flux
	args := []string{
		"upgrade", "--install", "flux2", "fluxcd/flux2",
		"--namespace", b.options.Namespace,
		"--create-namespace",
		"--wait",
	}

	if b.options.Version != "" {
		args = append(args, "--version", b.options.Version)
	}

	cmd = exec.CommandContext(ctx, "helm", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install Flux: %w: %s", err, string(output))
	}

	return nil
}

// waitForReady waits for the GitOps tool to be ready.
func (b *Bootstrapper) waitForReady(ctx context.Context) error {
	deadline := time.Now().Add(time.Duration(b.options.Timeout) * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if b.options.Tool == ToolArgoCD {
			// Check ArgoCD server deployment
			err := b.cluster.WaitForDeployment(ctx, b.options.Namespace, "argocd-server", 30)
			if err == nil {
				return nil
			}
		} else {
			// Check Flux controllers
			err := b.cluster.WaitForDeployment(ctx, b.options.Namespace, "source-controller", 30)
			if err == nil {
				return nil
			}
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for %s to be ready", b.options.Tool)
}

// configureRepository adds the repository to the GitOps tool.
func (b *Bootstrapper) configureRepository(ctx context.Context) error {
	if b.options.Tool == ToolArgoCD {
		return b.configureArgoCDRepo(ctx)
	}
	return b.configureFluxRepo(ctx)
}

// configureArgoCDRepo adds a repository to ArgoCD.
func (b *Bootstrapper) configureArgoCDRepo(ctx context.Context) error {
	repoSecret := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: repo-%s
  namespace: %s
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: git
  url: %s`, b.options.ProjectName, b.options.Namespace, b.options.RepoURL)

	return b.cluster.Apply(ctx, repoSecret)
}

// configureFluxRepo adds a repository to Flux.
func (b *Bootstrapper) configureFluxRepo(ctx context.Context) error {
	branch := b.options.RepoBranch
	if branch == "" {
		branch = "main"
	}

	gitRepo := fmt.Sprintf(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: %s
  namespace: %s
spec:
  interval: 1m
  url: %s
  ref:
    branch: %s`, b.options.ProjectName, b.options.Namespace, b.options.RepoURL, branch)

	return b.cluster.Apply(ctx, gitRepo)
}

// createAppOfApps creates the root application.
func (b *Bootstrapper) createAppOfApps(ctx context.Context) error {
	if b.options.Tool == ToolArgoCD {
		return b.createArgoCDAppOfApps(ctx)
	}
	return b.createFluxKustomization(ctx)
}

// createArgoCDAppOfApps creates an ArgoCD Application.
func (b *Bootstrapper) createArgoCDAppOfApps(ctx context.Context) error {
	path := b.options.RepoPath
	if path == "" {
		path = "argocd/applicationsets"
	}

	branch := b.options.RepoBranch
	if branch == "" {
		branch = "main"
	}

	app := fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s-root
  namespace: %s
spec:
  project: default
  source:
    repoURL: %s
    targetRevision: %s
    path: %s
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
  syncPolicy:
    automated:
      prune: true
      selfHeal: true`, b.options.ProjectName, b.options.Namespace, b.options.RepoURL, branch, path, b.options.Namespace)

	return b.cluster.Apply(ctx, app)
}

// createFluxKustomization creates a Flux Kustomization.
func (b *Bootstrapper) createFluxKustomization(ctx context.Context) error {
	path := b.options.RepoPath
	if path == "" {
		path = "./"
	}

	kustomization := fmt.Sprintf(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: %s
  namespace: %s
spec:
  interval: 10m
  sourceRef:
    kind: GitRepository
    name: %s
  path: %s
  prune: true`, b.options.ProjectName, b.options.Namespace, b.options.ProjectName, path)

	return b.cluster.Apply(ctx, kustomization)
}

// getArgoCDAccess gets the ArgoCD UI URL and initial admin password.
func (b *Bootstrapper) getArgoCDAccess(ctx context.Context) (string, string, error) {
	// Get password from secret
	output, err := b.cluster.RunCommand(ctx, "get", "secret", "argocd-initial-admin-secret",
		"-n", b.options.Namespace,
		"-o", "jsonpath={.data.password}")
	if err != nil {
		return "", "", err
	}

	// Decode base64 password
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("echo %s | base64 -d", strings.TrimSpace(output)))
	passwordBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to decode password: %w", err)
	}

	// Try to get the ArgoCD server URL (assuming LoadBalancer or use port-forward)
	url := fmt.Sprintf("https://argocd-server.%s.svc.cluster.local", b.options.Namespace)

	// Try to get external URL if available
	extOutput, err := b.cluster.RunCommand(ctx, "get", "svc", "argocd-server",
		"-n", b.options.Namespace,
		"-o", "jsonpath={.status.loadBalancer.ingress[0].hostname}")
	if err == nil && strings.TrimSpace(extOutput) != "" {
		url = fmt.Sprintf("https://%s", strings.TrimSpace(extOutput))
	}

	return url, strings.TrimSpace(string(passwordBytes)), nil
}

// Uninstall removes the GitOps tool from the cluster.
func (b *Bootstrapper) Uninstall(ctx context.Context) error {
	switch b.options.Tool {
	case ToolArgoCD:
		switch b.options.Mode {
		case ModeHelm:
			cmd := exec.CommandContext(ctx, "helm", "uninstall", "argocd", "-n", b.options.Namespace)
			if _, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to uninstall ArgoCD: %w", err)
			}
		case ModeManifest:
			version := b.options.Version
			if version == "" {
				version = "stable"
			}
			manifestURL := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", version)
			cmd := exec.CommandContext(ctx, "kubectl", "delete", "-n", b.options.Namespace, "-f", manifestURL)
			if _, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to delete ArgoCD manifests: %w", err)
			}
		}

	case ToolFlux:
		cmd := exec.CommandContext(ctx, "flux", "uninstall", "--namespace", b.options.Namespace, "--silent")
		if _, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to uninstall Flux: %w", err)
		}
	}

	// Delete namespace
	_, _ = b.cluster.RunCommand(ctx, "delete", "namespace", b.options.Namespace)

	return nil
}

// GetTool returns the configured GitOps tool.
func (b *Bootstrapper) GetTool() Tool {
	return b.options.Tool
}

// GetMode returns the configured installation mode.
func (b *Bootstrapper) GetMode() Mode {
	return b.options.Mode
}

// GetNamespace returns the configured namespace.
func (b *Bootstrapper) GetNamespace() string {
	return b.options.Namespace
}

// GetEnvFromToken retrieves a token from an environment variable.
func GetEnvFromToken(envVar string) string {
	return os.Getenv(envVar)
}

