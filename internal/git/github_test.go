package git

import (
	"context"
	"testing"
)

func TestNewGitHubProvider(t *testing.T) {
	provider := NewGitHubProvider()

	if provider == nil {
		t.Fatal("NewGitHubProvider() returned nil")
	}

	if provider.Name() != ProviderGitHub {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderGitHub)
	}

	if provider.GetInstance() != "github.com" {
		t.Errorf("GetInstance() = %v, want github.com", provider.GetInstance())
	}
}

func TestNewGitHubProviderWithToken(t *testing.T) {
	token := "ghp_test_token"
	provider := NewGitHubProviderWithToken(token)

	if provider == nil {
		t.Fatal("NewGitHubProviderWithToken() returned nil")
	}

	if provider.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", provider.GetToken(), token)
	}
}

func TestGitHubProvider_Name(t *testing.T) {
	provider := NewGitHubProvider()

	if provider.Name() != ProviderGitHub {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderGitHub)
	}
}

func TestGitHubProvider_Capabilities(t *testing.T) {
	provider := NewGitHubProvider()
	caps := provider.Capabilities()

	expectedCaps := []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityBranchProtection,
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

func TestGitHubProvider_SetToken(t *testing.T) {
	provider := NewGitHubProvider()

	if provider.GetToken() != "" {
		t.Errorf("GetToken() should be empty initially")
	}

	token := "ghp_new_token"
	provider.SetToken(token)

	if provider.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", provider.GetToken(), token)
	}
}

func TestGitHubProvider_Authenticate_Token(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "ghp_test_token",
	})

	if err != nil {
		t.Errorf("Authenticate() with token returned error: %v", err)
	}

	if provider.GetToken() != "ghp_test_token" {
		t.Errorf("Token not set after authentication")
	}
}

func TestGitHubProvider_Authenticate_EmptyTokenNoEnv(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err == nil {
		t.Error("Authenticate() with empty token should return error when env not set")
	}
}

func TestGitHubProvider_Authenticate_TokenFromEnv(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	t.Setenv("GITHUB_TOKEN", "ghp_env_token")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with GITHUB_TOKEN env returned error: %v", err)
	}

	if provider.GetToken() != "ghp_env_token" {
		t.Errorf("Token should be set from environment")
	}
}

func TestGitHubProvider_Authenticate_GHTokenFromEnv(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "ghp_gh_token")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with GH_TOKEN env returned error: %v", err)
	}

	if provider.GetToken() != "ghp_gh_token" {
		t.Errorf("Token should be set from GH_TOKEN environment")
	}
}

func TestGitHubProvider_Authenticate_SSH_NoKey(t *testing.T) {
	provider := NewGitHubProvider()
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

func TestGitHubProvider_Authenticate_OAuth(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthOAuth,
	})

	if err == nil {
		t.Error("Authenticate() with OAuth should return not implemented error")
	}
}

func TestGitHubProvider_Authenticate_UnsupportedMethod(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: "unsupported",
	})

	if err == nil {
		t.Error("Authenticate() with unsupported method should return error")
	}
}

func TestGitHubProvider_ValidateAccess_NotAuthenticated(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without authentication should return error")
	}
}

func TestGitHubProvider_ValidateAccess_NoToken(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	provider.auth = &AuthOptions{Method: AuthToken}

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without token should return error")
	}
}

func TestGitHubProvider_CreateRepository_NoToken(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without token should return error")
	}
}

func TestGitHubProvider_GetRepository_NoToken(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	_, err := provider.GetRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("GetRepository() without token should return error")
	}
}

func TestGitHubProvider_DeleteRepository_NoToken(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	err := provider.DeleteRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("DeleteRepository() without token should return error")
	}
}

func TestGitHubProvider_CreateWebhook_NoToken(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	_, err := provider.CreateWebhook(ctx, "owner", "repo", WebhookOptions{
		URL:    "https://example.com/webhook",
		Events: []string{"push"},
	})

	if err == nil {
		t.Error("CreateWebhook() without token should return error")
	}
}

func TestGitHubProvider_DeleteWebhook_NoToken(t *testing.T) {
	provider := NewGitHubProvider()
	ctx := context.Background()

	err := provider.DeleteWebhook(ctx, "owner", "repo", "123")

	if err == nil {
		t.Error("DeleteWebhook() without token should return error")
	}
}

func TestGitHubProvider_getGitEnv_WithToken(t *testing.T) {
	provider := NewGitHubProviderWithToken("test_token")
	env := provider.getGitEnv()

	hasGHToken := false
	hasGitHubToken := false

	for _, e := range env {
		if e == "GH_TOKEN=test_token" {
			hasGHToken = true
		}
		if e == "GITHUB_TOKEN=test_token" {
			hasGitHubToken = true
		}
	}

	if !hasGHToken {
		t.Error("getGitEnv() should include GH_TOKEN")
	}
	if !hasGitHubToken {
		t.Error("getGitEnv() should include GITHUB_TOKEN")
	}
}

func TestGitHubProvider_getGitEnv_WithSSH(t *testing.T) {
	provider := NewGitHubProvider()
	provider.auth = &AuthOptions{
		Method: AuthSSH,
		SSHKey: "/path/to/key",
	}

	env := provider.getGitEnv()

	hasSSHCommand := false
	for _, e := range env {
		if e == "GIT_SSH_COMMAND=ssh -i /path/to/key -o StrictHostKeyChecking=no" {
			hasSSHCommand = true
		}
	}

	if !hasSSHCommand {
		t.Error("getGitEnv() should include GIT_SSH_COMMAND when using SSH auth")
	}
}

func TestGitHubProvider_Interface(t *testing.T) {
	var _ Provider = (*GitHubProvider)(nil)
}
