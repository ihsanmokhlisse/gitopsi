package git

import (
	"context"
	"testing"
)

func TestNewGiteaProvider(t *testing.T) {
	tests := []struct {
		name     string
		instance string
		want     string
	}{
		{"Default instance", "", "gitea.com"},
		{"Custom instance", "gitea.company.com", "gitea.company.com"},
		{"Self-hosted", "git.example.org", "git.example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewGiteaProvider(tt.instance)
			if provider.GetInstance() != tt.want {
				t.Errorf("GetInstance() = %v, want %v", provider.GetInstance(), tt.want)
			}
		})
	}
}

func TestNewGiteaProviderWithToken(t *testing.T) {
	token := "gitea_test_token"
	instance := "gitea.company.com"
	provider := NewGiteaProviderWithToken(token, instance)

	if provider == nil {
		t.Fatal("NewGiteaProviderWithToken() returned nil")
	}

	if provider.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", provider.GetToken(), token)
	}

	if provider.GetInstance() != instance {
		t.Errorf("GetInstance() = %v, want %v", provider.GetInstance(), instance)
	}
}

func TestGiteaProvider_Name(t *testing.T) {
	provider := NewGiteaProvider("")

	if provider.Name() != ProviderGitea {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderGitea)
	}
}

func TestGiteaProvider_Capabilities(t *testing.T) {
	provider := NewGiteaProvider("")
	caps := provider.Capabilities()

	expectedCaps := []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
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

func TestGiteaProvider_Authenticate_Token(t *testing.T) {
	provider := NewGiteaProvider("gitea.example.com")
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "test_token",
	})

	if err != nil {
		t.Errorf("Authenticate() with token returned error: %v", err)
	}

	if provider.GetToken() != "test_token" {
		t.Errorf("Token not set after authentication")
	}
}

func TestGiteaProvider_Authenticate_EmptyTokenNoEnv(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	t.Setenv("GITEA_TOKEN", "")
	t.Setenv("GIT_TOKEN", "")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err == nil {
		t.Error("Authenticate() with empty token should return error when env not set")
	}
}

func TestGiteaProvider_Authenticate_TokenFromEnv(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	t.Setenv("GITEA_TOKEN", "gitea_env_token")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with GITEA_TOKEN env returned error: %v", err)
	}

	if provider.GetToken() != "gitea_env_token" {
		t.Errorf("Token should be set from environment")
	}
}

func TestGiteaProvider_Authenticate_SSH_NoKey(t *testing.T) {
	provider := NewGiteaProvider("")
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

func TestGiteaProvider_Authenticate_UnsupportedMethod(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: "unsupported",
	})

	if err == nil {
		t.Error("Authenticate() with unsupported method should return error")
	}
}

func TestGiteaProvider_ValidateAccess_NotAuthenticated(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without authentication should return error")
	}
}

func TestGiteaProvider_CreateRepository_NoToken(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without token should return error")
	}
}

func TestGiteaProvider_GetRepository_NoToken(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	_, err := provider.GetRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("GetRepository() without token should return error")
	}
}

func TestGiteaProvider_DeleteRepository_NoToken(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	err := provider.DeleteRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("DeleteRepository() without token should return error")
	}
}

func TestGiteaProvider_CreateWebhook_NoToken(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	_, err := provider.CreateWebhook(ctx, "owner", "repo", WebhookOptions{
		URL:    "https://example.com/webhook",
		Events: []string{"push"},
	})

	if err == nil {
		t.Error("CreateWebhook() without token should return error")
	}
}

func TestGiteaProvider_DeleteWebhook_NoToken(t *testing.T) {
	provider := NewGiteaProvider("")
	ctx := context.Background()

	err := provider.DeleteWebhook(ctx, "owner", "repo", "123")

	if err == nil {
		t.Error("DeleteWebhook() without token should return error")
	}
}

func TestGiteaProvider_SettersGetters(t *testing.T) {
	provider := NewGiteaProvider("")

	provider.SetToken("new_token")
	if provider.GetToken() != "new_token" {
		t.Errorf("GetToken() = %v, want new_token", provider.GetToken())
	}

	provider.SetInstance("new.gitea.com")
	if provider.GetInstance() != "new.gitea.com" {
		t.Errorf("GetInstance() = %v, want new.gitea.com", provider.GetInstance())
	}
}

func TestGiteaProvider_Interface(t *testing.T) {
	var _ Provider = (*GiteaProvider)(nil)
}

func TestGiteaProvider_getGitEnv_WithToken(t *testing.T) {
	provider := NewGiteaProviderWithToken("test_token", "gitea.com")
	env := provider.getGitEnv()

	hasGiteaToken := false
	for _, e := range env {
		if e == "GITEA_TOKEN=test_token" {
			hasGiteaToken = true
		}
	}

	if !hasGiteaToken {
		t.Error("getGitEnv() should include GITEA_TOKEN")
	}
}

func TestGiteaProvider_getGitEnv_WithSSH(t *testing.T) {
	provider := NewGiteaProvider("")
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

// Tests for Clone, Push, Pull, TestConnection

func TestGiteaProvider_Clone_Options(t *testing.T) {
	provider := NewGiteaProviderWithToken("test_token", "gitea.com")

	tests := []struct {
		name string
		opts CloneOptions
	}{
		{
			name: "basic clone",
			opts: CloneOptions{
				URL:  "https://gitea.com/owner/repo.git",
				Path: "/tmp/test-repo",
			},
		},
		{
			name: "clone with branch",
			opts: CloneOptions{
				URL:    "https://gitea.com/owner/repo.git",
				Path:   "/tmp/test-repo",
				Branch: "develop",
			},
		},
		{
			name: "clone with depth",
			opts: CloneOptions{
				URL:   "https://gitea.com/owner/repo.git",
				Path:  "/tmp/test-repo",
				Depth: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_ = provider.Clone(ctx, tt.opts)
		})
	}
}

func TestGiteaProvider_Push_Options(t *testing.T) {
	provider := NewGiteaProviderWithToken("test_token", "gitea.com")

	tests := []struct {
		name string
		opts PushOptions
	}{
		{
			name: "basic push",
			opts: PushOptions{
				Path:   "/tmp/test-repo",
				Remote: "origin",
				Branch: "main",
			},
		},
		{
			name: "push with force",
			opts: PushOptions{
				Path:   "/tmp/test-repo",
				Remote: "origin",
				Branch: "feature/test",
				Force:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_ = provider.Push(ctx, tt.opts)
		})
	}
}

func TestGiteaProvider_Pull_Options(t *testing.T) {
	provider := NewGiteaProviderWithToken("test_token", "gitea.com")

	tests := []struct {
		name string
		opts PullOptions
	}{
		{
			name: "basic pull",
			opts: PullOptions{
				Path:   "/tmp/test-repo",
				Remote: "origin",
				Branch: "main",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_ = provider.Pull(ctx, tt.opts)
		})
	}
}

func TestGiteaProvider_TestConnection_Behavior(t *testing.T) {
	provider := NewGiteaProviderWithToken("test_token", "gitea.com")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := provider.TestConnection(ctx)
	if err == nil {
		t.Log("TestConnection returned nil with canceled context")
	}
}
