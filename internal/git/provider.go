package git

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type ProviderType string

const (
	ProviderGitHub      ProviderType = "github"
	ProviderGitLab      ProviderType = "gitlab"
	ProviderGitea       ProviderType = "gitea"
	ProviderAzureDevOps ProviderType = "azure-devops"
	ProviderBitbucket   ProviderType = "bitbucket"
	ProviderGeneric     ProviderType = "generic"
)

type AuthMethod string

const (
	AuthSSH   AuthMethod = "ssh"
	AuthToken AuthMethod = "token"
	AuthOAuth AuthMethod = "oauth"
)

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityInternal Visibility = "internal"
)

type Repository struct {
	Name          string
	FullName      string
	Description   string
	URL           string
	SSHURL        string
	HTTPSURL      string
	CloneURL      string
	Visibility    Visibility
	DefaultBranch string
	Owner         string
}

type AuthOptions struct {
	Method   AuthMethod
	Token    string
	SSHKey   string
	Username string
	Password string
}

type CloneOptions struct {
	URL    string
	Path   string
	Branch string
	Depth  int
	Auth   *AuthOptions
}

type PushOptions struct {
	Path   string
	Remote string
	Branch string
	Force  bool
	Auth   *AuthOptions
}

type PullOptions struct {
	Path   string
	Remote string
	Branch string
	Auth   *AuthOptions
}

type CreateRepoOptions struct {
	Name        string
	Description string
	Visibility  Visibility
	AutoInit    bool
}

type WebhookOptions struct {
	URL    string
	Secret string
	Events []string
	Active bool
}

type Webhook struct {
	ID     string
	URL    string
	Events []string
	Active bool
}

type Capability string

const (
	CapabilityCreateRepo       Capability = "create-repo"
	CapabilityDeleteRepo       Capability = "delete-repo"
	CapabilityWebhooks         Capability = "webhooks"
	CapabilityDeployKeys       Capability = "deploy-keys"
	CapabilityBranchProtection Capability = "branch-protection"
	CapabilityCICD             Capability = "cicd"
)

type Provider interface {
	Name() ProviderType
	Capabilities() []Capability

	Authenticate(ctx context.Context, opts AuthOptions) error
	ValidateAccess(ctx context.Context) error

	Clone(ctx context.Context, opts CloneOptions) error
	Push(ctx context.Context, opts PushOptions) error
	Pull(ctx context.Context, opts PullOptions) error

	CreateRepository(ctx context.Context, opts CreateRepoOptions) (*Repository, error)
	GetRepository(ctx context.Context, owner, name string) (*Repository, error)
	DeleteRepository(ctx context.Context, owner, name string) error

	CreateWebhook(ctx context.Context, owner, repo string, opts WebhookOptions) (*Webhook, error)
	DeleteWebhook(ctx context.Context, owner, repo, webhookID string) error

	TestConnection(ctx context.Context) error
}

type ParsedURL struct {
	Provider    ProviderType
	Instance    string
	Owner       string
	Repository  string
	IsSSH       bool
	OriginalURL string
}

var (
	sshGitHubPattern      = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsGitHubPattern    = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	sshGitLabPattern      = regexp.MustCompile(`^git@gitlab\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsGitLabPattern    = regexp.MustCompile(`^https?://gitlab\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	sshBitbucketPattern   = regexp.MustCompile(`^git@bitbucket\.org:([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsBitbucketPattern = regexp.MustCompile(`^https?://bitbucket\.org/([^/]+)/([^/]+?)(?:\.git)?$`)
	sshAzurePattern       = regexp.MustCompile(`^git@ssh\.dev\.azure\.com:v3/([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsAzurePattern     = regexp.MustCompile(`^https?://dev\.azure\.com/([^/]+)/([^/]+)/_git/([^/]+?)(?:\.git)?$`)
	sshGenericPattern     = regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsGenericPattern   = regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`)
)

func ParseGitURL(gitURL string) (*ParsedURL, error) {
	gitURL = strings.TrimSpace(gitURL)
	if gitURL == "" {
		return nil, fmt.Errorf("git URL cannot be empty")
	}

	if matches := sshGitHubPattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderGitHub,
			Instance:    "github.com",
			Owner:       matches[1],
			Repository:  strings.TrimSuffix(matches[2], ".git"),
			IsSSH:       true,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := httpsGitHubPattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderGitHub,
			Instance:    "github.com",
			Owner:       matches[1],
			Repository:  strings.TrimSuffix(matches[2], ".git"),
			IsSSH:       false,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := sshGitLabPattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderGitLab,
			Instance:    "gitlab.com",
			Owner:       matches[1],
			Repository:  strings.TrimSuffix(matches[2], ".git"),
			IsSSH:       true,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := httpsGitLabPattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderGitLab,
			Instance:    "gitlab.com",
			Owner:       matches[1],
			Repository:  strings.TrimSuffix(matches[2], ".git"),
			IsSSH:       false,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := sshBitbucketPattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderBitbucket,
			Instance:    "bitbucket.org",
			Owner:       matches[1],
			Repository:  strings.TrimSuffix(matches[2], ".git"),
			IsSSH:       true,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := httpsBitbucketPattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderBitbucket,
			Instance:    "bitbucket.org",
			Owner:       matches[1],
			Repository:  strings.TrimSuffix(matches[2], ".git"),
			IsSSH:       false,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := sshAzurePattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderAzureDevOps,
			Instance:    "dev.azure.com",
			Owner:       matches[1] + "/" + matches[2],
			Repository:  strings.TrimSuffix(matches[3], ".git"),
			IsSSH:       true,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := httpsAzurePattern.FindStringSubmatch(gitURL); matches != nil {
		return &ParsedURL{
			Provider:    ProviderAzureDevOps,
			Instance:    "dev.azure.com",
			Owner:       matches[1] + "/" + matches[2],
			Repository:  strings.TrimSuffix(matches[3], ".git"),
			IsSSH:       false,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := sshGenericPattern.FindStringSubmatch(gitURL); matches != nil {
		instance := matches[1]
		provider := detectProviderFromHost(instance)
		return &ParsedURL{
			Provider:    provider,
			Instance:    instance,
			Owner:       matches[2],
			Repository:  strings.TrimSuffix(matches[3], ".git"),
			IsSSH:       true,
			OriginalURL: gitURL,
		}, nil
	}

	if matches := httpsGenericPattern.FindStringSubmatch(gitURL); matches != nil {
		instance := matches[1]
		provider := detectProviderFromHost(instance)
		return &ParsedURL{
			Provider:    provider,
			Instance:    instance,
			Owner:       matches[2],
			Repository:  strings.TrimSuffix(matches[3], ".git"),
			IsSSH:       false,
			OriginalURL: gitURL,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse git URL: %s", gitURL)
}

func detectProviderFromHost(host string) ProviderType {
	host = strings.ToLower(host)

	if strings.Contains(host, "gitlab") {
		return ProviderGitLab
	}
	if strings.Contains(host, "gitea") {
		return ProviderGitea
	}
	if strings.Contains(host, "bitbucket") {
		return ProviderBitbucket
	}
	if strings.Contains(host, "azure") || strings.Contains(host, "visualstudio") {
		return ProviderAzureDevOps
	}
	if strings.Contains(host, "github") {
		return ProviderGitHub
	}

	return ProviderGeneric
}

func (p *ParsedURL) GetHTTPSURL() string {
	switch p.Provider {
	case ProviderGitHub:
		return fmt.Sprintf("https://github.com/%s/%s.git", p.Owner, p.Repository)
	case ProviderGitLab:
		return fmt.Sprintf("https://%s/%s/%s.git", p.Instance, p.Owner, p.Repository)
	case ProviderBitbucket:
		return fmt.Sprintf("https://bitbucket.org/%s/%s.git", p.Owner, p.Repository)
	case ProviderAzureDevOps:
		parts := strings.Split(p.Owner, "/")
		if len(parts) == 2 {
			return fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", parts[0], parts[1], p.Repository)
		}
		return fmt.Sprintf("https://%s/%s/%s.git", p.Instance, p.Owner, p.Repository)
	default:
		return fmt.Sprintf("https://%s/%s/%s.git", p.Instance, p.Owner, p.Repository)
	}
}

func (p *ParsedURL) GetSSHURL() string {
	switch p.Provider {
	case ProviderGitHub:
		return fmt.Sprintf("git@github.com:%s/%s.git", p.Owner, p.Repository)
	case ProviderGitLab:
		return fmt.Sprintf("git@%s:%s/%s.git", p.Instance, p.Owner, p.Repository)
	case ProviderBitbucket:
		return fmt.Sprintf("git@bitbucket.org:%s/%s.git", p.Owner, p.Repository)
	case ProviderAzureDevOps:
		parts := strings.Split(p.Owner, "/")
		if len(parts) == 2 {
			return fmt.Sprintf("git@ssh.dev.azure.com:v3/%s/%s/%s", parts[0], parts[1], p.Repository)
		}
		return fmt.Sprintf("git@%s:%s/%s.git", p.Instance, p.Owner, p.Repository)
	default:
		return fmt.Sprintf("git@%s:%s/%s.git", p.Instance, p.Owner, p.Repository)
	}
}

func (p *ParsedURL) GetCloneURL(preferSSH bool) string {
	if preferSSH {
		return p.GetSSHURL()
	}
	return p.GetHTTPSURL()
}

func ValidateGitURL(gitURL string) error {
	_, err := ParseGitURL(gitURL)
	return err
}

func NormalizeGitURL(gitURL string) (string, error) {
	parsed, err := url.Parse(gitURL)
	if err != nil {
		return gitURL, nil
	}

	path := strings.TrimSuffix(parsed.Path, "/")
	if !strings.HasSuffix(path, ".git") {
		path += ".git"
	}
	parsed.Path = path

	return parsed.String(), nil
}
