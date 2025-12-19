package auth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCredentialTypes(t *testing.T) {
	tests := []struct {
		credType CredentialType
		expected string
	}{
		{CredentialTypeGit, "git"},
		{CredentialTypePlatform, "platform"},
		{CredentialTypeRegistry, "registry"},
	}

	for _, tt := range tests {
		if string(tt.credType) != tt.expected {
			t.Errorf("CredentialType mismatch: got %s, want %s", tt.credType, tt.expected)
		}
	}
}

func TestMethods(t *testing.T) {
	tests := []struct {
		method   Method
		expected string
	}{
		{MethodSSH, "ssh"},
		{MethodToken, "token"},
		{MethodBasic, "basic"},
		{MethodOAuth, "oauth"},
		{MethodServiceAccount, "service-account"},
		{MethodOIDC, "oidc"},
		{MethodAWSIRSA, "aws-irsa"},
		{MethodAzureAAD, "azure-aad"},
	}

	for _, tt := range tests {
		if string(tt.method) != tt.expected {
			t.Errorf("Method mismatch: got %s, want %s", tt.method, tt.expected)
		}
	}
}

func TestGitProviders(t *testing.T) {
	tests := []struct {
		provider GitProvider
		expected string
	}{
		{GitProviderGitHub, "github"},
		{GitProviderGitLab, "gitlab"},
		{GitProviderBitbucket, "bitbucket"},
		{GitProviderAzureDevOps, "azure-devops"},
		{GitProviderGitea, "gitea"},
	}

	for _, tt := range tests {
		if string(tt.provider) != tt.expected {
			t.Errorf("GitProvider mismatch: got %s, want %s", tt.provider, tt.expected)
		}
	}
}

func TestPlatformTypes(t *testing.T) {
	tests := []struct {
		platform PlatformType
		expected string
	}{
		{PlatformKubernetes, "kubernetes"},
		{PlatformOpenShift, "openshift"},
		{PlatformAWS, "aws"},
		{PlatformAzure, "azure"},
		{PlatformGCP, "gcp"},
	}

	for _, tt := range tests {
		if string(tt.platform) != tt.expected {
			t.Errorf("PlatformType mismatch: got %s, want %s", tt.platform, tt.expected)
		}
	}
}

func TestSecretFormats(t *testing.T) {
	tests := []struct {
		format   SecretFormat
		expected string
	}{
		{SecretFormatPlain, "plain"},
		{SecretFormatSealed, "sealed"},
		{SecretFormatSops, "sops"},
		{SecretFormatExternalSecret, "external-secret"},
		{SecretFormatVault, "vault"},
	}

	for _, tt := range tests {
		if string(tt.format) != tt.expected {
			t.Errorf("SecretFormat mismatch: got %s, want %s", tt.format, tt.expected)
		}
	}
}

func TestNewManager(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.store == nil {
		t.Fatal("Manager store is nil")
	}
	if manager.secretFormat != SecretFormatPlain {
		t.Errorf("Manager secretFormat: got %s, want %s", manager.secretFormat, SecretFormatPlain)
	}
}

func TestGitCredentialOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    GitCredentialOptions
		wantErr bool
	}{
		{
			name: "valid token auth",
			opts: GitCredentialOptions{
				Name:     "test",
				Provider: GitProviderGitHub,
				Method:   MethodToken,
				Token:    "ghp_xxxxxxxxxxxx",
			},
			wantErr: false,
		},
		{
			name: "valid ssh auth",
			opts: GitCredentialOptions{
				Name:          "test",
				Provider:      GitProviderGitLab,
				Method:        MethodSSH,
				SSHPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
			},
			wantErr: false,
		},
		{
			name: "valid basic auth",
			opts: GitCredentialOptions{
				Name:     "test",
				Provider: GitProviderBitbucket,
				Method:   MethodBasic,
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			opts: GitCredentialOptions{
				Provider: GitProviderGitHub,
				Method:   MethodToken,
				Token:    "test",
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			opts: GitCredentialOptions{
				Name:   "test",
				Method: MethodToken,
				Token:  "test",
			},
			wantErr: true,
		},
		{
			name: "missing method",
			opts: GitCredentialOptions{
				Name:     "test",
				Provider: GitProviderGitHub,
				Token:    "test",
			},
			wantErr: true,
		},
		{
			name: "token auth without token",
			opts: GitCredentialOptions{
				Name:     "test",
				Provider: GitProviderGitHub,
				Method:   MethodToken,
			},
			wantErr: true,
		},
		{
			name: "ssh auth without key",
			opts: GitCredentialOptions{
				Name:     "test",
				Provider: GitProviderGitHub,
				Method:   MethodSSH,
			},
			wantErr: true,
		},
		{
			name: "basic auth without password",
			opts: GitCredentialOptions{
				Name:     "test",
				Provider: GitProviderGitHub,
				Method:   MethodBasic,
				Username: "user",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlatformCredentialOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    PlatformCredentialOptions
		wantErr bool
	}{
		{
			name: "valid token auth",
			opts: PlatformCredentialOptions{
				Name:     "test",
				Platform: PlatformOpenShift,
				Method:   MethodToken,
				Token:    "sha256~xxxx",
			},
			wantErr: false,
		},
		{
			name: "valid aws irsa",
			opts: PlatformCredentialOptions{
				Name:       "test",
				Platform:   PlatformAWS,
				Method:     MethodAWSIRSA,
				AWSRoleARN: "arn:aws:iam::123456789:role/test",
			},
			wantErr: false,
		},
		{
			name: "valid azure aad",
			opts: PlatformCredentialOptions{
				Name:          "test",
				Platform:      PlatformAzure,
				Method:        MethodAzureAAD,
				AzureTenantID: "tenant-123",
				AzureClientID: "client-456",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			opts: PlatformCredentialOptions{
				Platform: PlatformOpenShift,
				Method:   MethodToken,
				Token:    "test",
			},
			wantErr: true,
		},
		{
			name: "token auth without token",
			opts: PlatformCredentialOptions{
				Name:     "test",
				Platform: PlatformKubernetes,
				Method:   MethodToken,
			},
			wantErr: true,
		},
		{
			name: "aws irsa without role arn",
			opts: PlatformCredentialOptions{
				Name:     "test",
				Platform: PlatformAWS,
				Method:   MethodAWSIRSA,
			},
			wantErr: true,
		},
		{
			name: "azure aad without tenant",
			opts: PlatformCredentialOptions{
				Name:          "test",
				Platform:      PlatformAzure,
				Method:        MethodAzureAAD,
				AzureClientID: "client-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryCredentialOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    RegistryCredentialOptions
		wantErr bool
	}{
		{
			name: "valid registry cred",
			opts: RegistryCredentialOptions{
				Name:     "test",
				URL:      "docker.io",
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			opts: RegistryCredentialOptions{
				URL:      "docker.io",
				Username: "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing url",
			opts: RegistryCredentialOptions{
				Name:     "test",
				Username: "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing credentials",
			opts: RegistryCredentialOptions{
				Name: "test",
				URL:  "docker.io",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_AddGitCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	opts := &GitCredentialOptions{
		Name:     "github-test",
		Provider: GitProviderGitHub,
		Method:   MethodToken,
		Token:    "ghp_xxxxxxxxxxxx",
		URL:      "https://github.com/test/repo",
	}

	cred, err := manager.AddGitCredential(ctx, opts)
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	if cred.Name != opts.Name {
		t.Errorf("Name: got %s, want %s", cred.Name, opts.Name)
	}
	if cred.Type != CredentialTypeGit {
		t.Errorf("Type: got %s, want %s", cred.Type, CredentialTypeGit)
	}
	if cred.Provider != string(opts.Provider) {
		t.Errorf("Provider: got %s, want %s", cred.Provider, opts.Provider)
	}
	if cred.Data.Token != opts.Token {
		t.Errorf("Token: got %s, want %s", cred.Data.Token, opts.Token)
	}
}

func TestManager_AddPlatformCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	opts := &PlatformCredentialOptions{
		Name:     "openshift-test",
		Platform: PlatformOpenShift,
		Method:   MethodToken,
		Token:    "sha256~xxxxxxxxxxxx",
		URL:      "https://api.cluster.example.com:6443",
	}

	cred, err := manager.AddPlatformCredential(ctx, opts)
	if err != nil {
		t.Fatalf("AddPlatformCredential failed: %v", err)
	}

	if cred.Name != opts.Name {
		t.Errorf("Name: got %s, want %s", cred.Name, opts.Name)
	}
	if cred.Type != CredentialTypePlatform {
		t.Errorf("Type: got %s, want %s", cred.Type, CredentialTypePlatform)
	}
	if cred.Data.Token != opts.Token {
		t.Errorf("Token: got %s, want %s", cred.Data.Token, opts.Token)
	}
}

func TestManager_AddRegistryCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	opts := &RegistryCredentialOptions{
		Name:     "registry-test",
		URL:      "docker.io",
		Username: "testuser",
		Password: "testpass",
	}

	cred, err := manager.AddRegistryCredential(ctx, opts)
	if err != nil {
		t.Fatalf("AddRegistryCredential failed: %v", err)
	}

	if cred.Name != opts.Name {
		t.Errorf("Name: got %s, want %s", cred.Name, opts.Name)
	}
	if cred.Type != CredentialTypeRegistry {
		t.Errorf("Type: got %s, want %s", cred.Type, CredentialTypeRegistry)
	}
	if cred.Data.Username != opts.Username {
		t.Errorf("Username: got %s, want %s", cred.Data.Username, opts.Username)
	}
}

func TestManager_GetCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	opts := &GitCredentialOptions{
		Name:     "test-cred",
		Provider: GitProviderGitHub,
		Method:   MethodToken,
		Token:    "test-token",
	}

	_, err := manager.AddGitCredential(ctx, opts)
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	// Get existing credential
	cred, err := manager.GetCredential(ctx, "test-cred")
	if err != nil {
		t.Fatalf("GetCredential failed: %v", err)
	}
	if cred.Name != "test-cred" {
		t.Errorf("Name: got %s, want %s", cred.Name, "test-cred")
	}

	// Get non-existing credential
	_, err = manager.GetCredential(ctx, "non-existing")
	if err == nil {
		t.Error("GetCredential should fail for non-existing credential")
	}
}

func TestManager_ListCredentials(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	// Add git credential
	_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
		Name:     "git-cred",
		Provider: GitProviderGitHub,
		Method:   MethodToken,
		Token:    "test-token",
	})
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	// Add platform credential
	_, err = manager.AddPlatformCredential(ctx, &PlatformCredentialOptions{
		Name:     "platform-cred",
		Platform: PlatformOpenShift,
		Method:   MethodToken,
		Token:    "test-token",
	})
	if err != nil {
		t.Fatalf("AddPlatformCredential failed: %v", err)
	}

	// List all
	all, err := manager.ListCredentials(ctx, "")
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("ListCredentials: got %d, want 2", len(all))
	}

	// List git only
	gitCreds, err := manager.ListCredentials(ctx, CredentialTypeGit)
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if len(gitCreds) != 1 {
		t.Errorf("ListCredentials git: got %d, want 1", len(gitCreds))
	}
}

func TestManager_DeleteCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
		Name:     "to-delete",
		Provider: GitProviderGitHub,
		Method:   MethodToken,
		Token:    "test-token",
	})
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	// Delete existing
	err = manager.DeleteCredential(ctx, "to-delete")
	if err != nil {
		t.Fatalf("DeleteCredential failed: %v", err)
	}

	// Verify deleted
	_, err = manager.GetCredential(ctx, "to-delete")
	if err == nil {
		t.Error("Credential should not exist after deletion")
	}

	// Delete non-existing
	err = manager.DeleteCredential(ctx, "non-existing")
	if err == nil {
		t.Error("DeleteCredential should fail for non-existing credential")
	}
}

func TestManager_TestCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	// Add valid token credential
	_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
		Name:     "valid-token",
		Provider: GitProviderGitHub,
		Method:   MethodToken,
		Token:    "ghp_xxxxxxxxxxxxxxxxxxxx",
	})
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	result, err := manager.TestCredential(ctx, "valid-token")
	if err != nil {
		t.Fatalf("TestCredential failed: %v", err)
	}
	if !result.Success {
		t.Errorf("TestCredential should succeed for valid token")
	}

	// Test non-existing credential
	_, err = manager.TestCredential(ctx, "non-existing")
	if err == nil {
		t.Error("TestCredential should fail for non-existing credential")
	}
}

func TestManager_GenerateKubernetesSecret(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	tests := []struct {
		name       string
		credType   CredentialType
		method     Method
		setupCred  func() error
		wantFields []string
	}{
		{
			name:     "git token secret",
			credType: CredentialTypeGit,
			method:   MethodToken,
			setupCred: func() error {
				_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
					Name:       "git-secret",
					Provider:   GitProviderGitHub,
					Method:     MethodToken,
					Token:      "test-token",
					Username:   "testuser",
					Namespace:  "default",
					SecretName: "my-git-secret",
				})
				return err
			},
			wantFields: []string{"kind: Secret", "name: my-git-secret", "namespace: default"},
		},
		{
			name:     "git ssh secret",
			credType: CredentialTypeGit,
			method:   MethodSSH,
			setupCred: func() error {
				_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
					Name:          "git-ssh",
					Provider:      GitProviderGitLab,
					Method:        MethodSSH,
					SSHPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
					Namespace:     "gitlab",
				})
				return err
			},
			wantFields: []string{"kind: Secret", "ssh-privatekey"},
		},
		{
			name:     "registry secret",
			credType: CredentialTypeRegistry,
			setupCred: func() error {
				_, err := manager.AddRegistryCredential(ctx, &RegistryCredentialOptions{
					Name:      "registry-secret",
					URL:       "docker.io",
					Username:  "user",
					Password:  "pass",
					Namespace: "registry-ns",
				})
				return err
			},
			wantFields: []string{"kind: Secret", "kubernetes.io/dockerconfigjson", ".dockerconfigjson"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setupCred(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Get credential name from test
			var credName string
			switch tt.name {
			case "git token secret":
				credName = "git-secret"
			case "git ssh secret":
				credName = "git-ssh"
			case "registry secret":
				credName = "registry-secret"
			}

			output, err := manager.GenerateKubernetesSecret(ctx, credName)
			if err != nil {
				t.Fatalf("GenerateKubernetesSecret failed: %v", err)
			}

			for _, field := range tt.wantFields {
				if !strings.Contains(output, field) {
					t.Errorf("Output missing expected field: %s", field)
				}
			}
		})
	}
}

func TestManager_GenerateArgoCDRepoSecret(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
		Name:     "argocd-repo",
		Provider: GitProviderGitHub,
		Method:   MethodToken,
		Token:    "test-token",
		Username: "git",
		URL:      "https://github.com/test/repo",
	})
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	output, err := manager.GenerateArgoCDRepoSecret(ctx, "argocd-repo", "argocd")
	if err != nil {
		t.Fatalf("GenerateArgoCDRepoSecret failed: %v", err)
	}

	wantFields := []string{
		"kind: Secret",
		"namespace: argocd",
		"argocd.argoproj.io/secret-type: repository",
		"type: git",
		"url: https://github.com/test/repo",
	}

	for _, field := range wantFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output missing expected field: %s", field)
		}
	}
}

func TestManager_GenerateFluxGitRepositorySecret(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store, SecretFormatPlain)
	ctx := context.Background()

	_, err := manager.AddGitCredential(ctx, &GitCredentialOptions{
		Name:          "flux-repo",
		Provider:      GitProviderGitLab,
		Method:        MethodSSH,
		SSHPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
		SSHPublicKey:  "ssh-rsa AAAA...",
	})
	if err != nil {
		t.Fatalf("AddGitCredential failed: %v", err)
	}

	output, err := manager.GenerateFluxGitRepositorySecret(ctx, "flux-repo", "flux-system")
	if err != nil {
		t.Fatalf("GenerateFluxGitRepositorySecret failed: %v", err)
	}

	wantFields := []string{
		"kind: Secret",
		"namespace: flux-system",
		"identity:",
		"identity.pub:",
	}

	for _, field := range wantFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output missing expected field: %s", field)
		}
	}
}

func TestLoadSSHKeyFromFile(t *testing.T) {
	// Create temp file with SSH key
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")
	keyContent := "-----BEGIN RSA PRIVATE KEY-----\ntest content\n-----END RSA PRIVATE KEY-----"

	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		t.Fatalf("Failed to write temp key: %v", err)
	}

	// Test loading
	loaded, err := LoadSSHKeyFromFile(keyPath)
	if err != nil {
		t.Fatalf("LoadSSHKeyFromFile failed: %v", err)
	}
	if loaded != keyContent {
		t.Errorf("Content mismatch: got %q, want %q", loaded, keyContent)
	}

	// Test non-existing file
	_, err = LoadSSHKeyFromFile("/non/existing/path")
	if err == nil {
		t.Error("LoadSSHKeyFromFile should fail for non-existing file")
	}
}

func TestMaskCredential(t *testing.T) {
	tests := []struct {
		input    string
		wantLen  int
		wantMask bool
	}{
		{"short", 5, false},            // <= 8 chars: all masked
		{"ghp_xxxxxxxxxxxx", 16, true}, // > 8 chars: partial mask
		{"a", 1, false},
	}

	for _, tt := range tests {
		result := MaskCredential(tt.input)
		if len(result) != tt.wantLen {
			t.Errorf("MaskCredential(%q): got len %d, want %d", tt.input, len(result), tt.wantLen)
		}
		if tt.wantMask {
			// Should have visible chars at start and end
			if !strings.HasPrefix(result, tt.input[:4]) {
				t.Errorf("MaskCredential should preserve first 4 chars")
			}
			if !strings.HasSuffix(result, tt.input[len(tt.input)-4:]) {
				t.Errorf("MaskCredential should preserve last 4 chars")
			}
		}
	}
}

func TestGetTokenFromEnv(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_TOKEN_VAR", "test-value")
	defer os.Unsetenv("TEST_TOKEN_VAR")

	result := GetTokenFromEnv("TEST_TOKEN_VAR")
	if result != "test-value" {
		t.Errorf("GetTokenFromEnv: got %q, want %q", result, "test-value")
	}

	// Test non-existing
	result = GetTokenFromEnv("NON_EXISTING_VAR")
	if result != "" {
		t.Errorf("GetTokenFromEnv should return empty for non-existing var")
	}
}

func TestCredential_Fields(t *testing.T) {
	now := time.Now()
	expires := now.Add(24 * time.Hour)

	cred := &Credential{
		Name:     "test",
		Type:     CredentialTypeGit,
		Provider: "github",
		Method:   MethodToken,
		Data: CredentialData{
			Token:    "test-token",
			Username: "testuser",
		},
		Metadata: CredentialMetadata{
			Description: "Test credential",
			URL:         "https://github.com/test/repo",
			Namespace:   "default",
			SecretName:  "my-secret",
			Labels:      map[string]string{"app": "test"},
			Annotations: map[string]string{"note": "test"},
			ExpiresAt:   &expires,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if cred.Name != "test" {
		t.Errorf("Name: got %s, want test", cred.Name)
	}
	if cred.Data.Token != "test-token" {
		t.Errorf("Token: got %s, want test-token", cred.Data.Token)
	}
	if cred.Metadata.Description != "Test credential" {
		t.Errorf("Description: got %s, want Test credential", cred.Metadata.Description)
	}
	if cred.Metadata.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
}
