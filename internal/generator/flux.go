package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/templates"
)

func (g *Generator) getFluxNamespace() string {
	if g.Config.Bootstrap.Namespace != "" {
		return g.Config.Bootstrap.Namespace
	}
	return "flux-system"
}

func (g *Generator) generateFlux() error {
	fmt.Println("🔄 Generating Flux configuration...")

	fluxNamespace := g.getFluxNamespace()

	if err := g.generateFluxGitRepository(fluxNamespace); err != nil {
		return err
	}

	if err := g.generateFluxKustomizations(fluxNamespace); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateFluxGitRepository(fluxNamespace string) error {
	repoURL := g.Config.Git.URL
	if repoURL == "" {
		repoURL = g.Config.Output.URL
	}
	if repoURL == "" {
		return fmt.Errorf("git.url is required to generate Flux GitRepository - Flux needs to sync from a Git repository")
	}

	branch := g.Config.Git.Branch
	if branch == "" {
		branch = "main"
	}

	gitRepoData := map[string]any{
		"Name":      g.Config.Project.Name,
		"Namespace": fluxNamespace,
		"Interval":  "1m",
		"URL":       repoURL,
		"Branch":    branch,
		"SecretRef": "",
	}

	content, err := templates.Render("flux/gitrepository.yaml.tmpl", gitRepoData)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s/sources/gitrepository.yaml", g.Config.Project.Name, g.Config.GitOpsTool)
	if err := g.Writer.WriteFile(path, content); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateFluxKustomizations(fluxNamespace string) error {
	for _, env := range g.Config.Environments {
		namespace := g.Config.Project.Name + "-" + env.Name

		if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
			kustomizationData := map[string]any{
				"Name":            fmt.Sprintf("%s-infra-%s", g.Config.Project.Name, env.Name),
				"Namespace":       fluxNamespace,
				"Interval":        "10m",
				"SourceName":      g.Config.Project.Name,
				"Path":            fmt.Sprintf("./infrastructure/overlays/%s", env.Name),
				"Prune":           true,
				"TargetNamespace": namespace,
				"HealthChecks":    []any{},
				"DependsOn":       []string{},
			}

			content, err := templates.Render("flux/kustomization.yaml.tmpl", kustomizationData)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("%s/%s/kustomizations/infra-%s.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, env.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}

		if g.Config.Scope == "application" || g.Config.Scope == "both" {
			dependsOn := []string{}
			if g.Config.Scope == "both" {
				dependsOn = append(dependsOn, fmt.Sprintf("%s-infra-%s", g.Config.Project.Name, env.Name))
			}

			kustomizationData := map[string]any{
				"Name":            fmt.Sprintf("%s-apps-%s", g.Config.Project.Name, env.Name),
				"Namespace":       fluxNamespace,
				"Interval":        "10m",
				"SourceName":      g.Config.Project.Name,
				"Path":            fmt.Sprintf("./applications/overlays/%s", env.Name),
				"Prune":           true,
				"TargetNamespace": namespace,
				"HealthChecks":    []any{},
				"DependsOn":       dependsOn,
			}

			content, err := templates.Render("flux/kustomization.yaml.tmpl", kustomizationData)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("%s/%s/kustomizations/apps-%s.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, env.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) generateFluxNotifications(fluxNamespace string) error {
	// Provider for notifications (e.g., Slack, Discord, etc.)
	providerData := map[string]any{
		"Name":      g.Config.Project.Name + "-alerts",
		"Namespace": fluxNamespace,
		"Type":      "slack",
		"Address":   "",
		"Channel":   "",
		"SecretRef": g.Config.Project.Name + "-slack-url",
	}

	content, err := templates.Render("flux/provider.yaml.tmpl", providerData)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s/notifications/provider.yaml", g.Config.Project.Name, g.Config.GitOpsTool)
	if err := g.Writer.WriteFile(path, content); err != nil {
		return err
	}

	// Create event sources for alert
	eventSources := make([]map[string]string, 0)
	for _, env := range g.Config.Environments {
		if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
			eventSources = append(eventSources, map[string]string{
				"Kind":      "Kustomization",
				"Name":      fmt.Sprintf("%s-infra-%s", g.Config.Project.Name, env.Name),
				"Namespace": fluxNamespace,
			})
		}
		if g.Config.Scope == "application" || g.Config.Scope == "both" {
			eventSources = append(eventSources, map[string]string{
				"Kind":      "Kustomization",
				"Name":      fmt.Sprintf("%s-apps-%s", g.Config.Project.Name, env.Name),
				"Namespace": fluxNamespace,
			})
		}
	}

	alertData := map[string]any{
		"Name":         g.Config.Project.Name + "-alerts",
		"Namespace":    fluxNamespace,
		"ProviderRef":  g.Config.Project.Name + "-alerts",
		"Severity":     "info",
		"EventSources": eventSources,
	}

	content, err = templates.Render("flux/alert.yaml.tmpl", alertData)
	if err != nil {
		return err
	}

	path = fmt.Sprintf("%s/%s/notifications/alert.yaml", g.Config.Project.Name, g.Config.GitOpsTool)
	return g.Writer.WriteFile(path, content)
}
