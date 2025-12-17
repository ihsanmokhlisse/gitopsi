package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/bootstrap"
	"github.com/ihsanmokhlisse/gitopsi/internal/cluster"
	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/generator"
	"github.com/ihsanmokhlisse/gitopsi/internal/git"
	outputpkg "github.com/ihsanmokhlisse/gitopsi/internal/output"
	"github.com/ihsanmokhlisse/gitopsi/internal/prompt"
)

var (
	gitURL        string
	gitToken      string
	pushAfterInit bool
	clusterURL    string
	clusterToken  string
	bootstrapFlag bool
	bootstrapMode string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new GitOps repository",
	Long: `Initialize a new GitOps repository structure with all necessary
manifests, documentation, and scripts.

Can run in interactive mode (default) or with a config file.
Optionally push to Git repository and bootstrap GitOps tool on cluster.

Examples:
  gitopsi init                                    # Interactive mode
  gitopsi init --config gitops.yaml               # Config file mode
  gitopsi init --dry-run                          # Preview without writing
  gitopsi init --git-url <url> --push             # Generate and push to Git
  gitopsi init --git-url <url> --cluster <url> --bootstrap  # Full E2E setup`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&gitURL, "git-url", "", "Git repository URL")
	initCmd.Flags().StringVar(&gitToken, "git-token", "", "Git authentication token (or use GITOPSI_GIT_TOKEN env)")
	initCmd.Flags().BoolVar(&pushAfterInit, "push", false, "Push generated code to Git repository")
	initCmd.Flags().StringVar(&clusterURL, "cluster", "", "Target cluster URL")
	initCmd.Flags().StringVar(&clusterToken, "cluster-token", "", "Cluster authentication token (or use GITOPSI_CLUSTER_TOKEN env)")
	initCmd.Flags().BoolVar(&bootstrapFlag, "bootstrap", false, "Bootstrap GitOps tool on cluster")
	initCmd.Flags().StringVar(&bootstrapMode, "bootstrap-mode", "helm", "Bootstrap mode: helm, olm, manifest")
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var cfg *config.Config
	var err error

	if cfgFile != "" {
		fmt.Printf("📄 Loading config from: %s\n", cfgFile)
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		fmt.Println("🎯 gitopsi - GitOps Repository Generator")
		fmt.Println()
		cfg, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
	}

	applyFlagOverrides(cfg)

	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	outputDir := GetOutput()
	if outputDir == "" {
		outputDir = "."
	}

	var absOutput string
	absOutput, err = filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}

	projectPath := filepath.Join(absOutput, cfg.Project.Name)

	if !dryRun {
		if _, statErr := os.Stat(projectPath); statErr == nil {
			return fmt.Errorf("directory already exists: %s", cfg.Project.Name)
		}
	}

	// Step 1: Authenticate to Git if needed
	var gitProvider git.Provider
	if shouldPush(cfg) {
		fmt.Println("\n🔐 Authenticating to Git...")
		gitProvider, err = authenticateGit(ctx, cfg)
		if err != nil {
			return fmt.Errorf("git authentication failed: %w", err)
		}
		fmt.Printf("   ✓ Connected to %s\n", gitProvider.GetInstance())
	}

	// Step 2: Generate files
	writer := outputpkg.New(absOutput, dryRun, verbose)
	gen := generator.New(cfg, writer, verbose)

	if dryRun {
		fmt.Println("\n🔍 DRY RUN - No files will be written")
	}

	fmt.Println("\n📁 Generating GitOps repository...")
	if genErr := gen.Generate(); genErr != nil {
		return genErr
	}

	if dryRun {
		fmt.Println("\n🔍 DRY RUN complete - no files were written")
		return nil
	}

	// Step 3: Push to Git if requested
	if shouldPush(cfg) && gitProvider != nil {
		fmt.Println("\n📤 Pushing to repository...")
		if pushErr := pushToGit(ctx, cfg, gitProvider, projectPath); pushErr != nil {
			return fmt.Errorf("failed to push to Git: %w", pushErr)
		}
		fmt.Printf("   ✓ Pushed to %s\n", cfg.Git.Branch)
	}

	// Step 4: Authenticate to cluster if needed
	var clusterConn *cluster.Cluster
	if shouldBootstrap(cfg) {
		fmt.Println("\n🔐 Authenticating to cluster...")
		clusterConn, err = authenticateCluster(ctx, cfg)
		if err != nil {
			return fmt.Errorf("cluster authentication failed: %w", err)
		}
		fmt.Printf("   ✓ Connected to cluster\n")
	}

	// Step 5: Bootstrap GitOps tool if requested
	if shouldBootstrap(cfg) && clusterConn != nil {
		fmt.Printf("\n🚀 Bootstrapping %s...\n", cfg.GitOpsTool)
		result, err := bootstrapCluster(ctx, cfg, clusterConn)
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}
		printBootstrapResult(result)
	}

	// Print summary
	printSummary(cfg, projectPath)

	return nil
}

func applyFlagOverrides(cfg *config.Config) {
	if gitURL != "" {
		cfg.Git.URL = gitURL
		cfg.Output.URL = gitURL
	}

	token := gitToken
	if token == "" {
		token = os.Getenv("GITOPSI_GIT_TOKEN")
	}
	if token != "" {
		cfg.Git.Auth.Token = token
		cfg.Git.Auth.Method = "token"
	}

	if pushAfterInit {
		cfg.Git.PushOnInit = true
	}

	if clusterURL != "" {
		cfg.Cluster.URL = clusterURL
	}

	cToken := clusterToken
	if cToken == "" {
		cToken = os.Getenv("GITOPSI_CLUSTER_TOKEN")
	}
	if cToken != "" {
		cfg.Cluster.Auth.Token = cToken
		cfg.Cluster.Auth.Method = "token"
	}

	if bootstrapFlag {
		cfg.Bootstrap.Enabled = true
	}

	if bootstrapMode != "" {
		cfg.Bootstrap.Mode = bootstrapMode
	}
}

func shouldPush(cfg *config.Config) bool {
	return cfg.Git.PushOnInit && cfg.Git.URL != ""
}

func shouldBootstrap(cfg *config.Config) bool {
	return cfg.Bootstrap.Enabled && cfg.Cluster.URL != ""
}

func authenticateGit(ctx context.Context, cfg *config.Config) (git.Provider, error) {
	providerType, instance, err := git.DetectProvider(cfg.Git.URL)
	if err != nil {
		return nil, err
	}

	provider, err := git.NewProvider(providerType, instance)
	if err != nil {
		return nil, err
	}

	authOpts := git.AuthOptions{
		Method: git.AuthMethod(cfg.Git.Auth.Method),
		Token:  cfg.Git.Auth.Token,
		SSHKey: cfg.Git.Auth.SSHKey,
	}

	if err := provider.Authenticate(ctx, authOpts); err != nil {
		return nil, err
	}

	if err := provider.ValidateAccess(ctx); err != nil {
		return nil, err
	}

	return provider, nil
}

func pushToGit(ctx context.Context, cfg *config.Config, provider git.Provider, projectPath string) error {
	// Initialize git repo
	if err := runGitCommand(ctx, projectPath, "init"); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// Add remote
	if err := runGitCommand(ctx, projectPath, "remote", "add", "origin", cfg.Git.URL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Add all files
	if err := runGitCommand(ctx, projectPath, "add", "."); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Commit
	if err := runGitCommand(ctx, projectPath, "commit", "-m", "feat: Initial GitOps repository structure"); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push
	branch := cfg.Git.Branch
	if branch == "" {
		branch = "main"
	}

	if err := provider.Push(ctx, git.PushOptions{
		Remote:      "origin",
		Branch:      branch,
		SetUpstream: true,
	}); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

func authenticateCluster(ctx context.Context, cfg *config.Config) (*cluster.Cluster, error) {
	c := cluster.New(cfg.Cluster.URL, cfg.Cluster.Name, cluster.Platform(cfg.Cluster.Platform))

	authOpts := &cluster.AuthOptions{
		Method:     cluster.AuthMethod(cfg.Cluster.Auth.Method),
		Token:      cfg.Cluster.Auth.Token,
		TokenEnv:   cfg.Cluster.Auth.TokenEnv,
		Kubeconfig: cfg.Cluster.Kubeconfig,
		Context:    cfg.Cluster.Context,
		CACert:     cfg.Cluster.Auth.CACert,
		SkipTLS:    cfg.Cluster.Auth.SkipTLS,
	}

	if err := c.Authenticate(authOpts); err != nil {
		return nil, err
	}

	if err := c.TestConnection(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func bootstrapCluster(ctx context.Context, cfg *config.Config, c *cluster.Cluster) (*bootstrap.Result, error) {
	opts := &bootstrap.Options{
		Tool:            bootstrap.Tool(cfg.GitOpsTool),
		Mode:            bootstrap.Mode(cfg.Bootstrap.Mode),
		Namespace:       cfg.Bootstrap.Namespace,
		Wait:            cfg.Bootstrap.Wait,
		Timeout:         cfg.Bootstrap.Timeout,
		ConfigureRepo:   cfg.Bootstrap.ConfigureRepo,
		RepoURL:         cfg.Git.URL,
		RepoBranch:      cfg.Git.Branch,
		RepoPath:        cfg.GitOpsTool + "/applicationsets",
		CreateAppOfApps: cfg.Bootstrap.CreateAppOfApps,
		SyncInitial:     cfg.Bootstrap.SyncInitial,
		ProjectName:     cfg.Project.Name,
	}

	b := bootstrap.New(c, opts)
	return b.Bootstrap(ctx)
}

func printBootstrapResult(result *bootstrap.Result) {
	if result.Ready {
		fmt.Printf("   ✓ %s installed in namespace %s\n", result.Tool, result.Namespace)
		if result.URL != "" {
			fmt.Printf("   ✓ UI available at: %s\n", result.URL)
		}
		if result.Username != "" {
			fmt.Printf("   ✓ Username: %s\n", result.Username)
			fmt.Println("   ✓ Password: Run 'gitopsi get-password argocd' to retrieve")
		}
	}
}

func printSummary(cfg *config.Config, projectPath string) {
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
	fmt.Println("✅ GitOps setup complete!")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("\n📁 Project: %s\n", projectPath)

	if cfg.Git.URL != "" && cfg.Git.PushOnInit {
		fmt.Printf("📦 Repository: %s\n", cfg.Git.URL)
	}

	if cfg.Bootstrap.Enabled && cfg.Cluster.URL != "" {
		fmt.Printf("🎯 Cluster: %s\n", cfg.Cluster.URL)
		fmt.Printf("🔄 GitOps Tool: %s\n", cfg.GitOpsTool)
	}

	fmt.Println("\n📚 Next steps:")
	if !cfg.Git.PushOnInit {
		fmt.Println("   1. cd " + cfg.Project.Name)
		fmt.Println("   2. git init && git add . && git commit -m 'Initial commit'")
		fmt.Println("   3. git remote add origin <your-repo-url>")
		fmt.Println("   4. git push -u origin main")
	}
	if !cfg.Bootstrap.Enabled {
		fmt.Println("   5. kubectl apply -f bootstrap/" + cfg.GitOpsTool + "/")
	}
	fmt.Println()
}

func runGitCommand(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w: %s", args[0], err, string(output))
	}
	return nil
}
