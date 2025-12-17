package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopology_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		topology Topology
		want     bool
	}{
		{"namespace-based", TopologyNamespaceBased, true},
		{"cluster-per-env", TopologyClusterPerEnv, true},
		{"multi-cluster", TopologyMultiCluster, true},
		{"invalid", Topology("invalid"), false},
		{"empty", Topology(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.topology.IsValid())
		})
	}
}

func TestParseTopology(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Topology
		wantErr bool
	}{
		{"namespace-based", "namespace-based", TopologyNamespaceBased, false},
		{"cluster-per-env", "cluster-per-env", TopologyClusterPerEnv, false},
		{"multi-cluster", "multi-cluster", TopologyMultiCluster, false},
		{"invalid", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTopology(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidTopologies(t *testing.T) {
	topologies := ValidTopologies()
	assert.Len(t, topologies, 3)
	assert.Contains(t, topologies, "namespace-based")
	assert.Contains(t, topologies, "cluster-per-env")
	assert.Contains(t, topologies, "multi-cluster")
}

func TestEnvironment_GetNamespace(t *testing.T) {
	tests := []struct {
		name        string
		env         Environment
		projectName string
		want        string
	}{
		{
			name:        "with namespace set",
			env:         Environment{Name: "dev", Namespace: "custom-ns"},
			projectName: "myproject",
			want:        "custom-ns",
		},
		{
			name:        "without namespace set",
			env:         Environment{Name: "dev"},
			projectName: "myproject",
			want:        "myproject-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.env.GetNamespace(tt.projectName))
		})
	}
}

func TestEnvironment_GetPrimaryCluster(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
		want *ClusterInfo
	}{
		{
			name: "no clusters",
			env:  Environment{Name: "dev", Clusters: []ClusterInfo{}},
			want: nil,
		},
		{
			name: "single cluster",
			env: Environment{
				Name: "dev",
				Clusters: []ClusterInfo{
					{Name: "dev-cluster", URL: "https://dev.k8s"},
				},
			},
			want: &ClusterInfo{Name: "dev-cluster", URL: "https://dev.k8s"},
		},
		{
			name: "multiple clusters with primary",
			env: Environment{
				Name: "prod",
				Clusters: []ClusterInfo{
					{Name: "prod-us", URL: "https://us.k8s"},
					{Name: "prod-eu", URL: "https://eu.k8s", Primary: true},
				},
			},
			want: &ClusterInfo{Name: "prod-eu", URL: "https://eu.k8s", Primary: true},
		},
		{
			name: "multiple clusters without primary",
			env: Environment{
				Name: "prod",
				Clusters: []ClusterInfo{
					{Name: "prod-us", URL: "https://us.k8s"},
					{Name: "prod-eu", URL: "https://eu.k8s"},
				},
			},
			want: &ClusterInfo{Name: "prod-us", URL: "https://us.k8s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.env.GetPrimaryCluster()
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.URL, got.URL)
			}
		})
	}
}

func TestEnvironment_HasCluster(t *testing.T) {
	env := Environment{
		Name: "dev",
		Clusters: []ClusterInfo{
			{Name: "dev-cluster", URL: "https://dev.k8s"},
		},
	}

	assert.True(t, env.HasCluster("dev-cluster"))
	assert.False(t, env.HasCluster("prod-cluster"))
}

func TestEnvironment_AddCluster(t *testing.T) {
	env := &Environment{Name: "dev", Clusters: []ClusterInfo{}}

	err := env.AddCluster(ClusterInfo{Name: "cluster1", URL: "https://c1.k8s"})
	require.NoError(t, err)
	assert.Len(t, env.Clusters, 1)

	err = env.AddCluster(ClusterInfo{Name: "cluster1", URL: "https://c1.k8s"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestEnvironment_RemoveCluster(t *testing.T) {
	env := &Environment{
		Name: "dev",
		Clusters: []ClusterInfo{
			{Name: "cluster1", URL: "https://c1.k8s"},
			{Name: "cluster2", URL: "https://c2.k8s"},
		},
	}

	err := env.RemoveCluster("cluster1")
	require.NoError(t, err)
	assert.Len(t, env.Clusters, 1)
	assert.Equal(t, "cluster2", env.Clusters[0].Name)

	err = env.RemoveCluster("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConfig_GetEnvironment(t *testing.T) {
	cfg := &Config{
		Environments: []*Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
	}

	assert.NotNil(t, cfg.GetEnvironment("dev"))
	assert.NotNil(t, cfg.GetEnvironment("prod"))
	assert.Nil(t, cfg.GetEnvironment("staging"))
}

func TestConfig_AddEnvironment(t *testing.T) {
	cfg := NewConfig()

	err := cfg.AddEnvironment(&Environment{Name: "dev"})
	require.NoError(t, err)
	assert.Len(t, cfg.Environments, 1)

	err = cfg.AddEnvironment(&Environment{Name: "dev"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfig_RemoveEnvironment(t *testing.T) {
	cfg := &Config{
		Environments: []*Environment{
			{Name: "dev"},
			{Name: "prod"},
		},
	}

	err := cfg.RemoveEnvironment("dev")
	require.NoError(t, err)
	assert.Len(t, cfg.Environments, 1)

	err = cfg.RemoveEnvironment("nonexistent")
	assert.Error(t, err)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid namespace-based",
			cfg: &Config{
				Topology: TopologyNamespaceBased,
				Cluster:  "https://k8s.local",
				Environments: []*Environment{
					{Name: "dev"},
				},
			},
			wantErr: false,
		},
		{
			name: "namespace-based without cluster",
			cfg: &Config{
				Topology: TopologyNamespaceBased,
				Environments: []*Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
			errMsg:  "cluster URL is required",
		},
		{
			name: "cluster-per-env without clusters",
			cfg: &Config{
				Topology: TopologyClusterPerEnv,
				Environments: []*Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
			errMsg:  "at least one cluster is required",
		},
		{
			name: "valid cluster-per-env",
			cfg: &Config{
				Topology: TopologyClusterPerEnv,
				Environments: []*Environment{
					{Name: "dev", Clusters: []ClusterInfo{{Name: "dev", URL: "https://dev.k8s"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "no environments",
			cfg: &Config{
				Topology:     TopologyNamespaceBased,
				Cluster:      "https://k8s.local",
				Environments: []*Environment{},
			},
			wantErr: true,
			errMsg:  "at least one environment is required",
		},
		{
			name: "duplicate environment names",
			cfg: &Config{
				Topology: TopologyNamespaceBased,
				Cluster:  "https://k8s.local",
				Environments: []*Environment{
					{Name: "dev"},
					{Name: "dev"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate environment name",
		},
		{
			name: "invalid topology",
			cfg: &Config{
				Topology: Topology("invalid"),
				Environments: []*Environment{
					{Name: "dev"},
				},
			},
			wantErr: true,
			errMsg:  "invalid topology",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetAllClusters(t *testing.T) {
	cfg := &Config{
		Cluster: "https://default.k8s",
		Environments: []*Environment{
			{
				Name: "dev",
				Clusters: []ClusterInfo{
					{Name: "dev", URL: "https://dev.k8s"},
				},
			},
			{
				Name: "prod",
				Clusters: []ClusterInfo{
					{Name: "prod-us", URL: "https://us.k8s"},
					{Name: "prod-eu", URL: "https://eu.k8s"},
				},
			},
		},
	}

	clusters := cfg.GetAllClusters()
	assert.Len(t, clusters, 4)
}

func TestConfig_Summary(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "namespace-based",
			cfg: &Config{
				Topology: TopologyNamespaceBased,
				Cluster:  "https://k8s.local",
				Environments: []*Environment{
					{Name: "dev"},
					{Name: "prod"},
				},
			},
			want: "Single cluster (https://k8s.local) with 2 environments (namespace-based)",
		},
		{
			name: "cluster-per-env",
			cfg: &Config{
				Topology: TopologyClusterPerEnv,
				Environments: []*Environment{
					{Name: "dev"},
					{Name: "prod"},
				},
			},
			want: "2 environments with dedicated clusters",
		},
		{
			name: "multi-cluster",
			cfg: &Config{
				Topology: TopologyMultiCluster,
				Environments: []*Environment{
					{Name: "prod", Clusters: []ClusterInfo{{Name: "us"}, {Name: "eu"}}},
				},
			},
			want: "1 environments across 2 clusters (multi-cluster)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.Summary())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("myproject")

	assert.Equal(t, TopologyNamespaceBased, cfg.Topology)
	assert.Len(t, cfg.Environments, 3)
	assert.Equal(t, "dev", cfg.Environments[0].Name)
	assert.Equal(t, "myproject-dev", cfg.Environments[0].Namespace)
}
