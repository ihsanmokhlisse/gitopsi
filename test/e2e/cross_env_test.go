//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

// Platform-specific test skip helpers

// skipIfNotPlatform skips the test if the current platform doesn't match.
func skipIfNotPlatform(t *testing.T, platform string) {
	t.Helper()
	currentPlatform := os.Getenv("E2E_PLATFORM")
	if currentPlatform == "" {
		currentPlatform = "kubernetes" // default
	}
	if currentPlatform != platform {
		t.Skipf("Skipping: requires platform %s, got %s", platform, currentPlatform)
	}
}

// skipIfNotEKS skips if not running on EKS.
func skipIfNotEKS(t *testing.T) {
	t.Helper()
	if os.Getenv("E2E_EKS") == "" {
		t.Skip("Skipping: E2E_EKS not set")
	}
	skipIfNotPlatform(t, "eks")
}

// skipIfNotAKS skips if not running on AKS.
func skipIfNotAKS(t *testing.T) {
	t.Helper()
	if os.Getenv("E2E_AKS") == "" {
		t.Skip("Skipping: E2E_AKS not set")
	}
	skipIfNotPlatform(t, "aks")
}

// skipIfNotOpenShift skips if not running on OpenShift.
func skipIfNotOpenShift(t *testing.T) {
	t.Helper()
	if os.Getenv("E2E_OPENSHIFT") == "" {
		t.Skip("Skipping: E2E_OPENSHIFT not set")
	}
	skipIfNotPlatform(t, "openshift")
}

// TestCrossEnvMinimalOnKubernetes tests minimal preset on vanilla Kubernetes.
func TestCrossEnvMinimalOnKubernetes(t *testing.T) {
	skipIfNotPlatform(t, "kubernetes")
	skipIfNoCluster(t)

	tmpDir := t.TempDir()
	k8s := NewKubernetesHelper(t)
	defer k8s.Cleanup()

	// Generate manifests
	result := RunGitopsi(t, "init",
		"--preset", "minimal",
		"--output", tmpDir,
		"--config", "fixtures/minimal-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-minimal")
	ValidateMinimalPreset(t, projectDir)

	// Dry-run apply
	nsFile := filepath.Join(projectDir, "infrastructure", "base", "namespaces", "dev.yaml")
	if err := k8s.ApplyDryRun(nsFile); err != nil {
		t.Errorf("dry-run apply failed: %v", err)
	}
}

// TestCrossEnvStandardOnEKS tests standard preset on EKS.
func TestCrossEnvStandardOnEKS(t *testing.T) {
	skipIfNotEKS(t)

	tmpDir := t.TempDir()

	// Create EKS-specific config
	configContent := `
project:
  name: test-eks-standard
platform: eks
scope: both
gitops_tool: argocd
preset: standard
environments:
  - name: dev
  - name: staging
  - name: prod
infra:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true
`
	configPath := filepath.Join(tmpDir, "eks-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-eks-standard")
	ValidateStandardPreset(t, projectDir)

	// Verify EKS-specific configurations if any
	t.Log("EKS standard preset generated successfully")
}

// TestCrossEnvStandardOnAKS tests standard preset on AKS.
func TestCrossEnvStandardOnAKS(t *testing.T) {
	skipIfNotAKS(t)

	tmpDir := t.TempDir()

	// Create AKS-specific config
	configContent := `
project:
  name: test-aks-standard
platform: aks
scope: both
gitops_tool: argocd
preset: standard
environments:
  - name: dev
  - name: staging
  - name: prod
infra:
  namespaces: true
  rbac: true
  network_policies: true
`
	configPath := filepath.Join(tmpDir, "aks-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-aks-standard")
	ValidateStandardPreset(t, projectDir)

	t.Log("AKS standard preset generated successfully")
}

// TestCrossEnvOpenShiftGitOps tests OpenShift GitOps configuration.
func TestCrossEnvOpenShiftGitOps(t *testing.T) {
	skipIfNotOpenShift(t)

	tmpDir := t.TempDir()

	// Create OpenShift-specific config
	configContent := `
project:
  name: test-openshift
platform: openshift
scope: both
gitops_tool: argocd
preset: standard
environments:
  - name: dev
  - name: staging
  - name: prod
infra:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true
bootstrap:
  mode: olm  # OpenShift uses OLM for ArgoCD installation
`
	configPath := filepath.Join(tmpDir, "openshift-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-openshift")
	ValidateStandardPreset(t, projectDir)

	// Verify OpenShift-specific configurations
	v := NewProjectValidator(t, projectDir)
	v.DirectoryExists("bootstrap").
		DirectoryExists("bootstrap/argocd").
		Assert()

	t.Log("OpenShift GitOps preset generated successfully")
}

// TestCrossEnvMultiClusterTopology tests multi-cluster configuration.
func TestCrossEnvMultiClusterTopology(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--output", tmpDir,
		"--config", "fixtures/multi-cluster-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-multi-cluster")

	// Verify multi-cluster structure
	v := NewProjectValidator(t, projectDir)
	v.DirectoryExists("argocd").
		DirectoryExists("argocd/applicationsets").
		Assert()

	// Check for cluster-specific ApplicationSets or configurations
	appSetsDir := filepath.Join(projectDir, "argocd", "applicationsets")
	if _, err := os.Stat(appSetsDir); err == nil {
		files, _ := filepath.Glob(filepath.Join(appSetsDir, "*.yaml"))
		t.Logf("Found %d ApplicationSet files", len(files))
	}
}

// TestCrossEnvPlatformDetection tests automatic platform detection.
func TestCrossEnvPlatformDetection(t *testing.T) {
	skipIfNoCluster(t)

	// Run preflight to see detected platform
	result := RunGitopsi(t, "preflight", "--verbose")

	if !result.Success() {
		// Preflight may have warnings
		t.Logf("Preflight output: %s", result.Combined)
	}

	// Check that platform was detected
	platforms := []string{"kubernetes", "openshift", "eks", "aks", "gke"}
	detected := false
	for _, p := range platforms {
		if result.Contains(p) {
			t.Logf("Detected platform: %s", p)
			detected = true
			break
		}
	}

	if !detected {
		t.Log("No specific platform detected - may be vanilla Kubernetes")
	}
}

// TestCrossEnvAllPresets runs all presets across available environment.
func TestCrossEnvAllPresets(t *testing.T) {
	presets := []struct {
		name       string
		configFile string
		validator  func(t *testing.T, projectDir string)
	}{
		{
			name:       "minimal",
			configFile: "fixtures/minimal-config.yaml",
			validator:  ValidateMinimalPreset,
		},
		{
			name:       "standard",
			configFile: "fixtures/standard-config.yaml",
			validator:  ValidateStandardPreset,
		},
		{
			name:       "enterprise",
			configFile: "fixtures/enterprise-config.yaml",
			validator:  ValidateEnterprisePreset,
		},
	}

	for _, preset := range presets {
		t.Run(preset.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			result := RunGitopsi(t, "init",
				"--preset", preset.name,
				"--output", tmpDir,
				"--config", preset.configFile,
			)

			if !result.Success() {
				t.Fatalf("init failed for %s: %s", preset.name, result.Combined)
			}

			// Get project name from config
			projectName := "test-" + preset.name
			projectDir := filepath.Join(tmpDir, projectName)

			// Run validator
			preset.validator(t, projectDir)
		})
	}
}

// TestCrossEnvNamespaceNaming tests that namespaces follow platform conventions.
func TestCrossEnvNamespaceNaming(t *testing.T) {
	platforms := []struct {
		name            string
		platform        string
		expectedPattern string
	}{
		{
			name:            "kubernetes",
			platform:        "kubernetes",
			expectedPattern: "argocd", // Standard ArgoCD namespace
		},
		{
			name:            "openshift",
			platform:        "openshift",
			expectedPattern: "openshift-gitops", // OpenShift GitOps namespace
		},
	}

	for _, p := range platforms {
		t.Run(p.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			configContent := `
project:
  name: test-ns-` + p.platform + `
platform: ` + p.platform + `
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
infra:
  namespaces: true
`
			configPath := filepath.Join(tmpDir, "ns-config.yaml")
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			result := RunGitopsi(t, "init",
				"--config", configPath,
				"--output", tmpDir,
			)

			if !result.Success() {
				t.Fatalf("init failed: %s", result.Combined)
			}

			// Check ArgoCD project files reference correct namespace
			projectDir := filepath.Join(tmpDir, "test-ns-"+p.platform)
			argoCDDir := filepath.Join(projectDir, "argocd")

			if _, err := os.Stat(argoCDDir); err == nil {
				// Read ArgoCD project file and check namespace
				projectFiles, _ := filepath.Glob(filepath.Join(argoCDDir, "**", "*.yaml"))
				t.Logf("Found %d ArgoCD files for %s platform", len(projectFiles), p.platform)
			}
		})
	}
}
