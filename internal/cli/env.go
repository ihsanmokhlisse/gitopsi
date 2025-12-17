package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/environment"
)

var (
	envTopology    string
	envNamespace   string
	envClusterURL  string
	envClusterName string
	envRegion      string
	envPrimary     bool
	envProjectPath string
	envFromEnv     string
	envToEnv       string
	envPromoteAll  bool
	envPromoteApp  string
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments and clusters",
	Long: `Manage GitOps environments and their cluster mappings.

Supports multiple deployment topologies:
  - namespace-based: All environments in one cluster, isolated by namespaces
  - cluster-per-env: Each environment has its own dedicated cluster
  - multi-cluster: Multiple clusters per environment (HA/multi-region)

Examples:
  gitopsi env create dev,staging,prod                    # Create namespaced environments
  gitopsi env create prod --cluster https://prod.k8s    # Create with dedicated cluster
  gitopsi env list                                       # List all environments
  gitopsi env show prod                                  # Show environment details
  gitopsi env add-cluster prod --url https://eu.k8s    # Add cluster to environment
  gitopsi promote myapp --from dev --to staging         # Promote application`,
}

var envCreateCmd = &cobra.Command{
	Use:   "create [environments]",
	Short: "Create one or more environments",
	Long: `Create environments with optional cluster mappings.

For namespace-based topology (default):
  gitopsi env create dev,staging,prod

For cluster-per-environment:
  gitopsi env create dev --cluster https://dev.k8s
  gitopsi env create prod --cluster https://prod.k8s

For multi-cluster environment:
  gitopsi env create prod --cluster https://us.k8s --cluster https://eu.k8s`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEnvCreate,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	RunE:  runEnvList,
}

var envShowCmd = &cobra.Command{
	Use:   "show [environment]",
	Short: "Show environment details",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvShow,
}

var envDeleteCmd = &cobra.Command{
	Use:   "delete [environment]",
	Short: "Delete an environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvDelete,
}

var envAddClusterCmd = &cobra.Command{
	Use:   "add-cluster [environment]",
	Short: "Add a cluster to an environment",
	Long: `Add a cluster to an environment for multi-cluster deployments.

Examples:
  gitopsi env add-cluster prod --url https://eu-west.k8s --name eu-west --region eu-west-1
  gitopsi env add-cluster prod --url https://us-east.k8s --name us-east --primary`,
	Args: cobra.ExactArgs(1),
	RunE: runEnvAddCluster,
}

var envRemoveClusterCmd = &cobra.Command{
	Use:   "remove-cluster [environment]",
	Short: "Remove a cluster from an environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvRemoveCluster,
}

var promoteCmd = &cobra.Command{
	Use:   "promote [application]",
	Short: "Promote application between environments",
	Long: `Promote an application from one environment to another.

Examples:
  gitopsi promote myapp --from dev --to staging
  gitopsi promote --all --from staging --to prod`,
	RunE: runPromote,
}

func init() {
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(promoteCmd)

	envCmd.AddCommand(envCreateCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envShowCmd)
	envCmd.AddCommand(envDeleteCmd)
	envCmd.AddCommand(envAddClusterCmd)
	envCmd.AddCommand(envRemoveClusterCmd)

	envCmd.PersistentFlags().StringVar(&envProjectPath, "project", ".", "Path to gitopsi project")
	envCmd.PersistentFlags().StringVar(&envTopology, "topology", "", "Environment topology: namespace-based, cluster-per-env, multi-cluster")

	envCreateCmd.Flags().StringVar(&envNamespace, "namespace", "", "Custom namespace for the environment")
	envCreateCmd.Flags().StringVar(&envClusterURL, "cluster", "", "Cluster URL for the environment")
	envCreateCmd.Flags().StringVar(&envClusterName, "cluster-name", "", "Name for the cluster")
	envCreateCmd.Flags().StringVar(&envRegion, "region", "", "Cluster region")
	envCreateCmd.Flags().BoolVar(&envPrimary, "primary", false, "Mark cluster as primary")

	envAddClusterCmd.Flags().StringVar(&envClusterURL, "url", "", "Cluster URL (required)")
	envAddClusterCmd.Flags().StringVar(&envClusterName, "name", "", "Cluster name (required)")
	envAddClusterCmd.Flags().StringVar(&envNamespace, "namespace", "", "Default namespace on the cluster")
	envAddClusterCmd.Flags().StringVar(&envRegion, "region", "", "Cluster region")
	envAddClusterCmd.Flags().BoolVar(&envPrimary, "primary", false, "Mark cluster as primary")
	_ = envAddClusterCmd.MarkFlagRequired("url")
	_ = envAddClusterCmd.MarkFlagRequired("name")

	envRemoveClusterCmd.Flags().StringVar(&envClusterName, "name", "", "Cluster name to remove (required)")
	_ = envRemoveClusterCmd.MarkFlagRequired("name")

	promoteCmd.Flags().StringVar(&envFromEnv, "from", "", "Source environment (required)")
	promoteCmd.Flags().StringVar(&envToEnv, "to", "", "Target environment (required)")
	promoteCmd.Flags().BoolVar(&envPromoteAll, "all", false, "Promote all applications")
	promoteCmd.Flags().StringVar(&envProjectPath, "project", ".", "Path to gitopsi project")
	_ = promoteCmd.MarkFlagRequired("from")
	_ = promoteCmd.MarkFlagRequired("to")
}

func getEnvManager() (*environment.Manager, error) {
	absPath, err := filepath.Abs(envProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", absPath)
	}

	mgr := environment.NewManager(absPath)
	if err := mgr.Load(); err != nil {
		return nil, fmt.Errorf("failed to load environment config: %w", err)
	}

	return mgr, nil
}

func runEnvCreate(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	envNames := strings.Split(args[0], ",")

	if envTopology != "" {
		topology, parseErr := environment.ParseTopology(envTopology)
		if parseErr != nil {
			return parseErr
		}
		if setErr := mgr.SetTopology(topology); setErr != nil {
			return setErr
		}
	}

	for _, name := range envNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		opts := environment.CreateEnvOptions{
			Namespace: envNamespace,
			Clusters:  []environment.ClusterInfo{},
		}

		if envClusterURL != "" {
			clusterName := envClusterName
			if clusterName == "" {
				clusterName = name + "-cluster"
			}
			opts.Clusters = append(opts.Clusters, environment.ClusterInfo{
				Name:      clusterName,
				URL:       envClusterURL,
				Namespace: envNamespace,
				Region:    envRegion,
				Primary:   envPrimary,
			})
		}

		if createErr := mgr.CreateEnvironment(name, opts); createErr != nil {
			pterm.Error.Printf("Failed to create environment %s: %v\n", name, createErr)
			continue
		}

		pterm.Success.Printf("Created environment: %s\n", name)
	}

	return nil
}

func runEnvList(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	envs := mgr.ListEnvironments()
	if len(envs) == 0 {
		pterm.Info.Println("No environments configured")
		return nil
	}

	pterm.DefaultSection.Println("Environments")
	pterm.Info.Printf("Topology: %s\n", mgr.Config().Topology)
	pterm.Println()

	tableData := pterm.TableData{
		{"Name", "Namespace", "Clusters", "Primary"},
	}

	for _, env := range envs {
		clusterCount := len(env.Clusters)
		clusterStr := fmt.Sprintf("%d", clusterCount)
		if clusterCount == 0 {
			clusterStr = "-"
		}

		primary := "-"
		if pc := env.GetPrimaryCluster(); pc != nil {
			primary = pc.Name
		}

		namespace := env.Namespace
		if namespace == "" {
			namespace = "-"
		}

		tableData = append(tableData, []string{env.Name, namespace, clusterStr, primary})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	return nil
}

func runEnvShow(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	envName := args[0]
	env := mgr.GetEnvironment(envName)
	if env == nil {
		return fmt.Errorf("environment %s not found", envName)
	}

	pterm.DefaultSection.Printf("Environment: %s\n", env.Name)

	pterm.Info.Printf("Namespace: %s\n", env.GetNamespace(""))
	pterm.Info.Printf("Clusters: %d\n", len(env.Clusters))

	if len(env.Clusters) > 0 {
		pterm.Println()
		pterm.DefaultSection.WithLevel(2).Println("Clusters")

		tableData := pterm.TableData{
			{"Name", "URL", "Region", "Primary"},
		}

		for _, c := range env.Clusters {
			primary := ""
			if c.Primary {
				primary = "✓"
			}
			region := c.Region
			if region == "" {
				region = "-"
			}
			tableData = append(tableData, []string{c.Name, c.URL, region, primary})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	}

	return nil
}

func runEnvDelete(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	envName := args[0]
	if deleteErr := mgr.DeleteEnvironment(envName); deleteErr != nil {
		return deleteErr
	}

	pterm.Success.Printf("Deleted environment: %s\n", envName)
	return nil
}

func runEnvAddCluster(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	envName := args[0]

	cluster := environment.ClusterInfo{
		Name:      envClusterName,
		URL:       envClusterURL,
		Namespace: envNamespace,
		Region:    envRegion,
		Primary:   envPrimary,
	}

	if addErr := mgr.AddClusterToEnvironment(envName, cluster); addErr != nil {
		return addErr
	}

	pterm.Success.Printf("Added cluster %s to environment %s\n", envClusterName, envName)
	return nil
}

func runEnvRemoveCluster(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	envName := args[0]

	if removeErr := mgr.RemoveClusterFromEnvironment(envName, envClusterName); removeErr != nil {
		return removeErr
	}

	pterm.Success.Printf("Removed cluster %s from environment %s\n", envClusterName, envName)
	return nil
}

func runPromote(cmd *cobra.Command, args []string) error {
	mgr, err := getEnvManager()
	if err != nil {
		return err
	}

	appName := ""
	if len(args) > 0 {
		appName = args[0]
	}

	if appName == "" && !envPromoteAll {
		return fmt.Errorf("specify an application name or use --all")
	}

	opts := environment.PromotionOptions{
		Application: appName,
		FromEnv:     envFromEnv,
		ToEnv:       envToEnv,
		All:         envPromoteAll,
		DryRun:      dryRun,
	}

	result, promoteErr := mgr.Promote(opts)
	if promoteErr != nil {
		return promoteErr
	}

	if dryRun {
		pterm.Warning.Println("DRY RUN - No changes made")
	}

	pterm.Success.Println(result.Message)

	if len(result.Changes) > 0 {
		pterm.Println()
		pterm.Info.Println("Changes:")
		for _, change := range result.Changes {
			pterm.Info.Printf("  • %s\n", change)
		}
	}

	return nil
}
