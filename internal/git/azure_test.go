package git

import (
	"context"
	"testing"
)

func TestNewAzureDevOpsProvider(t *testing.T) {
	provider := NewAzureDevOpsProvider("myorg", "myproject")

	if provider == nil {
		t.Fatal("NewAzureDevOpsProvider() returned nil")
	}

	if provider.Name() != ProviderAzureDevOps {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderAzureDevOps)
	}

	if provider.GetOrganization() != "myorg" {
		t.Errorf("GetOrganization() = %v, want myorg", provider.GetOrganization())
	}

	if provider.GetProject() != "myproject" {
		t.Errorf("GetProject() = %v, want myproject", provider.GetProject())
	}
}

func TestNewAzureDevOpsProviderWithToken(t *testing.T) {
	token := "azure_pat_token"
	provider := NewAzureDevOpsProviderWithToken(token, "org", "proj")

	if provider == nil {
		t.Fatal("NewAzureDevOpsProviderWithToken() returned nil")
	}

	if provider.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", provider.GetToken(), token)
	}
}

func TestAzureDevOpsProvider_Name(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")

	if provider.Name() != ProviderAzureDevOps {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderAzureDevOps)
	}
}

func TestAzureDevOpsProvider_Capabilities(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	caps := provider.Capabilities()

	expectedCaps := []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityCICD,
	}

	if len(caps) != len(expectedCaps) {
		t.Errorf("Capabilities() returned %d capabilities, want %d", len(caps), len(expectedCaps))
	}

	capMap := make(map[Capability]bool)
	for _, cap := range caps {
		capMap[cap] = true
	}

	for _, expected := range expectedCaps {
		if !capMap[expected] {
			t.Errorf("Capabilities() missing %v", expected)
		}
	}
}

func TestAzureDevOpsProvider_Authenticate_Token(t *testing.T) {
	provider := NewAzureDevOpsProvider("org", "proj")
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "test_pat",
	})

	if err != nil {
		t.Errorf("Authenticate() with token returned error: %v", err)
	}

	if provider.GetToken() != "test_pat" {
		t.Errorf("Token not set after authentication")
	}
}

func TestAzureDevOpsProvider_Authenticate_EmptyTokenNoEnv(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	t.Setenv("AZURE_DEVOPS_PAT", "")
	t.Setenv("AZURE_DEVOPS_TOKEN", "")
	t.Setenv("GIT_TOKEN", "")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err == nil {
		t.Error("Authenticate() with empty token should return error when env not set")
	}
}

func TestAzureDevOpsProvider_Authenticate_TokenFromEnv(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	t.Setenv("AZURE_DEVOPS_PAT", "azure_env_pat")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with AZURE_DEVOPS_PAT env returned error: %v", err)
	}

	if provider.GetToken() != "azure_env_pat" {
		t.Errorf("Token should be set from environment")
	}
}

func TestAzureDevOpsProvider_Authenticate_SSH_NoKey(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	t.Setenv("HOME", "/nonexistent")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthSSH,
		SSHKey: "/nonexistent/key",
	})

	if err == nil {
		t.Error("Authenticate() with non-existent SSH key should return error")
	}
}

func TestAzureDevOpsProvider_Authenticate_UnsupportedMethod(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: "unsupported",
	})

	if err == nil {
		t.Error("Authenticate() with unsupported method should return error")
	}
}

func TestAzureDevOpsProvider_ValidateAccess_NotAuthenticated(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without authentication should return error")
	}
}

func TestAzureDevOpsProvider_CreateRepository_NoToken(t *testing.T) {
	provider := NewAzureDevOpsProvider("org", "proj")
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without token should return error")
	}
}

func TestAzureDevOpsProvider_CreateRepository_NoOrgProject(t *testing.T) {
	provider := NewAzureDevOpsProviderWithToken("token", "", "")
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without org/project should return error")
	}
}

func TestAzureDevOpsProvider_GetRepository_NoToken(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	_, err := provider.GetRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("GetRepository() without token should return error")
	}
}

func TestAzureDevOpsProvider_DeleteRepository_NoToken(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	err := provider.DeleteRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("DeleteRepository() without token should return error")
	}
}

func TestAzureDevOpsProvider_CreateWebhook_NoToken(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	_, err := provider.CreateWebhook(ctx, "owner", "repo", WebhookOptions{
		URL:    "https://example.com/webhook",
		Events: []string{"push"},
	})

	if err == nil {
		t.Error("CreateWebhook() without token should return error")
	}
}

func TestAzureDevOpsProvider_DeleteWebhook_NoToken(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")
	ctx := context.Background()

	err := provider.DeleteWebhook(ctx, "owner", "repo", "123")

	if err == nil {
		t.Error("DeleteWebhook() without token should return error")
	}
}

func TestAzureDevOpsProvider_SettersGetters(t *testing.T) {
	provider := NewAzureDevOpsProvider("", "")

	provider.SetToken("new_token")
	if provider.GetToken() != "new_token" {
		t.Errorf("GetToken() = %v, want new_token", provider.GetToken())
	}

	provider.SetOrganization("new-org")
	if provider.GetOrganization() != "new-org" {
		t.Errorf("GetOrganization() = %v, want new-org", provider.GetOrganization())
	}

	provider.SetProject("new-project")
	if provider.GetProject() != "new-project" {
		t.Errorf("GetProject() = %v, want new-project", provider.GetProject())
	}
}

func TestAzureDevOpsProvider_Interface(t *testing.T) {
	var _ Provider = (*AzureDevOpsProvider)(nil)
}

