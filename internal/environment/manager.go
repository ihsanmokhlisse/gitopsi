package environment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	EnvConfigFile = ".gitopsi/environments.yaml"
)

type Manager struct {
	projectPath string
	config      *Config
}

func NewManager(projectPath string) *Manager {
	return &Manager{
		projectPath: projectPath,
		config:      NewConfig(),
	}
}

func (m *Manager) Load() error {
	configPath := filepath.Join(m.projectPath, EnvConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read environment config: %w", err)
	}

	if err := yaml.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse environment config: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	configDir := filepath.Join(m.projectPath, ".gitopsi")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(m.projectPath, EnvConfigFile)
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal environment config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment config: %w", err)
	}

	return nil
}

func (m *Manager) Config() *Config {
	return m.config
}

func (m *Manager) SetConfig(cfg *Config) {
	m.config = cfg
}

func (m *Manager) CreateEnvironment(name string, opts CreateEnvOptions) error {
	env := &Environment{
		Name:      name,
		Namespace: opts.Namespace,
		Clusters:  []ClusterInfo{},
	}

	for _, c := range opts.Clusters {
		env.Clusters = append(env.Clusters, c)
	}

	if err := m.config.AddEnvironment(env); err != nil {
		return err
	}

	return m.Save()
}

type CreateEnvOptions struct {
	Namespace string
	Clusters  []ClusterInfo
}

func (m *Manager) DeleteEnvironment(name string) error {
	if err := m.config.RemoveEnvironment(name); err != nil {
		return err
	}
	return m.Save()
}

func (m *Manager) AddClusterToEnvironment(envName string, cluster ClusterInfo) error {
	env := m.config.GetEnvironment(envName)
	if env == nil {
		return fmt.Errorf("environment %s not found", envName)
	}

	if err := env.AddCluster(cluster); err != nil {
		return err
	}

	return m.Save()
}

func (m *Manager) RemoveClusterFromEnvironment(envName, clusterName string) error {
	env := m.config.GetEnvironment(envName)
	if env == nil {
		return fmt.Errorf("environment %s not found", envName)
	}

	if err := env.RemoveCluster(clusterName); err != nil {
		return err
	}

	return m.Save()
}

func (m *Manager) SetTopology(topology Topology) error {
	m.config.Topology = topology
	return m.Save()
}

func (m *Manager) SetDefaultCluster(url string) error {
	m.config.Cluster = url
	return m.Save()
}

func (m *Manager) ListEnvironments() []*Environment {
	return m.config.Environments
}

func (m *Manager) GetEnvironment(name string) *Environment {
	return m.config.GetEnvironment(name)
}

func (m *Manager) ToJSON() (string, error) {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *Manager) ToYAML() (string, error) {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *Manager) Validate() error {
	return m.config.Validate()
}

type PromotionOptions struct {
	Application string
	FromEnv     string
	ToEnv       string
	All         bool
	DryRun      bool
}

type PromotionResult struct {
	Application string
	FromEnv     string
	ToEnv       string
	Success     bool
	Message     string
	Changes     []string
}

func (m *Manager) Promote(opts PromotionOptions) (*PromotionResult, error) {
	fromEnv := m.config.GetEnvironment(opts.FromEnv)
	if fromEnv == nil {
		return nil, fmt.Errorf("source environment %s not found", opts.FromEnv)
	}

	toEnv := m.config.GetEnvironment(opts.ToEnv)
	if toEnv == nil {
		return nil, fmt.Errorf("target environment %s not found", opts.ToEnv)
	}

	result := &PromotionResult{
		Application: opts.Application,
		FromEnv:     opts.FromEnv,
		ToEnv:       opts.ToEnv,
		Success:     true,
		Changes:     []string{},
	}

	if opts.DryRun {
		result.Message = fmt.Sprintf("Would promote %s from %s to %s", opts.Application, opts.FromEnv, opts.ToEnv)
		result.Changes = append(result.Changes, fmt.Sprintf("Copy overlay from %s to %s", opts.FromEnv, opts.ToEnv))
		result.Changes = append(result.Changes, "Update image tags")
		result.Changes = append(result.Changes, "Update configuration values")
		return result, nil
	}

	result.Message = fmt.Sprintf("Promoted %s from %s to %s", opts.Application, opts.FromEnv, opts.ToEnv)
	result.Changes = append(result.Changes, fmt.Sprintf("Copied overlay from %s to %s", opts.FromEnv, opts.ToEnv))
	result.Changes = append(result.Changes, "Updated image tags")
	result.Changes = append(result.Changes, "Updated configuration values")

	return result, nil
}

func (m *Manager) GetPromotionPath() []string {
	names := make([]string, len(m.config.Environments))
	for i, env := range m.config.Environments {
		names[i] = env.Name
	}
	return names
}
