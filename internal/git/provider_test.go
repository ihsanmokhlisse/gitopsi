package git

import (
	"testing"
)

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantErr    bool
		provider   ProviderType
		instance   string
		owner      string
		repository string
		isSSH      bool
	}{
		{
			name:       "GitHub SSH URL",
			url:        "git@github.com:owner/repo.git",
			wantErr:    false,
			provider:   ProviderGitHub,
			instance:   "github.com",
			owner:      "owner",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "GitHub SSH URL without .git",
			url:        "git@github.com:owner/repo",
			wantErr:    false,
			provider:   ProviderGitHub,
			instance:   "github.com",
			owner:      "owner",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "GitHub HTTPS URL",
			url:        "https://github.com/owner/repo.git",
			wantErr:    false,
			provider:   ProviderGitHub,
			instance:   "github.com",
			owner:      "owner",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:       "GitHub HTTPS URL without .git",
			url:        "https://github.com/owner/repo",
			wantErr:    false,
			provider:   ProviderGitHub,
			instance:   "github.com",
			owner:      "owner",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:       "GitLab SSH URL",
			url:        "git@gitlab.com:group/project.git",
			wantErr:    false,
			provider:   ProviderGitLab,
			instance:   "gitlab.com",
			owner:      "group",
			repository: "project",
			isSSH:      true,
		},
		{
			name:       "GitLab HTTPS URL",
			url:        "https://gitlab.com/group/project.git",
			wantErr:    false,
			provider:   ProviderGitLab,
			instance:   "gitlab.com",
			owner:      "group",
			repository: "project",
			isSSH:      false,
		},
		{
			name:       "Bitbucket SSH URL",
			url:        "git@bitbucket.org:workspace/repo.git",
			wantErr:    false,
			provider:   ProviderBitbucket,
			instance:   "bitbucket.org",
			owner:      "workspace",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "Bitbucket HTTPS URL",
			url:        "https://bitbucket.org/workspace/repo.git",
			wantErr:    false,
			provider:   ProviderBitbucket,
			instance:   "bitbucket.org",
			owner:      "workspace",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:       "Azure DevOps SSH URL",
			url:        "git@ssh.dev.azure.com:v3/org/project/repo",
			wantErr:    false,
			provider:   ProviderAzureDevOps,
			instance:   "dev.azure.com",
			owner:      "org/project",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "Azure DevOps HTTPS URL",
			url:        "https://dev.azure.com/org/project/_git/repo",
			wantErr:    false,
			provider:   ProviderAzureDevOps,
			instance:   "dev.azure.com",
			owner:      "org/project",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:       "Self-hosted GitLab SSH",
			url:        "git@gitlab.company.com:team/repo.git",
			wantErr:    false,
			provider:   ProviderGitLab,
			instance:   "gitlab.company.com",
			owner:      "team",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "Self-hosted GitLab HTTPS",
			url:        "https://gitlab.company.com/team/repo.git",
			wantErr:    false,
			provider:   ProviderGitLab,
			instance:   "gitlab.company.com",
			owner:      "team",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:       "Gitea SSH URL",
			url:        "git@gitea.example.com:org/repo.git",
			wantErr:    false,
			provider:   ProviderGitea,
			instance:   "gitea.example.com",
			owner:      "org",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "Gitea HTTPS URL",
			url:        "https://gitea.example.com/org/repo.git",
			wantErr:    false,
			provider:   ProviderGitea,
			instance:   "gitea.example.com",
			owner:      "org",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:       "Generic Git SSH URL",
			url:        "git@git.example.com:team/repo.git",
			wantErr:    false,
			provider:   ProviderGeneric,
			instance:   "git.example.com",
			owner:      "team",
			repository: "repo",
			isSSH:      true,
		},
		{
			name:       "Generic Git HTTPS URL",
			url:        "https://git.example.com/team/repo.git",
			wantErr:    false,
			provider:   ProviderGeneric,
			instance:   "git.example.com",
			owner:      "team",
			repository: "repo",
			isSSH:      false,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "Whitespace only URL",
			url:     "   ",
			wantErr: true,
		},
		{
			name:    "Invalid URL",
			url:     "not-a-valid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGitURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseGitURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseGitURL() unexpected error: %v", err)
				return
			}

			if result.Provider != tt.provider {
				t.Errorf("Provider = %v, want %v", result.Provider, tt.provider)
			}
			if result.Instance != tt.instance {
				t.Errorf("Instance = %v, want %v", result.Instance, tt.instance)
			}
			if result.Owner != tt.owner {
				t.Errorf("Owner = %v, want %v", result.Owner, tt.owner)
			}
			if result.Repository != tt.repository {
				t.Errorf("Repository = %v, want %v", result.Repository, tt.repository)
			}
			if result.IsSSH != tt.isSSH {
				t.Errorf("IsSSH = %v, want %v", result.IsSSH, tt.isSSH)
			}
		})
	}
}

func TestParsedURL_GetHTTPSURL(t *testing.T) {
	tests := []struct {
		name   string
		parsed *ParsedURL
		want   string
	}{
		{
			name: "GitHub",
			parsed: &ParsedURL{
				Provider:   ProviderGitHub,
				Instance:   "github.com",
				Owner:      "owner",
				Repository: "repo",
			},
			want: "https://github.com/owner/repo.git",
		},
		{
			name: "GitLab",
			parsed: &ParsedURL{
				Provider:   ProviderGitLab,
				Instance:   "gitlab.com",
				Owner:      "group",
				Repository: "project",
			},
			want: "https://gitlab.com/group/project.git",
		},
		{
			name: "GitLab self-hosted",
			parsed: &ParsedURL{
				Provider:   ProviderGitLab,
				Instance:   "gitlab.company.com",
				Owner:      "team",
				Repository: "repo",
			},
			want: "https://gitlab.company.com/team/repo.git",
		},
		{
			name: "Bitbucket",
			parsed: &ParsedURL{
				Provider:   ProviderBitbucket,
				Instance:   "bitbucket.org",
				Owner:      "workspace",
				Repository: "repo",
			},
			want: "https://bitbucket.org/workspace/repo.git",
		},
		{
			name: "Azure DevOps",
			parsed: &ParsedURL{
				Provider:   ProviderAzureDevOps,
				Instance:   "dev.azure.com",
				Owner:      "org/project",
				Repository: "repo",
			},
			want: "https://dev.azure.com/org/project/_git/repo",
		},
		{
			name: "Generic",
			parsed: &ParsedURL{
				Provider:   ProviderGeneric,
				Instance:   "git.example.com",
				Owner:      "team",
				Repository: "repo",
			},
			want: "https://git.example.com/team/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.parsed.GetHTTPSURL()
			if got != tt.want {
				t.Errorf("GetHTTPSURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsedURL_GetSSHURL(t *testing.T) {
	tests := []struct {
		name   string
		parsed *ParsedURL
		want   string
	}{
		{
			name: "GitHub",
			parsed: &ParsedURL{
				Provider:   ProviderGitHub,
				Instance:   "github.com",
				Owner:      "owner",
				Repository: "repo",
			},
			want: "git@github.com:owner/repo.git",
		},
		{
			name: "GitLab",
			parsed: &ParsedURL{
				Provider:   ProviderGitLab,
				Instance:   "gitlab.com",
				Owner:      "group",
				Repository: "project",
			},
			want: "git@gitlab.com:group/project.git",
		},
		{
			name: "GitLab self-hosted",
			parsed: &ParsedURL{
				Provider:   ProviderGitLab,
				Instance:   "gitlab.company.com",
				Owner:      "team",
				Repository: "repo",
			},
			want: "git@gitlab.company.com:team/repo.git",
		},
		{
			name: "Bitbucket",
			parsed: &ParsedURL{
				Provider:   ProviderBitbucket,
				Instance:   "bitbucket.org",
				Owner:      "workspace",
				Repository: "repo",
			},
			want: "git@bitbucket.org:workspace/repo.git",
		},
		{
			name: "Azure DevOps",
			parsed: &ParsedURL{
				Provider:   ProviderAzureDevOps,
				Instance:   "dev.azure.com",
				Owner:      "org/project",
				Repository: "repo",
			},
			want: "git@ssh.dev.azure.com:v3/org/project/repo",
		},
		{
			name: "Generic",
			parsed: &ParsedURL{
				Provider:   ProviderGeneric,
				Instance:   "git.example.com",
				Owner:      "team",
				Repository: "repo",
			},
			want: "git@git.example.com:team/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.parsed.GetSSHURL()
			if got != tt.want {
				t.Errorf("GetSSHURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsedURL_GetCloneURL(t *testing.T) {
	parsed := &ParsedURL{
		Provider:   ProviderGitHub,
		Instance:   "github.com",
		Owner:      "owner",
		Repository: "repo",
	}

	sshURL := parsed.GetCloneURL(true)
	if sshURL != "git@github.com:owner/repo.git" {
		t.Errorf("GetCloneURL(true) = %v, want git@github.com:owner/repo.git", sshURL)
	}

	httpsURL := parsed.GetCloneURL(false)
	if httpsURL != "https://github.com/owner/repo.git" {
		t.Errorf("GetCloneURL(false) = %v, want https://github.com/owner/repo.git", httpsURL)
	}
}

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid GitHub SSH", "git@github.com:owner/repo.git", false},
		{"Valid GitHub HTTPS", "https://github.com/owner/repo.git", false},
		{"Valid GitLab SSH", "git@gitlab.com:group/project.git", false},
		{"Valid GitLab HTTPS", "https://gitlab.com/group/project.git", false},
		{"Invalid empty", "", true},
		{"Invalid format", "not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectProviderFromHost(t *testing.T) {
	tests := []struct {
		host string
		want ProviderType
	}{
		{"gitlab.company.com", ProviderGitLab},
		{"my-gitlab.example.com", ProviderGitLab},
		{"gitea.example.com", ProviderGitea},
		{"my-gitea.local", ProviderGitea},
		{"bitbucket.mycompany.com", ProviderBitbucket},
		{"azure.visualstudio.com", ProviderAzureDevOps},
		{"github.enterprise.com", ProviderGitHub},
		{"git.example.com", ProviderGeneric},
		{"random.host.com", ProviderGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := detectProviderFromHost(tt.host)
			if got != tt.want {
				t.Errorf("detectProviderFromHost(%s) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestProviderTypeConstants(t *testing.T) {
	if ProviderGitHub != "github" {
		t.Errorf("ProviderGitHub = %v, want github", ProviderGitHub)
	}
	if ProviderGitLab != "gitlab" {
		t.Errorf("ProviderGitLab = %v, want gitlab", ProviderGitLab)
	}
	if ProviderGitea != "gitea" {
		t.Errorf("ProviderGitea = %v, want gitea", ProviderGitea)
	}
	if ProviderAzureDevOps != "azure-devops" {
		t.Errorf("ProviderAzureDevOps = %v, want azure-devops", ProviderAzureDevOps)
	}
	if ProviderBitbucket != "bitbucket" {
		t.Errorf("ProviderBitbucket = %v, want bitbucket", ProviderBitbucket)
	}
	if ProviderGeneric != "generic" {
		t.Errorf("ProviderGeneric = %v, want generic", ProviderGeneric)
	}
}

func TestAuthMethodConstants(t *testing.T) {
	if AuthSSH != "ssh" {
		t.Errorf("AuthSSH = %v, want ssh", AuthSSH)
	}
	if AuthToken != "token" {
		t.Errorf("AuthToken = %v, want token", AuthToken)
	}
	if AuthOAuth != "oauth" {
		t.Errorf("AuthOAuth = %v, want oauth", AuthOAuth)
	}
}

func TestVisibilityConstants(t *testing.T) {
	if VisibilityPublic != "public" {
		t.Errorf("VisibilityPublic = %v, want public", VisibilityPublic)
	}
	if VisibilityPrivate != "private" {
		t.Errorf("VisibilityPrivate = %v, want private", VisibilityPrivate)
	}
	if VisibilityInternal != "internal" {
		t.Errorf("VisibilityInternal = %v, want internal", VisibilityInternal)
	}
}

func TestCapabilityConstants(t *testing.T) {
	capabilities := []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityBranchProtection,
		CapabilityCICD,
	}

	expected := []string{
		"create-repo",
		"delete-repo",
		"webhooks",
		"deploy-keys",
		"branch-protection",
		"cicd",
	}

	for i, cap := range capabilities {
		if string(cap) != expected[i] {
			t.Errorf("Capability %d = %v, want %v", i, cap, expected[i])
		}
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "Already has .git",
			url:  "https://github.com/owner/repo.git",
			want: "https://github.com/owner/repo.git",
		},
		{
			name: "Missing .git",
			url:  "https://github.com/owner/repo",
			want: "https://github.com/owner/repo.git",
		},
		{
			name: "Trailing slash",
			url:  "https://github.com/owner/repo/",
			want: "https://github.com/owner/repo.git",
		},
		{
			name: "SSH URL (not modified)",
			url:  "git@github.com:owner/repo.git",
			want: "git@github.com:owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeGitURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeGitURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeGitURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
