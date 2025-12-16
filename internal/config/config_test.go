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

