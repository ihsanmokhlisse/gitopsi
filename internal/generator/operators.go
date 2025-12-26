package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ihsanmokhlisse/gitopsi/internal/operator"
)

func (g *Generator) generateOperators() error {
	if !g.Config.Operators.Enabled {
		return nil
	}

	enabledOps := getEnabledOperators(g.Config.Operators)
	if len(enabledOps) == 0 {
		return nil
	}

	fmt.Println("🔧 Generating operator manifests...")

	operatorsDir := filepath.Join(g.Config.Project.Name, "infrastructure", "base", "operators")
	if err := g.Writer.CreateDir(operatorsDir); err != nil {
		return fmt.Errorf("failed to create operators directory: %w", err)
	}

	var operatorDirs []string
	for _, op := range enabledOps {
		opDir := filepath.Join(operatorsDir, op.Name)
		if err := g.Writer.CreateDir(opDir); err != nil {
			return fmt.Errorf("failed to create operator directory %s: %w", op.Name, err)
		}
		operatorDirs = append(operatorDirs, op.Name)

		if err := g.generateSubscription(&op, opDir); err != nil {
			return fmt.Errorf("failed to generate subscription for %s: %w", op.Name, err)
		}

		if g.Config.Operators.CreateOperatorGroup {
			if err := g.generateOperatorGroup(&op, opDir); err != nil {
				return fmt.Errorf("failed to generate operator group for %s: %w", op.Name, err)
			}
		}

		if err := g.generateOperatorKustomization(&op, opDir); err != nil {
			return fmt.Errorf("failed to generate kustomization for %s: %w", op.Name, err)
		}
	}

	if err := g.generateOperatorsKustomization(operatorsDir, operatorDirs); err != nil {
		return fmt.Errorf("failed to generate operators kustomization: %w", err)
	}

	return nil
}

func getEnabledOperators(cfg operator.Config) []operator.Operator {
	var enabled []operator.Operator
	for _, op := range cfg.Operators {
		if op.Enabled {
			enabled = append(enabled, op)
		}
	}
	return enabled
}

func (g *Generator) generateSubscription(op *operator.Operator, dir string) error {
	manifest := op.ToSubscriptionManifest(
		g.Config.Operators.DefaultSource,
		g.Config.Operators.DefaultSourceNamespace,
	)

	content := fmt.Sprintf(`apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: %s
  namespace: %s
spec:
  channel: %s
  name: %s
  source: %s
  sourceNamespace: %s
  installPlanApproval: %s`,
		manifest.Name,
		manifest.Namespace,
		manifest.Channel,
		manifest.Name,
		manifest.Source,
		manifest.SourceNamespace,
		manifest.InstallPlanApproval,
	)

	if manifest.StartingCSV != "" {
		content += fmt.Sprintf("\n  startingCSV: %s", manifest.StartingCSV)
	}

	filePath := filepath.Join(dir, "subscription.yaml")
	return g.Writer.WriteFile(filePath, []byte(content+"\n"))
}

func (g *Generator) generateOperatorGroup(op *operator.Operator, dir string) error {
	manifest := op.ToGroupManifest()

	var content string
	if len(manifest.TargetNamespaces) > 0 {
		targetNsYAML := "  targetNamespaces:"
		for _, ns := range manifest.TargetNamespaces {
			targetNsYAML += fmt.Sprintf("\n    - %s", ns)
		}
		content = fmt.Sprintf(`apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: %s
  namespace: %s
spec:
%s
`,
			manifest.Name,
			manifest.Namespace,
			targetNsYAML,
		)
	} else {
		content = fmt.Sprintf(`apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: %s
  namespace: %s
spec: {}
`,
			manifest.Name,
			manifest.Namespace,
		)
	}

	filePath := filepath.Join(dir, "operatorgroup.yaml")
	return g.Writer.WriteFile(filePath, []byte(content))
}

func (g *Generator) generateOperatorKustomization(op *operator.Operator, dir string) error {
	resources := []string{"subscription.yaml"}
	if g.Config.Operators.CreateOperatorGroup {
		resources = append(resources, "operatorgroup.yaml")
	}

	content := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
%s
`,
		formatResourceList(resources),
	)

	filePath := filepath.Join(dir, "kustomization.yaml")
	return g.Writer.WriteFile(filePath, []byte(content))
}

func (g *Generator) generateOperatorsKustomization(dir string, operatorDirs []string) error {
	content := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
%s
`,
		formatResourceList(operatorDirs),
	)

	filePath := filepath.Join(dir, "kustomization.yaml")
	return g.Writer.WriteFile(filePath, []byte(content))
}

func formatResourceList(resources []string) string {
	var lines []string
	for _, r := range resources {
		lines = append(lines, fmt.Sprintf("  - %s", r))
	}
	return strings.Join(lines, "\n")
}
