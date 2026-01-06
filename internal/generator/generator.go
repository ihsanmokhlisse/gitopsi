package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
	"github.com/ihsanmokhlisse/gitopsi/internal/version"
)

// Generator handles the generation of GitOps repository structure and manifests.
type Generator struct {
	Config        *config.Config
	Writer        *output.Writer
	Verbose       bool
	VersionMapper *version.Mapper
	Deprecations  []version.DeprecationResult
}

// New creates a new Generator with the given configuration.
func New(cfg *config.Config, writer *output.Writer, verbose bool) *Generator {
	g := &Generator{
		Config:       cfg,
		Writer:       writer,
		Verbose:      verbose,
		Deprecations: []version.DeprecationResult{},
	}

	// Initialize version mapper if target version is specified
	targetVersion := ""
	if cfg.Version.Kubernetes != "" {
		targetVersion = cfg.Version.Kubernetes
	} else if cfg.Version.OpenShift != "" {
		// Convert OpenShift version to Kubernetes version
		if k8sVersion, ok := version.GetKubernetesVersionForOpenShift(cfg.Version.OpenShift); ok {
			targetVersion = k8sVersion
		}
	}

	if targetVersion != "" {
		vm, err := version.NewMapper(targetVersion, cfg.Platform)
		if err == nil {
			g.VersionMapper = vm
		}
	} else {
		// Create a mapper without target version for default API versions
		vm, _ := version.NewMapper("", cfg.Platform)
		g.VersionMapper = vm
	}

	return g
}

// GetAPIVersion returns the appropriate API version for a resource kind.
func (g *Generator) GetAPIVersion(kind string) string {
	if g.VersionMapper != nil {
		return g.VersionMapper.GetAPIVersion(kind)
	}
	if api, ok := version.DefaultAPIVersions[kind]; ok {
		return api
	}
	return ""
}

// CheckDeprecation checks if an API version is deprecated and records it.
func (g *Generator) CheckDeprecation(kind, apiVersion, name, filePath string) {
	if g.VersionMapper == nil {
		return
	}

	if result := g.VersionMapper.CheckDeprecation(kind, apiVersion); result != nil {
		result.Name = name
		result.FilePath = filePath
		g.Deprecations = append(g.Deprecations, *result)

		// Log warning if verbose
		if g.Verbose || g.Config.Version.WarnOnDeprecated {
			if result.Severity == "error" {
				fmt.Printf("  ❌ %s: %s\n", filePath, result.Message)
			} else {
				fmt.Printf("  ⚠️  %s: %s\n", filePath, result.Message)
			}
		}
	}
}

// GetDeprecations returns all recorded deprecations.
func (g *Generator) GetDeprecations() []version.DeprecationResult {
	return g.Deprecations
}

// HasCriticalDeprecations checks if there are any critical (error-level) deprecations.
func (g *Generator) HasCriticalDeprecations() bool {
	for _, d := range g.Deprecations {
		if d.Severity == "error" {
			return true
		}
	}
	return false
}

func (g *Generator) Generate() error {
	fmt.Printf("\n🚀 Generating GitOps repository: %s\n\n", g.Config.Project.Name)

	if err := g.generateStructure(); err != nil {
		return fmt.Errorf("failed to generate structure: %w", err)
	}

	if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
		if err := g.generateInfrastructure(); err != nil {
			return fmt.Errorf("failed to generate infrastructure: %w", err)
		}
	}

	if g.Config.Scope == "application" || g.Config.Scope == "both" {
		if err := g.generateApplications(); err != nil {
			return fmt.Errorf("failed to generate applications: %w", err)
		}
	}

	if err := g.generateGitOps(); err != nil {
		return fmt.Errorf("failed to generate gitops config: %w", err)
	}

	if g.Config.Docs.Readme {
		if err := g.generateDocs(); err != nil {
			return fmt.Errorf("failed to generate docs: %w", err)
		}
	}

	if err := g.generateBootstrap(); err != nil {
		return fmt.Errorf("failed to generate bootstrap: %w", err)
	}

	if err := g.generateScripts(); err != nil {
		return fmt.Errorf("failed to generate scripts: %w", err)
	}

	if err := g.generateOperators(); err != nil {
		return fmt.Errorf("failed to generate operators: %w", err)
	}

	fmt.Printf("\n✅ Generated: %s/\n", g.Config.Project.Name)
	return nil
}

func (g *Generator) generateStructure() error {
	fmt.Println("📁 Creating directory structure...")

	dirs := []string{
		g.Config.Project.Name,
		g.Config.Project.Name + "/docs",
		g.Config.Project.Name + "/bootstrap/" + g.Config.GitOpsTool,
		g.Config.Project.Name + "/scripts",
	}

	if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
		dirs = append(dirs,
			g.Config.Project.Name+"/infrastructure/base",
			g.Config.Project.Name+"/infrastructure/base/namespaces",
		)
		if g.Config.Infra.RBAC {
			dirs = append(dirs, g.Config.Project.Name+"/infrastructure/base/rbac")
		}
		if g.Config.Infra.NetworkPolicies {
			dirs = append(dirs, g.Config.Project.Name+"/infrastructure/base/network-policies")
		}
		if g.Config.Infra.ResourceQuotas {
			dirs = append(dirs, g.Config.Project.Name+"/infrastructure/base/resource-quotas")
		}
		for _, env := range g.Config.Environments {
			dirs = append(dirs, g.Config.Project.Name+"/infrastructure/overlays/"+env.Name)
		}
	}

	if g.Config.Scope == "application" || g.Config.Scope == "both" {
		dirs = append(dirs, g.Config.Project.Name+"/applications/base")
		for _, env := range g.Config.Environments {
			dirs = append(dirs, g.Config.Project.Name+"/applications/overlays/"+env.Name)
		}
	}

	// Add GitOps tool-specific directories
	if g.Config.GitOpsTool == "argocd" || g.Config.GitOpsTool == "both" {
		dirs = append(dirs,
			g.Config.Project.Name+"/argocd/projects",
			g.Config.Project.Name+"/argocd/applicationsets",
		)
		if g.Config.IsMultiCluster() {
			dirs = append(dirs, g.Config.Project.Name+"/argocd/clusters")
		}
	}

	// TODO: Flux support disabled - focus on ArgoCD first
	// Uncomment when Flux support is ready for production
	// if g.Config.GitOpsTool == "flux" || g.Config.GitOpsTool == "both" {
	// 	dirs = append(dirs,
	// 		g.Config.Project.Name+"/flux/sources",
	// 		g.Config.Project.Name+"/flux/kustomizations",
	// 		g.Config.Project.Name+"/flux/notifications",
	// 	)
	// }

	for _, dir := range dirs {
		if err := g.Writer.CreateDir(dir); err != nil {
			return err
		}
	}

	return nil
}
