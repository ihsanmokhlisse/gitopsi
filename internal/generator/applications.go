package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/templates"
)

func (g *Generator) generateApplications() error {
	fmt.Println("📦 Generating applications...")

	if len(g.Config.Apps) == 0 {
		g.Config.Apps = []config.Application{
			{Name: "sample-app", Image: "nginx:latest", Port: 80, Replicas: 1},
		}
	}

	var appDirs []string

	for _, app := range g.Config.Apps {
		appDir := g.Config.Project.Name + "/applications/base/" + app.Name
		if err := g.Writer.CreateDir(appDir); err != nil {
			return err
		}

		appDirs = append(appDirs, app.Name+"/")

		deployContent, err := templates.Render("kubernetes/deployment.yaml.tmpl", app)
		if err != nil {
			return err
		}
		if err := g.Writer.WriteFile(appDir+"/deployment.yaml", deployContent); err != nil {
			return err
		}

		svcContent, err := templates.Render("kubernetes/service.yaml.tmpl", app)
		if err != nil {
			return err
		}
		if err := g.Writer.WriteFile(appDir+"/service.yaml", svcContent); err != nil {
			return err
		}

		appKustomize := map[string]interface{}{
			"Resources": []string{"deployment.yaml", "service.yaml"},
		}
		kContent, err := templates.Render("kubernetes/kustomization.yaml.tmpl", appKustomize)
		if err != nil {
			return err
		}
		if err := g.Writer.WriteFile(appDir+"/kustomization.yaml", kContent); err != nil {
			return err
		}
	}

	baseKustomize := map[string]interface{}{
		"Resources": appDirs,
	}
	content, err := templates.Render("kubernetes/kustomization.yaml.tmpl", baseKustomize)
	if err != nil {
		return err
	}
	if err := g.Writer.WriteFile(g.Config.Project.Name+"/applications/base/kustomization.yaml", content); err != nil {
		return err
	}

	for _, env := range g.Config.Environments {
		overlayData := map[string]interface{}{
			"Resources": []string{"../../base"},
		}
		content, err := templates.Render("kubernetes/kustomization.yaml.tmpl", overlayData)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s/applications/overlays/%s/kustomization.yaml",
			g.Config.Project.Name, env.Name)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	return nil
}

