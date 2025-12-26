package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ihsanmokhlisse/gitopsi/internal/operator"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var operatorProjectPath string

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Manage Kubernetes operators in your GitOps configuration",
	Long: `Manage Kubernetes operators in your GitOps configuration.

This command provides subcommands to add, remove, and list operators 
that will be deployed via OLM (Operator Lifecycle Manager).`,
}

var operatorAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add an operator to the configuration",
	Long: `Add an operator to the GitOps configuration.

You can either add a preset operator by name or specify custom operator details.`,
	Args: cobra.ExactArgs(1),
	RunE: runOperatorAdd,
}

var operatorAddPresetCmd = &cobra.Command{
	Use:   "add-preset [preset]",
	Short: "Add a preset operator configuration",
	Long: `Add a preset operator from the available presets.

Use 'gitopsi operator presets' to see available presets.`,
	Args: cobra.ExactArgs(1),
	RunE: runOperatorAddPreset,
}

var operatorRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an operator from the configuration",
	Long:  `Remove an operator from the GitOps configuration by its name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runOperatorRemove,
}

var operatorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured operators",
	Long:  `List all operators currently configured in the GitOps configuration.`,
	RunE:  runOperatorList,
}

var operatorPresetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available operator presets",
	Long:  `List all available operator presets that can be added with 'add-preset'.`,
	RunE:  runOperatorPresets,
}

var (
	opNamespace         string
	opChannel           string
	opSource            string
	opSourceNamespace   string
	opInstallMode       string
	opInstallApproval   string
)

func init() {
	operatorCmd.AddCommand(operatorAddCmd)
	operatorCmd.AddCommand(operatorAddPresetCmd)
	operatorCmd.AddCommand(operatorRemoveCmd)
	operatorCmd.AddCommand(operatorListCmd)
	operatorCmd.AddCommand(operatorPresetsCmd)

	operatorCmd.PersistentFlags().StringVar(&operatorProjectPath, "project", ".", "Path to the gitopsi project")

	operatorAddCmd.Flags().StringVar(&opNamespace, "namespace", "", "Operator namespace (required)")
	operatorAddCmd.Flags().StringVar(&opChannel, "channel", "stable", "OLM channel")
	operatorAddCmd.Flags().StringVar(&opSource, "source", "community-operators", "CatalogSource name")
	operatorAddCmd.Flags().StringVar(&opSourceNamespace, "source-namespace", "openshift-marketplace", "CatalogSource namespace")
	operatorAddCmd.Flags().StringVar(&opInstallMode, "install-mode", "OwnNamespace", "Install mode: OwnNamespace, SingleNamespace, MultiNamespace, AllNamespaces")
	operatorAddCmd.Flags().StringVar(&opInstallApproval, "approval", "Automatic", "Install plan approval: Automatic or Manual")
}

func getOperatorManager() (*operator.Manager, error) {
	configPath := filepath.Join(operatorProjectPath, "gitopsi.yaml")

	cfg := operator.NewDefaultConfig()
	cfg.Enabled = true

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}

		var fullConfig struct {
			Operators operator.Config `yaml:"operators"`
		}
		if err := yaml.Unmarshal(data, &fullConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
		if fullConfig.Operators.Enabled {
			cfg = fullConfig.Operators
		}
	}

	return operator.NewManager(cfg), nil
}

func saveOperatorConfig(mgr *operator.Manager) error {
	configPath := filepath.Join(operatorProjectPath, "gitopsi.yaml")
	
	var fullConfig map[string]interface{}
	
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &fullConfig); err != nil {
			fullConfig = make(map[string]interface{})
		}
	} else {
		fullConfig = make(map[string]interface{})
	}

	fullConfig["operators"] = mgr.GetConfig()

	data, err := yaml.Marshal(fullConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func runOperatorAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	if opNamespace == "" {
		return fmt.Errorf("--namespace is required")
	}

	mgr, err := getOperatorManager()
	if err != nil {
		return err
	}

	op := &operator.Operator{
		Name:                name,
		Namespace:           opNamespace,
		Channel:             opChannel,
		Source:              opSource,
		SourceNamespace:     opSourceNamespace,
		InstallMode:         opInstallMode,
		InstallPlanApproval: opInstallApproval,
		Enabled:             true,
	}

	if err := mgr.AddOperator(op); err != nil {
		return fmt.Errorf("failed to add operator: %w", err)
	}

	if err := saveOperatorConfig(mgr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	pterm.Success.Printf("Added operator: %s\n", name)
	return nil
}

func runOperatorAddPreset(cmd *cobra.Command, args []string) error {
	presetName := args[0]

	mgr, err := getOperatorManager()
	if err != nil {
		return err
	}

	if err := mgr.AddPreset(presetName); err != nil {
		return fmt.Errorf("failed to add preset: %w", err)
	}

	if err := saveOperatorConfig(mgr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	preset, _ := operator.GetOperatorPreset(presetName)
	pterm.Success.Printf("Added preset operator: %s (%s)\n", presetName, preset.Name)
	return nil
}

func runOperatorRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := getOperatorManager()
	if err != nil {
		return err
	}

	if !mgr.RemoveOperator(name) {
		return fmt.Errorf("operator not found: %s", name)
	}

	if err := saveOperatorConfig(mgr); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	pterm.Success.Printf("Removed operator: %s\n", name)
	return nil
}

func runOperatorList(cmd *cobra.Command, args []string) error {
	mgr, err := getOperatorManager()
	if err != nil {
		return err
	}

	operators := mgr.GetEnabledOperators()
	if len(operators) == 0 {
		pterm.Info.Println("No operators configured")
		return nil
	}

	pterm.DefaultHeader.Println("Configured Operators")

	tableData := pterm.TableData{
		{"Name", "Namespace", "Channel", "Source", "Install Mode", "Enabled"},
	}

	for _, op := range operators {
		enabled := "No"
		if op.Enabled {
			enabled = "Yes"
		}
		tableData = append(tableData, []string{
			op.Name,
			op.Namespace,
			op.GetChannel(),
			op.GetSource("community-operators"),
			string(op.GetInstallMode()),
			enabled,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

func runOperatorPresets(cmd *cobra.Command, args []string) error {
	presets := operator.ListOperatorPresets()
	sort.Strings(presets)

	pterm.DefaultHeader.Println("Available Operator Presets")

	tableData := pterm.TableData{
		{"Preset Name", "Operator", "Namespace", "Channel", "Source"},
	}

	for _, name := range presets {
		preset, _ := operator.GetOperatorPreset(name)
		tableData = append(tableData, []string{
			name,
			preset.Name,
			preset.Namespace,
			preset.Channel,
			preset.Source,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

