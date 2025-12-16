package git

import (
	"context"
	"testing"
)

func TestNewGitLabProvider(t *testing.T) {
	provider := NewGitLabProvider()

	if provider == nil {
		t.Fatal("NewGitLabProvider() returned nil")
	}

	if provider.Name() != ProviderGitLab {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderGitLab)
	}

	if provider.GetInstance() != "gitlab.com" {
		t.Errorf("GetInstance() = %v, want gitlab.com", provider.GetInstance())
	}
}

func TestNewGitLabProviderWithInstance(t *testing.T) {
	tests := []struct {
		name     string
		instance string
		want     string
	}{
		{"Default instance", "", "gitlab.com"},
		{"Custom instance", "gitlab.company.com", "gitlab.company.com"},
		{"Self-hosted", "git.example.org", "git.example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewGitLabProviderWithInstance(tt.instance)
			if provider.GetInstance() != tt.want {
				t.Errorf("GetInstance() = %v, want %v", provider.GetInstance(), tt.want)
			}
		})
	}
}

func TestNewGitLabProviderWithToken(t *testing.T) {
	token := "glpat-test_token"
	instance := "gitlab.company.com"
	provider := NewGitLabProviderWithToken(token, instance)

	if provider == nil {
		t.Fatal("NewGitLabProviderWithToken() returned nil")
	}

	if provider.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", provider.GetToken(), token)
	}

	if provider.GetInstance() != instance {
		t.Errorf("GetInstance() = %v, want %v", provider.GetInstance(), instance)
	}
}

func TestNewGitLabProviderWithToken_DefaultInstance(t *testing.T) {
	token := "glpat-test_token"
	provider := NewGitLabProviderWithToken(token, "")

	if provider.GetInstance() != "gitlab.com" {
		t.Errorf("GetInstance() should default to gitlab.com, got %v", provider.GetInstance())
	}
}

func TestGitLabProvider_Name(t *testing.T) {
	provider := NewGitLabProvider()

	if provider.Name() != ProviderGitLab {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderGitLab)
	}
}

func TestGitLabProvider_Capabilities(t *testing.T) {
	provider := NewGitLabProvider()
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

func TestGitLabProvider_SetToken(t *testing.T) {
	provider := NewGitLabProvider()

	if provider.GetToken() != "" {
		t.Errorf("GetToken() should be empty initially")
	}

	token := "glpat-new_token"
	provider.SetToken(token)

	if provider.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", provider.GetToken(), token)
	}
}

func TestGitLabProvider_SetInstance(t *testing.T) {
	provider := NewGitLabProvider()

	newInstance := "gitlab.newcompany.com"
	provider.SetInstance(newInstance)

	if provider.GetInstance() != newInstance {
		t.Errorf("GetInstance() = %v, want %v", provider.GetInstance(), newInstance)
	}
}

func TestGitLabProvider_Authenticate_Token(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "glpat-test_token",
	})

	if err != nil {
		t.Errorf("Authenticate() with token returned error: %v", err)
	}

	if provider.GetToken() != "glpat-test_token" {
		t.Errorf("Token not set after authentication")
	}
}

func TestGitLabProvider_Authenticate_EmptyTokenNoEnv(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	t.Setenv("GITLAB_TOKEN", "")
	t.Setenv("GL_TOKEN", "")
	t.Setenv("GIT_TOKEN", "")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err == nil {
		t.Error("Authenticate() with empty token should return error when env not set")
	}
}

func TestGitLabProvider_Authenticate_TokenFromEnv(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	t.Setenv("GITLAB_TOKEN", "glpat-env_token")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with GITLAB_TOKEN env returned error: %v", err)
	}

	if provider.GetToken() != "glpat-env_token" {
		t.Errorf("Token should be set from environment")
	}
}

func TestGitLabProvider_Authenticate_GLTokenFromEnv(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	t.Setenv("GITLAB_TOKEN", "")
	t.Setenv("GL_TOKEN", "glpat-gl_token")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with GL_TOKEN env returned error: %v", err)
	}

	if provider.GetToken() != "glpat-gl_token" {
		t.Errorf("Token should be set from GL_TOKEN environment")
	}
}

func TestGitLabProvider_Authenticate_GITTokenFromEnv(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	t.Setenv("GITLAB_TOKEN", "")
	t.Setenv("GL_TOKEN", "")
	t.Setenv("GIT_TOKEN", "generic_git_token")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with GIT_TOKEN env returned error: %v", err)
	}

	if provider.GetToken() != "generic_git_token" {
		t.Errorf("Token should be set from GIT_TOKEN environment")
	}
}

func TestGitLabProvider_Authenticate_SSH_NoKey(t *testing.T) {
	provider := NewGitLabProvider()
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

func TestGitLabProvider_Authenticate_OAuth(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthOAuth,
	})

	if err == nil {
		t.Error("Authenticate() with OAuth should return not implemented error")
	}
}

func TestGitLabProvider_Authenticate_UnsupportedMethod(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: "unsupported",
	})

	if err == nil {
		t.Error("Authenticate() with unsupported method should return error")
	}
}

func TestGitLabProvider_ValidateAccess_NotAuthenticated(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without authentication should return error")
	}
}

func TestGitLabProvider_ValidateAccess_NoToken(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	provider.auth = &AuthOptions{Method: AuthToken}

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without token should return error")
	}
}

func TestGitLabProvider_CreateRepository_NoToken(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without token should return error")
	}
}

func TestGitLabProvider_GetRepository_NoToken(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	_, err := provider.GetRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("GetRepository() without token should return error")
	}
}

func TestGitLabProvider_DeleteRepository_NoToken(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	err := provider.DeleteRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("DeleteRepository() without token should return error")
	}
}

func TestGitLabProvider_CreateWebhook_NoToken(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	_, err := provider.CreateWebhook(ctx, "owner", "repo", WebhookOptions{
		URL:    "https://example.com/webhook",
		Events: []string{"push"},
	})

	if err == nil {
		t.Error("CreateWebhook() without token should return error")
	}
}

func TestGitLabProvider_DeleteWebhook_NoToken(t *testing.T) {
	provider := NewGitLabProvider()
	ctx := context.Background()

	err := provider.DeleteWebhook(ctx, "owner", "repo", "123")

	if err == nil {
		t.Error("DeleteWebhook() without token should return error")
	}
}

func TestGitLabProvider_getGitEnv_WithToken(t *testing.T) {
	provider := NewGitLabProviderWithToken("test_token", "gitlab.com")
	env := provider.getGitEnv()

	hasGLToken := false
	hasGitLabToken := false

	for _, e := range env {
		if e == "GL_TOKEN=test_token" {
			hasGLToken = true
		}
		if e == "GITLAB_TOKEN=test_token" {
			hasGitLabToken = true
		}
	}

	if !hasGLToken {
		t.Error("getGitEnv() should include GL_TOKEN")
	}
	if !hasGitLabToken {
		t.Error("getGitEnv() should include GITLAB_TOKEN")
	}
}

func TestGitLabProvider_getGitEnv_WithSSH(t *testing.T) {
	provider := NewGitLabProvider()
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

func TestGitLabProvider_Interface(t *testing.T) {
	var _ Provider = (*GitLabProvider)(nil)
}
