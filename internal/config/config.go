package config

type Config struct {
	Project      Project       `yaml:"project"`
	Output       Output        `yaml:"output"`
	Platform     string        `yaml:"platform"`
	Scope        string        `yaml:"scope"`
	GitOpsTool   string        `yaml:"gitops_tool"`
	Environments []Environment `yaml:"environments"`
	Infra        Infrastructure `yaml:"infrastructure"`
	Apps         []Application `yaml:"applications"`
	Docs         Documentation `yaml:"docs"`
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

