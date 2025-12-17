package config

// Config represents the complete gitopsi configuration.
type Config struct {
	Project      Project        `yaml:"project"`
	Output       Output         `yaml:"output"`
	Git          GitConfig      `yaml:"git"`
	Cluster      ClusterConfig  `yaml:"cluster"`
	Bootstrap    BootstrapConfig `yaml:"bootstrap"`
	Platform     string         `yaml:"platform"`
	Scope        string         `yaml:"scope"`
	GitOpsTool   string         `yaml:"gitops_tool"`
	Environments []Environment  `yaml:"environments"`
	Infra        Infrastructure `yaml:"infrastructure"`
	Apps         []Application  `yaml:"applications"`
	Docs         Documentation  `yaml:"docs"`
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
	Method    string `yaml:"method"`    // kubeconfig, token, oidc, service-account
	Token     string `yaml:"token"`     // Bearer token
	TokenEnv  string `yaml:"token_env"` // Env var containing token
	CACert    string `yaml:"ca_cert"`   // CA certificate path
	SkipTLS   bool   `yaml:"skip_tls"`  // Skip TLS verification (not recommended)
}

// BootstrapConfig holds GitOps tool bootstrap configuration.
type BootstrapConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Tool            string `yaml:"tool"`              // argocd, flux
	Mode            string `yaml:"mode"`              // helm, olm, manifest
	Namespace       string `yaml:"namespace"`         // Namespace to install GitOps tool
	Wait            bool   `yaml:"wait"`              // Wait for GitOps tool to be ready
	Timeout         int    `yaml:"timeout"`           // Timeout in seconds
	ConfigureRepo   bool   `yaml:"configure_repo"`    // Add repo to GitOps tool
	CreateAppOfApps bool   `yaml:"create_app_of_apps"` // Create root application
	SyncInitial     bool   `yaml:"sync_initial"`      // Trigger initial sync
}

type Environment struct {
	Name    string `yaml:"name"`
	Cluster string `yaml:"cluster"`
}

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
