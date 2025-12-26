package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RegistryType represents the type of pattern registry.
type RegistryType string

const (
	RegistryTypeOfficial  RegistryType = "official"
	RegistryTypeCommunity RegistryType = "community"
	RegistryTypePrivate   RegistryType = "private"
	RegistryTypeLocal     RegistryType = "local"
)

// Registry represents a pattern registry configuration.
type Registry struct {
	Name     string        `yaml:"name" json:"name"`
	Type     RegistryType  `yaml:"type" json:"type"`
	URL      string        `yaml:"url" json:"url"`
	Priority int           `yaml:"priority,omitempty" json:"priority,omitempty"`
	Auth     *RegistryAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
	Enabled  bool          `yaml:"enabled" json:"enabled"`
}

// RegistryAuth contains authentication for private registries.
type RegistryAuth struct {
	Type     string `yaml:"type" json:"type"` // token, basic, ssh
	Token    string `yaml:"token,omitempty" json:"token,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	SSHKey   string `yaml:"sshKey,omitempty" json:"sshKey,omitempty"`
}

// RegistryIndex represents the index of patterns in a registry.
type RegistryIndex struct {
	Version    string               `yaml:"version" json:"version"`
	Generated  time.Time            `yaml:"generated" json:"generated"`
	Patterns   []PatternIndexEntry  `yaml:"patterns" json:"patterns"`
	Categories []CategoryIndexEntry `yaml:"categories,omitempty" json:"categories,omitempty"`
}

// PatternIndexEntry represents a pattern entry in the registry index.
type PatternIndexEntry struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Category    string   `yaml:"category" json:"category"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Versions    []string `yaml:"versions" json:"versions"`
	Latest      string   `yaml:"latest" json:"latest"`
	Author      string   `yaml:"author,omitempty" json:"author,omitempty"`
	Rating      float64  `yaml:"rating,omitempty" json:"rating,omitempty"`
	Downloads   int      `yaml:"downloads,omitempty" json:"downloads,omitempty"`
	Verified    bool     `yaml:"verified,omitempty" json:"verified,omitempty"`
	Deprecated  bool     `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
}

// CategoryIndexEntry represents a category in the registry.
type CategoryIndexEntry struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Icon        string `yaml:"icon,omitempty" json:"icon,omitempty"`
	Count       int    `yaml:"count" json:"count"`
}

// RegistryManager manages multiple pattern registries.
type RegistryManager struct {
	registries []Registry
	cacheDir   string
	httpClient *http.Client
}

// NewRegistryManager creates a new registry manager.
func NewRegistryManager(cacheDir string) *RegistryManager {
	return &RegistryManager{
		registries: []Registry{
			{
				Name:     "official",
				Type:     RegistryTypeOfficial,
				URL:      "https://raw.githubusercontent.com/gitopsi/patterns/main",
				Priority: 100,
				Enabled:  true,
			},
		},
		cacheDir: cacheDir,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AddRegistry adds a custom registry.
func (rm *RegistryManager) AddRegistry(reg Registry) error {
	// Validate registry
	if reg.Name == "" {
		return fmt.Errorf("registry name is required")
	}
	if reg.URL == "" && reg.Type != RegistryTypeLocal {
		return fmt.Errorf("registry URL is required")
	}

	// Check for duplicates
	for _, r := range rm.registries {
		if r.Name == reg.Name {
			return fmt.Errorf("registry '%s' already exists", reg.Name)
		}
	}

	rm.registries = append(rm.registries, reg)

	// Sort by priority
	sort.Slice(rm.registries, func(i, j int) bool {
		return rm.registries[i].Priority > rm.registries[j].Priority
	})

	return nil
}

// RemoveRegistry removes a registry by name.
func (rm *RegistryManager) RemoveRegistry(name string) error {
	for i, r := range rm.registries {
		if r.Name == name {
			rm.registries = append(rm.registries[:i], rm.registries[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("registry '%s' not found", name)
}

// GetRegistry returns a registry by name.
func (rm *RegistryManager) GetRegistry(name string) (*Registry, error) {
	for _, r := range rm.registries {
		if r.Name == name {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("registry '%s' not found", name)
}

// ListRegistries returns all configured registries.
func (rm *RegistryManager) ListRegistries() []Registry {
	return rm.registries
}

// FetchIndex fetches the pattern index from a registry.
func (rm *RegistryManager) FetchIndex(ctx context.Context, registryName string) (*RegistryIndex, error) {
	reg, err := rm.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	if !reg.Enabled {
		return nil, fmt.Errorf("registry '%s' is disabled", registryName)
	}

	switch reg.Type {
	case RegistryTypeLocal:
		return rm.fetchLocalIndex(reg)
	default:
		return rm.fetchRemoteIndex(ctx, reg)
	}
}

// fetchLocalIndex fetches index from a local directory.
func (rm *RegistryManager) fetchLocalIndex(reg *Registry) (*RegistryIndex, error) {
	indexPath := filepath.Join(reg.URL, "index.yaml")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local index: %w", err)
	}

	var index RegistryIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &index, nil
}

// fetchRemoteIndex fetches index from a remote URL.
func (rm *RegistryManager) fetchRemoteIndex(ctx context.Context, reg *Registry) (*RegistryIndex, error) {
	indexURL := fmt.Sprintf("%s/index.yaml", strings.TrimSuffix(reg.URL, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if configured
	if reg.Auth != nil {
		switch reg.Auth.Type {
		case "token":
			req.Header.Set("Authorization", "Bearer "+reg.Auth.Token)
		case "basic":
			req.SetBasicAuth(reg.Auth.Username, reg.Auth.Password)
		}
	}

	resp, err := rm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch index: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var index RegistryIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	// Cache the index
	if err := rm.cacheIndex(reg.Name, &index); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to cache index: %v\n", err)
	}

	return &index, nil
}

// cacheIndex caches a registry index locally.
func (rm *RegistryManager) cacheIndex(registryName string, index *RegistryIndex) error {
	if rm.cacheDir == "" {
		return nil
	}

	cacheDir := filepath.Join(rm.cacheDir, "registries", registryName)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, "index.yaml"), data, 0644)
}

// GetCachedIndex returns a cached registry index.
func (rm *RegistryManager) GetCachedIndex(registryName string) (*RegistryIndex, error) {
	if rm.cacheDir == "" {
		return nil, fmt.Errorf("cache not configured")
	}

	cachePath := filepath.Join(rm.cacheDir, "registries", registryName, "index.yaml")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("cached index not found: %w", err)
	}

	var index RegistryIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse cached index: %w", err)
	}

	return &index, nil
}

// FetchPattern fetches a specific pattern from a registry.
func (rm *RegistryManager) FetchPattern(ctx context.Context, registryName, patternName, version string) (*Pattern, error) {
	reg, err := rm.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	switch reg.Type {
	case RegistryTypeLocal:
		return rm.fetchLocalPattern(reg, patternName, version)
	default:
		return rm.fetchRemotePattern(ctx, reg, patternName, version)
	}
}

// fetchLocalPattern fetches a pattern from a local directory.
func (rm *RegistryManager) fetchLocalPattern(reg *Registry, name, version string) (*Pattern, error) {
	patternPath := filepath.Join(reg.URL, "patterns", name, version, "pattern.yaml")
	return LoadPattern(patternPath)
}

// fetchRemotePattern fetches a pattern from a remote URL.
func (rm *RegistryManager) fetchRemotePattern(ctx context.Context, reg *Registry, name, version string) (*Pattern, error) {
	patternURL := fmt.Sprintf("%s/patterns/%s/%s/pattern.yaml",
		strings.TrimSuffix(reg.URL, "/"), name, version)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, patternURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if reg.Auth != nil && reg.Auth.Type == "token" {
		req.Header.Set("Authorization", "Bearer "+reg.Auth.Token)
	}

	resp, err := rm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pattern: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pattern not found: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
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

// SearchPatterns searches for patterns across all enabled registries.
func (rm *RegistryManager) SearchPatterns(ctx context.Context, query string, opts SearchOptions) ([]PatternSearchResult, error) {
	var results []PatternSearchResult
	seenPatterns := make(map[string]bool)

	query = strings.ToLower(query)

	for _, reg := range rm.registries {
		if !reg.Enabled {
			continue
		}

		index, err := rm.FetchIndex(ctx, reg.Name)
		if err != nil {
			// Try cached index
			index, err = rm.GetCachedIndex(reg.Name)
			if err != nil {
				continue
			}
		}

		for _, entry := range index.Patterns {
			if seenPatterns[entry.Name] {
				continue
			}

			if matchesSearch(entry, query, opts) {
				results = append(results, PatternSearchResult{
					Name:        entry.Name,
					Version:     entry.Latest,
					Description: entry.Description,
					Category:    entry.Category,
					Tags:        entry.Tags,
					Rating:      entry.Rating,
					Downloads:   entry.Downloads,
				})
				seenPatterns[entry.Name] = true
			}
		}
	}

	// Sort results by relevance
	sortSearchResults(results, query)

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// SearchOptions defines search parameters.
type SearchOptions struct {
	Category string
	Tags     []string
	Platform string
	Tool     string
	Limit    int
}

// matchesSearch checks if a pattern entry matches the search criteria.
func matchesSearch(entry PatternIndexEntry, query string, opts SearchOptions) bool {
	// Filter by category
	if opts.Category != "" && !strings.EqualFold(entry.Category, opts.Category) {
		return false
	}

	// Filter by tags
	if len(opts.Tags) > 0 {
		hasTag := false
		for _, tag := range opts.Tags {
			for _, entryTag := range entry.Tags {
				if strings.EqualFold(tag, entryTag) {
					hasTag = true
					break
				}
			}
		}
		if !hasTag {
			return false
		}
	}

	// Empty query matches all (after filters)
	if query == "" {
		return true
	}

	// Match query against name, description, and tags
	if strings.Contains(strings.ToLower(entry.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(entry.Description), query) {
		return true
	}
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}

// sortSearchResults sorts results by relevance to the query.
func sortSearchResults(results []PatternSearchResult, query string) {
	sort.Slice(results, func(i, j int) bool {
		// Exact name match first
		iExact := strings.EqualFold(results[i].Name, query)
		jExact := strings.EqualFold(results[j].Name, query)
		if iExact != jExact {
			return iExact
		}

		// Name starts with query
		iPrefix := strings.HasPrefix(strings.ToLower(results[i].Name), query)
		jPrefix := strings.HasPrefix(strings.ToLower(results[j].Name), query)
		if iPrefix != jPrefix {
			return iPrefix
		}

		// Higher rating
		if results[i].Rating != results[j].Rating {
			return results[i].Rating > results[j].Rating
		}

		// More downloads
		return results[i].Downloads > results[j].Downloads
	})
}

// GetPatternVersions returns all available versions of a pattern.
func (rm *RegistryManager) GetPatternVersions(ctx context.Context, patternName string) ([]PatternVersion, error) {
	for _, reg := range rm.registries {
		if !reg.Enabled {
			continue
		}

		index, err := rm.FetchIndex(ctx, reg.Name)
		if err != nil {
			continue
		}

		for _, entry := range index.Patterns {
			if entry.Name == patternName {
				var versions []PatternVersion
				for _, v := range entry.Versions {
					versions = append(versions, PatternVersion{
						Version: v,
					})
				}
				return versions, nil
			}
		}
	}

	return nil, fmt.Errorf("pattern '%s' not found", patternName)
}

// FindPattern finds a pattern by name across all registries.
func (rm *RegistryManager) FindPattern(ctx context.Context, name string) (*PatternIndexEntry, string, error) {
	for _, reg := range rm.registries {
		if !reg.Enabled {
			continue
		}

		index, err := rm.FetchIndex(ctx, reg.Name)
		if err != nil {
			continue
		}

		for _, entry := range index.Patterns {
			if entry.Name == name {
				return &entry, reg.Name, nil
			}
		}
	}

	return nil, "", fmt.Errorf("pattern '%s' not found in any registry", name)
}

// GetCategories returns all categories from all registries.
func (rm *RegistryManager) GetCategories(ctx context.Context) ([]CategoryIndexEntry, error) {
	categoryMap := make(map[string]*CategoryIndexEntry)

	for _, reg := range rm.registries {
		if !reg.Enabled {
			continue
		}

		index, err := rm.FetchIndex(ctx, reg.Name)
		if err != nil {
			continue
		}

		for _, cat := range index.Categories {
			if existing, ok := categoryMap[cat.Name]; ok {
				existing.Count += cat.Count
			} else {
				entry := cat
				categoryMap[cat.Name] = &entry
			}
		}

		// Also count from patterns if categories not defined
		if len(index.Categories) == 0 {
			for _, pattern := range index.Patterns {
				if pattern.Category != "" {
					if _, ok := categoryMap[pattern.Category]; !ok {
						categoryMap[pattern.Category] = &CategoryIndexEntry{
							Name:        pattern.Category,
							Description: CategoryDescription(PatternCategory(pattern.Category)),
						}
					}
					categoryMap[pattern.Category].Count++
				}
			}
		}
	}

	var categories []CategoryIndexEntry
	for _, cat := range categoryMap {
		categories = append(categories, *cat)
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories, nil
}

// ToJSON converts the registry manager state to JSON.
func (rm *RegistryManager) ToJSON() (string, error) {
	data, err := json.MarshalIndent(rm.registries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
