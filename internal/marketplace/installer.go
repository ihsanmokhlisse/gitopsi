package marketplace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// InstallOptions defines options for pattern installation.
type InstallOptions struct {
	Version      string
	Config       map[string]any
	Environments []string
	DryRun       bool
	Force        bool
	SkipDeps     bool
	AutoApprove  bool
}

// InstallResult represents the result of a pattern installation.
type InstallResult struct {
	Pattern       string
	Version       string
	Success       bool
	Message       string
	GeneratedPath []string
	AccessInfo    map[string]string
	Dependencies  []DependencyResult
	Errors        []string
	Warnings      []string
}

// DependencyResult represents the result of installing a dependency.
type DependencyResult struct {
	Name     string
	Version  string
	Status   string // installed, skipped, failed
	Optional bool
	Message  string
}

// Installer handles pattern installation.
type Installer struct {
	registry    *RegistryManager
	projectPath string
	stateFile   string
	gitOpsTool  string
	platform    string
	installed   map[string]*InstalledPattern
}

// NewInstaller creates a new pattern installer.
func NewInstaller(registry *RegistryManager, projectPath, gitOpsTool, platform string) *Installer {
	return &Installer{
		registry:    registry,
		projectPath: projectPath,
		stateFile:   filepath.Join(projectPath, ".gitopsi", "patterns.yaml"),
		gitOpsTool:  gitOpsTool,
		platform:    platform,
		installed:   make(map[string]*InstalledPattern),
	}
}

// LoadState loads the installed patterns state.
func (i *Installer) LoadState() error {
	data, err := os.ReadFile(i.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state struct {
		Patterns map[string]*InstalledPattern `yaml:"patterns"`
	}
	if err := yaml.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	i.installed = state.Patterns
	return nil
}

// SaveState saves the installed patterns state.
func (i *Installer) SaveState() error {
	state := struct {
		Version  string                       `yaml:"version"`
		Updated  time.Time                    `yaml:"updated"`
		Patterns map[string]*InstalledPattern `yaml:"patterns"`
	}{
		Version:  "1.0",
		Updated:  time.Now(),
		Patterns: i.installed,
	}

	data, err := yaml.Marshal(&state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	dir := filepath.Dir(i.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	return os.WriteFile(i.stateFile, data, 0644)
}

// Install installs a pattern.
func (i *Installer) Install(ctx context.Context, patternName string, opts InstallOptions) (*InstallResult, error) {
	result := &InstallResult{
		Pattern:      patternName,
		AccessInfo:   make(map[string]string),
		Dependencies: []DependencyResult{},
	}

	// Load current state
	if err := i.LoadState(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to load state: %v", err))
	}

	// Check if already installed
	if existing, ok := i.installed[patternName]; ok && !opts.Force {
		result.Message = fmt.Sprintf("Pattern '%s' is already installed (version %s). Use --force to reinstall.",
			patternName, existing.Pattern.Metadata.Version)
		return result, nil
	}

	// Find and fetch the pattern
	entry, registryName, err := i.registry.FindPattern(ctx, patternName)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	version := opts.Version
	if version == "" {
		version = entry.Latest
	}
	result.Version = version

	pattern, err := i.registry.FetchPattern(ctx, registryName, patternName, version)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to fetch pattern: %v", err))
		return result, err
	}

	// Check compatibility
	if !pattern.IsCompatibleWithPlatform(i.platform) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Pattern may not be fully compatible with platform '%s'", i.platform))
	}
	if !pattern.IsCompatibleWithTool(i.gitOpsTool) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Pattern may not be fully compatible with GitOps tool '%s'", i.gitOpsTool))
	}

	// Install dependencies first
	if !opts.SkipDeps {
		for _, dep := range pattern.Spec.Dependencies {
			depResult := i.installDependency(ctx, dep, opts)
			result.Dependencies = append(result.Dependencies, depResult)
			if depResult.Status == "failed" && !dep.Optional {
				result.Success = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("required dependency '%s' failed to install", dep.Name))
				return result, fmt.Errorf("dependency installation failed")
			}
		}
	}

	// Merge config with defaults
	config := pattern.MergeConfigWithDefaults(opts.Config)

	// Validate config
	if configErr := pattern.ValidateConfig(config); configErr != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("config validation failed: %v", configErr))
		return result, configErr
	}

	// Determine target environments
	environments := opts.Environments
	if len(environments) == 0 {
		environments = []string{"dev"} // Default
	}

	// Dry run check
	if opts.DryRun {
		result.Message = "Dry run - no changes made"
		paths, planErr := i.planGeneration(pattern, config, environments)
		if planErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("planning error: %v", planErr))
		}
		result.GeneratedPath = paths
		return result, nil
	}

	// Generate pattern files
	generatedPaths, err := i.generatePattern(pattern, config, environments)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to generate pattern: %v", err))
		return result, err
	}
	result.GeneratedPath = generatedPaths

	// Record installation
	installedPattern := &InstalledPattern{
		Pattern:      *pattern,
		InstalledAt:  time.Now(),
		Config:       config,
		Environments: environments,
		Status:       "installed",
		Paths:        generatedPaths,
	}
	i.installed[patternName] = installedPattern

	// Save state
	if err := i.SaveState(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to save state: %v", err))
	}

	result.Success = true
	result.Message = fmt.Sprintf("Pattern '%s' version %s installed successfully", patternName, version)

	return result, nil
}

// installDependency installs a single dependency.
func (i *Installer) installDependency(ctx context.Context, dep Dependency, opts InstallOptions) DependencyResult {
	result := DependencyResult{
		Name:     dep.Name,
		Version:  dep.Version,
		Optional: dep.Optional,
	}

	// Check if already installed
	if existing, ok := i.installed[dep.Name]; ok {
		result.Status = "skipped"
		result.Message = fmt.Sprintf("already installed (v%s)", existing.Pattern.Metadata.Version)
		return result
	}

	// Try to install
	depOpts := InstallOptions{
		Version:     dep.Version,
		Config:      make(map[string]any),
		DryRun:      opts.DryRun,
		AutoApprove: opts.AutoApprove,
		SkipDeps:    false, // Install transitive deps
	}

	installResult, err := i.Install(ctx, dep.Name, depOpts)
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		return result
	}

	result.Status = "installed"
	result.Version = installResult.Version
	result.Message = "installed successfully"
	return result
}

// planGeneration returns the paths that would be generated.
func (i *Installer) planGeneration(pattern *Pattern, config map[string]any, environments []string) ([]string, error) {
	var paths []string

	basePath := filepath.Join(i.projectPath, "infrastructure", pattern.Metadata.Category, pattern.Metadata.Name)

	for _, comp := range pattern.Spec.Components {
		switch comp.Type {
		case ComponentTypeHelm:
			paths = append(paths,
				filepath.Join(basePath, "base", "helm-release.yaml"),
				filepath.Join(basePath, "base", "values.yaml"),
			)
		case ComponentTypeKustomize:
			paths = append(paths, filepath.Join(basePath, "base", "kustomization.yaml"))
		case ComponentTypeManifest:
			paths = append(paths, filepath.Join(basePath, "base", comp.Name+".yaml"))
		}
	}

	// Overlay paths for each environment
	for _, env := range environments {
		paths = append(paths, filepath.Join(basePath, "overlays", env, "kustomization.yaml"))
	}

	// ArgoCD application
	argoCDPath := filepath.Join(i.projectPath, i.gitOpsTool, "applications", pattern.Metadata.Name+".yaml")
	paths = append(paths, argoCDPath)

	return paths, nil
}

// generatePattern generates the pattern files.
func (i *Installer) generatePattern(pattern *Pattern, config map[string]any, environments []string) ([]string, error) {
	var generatedPaths []string

	basePath := filepath.Join(i.projectPath, "infrastructure", pattern.Metadata.Category, pattern.Metadata.Name)

	// Create base directory
	baseDir := filepath.Join(basePath, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Generate component files
	for idx := range pattern.Spec.Components {
		paths, err := i.generateComponent(baseDir, pattern, &pattern.Spec.Components[idx], config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate component '%s': %w", pattern.Spec.Components[idx].Name, err)
		}
		generatedPaths = append(generatedPaths, paths...)
	}

	// Generate base kustomization
	kustomizePath := filepath.Join(baseDir, "kustomization.yaml")
	if err := i.generateBaseKustomization(kustomizePath, pattern); err != nil {
		return nil, err
	}
	generatedPaths = append(generatedPaths, kustomizePath)

	// Generate overlays for each environment
	for _, env := range environments {
		overlayDir := filepath.Join(basePath, "overlays", env)
		if err := os.MkdirAll(overlayDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create overlay directory: %w", err)
		}

		overlayPath := filepath.Join(overlayDir, "kustomization.yaml")
		if err := i.generateOverlayKustomization(overlayPath, env, config); err != nil {
			return nil, err
		}
		generatedPaths = append(generatedPaths, overlayPath)
	}

	// Generate ArgoCD application
	argoCDPaths, err := i.generateArgoCDApplication(pattern, config, environments)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ArgoCD application: %w", err)
	}
	generatedPaths = append(generatedPaths, argoCDPaths...)

	return generatedPaths, nil
}

// generateComponent generates files for a single component.
func (i *Installer) generateComponent(baseDir string, pattern *Pattern, comp *Component, config map[string]any) ([]string, error) {
	var paths []string

	switch comp.Type {
	case ComponentTypeHelm:
		// Generate HelmRelease
		helmRelease := map[string]any{
			"apiVersion": "source.toolkit.fluxcd.io/v1beta2",
			"kind":       "HelmRepository",
			"metadata": map[string]any{
				"name": comp.Name,
			},
			"spec": map[string]any{
				"interval": "1h",
				"url":      comp.Repository,
			},
		}

		data, err := yaml.Marshal(helmRelease)
		if err != nil {
			return nil, err
		}

		helmRepoPath := filepath.Join(baseDir, comp.Name+"-repo.yaml")
		if writeErr := os.WriteFile(helmRepoPath, data, 0644); writeErr != nil {
			return nil, writeErr
		}
		paths = append(paths, helmRepoPath)

		// Generate HelmRelease
		release := map[string]any{
			"apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
			"kind":       "HelmRelease",
			"metadata": map[string]any{
				"name": comp.Name,
			},
			"spec": map[string]any{
				"interval": "5m",
				"chart": map[string]any{
					"spec": map[string]any{
						"chart":   comp.Chart,
						"version": comp.Version,
						"sourceRef": map[string]any{
							"kind": "HelmRepository",
							"name": comp.Name,
						},
					},
				},
				"values": mergeValues(comp.Values, config),
			},
		}

		releaseData, err := yaml.Marshal(release)
		if err != nil {
			return nil, err
		}

		releasePath := filepath.Join(baseDir, comp.Name+"-release.yaml")
		if err := os.WriteFile(releasePath, releaseData, 0644); err != nil {
			return nil, err
		}
		paths = append(paths, releasePath)

	case ComponentTypeKustomize:
		// Generate kustomization reference
		kustomization := map[string]any{
			"apiVersion": "kustomize.toolkit.fluxcd.io/v1",
			"kind":       "Kustomization",
			"metadata": map[string]any{
				"name": comp.Name,
			},
			"spec": map[string]any{
				"interval": "5m",
				"path":     comp.Path,
				"prune":    true,
			},
		}

		data, err := yaml.Marshal(kustomization)
		if err != nil {
			return nil, err
		}

		kustomizePath := filepath.Join(baseDir, comp.Name+"-kustomization.yaml")
		if err := os.WriteFile(kustomizePath, data, 0644); err != nil {
			return nil, err
		}
		paths = append(paths, kustomizePath)

	case ComponentTypeManifest:
		// Copy or generate manifest
		manifestPath := filepath.Join(baseDir, comp.Name+".yaml")
		// For now, generate a placeholder
		manifest := fmt.Sprintf("# Manifest for %s\n# TODO: Add actual manifest content\n", comp.Name)
		if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
			return nil, err
		}
		paths = append(paths, manifestPath)
	}

	return paths, nil
}

// generateBaseKustomization generates the base kustomization.yaml.
func (i *Installer) generateBaseKustomization(path string, pattern *Pattern) error {
	var resources []string
	for _, comp := range pattern.Spec.Components {
		switch comp.Type {
		case ComponentTypeHelm:
			resources = append(resources, comp.Name+"-repo.yaml", comp.Name+"-release.yaml")
		case ComponentTypeKustomize:
			resources = append(resources, comp.Name+"-kustomization.yaml")
		case ComponentTypeManifest:
			resources = append(resources, comp.Name+".yaml")
		}
	}

	kustomization := map[string]any{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"resources":  resources,
	}

	data, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// generateOverlayKustomization generates an overlay kustomization.yaml.
func (i *Installer) generateOverlayKustomization(path, env string, config map[string]any) error {
	kustomization := map[string]any{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"resources": []string{
			"../../base",
		},
		"commonLabels": map[string]string{
			"environment": env,
		},
	}

	data, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// generateArgoCDApplication generates ArgoCD Application resources.
func (i *Installer) generateArgoCDApplication(pattern *Pattern, config map[string]any, environments []string) ([]string, error) {
	var paths []string

	appDir := filepath.Join(i.projectPath, i.gitOpsTool, "applications")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	for _, env := range environments {
		appName := fmt.Sprintf("%s-%s", pattern.Metadata.Name, env)
		app := map[string]any{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]any{
				"name": appName,
			},
			"spec": map[string]any{
				"project": "default",
				"source": map[string]any{
					"repoURL":        "{{ .RepoURL }}",
					"targetRevision": "HEAD",
					"path":           fmt.Sprintf("infrastructure/%s/%s/overlays/%s", pattern.Metadata.Category, pattern.Metadata.Name, env),
				},
				"destination": map[string]any{
					"server":    "https://kubernetes.default.svc",
					"namespace": pattern.Metadata.Name,
				},
				"syncPolicy": map[string]any{
					"automated": map[string]any{
						"prune":    true,
						"selfHeal": true,
					},
				},
			},
		}

		data, err := yaml.Marshal(app)
		if err != nil {
			return nil, err
		}

		appPath := filepath.Join(appDir, appName+".yaml")
		if err := os.WriteFile(appPath, data, 0644); err != nil {
			return nil, err
		}
		paths = append(paths, appPath)
	}

	return paths, nil
}

// Uninstall removes an installed pattern.
func (i *Installer) Uninstall(ctx context.Context, patternName string, opts UninstallOptions) error {
	if err := i.LoadState(); err != nil {
		return err
	}

	installed, ok := i.installed[patternName]
	if !ok {
		return fmt.Errorf("pattern '%s' is not installed", patternName)
	}

	// Remove generated files
	if !opts.KeepFiles {
		for _, path := range installed.Paths {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				if !opts.Force {
					return fmt.Errorf("failed to remove %s: %w", path, err)
				}
			}
		}
	}

	// Remove from state
	delete(i.installed, patternName)

	return i.SaveState()
}

// UninstallOptions defines options for pattern uninstallation.
type UninstallOptions struct {
	Force     bool
	KeepFiles bool
}

// Update updates an installed pattern to a new version.
func (i *Installer) Update(ctx context.Context, patternName string, opts UpdateOptions) (*InstallResult, error) {
	if err := i.LoadState(); err != nil {
		return nil, err
	}

	installed, ok := i.installed[patternName]
	if !ok {
		return nil, fmt.Errorf("pattern '%s' is not installed", patternName)
	}

	// Get target version
	targetVersion := opts.Version
	if targetVersion == "" {
		// Get latest version
		entry, _, err := i.registry.FindPattern(ctx, patternName)
		if err != nil {
			return nil, err
		}
		targetVersion = entry.Latest
	}

	// Check if update is needed
	if installed.Pattern.Metadata.Version == targetVersion && !opts.Force {
		return &InstallResult{
			Pattern: patternName,
			Version: targetVersion,
			Success: true,
			Message: "Already at target version",
		}, nil
	}

	// Install new version with existing config
	installOpts := InstallOptions{
		Version:      targetVersion,
		Config:       installed.Config,
		Environments: installed.Environments,
		Force:        true,
	}

	return i.Install(ctx, patternName, installOpts)
}

// UpdateOptions defines options for pattern update.
type UpdateOptions struct {
	Version string
	Force   bool
}

// ListInstalled returns all installed patterns.
func (i *Installer) ListInstalled() ([]InstalledPattern, error) {
	if err := i.LoadState(); err != nil {
		return nil, err
	}

	var patterns []InstalledPattern
	for _, p := range i.installed {
		patterns = append(patterns, *p)
	}

	return patterns, nil
}

// GetInstalled returns information about an installed pattern.
func (i *Installer) GetInstalled(name string) (*InstalledPattern, error) {
	if err := i.LoadState(); err != nil {
		return nil, err
	}

	installed, ok := i.installed[name]
	if !ok {
		return nil, fmt.Errorf("pattern '%s' is not installed", name)
	}

	return installed, nil
}

// GetStatus checks the status of installed patterns.
func (i *Installer) GetStatus(ctx context.Context) (map[string]string, error) {
	if err := i.LoadState(); err != nil {
		return nil, err
	}

	status := make(map[string]string)
	for name, installed := range i.installed {
		// Check if files still exist
		allExist := true
		for _, path := range installed.Paths {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				allExist = false
				break
			}
		}

		if allExist {
			status[name] = "healthy"
		} else {
			status[name] = "degraded"
		}
	}

	return status, nil
}

// mergeValues merges component values with user config.
func mergeValues(compValues, config map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy component values
	for k, v := range compValues {
		result[k] = v
	}

	// Override with config
	for k, v := range config {
		result[k] = v
	}

	return result
}

// ScaffoldPattern creates a new pattern scaffold.
func ScaffoldPattern(name, category, outputDir string) (*Pattern, error) {
	pattern := NewPattern(name, "0.1.0", fmt.Sprintf("A GitOps pattern for %s", name))
	pattern.Metadata.Category = category
	pattern.Metadata.Author = "your-name"
	pattern.Metadata.Tags = []string{category}

	// Add sample component
	pattern.Spec.Components = []Component{
		{
			Name:       name,
			Type:       ComponentTypeHelm,
			Chart:      "example/chart",
			Version:    "1.0.0",
			Repository: "https://charts.example.com",
		},
	}

	// Add sample config
	pattern.Spec.Config = map[string]ConfigItem{
		"replicas": {
			Type:        ConfigTypeInteger,
			Default:     1,
			Description: "Number of replicas",
		},
		"enabled": {
			Type:        ConfigTypeBoolean,
			Default:     true,
			Description: "Enable the component",
		},
	}

	// Add sample validation
	pattern.Spec.Validation = []ValidationCheck{
		{
			Name:    "deployment-ready",
			Check:   fmt.Sprintf("deployment/%s ready", name),
			Timeout: "5m",
		},
	}

	// Create output directory
	patternDir := filepath.Join(outputDir, name)
	if err := os.MkdirAll(patternDir, 0755); err != nil {
		return nil, err
	}

	// Save pattern.yaml
	patternPath := filepath.Join(patternDir, "pattern.yaml")
	if err := pattern.Save(patternPath); err != nil {
		return nil, err
	}

	// Create README template
	readmeContent := `# {{ .Name }}

{{ .Description }}

## Installation

` + "```bash" + `
gitopsi install {{ .Name }}
` + "```" + `

## Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
{{range $key, $item := .Config}}| {{ $key }} | {{ $item.Type }} | {{ $item.Default }} | {{ $item.Description }} |
{{end}}

## Components

{{range .Components}}
- **{{ .Name }}**: {{ .Type }} ({{ .Version }})
{{end}}
`

	tmpl, err := template.New("readme").Parse(readmeContent)
	if err != nil {
		return nil, err
	}

	readmePath := filepath.Join(patternDir, "README.md")
	readmeFile, err := os.Create(readmePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := readmeFile.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close readme file: %v\n", cerr)
		}
	}()

	data := struct {
		Name        string
		Description string
		Config      map[string]ConfigItem
		Components  []Component
	}{
		Name:        pattern.Metadata.Name,
		Description: pattern.Metadata.Description,
		Config:      pattern.Spec.Config,
		Components:  pattern.Spec.Components,
	}

	if err := tmpl.Execute(readmeFile, data); err != nil {
		return nil, err
	}

	return pattern, nil
}

// ValidatePattern validates a pattern directory.
func ValidatePattern(patternDir string) ([]string, error) {
	var errors []string

	// Check pattern.yaml exists
	patternPath := filepath.Join(patternDir, "pattern.yaml")
	if _, err := os.Stat(patternPath); os.IsNotExist(err) {
		errors = append(errors, "pattern.yaml is missing")
		return errors, fmt.Errorf("pattern.yaml not found")
	}

	// Load and validate pattern
	pattern, err := LoadPattern(patternPath)
	if err != nil {
		errors = append(errors, err.Error())
		return errors, err
	}

	// Check README exists
	readmePath := filepath.Join(patternDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		errors = append(errors, "README.md is recommended but missing")
	}

	// Validate components have valid types
	validTypes := map[ComponentType]bool{
		ComponentTypeHelm:      true,
		ComponentTypeKustomize: true,
		ComponentTypeManifest:  true,
		ComponentTypeOperator:  true,
	}
	for _, comp := range pattern.Spec.Components {
		if !validTypes[comp.Type] {
			errors = append(errors, fmt.Sprintf("component '%s' has invalid type '%s'", comp.Name, comp.Type))
		}
	}

	return errors, nil
}

// GenerateIndex generates an index file for a pattern directory.
func GenerateIndex(patternsDir, outputPath string) error {
	var entries []PatternIndexEntry

	// Walk patterns directory
	err := filepath.Walk(patternsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "pattern.yaml" {
			pattern, err := LoadPattern(path)
			if err != nil {
				return nil // Skip invalid patterns
			}

			entry := PatternIndexEntry{
				Name:        pattern.Metadata.Name,
				Description: pattern.Metadata.Description,
				Category:    pattern.Metadata.Category,
				Tags:        pattern.Metadata.Tags,
				Versions:    []string{pattern.Metadata.Version},
				Latest:      pattern.Metadata.Version,
				Author:      pattern.Metadata.Author,
			}
			entries = append(entries, entry)
		}

		return nil
	})
	if err != nil {
		return err
	}

	index := RegistryIndex{
		Version:   "1.0",
		Generated: time.Now(),
		Patterns:  entries,
	}

	// Build categories from patterns
	categoryCount := make(map[string]int)
	for _, entry := range entries {
		if entry.Category != "" {
			categoryCount[entry.Category]++
		}
	}

	for cat, count := range categoryCount {
		index.Categories = append(index.Categories, CategoryIndexEntry{
			Name:        cat,
			Description: CategoryDescription(PatternCategory(cat)),
			Count:       count,
		})
	}

	data, err := yaml.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// CheckUpdates checks for available updates for installed patterns.
func (i *Installer) CheckUpdates(ctx context.Context) (map[string]string, error) {
	if err := i.LoadState(); err != nil {
		return nil, err
	}

	updates := make(map[string]string)

	for name, installed := range i.installed {
		entry, _, err := i.registry.FindPattern(ctx, name)
		if err != nil {
			continue
		}

		if entry.Latest != installed.Pattern.Metadata.Version {
			updates[name] = entry.Latest
		}
	}

	return updates, nil
}

// GetDependencyTree returns the dependency tree for a pattern.
func (i *Installer) GetDependencyTree(ctx context.Context, patternName string) (map[string][]string, error) {
	tree := make(map[string][]string)

	entry, registryName, err := i.registry.FindPattern(ctx, patternName)
	if err != nil {
		return nil, err
	}

	pattern, err := i.registry.FetchPattern(ctx, registryName, patternName, entry.Latest)
	if err != nil {
		return nil, err
	}

	return i.buildDepTree(ctx, pattern, tree)
}

// buildDepTree recursively builds the dependency tree.
func (i *Installer) buildDepTree(ctx context.Context, pattern *Pattern, tree map[string][]string) (map[string][]string, error) {
	var deps []string
	for _, dep := range pattern.Spec.Dependencies {
		deps = append(deps, dep.Name)

		// Get transitive dependencies
		entry, registryName, err := i.registry.FindPattern(ctx, dep.Name)
		if err != nil {
			continue
		}

		depPattern, err := i.registry.FetchPattern(ctx, registryName, dep.Name, entry.Latest)
		if err != nil {
			continue
		}

		subTree, _ := i.buildDepTree(ctx, depPattern, tree)
		for k, v := range subTree {
			tree[k] = v
		}
	}

	tree[pattern.Metadata.Name] = deps
	return tree, nil
}

// ConflictCheck checks for conflicts with existing patterns.
func (i *Installer) ConflictCheck(ctx context.Context, patternName string) ([]string, error) {
	var conflicts []string

	entry, registryName, err := i.registry.FindPattern(ctx, patternName)
	if err != nil {
		return nil, err
	}

	pattern, err := i.registry.FetchPattern(ctx, registryName, patternName, entry.Latest)
	if err != nil {
		return nil, err
	}

	if err := i.LoadState(); err != nil {
		return nil, err
	}

	// Check for namespace conflicts
	for _, comp := range pattern.Spec.Components {
		if comp.Namespace != "" {
			for installedName, installed := range i.installed {
				for _, installedComp := range installed.Pattern.Spec.Components {
					if installedComp.Namespace == comp.Namespace {
						conflicts = append(conflicts,
							fmt.Sprintf("namespace conflict with '%s': both use namespace '%s'",
								installedName, comp.Namespace))
					}
				}
			}
		}
	}

	// Check for component name conflicts
	for _, comp := range pattern.Spec.Components {
		for installedName, installed := range i.installed {
			for _, installedComp := range installed.Pattern.Spec.Components {
				if installedComp.Name == comp.Name && installedName != patternName {
					conflicts = append(conflicts,
						fmt.Sprintf("component name conflict with '%s': both have component '%s'",
							installedName, comp.Name))
				}
			}
		}
	}

	return conflicts, nil
}

// GetInstallPath returns the path where pattern files are installed.
func GetInstallPath(projectPath, category, patternName string) string {
	if category == "" {
		category = "other"
	}
	return filepath.Join(projectPath, "infrastructure", strings.ToLower(category), patternName)
}
