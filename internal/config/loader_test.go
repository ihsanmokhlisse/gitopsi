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

