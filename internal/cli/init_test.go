package cli

import (
	"os"
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/progress"
)

func TestApplyFlagOverrides_GitURL(t *testing.T) {
	cfg := config.NewDefaultConfig()

	gitURL = "https://github.com/org/repo.git"
	defer func() { gitURL = "" }()

	applyFlagOverrides(cfg)

	if cfg.Git.URL != "https://github.com/org/repo.git" {
		t.Errorf("Git.URL = %v, want https://github.com/org/repo.git", cfg.Git.URL)
	}
	if cfg.Output.URL != "https://github.com/org/repo.git" {
		t.Errorf("Output.URL = %v, want https://github.com/org/repo.git", cfg.Output.URL)
	}
}

func TestApplyFlagOverrides_GitToken(t *testing.T) {
	cfg := config.NewDefaultConfig()

	gitToken = "test-token-12345"
	defer func() { gitToken = "" }()

	applyFlagOverrides(cfg)

	if cfg.Git.Auth.Token != "test-token-12345" {
		t.Errorf("Git.Auth.Token = %v, want test-token-12345", cfg.Git.Auth.Token)
	}
	if cfg.Git.Auth.Method != "token" {
		t.Errorf("Git.Auth.Method = %v, want token", cfg.Git.Auth.Method)
	}
}

func TestApplyFlagOverrides_GitTokenFromEnv(t *testing.T) {
	cfg := config.NewDefaultConfig()

	os.Setenv("GITOPSI_GIT_TOKEN", "env-token-12345")
	defer os.Unsetenv("GITOPSI_GIT_TOKEN")

	applyFlagOverrides(cfg)

	if cfg.Git.Auth.Token != "env-token-12345" {
		t.Errorf("Git.Auth.Token = %v, want env-token-12345", cfg.Git.Auth.Token)
	}
}

func TestApplyFlagOverrides_PushFlag(t *testing.T) {
	cfg := config.NewDefaultConfig()

	pushAfterInit = true
	defer func() { pushAfterInit = false }()

	applyFlagOverrides(cfg)

	if !cfg.Git.PushOnInit {
		t.Error("Git.PushOnInit should be true")
	}
}

func TestApplyFlagOverrides_ClusterURL(t *testing.T) {
	cfg := config.NewDefaultConfig()

	clusterURL = "https://api.cluster.example.com:6443"
	defer func() { clusterURL = "" }()

	applyFlagOverrides(cfg)

	if cfg.Cluster.URL != "https://api.cluster.example.com:6443" {
		t.Errorf("Cluster.URL = %v, want https://api.cluster.example.com:6443", cfg.Cluster.URL)
	}
}

func TestApplyFlagOverrides_ClusterToken(t *testing.T) {
	cfg := config.NewDefaultConfig()

	clusterToken = "cluster-token-12345"
	defer func() { clusterToken = "" }()

	applyFlagOverrides(cfg)

	if cfg.Cluster.Auth.Token != "cluster-token-12345" {
		t.Errorf("Cluster.Auth.Token = %v, want cluster-token-12345", cfg.Cluster.Auth.Token)
	}
	if cfg.Cluster.Auth.Method != "token" {
		t.Errorf("Cluster.Auth.Method = %v, want token", cfg.Cluster.Auth.Method)
	}
}

func TestApplyFlagOverrides_ClusterTokenFromEnv(t *testing.T) {
	cfg := config.NewDefaultConfig()

	os.Setenv("GITOPSI_CLUSTER_TOKEN", "env-cluster-token-12345")
	defer os.Unsetenv("GITOPSI_CLUSTER_TOKEN")

	applyFlagOverrides(cfg)

	if cfg.Cluster.Auth.Token != "env-cluster-token-12345" {
		t.Errorf("Cluster.Auth.Token = %v, want env-cluster-token-12345", cfg.Cluster.Auth.Token)
	}
}

func TestApplyFlagOverrides_Bootstrap(t *testing.T) {
	cfg := config.NewDefaultConfig()

	bootstrapFlag = true
	defer func() { bootstrapFlag = false }()

	applyFlagOverrides(cfg)

	if !cfg.Bootstrap.Enabled {
		t.Error("Bootstrap.Enabled should be true")
	}
}

func TestApplyFlagOverrides_BootstrapMode(t *testing.T) {
	cfg := config.NewDefaultConfig()

	bootstrapMode = "olm"
	defer func() { bootstrapMode = "" }()

	applyFlagOverrides(cfg)

	if cfg.Bootstrap.Mode != "olm" {
		t.Errorf("Bootstrap.Mode = %v, want olm", cfg.Bootstrap.Mode)
	}
}

func TestShouldPush(t *testing.T) {
	tests := []struct {
		name       string
		pushOnInit bool
		gitURL     string
		want       bool
	}{
		{
			name:       "push enabled with URL",
			pushOnInit: true,
			gitURL:     "https://github.com/org/repo.git",
			want:       true,
		},
		{
			name:       "push disabled",
			pushOnInit: false,
			gitURL:     "https://github.com/org/repo.git",
			want:       false,
		},
		{
			name:       "push enabled but no URL",
			pushOnInit: true,
			gitURL:     "",
			want:       false,
		},
		{
			name:       "both disabled",
			pushOnInit: false,
			gitURL:     "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.Git.PushOnInit = tt.pushOnInit
			cfg.Git.URL = tt.gitURL

			if got := shouldPush(cfg); got != tt.want {
				t.Errorf("shouldPush() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldBootstrap(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		clusterURL string
		want       bool
	}{
		{
			name:       "bootstrap enabled with cluster URL",
			enabled:    true,
			clusterURL: "https://api.cluster.example.com:6443",
			want:       true,
		},
		{
			name:       "bootstrap disabled",
			enabled:    false,
			clusterURL: "https://api.cluster.example.com:6443",
			want:       false,
		},
		{
			name:       "bootstrap enabled but no cluster URL",
			enabled:    true,
			clusterURL: "",
			want:       true, // cluster URL is auto-detected when bootstrap runs
		},
		{
			name:       "both disabled",
			enabled:    false,
			clusterURL: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.Bootstrap.Enabled = tt.enabled
			cfg.Cluster.URL = tt.clusterURL

			if got := shouldBootstrap(cfg); got != tt.want {
				t.Errorf("shouldBootstrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProgressSummary(t *testing.T) {
	tempDir := t.TempDir()

	summary := &progress.SetupSummary{
		Setup: progress.SetupInfo{
			Version: "test-project",
		},
		Git: progress.GitInfo{
			URL:    "https://github.com/org/repo.git",
			Branch: "main",
		},
	}

	// Save and load should work
	err := progress.SaveSummary(tempDir, summary)
	if err != nil {
		t.Fatalf("SaveSummary failed: %v", err)
	}

	loaded, err := progress.LoadSummary(tempDir)
	if err != nil {
		t.Fatalf("LoadSummary failed: %v", err)
	}

	if loaded.Git.URL != summary.Git.URL {
		t.Errorf("Expected Git URL %s, got %s", summary.Git.URL, loaded.Git.URL)
	}
}

func TestProgressSummary_WithGitPush(t *testing.T) {
	summary := &progress.SetupSummary{
		Setup: progress.SetupInfo{
			Version: "test-project",
		},
		Git: progress.GitInfo{
			URL:    "https://github.com/org/repo.git",
			Branch: "main",
			Status: "synced",
		},
	}

	p := progress.New("Test", "test-project")
	p.SetQuiet(true)

	// Should not panic
	p.ShowSummary(summary)
}

func TestProgressSummary_WithBootstrap(t *testing.T) {
	summary := &progress.SetupSummary{
		Setup: progress.SetupInfo{
			Version: "test-project",
		},
		GitOpsTool: progress.GitOpsToolInfo{
			Name:      "argocd",
			URL:       "https://argocd.example.com",
			Namespace: "argocd",
			Status:    "healthy",
		},
		Cluster: progress.ClusterInfo{
			URL:      "https://api.cluster.example.com:6443",
			Platform: "kubernetes",
			Status:   "connected",
		},
	}

	p := progress.New("Test", "test-project")
	p.SetQuiet(true)

	// Should not panic
	p.ShowSummary(summary)
}

func TestApplyFlagOverrides_AllFlags(t *testing.T) {
	cfg := config.NewDefaultConfig()

	// Set all flags
	gitURL = "https://github.com/org/repo.git"
	gitToken = "git-token"
	pushAfterInit = true
	clusterURL = "https://api.cluster.example.com:6443"
	clusterToken = "cluster-token"
	bootstrapFlag = true
	bootstrapMode = "manifest"

	defer func() {
		gitURL = ""
		gitToken = ""
		pushAfterInit = false
		clusterURL = ""
		clusterToken = ""
		bootstrapFlag = false
		bootstrapMode = ""
	}()

	applyFlagOverrides(cfg)

	// Verify all overrides
	if cfg.Git.URL != "https://github.com/org/repo.git" {
		t.Errorf("Git.URL = %v, want https://github.com/org/repo.git", cfg.Git.URL)
	}
	if cfg.Git.Auth.Token != "git-token" {
		t.Errorf("Git.Auth.Token = %v, want git-token", cfg.Git.Auth.Token)
	}
	if !cfg.Git.PushOnInit {
		t.Error("Git.PushOnInit should be true")
	}
	if cfg.Cluster.URL != "https://api.cluster.example.com:6443" {
		t.Errorf("Cluster.URL = %v, want https://api.cluster.example.com:6443", cfg.Cluster.URL)
	}
	if cfg.Cluster.Auth.Token != "cluster-token" {
		t.Errorf("Cluster.Auth.Token = %v, want cluster-token", cfg.Cluster.Auth.Token)
	}
	if !cfg.Bootstrap.Enabled {
		t.Error("Bootstrap.Enabled should be true")
	}
	if cfg.Bootstrap.Mode != "manifest" {
		t.Errorf("Bootstrap.Mode = %v, want manifest", cfg.Bootstrap.Mode)
	}
}
