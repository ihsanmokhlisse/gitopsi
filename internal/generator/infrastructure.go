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

	if g.Config.Infra.RBAC {
		if err := g.generateRBAC(); err != nil {
			return err
		}
	}

	if g.Config.Infra.NetworkPolicies {
		if err := g.generateNetworkPolicies(); err != nil {
			return err
		}
	}

	if g.Config.Infra.ResourceQuotas {
		if err := g.generateResourceQuotas(); err != nil {
			return err
		}
	}

	resources := []string{"namespaces/"}
	if g.Config.Infra.RBAC {
		resources = append(resources, "rbac/")
	}

	kustomizeData := map[string]interface{}{
		"Resources": resources,
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

func (g *Generator) generateRBAC() error {
	for _, env := range g.Config.Environments {
		rbacData := map[string]string{
			"Name":      g.Config.Project.Name,
			"Namespace": g.Config.Project.Name + "-" + env.Name,
			"Env":       env.Name,
		}

		content, err := templates.Render("infrastructure/rbac.yaml.tmpl", rbacData)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s/infrastructure/base/rbac/%s.yaml",
			g.Config.Project.Name, env.Name)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateNetworkPolicies() error {
	for _, env := range g.Config.Environments {
		npData := map[string]string{
			"Name":      g.Config.Project.Name,
			"Namespace": g.Config.Project.Name + "-" + env.Name,
			"Env":       env.Name,
		}

		content, err := templates.Render("infrastructure/networkpolicy.yaml.tmpl", npData)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s/infrastructure/base/network-policies/%s.yaml",
			g.Config.Project.Name, env.Name)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateResourceQuotas() error {
	quotaDefaults := map[string]map[string]string{
		"dev":     {"RequestsCPU": "2", "RequestsMemory": "4Gi", "LimitsCPU": "4", "LimitsMemory": "8Gi", "MaxPods": "20", "MaxServices": "10", "MaxConfigMaps": "20", "MaxSecrets": "20"},
		"staging": {"RequestsCPU": "4", "RequestsMemory": "8Gi", "LimitsCPU": "8", "LimitsMemory": "16Gi", "MaxPods": "50", "MaxServices": "20", "MaxConfigMaps": "50", "MaxSecrets": "50"},
		"prod":    {"RequestsCPU": "8", "RequestsMemory": "16Gi", "LimitsCPU": "16", "LimitsMemory": "32Gi", "MaxPods": "100", "MaxServices": "50", "MaxConfigMaps": "100", "MaxSecrets": "100"},
	}

	defaultQuota := map[string]string{
		"RequestsCPU": "4", "RequestsMemory": "8Gi", "LimitsCPU": "8", "LimitsMemory": "16Gi",
		"MaxPods": "50", "MaxServices": "20", "MaxConfigMaps": "50", "MaxSecrets": "50",
	}

	for _, env := range g.Config.Environments {
		quota, ok := quotaDefaults[env.Name]
		if !ok {
			quota = defaultQuota
		}

		rqData := map[string]string{
			"Name":           g.Config.Project.Name,
			"Namespace":      g.Config.Project.Name + "-" + env.Name,
			"Env":            env.Name,
			"RequestsCPU":    quota["RequestsCPU"],
			"RequestsMemory": quota["RequestsMemory"],
			"LimitsCPU":      quota["LimitsCPU"],
			"LimitsMemory":   quota["LimitsMemory"],
			"MaxPods":        quota["MaxPods"],
			"MaxServices":    quota["MaxServices"],
			"MaxConfigMaps":  quota["MaxConfigMaps"],
			"MaxSecrets":     quota["MaxSecrets"],
		}

		content, err := templates.Render("infrastructure/resourcequota.yaml.tmpl", rqData)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s/infrastructure/base/resource-quotas/%s.yaml",
			g.Config.Project.Name, env.Name)
		if err := g.Writer.WriteFile(path, content); err != nil {
			return err
		}
	}
	return nil
}
