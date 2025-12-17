package environment

import (
	"fmt"
)

type Topology string

const (
	TopologyNamespaceBased   Topology = "namespace-based"
	TopologyClusterPerEnv    Topology = "cluster-per-env"
	TopologyMultiCluster     Topology = "multi-cluster"
)

func (t Topology) String() string {
	return string(t)
}

func (t Topology) IsValid() bool {
	switch t {
	case TopologyNamespaceBased, TopologyClusterPerEnv, TopologyMultiCluster:
		return true
	}
	return false
}

func ParseTopology(s string) (Topology, error) {
	t := Topology(s)
	if !t.IsValid() {
		return "", fmt.Errorf("invalid topology: %s (valid: namespace-based, cluster-per-env, multi-cluster)", s)
	}
	return t, nil
}

func ValidTopologies() []string {
	return []string{
		string(TopologyNamespaceBased),
		string(TopologyClusterPerEnv),
		string(TopologyMultiCluster),
	}
}

type ClusterInfo struct {
	Name      string `yaml:"name" json:"name"`
	URL       string `yaml:"url" json:"url"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Region    string `yaml:"region,omitempty" json:"region,omitempty"`
	Primary   bool   `yaml:"primary,omitempty" json:"primary,omitempty"`
}

type Environment struct {
	Name      string        `yaml:"name" json:"name"`
	Namespace string        `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Clusters  []ClusterInfo `yaml:"clusters,omitempty" json:"clusters,omitempty"`
}

func (e *Environment) GetNamespace(projectName string) string {
	if e.Namespace != "" {
		return e.Namespace
	}
	return fmt.Sprintf("%s-%s", projectName, e.Name)
}

func (e *Environment) GetPrimaryCluster() *ClusterInfo {
	if len(e.Clusters) == 0 {
		return nil
	}
	for i := range e.Clusters {
		if e.Clusters[i].Primary {
			return &e.Clusters[i]
		}
	}
	return &e.Clusters[0]
}

func (e *Environment) HasCluster(name string) bool {
	for _, c := range e.Clusters {
		if c.Name == name {
			return true
		}
	}
	return false
}

func (e *Environment) AddCluster(cluster ClusterInfo) error {
	if e.HasCluster(cluster.Name) {
		return fmt.Errorf("cluster %s already exists in environment %s", cluster.Name, e.Name)
	}
	e.Clusters = append(e.Clusters, cluster)
	return nil
}

func (e *Environment) RemoveCluster(name string) error {
	for i, c := range e.Clusters {
		if c.Name == name {
			e.Clusters = append(e.Clusters[:i], e.Clusters[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("cluster %s not found in environment %s", name, e.Name)
}

type Config struct {
	Topology     Topology       `yaml:"topology" json:"topology"`
	Cluster      string         `yaml:"cluster,omitempty" json:"cluster,omitempty"`
	Environments []*Environment `yaml:"environments" json:"environments"`
}

func NewConfig() *Config {
	return &Config{
		Topology:     TopologyNamespaceBased,
		Environments: []*Environment{},
	}
}

func DefaultConfig(projectName string) *Config {
	return &Config{
		Topology: TopologyNamespaceBased,
		Environments: []*Environment{
			{Name: "dev", Namespace: projectName + "-dev"},
			{Name: "staging", Namespace: projectName + "-staging"},
			{Name: "prod", Namespace: projectName + "-prod"},
		},
	}
}

func (c *Config) GetEnvironment(name string) *Environment {
	for _, e := range c.Environments {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func (c *Config) HasEnvironment(name string) bool {
	return c.GetEnvironment(name) != nil
}

func (c *Config) AddEnvironment(env *Environment) error {
	if c.HasEnvironment(env.Name) {
		return fmt.Errorf("environment %s already exists", env.Name)
	}
	c.Environments = append(c.Environments, env)
	return nil
}

func (c *Config) RemoveEnvironment(name string) error {
	for i, e := range c.Environments {
		if e.Name == name {
			c.Environments = append(c.Environments[:i], c.Environments[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("environment %s not found", name)
}

func (c *Config) Validate() error {
	if !c.Topology.IsValid() {
		return fmt.Errorf("invalid topology: %s", c.Topology)
	}

	if len(c.Environments) == 0 {
		return fmt.Errorf("at least one environment is required")
	}

	switch c.Topology {
	case TopologyNamespaceBased:
		if c.Cluster == "" {
			return fmt.Errorf("cluster URL is required for namespace-based topology")
		}
	case TopologyClusterPerEnv, TopologyMultiCluster:
		for _, env := range c.Environments {
			if len(env.Clusters) == 0 {
				return fmt.Errorf("at least one cluster is required for environment %s in %s topology", env.Name, c.Topology)
			}
		}
	}

	names := make(map[string]bool)
	for _, env := range c.Environments {
		if names[env.Name] {
			return fmt.Errorf("duplicate environment name: %s", env.Name)
		}
		names[env.Name] = true
	}

	return nil
}

func (c *Config) GetAllClusters() []ClusterInfo {
	seen := make(map[string]bool)
	var clusters []ClusterInfo

	if c.Cluster != "" {
		clusters = append(clusters, ClusterInfo{
			Name: "default",
			URL:  c.Cluster,
		})
		seen[c.Cluster] = true
	}

	for _, env := range c.Environments {
		for _, cluster := range env.Clusters {
			if !seen[cluster.URL] {
				clusters = append(clusters, cluster)
				seen[cluster.URL] = true
			}
		}
	}

	return clusters
}

func (c *Config) Summary() string {
	switch c.Topology {
	case TopologyNamespaceBased:
		return fmt.Sprintf("Single cluster (%s) with %d environments (namespace-based)",
			c.Cluster, len(c.Environments))
	case TopologyClusterPerEnv:
		return fmt.Sprintf("%d environments with dedicated clusters", len(c.Environments))
	case TopologyMultiCluster:
		totalClusters := 0
		for _, env := range c.Environments {
			totalClusters += len(env.Clusters)
		}
		return fmt.Sprintf("%d environments across %d clusters (multi-cluster)", len(c.Environments), totalClusters)
	}
	return "Unknown topology"
}

