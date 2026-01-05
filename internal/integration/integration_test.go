package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/environment"
	"github.com/ihsanmokhlisse/gitopsi/internal/generator"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
	"github.com/ihsanmokhlisse/gitopsi/internal/validate"
)

func TestIntegration_InitFlow_ConfigToGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: integration-test
  description: Integration test project

platform: kubernetes
scope: both
gitops_tool: argocd

git:
  url: https://github.com/test/integration-test.git

environments:
  - name: dev
  - name: staging
  - name: prod

applications:
  - name: api
    image: api:v1
    port: 8080
    replicas: 2

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true

docs:
  readme: true
  architecture: true
`

	configPath := filepath.Join(tmpDir, "gitopsi.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Should write config file")

	cfg, err := config.Load(configPath)
	require.NoError(t, err, "Should load config from file")

	assert.Equal(t, "integration-test", cfg.Project.Name)
	assert.Equal(t, "kubernetes", cfg.Platform)
	assert.Equal(t, 3, len(cfg.Environments))

	outputDir := filepath.Join(tmpDir, "output")
	writer := output.New(outputDir, false, false)
	gen := generator.New(cfg, writer, false)

	err = gen.Generate()
	require.NoError(t, err, "Generation should succeed")

	expectedFiles := []string{
		"integration-test/README.md",
		"integration-test/infrastructure/base/kustomization.yaml",
		"integration-test/infrastructure/base/namespaces/dev.yaml",
		"integration-test/infrastructure/overlays/dev/kustomization.yaml",
		"integration-test/applications/base/api/deployment.yaml",
		"integration-test/applications/base/api/service.yaml",
		"integration-test/argocd/projects/infrastructure.yaml",
		"integration-test/argocd/applicationsets/infra-dev.yaml",
		"integration-test/scripts/bootstrap.sh",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(outputDir, file)
		_, err := os.Stat(fullPath)
		assert.False(t, os.IsNotExist(err), "File should exist: %s", file)
	}
}

func TestIntegration_InitFlow_OpenShiftPlatform(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "openshift-test"},
		Platform:   "openshift",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Apps: []config.Application{
			{Name: "web", Image: "nginx:latest", Port: 8080, Replicas: 1},
		},
		Infra: config.Infrastructure{Namespaces: true},
		Docs:  config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generation should succeed for OpenShift")

	projectFile := filepath.Join(tmpDir, "openshift-test/argocd/projects/infrastructure.yaml")
	content, err := os.ReadFile(projectFile)
	require.NoError(t, err, "Should read project file")

	assert.Contains(t, string(content), "namespace: openshift-gitops",
		"OpenShift should use openshift-gitops namespace")
}

func TestIntegration_InitFlow_AllPlatforms(t *testing.T) {
	platforms := []struct {
		name              string
		expectedNamespace string
	}{
		{"kubernetes", "argocd"},
		{"openshift", "openshift-gitops"},
		{"aks", "argocd"},
		{"eks", "argocd"},
	}

	for _, tc := range platforms {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "platform-" + tc.name},
				Platform:   tc.name,
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Docs: config.Documentation{Readme: true},
			}

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			err := gen.Generate()
			require.NoError(t, err, "Generation should succeed for %s", tc.name)

			projectFile := filepath.Join(tmpDir, "platform-"+tc.name+"/argocd/projects/infrastructure.yaml")
			content, err := os.ReadFile(projectFile)
			require.NoError(t, err, "Should read project file for %s", tc.name)

			assert.Contains(t, string(content), "namespace: "+tc.expectedNamespace,
				"Platform %s should use namespace %s", tc.name, tc.expectedNamespace)
		})
	}
}

func TestIntegration_InitFlow_WithPresets(t *testing.T) {
	presets := []config.Preset{config.PresetMinimal, config.PresetStandard, config.PresetEnterprise}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := config.NewDefaultConfig()
			cfg.Project.Name = "preset-" + string(preset)
			cfg.Platform = "kubernetes"
			cfg.GitOpsTool = "argocd"
			cfg.Git.URL = "https://github.com/test/preset-test.git"
			cfg.Environments = []config.Environment{{Name: "dev"}}
			cfg.Preset = preset
			cfg.ApplyPreset()

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			err := gen.Generate()
			require.NoError(t, err, "Generation should succeed for preset %s", preset)

			projectDir := filepath.Join(tmpDir, "preset-"+string(preset))
			_, err = os.Stat(projectDir)
			assert.False(t, os.IsNotExist(err), "Project directory should exist for preset %s", preset)
		})
	}
}

func TestIntegration_ValidateFlow_GeneratedManifests(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "validate-test"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/validate-test.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{
			{Name: "app", Image: "nginx:latest", Port: 80, Replicas: 1},
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
	require.NoError(t, err, "Generation should succeed")

	projectPath := filepath.Join(tmpDir, "validate-test")

	validator := validate.New(&validate.Options{
		Path:   projectPath,
		FailOn: validate.SeverityHigh,
	})

	ctx := context.Background()
	result, err := validator.Validate(ctx)
	require.NoError(t, err, "Validation should complete")

	assert.NotNil(t, result, "Validation result should not be nil")
	assert.Greater(t, result.TotalManifests, 0, "Should find files to validate")

	assert.False(t, validator.ShouldFail(result),
		"Generated manifests should not have critical errors")
}

func TestIntegration_ValidateFlow_WithSecurityChecks(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "security-test"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/security-test.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{
			{Name: "secure-app", Image: "nginx:latest", Port: 80, Replicas: 1},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err, "Generation should succeed")

	projectPath := filepath.Join(tmpDir, "security-test")

	validator := validate.New(&validate.Options{
		Path:     projectPath,
		FailOn:   validate.SeverityHigh,
		Security: true,
	})

	ctx := context.Background()
	result, err := validator.Validate(ctx)
	require.NoError(t, err, "Validation should complete")

	securityResult := result.Categories[validate.CategorySecurity]
	assert.NotNil(t, securityResult, "Security category results should be present")
}

func TestIntegration_EnvironmentFlow_CreateAndManage(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := environment.NewManager(tmpDir)

	err := mgr.CreateEnvironment("development", environment.CreateEnvOptions{
		Namespace: "dev-ns",
	})
	require.NoError(t, err, "Should create development environment")

	err = mgr.CreateEnvironment("staging", environment.CreateEnvOptions{
		Namespace: "staging-ns",
	})
	require.NoError(t, err, "Should create staging environment")

	err = mgr.CreateEnvironment("production", environment.CreateEnvOptions{
		Namespace: "prod-ns",
	})
	require.NoError(t, err, "Should create production environment")

	envs := mgr.ListEnvironments()
	assert.Len(t, envs, 3, "Should have 3 environments")

	dev := mgr.GetEnvironment("development")
	require.NotNil(t, dev, "Should get development environment")
	assert.Equal(t, "development", dev.Name)
	assert.Equal(t, "dev-ns", dev.Namespace)

	err = mgr.AddClusterToEnvironment("production", environment.ClusterInfo{
		Name:    "prod-cluster-1",
		URL:     "https://prod-1.k8s.local:6443",
		Primary: true,
	})
	require.NoError(t, err, "Should add cluster to production")

	prod := mgr.GetEnvironment("production")
	require.NotNil(t, prod, "Should get production environment")
	assert.Len(t, prod.Clusters, 1, "Production should have 1 cluster")
	assert.Equal(t, "prod-cluster-1", prod.Clusters[0].Name)

	err = mgr.DeleteEnvironment("staging")
	require.NoError(t, err, "Should delete staging environment")

	envs = mgr.ListEnvironments()
	assert.Len(t, envs, 2, "Should have 2 environments after deletion")
}

func TestIntegration_EnvironmentFlow_Promotion(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := environment.NewManager(tmpDir)

	err := mgr.CreateEnvironment("dev", environment.CreateEnvOptions{})
	require.NoError(t, err)

	err = mgr.CreateEnvironment("staging", environment.CreateEnvOptions{})
	require.NoError(t, err)

	err = mgr.CreateEnvironment("prod", environment.CreateEnvOptions{})
	require.NoError(t, err)

	result, err := mgr.Promote(environment.PromotionOptions{
		Application: "my-app",
		FromEnv:     "dev",
		ToEnv:       "staging",
	})
	require.NoError(t, err, "Promotion should succeed")

	assert.Equal(t, "my-app", result.Application)
	assert.Equal(t, "dev", result.FromEnv)
	assert.Equal(t, "staging", result.ToEnv)
}

func TestIntegration_FullWorkflow_ConfigToValidation(t *testing.T) {
	tmpDir := t.TempDir()

	configYAML := `
project:
  name: full-workflow
  description: Full integration test

platform: kubernetes
scope: both
gitops_tool: argocd

git:
  url: https://github.com/test/full-workflow.git
  branch: main

environments:
  - name: dev
    namespace: full-workflow-dev
  - name: staging
    namespace: full-workflow-staging
  - name: prod
    namespace: full-workflow-prod

applications:
  - name: frontend
    image: nginx:1.25
    port: 80
    replicas: 2
  - name: backend
    image: node:20-alpine
    port: 3000
    replicas: 3

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

docs:
  readme: true
  architecture: true
  onboarding: true
`

	configPath := filepath.Join(tmpDir, "gitopsi.yaml")
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	cfg, err := config.Load(configPath)
	require.NoError(t, err, "Should load config")

	outputDir := filepath.Join(tmpDir, "output")
	writer := output.New(outputDir, false, false)
	gen := generator.New(cfg, writer, false)

	err = gen.Generate()
	require.NoError(t, err, "Generation should succeed")

	projectPath := filepath.Join(outputDir, "full-workflow")

	var files []string
	err = filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(projectPath, path)
			files = append(files, relPath)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Greater(t, len(files), 20, "Should generate many files")

	validator := validate.New(&validate.Options{
		Path:   projectPath,
		FailOn: validate.SeverityHigh,
	})

	ctx := context.Background()
	result, err := validator.Validate(ctx)
	require.NoError(t, err, "Validation should complete")
	assert.False(t, validator.ShouldFail(result), "Should pass validation")

	readmePath := filepath.Join(projectPath, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Contains(t, string(readmeContent), "full-workflow")
	assert.Contains(t, string(readmeContent), "kubernetes")

	infraKustomization := filepath.Join(projectPath, "infrastructure/base/kustomization.yaml")
	infraContent, err := os.ReadFile(infraKustomization)
	require.NoError(t, err)
	assert.Contains(t, string(infraContent), "namespaces/")
	assert.Contains(t, string(infraContent), "rbac/")

	for _, env := range []string{"dev", "staging", "prod"} {
		nsFile := filepath.Join(projectPath, "infrastructure/base/namespaces", env+".yaml")
		nsContent, err := os.ReadFile(nsFile)
		require.NoError(t, err, "Should read namespace file for %s", env)
		assert.Contains(t, string(nsContent), "kind: Namespace")
		assert.Contains(t, string(nsContent), "full-workflow-"+env)
	}

	for _, app := range []string{"frontend", "backend"} {
		deployFile := filepath.Join(projectPath, "applications/base", app, "deployment.yaml")
		deployContent, err := os.ReadFile(deployFile)
		require.NoError(t, err, "Should read deployment for %s", app)
		assert.Contains(t, string(deployContent), "kind: Deployment")
		assert.Contains(t, string(deployContent), "name: "+app)
	}
}

func TestIntegration_MultiEnvironment_Topologies(t *testing.T) {
	topologies := []struct {
		name     string
		topology environment.Topology
	}{
		{"namespace-based", environment.TopologyNamespaceBased},
		{"cluster-per-env", environment.TopologyClusterPerEnv},
		{"multi-cluster", environment.TopologyMultiCluster},
	}

	for _, tc := range topologies {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			mgr := environment.NewManager(tmpDir)

			err := mgr.SetTopology(tc.topology)
			require.NoError(t, err, "Should set topology %s", tc.name)

			err = mgr.CreateEnvironment("test-env", environment.CreateEnvOptions{
				Namespace: "test-ns",
			})
			require.NoError(t, err, "Should create environment with topology %s", tc.name)

			cfg := mgr.Config()
			assert.Equal(t, tc.topology, cfg.Topology)
		})
	}
}

func TestIntegration_GitOpsToolSelection(t *testing.T) {
	tools := []string{"argocd", "flux"}

	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			// TODO: Flux support is disabled - focus on ArgoCD first
			if tool == "flux" {
				t.Skip("Flux support is disabled - focusing on ArgoCD first")
			}
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "tool-" + tool},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: tool,
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Docs: config.Documentation{Readme: true},
			}

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			err := gen.Generate()
			require.NoError(t, err, "Generation should succeed for %s", tool)

			toolDir := filepath.Join(tmpDir, "tool-"+tool, tool)
			_, err = os.Stat(toolDir)
			assert.False(t, os.IsNotExist(err), "%s directory should exist", tool)
		})
	}
}

func TestIntegration_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		shouldError bool
	}{
		{
			name: "valid minimal config",
			config: `
project:
  name: valid-minimal
platform: kubernetes
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
`,
			shouldError: false,
		},
		{
			name: "missing project name",
			config: `
platform: kubernetes
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
`,
			shouldError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configPath, []byte(tc.config), 0644)
			require.NoError(t, err)

			cfg, err := config.Load(configPath)

			if tc.shouldError {
				if err == nil && cfg != nil {
					err = cfg.Validate()
				}
				assert.Error(t, err, "Should error for: %s", tc.name)
			} else {
				require.NoError(t, err, "Should not error for: %s", tc.name)
			}
		})
	}
}

func TestIntegration_OutputFormats(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "output-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/output-test.git"},
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
	require.NoError(t, err)

	projectPath := filepath.Join(tmpDir, "output-test")

	validator := validate.New(&validate.Options{
		Path:   projectPath,
		FailOn: validate.SeverityHigh,
	})

	ctx := context.Background()
	result, err := validator.Validate(ctx)
	require.NoError(t, err)

	jsonOutput, err := result.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonOutput, "path")
	assert.Contains(t, jsonOutput, "total_manifests")

	yamlOutput, err := result.ToYAML()
	require.NoError(t, err)
	assert.Contains(t, yamlOutput, "path:")
}

func TestIntegration_KustomizationStructure(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "kustomize-struct"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/kustomize-struct.git"},
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
	require.NoError(t, err)

	baseKustomization := filepath.Join(tmpDir, "kustomize-struct/infrastructure/base/kustomization.yaml")
	baseContent, err := os.ReadFile(baseKustomization)
	require.NoError(t, err)

	var kustomization map[string]interface{}
	err = yaml.Unmarshal(baseContent, &kustomization)
	require.NoError(t, err, "Should parse kustomization.yaml")

	assert.Contains(t, kustomization, "apiVersion")
	assert.Contains(t, kustomization, "kind")
	assert.Contains(t, kustomization, "resources")

	subdirs := []string{"namespaces", "rbac", "network-policies", "resource-quotas"}
	for _, subdir := range subdirs {
		kustomizePath := filepath.Join(tmpDir, "kustomize-struct/infrastructure/base", subdir, "kustomization.yaml")
		_, err := os.Stat(kustomizePath)
		assert.False(t, os.IsNotExist(err), "Subdirectory %s should have kustomization.yaml", subdir)
	}

	for _, env := range []string{"dev", "staging", "prod"} {
		overlayPath := filepath.Join(tmpDir, "kustomize-struct/infrastructure/overlays", env, "kustomization.yaml")
		overlayContent, err := os.ReadFile(overlayPath)
		require.NoError(t, err, "Should read %s overlay", env)
		assert.Contains(t, string(overlayContent), "resources:")
	}
}

func TestIntegration_ApplicationGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	apps := []config.Application{
		{Name: "frontend", Image: "nginx:1.25", Port: 80, Replicas: 2},
		{Name: "backend", Image: "node:20", Port: 3000, Replicas: 3},
		{Name: "api", Image: "python:3.11", Port: 8000, Replicas: 2},
	}

	cfg := &config.Config{
		Project:    config.Project{Name: "multi-app"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/multi-app.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: apps,
	}

	writer := output.New(tmpDir, false, false)
	gen := generator.New(cfg, writer, false)

	err := gen.Generate()
	require.NoError(t, err)

	for _, app := range apps {
		deployPath := filepath.Join(tmpDir, "multi-app/applications/base", app.Name, "deployment.yaml")
		deployContent, err := os.ReadFile(deployPath)
		require.NoError(t, err, "Should read deployment for %s", app.Name)

		assert.Contains(t, string(deployContent), "name: "+app.Name)
		assert.Contains(t, string(deployContent), "image: "+app.Image)

		svcPath := filepath.Join(tmpDir, "multi-app/applications/base", app.Name, "service.yaml")
		svcContent, err := os.ReadFile(svcPath)
		require.NoError(t, err, "Should read service for %s", app.Name)
		assert.Contains(t, string(svcContent), "kind: Service")

		kustomizePath := filepath.Join(tmpDir, "multi-app/applications/base", app.Name, "kustomization.yaml")
		_, err = os.Stat(kustomizePath)
		assert.False(t, os.IsNotExist(err), "%s should have kustomization.yaml", app.Name)
	}
}

func TestIntegration_ScopeSelection(t *testing.T) {
	scopes := []struct {
		name     string
		scope    string
		hasInfra bool
		hasApps  bool
	}{
		{"infrastructure only", "infrastructure", true, false},
		{"application only", "application", false, true},
		{"both", "both", true, true},
	}

	for _, tc := range scopes {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "scope-" + tc.scope},
				Platform:   "kubernetes",
				Scope:      tc.scope,
				GitOpsTool: "argocd",
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Apps: []config.Application{
					{Name: "app", Image: "nginx", Port: 80, Replicas: 1},
				},
				Infra: config.Infrastructure{Namespaces: true},
			}

			writer := output.New(tmpDir, false, false)
			gen := generator.New(cfg, writer, false)

			err := gen.Generate()
			require.NoError(t, err, "Generation should succeed for scope %s", tc.scope)

			infraPath := filepath.Join(tmpDir, "scope-"+tc.scope, "infrastructure/base")
			_, infraErr := os.Stat(infraPath)

			appsPath := filepath.Join(tmpDir, "scope-"+tc.scope, "applications/base")
			_, appsErr := os.Stat(appsPath)

			if tc.hasInfra {
				assert.False(t, os.IsNotExist(infraErr), "Scope %s should have infrastructure", tc.scope)
			}
			if tc.hasApps {
				assert.False(t, os.IsNotExist(appsErr), "Scope %s should have applications", tc.scope)
			}
		})
	}
}
