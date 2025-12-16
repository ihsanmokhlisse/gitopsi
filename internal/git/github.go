package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type HubProvider struct {
	token    string
	instance string
	auth     *AuthOptions
}

type GitHubProvider = HubProvider

func NewGitHubProvider() *HubProvider {
	return &HubProvider{
		instance: "github.com",
	}
}

func NewGitHubProviderWithToken(token string) *HubProvider {
	return &HubProvider{
		token:    token,
		instance: "github.com",
	}
}

func (g *HubProvider) Name() ProviderType {
	return ProviderGitHub
}

func (g *HubProvider) Capabilities() []Capability {
	return []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityBranchProtection,
		CapabilityCICD,
	}
}

func (g *HubProvider) Authenticate(ctx context.Context, opts AuthOptions) error {
	g.auth = &opts

	switch opts.Method {
	case AuthToken:
		if opts.Token == "" {
			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				token = os.Getenv("GH_TOKEN")
			}
			if token == "" {
				return fmt.Errorf("GitHub token not provided and GITHUB_TOKEN/GH_TOKEN not set")
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
		return fmt.Errorf("OAuth authentication not yet implemented")
	default:
		return fmt.Errorf("unsupported authentication method: %s", opts.Method)
	}

	return nil
}

func (g *HubProvider) ValidateAccess(ctx context.Context) error {
	if g.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	switch g.auth.Method {
	case AuthToken:
		if g.token == "" {
			return fmt.Errorf("no token available")
		}
		cmd := exec.CommandContext(ctx, "gh", "auth", "status")
		cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", g.token))
		if err := cmd.Run(); err != nil {
			cmd = exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", g.token), "https://api.github.com/user")
			output, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("failed to validate GitHub token: %w", err)
			}
			if strings.Contains(string(output), "Bad credentials") {
				return fmt.Errorf("invalid GitHub token")
			}
		}
	case AuthSSH:
		cmd := exec.CommandContext(ctx, "ssh", "-T", "-o", "StrictHostKeyChecking=no", "git@github.com")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		if !strings.Contains(outputStr, "successfully authenticated") && !strings.Contains(outputStr, "Hi ") {
			return fmt.Errorf("SSH authentication failed: %s", outputStr)
		}
	}

	return nil
}

func (g *HubProvider) Clone(ctx context.Context, opts CloneOptions) error {
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

func (g *HubProvider) Push(ctx context.Context, opts PushOptions) error {
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

func (g *HubProvider) Pull(ctx context.Context, opts PullOptions) error {
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

func (g *HubProvider) CreateRepository(ctx context.Context, opts CreateRepoOptions) (*Repository, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to create repository")
	}

	args := []string{"repo", "create", opts.Name}

	switch opts.Visibility {
	case VisibilityPublic:
		args = append(args, "--public")
	case VisibilityPrivate:
		args = append(args, "--private")
	case VisibilityInternal:
		args = append(args, "--internal")
	default:
		args = append(args, "--private")
	}

	if opts.Description != "" {
		args = append(args, "--description", opts.Description)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", g.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w: %s", err, string(output))
	}

	outputStr := strings.TrimSpace(string(output))
	parts := strings.Split(outputStr, "/")
	owner := ""
	repoName := opts.Name
	if len(parts) >= 2 {
		owner = parts[len(parts)-2]
		repoName = parts[len(parts)-1]
	}

	return &Repository{
		Name:        repoName,
		FullName:    fmt.Sprintf("%s/%s", owner, repoName),
		Description: opts.Description,
		URL:         fmt.Sprintf("https://github.com/%s/%s", owner, repoName),
		HTTPSURL:    fmt.Sprintf("https://github.com/%s/%s.git", owner, repoName),
		SSHURL:      fmt.Sprintf("git@github.com:%s/%s.git", owner, repoName),
		Visibility:  opts.Visibility,
		Owner:       owner,
	}, nil
}

func (g *HubProvider) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to get repository")
	}

	cmd := exec.CommandContext(ctx, "gh", "repo", "view", fmt.Sprintf("%s/%s", owner, name), "--json", "name,description,url,sshUrl,visibility,defaultBranchRef")
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", g.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w: %s", err, string(output))
	}

	return &Repository{
		Name:     name,
		FullName: fmt.Sprintf("%s/%s", owner, name),
		URL:      fmt.Sprintf("https://github.com/%s/%s", owner, name),
		HTTPSURL: fmt.Sprintf("https://github.com/%s/%s.git", owner, name),
		SSHURL:   fmt.Sprintf("git@github.com:%s/%s.git", owner, name),
		Owner:    owner,
	}, nil
}

func (g *HubProvider) DeleteRepository(ctx context.Context, owner, name string) error {
	if g.token == "" {
		return fmt.Errorf("token required to delete repository")
	}

	cmd := exec.CommandContext(ctx, "gh", "repo", "delete", fmt.Sprintf("%s/%s", owner, name), "--yes")
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", g.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w: %s", err, string(output))
	}

	return nil
}

func (g *HubProvider) CreateWebhook(ctx context.Context, owner, repo string, opts WebhookOptions) (*Webhook, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to create webhook")
	}

	events := strings.Join(opts.Events, ",")

	cmd := exec.CommandContext(ctx, "gh", "api",
		fmt.Sprintf("/repos/%s/%s/hooks", owner, repo),
		"-X", "POST",
		"-f", fmt.Sprintf("name=web"),
		"-f", fmt.Sprintf("config[url]=%s", opts.URL),
		"-f", fmt.Sprintf("config[content_type]=json"),
		"-f", fmt.Sprintf("config[secret]=%s", opts.Secret),
		"-f", fmt.Sprintf("events=%s", events),
		"-f", fmt.Sprintf("active=%t", opts.Active),
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", g.token))

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

func (g *HubProvider) DeleteWebhook(ctx context.Context, owner, repo, webhookID string) error {
	if g.token == "" {
		return fmt.Errorf("token required to delete webhook")
	}

	cmd := exec.CommandContext(ctx, "gh", "api",
		fmt.Sprintf("/repos/%s/%s/hooks/%s", owner, repo, webhookID),
		"-X", "DELETE",
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", g.token))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w: %s", err, string(output))
	}

	return nil
}

func (g *HubProvider) TestConnection(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "https://github.com")
	cmd.Env = g.getGitEnv()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot connect to GitHub: %w", err)
	}

	return nil
}

func (g *HubProvider) getGitEnv() []string {
	env := os.Environ()

	if g.token != "" {
		env = append(env,
			fmt.Sprintf("GH_TOKEN=%s", g.token),
			fmt.Sprintf("GITHUB_TOKEN=%s", g.token),
		)
	}

	if g.auth != nil && g.auth.Method == AuthSSH && g.auth.SSHKey != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", g.auth.SSHKey))
	}

	return env
}

func (g *HubProvider) GetToken() string {
	return g.token
}

func (g *HubProvider) SetToken(token string) {
	g.token = token
}

func (g *HubProvider) GetInstance() string {
	return g.instance
}
