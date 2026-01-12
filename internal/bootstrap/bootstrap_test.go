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

func TestGetArgoCDHelmConfig_Default(t *testing.T) {
	b := New(nil, &Options{
		Tool:    ToolArgoCD,
		Mode:    ModeHelm,
		Version: "2.9.0",
	})

	cfg := b.getArgoCDHelmConfig()
	if cfg.Repo != "https://argoproj.github.io/argo-helm" {
		t.Errorf("Default Repo = %s, want https://argoproj.github.io/argo-helm", cfg.Repo)
	}
	if cfg.Chart != "argo-cd" {
		t.Errorf("Default Chart = %s, want argo-cd", cfg.Chart)
	}
	if cfg.Version != "2.9.0" {
		t.Errorf("Version = %s, want 2.9.0", cfg.Version)
	}
}

func TestGetArgoCDHelmConfig_Custom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			Repo:    "https://custom.repo.io",
			Chart:   "custom-argocd",
			Version: "1.0.0",
		},
	})

	cfg := b.getArgoCDHelmConfig()
	if cfg.Repo != "https://custom.repo.io" {
		t.Errorf("Custom Repo = %s, want https://custom.repo.io", cfg.Repo)
	}
	if cfg.Chart != "custom-argocd" {
		t.Errorf("Custom Chart = %s, want custom-argocd", cfg.Chart)
	}
}

func TestGetArgoCDHelmConfig_PartialCustom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			// Repo and Chart are empty - should use defaults
		},
	})

	cfg := b.getArgoCDHelmConfig()
	if cfg.Repo != "https://argoproj.github.io/argo-helm" {
		t.Errorf("Repo should use default, got %s", cfg.Repo)
	}
	if cfg.Chart != "argo-cd" {
		t.Errorf("Chart should use default, got %s", cfg.Chart)
	}
}

func TestGetFluxHelmConfig_Default(t *testing.T) {
	b := New(nil, &Options{
		Tool:    ToolFlux,
		Mode:    ModeHelm,
		Version: "2.0.0",
	})

	cfg := b.getFluxHelmConfig()
	if cfg.Repo != "https://fluxcd-community.github.io/helm-charts" {
		t.Errorf("Default Repo = %s, want https://fluxcd-community.github.io/helm-charts", cfg.Repo)
	}
	if cfg.Chart != "flux2" {
		t.Errorf("Default Chart = %s, want flux2", cfg.Chart)
	}
}

func TestGetFluxHelmConfig_Custom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolFlux,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			Repo:    "https://custom.flux.io",
			Chart:   "custom-flux",
			Version: "2.0.0",
		},
	})

	cfg := b.getFluxHelmConfig()
	if cfg.Repo != "https://custom.flux.io" {
		t.Errorf("Custom Repo = %s, want https://custom.flux.io", cfg.Repo)
	}
	if cfg.Chart != "custom-flux" {
		t.Errorf("Custom Chart = %s, want custom-flux", cfg.Chart)
	}
}

func TestGetFluxHelmConfig_PartialCustom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolFlux,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			// Empty - should use defaults
		},
	})

	cfg := b.getFluxHelmConfig()
	if cfg.Repo != "https://fluxcd-community.github.io/helm-charts" {
		t.Errorf("Repo should use default, got %s", cfg.Repo)
	}
	if cfg.Chart != "flux2" {
		t.Errorf("Chart should use default, got %s", cfg.Chart)
	}
}

func TestGetArgoCDOLMConfig_Default(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeOLM,
	})

	cfg := b.getArgoCDOLMConfig()
	if cfg.Channel != "alpha" {
		t.Errorf("Default Channel = %s, want alpha", cfg.Channel)
	}
	if cfg.Source != "community-operators" {
		t.Errorf("Default Source = %s, want community-operators", cfg.Source)
	}
	if cfg.SourceNamespace != "openshift-marketplace" {
		t.Errorf("Default SourceNamespace = %s, want openshift-marketplace", cfg.SourceNamespace)
	}
	if cfg.Approval != "Automatic" {
		t.Errorf("Default Approval = %s, want Automatic", cfg.Approval)
	}
}

func TestGetArgoCDOLMConfig_Custom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeOLM,
		OLM: &OLMConfig{
			Channel:         "stable",
			Source:          "custom-catalog",
			SourceNamespace: "custom-ns",
			Approval:        "Manual",
		},
	})

	cfg := b.getArgoCDOLMConfig()
	if cfg.Channel != "stable" {
		t.Errorf("Custom Channel = %s, want stable", cfg.Channel)
	}
	if cfg.Source != "custom-catalog" {
		t.Errorf("Custom Source = %s, want custom-catalog", cfg.Source)
	}
}

func TestGetArgoCDOLMConfig_PartialCustom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeOLM,
		OLM:  &OLMConfig{
			// All empty - should use defaults
		},
	})

	cfg := b.getArgoCDOLMConfig()
	if cfg.Channel != "alpha" {
		t.Errorf("Channel should use default, got %s", cfg.Channel)
	}
	if cfg.Source != "community-operators" {
		t.Errorf("Source should use default, got %s", cfg.Source)
	}
}

func TestGetArgoCDManifestConfig_Default(t *testing.T) {
	b := New(nil, &Options{
		Tool:    ToolArgoCD,
		Mode:    ModeManifest,
		Version: "v2.9.0",
	})

	cfg := b.getArgoCDManifestConfig()
	expectedURL := "https://raw.githubusercontent.com/argoproj/argo-cd/v2.9.0/manifests/install.yaml"
	if cfg.URL != expectedURL {
		t.Errorf("URL = %s, want %s", cfg.URL, expectedURL)
	}
}

func TestGetArgoCDManifestConfig_DefaultVersion(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeManifest,
		// No version specified
	})

	cfg := b.getArgoCDManifestConfig()
	expectedURL := "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"
	if cfg.URL != expectedURL {
		t.Errorf("URL = %s, want %s", cfg.URL, expectedURL)
	}
}

func TestGetArgoCDManifestConfig_Custom(t *testing.T) {
	customURL := "https://example.com/argocd/install.yaml"
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeManifest,
		Manifest: &ManifestConfig{
			URL: customURL,
		},
	})

	cfg := b.getArgoCDManifestConfig()
	if cfg.URL != customURL {
		t.Errorf("URL = %s, want %s", cfg.URL, customURL)
	}
}

func TestGetArgoCDKustomizeConfig_Default(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeKustomize,
	})

	cfg := b.getArgoCDKustomizeConfig()
	if cfg.URL != "https://github.com/argoproj/argo-cd/manifests/cluster-install" {
		t.Errorf("URL = %s, want default ArgoCD kustomize URL", cfg.URL)
	}
	if cfg.Path != "cluster-install" {
		t.Errorf("Path = %s, want cluster-install", cfg.Path)
	}
}

func TestGetArgoCDKustomizeConfig_Custom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeKustomize,
		Kustomize: &KustomizeConfig{
			URL:  "https://custom.kustomize.io/argocd",
			Path: "overlays/prod",
		},
	})

	cfg := b.getArgoCDKustomizeConfig()
	if cfg.URL != "https://custom.kustomize.io/argocd" {
		t.Errorf("URL = %s, want https://custom.kustomize.io/argocd", cfg.URL)
	}
}

func TestGetFluxKustomizeConfig_Default(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolFlux,
		Mode: ModeKustomize,
	})

	cfg := b.getFluxKustomizeConfig()
	if cfg.URL != "https://github.com/fluxcd/flux2/manifests/install" {
		t.Errorf("URL = %s, want default Flux kustomize URL", cfg.URL)
	}
	if cfg.Path != "install" {
		t.Errorf("Path = %s, want install", cfg.Path)
	}
}

func TestGetFluxKustomizeConfig_Custom(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolFlux,
		Mode: ModeKustomize,
		Kustomize: &KustomizeConfig{
			URL:  "https://custom.kustomize.io/flux",
			Path: "overlays/staging",
		},
	})

	cfg := b.getFluxKustomizeConfig()
	if cfg.URL != "https://custom.kustomize.io/flux" {
		t.Errorf("URL = %s, want https://custom.kustomize.io/flux", cfg.URL)
	}
}

func TestDefaultHelmConfig_UnknownTool(t *testing.T) {
	cfg := DefaultHelmConfig(Tool("unknown"))
	if cfg != nil {
		t.Errorf("DefaultHelmConfig should return nil for unknown tool, got %v", cfg)
	}
}

func TestModeDescription_Unknown(t *testing.T) {
	desc := ModeDescription(Mode("unknown"))
	// Unknown mode returns just the mode name
	if desc != "unknown" {
		t.Errorf("ModeDescription for unknown = %s, want 'unknown'", desc)
	}
}

func TestSuggestMode_GKE(t *testing.T) {
	mode := SuggestMode(ToolArgoCD, "gke")
	if mode != ModeHelm {
		t.Errorf("SuggestMode for GKE = %v, want ModeHelm", mode)
	}
}

func TestBootstrapper_Nil_Cluster(t *testing.T) {
	// Should not panic with nil cluster
	b := New(nil, &Options{
		Tool:      ToolArgoCD,
		Namespace: "argocd",
	})

	if b.GetTool() != ToolArgoCD {
		t.Errorf("GetTool() = %v, want %v", b.GetTool(), ToolArgoCD)
	}
	if b.cluster != nil {
		t.Error("cluster should be nil")
	}
}

func TestOptions_WithModeConfigs(t *testing.T) {
	opts := &Options{
		Tool: ToolArgoCD,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			Repo:  "https://custom.repo.io",
			Chart: "custom-chart",
		},
		OLM: &OLMConfig{
			Channel: "stable",
		},
		Manifest: &ManifestConfig{
			URL: "https://custom.manifest.io",
		},
		Kustomize: &KustomizeConfig{
			URL: "https://custom.kustomize.io",
		},
	}

	b := New(nil, opts)

	if b.options.Helm == nil {
		t.Error("Helm config should not be nil")
	}
	if b.options.OLM == nil {
		t.Error("OLM config should not be nil")
	}
	if b.options.Manifest == nil {
		t.Error("Manifest config should not be nil")
	}
	if b.options.Kustomize == nil {
		t.Error("Kustomize config should not be nil")
	}
}

func TestNamespace_Preserved(t *testing.T) {
	b := New(nil, &Options{
		Tool:      ToolArgoCD,
		Namespace: "openshift-gitops",
	})

	if b.GetNamespace() != "openshift-gitops" {
		t.Errorf("Namespace = %s, want openshift-gitops", b.GetNamespace())
	}
}

func TestValidModes_AKS(t *testing.T) {
	modes := ValidModes(ToolArgoCD, "aks")
	hasOLM := false
	for _, m := range modes {
		if m == ModeOLM {
			hasOLM = true
			break
		}
	}
	if hasOLM {
		t.Error("ValidModes for AKS should not include OLM")
	}
}

func TestValidModes_FluxOnOpenShift(t *testing.T) {
	modes := ValidModes(ToolFlux, "openshift")
	hasOLM := false
	for _, m := range modes {
		if m == ModeOLM {
			hasOLM = true
			break
		}
	}
	if hasOLM {
		t.Error("ValidModes for Flux on OpenShift should not include OLM")
	}
}

func TestIsValidMode_ManifestOnEKS(t *testing.T) {
	if !IsValidMode(ModeManifest, ToolArgoCD, "eks") {
		t.Error("Manifest mode should be valid for ArgoCD on EKS")
	}
}

func TestIsValidMode_KustomizeOnAKS(t *testing.T) {
	if !IsValidMode(ModeKustomize, ToolArgoCD, "aks") {
		t.Error("Kustomize mode should be valid for ArgoCD on AKS")
	}
}

func TestIsValidMode_InvalidMode(t *testing.T) {
	if IsValidMode(Mode("invalid"), ToolArgoCD, "kubernetes") {
		t.Error("Invalid mode should not be valid")
	}
}

func TestModeDescription_EmptyMode(t *testing.T) {
	desc := ModeDescription(Mode(""))
	if desc != "" {
		t.Errorf("ModeDescription for empty mode = %s, want empty string", desc)
	}
}

func TestDefaultHelmConfig_EmptyTool(t *testing.T) {
	cfg := DefaultHelmConfig(Tool(""))
	if cfg != nil {
		t.Error("DefaultHelmConfig for empty tool should return nil")
	}
}

func TestNew_ZeroTimeout(t *testing.T) {
	b := New(nil, &Options{
		Tool:    ToolFlux,
		Timeout: 0,
	})

	if b.options.Timeout != 300 {
		t.Errorf("Zero timeout should default to 300, got %d", b.options.Timeout)
	}
}

func TestNew_NegativeTimeout(t *testing.T) {
	b := New(nil, &Options{
		Tool:    ToolArgoCD,
		Timeout: -1,
	})

	// Negative timeout is preserved (no validation in New)
	if b.options.Timeout != -1 {
		t.Errorf("Negative timeout = %d, want -1", b.options.Timeout)
	}
}

func TestOptions_EmptyRepoURL(t *testing.T) {
	opts := &Options{
		Tool:          ToolArgoCD,
		ConfigureRepo: true,
		RepoURL:       "",
	}

	b := New(nil, opts)
	if b.options.RepoURL != "" {
		t.Errorf("RepoURL = %s, want empty", b.options.RepoURL)
	}
}

func TestOptions_RepoBranchDefault(t *testing.T) {
	opts := &Options{
		Tool:       ToolArgoCD,
		RepoBranch: "",
	}

	b := New(nil, opts)
	// RepoBranch defaults are handled at runtime in createArgoCDAppOfApps
	if b.options.RepoBranch != "" {
		t.Errorf("RepoBranch = %s, want empty (defaults applied at runtime)", b.options.RepoBranch)
	}
}

func TestHelmConfig_EmptySetValues(t *testing.T) {
	cfg := &HelmConfig{
		Repo:      "https://charts.example.com",
		Chart:     "my-chart",
		SetValues: nil,
	}

	if cfg.SetValues != nil {
		t.Error("SetValues should be nil")
	}
}

func TestHelmConfig_EmptyValues(t *testing.T) {
	cfg := &HelmConfig{
		Repo:   "https://charts.example.com",
		Chart:  "my-chart",
		Values: nil,
	}

	if cfg.Values != nil {
		t.Error("Values should be nil")
	}
}

func TestManifestConfig_EmptyPaths(t *testing.T) {
	cfg := &ManifestConfig{
		URL:   "https://example.com/install.yaml",
		Paths: nil,
	}

	if cfg.Paths != nil {
		t.Error("Paths should be nil")
	}
}

func TestKustomizeConfig_EmptyPatches(t *testing.T) {
	cfg := &KustomizeConfig{
		URL:     "https://github.com/example/repo",
		Path:    "overlays/prod",
		Patches: nil,
	}

	if cfg.Patches != nil {
		t.Error("Patches should be nil")
	}
}

func TestResult_EmptyFields(t *testing.T) {
	result := &Result{}

	if result.Tool != "" {
		t.Errorf("Tool = %v, want empty", result.Tool)
	}
	if result.URL != "" {
		t.Errorf("URL = %s, want empty", result.URL)
	}
	if result.Username != "" {
		t.Errorf("Username = %s, want empty", result.Username)
	}
	if result.Password != "" {
		t.Errorf("Password = %s, want empty", result.Password)
	}
	if result.Namespace != "" {
		t.Errorf("Namespace = %s, want empty", result.Namespace)
	}
	if result.Ready {
		t.Error("Ready should be false by default")
	}
	if result.Message != "" {
		t.Errorf("Message = %s, want empty", result.Message)
	}
}

func TestGetArgoCDHelmConfig_CustomSetValues(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			SetValues: map[string]string{
				"server.insecure":     "true",
				"controller.replicas": "2",
			},
		},
	})

	cfg := b.getArgoCDHelmConfig()
	if cfg.SetValues["server.insecure"] != "true" {
		t.Errorf("SetValues[server.insecure] = %s, want true", cfg.SetValues["server.insecure"])
	}
	if cfg.SetValues["controller.replicas"] != "2" {
		t.Errorf("SetValues[controller.replicas] = %s, want 2", cfg.SetValues["controller.replicas"])
	}
}

func TestGetFluxHelmConfig_CustomSetValues(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolFlux,
		Mode: ModeHelm,
		Helm: &HelmConfig{
			SetValues: map[string]string{
				"installCRDs": "true",
			},
		},
	})

	cfg := b.getFluxHelmConfig()
	if cfg.SetValues["installCRDs"] != "true" {
		t.Errorf("SetValues[installCRDs] = %s, want true", cfg.SetValues["installCRDs"])
	}
}

func TestSuggestMode_UnknownPlatform(t *testing.T) {
	mode := SuggestMode(ToolArgoCD, "unknown-platform")
	if mode != ModeHelm {
		t.Errorf("SuggestMode for unknown platform = %v, want ModeHelm", mode)
	}
}

func TestSuggestMode_EmptyPlatform(t *testing.T) {
	mode := SuggestMode(ToolArgoCD, "")
	if mode != ModeHelm {
		t.Errorf("SuggestMode for empty platform = %v, want ModeHelm", mode)
	}
}

func TestValidModes_EmptyPlatform(t *testing.T) {
	modes := ValidModes(ToolArgoCD, "")
	if len(modes) < 3 {
		t.Errorf("Expected at least 3 modes for empty platform, got %d", len(modes))
	}
	// Should not include OLM for non-openshift
	hasOLM := false
	for _, m := range modes {
		if m == ModeOLM {
			hasOLM = true
			break
		}
	}
	if hasOLM {
		t.Error("ValidModes for empty platform should not include OLM")
	}
}

func TestGetArgoCDOLMConfig_PartialWithApproval(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeOLM,
		OLM: &OLMConfig{
			Approval: "Manual",
			// Others empty - should use defaults
		},
	})

	cfg := b.getArgoCDOLMConfig()
	if cfg.Approval != "Manual" {
		t.Errorf("Approval = %s, want Manual", cfg.Approval)
	}
	if cfg.Channel != "alpha" {
		t.Errorf("Channel should default to alpha, got %s", cfg.Channel)
	}
}

func TestGetArgoCDManifestConfig_WithPaths(t *testing.T) {
	b := New(nil, &Options{
		Tool: ToolArgoCD,
		Mode: ModeManifest,
		Manifest: &ManifestConfig{
			URL:   "https://example.com/install.yaml",
			Paths: []string{"/path/to/extra1.yaml", "/path/to/extra2.yaml"},
		},
	})

	cfg := b.getArgoCDManifestConfig()
	if len(cfg.Paths) != 2 {
		t.Errorf("Paths length = %d, want 2", len(cfg.Paths))
	}
}
