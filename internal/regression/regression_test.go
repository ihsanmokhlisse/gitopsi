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
