package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	content := `
project:
  name: test-project
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
  - name: prod
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %s", cfg.Project.Name)
	}

	if cfg.Platform != "kubernetes" {
		t.Errorf("expected platform 'kubernetes', got %s", cfg.Platform)
	}

	if len(cfg.Environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(cfg.Environments))
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	content := `{invalid yaml content`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "output.yaml")

	cfg := &Config{
		Project:    Project{Name: "saved-project"},
		Platform:   "openshift",
		Scope:      "infrastructure",
		GitOpsTool: "flux",
		Environments: []Environment{
			{Name: "staging"},
		},
	}

	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to reload saved config: %v", err)
	}

	if loaded.Project.Name != cfg.Project.Name {
		t.Errorf("saved config mismatch: expected %s, got %s",
			cfg.Project.Name, loaded.Project.Name)
	}
}

func TestSaveToInvalidPath(t *testing.T) {
	cfg := &Config{
		Project: Project{Name: "test"},
	}

	err := Save(cfg, "/nonexistent/directory/config.yaml")
	if err == nil {
		t.Error("Save() should fail for invalid path")
	}
}

func TestSaveCompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "complete.yaml")

	cfg := &Config{
		Project: Project{
			Name:        "complete-project",
			Description: "A complete test project",
		},
		Platform:   "eks",
		Scope:      "both",
		GitOpsTool: "both",
		Output: Output{
			Type:   "git",
			URL:    "https://github.com/test/repo.git",
			Branch: "main",
		},
		Environments: []Environment{
			{Name: "dev", Cluster: "https://dev.k8s"},
			{Name: "prod", Cluster: "https://prod.k8s"},
		},
		Infra: Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
		Apps: []Application{
			{Name: "api", Image: "api:v1", Port: 8080, Replicas: 3},
			{Name: "web", Image: "web:v1", Port: 80, Replicas: 2},
		},
		Docs: Documentation{
			Readme:       true,
			Architecture: true,
			Onboarding:   true,
		},
	}

	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Platform != cfg.Platform {
		t.Errorf("Platform mismatch: %s vs %s", loaded.Platform, cfg.Platform)
	}
	if loaded.Output.URL != cfg.Output.URL {
		t.Errorf("Output URL mismatch: %s vs %s", loaded.Output.URL, cfg.Output.URL)
	}
	if len(loaded.Apps) != len(cfg.Apps) {
		t.Errorf("Apps count mismatch: %d vs %d", len(loaded.Apps), len(cfg.Apps))
	}
}

func TestLoadPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yaml")

	content := `
project:
  name: partial-project
platform: kubernetes
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Project.Name != "partial-project" {
		t.Error("Project name not loaded")
	}

	if cfg.Scope != "both" {
		t.Error("Default scope not applied")
	}

	if cfg.GitOpsTool != "argocd" {
		t.Error("Default gitops tool not applied")
	}
}
