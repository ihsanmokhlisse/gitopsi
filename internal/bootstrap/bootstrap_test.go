package bootstrap

import (
	"testing"
)

func TestToolConstants(t *testing.T) {
	tests := []struct {
		tool Tool
		want string
	}{
		{ToolArgoCD, "argocd"},
		{ToolFlux, "flux"},
	}

	for _, tt := range tests {
		if string(tt.tool) != tt.want {
			t.Errorf("Tool %v = %s, want %s", tt.tool, string(tt.tool), tt.want)
		}
	}
}

func TestModeConstants(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeHelm, "helm"},
		{ModeOLM, "olm"},
		{ModeManifest, "manifest"},
		{ModeKustomize, "kustomize"},
	}

	for _, tt := range tests {
		if string(tt.mode) != tt.want {
			t.Errorf("Mode %v = %s, want %s", tt.mode, string(tt.mode), tt.want)
		}
	}
}

func TestNew_Defaults(t *testing.T) {
	tests := []struct {
		name              string
		tool              Tool
		expectedNamespace string
	}{
		{
			name:              "ArgoCD default namespace",
			tool:              ToolArgoCD,
			expectedNamespace: "argocd",
		},
		{
			name:              "Flux default namespace",
			tool:              ToolFlux,
			expectedNamespace: "flux-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(nil, &Options{Tool: tt.tool})

			if b.GetNamespace() != tt.expectedNamespace {
				t.Errorf("GetNamespace() = %v, want %v", b.GetNamespace(), tt.expectedNamespace)
			}
			if b.GetTool() != tt.tool {
				t.Errorf("GetTool() = %v, want %v", b.GetTool(), tt.tool)
			}
		})
	}
}

func TestNew_CustomNamespace(t *testing.T) {
	b := New(nil, &Options{
		Tool:      ToolArgoCD,
		Namespace: "custom-namespace",
	})

	if b.GetNamespace() != "custom-namespace" {
		t.Errorf("GetNamespace() = %v, want custom-namespace", b.GetNamespace())
	}
}

func TestNew_DefaultTimeout(t *testing.T) {
	b := New(nil, &Options{Tool: ToolArgoCD})

	if b.options.Timeout != 300 {
		t.Errorf("Default timeout = %v, want 300", b.options.Timeout)
	}
}

func TestNew_CustomTimeout(t *testing.T) {
	b := New(nil, &Options{
		Tool:    ToolArgoCD,
		Timeout: 600,
	})

	if b.options.Timeout != 600 {
		t.Errorf("Timeout = %v, want 600", b.options.Timeout)
	}
}

func TestGetMode(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"helm mode", ModeHelm},
		{"olm mode", ModeOLM},
		{"manifest mode", ModeManifest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(nil, &Options{
				Tool: ToolArgoCD,
				Mode: tt.mode,
			})

			if b.GetMode() != tt.mode {
				t.Errorf("GetMode() = %v, want %v", b.GetMode(), tt.mode)
			}
		})
	}
}

func TestGetEnvFromToken(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		want     string
	}{
		{
			name:     "existing env var",
			envName:  "TEST_TOKEN_VAR",
			envValue: "test-token-value",
			want:     "test-token-value",
		},
		{
			name:    "non-existing env var",
			envName: "NON_EXISTING_VAR_12345",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envName, tt.envValue)
			}

			got := GetEnvFromToken(tt.envName)
			if got != tt.want {
				t.Errorf("GetEnvFromToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_AllFields(t *testing.T) {
	opts := &Options{
		Tool:            ToolArgoCD,
		Mode:            ModeHelm,
		Namespace:       "test-ns",
		Version:         "v2.9.0",
		Wait:            true,
		Timeout:         600,
		ConfigureRepo:   true,
		RepoURL:         "https://github.com/org/repo.git",
		RepoBranch:      "main",
		RepoPath:        "clusters/prod",
		CreateAppOfApps: true,
		SyncInitial:     true,
		ProjectName:     "my-project",
	}

	b := New(nil, opts)

	if b.options.Tool != ToolArgoCD {
		t.Errorf("Tool = %v, want %v", b.options.Tool, ToolArgoCD)
	}
	if b.options.Mode != ModeHelm {
		t.Errorf("Mode = %v, want %v", b.options.Mode, ModeHelm)
	}
	if b.options.Namespace != "test-ns" {
		t.Errorf("Namespace = %v, want test-ns", b.options.Namespace)
	}
	if b.options.Version != "v2.9.0" {
		t.Errorf("Version = %v, want v2.9.0", b.options.Version)
	}
	if !b.options.Wait {
		t.Error("Wait should be true")
	}
	if b.options.Timeout != 600 {
		t.Errorf("Timeout = %v, want 600", b.options.Timeout)
	}
	if !b.options.ConfigureRepo {
		t.Error("ConfigureRepo should be true")
	}
	if b.options.RepoURL != "https://github.com/org/repo.git" {
		t.Errorf("RepoURL = %v, want https://github.com/org/repo.git", b.options.RepoURL)
	}
	if b.options.RepoBranch != "main" {
		t.Errorf("RepoBranch = %v, want main", b.options.RepoBranch)
	}
	if b.options.RepoPath != "clusters/prod" {
		t.Errorf("RepoPath = %v, want clusters/prod", b.options.RepoPath)
	}
	if !b.options.CreateAppOfApps {
		t.Error("CreateAppOfApps should be true")
	}
	if !b.options.SyncInitial {
		t.Error("SyncInitial should be true")
	}
	if b.options.ProjectName != "my-project" {
		t.Errorf("ProjectName = %v, want my-project", b.options.ProjectName)
	}
}

func TestResult_Fields(t *testing.T) {
	result := &Result{
		Tool:      ToolArgoCD,
		URL:       "https://argocd.example.com",
		Username:  "admin",
		Password:  "secret123",
		Namespace: "argocd",
		Ready:     true,
		Message:   "ArgoCD installed successfully",
	}

	if result.Tool != ToolArgoCD {
		t.Errorf("Tool = %v, want %v", result.Tool, ToolArgoCD)
	}
	if result.URL != "https://argocd.example.com" {
		t.Errorf("URL = %v, want https://argocd.example.com", result.URL)
	}
	if result.Username != "admin" {
		t.Errorf("Username = %v, want admin", result.Username)
	}
	if result.Password != "secret123" {
		t.Errorf("Password = %v, want secret123", result.Password)
	}
	if result.Namespace != "argocd" {
		t.Errorf("Namespace = %v, want argocd", result.Namespace)
	}
	if !result.Ready {
		t.Error("Ready should be true")
	}
	if result.Message != "ArgoCD installed successfully" {
		t.Errorf("Message = %v, want 'ArgoCD installed successfully'", result.Message)
	}
}

func TestValidModes(t *testing.T) {
	tests := []struct {
		name          string
		tool          Tool
		platform      string
		expectedModes []Mode
		includesOLM   bool
	}{
		{
			name:          "ArgoCD on OpenShift includes OLM",
			tool:          ToolArgoCD,
			platform:      "openshift",
			expectedModes: []Mode{ModeHelm, ModeManifest, ModeKustomize, ModeOLM},
			includesOLM:   true,
		},
		{
			name:          "ArgoCD on Kubernetes excludes OLM",
			tool:          ToolArgoCD,
			platform:      "kubernetes",
			expectedModes: []Mode{ModeHelm, ModeManifest, ModeKustomize},
			includesOLM:   false,
		},
		{
			name:          "Flux on OpenShift excludes OLM",
			tool:          ToolFlux,
			platform:      "openshift",
			expectedModes: []Mode{ModeHelm, ModeManifest, ModeKustomize},
			includesOLM:   false,
		},
		{
			name:          "ArgoCD on EKS excludes OLM",
			tool:          ToolArgoCD,
			platform:      "eks",
			expectedModes: []Mode{ModeHelm, ModeManifest, ModeKustomize},
			includesOLM:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modes := ValidModes(tt.tool, tt.platform)

			hasOLM := false
			for _, m := range modes {
				if m == ModeOLM {
					hasOLM = true
					break
				}
			}

			if hasOLM != tt.includesOLM {
				t.Errorf("ValidModes() OLM presence = %v, want %v", hasOLM, tt.includesOLM)
			}

			if len(modes) < 3 {
				t.Errorf("Expected at least 3 modes, got %d", len(modes))
			}
		})
	}
}

func TestSuggestMode(t *testing.T) {
	tests := []struct {
		name     string
		tool     Tool
		platform string
		want     Mode
	}{
		{
			name:     "ArgoCD on OpenShift suggests OLM",
			tool:     ToolArgoCD,
			platform: "openshift",
			want:     ModeOLM,
		},
		{
			name:     "Flux on OpenShift suggests Helm",
			tool:     ToolFlux,
			platform: "openshift",
			want:     ModeHelm,
		},
		{
			name:     "ArgoCD on EKS suggests Helm",
			tool:     ToolArgoCD,
			platform: "eks",
			want:     ModeHelm,
		},
		{
			name:     "ArgoCD on AKS suggests Helm",
			tool:     ToolArgoCD,
			platform: "aks",
			want:     ModeHelm,
		},
		{
			name:     "ArgoCD on Kubernetes suggests Helm",
			tool:     ToolArgoCD,
			platform: "kubernetes",
			want:     ModeHelm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestMode(tt.tool, tt.platform)
			if got != tt.want {
				t.Errorf("SuggestMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		tool     Tool
		platform string
		want     bool
	}{
		{
			name:     "Helm is valid for ArgoCD on Kubernetes",
			mode:     ModeHelm,
			tool:     ToolArgoCD,
			platform: "kubernetes",
			want:     true,
		},
		{
			name:     "OLM is valid for ArgoCD on OpenShift",
			mode:     ModeOLM,
			tool:     ToolArgoCD,
			platform: "openshift",
			want:     true,
		},
		{
			name:     "OLM is not valid for ArgoCD on Kubernetes",
			mode:     ModeOLM,
			tool:     ToolArgoCD,
			platform: "kubernetes",
			want:     false,
		},
		{
			name:     "Kustomize is valid for Flux",
			mode:     ModeKustomize,
			tool:     ToolFlux,
			platform: "kubernetes",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMode(tt.mode, tt.tool, tt.platform)
			if got != tt.want {
				t.Errorf("IsValidMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModeDescription(t *testing.T) {
	tests := []struct {
		mode Mode
	}{
		{ModeHelm},
		{ModeOLM},
		{ModeManifest},
		{ModeKustomize},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			desc := ModeDescription(tt.mode)
			if desc == "" {
				t.Errorf("ModeDescription(%s) returned empty string", tt.mode)
			}
			if desc == string(tt.mode) && tt.mode != "" {
				t.Errorf("ModeDescription(%s) returned just the mode name", tt.mode)
			}
		})
	}
}

func TestDefaultHelmConfig(t *testing.T) {
	tests := []struct {
		tool          Tool
		expectedRepo  string
		expectedChart string
	}{
		{
			tool:          ToolArgoCD,
			expectedRepo:  "https://argoproj.github.io/argo-helm",
			expectedChart: "argo-cd",
		},
		{
			tool:          ToolFlux,
			expectedRepo:  "https://fluxcd-community.github.io/helm-charts",
			expectedChart: "flux2",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.tool), func(t *testing.T) {
			cfg := DefaultHelmConfig(tt.tool)
			if cfg == nil {
				t.Fatal("DefaultHelmConfig returned nil")
			}
			if cfg.Repo != tt.expectedRepo {
				t.Errorf("Repo = %v, want %v", cfg.Repo, tt.expectedRepo)
			}
			if cfg.Chart != tt.expectedChart {
				t.Errorf("Chart = %v, want %v", cfg.Chart, tt.expectedChart)
			}
		})
	}
}

func TestDefaultOLMConfig(t *testing.T) {
	cfg := DefaultOLMConfig()
	if cfg == nil {
		t.Fatal("DefaultOLMConfig returned nil")
	}
	if cfg.Channel == "" {
		t.Error("Channel should not be empty")
	}
	if cfg.Source == "" {
		t.Error("Source should not be empty")
	}
	if cfg.SourceNamespace == "" {
		t.Error("SourceNamespace should not be empty")
	}
	if cfg.Approval == "" {
		t.Error("Approval should not be empty")
	}
}

func TestGetMode_AllModes(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"helm mode", ModeHelm},
		{"olm mode", ModeOLM},
		{"manifest mode", ModeManifest},
		{"kustomize mode", ModeKustomize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(nil, &Options{
				Tool: ToolArgoCD,
				Mode: tt.mode,
			})

			if b.GetMode() != tt.mode {
				t.Errorf("GetMode() = %v, want %v", b.GetMode(), tt.mode)
			}
		})
	}
}

func TestHelmConfig_Fields(t *testing.T) {
	cfg := &HelmConfig{
		Repo:    "https://charts.example.com",
		Chart:   "my-chart",
		Version: "1.0.0",
		Values:  map[string]any{"key": "value"},
		SetValues: map[string]string{
			"image.tag": "latest",
		},
	}

	if cfg.Repo != "https://charts.example.com" {
		t.Errorf("Repo = %v, want https://charts.example.com", cfg.Repo)
	}
	if cfg.Chart != "my-chart" {
		t.Errorf("Chart = %v, want my-chart", cfg.Chart)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", cfg.Version)
	}
	if cfg.Values["key"] != "value" {
		t.Errorf("Values[key] = %v, want value", cfg.Values["key"])
	}
	if cfg.SetValues["image.tag"] != "latest" {
		t.Errorf("SetValues[image.tag] = %v, want latest", cfg.SetValues["image.tag"])
	}
}

func TestOLMConfig_Fields(t *testing.T) {
	cfg := &OLMConfig{
		Channel:         "stable",
		Source:          "community-operators",
		SourceNamespace: "openshift-marketplace",
		Approval:        "Manual",
	}

	if cfg.Channel != "stable" {
		t.Errorf("Channel = %v, want stable", cfg.Channel)
	}
	if cfg.Source != "community-operators" {
		t.Errorf("Source = %v, want community-operators", cfg.Source)
	}
	if cfg.SourceNamespace != "openshift-marketplace" {
		t.Errorf("SourceNamespace = %v, want openshift-marketplace", cfg.SourceNamespace)
	}
	if cfg.Approval != "Manual" {
		t.Errorf("Approval = %v, want Manual", cfg.Approval)
	}
}

func TestKustomizeConfig_Fields(t *testing.T) {
	cfg := &KustomizeConfig{
		URL:     "https://github.com/example/repo/manifests",
		Path:    "overlays/prod",
		Patches: []string{"patch1.yaml", "patch2.yaml"},
	}

	if cfg.URL != "https://github.com/example/repo/manifests" {
		t.Errorf("URL = %v, want https://github.com/example/repo/manifests", cfg.URL)
	}
	if cfg.Path != "overlays/prod" {
		t.Errorf("Path = %v, want overlays/prod", cfg.Path)
	}
	if len(cfg.Patches) != 2 {
		t.Errorf("Patches length = %v, want 2", len(cfg.Patches))
	}
}

func TestManifestConfig_Fields(t *testing.T) {
	cfg := &ManifestConfig{
		URL:       "https://raw.github.com/example/install.yaml",
		Paths:     []string{"/path/to/manifest1.yaml", "/path/to/manifest2.yaml"},
		Namespace: "custom-ns",
	}

	if cfg.URL != "https://raw.github.com/example/install.yaml" {
		t.Errorf("URL = %v, want https://raw.github.com/example/install.yaml", cfg.URL)
	}
	if len(cfg.Paths) != 2 {
		t.Errorf("Paths length = %v, want 2", len(cfg.Paths))
	}
	if cfg.Namespace != "custom-ns" {
		t.Errorf("Namespace = %v, want custom-ns", cfg.Namespace)
	}
}
