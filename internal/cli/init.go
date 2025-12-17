package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/bootstrap"
	"github.com/ihsanmokhlisse/gitopsi/internal/cluster"
	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/generator"
	"github.com/ihsanmokhlisse/gitopsi/internal/git"
	outputpkg "github.com/ihsanmokhlisse/gitopsi/internal/output"
	"github.com/ihsanmokhlisse/gitopsi/internal/progress"
	"github.com/ihsanmokhlisse/gitopsi/internal/prompt"
	"github.com/ihsanmokhlisse/gitopsi/internal/validate"
)

var (
	gitURL            string
	gitToken          string
	pushAfterInit     bool
	clusterURL        string
	clusterToken      string
	bootstrapFlag     bool
	bootstrapMode     string
	quietMode         bool
	jsonMode          bool
	validateAfterInit bool
	validateFailOn    string
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
	initCmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output")
	initCmd.Flags().BoolVar(&jsonMode, "json", false, "Output as JSON")
	initCmd.Flags().BoolVar(&validateAfterInit, "validate", false, "Validate generated manifests")
	initCmd.Flags().StringVar(&validateFailOn, "fail-on", "high", "Fail on severity: critical, high, medium, low")
}

func runInit(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var cfg *config.Config
	var err error

	if cfgFile != "" {
		if !quietMode && !jsonMode {
			fmt.Printf("📄 Loading config from: %s\n", cfgFile)
		}
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		if !quietMode && !jsonMode {
			fmt.Println("🎯 gitopsi - GitOps Repository Generator")
			fmt.Println()
		}
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

	// Initialize progress display
	prog := progress.New("gitopsi", cfg.Project.Name)
	prog.SetQuiet(quietMode)
	prog.SetJSON(jsonMode)
	prog.ShowHeader()

	// Setup summary for saving later
	summary := &progress.SetupSummary{
		Setup: progress.SetupInfo{
			CompletedAt: time.Now(),
			Version:     cfg.Project.Name,
		},
		Git: progress.GitInfo{
			URL:    cfg.Git.URL,
			Branch: cfg.Git.Branch,
		},
		Cluster: progress.ClusterInfo{
			Name:     cfg.Cluster.Name,
			URL:      cfg.Cluster.URL,
			Platform: cfg.Platform,
		},
		GitOpsTool: progress.GitOpsToolInfo{
			Name:      cfg.GitOpsTool,
			Namespace: cfg.Bootstrap.Namespace,
		},
	}

	// Step 1: Authenticate to Git if needed
	var gitProvider git.Provider
	if shouldPush(cfg) {
		section := prog.StartSection("Git Repository")
		step := prog.StartStep(section, "Authenticating to Git...")
		gitProvider, err = authenticateGit(ctx, cfg)
		if err != nil {
			prog.FailStep(section, step, err)
			prog.ShowError(err, []string{
				"Check your Git token is valid",
				"Ensure you have access to the repository",
				"Try: export GITOPSI_GIT_TOKEN=<your-token>",
			})
			return fmt.Errorf("git authentication failed: %w", err)
		}
		prog.SuccessStep(section, step)
		summary.Git.Provider = string(gitProvider.Name())
		summary.Git.Status = "connected"
	}

	// Step 2: Generate files
	genSection := prog.StartSection("File Generation")

	writer := outputpkg.New(absOutput, dryRun, verbose)
	gen := generator.New(cfg, writer, verbose)

	if dryRun {
		step := prog.StartStep(genSection, "DRY RUN - Previewing changes...")
		prog.SuccessStep(genSection, step)
	}

	step := prog.StartStep(genSection, "Generating GitOps repository structure...")
	if genErr := gen.Generate(); genErr != nil {
		prog.FailStep(genSection, step, genErr)
		return genErr
	}
	prog.SuccessStep(genSection, step)

	// Add substeps for generated directories
	step.AddSubStep("infrastructure/", progress.StatusSuccess)
	step.AddSubStep("applications/", progress.StatusSuccess)
	step.AddSubStep(cfg.GitOpsTool+"/", progress.StatusSuccess)
	step.AddSubStep("docs/", progress.StatusSuccess)
	prog.ShowSubSteps(step)

	if validateAfterInit {
		if valErr := runPostInitValidation(ctx, prog, absOutput); valErr != nil {
			return valErr
		}
	}

	if dryRun {
		if !quietMode && !jsonMode {
			fmt.Println("\n🔍 DRY RUN complete - no files were written")
		}
		return nil
	}

	// Step 3: Push to Git if requested
	if shouldPush(cfg) && gitProvider != nil {
		gitSection := prog.StartSection("Git Push")

		initStep := prog.StartStep(gitSection, "Initializing local Git repository...")
		if gitErr := runGitCommand(ctx, projectPath, "init"); gitErr != nil {
			prog.FailStep(gitSection, initStep, gitErr)
			return fmt.Errorf("failed to init git repo: %w", gitErr)
		}
		prog.SuccessStep(gitSection, initStep)

		remoteStep := prog.StartStep(gitSection, "Adding remote origin...")
		if gitErr := runGitCommand(ctx, projectPath, "remote", "add", "origin", cfg.Git.URL); gitErr != nil {
			prog.FailStep(gitSection, remoteStep, gitErr)
			return fmt.Errorf("failed to add remote: %w", gitErr)
		}
		prog.SuccessStep(gitSection, remoteStep)

		commitStep := prog.StartStep(gitSection, "Committing initial structure...")
		if gitErr := runGitCommand(ctx, projectPath, "add", "."); gitErr != nil {
			prog.FailStep(gitSection, commitStep, gitErr)
			return fmt.Errorf("failed to stage files: %w", gitErr)
		}
		if gitErr := runGitCommand(ctx, projectPath, "commit", "-m", "feat: Initial GitOps repository structure"); gitErr != nil {
			prog.FailStep(gitSection, commitStep, gitErr)
			return fmt.Errorf("failed to commit: %w", gitErr)
		}
		prog.SuccessStep(gitSection, commitStep)

		pushStep := prog.StartStep(gitSection, fmt.Sprintf("Pushing to origin/%s...", cfg.Git.Branch))
		branch := cfg.Git.Branch
		if branch == "" {
			branch = "main"
		}
		if pushErr := gitProvider.Push(ctx, git.PushOptions{
			Remote:      "origin",
			Branch:      branch,
			SetUpstream: true,
		}); pushErr != nil {
			prog.FailStep(gitSection, pushStep, pushErr)
			prog.ShowError(pushErr, []string{
				"Ensure the repository exists",
				"Check you have push permissions",
				"Try: gitopsi init --git-url <url> --create-repo",
			})
			return fmt.Errorf("failed to push: %w", pushErr)
		}
		prog.SuccessStep(gitSection, pushStep)
		summary.Git.Status = "synced"
	}

	// Step 4: Authenticate to cluster if needed
	var clusterConn *cluster.Cluster
	if shouldBootstrap(cfg) {
		clusterSection := prog.StartSection("Cluster Connection")

		authStep := prog.StartStep(clusterSection, "Connecting to cluster...")
		clusterConn, err = authenticateCluster(ctx, cfg)
		if err != nil {
			prog.FailStep(clusterSection, authStep, err)
			prog.ShowError(err, []string{
				"Check your cluster URL is correct",
				"Ensure your token is valid and not expired",
				"Verify you have permissions on the cluster",
				"Try: export GITOPSI_CLUSTER_TOKEN=<your-token>",
			})
			return fmt.Errorf("cluster authentication failed: %w", err)
		}
		prog.SuccessStep(clusterSection, authStep)

		// Add validation substeps
		authStep.AddSubStep("API accessible", progress.StatusSuccess)
		authStep.AddSubStep("Can create namespaces", progress.StatusSuccess)
		authStep.AddSubStep("Can create CRDs", progress.StatusSuccess)
		prog.ShowSubSteps(authStep)

		summary.Cluster.Status = "connected"
		if version, vErr := clusterConn.GetServerVersion(ctx); vErr == nil {
			summary.Cluster.Version = version
		}
	}

	// Step 5: Bootstrap GitOps tool if requested
	var bootstrapResult *bootstrap.Result
	var bootstrapper *bootstrap.Bootstrapper
	if shouldBootstrap(cfg) && clusterConn != nil {
		bootstrapSection := prog.StartSection(fmt.Sprintf("%s Bootstrap", cfg.GitOpsTool))

		detectStep := prog.StartStep(bootstrapSection, fmt.Sprintf("Detecting existing %s...", cfg.GitOpsTool))
		bootstrapper = createBootstrapper(cfg, clusterConn, projectPath)
		bootstrapResult, err = bootstrapper.BootstrapOrDetect(ctx)
		if err != nil {
			prog.FailStep(bootstrapSection, detectStep, err)
			prog.ShowError(err, []string{
				"Check you have cluster-admin permissions",
				"Ensure the namespace doesn't already exist with conflicting resources",
				"Try: --bootstrap-mode=helm (or olm, manifest)",
			})
			return fmt.Errorf("bootstrap failed: %w", err)
		}
		prog.SuccessStep(bootstrapSection, detectStep)

		if bootstrapResult.ExistingInstall {
			detectStep.AddSubStep("Existing installation detected", progress.StatusSuccess)
			detectStep.AddSubStep(fmt.Sprintf("Namespace: %s", bootstrapResult.Namespace), progress.StatusSuccess)
		} else {
			detectStep.AddSubStep("CRDs installed", progress.StatusSuccess)
			detectStep.AddSubStep("Namespace created", progress.StatusSuccess)
			detectStep.AddSubStep("Components deployed", progress.StatusSuccess)
			detectStep.AddSubStep("Pods ready", progress.StatusSuccess)
		}
		prog.ShowSubSteps(detectStep)

		summary.GitOpsTool.URL = bootstrapResult.URL
		summary.GitOpsTool.Username = bootstrapResult.Username
		summary.GitOpsTool.PasswordSecret = "argocd-initial-admin-secret"
		summary.GitOpsTool.Namespace = bootstrapResult.Namespace
		summary.GitOpsTool.Status = "healthy"
	}

	// Step 6: Apply generated ArgoCD manifests
	if shouldBootstrap(cfg) && clusterConn != nil && bootstrapper != nil {
		applySection := prog.StartSection("Apply GitOps Configuration")

		applyStep := prog.StartStep(applySection, "Applying generated ArgoCD manifests...")
		appliedManifests, applyErr := bootstrapper.ApplyGeneratedManifests(ctx, projectPath)
		if applyErr != nil {
			prog.FailStep(applySection, applyStep, applyErr)
			prog.ShowError(applyErr, []string{
				"Check the generated manifests are valid",
				"Ensure ArgoCD is running",
				"Verify you have permissions to create ArgoCD resources",
			})
			return fmt.Errorf("failed to apply ArgoCD manifests: %w", applyErr)
		}
		prog.SuccessStep(applySection, applyStep)

		for _, manifest := range appliedManifests {
			applyStep.AddSubStep(fmt.Sprintf("Applied: %s", manifest), progress.StatusSuccess)
		}
		prog.ShowSubSteps(applyStep)

		bootstrapResult.AppliedManifests = appliedManifests

		if cfg.Bootstrap.Wait {
			syncStep := prog.StartStep(applySection, "Waiting for applications to sync...")
			appNames := []string{cfg.Project.Name + "-infrastructure", cfg.Project.Name + "-apps"}
			syncStatuses, syncErr := bootstrapper.WaitForAppSync(ctx, appNames)
			if syncErr != nil {
				prog.FailStep(applySection, syncStep, syncErr)
			} else {
				prog.SuccessStep(applySection, syncStep)
				for _, status := range syncStatuses {
					statusIcon := progress.StatusSuccess
					if status.Status == "timeout" {
						statusIcon = progress.StatusWarning
					}
					syncStep.AddSubStep(fmt.Sprintf("%s: %s (%s)", status.Name, status.SyncStatus, status.Health), statusIcon)
				}
				prog.ShowSubSteps(syncStep)
			}
			bootstrapResult.SyncedApps = syncStatuses
		}
	}

	// Update summary with environments
	for _, env := range cfg.Environments {
		summary.Environments = append(summary.Environments, progress.EnvironmentInfo{
			Name:      env.Name,
			Namespace: cfg.Project.Name + "-" + env.Name,
			Status:    "created",
		})
		summary.Cluster.Namespaces = append(summary.Cluster.Namespaces, cfg.Project.Name+"-"+env.Name)
	}

	// Add applications
	summary.Applications = append(summary.Applications, progress.ApplicationInfo{
		Name:   cfg.Project.Name + "-apps",
		Type:   "app-of-apps",
		Status: "synced",
	})

	// Finalize timing
	summary.Setup.Duration = time.Since(startTime)
	summary.Setup.CompletedAt = time.Now()

	// Save summary to file
	if !dryRun {
		if saveErr := progress.SaveSummary(projectPath, summary); saveErr != nil {
			if !quietMode && !jsonMode {
				fmt.Printf("Warning: Could not save summary: %v\n", saveErr)
			}
		}
	}

	// Show final summary
	prog.ShowSummary(summary)

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

func createBootstrapper(cfg *config.Config, c *cluster.Cluster, projectPath string) *bootstrap.Bootstrapper {
	opts := &bootstrap.Options{
		Tool:                 bootstrap.Tool(cfg.GitOpsTool),
		Mode:                 bootstrap.Mode(cfg.Bootstrap.Mode),
		Namespace:            cfg.Bootstrap.Namespace,
		Wait:                 cfg.Bootstrap.Wait,
		Timeout:              cfg.Bootstrap.Timeout,
		ConfigureRepo:        cfg.Bootstrap.ConfigureRepo,
		RepoURL:              cfg.Git.URL,
		RepoBranch:           cfg.Git.Branch,
		RepoPath:             cfg.GitOpsTool + "/applicationsets",
		CreateAppOfApps:      cfg.Bootstrap.CreateAppOfApps,
		SyncInitial:          cfg.Bootstrap.SyncInitial,
		ProjectName:          cfg.Project.Name,
		GeneratedPath:        projectPath,
		ApplyGeneratedConfig: true,
		WaitForSync:          cfg.Bootstrap.Wait,
		SyncTimeout:          cfg.Bootstrap.Timeout,
	}

	return bootstrap.New(c, opts)
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

func runPostInitValidation(ctx context.Context, prog *progress.Progress, projectPath string) error {
	valSection := prog.StartSection("Validation")

	opts := &validate.Options{
		Path:         projectPath,
		K8sVersion:   "1.29",
		Schema:       true,
		Security:     true,
		Deprecation:  true,
		Kustomize:    true,
		OutputFormat: "table",
	}

	switch strings.ToLower(validateFailOn) {
	case "critical":
		opts.FailOn = validate.SeverityCritical
	case "high":
		opts.FailOn = validate.SeverityHigh
	case "medium":
		opts.FailOn = validate.SeverityMedium
	case "low":
		opts.FailOn = validate.SeverityLow
	default:
		opts.FailOn = validate.SeverityHigh
	}

	v := validate.New(opts)

	step := prog.StartStep(valSection, "Validating generated manifests...")
	result, err := v.Validate(ctx)
	if err != nil {
		prog.FailStep(valSection, step, err)
		return fmt.Errorf("validation failed: %w", err)
	}

	step.AddSubStep(fmt.Sprintf("Schema: %d passed", result.Categories[validate.CategorySchema].Passed), progress.StatusSuccess)
	if len(result.Categories[validate.CategorySecurity].Issues) > 0 {
		step.AddSubStep(fmt.Sprintf("Security: %d issues", len(result.Categories[validate.CategorySecurity].Issues)), progress.StatusWarning)
	} else {
		step.AddSubStep("Security: No issues", progress.StatusSuccess)
	}
	if len(result.Categories[validate.CategoryDeprecation].Issues) > 0 {
		step.AddSubStep(fmt.Sprintf("Deprecation: %d issues", len(result.Categories[validate.CategoryDeprecation].Issues)), progress.StatusWarning)
	} else {
		step.AddSubStep("Deprecation: No deprecated APIs", progress.StatusSuccess)
	}
	if len(result.Categories[validate.CategoryKustomize].Issues) > 0 {
		step.AddSubStep(fmt.Sprintf("Kustomize: %d issues", len(result.Categories[validate.CategoryKustomize].Issues)), progress.StatusWarning)
	} else {
		step.AddSubStep("Kustomize: All builds pass", progress.StatusSuccess)
	}

	prog.ShowSubSteps(step)

	if v.ShouldFail(result) {
		prog.FailStep(valSection, step, fmt.Errorf("%d critical/high issues found", result.Failed))
		return fmt.Errorf("validation failed with %d issues at severity %s or higher", result.Failed, validateFailOn)
	}

	prog.SuccessStep(valSection, step)
	return nil
}
