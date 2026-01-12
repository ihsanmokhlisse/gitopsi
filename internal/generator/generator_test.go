package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
)

const testGitURL = "https://github.com/test/repo.git"

func TestNew(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "test"

	writer := output.New("/tmp", false, false)
	gen := New(cfg, writer, false)

	if gen == nil {
		t.Fatal("New() returned nil")
	}

	if gen.Config != cfg {
		t.Error("New() config mismatch")
	}

	if gen.Writer != writer {
		t.Error("New() writer mismatch")
	}
}

func TestGenerateDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-project"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Apps: []config.Application{
			{Name: "web", Image: "nginx:latest", Port: 80, Replicas: 1},
		},
		Docs: config.Documentation{
			Readme:       true,
			Architecture: true,
		},
	}

	writer := output.New(tmpDir, true, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	projectDir := filepath.Join(tmpDir, "test-project")
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Error("Dry run should not create directories")
	}
}

func TestGenerateStructure(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-structure"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateStructure()
	if err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	expectedDirs := []string{
		"test-structure",
		"test-structure/docs",
		"test-structure/bootstrap/argocd",
		"test-structure/scripts",
		"test-structure/infrastructure/base",
		"test-structure/infrastructure/overlays/dev",
		"test-structure/infrastructure/overlays/staging",
		"test-structure/infrastructure/overlays/prod",
		"test-structure/applications/base",
		"test-structure/applications/overlays/dev",
		"test-structure/argocd/projects",
		"test-structure/argocd/applicationsets",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}
}

func TestGenerateInfrastructure(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-infra"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateInfrastructure()
	if err != nil {
		t.Fatalf("generateInfrastructure() error = %v", err)
	}

	expectedFiles := []string{
		"test-infra/infrastructure/base/namespaces/dev.yaml",
		"test-infra/infrastructure/base/namespaces/prod.yaml",
		"test-infra/infrastructure/base/namespaces/kustomization.yaml",
		"test-infra/infrastructure/base/kustomization.yaml",
		"test-infra/infrastructure/overlays/dev/kustomization.yaml",
		"test-infra/infrastructure/overlays/prod/kustomization.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}

	nsContent, err := os.ReadFile(filepath.Join(tmpDir, "test-infra/infrastructure/base/namespaces/dev.yaml"))
	if err != nil {
		t.Fatalf("Failed to read namespace file: %v", err)
	}

	if !strings.Contains(string(nsContent), "kind: Namespace") {
		t.Error("Namespace file missing 'kind: Namespace'")
	}

	// Verify subdirectory kustomization.yaml has correct resources
	nsKustomization, err := os.ReadFile(filepath.Join(tmpDir, "test-infra/infrastructure/base/namespaces/kustomization.yaml"))
	if err != nil {
		t.Fatalf("Failed to read namespaces kustomization.yaml: %v", err)
	}

	if !strings.Contains(string(nsKustomization), "dev.yaml") {
		t.Error("Namespaces kustomization.yaml missing 'dev.yaml'")
	}
	if !strings.Contains(string(nsKustomization), "prod.yaml") {
		t.Error("Namespaces kustomization.yaml missing 'prod.yaml'")
	}
}

func TestGenerateInfrastructureSubdirKustomization(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "subdir-kustomize"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
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
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateInfrastructure()
	if err != nil {
		t.Fatalf("generateInfrastructure() error = %v", err)
	}

	// Verify all subdirectory kustomization.yaml files are created
	subdirs := []string{"namespaces", "rbac", "network-policies", "resource-quotas"}
	for _, subdir := range subdirs {
		kustomizePath := filepath.Join(tmpDir, "subdir-kustomize/infrastructure/base", subdir, "kustomization.yaml")
		if _, err := os.Stat(kustomizePath); os.IsNotExist(err) {
			t.Errorf("Subdirectory kustomization.yaml not created: %s", kustomizePath)
		}

		content, err := os.ReadFile(kustomizePath)
		if err != nil {
			t.Fatalf("Failed to read %s/kustomization.yaml: %v", subdir, err)
		}

		// Each subdirectory should list the environment files
		for _, env := range cfg.Environments {
			expectedFile := env.Name + ".yaml"
			if !strings.Contains(string(content), expectedFile) {
				t.Errorf("%s/kustomization.yaml missing resource: %s", subdir, expectedFile)
			}
		}
	}

	// Verify base kustomization.yaml references subdirectories (not individual files)
	baseKustomization, err := os.ReadFile(filepath.Join(tmpDir, "subdir-kustomize/infrastructure/base/kustomization.yaml"))
	if err != nil {
		t.Fatalf("Failed to read base kustomization.yaml: %v", err)
	}

	for _, subdir := range subdirs {
		if !strings.Contains(string(baseKustomization), subdir+"/") {
			t.Errorf("Base kustomization.yaml missing subdirectory reference: %s/", subdir)
		}
	}
}

func TestGenerateApplications(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-apps"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{
			{Name: "frontend", Image: "nginx:latest", Port: 80, Replicas: 2},
			{Name: "backend", Image: "node:18", Port: 3000, Replicas: 3},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateApplications()
	if err != nil {
		t.Fatalf("generateApplications() error = %v", err)
	}

	expectedFiles := []string{
		"test-apps/applications/base/frontend/deployment.yaml",
		"test-apps/applications/base/frontend/service.yaml",
		"test-apps/applications/base/frontend/kustomization.yaml",
		"test-apps/applications/base/backend/deployment.yaml",
		"test-apps/applications/base/backend/service.yaml",
		"test-apps/applications/base/kustomization.yaml",
		"test-apps/applications/overlays/dev/kustomization.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}

	deployContent, err := os.ReadFile(filepath.Join(tmpDir, "test-apps/applications/base/frontend/deployment.yaml"))
	if err != nil {
		t.Fatalf("Failed to read deployment file: %v", err)
	}

	checks := []string{"kind: Deployment", "name: frontend", "image: nginx:latest", "replicas: 2"}
	for _, check := range checks {
		if !strings.Contains(string(deployContent), check) {
			t.Errorf("Deployment missing: %s", check)
		}
	}
}

func TestGenerateArgoCD(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-argocd"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://dev.k8s.local"},
			{Name: "prod", Cluster: "https://prod.k8s.local"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	expectedFiles := []string{
		"test-argocd/argocd/projects/infrastructure.yaml",
		"test-argocd/argocd/projects/applications.yaml",
		"test-argocd/argocd/applicationsets/infra-dev.yaml",
		"test-argocd/argocd/applicationsets/apps-dev.yaml",
		"test-argocd/argocd/applicationsets/infra-prod.yaml",
		"test-argocd/argocd/applicationsets/apps-prod.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}

	projectContent, err := os.ReadFile(filepath.Join(tmpDir, "test-argocd/argocd/projects/infrastructure.yaml"))
	if err != nil {
		t.Fatalf("Failed to read project file: %v", err)
	}

	if !strings.Contains(string(projectContent), "kind: AppProject") {
		t.Error("Project file missing 'kind: AppProject'")
	}

	if !strings.Contains(string(projectContent), "namespace: argocd") {
		t.Error("Project file should use 'namespace: argocd' for kubernetes platform")
	}
}

func TestGenerateArgoCDOpenShiftNamespace(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-openshift"},
		Platform:   "openshift",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://api.ocp.local:6443"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	projectContent, err := os.ReadFile(filepath.Join(tmpDir, "test-openshift/argocd/projects/infrastructure.yaml"))
	if err != nil {
		t.Fatalf("Failed to read project file: %v", err)
	}

	if !strings.Contains(string(projectContent), "namespace: openshift-gitops") {
		t.Errorf("OpenShift platform should use 'namespace: openshift-gitops', got: %s", string(projectContent))
	}

	appContent, err := os.ReadFile(filepath.Join(tmpDir, "test-openshift/argocd/applicationsets/infra-dev.yaml"))
	if err != nil {
		t.Fatalf("Failed to read application file: %v", err)
	}

	if !strings.Contains(string(appContent), "namespace: openshift-gitops") {
		t.Errorf("OpenShift applications should use 'namespace: openshift-gitops', got: %s", string(appContent))
	}
}

func TestGenerateArgoCDCustomNamespace(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-custom-ns"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Bootstrap: config.BootstrapConfig{
			Namespace: "custom-argocd-ns",
		},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	projectContent, err := os.ReadFile(filepath.Join(tmpDir, "test-custom-ns/argocd/projects/infrastructure.yaml"))
	if err != nil {
		t.Fatalf("Failed to read project file: %v", err)
	}

	if !strings.Contains(string(projectContent), "namespace: custom-argocd-ns") {
		t.Errorf("Custom namespace should be used, got: %s", string(projectContent))
	}
}

func TestGenerateDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "test-docs",
			Description: "Test project description",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{
			Readme: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateDocs()
	if err != nil {
		t.Fatalf("generateDocs() error = %v", err)
	}

	readmePath := filepath.Join(tmpDir, "test-docs/README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Fatal("README.md not created")
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README: %v", err)
	}

	checks := []string{"test-docs", "kubernetes", "argocd"}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("README missing: %s", check)
		}
	}
}

func TestGenerateBootstrap(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-bootstrap"},
		GitOpsTool: "argocd",
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := os.MkdirAll(filepath.Join(tmpDir, "test-bootstrap/bootstrap/argocd"), 0755); err != nil {
		t.Fatalf("Failed to create bootstrap dir: %v", err)
	}

	err := gen.generateBootstrap()
	if err != nil {
		t.Fatalf("generateBootstrap() error = %v", err)
	}

	nsPath := filepath.Join(tmpDir, "test-bootstrap/bootstrap/argocd/namespace.yaml")
	if _, err := os.Stat(nsPath); os.IsNotExist(err) {
		t.Fatal("Bootstrap namespace.yaml not created")
	}
}

func TestGenerateScripts(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "test-scripts"},
		GitOpsTool: "argocd",
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := os.MkdirAll(filepath.Join(tmpDir, "test-scripts/scripts"), 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}

	err := gen.generateScripts()
	if err != nil {
		t.Fatalf("generateScripts() error = %v", err)
	}

	expectedScripts := []string{
		"test-scripts/scripts/bootstrap.sh",
		"test-scripts/scripts/validate.sh",
	}

	for _, script := range expectedScripts {
		fullPath := filepath.Join(tmpDir, script)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Script not created: %s", script)
		}
	}
}

func TestGenerateFullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "full-test",
			Description: "Full workflow test",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
		Apps: []config.Application{
			{Name: "api", Image: "api:v1", Port: 8080, Replicas: 2},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
			RBAC:       true,
		},
		Docs: config.Documentation{
			Readme: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expectedDirs := []string{
		"full-test",
		"full-test/infrastructure/base",
		"full-test/infrastructure/overlays/dev",
		"full-test/applications/base/api",
		"full-test/argocd/projects",
		"full-test/bootstrap/argocd",
		"full-test/scripts",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Directory not created: %s", dir)
		}
	}

	expectedFiles := []string{
		"full-test/README.md",
		"full-test/infrastructure/base/kustomization.yaml",
		"full-test/applications/base/api/deployment.yaml",
		"full-test/argocd/projects/infrastructure.yaml",
		"full-test/scripts/bootstrap.sh",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("File not created: %s", file)
		}
	}
}

func TestGenerateInfrastructureOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "infra-only"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	infraPath := filepath.Join(tmpDir, "infra-only/infrastructure/base")
	if _, err := os.Stat(infraPath); os.IsNotExist(err) {
		t.Error("Infrastructure directory not created")
	}
}

func TestGenerateApplicationsOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "apps-only"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{
			{Name: "web", Image: "nginx", Port: 80, Replicas: 1},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	appsPath := filepath.Join(tmpDir, "apps-only/applications/base")
	if _, err := os.Stat(appsPath); os.IsNotExist(err) {
		t.Error("Applications directory not created")
	}
}

func TestGenerateWithoutDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-docs"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{Readme: false},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestGenerateVerbose(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "verbose-test"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, true)
	gen := New(cfg, writer, true)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestGenerateEmptyApps(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "empty-apps"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{},
		Docs: config.Documentation{Readme: false},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	sampleAppPath := filepath.Join(tmpDir, "empty-apps/applications/base/sample-app")
	if _, err := os.Stat(sampleAppPath); os.IsNotExist(err) {
		t.Error("Sample app should be created when apps list is empty")
	}
}

func TestGenerateMultipleApps(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "multi-apps"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
		Apps: []config.Application{
			{Name: "frontend", Image: "nginx:latest", Port: 80, Replicas: 2},
			{Name: "backend", Image: "node:18", Port: 3000, Replicas: 3},
			{Name: "api", Image: "python:3.11", Port: 8000, Replicas: 2},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, app := range cfg.Apps {
		appPath := filepath.Join(tmpDir, "multi-apps/applications/base", app.Name)
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			t.Errorf("App directory not created: %s", app.Name)
		}

		deployPath := filepath.Join(appPath, "deployment.yaml")
		if _, err := os.Stat(deployPath); os.IsNotExist(err) {
			t.Errorf("Deployment not created for: %s", app.Name)
		}
	}
}

func TestGenerateMultipleEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "multi-env"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://dev.k8s.local"},
			{Name: "qa", Cluster: "https://qa.k8s.local"},
			{Name: "staging", Cluster: "https://staging.k8s.local"},
			{Name: "prod", Cluster: "https://prod.k8s.local"},
		},
		Apps: []config.Application{
			{Name: "app", Image: "app:v1", Port: 80, Replicas: 1},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, env := range cfg.Environments {
		infraOverlay := filepath.Join(tmpDir, "multi-env/infrastructure/overlays", env.Name)
		if _, err := os.Stat(infraOverlay); os.IsNotExist(err) {
			t.Errorf("Infrastructure overlay not created for: %s", env.Name)
		}

		appOverlay := filepath.Join(tmpDir, "multi-env/applications/overlays", env.Name)
		if _, err := os.Stat(appOverlay); os.IsNotExist(err) {
			t.Errorf("Application overlay not created for: %s", env.Name)
		}
	}
}

func TestGenerateWithGitURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "git-url-test"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output: config.Output{
			Type:   "git",
			URL:    "https://github.com/myorg/myrepo.git",
			Branch: "main",
		},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	appFile := filepath.Join(tmpDir, "git-url-test/argocd/applicationsets/apps-dev.yaml")
	content, err := os.ReadFile(appFile)
	if err != nil {
		t.Fatalf("Failed to read ArgoCD app: %v", err)
	}

	if !strings.Contains(string(content), "https://github.com/myorg/myrepo.git") {
		t.Error("ArgoCD application should contain the git URL")
	}
}

func TestGenerateInfrastructureContent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "infra-content"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateInfrastructure()
	if err != nil {
		t.Fatalf("generateInfrastructure() error = %v", err)
	}

	nsFile := filepath.Join(tmpDir, "infra-content/infrastructure/base/namespaces/dev.yaml")
	content, err := os.ReadFile(nsFile)
	if err != nil {
		t.Fatalf("Failed to read namespace file: %v", err)
	}

	checks := []string{
		"apiVersion: v1",
		"kind: Namespace",
		"name: infra-content-dev",
		"env: dev",
	}

	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("Namespace file missing: %s", check)
		}
	}
}

func TestGenerateApplicationsContent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "apps-content"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{
			{Name: "myapp", Image: "myregistry/myapp:v1", Port: 9000, Replicas: 4},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateApplications()
	if err != nil {
		t.Fatalf("generateApplications() error = %v", err)
	}

	deployFile := filepath.Join(tmpDir, "apps-content/applications/base/myapp/deployment.yaml")
	content, err := os.ReadFile(deployFile)
	if err != nil {
		t.Fatalf("Failed to read deployment: %v", err)
	}

	checks := []string{
		"name: myapp",
		"image: myregistry/myapp:v1",
		"containerPort: 9000",
		"replicas: 4",
	}

	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("Deployment missing: %s", check)
		}
	}

	svcFile := filepath.Join(tmpDir, "apps-content/applications/base/myapp/service.yaml")
	svcContent, err := os.ReadFile(svcFile)
	if err != nil {
		t.Fatalf("Failed to read service: %v", err)
	}

	if !strings.Contains(string(svcContent), "port: 9000") {
		t.Error("Service should have port 9000")
	}
}

func TestGeneratorConfigFields(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "test"

	writer := output.New("/tmp", true, true)
	gen := New(cfg, writer, true)

	if gen.Config != cfg {
		t.Error("Config not set correctly")
	}
	if gen.Writer != writer {
		t.Error("Writer not set correctly")
	}
	if !gen.Verbose {
		t.Error("Verbose not set correctly")
	}
}

func TestGenerateAllPlatforms(t *testing.T) {
	platforms := []string{"kubernetes", "openshift", "aks", "eks"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "platform-" + platform},
				Platform:   platform,
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     config.Output{Type: "local"},
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Docs: config.Documentation{Readme: true},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() for %s error = %v", platform, err)
			}
		})
	}
}

func TestGenerateAllScopes(t *testing.T) {
	scopes := []string{"infrastructure", "application", "both"}

	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "scope-" + scope},
				Platform:   "kubernetes",
				Scope:      scope,
				GitOpsTool: "argocd",
				Output:     config.Output{Type: "local"},
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Apps: []config.Application{
					{Name: "app", Image: "nginx", Port: 80, Replicas: 1},
				},
				Docs: config.Documentation{Readme: true},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() for scope %s error = %v", scope, err)
			}
		})
	}
}

func TestGenerateFluxTool(t *testing.T) {
	// TODO: Flux support is disabled - focus on ArgoCD first
	t.Skip("Flux support is disabled - focusing on ArgoCD first")
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "flux-test"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "flux",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with flux error = %v", err)
	}

	fluxDir := filepath.Join(tmpDir, "flux-test/flux")
	if _, err := os.Stat(fluxDir); os.IsNotExist(err) {
		t.Error("Flux directory not created")
	}
}

func TestGenerateBothTools(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "both-tools"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "both",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with both tools error = %v", err)
	}
}

func TestGenerateWithAllDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "full-docs",
			Description: "Project with all documentation",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{
			Readme:       true,
			Architecture: true,
			Onboarding:   true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	readmePath := filepath.Join(tmpDir, "full-docs/README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md not created")
	}
}

func TestGenerateWithVerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "verbose-gen"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: []config.Application{
			{Name: "app", Image: "nginx", Port: 80, Replicas: 1},
		},
		Docs: config.Documentation{Readme: true},
	}

	writer := output.New(tmpDir, false, true)
	gen := New(cfg, writer, true)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with verbose error = %v", err)
	}
}

func TestGenerateStructureCreatesAllDirs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "structure-test"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Git:        config.GitConfig{URL: testGitURL},
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

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateStructure()
	if err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	expectedDirs := []string{
		"structure-test",
		"structure-test/docs",
		"structure-test/bootstrap/argocd",
		"structure-test/scripts",
		"structure-test/infrastructure/base",
		"structure-test/infrastructure/base/namespaces",
		"structure-test/infrastructure/base/rbac",
		"structure-test/infrastructure/base/network-policies",
		"structure-test/infrastructure/base/resource-quotas",
		"structure-test/infrastructure/overlays/dev",
		"structure-test/infrastructure/overlays/prod",
		"structure-test/applications/base",
		"structure-test/applications/overlays/dev",
		"structure-test/applications/overlays/prod",
		"structure-test/argocd/projects",
		"structure-test/argocd/applicationsets",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Directory not created: %s", dir)
		}
	}
}

func TestGenerateGitOpsForAllTools(t *testing.T) {
	tools := []string{"argocd", "flux"}

	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			// TODO: Flux support is disabled - focus on ArgoCD first
			if tool == "flux" {
				t.Skip("Flux support is disabled - focusing on ArgoCD first")
			}
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "gitops-" + tool},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: tool,
				Output:     config.Output{URL: "https://github.com/test/repo.git"},
				Environments: []config.Environment{
					{Name: "dev"},
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			if err := gen.generateStructure(); err != nil {
				t.Fatalf("generateStructure() error = %v", err)
			}

			err := gen.generateGitOps()
			if err != nil {
				t.Fatalf("generateGitOps() for %s error = %v", tool, err)
			}

			toolDir := filepath.Join(tmpDir, "gitops-"+tool, tool)
			if _, err := os.Stat(toolDir); os.IsNotExist(err) {
				t.Errorf("%s directory not created", tool)
			}
		})
	}
}

func TestGenerateRBAC(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "rbac-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
			RBAC:       true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateRBAC()
	if err != nil {
		t.Fatalf("generateRBAC() error = %v", err)
	}

	expectedFiles := []string{
		"rbac-test/infrastructure/base/rbac/dev.yaml",
		"rbac-test/infrastructure/base/rbac/prod.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("RBAC file not created: %s", file)
		}
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "rbac-test/infrastructure/base/rbac/dev.yaml"))
	if err != nil {
		t.Fatalf("Failed to read RBAC file: %v", err)
	}

	checks := []string{"kind: Role", "kind: RoleBinding", "rbac-test-dev"}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("RBAC file missing: %s", check)
		}
	}
}

func TestGenerateNetworkPolicies(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "np-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			NetworkPolicies: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateNetworkPolicies()
	if err != nil {
		t.Fatalf("generateNetworkPolicies() error = %v", err)
	}

	npFile := filepath.Join(tmpDir, "np-test/infrastructure/base/network-policies/dev.yaml")
	if _, err := os.Stat(npFile); os.IsNotExist(err) {
		t.Fatal("NetworkPolicy file not created")
	}

	content, err := os.ReadFile(npFile)
	if err != nil {
		t.Fatalf("Failed to read NetworkPolicy file: %v", err)
	}

	checks := []string{"kind: NetworkPolicy", "np-test-network-policy", "Ingress", "Egress"}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("NetworkPolicy file missing: %s", check)
		}
	}
}

func TestGenerateResourceQuotas(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "rq-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
			{Name: "custom"},
		},
		Infra: config.Infrastructure{
			Namespaces:     true,
			ResourceQuotas: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateResourceQuotas()
	if err != nil {
		t.Fatalf("generateResourceQuotas() error = %v", err)
	}

	expectedFiles := []string{
		"rq-test/infrastructure/base/resource-quotas/dev.yaml",
		"rq-test/infrastructure/base/resource-quotas/staging.yaml",
		"rq-test/infrastructure/base/resource-quotas/prod.yaml",
		"rq-test/infrastructure/base/resource-quotas/custom.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("ResourceQuota file not created: %s", file)
		}
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "rq-test/infrastructure/base/resource-quotas/prod.yaml"))
	if err != nil {
		t.Fatalf("Failed to read ResourceQuota file: %v", err)
	}

	if !strings.Contains(string(content), "kind: ResourceQuota") {
		t.Error("ResourceQuota file missing kind")
	}

	if !strings.Contains(string(content), "limits.cpu") {
		t.Error("ResourceQuota file missing limits.cpu")
	}
}

func TestGenerateAllInfrastructure(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "full-infra"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
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

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateInfrastructure()
	if err != nil {
		t.Fatalf("generateInfrastructure() error = %v", err)
	}

	expectedDirs := []string{
		"full-infra/infrastructure/base/namespaces",
		"full-infra/infrastructure/base/rbac",
		"full-infra/infrastructure/base/network-policies",
		"full-infra/infrastructure/base/resource-quotas",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Infrastructure directory not created: %s", dir)
		}
	}
}

func TestGenerateArchitectureDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "arch-test",
			Description: "Test project",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
		Docs: config.Documentation{
			Readme:       true,
			Architecture: true,
			Onboarding:   false,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateDocs()
	if err != nil {
		t.Fatalf("generateDocs() error = %v", err)
	}

	archFile := filepath.Join(tmpDir, "arch-test/docs/ARCHITECTURE.md")
	if _, err := os.Stat(archFile); os.IsNotExist(err) {
		t.Fatal("ARCHITECTURE.md not created")
	}

	content, err := os.ReadFile(archFile)
	if err != nil {
		t.Fatalf("Failed to read ARCHITECTURE.md: %v", err)
	}

	checks := []string{"arch-test", "kubernetes", "argocd", "dev", "prod", "Infrastructure Layer"}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("ARCHITECTURE.md missing: %s", check)
		}
	}
}

func TestGenerateOnboardingDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "onboard-test",
			Description: "Test project",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://dev.k8s.local"},
		},
		Docs: config.Documentation{
			Readme:       false,
			Architecture: false,
			Onboarding:   true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateDocs()
	if err != nil {
		t.Fatalf("generateDocs() error = %v", err)
	}

	onboardFile := filepath.Join(tmpDir, "onboard-test/docs/ONBOARDING.md")
	if _, err := os.Stat(onboardFile); os.IsNotExist(err) {
		t.Fatal("ONBOARDING.md not created")
	}

	content, err := os.ReadFile(onboardFile)
	if err != nil {
		t.Fatalf("Failed to read ONBOARDING.md: %v", err)
	}

	checks := []string{"onboard-test", "Prerequisites", "Quick Start", "bootstrap.sh", "argocd"}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("ONBOARDING.md missing: %s", check)
		}
	}
}

func TestGenerateAllDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "all-docs",
			Description: "Test project with all docs",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{
			Readme:       true,
			Architecture: true,
			Onboarding:   true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateDocs()
	if err != nil {
		t.Fatalf("generateDocs() error = %v", err)
	}

	expectedFiles := []string{
		"all-docs/README.md",
		"all-docs/docs/ARCHITECTURE.md",
		"all-docs/docs/ONBOARDING.md",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Doc file not created: %s", file)
		}
	}
}

func TestGenerateNoDocs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-docs"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{
			Readme:       false,
			Architecture: false,
			Onboarding:   false,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateDocs()
	if err != nil {
		t.Fatalf("generateDocs() error = %v", err)
	}

	notExpectedFiles := []string{
		"no-docs/README.md",
		"no-docs/docs/ARCHITECTURE.md",
		"no-docs/docs/ONBOARDING.md",
	}

	for _, file := range notExpectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			t.Errorf("Doc file should not be created when disabled: %s", file)
		}
	}
}

func TestGenerateCompleteProject(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "complete-project",
			Description: "A complete GitOps project",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output: config.Output{
			Type: "local",
			URL:  "https://github.com/test/repo.git",
		},
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://dev.k8s.local"},
			{Name: "staging", Cluster: "https://staging.k8s.local"},
			{Name: "prod", Cluster: "https://prod.k8s.local"},
		},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
		Apps: []config.Application{
			{Name: "frontend", Image: "nginx:latest", Port: 80, Replicas: 2},
			{Name: "backend", Image: "node:18", Port: 3000, Replicas: 3},
		},
		Docs: config.Documentation{
			Readme:       true,
			Architecture: true,
			Onboarding:   true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expectedItems := []string{
		"complete-project/README.md",
		"complete-project/docs/ARCHITECTURE.md",
		"complete-project/docs/ONBOARDING.md",
		"complete-project/infrastructure/base/namespaces/dev.yaml",
		"complete-project/infrastructure/base/rbac/dev.yaml",
		"complete-project/infrastructure/base/network-policies/dev.yaml",
		"complete-project/infrastructure/base/resource-quotas/dev.yaml",
		"complete-project/applications/base/frontend/deployment.yaml",
		"complete-project/applications/base/backend/deployment.yaml",
		"complete-project/argocd/projects/infrastructure.yaml",
		"complete-project/argocd/applicationsets/infra-dev.yaml",
		"complete-project/scripts/bootstrap.sh",
		"complete-project/scripts/validate.sh",
	}

	for _, item := range expectedItems {
		fullPath := filepath.Join(tmpDir, item)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected item not created: %s", item)
		}
	}
}

// Multi-cluster ArgoCD Tests

func TestGenerateMultiClusterArgoCD_ClusterPerEnv(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "mc-cluster-per-env"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{
				Name: "nonprod",
				Clusters: []config.EnvironmentCluster{
					{Name: "dev-cluster", URL: "https://dev.k8s.local:6443", Region: "us-east-1", Primary: true},
				},
			},
			{
				Name: "prod",
				Clusters: []config.EnvironmentCluster{
					{Name: "prod-cluster", URL: "https://prod.k8s.local:6443", Region: "us-west-2", Primary: true},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	// Verify cluster secrets are created
	expectedSecrets := []string{
		"mc-cluster-per-env/argocd/clusters/dev-cluster.yaml",
		"mc-cluster-per-env/argocd/clusters/prod-cluster.yaml",
	}

	for _, file := range expectedSecrets {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Cluster secret not created: %s", file)
		}
	}

	// Verify cluster-per-env applicationsets are created
	expectedAppSets := []string{
		"mc-cluster-per-env/argocd/applicationsets/infra-nonprod-cluster.yaml",
		"mc-cluster-per-env/argocd/applicationsets/infra-prod-cluster.yaml",
		"mc-cluster-per-env/argocd/applicationsets/apps-nonprod-cluster.yaml",
		"mc-cluster-per-env/argocd/applicationsets/apps-prod-cluster.yaml",
	}

	for _, file := range expectedAppSets {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("ApplicationSet not created: %s", file)
		}
	}

	// Verify cluster secret content
	secretContent, err := os.ReadFile(filepath.Join(tmpDir, "mc-cluster-per-env/argocd/clusters/dev-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read cluster secret: %v", err)
	}

	secretChecks := []string{
		"kind: Secret",
		"name: dev-cluster",
		"argocd.argoproj.io/secret-type: cluster",
		"env: nonprod",
		"region: us-east-1",
		"primary: \"true\"",
		"server: https://dev.k8s.local:6443",
	}

	for _, check := range secretChecks {
		if !strings.Contains(string(secretContent), check) {
			t.Errorf("Cluster secret missing: %s", check)
		}
	}

	// Verify applicationset content
	appSetContent, err := os.ReadFile(filepath.Join(tmpDir, "mc-cluster-per-env/argocd/applicationsets/infra-nonprod-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read applicationset: %v", err)
	}

	appSetChecks := []string{
		"kind: ApplicationSet",
		"name: mc-cluster-per-env-infra-nonprod",
		"clusters:",
		"matchLabels:",
		"env: nonprod",
		"project: infrastructure",
		"repoURL: https://github.com/test/repo.git",
	}

	for _, check := range appSetChecks {
		if !strings.Contains(string(appSetContent), check) {
			t.Errorf("ApplicationSet missing: %s", check)
		}
	}
}

func TestGenerateMultiClusterArgoCD_MultiCluster(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "mc-multi-cluster"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyMultiCluster,
		Output:     config.Output{URL: "https://github.com/test/repo.git", Branch: "main"},
		Environments: []config.Environment{
			{
				Name: "dev",
				Clusters: []config.EnvironmentCluster{
					{Name: "dev-us-east", URL: "https://dev-east.k8s.local:6443", Region: "us-east-1", Primary: true},
					{Name: "dev-us-west", URL: "https://dev-west.k8s.local:6443", Region: "us-west-2"},
				},
			},
			{
				Name: "prod",
				Clusters: []config.EnvironmentCluster{
					{Name: "prod-us-east", URL: "https://prod-east.k8s.local:6443", Region: "us-east-1", Primary: true},
					{Name: "prod-eu-west", URL: "https://prod-eu.k8s.local:6443", Region: "eu-west-1"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	// Verify all cluster secrets are created
	expectedSecrets := []string{
		"mc-multi-cluster/argocd/clusters/dev-us-east.yaml",
		"mc-multi-cluster/argocd/clusters/dev-us-west.yaml",
		"mc-multi-cluster/argocd/clusters/prod-us-east.yaml",
		"mc-multi-cluster/argocd/clusters/prod-eu-west.yaml",
	}

	for _, file := range expectedSecrets {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Cluster secret not created: %s", file)
		}
	}

	// Verify multi-cluster applicationsets are created
	expectedAppSets := []string{
		"mc-multi-cluster/argocd/applicationsets/infra-multi-cluster.yaml",
		"mc-multi-cluster/argocd/applicationsets/apps-multi-cluster.yaml",
	}

	for _, file := range expectedAppSets {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("ApplicationSet not created: %s", file)
		}
	}

	// Verify multi-cluster applicationset content
	appSetContent, err := os.ReadFile(filepath.Join(tmpDir, "mc-multi-cluster/argocd/applicationsets/infra-multi-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read multi-cluster applicationset: %v", err)
	}

	appSetChecks := []string{
		"kind: ApplicationSet",
		"name: mc-multi-cluster-infra-multi-cluster",
		"matrix:",
		"generators:",
		"list:",
		"elements:",
		"- env: dev",
		"- env: prod",
		"clusters:",
		"project: infrastructure",
		"repoURL: https://github.com/test/repo.git",
		"targetRevision: main",
	}

	for _, check := range appSetChecks {
		if !strings.Contains(string(appSetContent), check) {
			t.Errorf("Multi-cluster ApplicationSet missing: %s", check)
		}
	}
}

func TestGenerateClusterSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "cluster-secrets-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Environments: []config.Environment{
			{
				Name: "dev",
				Clusters: []config.EnvironmentCluster{
					{Name: "dev-primary", URL: "https://dev.k8s.local:6443", Region: "us-east-1", Primary: true},
					{Name: "dev-secondary", URL: "https://dev2.k8s.local:6443", Region: "us-west-2"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateClusterSecrets("argocd")
	if err != nil {
		t.Fatalf("generateClusterSecrets() error = %v", err)
	}

	// Verify primary cluster secret
	primaryContent, err := os.ReadFile(filepath.Join(tmpDir, "cluster-secrets-test/argocd/clusters/dev-primary.yaml"))
	if err != nil {
		t.Fatalf("Failed to read primary cluster secret: %v", err)
	}

	if !strings.Contains(string(primaryContent), "primary: \"true\"") {
		t.Error("Primary cluster secret missing primary label")
	}

	if !strings.Contains(string(primaryContent), "region: us-east-1") {
		t.Error("Primary cluster secret missing region label")
	}

	// Verify secondary cluster secret
	secondaryContent, err := os.ReadFile(filepath.Join(tmpDir, "cluster-secrets-test/argocd/clusters/dev-secondary.yaml"))
	if err != nil {
		t.Fatalf("Failed to read secondary cluster secret: %v", err)
	}

	// Secondary should NOT have primary label
	if strings.Contains(string(secondaryContent), "primary:") {
		t.Error("Secondary cluster should not have primary label")
	}

	if !strings.Contains(string(secondaryContent), "region: us-west-2") {
		t.Error("Secondary cluster secret missing region label")
	}
}

func TestGenerateClusterPerEnvApplicationSets(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "cluster-per-env-appsets"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Output:     config.Output{Branch: "develop"},
		Environments: []config.Environment{
			{Name: "dev", Namespace: "my-app-dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateClusterPerEnvApplicationSets("argocd", "https://github.com/test/repo.git", "develop")
	if err != nil {
		t.Fatalf("generateClusterPerEnvApplicationSets() error = %v", err)
	}

	// Verify applicationsets for each environment and scope
	expectedFiles := []string{
		"cluster-per-env-appsets/argocd/applicationsets/infra-dev-cluster.yaml",
		"cluster-per-env-appsets/argocd/applicationsets/apps-dev-cluster.yaml",
		"cluster-per-env-appsets/argocd/applicationsets/infra-staging-cluster.yaml",
		"cluster-per-env-appsets/argocd/applicationsets/apps-staging-cluster.yaml",
		"cluster-per-env-appsets/argocd/applicationsets/infra-prod-cluster.yaml",
		"cluster-per-env-appsets/argocd/applicationsets/apps-prod-cluster.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("ApplicationSet not created: %s", file)
		}
	}

	// Verify custom namespace is used
	devContent, err := os.ReadFile(filepath.Join(tmpDir, "cluster-per-env-appsets/argocd/applicationsets/infra-dev-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read dev applicationset: %v", err)
	}

	if !strings.Contains(string(devContent), "namespace: my-app-dev") {
		t.Error("Dev applicationset should use custom namespace")
	}

	if !strings.Contains(string(devContent), "targetRevision: develop") {
		t.Error("ApplicationSet should use develop branch")
	}
}

func TestGenerateMultiClusterApplicationSets(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "multi-appsets"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyMultiCluster,
		Environments: []config.Environment{
			{Name: "dev", Namespace: "app-dev"},
			{Name: "staging", Namespace: "app-staging"},
			{Name: "prod", Namespace: "app-prod"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateMultiClusterApplicationSets("argocd", "https://github.com/test/repo.git", "main")
	if err != nil {
		t.Fatalf("generateMultiClusterApplicationSets() error = %v", err)
	}

	// Only two applicationsets should be created (one for infra, one for apps)
	expectedFiles := []string{
		"multi-appsets/argocd/applicationsets/infra-multi-cluster.yaml",
		"multi-appsets/argocd/applicationsets/apps-multi-cluster.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Multi-cluster ApplicationSet not created: %s", file)
		}
	}

	// Verify environment list in applicationset
	infraContent, err := os.ReadFile(filepath.Join(tmpDir, "multi-appsets/argocd/applicationsets/infra-multi-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read infra applicationset: %v", err)
	}

	envChecks := []string{
		"- env: dev",
		"namespace: app-dev",
		"- env: staging",
		"namespace: app-staging",
		"- env: prod",
		"namespace: app-prod",
	}

	for _, check := range envChecks {
		if !strings.Contains(string(infraContent), check) {
			t.Errorf("Multi-cluster ApplicationSet missing: %s", check)
		}
	}
}

func TestGenerateMultiClusterArgoCD_MissingGitURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-git-url"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Environments: []config.Environment{
			{
				Name: "dev",
				Clusters: []config.EnvironmentCluster{
					{Name: "dev-cluster", URL: "https://dev.k8s.local:6443"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err == nil {
		t.Fatal("Expected error when git.url is missing for multi-cluster ArgoCD")
	}

	if !strings.Contains(err.Error(), "git.url is required") {
		t.Errorf("Expected 'git.url is required' error, got: %v", err)
	}
}

func TestGenerateMultiClusterArgoCD_InfrastructureOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "mc-infra-only"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{
				Name: "prod",
				Clusters: []config.EnvironmentCluster{
					{Name: "prod-cluster", URL: "https://prod.k8s.local:6443"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	// Infra applicationset should exist
	infraPath := filepath.Join(tmpDir, "mc-infra-only/argocd/applicationsets/infra-prod-cluster.yaml")
	if _, err := os.Stat(infraPath); os.IsNotExist(err) {
		t.Error("Infrastructure applicationset should be created")
	}

	// Apps applicationset should NOT exist
	appsPath := filepath.Join(tmpDir, "mc-infra-only/argocd/applicationsets/apps-prod-cluster.yaml")
	if _, err := os.Stat(appsPath); !os.IsNotExist(err) {
		t.Error("Application applicationset should not be created for infrastructure-only scope")
	}
}

func TestGenerateMultiClusterArgoCD_ApplicationOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "mc-apps-only"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Topology:   config.TopologyMultiCluster,
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{
				Name: "dev",
				Clusters: []config.EnvironmentCluster{
					{Name: "dev-cluster", URL: "https://dev.k8s.local:6443"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	// Apps applicationset should exist
	appsPath := filepath.Join(tmpDir, "mc-apps-only/argocd/applicationsets/apps-multi-cluster.yaml")
	if _, err := os.Stat(appsPath); os.IsNotExist(err) {
		t.Error("Application applicationset should be created")
	}

	// Infra applicationset should NOT exist
	infraPath := filepath.Join(tmpDir, "mc-apps-only/argocd/applicationsets/infra-multi-cluster.yaml")
	if _, err := os.Stat(infraPath); !os.IsNotExist(err) {
		t.Error("Infrastructure applicationset should not be created for application-only scope")
	}
}

func TestGenerateMultiClusterArgoCD_OpenShiftNamespace(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "mc-openshift"},
		Platform:   "openshift",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Output:     config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{
				Name: "dev",
				Clusters: []config.EnvironmentCluster{
					{Name: "ocp-dev", URL: "https://api.ocp.local:6443"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	// Verify cluster secret uses openshift-gitops namespace
	secretContent, err := os.ReadFile(filepath.Join(tmpDir, "mc-openshift/argocd/clusters/ocp-dev.yaml"))
	if err != nil {
		t.Fatalf("Failed to read cluster secret: %v", err)
	}

	if !strings.Contains(string(secretContent), "namespace: openshift-gitops") {
		t.Error("OpenShift cluster secret should use openshift-gitops namespace")
	}

	// Verify applicationset uses openshift-gitops namespace
	appSetContent, err := os.ReadFile(filepath.Join(tmpDir, "mc-openshift/argocd/applicationsets/infra-dev-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read applicationset: %v", err)
	}

	if !strings.Contains(string(appSetContent), "namespace: openshift-gitops") {
		t.Error("OpenShift applicationset should use openshift-gitops namespace")
	}
}

func TestGenerateMultiClusterArgoCD_CustomNamespace(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "mc-custom-ns"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   config.TopologyClusterPerEnv,
		Bootstrap: config.BootstrapConfig{
			Namespace: "my-argocd",
		},
		Output: config.Output{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{
				Name: "dev",
				Clusters: []config.EnvironmentCluster{
					{Name: "dev-cluster", URL: "https://dev.k8s.local:6443"},
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() error = %v", err)
	}

	// Verify cluster secret uses custom namespace
	secretContent, err := os.ReadFile(filepath.Join(tmpDir, "mc-custom-ns/argocd/clusters/dev-cluster.yaml"))
	if err != nil {
		t.Fatalf("Failed to read cluster secret: %v", err)
	}

	if !strings.Contains(string(secretContent), "namespace: my-argocd") {
		t.Error("Cluster secret should use custom namespace")
	}
}

func TestGenerateArgoCDNamespace(t *testing.T) {
	tests := []struct {
		name      string
		platform  string
		bootstrap config.BootstrapConfig
		expected  string
	}{
		{
			name:      "kubernetes default",
			platform:  "kubernetes",
			bootstrap: config.BootstrapConfig{},
			expected:  "argocd",
		},
		{
			name:      "openshift default",
			platform:  "openshift",
			bootstrap: config.BootstrapConfig{},
			expected:  "openshift-gitops",
		},
		{
			name:      "eks default",
			platform:  "eks",
			bootstrap: config.BootstrapConfig{},
			expected:  "argocd",
		},
		{
			name:      "aks default",
			platform:  "aks",
			bootstrap: config.BootstrapConfig{},
			expected:  "argocd",
		},
		{
			name:      "custom namespace overrides kubernetes",
			platform:  "kubernetes",
			bootstrap: config.BootstrapConfig{Namespace: "custom-argocd"},
			expected:  "custom-argocd",
		},
		{
			name:      "custom namespace overrides openshift",
			platform:  "openshift",
			bootstrap: config.BootstrapConfig{Namespace: "my-gitops"},
			expected:  "my-gitops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Project:   config.Project{Name: "test"},
				Platform:  tt.platform,
				Bootstrap: tt.bootstrap,
			}
			writer := output.New("/tmp", true, false)
			gen := New(cfg, writer, false)

			got := gen.getArgoCDNamespace()
			if got != tt.expected {
				t.Errorf("getArgoCDNamespace() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNew_VerboseMode(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "verbose-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, true)

	if !gen.Verbose {
		t.Error("Verbose should be true")
	}
}

func TestNew_VersionMapper_OpenShiftVersion(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "ocp-version-test"
	cfg.Platform = "openshift"
	cfg.Version.OpenShift = "4.14"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)

	if gen.VersionMapper == nil {
		t.Error("VersionMapper should be initialized for OpenShift with version")
	}
}

func TestNew_VersionMapper_KubernetesVersion(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "k8s-version-test"
	cfg.Platform = "kubernetes"
	cfg.Version.Kubernetes = "1.29"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)

	if gen.VersionMapper == nil {
		t.Error("VersionMapper should be initialized for Kubernetes with version")
	}
}

func TestGetAPIVersion_NilVersionMapper(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "nil-mapper-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)
	gen.VersionMapper = nil

	apiVersion := gen.GetAPIVersion("Namespace")
	if apiVersion != "v1" {
		t.Errorf("GetAPIVersion() = %s, want v1", apiVersion)
	}
}

func TestGetAPIVersion_UnknownKind(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "unknown-kind-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)
	gen.VersionMapper = nil

	apiVersion := gen.GetAPIVersion("UnknownKindXYZ")
	if apiVersion != "" {
		t.Errorf("GetAPIVersion() for unknown kind = %s, want empty", apiVersion)
	}
}

func TestCheckDeprecation_NilMapper(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "nil-mapper-dep-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)
	gen.VersionMapper = nil

	// Should not panic
	gen.CheckDeprecation("Deployment", "apps/v1", "my-deployment", "test.yaml")

	if len(gen.Deprecations) != 0 {
		t.Errorf("Deprecations should be empty when VersionMapper is nil, got %d", len(gen.Deprecations))
	}
}

func TestGetDeprecations_Empty(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "empty-dep-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)

	deps := gen.GetDeprecations()
	if len(deps) != 0 {
		t.Errorf("GetDeprecations() should return empty slice, got %d", len(deps))
	}
}

func TestHasCriticalDeprecations_NoCritical(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "no-critical-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)

	if gen.HasCriticalDeprecations() {
		t.Error("HasCriticalDeprecations() should be false with no deprecations")
	}
}

func TestGenerateStructure_SingleEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "single-env"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Environments: []config.Environment{
			{Name: "production"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateStructure()
	if err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	expectedDirs := []string{
		"single-env/infrastructure/overlays/production",
		"single-env/applications/overlays/production",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}
}

func TestGenerateStructure_InfrastructureOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "infra-only"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateStructure()
	if err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	// Should have infrastructure dirs
	infraDir := filepath.Join(tmpDir, "infra-only/infrastructure/overlays/dev")
	if _, err := os.Stat(infraDir); os.IsNotExist(err) {
		t.Error("Infrastructure directory not created")
	}

	// Should NOT have applications dirs
	appsDir := filepath.Join(tmpDir, "infra-only/applications/overlays/dev")
	if _, err := os.Stat(appsDir); !os.IsNotExist(err) {
		t.Error("Applications directory should not be created for infrastructure-only scope")
	}
}

func TestGenerateStructure_ApplicationOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "apps-only"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateStructure()
	if err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	// Should have applications dirs
	appsDir := filepath.Join(tmpDir, "apps-only/applications/overlays/dev")
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		t.Error("Applications directory not created")
	}

	// Should NOT have infrastructure dirs
	infraDir := filepath.Join(tmpDir, "apps-only/infrastructure/overlays/dev")
	if _, err := os.Stat(infraDir); !os.IsNotExist(err) {
		t.Error("Infrastructure directory should not be created for application-only scope")
	}
}

func TestGenerateStructure_WithAllInfraOptions(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "all-infra"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Output:     config.Output{Type: "local"},
		Infra: config.Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateStructure()
	if err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	expectedDirs := []string{
		"all-infra/infrastructure/base/namespaces",
		"all-infra/infrastructure/base/rbac",
		"all-infra/infrastructure/base/network-policies",
		"all-infra/infrastructure/base/resource-quotas",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}
}

func TestGenerateArgoCDNamespace_EmptyPlatform(t *testing.T) {
	cfg := &config.Config{
		Project:  config.Project{Name: "test"},
		Platform: "",
	}
	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)

	ns := gen.getArgoCDNamespace()
	if ns != "argocd" {
		t.Errorf("getArgoCDNamespace() for empty platform = %s, want argocd", ns)
	}
}

func TestGenerateArgoCDNamespace_GKEPlatform(t *testing.T) {
	cfg := &config.Config{
		Project:  config.Project{Name: "test"},
		Platform: "gke",
	}
	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, false)

	ns := gen.getArgoCDNamespace()
	if ns != "argocd" {
		t.Errorf("getArgoCDNamespace() for GKE = %s, want argocd", ns)
	}
}

func TestGeneratorFields(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "field-test"

	writer := output.New("/tmp", true, false)
	gen := New(cfg, writer, true)

	if gen.Config == nil {
		t.Error("Config should not be nil")
	}
	if gen.Writer == nil {
		t.Error("Writer should not be nil")
	}
	if !gen.Verbose {
		t.Error("Verbose should be true")
	}
	if gen.Deprecations == nil {
		t.Error("Deprecations should be initialized")
	}
	if len(gen.Deprecations) != 0 {
		t.Errorf("Deprecations should be empty, got %d", len(gen.Deprecations))
	}
}

// Additional ArgoCD edge case tests

func TestGenerateGitOps_ArgoCDOnly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "argocd-only"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateGitOps()
	if err != nil {
		t.Fatalf("generateGitOps() error = %v", err)
	}

	// Verify ArgoCD files were created
	projectPath := filepath.Join(tmpDir, "argocd-only/argocd/projects/infrastructure.yaml")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("ArgoCD infrastructure project not created")
	}
}

func TestGenerateGitOps_FluxDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "flux-disabled"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "flux",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	// Flux is currently disabled, should not error but also not create files
	err := gen.generateGitOps()
	if err != nil {
		t.Fatalf("generateGitOps() should not error even for flux: %v", err)
	}
}

func TestGenerateGitOps_BothTools(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "both-tools"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "both",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateGitOps()
	if err != nil {
		t.Fatalf("generateGitOps() error = %v", err)
	}

	// ArgoCD files should be created even with "both" setting
	projectPath := filepath.Join(tmpDir, "both-tools/both/projects/infrastructure.yaml")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("ArgoCD infrastructure project not created for 'both' tool setting")
	}
}

func TestGetArgoCDNamespace_AllPlatforms(t *testing.T) {
	tests := []struct {
		platform  string
		namespace string
		expected  string
	}{
		{"kubernetes", "", "argocd"},
		{"openshift", "", "openshift-gitops"},
		{"eks", "", "argocd"},
		{"aks", "", "argocd"},
		{"gke", "", "argocd"},
		{"kubernetes", "custom-ns", "custom-ns"},
		{"openshift", "custom-ns", "custom-ns"},
	}

	for _, tt := range tests {
		t.Run(tt.platform+"_"+tt.namespace, func(t *testing.T) {
			cfg := &config.Config{
				Platform: tt.platform,
				Bootstrap: config.Bootstrap{
					Namespace: tt.namespace,
				},
			}
			writer := output.New("/tmp", true, false)
			gen := New(cfg, writer, false)

			ns := gen.getArgoCDNamespace()
			if ns != tt.expected {
				t.Errorf("getArgoCDNamespace() = %s, want %s", ns, tt.expected)
			}
		})
	}
}

func TestGenerateArgoCD_MissingGitURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-git-url"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: ""}, // Empty URL
		Output:     config.Output{URL: ""},    // Also empty
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err == nil {
		t.Error("generateArgoCD() should error when git.url is missing")
	}

	if !strings.Contains(err.Error(), "git.url is required") {
		t.Errorf("Expected 'git.url is required' error, got: %v", err)
	}
}

func TestGenerateArgoCD_UsesOutputURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "output-url"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: ""},
		Output:     config.Output{URL: "https://github.com/fallback/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() should use output.url as fallback: %v", err)
	}
}

func TestGenerateArgoCDProjects_InfrastructureScope(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "infra-scope"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCDProjects("argocd")
	if err != nil {
		t.Fatalf("generateArgoCDProjects() error = %v", err)
	}

	infraPath := filepath.Join(tmpDir, "infra-scope/argocd/projects/infrastructure.yaml")
	if _, err := os.Stat(infraPath); os.IsNotExist(err) {
		t.Error("Infrastructure project should be created")
	}

	appsPath := filepath.Join(tmpDir, "infra-scope/argocd/projects/applications.yaml")
	if _, err := os.Stat(appsPath); !os.IsNotExist(err) {
		t.Error("Applications project should NOT be created for infrastructure scope")
	}
}

func TestGenerateArgoCDProjects_ApplicationScope(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "app-scope"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCDProjects("argocd")
	if err != nil {
		t.Fatalf("generateArgoCDProjects() error = %v", err)
	}

	infraPath := filepath.Join(tmpDir, "app-scope/argocd/projects/infrastructure.yaml")
	if _, err := os.Stat(infraPath); !os.IsNotExist(err) {
		t.Error("Infrastructure project should NOT be created for application scope")
	}

	appsPath := filepath.Join(tmpDir, "app-scope/argocd/projects/applications.yaml")
	if _, err := os.Stat(appsPath); os.IsNotExist(err) {
		t.Error("Applications project should be created")
	}
}

func TestGenerateSingleClusterArgoCD_MultipleEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "multi-env"},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateSingleClusterArgoCD("argocd")
	if err != nil {
		t.Fatalf("generateSingleClusterArgoCD() error = %v", err)
	}

	// Verify files for each environment
	envs := []string{"dev", "staging", "prod"}
	for _, env := range envs {
		infraPath := filepath.Join(tmpDir, "multi-env/argocd/applicationsets/infra-"+env+".yaml")
		if _, err := os.Stat(infraPath); os.IsNotExist(err) {
			t.Errorf("Infrastructure application not created for %s", env)
		}

		appsPath := filepath.Join(tmpDir, "multi-env/argocd/applicationsets/apps-"+env+".yaml")
		if _, err := os.Stat(appsPath); os.IsNotExist(err) {
			t.Errorf("Applications application not created for %s", env)
		}
	}
}

func TestGenerateSingleClusterArgoCD_WithClusterURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "with-cluster"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev", Cluster: "https://dev-cluster.example.com"},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateSingleClusterArgoCD("argocd")
	if err != nil {
		t.Fatalf("generateSingleClusterArgoCD() error = %v", err)
	}

	// Verify content includes the cluster URL
	content, err := os.ReadFile(filepath.Join(tmpDir, "with-cluster/argocd/applicationsets/infra-dev.yaml"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "dev-cluster.example.com") {
		t.Error("Application should contain the cluster URL")
	}
}

func TestGenerateArgoCD_EmptyEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:      config.Project{Name: "no-envs"},
		Platform:     "kubernetes",
		Scope:        "both",
		GitOpsTool:   "argocd",
		Git:          config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{}, // Empty
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	if err := gen.generateStructure(); err != nil {
		t.Fatalf("generateStructure() error = %v", err)
	}

	err := gen.generateArgoCD()
	if err != nil {
		t.Fatalf("generateArgoCD() should handle empty environments: %v", err)
	}

	// Projects should still be created
	infraPath := filepath.Join(tmpDir, "no-envs/argocd/projects/infrastructure.yaml")
	if _, err := os.Stat(infraPath); os.IsNotExist(err) {
		t.Error("Infrastructure project should be created even with no environments")
	}
}

func TestEnvInfo_Struct(t *testing.T) {
	env := envInfo{
		Name:      "production",
		Namespace: "prod-ns",
	}

	if env.Name != "production" {
		t.Errorf("Name = %s, want production", env.Name)
	}
	if env.Namespace != "prod-ns" {
		t.Errorf("Namespace = %s, want prod-ns", env.Namespace)
	}
}

func TestEnvInfo_EmptyValues(t *testing.T) {
	env := envInfo{}

	if env.Name != "" {
		t.Errorf("Name should be empty, got %s", env.Name)
	}
	if env.Namespace != "" {
		t.Errorf("Namespace should be empty, got %s", env.Namespace)
	}
}

// ============================================================================
// Edge Case Tests - Boundary conditions and unusual inputs
// ============================================================================

func TestEdgeCase_VeryLongProjectName(t *testing.T) {
	tmpDir := t.TempDir()

	longName := "this-is-an-extremely-long-project-name-that-exceeds-typical-limits-and-should-be-handled-gracefully-by-the-generator"

	cfg := &config.Config{
		Project:    config.Project{Name: longName},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	// Should either succeed or fail gracefully - not panic
	_ = gen.Generate()
}

func TestEdgeCase_SingleCharacterNames(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "a"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "d"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Single char names should work: %v", err)
	}

	// Verify namespace file was created
	nsFile := filepath.Join(tmpDir, "a/infrastructure/base/namespaces/d.yaml")
	if _, err := os.Stat(nsFile); os.IsNotExist(err) {
		t.Error("Namespace file should be created for single char env name")
	}
}

func TestEdgeCase_NumericEnvironmentNames(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "numeric-envs"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "env1"},
			{Name: "env2"},
			{Name: "env3"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Numeric env names should work: %v", err)
	}

	for _, env := range cfg.Environments {
		nsFile := filepath.Join(tmpDir, "numeric-envs/infrastructure/base/namespaces", env.Name+".yaml")
		if _, err := os.Stat(nsFile); os.IsNotExist(err) {
			t.Errorf("Namespace file should be created for %s", env.Name)
		}
	}
}

func TestEdgeCase_HyphenatedNames(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "my-awesome-project"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev-us-east-1"},
			{Name: "prod-eu-west-2"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Hyphenated names should work: %v", err)
	}
}

func TestEdgeCase_UnderscoreNames(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "my_project_with_underscores"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev_environment"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Underscore names should work: %v", err)
	}
}

func TestEdgeCase_ManyEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 10 environments
	envs := make([]config.Environment, 10)
	for i := 0; i < 10; i++ {
		envs[i] = config.Environment{Name: fmt.Sprintf("env%d", i+1)}
	}

	cfg := &config.Config{
		Project:      config.Project{Name: "many-envs"},
		Platform:     "kubernetes",
		Scope:        "infrastructure",
		GitOpsTool:   "argocd",
		Git:          config.GitConfig{URL: testGitURL},
		Environments: envs,
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Many environments should work: %v", err)
	}

	// Verify all namespace files were created
	for i := 0; i < 10; i++ {
		envName := fmt.Sprintf("env%d", i+1)
		nsFile := filepath.Join(tmpDir, "many-envs/infrastructure/base/namespaces", envName+".yaml")
		if _, err := os.Stat(nsFile); os.IsNotExist(err) {
			t.Errorf("Namespace file should be created for %s", envName)
		}
	}
}

func TestEdgeCase_ManyApplications(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 10 applications
	apps := make([]config.Application, 10)
	for i := 0; i < 10; i++ {
		apps[i] = config.Application{
			Name:     fmt.Sprintf("app%d", i+1),
			Image:    fmt.Sprintf("nginx:1.%d", i),
			Port:     8080 + i,
			Replicas: i + 1,
		}
	}

	cfg := &config.Config{
		Project:    config.Project{Name: "many-apps"},
		Platform:   "kubernetes",
		Scope:      "application",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Apps: apps,
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Many applications should work: %v", err)
	}

	// Verify all app directories were created
	for i := 0; i < 10; i++ {
		appName := fmt.Sprintf("app%d", i+1)
		appDir := filepath.Join(tmpDir, "many-apps/applications/base", appName)
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			t.Errorf("App directory should be created for %s", appName)
		}
	}
}

func TestEdgeCase_AllInfraEnabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "all-infra"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
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
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("All infra enabled should work: %v", err)
	}

	// Verify all infra directories have content
	subdirs := []string{"namespaces", "rbac", "network-policies", "resource-quotas"}
	for _, subdir := range subdirs {
		path := filepath.Join(tmpDir, "all-infra/infrastructure/base", subdir, "dev.yaml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s/dev.yaml should exist", subdir)
		}
	}
}

func TestEdgeCase_NoInfraEnabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "no-infra"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
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
	gen := New(cfg, writer, false)

	// Should handle gracefully
	err := gen.Generate()
	// May fail or succeed depending on implementation
	_ = err
}

func TestEdgeCase_SpecialPortNumbers(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port-80", 80},
		{"port-443", 443},
		{"port-8080", 8080},
		{"port-high", 65535},
		{"port-1024", 1024},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: tc.name},
				Platform:   "kubernetes",
				Scope:      "application",
				GitOpsTool: "argocd",
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Apps: []config.Application{
					{Name: "app", Image: "nginx", Port: tc.port, Replicas: 1},
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Port %d should work: %v", tc.port, err)
			}

			// Verify port is in deployment
			deployFile := filepath.Join(tmpDir, tc.name+"/applications/base/app/deployment.yaml")
			content, err := os.ReadFile(deployFile)
			if err != nil {
				t.Fatalf("Failed to read deployment: %v", err)
			}

			portStr := fmt.Sprintf("containerPort: %d", tc.port)
			if !strings.Contains(string(content), portStr) {
				t.Errorf("Deployment should have %s", portStr)
			}
		})
	}
}

func TestEdgeCase_ReplicaCounts(t *testing.T) {
	tests := []struct {
		name     string
		replicas int
	}{
		{"one-replica", 1},
		{"three-replicas", 3},
		{"ten-replicas", 10},
		{"zero-replicas", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: tc.name},
				Platform:   "kubernetes",
				Scope:      "application",
				GitOpsTool: "argocd",
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Apps: []config.Application{
					{Name: "app", Image: "nginx", Port: 80, Replicas: tc.replicas},
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Replicas %d should work: %v", tc.replicas, err)
			}

			// Verify replicas is in deployment
			deployFile := filepath.Join(tmpDir, tc.name+"/applications/base/app/deployment.yaml")
			content, err := os.ReadFile(deployFile)
			if err != nil {
				t.Fatalf("Failed to read deployment: %v", err)
			}

			replicaStr := fmt.Sprintf("replicas: %d", tc.replicas)
			if !strings.Contains(string(content), replicaStr) {
				t.Errorf("Deployment should have %s", replicaStr)
			}
		})
	}
}

func TestEdgeCase_VariousImageFormats(t *testing.T) {
	tests := []struct {
		name  string
		image string
	}{
		{"simple-image", "nginx"},
		{"with-tag", "nginx:latest"},
		{"with-version", "nginx:1.21.0"},
		{"with-registry", "gcr.io/project/app:v1"},
		{"with-port", "registry.example.com:5000/app:v1"},
		{"sha-digest", "nginx@sha256:abc123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: tc.name},
				Platform:   "kubernetes",
				Scope:      "application",
				GitOpsTool: "argocd",
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Apps: []config.Application{
					{Name: "app", Image: tc.image, Port: 80, Replicas: 1},
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Image %s should work: %v", tc.image, err)
			}

			// Verify image is in deployment
			deployFile := filepath.Join(tmpDir, tc.name+"/applications/base/app/deployment.yaml")
			content, err := os.ReadFile(deployFile)
			if err != nil {
				t.Fatalf("Failed to read deployment: %v", err)
			}

			if !strings.Contains(string(content), "image: "+tc.image) {
				t.Errorf("Deployment should have image: %s", tc.image)
			}
		})
	}
}

func TestEdgeCase_DryRunPreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an existing file
	existingDir := filepath.Join(tmpDir, "existing-project")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("Failed to create existing dir: %v", err)
	}

	existingFile := filepath.Join(existingDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	cfg := &config.Config{
		Project:    config.Project{Name: "existing-project"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, true, false) // dryRun = true
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Dry run should succeed: %v", err)
	}

	// Existing file should be preserved
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Existing file should still exist: %v", err)
	}

	if string(content) != "existing content" {
		t.Error("Existing content should be preserved")
	}
}

func TestEdgeCase_VerboseModeOutput(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "verbose-test"},
		Platform:   "kubernetes",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Infra: config.Infrastructure{
			Namespaces: true,
		},
	}

	writer := output.New(tmpDir, false, true) // verbose = true
	gen := New(cfg, writer, true)             // verbose = true

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Verbose mode should work: %v", err)
	}
}

func TestEdgeCase_MultiClusterURLFormats(t *testing.T) {
	tests := []struct {
		name       string
		clusterURL string
	}{
		{"https-standard", "https://cluster.example.com:6443"},
		{"https-no-port", "https://cluster.example.com"},
		{"localhost", "https://localhost:6443"},
		{"ip-address", "https://192.168.1.100:6443"},
		{"aws-eks", "https://ABC123.gr7.us-west-2.eks.amazonaws.com"},
		{"azure-aks", "https://myaks-dns-abc123.hcp.westus2.azmk8s.io:443"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: tc.name},
				Platform:   "kubernetes",
				Scope:      "infrastructure",
				GitOpsTool: "argocd",
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev", ClusterURL: tc.clusterURL},
				},
				Infra: config.Infrastructure{
					Namespaces: true,
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Cluster URL %s should work: %v", tc.clusterURL, err)
			}
		})
	}
}

func TestEdgeCase_ProjectDescriptionWithSpecialChars(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project: config.Project{
			Name:        "special-desc",
			Description: "Project with special chars: <>&\"'",
		},
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: testGitURL},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Docs: config.Documentation{
			Readme: true,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Special chars in description should work: %v", err)
	}
}

func TestEdgeCase_AllPlatformsWithBootstrap(t *testing.T) {
	platforms := []struct {
		platform  string
		namespace string
	}{
		{"kubernetes", "argocd"},
		{"openshift", "openshift-gitops"},
		{"eks", "argocd"},
		{"aks", "argocd"},
	}

	for _, tc := range platforms {
		t.Run(tc.platform, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				Project:    config.Project{Name: "bootstrap-" + tc.platform},
				Platform:   tc.platform,
				Scope:      "both",
				GitOpsTool: "argocd",
				Git:        config.GitConfig{URL: testGitURL},
				Environments: []config.Environment{
					{Name: "dev"},
				},
				Bootstrap: config.BootstrapConfig{
					Enabled: true,
					Mode:    "helm",
				},
			}

			writer := output.New(tmpDir, false, false)
			gen := New(cfg, writer, false)

			err := gen.Generate()
			if err != nil {
				t.Fatalf("Platform %s should work: %v", tc.platform, err)
			}
		})
	}
}
