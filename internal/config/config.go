package config

// Preset defines a configuration preset type
type Preset string

const (
	PresetMinimal    Preset = "minimal"
	PresetStandard   Preset = "standard"
	PresetEnterprise Preset = "enterprise"
	PresetCustom     Preset = "custom"
)

// Config represents the complete gitopsi configuration.
type Config struct {
	Preset       Preset              `yaml:"preset,omitempty"`
	Project      Project             `yaml:"project"`
	Structure    StructureConfig     `yaml:"structure,omitempty"`
	Generate     GenerateConfig      `yaml:"generate,omitempty"`
	Output       Output              `yaml:"output"`
	Git          GitConfig           `yaml:"git"`
	Cluster      ClusterConfig       `yaml:"cluster"`
	Bootstrap    BootstrapConfig     `yaml:"bootstrap"`
	Platform     string              `yaml:"platform"`
	Scope        string              `yaml:"scope"`
	GitOpsTool   string              `yaml:"gitops_tool"`
	Topology     EnvironmentTopology `yaml:"topology,omitempty"`
	Environments []Environment       `yaml:"environments"`
	Infra        Infrastructure      `yaml:"infrastructure"`
	Apps         []Application       `yaml:"applications"`
	Docs         Documentation       `yaml:"docs"`
}

// StructureConfig defines custom directory structure
type StructureConfig struct {
	InfrastructureDir string      `yaml:"infrastructure_dir,omitempty"`
	ApplicationsDir   string      `yaml:"applications_dir,omitempty"`
	BootstrapDir      string      `yaml:"bootstrap_dir,omitempty"`
	ScriptsDir        string      `yaml:"scripts_dir,omitempty"`
	DocsDir           string      `yaml:"docs_dir,omitempty"`
	CustomDirs        []CustomDir `yaml:"custom_dirs,omitempty"`
}

// CustomDir defines a custom directory to create
type CustomDir struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description,omitempty"`
}

// GenerateConfig defines what components to generate
type GenerateConfig struct {
	Infrastructure GenerateInfra   `yaml:"infrastructure,omitempty"`
	Applications   GenerateApps    `yaml:"applications,omitempty"`
	GitOps         GenerateGitOps  `yaml:"gitops,omitempty"`
	Docs           GenerateDocs    `yaml:"docs,omitempty"`
	Scripts        GenerateScripts `yaml:"scripts,omitempty"`
}

// GenerateInfra defines infrastructure generation options
type GenerateInfra struct {
	Namespaces      *bool `yaml:"namespaces,omitempty"`
	RBAC            *bool `yaml:"rbac,omitempty"`
	NetworkPolicies *bool `yaml:"network_policies,omitempty"`
	ResourceQuotas  *bool `yaml:"resource_quotas,omitempty"`
	LimitRanges     *bool `yaml:"limit_ranges,omitempty"`
	ServiceAccounts *bool `yaml:"service_accounts,omitempty"`
}

// GenerateApps defines application generation options
type GenerateApps struct {
	Deployments     *bool `yaml:"deployments,omitempty"`
	Services        *bool `yaml:"services,omitempty"`
	Ingress         *bool `yaml:"ingress,omitempty"`
	HPA             *bool `yaml:"hpa,omitempty"`
	PDB             *bool `yaml:"pdb,omitempty"`
	ServiceAccounts *bool `yaml:"service_accounts,omitempty"`
}

// GenerateGitOps defines GitOps resource generation options
type GenerateGitOps struct {
	Projects        *bool `yaml:"projects,omitempty"`
	Applications    *bool `yaml:"applications,omitempty"`
	ApplicationSets *bool `yaml:"applicationsets,omitempty"`
	AppOfApps       *bool `yaml:"app_of_apps,omitempty"`
}

// GenerateDocs defines documentation generation options
type GenerateDocs struct {
	Readme       *bool `yaml:"readme,omitempty"`
	Architecture *bool `yaml:"architecture,omitempty"`
	Onboarding   *bool `yaml:"onboarding,omitempty"`
	Runbook      *bool `yaml:"runbook,omitempty"`
	ADR          *bool `yaml:"adr,omitempty"`
}

// GenerateScripts defines script generation options
type GenerateScripts struct {
	Bootstrap *bool `yaml:"bootstrap,omitempty"`
	Validate  *bool `yaml:"validate,omitempty"`
	Sync      *bool `yaml:"sync,omitempty"`
	Rollback  *bool `yaml:"rollback,omitempty"`
}

type Project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Output struct {
	Type   string `yaml:"type"`
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
}

type GitConfig struct {
	URL             string      `yaml:"url"`
	Branch          string      `yaml:"branch"`
	Provider        GitProvider `yaml:"provider"`
	Auth            GitAuth     `yaml:"auth"`
	PushOnInit      bool        `yaml:"push_on_init"`
	CreateIfMissing bool        `yaml:"create_if_missing"`
}

type GitProvider struct {
	Name     string `yaml:"name"`
	Instance string `yaml:"instance"`
}

type GitAuth struct {
	Method   string `yaml:"method"`
	Token    string `yaml:"token"`
	SSHKey   string `yaml:"ssh_key"`
	TokenEnv string `yaml:"token_env"`
}

// ClusterConfig holds target cluster configuration.
type ClusterConfig struct {
	URL        string      `yaml:"url"`
	Name       string      `yaml:"name"`
	Auth       ClusterAuth `yaml:"auth"`
	Platform   string      `yaml:"platform"`   // kubernetes, openshift, aks, eks, gke
	Kubeconfig string      `yaml:"kubeconfig"` // Path to kubeconfig file
	Context    string      `yaml:"context"`    // Kubeconfig context to use
}

// ClusterAuth holds cluster authentication configuration.
type ClusterAuth struct {
	Method   string `yaml:"method"`    // kubeconfig, token, oidc, service-account
	Token    string `yaml:"token"`     // Bearer token
	TokenEnv string `yaml:"token_env"` // Env var containing token
	CACert   string `yaml:"ca_cert"`   // CA certificate path
	SkipTLS  bool   `yaml:"skip_tls"`  // Skip TLS verification (not recommended)
}

// BootstrapConfig holds GitOps tool bootstrap configuration.
type BootstrapConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Tool            string `yaml:"tool"`               // argocd, flux
	Mode            string `yaml:"mode"`               // helm, olm, manifest, kustomize
	Namespace       string `yaml:"namespace"`          // Namespace to install GitOps tool
	Wait            bool   `yaml:"wait"`               // Wait for GitOps tool to be ready
	Timeout         int    `yaml:"timeout"`            // Timeout in seconds
	ConfigureRepo   bool   `yaml:"configure_repo"`     // Add repo to GitOps tool
	CreateAppOfApps bool   `yaml:"create_app_of_apps"` // Create root application
	SyncInitial     bool   `yaml:"sync_initial"`       // Trigger initial sync
	Version         string `yaml:"version,omitempty"`  // Tool version

	// Mode-specific configurations
	Helm      *BootstrapHelmConfig      `yaml:"helm,omitempty"`
	OLM       *BootstrapOLMConfig       `yaml:"olm,omitempty"`
	Manifest  *BootstrapManifestConfig  `yaml:"manifest,omitempty"`
	Kustomize *BootstrapKustomizeConfig `yaml:"kustomize,omitempty"`
}

// BootstrapHelmConfig holds Helm-specific bootstrap configuration.
type BootstrapHelmConfig struct {
	Repo      string            `yaml:"repo,omitempty"`
	Chart     string            `yaml:"chart,omitempty"`
	Version   string            `yaml:"version,omitempty"`
	Values    map[string]any    `yaml:"values,omitempty"`
	SetValues map[string]string `yaml:"set_values,omitempty"`
}

// BootstrapOLMConfig holds OLM-specific bootstrap configuration.
type BootstrapOLMConfig struct {
	Channel         string `yaml:"channel,omitempty"`
	Source          string `yaml:"source,omitempty"`
	SourceNamespace string `yaml:"source_namespace,omitempty"`
	Approval        string `yaml:"approval,omitempty"` // Automatic, Manual
}

// BootstrapManifestConfig holds manifest-specific bootstrap configuration.
type BootstrapManifestConfig struct {
	URL   string   `yaml:"url,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
}

// BootstrapKustomizeConfig holds Kustomize-specific bootstrap configuration.
type BootstrapKustomizeConfig struct {
	URL     string   `yaml:"url,omitempty"`
	Path    string   `yaml:"path,omitempty"`
	Patches []string `yaml:"patches,omitempty"`
}

type Environment struct {
	Name      string               `yaml:"name"`
	Cluster   string               `yaml:"cluster,omitempty"`
	Namespace string               `yaml:"namespace,omitempty"`
	Clusters  []EnvironmentCluster `yaml:"clusters,omitempty"`
}

type EnvironmentCluster struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Namespace string `yaml:"namespace,omitempty"`
	Region    string `yaml:"region,omitempty"`
	Primary   bool   `yaml:"primary,omitempty"`
}

type EnvironmentTopology string

const (
	TopologyNamespaceBased EnvironmentTopology = "namespace-based"
	TopologyClusterPerEnv  EnvironmentTopology = "cluster-per-env"
	TopologyMultiCluster   EnvironmentTopology = "multi-cluster"
)

type Infrastructure struct {
	Namespaces      bool `yaml:"namespaces"`
	RBAC            bool `yaml:"rbac"`
	NetworkPolicies bool `yaml:"network_policies"`
	ResourceQuotas  bool `yaml:"resource_quotas"`
}

type Application struct {
	Name     string `yaml:"name"`
	Image    string `yaml:"image"`
	Port     int    `yaml:"port"`
	Replicas int    `yaml:"replicas"`
}

type Documentation struct {
	Readme       bool `yaml:"readme"`
	Architecture bool `yaml:"architecture"`
	Onboarding   bool `yaml:"onboarding"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Platform:   "kubernetes",
		Scope:      "both",
		GitOpsTool: "argocd",
		Topology:   TopologyNamespaceBased,
		Output: Output{
			Type:   "local",
			Branch: "main",
		},
		Git: GitConfig{
			Branch:     "main",
			PushOnInit: false,
			Auth: GitAuth{
				Method: "ssh",
			},
		},
		Cluster: ClusterConfig{
			Platform: "kubernetes",
			Auth: ClusterAuth{
				Method: "kubeconfig",
			},
		},
		Bootstrap: BootstrapConfig{
			Enabled:         false,
			Tool:            "argocd",
			Mode:            "helm",
			Namespace:       "argocd",
			Wait:            true,
			Timeout:         300,
			ConfigureRepo:   true,
			CreateAppOfApps: true,
			SyncInitial:     true,
		},
		Environments: []Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
		Infra: Infrastructure{
			Namespaces:      true,
			RBAC:            true,
			NetworkPolicies: true,
			ResourceQuotas:  true,
		},
		Docs: Documentation{
			Readme:       true,
			Architecture: true,
			Onboarding:   true,
		},
	}
}

func (t EnvironmentTopology) IsValid() bool {
	switch t {
	case TopologyNamespaceBased, TopologyClusterPerEnv, TopologyMultiCluster, "":
		return true
	}
	return false
}

func (c *Config) GetEnvironmentNamespace(envName string) string {
	for _, env := range c.Environments {
		if env.Name == envName {
			if env.Namespace != "" {
				return env.Namespace
			}
			return c.Project.Name + "-" + env.Name
		}
	}
	return c.Project.Name + "-" + envName
}

func (c *Config) GetEnvironmentClusters(envName string) []EnvironmentCluster {
	for _, env := range c.Environments {
		if env.Name == envName {
			return env.Clusters
		}
	}
	return nil
}

func (c *Config) IsMultiCluster() bool {
	return c.Topology == TopologyClusterPerEnv || c.Topology == TopologyMultiCluster
}

// ApplyPreset applies a preset configuration
func (c *Config) ApplyPreset() {
	switch c.Preset {
	case PresetMinimal:
		c.applyMinimalPreset()
	case PresetStandard:
		c.applyStandardPreset()
	case PresetEnterprise:
		c.applyEnterprisePreset()
	}
}

func (c *Config) applyMinimalPreset() {
	c.Infra = Infrastructure{
		Namespaces:      true,
		RBAC:            false,
		NetworkPolicies: false,
		ResourceQuotas:  false,
	}
	c.Docs = Documentation{
		Readme:       true,
		Architecture: false,
		Onboarding:   false,
	}
	if len(c.Environments) == 0 {
		c.Environments = []Environment{{Name: "dev"}}
	}
}

func (c *Config) applyStandardPreset() {
	c.Infra = Infrastructure{
		Namespaces:      true,
		RBAC:            true,
		NetworkPolicies: true,
		ResourceQuotas:  true,
	}
	c.Docs = Documentation{
		Readme:       true,
		Architecture: true,
		Onboarding:   true,
	}
	if len(c.Environments) == 0 {
		c.Environments = []Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		}
	}
}

func (c *Config) applyEnterprisePreset() {
	c.Infra = Infrastructure{
		Namespaces:      true,
		RBAC:            true,
		NetworkPolicies: true,
		ResourceQuotas:  true,
	}
	c.Docs = Documentation{
		Readme:       true,
		Architecture: true,
		Onboarding:   true,
	}
	// Enterprise adds custom directories for monitoring and policies
	c.Structure.CustomDirs = append(c.Structure.CustomDirs,
		CustomDir{Path: "monitoring/dashboards", Description: "Grafana dashboards"},
		CustomDir{Path: "monitoring/alerts", Description: "Alertmanager rules"},
		CustomDir{Path: "policies/opa", Description: "OPA/Gatekeeper policies"},
	)
	if len(c.Environments) == 0 {
		c.Environments = []Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		}
	}
}

// GetInfrastructureDir returns the infrastructure directory name
func (c *Config) GetInfrastructureDir() string {
	if c.Structure.InfrastructureDir != "" {
		return c.Structure.InfrastructureDir
	}
	return "infrastructure"
}

// GetApplicationsDir returns the applications directory name
func (c *Config) GetApplicationsDir() string {
	if c.Structure.ApplicationsDir != "" {
		return c.Structure.ApplicationsDir
	}
	return "applications"
}

// GetBootstrapDir returns the bootstrap directory name
func (c *Config) GetBootstrapDir() string {
	if c.Structure.BootstrapDir != "" {
		return c.Structure.BootstrapDir
	}
	return "bootstrap"
}

// GetScriptsDir returns the scripts directory name
func (c *Config) GetScriptsDir() string {
	if c.Structure.ScriptsDir != "" {
		return c.Structure.ScriptsDir
	}
	return "scripts"
}

// GetDocsDir returns the docs directory name
func (c *Config) GetDocsDir() string {
	if c.Structure.DocsDir != "" {
		return c.Structure.DocsDir
	}
	return "docs"
}

// ShouldGenerateNamespaces returns whether to generate namespaces
func (c *Config) ShouldGenerateNamespaces() bool {
	if c.Generate.Infrastructure.Namespaces != nil {
		return *c.Generate.Infrastructure.Namespaces
	}
	return c.Infra.Namespaces
}

// ShouldGenerateRBAC returns whether to generate RBAC
func (c *Config) ShouldGenerateRBAC() bool {
	if c.Generate.Infrastructure.RBAC != nil {
		return *c.Generate.Infrastructure.RBAC
	}
	return c.Infra.RBAC
}

// ShouldGenerateNetworkPolicies returns whether to generate network policies
func (c *Config) ShouldGenerateNetworkPolicies() bool {
	if c.Generate.Infrastructure.NetworkPolicies != nil {
		return *c.Generate.Infrastructure.NetworkPolicies
	}
	return c.Infra.NetworkPolicies
}

// ShouldGenerateResourceQuotas returns whether to generate resource quotas
func (c *Config) ShouldGenerateResourceQuotas() bool {
	if c.Generate.Infrastructure.ResourceQuotas != nil {
		return *c.Generate.Infrastructure.ResourceQuotas
	}
	return c.Infra.ResourceQuotas
}

// ShouldGenerateReadme returns whether to generate README
func (c *Config) ShouldGenerateReadme() bool {
	if c.Generate.Docs.Readme != nil {
		return *c.Generate.Docs.Readme
	}
	return c.Docs.Readme
}

// ShouldGenerateArchitecture returns whether to generate architecture docs
func (c *Config) ShouldGenerateArchitecture() bool {
	if c.Generate.Docs.Architecture != nil {
		return *c.Generate.Docs.Architecture
	}
	return c.Docs.Architecture
}

// ShouldGenerateOnboarding returns whether to generate onboarding docs
func (c *Config) ShouldGenerateOnboarding() bool {
	if c.Generate.Docs.Onboarding != nil {
		return *c.Generate.Docs.Onboarding
	}
	return c.Docs.Onboarding
}

// ValidPresets returns list of valid preset names
func ValidPresets() []string {
	return []string{string(PresetMinimal), string(PresetStandard), string(PresetEnterprise), string(PresetCustom)}
}

// IsValidPreset checks if a preset name is valid
func IsValidPreset(preset string) bool {
	for _, p := range ValidPresets() {
		if p == preset {
			return true
		}
	}
	return preset == ""
}
