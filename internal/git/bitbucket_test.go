package git

import (
	"context"
	"testing"
)

func TestNewBitbucketProvider(t *testing.T) {
	provider := NewBitbucketProvider("myworkspace")

	if provider == nil {
		t.Fatal("NewBitbucketProvider() returned nil")
	}

	if provider.Name() != ProviderBitbucket {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderBitbucket)
	}

	if provider.GetWorkspace() != "myworkspace" {
		t.Errorf("GetWorkspace() = %v, want myworkspace", provider.GetWorkspace())
	}
}

func TestNewBitbucketProviderWithToken(t *testing.T) {
	username := "testuser"
	appPassword := "app_password"
	workspace := "testworkspace"
	provider := NewBitbucketProviderWithToken(username, appPassword, workspace)

	if provider == nil {
		t.Fatal("NewBitbucketProviderWithToken() returned nil")
	}

	if provider.GetUsername() != username {
		t.Errorf("GetUsername() = %v, want %v", provider.GetUsername(), username)
	}

	if provider.GetAppPassword() != appPassword {
		t.Errorf("GetAppPassword() = %v, want %v", provider.GetAppPassword(), appPassword)
	}

	if provider.GetWorkspace() != workspace {
		t.Errorf("GetWorkspace() = %v, want %v", provider.GetWorkspace(), workspace)
	}
}

func TestBitbucketProvider_Name(t *testing.T) {
	provider := NewBitbucketProvider("")

	if provider.Name() != ProviderBitbucket {
		t.Errorf("Name() = %v, want %v", provider.Name(), ProviderBitbucket)
	}
}

func TestBitbucketProvider_Capabilities(t *testing.T) {
	provider := NewBitbucketProvider("")
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

func TestBitbucketProvider_Authenticate_Token(t *testing.T) {
	provider := NewBitbucketProvider("workspace")
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method:   AuthToken,
		Token:    "app_password",
		Username: "testuser",
	})

	if err != nil {
		t.Errorf("Authenticate() with token returned error: %v", err)
	}

	if provider.GetAppPassword() != "app_password" {
		t.Errorf("AppPassword not set after authentication")
	}

	if provider.GetUsername() != "testuser" {
		t.Errorf("Username not set after authentication")
	}
}

func TestBitbucketProvider_Authenticate_EmptyTokenNoEnv(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	t.Setenv("BITBUCKET_APP_PASSWORD", "")
	t.Setenv("BITBUCKET_TOKEN", "")
	t.Setenv("BITBUCKET_USERNAME", "testuser")
	t.Setenv("GIT_TOKEN", "")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err == nil {
		t.Error("Authenticate() with empty token should return error when env not set")
	}
}

func TestBitbucketProvider_Authenticate_NoUsername(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	t.Setenv("BITBUCKET_APP_PASSWORD", "password")
	t.Setenv("BITBUCKET_USERNAME", "")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "password",
	})

	if err == nil {
		t.Error("Authenticate() without username should return error")
	}
}

func TestBitbucketProvider_Authenticate_TokenFromEnv(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	t.Setenv("BITBUCKET_APP_PASSWORD", "bb_env_password")
	t.Setenv("BITBUCKET_USERNAME", "envuser")

	err := provider.Authenticate(ctx, AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err != nil {
		t.Errorf("Authenticate() with BITBUCKET_APP_PASSWORD env returned error: %v", err)
	}

	if provider.GetAppPassword() != "bb_env_password" {
		t.Errorf("AppPassword should be set from environment")
	}
}

func TestBitbucketProvider_Authenticate_SSH_NoKey(t *testing.T) {
	provider := NewBitbucketProvider("")
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

func TestBitbucketProvider_Authenticate_UnsupportedMethod(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	err := provider.Authenticate(ctx, AuthOptions{
		Method: "unsupported",
	})

	if err == nil {
		t.Error("Authenticate() with unsupported method should return error")
	}
}

func TestBitbucketProvider_ValidateAccess_NotAuthenticated(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	err := provider.ValidateAccess(ctx)

	if err == nil {
		t.Error("ValidateAccess() without authentication should return error")
	}
}

func TestBitbucketProvider_CreateRepository_NoCredentials(t *testing.T) {
	provider := NewBitbucketProvider("workspace")
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without credentials should return error")
	}
}

func TestBitbucketProvider_CreateRepository_NoWorkspace(t *testing.T) {
	provider := NewBitbucketProviderWithToken("user", "pass", "")
	ctx := context.Background()

	_, err := provider.CreateRepository(ctx, CreateRepoOptions{
		Name:       "test-repo",
		Visibility: VisibilityPrivate,
	})

	if err == nil {
		t.Error("CreateRepository() without workspace should return error")
	}
}

func TestBitbucketProvider_GetRepository_NoCredentials(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	_, err := provider.GetRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("GetRepository() without credentials should return error")
	}
}

func TestBitbucketProvider_DeleteRepository_NoCredentials(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	err := provider.DeleteRepository(ctx, "owner", "repo")

	if err == nil {
		t.Error("DeleteRepository() without credentials should return error")
	}
}

func TestBitbucketProvider_CreateWebhook_NoCredentials(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	_, err := provider.CreateWebhook(ctx, "owner", "repo", WebhookOptions{
		URL:    "https://example.com/webhook",
		Events: []string{"push"},
	})

	if err == nil {
		t.Error("CreateWebhook() without credentials should return error")
	}
}

func TestBitbucketProvider_DeleteWebhook_NoCredentials(t *testing.T) {
	provider := NewBitbucketProvider("")
	ctx := context.Background()

	err := provider.DeleteWebhook(ctx, "owner", "repo", "123")

	if err == nil {
		t.Error("DeleteWebhook() without credentials should return error")
	}
}

func TestBitbucketProvider_SettersGetters(t *testing.T) {
	provider := NewBitbucketProvider("")

	provider.SetUsername("new_user")
	if provider.GetUsername() != "new_user" {
		t.Errorf("GetUsername() = %v, want new_user", provider.GetUsername())
	}

	provider.SetAppPassword("new_password")
	if provider.GetAppPassword() != "new_password" {
		t.Errorf("GetAppPassword() = %v, want new_password", provider.GetAppPassword())
	}

	provider.SetWorkspace("new_workspace")
	if provider.GetWorkspace() != "new_workspace" {
		t.Errorf("GetWorkspace() = %v, want new_workspace", provider.GetWorkspace())
	}
}

func TestBitbucketProvider_Interface(t *testing.T) {
	var _ Provider = (*BitbucketProvider)(nil)
}

func TestBitbucketProvider_getGitEnv_WithCredentials(t *testing.T) {
	provider := NewBitbucketProviderWithToken("user", "password", "workspace")
	env := provider.getGitEnv()

	hasUsername := false
	hasPassword := false
	for _, e := range env {
		if e == "BITBUCKET_USERNAME=user" {
			hasUsername = true
		}
		if e == "BITBUCKET_APP_PASSWORD=password" {
			hasPassword = true
		}
	}

	if !hasUsername {
		t.Error("getGitEnv() should include BITBUCKET_USERNAME")
	}
	if !hasPassword {
		t.Error("getGitEnv() should include BITBUCKET_APP_PASSWORD")
	}
}

func TestBitbucketProvider_getGitEnv_WithSSH(t *testing.T) {
	provider := NewBitbucketProvider("")
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
