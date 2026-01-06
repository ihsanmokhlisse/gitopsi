package generator

import (
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
