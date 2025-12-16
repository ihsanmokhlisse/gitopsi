package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
)

type Generator struct {
	Config  *config.Config
	Writer  *output.Writer
	Verbose bool
}

func New(cfg *config.Config, writer *output.Writer, verbose bool) *Generator {
	return &Generator{
		Config:  cfg,
		Writer:  writer,
		Verbose: verbose,
	}
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
			g.Config.Project.Name+"/infrastructure/base/rbac",
		)
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

	dirs = append(dirs,
		g.Config.Project.Name+"/"+g.Config.GitOpsTool+"/projects",
		g.Config.Project.Name+"/"+g.Config.GitOpsTool+"/applicationsets",
	)

	for _, dir := range dirs {
		if err := g.Writer.CreateDir(dir); err != nil {
			return err
		}
	}

	return nil
}

