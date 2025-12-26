package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/organization"
)

var (
	orgDomain       string
	orgTeamName     string
	orgProjectPath  string
	teamOwners      string
	teamCostCenter  string
	teamAdmin       bool
	quotaCPU        string
	quotaMemory     string
	quotaStorage    string
	quotaNamespaces int
	projectEnvs     string
	projectRepo     string
	projectPCI      bool
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organization",
	Long: `Manage organization structure including teams and projects.

Examples:
  # Initialize an organization
  gitopsi org init acme-corp --domain acme.com

  # Show organization details
  gitopsi org show`,
}

var orgInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new organization",
	Long: `Initialize a new organization with the given name and optional domain.

Example:
  gitopsi org init acme-corp --domain acme.com`,
	Args: cobra.ExactArgs(1),
	RunE: runOrgInit,
}

var orgShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show organization details",
	RunE:  runOrgShow,
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage teams",
	Long: `Manage teams within the organization.

Examples:
  # Create a new team
  gitopsi team create frontend --owners frontend-leads@acme.com

  # List all teams
  gitopsi team list

  # Set team quotas
  gitopsi team set-quota frontend --cpu 50 --memory 100Gi`,
}

var teamCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new team",
	Long: `Create a new team within the organization.

Example:
  gitopsi team create frontend --owners frontend@acme.com --cost-center CC-1234`,
	Args: cobra.ExactArgs(1),
	RunE: runTeamCreate,
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE:  runTeamList,
}

var teamShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show team details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTeamShow,
}

var teamDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a team",
	Args:  cobra.ExactArgs(1),
	RunE:  runTeamDelete,
}

var teamSetQuotaCmd = &cobra.Command{
	Use:   "set-quota [team]",
	Short: "Set team quotas",
	Long: `Set resource quotas for a team.

Example:
  gitopsi team set-quota frontend --cpu 50 --memory 100Gi --namespaces 10`,
	Args: cobra.ExactArgs(1),
	RunE: runTeamSetQuota,
}

var teamAddClusterCmd = &cobra.Command{
	Use:   "add-cluster [team] [cluster]",
	Short: "Add a cluster to team's allowed list",
	Args:  cobra.ExactArgs(2),
	RunE:  runTeamAddCluster,
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long: `Manage projects within teams.

Examples:
  # Create a new project
  gitopsi project create web-app --team frontend

  # List projects in a team
  gitopsi project list --team frontend

  # Add environment to project
  gitopsi project add-env web-app staging --team frontend`,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new project",
	Long: `Create a new project within a team.

Example:
  gitopsi project create web-app --team frontend --envs dev,staging,prod`,
	Args: cobra.ExactArgs(1),
	RunE: runProjectCreate,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in a team",
	RunE:  runProjectList,
}

var projectShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectShow,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDelete,
}

var projectAddEnvCmd = &cobra.Command{
	Use:   "add-env [project] [environment]",
	Short: "Add an environment to a project",
	Args:  cobra.ExactArgs(2),
	RunE:  runProjectAddEnv,
}

func init() {
	rootCmd.AddCommand(orgCmd)
	rootCmd.AddCommand(teamCmd)
	rootCmd.AddCommand(projectCmd)

	// Org commands
	orgCmd.AddCommand(orgInitCmd)
	orgCmd.AddCommand(orgShowCmd)

	orgInitCmd.Flags().StringVar(&orgDomain, "domain", "", "Organization domain")

	// Team commands
	teamCmd.AddCommand(teamCreateCmd)
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamShowCmd)
	teamCmd.AddCommand(teamDeleteCmd)
	teamCmd.AddCommand(teamSetQuotaCmd)
	teamCmd.AddCommand(teamAddClusterCmd)

	teamCreateCmd.Flags().StringVar(&teamOwners, "owners", "", "Team owners (comma-separated emails)")
	teamCreateCmd.Flags().StringVar(&teamCostCenter, "cost-center", "", "Cost center for billing")
	teamCreateCmd.Flags().BoolVar(&teamAdmin, "admin", false, "Is this an admin team")
	teamCreateCmd.Flags().StringVar(&quotaCPU, "quota-cpu", "", "CPU quota")
	teamCreateCmd.Flags().StringVar(&quotaMemory, "quota-memory", "", "Memory quota")

	teamSetQuotaCmd.Flags().StringVar(&quotaCPU, "cpu", "", "CPU quota")
	teamSetQuotaCmd.Flags().StringVar(&quotaMemory, "memory", "", "Memory quota")
	teamSetQuotaCmd.Flags().StringVar(&quotaStorage, "storage", "", "Storage quota")
	teamSetQuotaCmd.Flags().IntVar(&quotaNamespaces, "namespaces", 0, "Maximum namespaces")

	// Project commands
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectAddEnvCmd)

	projectCreateCmd.Flags().StringVar(&orgTeamName, "team", "", "Team name (required)")
	projectCreateCmd.Flags().StringVar(&projectEnvs, "envs", "", "Environments (comma-separated)")
	projectCreateCmd.Flags().StringVar(&projectRepo, "repo", "", "Source repository URL")
	projectCreateCmd.Flags().BoolVar(&projectPCI, "pci", false, "PCI compliant project")
	_ = projectCreateCmd.MarkFlagRequired("team")

	projectListCmd.Flags().StringVar(&orgTeamName, "team", "", "Team name (required)")
	_ = projectListCmd.MarkFlagRequired("team")

	projectShowCmd.Flags().StringVar(&orgTeamName, "team", "", "Team name (required)")
	_ = projectShowCmd.MarkFlagRequired("team")

	projectDeleteCmd.Flags().StringVar(&orgTeamName, "team", "", "Team name (required)")
	_ = projectDeleteCmd.MarkFlagRequired("team")

	projectAddEnvCmd.Flags().StringVar(&orgTeamName, "team", "", "Team name (required)")
	_ = projectAddEnvCmd.MarkFlagRequired("team")
}

func getOrgManager() (*organization.Manager, error) {
	configPath := "organization.yaml"
	if orgProjectPath != "" {
		configPath = orgProjectPath + "/organization.yaml"
	}
	return organization.NewManager(configPath)
}

func runOrgInit(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	org, err := manager.InitOrganization(ctx, name, orgDomain)
	if err != nil {
		return fmt.Errorf("failed to initialize organization: %w", err)
	}

	pterm.Success.Printf("Organization '%s' initialized successfully\n", org.Name)
	if org.Domain != "" {
		pterm.Info.Printf("Domain: %s\n", org.Domain)
	}

	return nil
}

func runOrgShow(_ *cobra.Command, _ []string) error {
	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	org := manager.GetOrganization()
	if org == nil {
		return fmt.Errorf("no organization initialized")
	}

	pterm.DefaultSection.Println("Organization")
	tableData := [][]string{
		{"Name", org.Name},
		{"Domain", org.Domain},
		{"Teams", fmt.Sprintf("%d", len(org.Teams))},
	}

	// Count total projects
	totalProjects := 0
	for _, t := range org.Teams {
		totalProjects += len(t.Projects)
	}
	tableData = append(tableData, []string{"Projects", fmt.Sprintf("%d", totalProjects)})

	_ = pterm.DefaultTable.WithData(tableData).Render()

	// Show policies
	pterm.DefaultSection.Println("Policies")
	policyData := [][]string{
		{"Default CPU Quota", org.Policies.ResourceQuotas.CPU},
		{"Default Memory Quota", org.Policies.ResourceQuotas.Memory},
		{"Pod Security Level", org.Policies.PodSecurity.Level},
	}
	_ = pterm.DefaultTable.WithData(policyData).Render()

	return nil
}

func runTeamCreate(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	var owners []string
	if teamOwners != "" {
		owners = strings.Split(teamOwners, ",")
	}

	opts := &organization.TeamOptions{
		Name:       name,
		Owners:     owners,
		IsAdmin:    teamAdmin,
		CostCenter: teamCostCenter,
		Quotas: organization.TeamQuotas{
			CPU:    quotaCPU,
			Memory: quotaMemory,
		},
	}

	team, err := manager.CreateTeam(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	pterm.Success.Printf("Team '%s' created successfully\n", team.Name)

	return nil
}

func runTeamList(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	teams, err := manager.ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	if len(teams) == 0 {
		pterm.Info.Println("No teams found")
		return nil
	}

	tableData := [][]string{
		{"NAME", "OWNERS", "PROJECTS", "QUOTAS (CPU/MEM)", "ADMIN"},
	}

	for _, t := range teams {
		owners := strings.Join(t.Owners, ", ")
		if len(owners) > 30 {
			owners = owners[:27] + "..."
		}
		quotas := fmt.Sprintf("%s / %s", t.Quotas.CPU, t.Quotas.Memory)
		admin := "No"
		if t.IsAdmin {
			admin = "Yes"
		}
		tableData = append(tableData, []string{
			t.Name,
			owners,
			fmt.Sprintf("%d", len(t.Projects)),
			quotas,
			admin,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

func runTeamShow(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	team, err := manager.GetTeam(ctx, name)
	if err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	pterm.DefaultSection.Println("Team: " + team.Name)

	tableData := [][]string{
		{"Description", team.Description},
		{"Owners", strings.Join(team.Owners, ", ")},
		{"Cost Center", team.CostCenter},
		{"Admin", fmt.Sprintf("%v", team.IsAdmin)},
		{"Projects", fmt.Sprintf("%d", len(team.Projects))},
	}
	_ = pterm.DefaultTable.WithData(tableData).Render()

	pterm.DefaultSection.Println("Quotas")
	quotaData := [][]string{
		{"CPU", team.Quotas.CPU},
		{"Memory", team.Quotas.Memory},
		{"Storage", team.Quotas.Storage},
		{"Namespaces", fmt.Sprintf("%d", team.Quotas.Namespaces)},
	}
	_ = pterm.DefaultTable.WithData(quotaData).Render()

	if len(team.AllowedClusters) > 0 {
		pterm.DefaultSection.Println("Allowed Clusters")
		for _, c := range team.AllowedClusters {
			fmt.Printf("  - %s\n", c)
		}
	}

	if len(team.Projects) > 0 {
		pterm.DefaultSection.Println("Projects")
		for _, p := range team.Projects {
			fmt.Printf("  - %s (%s)\n", p.Name, strings.Join(p.Environments, ", "))
		}
	}

	return nil
}

func runTeamDelete(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	if err := manager.DeleteTeam(ctx, name); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	pterm.Success.Printf("Team '%s' deleted successfully\n", name)
	return nil
}

func runTeamSetQuota(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	quotas := organization.TeamQuotas{
		CPU:        quotaCPU,
		Memory:     quotaMemory,
		Storage:    quotaStorage,
		Namespaces: quotaNamespaces,
	}

	if err := manager.SetTeamQuota(ctx, name, quotas); err != nil {
		return fmt.Errorf("failed to set quota: %w", err)
	}

	pterm.Success.Printf("Quota updated for team '%s'\n", name)
	return nil
}

func runTeamAddCluster(_ *cobra.Command, args []string) error {
	teamName := args[0]
	clusterName := args[1]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	if err := manager.AddClusterToTeam(ctx, teamName, clusterName); err != nil {
		return fmt.Errorf("failed to add cluster: %w", err)
	}

	pterm.Success.Printf("Cluster '%s' added to team '%s'\n", clusterName, teamName)
	return nil
}

func runProjectCreate(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	var envs []string
	if projectEnvs != "" {
		envs = strings.Split(projectEnvs, ",")
	}

	opts := &organization.ProjectOptions{
		Name:         name,
		Environments: envs,
		SourceRepo:   projectRepo,
		PCICompliant: projectPCI,
	}

	project, err := manager.CreateProject(ctx, orgTeamName, opts)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	pterm.Success.Printf("Project '%s' created in team '%s'\n", project.Name, orgTeamName)
	pterm.Info.Printf("Environments: %s\n", strings.Join(project.Environments, ", "))

	return nil
}

func runProjectList(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	projects, err := manager.ListProjects(ctx, orgTeamName)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		pterm.Info.Printf("No projects found in team '%s'\n", orgTeamName)
		return nil
	}

	tableData := [][]string{
		{"NAME", "ENVIRONMENTS", "REPO", "PCI"},
	}

	for _, p := range projects {
		pci := "No"
		if p.PCICompliant {
			pci = "Yes"
		}
		repo := p.SourceRepo
		if len(repo) > 40 {
			repo = repo[:37] + "..."
		}
		tableData = append(tableData, []string{
			p.Name,
			strings.Join(p.Environments, ", "),
			repo,
			pci,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

func runProjectShow(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	project, err := manager.GetProject(ctx, orgTeamName, name)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	pterm.DefaultSection.Println("Project: " + project.Name)

	tableData := [][]string{
		{"Team", orgTeamName},
		{"Description", project.Description},
		{"Environments", strings.Join(project.Environments, ", ")},
		{"Source Repo", project.SourceRepo},
		{"PCI Compliant", fmt.Sprintf("%v", project.PCICompliant)},
	}
	_ = pterm.DefaultTable.WithData(tableData).Render()

	if len(project.Owners) > 0 {
		pterm.DefaultSection.Println("Owners")
		for _, o := range project.Owners {
			fmt.Printf("  - %s\n", o)
		}
	}

	return nil
}

func runProjectDelete(_ *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	if err := manager.DeleteProject(ctx, orgTeamName, name); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	pterm.Success.Printf("Project '%s' deleted from team '%s'\n", name, orgTeamName)
	return nil
}

func runProjectAddEnv(_ *cobra.Command, args []string) error {
	projectName := args[0]
	envName := args[1]
	ctx := context.Background()

	manager, err := getOrgManager()
	if err != nil {
		return err
	}

	if err := manager.AddEnvironmentToProject(ctx, orgTeamName, projectName, envName); err != nil {
		return fmt.Errorf("failed to add environment: %w", err)
	}

	pterm.Success.Printf("Environment '%s' added to project '%s'\n", envName, projectName)
	return nil
}
