package config

import (
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.Platform != "kubernetes" {
		t.Errorf("expected platform 'kubernetes', got %s", cfg.Platform)
	}

	if cfg.Scope != "both" {
		t.Errorf("expected scope 'both', got %s", cfg.Scope)
	}

	if cfg.GitOpsTool != "argocd" {
		t.Errorf("expected gitops_tool 'argocd', got %s", cfg.GitOpsTool)
	}

	if len(cfg.Environments) != 3 {
		t.Errorf("expected 3 environments, got %d", len(cfg.Environments))
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Project:    Project{Name: "test"},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     Output{Type: "local"},
				Environments: []Environment{
					{Name: "dev"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			cfg: &Config{
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     Output{Type: "local"},
				Environments: []Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid platform",
			cfg: &Config{
				Project:    Project{Name: "test"},
				Platform:   "invalid",
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     Output{Type: "local"},
				Environments: []Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid scope",
			cfg: &Config{
				Project:    Project{Name: "test"},
				Platform:   "kubernetes",
				Scope:      "invalid",
				GitOpsTool: "argocd",
				Output:     Output{Type: "local"},
				Environments: []Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
		},
		{
			name: "git output without url",
			cfg: &Config{
				Project:    Project{Name: "test"},
				Platform:   "kubernetes",
				Scope:      "both",
				GitOpsTool: "argocd",
				Output:     Output{Type: "git"},
				Environments: []Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
		},
		{
			name: "no environments",
			cfg: &Config{
				Project:      Project{Name: "test"},
				Platform:     "kubernetes",
				Scope:        "both",
				GitOpsTool:   "argocd",
				Output:       Output{Type: "local"},
				Environments: []Environment{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidPlatforms(t *testing.T) {
	platforms := ValidPlatforms()
	if len(platforms) != 4 {
		t.Errorf("expected 4 platforms, got %d", len(platforms))
	}

	expected := map[string]bool{
		"kubernetes": true,
		"openshift":  true,
		"aks":        true,
		"eks":        true,
	}

	for _, p := range platforms {
		if !expected[p] {
			t.Errorf("unexpected platform: %s", p)
		}
	}
}

func TestValidScopes(t *testing.T) {
	scopes := ValidScopes()
	if len(scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(scopes))
	}
}

func TestValidGitOpsTools(t *testing.T) {
	tools := ValidGitOpsTools()
	if len(tools) != 3 {
		t.Errorf("expected 3 gitops tools, got %d", len(tools))
	}
}

func TestConfigValidateAllScenarios(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name: "valid kubernetes both argocd",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "kubernetes"
				c.Scope = "both"
				c.GitOpsTool = "argocd"
			},
			wantErr: false,
		},
		{
			name: "valid openshift infrastructure flux",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "openshift"
				c.Scope = "infrastructure"
				c.GitOpsTool = "flux"
			},
			wantErr: false,
		},
		{
			name: "valid aks application both tools",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "aks"
				c.Scope = "application"
				c.GitOpsTool = "both"
			},
			wantErr: false,
		},
		{
			name: "valid eks with git url",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "eks"
				c.Scope = "both"
				c.GitOpsTool = "argocd"
				c.Output.Type = "git"
				c.Output.URL = "https://github.com/test/repo.git"
			},
			wantErr: false,
		},
		{
			name: "invalid gitops tool",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "kubernetes"
				c.Scope = "both"
				c.GitOpsTool = "invalid"
			},
			wantErr: true,
		},
		{
			name: "invalid output type",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "kubernetes"
				c.Scope = "both"
				c.GitOpsTool = "argocd"
				c.Output.Type = "invalid"
			},
			wantErr: true,
		},
		{
			name: "environment without name",
			modify: func(c *Config) {
				c.Project.Name = "test"
				c.Platform = "kubernetes"
				c.Scope = "both"
				c.GitOpsTool = "argocd"
				c.Environments = []Environment{{Name: ""}}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfigInfrastructure(t *testing.T) {
	cfg := NewDefaultConfig()

	if !cfg.Infra.Namespaces {
		t.Error("Default namespaces should be true")
	}
	if !cfg.Infra.RBAC {
		t.Error("Default RBAC should be true")
	}
	if !cfg.Infra.NetworkPolicies {
		t.Error("Default network policies should be true")
	}
	if !cfg.Infra.ResourceQuotas {
		t.Error("Default resource quotas should be true")
	}
}

func TestDefaultConfigDocs(t *testing.T) {
	cfg := NewDefaultConfig()

	if !cfg.Docs.Readme {
		t.Error("Default readme should be true")
	}
	if !cfg.Docs.Architecture {
		t.Error("Default architecture should be true")
	}
	if !cfg.Docs.Onboarding {
		t.Error("Default onboarding should be true")
	}
}

func TestDefaultConfigOutput(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.Output.Type != "local" {
		t.Errorf("Default output type should be local, got %s", cfg.Output.Type)
	}
	if cfg.Output.Branch != "main" {
		t.Errorf("Default branch should be main, got %s", cfg.Output.Branch)
	}
}

func TestDefaultConfigEnvironments(t *testing.T) {
	cfg := NewDefaultConfig()

	expectedEnvs := []string{"dev", "staging", "prod"}
	if len(cfg.Environments) != len(expectedEnvs) {
		t.Fatalf("Expected %d environments, got %d", len(expectedEnvs), len(cfg.Environments))
	}

	for i, expected := range expectedEnvs {
		if cfg.Environments[i].Name != expected {
			t.Errorf("Environment %d should be %s, got %s", i, expected, cfg.Environments[i].Name)
		}
	}
}

