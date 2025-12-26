// Package multirepo provides multi-repository support for GitOps.
package multirepo

import (
	"fmt"
	"strings"
)

// Repository represents a Git repository configuration.
type Repository struct {
	// Name is the unique identifier for this repository
	Name string `yaml:"name" json:"name"`
	// URL is the Git repository URL
	URL string `yaml:"url" json:"url"`
	// Branch is the default branch to use
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
	// Path is the base path within the repository
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Type indicates what this repository contains
	Type RepositoryType `yaml:"type" json:"type"`
	// Credentials holds authentication configuration
	Credentials *Credentials `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	// SyncPolicy defines how to sync from this repo
	SyncPolicy *SyncPolicy `yaml:"sync_policy,omitempty" json:"sync_policy,omitempty"`
}

// RepositoryType defines the type of content in a repository.
type RepositoryType string

const (
	// RepoTypeMonorepo contains everything in one repository
	RepoTypeMonorepo RepositoryType = "monorepo"
	// RepoTypeApplications contains only application manifests
	RepoTypeApplications RepositoryType = "applications"
	// RepoTypeInfrastructure contains only infrastructure manifests
	RepoTypeInfrastructure RepositoryType = "infrastructure"
	// RepoTypeConfig contains cluster/environment configuration
	RepoTypeConfig RepositoryType = "config"
	// RepoTypeHelmCharts contains Helm charts
	RepoTypeHelmCharts RepositoryType = "helm-charts"
	// RepoTypeKustomize contains Kustomize bases
	RepoTypeKustomize RepositoryType = "kustomize"
)

// Credentials holds repository authentication configuration.
type Credentials struct {
	// Type is the credential type: ssh, token, basic
	Type string `yaml:"type" json:"type"`
	// SecretRef references a Kubernetes Secret
	SecretRef *SecretRef `yaml:"secret_ref,omitempty" json:"secret_ref,omitempty"`
	// SSHPrivateKeyPath is the path to SSH private key (for local use)
	SSHPrivateKeyPath string `yaml:"ssh_private_key_path,omitempty" json:"ssh_private_key_path,omitempty"`
	// Username for basic auth
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// TokenEnv is the environment variable containing the token
	TokenEnv string `yaml:"token_env,omitempty" json:"token_env,omitempty"`
}

// SecretRef references a Kubernetes Secret.
type SecretRef struct {
	// Name of the Secret
	Name string `yaml:"name" json:"name"`
	// Namespace of the Secret
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// Key in the Secret data
	Key string `yaml:"key,omitempty" json:"key,omitempty"`
}

// SyncPolicy defines how ArgoCD should sync from this repository.
type SyncPolicy struct {
	// Automated enables automatic sync
	Automated bool `yaml:"automated,omitempty" json:"automated,omitempty"`
	// Prune enables automatic pruning
	Prune bool `yaml:"prune,omitempty" json:"prune,omitempty"`
	// SelfHeal enables automatic self-healing
	SelfHeal bool `yaml:"self_heal,omitempty" json:"self_heal,omitempty"`
	// SyncOptions are additional sync options
	SyncOptions []string `yaml:"sync_options,omitempty" json:"sync_options,omitempty"`
}

// Config holds the multi-repository configuration.
type Config struct {
	// Enabled controls whether multi-repo support is enabled
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Primary is the main repository (usually the GitOps repo itself)
	Primary *Repository `yaml:"primary,omitempty" json:"primary,omitempty"`
	// Repositories is the list of additional repositories
	Repositories []Repository `yaml:"repositories,omitempty" json:"repositories,omitempty"`
	// ApplicationMappings maps applications to repositories
	ApplicationMappings []ApplicationMapping `yaml:"application_mappings,omitempty" json:"application_mappings,omitempty"`
}

// ApplicationMapping maps an application to a specific repository.
type ApplicationMapping struct {
	// Application is the application name or pattern (supports wildcards)
	Application string `yaml:"application" json:"application"`
	// Repository is the repository name to use
	Repository string `yaml:"repository" json:"repository"`
	// Path overrides the repository's default path
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Branch overrides the repository's default branch
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
}

// NewDefaultConfig creates a default multi-repo configuration.
func NewDefaultConfig() Config {
	return Config{
		Enabled:             false,
		Repositories:        []Repository{},
		ApplicationMappings: []ApplicationMapping{},
	}
}

// Validate validates the repository configuration.
func (r *Repository) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("repository name is required")
	}
	if r.URL == "" {
		return fmt.Errorf("repository URL is required for %s", r.Name)
	}
	if r.Type == "" {
		return fmt.Errorf("repository type is required for %s", r.Name)
	}

	validTypes := []RepositoryType{
		RepoTypeMonorepo,
		RepoTypeApplications,
		RepoTypeInfrastructure,
		RepoTypeConfig,
		RepoTypeHelmCharts,
		RepoTypeKustomize,
	}
	valid := false
	for _, t := range validTypes {
		if r.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid repository type: %s", r.Type)
	}

	return nil
}

// GetBranch returns the branch with default fallback.
func (r *Repository) GetBranch() string {
	if r.Branch != "" {
		return r.Branch
	}
	return "main"
}

// GetPath returns the path with default fallback.
func (r *Repository) GetPath() string {
	if r.Path != "" {
		return r.Path
	}
	return "."
}

// GetCredentialType returns the credential type with default.
func (r *Repository) GetCredentialType() string {
	if r.Credentials != nil && r.Credentials.Type != "" {
		return r.Credentials.Type
	}
	// Try to infer from URL
	if strings.HasPrefix(r.URL, "git@") || strings.Contains(r.URL, "ssh://") {
		return "ssh"
	}
	return "https"
}

// Manager handles multi-repository operations.
type Manager struct {
	config Config
}

// NewManager creates a new multi-repo manager.
func NewManager(config Config) *Manager {
	return &Manager{config: config}
}

// GetConfig returns the current configuration.
func (m *Manager) GetConfig() Config {
	return m.config
}

// AddRepository adds a repository to the configuration.
func (m *Manager) AddRepository(repo *Repository) error {
	if err := repo.Validate(); err != nil {
		return err
	}

	// Check for duplicate names
	for _, existing := range m.config.Repositories {
		if existing.Name == repo.Name {
			return fmt.Errorf("repository with name %s already exists", repo.Name)
		}
	}

	m.config.Repositories = append(m.config.Repositories, *repo)
	return nil
}

// RemoveRepository removes a repository by name.
func (m *Manager) RemoveRepository(name string) bool {
	for i, repo := range m.config.Repositories {
		if repo.Name == name {
			m.config.Repositories = append(m.config.Repositories[:i], m.config.Repositories[i+1:]...)
			return true
		}
	}
	return false
}

// GetRepository returns a repository by name.
func (m *Manager) GetRepository(name string) (*Repository, bool) {
	for i := range m.config.Repositories {
		if m.config.Repositories[i].Name == name {
			return &m.config.Repositories[i], true
		}
	}
	return nil, false
}

// ListRepositories returns all configured repositories.
func (m *Manager) ListRepositories() []Repository {
	return m.config.Repositories
}

// AddMapping adds an application mapping.
func (m *Manager) AddMapping(mapping ApplicationMapping) error {
	if mapping.Application == "" {
		return fmt.Errorf("application name/pattern is required")
	}
	if mapping.Repository == "" {
		return fmt.Errorf("repository name is required")
	}

	// Verify repository exists
	if _, exists := m.GetRepository(mapping.Repository); !exists {
		if m.config.Primary == nil || m.config.Primary.Name != mapping.Repository {
			return fmt.Errorf("repository %s does not exist", mapping.Repository)
		}
	}

	m.config.ApplicationMappings = append(m.config.ApplicationMappings, mapping)
	return nil
}

// GetMappingForApplication finds the mapping for an application.
func (m *Manager) GetMappingForApplication(appName string) (*ApplicationMapping, bool) {
	// First try exact match
	for i, mapping := range m.config.ApplicationMappings {
		if mapping.Application == appName {
			return &m.config.ApplicationMappings[i], true
		}
	}

	// Then try wildcard match
	for i, mapping := range m.config.ApplicationMappings {
		if matchWildcard(mapping.Application, appName) {
			return &m.config.ApplicationMappings[i], true
		}
	}

	return nil, false
}

// GetRepositoryForApplication returns the repository for an application.
func (m *Manager) GetRepositoryForApplication(appName string) (*Repository, error) {
	mapping, found := m.GetMappingForApplication(appName)
	if !found {
		// Return primary if no mapping found
		if m.config.Primary != nil {
			return m.config.Primary, nil
		}
		return nil, fmt.Errorf("no repository mapping found for application: %s", appName)
	}

	repo, exists := m.GetRepository(mapping.Repository)
	if !exists {
		if m.config.Primary != nil && m.config.Primary.Name == mapping.Repository {
			return m.config.Primary, nil
		}
		return nil, fmt.Errorf("mapped repository %s not found", mapping.Repository)
	}

	return repo, nil
}

// SetPrimary sets the primary repository.
func (m *Manager) SetPrimary(repo *Repository) error {
	if err := repo.Validate(); err != nil {
		return err
	}
	m.config.Primary = repo
	return nil
}

// GetPrimary returns the primary repository.
func (m *Manager) GetPrimary() *Repository {
	return m.config.Primary
}

// ValidateAll validates all repositories and mappings.
func (m *Manager) ValidateAll() []error {
	var errors []error

	if m.config.Primary != nil {
		if err := m.config.Primary.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("primary repository: %w", err))
		}
	}

	for i := range m.config.Repositories {
		if err := m.config.Repositories[i].Validate(); err != nil {
			errors = append(errors, fmt.Errorf("repository %s: %w", m.config.Repositories[i].Name, err))
		}
	}

	for _, mapping := range m.config.ApplicationMappings {
		if _, exists := m.GetRepository(mapping.Repository); !exists {
			if m.config.Primary == nil || m.config.Primary.Name != mapping.Repository {
				errors = append(errors, fmt.Errorf("mapping references non-existent repository: %s", mapping.Repository))
			}
		}
	}

	return errors
}

// GetRepositoriesByType returns repositories of a specific type.
func (m *Manager) GetRepositoriesByType(repoType RepositoryType) []Repository {
	var repos []Repository
	for _, repo := range m.config.Repositories {
		if repo.Type == repoType {
			repos = append(repos, repo)
		}
	}
	return repos
}

// matchWildcard performs simple wildcard matching.
func matchWildcard(pattern, str string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(str, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(str, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(str, pattern[:len(pattern)-1])
	}
	return pattern == str
}

// RepositoryManifest holds data for ArgoCD repository secret template.
type RepositoryManifest struct {
	Name          string
	Namespace     string
	URL           string
	Type          string // git, helm
	Username      string
	Password      string
	SSHPrivateKey string
	TLSClientCert string
	TLSClientKey  string
	Insecure      bool
	EnableLFS     bool
	Project       string
}

// ToRepositoryManifest creates an ArgoCD repository manifest from Repository.
func (r *Repository) ToRepositoryManifest(namespace, project string) RepositoryManifest {
	manifest := RepositoryManifest{
		Name:      r.Name,
		Namespace: namespace,
		URL:       r.URL,
		Type:      "git",
		Project:   project,
	}

	if r.Type == RepoTypeHelmCharts {
		manifest.Type = "helm"
	}

	return manifest
}
