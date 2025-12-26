package multirepo

import (
	"testing"
)

func TestRepository_Validate(t *testing.T) {
	tests := []struct {
		name    string
		repo    Repository
		wantErr bool
	}{
		{
			name: "valid repository",
			repo: Repository{
				Name: "test-repo",
				URL:  "https://github.com/org/repo",
				Type: RepoTypeApplications,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			repo: Repository{
				URL:  "https://github.com/org/repo",
				Type: RepoTypeApplications,
			},
			wantErr: true,
		},
		{
			name: "missing URL",
			repo: Repository{
				Name: "test-repo",
				Type: RepoTypeApplications,
			},
			wantErr: true,
		},
		{
			name: "missing type",
			repo: Repository{
				Name: "test-repo",
				URL:  "https://github.com/org/repo",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			repo: Repository{
				Name: "test-repo",
				URL:  "https://github.com/org/repo",
				Type: "invalid-type",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_GetBranch(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{"custom branch", "develop", "develop"},
		{"default branch", "", "main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := Repository{Branch: tt.branch}
			if got := repo.GetBranch(); got != tt.want {
				t.Errorf("GetBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"custom path", "apps/production", "apps/production"},
		{"default path", "", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := Repository{Path: tt.path}
			if got := repo.GetPath(); got != tt.want {
				t.Errorf("GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetCredentialType(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		credentials *Credentials
		want        string
	}{
		{"https URL", "https://github.com/org/repo", nil, "https"},
		{"ssh git@ URL", "git@github.com:org/repo.git", nil, "ssh"},
		{"ssh:// URL", "ssh://git@github.com/org/repo", nil, "ssh"},
		{"explicit token", "https://github.com/org/repo", &Credentials{Type: "token"}, "token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := Repository{URL: tt.url, Credentials: tt.credentials}
			if got := repo.GetCredentialType(); got != tt.want {
				t.Errorf("GetCredentialType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.config.Enabled {
		t.Error("Default config should have Enabled = false")
	}
}

func TestManager_AddRepository(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	repo := &Repository{
		Name: "test-repo",
		URL:  "https://github.com/org/repo",
		Type: RepoTypeApplications,
	}

	err := manager.AddRepository(repo)
	if err != nil {
		t.Errorf("AddRepository() error = %v", err)
	}

	if len(manager.ListRepositories()) != 1 {
		t.Error("Repository should be added")
	}
}

func TestManager_AddRepository_Duplicate(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	repo := &Repository{
		Name: "test-repo",
		URL:  "https://github.com/org/repo",
		Type: RepoTypeApplications,
	}

	_ = manager.AddRepository(repo)
	err := manager.AddRepository(repo)
	if err == nil {
		t.Error("AddRepository() should fail for duplicate name")
	}
}

func TestManager_AddRepository_Invalid(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	repo := &Repository{
		Name: "test-repo",
		// Missing URL
	}

	err := manager.AddRepository(repo)
	if err == nil {
		t.Error("AddRepository() should fail for invalid repository")
	}
}

func TestManager_RemoveRepository(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	_ = manager.AddRepository(&Repository{Name: "repo1", URL: "https://url1", Type: RepoTypeApplications})
	_ = manager.AddRepository(&Repository{Name: "repo2", URL: "https://url2", Type: RepoTypeInfrastructure})

	removed := manager.RemoveRepository("repo1")
	if !removed {
		t.Error("RemoveRepository() should return true")
	}

	if len(manager.ListRepositories()) != 1 {
		t.Error("Should have 1 repository after removal")
	}
}

func TestManager_RemoveRepository_NotFound(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	removed := manager.RemoveRepository("nonexistent")
	if removed {
		t.Error("RemoveRepository() should return false for nonexistent repo")
	}
}

func TestManager_GetRepository(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	_ = manager.AddRepository(&Repository{Name: "repo1", URL: "https://url1", Type: RepoTypeApplications})

	repo, found := manager.GetRepository("repo1")
	if !found {
		t.Error("GetRepository() should find repo1")
	}
	if repo.Name != "repo1" {
		t.Errorf("GetRepository() name = %v, want repo1", repo.Name)
	}

	_, found = manager.GetRepository("nonexistent")
	if found {
		t.Error("GetRepository() should not find nonexistent")
	}
}

func TestManager_AddMapping(t *testing.T) {
	manager := NewManager(NewDefaultConfig())
	_ = manager.AddRepository(&Repository{Name: "apps-repo", URL: "https://url", Type: RepoTypeApplications})

	mapping := ApplicationMapping{
		Application: "my-app",
		Repository:  "apps-repo",
	}

	err := manager.AddMapping(mapping)
	if err != nil {
		t.Errorf("AddMapping() error = %v", err)
	}
}

func TestManager_AddMapping_InvalidRepo(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	mapping := ApplicationMapping{
		Application: "my-app",
		Repository:  "nonexistent-repo",
	}

	err := manager.AddMapping(mapping)
	if err == nil {
		t.Error("AddMapping() should fail for nonexistent repository")
	}
}

func TestManager_GetMappingForApplication(t *testing.T) {
	manager := NewManager(NewDefaultConfig())
	_ = manager.AddRepository(&Repository{Name: "apps-repo", URL: "https://url", Type: RepoTypeApplications})

	// Add exact match mapping
	_ = manager.AddMapping(ApplicationMapping{Application: "my-app", Repository: "apps-repo"})
	// Add wildcard mapping
	_ = manager.AddMapping(ApplicationMapping{Application: "frontend-*", Repository: "apps-repo"})

	tests := []struct {
		name    string
		appName string
		want    bool
	}{
		{"exact match", "my-app", true},
		{"wildcard match", "frontend-service", true},
		{"no match", "backend-api", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := manager.GetMappingForApplication(tt.appName)
			if found != tt.want {
				t.Errorf("GetMappingForApplication(%s) found = %v, want %v", tt.appName, found, tt.want)
			}
		})
	}
}

func TestManager_SetPrimary(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	repo := &Repository{
		Name: "primary-repo",
		URL:  "https://github.com/org/main-repo",
		Type: RepoTypeMonorepo,
	}

	err := manager.SetPrimary(repo)
	if err != nil {
		t.Errorf("SetPrimary() error = %v", err)
	}

	primary := manager.GetPrimary()
	if primary == nil {
		t.Fatal("GetPrimary() should not return nil")
	}
	if primary.Name != "primary-repo" {
		t.Errorf("Primary name = %v, want primary-repo", primary.Name)
	}
}

func TestManager_GetRepositoryForApplication(t *testing.T) {
	manager := NewManager(NewDefaultConfig())

	// Set up primary
	_ = manager.SetPrimary(&Repository{Name: "main", URL: "https://main", Type: RepoTypeMonorepo})

	// Add specific repo and mapping
	_ = manager.AddRepository(&Repository{Name: "apps", URL: "https://apps", Type: RepoTypeApplications})
	_ = manager.AddMapping(ApplicationMapping{Application: "frontend-*", Repository: "apps"})

	tests := []struct {
		name     string
		appName  string
		wantRepo string
		wantErr  bool
	}{
		{"mapped app", "frontend-web", "apps", false},
		{"unmapped app falls to primary", "backend-api", "main", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := manager.GetRepositoryForApplication(tt.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositoryForApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && repo.Name != tt.wantRepo {
				t.Errorf("GetRepositoryForApplication() repo = %v, want %v", repo.Name, tt.wantRepo)
			}
		})
	}
}

func TestManager_ValidateAll(t *testing.T) {
	config := NewDefaultConfig()
	config.Repositories = []Repository{
		{Name: "valid", URL: "https://url", Type: RepoTypeApplications},
		{Name: "invalid"},
	}
	manager := NewManager(config)

	errors := manager.ValidateAll()
	if len(errors) != 1 {
		t.Errorf("ValidateAll() errors count = %d, want 1", len(errors))
	}
}

func TestManager_GetRepositoriesByType(t *testing.T) {
	manager := NewManager(NewDefaultConfig())
	_ = manager.AddRepository(&Repository{Name: "app1", URL: "https://app1", Type: RepoTypeApplications})
	_ = manager.AddRepository(&Repository{Name: "app2", URL: "https://app2", Type: RepoTypeApplications})
	_ = manager.AddRepository(&Repository{Name: "infra", URL: "https://infra", Type: RepoTypeInfrastructure})

	apps := manager.GetRepositoriesByType(RepoTypeApplications)
	if len(apps) != 2 {
		t.Errorf("GetRepositoriesByType(applications) count = %d, want 2", len(apps))
	}

	infra := manager.GetRepositoriesByType(RepoTypeInfrastructure)
	if len(infra) != 1 {
		t.Errorf("GetRepositoriesByType(infrastructure) count = %d, want 1", len(infra))
	}
}

func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		pattern string
		str     string
		want    bool
	}{
		{"*", "anything", true},
		{"prefix-*", "prefix-app", true},
		{"prefix-*", "other-app", false},
		{"*-suffix", "app-suffix", true},
		{"*-suffix", "app-other", false},
		{"*middle*", "has-middle-here", true},
		{"*middle*", "no-match", false},
		{"exact", "exact", true},
		{"exact", "different", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.str, func(t *testing.T) {
			if got := matchWildcard(tt.pattern, tt.str); got != tt.want {
				t.Errorf("matchWildcard(%q, %q) = %v, want %v", tt.pattern, tt.str, got, tt.want)
			}
		})
	}
}

func TestRepositoryTypeConstants(t *testing.T) {
	if RepoTypeMonorepo != "monorepo" {
		t.Error("RepoTypeMonorepo constant wrong")
	}
	if RepoTypeApplications != "applications" {
		t.Error("RepoTypeApplications constant wrong")
	}
	if RepoTypeInfrastructure != "infrastructure" {
		t.Error("RepoTypeInfrastructure constant wrong")
	}
	if RepoTypeConfig != "config" {
		t.Error("RepoTypeConfig constant wrong")
	}
	if RepoTypeHelmCharts != "helm-charts" {
		t.Error("RepoTypeHelmCharts constant wrong")
	}
	if RepoTypeKustomize != "kustomize" {
		t.Error("RepoTypeKustomize constant wrong")
	}
}

func TestRepository_ToRepositoryManifest(t *testing.T) {
	repo := Repository{
		Name: "test-repo",
		URL:  "https://github.com/org/repo",
		Type: RepoTypeApplications,
	}

	manifest := repo.ToRepositoryManifest("argocd", "default")

	if manifest.Name != "test-repo" {
		t.Errorf("Name = %v, want test-repo", manifest.Name)
	}
	if manifest.Namespace != "argocd" {
		t.Errorf("Namespace = %v, want argocd", manifest.Namespace)
	}
	if manifest.URL != "https://github.com/org/repo" {
		t.Errorf("URL = %v, want https://github.com/org/repo", manifest.URL)
	}
	if manifest.Type != "git" {
		t.Errorf("Type = %v, want git", manifest.Type)
	}
	if manifest.Project != "default" {
		t.Errorf("Project = %v, want default", manifest.Project)
	}
}

func TestRepository_ToRepositoryManifest_HelmType(t *testing.T) {
	repo := Repository{
		Name: "helm-repo",
		URL:  "https://charts.helm.sh/stable",
		Type: RepoTypeHelmCharts,
	}

	manifest := repo.ToRepositoryManifest("argocd", "default")

	if manifest.Type != "helm" {
		t.Errorf("Type = %v, want helm", manifest.Type)
	}
}
