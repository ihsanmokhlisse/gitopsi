package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
)

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

