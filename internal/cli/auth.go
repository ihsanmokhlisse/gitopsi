package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/auth"
)

var (
	authProvider   string
	authMethod     string
	authToken      string
	authUsername   string
	authPassword   string
	authSSHKeyFile string
	authURL        string
	authNamespace  string
	authSecretName string
	authPlatform   string
	authRegistry   string
	authFormat     string
	authRoleARN    string
	authTenantID   string
	authClientID   string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long: `Manage authentication credentials for Git providers, platforms, and registries.

Examples:
  # Add GitHub token
  gitopsi auth add git --provider github --method token --token $GITHUB_TOKEN

  # Add SSH key for GitLab
  gitopsi auth add git --provider gitlab --method ssh --ssh-key ~/.ssh/id_rsa

  # Add OpenShift credentials
  gitopsi auth add platform --platform openshift --method token --token $OCP_TOKEN

  # Add registry credentials
  gitopsi auth add registry --url registry.example.com --username user --password pass

  # List all credentials
  gitopsi auth list

  # Test credentials
  gitopsi auth test my-github-cred`,
}

var authAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add credentials",
	Long:  "Add Git, platform, or registry credentials.",
}

var authAddGitCmd = &cobra.Command{
	Use:   "git [name]",
	Short: "Add Git provider credentials",
	Long: `Add credentials for Git providers (GitHub, GitLab, Bitbucket, etc.)

Examples:
  gitopsi auth add git github-main --provider github --method token --token $GITHUB_TOKEN
  gitopsi auth add git gitlab-ssh --provider gitlab --method ssh --ssh-key ~/.ssh/id_rsa`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthAddGit,
}

var authAddPlatformCmd = &cobra.Command{
	Use:   "platform [name]",
	Short: "Add platform credentials",
	Long: `Add credentials for platforms (Kubernetes, OpenShift, AWS, Azure, GCP)

Examples:
  gitopsi auth add platform openshift-prod --platform openshift --method token --token $OCP_TOKEN
  gitopsi auth add platform aws-prod --platform aws --method aws-irsa --role-arn arn:aws:iam::xxx`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthAddPlatform,
}

var authAddRegistryCmd = &cobra.Command{
	Use:   "registry [name]",
	Short: "Add container registry credentials",
	Long: `Add credentials for container registries

Examples:
  gitopsi auth add registry docker-hub --url docker.io --username user --password pass
  gitopsi auth add registry quay --url quay.io --username user --password pass`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthAddRegistry,
}

var authListCmd = &cobra.Command{
	Use:   "list [type]",
	Short: "List credentials",
	Long: `List all credentials or filter by type (git, platform, registry)

Examples:
  gitopsi auth list
  gitopsi auth list git
  gitopsi auth list platform`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAuthList,
}

var authTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test credential validity",
	Long: `Test if a credential is valid

Examples:
  gitopsi auth test github-main
  gitopsi auth test openshift-prod`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthTest,
}

var authDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a credential",
	Long: `Delete a stored credential

Examples:
  gitopsi auth delete github-main`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthDelete,
}

var authGenerateCmd = &cobra.Command{
	Use:   "generate [name]",
	Short: "Generate Kubernetes secret from credential",
	Long: `Generate a Kubernetes Secret manifest from a stored credential

Examples:
  gitopsi auth generate github-main
  gitopsi auth generate github-main --format argocd
  gitopsi auth generate github-main --format flux`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthGenerate,
}

func init() {
	rootCmd.AddCommand(authCmd)

	// Add subcommands
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authTestCmd)
	authCmd.AddCommand(authDeleteCmd)
	authCmd.AddCommand(authGenerateCmd)

	// Add type-specific add commands
	authAddCmd.AddCommand(authAddGitCmd)
	authAddCmd.AddCommand(authAddPlatformCmd)
	authAddCmd.AddCommand(authAddRegistryCmd)

	// Git credentials flags
	authAddGitCmd.Flags().StringVar(&authProvider, "provider", "", "Git provider: github, gitlab, bitbucket, azure-devops, gitea")
	authAddGitCmd.Flags().StringVar(&authMethod, "method", "", "Auth method: ssh, token, basic, oauth")
	authAddGitCmd.Flags().StringVar(&authToken, "token", "", "Access token (or use env var)")
	authAddGitCmd.Flags().StringVar(&authUsername, "username", "", "Username for basic auth")
	authAddGitCmd.Flags().StringVar(&authPassword, "password", "", "Password for basic auth")
	authAddGitCmd.Flags().StringVar(&authSSHKeyFile, "ssh-key", "", "Path to SSH private key file")
	authAddGitCmd.Flags().StringVar(&authURL, "url", "", "Git repository URL")
	authAddGitCmd.Flags().StringVar(&authNamespace, "namespace", "", "Kubernetes namespace for generated secret")
	authAddGitCmd.Flags().StringVar(&authSecretName, "secret-name", "", "Name for generated Kubernetes secret")

	// Platform credentials flags
	authAddPlatformCmd.Flags().StringVar(&authPlatform, "platform", "", "Platform: kubernetes, openshift, aws, azure, gcp")
	authAddPlatformCmd.Flags().StringVar(&authMethod, "method", "", "Auth method: token, service-account, oidc, aws-irsa, azure-aad")
	authAddPlatformCmd.Flags().StringVar(&authToken, "token", "", "Access token")
	authAddPlatformCmd.Flags().StringVar(&authURL, "url", "", "Platform API URL")
	authAddPlatformCmd.Flags().StringVar(&authRoleARN, "role-arn", "", "AWS IAM Role ARN for IRSA")
	authAddPlatformCmd.Flags().StringVar(&authTenantID, "tenant-id", "", "Azure tenant ID")
	authAddPlatformCmd.Flags().StringVar(&authClientID, "client-id", "", "Client ID for OIDC/Azure")
	authAddPlatformCmd.Flags().StringVar(&authNamespace, "namespace", "", "Kubernetes namespace")
	authAddPlatformCmd.Flags().StringVar(&authSecretName, "secret-name", "", "Secret name")

	// Registry credentials flags
	authAddRegistryCmd.Flags().StringVar(&authURL, "url", "", "Registry URL")
	authAddRegistryCmd.Flags().StringVar(&authUsername, "username", "", "Registry username")
	authAddRegistryCmd.Flags().StringVar(&authPassword, "password", "", "Registry password")
	authAddRegistryCmd.Flags().StringVar(&authNamespace, "namespace", "", "Kubernetes namespace")
	authAddRegistryCmd.Flags().StringVar(&authSecretName, "secret-name", "", "Secret name")

	// Generate flags
	authGenerateCmd.Flags().StringVar(&authFormat, "format", "k8s", "Output format: k8s, argocd, flux")
	authGenerateCmd.Flags().StringVar(&authNamespace, "namespace", "", "Override namespace for generated secret")

	// Mark required flags
	_ = authAddGitCmd.MarkFlagRequired("provider")
	_ = authAddGitCmd.MarkFlagRequired("method")
	_ = authAddPlatformCmd.MarkFlagRequired("platform")
	_ = authAddPlatformCmd.MarkFlagRequired("method")
	_ = authAddRegistryCmd.MarkFlagRequired("url")
	_ = authAddRegistryCmd.MarkFlagRequired("username")
	_ = authAddRegistryCmd.MarkFlagRequired("password")
}

func getAuthManager() (*auth.Manager, error) {
	storePath := auth.GetDefaultStorePath()
	store, err := auth.NewFileStore(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize credential store: %w", err)
	}
	return auth.NewManager(store, auth.SecretFormatPlain), nil
}

func runAuthAddGit(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	opts := &auth.GitCredentialOptions{
		Name:       name,
		Provider:   auth.GitProvider(authProvider),
		Method:     auth.Method(authMethod),
		URL:        authURL,
		Namespace:  authNamespace,
		SecretName: authSecretName,
		Username:   authUsername,
	}

	// Get token from flag or environment
	switch opts.Method {
	case auth.MethodSSH:
		if authSSHKeyFile == "" {
			return fmt.Errorf("--ssh-key is required for SSH authentication")
		}
		key, err := auth.LoadSSHKeyFromFile(authSSHKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load SSH key: %w", err)
		}
		opts.SSHPrivateKey = key
		knownHosts, _ := auth.LoadSSHKnownHosts("")
		opts.SSHKnownHosts = knownHosts
	case auth.MethodToken:
		opts.Token = getTokenValue(authToken, authProvider)
		if opts.Token == "" {
			return fmt.Errorf("--token is required or set %s_TOKEN environment variable", strings.ToUpper(authProvider))
		}
	case auth.MethodBasic:
		if authUsername == "" || authPassword == "" {
			return fmt.Errorf("--username and --password are required for basic authentication")
		}
		opts.Username = authUsername
		opts.Password = authPassword
	case auth.MethodOAuth:
		opts.Token = getTokenValue(authToken, authProvider)
		opts.ClientID = authClientID
	}

	cred, err := manager.AddGitCredential(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to add credential: %w", err)
	}

	pterm.Success.Printf("Git credential '%s' added successfully\n", cred.Name)
	pterm.Info.Printf("Provider: %s, Method: %s\n", cred.Provider, cred.Method)

	return nil
}

func runAuthAddPlatform(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	opts := &auth.PlatformCredentialOptions{
		Name:       name,
		Platform:   auth.PlatformType(authPlatform),
		Method:     auth.Method(authMethod),
		URL:        authURL,
		Namespace:  authNamespace,
		SecretName: authSecretName,
	}

	switch opts.Method {
	case auth.MethodToken, auth.MethodServiceAccount:
		opts.Token = getTokenValue(authToken, authPlatform)
		if opts.Token == "" {
			return fmt.Errorf("--token is required or set %s_TOKEN environment variable", strings.ToUpper(authPlatform))
		}
	case auth.MethodOIDC:
		opts.ClientID = authClientID
		if opts.ClientID == "" {
			return fmt.Errorf("--client-id is required for OIDC authentication")
		}
	case auth.MethodAWSIRSA:
		opts.AWSRoleARN = authRoleARN
		if opts.AWSRoleARN == "" {
			return fmt.Errorf("--role-arn is required for AWS IRSA")
		}
	case auth.MethodAzureAAD:
		opts.AzureTenantID = authTenantID
		opts.AzureClientID = authClientID
		if opts.AzureTenantID == "" || opts.AzureClientID == "" {
			return fmt.Errorf("--tenant-id and --client-id are required for Azure AAD")
		}
	}

	cred, err := manager.AddPlatformCredential(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to add credential: %w", err)
	}

	pterm.Success.Printf("Platform credential '%s' added successfully\n", cred.Name)
	pterm.Info.Printf("Platform: %s, Method: %s\n", cred.Provider, cred.Method)

	return nil
}

func runAuthAddRegistry(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	opts := &auth.RegistryCredentialOptions{
		Name:       name,
		URL:        authURL,
		Username:   authUsername,
		Password:   authPassword,
		Namespace:  authNamespace,
		SecretName: authSecretName,
	}

	cred, err := manager.AddRegistryCredential(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to add credential: %w", err)
	}

	pterm.Success.Printf("Registry credential '%s' added successfully\n", cred.Name)
	pterm.Info.Printf("URL: %s\n", cred.Metadata.URL)

	return nil
}

func runAuthList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	var credType auth.CredentialType
	if len(args) > 0 {
		credType = auth.CredentialType(args[0])
	}

	creds, err := manager.ListCredentials(ctx, credType)
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	if len(creds) == 0 {
		pterm.Info.Println("No credentials found")
		return nil
	}

	tableData := [][]string{
		{"NAME", "TYPE", "PROVIDER", "METHOD", "URL"},
	}

	for _, cred := range creds {
		url := cred.Metadata.URL
		if len(url) > 40 {
			url = url[:37] + "..."
		}
		tableData = append(tableData, []string{
			cred.Name,
			string(cred.Type),
			cred.Provider,
			string(cred.Method),
			url,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

func runAuthTest(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	pterm.Info.Printf("Testing credential '%s'...\n", name)

	result, err := manager.TestCredential(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to test credential: %w", err)
	}

	if result.Success {
		pterm.Success.Printf("Credential '%s' is valid\n", name)
		pterm.Info.Println(result.Message)
	} else {
		pterm.Error.Printf("Credential '%s' validation failed\n", name)
		pterm.Error.Println(result.Message)
		return fmt.Errorf("credential test failed")
	}

	return nil
}

func runAuthDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	if err := manager.DeleteCredential(ctx, name); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	pterm.Success.Printf("Credential '%s' deleted successfully\n", name)
	return nil
}

func runAuthGenerate(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getAuthManager()
	if err != nil {
		return err
	}

	var output string

	switch authFormat {
	case "k8s", "kubernetes":
		output, err = manager.GenerateKubernetesSecret(ctx, name)
	case "argocd":
		ns := authNamespace
		if ns == "" {
			ns = "argocd"
		}
		output, err = manager.GenerateArgoCDRepoSecret(ctx, name, ns)
	case "flux":
		ns := authNamespace
		if ns == "" {
			ns = "flux-system"
		}
		output, err = manager.GenerateFluxGitRepositorySecret(ctx, name, ns)
	default:
		return fmt.Errorf("unsupported format: %s", authFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to generate secret: %w", err)
	}

	fmt.Println(output)
	return nil
}

func getTokenValue(tokenFlag, provider string) string {
	if tokenFlag != "" {
		// Check if it's an environment variable reference
		if strings.HasPrefix(tokenFlag, "$") {
			envVar := strings.TrimPrefix(tokenFlag, "$")
			return os.Getenv(envVar)
		}
		return tokenFlag
	}

	// Try common environment variables
	envVars := []string{
		strings.ToUpper(provider) + "_TOKEN",
		"GITOPSI_" + strings.ToUpper(provider) + "_TOKEN",
		"GIT_TOKEN",
		"GITOPSI_GIT_TOKEN",
	}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			return val
		}
	}

	return ""
}
