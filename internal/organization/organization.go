// Package organization provides multi-tenancy management for GitOps,
// supporting organizations, teams, and projects with proper isolation,
// quotas, and governance.
package organization

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Organization represents a top-level organization in the GitOps hierarchy.
type Organization struct {
	Name        string            `yaml:"name" json:"name"`
	Domain      string            `yaml:"domain,omitempty" json:"domain,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Policies    Policies          `yaml:"policies,omitempty" json:"policies,omitempty"`
	Teams       []*Team           `yaml:"teams,omitempty" json:"teams,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
}

// Policies defines organization-wide policies.
type Policies struct {
	ResourceQuotas    ResourceQuotaPolicy `yaml:"resource_quotas,omitempty" json:"resource_quotas,omitempty"`
	NetworkPolicies   NetworkPolicy       `yaml:"network_policies,omitempty" json:"network_policies,omitempty"`
	PodSecurity       PodSecurityPolicy   `yaml:"pod_security,omitempty" json:"pod_security,omitempty"`
	AllowedNamespaces []string            `yaml:"allowed_namespaces,omitempty" json:"allowed_namespaces,omitempty"`
	AllowedClusters   []string            `yaml:"allowed_clusters,omitempty" json:"allowed_clusters,omitempty"`
}

// ResourceQuotaPolicy defines default resource quotas.
type ResourceQuotaPolicy struct {
	CPU        string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory     string `yaml:"memory,omitempty" json:"memory,omitempty"`
	Storage    string `yaml:"storage,omitempty" json:"storage,omitempty"`
	Pods       string `yaml:"pods,omitempty" json:"pods,omitempty"`
	Services   string `yaml:"services,omitempty" json:"services,omitempty"`
	Secrets    string `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	ConfigMaps string `yaml:"configmaps,omitempty" json:"configmaps,omitempty"`
}

// NetworkPolicy defines network isolation rules.
type NetworkPolicy struct {
	DenyAllIngress     bool `yaml:"deny_all_ingress,omitempty" json:"deny_all_ingress,omitempty"`
	AllowSameNamespace bool `yaml:"allow_same_namespace,omitempty" json:"allow_same_namespace,omitempty"`
	AllowSameTeam      bool `yaml:"allow_same_team,omitempty" json:"allow_same_team,omitempty"`
	AllowMonitoring    bool `yaml:"allow_monitoring,omitempty" json:"allow_monitoring,omitempty"`
	AllowIngress       bool `yaml:"allow_ingress,omitempty" json:"allow_ingress,omitempty"`
}

// PodSecurityPolicy defines pod security standards.
type PodSecurityPolicy struct {
	Level     string `yaml:"level,omitempty" json:"level,omitempty"` // privileged, baseline, restricted
	Enforce   bool   `yaml:"enforce,omitempty" json:"enforce,omitempty"`
	AuditMode bool   `yaml:"audit_mode,omitempty" json:"audit_mode,omitempty"`
	WarnMode  bool   `yaml:"warn_mode,omitempty" json:"warn_mode,omitempty"`
}

// Team represents a team within an organization.
type Team struct {
	Name            string            `yaml:"name" json:"name"`
	Description     string            `yaml:"description,omitempty" json:"description,omitempty"`
	Owners          []string          `yaml:"owners,omitempty" json:"owners,omitempty"`
	Members         []string          `yaml:"members,omitempty" json:"members,omitempty"`
	IsAdmin         bool              `yaml:"admin,omitempty" json:"admin,omitempty"`
	Quotas          TeamQuotas        `yaml:"quotas,omitempty" json:"quotas,omitempty"`
	AllowedClusters []string          `yaml:"allowed_clusters,omitempty" json:"allowed_clusters,omitempty"`
	Projects        []*Project        `yaml:"projects,omitempty" json:"projects,omitempty"`
	Labels          map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	CostCenter      string            `yaml:"cost_center,omitempty" json:"cost_center,omitempty"`
	CreatedAt       time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt       time.Time         `yaml:"updated_at" json:"updated_at"`
}

// TeamQuotas defines resource limits for a team.
type TeamQuotas struct {
	CPU        string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory     string `yaml:"memory,omitempty" json:"memory,omitempty"`
	Storage    string `yaml:"storage,omitempty" json:"storage,omitempty"`
	Namespaces int    `yaml:"namespaces,omitempty" json:"namespaces,omitempty"`
	Pods       string `yaml:"pods,omitempty" json:"pods,omitempty"`
}

// Project represents a project within a team.
type Project struct {
	Name         string            `yaml:"name" json:"name"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	Owners       []string          `yaml:"owners,omitempty" json:"owners,omitempty"`
	Environments []string          `yaml:"environments,omitempty" json:"environments,omitempty"`
	SourceRepo   string            `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations  map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	PCICompliant bool              `yaml:"pci_compliant,omitempty" json:"pci_compliant,omitempty"`
	CreatedAt    time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `yaml:"updated_at" json:"updated_at"`
}

// Manager handles organization operations.
type Manager struct {
	configPath   string
	organization *Organization
}

// NewManager creates a new organization manager.
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load organization: %w", err)
		}
	}

	return m, nil
}

// Load loads the organization from the config file.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var org Organization
	if err := yaml.Unmarshal(data, &org); err != nil {
		return err
	}

	m.organization = &org
	return nil
}

// Save saves the organization to the config file.
func (m *Manager) Save() error {
	if m.organization == nil {
		return fmt.Errorf("no organization to save")
	}

	m.organization.UpdatedAt = time.Now()

	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(m.organization)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// InitOrganization initializes a new organization.
func (m *Manager) InitOrganization(_ context.Context, name, domain string) (*Organization, error) {
	if m.organization != nil {
		return nil, fmt.Errorf("organization already exists")
	}

	if name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	m.organization = &Organization{
		Name:   name,
		Domain: domain,
		Policies: Policies{
			ResourceQuotas: ResourceQuotaPolicy{
				CPU:    "10",
				Memory: "20Gi",
			},
			NetworkPolicies: NetworkPolicy{
				DenyAllIngress:     true,
				AllowSameNamespace: true,
			},
			PodSecurity: PodSecurityPolicy{
				Level:   "restricted",
				Enforce: true,
			},
		},
		Teams:     []*Team{},
		Labels:    make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return m.organization, m.Save()
}

// GetOrganization returns the current organization.
func (m *Manager) GetOrganization() *Organization {
	return m.organization
}

// SetPolicy updates organization policies.
func (m *Manager) SetPolicy(_ context.Context, policy string, value interface{}) error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	switch policy {
	case "resource-quotas":
		if quotas, ok := value.(ResourceQuotaPolicy); ok {
			m.organization.Policies.ResourceQuotas = quotas
		} else {
			return fmt.Errorf("invalid resource quotas value")
		}
	case "network-policies":
		if np, ok := value.(NetworkPolicy); ok {
			m.organization.Policies.NetworkPolicies = np
		} else {
			return fmt.Errorf("invalid network policies value")
		}
	case "pod-security":
		if ps, ok := value.(PodSecurityPolicy); ok {
			m.organization.Policies.PodSecurity = ps
		} else {
			return fmt.Errorf("invalid pod security policy value")
		}
	default:
		return fmt.Errorf("unknown policy: %s", policy)
	}

	return m.Save()
}

// CreateTeam creates a new team within the organization.
func (m *Manager) CreateTeam(_ context.Context, opts *TeamOptions) (*Team, error) {
	if m.organization == nil {
		return nil, fmt.Errorf("no organization initialized")
	}

	if opts.Name == "" {
		return nil, fmt.Errorf("team name is required")
	}

	// Check for duplicate team
	for _, t := range m.organization.Teams {
		if t.Name == opts.Name {
			return nil, fmt.Errorf("team %q already exists", opts.Name)
		}
	}

	team := &Team{
		Name:            opts.Name,
		Description:     opts.Description,
		Owners:          opts.Owners,
		Members:         opts.Members,
		IsAdmin:         opts.IsAdmin,
		Quotas:          opts.Quotas,
		AllowedClusters: opts.AllowedClusters,
		Projects:        []*Project{},
		Labels:          opts.Labels,
		CostCenter:      opts.CostCenter,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	m.organization.Teams = append(m.organization.Teams, team)

	return team, m.Save()
}

// TeamOptions contains options for creating a team.
type TeamOptions struct {
	Name            string
	Description     string
	Owners          []string
	Members         []string
	IsAdmin         bool
	Quotas          TeamQuotas
	AllowedClusters []string
	Labels          map[string]string
	CostCenter      string
}

// GetTeam returns a team by name.
func (m *Manager) GetTeam(_ context.Context, name string) (*Team, error) {
	if m.organization == nil {
		return nil, fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == name {
			return t, nil
		}
	}

	return nil, fmt.Errorf("team %q not found", name)
}

// ListTeams returns all teams.
func (m *Manager) ListTeams(_ context.Context) ([]*Team, error) {
	if m.organization == nil {
		return nil, fmt.Errorf("no organization initialized")
	}

	return m.organization.Teams, nil
}

// DeleteTeam removes a team.
func (m *Manager) DeleteTeam(_ context.Context, name string) error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	for i, t := range m.organization.Teams {
		if t.Name == name {
			m.organization.Teams = append(m.organization.Teams[:i], m.organization.Teams[i+1:]...)
			return m.Save()
		}
	}

	return fmt.Errorf("team %q not found", name)
}

// SetTeamQuota updates a team's quotas.
func (m *Manager) SetTeamQuota(_ context.Context, teamName string, quotas TeamQuotas) error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			t.Quotas = quotas
			t.UpdatedAt = time.Now()
			return m.Save()
		}
	}

	return fmt.Errorf("team %q not found", teamName)
}

// AddClusterToTeam adds a cluster to a team's allowed list.
func (m *Manager) AddClusterToTeam(_ context.Context, teamName, clusterName string) error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			for _, c := range t.AllowedClusters {
				if c == clusterName {
					return nil // Already exists
				}
			}
			t.AllowedClusters = append(t.AllowedClusters, clusterName)
			t.UpdatedAt = time.Now()
			return m.Save()
		}
	}

	return fmt.Errorf("team %q not found", teamName)
}

// CreateProject creates a new project within a team.
func (m *Manager) CreateProject(_ context.Context, teamName string, opts *ProjectOptions) (*Project, error) {
	if m.organization == nil {
		return nil, fmt.Errorf("no organization initialized")
	}

	var team *Team
	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			team = t
			break
		}
	}

	if team == nil {
		return nil, fmt.Errorf("team %q not found", teamName)
	}

	if opts.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// Check for duplicate project
	for _, p := range team.Projects {
		if p.Name == opts.Name {
			return nil, fmt.Errorf("project %q already exists in team %q", opts.Name, teamName)
		}
	}

	project := &Project{
		Name:         opts.Name,
		Description:  opts.Description,
		Owners:       opts.Owners,
		Environments: opts.Environments,
		SourceRepo:   opts.SourceRepo,
		Labels:       opts.Labels,
		Annotations:  opts.Annotations,
		PCICompliant: opts.PCICompliant,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if len(project.Environments) == 0 {
		project.Environments = []string{"dev", "staging", "prod"}
	}

	team.Projects = append(team.Projects, project)
	team.UpdatedAt = time.Now()

	return project, m.Save()
}

// ProjectOptions contains options for creating a project.
type ProjectOptions struct {
	Name         string
	Description  string
	Owners       []string
	Environments []string
	SourceRepo   string
	Labels       map[string]string
	Annotations  map[string]string
	PCICompliant bool
}

// GetProject returns a project by name.
func (m *Manager) GetProject(_ context.Context, teamName, projectName string) (*Project, error) {
	if m.organization == nil {
		return nil, fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			for _, p := range t.Projects {
				if p.Name == projectName {
					return p, nil
				}
			}
			return nil, fmt.Errorf("project %q not found in team %q", projectName, teamName)
		}
	}

	return nil, fmt.Errorf("team %q not found", teamName)
}

// ListProjects returns all projects in a team.
func (m *Manager) ListProjects(_ context.Context, teamName string) ([]*Project, error) {
	if m.organization == nil {
		return nil, fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			return t.Projects, nil
		}
	}

	return nil, fmt.Errorf("team %q not found", teamName)
}

// DeleteProject removes a project from a team.
func (m *Manager) DeleteProject(_ context.Context, teamName, projectName string) error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			for i, p := range t.Projects {
				if p.Name == projectName {
					t.Projects = append(t.Projects[:i], t.Projects[i+1:]...)
					t.UpdatedAt = time.Now()
					return m.Save()
				}
			}
			return fmt.Errorf("project %q not found in team %q", projectName, teamName)
		}
	}

	return fmt.Errorf("team %q not found", teamName)
}

// AddEnvironmentToProject adds an environment to a project.
func (m *Manager) AddEnvironmentToProject(_ context.Context, teamName, projectName, envName string) error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	for _, t := range m.organization.Teams {
		if t.Name == teamName {
			for _, p := range t.Projects {
				if p.Name == projectName {
					for _, e := range p.Environments {
						if e == envName {
							return nil // Already exists
						}
					}
					p.Environments = append(p.Environments, envName)
					p.UpdatedAt = time.Now()
					t.UpdatedAt = time.Now()
					return m.Save()
				}
			}
			return fmt.Errorf("project %q not found in team %q", projectName, teamName)
		}
	}

	return fmt.Errorf("team %q not found", teamName)
}

// GenerateNamespace generates a namespace name for a team project environment.
func GenerateNamespace(org, team, project, env string) string {
	parts := []string{}
	if team != "" {
		parts = append(parts, team)
	}
	if project != "" {
		parts = append(parts, project)
	}
	if env != "" {
		parts = append(parts, env)
	}
	return strings.Join(parts, "-")
}

// ToYAML returns the organization as YAML.
func (m *Manager) ToYAML() (string, error) {
	if m.organization == nil {
		return "", fmt.Errorf("no organization initialized")
	}

	data, err := yaml.Marshal(m.organization)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GenerateAppProjectName generates an ArgoCD AppProject name for a team.
func GenerateAppProjectName(org, team string) string {
	if org != "" {
		return fmt.Sprintf("%s-%s", org, team)
	}
	return fmt.Sprintf("%s-team", team)
}

// GenerateSourceRepoPattern generates a source repo pattern for a team.
func GenerateSourceRepoPattern(domain, team string) string {
	return fmt.Sprintf("https://%s/%s-*", domain, team)
}

// GenerateDestinationPattern generates a namespace destination pattern for a team.
func GenerateDestinationPattern(team string) string {
	return fmt.Sprintf("%s-*", team)
}

// Validate validates the organization structure.
func (m *Manager) Validate() error {
	if m.organization == nil {
		return fmt.Errorf("no organization initialized")
	}

	if m.organization.Name == "" {
		return fmt.Errorf("organization name is required")
	}

	// Check for duplicate team names
	teamNames := make(map[string]bool)
	for _, t := range m.organization.Teams {
		if t.Name == "" {
			return fmt.Errorf("team name is required")
		}
		if teamNames[t.Name] {
			return fmt.Errorf("duplicate team name: %s", t.Name)
		}
		teamNames[t.Name] = true

		// Check for duplicate project names within team
		projectNames := make(map[string]bool)
		for _, p := range t.Projects {
			if p.Name == "" {
				return fmt.Errorf("project name is required in team %s", t.Name)
			}
			if projectNames[p.Name] {
				return fmt.Errorf("duplicate project name %s in team %s", p.Name, t.Name)
			}
			projectNames[p.Name] = true
		}
	}

	return nil
}
