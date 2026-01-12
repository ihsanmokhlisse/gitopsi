package progress

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// Tests for SetupSummary struct

func TestSetupSummary_Empty(t *testing.T) {
	summary := SetupSummary{}

	if summary.Setup.CompletedAt != (time.Time{}) {
		t.Error("CompletedAt should be zero value")
	}
	if summary.Git.URL != "" {
		t.Error("Git URL should be empty")
	}
	if summary.Cluster.Name != "" {
		t.Error("Cluster Name should be empty")
	}
	if summary.GitOpsTool.Name != "" {
		t.Error("GitOpsTool Name should be empty")
	}
	if len(summary.Environments) != 0 {
		t.Error("Environments should be empty")
	}
	if len(summary.Applications) != 0 {
		t.Error("Applications should be empty")
	}
}

func TestSetupSummary_AllFields(t *testing.T) {
	now := time.Now()
	summary := SetupSummary{
		Setup: SetupInfo{
			CompletedAt: now,
			Duration:    5 * time.Minute,
			Version:     "1.0.0",
		},
		Git: GitInfo{
			URL:      "https://github.com/org/repo",
			Branch:   "main",
			WebURL:   "https://github.com/org/repo",
			Provider: "github",
			Status:   "connected",
		},
		Cluster: ClusterInfo{
			Name:       "prod-cluster",
			URL:        "https://k8s.example.com",
			Platform:   "kubernetes",
			Version:    "1.28.0",
			Status:     "healthy",
			Namespaces: []string{"dev", "staging", "prod"},
		},
		GitOpsTool: GitOpsToolInfo{
			Name:      "argocd",
			URL:       "https://argocd.example.com",
			Username:  "admin",
			Password:  "secret",
			Namespace: "argocd",
			Version:   "2.9.0",
			Status:    "ready",
			PodCount:  "3/3",
		},
		Environments: []EnvironmentInfo{
			{Name: "dev", Namespace: "dev", Status: "active"},
			{Name: "staging", Namespace: "staging", Status: "active"},
			{Name: "prod", Namespace: "prod", Status: "active"},
		},
		Applications: []ApplicationInfo{
			{Name: "app-of-apps", Type: "Application", Status: "synced", Children: []string{"infra", "apps"}},
		},
	}

	if summary.Setup.Version != "1.0.0" {
		t.Errorf("Setup.Version = %s, want 1.0.0", summary.Setup.Version)
	}
	if summary.Git.Provider != "github" {
		t.Errorf("Git.Provider = %s, want github", summary.Git.Provider)
	}
	if len(summary.Cluster.Namespaces) != 3 {
		t.Errorf("Cluster.Namespaces count = %d, want 3", len(summary.Cluster.Namespaces))
	}
	if summary.GitOpsTool.PodCount != "3/3" {
		t.Errorf("GitOpsTool.PodCount = %s, want 3/3", summary.GitOpsTool.PodCount)
	}
	if len(summary.Environments) != 3 {
		t.Errorf("Environments count = %d, want 3", len(summary.Environments))
	}
	if len(summary.Applications[0].Children) != 2 {
		t.Errorf("Applications[0].Children count = %d, want 2", len(summary.Applications[0].Children))
	}
}

// Tests for SetupInfo struct

func TestSetupInfo_ZeroDuration(t *testing.T) {
	info := SetupInfo{
		CompletedAt: time.Now(),
		Duration:    0,
		Version:     "0.1.0",
	}

	if info.Duration != 0 {
		t.Errorf("Duration = %v, want 0", info.Duration)
	}
}

func TestSetupInfo_LongDuration(t *testing.T) {
	info := SetupInfo{
		CompletedAt: time.Now(),
		Duration:    24 * time.Hour,
		Version:     "1.0.0",
	}

	if info.Duration != 24*time.Hour {
		t.Errorf("Duration = %v, want 24h", info.Duration)
	}
}

func TestSetupInfo_EmptyVersion(t *testing.T) {
	info := SetupInfo{
		Version: "",
	}

	if info.Version != "" {
		t.Errorf("Version = %s, want empty", info.Version)
	}
}

// Tests for GitInfo struct

func TestGitInfo_AllFields(t *testing.T) {
	git := GitInfo{
		URL:      "git@github.com:org/repo.git",
		Branch:   "develop",
		WebURL:   "https://github.com/org/repo",
		Provider: "github",
		Status:   "connected",
	}

	if git.URL != "git@github.com:org/repo.git" {
		t.Errorf("URL = %s", git.URL)
	}
	if git.Branch != "develop" {
		t.Errorf("Branch = %s, want develop", git.Branch)
	}
	if git.WebURL != "https://github.com/org/repo" {
		t.Errorf("WebURL = %s", git.WebURL)
	}
	if git.Provider != "github" {
		t.Errorf("Provider = %s, want github", git.Provider)
	}
	if git.Status != "connected" {
		t.Errorf("Status = %s, want connected", git.Status)
	}
}

func TestGitInfo_DifferentProviders(t *testing.T) {
	providers := []string{"github", "gitlab", "bitbucket", "azure", "gitea"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			git := GitInfo{Provider: provider}
			if git.Provider != provider {
				t.Errorf("Provider = %s, want %s", git.Provider, provider)
			}
		})
	}
}

func TestGitInfo_StatusValues(t *testing.T) {
	statuses := []string{"connected", "disconnected", "synced", "error", "pending"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			git := GitInfo{Status: status}
			if git.Status != status {
				t.Errorf("Status = %s, want %s", git.Status, status)
			}
		})
	}
}

// Tests for ClusterInfo struct

func TestClusterInfo_AllFields(t *testing.T) {
	cluster := ClusterInfo{
		Name:       "production",
		URL:        "https://api.k8s.example.com:6443",
		Platform:   "kubernetes",
		Version:    "1.29.0",
		Status:     "healthy",
		Namespaces: []string{"default", "kube-system", "app"},
	}

	if cluster.Name != "production" {
		t.Errorf("Name = %s, want production", cluster.Name)
	}
	if cluster.URL != "https://api.k8s.example.com:6443" {
		t.Errorf("URL = %s", cluster.URL)
	}
	if cluster.Platform != "kubernetes" {
		t.Errorf("Platform = %s, want kubernetes", cluster.Platform)
	}
	if cluster.Version != "1.29.0" {
		t.Errorf("Version = %s, want 1.29.0", cluster.Version)
	}
	if cluster.Status != "healthy" {
		t.Errorf("Status = %s, want healthy", cluster.Status)
	}
	if len(cluster.Namespaces) != 3 {
		t.Errorf("Namespaces count = %d, want 3", len(cluster.Namespaces))
	}
}

func TestClusterInfo_NilNamespaces_Minimal(t *testing.T) {
	cluster := ClusterInfo{
		Name:       "minimal",
		Namespaces: nil,
	}

	if cluster.Namespaces != nil {
		t.Error("Namespaces should be nil")
	}
}

func TestClusterInfo_DifferentPlatforms(t *testing.T) {
	platforms := []string{"kubernetes", "openshift", "eks", "aks", "gke"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			cluster := ClusterInfo{Platform: platform}
			if cluster.Platform != platform {
				t.Errorf("Platform = %s, want %s", cluster.Platform, platform)
			}
		})
	}
}

func TestClusterInfo_StatusValues(t *testing.T) {
	statuses := []string{"healthy", "degraded", "error", "disconnected", "unknown"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			cluster := ClusterInfo{Status: status}
			if cluster.Status != status {
				t.Errorf("Status = %s, want %s", cluster.Status, status)
			}
		})
	}
}

// Tests for GitOpsToolInfo struct

func TestGitOpsToolInfo_AllFields(t *testing.T) {
	tool := GitOpsToolInfo{
		Name:           "argocd",
		URL:            "https://argocd.example.com",
		Username:       "admin",
		Password:       "supersecret",
		PasswordSecret: "argocd-initial-admin-secret",
		Namespace:      "argocd",
		Version:        "2.10.0",
		Status:         "ready",
		PodCount:       "5/5",
	}

	if tool.Name != "argocd" {
		t.Errorf("Name = %s, want argocd", tool.Name)
	}
	if tool.URL != "https://argocd.example.com" {
		t.Errorf("URL = %s", tool.URL)
	}
	if tool.Username != "admin" {
		t.Errorf("Username = %s, want admin", tool.Username)
	}
	if tool.Password != "supersecret" {
		t.Errorf("Password = %s, want supersecret", tool.Password)
	}
	if tool.PasswordSecret != "argocd-initial-admin-secret" {
		t.Errorf("PasswordSecret = %s", tool.PasswordSecret)
	}
	if tool.Namespace != "argocd" {
		t.Errorf("Namespace = %s, want argocd", tool.Namespace)
	}
	if tool.Version != "2.10.0" {
		t.Errorf("Version = %s, want 2.10.0", tool.Version)
	}
	if tool.Status != "ready" {
		t.Errorf("Status = %s, want ready", tool.Status)
	}
	if tool.PodCount != "5/5" {
		t.Errorf("PodCount = %s, want 5/5", tool.PodCount)
	}
}

func TestGitOpsToolInfo_OpenShiftGitOps(t *testing.T) {
	tool := GitOpsToolInfo{
		Name:      "openshift-gitops",
		Namespace: "openshift-gitops",
	}

	if tool.Name != "openshift-gitops" {
		t.Errorf("Name = %s, want openshift-gitops", tool.Name)
	}
	if tool.Namespace != "openshift-gitops" {
		t.Errorf("Namespace = %s, want openshift-gitops", tool.Namespace)
	}
}

func TestGitOpsToolInfo_Flux(t *testing.T) {
	tool := GitOpsToolInfo{
		Name:      "flux",
		Namespace: "flux-system",
		Version:   "2.2.0",
	}

	if tool.Name != "flux" {
		t.Errorf("Name = %s, want flux", tool.Name)
	}
	if tool.Namespace != "flux-system" {
		t.Errorf("Namespace = %s, want flux-system", tool.Namespace)
	}
}

func TestGitOpsToolInfo_EmptyPassword(t *testing.T) {
	tool := GitOpsToolInfo{
		Name:     "argocd",
		Password: "",
	}

	if tool.Password != "" {
		t.Errorf("Password = %s, want empty", tool.Password)
	}
}

// Tests for EnvironmentInfo struct

func TestEnvironmentInfo_AllFields(t *testing.T) {
	env := EnvironmentInfo{
		Name:      "production",
		Namespace: "prod",
		Status:    "active",
	}

	if env.Name != "production" {
		t.Errorf("Name = %s, want production", env.Name)
	}
	if env.Namespace != "prod" {
		t.Errorf("Namespace = %s, want prod", env.Namespace)
	}
	if env.Status != "active" {
		t.Errorf("Status = %s, want active", env.Status)
	}
}

func TestEnvironmentInfo_DifferentEnvironments(t *testing.T) {
	environments := []struct {
		name      string
		namespace string
	}{
		{"dev", "development"},
		{"staging", "staging"},
		{"prod", "production"},
		{"qa", "qa"},
		{"uat", "uat"},
	}

	for _, env := range environments {
		t.Run(env.name, func(t *testing.T) {
			info := EnvironmentInfo{Name: env.name, Namespace: env.namespace}
			if info.Name != env.name {
				t.Errorf("Name = %s, want %s", info.Name, env.name)
			}
			if info.Namespace != env.namespace {
				t.Errorf("Namespace = %s, want %s", info.Namespace, env.namespace)
			}
		})
	}
}

func TestEnvironmentInfo_StatusValues(t *testing.T) {
	statuses := []string{"active", "inactive", "pending", "error", "degraded"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			env := EnvironmentInfo{Status: status}
			if env.Status != status {
				t.Errorf("Status = %s, want %s", env.Status, status)
			}
		})
	}
}

// Tests for ApplicationInfo struct

func TestApplicationInfo_AllFields(t *testing.T) {
	app := ApplicationInfo{
		Name:     "my-app",
		Type:     "Application",
		Status:   "synced",
		Children: []string{"frontend", "backend", "database"},
	}

	if app.Name != "my-app" {
		t.Errorf("Name = %s, want my-app", app.Name)
	}
	if app.Type != "Application" {
		t.Errorf("Type = %s, want Application", app.Type)
	}
	if app.Status != "synced" {
		t.Errorf("Status = %s, want synced", app.Status)
	}
	if len(app.Children) != 3 {
		t.Errorf("Children count = %d, want 3", len(app.Children))
	}
}

func TestApplicationInfo_NoChildren(t *testing.T) {
	app := ApplicationInfo{
		Name:     "simple-app",
		Type:     "Application",
		Status:   "synced",
		Children: nil,
	}

	if app.Children != nil {
		t.Error("Children should be nil")
	}
}

func TestApplicationInfo_EmptyChildren(t *testing.T) {
	app := ApplicationInfo{
		Name:     "simple-app",
		Type:     "Application",
		Status:   "synced",
		Children: []string{},
	}

	if len(app.Children) != 0 {
		t.Errorf("Children count = %d, want 0", len(app.Children))
	}
}

func TestApplicationInfo_TypeValues(t *testing.T) {
	types := []string{"Application", "ApplicationSet", "AppProject"}

	for _, appType := range types {
		t.Run(appType, func(t *testing.T) {
			app := ApplicationInfo{Type: appType}
			if app.Type != appType {
				t.Errorf("Type = %s, want %s", app.Type, appType)
			}
		})
	}
}

func TestApplicationInfo_StatusValues(t *testing.T) {
	statuses := []string{"synced", "out-of-sync", "progressing", "degraded", "healthy", "missing", "unknown"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			app := ApplicationInfo{Status: status}
			if app.Status != status {
				t.Errorf("Status = %s, want %s", app.Status, status)
			}
		})
	}
}

// Tests for formatStatus function

func TestFormatStatus_Connected(t *testing.T) {
	result := formatStatus("connected")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'connected'")
	}
}

func TestFormatStatus_Synced(t *testing.T) {
	result := formatStatus("synced")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'synced'")
	}
}

func TestFormatStatus_Healthy(t *testing.T) {
	result := formatStatus("healthy")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'healthy'")
	}
}

func TestFormatStatus_Ready(t *testing.T) {
	result := formatStatus("ready")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'ready'")
	}
}

func TestFormatStatus_Warning(t *testing.T) {
	result := formatStatus("warning")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'warning'")
	}
}

func TestFormatStatus_Degraded(t *testing.T) {
	result := formatStatus("degraded")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'degraded'")
	}
}

func TestFormatStatus_Failed(t *testing.T) {
	result := formatStatus("failed")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'failed'")
	}
}

func TestFormatStatus_Error(t *testing.T) {
	result := formatStatus("error")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'error'")
	}
}

func TestFormatStatus_Disconnected(t *testing.T) {
	result := formatStatus("disconnected")
	if result == "" {
		t.Error("formatStatus should return non-empty string for 'disconnected'")
	}
}

func TestFormatStatus_Unknown(t *testing.T) {
	result := formatStatus("unknown")
	if result != "unknown" {
		t.Errorf("formatStatus('unknown') = %s, want 'unknown'", result)
	}
}

func TestFormatStatus_Empty(t *testing.T) {
	result := formatStatus("")
	if result != "" {
		t.Errorf("formatStatus('') = %s, want empty", result)
	}
}

func TestFormatStatus_Custom(t *testing.T) {
	result := formatStatus("custom-status")
	if result != "custom-status" {
		t.Errorf("formatStatus('custom-status') = %s, want 'custom-status'", result)
	}
}

// Tests for SaveSummary function

func TestSaveSummary_Success(t *testing.T) {
	tmpDir := t.TempDir()

	summary := &SetupSummary{
		Setup: SetupInfo{
			CompletedAt: time.Now(),
			Duration:    5 * time.Minute,
			Version:     "1.0.0",
		},
		Git: GitInfo{
			URL:      "https://github.com/org/repo",
			Branch:   "main",
			Provider: "github",
			Status:   "connected",
		},
	}

	err := SaveSummary(tmpDir, summary)
	if err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Verify file exists
	summaryPath := filepath.Join(tmpDir, ".gitopsi", "setup-summary.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Summary file was not created")
	}
}

func TestSaveSummary_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	summary := &SetupSummary{
		Setup: SetupInfo{Version: "1.0.0"},
	}

	err := SaveSummary(tmpDir, summary)
	if err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Verify directory was created
	gitopsiDir := filepath.Join(tmpDir, ".gitopsi")
	info, err := os.Stat(gitopsiDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error(".gitopsi should be a directory")
	}
}

func TestSaveSummary_FileContent(t *testing.T) {
	tmpDir := t.TempDir()

	summary := &SetupSummary{
		Setup: SetupInfo{
			Version: "1.0.0",
		},
		Git: GitInfo{
			URL:    "https://github.com/test/repo",
			Branch: "main",
		},
	}

	err := SaveSummary(tmpDir, summary)
	if err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Read and verify content
	summaryPath := filepath.Join(tmpDir, ".gitopsi", "setup-summary.yaml")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	var loaded SetupSummary
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal summary: %v", err)
	}

	if loaded.Setup.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", loaded.Setup.Version)
	}
	if loaded.Git.URL != "https://github.com/test/repo" {
		t.Errorf("Git.URL = %s", loaded.Git.URL)
	}
}

func TestSaveSummary_InvalidPath(t *testing.T) {
	// Use a path that will fail to create
	invalidPath := "/nonexistent/path/that/does/not/exist"

	summary := &SetupSummary{}
	err := SaveSummary(invalidPath, summary)
	if err == nil {
		t.Error("SaveSummary() should error for invalid path")
	}
}

func TestSaveSummary_EmptySummary(t *testing.T) {
	tmpDir := t.TempDir()

	summary := &SetupSummary{}

	err := SaveSummary(tmpDir, summary)
	if err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Should still create a valid file
	summaryPath := filepath.Join(tmpDir, ".gitopsi", "setup-summary.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Summary file was not created")
	}
}

func TestSaveSummary_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial summary
	summary1 := &SetupSummary{
		Setup: SetupInfo{Version: "1.0.0"},
	}
	if err := SaveSummary(tmpDir, summary1); err != nil {
		t.Fatalf("First SaveSummary() error = %v", err)
	}

	// Save a new summary
	summary2 := &SetupSummary{
		Setup: SetupInfo{Version: "2.0.0"},
	}
	if err := SaveSummary(tmpDir, summary2); err != nil {
		t.Fatalf("Second SaveSummary() error = %v", err)
	}

	// Verify it was overwritten
	summaryPath := filepath.Join(tmpDir, ".gitopsi", "setup-summary.yaml")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	var loaded SetupSummary
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal summary: %v", err)
	}

	if loaded.Setup.Version != "2.0.0" {
		t.Errorf("Version = %s, want 2.0.0 (should be overwritten)", loaded.Setup.Version)
	}
}

// Tests for LoadSummary function

func TestLoadSummary_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Save first
	original := &SetupSummary{
		Setup: SetupInfo{
			Version:  "1.0.0",
			Duration: 10 * time.Minute,
		},
		Git: GitInfo{
			URL:      "https://github.com/org/repo",
			Provider: "github",
		},
	}
	if err := SaveSummary(tmpDir, original); err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Load
	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	if loaded.Setup.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", loaded.Setup.Version)
	}
	if loaded.Git.URL != "https://github.com/org/repo" {
		t.Errorf("Git.URL = %s", loaded.Git.URL)
	}
	if loaded.Git.Provider != "github" {
		t.Errorf("Git.Provider = %s, want github", loaded.Git.Provider)
	}
}

func TestLoadSummary_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadSummary(tmpDir)
	if err == nil {
		t.Error("LoadSummary() should error for non-existent file")
	}
}

func TestLoadSummary_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid YAML file
	gitopsiDir := filepath.Join(tmpDir, ".gitopsi")
	if err := os.MkdirAll(gitopsiDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	invalidYAML := "invalid: yaml: content: ["
	summaryPath := filepath.Join(gitopsiDir, "setup-summary.yaml")
	if err := os.WriteFile(summaryPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := LoadSummary(tmpDir)
	if err == nil {
		t.Error("LoadSummary() should error for invalid YAML")
	}
}

func TestLoadSummary_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty file
	gitopsiDir := filepath.Join(tmpDir, ".gitopsi")
	if err := os.MkdirAll(gitopsiDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	summaryPath := filepath.Join(gitopsiDir, "setup-summary.yaml")
	if err := os.WriteFile(summaryPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Empty file should unmarshal to empty struct
	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	// Should get zero values
	if loaded.Setup.Version != "" {
		t.Errorf("Version should be empty, got %s", loaded.Setup.Version)
	}
}

func TestLoadSummary_PartialData(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with partial data
	gitopsiDir := filepath.Join(tmpDir, ".gitopsi")
	if err := os.MkdirAll(gitopsiDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	partialYAML := `setup:
  version: "1.0.0"
`
	summaryPath := filepath.Join(gitopsiDir, "setup-summary.yaml")
	if err := os.WriteFile(summaryPath, []byte(partialYAML), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	if loaded.Setup.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", loaded.Setup.Version)
	}
	// Other fields should be zero values
	if loaded.Git.URL != "" {
		t.Errorf("Git.URL should be empty, got %s", loaded.Git.URL)
	}
}

// Tests for SaveSummary and LoadSummary round-trip

func TestSaveLoadSummary_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	original := &SetupSummary{
		Setup: SetupInfo{
			CompletedAt: time.Now().Truncate(time.Second), // Truncate for comparison
			Duration:    15 * time.Minute,
			Version:     "2.0.0",
		},
		Git: GitInfo{
			URL:      "https://gitlab.com/org/repo",
			Branch:   "develop",
			WebURL:   "https://gitlab.com/org/repo",
			Provider: "gitlab",
			Status:   "synced",
		},
		Cluster: ClusterInfo{
			Name:       "dev-cluster",
			URL:        "https://api.k8s.dev.example.com",
			Platform:   "eks",
			Version:    "1.27.0",
			Status:     "healthy",
			Namespaces: []string{"app1", "app2"},
		},
		GitOpsTool: GitOpsToolInfo{
			Name:      "argocd",
			URL:       "https://argocd.dev.example.com",
			Username:  "admin",
			Password:  "secret123",
			Namespace: "argocd",
			Version:   "2.8.5",
			Status:    "ready",
			PodCount:  "4/4",
		},
		Environments: []EnvironmentInfo{
			{Name: "dev", Namespace: "dev", Status: "active"},
		},
		Applications: []ApplicationInfo{
			{
				Name:     "root-app",
				Type:     "Application",
				Status:   "synced",
				Children: []string{"child1", "child2"},
			},
		},
	}

	// Save
	if err := SaveSummary(tmpDir, original); err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Load
	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	// Compare key fields
	if loaded.Setup.Version != original.Setup.Version {
		t.Errorf("Setup.Version = %s, want %s", loaded.Setup.Version, original.Setup.Version)
	}
	if loaded.Git.URL != original.Git.URL {
		t.Errorf("Git.URL = %s, want %s", loaded.Git.URL, original.Git.URL)
	}
	if loaded.Git.Provider != original.Git.Provider {
		t.Errorf("Git.Provider = %s, want %s", loaded.Git.Provider, original.Git.Provider)
	}
	if loaded.Cluster.Name != original.Cluster.Name {
		t.Errorf("Cluster.Name = %s, want %s", loaded.Cluster.Name, original.Cluster.Name)
	}
	if loaded.Cluster.Platform != original.Cluster.Platform {
		t.Errorf("Cluster.Platform = %s, want %s", loaded.Cluster.Platform, original.Cluster.Platform)
	}
	if len(loaded.Cluster.Namespaces) != len(original.Cluster.Namespaces) {
		t.Errorf("Cluster.Namespaces count = %d, want %d", len(loaded.Cluster.Namespaces), len(original.Cluster.Namespaces))
	}
	if loaded.GitOpsTool.Name != original.GitOpsTool.Name {
		t.Errorf("GitOpsTool.Name = %s, want %s", loaded.GitOpsTool.Name, original.GitOpsTool.Name)
	}
	if loaded.GitOpsTool.Password != original.GitOpsTool.Password {
		t.Errorf("GitOpsTool.Password = %s, want %s", loaded.GitOpsTool.Password, original.GitOpsTool.Password)
	}
	if len(loaded.Environments) != len(original.Environments) {
		t.Errorf("Environments count = %d, want %d", len(loaded.Environments), len(original.Environments))
	}
	if len(loaded.Applications) != len(original.Applications) {
		t.Errorf("Applications count = %d, want %d", len(loaded.Applications), len(original.Applications))
	}
	if loaded.Applications[0].Name != original.Applications[0].Name {
		t.Errorf("Applications[0].Name = %s, want %s", loaded.Applications[0].Name, original.Applications[0].Name)
	}
	if len(loaded.Applications[0].Children) != len(original.Applications[0].Children) {
		t.Errorf("Applications[0].Children count = %d, want %d", len(loaded.Applications[0].Children), len(original.Applications[0].Children))
	}
}

func TestSaveLoadSummary_WithSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()

	original := &SetupSummary{
		Git: GitInfo{
			URL: "https://github.com/org/repo-with-special-chars!@#$%",
		},
		GitOpsTool: GitOpsToolInfo{
			Password: "p@$$w0rd!#$%^&*()",
		},
	}

	// Save
	if err := SaveSummary(tmpDir, original); err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Load
	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	if loaded.Git.URL != original.Git.URL {
		t.Errorf("Git.URL = %s, want %s", loaded.Git.URL, original.Git.URL)
	}
	if loaded.GitOpsTool.Password != original.GitOpsTool.Password {
		t.Errorf("GitOpsTool.Password = %s, want %s", loaded.GitOpsTool.Password, original.GitOpsTool.Password)
	}
}

func TestSaveLoadSummary_ManyEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	envs := make([]EnvironmentInfo, 10)
	for i := 0; i < 10; i++ {
		envs[i] = EnvironmentInfo{
			Name:      "env-" + string(rune('a'+i)),
			Namespace: "ns-" + string(rune('a'+i)),
			Status:    "active",
		}
	}

	original := &SetupSummary{
		Environments: envs,
	}

	// Save
	if err := SaveSummary(tmpDir, original); err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Load
	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	if len(loaded.Environments) != 10 {
		t.Errorf("Environments count = %d, want 10", len(loaded.Environments))
	}
}

func TestSaveLoadSummary_ManyApplications(t *testing.T) {
	tmpDir := t.TempDir()

	apps := make([]ApplicationInfo, 20)
	for i := 0; i < 20; i++ {
		apps[i] = ApplicationInfo{
			Name:     "app-" + string(rune('a'+i%26)),
			Type:     "Application",
			Status:   "synced",
			Children: []string{"child1", "child2", "child3"},
		}
	}

	original := &SetupSummary{
		Applications: apps,
	}

	// Save
	if err := SaveSummary(tmpDir, original); err != nil {
		t.Fatalf("SaveSummary() error = %v", err)
	}

	// Load
	loaded, err := LoadSummary(tmpDir)
	if err != nil {
		t.Fatalf("LoadSummary() error = %v", err)
	}

	if len(loaded.Applications) != 20 {
		t.Errorf("Applications count = %d, want 20", len(loaded.Applications))
	}

	for i, app := range loaded.Applications {
		if len(app.Children) != 3 {
			t.Errorf("Applications[%d].Children count = %d, want 3", i, len(app.Children))
		}
	}
}

// Edge cases

func TestSetupSummary_NilSlices(t *testing.T) {
	summary := SetupSummary{
		Cluster:      ClusterInfo{Namespaces: nil},
		Environments: nil,
		Applications: nil,
	}

	if summary.Cluster.Namespaces != nil {
		t.Error("Namespaces should be nil")
	}
	if summary.Environments != nil {
		t.Error("Environments should be nil")
	}
	if summary.Applications != nil {
		t.Error("Applications should be nil")
	}
}

func TestClusterInfo_ManyNamespaces(t *testing.T) {
	namespaces := make([]string, 100)
	for i := 0; i < 100; i++ {
		namespaces[i] = "ns-" + string(rune('0'+i/10)) + string(rune('0'+i%10))
	}

	cluster := ClusterInfo{
		Namespaces: namespaces,
	}

	if len(cluster.Namespaces) != 100 {
		t.Errorf("Namespaces count = %d, want 100", len(cluster.Namespaces))
	}
}

func TestApplicationInfo_ManyChildren(t *testing.T) {
	children := make([]string, 50)
	for i := 0; i < 50; i++ {
		children[i] = "child-" + string(rune('a'+i%26))
	}

	app := ApplicationInfo{
		Children: children,
	}

	if len(app.Children) != 50 {
		t.Errorf("Children count = %d, want 50", len(app.Children))
	}
}

func TestSetupInfo_VeryLongDuration(t *testing.T) {
	info := SetupInfo{
		Duration: 30 * 24 * time.Hour, // 30 days
	}

	if info.Duration != 30*24*time.Hour {
		t.Errorf("Duration = %v, want 720h", info.Duration)
	}
}

func TestSetupInfo_NegativeDuration(t *testing.T) {
	info := SetupInfo{
		Duration: -5 * time.Minute,
	}

	if info.Duration != -5*time.Minute {
		t.Errorf("Duration = %v, want -5m", info.Duration)
	}
}
