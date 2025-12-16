package git

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type BitbucketProvider struct {
	username    string
	appPassword string
	workspace   string
	auth        *AuthOptions
}

func NewBitbucketProvider(workspace string) *BitbucketProvider {
	return &BitbucketProvider{
		workspace: workspace,
	}
}

func NewBitbucketProviderWithToken(username, appPassword, workspace string) *BitbucketProvider {
	return &BitbucketProvider{
		username:    username,
		appPassword: appPassword,
		workspace:   workspace,
	}
}

func (b *BitbucketProvider) Name() ProviderType {
	return ProviderBitbucket
}

func (b *BitbucketProvider) Capabilities() []Capability {
	return []Capability{
		CapabilityCreateRepo,
		CapabilityDeleteRepo,
		CapabilityWebhooks,
		CapabilityDeployKeys,
		CapabilityCICD,
	}
}

func (b *BitbucketProvider) Authenticate(ctx context.Context, opts AuthOptions) error {
	b.auth = &opts

	switch opts.Method {
	case AuthToken:
		if opts.Username == "" {
			b.username = os.Getenv("BITBUCKET_USERNAME")
		} else {
			b.username = opts.Username
		}

		if opts.Token == "" {
			b.appPassword = os.Getenv("BITBUCKET_APP_PASSWORD")
			if b.appPassword == "" {
				b.appPassword = os.Getenv("BITBUCKET_TOKEN")
			}
			if b.appPassword == "" {
				b.appPassword = os.Getenv("GIT_TOKEN")
			}
			if b.appPassword == "" {
				return fmt.Errorf("Bitbucket app password not provided and BITBUCKET_APP_PASSWORD not set")
			}
		} else {
			b.appPassword = opts.Token
		}

		if b.username == "" {
			return fmt.Errorf("Bitbucket username is required")
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
		b.auth.SSHKey = opts.SSHKey
	default:
		return fmt.Errorf("unsupported authentication method: %s", opts.Method)
	}

	return nil
}

func (b *BitbucketProvider) ValidateAccess(ctx context.Context) error {
	if b.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	switch b.auth.Method {
	case AuthToken:
		if b.appPassword == "" || b.username == "" {
			return fmt.Errorf("no credentials available")
		}
		apiURL := "https://api.bitbucket.org/2.0/user"
		basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.appPassword)))
		cmd := exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("Authorization: Basic %s", basicAuth), apiURL)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to validate Bitbucket credentials: %w", err)
		}
		if strings.Contains(string(output), "401") || strings.Contains(string(output), "Unauthorized") {
			return fmt.Errorf("invalid Bitbucket credentials")
		}
	case AuthSSH:
		cmd := exec.CommandContext(ctx, "ssh", "-T", "-o", "StrictHostKeyChecking=no", "git@bitbucket.org")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		if strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("SSH authentication failed: %s", outputStr)
		}
	}

	return nil
}

func (b *BitbucketProvider) Clone(ctx context.Context, opts CloneOptions) error {
	args := []string{"clone"}

	if opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}

	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	args = append(args, opts.URL, opts.Path)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = b.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}

	return nil
}

func (b *BitbucketProvider) Push(ctx context.Context, opts PushOptions) error {
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
	cmd.Env = b.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w: %s", err, string(output))
	}

	return nil
}

func (b *BitbucketProvider) Pull(ctx context.Context, opts PullOptions) error {
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
	cmd.Env = b.getGitEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w: %s", err, string(output))
	}

	return nil
}

func (b *BitbucketProvider) CreateRepository(ctx context.Context, opts CreateRepoOptions) (*Repository, error) {
	if b.appPassword == "" || b.username == "" {
		return nil, fmt.Errorf("credentials required to create repository")
	}

	if b.workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}

	isPrivate := "true"
	if opts.Visibility == VisibilityPublic {
		isPrivate = "false"
	}

	apiURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", b.workspace, opts.Name)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.appPassword)))

	payload := fmt.Sprintf(`{"scm":"git","is_private":%s,"description":"%s"}`, isPrivate, opts.Description)

	args := []string{
		"-s",
		"-H", fmt.Sprintf("Authorization: Basic %s", basicAuth),
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
		FullName:    fmt.Sprintf("%s/%s", b.workspace, opts.Name),
		Description: opts.Description,
		URL:         fmt.Sprintf("https://bitbucket.org/%s/%s", b.workspace, opts.Name),
		HTTPSURL:    fmt.Sprintf("https://bitbucket.org/%s/%s.git", b.workspace, opts.Name),
		SSHURL:      fmt.Sprintf("git@bitbucket.org:%s/%s.git", b.workspace, opts.Name),
		Visibility:  opts.Visibility,
		Owner:       b.workspace,
	}, nil
}

func (b *BitbucketProvider) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	if b.appPassword == "" || b.username == "" {
		return nil, fmt.Errorf("credentials required to get repository")
	}

	workspace := owner
	if workspace == "" {
		workspace = b.workspace
	}

	apiURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", workspace, name)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.appPassword)))

	cmd := exec.CommandContext(ctx, "curl", "-s", "-H", fmt.Sprintf("Authorization: Basic %s", basicAuth), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), "404") || strings.Contains(string(output), "not found") {
		return nil, fmt.Errorf("repository not found: %s/%s", workspace, name)
	}

	return &Repository{
		Name:     name,
		FullName: fmt.Sprintf("%s/%s", workspace, name),
		URL:      fmt.Sprintf("https://bitbucket.org/%s/%s", workspace, name),
		HTTPSURL: fmt.Sprintf("https://bitbucket.org/%s/%s.git", workspace, name),
		SSHURL:   fmt.Sprintf("git@bitbucket.org:%s/%s.git", workspace, name),
		Owner:    workspace,
	}, nil
}

func (b *BitbucketProvider) DeleteRepository(ctx context.Context, owner, name string) error {
	if b.appPassword == "" || b.username == "" {
		return fmt.Errorf("credentials required to delete repository")
	}

	workspace := owner
	if workspace == "" {
		workspace = b.workspace
	}

	apiURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", workspace, name)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.appPassword)))

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-H", fmt.Sprintf("Authorization: Basic %s", basicAuth), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w: %s", err, string(output))
	}

	return nil
}

func (b *BitbucketProvider) CreateWebhook(ctx context.Context, owner, repo string, opts WebhookOptions) (*Webhook, error) {
	if b.appPassword == "" || b.username == "" {
		return nil, fmt.Errorf("credentials required to create webhook")
	}

	workspace := owner
	if workspace == "" {
		workspace = b.workspace
	}

	apiURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/hooks", workspace, repo)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.appPassword)))

	events := make([]string, 0, len(opts.Events))
	for _, e := range opts.Events {
		switch e {
		case "push":
			events = append(events, `"repo:push"`)
		case "pull_request":
			events = append(events, `"pullrequest:created"`, `"pullrequest:updated"`)
		default:
			events = append(events, fmt.Sprintf(`"%s"`, e))
		}
	}

	payload := fmt.Sprintf(`{
		"description": "gitopsi webhook",
		"url": "%s",
		"active": %t,
		"events": [%s]
	}`, opts.URL, opts.Active, strings.Join(events, ","))

	args := []string{
		"-s",
		"-H", fmt.Sprintf("Authorization: Basic %s", basicAuth),
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

func (b *BitbucketProvider) DeleteWebhook(ctx context.Context, owner, repo, webhookID string) error {
	if b.appPassword == "" || b.username == "" {
		return fmt.Errorf("credentials required to delete webhook")
	}

	workspace := owner
	if workspace == "" {
		workspace = b.workspace
	}

	apiURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/hooks/%s", workspace, repo, webhookID)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.appPassword)))

	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "DELETE", "-H", fmt.Sprintf("Authorization: Basic %s", basicAuth), apiURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w: %s", err, string(output))
	}

	return nil
}

func (b *BitbucketProvider) TestConnection(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "https://api.bitbucket.org/2.0")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot connect to Bitbucket: %w", err)
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode != "200" && statusCode != "302" {
		return fmt.Errorf("Bitbucket returned status %s", statusCode)
	}

	return nil
}

func (b *BitbucketProvider) getGitEnv() []string {
	env := os.Environ()

	if b.appPassword != "" && b.username != "" {
		env = append(env,
			fmt.Sprintf("BITBUCKET_USERNAME=%s", b.username),
			fmt.Sprintf("BITBUCKET_APP_PASSWORD=%s", b.appPassword),
		)
	}

	if b.auth != nil && b.auth.Method == AuthSSH && b.auth.SSHKey != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", b.auth.SSHKey))
	}

	return env
}

func (b *BitbucketProvider) GetUsername() string {
	return b.username
}

func (b *BitbucketProvider) SetUsername(username string) {
	b.username = username
}

func (b *BitbucketProvider) GetAppPassword() string {
	return b.appPassword
}

func (b *BitbucketProvider) SetAppPassword(password string) {
	b.appPassword = password
}

func (b *BitbucketProvider) GetWorkspace() string {
	return b.workspace
}

func (b *BitbucketProvider) SetWorkspace(workspace string) {
	b.workspace = workspace
}

