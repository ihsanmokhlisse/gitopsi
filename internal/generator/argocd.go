package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
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

	if err := g.generateArgoCDProjects(argoCDNamespace); err != nil {
		return err
	}

	if g.Config.IsMultiCluster() {
		return g.generateMultiClusterArgoCD(argoCDNamespace)
	}

	return g.generateSingleClusterArgoCD(argoCDNamespace)
}

func (g *Generator) generateArgoCDProjects(argoCDNamespace string) error {
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

	return nil
}

func (g *Generator) generateSingleClusterArgoCD(argoCDNamespace string) error {
	repoURL := g.Config.Output.URL
	if repoURL == "" {
		repoURL = g.Config.Git.URL
	}
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

func (g *Generator) generateMultiClusterArgoCD(argoCDNamespace string) error {
	repoURL := g.Config.Output.URL
	if repoURL == "" {
		repoURL = g.Config.Git.URL
	}
	if repoURL == "" {
		repoURL = "https://github.com/org/" + g.Config.Project.Name + ".git"
	}

	branch := g.Config.Output.Branch
	if branch == "" {
		branch = "main"
	}

	if err := g.generateClusterSecrets(argoCDNamespace); err != nil {
		return err
	}

	switch g.Config.Topology {
	case config.TopologyClusterPerEnv:
		return g.generateClusterPerEnvApplicationSets(argoCDNamespace, repoURL, branch)
	case config.TopologyMultiCluster:
		return g.generateMultiClusterApplicationSets(argoCDNamespace, repoURL, branch)
	}

	return nil
}

func (g *Generator) generateClusterSecrets(argoCDNamespace string) error {
	for _, env := range g.Config.Environments {
		for _, cluster := range env.Clusters {
			secretData := map[string]any{
				"Name":            cluster.Name,
				"URL":             cluster.URL,
				"Environment":     env.Name,
				"Region":          cluster.Region,
				"Primary":         cluster.Primary,
				"ArgoCDNamespace": argoCDNamespace,
			}
			content, err := templates.Render("argocd/cluster-secret.yaml.tmpl", secretData)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/clusters/%s.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, cluster.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) generateClusterPerEnvApplicationSets(argoCDNamespace, repoURL, branch string) error {
	for _, env := range g.Config.Environments {
		namespace := g.Config.GetEnvironmentNamespace(env.Name)

		if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
			appSetData := map[string]any{
				"Name":            g.Config.Project.Name + "-infra",
				"Environment":     env.Name,
				"Project":         "infrastructure",
				"RepoURL":         repoURL,
				"Branch":          branch,
				"Path":            "infrastructure",
				"Namespace":       namespace,
				"ArgoCDNamespace": argoCDNamespace,
			}
			content, err := templates.Render("argocd/applicationset-cluster.yaml.tmpl", appSetData)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/applicationsets/infra-%s-cluster.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, env.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}

		if g.Config.Scope == "application" || g.Config.Scope == "both" {
			appSetData := map[string]any{
				"Name":            g.Config.Project.Name + "-apps",
				"Environment":     env.Name,
				"Project":         "applications",
				"RepoURL":         repoURL,
				"Branch":          branch,
				"Path":            "applications",
				"Namespace":       namespace,
				"ArgoCDNamespace": argoCDNamespace,
			}
			content, err := templates.Render("argocd/applicationset-cluster.yaml.tmpl", appSetData)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/applicationsets/apps-%s-cluster.yaml",
				g.Config.Project.Name, g.Config.GitOpsTool, env.Name)
			if err := g.Writer.WriteFile(path, content); err != nil {
				return err
			}
		}
	}

	return nil
}

type envInfo struct {
	Name      string
	Namespace string
}

func (g *Generator) generateMultiClusterApplicationSets(argoCDNamespace, repoURL, branch string) error {
	envList := make([]envInfo, 0, len(g.Config.Environments))
	for _, env := range g.Config.Environments {
		envList = append(envList, envInfo{
			Name:      env.Name,
			Namespace: g.Config.GetEnvironmentNamespace(env.Name),
		})
	}

	if g.Config.Scope == "infrastructure" || g.Config.Scope == "both" {
		appSetData := map[string]any{
			"Name":            g.Config.Project.Name + "-infra",
			"Environments":    envList,
			"Project":         "infrastructure",
			"RepoURL":         repoURL,
			"Branch":          branch,
			"Path":            "infrastructure",
			"ArgoCDNamespace": argoCDNamespace,
		}
		content, err := templates.Render("argocd/applicationset-matrix.yaml.tmpl", appSetData)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/%s/applicationsets/infra-multi-cluster.yaml",
			g.Config.Project.Name, g.Config.GitOpsTool)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	if g.Config.Scope == "application" || g.Config.Scope == "both" {
		appSetData := map[string]any{
			"Name":            g.Config.Project.Name + "-apps",
			"Environments":    envList,
			"Project":         "applications",
			"RepoURL":         repoURL,
			"Branch":          branch,
			"Path":            "applications",
			"ArgoCDNamespace": argoCDNamespace,
		}
		content, err := templates.Render("argocd/applicationset-matrix.yaml.tmpl", appSetData)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/%s/applicationsets/apps-multi-cluster.yaml",
			g.Config.Project.Name, g.Config.GitOpsTool)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	return nil
}
