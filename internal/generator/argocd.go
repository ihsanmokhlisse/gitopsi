package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/templates"
)

func (g *Generator) generateGitOps() error {
	if g.Config.GitOpsTool == "argocd" || g.Config.GitOpsTool == "both" {
		if err := g.generateArgoCD(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) getArgoCDNamespace() string {
	if g.Config.Bootstrap.Namespace != "" {
		return g.Config.Bootstrap.Namespace
	}
	if g.Config.Platform == "openshift" {
		return "openshift-gitops"
	}
	return "argocd"
}

func (g *Generator) generateArgoCD() error {
	fmt.Println("🔄 Generating ArgoCD configuration...")

	argoCDNamespace := g.getArgoCDNamespace()

	if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
		projectData := map[string]string{
			"Name":            "infrastructure",
			"Description":     "Infrastructure resources",
			"ArgoCDNamespace": argoCDNamespace,
		}
		content, err := templates.Render("argocd/project.yaml.tmpl", projectData)
		if err != nil {
			return err
		}
		path := g.Config.Project.Name + "/" + g.Config.GitOpsTool + "/projects/infrastructure.yaml"
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	if g.Config.Scope == "application" || g.Config.Scope == "both" {
		projectData := map[string]string{
			"Name":            "applications",
			"Description":     "Application deployments",
			"ArgoCDNamespace": argoCDNamespace,
		}
		content, err := templates.Render("argocd/project.yaml.tmpl", projectData)
		if err != nil {
			return err
		}
		path := g.Config.Project.Name + "/" + g.Config.GitOpsTool + "/projects/applications.yaml"
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	repoURL := g.Config.Output.URL
	if repoURL == "" {
		repoURL = "https://github.com/org/" + g.Config.Project.Name + ".git"
	}

	for _, env := range g.Config.Environments {
		if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
			appData := map[string]string{
				"Name":            fmt.Sprintf("%s-infra-%s", g.Config.Project.Name, env.Name),
				"Project":         "infrastructure",
				"RepoURL":         repoURL,
				"Path":            fmt.Sprintf("infrastructure/overlays/%s", env.Name),
				"Server":          env.Cluster,
				"Namespace":       g.Config.Project.Name + "-" + env.Name,
				"TargetRevision":  "HEAD",
				"ArgoCDNamespace": argoCDNamespace,
			}
			content, err := templates.Render("argocd/application.yaml.tmpl", appData)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/applicationsets/infra-%s.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, env.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}

		if g.Config.Scope == "application" || g.Config.Scope == "both" {
			appData := map[string]string{
				"Name":            fmt.Sprintf("%s-apps-%s", g.Config.Project.Name, env.Name),
				"Project":         "applications",
				"RepoURL":         repoURL,
				"Path":            fmt.Sprintf("applications/overlays/%s", env.Name),
				"Server":          env.Cluster,
				"Namespace":       g.Config.Project.Name + "-" + env.Name,
				"TargetRevision":  "HEAD",
				"ArgoCDNamespace": argoCDNamespace,
			}
			content, err := templates.Render("argocd/application.yaml.tmpl", appData)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/applicationsets/apps-%s.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, env.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}
	}

	return nil
}
