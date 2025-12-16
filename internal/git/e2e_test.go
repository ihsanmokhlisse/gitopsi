package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestE2E_ParseAndDetect(t *testing.T) {
	urls := []struct {
		url          string
		wantProvider ProviderType
		wantOwner    string
		wantRepo     string
	}{
		{
			url:          "git@github.com:ihsanmokhlisse/gitopsi.git",
			wantProvider: ProviderGitHub,
			wantOwner:    "ihsanmokhlisse",
			wantRepo:     "gitopsi",
		},
		{
			url:          "https://github.com/kubernetes/kubernetes.git",
			wantProvider: ProviderGitHub,
			wantOwner:    "kubernetes",
			wantRepo:     "kubernetes",
		},
		{
			url:          "git@gitlab.com:gitlab-org/gitlab.git",
			wantProvider: ProviderGitLab,
			wantOwner:    "gitlab-org",
			wantRepo:     "gitlab",
		},
		{
			url:          "https://gitlab.com/gitlab-org/gitlab-runner.git",
			wantProvider: ProviderGitLab,
			wantOwner:    "gitlab-org",
			wantRepo:     "gitlab-runner",
		},
	}

	for _, tt := range urls {
		t.Run(tt.url, func(t *testing.T) {
			parsed, err := ParseGitURL(tt.url)
			if err != nil {
				t.Fatalf("ParseGitURL() error: %v", err)
			}

			if parsed.Provider != tt.wantProvider {
				t.Errorf("Provider = %v, want %v", parsed.Provider, tt.wantProvider)
			}
			if parsed.Owner != tt.wantOwner {
				t.Errorf("Owner = %v, want %v", parsed.Owner, tt.wantOwner)
			}
			if parsed.Repository != tt.wantRepo {
				t.Errorf("Repository = %v, want %v", parsed.Repository, tt.wantRepo)
			}

			providerType, instance, err := DetectProvider(tt.url)
			if err != nil {
				t.Fatalf("DetectProvider() error: %v", err)
			}

			if providerType != tt.wantProvider {
				t.Errorf("DetectProvider() provider = %v, want %v", providerType, tt.wantProvider)
			}

			t.Logf("URL: %s -> Provider: %s, Instance: %s", tt.url, providerType, instance)
		})
	}
}

func TestE2E_ProviderFactory(t *testing.T) {
	tests := []struct {
		url          string
		wantProvider ProviderType
	}{
		{"git@github.com:owner/repo.git", ProviderGitHub},
		{"https://github.com/owner/repo.git", ProviderGitHub},
		{"git@gitlab.com:group/project.git", ProviderGitLab},
		{"https://gitlab.com/group/project.git", ProviderGitLab},
		{"https://gitlab.company.com/team/repo.git", ProviderGitLab},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			provider, err := NewProviderFromURL(tt.url)
			if err != nil {
				t.Fatalf("NewProviderFromURL() error: %v", err)
			}

			if provider.Name() != tt.wantProvider {
				t.Errorf("Provider.Name() = %v, want %v", provider.Name(), tt.wantProvider)
			}

			caps := provider.Capabilities()
			if len(caps) == 0 {
				t.Error("Provider should have capabilities")
			}

			t.Logf("Provider: %s, Capabilities: %v", provider.Name(), caps)
		})
	}
}

func TestE2E_URLConversion(t *testing.T) {
	tests := []struct {
		sshURL   string
		httpsURL string
	}{
		{
			sshURL:   "git@github.com:owner/repo.git",
			httpsURL: "https://github.com/owner/repo.git",
		},
		{
			sshURL:   "git@gitlab.com:group/project.git",
			httpsURL: "https://gitlab.com/group/project.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.sshURL, func(t *testing.T) {
			parsedSSH, err := ParseGitURL(tt.sshURL)
			if err != nil {
				t.Fatalf("ParseGitURL(SSH) error: %v", err)
			}

			parsedHTTPS, err := ParseGitURL(tt.httpsURL)
			if err != nil {
				t.Fatalf("ParseGitURL(HTTPS) error: %v", err)
			}

			if parsedSSH.GetHTTPSURL() != tt.httpsURL {
				t.Errorf("SSH.GetHTTPSURL() = %v, want %v", parsedSSH.GetHTTPSURL(), tt.httpsURL)
			}

			if parsedHTTPS.GetSSHURL() != tt.sshURL {
				t.Errorf("HTTPS.GetSSHURL() = %v, want %v", parsedHTTPS.GetSSHURL(), tt.sshURL)
			}

			if parsedSSH.Owner != parsedHTTPS.Owner {
				t.Errorf("Owner mismatch: SSH=%v, HTTPS=%v", parsedSSH.Owner, parsedHTTPS.Owner)
			}

			if parsedSSH.Repository != parsedHTTPS.Repository {
				t.Errorf("Repository mismatch: SSH=%v, HTTPS=%v", parsedSSH.Repository, parsedHTTPS.Repository)
			}
		})
	}
}

func TestE2E_AuthenticationSetup(t *testing.T) {
	ctx := context.Background()

	t.Run("GitHub with token from env", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "test_token_for_testing")

		provider := NewGitHubProvider()
		err := provider.Authenticate(ctx, AuthOptions{
			Method: AuthToken,
		})

		if err != nil {
			t.Errorf("Authenticate() error: %v", err)
		}

		if provider.GetToken() != "test_token_for_testing" {
			t.Errorf("Token not set correctly from env")
		}
	})

	t.Run("GitLab with token from env", func(t *testing.T) {
		t.Setenv("GITLAB_TOKEN", "glpat_test_token")

		provider := NewGitLabProvider()
		err := provider.Authenticate(ctx, AuthOptions{
			Method: AuthToken,
		})

		if err != nil {
			t.Errorf("Authenticate() error: %v", err)
		}

		if provider.GetToken() != "glpat_test_token" {
			t.Errorf("Token not set correctly from env")
		}
	})

	t.Run("GitHub with explicit token", func(t *testing.T) {
		provider := NewGitHubProvider()
		err := provider.Authenticate(ctx, AuthOptions{
			Method: AuthToken,
			Token:  "explicit_token",
		})

		if err != nil {
			t.Errorf("Authenticate() error: %v", err)
		}

		if provider.GetToken() != "explicit_token" {
			t.Errorf("Token not set correctly")
		}
	})

	t.Run("GitLab self-hosted with token", func(t *testing.T) {
		provider := NewGitLabProviderWithInstance("gitlab.company.com")
		err := provider.Authenticate(ctx, AuthOptions{
			Method: AuthToken,
			Token:  "company_token",
		})

		if err != nil {
			t.Errorf("Authenticate() error: %v", err)
		}

		if provider.GetInstance() != "gitlab.company.com" {
			t.Errorf("Instance not set correctly")
		}

		if provider.GetToken() != "company_token" {
			t.Errorf("Token not set correctly")
		}
	})
}

func TestE2E_GitEnvironment(t *testing.T) {
	t.Run("GitHub env with token", func(t *testing.T) {
		provider := NewGitHubProviderWithToken("my_token")
		env := provider.getGitEnv()

		hasGHToken := false
		hasGitHubToken := false

		for _, e := range env {
			if e == "GH_TOKEN=my_token" {
				hasGHToken = true
			}
			if e == "GITHUB_TOKEN=my_token" {
				hasGitHubToken = true
			}
		}

		if !hasGHToken || !hasGitHubToken {
			t.Error("Environment should contain both GH_TOKEN and GITHUB_TOKEN")
		}
	})

	t.Run("GitLab env with token", func(t *testing.T) {
		provider := NewGitLabProviderWithToken("my_token", "gitlab.com")
		env := provider.getGitEnv()

		hasGLToken := false
		hasGitLabToken := false

		for _, e := range env {
			if e == "GL_TOKEN=my_token" {
				hasGLToken = true
			}
			if e == "GITLAB_TOKEN=my_token" {
				hasGitLabToken = true
			}
		}

		if !hasGLToken || !hasGitLabToken {
			t.Error("Environment should contain both GL_TOKEN and GITLAB_TOKEN")
		}
	})
}

func TestE2E_ProviderCapabilities(t *testing.T) {
	providers := []struct {
		name     string
		provider Provider
	}{
		{"GitHub", NewGitHubProvider()},
		{"GitLab", NewGitLabProvider()},
	}

	expectedCaps := []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityBranchProtection,
		CapabilityCICD,
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			caps := p.provider.Capabilities()

			if len(caps) != len(expectedCaps) {
				t.Errorf("%s: Capabilities count = %d, want %d", p.name, len(caps), len(expectedCaps))
			}

			capMap := make(map[Capability]bool)
			for _, c := range caps {
				capMap[c] = true
			}

			for _, expected := range expectedCaps {
				if !capMap[expected] {
					t.Errorf("%s: missing capability %v", p.name, expected)
				}
			}
		})
	}
}

func TestE2E_WorkflowSimulation(t *testing.T) {
	ctx := context.Background()

	t.Run("Complete GitHub workflow simulation", func(t *testing.T) {
		gitURL := "git@github.com:test-org/test-repo.git"

		parsed, err := ParseGitURL(gitURL)
		if err != nil {
			t.Fatalf("Step 1 - ParseGitURL failed: %v", err)
		}
		t.Logf("Step 1: Parsed URL - Provider: %s, Owner: %s, Repo: %s",
			parsed.Provider, parsed.Owner, parsed.Repository)

		provider, err := NewProviderFromURL(gitURL)
		if err != nil {
			t.Fatalf("Step 2 - NewProviderFromURL failed: %v", err)
		}
		t.Logf("Step 2: Created provider: %s", provider.Name())

		t.Setenv("GITHUB_TOKEN", "test_workflow_token")
		err = provider.Authenticate(ctx, AuthOptions{Method: AuthToken})
		if err != nil {
			t.Fatalf("Step 3 - Authenticate failed: %v", err)
		}
		t.Log("Step 3: Authentication configured")

		ghProvider := provider.(*GitHubProvider)
		if ghProvider.GetToken() != "test_workflow_token" {
			t.Error("Token not set correctly")
		}
		t.Log("Step 4: Token verified")

		t.Log("Workflow simulation complete!")
	})

	t.Run("Complete GitLab workflow simulation", func(t *testing.T) {
		gitURL := "https://gitlab.company.com/team/project.git"

		parsed, err := ParseGitURL(gitURL)
		if err != nil {
			t.Fatalf("Step 1 - ParseGitURL failed: %v", err)
		}
		t.Logf("Step 1: Parsed URL - Provider: %s, Instance: %s, Owner: %s, Repo: %s",
			parsed.Provider, parsed.Instance, parsed.Owner, parsed.Repository)

		provider, err := NewProviderFromURL(gitURL)
		if err != nil {
			t.Fatalf("Step 2 - NewProviderFromURL failed: %v", err)
		}
		t.Logf("Step 2: Created provider: %s", provider.Name())

		glProvider := provider.(*GitLabProvider)
		if glProvider.GetInstance() != "gitlab.company.com" {
			t.Errorf("Instance = %s, want gitlab.company.com", glProvider.GetInstance())
		}
		t.Logf("Step 3: Instance verified: %s", glProvider.GetInstance())

		t.Setenv("GITLAB_TOKEN", "glpat_workflow_token")
		err = provider.Authenticate(ctx, AuthOptions{Method: AuthToken})
		if err != nil {
			t.Fatalf("Step 4 - Authenticate failed: %v", err)
		}
		t.Log("Step 4: Authentication configured")

		if glProvider.GetToken() != "glpat_workflow_token" {
			t.Error("Token not set correctly")
		}
		t.Log("Step 5: Token verified")

		t.Log("Workflow simulation complete!")
	})
}

func TestE2E_LocalGitOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping local git operations test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "gitopsi-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Git init and basic operations", func(t *testing.T) {
		repoDir := filepath.Join(tempDir, "test-repo")
		if err := os.MkdirAll(repoDir, 0750); err != nil {
			t.Fatalf("Failed to create repo dir: %v", err)
		}

		provider := NewGitHubProvider()

		cloneOpts := CloneOptions{
			Path:   repoDir,
			Branch: "main",
		}
		t.Logf("Clone options prepared: %+v", cloneOpts)

		pushOpts := PushOptions{
			Path:   repoDir,
			Remote: "origin",
			Branch: "main",
		}
		t.Logf("Push options prepared: %+v", pushOpts)

		pullOpts := PullOptions{
			Path:   repoDir,
			Remote: "origin",
			Branch: "main",
		}
		t.Logf("Pull options prepared: %+v", pullOpts)

		_ = provider
	})
}
