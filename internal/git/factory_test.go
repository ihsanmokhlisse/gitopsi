package git

import (
	"testing"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		instance     string
		wantErr      bool
		wantType     ProviderType
	}{
		{
			name:         "GitHub provider",
			providerType: ProviderGitHub,
			instance:     "",
			wantErr:      false,
			wantType:     ProviderGitHub,
		},
		{
			name:         "GitLab provider",
			providerType: ProviderGitLab,
			instance:     "gitlab.com",
			wantErr:      false,
			wantType:     ProviderGitLab,
		},
		{
			name:         "GitLab self-hosted",
			providerType: ProviderGitLab,
			instance:     "gitlab.company.com",
			wantErr:      false,
			wantType:     ProviderGitLab,
		},
		{
			name:         "Gitea not implemented",
			providerType: ProviderGitea,
			instance:     "",
			wantErr:      true,
		},
		{
			name:         "Azure DevOps not implemented",
			providerType: ProviderAzureDevOps,
			instance:     "",
			wantErr:      true,
		},
		{
			name:         "Bitbucket not implemented",
			providerType: ProviderBitbucket,
			instance:     "",
			wantErr:      true,
		},
		{
			name:         "Generic not implemented",
			providerType: ProviderGeneric,
			instance:     "",
			wantErr:      true,
		},
		{
			name:         "Unknown provider",
			providerType: "unknown",
			instance:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.providerType, tt.instance)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProvider() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewProvider() unexpected error: %v", err)
				return
			}

			if provider.Name() != tt.wantType {
				t.Errorf("Provider.Name() = %v, want %v", provider.Name(), tt.wantType)
			}
		})
	}
}

func TestNewProviderFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantErr  bool
		wantType ProviderType
	}{
		{
			name:     "GitHub SSH URL",
			url:      "git@github.com:owner/repo.git",
			wantErr:  false,
			wantType: ProviderGitHub,
		},
		{
			name:     "GitHub HTTPS URL",
			url:      "https://github.com/owner/repo.git",
			wantErr:  false,
			wantType: ProviderGitHub,
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:group/project.git",
			wantErr:  false,
			wantType: ProviderGitLab,
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/group/project.git",
			wantErr:  false,
			wantType: ProviderGitLab,
		},
		{
			name:     "GitLab self-hosted",
			url:      "https://gitlab.company.com/team/repo.git",
			wantErr:  false,
			wantType: ProviderGitLab,
		},
		{
			name:    "Bitbucket (not implemented)",
			url:     "git@bitbucket.org:workspace/repo.git",
			wantErr: true,
		},
		{
			name:    "Azure DevOps (not implemented)",
			url:     "https://dev.azure.com/org/project/_git/repo",
			wantErr: true,
		},
		{
			name:    "Invalid URL",
			url:     "invalid-url",
			wantErr: true,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProviderFromURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProviderFromURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewProviderFromURL() unexpected error: %v", err)
				return
			}

			if provider.Name() != tt.wantType {
				t.Errorf("Provider.Name() = %v, want %v", provider.Name(), tt.wantType)
			}
		})
	}
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantErr      bool
		wantProvider ProviderType
		wantInstance string
	}{
		{
			name:         "GitHub",
			url:          "git@github.com:owner/repo.git",
			wantErr:      false,
			wantProvider: ProviderGitHub,
			wantInstance: "github.com",
		},
		{
			name:         "GitLab",
			url:          "https://gitlab.com/group/project.git",
			wantErr:      false,
			wantProvider: ProviderGitLab,
			wantInstance: "gitlab.com",
		},
		{
			name:         "GitLab self-hosted",
			url:          "https://gitlab.company.com/team/repo.git",
			wantErr:      false,
			wantProvider: ProviderGitLab,
			wantInstance: "gitlab.company.com",
		},
		{
			name:         "Bitbucket",
			url:          "git@bitbucket.org:workspace/repo.git",
			wantErr:      false,
			wantProvider: ProviderBitbucket,
			wantInstance: "bitbucket.org",
		},
		{
			name:         "Azure DevOps",
			url:          "https://dev.azure.com/org/project/_git/repo",
			wantErr:      false,
			wantProvider: ProviderAzureDevOps,
			wantInstance: "dev.azure.com",
		},
		{
			name:    "Invalid URL",
			url:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, instance, err := DetectProvider(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectProvider() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DetectProvider() unexpected error: %v", err)
				return
			}

			if provider != tt.wantProvider {
				t.Errorf("Provider = %v, want %v", provider, tt.wantProvider)
			}
			if instance != tt.wantInstance {
				t.Errorf("Instance = %v, want %v", instance, tt.wantInstance)
			}
		})
	}
}

func TestGetSupportedProviders(t *testing.T) {
	providers := GetSupportedProviders()

	if len(providers) != 6 {
		t.Errorf("GetSupportedProviders() returned %d providers, want 6", len(providers))
	}

	expected := []ProviderType{
		ProviderGitHub,
		ProviderGitLab,
		ProviderGitea,
		ProviderAzureDevOps,
		ProviderBitbucket,
		ProviderGeneric,
	}

	for i, p := range expected {
		if providers[i] != p {
			t.Errorf("GetSupportedProviders()[%d] = %v, want %v", i, providers[i], p)
		}
	}
}

func TestGetImplementedProviders(t *testing.T) {
	providers := GetImplementedProviders()

	if len(providers) != 2 {
		t.Errorf("GetImplementedProviders() returned %d providers, want 2", len(providers))
	}

	hasGitHub := false
	hasGitLab := false

	for _, p := range providers {
		if p == ProviderGitHub {
			hasGitHub = true
		}
		if p == ProviderGitLab {
			hasGitLab = true
		}
	}

	if !hasGitHub {
		t.Error("GetImplementedProviders() should include GitHub")
	}
	if !hasGitLab {
		t.Error("GetImplementedProviders() should include GitLab")
	}
}

func TestIsProviderImplemented(t *testing.T) {
	tests := []struct {
		provider ProviderType
		want     bool
	}{
		{ProviderGitHub, true},
		{ProviderGitLab, true},
		{ProviderGitea, false},
		{ProviderAzureDevOps, false},
		{ProviderBitbucket, false},
		{ProviderGeneric, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			got := IsProviderImplemented(tt.provider)
			if got != tt.want {
				t.Errorf("IsProviderImplemented(%s) = %v, want %v", tt.provider, got, tt.want)
			}
		})
	}
}

func TestGetProviderDisplayName(t *testing.T) {
	tests := []struct {
		provider ProviderType
		want     string
	}{
		{ProviderGitHub, "GitHub"},
		{ProviderGitLab, "GitLab"},
		{ProviderGitea, "Gitea"},
		{ProviderAzureDevOps, "Azure DevOps"},
		{ProviderBitbucket, "Bitbucket"},
		{ProviderGeneric, "Generic Git"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			got := GetProviderDisplayName(tt.provider)
			if got != tt.want {
				t.Errorf("GetProviderDisplayName(%s) = %v, want %v", tt.provider, got, tt.want)
			}
		})
	}
}

func TestGetAuthMethodsForProvider(t *testing.T) {
	tests := []struct {
		provider    ProviderType
		wantMethods []AuthMethod
	}{
		{ProviderGitHub, []AuthMethod{AuthSSH, AuthToken, AuthOAuth}},
		{ProviderGitLab, []AuthMethod{AuthSSH, AuthToken, AuthOAuth}},
		{ProviderGitea, []AuthMethod{AuthSSH, AuthToken}},
		{ProviderAzureDevOps, []AuthMethod{AuthSSH, AuthToken, AuthOAuth}},
		{ProviderBitbucket, []AuthMethod{AuthSSH, AuthToken, AuthOAuth}},
		{ProviderGeneric, []AuthMethod{AuthSSH, AuthToken}},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			got := GetAuthMethodsForProvider(tt.provider)

			if len(got) != len(tt.wantMethods) {
				t.Errorf("GetAuthMethodsForProvider(%s) returned %d methods, want %d", tt.provider, len(got), len(tt.wantMethods))
				return
			}

			for i, method := range tt.wantMethods {
				if got[i] != method {
					t.Errorf("GetAuthMethodsForProvider(%s)[%d] = %v, want %v", tt.provider, i, got[i], method)
				}
			}
		})
	}
}

func TestGetAuthMethodDisplayName(t *testing.T) {
	tests := []struct {
		method AuthMethod
		want   string
	}{
		{AuthSSH, "SSH Key"},
		{AuthToken, "Personal Access Token"},
		{AuthOAuth, "OAuth (Browser)"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			got := GetAuthMethodDisplayName(tt.method)
			if got != tt.want {
				t.Errorf("GetAuthMethodDisplayName(%s) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}
