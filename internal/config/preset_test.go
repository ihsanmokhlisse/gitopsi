package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresetConstants(t *testing.T) {
	assert.Equal(t, Preset("minimal"), PresetMinimal)
	assert.Equal(t, Preset("standard"), PresetStandard)
	assert.Equal(t, Preset("enterprise"), PresetEnterprise)
	assert.Equal(t, Preset("custom"), PresetCustom)
}

func TestValidPresets(t *testing.T) {
	presets := ValidPresets()
	assert.Contains(t, presets, "minimal")
	assert.Contains(t, presets, "standard")
	assert.Contains(t, presets, "enterprise")
	assert.Contains(t, presets, "custom")
	assert.Len(t, presets, 4)
}

func TestIsValidPreset(t *testing.T) {
	assert.True(t, IsValidPreset("minimal"))
	assert.True(t, IsValidPreset("standard"))
	assert.True(t, IsValidPreset("enterprise"))
	assert.True(t, IsValidPreset("custom"))
	assert.True(t, IsValidPreset(""))
	assert.False(t, IsValidPreset("invalid"))
	assert.False(t, IsValidPreset("MINIMAL"))
}

func TestApplyMinimalPreset(t *testing.T) {
	cfg := &Config{
		Preset:  PresetMinimal,
		Project: Project{Name: "test"},
	}

	cfg.ApplyPreset()

	assert.True(t, cfg.Infra.Namespaces)
	assert.False(t, cfg.Infra.RBAC)
	assert.False(t, cfg.Infra.NetworkPolicies)
	assert.False(t, cfg.Infra.ResourceQuotas)
	assert.True(t, cfg.Docs.Readme)
	assert.False(t, cfg.Docs.Architecture)
	assert.False(t, cfg.Docs.Onboarding)
	assert.Len(t, cfg.Environments, 1)
	assert.Equal(t, "dev", cfg.Environments[0].Name)
}

func TestApplyStandardPreset(t *testing.T) {
	cfg := &Config{
		Preset:  PresetStandard,
		Project: Project{Name: "test"},
	}

	cfg.ApplyPreset()

	assert.True(t, cfg.Infra.Namespaces)
	assert.True(t, cfg.Infra.RBAC)
	assert.True(t, cfg.Infra.NetworkPolicies)
	assert.True(t, cfg.Infra.ResourceQuotas)
	assert.True(t, cfg.Docs.Readme)
	assert.True(t, cfg.Docs.Architecture)
	assert.True(t, cfg.Docs.Onboarding)
	assert.Len(t, cfg.Environments, 3)
}

func TestApplyEnterprisePreset(t *testing.T) {
	cfg := &Config{
		Preset:  PresetEnterprise,
		Project: Project{Name: "test"},
	}

	cfg.ApplyPreset()

	assert.True(t, cfg.Infra.Namespaces)
	assert.True(t, cfg.Infra.RBAC)
	assert.True(t, cfg.Infra.NetworkPolicies)
	assert.True(t, cfg.Infra.ResourceQuotas)
	assert.True(t, cfg.Docs.Readme)
	assert.True(t, cfg.Docs.Architecture)
	assert.True(t, cfg.Docs.Onboarding)
	assert.GreaterOrEqual(t, len(cfg.Structure.CustomDirs), 3)
	assert.Len(t, cfg.Environments, 3)
}

func TestPresetPreservesExistingEnvironments(t *testing.T) {
	cfg := &Config{
		Preset:  PresetMinimal,
		Project: Project{Name: "test"},
		Environments: []Environment{
			{Name: "custom-env"},
		},
	}

	cfg.ApplyPreset()

	assert.Len(t, cfg.Environments, 1)
	assert.Equal(t, "custom-env", cfg.Environments[0].Name)
}

func TestGetInfrastructureDir(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "infrastructure", cfg.GetInfrastructureDir())

	cfg.Structure.InfrastructureDir = "infra"
	assert.Equal(t, "infra", cfg.GetInfrastructureDir())
}

func TestGetApplicationsDir(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "applications", cfg.GetApplicationsDir())

	cfg.Structure.ApplicationsDir = "apps"
	assert.Equal(t, "apps", cfg.GetApplicationsDir())
}

func TestGetBootstrapDir(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "bootstrap", cfg.GetBootstrapDir())

	cfg.Structure.BootstrapDir = "setup"
	assert.Equal(t, "setup", cfg.GetBootstrapDir())
}

func TestGetScriptsDir(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "scripts", cfg.GetScriptsDir())

	cfg.Structure.ScriptsDir = "bin"
	assert.Equal(t, "bin", cfg.GetScriptsDir())
}

func TestGetDocsDir(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "docs", cfg.GetDocsDir())

	cfg.Structure.DocsDir = "documentation"
	assert.Equal(t, "documentation", cfg.GetDocsDir())
}

func TestShouldGenerateNamespaces(t *testing.T) {
	cfg := &Config{Infra: Infrastructure{Namespaces: true}}
	assert.True(t, cfg.ShouldGenerateNamespaces())

	cfg.Infra.Namespaces = false
	assert.False(t, cfg.ShouldGenerateNamespaces())

	// Override via Generate
	val := true
	cfg.Generate.Infrastructure.Namespaces = &val
	assert.True(t, cfg.ShouldGenerateNamespaces())
}

func TestShouldGenerateRBAC(t *testing.T) {
	cfg := &Config{Infra: Infrastructure{RBAC: true}}
	assert.True(t, cfg.ShouldGenerateRBAC())

	cfg.Infra.RBAC = false
	assert.False(t, cfg.ShouldGenerateRBAC())

	val := true
	cfg.Generate.Infrastructure.RBAC = &val
	assert.True(t, cfg.ShouldGenerateRBAC())
}

func TestShouldGenerateNetworkPolicies(t *testing.T) {
	cfg := &Config{Infra: Infrastructure{NetworkPolicies: true}}
	assert.True(t, cfg.ShouldGenerateNetworkPolicies())

	cfg.Infra.NetworkPolicies = false
	assert.False(t, cfg.ShouldGenerateNetworkPolicies())

	val := true
	cfg.Generate.Infrastructure.NetworkPolicies = &val
	assert.True(t, cfg.ShouldGenerateNetworkPolicies())
}

func TestShouldGenerateResourceQuotas(t *testing.T) {
	cfg := &Config{Infra: Infrastructure{ResourceQuotas: true}}
	assert.True(t, cfg.ShouldGenerateResourceQuotas())

	cfg.Infra.ResourceQuotas = false
	assert.False(t, cfg.ShouldGenerateResourceQuotas())

	val := true
	cfg.Generate.Infrastructure.ResourceQuotas = &val
	assert.True(t, cfg.ShouldGenerateResourceQuotas())
}

func TestShouldGenerateDocs(t *testing.T) {
	cfg := &Config{Docs: Documentation{Readme: true, Architecture: true, Onboarding: true}}
	assert.True(t, cfg.ShouldGenerateReadme())
	assert.True(t, cfg.ShouldGenerateArchitecture())
	assert.True(t, cfg.ShouldGenerateOnboarding())

	cfg.Docs = Documentation{Readme: false, Architecture: false, Onboarding: false}
	assert.False(t, cfg.ShouldGenerateReadme())
	assert.False(t, cfg.ShouldGenerateArchitecture())
	assert.False(t, cfg.ShouldGenerateOnboarding())

	// Override via Generate
	val := true
	cfg.Generate.Docs.Readme = &val
	cfg.Generate.Docs.Architecture = &val
	cfg.Generate.Docs.Onboarding = &val
	assert.True(t, cfg.ShouldGenerateReadme())
	assert.True(t, cfg.ShouldGenerateArchitecture())
	assert.True(t, cfg.ShouldGenerateOnboarding())
}

func TestCustomDirs(t *testing.T) {
	cfg := &Config{
		Structure: StructureConfig{
			CustomDirs: []CustomDir{
				{Path: "monitoring/dashboards", Description: "Grafana dashboards"},
				{Path: "policies/opa", Description: "OPA policies"},
			},
		},
	}

	assert.Len(t, cfg.Structure.CustomDirs, 2)
	assert.Equal(t, "monitoring/dashboards", cfg.Structure.CustomDirs[0].Path)
	assert.Equal(t, "Grafana dashboards", cfg.Structure.CustomDirs[0].Description)
}
