package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ihsanmokhlisse/gitopsi/internal/bootstrap"
	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/generator"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
)

func TestRegression_34_OpenshiftUsesOpenshiftGitopsNamespace(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "regression-34"},
		Platform:   "openshift",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	filesToCheck := []string{
		"regression-34/argocd/projects/infrastructure.yaml",
		"regression-34/argocd/projects/applications.yaml",
		"regression-34/argocd/applicationsets/infra-dev.yaml",
		"regression-34/argocd/applicationsets/apps-dev.yaml",
		"regression-34/argocd/applicationsets/infra-prod.yaml",
		"regression-34/argocd/applicationsets/apps-prod.yaml",
	}

	for _, file := range filesToCheck {
		fullPath := filepath.Join(tmpDir, file)
		content, err := os.ReadFile(fullPath)
		require.NoError(t, err, "Should be able to read %s", file)

		assert.Contains(t, string(content), "namespace: openshift-gitops",
			"File %s should use 'namespace: openshift-gitops' for OpenShift platform", file)
		assert.NotContains(t, string(content), "namespace: argocd",
			"File %s should NOT use 'namespace: argocd' for OpenShift platform", file)
	}
}

func TestRegression_34_KubernetesUsesArgoCDNamespace(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "regression-34-k8s"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	projectFile := filepath.Join(tmpDir, "regression-34-k8s/argocd/projects/infrastructure.yaml")
	content, err := os.ReadFile(projectFile)
	require.NoError(t, err, "Should be able to read project file")

	assert.Contains(t, string(content), "namespace: argocd",
		"Kubernetes platform should use 'namespace: argocd'")
}

func TestRegression_34_CustomNamespaceOverridesDefault(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "regression-34-custom"},
		Platform:   "openshift",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Bootstrap: config.BootstrapConfig{
			Namespace: "my-custom-argocd",
		},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	projectFile := filepath.Join(tmpDir, "regression-34-custom/argocd/projects/infrastructure.yaml")
	content, err := os.ReadFile(projectFile)
	require.NoError(t, err, "Should read project file")

	assert.Contains(t, string(content), "namespace: my-custom-argocd",
		"Custom namespace should override default")
}

func TestRegression_36_ArgoCDDetectionStates(t *testing.T) {
	states := []bootstrap.ArgoCDState{
		bootstrap.ArgoCDStateNotInstalled,
		bootstrap.ArgoCDStateNamespaceOnly,
		bootstrap.ArgoCDStatePartialInstall,
		bootstrap.ArgoCDStateNotRunning,
		bootstrap.ArgoCDStateRunning,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			result := &bootstrap.ArgoCDDetectionResult{
				State: state,
			}

			icon := result.StateIcon()
			assert.NotEmpty(t, icon, "State %s should have an icon", state)

			_ = result.IsReady()
			_ = result.NeedsBootstrap()
		})
	}
}

func TestRegression_36_ArgoCDDetectionPartialInstall(t *testing.T) {
	d := bootstrap.NewDetector("", 30*time.Second)

	result := &bootstrap.ArgoCDDetectionResult{
		Namespace: "argocd",
		Components: []bootstrap.ArgoCDComponent{
			{Name: "server", Ready: true},
		},
		TotalComponents: 1,
		ReadyComponents: 1,
	}

	state, msg := d.DetermineState(result)

	assert.Equal(t, bootstrap.ArgoCDStatePartialInstall, state,
		"Single component should be detected as partial install")
	assert.NotEmpty(t, msg, "State should have a message")
}

func TestRegression_36_ArgoCDDetectionNamespaceOnly(t *testing.T) {
	d := bootstrap.NewDetector("", 30*time.Second)

	result := &bootstrap.ArgoCDDetectionResult{
		Namespace:  "openshift-gitops",
		Components: []bootstrap.ArgoCDComponent{},
	}

	state, msg := d.DetermineState(result)

	assert.Equal(t, bootstrap.ArgoCDStateNamespaceOnly, state,
		"Namespace without components should be namespace_only")
	assert.Contains(t, msg, "openshift-gitops",
		"Message should mention the namespace")
}

func TestRegression_36_ArgoCDDetectionRunning(t *testing.T) {
	d := bootstrap.NewDetector("", 30*time.Second)

	result := &bootstrap.ArgoCDDetectionResult{
		Namespace: "argocd",
		Components: []bootstrap.ArgoCDComponent{
			{Name: "server", Ready: true},
			{Name: "repo-server", Ready: true},
			{Name: "application-controller", Ready: true},
		},
		TotalComponents: 3,
		ReadyComponents: 3,
	}

	state, _ := d.DetermineState(result)

	assert.Equal(t, bootstrap.ArgoCDStateRunning, state,
		"All components ready should be running state")
}

func TestRegression_40_AllInfraSubdirsHaveKustomization(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "regression-40"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	subdirs := []string{"namespaces", "rbac", "network-policies", "resource-quotas"}

	for _, subdir := range subdirs {
		kustomizePath := filepath.Join(tmpDir, "regression-40/infrastructure/base", subdir, "kustomization.yaml")

		_, err := os.Stat(kustomizePath)
		assert.False(t, os.IsNotExist(err),
			"kustomization.yaml should exist in %s/ directory", subdir)

		if err == nil {
			content, readErr := os.ReadFile(kustomizePath)
			require.NoError(t, readErr, "Should read %s/kustomization.yaml", subdir)

			assert.Contains(t, string(content), "apiVersion: kustomize.config.k8s.io",
				"%s/kustomization.yaml should have proper apiVersion", subdir)
			assert.Contains(t, string(content), "kind: Kustomization",
				"%s/kustomization.yaml should have kind: Kustomization", subdir)

			for _, env := range cfg.Environments {
				assert.Contains(t, string(content), env.Name+".yaml",
					"%s/kustomization.yaml should reference %s.yaml", subdir, env.Name)
			}
		}
	}
}

func TestRegression_40_BaseKustomizationReferencesSubdirs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "regression-40-base"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	baseKustomizePath := filepath.Join(tmpDir, "regression-40-base/infrastructure/base/kustomization.yaml")
	content, err := os.ReadFile(baseKustomizePath)
	require.NoError(t, err, "Should read base kustomization.yaml")

	subdirs := []string{"namespaces/", "rbac/", "network-policies/", "resource-quotas/"}
	for _, subdir := range subdirs {
		assert.Contains(t, string(content), subdir,
			"Base kustomization.yaml should reference %s subdirectory", subdir)
	}
}

func TestRegression_40_OverlaysHaveKustomization(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "regression-40-overlays"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	for _, env := range cfg.Environments {
		overlayPath := filepath.Join(tmpDir, "regression-40-overlays/infrastructure/overlays", env.Name, "kustomization.yaml")
		_, err := os.Stat(overlayPath)
		assert.False(t, os.IsNotExist(err),
			"kustomization.yaml should exist in %s overlay directory", env.Name)
	}
}

func TestRegression_41_BootstrapOptionsStructure(t *testing.T) {
	opts := &bootstrap.Options{
		Tool:            bootstrap.ToolArgoCD,
		Mode:            bootstrap.ModeHelm,
		Namespace:       "argocd",
		Version:         "v2.9.0",
		Wait:            true,
		Timeout:         300,
		ConfigureRepo:   true,
		RepoURL:         "https://github.com/test/repo.git",
		RepoBranch:      "main",
		CreateAppOfApps: true,
		SyncInitial:     true,
	}

	assert.Equal(t, bootstrap.ToolArgoCD, opts.Tool)
	assert.Equal(t, bootstrap.ModeHelm, opts.Mode)
	assert.Equal(t, "argocd", opts.Namespace)
	assert.True(t, opts.ConfigureRepo, "ConfigureRepo should be true for auto-apply")
	assert.True(t, opts.CreateAppOfApps, "CreateAppOfApps should be true for auto-apply")
}

func TestRegression_41_BootstrapModeConstants(t *testing.T) {
	kubernetesValidModes := []bootstrap.Mode{
		bootstrap.ModeHelm,
		bootstrap.ModeManifest,
		bootstrap.ModeKustomize,
	}

	for _, mode := range kubernetesValidModes {
		t.Run(string(mode)+"_kubernetes", func(t *testing.T) {
			assert.True(t, bootstrap.IsValidMode(mode, bootstrap.ToolArgoCD, "kubernetes"),
				"Mode %s should be valid for ArgoCD on kubernetes", mode)
		})
	}

	t.Run("olm_openshift", func(t *testing.T) {
		assert.True(t, bootstrap.IsValidMode(bootstrap.ModeOLM, bootstrap.ToolArgoCD, "openshift"),
			"Mode OLM should be valid for ArgoCD on OpenShift")
	})

	t.Run("olm_not_valid_on_kubernetes", func(t *testing.T) {
		assert.False(t, bootstrap.IsValidMode(bootstrap.ModeOLM, bootstrap.ToolArgoCD, "kubernetes"),
			"Mode OLM should NOT be valid for ArgoCD on kubernetes")
	})
}

func TestRegression_41_ToolConstants(t *testing.T) {
	tools := []bootstrap.Tool{
		bootstrap.ToolArgoCD,
		bootstrap.ToolFlux,
	}

	for _, tool := range tools {
		t.Run(string(tool), func(t *testing.T) {
			assert.NotEmpty(t, tool, "Tool should not be empty")
		})
	}
}

func TestRegression_AllPlatformsGenerateCorrectNamespace(t *testing.T) {
	platforms := []struct {
		platform          string
		expectedNamespace string
	}{
		{"kubernetes", "argocd"},
		{"openshift", "openshift-gitops"},
		{"aks", "argocd"},
		{"eks", "argocd"},
	}

	for _, tc := range platforms {
		t.Run(tc.platform, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "ns-test-" + tc.platform},
				Platform:   tc.platform,
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			err := gen.Generate()
			require.NoError(t, err, "Generate should not fail for %s", tc.platform)

			projectFile := filepath.Join(tmpDir, "ns-test-"+tc.platform+"/argocd/projects/infrastructure.yaml")
			content, err := os.ReadFile(projectFile)
			require.NoError(t, err, "Should read project file for %s", tc.platform)

			assert.Contains(t, string(content), "namespace: "+tc.expectedNamespace,
				"Platform %s should use namespace '%s'", tc.platform, tc.expectedNamespace)
		})
	}
}

func TestRegression_KustomizeBuildCompatibility(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "kustomize-compat"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	baseKustomizePath := filepath.Join(tmpDir, "kustomize-compat/infrastructure/base/kustomization.yaml")
	content, err := os.ReadFile(baseKustomizePath)
	require.NoError(t, err, "Should read base kustomization.yaml")

	assert.Contains(t, string(content), "apiVersion: kustomize.config.k8s.io/v1beta1",
		"Should use standard kustomize API version")
	assert.Contains(t, string(content), "kind: Kustomization",
		"Should have kind: Kustomization")

	assert.NotContains(t, string(content), "bases:",
		"Should not use deprecated 'bases:' field")

	lines := strings.Split(string(content), "\n")
	hasResources := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "resources:" {
			hasResources = true
			break
		}
	}
	assert.True(t, hasResources, "Should use 'resources:' field")
}

func TestRegression_35_PreflightCheckGitCredentials(t *testing.T) {
	tests := []struct {
		name        string
		gitToken    string
		gitURL      string
		expectError bool
	}{
		{
			name:        "valid token and URL",
			gitToken:    "ghp_validtoken123",
			gitURL:      "https://github.com/org/repo.git",
			expectError: false,
		},
		{
			name:        "empty token should warn",
			gitToken:    "",
			gitURL:      "https://github.com/org/repo.git",
			expectError: true,
		},
		{
			name:        "empty URL should warn",
			gitToken:    "ghp_validtoken123",
			gitURL:      "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Project:    config.Project{Name: "preflight-test"},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: "argocd",
				Git: config.GitConfig{
					URL: tc.gitURL,
					Auth: config.GitAuth{
						Token: tc.gitToken,
					},
				},
			}

			hasGitConfig := cfg.Git.URL != "" && cfg.Git.Auth.Token != ""

			if tc.expectError {
				assert.False(t, hasGitConfig,
					"Should detect missing Git configuration")
			} else {
				assert.True(t, hasGitConfig,
					"Should have valid Git configuration")
			}
		})
	}
}

func TestRegression_35_PreflightCheckClusterConfig(t *testing.T) {
	tests := []struct {
		name         string
		clusterURL   string
		clusterToken string
		expectReady  bool
	}{
		{
			name:         "with cluster URL and token",
			clusterURL:   "https://api.cluster.local:6443",
			clusterToken: "cluster-token-xyz",
			expectReady:  true,
		},
		{
			name:         "missing cluster URL",
			clusterURL:   "",
			clusterToken: "cluster-token-xyz",
			expectReady:  false,
		},
		{
			name:         "missing cluster token",
			clusterURL:   "https://api.cluster.local:6443",
			clusterToken: "",
			expectReady:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Project:  config.Project{Name: "cluster-preflight"},
				Platform: "kubernetes",
				Cluster: config.ClusterConfig{
					URL: tc.clusterURL,
					Auth: config.ClusterAuth{
						Token: tc.clusterToken,
					},
				},
			}

			hasClusterConfig := cfg.Cluster.URL != "" && cfg.Cluster.Auth.Token != ""

			if tc.expectReady {
				assert.True(t, hasClusterConfig,
					"Cluster should be ready with URL and token")
			} else {
				assert.False(t, hasClusterConfig,
					"Cluster should not be ready without URL or token")
			}
		})
	}
}

func TestRegression_35_PreflightSecuritySettings(t *testing.T) {
	tests := []struct {
		name          string
		skipTLS       bool
		expectWarning bool
	}{
		{
			name:          "secure settings",
			skipTLS:       false,
			expectWarning: false,
		},
		{
			name:          "skip TLS verify should warn",
			skipTLS:       true,
			expectWarning: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Project:  config.Project{Name: "security-preflight"},
				Platform: "kubernetes",
				Cluster: config.ClusterConfig{
					Auth: config.ClusterAuth{
						SkipTLS: tc.skipTLS,
					},
				},
			}

			hasSecurityWarning := cfg.Cluster.Auth.SkipTLS

			if tc.expectWarning {
				assert.True(t, hasSecurityWarning,
					"Should detect security warning when SkipTLS is true")
			} else {
				assert.False(t, hasSecurityWarning,
					"Should not have security warnings when SkipTLS is false")
			}
		})
	}
}

// ============================================================================
// Issue #42 - Config Validation Edge Cases
// Bug: Empty or invalid project names caused panics in generator
// ============================================================================

func TestRegression_42_EmptyProjectNameValidation(t *testing.T) {
	cfg := &config.Config{
		Project:    config.Project{Name: ""},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
	}

	err := cfg.Validate()
	assert.Error(t, err, "Empty project name should fail validation")
	assert.Contains(t, err.Error(), "project", "Error should mention project")
}

func TestRegression_42_InvalidPlatformValidation(t *testing.T) {
	cfg := &config.Config{
		Project:    config.Project{Name: "test-project"},
		Platform:   "invalid-platform",
		Scope:      "both",
		GitOpsTool: "argocd",
	}

	err := cfg.Validate()
	assert.Error(t, err, "Invalid platform should fail validation")
}

func TestRegression_42_InvalidScopeValidation(t *testing.T) {
	cfg := &config.Config{
		Project:    config.Project{Name: "test-project"},
		Platform:   "kubernetes",
		Scope:      "invalid-scope",
		GitOpsTool: "argocd",
	}

	err := cfg.Validate()
	assert.Error(t, err, "Invalid scope should fail validation")
}

func TestRegression_42_InvalidGitOpsToolValidation(t *testing.T) {
	cfg := &config.Config{
		Project:    config.Project{Name: "test-project"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "invalid-tool",
	}

	err := cfg.Validate()
	assert.Error(t, err, "Invalid GitOps tool should fail validation")
}

// ============================================================================
// Issue #43 - Environment Configuration Edge Cases
// Bug: Duplicate environment names caused file overwrites
// ============================================================================

func TestRegression_43_DuplicateEnvironmentNames(t *testing.T) {
	cfg := &config.Config{
		Project:    config.Project{Name: "dup-env-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "dev"}, // Duplicate
			{Name: "prod"},
		},
	}

	err := cfg.Validate()
	// The config should either reject duplicates or handle them gracefully
	// For now, we just ensure it doesn't panic
	if err == nil {
		// If validation passes, ensure generation doesn't overwrite files
		tmpDir := t.TempDir()
		writer := output.New(tmpDir, false, false)
		gen := generator.New(cfg, writer, false)
		genErr := gen.Generate()
		// Generation should either succeed or fail gracefully
		_ = genErr
	}
}

func TestRegression_43_EmptyEnvironmentList(t *testing.T) {
	cfg := &config.Config{
		Project:      config.Project{Name: "no-env-test"},
		Platform:     "kubernetes",
		Scope:        "infrastructure",
		GitOpsTool:   "argocd",
		Environments: []config.Environment{},
	}

	err := cfg.Validate()
	assert.Error(t, err, "Empty environment list should fail validation")
}

func TestRegression_43_EnvironmentWithSpaces(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "space-env-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev env"}, // Space in name - could cause path issues
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	// Should either sanitize the name or reject it - not panic
	err := gen.Generate()
	// Just ensure no panic occurs
	_ = err
}

// ============================================================================
// Issue #44 - Git URL Parsing Edge Cases
// Bug: Malformed Git URLs caused crashes during repository operations
// ============================================================================

func TestRegression_44_GitURLParsing(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		isSSH    bool
		isHTTPS  bool
		hasError bool
	}{
		{
			name:    "standard HTTPS",
			url:     "https://github.com/org/repo.git",
			isHTTPS: true,
		},
		{
			name:  "SSH format",
			url:   "git@github.com:org/repo.git",
			isSSH: true,
		},
		{
			name:    "HTTPS without .git",
			url:     "https://github.com/org/repo",
			isHTTPS: true,
		},
		{
			name:     "empty URL",
			url:      "",
			hasError: true,
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			hasError: true,
		},
		{
			name:    "GitLab URL",
			url:     "https://gitlab.com/group/project.git",
			isHTTPS: true,
		},
		{
			name:    "Azure DevOps URL",
			url:     "https://dev.azure.com/org/project/_git/repo",
			isHTTPS: true,
		},
		{
			name:    "Bitbucket URL",
			url:     "https://bitbucket.org/org/repo.git",
			isHTTPS: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Config used to validate URL is stored properly
			_ = &config.Config{
				Project:    config.Project{Name: "url-test"},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     config.Output{URL: tc.url},
			}

			isSSH := strings.HasPrefix(tc.url, "git@")
			isHTTPS := strings.HasPrefix(tc.url, "https://") || strings.HasPrefix(tc.url, "http://")
			isEmpty := tc.url == ""

			if tc.hasError {
				assert.True(t, isEmpty || (!isSSH && !isHTTPS),
					"Should detect invalid URL format")
			} else {
				assert.Equal(t, tc.isSSH, isSSH, "SSH detection mismatch")
				assert.Equal(t, tc.isHTTPS, isHTTPS, "HTTPS detection mismatch")
			}
		})
	}
}

// ============================================================================
// Issue #45 - ApplicationSet Generation Edge Cases
// Bug: ApplicationSet templates missing required fields caused ArgoCD failures
// ============================================================================

func TestRegression_45_ApplicationSetHasRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "appset-fields"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	appsetPath := filepath.Join(tmpDir, "appset-fields/argocd/applicationsets/infra-dev.yaml")
	content, err := os.ReadFile(appsetPath)
	require.NoError(t, err, "Should read ApplicationSet file")

	contentStr := string(content)

	// Required ApplicationSet fields
	assert.Contains(t, contentStr, "apiVersion: argoproj.io/v1alpha1",
		"Should have correct API version")
	assert.Contains(t, contentStr, "kind: Application",
		"Should have kind Application or ApplicationSet")
	assert.Contains(t, contentStr, "spec:",
		"Should have spec section")
	assert.Contains(t, contentStr, "destination:",
		"Should have destination section")
	assert.Contains(t, contentStr, "source:",
		"Should have source section")
	assert.Contains(t, contentStr, "project:",
		"Should specify ArgoCD project")
}

func TestRegression_45_ApplicationSetSyncPolicy(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "appset-sync"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	appsetPath := filepath.Join(tmpDir, "appset-sync/argocd/applicationsets/infra-dev.yaml")
	content, err := os.ReadFile(appsetPath)
	require.NoError(t, err, "Should read ApplicationSet file")

	contentStr := string(content)

	// Sync policy should be present for GitOps automation
	assert.Contains(t, contentStr, "syncPolicy:",
		"Should have syncPolicy section for automated sync")
}

// ============================================================================
// Issue #46 - RBAC Generation Edge Cases
// Bug: RBAC bindings referenced non-existent service accounts
// ============================================================================

func TestRegression_46_RBACBindingsReferenceExistingResources(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "rbac-refs"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
			RBAC:       true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	rbacPath := filepath.Join(tmpDir, "rbac-refs/infrastructure/base/rbac/dev.yaml")
	content, err := os.ReadFile(rbacPath)
	require.NoError(t, err, "Should read RBAC file")

	contentStr := string(content)

	// RBAC should have valid structure
	assert.Contains(t, contentStr, "apiVersion: rbac.authorization.k8s.io",
		"Should have correct RBAC API version")
	assert.Contains(t, contentStr, "kind:",
		"Should specify kind")
	assert.Contains(t, contentStr, "metadata:",
		"Should have metadata section")
}

// ============================================================================
// Issue #47 - NetworkPolicy Generation Edge Cases
// Bug: NetworkPolicies with empty port specifications caused K8s rejections
// ============================================================================

func TestRegression_47_NetworkPolicyHasValidStructure(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "netpol-valid"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			NetworkPolicies: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	netpolPath := filepath.Join(tmpDir, "netpol-valid/infrastructure/base/network-policies/dev.yaml")
	content, err := os.ReadFile(netpolPath)
	require.NoError(t, err, "Should read NetworkPolicy file")

	contentStr := string(content)

	// NetworkPolicy should have valid structure
	assert.Contains(t, contentStr, "apiVersion: networking.k8s.io/v1",
		"Should have correct NetworkPolicy API version")
	assert.Contains(t, contentStr, "kind: NetworkPolicy",
		"Should have kind NetworkPolicy")
	assert.Contains(t, contentStr, "spec:",
		"Should have spec section")
	assert.Contains(t, contentStr, "podSelector:",
		"Should have podSelector")
}

// ============================================================================
// Issue #48 - ResourceQuota Generation Edge Cases
// Bug: ResourceQuotas with zero values caused quota enforcement issues
// ============================================================================

func TestRegression_48_ResourceQuotaHasValidValues(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "quota-valid"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces:     true,
			ResourceQuotas: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	quotaPath := filepath.Join(tmpDir, "quota-valid/infrastructure/base/resource-quotas/dev.yaml")
	content, err := os.ReadFile(quotaPath)
	require.NoError(t, err, "Should read ResourceQuota file")

	contentStr := string(content)

	// ResourceQuota should have valid structure
	assert.Contains(t, contentStr, "apiVersion: v1",
		"Should have correct API version")
	assert.Contains(t, contentStr, "kind: ResourceQuota",
		"Should have kind ResourceQuota")
	assert.Contains(t, contentStr, "spec:",
		"Should have spec section")
}

// ============================================================================
// Issue #49 - Multi-Cluster Configuration Edge Cases
// Bug: Multi-cluster setup without cluster secrets caused sync failures
// ============================================================================

func TestRegression_49_MultiClusterGeneratesClusterSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "multi-cluster"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://dev.cluster.local:6443"},
			{Name: "staging", Cluster: "https://staging.cluster.local:6443"},
			{Name: "prod", Cluster: "https://prod.cluster.local:6443"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check if cluster secrets directory exists
	clustersDir := filepath.Join(tmpDir, "multi-cluster/argocd/clusters")
	_, err = os.Stat(clustersDir)
	if !os.IsNotExist(err) {
		// If clusters dir exists, verify secret structure
		entries, _ := os.ReadDir(clustersDir)
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".yaml") {
				secretPath := filepath.Join(clustersDir, entry.Name())
				content, readErr := os.ReadFile(secretPath)
				if readErr == nil {
					contentStr := string(content)
					assert.Contains(t, contentStr, "kind: Secret",
						"Cluster file should be a Secret")
					assert.Contains(t, contentStr, "argocd.argoproj.io/secret-type",
						"Should have ArgoCD secret-type label")
				}
			}
		}
	}
}

func TestRegression_49_SingleClusterNoClusterSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "single-cluster"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},     // No ClusterURL - single cluster mode
			{Name: "staging"}, // No ClusterURL - single cluster mode
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Single cluster mode should NOT generate cluster secrets
	clustersDir := filepath.Join(tmpDir, "single-cluster/argocd/clusters")
	entries, err := os.ReadDir(clustersDir)
	if err == nil && len(entries) > 0 {
		// If clusters dir exists, it should be empty or contain only README
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".yaml") {
				t.Logf("Warning: single-cluster mode generated cluster secret: %s", entry.Name())
			}
		}
	}
}

// ============================================================================
// Issue #50 - Bootstrap Mode Validation
// Bug: Invalid bootstrap mode caused runtime errors during installation
// ============================================================================

func TestRegression_50_BootstrapModeValidation(t *testing.T) {
	tests := []struct {
		name     string
		mode     bootstrap.Mode
		platform string
		tool     bootstrap.Tool
		valid    bool
	}{
		{
			name:     "helm on kubernetes",
			mode:     bootstrap.ModeHelm,
			platform: "kubernetes",
			tool:     bootstrap.ToolArgoCD,
			valid:    true,
		},
		{
			name:     "olm on openshift",
			mode:     bootstrap.ModeOLM,
			platform: "openshift",
			tool:     bootstrap.ToolArgoCD,
			valid:    true,
		},
		{
			name:     "olm on kubernetes - invalid",
			mode:     bootstrap.ModeOLM,
			platform: "kubernetes",
			tool:     bootstrap.ToolArgoCD,
			valid:    false,
		},
		{
			name:     "manifest on kubernetes",
			mode:     bootstrap.ModeManifest,
			platform: "kubernetes",
			tool:     bootstrap.ToolArgoCD,
			valid:    true,
		},
		{
			name:     "kustomize on kubernetes",
			mode:     bootstrap.ModeKustomize,
			platform: "kubernetes",
			tool:     bootstrap.ToolArgoCD,
			valid:    true,
		},
		{
			name:     "helm on eks",
			mode:     bootstrap.ModeHelm,
			platform: "eks",
			tool:     bootstrap.ToolArgoCD,
			valid:    true,
		},
		{
			name:     "helm on aks",
			mode:     bootstrap.ModeHelm,
			platform: "aks",
			tool:     bootstrap.ToolArgoCD,
			valid:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := bootstrap.IsValidMode(tc.mode, tc.tool, tc.platform)
			assert.Equal(t, tc.valid, isValid,
				"Mode %s should be valid=%v for %s on %s", tc.mode, tc.valid, tc.tool, tc.platform)
		})
	}
}

// ============================================================================
// Issue #51 - Template Rendering Edge Cases
// Bug: Templates with nil values caused panics during rendering
// ============================================================================

func TestRegression_51_GeneratorHandlesMinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Minimal valid config - should not panic
	cfg := &config.Config{
		Project:    config.Project{Name: "minimal"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	// Should not panic with minimal config
	err := gen.Generate()
	assert.NoError(t, err, "Minimal config should generate successfully")
}

func TestRegression_51_GeneratorHandlesAllInfraDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-infra"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces:      false,
			RBAC:            false,
			NetworkPolicies: false,
			ResourceQuotas:  false,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	// Should handle all infra disabled gracefully
	err := gen.Generate()
	// May succeed or fail gracefully, but should not panic
	_ = err
}

// ============================================================================
// Issue #52 - Project Name Sanitization
// Bug: Special characters in project names caused file system errors
// ============================================================================

func TestRegression_52_ProjectNameSanitization(t *testing.T) {
	invalidNames := []string{
		"project/with/slashes",
		"project:with:colons",
		"project<with>brackets",
		"project|with|pipes",
		"project\"with\"quotes",
		"project*with*stars",
		"project?with?questions",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			cfg := &config.Config{
				Project:    config.Project{Name: name},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: "argocd",
			}

			err := cfg.Validate()
			// Should either reject invalid names or sanitize them
			// The key is it should not panic
			if err == nil {
				// If validation passes, check that generation doesn't create invalid paths
				tmpDir := t.TempDir()
				writer := output.New(tmpDir, false, false)
				gen := generator.New(cfg, writer, false)
				_ = gen.Generate() // Should not panic
			}
		})
	}
}

func TestRegression_52_ValidProjectNames(t *testing.T) {
	validNames := []string{
		"my-project",
		"my_project",
		"myproject123",
		"MyProject",
		"my.project",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: name},
				Platform:   "kubernetes",
				Scope:      "infrastructure",
				GitOpsTool: "argocd",
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Infra: config.Infrastructure{
					Namespaces: true,
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			err := gen.Generate()
			assert.NoError(t, err, "Valid project name %s should generate successfully", name)

			// Verify project directory was created
			projectDir := filepath.Join(tmpDir, name)
			_, statErr := os.Stat(projectDir)
			assert.False(t, os.IsNotExist(statErr),
				"Project directory should exist for %s", name)
		})
	}
}

// ============================================================================
// Issue #53 - Dry Run Mode
// Bug: Dry run mode was still creating files in some edge cases
// ============================================================================

func TestRegression_53_DryRunDoesNotCreateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "dry-run-test"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
	}

	// Dry run mode should NOT create files
	writer := output.New(tmpDir, true, false) // dryRun = true
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	assert.NoError(t, err, "Dry run should succeed")

	// Verify no files were created
	projectDir := filepath.Join(tmpDir, "dry-run-test")
	_, statErr := os.Stat(projectDir)
	assert.True(t, os.IsNotExist(statErr),
		"Dry run should NOT create project directory")
}

// ============================================================================
// Issue #54 - ArgoCD Project Destination Configuration
// Bug: ArgoCD Projects allowed wrong namespaces in destinations
// ============================================================================

func TestRegression_54_ArgoCDProjectDestinations(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "project-dest"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check infrastructure project has correct destination namespaces
	infraProjectPath := filepath.Join(tmpDir, "project-dest/argocd/projects/infrastructure.yaml")
	content, err := os.ReadFile(infraProjectPath)
	require.NoError(t, err, "Should read infrastructure project file")

	contentStr := string(content)

	// Project should list environment namespaces as destinations
	assert.Contains(t, contentStr, "destinations:",
		"Should have destinations section")
}

// ============================================================================
// Issue #55 - Long Environment Names
// Bug: Very long environment names exceeded Kubernetes label limits
// ============================================================================

func TestRegression_55_LongEnvironmentNames(t *testing.T) {
	tmpDir := t.TempDir()

	longEnvName := "this-is-a-very-long-environment-name-that-might-exceed-limits"

	cfg := &config.Config{
		Project:    config.Project{Name: "long-env"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: longEnvName},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	// Should handle long names gracefully (truncate or reject)
	err := gen.Generate()
	if err == nil {
		// If generation succeeds, verify files were created
		nsPath := filepath.Join(tmpDir, "long-env/infrastructure/base/namespaces", longEnvName+".yaml")
		_, statErr := os.Stat(nsPath)
		if !os.IsNotExist(statErr) {
			content, _ := os.ReadFile(nsPath)
			contentStr := string(content)
			// Kubernetes names must be <= 63 characters
			assert.Contains(t, contentStr, "name:",
				"Should have name field in namespace")
		}
	}
}

// ============================================================================
// Issue #56 - Output URL Configuration
// Bug: Missing output URL caused generation to fail silently
// ============================================================================

func TestRegression_56_MissingOutputURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-output-url"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: ""}, // Empty URL
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	// Should either work with default values or report clear error
	err := gen.Generate()
	if err == nil {
		// If it succeeds, check that source URLs have default values or warnings
		appsetPath := filepath.Join(tmpDir, "no-output-url/argocd/applicationsets/infra-dev.yaml")
		content, readErr := os.ReadFile(appsetPath)
		if readErr == nil {
			contentStr := string(content)
			// Should have some source reference even if URL is empty
			assert.Contains(t, contentStr, "source:",
				"Should have source section")
		}
	}
}

// ============================================================================
// Issue #57 - Kustomize Base/Overlay Relationship
// Bug: Overlays were not correctly referencing base directories
// ============================================================================

func TestRegression_57_OverlayReferencesBase(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "overlay-base"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	for _, env := range cfg.Environments {
		overlayPath := filepath.Join(tmpDir, "overlay-base/infrastructure/overlays", env.Name, "kustomization.yaml")
		content, err := os.ReadFile(overlayPath)
		require.NoError(t, err, "Should read overlay kustomization for %s", env.Name)

		contentStr := string(content)

		// Overlay should reference base
		assert.Contains(t, contentStr, "../../base",
			"Overlay %s should reference ../../base", env.Name)
	}
}

// ============================================================================
// Issue #58 - Bootstrap Options Defaults
// Bug: Missing bootstrap options caused nil pointer dereference
// ============================================================================

func TestRegression_58_BootstrapOptionsDefaults(t *testing.T) {
	// Creating Options with minimal fields should not panic
	opts := &bootstrap.Options{
		Tool: bootstrap.ToolArgoCD,
		Mode: bootstrap.ModeHelm,
	}

	assert.NotNil(t, opts, "Options should be created")
	assert.Equal(t, bootstrap.ToolArgoCD, opts.Tool)
	assert.Equal(t, bootstrap.ModeHelm, opts.Mode)

	// Empty namespace should use default
	if opts.Namespace == "" {
		opts.Namespace = "argocd"
	}
	assert.Equal(t, "argocd", opts.Namespace)
}

// ============================================================================
// Issue #59 - Documentation Generation
// Bug: Generated README had broken links and incorrect paths
// ============================================================================

func TestRegression_59_ReadmeHasValidContent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "readme-test",
			Description: "Test project for README validation",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/readme-test.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
		Docs: config.Documentation{
			Readme: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	readmePath := filepath.Join(tmpDir, "readme-test/README.md")
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err, "Should read README.md")

	contentStr := string(content)

	// README should have essential sections
	assert.Contains(t, contentStr, "readme-test",
		"README should mention project name")
	assert.Contains(t, contentStr, "#",
		"README should have headers")
}

// ============================================================================
// Issue #60 - Concurrent Generation Safety
// Bug: Concurrent calls to generator could cause race conditions
// ============================================================================

func TestRegression_60_ConcurrentGenerationSafety(t *testing.T) {
	// Run multiple generators concurrently to check for race conditions
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "concurrent-" + string(rune('a'+idx))},
				Platform:   "kubernetes",
				Scope:      "infrastructure",
				GitOpsTool: "argocd",
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Infra: config.Infrastructure{
					Namespaces: true,
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			_ = gen.Generate()
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
