package marketplace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPattern(t *testing.T) {
	pattern := NewPattern("test-pattern", "1.0.0", "A test pattern")

	if pattern.APIVersion != "gitopsi.io/v1" {
		t.Errorf("Expected APIVersion 'gitopsi.io/v1', got '%s'", pattern.APIVersion)
	}
	if pattern.Kind != "Pattern" {
		t.Errorf("Expected Kind 'Pattern', got '%s'", pattern.Kind)
	}
	if pattern.Metadata.Name != "test-pattern" {
		t.Errorf("Expected Name 'test-pattern', got '%s'", pattern.Metadata.Name)
	}
	if pattern.Metadata.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", pattern.Metadata.Version)
	}
	if pattern.Metadata.License != "MIT" {
		t.Errorf("Expected License 'MIT', got '%s'", pattern.Metadata.License)
	}
}

func TestPatternValidate(t *testing.T) {
	tests := []struct {
		name      string
		pattern   Pattern
		wantError bool
	}{
		{
			name: "valid pattern",
			pattern: Pattern{
				APIVersion: "gitopsi.io/v1",
				Kind:       "Pattern",
				Metadata: PatternMetadata{
					Name:        "test",
					Version:     "1.0.0",
					Description: "Test pattern",
				},
			},
			wantError: false,
		},
		{
			name: "missing apiVersion",
			pattern: Pattern{
				Kind: "Pattern",
				Metadata: PatternMetadata{
					Name:        "test",
					Version:     "1.0.0",
					Description: "Test pattern",
				},
			},
			wantError: true,
		},
		{
			name: "wrong kind",
			pattern: Pattern{
				APIVersion: "gitopsi.io/v1",
				Kind:       "Wrong",
				Metadata: PatternMetadata{
					Name:        "test",
					Version:     "1.0.0",
					Description: "Test pattern",
				},
			},
			wantError: true,
		},
		{
			name: "missing name",
			pattern: Pattern{
				APIVersion: "gitopsi.io/v1",
				Kind:       "Pattern",
				Metadata: PatternMetadata{
					Version:     "1.0.0",
					Description: "Test pattern",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pattern.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestPatternGetFullName(t *testing.T) {
	pattern := NewPattern("my-pattern", "2.1.0", "Test")
	expected := "my-pattern@2.1.0"
	if got := pattern.GetFullName(); got != expected {
		t.Errorf("GetFullName() = %s, want %s", got, expected)
	}
}

func TestPatternHasDependency(t *testing.T) {
	pattern := NewPattern("test", "1.0.0", "Test")
	pattern.Spec.Dependencies = []Dependency{
		{Name: "dep1", Version: "1.0.0"},
		{Name: "dep2", Version: "2.0.0", Optional: true},
	}

	if !pattern.HasDependency("dep1") {
		t.Error("HasDependency() should return true for dep1")
	}
	if !pattern.HasDependency("dep2") {
		t.Error("HasDependency() should return true for dep2")
	}
	if pattern.HasDependency("dep3") {
		t.Error("HasDependency() should return false for dep3")
	}
}

func TestPatternIsCompatibleWithPlatform(t *testing.T) {
	tests := []struct {
		name       string
		platforms  []PlatformRequirement
		platform   string
		compatible bool
	}{
		{
			name:       "no restrictions",
			platforms:  nil,
			platform:   "kubernetes",
			compatible: true,
		},
		{
			name: "kubernetes compatible",
			platforms: []PlatformRequirement{
				{Name: "kubernetes"},
				{Name: "openshift"},
			},
			platform:   "kubernetes",
			compatible: true,
		},
		{
			name: "openshift compatible",
			platforms: []PlatformRequirement{
				{Name: "kubernetes"},
				{Name: "openshift"},
			},
			platform:   "OpenShift",
			compatible: true,
		},
		{
			name: "not compatible",
			platforms: []PlatformRequirement{
				{Name: "kubernetes"},
			},
			platform:   "openshift",
			compatible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewPattern("test", "1.0.0", "Test")
			pattern.Spec.Platforms = tt.platforms

			if got := pattern.IsCompatibleWithPlatform(tt.platform); got != tt.compatible {
				t.Errorf("IsCompatibleWithPlatform() = %v, want %v", got, tt.compatible)
			}
		})
	}
}

func TestPatternIsCompatibleWithTool(t *testing.T) {
	pattern := NewPattern("test", "1.0.0", "Test")
	pattern.Spec.GitOpsTools = []ToolRequirement{
		{Name: "argocd", MinVersion: "2.8"},
		{Name: "flux", MinVersion: "2.0"},
	}

	if !pattern.IsCompatibleWithTool("argocd") {
		t.Error("Should be compatible with argocd")
	}
	if !pattern.IsCompatibleWithTool("ArgoCD") {
		t.Error("Should be compatible with ArgoCD (case insensitive)")
	}
	if pattern.IsCompatibleWithTool("jenkins") {
		t.Error("Should not be compatible with jenkins")
	}
}

func TestPatternValidateConfig(t *testing.T) {
	pattern := NewPattern("test", "1.0.0", "Test")
	pattern.Spec.Config = map[string]ConfigItem{
		"replicas": {
			Type:     ConfigTypeInteger,
			Required: true,
		},
		"enabled": {
			Type:    ConfigTypeBoolean,
			Default: true,
		},
		"name": {
			Type: ConfigTypeString,
		},
	}

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
	}{
		{
			name:      "valid config",
			config:    map[string]any{"replicas": 3, "enabled": true},
			wantError: false,
		},
		{
			name:      "missing required",
			config:    map[string]any{"enabled": true},
			wantError: true,
		},
		{
			name:      "wrong type",
			config:    map[string]any{"replicas": "three"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pattern.ValidateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestPatternMergeConfigWithDefaults(t *testing.T) {
	pattern := NewPattern("test", "1.0.0", "Test")
	pattern.Spec.Config = map[string]ConfigItem{
		"replicas": {Type: ConfigTypeInteger, Default: 1},
		"enabled":  {Type: ConfigTypeBoolean, Default: true},
		"name":     {Type: ConfigTypeString, Default: "default"},
	}

	config := map[string]any{
		"replicas": 3,
	}

	merged := pattern.MergeConfigWithDefaults(config)

	if merged["replicas"] != 3 {
		t.Errorf("Expected replicas=3, got %v", merged["replicas"])
	}
	if merged["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", merged["enabled"])
	}
	if merged["name"] != "default" {
		t.Errorf("Expected name='default', got %v", merged["name"])
	}
}

func TestPatternSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	patternPath := filepath.Join(tmpDir, "pattern.yaml")

	// Create and save pattern
	pattern := NewPattern("test-pattern", "1.0.0", "A test pattern")
	pattern.Metadata.Category = "observability"
	pattern.Metadata.Tags = []string{"monitoring", "test"}
	pattern.Spec.Components = []Component{
		{Name: "prometheus", Type: ComponentTypeHelm, Chart: "prometheus-community/prometheus"},
	}

	if err := pattern.Save(patternPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load pattern
	loaded, err := LoadPattern(patternPath)
	if err != nil {
		t.Fatalf("LoadPattern() error = %v", err)
	}

	if loaded.Metadata.Name != pattern.Metadata.Name {
		t.Errorf("Loaded name = %s, want %s", loaded.Metadata.Name, pattern.Metadata.Name)
	}
	if loaded.Metadata.Category != pattern.Metadata.Category {
		t.Errorf("Loaded category = %s, want %s", loaded.Metadata.Category, pattern.Metadata.Category)
	}
	if len(loaded.Spec.Components) != 1 {
		t.Errorf("Loaded components count = %d, want 1", len(loaded.Spec.Components))
	}
}

func TestGetCategories(t *testing.T) {
	categories := GetCategories()
	if len(categories) != 8 {
		t.Errorf("Expected 8 categories, got %d", len(categories))
	}

	// Check all categories have descriptions
	for _, cat := range categories {
		desc := CategoryDescription(cat)
		if desc == "" {
			t.Errorf("Category %s has no description", cat)
		}
	}
}

func TestNewRegistryManager(t *testing.T) {
	rm := NewRegistryManager("/tmp/cache")

	registries := rm.ListRegistries()
	if len(registries) != 1 {
		t.Errorf("Expected 1 default registry, got %d", len(registries))
	}

	if registries[0].Name != "official" {
		t.Errorf("Expected 'official' registry, got '%s'", registries[0].Name)
	}
}

func TestRegistryManagerAddRemove(t *testing.T) {
	rm := NewRegistryManager("/tmp/cache")

	// Add registry
	reg := Registry{
		Name:     "custom",
		Type:     RegistryTypePrivate,
		URL:      "https://my-registry.com",
		Priority: 50,
		Enabled:  true,
	}

	if err := rm.AddRegistry(reg); err != nil {
		t.Fatalf("AddRegistry() error = %v", err)
	}

	if len(rm.ListRegistries()) != 2 {
		t.Errorf("Expected 2 registries after add")
	}

	// Try to add duplicate
	if err := rm.AddRegistry(reg); err == nil {
		t.Error("Expected error for duplicate registry")
	}

	// Remove registry
	if err := rm.RemoveRegistry("custom"); err != nil {
		t.Fatalf("RemoveRegistry() error = %v", err)
	}

	if len(rm.ListRegistries()) != 1 {
		t.Errorf("Expected 1 registry after remove")
	}

	// Try to remove non-existent
	if err := rm.RemoveRegistry("nonexistent"); err == nil {
		t.Error("Expected error for non-existent registry")
	}
}

func TestSearchOptions(t *testing.T) {
	opts := SearchOptions{
		Category: "monitoring",
		Tags:     []string{"prometheus"},
		Platform: "kubernetes",
		Tool:     "argocd",
		Limit:    10,
	}

	if opts.Category != "monitoring" {
		t.Errorf("Category = %s, want 'monitoring'", opts.Category)
	}
	if opts.Limit != 10 {
		t.Errorf("Limit = %d, want 10", opts.Limit)
	}
}

func TestMatchesSearch(t *testing.T) {
	entry := PatternIndexEntry{
		Name:        "prometheus-stack",
		Description: "Complete monitoring solution",
		Category:    "observability",
		Tags:        []string{"monitoring", "prometheus", "grafana"},
	}

	tests := []struct {
		name    string
		query   string
		opts    SearchOptions
		matches bool
	}{
		{
			name:    "name match",
			query:   "prometheus",
			opts:    SearchOptions{},
			matches: true,
		},
		{
			name:    "description match",
			query:   "monitoring",
			opts:    SearchOptions{},
			matches: true,
		},
		{
			name:    "tag match",
			query:   "grafana",
			opts:    SearchOptions{},
			matches: true,
		},
		{
			name:    "category filter match",
			query:   "",
			opts:    SearchOptions{Category: "observability"},
			matches: true,
		},
		{
			name:    "category filter no match",
			query:   "",
			opts:    SearchOptions{Category: "security"},
			matches: false,
		},
		{
			name:    "no match",
			query:   "kubernetes",
			opts:    SearchOptions{},
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesSearch(entry, tt.query, tt.opts); got != tt.matches {
				t.Errorf("matchesSearch() = %v, want %v", got, tt.matches)
			}
		})
	}
}

func TestInstallerStateManagement(t *testing.T) {
	tmpDir := t.TempDir()
	rm := NewRegistryManager(filepath.Join(tmpDir, "cache"))
	installer := NewInstaller(rm, tmpDir, "argocd", "kubernetes")

	// Initially no patterns installed
	patterns, err := installer.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled() error = %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("Expected 0 installed patterns, got %d", len(patterns))
	}

	// Manually add an installed pattern
	installer.installed["test-pattern"] = &InstalledPattern{
		Pattern: Pattern{
			APIVersion: "gitopsi.io/v1",
			Kind:       "Pattern",
			Metadata: PatternMetadata{
				Name:    "test-pattern",
				Version: "1.0.0",
			},
		},
		InstalledAt: time.Now(),
		Status:      "installed",
	}

	// Save state
	if err := installer.SaveState(); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	// Create new installer and load state
	installer2 := NewInstaller(rm, tmpDir, "argocd", "kubernetes")
	if err := installer2.LoadState(); err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	patterns, err = installer2.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled() error = %v", err)
	}
	if len(patterns) != 1 {
		t.Errorf("Expected 1 installed pattern, got %d", len(patterns))
	}
}

func TestScaffoldPattern(t *testing.T) {
	tmpDir := t.TempDir()

	pattern, err := ScaffoldPattern("my-new-pattern", "observability", tmpDir)
	if err != nil {
		t.Fatalf("ScaffoldPattern() error = %v", err)
	}

	if pattern.Metadata.Name != "my-new-pattern" {
		t.Errorf("Name = %s, want 'my-new-pattern'", pattern.Metadata.Name)
	}
	if pattern.Metadata.Category != "observability" {
		t.Errorf("Category = %s, want 'observability'", pattern.Metadata.Category)
	}

	// Check files were created
	patternPath := filepath.Join(tmpDir, "my-new-pattern", "pattern.yaml")
	if _, err := os.Stat(patternPath); os.IsNotExist(err) {
		t.Error("pattern.yaml was not created")
	}

	readmePath := filepath.Join(tmpDir, "my-new-pattern", "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md was not created")
	}
}

func TestValidatePattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid pattern
	_, err := ScaffoldPattern("valid-pattern", "security", tmpDir)
	if err != nil {
		t.Fatalf("ScaffoldPattern() error = %v", err)
	}

	errors, err := ValidatePattern(filepath.Join(tmpDir, "valid-pattern"))
	if err != nil {
		t.Fatalf("ValidatePattern() error = %v", err)
	}
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Test non-existent directory
	_, err = ValidatePattern(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestNewMarketplace(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)

	if mp.projectPath != tmpDir {
		t.Errorf("projectPath = %s, want %s", mp.projectPath, tmpDir)
	}

	if mp.registry == nil {
		t.Error("registry should not be nil")
	}

	registries := mp.ListRegistries()
	if len(registries) != 1 {
		t.Errorf("Expected 1 registry, got %d", len(registries))
	}
}

func TestMarketplaceConfigure(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)

	// Before configure, installer should be nil
	if mp.installer != nil {
		t.Error("installer should be nil before Configure()")
	}

	mp.Configure("argocd", "kubernetes")

	if mp.installer == nil {
		t.Error("installer should not be nil after Configure()")
	}
}

func TestMarketplaceListInstalled_NotConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)

	_, err := mp.ListInstalled()
	if err == nil {
		t.Error("Expected error when not configured")
	}
}

func TestMarketplaceCreatePattern(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)

	pattern, err := mp.CreatePattern("test-pattern", "security")
	if err != nil {
		t.Fatalf("CreatePattern() error = %v", err)
	}

	if pattern.Metadata.Name != "test-pattern" {
		t.Errorf("Name = %s, want 'test-pattern'", pattern.Metadata.Name)
	}
}

func TestMarketplaceValidatePattern(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)

	// Create a pattern to validate
	_, err := mp.CreatePattern("to-validate", "networking")
	if err != nil {
		t.Fatalf("CreatePattern() error = %v", err)
	}

	errors, err := mp.ValidatePattern(filepath.Join(tmpDir, "to-validate"))
	if err != nil {
		t.Fatalf("ValidatePattern() error = %v", err)
	}
	if len(errors) > 0 {
		t.Logf("Validation warnings: %v", errors)
	}
}

func TestMarketplaceGetMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)
	mp.Configure("argocd", "kubernetes")

	metrics, err := mp.GetMetrics()
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}

	if metrics.InstalledCount != 0 {
		t.Errorf("InstalledCount = %d, want 0", metrics.InstalledCount)
	}
	if metrics.RegistriesCount != 1 {
		t.Errorf("RegistriesCount = %d, want 1", metrics.RegistriesCount)
	}
}

func TestGetOfficialPatterns(t *testing.T) {
	patterns := GetOfficialPatterns()

	if len(patterns) < 10 {
		t.Errorf("Expected at least 10 official patterns, got %d", len(patterns))
	}

	// Check prometheus-stack exists
	found := false
	for _, p := range patterns {
		if p.Name == "prometheus-stack" {
			found = true
			if !p.Verified {
				t.Error("prometheus-stack should be verified")
			}
			break
		}
	}
	if !found {
		t.Error("prometheus-stack should be in official patterns")
	}
}

func TestCategoryInfo(t *testing.T) {
	tests := []struct {
		category   string
		expectIcon bool
		expectDesc bool
	}{
		{"observability", true, true},
		{"security", true, true},
		{"networking", true, true},
		{"unknown", true, true}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			icon, desc := CategoryInfo(tt.category)
			if tt.expectIcon && icon == "" {
				t.Error("Expected icon")
			}
			if tt.expectDesc && desc == "" {
				t.Error("Expected description")
			}
		})
	}
}

func TestInstallOptionsDefaults(t *testing.T) {
	opts := InstallOptions{}

	if opts.DryRun {
		t.Error("DryRun should be false by default")
	}
	if opts.Force {
		t.Error("Force should be false by default")
	}
	if opts.SkipDeps {
		t.Error("SkipDeps should be false by default")
	}
}

func TestInstallResult(t *testing.T) {
	result := InstallResult{
		Pattern: "test",
		Version: "1.0.0",
		Success: true,
		Message: "Installed successfully",
	}

	if result.Pattern != "test" {
		t.Errorf("Pattern = %s, want 'test'", result.Pattern)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestInstalledPattern(t *testing.T) {
	installed := InstalledPattern{
		Pattern: Pattern{
			Metadata: PatternMetadata{
				Name:    "test",
				Version: "1.0.0",
			},
		},
		InstalledAt:  time.Now(),
		Status:       "installed",
		Environments: []string{"dev", "staging"},
		Config:       map[string]any{"replicas": 3},
	}

	if installed.Pattern.Metadata.Name != "test" {
		t.Errorf("Pattern name = %s, want 'test'", installed.Pattern.Metadata.Name)
	}
	if len(installed.Environments) != 2 {
		t.Errorf("Environments count = %d, want 2", len(installed.Environments))
	}
}

func TestPatternSuggestion(t *testing.T) {
	suggestion := PatternSuggestion{
		Pattern:  "prometheus-stack",
		Reason:   "No monitoring detected",
		Category: "observability",
		Priority: 1,
	}

	if suggestion.Pattern != "prometheus-stack" {
		t.Errorf("Pattern = %s, want 'prometheus-stack'", suggestion.Pattern)
	}
	if suggestion.Priority != 1 {
		t.Errorf("Priority = %d, want 1", suggestion.Priority)
	}
}

func TestConfigTypeConstants(t *testing.T) {
	tests := []struct {
		configType ConfigType
		expected   string
	}{
		{ConfigTypeString, "string"},
		{ConfigTypeInteger, "integer"},
		{ConfigTypeBoolean, "boolean"},
		{ConfigTypeSecret, "secret"},
		{ConfigTypeArray, "array"},
		{ConfigTypeObject, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.configType) != tt.expected {
				t.Errorf("ConfigType = %s, want %s", tt.configType, tt.expected)
			}
		})
	}
}

func TestComponentTypeConstants(t *testing.T) {
	tests := []struct {
		compType ComponentType
		expected string
	}{
		{ComponentTypeHelm, "helm"},
		{ComponentTypeKustomize, "kustomize"},
		{ComponentTypeManifest, "manifest"},
		{ComponentTypeOperator, "operator"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.compType) != tt.expected {
				t.Errorf("ComponentType = %s, want %s", tt.compType, tt.expected)
			}
		})
	}
}

func TestRegistryTypeConstants(t *testing.T) {
	tests := []struct {
		regType  RegistryType
		expected string
	}{
		{RegistryTypeOfficial, "official"},
		{RegistryTypeCommunity, "community"},
		{RegistryTypePrivate, "private"},
		{RegistryTypeLocal, "local"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.regType) != tt.expected {
				t.Errorf("RegistryType = %s, want %s", tt.regType, tt.expected)
			}
		})
	}
}

func TestGetInstallPath(t *testing.T) {
	tests := []struct {
		projectPath string
		category    string
		patternName string
		expected    string
	}{
		{"/project", "observability", "prometheus", "/project/infrastructure/observability/prometheus"},
		{"/project", "", "test", "/project/infrastructure/other/test"},
		{"/project", "Security", "vault", "/project/infrastructure/security/vault"},
	}

	for _, tt := range tests {
		t.Run(tt.patternName, func(t *testing.T) {
			got := GetInstallPath(tt.projectPath, tt.category, tt.patternName)
			if got != tt.expected {
				t.Errorf("GetInstallPath() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestMarketplaceSearch_NoRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)
	mp.Configure("argocd", "kubernetes")

	// Search will try to fetch from registry which doesn't exist
	// Should not panic and return empty results
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results, err := mp.Search(ctx, "prometheus", SearchOptions{})
	// May error due to network, but should not panic
	if err != nil {
		t.Logf("Search returned error (expected for no network): %v", err)
	}
	_ = results // May be nil or empty
}

func TestMarketplaceToYAML(t *testing.T) {
	tmpDir := t.TempDir()
	mp := NewMarketplace(tmpDir)

	yaml, err := mp.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if yaml == "" {
		t.Error("Expected non-empty YAML output")
	}

	if !containsString(yaml, "projectPath") {
		t.Error("YAML should contain projectPath")
	}
}

func TestPatternToYAML(t *testing.T) {
	pattern := NewPattern("test", "1.0.0", "Test pattern")
	pattern.Metadata.Category = "security"

	yaml, err := pattern.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if !containsString(yaml, "name: test") {
		t.Error("YAML should contain pattern name")
	}
	if !containsString(yaml, "category: security") {
		t.Error("YAML should contain category")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}
