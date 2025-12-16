package generator

import (
	"fmt"

	"github.com/ihsanmokhlisse/gitopsi/internal/templates"
)

func (g *Generator) generateDocs() error {
	fmt.Println("📚 Generating documentation...")

	if g.Config.Docs.Readme {
		content, err := templates.Render("docs/README.md.tmpl", g.Config)
		if err != nil {
			return err
		}

		path := g.Config.Project.Name + "/README.md"
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	if g.Config.Docs.Architecture {
		content, err := templates.Render("docs/ARCHITECTURE.md.tmpl", g.Config)
		if err != nil {
			return err
		}

		path := g.Config.Project.Name + "/docs/ARCHITECTURE.md"
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	if g.Config.Docs.Onboarding {
		content, err := templates.Render("docs/ONBOARDING.md.tmpl", g.Config)
		if err != nil {
			return err
		}

		path := g.Config.Project.Name + "/docs/ONBOARDING.md"
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateBootstrap() error {
	fmt.Println("🔧 Generating bootstrap...")

	bootstrapContent := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, g.Config.GitOpsTool)

	path := fmt.Sprintf("%s/bootstrap/%s/namespace.yaml",
		g.Config.Project.Name, g.Config.GitOpsTool)
	if err := g.Writer.WriteFile(path, []byte(bootstrapContent)); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateScripts() error {
	fmt.Println("📜 Generating scripts...")

	bootstrapScript := fmt.Sprintf(`#!/bin/bash
set -e

echo "Bootstrapping %s..."

# Apply namespace
kubectl apply -f bootstrap/%s/namespace.yaml

# Apply GitOps tool
echo "Apply your %s installation manifests here"

echo "Bootstrap complete!"
`, g.Config.Project.Name, g.Config.GitOpsTool, g.Config.GitOpsTool)

	path := g.Config.Project.Name + "/scripts/bootstrap.sh"
	if err := g.Writer.WriteFile(path, []byte(bootstrapScript)); err != nil {
		return err
	}

	validateScript := `#!/bin/bash
set -e

echo "Validating repository..."

# Check YAML syntax
find . -name "*.yaml" -o -name "*.yml" | while read f; do
  yamllint "$f" 2>/dev/null || echo "Warning: yamllint not installed"
  break
done

# Build with kustomize
for dir in infrastructure applications; do
  if [ -d "$dir/base" ]; then
    echo "Building $dir..."
    kustomize build "$dir/base" > /dev/null
  fi
done

echo "Validation complete!"
`

	path = g.Config.Project.Name + "/scripts/validate.sh"
	if err := g.Writer.WriteFile(path, []byte(validateScript)); err != nil {
		return err
	}

	return nil
}
