package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/templates"
)

func (g *Generator) generateInfrastructure() error {
	fmt.Println("🏗️  Generating infrastructure...")

	for _, env := range g.Config.Environments {
		nsData := map[string]string{
			"Name": g.Config.Project.Name + "-" + env.Name,
			"Env":  env.Name,
		}

		content, err := templates.Render("infrastructure/namespace.yaml.tmpl", nsData)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s/infrastructure/base/namespaces/%s.yaml",
			g.Config.Project.Name, env.Name)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	kustomizeData := map[string]interface{}{
		"Resources": []string{"namespaces/"},
	}

	content, err := templates.Render("kubernetes/kustomization.yaml.tmpl", kustomizeData)
	if err != nil {
		return err
	}

	path := g.Config.Project.Name + "/infrastructure/base/kustomization.yaml"
	if err := g.Writer.WriteFile(path, content); err != nil {
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

		path := fmt.Sprintf("%s/infrastructure/overlays/%s/kustomization.yaml",
			g.Config.Project.Name, env.Name)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	return nil
}

