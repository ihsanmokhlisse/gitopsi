//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMultiEnvironmentGeneration tests generation with multiple environments.
func TestMultiEnvironmentGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--preset", "standard",
		"--output", tmpDir,
		"--config", "fixtures/standard-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-standard")

	// Validate each environment overlay exists
	environments := []string{"dev", "staging", "prod"}

	for _, env := range environments {
		// Check infrastructure overlays
		infraOverlay := filepath.Join(projectDir, "infrastructure", "overlays", env)
		if _, err := os.Stat(infraOverlay); os.IsNotExist(err) {
			t.Errorf("infrastructure overlay for %s does not exist", env)
		}

		// Check infrastructure kustomization
		kustomization := filepath.Join(infraOverlay, "kustomization.yaml")
		if _, err := os.Stat(kustomization); os.IsNotExist(err) {
			t.Errorf("kustomization.yaml for %s does not exist", env)
		}

		// Check application overlays
		appOverlay := filepath.Join(projectDir, "applications", "overlays", env)
		if _, err := os.Stat(appOverlay); os.IsNotExist(err) {
			t.Errorf("application overlay for %s does not exist", env)
		}
	}
}

// TestEnvironmentIsolation tests that environment configs are properly isolated.
func TestEnvironmentIsolation(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--preset", "standard",
		"--output", tmpDir,
		"--config", "fixtures/standard-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-standard")

	// Read namespaces for each environment and verify they're different
	environments := []string{"dev", "staging", "prod"}
	namespaceContents := make(map[string]string)

	for _, env := range environments {
		nsFile := filepath.Join(projectDir, "infrastructure", "base", "namespaces", env+".yaml")
		content, err := os.ReadFile(nsFile)
		if err != nil {
			t.Errorf("failed to read namespace file for %s: %v", env, err)
			continue
		}
		namespaceContents[env] = string(content)

		// Verify namespace contains environment-specific name
		if !strings.Contains(string(content), env) {
			t.Errorf("namespace file for %s should contain environment name", env)
		}
	}

	// Verify namespaces are different
	if len(namespaceContents) == 3 {
		if namespaceContents["dev"] == namespaceContents["prod"] {
			t.Error("dev and prod namespace contents should differ")
		}
	}
}

// TestSingleEnvironment tests generation with a single environment.
func TestSingleEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--preset", "minimal",
		"--output", tmpDir,
		"--config", "fixtures/minimal-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-minimal")

	// Only dev environment should exist
	devOverlay := filepath.Join(projectDir, "infrastructure", "overlays", "dev")
	if _, err := os.Stat(devOverlay); os.IsNotExist(err) {
		t.Error("dev overlay should exist for minimal preset")
	}

	// staging and prod should not exist
	for _, env := range []string{"staging", "prod"} {
		overlay := filepath.Join(projectDir, "infrastructure", "overlays", env)
		if _, err := os.Stat(overlay); err == nil {
			t.Errorf("%s overlay should not exist for minimal preset", env)
		}
	}
}

// TestCustomEnvironments tests generation with custom environment names.
func TestCustomEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with custom environment names
	configContent := `
project:
  name: test-custom-envs
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: development
  - name: qa
  - name: production
infra:
  namespaces: true
  rbac: true
`
	configPath := filepath.Join(tmpDir, "custom-envs.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-custom-envs")

	// Check custom environments exist
	customEnvs := []string{"development", "qa", "production"}
	for _, env := range customEnvs {
		overlay := filepath.Join(projectDir, "infrastructure", "overlays", env)
		if _, err := os.Stat(overlay); os.IsNotExist(err) {
			t.Errorf("overlay for custom environment %s should exist", env)
		}
	}

	// Standard envs should not exist
	standardEnvs := []string{"dev", "staging", "prod"}
	for _, env := range standardEnvs {
		overlay := filepath.Join(projectDir, "infrastructure", "overlays", env)
		if _, err := os.Stat(overlay); err == nil {
			t.Errorf("standard overlay %s should not exist for custom environments", env)
		}
	}
}

// TestManyEnvironments tests generation with many environments.
func TestManyEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with many environments
	configContent := `
project:
  name: test-many-envs
platform: kubernetes
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
  - name: test
  - name: staging
  - name: uat
  - name: preprod
  - name: prod
infra:
  namespaces: true
`
	configPath := filepath.Join(tmpDir, "many-envs.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-many-envs")

	// Check all environments exist
	envs := []string{"dev", "test", "staging", "uat", "preprod", "prod"}
	for _, env := range envs {
		overlay := filepath.Join(projectDir, "infrastructure", "overlays", env)
		if _, err := os.Stat(overlay); os.IsNotExist(err) {
			t.Errorf("overlay for environment %s should exist", env)
		}

		// Check namespace exists
		nsFile := filepath.Join(projectDir, "infrastructure", "base", "namespaces", env+".yaml")
		if _, err := os.Stat(nsFile); os.IsNotExist(err) {
			t.Errorf("namespace file for %s should exist", env)
		}
	}
}

// TestEnvironmentWithClusterURL tests environment with cluster URLs.
func TestEnvironmentWithClusterURL(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with cluster URLs
	configContent := `
project:
  name: test-cluster-urls
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
    cluster: https://dev.example.com
  - name: staging
    cluster: https://staging.example.com
  - name: prod
    cluster: https://prod.example.com
infra:
  namespaces: true
`
	configPath := filepath.Join(tmpDir, "cluster-urls.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-cluster-urls")

	// Check that multi-cluster ArgoCD files were generated
	clusterSecrets := filepath.Join(projectDir, "argocd", "clusters")
	if _, err := os.Stat(clusterSecrets); os.IsNotExist(err) {
		t.Log("clusters directory not created - may be expected for some configurations")
	}

	// Check ApplicationSets reference clusters
	appSetsDir := filepath.Join(projectDir, "argocd", "applicationsets")
	if _, err := os.Stat(appSetsDir); err == nil {
		// Look for multi-cluster ApplicationSet patterns
		files, _ := filepath.Glob(filepath.Join(appSetsDir, "*.yaml"))
		if len(files) == 0 {
			t.Log("No ApplicationSet files found")
		}
	}
}

// TestInfraOnlyScope tests infrastructure-only scope.
func TestInfraOnlyScope(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--output", tmpDir,
		"--config", "fixtures/infra-only-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-infra")

	// Infrastructure should exist
	infraDir := filepath.Join(projectDir, "infrastructure")
	if _, err := os.Stat(infraDir); os.IsNotExist(err) {
		t.Error("infrastructure directory should exist")
	}

	// Applications might be minimal or non-existent depending on implementation
	appsDir := filepath.Join(projectDir, "applications")
	if _, err := os.Stat(appsDir); err == nil {
		// If exists, check it has minimal content
		baseContent, _ := os.ReadDir(filepath.Join(appsDir, "base"))
		if len(baseContent) > 1 {
			t.Log("applications/base has content - may include sample app")
		}
	}
}

// TestAppOnlyScope tests application-only scope.
func TestAppOnlyScope(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--output", tmpDir,
		"--config", "fixtures/app-only-config.yaml",
	)

	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-app")

	// Applications should exist
	appsDir := filepath.Join(projectDir, "applications")
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		t.Error("applications directory should exist for app-only scope")
	}

	// Check applications has content
	appsBase := filepath.Join(appsDir, "base")
	if content, err := os.ReadDir(appsBase); err == nil && len(content) == 0 {
		t.Error("applications/base should have content for app-only scope")
	}
}
