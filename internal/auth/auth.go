// Package auth provides authentication management for Git providers,
// platforms, and registries in GitOps workflows.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// CredentialType represents the type of credential.
type CredentialType string

const (
	// CredentialTypeGit represents Git repository credentials.
	CredentialTypeGit CredentialType = "git"
	// CredentialTypePlatform represents platform credentials (K8s, OpenShift).
	CredentialTypePlatform CredentialType = "platform"
	// CredentialTypeRegistry represents container registry credentials.
	CredentialTypeRegistry CredentialType = "registry"
)

// Method represents the authentication method.
type Method string

const (
	// MethodSSH uses SSH key authentication.
	MethodSSH Method = "ssh"
	// MethodToken uses personal access token.
	MethodToken Method = "token"
	// MethodOAuth uses OAuth authentication.
	MethodOAuth Method = "oauth"
	// MethodBasic uses basic auth (username/password).
	MethodBasic Method = "basic"
	// MethodServiceAccount uses Kubernetes service account.
	MethodServiceAccount Method = "service-account"
	// MethodOIDC uses OpenID Connect.
	MethodOIDC Method = "oidc"
	// MethodAWSIRSA uses AWS IAM Roles for Service Accounts.
	MethodAWSIRSA Method = "aws-irsa"
	// MethodAzureAAD uses Azure Active Directory Pod Identity.
	MethodAzureAAD Method = "azure-aad"
)

// GitProvider represents a Git hosting provider.
type GitProvider string

const (
	GitProviderGitHub      GitProvider = "github"
	GitProviderGitLab      GitProvider = "gitlab"
	GitProviderBitbucket   GitProvider = "bitbucket"
	GitProviderAzureDevOps GitProvider = "azure-devops"
	GitProviderGitea       GitProvider = "gitea"
)

// PlatformType represents the target platform type.
type PlatformType string

const (
	PlatformKubernetes PlatformType = "kubernetes"
	PlatformOpenShift  PlatformType = "openshift"
	PlatformAWS        PlatformType = "aws"
	PlatformAzure      PlatformType = "azure"
	PlatformGCP        PlatformType = "gcp"
)

// SecretFormat represents the format for storing secrets.
type SecretFormat string

const (
	SecretFormatPlain          SecretFormat = "plain"
	SecretFormatSealed         SecretFormat = "sealed"
	SecretFormatSops           SecretFormat = "sops"
	SecretFormatExternalSecret SecretFormat = "external-secret"
	SecretFormatVault          SecretFormat = "vault"
)

// Credential represents a stored credential.
type Credential struct {
	// Name is a unique identifier for the credential
	Name string `yaml:"name" json:"name"`
	// Type is the credential type (git, platform, registry)
	Type CredentialType `yaml:"type" json:"type"`
	// Provider is the service provider (github, gitlab, openshift, etc.)
	Provider string `yaml:"provider" json:"provider"`
	// Method is the authentication method
	Method Method `yaml:"method" json:"method"`
	// Data contains the actual credential data
	Data CredentialData `yaml:"data" json:"data"`
	// Metadata contains additional information
	Metadata CredentialMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	// CreatedAt is when the credential was created
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	// UpdatedAt is when the credential was last updated
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

// CredentialData holds the actual credential values.
type CredentialData struct {
	// Token is the access token or API key
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
	// Username for basic auth
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Password for basic auth
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	// SSHPrivateKey is the SSH private key content
	SSHPrivateKey string `yaml:"ssh_private_key,omitempty" json:"ssh_private_key,omitempty"`
	// SSHPublicKey is the SSH public key content
	SSHPublicKey string `yaml:"ssh_public_key,omitempty" json:"ssh_public_key,omitempty"`
	// SSHKnownHosts is the SSH known_hosts content
	SSHKnownHosts string `yaml:"ssh_known_hosts,omitempty" json:"ssh_known_hosts,omitempty"`
	// ClientID for OAuth
	ClientID string `yaml:"client_id,omitempty" json:"client_id,omitempty"`
	// ClientSecret for OAuth
	ClientSecret string `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`
	// CACert is the CA certificate for TLS verification
	CACert string `yaml:"ca_cert,omitempty" json:"ca_cert,omitempty"`
	// TLSCert is the client certificate
	TLSCert string `yaml:"tls_cert,omitempty" json:"tls_cert,omitempty"`
	// TLSKey is the client private key
	TLSKey string `yaml:"tls_key,omitempty" json:"tls_key,omitempty"`
	// AWSRoleARN for AWS IRSA
	AWSRoleARN string `yaml:"aws_role_arn,omitempty" json:"aws_role_arn,omitempty"`
	// AzureTenantID for Azure AAD
	AzureTenantID string `yaml:"azure_tenant_id,omitempty" json:"azure_tenant_id,omitempty"`
	// AzureClientID for Azure AAD
	AzureClientID string `yaml:"azure_client_id,omitempty" json:"azure_client_id,omitempty"`
}

// CredentialMetadata contains additional information about a credential.
type CredentialMetadata struct {
	// Description of the credential
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// URL is the target URL (e.g., Git repo URL, registry URL)
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Namespace is the K8s namespace where secrets should be created
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// SecretName is the name of the K8s secret to generate
	SecretName string `yaml:"secret_name,omitempty" json:"secret_name,omitempty"`
	// Labels to apply to generated secrets
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Annotations to apply to generated secrets
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	// ExpiresAt is when the credential expires
	ExpiresAt *time.Time `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// Manager handles credential operations.
type Manager struct {
	store        Store
	secretFormat SecretFormat
}

// Store interface for credential storage.
type Store interface {
	// Save stores a credential
	Save(ctx context.Context, cred *Credential) error
	// Get retrieves a credential by name
	Get(ctx context.Context, name string) (*Credential, error)
	// List returns all credentials of a given type
	List(ctx context.Context, credType CredentialType) ([]*Credential, error)
	// Delete removes a credential
	Delete(ctx context.Context, name string) error
	// Exists checks if a credential exists
	Exists(ctx context.Context, name string) (bool, error)
}

// NewManager creates a new credential manager.
func NewManager(store Store, format SecretFormat) *Manager {
	return &Manager{
		store:        store,
		secretFormat: format,
	}
}

// AddGitCredential adds a Git provider credential.
func (m *Manager) AddGitCredential(ctx context.Context, opts *GitCredentialOptions) (*Credential, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid git credential options: %w", err)
	}

	cred := &Credential{
		Name:      opts.Name,
		Type:      CredentialTypeGit,
		Provider:  string(opts.Provider),
		Method:    opts.Method,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: CredentialMetadata{
			Description: opts.Description,
			URL:         opts.URL,
			Namespace:   opts.Namespace,
			SecretName:  opts.SecretName,
		},
	}

	// Set credential data based on method
	switch opts.Method {
	case MethodSSH:
		cred.Data.SSHPrivateKey = opts.SSHPrivateKey
		cred.Data.SSHPublicKey = opts.SSHPublicKey
		cred.Data.SSHKnownHosts = opts.SSHKnownHosts
	case MethodToken:
		cred.Data.Token = opts.Token
		cred.Data.Username = opts.Username
	case MethodBasic:
		cred.Data.Username = opts.Username
		cred.Data.Password = opts.Password
	case MethodOAuth:
		cred.Data.ClientID = opts.ClientID
		cred.Data.ClientSecret = opts.ClientSecret
		cred.Data.Token = opts.Token
	}

	if err := m.store.Save(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}

	return cred, nil
}

// GitCredentialOptions contains options for creating Git credentials.
type GitCredentialOptions struct {
	Name          string
	Provider      GitProvider
	Method        Method
	URL           string
	Description   string
	Namespace     string
	SecretName    string
	Token         string
	Username      string
	Password      string
	SSHPrivateKey string
	SSHPublicKey  string
	SSHKnownHosts string
	ClientID      string
	ClientSecret  string
}

// Validate validates the Git credential options.
func (o *GitCredentialOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("name is required")
	}
	if o.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if o.Method == "" {
		return fmt.Errorf("method is required")
	}

	switch o.Method {
	case MethodSSH:
		if o.SSHPrivateKey == "" {
			return fmt.Errorf("ssh_private_key is required for SSH auth")
		}
	case MethodToken:
		if o.Token == "" {
			return fmt.Errorf("token is required for token auth")
		}
	case MethodBasic:
		if o.Username == "" || o.Password == "" {
			return fmt.Errorf("username and password are required for basic auth")
		}
	case MethodOAuth:
		if o.Token == "" && (o.ClientID == "" || o.ClientSecret == "") {
			return fmt.Errorf("token or client_id/client_secret required for OAuth")
		}
	}

	return nil
}

// AddPlatformCredential adds a platform credential (K8s, OpenShift, Cloud).
func (m *Manager) AddPlatformCredential(ctx context.Context, opts *PlatformCredentialOptions) (*Credential, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid platform credential options: %w", err)
	}

	cred := &Credential{
		Name:      opts.Name,
		Type:      CredentialTypePlatform,
		Provider:  string(opts.Platform),
		Method:    opts.Method,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: CredentialMetadata{
			Description: opts.Description,
			URL:         opts.URL,
			Namespace:   opts.Namespace,
			SecretName:  opts.SecretName,
		},
	}

	// Set credential data based on method
	switch opts.Method {
	case MethodToken:
		cred.Data.Token = opts.Token
	case MethodServiceAccount:
		cred.Data.Token = opts.Token
		cred.Data.CACert = opts.CACert
	case MethodOIDC:
		cred.Data.ClientID = opts.ClientID
		cred.Data.ClientSecret = opts.ClientSecret
	case MethodAWSIRSA:
		cred.Data.AWSRoleARN = opts.AWSRoleARN
	case MethodAzureAAD:
		cred.Data.AzureTenantID = opts.AzureTenantID
		cred.Data.AzureClientID = opts.AzureClientID
	case MethodBasic:
		cred.Data.Username = opts.Username
		cred.Data.Password = opts.Password
	}

	if err := m.store.Save(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}

	return cred, nil
}

// PlatformCredentialOptions contains options for creating platform credentials.
type PlatformCredentialOptions struct {
	Name          string
	Platform      PlatformType
	Method        Method
	URL           string
	Description   string
	Namespace     string
	SecretName    string
	Token         string
	Username      string
	Password      string
	CACert        string
	ClientID      string
	ClientSecret  string
	AWSRoleARN    string
	AzureTenantID string
	AzureClientID string
}

// Validate validates the platform credential options.
func (o *PlatformCredentialOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("name is required")
	}
	if o.Platform == "" {
		return fmt.Errorf("platform is required")
	}
	if o.Method == "" {
		return fmt.Errorf("method is required")
	}

	switch o.Method {
	case MethodToken, MethodServiceAccount:
		if o.Token == "" {
			return fmt.Errorf("token is required for token/service-account auth")
		}
	case MethodOIDC:
		if o.ClientID == "" {
			return fmt.Errorf("client_id is required for OIDC auth")
		}
	case MethodAWSIRSA:
		if o.AWSRoleARN == "" {
			return fmt.Errorf("aws_role_arn is required for AWS IRSA")
		}
	case MethodAzureAAD:
		if o.AzureTenantID == "" || o.AzureClientID == "" {
			return fmt.Errorf("azure_tenant_id and azure_client_id required for Azure AAD")
		}
	case MethodBasic:
		if o.Username == "" || o.Password == "" {
			return fmt.Errorf("username and password required for basic auth")
		}
	}

	return nil
}

// AddRegistryCredential adds a container registry credential.
func (m *Manager) AddRegistryCredential(ctx context.Context, opts *RegistryCredentialOptions) (*Credential, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registry credential options: %w", err)
	}

	cred := &Credential{
		Name:      opts.Name,
		Type:      CredentialTypeRegistry,
		Provider:  opts.Registry,
		Method:    MethodBasic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Data: CredentialData{
			Username: opts.Username,
			Password: opts.Password,
		},
		Metadata: CredentialMetadata{
			Description: opts.Description,
			URL:         opts.URL,
			Namespace:   opts.Namespace,
			SecretName:  opts.SecretName,
		},
	}

	if err := m.store.Save(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}

	return cred, nil
}

// RegistryCredentialOptions contains options for creating registry credentials.
type RegistryCredentialOptions struct {
	Name        string
	Registry    string
	URL         string
	Username    string
	Password    string
	Description string
	Namespace   string
	SecretName  string
}

// Validate validates the registry credential options.
func (o *RegistryCredentialOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("name is required")
	}
	if o.URL == "" {
		return fmt.Errorf("url is required")
	}
	if o.Username == "" || o.Password == "" {
		return fmt.Errorf("username and password are required")
	}
	return nil
}

// GetCredential retrieves a credential by name.
func (m *Manager) GetCredential(ctx context.Context, name string) (*Credential, error) {
	return m.store.Get(ctx, name)
}

// ListCredentials lists all credentials of a given type.
func (m *Manager) ListCredentials(ctx context.Context, credType CredentialType) ([]*Credential, error) {
	return m.store.List(ctx, credType)
}

// DeleteCredential deletes a credential by name.
func (m *Manager) DeleteCredential(ctx context.Context, name string) error {
	return m.store.Delete(ctx, name)
}

// TestCredential tests if a credential is valid.
func (m *Manager) TestCredential(ctx context.Context, name string) (*TestResult, error) {
	cred, err := m.store.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	result := &TestResult{
		Name:     name,
		Type:     cred.Type,
		Provider: cred.Provider,
		TestedAt: time.Now(),
	}

	switch cred.Type {
	case CredentialTypeGit:
		result.Success, result.Message = m.testGitCredential(ctx, cred)
	case CredentialTypePlatform:
		result.Success, result.Message = m.testPlatformCredential(ctx, cred)
	case CredentialTypeRegistry:
		result.Success, result.Message = m.testRegistryCredential(ctx, cred)
	default:
		result.Success = false
		result.Message = fmt.Sprintf("unsupported credential type: %s", cred.Type)
	}

	return result, nil
}

// TestResult contains the result of testing a credential.
type TestResult struct {
	Name     string         `json:"name"`
	Type     CredentialType `json:"type"`
	Provider string         `json:"provider"`
	Success  bool           `json:"success"`
	Message  string         `json:"message"`
	TestedAt time.Time      `json:"tested_at"`
}

func (m *Manager) testGitCredential(_ context.Context, cred *Credential) (success bool, message string) {
	// Basic validation - actual testing would involve Git operations
	switch cred.Method {
	case MethodSSH:
		if cred.Data.SSHPrivateKey == "" {
			return false, "SSH private key is empty"
		}
		return true, "SSH key is present (connection test requires Git operation)"
	case MethodToken:
		if cred.Data.Token == "" {
			return false, "Token is empty"
		}
		if len(cred.Data.Token) < 10 {
			return false, "Token appears to be too short"
		}
		return true, "Token is present and valid format"
	case MethodBasic:
		if cred.Data.Username == "" || cred.Data.Password == "" {
			return false, "Username or password is empty"
		}
		return true, "Basic auth credentials are present"
	default:
		return false, fmt.Sprintf("unsupported auth method: %s", cred.Method)
	}
}

func (m *Manager) testPlatformCredential(_ context.Context, cred *Credential) (success bool, message string) {
	switch cred.Method {
	case MethodToken, MethodServiceAccount:
		if cred.Data.Token == "" {
			return false, "Token is empty"
		}
		return true, "Token is present"
	case MethodOIDC:
		if cred.Data.ClientID == "" {
			return false, "Client ID is empty"
		}
		return true, "OIDC credentials are present"
	case MethodAWSIRSA:
		if cred.Data.AWSRoleARN == "" {
			return false, "AWS Role ARN is empty"
		}
		return true, "AWS IRSA configuration is present"
	case MethodAzureAAD:
		if cred.Data.AzureTenantID == "" || cred.Data.AzureClientID == "" {
			return false, "Azure tenant ID or client ID is empty"
		}
		return true, "Azure AAD configuration is present"
	default:
		return false, fmt.Sprintf("unsupported auth method: %s", cred.Method)
	}
}

func (m *Manager) testRegistryCredential(_ context.Context, cred *Credential) (success bool, message string) {
	if cred.Data.Username == "" || cred.Data.Password == "" {
		return false, "Username or password is empty"
	}
	return true, "Registry credentials are present"
}

// GenerateKubernetesSecret generates a Kubernetes Secret manifest for a credential.
func (m *Manager) GenerateKubernetesSecret(ctx context.Context, name string) (string, error) {
	cred, err := m.store.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("credential not found: %w", err)
	}

	namespace := cred.Metadata.Namespace
	if namespace == "" {
		namespace = "default"
	}

	secretName := cred.Metadata.SecretName
	if secretName == "" {
		secretName = cred.Name
	}

	labels := cred.Metadata.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app.kubernetes.io/managed-by"] = "gitopsi"

	switch cred.Type {
	case CredentialTypeGit:
		return m.generateGitSecret(cred, namespace, secretName, labels)
	case CredentialTypePlatform:
		return m.generatePlatformSecret(cred, namespace, secretName, labels)
	case CredentialTypeRegistry:
		return m.generateRegistrySecret(cred, namespace, secretName, labels)
	default:
		return "", fmt.Errorf("unsupported credential type: %s", cred.Type)
	}
}

func (m *Manager) generateGitSecret(cred *Credential, namespace, secretName string, labels map[string]string) (string, error) {
	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      secretName,
			"namespace": namespace,
			"labels":    labels,
		},
		"type": "Opaque",
	}

	stringData := make(map[string]string)

	switch cred.Method {
	case MethodSSH:
		secret["type"] = "kubernetes.io/ssh-auth"
		stringData["ssh-privatekey"] = cred.Data.SSHPrivateKey
		if cred.Data.SSHKnownHosts != "" {
			stringData["known_hosts"] = cred.Data.SSHKnownHosts
		}
	case MethodToken:
		stringData["username"] = cred.Data.Username
		stringData["password"] = cred.Data.Token
	case MethodBasic:
		secret["type"] = "kubernetes.io/basic-auth"
		stringData["username"] = cred.Data.Username
		stringData["password"] = cred.Data.Password
	}

	secret["stringData"] = stringData

	return toYAML(secret)
}

func (m *Manager) generatePlatformSecret(cred *Credential, namespace, secretName string, labels map[string]string) (string, error) {
	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      secretName,
			"namespace": namespace,
			"labels":    labels,
		},
		"type": "Opaque",
	}

	stringData := make(map[string]string)

	switch cred.Method {
	case MethodToken, MethodServiceAccount:
		stringData["token"] = cred.Data.Token
		if cred.Data.CACert != "" {
			stringData["ca.crt"] = cred.Data.CACert
		}
	case MethodOIDC:
		stringData["client-id"] = cred.Data.ClientID
		if cred.Data.ClientSecret != "" {
			stringData["client-secret"] = cred.Data.ClientSecret
		}
	case MethodBasic:
		stringData["username"] = cred.Data.Username
		stringData["password"] = cred.Data.Password
	}

	secret["stringData"] = stringData

	return toYAML(secret)
}

func (m *Manager) generateRegistrySecret(cred *Credential, namespace, secretName string, labels map[string]string) (string, error) {
	// Generate docker-registry type secret
	registryURL := cred.Metadata.URL
	if registryURL == "" {
		registryURL = "https://index.docker.io/v1/"
	}

	authStr := base64.StdEncoding.EncodeToString([]byte(cred.Data.Username + ":" + cred.Data.Password))
	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			registryURL: map[string]string{
				"username": cred.Data.Username,
				"password": cred.Data.Password,
				"auth":     authStr,
			},
		},
	}
	dockerConfigBytes, err := json.Marshal(dockerConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal docker config: %w", err)
	}
	dockerConfigJSON := string(dockerConfigBytes)

	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      secretName,
			"namespace": namespace,
			"labels":    labels,
		},
		"type": "kubernetes.io/dockerconfigjson",
		"stringData": map[string]string{
			".dockerconfigjson": dockerConfigJSON,
		},
	}

	return toYAML(secret)
}

// GenerateArgoCDRepoSecret generates an ArgoCD repository secret.
func (m *Manager) GenerateArgoCDRepoSecret(ctx context.Context, name, argoCDNamespace string) (string, error) {
	cred, err := m.store.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("credential not found: %w", err)
	}

	if cred.Type != CredentialTypeGit {
		return "", fmt.Errorf("credential must be of type 'git'")
	}

	secretName := cred.Metadata.SecretName
	if secretName == "" {
		secretName = fmt.Sprintf("repo-%s", cred.Name)
	}

	labels := map[string]string{
		"argocd.argoproj.io/secret-type": "repository",
		"app.kubernetes.io/managed-by":   "gitopsi",
	}

	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      secretName,
			"namespace": argoCDNamespace,
			"labels":    labels,
		},
		"type": "Opaque",
	}

	stringData := map[string]string{
		"type": "git",
		"url":  cred.Metadata.URL,
	}

	switch cred.Method {
	case MethodSSH:
		stringData["sshPrivateKey"] = cred.Data.SSHPrivateKey
	case MethodToken:
		stringData["username"] = cred.Data.Username
		if stringData["username"] == "" {
			stringData["username"] = "git"
		}
		stringData["password"] = cred.Data.Token
	case MethodBasic:
		stringData["username"] = cred.Data.Username
		stringData["password"] = cred.Data.Password
	}

	secret["stringData"] = stringData

	return toYAML(secret)
}

// GenerateFluxGitRepositorySecret generates a Flux GitRepository secret.
func (m *Manager) GenerateFluxGitRepositorySecret(ctx context.Context, name, namespace string) (string, error) {
	cred, err := m.store.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("credential not found: %w", err)
	}

	if cred.Type != CredentialTypeGit {
		return "", fmt.Errorf("credential must be of type 'git'")
	}

	secretName := cred.Metadata.SecretName
	if secretName == "" {
		secretName = fmt.Sprintf("flux-git-%s", cred.Name)
	}

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "gitopsi",
	}

	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      secretName,
			"namespace": namespace,
			"labels":    labels,
		},
		"type": "Opaque",
	}

	stringData := make(map[string]string)

	switch cred.Method {
	case MethodSSH:
		stringData["identity"] = cred.Data.SSHPrivateKey
		stringData["identity.pub"] = cred.Data.SSHPublicKey
		if cred.Data.SSHKnownHosts != "" {
			stringData["known_hosts"] = cred.Data.SSHKnownHosts
		}
	case MethodToken:
		stringData["username"] = cred.Data.Username
		if stringData["username"] == "" {
			stringData["username"] = "git"
		}
		stringData["password"] = cred.Data.Token
	case MethodBasic:
		stringData["username"] = cred.Data.Username
		stringData["password"] = cred.Data.Password
	}

	secret["stringData"] = stringData

	return toYAML(secret)
}

// LoadSSHKeyFromFile loads an SSH private key from a file.
func LoadSSHKeyFromFile(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH key: %w", err)
	}

	return string(data), nil
}

// LoadSSHKnownHosts loads known_hosts from a file.
func LoadSSHKnownHosts(path string) (string, error) {
	if path == "" {
		// Try default location
		home, err := os.UserHomeDir()
		if err != nil {
			return "", nil
		}
		path = filepath.Join(home, ".ssh", "known_hosts")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil // Known hosts is optional
	}

	return string(data), nil
}

// GetTokenFromEnv retrieves a token from environment variable.
func GetTokenFromEnv(envVar string) string {
	return os.Getenv(envVar)
}

func toYAML(v interface{}) (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MaskCredential returns a masked version of the credential for display.
func MaskCredential(value string) string {
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
