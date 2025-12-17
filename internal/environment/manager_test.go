package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager("/tmp/test-project")
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.config)
	assert.Equal(t, "/tmp/test-project", mgr.projectPath)
}

func TestManager_SaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.config = &Config{
		Topology: TopologyClusterPerEnv,
		Environments: []*Environment{
			{Name: "dev", Clusters: []ClusterInfo{{Name: "dev-cluster", URL: "https://dev.k8s"}}},
			{Name: "prod", Clusters: []ClusterInfo{{Name: "prod-cluster", URL: "https://prod.k8s"}}},
		},
	}

	err = mgr.Save()
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, EnvConfigFile)
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	newMgr := NewManager(tmpDir)
	err = newMgr.Load()
	require.NoError(t, err)

	assert.Equal(t, TopologyClusterPerEnv, newMgr.config.Topology)
	assert.Len(t, newMgr.config.Environments, 2)
}

func TestManager_LoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	err = mgr.Load()
	assert.NoError(t, err)
}

func TestManager_CreateEnvironment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	err = mgr.CreateEnvironment("dev", CreateEnvOptions{
		Namespace: "myproject-dev",
		Clusters: []ClusterInfo{
			{Name: "dev-cluster", URL: "https://dev.k8s"},
		},
	})
	require.NoError(t, err)

	assert.True(t, mgr.config.HasEnvironment("dev"))
	env := mgr.config.GetEnvironment("dev")
	assert.Equal(t, "myproject-dev", env.Namespace)
	assert.Len(t, env.Clusters, 1)
}

func TestManager_DeleteEnvironment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.config.Environments = []*Environment{{Name: "dev"}, {Name: "prod"}}

	err = mgr.DeleteEnvironment("dev")
	require.NoError(t, err)
	assert.Len(t, mgr.config.Environments, 1)
	assert.False(t, mgr.config.HasEnvironment("dev"))
}

func TestManager_AddClusterToEnvironment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.config.Environments = []*Environment{{Name: "prod", Clusters: []ClusterInfo{}}}

	err = mgr.AddClusterToEnvironment("prod", ClusterInfo{
		Name:   "prod-eu",
		URL:    "https://eu.k8s",
		Region: "eu-west-1",
	})
	require.NoError(t, err)

	env := mgr.config.GetEnvironment("prod")
	assert.Len(t, env.Clusters, 1)
	assert.Equal(t, "prod-eu", env.Clusters[0].Name)
}

func TestManager_AddClusterToNonExistentEnv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	err = mgr.AddClusterToEnvironment("nonexistent", ClusterInfo{Name: "cluster", URL: "https://k8s"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_RemoveClusterFromEnvironment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.config.Environments = []*Environment{{
		Name: "prod",
		Clusters: []ClusterInfo{
			{Name: "prod-us", URL: "https://us.k8s"},
			{Name: "prod-eu", URL: "https://eu.k8s"},
		},
	}}

	err = mgr.RemoveClusterFromEnvironment("prod", "prod-us")
	require.NoError(t, err)

	env := mgr.config.GetEnvironment("prod")
	assert.Len(t, env.Clusters, 1)
	assert.Equal(t, "prod-eu", env.Clusters[0].Name)
}

func TestManager_SetTopology(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	err = mgr.SetTopology(TopologyMultiCluster)
	require.NoError(t, err)
	assert.Equal(t, TopologyMultiCluster, mgr.config.Topology)
}

func TestManager_SetDefaultCluster(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	err = mgr.SetDefaultCluster("https://k8s.local")
	require.NoError(t, err)
	assert.Equal(t, "https://k8s.local", mgr.config.Cluster)
}

func TestManager_ListEnvironments(t *testing.T) {
	mgr := NewManager("/tmp")
	mgr.config.Environments = []*Environment{
		{Name: "dev"},
		{Name: "staging"},
		{Name: "prod"},
	}

	envs := mgr.ListEnvironments()
	assert.Len(t, envs, 3)
}

func TestManager_ToJSON(t *testing.T) {
	mgr := NewManager("/tmp")
	mgr.config = &Config{
		Topology: TopologyNamespaceBased,
		Cluster:  "https://k8s.local",
		Environments: []*Environment{
			{Name: "dev"},
		},
	}

	json, err := mgr.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, json, "namespace-based")
	assert.Contains(t, json, "https://k8s.local")
}

func TestManager_ToYAML(t *testing.T) {
	mgr := NewManager("/tmp")
	mgr.config = &Config{
		Topology: TopologyNamespaceBased,
		Cluster:  "https://k8s.local",
		Environments: []*Environment{
			{Name: "dev"},
		},
	}

	yaml, err := mgr.ToYAML()
	require.NoError(t, err)
	assert.Contains(t, yaml, "topology: namespace-based")
	assert.Contains(t, yaml, "cluster: https://k8s.local")
}

func TestManager_Promote(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.config = &Config{
		Environments: []*Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
	}

	result, err := mgr.Promote(PromotionOptions{
		Application: "myapp",
		FromEnv:     "dev",
		ToEnv:       "staging",
		DryRun:      true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "Would promote")
	assert.NotEmpty(t, result.Changes)
}

func TestManager_PromoteInvalidEnv(t *testing.T) {
	mgr := NewManager("/tmp")
	mgr.config = &Config{
		Environments: []*Environment{
			{Name: "dev"},
		},
	}

	_, err := mgr.Promote(PromotionOptions{
		Application: "myapp",
		FromEnv:     "dev",
		ToEnv:       "nonexistent",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_GetPromotionPath(t *testing.T) {
	mgr := NewManager("/tmp")
	mgr.config = &Config{
		Environments: []*Environment{
			{Name: "dev"},
			{Name: "staging"},
			{Name: "prod"},
		},
	}

	path := mgr.GetPromotionPath()
	assert.Equal(t, []string{"dev", "staging", "prod"}, path)
}

