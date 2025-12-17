// Package cluster provides Kubernetes cluster authentication and operations.
package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AuthMethod represents the authentication method for a cluster.
type AuthMethod string

const (
	AuthKubeconfig     AuthMethod = "kubeconfig"
	AuthToken          AuthMethod = "token"
	AuthOIDC           AuthMethod = "oidc"
	AuthServiceAccount AuthMethod = "service-account"
)

// Platform represents the Kubernetes platform type.
type Platform string

const (
	PlatformKubernetes Platform = "kubernetes"
	PlatformOpenShift  Platform = "openshift"
	PlatformAKS        Platform = "aks"
	PlatformEKS        Platform = "eks"
	PlatformGKE        Platform = "gke"
)

// AuthOptions holds cluster authentication configuration.
type AuthOptions struct {
	Method     AuthMethod
	Token      string
	TokenEnv   string
	Kubeconfig string
	Context    string
	CACert     string
	SkipTLS    bool
}

// Cluster represents a Kubernetes cluster connection.
type Cluster struct {
	URL      string
	Name     string
	Platform Platform
	auth     *AuthOptions
}

// New creates a new Cluster instance.
func New(url, name string, platform Platform) *Cluster {
	return &Cluster{
		URL:      url,
		Name:     name,
		Platform: platform,
	}
}

// Authenticate sets the authentication options for the cluster.
func (c *Cluster) Authenticate(opts *AuthOptions) error {
	c.auth = opts

	switch opts.Method {
	case AuthKubeconfig:
		if opts.Kubeconfig == "" {
			// Use default kubeconfig
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			opts.Kubeconfig = filepath.Join(home, ".kube", "config")
			c.auth.Kubeconfig = opts.Kubeconfig
		}
		if _, err := os.Stat(opts.Kubeconfig); os.IsNotExist(err) {
			return fmt.Errorf("kubeconfig file not found: %s", opts.Kubeconfig)
		}

	case AuthToken:
		token := opts.Token
		if token == "" && opts.TokenEnv != "" {
			token = os.Getenv(opts.TokenEnv)
		}
		if token == "" {
			return fmt.Errorf("token is required for token authentication")
		}
		c.auth.Token = token

	case AuthOIDC:
		// OIDC typically uses kubeconfig with OIDC provider configured
		if opts.Kubeconfig == "" {
			return fmt.Errorf("kubeconfig is required for OIDC authentication")
		}

	case AuthServiceAccount:
		// Service account uses in-cluster config or mounted token
		// #nosec G101 - This is a well-known Kubernetes path, not a hardcoded credential
		saTokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
		if _, err := os.Stat(saTokenPath); os.IsNotExist(err) {
			return fmt.Errorf("service account token not found (not running in cluster?)")
		}

	default:
		return fmt.Errorf("unsupported authentication method: %s", opts.Method)
	}

	return nil
}

// TestConnection tests the connection to the cluster.
func (c *Cluster) TestConnection(ctx context.Context) error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs("cluster-info")
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w: %s", err, string(output))
	}

	return nil
}

// GetServerVersion returns the Kubernetes server version.
func (c *Cluster) GetServerVersion(ctx context.Context) (string, error) {
	if c.auth == nil {
		return "", fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs("version", "--short")
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get server version: %w: %s", err, string(output))
	}

	// Parse server version from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Server Version:") {
			return strings.TrimPrefix(line, "Server Version:"), nil
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateNamespace creates a namespace in the cluster.
func (c *Cluster) CreateNamespace(ctx context.Context, name string) error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs("create", "namespace", name)
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if namespace already exists
		if strings.Contains(string(output), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create namespace: %w: %s", err, string(output))
	}

	return nil
}

// Apply applies a manifest to the cluster.
func (c *Cluster) Apply(ctx context.Context, manifest string) error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs("apply", "-f", "-")
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()
	cmd.Stdin = strings.NewReader(manifest)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply manifest: %w: %s", err, string(output))
	}

	return nil
}

// ApplyFile applies a manifest file to the cluster.
func (c *Cluster) ApplyFile(ctx context.Context, path string) error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs("apply", "-f", path)
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply file: %w: %s", err, string(output))
	}

	return nil
}

// WaitForDeployment waits for a deployment to be ready.
func (c *Cluster) WaitForDeployment(ctx context.Context, namespace, name string, timeout int) error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs(
		"wait", "--for=condition=available",
		fmt.Sprintf("deployment/%s", name),
		"-n", namespace,
		fmt.Sprintf("--timeout=%ds", timeout),
	)
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deployment not ready: %w: %s", err, string(output))
	}

	return nil
}

// GetPodLogs gets logs from a pod.
func (c *Cluster) GetPodLogs(ctx context.Context, namespace, pod string, lines int) (string, error) {
	if c.auth == nil {
		return "", fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs("logs", pod, "-n", namespace, fmt.Sprintf("--tail=%d", lines))
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w: %s", err, string(output))
	}

	return string(output), nil
}

// RunCommand runs a kubectl command with the cluster's authentication.
func (c *Cluster) RunCommand(ctx context.Context, kubectlArgs ...string) (string, error) {
	if c.auth == nil {
		return "", fmt.Errorf("not authenticated")
	}

	args := c.buildKubectlArgs(kubectlArgs...)
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = c.getKubeEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kubectl command failed: %w: %s", err, string(output))
	}

	return string(output), nil
}

// buildKubectlArgs builds kubectl arguments with authentication options.
func (c *Cluster) buildKubectlArgs(args ...string) []string {
	result := make([]string, 0, len(args)+6)

	// Add server URL if specified
	if c.URL != "" && c.auth != nil && c.auth.Method == AuthToken {
		result = append(result, "--server", c.URL)
	}

	// Add authentication args
	if c.auth != nil {
		switch c.auth.Method {
		case AuthKubeconfig:
			if c.auth.Kubeconfig != "" {
				result = append(result, "--kubeconfig", c.auth.Kubeconfig)
			}
			if c.auth.Context != "" {
				result = append(result, "--context", c.auth.Context)
			}

		case AuthToken:
			result = append(result, "--token", c.auth.Token)
			if c.auth.CACert != "" {
				result = append(result, "--certificate-authority", c.auth.CACert)
			}
			if c.auth.SkipTLS {
				result = append(result, "--insecure-skip-tls-verify")
			}
		}
	}

	// Add the actual command args
	result = append(result, args...)

	return result
}

// getKubeEnv returns environment variables for kubectl commands.
func (c *Cluster) getKubeEnv() []string {
	env := os.Environ()

	if c.auth != nil && c.auth.Kubeconfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", c.auth.Kubeconfig))
	}

	return env
}

// GetURL returns the cluster URL.
func (c *Cluster) GetURL() string {
	return c.URL
}

// GetName returns the cluster name.
func (c *Cluster) GetName() string {
	return c.Name
}

// GetPlatform returns the cluster platform.
func (c *Cluster) GetPlatform() Platform {
	return c.Platform
}

// IsAuthenticated returns true if the cluster has authentication configured.
func (c *Cluster) IsAuthenticated() bool {
	return c.auth != nil
}

// GetAuthMethod returns the configured authentication method.
func (c *Cluster) GetAuthMethod() AuthMethod {
	if c.auth == nil {
		return ""
	}
	return c.auth.Method
}
