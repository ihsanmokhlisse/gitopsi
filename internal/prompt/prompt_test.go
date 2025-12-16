package prompt

import (
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
)

func TestRunReturnsConfig(t *testing.T) {
	t.Skip("Requires interactive input - tested via integration tests")
}

func TestValidPlatformsUsedInPrompt(t *testing.T) {
	platforms := config.ValidPlatforms()
	if len(platforms) == 0 {
		t.Error("ValidPlatforms() should return platforms")
	}

	expected := []string{"kubernetes", "openshift", "aks", "eks"}
	for _, p := range expected {
		found := false
		for _, v := range platforms {
			if v == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing platform: %s", p)
		}
	}
}

func TestValidScopesUsedInPrompt(t *testing.T) {
	scopes := config.ValidScopes()
	if len(scopes) == 0 {
		t.Error("ValidScopes() should return scopes")
	}

	expected := []string{"infrastructure", "application", "both"}
	for _, s := range expected {
		found := false
		for _, v := range scopes {
			if v == s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing scope: %s", s)
		}
	}
}

func TestValidGitOpsToolsUsedInPrompt(t *testing.T) {
	tools := config.ValidGitOpsTools()
	if len(tools) == 0 {
		t.Error("ValidGitOpsTools() should return tools")
	}

	expected := []string{"argocd", "flux", "both"}
	for _, tool := range expected {
		found := false
		for _, v := range tools {
			if v == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing tool: %s", tool)
		}
	}
}

func TestDefaultConfigValuesForPrompt(t *testing.T) {
	cfg := config.NewDefaultConfig()

	if cfg.Platform != "kubernetes" {
		t.Errorf("Default platform should be kubernetes, got %s", cfg.Platform)
	}

	if cfg.Scope != "both" {
		t.Errorf("Default scope should be both, got %s", cfg.Scope)
	}

	if cfg.GitOpsTool != "argocd" {
		t.Errorf("Default gitops tool should be argocd, got %s", cfg.GitOpsTool)
	}

	if len(cfg.Environments) != 3 {
		t.Errorf("Default should have 3 environments, got %d", len(cfg.Environments))
	}
}

