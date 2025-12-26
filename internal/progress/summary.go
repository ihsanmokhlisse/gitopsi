package progress

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

// SetupSummary contains all setup information.
type SetupSummary struct {
	Setup        SetupInfo         `yaml:"setup"`
	Git          GitInfo           `yaml:"git"`
	Cluster      ClusterInfo       `yaml:"cluster"`
	GitOpsTool   GitOpsToolInfo    `yaml:"gitops_tool"`
	Environments []EnvironmentInfo `yaml:"environments"`
	Applications []ApplicationInfo `yaml:"applications"`
}

// SetupInfo contains setup metadata.
type SetupInfo struct {
	CompletedAt time.Time     `yaml:"completed_at"`
	Duration    time.Duration `yaml:"duration"`
	Version     string        `yaml:"version"`
}

// GitInfo contains Git repository information.
type GitInfo struct {
	URL      string `yaml:"url"`
	Branch   string `yaml:"branch"`
	WebURL   string `yaml:"web_url"`
	Provider string `yaml:"provider"`
	Status   string `yaml:"status"`
}

// ClusterInfo contains cluster information.
type ClusterInfo struct {
	Name       string   `yaml:"name"`
	URL        string   `yaml:"url"`
	Platform   string   `yaml:"platform"`
	Version    string   `yaml:"version"`
	Status     string   `yaml:"status"`
	Namespaces []string `yaml:"namespaces"`
}

// GitOpsToolInfo contains GitOps tool information.
type GitOpsToolInfo struct {
	Name           string `yaml:"name"`
	URL            string `yaml:"url"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password,omitempty"`
	PasswordSecret string `yaml:"password_secret,omitempty"`
	Namespace      string `yaml:"namespace"`
	Version        string `yaml:"version"`
	Status         string `yaml:"status"`
	PodCount       string `yaml:"pod_count"`
}

// EnvironmentInfo contains environment information.
type EnvironmentInfo struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Status    string `yaml:"status"`
}

// ApplicationInfo contains application information.
type ApplicationInfo struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Status   string   `yaml:"status"`
	Children []string `yaml:"children,omitempty"`
}

// ShowSummary displays the complete setup summary.
func (p *Progress) ShowSummary(summary *SetupSummary) {
	if p.jsonOutput {
		data, err := yaml.Marshal(summary)
		if err == nil {
			fmt.Println(string(data))
		}
		return
	}

	if p.quiet {
		fmt.Printf("✅ Setup complete! ArgoCD: %s\n", summary.GitOpsTool.URL)
		return
	}

	fmt.Println()
	pterm.DefaultBox.WithTitle(pterm.Bold.Sprint("✅ GitOps Setup Complete!")).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgGreen)).
		Println(fmt.Sprintf("Duration: %s", formatDuration(summary.Setup.Duration)))

	// Git Repository section
	p.showGitSection(&summary.Git)

	// GitOps Tool section
	p.showGitOpsSection(&summary.GitOpsTool)

	// Cluster section
	p.showClusterSection(&summary.Cluster)

	// Applications section
	if len(summary.Applications) > 0 {
		p.showApplicationsSection(summary.Applications)
	}

	// Quick Commands section
	p.showQuickCommands(summary)

	// Documentation section
	p.showDocumentation(summary)
}

func (p *Progress) showGitSection(git *GitInfo) {
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Git Repository")

	data := pterm.TableData{
		{"URL:", git.URL},
		{"Branch:", git.Branch},
		{"Provider:", git.Provider},
		{"Status:", formatStatus(git.Status)},
	}
	if git.WebURL != "" {
		data = append(data, []string{"Web:", git.WebURL})
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()
}

func (p *Progress) showGitOpsSection(tool *GitOpsToolInfo) {
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println(tool.Name)

	passwordDisplay := tool.Password
	if passwordDisplay == "" {
		passwordDisplay = fmt.Sprintf("Run: gitopsi get-password %s", tool.Name)
	}

	data := pterm.TableData{
		{"URL:", tool.URL},
		{"Username:", tool.Username},
		{"Password:", passwordDisplay},
		{"Namespace:", tool.Namespace},
		{"Version:", tool.Version},
		{"Status:", formatStatus(tool.Status)},
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()

	fmt.Println()
	pterm.Info.Println("CLI Login:")
	if tool.Password != "" {
		pterm.Printf("  %s login %s --username %s --password %s\n",
			tool.Name, tool.URL, tool.Username, tool.Password)
	} else {
		pterm.Printf("  %s login %s --username %s --password <password>\n",
			tool.Name, tool.URL, tool.Username)
	}
}

func (p *Progress) showClusterSection(cluster *ClusterInfo) {
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Cluster")

	data := pterm.TableData{
		{"Name:", cluster.Name},
		{"URL:", cluster.URL},
		{"Platform:", cluster.Platform},
		{"Version:", cluster.Version},
		{"Status:", formatStatus(cluster.Status)},
	}

	_ = pterm.DefaultTable.WithData(data).WithLeftAlignment().Render()

	if len(cluster.Namespaces) > 0 {
		fmt.Println()
		pterm.Info.Println("Namespaces Created:")
		for _, ns := range cluster.Namespaces {
			pterm.Printf("  %s %s\n", pterm.Green("✓"), ns)
		}
	}
}

func (p *Progress) showApplicationsSection(apps []ApplicationInfo) {
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Applications")

	for _, app := range apps {
		icon := getStatusIcon(StepStatus(app.Status))
		pterm.Printf("%s %s (%s)\n", icon, app.Name, app.Type)
		for _, child := range app.Children {
			pterm.Printf("   %s %s\n", pterm.Green("✓"), child)
		}
	}
}

func (p *Progress) showQuickCommands(summary *SetupSummary) {
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Quick Commands")

	commands := [][]string{
		{"Open UI:", fmt.Sprintf("gitopsi open %s", summary.GitOpsTool.Name)},
		{"Get Password:", fmt.Sprintf("gitopsi get-password %s", summary.GitOpsTool.Name)},
		{"Check Status:", "gitopsi status"},
		{"Add Application:", "gitopsi add app <name> --image <image>"},
	}

	for _, cmd := range commands {
		pterm.Printf("  %s\n    %s\n", pterm.Bold.Sprint(cmd[0]), pterm.Cyan(cmd[1]))
	}
}

func (p *Progress) showDocumentation(summary *SetupSummary) {
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Documentation")

	pterm.Info.Println("Generated docs available at:")
	pterm.Printf("  README:       %s/README.md\n", p.projectName)
	pterm.Printf("  Architecture: %s/docs/ARCHITECTURE.md\n", p.projectName)
	pterm.Printf("  Onboarding:   %s/docs/ONBOARDING.md\n", p.projectName)

	fmt.Println()
	pterm.Success.Printf("Summary saved to: %s/.gitopsi/setup-summary.yaml\n", p.projectName)
}

func formatStatus(status string) string {
	switch status {
	case "connected", "synced", "healthy", "ready":
		return pterm.Green("✓ " + status)
	case "warning", "degraded":
		return pterm.Yellow("⚠ " + status)
	case "failed", "error", "disconnected":
		return pterm.Red("✗ " + status)
	default:
		return status
	}
}

// SaveSummary saves the setup summary to a file.
func SaveSummary(projectPath string, summary *SetupSummary) error {
	gitopsiDir := filepath.Join(projectPath, ".gitopsi")
	if err := os.MkdirAll(gitopsiDir, 0750); err != nil {
		return fmt.Errorf("failed to create .gitopsi directory: %w", err)
	}

	summaryPath := filepath.Join(gitopsiDir, "setup-summary.yaml")
	data, err := yaml.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := os.WriteFile(summaryPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}

// LoadSummary loads the setup summary from a file.
func LoadSummary(projectPath string) (*SetupSummary, error) {
	summaryPath := filepath.Join(projectPath, ".gitopsi", "setup-summary.yaml")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read summary file: %w", err)
	}

	var summary SetupSummary
	if err := yaml.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
	}

	return &summary, nil
}
