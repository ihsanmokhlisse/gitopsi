package marketplace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Marketplace provides the main interface for pattern marketplace operations.
type Marketplace struct {
	registry    *RegistryManager
	installer   *Installer
	projectPath string
	cacheDir    string
}

// NewMarketplace creates a new marketplace instance.
func NewMarketplace(projectPath string) *Marketplace {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	cacheDir := filepath.Join(homeDir, ".gitopsi", "cache")

	registry := NewRegistryManager(cacheDir)

	return &Marketplace{
		registry:    registry,
		projectPath: projectPath,
		cacheDir:    cacheDir,
	}
}

// Configure configures the marketplace with GitOps settings.
func (m *Marketplace) Configure(gitOpsTool, platform string) {
	m.installer = NewInstaller(m.registry, m.projectPath, gitOpsTool, platform)
}

// GetRegistry returns the registry manager.
func (m *Marketplace) GetRegistry() *RegistryManager {
	return m.registry
}

// GetInstaller returns the installer.
func (m *Marketplace) GetInstaller() *Installer {
	return m.installer
}

// Search searches for patterns in the marketplace.
func (m *Marketplace) Search(ctx context.Context, query string, opts SearchOptions) ([]PatternSearchResult, error) {
	results, err := m.registry.SearchPatterns(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	// Mark installed patterns
	if m.installer != nil {
		if err := m.installer.LoadState(); err == nil {
			for i, result := range results {
				if _, ok := m.installer.installed[result.Name]; ok {
					results[i].Installed = true
				}
			}
		}
	}

	return results, nil
}

// ListCategories returns all available categories.
func (m *Marketplace) ListCategories(ctx context.Context) ([]CategoryIndexEntry, error) {
	return m.registry.GetCategories(ctx)
}

// GetPatternInfo returns detailed information about a pattern.
func (m *Marketplace) GetPatternInfo(ctx context.Context, name string) (*PatternInfo, error) {
	entry, registryName, err := m.registry.FindPattern(ctx, name)
	if err != nil {
		return nil, err
	}

	// Fetch full pattern for latest version
	pattern, err := m.registry.FetchPattern(ctx, registryName, name, entry.Latest)
	if err != nil {
		return nil, err
	}

	info := &PatternInfo{
		Pattern:   *pattern,
		Registry:  registryName,
		Versions:  entry.Versions,
		Rating:    entry.Rating,
		Downloads: entry.Downloads,
		Verified:  entry.Verified,
	}

	// Check if installed
	if m.installer != nil {
		if installed, err := m.installer.GetInstalled(name); err == nil {
			info.Installed = true
			info.InstalledVersion = installed.Pattern.Metadata.Version
			info.InstalledAt = installed.InstalledAt.String()
		}
	}

	return info, nil
}

// PatternInfo contains detailed pattern information.
type PatternInfo struct {
	Pattern          Pattern  `yaml:"pattern" json:"pattern"`
	Registry         string   `yaml:"registry" json:"registry"`
	Versions         []string `yaml:"versions" json:"versions"`
	Rating           float64  `yaml:"rating,omitempty" json:"rating,omitempty"`
	Downloads        int      `yaml:"downloads,omitempty" json:"downloads,omitempty"`
	Verified         bool     `yaml:"verified,omitempty" json:"verified,omitempty"`
	Installed        bool     `yaml:"installed" json:"installed"`
	InstalledVersion string   `yaml:"installedVersion,omitempty" json:"installedVersion,omitempty"`
	InstalledAt      string   `yaml:"installedAt,omitempty" json:"installedAt,omitempty"`
}

// Install installs a pattern.
func (m *Marketplace) Install(ctx context.Context, name string, opts InstallOptions) (*InstallResult, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.Install(ctx, name, opts)
}

// Uninstall removes a pattern.
func (m *Marketplace) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	if m.installer == nil {
		return fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.Uninstall(ctx, name, opts)
}

// Update updates a pattern to a newer version.
func (m *Marketplace) Update(ctx context.Context, name string, opts UpdateOptions) (*InstallResult, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.Update(ctx, name, opts)
}

// ListInstalled returns all installed patterns.
func (m *Marketplace) ListInstalled() ([]InstalledPattern, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.ListInstalled()
}

// GetStatus returns the status of all installed patterns.
func (m *Marketplace) GetStatus(ctx context.Context) (map[string]string, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.GetStatus(ctx)
}

// CheckUpdates checks for available updates.
func (m *Marketplace) CheckUpdates(ctx context.Context) (map[string]string, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.CheckUpdates(ctx)
}

// AddRegistry adds a custom registry.
func (m *Marketplace) AddRegistry(reg Registry) error {
	return m.registry.AddRegistry(reg)
}

// RemoveRegistry removes a registry.
func (m *Marketplace) RemoveRegistry(name string) error {
	return m.registry.RemoveRegistry(name)
}

// ListRegistries returns all configured registries.
func (m *Marketplace) ListRegistries() []Registry {
	return m.registry.ListRegistries()
}

// CreatePattern scaffolds a new pattern.
func (m *Marketplace) CreatePattern(name, category string) (*Pattern, error) {
	outputDir := m.projectPath
	if outputDir == "" {
		outputDir = "."
	}
	return ScaffoldPattern(name, category, outputDir)
}

// ValidatePattern validates a pattern directory.
func (m *Marketplace) ValidatePattern(patternDir string) ([]string, error) {
	return ValidatePattern(patternDir)
}

// PublishPattern publishes a pattern to a registry.
func (m *Marketplace) PublishPattern(ctx context.Context, patternDir, registryName string) error {
	// Validate first
	errors, err := ValidatePattern(patternDir)
	if err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("validation warnings: %v", errors)
	}

	// For now, only local registries support publishing
	reg, err := m.registry.GetRegistry(registryName)
	if err != nil {
		return err
	}

	if reg.Type != RegistryTypeLocal {
		return fmt.Errorf("publishing to remote registries is not yet supported")
	}

	// Copy pattern to registry
	pattern, err := LoadPattern(filepath.Join(patternDir, "pattern.yaml"))
	if err != nil {
		return err
	}

	destDir := filepath.Join(reg.URL, "patterns", pattern.Metadata.Name, pattern.Metadata.Version)
	if mkdirErr := os.MkdirAll(destDir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create destination directory: %w", mkdirErr)
	}

	// Copy all files
	err = filepath.Walk(patternDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, relErr := filepath.Rel(patternDir, path)
		if relErr != nil {
			return relErr
		}
		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		return os.WriteFile(destPath, data, 0644)
	})
	if err != nil {
		return fmt.Errorf("failed to copy pattern: %w", err)
	}

	// Regenerate index
	patternsDir := filepath.Join(reg.URL, "patterns")
	indexPath := filepath.Join(reg.URL, "index.yaml")
	return GenerateIndex(patternsDir, indexPath)
}

// GetDependencies returns the dependencies of a pattern.
func (m *Marketplace) GetDependencies(ctx context.Context, name string) ([]Dependency, error) {
	info, err := m.GetPatternInfo(ctx, name)
	if err != nil {
		return nil, err
	}
	return info.Pattern.Spec.Dependencies, nil
}

// GetDependencyTree returns the full dependency tree.
func (m *Marketplace) GetDependencyTree(ctx context.Context, name string) (map[string][]string, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.GetDependencyTree(ctx, name)
}

// ConflictCheck checks for conflicts with existing patterns.
func (m *Marketplace) ConflictCheck(ctx context.Context, name string) ([]string, error) {
	if m.installer == nil {
		return nil, fmt.Errorf("marketplace not configured, call Configure() first")
	}
	return m.installer.ConflictCheck(ctx, name)
}

// GetPopularPatterns returns the most popular patterns.
func (m *Marketplace) GetPopularPatterns(ctx context.Context, limit int) ([]PatternSearchResult, error) {
	results, err := m.Search(ctx, "", SearchOptions{Limit: 100})
	if err != nil {
		return nil, err
	}

	// Sort by downloads
	sort.Slice(results, func(i, j int) bool {
		return results[i].Downloads > results[j].Downloads
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetRecommendedPatterns returns recommended patterns based on current setup.
func (m *Marketplace) GetRecommendedPatterns(ctx context.Context, limit int) ([]PatternSearchResult, error) {
	var recommended []PatternSearchResult

	installed, err := m.ListInstalled()
	if err != nil {
		installed = nil
	}
	installedNames := make(map[string]bool)
	for idx := range installed {
		installedNames[installed[idx].Pattern.Metadata.Name] = true
	}

	// Get all patterns
	results, err := m.Search(ctx, "", SearchOptions{})
	if err != nil {
		return nil, err
	}

	// Recommend based on:
	// 1. Not already installed
	// 2. High rating
	// 3. Related to installed patterns (same category)
	installedCategories := make(map[string]bool)
	for idx := range installed {
		installedCategories[installed[idx].Pattern.Metadata.Category] = true
	}

	for _, result := range results {
		if installedNames[result.Name] {
			continue
		}

		// Prioritize same category
		if installedCategories[result.Category] {
			recommended = append(recommended, result)
		}
	}

	// Sort by rating
	sort.Slice(recommended, func(i, j int) bool {
		return recommended[i].Rating > recommended[j].Rating
	})

	if limit > 0 && len(recommended) > limit {
		recommended = recommended[:limit]
	}

	return recommended, nil
}

// ExportConfig exports installed patterns to a config file.
func (m *Marketplace) ExportConfig(outputPath string) error {
	installed, err := m.ListInstalled()
	if err != nil {
		return err
	}

	type PatternConfig struct {
		Name    string         `yaml:"name"`
		Version string         `yaml:"version"`
		Config  map[string]any `yaml:"config,omitempty"`
	}

	patterns := make([]PatternConfig, 0, len(installed))
	for idx := range installed {
		patterns = append(patterns, PatternConfig{
			Name:    installed[idx].Pattern.Metadata.Name,
			Version: installed[idx].Pattern.Metadata.Version,
			Config:  installed[idx].Config,
		})
	}

	config := struct {
		Patterns []PatternConfig `yaml:"patterns"`
	}{
		Patterns: patterns,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// ImportConfig imports patterns from a config file.
func (m *Marketplace) ImportConfig(ctx context.Context, configPath string) ([]InstallResult, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config struct {
		Patterns []struct {
			Name         string         `yaml:"name"`
			Version      string         `yaml:"version,omitempty"`
			Config       map[string]any `yaml:"config,omitempty"`
			Environments []string       `yaml:"environments,omitempty"`
		} `yaml:"patterns"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	var results []InstallResult
	for _, p := range config.Patterns {
		opts := InstallOptions{
			Version:      p.Version,
			Config:       p.Config,
			Environments: p.Environments,
		}

		result, err := m.Install(ctx, p.Name, opts)
		if err != nil {
			results = append(results, InstallResult{
				Pattern: p.Name,
				Success: false,
				Errors:  []string{err.Error()},
			})
		} else {
			results = append(results, *result)
		}
	}

	return results, nil
}

// GetRecommendedPatternsForCategory returns recommended patterns for a category.
func (m *Marketplace) GetRecommendedPatternsForCategory(ctx context.Context, category string, limit int) ([]PatternSearchResult, error) {
	results, err := m.Search(ctx, "", SearchOptions{
		Category: category,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	// Sort by rating
	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	return results, nil
}

// GetPatternsByTag returns patterns with a specific tag.
func (m *Marketplace) GetPatternsByTag(ctx context.Context, tag string) ([]PatternSearchResult, error) {
	return m.Search(ctx, "", SearchOptions{
		Tags: []string{tag},
	})
}

// SuggestPatterns suggests patterns based on project analysis.
func (m *Marketplace) SuggestPatterns(ctx context.Context) ([]PatternSuggestion, error) {
	var suggestions []PatternSuggestion

	// Analyze project structure
	hasMonitoring := false
	hasLogging := false
	hasIngress := false
	hasSecrets := false

	infraPath := filepath.Join(m.projectPath, "infrastructure")
	if _, err := os.Stat(infraPath); err == nil {
		// Check what's already configured
		if _, err := os.Stat(filepath.Join(infraPath, "monitoring")); err == nil {
			hasMonitoring = true
		}
		if _, err := os.Stat(filepath.Join(infraPath, "logging")); err == nil {
			hasLogging = true
		}
		if _, err := os.Stat(filepath.Join(infraPath, "ingress")); err == nil {
			hasIngress = true
		}
		if _, err := os.Stat(filepath.Join(infraPath, "secrets")); err == nil {
			hasSecrets = true
		}
	}

	// Suggest based on what's missing
	if !hasMonitoring {
		suggestions = append(suggestions, PatternSuggestion{
			Pattern:  "prometheus-stack",
			Reason:   "No monitoring stack detected",
			Category: string(CategoryObservability),
			Priority: 1,
		})
	}

	if !hasLogging {
		suggestions = append(suggestions, PatternSuggestion{
			Pattern:  "loki-stack",
			Reason:   "No logging solution detected",
			Category: string(CategoryObservability),
			Priority: 2,
		})
	}

	if !hasIngress {
		suggestions = append(suggestions, PatternSuggestion{
			Pattern:  "nginx-ingress",
			Reason:   "No ingress controller detected",
			Category: string(CategoryNetworking),
			Priority: 1,
		})
	}

	if !hasSecrets {
		suggestions = append(suggestions, PatternSuggestion{
			Pattern:  "sealed-secrets",
			Reason:   "No secrets management detected",
			Category: string(CategorySecurity),
			Priority: 1,
		})
	}

	// Sort by priority
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Priority < suggestions[j].Priority
	})

	return suggestions, nil
}

// PatternSuggestion represents a suggested pattern.
type PatternSuggestion struct {
	Pattern  string `yaml:"pattern" json:"pattern"`
	Reason   string `yaml:"reason" json:"reason"`
	Category string `yaml:"category" json:"category"`
	Priority int    `yaml:"priority" json:"priority"`
}

// GetMetrics returns marketplace usage metrics.
func (m *Marketplace) GetMetrics() (*Metrics, error) {
	installed, err := m.ListInstalled()
	if err != nil {
		installed = nil // Handle gracefully
	}

	metrics := &Metrics{
		InstalledCount:  len(installed),
		RegistriesCount: len(m.registry.registries),
		Categories:      make(map[string]int),
	}

	for idx := range installed {
		cat := installed[idx].Pattern.Metadata.Category
		if cat == "" {
			cat = "other"
		}
		metrics.Categories[cat]++
	}

	return metrics, nil
}

// Metrics contains marketplace usage metrics.
type Metrics struct {
	InstalledCount  int            `yaml:"installedCount" json:"installedCount"`
	RegistriesCount int            `yaml:"registriesCount" json:"registriesCount"`
	Categories      map[string]int `yaml:"categories" json:"categories"`
}

// ToYAML converts marketplace state to YAML.
func (m *Marketplace) ToYAML() (string, error) {
	state := struct {
		ProjectPath string     `yaml:"projectPath"`
		CacheDir    string     `yaml:"cacheDir"`
		Registries  []Registry `yaml:"registries"`
	}{
		ProjectPath: m.projectPath,
		CacheDir:    m.cacheDir,
		Registries:  m.registry.registries,
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetOfficialPatterns returns a list of official/curated patterns.
func GetOfficialPatterns() []PatternIndexEntry {
	return []PatternIndexEntry{
		{
			Name:        "prometheus-stack",
			Description: "Complete Prometheus + Grafana monitoring stack",
			Category:    string(CategoryObservability),
			Tags:        []string{"monitoring", "observability", "alerting", "prometheus", "grafana"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "loki-stack",
			Description: "Grafana Loki for log aggregation",
			Category:    string(CategoryObservability),
			Tags:        []string{"logging", "observability", "loki", "grafana"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "cert-manager",
			Description: "Automatic TLS certificate management",
			Category:    string(CategorySecurity),
			Tags:        []string{"certificates", "tls", "security", "letsencrypt"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "sealed-secrets",
			Description: "Bitnami Sealed Secrets for GitOps-safe secrets",
			Category:    string(CategorySecurity),
			Tags:        []string{"secrets", "encryption", "security"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "vault-integration",
			Description: "HashiCorp Vault integration with External Secrets",
			Category:    string(CategorySecurity),
			Tags:        []string{"secrets", "vault", "security"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "nginx-ingress",
			Description: "NGINX Ingress Controller",
			Category:    string(CategoryNetworking),
			Tags:        []string{"ingress", "networking", "nginx"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "istio-mesh",
			Description: "Istio Service Mesh with Kiali and Jaeger",
			Category:    string(CategoryNetworking),
			Tags:        []string{"service-mesh", "istio", "networking", "security"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "postgresql-operator",
			Description: "CloudNativePG PostgreSQL Operator",
			Category:    string(CategoryData),
			Tags:        []string{"database", "postgresql", "operator"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "redis-operator",
			Description: "Redis Operator for high availability Redis",
			Category:    string(CategoryData),
			Tags:        []string{"cache", "redis", "operator"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "kyverno-policies",
			Description: "Kyverno policy engine with best practice policies",
			Category:    string(CategorySecurity),
			Tags:        []string{"policies", "security", "compliance", "kyverno"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "tekton-pipelines",
			Description: "Tekton Pipelines for CI/CD",
			Category:    string(CategoryCICD),
			Tags:        []string{"ci", "cd", "pipelines", "tekton"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
		{
			Name:        "external-dns",
			Description: "External DNS for automatic DNS management",
			Category:    string(CategoryNetworking),
			Tags:        []string{"dns", "networking"},
			Latest:      "1.0.0",
			Versions:    []string{"1.0.0"},
			Verified:    true,
		},
	}
}

// CategoryInfo returns information about a category.
func CategoryInfo(category string) (icon, description string) {
	info := map[string]struct {
		icon        string
		description string
	}{
		string(CategoryInfrastructure): {"🏗️", "Networking, security, storage, and compute infrastructure"},
		string(CategoryObservability):  {"📊", "Monitoring, logging, tracing, and dashboards"},
		string(CategorySecurity):       {"🔐", "Secrets management, policies, scanning, and certificates"},
		string(CategoryNetworking):     {"🌐", "Ingress controllers, service mesh, and DNS"},
		string(CategoryData):           {"💾", "Databases, caching, messaging, and storage"},
		string(CategoryCICD):           {"🚀", "Pipelines, workflows, and testing"},
		string(CategoryPlatforms):      {"📱", "Platform-specific patterns (OpenShift, AWS, Azure, GCP)"},
		string(CategoryEnterprise):     {"🏢", "Multi-tenancy, compliance, and cost management"},
	}

	if cat, ok := info[strings.ToLower(category)]; ok {
		return cat.icon, cat.description
	}
	return "📦", "Other patterns"
}
