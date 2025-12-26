// Package marketplace provides GitOps pattern management and marketplace functionality.
package marketplace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// PatternCategory represents a category of patterns.
type PatternCategory string

const (
	CategoryInfrastructure PatternCategory = "infrastructure"
	CategoryObservability  PatternCategory = "observability"
	CategorySecurity       PatternCategory = "security"
	CategoryNetworking     PatternCategory = "networking"
	CategoryData           PatternCategory = "data"
	CategoryCICD           PatternCategory = "cicd"
	CategoryPlatforms      PatternCategory = "platforms"
	CategoryEnterprise     PatternCategory = "enterprise"
)

// ComponentType represents the type of component in a pattern.
type ComponentType string

const (
	ComponentTypeHelm      ComponentType = "helm"
	ComponentTypeKustomize ComponentType = "kustomize"
	ComponentTypeManifest  ComponentType = "manifest"
	ComponentTypeOperator  ComponentType = "operator"
)

// ConfigType represents the type of configuration value.
type ConfigType string

const (
	ConfigTypeString  ConfigType = "string"
	ConfigTypeInteger ConfigType = "integer"
	ConfigTypeBoolean ConfigType = "boolean"
	ConfigTypeSecret  ConfigType = "secret"
	ConfigTypeArray   ConfigType = "array"
	ConfigTypeObject  ConfigType = "object"
)

// Pattern represents a GitOps pattern definition.
type Pattern struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	Metadata   PatternMetadata `yaml:"metadata" json:"metadata"`
	Spec       PatternSpec     `yaml:"spec" json:"spec"`
}

// PatternMetadata contains pattern metadata.
type PatternMetadata struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Author      string   `yaml:"author" json:"author"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Repository  string   `yaml:"repository,omitempty" json:"repository,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Category    string   `yaml:"category,omitempty" json:"category,omitempty"`
	Icon        string   `yaml:"icon,omitempty" json:"icon,omitempty"`
}

// PatternSpec contains the pattern specification.
type PatternSpec struct {
	Platforms    []PlatformRequirement `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	GitOpsTools  []ToolRequirement     `yaml:"gitops_tools,omitempty" json:"gitops_tools,omitempty"`
	Dependencies []Dependency          `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Components   []Component           `yaml:"components,omitempty" json:"components,omitempty"`
	Config       map[string]ConfigItem `yaml:"config,omitempty" json:"config,omitempty"`
	Validation   []ValidationCheck     `yaml:"validation,omitempty" json:"validation,omitempty"`
	Docs         PatternDocs           `yaml:"docs,omitempty" json:"docs,omitempty"`
	Hooks        PatternHooks          `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// PlatformRequirement defines platform compatibility.
type PlatformRequirement struct {
	Name       string `yaml:"name" json:"name"`
	MinVersion string `yaml:"minVersion,omitempty" json:"minVersion,omitempty"`
	MaxVersion string `yaml:"maxVersion,omitempty" json:"maxVersion,omitempty"`
}

// ToolRequirement defines GitOps tool compatibility.
type ToolRequirement struct {
	Name       string `yaml:"name" json:"name"`
	MinVersion string `yaml:"minVersion,omitempty" json:"minVersion,omitempty"`
}

// Dependency defines a pattern dependency.
type Dependency struct {
	Name     string `yaml:"name" json:"name"`
	Version  string `yaml:"version,omitempty" json:"version,omitempty"`
	Optional bool   `yaml:"optional,omitempty" json:"optional,omitempty"`
	Reason   string `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// Component defines a component in the pattern.
type Component struct {
	Name       string            `yaml:"name" json:"name"`
	Type       ComponentType     `yaml:"type" json:"type"`
	Chart      string            `yaml:"chart,omitempty" json:"chart,omitempty"`
	Repository string            `yaml:"repository,omitempty" json:"repository,omitempty"`
	Version    string            `yaml:"version,omitempty" json:"version,omitempty"`
	Path       string            `yaml:"path,omitempty" json:"path,omitempty"`
	Values     map[string]any    `yaml:"values,omitempty" json:"values,omitempty"`
	Namespace  string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ConfigItem defines a configuration option.
type ConfigItem struct {
	Type        ConfigType `yaml:"type" json:"type"`
	Default     any        `yaml:"default,omitempty" json:"default,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool       `yaml:"required,omitempty" json:"required,omitempty"`
	Enum        []string   `yaml:"enum,omitempty" json:"enum,omitempty"`
	Min         *int       `yaml:"min,omitempty" json:"min,omitempty"`
	Max         *int       `yaml:"max,omitempty" json:"max,omitempty"`
	Pattern     string     `yaml:"pattern,omitempty" json:"pattern,omitempty"`
}

// ValidationCheck defines a post-install validation.
type ValidationCheck struct {
	Name    string `yaml:"name" json:"name"`
	Check   string `yaml:"check" json:"check"`
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// PatternDocs contains documentation references.
type PatternDocs struct {
	Readme          string `yaml:"readme,omitempty" json:"readme,omitempty"`
	Architecture    string `yaml:"architecture,omitempty" json:"architecture,omitempty"`
	Troubleshooting string `yaml:"troubleshooting,omitempty" json:"troubleshooting,omitempty"`
	Changelog       string `yaml:"changelog,omitempty" json:"changelog,omitempty"`
}

// PatternHooks defines lifecycle hooks.
type PatternHooks struct {
	PreInstall  string `yaml:"preInstall,omitempty" json:"preInstall,omitempty"`
	PostInstall string `yaml:"postInstall,omitempty" json:"postInstall,omitempty"`
	PreUpdate   string `yaml:"preUpdate,omitempty" json:"preUpdate,omitempty"`
	PostUpdate  string `yaml:"postUpdate,omitempty" json:"postUpdate,omitempty"`
	PreDelete   string `yaml:"preDelete,omitempty" json:"preDelete,omitempty"`
	PostDelete  string `yaml:"postDelete,omitempty" json:"postDelete,omitempty"`
}

// InstalledPattern represents an installed pattern with its configuration.
type InstalledPattern struct {
	Pattern      Pattern           `yaml:"pattern" json:"pattern"`
	InstalledAt  time.Time         `yaml:"installedAt" json:"installedAt"`
	UpdatedAt    time.Time         `yaml:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	Config       map[string]any    `yaml:"config,omitempty" json:"config,omitempty"`
	Environments []string          `yaml:"environments,omitempty" json:"environments,omitempty"`
	Status       string            `yaml:"status" json:"status"`
	Health       string            `yaml:"health,omitempty" json:"health,omitempty"`
	Paths        []string          `yaml:"paths,omitempty" json:"paths,omitempty"`
	Annotations  map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// PatternVersion represents a specific version of a pattern.
type PatternVersion struct {
	Version     string    `yaml:"version" json:"version"`
	ReleasedAt  time.Time `yaml:"releasedAt" json:"releasedAt"`
	Changelog   string    `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	Deprecated  bool      `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	MinUpgrade  string    `yaml:"minUpgrade,omitempty" json:"minUpgrade,omitempty"`
	BreakingAPI bool      `yaml:"breakingAPI,omitempty" json:"breakingAPI,omitempty"`
}

// PatternSearchResult represents a search result item.
type PatternSearchResult struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Category    string   `yaml:"category" json:"category"`
	Tags        []string `yaml:"tags" json:"tags"`
	Rating      float64  `yaml:"rating" json:"rating"`
	Downloads   int      `yaml:"downloads" json:"downloads"`
	Installed   bool     `yaml:"installed" json:"installed"`
}

// NewPattern creates a new pattern with default values.
func NewPattern(name, version, description string) *Pattern {
	return &Pattern{
		APIVersion: "gitopsi.io/v1",
		Kind:       "Pattern",
		Metadata: PatternMetadata{
			Name:        name,
			Version:     version,
			Description: description,
			License:     "MIT",
		},
		Spec: PatternSpec{
			Config: make(map[string]ConfigItem),
		},
	}
}

// LoadPattern loads a pattern from a YAML file.
func LoadPattern(path string) (*Pattern, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pattern file: %w", err)
	}

	var pattern Pattern
	if err := yaml.Unmarshal(data, &pattern); err != nil {
		return nil, fmt.Errorf("failed to parse pattern: %w", err)
	}

	if err := pattern.Validate(); err != nil {
		return nil, err
	}

	return &pattern, nil
}

// Save saves the pattern to a YAML file.
func (p *Pattern) Save(path string) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal pattern: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write pattern file: %w", err)
	}

	return nil
}

// Validate validates the pattern definition.
func (p *Pattern) Validate() error {
	var errors []string

	if p.APIVersion == "" {
		errors = append(errors, "apiVersion is required")
	}
	if p.Kind != "Pattern" {
		errors = append(errors, "kind must be 'Pattern'")
	}
	if p.Metadata.Name == "" {
		errors = append(errors, "metadata.name is required")
	}
	if p.Metadata.Version == "" {
		errors = append(errors, "metadata.version is required")
	}
	if p.Metadata.Description == "" {
		errors = append(errors, "metadata.description is required")
	}

	// Validate components
	for i, comp := range p.Spec.Components {
		if comp.Name == "" {
			errors = append(errors, fmt.Sprintf("components[%d].name is required", i))
		}
		if comp.Type == "" {
			errors = append(errors, fmt.Sprintf("components[%d].type is required", i))
		}
	}

	// Validate config items
	for key, item := range p.Spec.Config {
		if item.Type == "" {
			errors = append(errors, fmt.Sprintf("config.%s.type is required", key))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("pattern validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetFullName returns the full pattern name with version.
func (p *Pattern) GetFullName() string {
	return fmt.Sprintf("%s@%s", p.Metadata.Name, p.Metadata.Version)
}

// HasDependency checks if the pattern has a specific dependency.
func (p *Pattern) HasDependency(name string) bool {
	for _, dep := range p.Spec.Dependencies {
		if dep.Name == name {
			return true
		}
	}
	return false
}

// GetDependency returns a specific dependency.
func (p *Pattern) GetDependency(name string) *Dependency {
	for _, dep := range p.Spec.Dependencies {
		if dep.Name == name {
			return &dep
		}
	}
	return nil
}

// IsCompatibleWithPlatform checks if the pattern is compatible with a platform.
func (p *Pattern) IsCompatibleWithPlatform(platform string) bool {
	if len(p.Spec.Platforms) == 0 {
		return true // No restrictions
	}

	for _, req := range p.Spec.Platforms {
		if strings.EqualFold(req.Name, platform) {
			return true
		}
	}
	return false
}

// IsCompatibleWithTool checks if the pattern is compatible with a GitOps tool.
func (p *Pattern) IsCompatibleWithTool(tool string) bool {
	if len(p.Spec.GitOpsTools) == 0 {
		return true // No restrictions
	}

	for _, req := range p.Spec.GitOpsTools {
		if strings.EqualFold(req.Name, tool) {
			return true
		}
	}
	return false
}

// GetRequiredConfig returns a list of required configuration keys.
func (p *Pattern) GetRequiredConfig() []string {
	var required []string
	for key, item := range p.Spec.Config {
		if item.Required {
			required = append(required, key)
		}
	}
	return required
}

// ValidateConfig validates provided configuration against the pattern spec.
func (p *Pattern) ValidateConfig(config map[string]any) error {
	var errors []string

	// Check required fields
	for key, item := range p.Spec.Config {
		if item.Required {
			if _, ok := config[key]; !ok {
				errors = append(errors, fmt.Sprintf("required config '%s' is missing", key))
			}
		}
	}

	// Validate types and constraints
	for key, value := range config {
		item, ok := p.Spec.Config[key]
		if !ok {
			continue // Unknown config, allow it
		}

		if err := validateConfigValue(key, value, &item); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateConfigValue validates a single configuration value.
func validateConfigValue(key string, value any, item *ConfigItem) error {
	switch item.Type {
	case ConfigTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s must be a string", key)
		}
	case ConfigTypeInteger:
		switch v := value.(type) {
		case int, int32, int64, float64:
			// Check min/max constraints
			var intVal int
			switch vv := v.(type) {
			case int:
				intVal = vv
			case int32:
				intVal = int(vv)
			case int64:
				intVal = int(vv)
			case float64:
				intVal = int(vv)
			}
			if item.Min != nil && intVal < *item.Min {
				return fmt.Errorf("%s must be at least %d", key, *item.Min)
			}
			if item.Max != nil && intVal > *item.Max {
				return fmt.Errorf("%s must be at most %d", key, *item.Max)
			}
		default:
			return fmt.Errorf("%s must be an integer", key)
		}
	case ConfigTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s must be a boolean", key)
		}
	case ConfigTypeArray:
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("%s must be an array", key)
		}
	}

	// Check enum constraints
	if len(item.Enum) > 0 {
		strVal, ok := value.(string)
		if ok {
			found := false
			for _, e := range item.Enum {
				if e == strVal {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%s must be one of: %v", key, item.Enum)
			}
		}
	}

	return nil
}

// MergeConfigWithDefaults merges provided config with default values.
func (p *Pattern) MergeConfigWithDefaults(config map[string]any) map[string]any {
	result := make(map[string]any)

	// Apply defaults first
	for key, item := range p.Spec.Config {
		if item.Default != nil {
			result[key] = item.Default
		}
	}

	// Override with provided config
	for key, value := range config {
		result[key] = value
	}

	return result
}

// ToYAML converts the pattern to YAML.
func (p *Pattern) ToYAML() (string, error) {
	data, err := yaml.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetCategories returns all available pattern categories.
func GetCategories() []PatternCategory {
	return []PatternCategory{
		CategoryInfrastructure,
		CategoryObservability,
		CategorySecurity,
		CategoryNetworking,
		CategoryData,
		CategoryCICD,
		CategoryPlatforms,
		CategoryEnterprise,
	}
}

// CategoryDescription returns a description for a category.
func CategoryDescription(cat PatternCategory) string {
	descriptions := map[PatternCategory]string{
		CategoryInfrastructure: "Networking, security, storage, and compute infrastructure",
		CategoryObservability:  "Monitoring, logging, tracing, and dashboards",
		CategorySecurity:       "Secrets management, policies, scanning, and certificates",
		CategoryNetworking:     "Ingress controllers, service mesh, and DNS",
		CategoryData:           "Databases, caching, messaging, and storage",
		CategoryCICD:           "Pipelines, workflows, and testing",
		CategoryPlatforms:      "Platform-specific patterns (OpenShift, AWS, Azure, GCP)",
		CategoryEnterprise:     "Multi-tenancy, compliance, and cost management",
	}
	return descriptions[cat]
}
