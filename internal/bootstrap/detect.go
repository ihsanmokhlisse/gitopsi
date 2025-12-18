// Package bootstrap provides GitOps tool installation and configuration.
package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ArgoCDType string

const (
	ArgoCDTypeCommunity    ArgoCDType = "community"
	ArgoCDTypeRedHat       ArgoCDType = "redhat"
	ArgoCDTypeUnknown      ArgoCDType = "unknown"
	ArgoCDTypeNotInstalled ArgoCDType = "not_installed"
)

// ArgoCDState represents the installation state of ArgoCD
type ArgoCDState string

const (
	ArgoCDStateNotInstalled   ArgoCDState = "not_installed"   // No namespace found
	ArgoCDStateNamespaceOnly  ArgoCDState = "namespace_only"  // Namespace exists, no workloads
	ArgoCDStatePartialInstall ArgoCDState = "partial_install" // Some components missing
	ArgoCDStateNotRunning     ArgoCDState = "not_running"     // All components exist, not ready
	ArgoCDStateRunning        ArgoCDState = "running"         // Fully operational
)

type InstallMethod string

const (
	InstallMethodOperator InstallMethod = "operator"
	InstallMethodManifest InstallMethod = "manifest"
	InstallMethodHelm     InstallMethod = "helm"
	InstallMethodOLM      InstallMethod = "olm"
	InstallMethodUnknown  InstallMethod = "unknown"
)

type OperatorSource string

const (
	OperatorSourceRedHat      OperatorSource = "redhat-operators"
	OperatorSourceCommunity   OperatorSource = "community-operators"
	OperatorSourceCertified   OperatorSource = "certified-operators"
	OperatorSourceMarketplace OperatorSource = "redhat-marketplace"
	OperatorSourceUnknown     OperatorSource = "unknown"
)

type ComponentStatus string

const (
	StatusRunning    ComponentStatus = "running"
	StatusNotRunning ComponentStatus = "not_running"
	StatusDegraded   ComponentStatus = "degraded"
	StatusUnknown    ComponentStatus = "unknown"
)

type ArgoCDComponent struct {
	Name      string          `json:"name"`
	Ready     bool            `json:"ready"`
	Replicas  int             `json:"replicas"`
	Available int             `json:"available"`
	Status    ComponentStatus `json:"status"`
	Image     string          `json:"image,omitempty"`
}

type ArgoCDDetectionResult struct {
	Installed       bool              `json:"installed"`
	State           ArgoCDState       `json:"state"`
	StateMessage    string            `json:"state_message"`
	Type            ArgoCDType        `json:"type"`
	InstallMethod   InstallMethod     `json:"install_method"`
	OperatorSource  OperatorSource    `json:"operator_source,omitempty"`
	Namespace       string            `json:"namespace"`
	Version         string            `json:"version,omitempty"`
	URL             string            `json:"url,omitempty"`
	Running         bool              `json:"running"`
	Components      []ArgoCDComponent `json:"components"`
	TotalComponents int               `json:"total_components"`
	ReadyComponents int               `json:"ready_components"`
	HealthStatus    string            `json:"health_status"`
	SyncStatus      string            `json:"sync_status,omitempty"`
	AppCount        int               `json:"app_count"`
	DetectedAt      time.Time         `json:"detected_at"`
	Issues          []string          `json:"issues,omitempty"`
	Recommendations []string          `json:"recommendations,omitempty"`
}

type Detector struct {
	kubeContext string
	timeout     time.Duration
}

func NewDetector(kubeContext string, timeout time.Duration) *Detector {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Detector{
		kubeContext: kubeContext,
		timeout:     timeout,
	}
}

func (d *Detector) DetectArgoCD(ctx context.Context) (*ArgoCDDetectionResult, error) {
	result := &ArgoCDDetectionResult{
		DetectedAt:      time.Now(),
		Components:      []ArgoCDComponent{},
		Issues:          []string{},
		Recommendations: []string{},
	}

	namespaces := []string{"openshift-gitops", "argocd", "gitops"}
	for _, ns := range namespaces {
		if d.namespaceExists(ctx, ns) {
			result.Namespace = ns
			break
		}
	}

	if result.Namespace == "" {
		result.Installed = false
		result.State = ArgoCDStateNotInstalled
		result.StateMessage = "ArgoCD is not installed - no known namespace found"
		result.Type = ArgoCDTypeNotInstalled
		result.HealthStatus = "not_installed"
		result.Issues = append(result.Issues, "ArgoCD is not installed - no known namespace found")
		result.Recommendations = append(result.Recommendations, "Use 'gitopsi init --bootstrap' to install ArgoCD")
		result.Recommendations = append(result.Recommendations, "Or install OpenShift GitOps operator from OperatorHub")
		return result, nil
	}

	// Namespace exists, detect components
	result.Components = d.detectComponents(ctx, result.Namespace)
	result.TotalComponents = len(result.Components)
	result.ReadyComponents = d.countReadyComponents(result.Components)

	// Determine state based on components
	result.State, result.StateMessage = d.determineState(result)

	if result.State == ArgoCDStateNamespaceOnly {
		result.Installed = false
		result.Type = ArgoCDTypeUnknown
		result.InstallMethod = InstallMethodUnknown
		result.HealthStatus = "namespace_only"
		result.Issues = append(result.Issues, "Namespace exists but no ArgoCD components found")
		result.Recommendations = append(result.Recommendations, "Install ArgoCD using: gitopsi init --bootstrap")
		result.Recommendations = append(result.Recommendations, "Or install OpenShift GitOps operator from OperatorHub")
		return result, nil
	}

	result.Installed = true
	result.Type = d.detectArgoCDType(ctx, result.Namespace)
	result.InstallMethod = d.detectInstallMethod(ctx, result.Namespace)

	if result.InstallMethod == InstallMethodOLM || result.InstallMethod == InstallMethodOperator {
		result.OperatorSource = d.detectOperatorSource(ctx, result.Namespace)
	}

	result.Version = d.detectVersion(ctx, result.Namespace)
	result.URL = d.detectURL(ctx, result.Namespace)
	result.Running = d.isRunning(result.Components)
	result.HealthStatus = d.determineHealthStatus(result.Components)
	result.AppCount = d.countApplications(ctx, result.Namespace)

	d.analyzeAndRecommend(result)

	return result, nil
}

func (d *Detector) countReadyComponents(components []ArgoCDComponent) int {
	count := 0
	for _, c := range components {
		if c.Ready {
			count++
		}
	}
	return count
}

func (d *Detector) determineState(result *ArgoCDDetectionResult) (ArgoCDState, string) {
	if len(result.Components) == 0 {
		return ArgoCDStateNamespaceOnly, fmt.Sprintf("Namespace '%s' exists but no ArgoCD components found", result.Namespace)
	}

	readyCount := result.ReadyComponents
	totalCount := result.TotalComponents

	// Check for expected minimum components (server, repo-server, application-controller)
	expectedComponents := 3
	if totalCount < expectedComponents {
		return ArgoCDStatePartialInstall, fmt.Sprintf("Only %d/%d expected components found", totalCount, expectedComponents)
	}

	if readyCount == 0 {
		return ArgoCDStateNotRunning, fmt.Sprintf("All %d components exist but none are running", totalCount)
	}

	if readyCount < totalCount {
		return ArgoCDStatePartialInstall, fmt.Sprintf("%d/%d components ready", readyCount, totalCount)
	}

	return ArgoCDStateRunning, fmt.Sprintf("All %d components running", totalCount)
}

func (d *Detector) namespaceExists(ctx context.Context, namespace string) bool {
	args := []string{"get", "namespace", namespace, "--ignore-not-found", "-o", "name"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

func (d *Detector) detectArgoCDType(ctx context.Context, namespace string) ArgoCDType {
	if namespace == "openshift-gitops" {
		return ArgoCDTypeRedHat
	}

	args := []string{"get", "deployments", "-n", namespace, "-o", "jsonpath={.items[*].spec.template.spec.containers[*].image}"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return ArgoCDTypeUnknown
	}

	images := string(output)
	if strings.Contains(images, "registry.redhat.io") || strings.Contains(images, "quay.io/openshift-gitops") {
		return ArgoCDTypeRedHat
	}
	if strings.Contains(images, "quay.io/argoproj") || strings.Contains(images, "ghcr.io/argoproj") {
		return ArgoCDTypeCommunity
	}

	return ArgoCDTypeUnknown
}

func (d *Detector) detectInstallMethod(ctx context.Context, namespace string) InstallMethod {
	args := []string{"get", "subscription", "-n", namespace, "--ignore-not-found", "-o", "name"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx1, cancel1 := context.WithTimeout(ctx, d.timeout)
	defer cancel1()
	cmd := exec.CommandContext(ctx1, "kubectl", args...)
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		return InstallMethodOLM
	}

	args = []string{"get", "subscription", "-n", "openshift-operators", "--ignore-not-found", "-o", "name"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx2, cancel2 := context.WithTimeout(ctx, d.timeout)
	defer cancel2()
	cmd = exec.CommandContext(ctx2, "kubectl", args...)
	output, err = cmd.Output()
	if err == nil && (strings.Contains(string(output), "gitops") || strings.Contains(string(output), "argocd")) {
		return InstallMethodOLM
	}

	args = []string{"get", "deployment", "-n", namespace, "-l", "helm.sh/chart", "--ignore-not-found", "-o", "name"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx3, cancel3 := context.WithTimeout(ctx, d.timeout)
	defer cancel3()
	cmd = exec.CommandContext(ctx3, "kubectl", args...)
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		return InstallMethodHelm
	}

	args = []string{"get", "argocd", "-n", namespace, "--ignore-not-found", "-o", "name"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx4, cancel4 := context.WithTimeout(ctx, d.timeout)
	defer cancel4()
	cmd = exec.CommandContext(ctx4, "kubectl", args...)
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		return InstallMethodOperator
	}

	return InstallMethodManifest
}

func (d *Detector) detectOperatorSource(ctx context.Context, namespace string) OperatorSource {
	args := []string{"get", "subscription", "-A", "-o", "jsonpath={.items[?(@.spec.name==\"openshift-gitops-operator\")].spec.source}"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return OperatorSourceUnknown
	}

	source := strings.TrimSpace(string(output))
	switch source {
	case "redhat-operators":
		return OperatorSourceRedHat
	case "community-operators":
		return OperatorSourceCommunity
	case "certified-operators":
		return OperatorSourceCertified
	case "redhat-marketplace":
		return OperatorSourceMarketplace
	default:
		return OperatorSourceUnknown
	}
}

func (d *Detector) detectVersion(ctx context.Context, namespace string) string {
	args := []string{"get", "deployment", "-n", namespace, "-l", "app.kubernetes.io/name=argocd-server",
		"-o", "jsonpath={.items[0].spec.template.spec.containers[0].image}"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	image := string(output)
	parts := strings.Split(image, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func (d *Detector) detectURL(ctx context.Context, namespace string) string {
	args := []string{"get", "route", "-n", namespace, "-l", "app.kubernetes.io/name=openshift-gitops-server",
		"-o", "jsonpath={.items[0].spec.host}"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx1, cancel1 := context.WithTimeout(ctx, d.timeout)
	defer cancel1()
	cmd := exec.CommandContext(ctx1, "kubectl", args...)
	output, err := cmd.Output()
	if err == nil {
		if host := strings.TrimSpace(string(output)); host != "" {
			return "https://" + host
		}
	}

	args = []string{"get", "route", "-n", namespace, "-o", "jsonpath={.items[0].spec.host}"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx2, cancel2 := context.WithTimeout(ctx, d.timeout)
	defer cancel2()
	cmd = exec.CommandContext(ctx2, "kubectl", args...)
	output, err = cmd.Output()
	if err == nil {
		if host := strings.TrimSpace(string(output)); host != "" {
			return "https://" + host
		}
	}

	args = []string{"get", "ingress", "-n", namespace, "-o", "jsonpath={.items[0].spec.rules[0].host}"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx3, cancel3 := context.WithTimeout(ctx, d.timeout)
	defer cancel3()
	cmd = exec.CommandContext(ctx3, "kubectl", args...)
	output, err = cmd.Output()
	if err == nil {
		if host := strings.TrimSpace(string(output)); host != "" {
			return "https://" + host
		}
	}

	return ""
}

func (d *Detector) detectComponents(ctx context.Context, namespace string) []ArgoCDComponent {
	components := []ArgoCDComponent{}
	componentNames := []string{
		"argocd-server",
		"argocd-repo-server",
		"argocd-application-controller",
		"argocd-redis",
		"argocd-dex-server",
		"argocd-applicationset-controller",
		"openshift-gitops-server",
		"openshift-gitops-repo-server",
		"openshift-gitops-application-controller",
		"openshift-gitops-redis",
		"openshift-gitops-dex-server",
		"openshift-gitops-applicationset-controller",
	}

	for _, name := range componentNames {
		component := d.getComponentStatus(ctx, namespace, name)
		if component != nil {
			components = append(components, *component)
		}
	}

	return components
}

func (d *Detector) getComponentStatus(ctx context.Context, namespace, name string) *ArgoCDComponent {
	args := []string{"get", "deployment", name, "-n", namespace, "-o", "json", "--ignore-not-found"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil
	}

	var deployment struct {
		Spec struct {
			Replicas int `json:"replicas"`
		} `json:"spec"`
		Status struct {
			ReadyReplicas     int `json:"readyReplicas"`
			AvailableReplicas int `json:"availableReplicas"`
		} `json:"status"`
	}

	if err := json.Unmarshal(output, &deployment); err != nil {
		return nil
	}

	component := &ArgoCDComponent{
		Name:      name,
		Replicas:  deployment.Spec.Replicas,
		Available: deployment.Status.AvailableReplicas,
		Ready:     deployment.Status.ReadyReplicas >= deployment.Spec.Replicas,
	}

	switch {
	case deployment.Status.ReadyReplicas >= deployment.Spec.Replicas:
		component.Status = StatusRunning
	case deployment.Status.ReadyReplicas > 0:
		component.Status = StatusDegraded
	default:
		component.Status = StatusNotRunning
	}

	return component
}

func (d *Detector) isRunning(components []ArgoCDComponent) bool {
	if len(components) == 0 {
		return false
	}
	for _, c := range components {
		if !c.Ready {
			return false
		}
	}
	return true
}

func (d *Detector) determineHealthStatus(components []ArgoCDComponent) string {
	if len(components) == 0 {
		return "unknown"
	}

	healthy := 0
	degraded := 0
	for _, c := range components {
		switch c.Status {
		case StatusRunning:
			healthy++
		case StatusDegraded:
			degraded++
		}
	}

	switch {
	case healthy == len(components):
		return "healthy"
	case degraded > 0 || healthy > 0:
		return "degraded"
	default:
		return "unhealthy"
	}
}

func (d *Detector) countApplications(ctx context.Context, namespace string) int {
	args := []string{"get", "applications.argoproj.io", "-n", namespace, "-o", "name"}
	if d.kubeContext != "" {
		args = append([]string{"--context", d.kubeContext}, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

func (d *Detector) analyzeAndRecommend(result *ArgoCDDetectionResult) {
	if !result.Running {
		result.Issues = append(result.Issues, "ArgoCD is not fully running")
		for _, c := range result.Components {
			if !c.Ready {
				result.Issues = append(result.Issues, fmt.Sprintf("Component %s is not ready (%d/%d replicas)", c.Name, c.Available, c.Replicas))
			}
		}
		result.Recommendations = append(result.Recommendations, "Check pod logs and events for troubleshooting")
	}

	if result.Type == ArgoCDTypeCommunity && result.Namespace == "openshift-gitops" {
		result.Issues = append(result.Issues, "Community ArgoCD installed in OpenShift GitOps namespace")
		result.Recommendations = append(result.Recommendations, "Consider using Red Hat OpenShift GitOps operator for full support")
	}

	if result.Type == ArgoCDTypeRedHat && result.OperatorSource == OperatorSourceCommunity {
		result.Issues = append(result.Issues, "Using community operator source on OpenShift")
		result.Recommendations = append(result.Recommendations, "Consider switching to redhat-operators for official support")
	}

	if result.URL == "" {
		result.Issues = append(result.Issues, "No external URL detected for ArgoCD")
		result.Recommendations = append(result.Recommendations, "Create a Route or Ingress to access ArgoCD UI externally")
	}

	if result.Version != "" {
		if strings.HasPrefix(result.Version, "v1") {
			result.Issues = append(result.Issues, "ArgoCD v1.x is outdated")
			result.Recommendations = append(result.Recommendations, "Upgrade to ArgoCD v2.x for latest features and security fixes")
		}
	}
}

func (r *ArgoCDDetectionResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *ArgoCDDetectionResult) Summary() string {
	var sb strings.Builder
	sb.WriteString("=== ArgoCD Detection Summary ===\n")
	sb.WriteString(fmt.Sprintf("State:         %s\n", r.State))
	sb.WriteString(fmt.Sprintf("Message:       %s\n", r.StateMessage))
	sb.WriteString(fmt.Sprintf("Installed:     %v\n", r.Installed))
	if r.Namespace != "" {
		sb.WriteString(fmt.Sprintf("Namespace:     %s\n", r.Namespace))
	}
	if r.Installed {
		sb.WriteString(fmt.Sprintf("Type:          %s\n", r.Type))
		sb.WriteString(fmt.Sprintf("Install Method: %s\n", r.InstallMethod))
		if r.OperatorSource != "" && r.OperatorSource != OperatorSourceUnknown {
			sb.WriteString(fmt.Sprintf("Operator Source: %s\n", r.OperatorSource))
		}
		if r.Version != "" {
			sb.WriteString(fmt.Sprintf("Version:       %s\n", r.Version))
		}
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("URL:           %s\n", r.URL))
		}
		sb.WriteString(fmt.Sprintf("Running:       %v\n", r.Running))
		sb.WriteString(fmt.Sprintf("Health:        %s\n", r.HealthStatus))
		sb.WriteString(fmt.Sprintf("Applications:  %d\n", r.AppCount))
	}

	if len(r.Components) > 0 {
		sb.WriteString(fmt.Sprintf("\nComponents (%d/%d ready):\n", r.ReadyComponents, r.TotalComponents))
		for _, c := range r.Components {
			status := "✅"
			if !c.Ready {
				status = "❌"
			}
			sb.WriteString(fmt.Sprintf("  %s %s (%d/%d replicas)\n", status, c.Name, c.Available, c.Replicas))
		}
	}

	if len(r.Issues) > 0 {
		sb.WriteString("\nIssues:\n")
		for _, issue := range r.Issues {
			sb.WriteString(fmt.Sprintf("  ⚠️  %s\n", issue))
		}
	}

	if len(r.Recommendations) > 0 {
		sb.WriteString("\nRecommendations:\n")
		for _, rec := range r.Recommendations {
			sb.WriteString(fmt.Sprintf("  💡 %s\n", rec))
		}
	}

	return sb.String()
}

// StateIcon returns an icon representing the current state
func (r *ArgoCDDetectionResult) StateIcon() string {
	switch r.State {
	case ArgoCDStateRunning:
		return "✅"
	case ArgoCDStatePartialInstall, ArgoCDStateNotRunning:
		return "⚠️"
	case ArgoCDStateNamespaceOnly, ArgoCDStateNotInstalled:
		return "❌"
	default:
		return "❓"
	}
}

// IsReady returns true if ArgoCD is fully operational
func (r *ArgoCDDetectionResult) IsReady() bool {
	return r.State == ArgoCDStateRunning
}

// NeedsBootstrap returns true if ArgoCD needs to be installed
func (r *ArgoCDDetectionResult) NeedsBootstrap() bool {
	return r.State == ArgoCDStateNotInstalled || r.State == ArgoCDStateNamespaceOnly
}
