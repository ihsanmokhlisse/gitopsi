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
	ModeHelm      Mode = "helm"
	ModeOLM       Mode = "olm"
	ModeManifest  Mode = "manifest"
	ModeKustomize Mode = "kustomize"
)

// HelmConfig holds Helm-specific configuration.
type HelmConfig struct {
	Repo      string            `yaml:"repo"`
	Chart     string            `yaml:"chart"`
	Version   string            `yaml:"version"`
	Namespace string            `yaml:"namespace"`
	Values    map[string]any    `yaml:"values"`
	SetValues map[string]string `yaml:"set_values"`
}

// OLMConfig holds OLM-specific configuration.
type OLMConfig struct {
	Channel         string `yaml:"channel"`
	Source          string `yaml:"source"`
	SourceNamespace string `yaml:"source_namespace"`
	Approval        string `yaml:"approval"`
}

// ManifestConfig holds manifest-specific configuration.
type ManifestConfig struct {
	URL       string   `yaml:"url"`
	Paths     []string `yaml:"paths"`
	Namespace string   `yaml:"namespace"`
}

// KustomizeConfig holds Kustomize-specific configuration.
type KustomizeConfig struct {
	URL     string   `yaml:"url"`
	Path    string   `yaml:"path"`
	Patches []string `yaml:"patches"`
}

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

	// Generated manifest paths for auto-apply
	GeneratedPath        string
	ApplyGeneratedConfig bool
	WaitForSync          bool
	SyncTimeout          int

	// Mode-specific configurations
	Helm      *HelmConfig      `yaml:"helm,omitempty"`
	OLM       *OLMConfig       `yaml:"olm,omitempty"`
	Manifest  *ManifestConfig  `yaml:"manifest,omitempty"`
	Kustomize *KustomizeConfig `yaml:"kustomize,omitempty"`
}

// Result holds the bootstrap result.
type Result struct {
	Tool             Tool
	URL              string
	Username         string
	Password         string
	Namespace        string
	Ready            bool
	Message          string
	ExistingInstall  bool
	AppliedManifests []string
	SyncedApps       []SyncStatus
}

// SyncStatus represents the sync status of an application.
type SyncStatus struct {
	Name       string
	Status     string
	Health     string
	SyncStatus string
	Message    string
}

// Bootstrapper handles GitOps tool installation.
type Bootstrapper struct {
	cluster *cluster.Cluster
	options *Options
}

// New creates a new Bootstrapper instance.
func New(c *cluster.Cluster, opts *Options) *Bootstrapper {
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
		options: opts,
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

// DetectExistingArgoCD checks if ArgoCD is already installed and returns detection result.
func (b *Bootstrapper) DetectExistingArgoCD(ctx context.Context) (*ArgoCDDetectionResult, error) {
	timeout := time.Duration(b.options.Timeout) * time.Second
	detector := NewDetector(b.cluster.Name, timeout)
	return detector.DetectArgoCD(ctx)
}

// BootstrapOrDetect either bootstraps ArgoCD or detects existing installation.
func (b *Bootstrapper) BootstrapOrDetect(ctx context.Context) (*Result, error) {
	result := &Result{
		Tool:      b.options.Tool,
		Namespace: b.options.Namespace,
	}

	if b.options.Tool == ToolArgoCD {
		detection, err := b.DetectExistingArgoCD(ctx)
		if err == nil && detection.Installed && detection.Running {
			result.ExistingInstall = true
			result.Ready = true
			result.Namespace = detection.Namespace
			result.URL = detection.URL
			result.Username = "admin"
			result.Message = fmt.Sprintf("Existing %s ArgoCD detected in namespace %s", detection.Type, detection.Namespace)
			return result, nil
		}
	}

	return b.Bootstrap(ctx)
}

// ApplyGeneratedManifests applies the generated ArgoCD manifests from the output directory.
func (b *Bootstrapper) ApplyGeneratedManifests(ctx context.Context, projectPath string) ([]string, error) {
	if b.options.Tool != ToolArgoCD {
		return nil, nil
	}

	var appliedManifests []string
	argoCDPath := fmt.Sprintf("%s/argocd", projectPath)

	if _, err := os.Stat(argoCDPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("argocd directory not found at %s", argoCDPath)
	}

	projectsPath := fmt.Sprintf("%s/projects", argoCDPath)
	if _, err := os.Stat(projectsPath); err == nil {
		files, err := os.ReadDir(projectsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read projects directory: %w", err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
				continue
			}
			filePath := fmt.Sprintf("%s/%s", projectsPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", file.Name(), err)
			}
			if err := b.cluster.Apply(ctx, string(content)); err != nil {
				return nil, fmt.Errorf("failed to apply %s: %w", file.Name(), err)
			}
			appliedManifests = append(appliedManifests, file.Name())
		}
	}

	applicationsPath := fmt.Sprintf("%s/applications", argoCDPath)
	if _, err := os.Stat(applicationsPath); err == nil {
		files, err := os.ReadDir(applicationsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read applications directory: %w", err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
				continue
			}
			filePath := fmt.Sprintf("%s/%s", applicationsPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", file.Name(), err)
			}
			if err := b.cluster.Apply(ctx, string(content)); err != nil {
				return nil, fmt.Errorf("failed to apply %s: %w", file.Name(), err)
			}
			appliedManifests = append(appliedManifests, file.Name())
		}
	}

	appsetsPath := fmt.Sprintf("%s/applicationsets", argoCDPath)
	if _, err := os.Stat(appsetsPath); err == nil {
		files, err := os.ReadDir(appsetsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read applicationsets directory: %w", err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
				continue
			}
			filePath := fmt.Sprintf("%s/%s", appsetsPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", file.Name(), err)
			}
			if err := b.cluster.Apply(ctx, string(content)); err != nil {
				return nil, fmt.Errorf("failed to apply %s: %w", file.Name(), err)
			}
			appliedManifests = append(appliedManifests, file.Name())
		}
	}

	return appliedManifests, nil
}

// WaitForAppSync waits for ArgoCD applications to sync.
func (b *Bootstrapper) WaitForAppSync(ctx context.Context, appNames []string) ([]SyncStatus, error) {
	if b.options.Tool != ToolArgoCD {
		return nil, nil
	}

	timeout := b.options.SyncTimeout
	if timeout == 0 {
		timeout = 300
	}

	syncStatuses := make([]SyncStatus, 0, len(appNames))
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for _, appName := range appNames {
		status := SyncStatus{Name: appName}
		for time.Now().Before(deadline) {
			cmd := exec.CommandContext(ctx, "kubectl", "get", "application", appName,
				"-n", b.options.Namespace, "-o", "jsonpath={.status.sync.status},{.status.health.status}")
			output, err := cmd.Output()
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			parts := strings.Split(string(output), ",")
			if len(parts) >= 2 {
				status.SyncStatus = parts[0]
				status.Health = parts[1]
				if parts[0] == "Synced" && (parts[1] == "Healthy" || parts[1] == "Progressing") {
					status.Status = "ready"
					break
				}
			}
			time.Sleep(5 * time.Second)
		}
		if status.Status == "" {
			status.Status = "timeout"
			status.Message = "Sync did not complete within timeout"
		}
		syncStatuses = append(syncStatuses, status)
	}

	return syncStatuses, nil
}

// GetApplicationStatus returns the status of an ArgoCD application.
func (b *Bootstrapper) GetApplicationStatus(ctx context.Context, appName string) (*SyncStatus, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "get", "application", appName,
		"-n", b.options.Namespace, "-o", "jsonpath={.status.sync.status},{.status.health.status},{.status.operationState.message}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get application status: %w", err)
	}

	parts := strings.Split(string(output), ",")
	status := &SyncStatus{Name: appName}
	if len(parts) >= 1 {
		status.SyncStatus = parts[0]
	}
	if len(parts) >= 2 {
		status.Health = parts[1]
	}
	if len(parts) >= 3 {
		status.Message = parts[2]
	}
	status.Status = "active"
	return status, nil
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
	case ModeKustomize:
		return b.installArgoCDKustomize(ctx)
	default:
		return fmt.Errorf("unsupported installation mode: %s", b.options.Mode)
	}
}

// installArgoCDHelm installs ArgoCD using Helm.
func (b *Bootstrapper) installArgoCDHelm(ctx context.Context) error {
	helmCfg := b.getArgoCDHelmConfig()

	// Add ArgoCD Helm repo
	cmd := exec.CommandContext(ctx, "helm", "repo", "add", "argo", helmCfg.Repo)
	if output, err := cmd.CombinedOutput(); err != nil {
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
		"upgrade", "--install", "argocd", fmt.Sprintf("argo/%s", helmCfg.Chart),
		"--namespace", b.options.Namespace,
		"--create-namespace",
		"--wait",
	}

	if helmCfg.Version != "" {
		args = append(args, "--version", helmCfg.Version)
	} else if b.options.Version != "" {
		args = append(args, "--version", b.options.Version)
	}

	// Add set values
	for k, v := range helmCfg.SetValues {
		args = append(args, "--set", fmt.Sprintf("%s=%s", k, v))
	}

	cmd = exec.CommandContext(ctx, "helm", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w: %s", err, string(output))
	}

	return nil
}

// installArgoCDManifest installs ArgoCD using manifests.
func (b *Bootstrapper) installArgoCDManifest(ctx context.Context) error {
	manifestCfg := b.getArgoCDManifestConfig()

	manifestURL := manifestCfg.URL
	if manifestURL == "" {
		version := b.options.Version
		if version == "" {
			version = "stable"
		}
		manifestURL = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", version)
	}

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-n", b.options.Namespace, "-f", manifestURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply ArgoCD manifests: %w: %s", err, string(output))
	}

	// Apply additional manifests if specified
	for _, path := range manifestCfg.Paths {
		cmd = exec.CommandContext(ctx, "kubectl", "apply", "-n", b.options.Namespace, "-f", path)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to apply manifest %s: %w: %s", path, err, string(output))
		}
	}

	return nil
}

// installArgoCDOLM installs ArgoCD using OLM.
func (b *Bootstrapper) installArgoCDOLM(ctx context.Context) error {
	olmCfg := b.getArgoCDOLMConfig()

	// Check if OLM is installed
	cmd := exec.CommandContext(ctx, "kubectl", "get", "crd", "subscriptions.operators.coreos.com")
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("OLM not installed on cluster. OLM is required for this installation mode")
	}

	// Create OperatorGroup
	operatorGroup := fmt.Sprintf(`apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: argocd-operator
  namespace: %s
spec:
  targetNamespaces:
    - %s`, b.options.Namespace, b.options.Namespace)

	if err := b.cluster.Apply(ctx, operatorGroup); err != nil {
		return fmt.Errorf("failed to create OperatorGroup: %w", err)
	}

	// Create ArgoCD subscription
	subscription := fmt.Sprintf(`apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: argocd-operator
  namespace: %s
spec:
  channel: %s
  name: argocd-operator
  source: %s
  sourceNamespace: %s
  installPlanApproval: %s`, b.options.Namespace, olmCfg.Channel, olmCfg.Source, olmCfg.SourceNamespace, olmCfg.Approval)

	return b.cluster.Apply(ctx, subscription)
}

// installFlux installs Flux using the specified mode.
func (b *Bootstrapper) installFlux(ctx context.Context) error {
	switch b.options.Mode {
	case ModeManifest:
		return b.installFluxManifest(ctx)
	case ModeHelm:
		return b.installFluxHelm(ctx)
	case ModeKustomize:
		return b.installFluxKustomize(ctx)
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
	helmCfg := b.getFluxHelmConfig()

	// Add Flux Helm repo
	cmd := exec.CommandContext(ctx, "helm", "repo", "add", "fluxcd", helmCfg.Repo)
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
		"upgrade", "--install", "flux2", fmt.Sprintf("fluxcd/%s", helmCfg.Chart),
		"--namespace", b.options.Namespace,
		"--create-namespace",
		"--wait",
	}

	if helmCfg.Version != "" {
		args = append(args, "--version", helmCfg.Version)
	} else if b.options.Version != "" {
		args = append(args, "--version", b.options.Version)
	}

	// Add set values
	for k, v := range helmCfg.SetValues {
		args = append(args, "--set", fmt.Sprintf("%s=%s", k, v))
	}

	cmd = exec.CommandContext(ctx, "helm", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install Flux: %w: %s", err, string(output))
	}

	return nil
}

// installArgoCDKustomize installs ArgoCD using Kustomize.
func (b *Bootstrapper) installArgoCDKustomize(ctx context.Context) error {
	kustomizeCfg := b.getArgoCDKustomizeConfig()

	kustomizeURL := kustomizeCfg.URL
	if kustomizeURL == "" {
		kustomizeURL = "https://github.com/argoproj/argo-cd/manifests/cluster-install"
		if kustomizeCfg.Path != "" {
			kustomizeURL = fmt.Sprintf("https://github.com/argoproj/argo-cd/manifests/%s", kustomizeCfg.Path)
		}
	}

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-k", kustomizeURL, "-n", b.options.Namespace)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply ArgoCD Kustomize: %w: %s", err, string(output))
	}

	return nil
}

// installFluxKustomize installs Flux using Kustomize.
func (b *Bootstrapper) installFluxKustomize(ctx context.Context) error {
	kustomizeCfg := b.getFluxKustomizeConfig()

	kustomizeURL := kustomizeCfg.URL
	if kustomizeURL == "" {
		kustomizeURL = "https://github.com/fluxcd/flux2/manifests/install"
		if kustomizeCfg.Path != "" {
			kustomizeURL = fmt.Sprintf("https://github.com/fluxcd/flux2/manifests/%s", kustomizeCfg.Path)
		}
	}

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-k", kustomizeURL, "-n", b.options.Namespace)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply Flux Kustomize: %w: %s", err, string(output))
	}

	return nil
}

// getArgoCDHelmConfig returns the ArgoCD Helm configuration with defaults.
func (b *Bootstrapper) getArgoCDHelmConfig() *HelmConfig {
	if b.options.Helm != nil {
		cfg := *b.options.Helm
		if cfg.Repo == "" {
			cfg.Repo = "https://argoproj.github.io/argo-helm"
		}
		if cfg.Chart == "" {
			cfg.Chart = "argo-cd"
		}
		return &cfg
	}
	return &HelmConfig{
		Repo:    "https://argoproj.github.io/argo-helm",
		Chart:   "argo-cd",
		Version: b.options.Version,
	}
}

// getFluxHelmConfig returns the Flux Helm configuration with defaults.
func (b *Bootstrapper) getFluxHelmConfig() *HelmConfig {
	if b.options.Helm != nil {
		cfg := *b.options.Helm
		if cfg.Repo == "" {
			cfg.Repo = "https://fluxcd-community.github.io/helm-charts"
		}
		if cfg.Chart == "" {
			cfg.Chart = "flux2"
		}
		return &cfg
	}
	return &HelmConfig{
		Repo:    "https://fluxcd-community.github.io/helm-charts",
		Chart:   "flux2",
		Version: b.options.Version,
	}
}

// getArgoCDOLMConfig returns the ArgoCD OLM configuration with defaults.
func (b *Bootstrapper) getArgoCDOLMConfig() *OLMConfig {
	if b.options.OLM != nil {
		cfg := *b.options.OLM
		if cfg.Channel == "" {
			cfg.Channel = "alpha"
		}
		if cfg.Source == "" {
			cfg.Source = "community-operators"
		}
		if cfg.SourceNamespace == "" {
			cfg.SourceNamespace = "openshift-marketplace"
		}
		if cfg.Approval == "" {
			cfg.Approval = "Automatic"
		}
		return &cfg
	}
	return &OLMConfig{
		Channel:         "alpha",
		Source:          "community-operators",
		SourceNamespace: "openshift-marketplace",
		Approval:        "Automatic",
	}
}

// getArgoCDManifestConfig returns the ArgoCD manifest configuration with defaults.
func (b *Bootstrapper) getArgoCDManifestConfig() *ManifestConfig {
	if b.options.Manifest != nil {
		return b.options.Manifest
	}
	version := b.options.Version
	if version == "" {
		version = "stable"
	}
	return &ManifestConfig{
		URL: fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", version),
	}
}

// getArgoCDKustomizeConfig returns the ArgoCD Kustomize configuration with defaults.
func (b *Bootstrapper) getArgoCDKustomizeConfig() *KustomizeConfig {
	if b.options.Kustomize != nil {
		return b.options.Kustomize
	}
	return &KustomizeConfig{
		URL:  "https://github.com/argoproj/argo-cd/manifests/cluster-install",
		Path: "cluster-install",
	}
}

// getFluxKustomizeConfig returns the Flux Kustomize configuration with defaults.
func (b *Bootstrapper) getFluxKustomizeConfig() *KustomizeConfig {
	if b.options.Kustomize != nil {
		return b.options.Kustomize
	}
	return &KustomizeConfig{
		URL:  "https://github.com/fluxcd/flux2/manifests/install",
		Path: "install",
	}
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
func (b *Bootstrapper) getArgoCDAccess(ctx context.Context) (accessURL, accessPassword string, accessErr error) {
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
	accessURL = fmt.Sprintf("https://argocd-server.%s.svc.cluster.local", b.options.Namespace)

	// Try to get external URL if available
	extOutput, err := b.cluster.RunCommand(ctx, "get", "svc", "argocd-server",
		"-n", b.options.Namespace,
		"-o", "jsonpath={.status.loadBalancer.ingress[0].hostname}")
	if err == nil && strings.TrimSpace(extOutput) != "" {
		accessURL = fmt.Sprintf("https://%s", strings.TrimSpace(extOutput))
	}

	accessPassword = strings.TrimSpace(string(passwordBytes))
	return accessURL, accessPassword, nil
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

	// Delete namespace (ignore error as it may not exist)
	if _, delErr := b.cluster.RunCommand(ctx, "delete", "namespace", b.options.Namespace); delErr != nil {
		// Namespace deletion is best-effort, log but don't fail
		_ = delErr
	}

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

// ValidModes returns valid bootstrap modes for a tool on a platform.
func ValidModes(tool Tool, platform string) []Mode {
	modes := []Mode{ModeHelm, ModeManifest, ModeKustomize}

	// OLM is only available on OpenShift
	if platform == "openshift" && tool == ToolArgoCD {
		modes = append(modes, ModeOLM)
	}

	return modes
}

// SuggestMode suggests the best bootstrap mode for a given platform and tool.
func SuggestMode(tool Tool, platform string) Mode {
	switch platform {
	case "openshift":
		if tool == ToolArgoCD {
			return ModeOLM
		}
		return ModeHelm
	case "eks", "aks", "gke":
		return ModeHelm
	default:
		return ModeHelm
	}
}

// IsValidMode checks if a mode is valid for a given tool and platform.
func IsValidMode(mode Mode, tool Tool, platform string) bool {
	validModes := ValidModes(tool, platform)
	for _, m := range validModes {
		if m == mode {
			return true
		}
	}
	return false
}

// ModeDescription returns a human-readable description of a bootstrap mode.
func ModeDescription(mode Mode) string {
	switch mode {
	case ModeHelm:
		return "Helm chart - Official Helm charts with full customization support"
	case ModeOLM:
		return "Operator Lifecycle Manager - Managed installation via OpenShift OperatorHub"
	case ModeManifest:
		return "Raw manifests - Direct YAML manifests for air-gapped or custom setups"
	case ModeKustomize:
		return "Kustomize - Official Kustomize installations with overlay support"
	default:
		return string(mode)
	}
}

// DefaultHelmConfig returns default Helm configuration for a tool.
func DefaultHelmConfig(tool Tool) *HelmConfig {
	switch tool {
	case ToolArgoCD:
		return &HelmConfig{
			Repo:  "https://argoproj.github.io/argo-helm",
			Chart: "argo-cd",
		}
	case ToolFlux:
		return &HelmConfig{
			Repo:  "https://fluxcd-community.github.io/helm-charts",
			Chart: "flux2",
		}
	default:
		return nil
	}
}

// DefaultOLMConfig returns default OLM configuration for ArgoCD.
func DefaultOLMConfig() *OLMConfig {
	return &OLMConfig{
		Channel:         "alpha",
		Source:          "community-operators",
		SourceNamespace: "openshift-marketplace",
		Approval:        "Automatic",
	}
}
