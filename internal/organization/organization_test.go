package organization

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")

	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestInitOrganization(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)

	org, err := manager.InitOrganization(ctx, "acme-corp", "acme.com")
	if err != nil {
		t.Fatalf("InitOrganization failed: %v", err)
	}

	if org.Name != "acme-corp" {
		t.Errorf("Name: got %s, want acme-corp", org.Name)
	}
	if org.Domain != "acme.com" {
		t.Errorf("Domain: got %s, want acme.com", org.Domain)
	}

	// Try to init again - should fail
	_, err = manager.InitOrganization(ctx, "another-org", "another.com")
	if err == nil {
		t.Error("InitOrganization should fail if org already exists")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should have been created")
	}
}

func TestInitOrganization_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)

	_, err := manager.InitOrganization(ctx, "", "acme.com")
	if err == nil {
		t.Error("InitOrganization should fail with empty name")
	}
}

func TestCreateTeam(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")

	opts := &TeamOptions{
		Name:        "frontend",
		Description: "Frontend team",
		Owners:      []string{"frontend-leads@acme.com"},
		Quotas: TeamQuotas{
			CPU:    "20",
			Memory: "40Gi",
		},
		AllowedClusters: []string{"dev-cluster", "prod-cluster"},
	}

	team, err := manager.CreateTeam(ctx, opts)
	if err != nil {
		t.Fatalf("CreateTeam failed: %v", err)
	}

	if team.Name != "frontend" {
		t.Errorf("Name: got %s, want frontend", team.Name)
	}
	if team.Quotas.CPU != "20" {
		t.Errorf("Quotas.CPU: got %s, want 20", team.Quotas.CPU)
	}
	if len(team.AllowedClusters) != 2 {
		t.Errorf("AllowedClusters: got %d, want 2", len(team.AllowedClusters))
	}

	// Try to create duplicate team
	_, err = manager.CreateTeam(ctx, opts)
	if err == nil {
		t.Error("CreateTeam should fail for duplicate team")
	}
}

func TestCreateTeam_NoOrg(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)

	_, err := manager.CreateTeam(ctx, &TeamOptions{Name: "test"})
	if err == nil {
		t.Error("CreateTeam should fail if no org initialized")
	}
}

func TestGetTeam(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "backend"})

	team, err := manager.GetTeam(ctx, "backend")
	if err != nil {
		t.Fatalf("GetTeam failed: %v", err)
	}
	if team.Name != "backend" {
		t.Errorf("Name: got %s, want backend", team.Name)
	}

	// Get non-existent team
	_, err = manager.GetTeam(ctx, "non-existent")
	if err == nil {
		t.Error("GetTeam should fail for non-existent team")
	}
}

func TestListTeams(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "backend"})
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "platform"})

	teams, err := manager.ListTeams(ctx)
	if err != nil {
		t.Fatalf("ListTeams failed: %v", err)
	}
	if len(teams) != 3 {
		t.Errorf("ListTeams: got %d, want 3", len(teams))
	}
}

func TestDeleteTeam(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "to-delete"})

	err := manager.DeleteTeam(ctx, "to-delete")
	if err != nil {
		t.Fatalf("DeleteTeam failed: %v", err)
	}

	// Verify deleted
	_, err = manager.GetTeam(ctx, "to-delete")
	if err == nil {
		t.Error("Team should not exist after deletion")
	}

	// Delete non-existent team
	err = manager.DeleteTeam(ctx, "non-existent")
	if err == nil {
		t.Error("DeleteTeam should fail for non-existent team")
	}
}

func TestSetTeamQuota(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})

	quotas := TeamQuotas{
		CPU:    "50",
		Memory: "100Gi",
	}

	err := manager.SetTeamQuota(ctx, "frontend", quotas)
	if err != nil {
		t.Fatalf("SetTeamQuota failed: %v", err)
	}

	team, _ := manager.GetTeam(ctx, "frontend")
	if team.Quotas.CPU != "50" {
		t.Errorf("Quotas.CPU: got %s, want 50", team.Quotas.CPU)
	}
}

func TestAddClusterToTeam(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})

	err := manager.AddClusterToTeam(ctx, "frontend", "prod-cluster")
	if err != nil {
		t.Fatalf("AddClusterToTeam failed: %v", err)
	}

	team, _ := manager.GetTeam(ctx, "frontend")
	if len(team.AllowedClusters) != 1 {
		t.Errorf("AllowedClusters: got %d, want 1", len(team.AllowedClusters))
	}

	// Add same cluster again - should not duplicate
	_ = manager.AddClusterToTeam(ctx, "frontend", "prod-cluster")
	team, _ = manager.GetTeam(ctx, "frontend")
	if len(team.AllowedClusters) != 1 {
		t.Errorf("AllowedClusters should not duplicate: got %d", len(team.AllowedClusters))
	}
}

func TestCreateProject(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})

	opts := &ProjectOptions{
		Name:         "web-app",
		Description:  "Main web application",
		Environments: []string{"dev", "staging", "prod"},
		SourceRepo:   "https://github.com/acme/web-app",
	}

	project, err := manager.CreateProject(ctx, "frontend", opts)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	if project.Name != "web-app" {
		t.Errorf("Name: got %s, want web-app", project.Name)
	}
	if len(project.Environments) != 3 {
		t.Errorf("Environments: got %d, want 3", len(project.Environments))
	}

	// Create duplicate project
	_, err = manager.CreateProject(ctx, "frontend", opts)
	if err == nil {
		t.Error("CreateProject should fail for duplicate project")
	}
}

func TestCreateProject_DefaultEnvironments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})

	opts := &ProjectOptions{
		Name: "api",
		// No environments specified
	}

	project, err := manager.CreateProject(ctx, "frontend", opts)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Should have default environments
	if len(project.Environments) != 3 {
		t.Errorf("Default environments: got %d, want 3", len(project.Environments))
	}
}

func TestGetProject(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "backend"})
	_, _ = manager.CreateProject(ctx, "backend", &ProjectOptions{Name: "api-gateway"})

	project, err := manager.GetProject(ctx, "backend", "api-gateway")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if project.Name != "api-gateway" {
		t.Errorf("Name: got %s, want api-gateway", project.Name)
	}

	// Get non-existent project
	_, err = manager.GetProject(ctx, "backend", "non-existent")
	if err == nil {
		t.Error("GetProject should fail for non-existent project")
	}

	// Get project from non-existent team
	_, err = manager.GetProject(ctx, "non-existent", "api-gateway")
	if err == nil {
		t.Error("GetProject should fail for non-existent team")
	}
}

func TestListProjects(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "backend"})
	_, _ = manager.CreateProject(ctx, "backend", &ProjectOptions{Name: "api-gateway"})
	_, _ = manager.CreateProject(ctx, "backend", &ProjectOptions{Name: "order-service"})

	projects, err := manager.ListProjects(ctx, "backend")
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("ListProjects: got %d, want 2", len(projects))
	}
}

func TestDeleteProject(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "backend"})
	_, _ = manager.CreateProject(ctx, "backend", &ProjectOptions{Name: "to-delete"})

	err := manager.DeleteProject(ctx, "backend", "to-delete")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify deleted
	_, err = manager.GetProject(ctx, "backend", "to-delete")
	if err == nil {
		t.Error("Project should not exist after deletion")
	}
}

func TestAddEnvironmentToProject(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})
	_, _ = manager.CreateProject(ctx, "frontend", &ProjectOptions{
		Name:         "web-app",
		Environments: []string{"dev"},
	})

	err := manager.AddEnvironmentToProject(ctx, "frontend", "web-app", "staging")
	if err != nil {
		t.Fatalf("AddEnvironmentToProject failed: %v", err)
	}

	project, _ := manager.GetProject(ctx, "frontend", "web-app")
	if len(project.Environments) != 2 {
		t.Errorf("Environments: got %d, want 2", len(project.Environments))
	}

	// Add duplicate environment - should not duplicate
	_ = manager.AddEnvironmentToProject(ctx, "frontend", "web-app", "staging")
	project, _ = manager.GetProject(ctx, "frontend", "web-app")
	if len(project.Environments) != 2 {
		t.Errorf("Environments should not duplicate: got %d", len(project.Environments))
	}
}

func TestGenerateNamespace(t *testing.T) {
	tests := []struct {
		org, team, project, env string
		want                    string
	}{
		{"acme", "frontend", "web-app", "dev", "frontend-web-app-dev"},
		{"", "frontend", "web-app", "dev", "frontend-web-app-dev"},
		{"", "frontend", "", "dev", "frontend-dev"},
		{"", "", "web-app", "dev", "web-app-dev"},
	}

	for _, tt := range tests {
		got := GenerateNamespace(tt.org, tt.team, tt.project, tt.env)
		if got != tt.want {
			t.Errorf("GenerateNamespace(%s,%s,%s,%s): got %s, want %s",
				tt.org, tt.team, tt.project, tt.env, got, tt.want)
		}
	}
}

func TestGenerateAppProjectName(t *testing.T) {
	tests := []struct {
		org, team string
		want      string
	}{
		{"acme", "frontend", "acme-frontend"},
		{"", "frontend", "frontend-team"},
	}

	for _, tt := range tests {
		got := GenerateAppProjectName(tt.org, tt.team)
		if got != tt.want {
			t.Errorf("GenerateAppProjectName(%s,%s): got %s, want %s",
				tt.org, tt.team, got, tt.want)
		}
	}
}

func TestValidate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})
	_, _ = manager.CreateProject(ctx, "frontend", &ProjectOptions{Name: "web-app"})

	err := manager.Validate()
	if err != nil {
		t.Errorf("Validate should pass: %v", err)
	}
}

func TestValidate_DuplicateTeam(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")

	// Manually add duplicate teams
	manager.organization.Teams = []*Team{
		{Name: "frontend"},
		{Name: "frontend"},
	}

	err := manager.Validate()
	if err == nil {
		t.Error("Validate should fail for duplicate teams")
	}
}

func TestLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	// Create and save
	manager1, _ := NewManager(configPath)
	_, _ = manager1.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager1.CreateTeam(ctx, &TeamOptions{Name: "frontend"})
	_, _ = manager1.CreateProject(ctx, "frontend", &ProjectOptions{Name: "web-app"})

	// Load in new manager
	manager2, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager (reload) failed: %v", err)
	}

	org := manager2.GetOrganization()
	if org.Name != "acme-corp" {
		t.Errorf("Name: got %s, want acme-corp", org.Name)
	}
	if len(org.Teams) != 1 {
		t.Errorf("Teams: got %d, want 1", len(org.Teams))
	}
	if len(org.Teams[0].Projects) != 1 {
		t.Errorf("Projects: got %d, want 1", len(org.Teams[0].Projects))
	}
}

func TestToYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")
	_, _ = manager.CreateTeam(ctx, &TeamOptions{Name: "frontend"})

	yamlStr, err := manager.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	if yamlStr == "" {
		t.Error("ToYAML should return non-empty string")
	}

	// Check for expected content
	if !contains(yamlStr, "name: acme-corp") {
		t.Error("YAML should contain organization name")
	}
	if !contains(yamlStr, "name: frontend") {
		t.Error("YAML should contain team name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSetPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")

	quotas := ResourceQuotaPolicy{
		CPU:    "100",
		Memory: "200Gi",
	}

	err := manager.SetPolicy(ctx, "resource-quotas", quotas)
	if err != nil {
		t.Fatalf("SetPolicy failed: %v", err)
	}

	org := manager.GetOrganization()
	if org.Policies.ResourceQuotas.CPU != "100" {
		t.Errorf("CPU: got %s, want 100", org.Policies.ResourceQuotas.CPU)
	}
}

func TestSetPolicy_InvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "organization.yaml")
	ctx := context.Background()

	manager, _ := NewManager(configPath)
	_, _ = manager.InitOrganization(ctx, "acme-corp", "acme.com")

	err := manager.SetPolicy(ctx, "resource-quotas", "invalid")
	if err == nil {
		t.Error("SetPolicy should fail for invalid value type")
	}

	err = manager.SetPolicy(ctx, "unknown-policy", nil)
	if err == nil {
		t.Error("SetPolicy should fail for unknown policy")
	}
}




