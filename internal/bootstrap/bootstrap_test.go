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
