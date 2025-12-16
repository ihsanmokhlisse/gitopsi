package config

import (
	"fmt"
	"slices"
)

var (
	validPlatforms   = []string{"kubernetes", "openshift", "aks", "eks"}
	validScopes      = []string{"infrastructure", "application", "both"}
	validGitOpsTools = []string{"argocd", "flux", "both"}
	validOutputTypes = []string{"local", "git"}
)

func (c *Config) Validate() error {
	if c.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if !slices.Contains(validPlatforms, c.Platform) {
		return fmt.Errorf("invalid platform: %s (valid: %v)", c.Platform, validPlatforms)
	}

	if !slices.Contains(validScopes, c.Scope) {
		return fmt.Errorf("invalid scope: %s (valid: %v)", c.Scope, validScopes)
	}

	if !slices.Contains(validGitOpsTools, c.GitOpsTool) {
		return fmt.Errorf("invalid gitops_tool: %s (valid: %v)", c.GitOpsTool, validGitOpsTools)
	}

	if !slices.Contains(validOutputTypes, c.Output.Type) {
		return fmt.Errorf("invalid output type: %s (valid: %v)", c.Output.Type, validOutputTypes)
	}

	if c.Output.Type == "git" && c.Output.URL == "" {
		return fmt.Errorf("git URL is required when output type is 'git'")
	}

	if len(c.Environments) == 0 {
		return fmt.Errorf("at least one environment is required")
	}

	for i, env := range c.Environments {
		if env.Name == "" {
			return fmt.Errorf("environment %d: name is required", i)
		}
	}

	return nil
}

func ValidPlatforms() []string {
	return validPlatforms
}

func ValidScopes() []string {
	return validScopes
}

func ValidGitOpsTools() []string {
	return validGitOpsTools
}

