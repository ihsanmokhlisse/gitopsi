package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type AzureDevOpsProvider struct {
	token        string
	organization string
	project      string
	auth         *AuthOptions
}

func NewAzureDevOpsProvider(organization, project string) *AzureDevOpsProvider {
	return &AzureDevOpsProvider{
		organization: organization,
		project:      project,
	}
}

func NewAzureDevOpsProviderWithToken(token, organization, project string) *AzureDevOpsProvider {
	return &AzureDevOpsProvider{
		token:        token,
		organization: organization,
		project:      project,
	}
}

func (a *AzureDevOpsProvider) Name() ProviderType {
	return ProviderAzureDevOps
}

func (a *AzureDevOpsProvider) Capabilities() []Capability {
	return []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityCICD,
	}
}

func (a *AzureDevOpsProvider) Authenticate(ctx context.Context, opts AuthOptions) error {
	a.auth = &opts

	switch opts.Method {
	case AuthToken:
		if opts.Token == "" {
			token := os.Getenv("AZURE_DEVOPS_PAT")
			if token == "" {
				token = os.Getenv("AZURE_DEVOPS_TOKEN")
			}
			if token == "" {
				token = os.Getenv("GIT_TOKEN")
			}
			if token == "" {
				return fmt.Errorf("Azure DevOps PAT not provided and AZURE_DEVOPS_PAT not set")
			}
			a.token = token
		} else {
			a.token = opts.Token
		}
	case AuthSSH:
		if opts.SSHKey == "" {
			opts.SSHKey = os.ExpandEnv("$HOME/.ssh/id_rsa")
		}
		if _, err := os.Stat(opts.SSHKey); os.IsNotExist(err) {
			ed25519Key := os.ExpandEnv("$HOME/.ssh/id_ed25519")
			if _, err := os.Stat(ed25519Key); err == nil {
				opts.SSHKey = ed25519Key
			} else {
				return fmt.Errorf("SSH key not found: %s", opts.SSHKey)
			}
		}
		a.auth.SSHKey = opts.SSHKey
	default:
		return fmt.Errorf("unsupported authentication method: %s", opts.Method)
	}

	return nil
}

func (a *AzureDevOpsProvider) ValidateAccess(ctx context.Context) error {
	if a.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	switch a.auth.Method {
	case AuthToken:
		if a.token == "" {
			return fmt.Errorf("no token available")
		}
		apiURL := fmt.Sprintf("https://dev.azure.com/%s/_apis/projects?api-version=7.0", a.organization)
		basicAuth := fmt.Sprintf(":%s", a.token)
		cmd := exec.CommandContext(ctx, "curl", "-s", "-u", basicAuth, apiURL)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to validate Azure DevOps PAT: %w", err)
		}
		if strings.Contains(string(output), "401") || strings.Contains(string(output), "Unauthorized") {
			return fmt.Errorf("invalid Azure DevOps PAT")
		}
	case AuthSSH:
		cmd := exec.CommandContext(ctx, "ssh", "-T", "-o", "StrictHostKeyChecking=no", "git@ssh.dev.azure.com")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		if strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("SSH authentication failed: %s", outputStr)
		}
	}

	return nil
}

func (a *AzureDevOpsProvider) Clone(ctx context.Context, opts CloneOptions) error {
	args := []string{"clone"}

	if opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}

	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	args = append(args, opts.URL, opts.Path)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = a.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}

	return nil
}

func (a *AzureDevOpsProvider) Push(ctx context.Context, opts PushOptions) error {
	args := []string{"push"}

	if opts.Remote == "" {
		opts.Remote = "origin"
	}

	args = append(args, opts.Remote)

	if opts.Branch != "" {
		args = append(args, opts.Branch)
	}

	if opts.Force {
		args = append(args, "--force")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = opts.Path
	cmd.Env = a.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w: %s", err, string(output))
	}

	return nil
}

func (a *AzureDevOpsProvider) Pull(ctx context.Context, opts PullOptions) error {
	args := []string{"pull"}

	if opts.Remote == "" {
		opts.Remote = "origin"
	}

	args = append(args, opts.Remote)

	if opts.Branch != "" {
		args = append(args, opts.Branch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = opts.Path
	cmd.Env = a.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w: %s", err, string(output))
	}

	return nil
}

func (a *AzureDevOpsProvider) CreateRepository(ctx context.Context, opts CreateRepoOptions) (*Repository, error) {
	if a.token == "" {
		return nil, fmt.Errorf("token required to create repository")
	}

	if a.organization == "" || a.project == "" {
		return nil, fmt.Errorf("organization and project are required")
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories?api-version=7.0", a.organization, a.project)
	basicAuth := fmt.Sprintf(":%s", a.token)

	payload := fmt.Sprintf(`{"name":%q}`, opts.Name)

	args := []string{
		"-s",
		"-u", basicAuth,
		"-H", "Content-Type: application/json",
		"-X", "POST",
		"-d", payload,
		apiURL,
	}

	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), "error") {
		return nil, fmt.Errorf("failed to create repository: %s", string(output))
	}

	return &Repository{
		Name:        opts.Name,
		Description: opts.Description,
		URL:         fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", a.organization, a.project, opts.Name),
		HTTPSURL:    fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", a.organization, a.project, opts.Name),
		SSHURL:      fmt.Sprintf("git@ssh.dev.azure.com:v3/%s/%s/%s", a.organization, a.project, opts.Name),
		Visibility:  opts.Visibility,
		Owner:       fmt.Sprintf("%s/%s", a.organization, a.project),
	}, nil
}

func (a *AzureDevOpsProvider) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	if a.token == "" {
		return nil, fmt.Errorf("token required to get repository")
	}

	parts := strings.Split(owner, "/")
	org := a.organization
	proj := a.project
	if len(parts) == 2 {
		org = parts[0]
		proj = parts[1]
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories/%s?api-version=7.0", org, proj, name)
	basicAuth := fmt.Sprintf(":%s", a.token)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-u", basicAuth, apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), "404") || strings.Contains(string(output), "not found") {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, name)
	}

	return &Repository{
		Name:     name,
		FullName: fmt.Sprintf("%s/%s/%s", org, proj, name),
		URL:      fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", org, proj, name),
		HTTPSURL: fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", org, proj, name),
		SSHURL:   fmt.Sprintf("git@ssh.dev.azure.com:v3/%s/%s/%s", org, proj, name),
		Owner:    owner,
	}, nil
}

func (a *AzureDevOpsProvider) DeleteRepository(ctx context.Context, owner, name string) error {
	if a.token == "" {
		return fmt.Errorf("token required to delete repository")
	}

	parts := strings.Split(owner, "/")
	org := a.organization
	proj := a.project
	if len(parts) == 2 {
		org = parts[0]
		proj = parts[1]
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories/%s?api-version=7.0", org, proj, name)
	basicAuth := fmt.Sprintf(":%s", a.token)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-u", basicAuth, apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w: %s", err, string(output))
	}

	return nil
}

func (a *AzureDevOpsProvider) CreateWebhook(ctx context.Context, owner, repo string, opts WebhookOptions) (*Webhook, error) {
	if a.token == "" {
		return nil, fmt.Errorf("token required to create webhook")
	}

	parts := strings.Split(owner, "/")
	org := a.organization
	proj := a.project
	if len(parts) == 2 {
		org = parts[0]
		proj = parts[1]
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/_apis/hooks/subscriptions?api-version=7.0", org)
	basicAuth := fmt.Sprintf(":%s", a.token)

	payload := fmt.Sprintf(`{
		"publisherId": "tfs",
		"eventType": "git.push",
		"resourceVersion": "1.0",
		"consumerId": "webHooks",
		"consumerActionId": "httpRequest",
		"publisherInputs": {
			"projectId": "%s",
			"repository": "%s"
		},
		"consumerInputs": {
			"url": "%s"
		}
	}`, proj, repo, opts.URL)

	args := []string{
		"-s",
		"-u", basicAuth,
		"-H", "Content-Type: application/json",
		"-X", "POST",
		"-d", payload,
		apiURL,
	}

	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w: %s", err, string(output))
	}

	return &Webhook{
		URL:    opts.URL,
		Events: opts.Events,
		Active: opts.Active,
	}, nil
}

func (a *AzureDevOpsProvider) DeleteWebhook(ctx context.Context, owner, repo, webhookID string) error {
	if a.token == "" {
		return fmt.Errorf("token required to delete webhook")
	}

	parts := strings.Split(owner, "/")
	org := a.organization
	if len(parts) >= 1 {
		org = parts[0]
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/_apis/hooks/subscriptions/%s?api-version=7.0", org, webhookID)
	basicAuth := fmt.Sprintf(":%s", a.token)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-u", basicAuth, apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w: %s", err, string(output))
	}

	return nil
}

func (a *AzureDevOpsProvider) TestConnection(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "https://dev.azure.com")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot connect to Azure DevOps: %w", err)
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode != "200" && statusCode != "302" && statusCode != "203" {
		return fmt.Errorf("Azure DevOps returned status %s", statusCode)
	}

	return nil
}

func (a *AzureDevOpsProvider) getGitEnv() []string {
	env := os.Environ()

	if a.token != "" {
		env = append(env,
			fmt.Sprintf("AZURE_DEVOPS_PAT=%s", a.token),
			fmt.Sprintf("GIT_ASKPASS=echo"),
			fmt.Sprintf("GIT_USERNAME="),
			fmt.Sprintf("GIT_PASSWORD=%s", a.token),
		)
	}

	if a.auth != nil && a.auth.Method == AuthSSH && a.auth.SSHKey != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", a.auth.SSHKey))
	}

	return env
}

func (a *AzureDevOpsProvider) GetToken() string {
	return a.token
}

func (a *AzureDevOpsProvider) SetToken(token string) {
	a.token = token
}

func (a *AzureDevOpsProvider) GetOrganization() string {
	return a.organization
}

func (a *AzureDevOpsProvider) SetOrganization(org string) {
	a.organization = org
}

func (a *AzureDevOpsProvider) GetProject() string {
	return a.project
}

func (a *AzureDevOpsProvider) SetProject(project string) {
	a.project = project
}
