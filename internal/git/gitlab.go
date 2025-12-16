package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GitLabProvider struct {
	token    string
	instance string
	auth     *AuthOptions
}

func NewGitLabProvider() *GitLabProvider {
	return &GitLabProvider{
		instance: "gitlab.com",
	}
}

func NewGitLabProviderWithInstance(instance string) *GitLabProvider {
	if instance == "" {
		instance = "gitlab.com"
	}
	return &GitLabProvider{
		instance: instance,
	}
}

func NewGitLabProviderWithToken(token, instance string) *GitLabProvider {
	if instance == "" {
		instance = "gitlab.com"
	}
	return &GitLabProvider{
		token:    token,
		instance: instance,
	}
}

func (g *GitLabProvider) Name() ProviderType {
	return ProviderGitLab
}

func (g *GitLabProvider) Capabilities() []Capability {
	return []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityBranchProtection,
		CapabilityCICD,
	}
}

func (g *GitLabProvider) Authenticate(ctx context.Context, opts AuthOptions) error {
	g.auth = &opts

	switch opts.Method {
	case AuthToken:
		if opts.Token == "" {
			token := os.Getenv("GITLAB_TOKEN")
			if token == "" {
				token = os.Getenv("GL_TOKEN")
			}
			if token == "" {
				token = os.Getenv("GIT_TOKEN")
			}
			if token == "" {
				return fmt.Errorf("GitLab token not provided and GITLAB_TOKEN/GL_TOKEN not set")
			}
			g.token = token
		} else {
			g.token = opts.Token
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
		g.auth.SSHKey = opts.SSHKey
	case AuthOAuth:
		return fmt.Errorf("OAuth authentication not yet implemented for GitLab")
	default:
		return fmt.Errorf("unsupported authentication method: %s", opts.Method)
	}

	return nil
}

func (g *GitLabProvider) ValidateAccess(ctx context.Context) error {
	if g.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	switch g.auth.Method {
	case AuthToken:
		if g.token == "" {
			return fmt.Errorf("no token available")
		}
		apiURL := fmt.Sprintf("https://%s/api/v4/user", g.instance)
		cmd := exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("PRIVATE-TOKEN: %s", g.token), apiURL)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to validate GitLab token: %w", err)
		}
		if strings.Contains(string(output), "401") || strings.Contains(string(output), "Unauthorized") {
			return fmt.Errorf("invalid GitLab token")
		}
	case AuthSSH:
		cmd := exec.CommandContext(ctx, "ssh", "-T", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("git@%s", g.instance))
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		if !strings.Contains(outputStr, "Welcome to GitLab") && !strings.Contains(outputStr, "successfully authenticated") {
			return fmt.Errorf("SSH authentication failed: %s", outputStr)
		}
	}

	return nil
}

func (g *GitLabProvider) Clone(ctx context.Context, opts CloneOptions) error {
	args := []string{"clone"}

	if opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}

	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	args = append(args, opts.URL, opts.Path)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = g.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}

	return nil
}

func (g *GitLabProvider) Push(ctx context.Context, opts PushOptions) error {
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
	cmd.Env = g.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w: %s", err, string(output))
	}

	return nil
}

func (g *GitLabProvider) Pull(ctx context.Context, opts PullOptions) error {
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
	cmd.Env = g.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w: %s", err, string(output))
	}

	return nil
}

func (g *GitLabProvider) CreateRepository(ctx context.Context, opts CreateRepoOptions) (*Repository, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to create repository")
	}

	visibility := "private"
	switch opts.Visibility {
	case VisibilityPublic:
		visibility = "public"
	case VisibilityInternal:
		visibility = "internal"
	case VisibilityPrivate:
		visibility = "private"
	}

	apiURL := fmt.Sprintf("https://%s/api/v4/projects", g.instance)

	args := []string{
		"-s",
		"-H", fmt.Sprintf("PRIVATE-TOKEN: %s", g.token),
		"-X", "POST",
		"-d", fmt.Sprintf("name=%s", opts.Name),
		"-d", fmt.Sprintf("visibility=%s", visibility),
	}

	if opts.Description != "" {
		args = append(args, "-d", fmt.Sprintf("description=%s", opts.Description))
	}

	args = append(args, apiURL)

	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), "error") || strings.Contains(string(output), "message") {
		if strings.Contains(string(output), "has already been taken") {
			return nil, fmt.Errorf("repository already exists: %s", opts.Name)
		}
	}

	return &Repository{
		Name:        opts.Name,
		Description: opts.Description,
		URL:         fmt.Sprintf("https://%s/%s", g.instance, opts.Name),
		HTTPSURL:    fmt.Sprintf("https://%s/%s.git", g.instance, opts.Name),
		SSHURL:      fmt.Sprintf("git@%s:%s.git", g.instance, opts.Name),
		Visibility:  opts.Visibility,
	}, nil
}

func (g *GitLabProvider) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to get repository")
	}

	projectPath := fmt.Sprintf("%s/%s", owner, name)
	projectPath = strings.ReplaceAll(projectPath, "/", "%2F")

	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s", g.instance, projectPath)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("PRIVATE-TOKEN: %s", g.token), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), "404") || strings.Contains(string(output), "not found") {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, name)
	}

	return &Repository{
		Name:     name,
		FullName: fmt.Sprintf("%s/%s", owner, name),
		URL:      fmt.Sprintf("https://%s/%s/%s", g.instance, owner, name),
		HTTPSURL: fmt.Sprintf("https://%s/%s/%s.git", g.instance, owner, name),
		SSHURL:   fmt.Sprintf("git@%s:%s/%s.git", g.instance, owner, name),
		Owner:    owner,
	}, nil
}

func (g *GitLabProvider) DeleteRepository(ctx context.Context, owner, name string) error {
	if g.token == "" {
		return fmt.Errorf("token required to delete repository")
	}

	projectPath := fmt.Sprintf("%s/%s", owner, name)
	projectPath = strings.ReplaceAll(projectPath, "/", "%2F")

	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s", g.instance, projectPath)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-H", fmt.Sprintf("PRIVATE-TOKEN: %s", g.token), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w: %s", err, string(output))
	}

	return nil
}

func (g *GitLabProvider) CreateWebhook(ctx context.Context, owner, repo string, opts WebhookOptions) (*Webhook, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to create webhook")
	}

	projectPath := fmt.Sprintf("%s/%s", owner, repo)
	projectPath = strings.ReplaceAll(projectPath, "/", "%2F")

	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/hooks", g.instance, projectPath)

	args := []string{
		"-s",
		"-H", fmt.Sprintf("PRIVATE-TOKEN: %s", g.token),
		"-X", "POST",
		"-d", fmt.Sprintf("url=%s", opts.URL),
		"-d", fmt.Sprintf("token=%s", opts.Secret),
		"-d", "push_events=true",
	}

	for _, event := range opts.Events {
		switch event {
		case "push":
			args = append(args, "-d", "push_events=true")
		case "tag_push":
			args = append(args, "-d", "tag_push_events=true")
		case "merge_request":
			args = append(args, "-d", "merge_requests_events=true")
		}
	}

	args = append(args, apiURL)

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

func (g *GitLabProvider) DeleteWebhook(ctx context.Context, owner, repo, webhookID string) error {
	if g.token == "" {
		return fmt.Errorf("token required to delete webhook")
	}

	projectPath := fmt.Sprintf("%s/%s", owner, repo)
	projectPath = strings.ReplaceAll(projectPath, "/", "%2F")

	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/hooks/%s", g.instance, projectPath, webhookID)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-H", fmt.Sprintf("PRIVATE-TOKEN: %s", g.token), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w: %s", err, string(output))
	}

	return nil
}

func (g *GitLabProvider) TestConnection(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", fmt.Sprintf("https://%s", g.instance))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot connect to GitLab: %w", err)
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode != "200" && statusCode != "302" {
		return fmt.Errorf("GitLab returned status %s", statusCode)
	}

	return nil
}

func (g *GitLabProvider) getGitEnv() []string {
	env := os.Environ()

	if g.token != "" {
		env = append(env,
			fmt.Sprintf("GITLAB_TOKEN=%s", g.token),
			fmt.Sprintf("GL_TOKEN=%s", g.token),
		)
	}

	if g.auth != nil && g.auth.Method == AuthSSH && g.auth.SSHKey != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", g.auth.SSHKey))
	}

	return env
}

func (g *GitLabProvider) GetToken() string {
	return g.token
}

func (g *GitLabProvider) SetToken(token string) {
	g.token = token
}

func (g *GitLabProvider) GetInstance() string {
	return g.instance
}

func (g *GitLabProvider) SetInstance(instance string) {
	g.instance = instance
}
