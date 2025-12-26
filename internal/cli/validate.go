package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/validate"
)

var (
	validateK8sVersion    string
	validateArgoCDVersion string
	validateSchema        bool
	validateSecurity      bool
	validateDeprecation   bool
	validateKustomize     bool
	validateAll           bool
	validateCmdFailOn     string
	validateOutputFormat  string
	validateFix           bool
)

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate GitOps manifests",
	Long: `Validate GitOps manifests for schema compliance, security issues,
deprecated APIs, and best practices.

Examples:
  gitopsi validate ./my-platform/                    # Validate all checks
  gitopsi validate ./my-platform/ --security         # Security scan only
  gitopsi validate ./my-platform/ --deprecation      # Deprecated API check only
  gitopsi validate ./my-platform/ --k8s-version 1.29 # Specific K8s version
  gitopsi validate ./my-platform/ --fail-on high     # Fail on high+ severity
  gitopsi validate ./my-platform/ --output json      # JSON output`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVar(&validateK8sVersion, "k8s-version", "1.29", "Target Kubernetes version")
	validateCmd.Flags().StringVar(&validateArgoCDVersion, "argocd-version", "2.10", "Target ArgoCD version")
	validateCmd.Flags().BoolVar(&validateSchema, "schema", false, "Run schema validation only")
	validateCmd.Flags().BoolVar(&validateSecurity, "security", false, "Run security scan only")
	validateCmd.Flags().BoolVar(&validateDeprecation, "deprecation", false, "Run deprecation check only")
	validateCmd.Flags().BoolVar(&validateKustomize, "kustomize", false, "Run kustomize validation only")
	validateCmd.Flags().BoolVar(&validateAll, "all", true, "Run all validations (default)")
	validateCmd.Flags().StringVar(&validateCmdFailOn, "fail-on", "high", "Fail on severity: critical, high, medium, low")
	validateCmd.Flags().StringVar(&validateOutputFormat, "output", "table", "Output format: table, json, yaml")
	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "Auto-fix fixable issues")
}

func runValidate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	opts := &validate.Options{
		Path:          path,
		K8sVersion:    validateK8sVersion,
		ArgoCDVersion: validateArgoCDVersion,
		OutputFormat:  validateOutputFormat,
		Fix:           validateFix,
	}

	if validateSchema || validateSecurity || validateDeprecation || validateKustomize {
		opts.Schema = validateSchema
		opts.Security = validateSecurity
		opts.Deprecation = validateDeprecation
		opts.Kustomize = validateKustomize
	} else {
		opts.Schema = true
		opts.Security = true
		opts.Deprecation = true
		opts.Kustomize = true
	}

	switch strings.ToLower(validateCmdFailOn) {
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

	if validateOutputFormat != "json" && validateOutputFormat != "yaml" {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
			WithTextStyle(pterm.NewStyle(pterm.FgWhite)).
			Println("gitopsi validate")
		pterm.Info.Printf("Validating: %s\n", path)
		pterm.Info.Printf("K8s Version: %s | ArgoCD Version: %s\n", validateK8sVersion, validateArgoCDVersion)
		pterm.Println()
	}

	validator := validate.New(opts)
	result, err := validator.Validate(ctx)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	switch validateOutputFormat {
	case "json":
		jsonOutput, jsonErr := result.ToJSON()
		if jsonErr != nil {
			return jsonErr
		}
		fmt.Println(jsonOutput)
	case "yaml":
		yamlOutput, yamlErr := result.ToYAML()
		if yamlErr != nil {
			return yamlErr
		}
		fmt.Println(yamlOutput)
	default:
		printValidationResult(result)
	}

	if validator.ShouldFail(result) {
		return fmt.Errorf("validation failed with %d issues at severity %s or higher", result.Failed, validateCmdFailOn)
	}

	return nil
}

func printValidationResult(result *validate.ValidationResult) {
	if catResult, ok := result.Categories[validate.CategorySchema]; ok {
		pterm.DefaultSection.Println("📋 Schema Validation")
		if len(catResult.Issues) == 0 {
			pterm.Success.Printf("✅ %d manifests validated against Kubernetes schema\n", catResult.Passed)
		} else {
			pterm.Warning.Printf("⚠️  %d issues found\n", len(catResult.Issues))
			printIssues(catResult.Issues)
		}
		pterm.Println()
	}

	if catResult, ok := result.Categories[validate.CategorySecurity]; ok {
		pterm.DefaultSection.Println("🔒 Security Scan")
		if len(catResult.Issues) == 0 {
			pterm.Success.Printf("✅ No security issues found\n")
		} else {
			pterm.Warning.Printf("⚠️  %d security issues found\n", len(catResult.Issues))
			printIssues(catResult.Issues)
		}
		pterm.Println()
	}

	if catResult, ok := result.Categories[validate.CategoryDeprecation]; ok {
		pterm.DefaultSection.Println("⚠️  Deprecation Check")
		if len(catResult.Issues) == 0 {
			pterm.Success.Printf("✅ No deprecated APIs found\n")
		} else {
			pterm.Warning.Printf("⚠️  %d deprecated APIs found\n", len(catResult.Issues))
			printIssues(catResult.Issues)
		}
		pterm.Println()
	}

	if catResult, ok := result.Categories[validate.CategoryKustomize]; ok {
		pterm.DefaultSection.Println("📦 Kustomize Validation")
		if len(catResult.Issues) == 0 {
			pterm.Success.Printf("✅ All kustomizations build successfully\n")
		} else {
			pterm.Warning.Printf("⚠️  %d kustomization issues found\n", len(catResult.Issues))
			printIssues(catResult.Issues)
		}
		pterm.Println()
	}

	pterm.DefaultSection.Println("📊 Summary")

	tableData := pterm.TableData{
		{"Metric", "Count"},
		{"Total Manifests", fmt.Sprintf("%d", result.TotalManifests)},
		{"Passed", fmt.Sprintf("%d", result.Passed)},
		{"Warnings", fmt.Sprintf("%d", result.Warnings)},
		{"Failed", fmt.Sprintf("%d", result.Failed)},
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	pterm.Println()
	switch {
	case result.Failed > 0:
		pterm.Error.Printf("❌ Validation failed with %d critical/high issues\n", result.Failed)
	case result.Warnings > 0:
		pterm.Warning.Printf("⚠️  Validation passed with %d warnings\n", result.Warnings)
	default:
		pterm.Success.Printf("✅ All validations passed!\n")
	}
}

func printIssues(issues []validate.Issue) {
	for _, issue := range issues {
		severityColor := pterm.FgYellow
		switch issue.Severity {
		case validate.SeverityCritical:
			severityColor = pterm.FgRed
		case validate.SeverityHigh:
			severityColor = pterm.FgLightRed
		case validate.SeverityMedium:
			severityColor = pterm.FgYellow
		case validate.SeverityLow:
			severityColor = pterm.FgCyan
		}

		pterm.Printf("  ")
		pterm.NewStyle(severityColor).Printf("[%s]", strings.ToUpper(string(issue.Severity)))
		pterm.Printf(" %s\n", issue.File)
		pterm.Printf("    %s: %s\n", issue.Rule, issue.Message)
		if issue.Suggestion != "" {
			pterm.Printf("    💡 %s\n", issue.Suggestion)
		}
	}
}
