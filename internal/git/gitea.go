package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GiteaProvider struct {
	token    string
	instance string
	auth     *AuthOptions
}

func NewGiteaProvider(instance string) *GiteaProvider {
	if instance == "" {
		instance = "gitea.com"
	}
	return &GiteaProvider{
		instance: instance,
	}
}

func NewGiteaProviderWithToken(token, instance string) *GiteaProvider {
	if instance == "" {
		instance = "gitea.com"
	}
	return &GiteaProvider{
		token:    token,
		instance: instance,
	}
}

func (g *GiteaProvider) Name() ProviderType {
	return ProviderGitea
}

func (g *GiteaProvider) Capabilities() []Capability {
	return []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
	}
}

func (g *GiteaProvider) Authenticate(ctx context.Context, opts AuthOptions) error {
	g.auth = &opts

	switch opts.Method {
	case AuthToken:
		if opts.Token == "" {
			token := os.Getenv("GITEA_TOKEN")
			if token == "" {
				token = os.Getenv("GIT_TOKEN")
			}
			if token == "" {
				return fmt.Errorf("Gitea token not provided and GITEA_TOKEN not set")
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
	default:
		return fmt.Errorf("unsupported authentication method: %s", opts.Method)
	}

	return nil
}

func (g *GiteaProvider) ValidateAccess(ctx context.Context) error {
	if g.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	switch g.auth.Method {
	case AuthToken:
		if g.token == "" {
			return fmt.Errorf("no token available")
		}
		apiURL := fmt.Sprintf("https://%s/api/v1/user", g.instance)
		cmd := exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", g.token), apiURL)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to validate Gitea token: %w", err)
		}
		if strings.Contains(string(output), "401") || strings.Contains(string(output), "Unauthorized") {
			return fmt.Errorf("invalid Gitea token")
		}
	case AuthSSH:
		cmd := exec.CommandContext(ctx, "ssh", "-T", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("git@%s", g.instance))
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		if strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("SSH authentication failed: %s", outputStr)
		}
	}

	return nil
}

func (g *GiteaProvider) Clone(ctx context.Context, opts CloneOptions) error {
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

func (g *GiteaProvider) Push(ctx context.Context, opts PushOptions) error {
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

func (g *GiteaProvider) Pull(ctx context.Context, opts PullOptions) error {
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

func (g *GiteaProvider) CreateRepository(ctx context.Context, opts CreateRepoOptions) (*Repository, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to create repository")
	}

	visibility := "true"
	if opts.Visibility == VisibilityPublic {
		visibility = "false"
	}

	apiURL := fmt.Sprintf("https://%s/api/v1/user/repos", g.instance)

	args := []string{
		"-s",
		"-H", fmt.Sprintf("Authorization: token %s", g.token),
		"-H", "Content-Type: application/json",
		"-X", "POST",
		"-d", fmt.Sprintf(`{"name":%q,"description":%q,"private":%s}`, opts.Name, opts.Description, visibility),
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
		URL:         fmt.Sprintf("https://%s/%s", g.instance, opts.Name),
		HTTPSURL:    fmt.Sprintf("https://%s/%s.git", g.instance, opts.Name),
		SSHURL:      fmt.Sprintf("git@%s:%s.git", g.instance, opts.Name),
		Visibility:  opts.Visibility,
	}, nil
}

func (g *GiteaProvider) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to get repository")
	}

	apiURL := fmt.Sprintf("https://%s/api/v1/repos/%s/%s", g.instance, owner, name)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", g.token), apiURL)
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

func (g *GiteaProvider) DeleteRepository(ctx context.Context, owner, name string) error {
	if g.token == "" {
		return fmt.Errorf("token required to delete repository")
	}

	apiURL := fmt.Sprintf("https://%s/api/v1/repos/%s/%s", g.instance, owner, name)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-H", fmt.Sprintf("Authorization: token %s", g.token), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w: %s", err, string(output))
	}

	return nil
}

func (g *GiteaProvider) CreateWebhook(ctx context.Context, owner, repo string, opts WebhookOptions) (*Webhook, error) {
	if g.token == "" {
		return nil, fmt.Errorf("token required to create webhook")
	}

	apiURL := fmt.Sprintf("https://%s/api/v1/repos/%s/%s/hooks", g.instance, owner, repo)

	events := make([]string, 0, len(opts.Events))
	for _, e := range opts.Events {
		events = append(events, fmt.Sprintf("%q", e))
	}

	payload := fmt.Sprintf(`{
		"type": "gitea",
		"config": {
			"url": "%s",
			"content_type": "json",
			"secret": "%s"
		},
		"events": [%s],
		"active": %t
	}`, opts.URL, opts.Secret, strings.Join(events, ","), opts.Active)

	args := []string{
		"-s",
		"-H", fmt.Sprintf("Authorization: token %s", g.token),
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

func (g *GiteaProvider) DeleteWebhook(ctx context.Context, owner, repo, webhookID string) error {
	if g.token == "" {
		return fmt.Errorf("token required to delete webhook")
	}

	apiURL := fmt.Sprintf("https://%s/api/v1/repos/%s/%s/hooks/%s", g.instance, owner, repo, webhookID)

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-H", fmt.Sprintf("Authorization: token %s", g.token), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w: %s", err, string(output))
	}

	return nil
}

func (g *GiteaProvider) TestConnection(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", fmt.Sprintf("https://%s", g.instance))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot connect to Gitea: %w", err)
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode != "200" && statusCode != "302" {
		return fmt.Errorf("Gitea returned status %s", statusCode)
	}

	return nil
}

func (g *GiteaProvider) getGitEnv() []string {
	env := os.Environ()

	if g.token != "" {
		env = append(env, fmt.Sprintf("GITEA_TOKEN=%s", g.token))
	}

	if g.auth != nil && g.auth.Method == AuthSSH && g.auth.SSHKey != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", g.auth.SSHKey))
	}

	return env
}

func (g *GiteaProvider) GetToken() string {
	return g.token
}

func (g *GiteaProvider) SetToken(token string) {
	g.token = token
}

func (g *GiteaProvider) GetInstance() string {
	return g.instance
}

func (g *GiteaProvider) SetInstance(instance string) {
	g.instance = instance
}
