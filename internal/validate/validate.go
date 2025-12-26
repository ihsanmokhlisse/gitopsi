package validate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Category string

const (
	CategorySchema       Category = "schema"
	CategorySecurity     Category = "security"
	CategoryDeprecation  Category = "deprecation"
	CategoryBestPractice Category = "best-practice"
	CategoryKustomize    Category = "kustomize"
)

type Issue struct {
	File       string   `json:"file" yaml:"file"`
	Line       int      `json:"line,omitempty" yaml:"line,omitempty"`
	Category   Category `json:"category" yaml:"category"`
	Severity   Severity `json:"severity" yaml:"severity"`
	Rule       string   `json:"rule" yaml:"rule"`
	Message    string   `json:"message" yaml:"message"`
	Suggestion string   `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
	Fixable    bool     `json:"fixable" yaml:"fixable"`
}

type ValidationResult struct {
	Path           string                       `json:"path" yaml:"path"`
	TotalManifests int                          `json:"total_manifests" yaml:"total_manifests"`
	Passed         int                          `json:"passed" yaml:"passed"`
	Warnings       int                          `json:"warnings" yaml:"warnings"`
	Failed         int                          `json:"failed" yaml:"failed"`
	Issues         []Issue                      `json:"issues" yaml:"issues"`
	Categories     map[Category]*CategoryResult `json:"categories" yaml:"categories"`
}

type CategoryResult struct {
	Passed int     `json:"passed" yaml:"passed"`
	Failed int     `json:"failed" yaml:"failed"`
	Issues []Issue `json:"issues" yaml:"issues"`
}

type Options struct {
	Path          string
	K8sVersion    string
	ArgoCDVersion string
	Schema        bool
	Security      bool
	Deprecation   bool
	BestPractice  bool
	Kustomize     bool
	FailOn        Severity
	OutputFormat  string
	Fix           bool
}

func DefaultOptions() *Options {
	return &Options{
		K8sVersion:    "1.29",
		ArgoCDVersion: "2.10",
		Schema:        true,
		Security:      true,
		Deprecation:   true,
		BestPractice:  true,
		Kustomize:     true,
		FailOn:        SeverityHigh,
		OutputFormat:  "table",
	}
}

type Validator struct {
	opts *Options
}

func New(opts *Options) *Validator {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &Validator{opts: opts}
}

func (v *Validator) Validate(ctx context.Context) (*ValidationResult, error) {
	if _, err := os.Stat(v.opts.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", v.opts.Path)
	}

	result := &ValidationResult{
		Path:       v.opts.Path,
		Issues:     []Issue{},
		Categories: make(map[Category]*CategoryResult),
	}

	manifests, err := v.findManifests()
	if err != nil {
		return nil, fmt.Errorf("failed to find manifests: %w", err)
	}
	result.TotalManifests = len(manifests)

	if v.opts.Schema {
		if schemaErr := v.validateSchema(ctx, manifests, result); schemaErr != nil {
			return nil, fmt.Errorf("schema validation failed: %w", schemaErr)
		}
	}

	if v.opts.Security {
		if secErr := v.validateSecurity(ctx, result); secErr != nil {
			return nil, fmt.Errorf("security validation failed: %w", secErr)
		}
	}

	if v.opts.Deprecation {
		if depErr := v.validateDeprecation(ctx, manifests, result); depErr != nil {
			return nil, fmt.Errorf("deprecation check failed: %w", depErr)
		}
	}

	if v.opts.Kustomize {
		if kusErr := v.validateKustomize(ctx, result); kusErr != nil {
			return nil, fmt.Errorf("kustomize validation failed: %w", kusErr)
		}
	}

	v.calculateSummary(result)

	return result, nil
}

func (v *Validator) findManifests() ([]string, error) {
	var manifests []string

	err := filepath.Walk(v.opts.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			if !strings.Contains(path, "kustomization") {
				manifests = append(manifests, path)
			}
		}
		return nil
	})

	return manifests, err
}

func (v *Validator) validateSchema(ctx context.Context, manifests []string, result *ValidationResult) error {
	catResult := &CategoryResult{Issues: []Issue{}}
	result.Categories[CategorySchema] = catResult

	kubeconformPath, err := exec.LookPath("kubeconform")
	if err != nil {
		for _, m := range manifests {
			if validateErr := v.basicYAMLValidation(m, catResult); validateErr != nil {
				continue
			}
		}
		catResult.Passed = len(manifests) - catResult.Failed
		result.Issues = append(result.Issues, catResult.Issues...)
		return nil
	}

	for _, manifest := range manifests {
		args := []string{
			"-kubernetes-version", v.opts.K8sVersion,
			"-summary",
			"-output", "json",
			manifest,
		}

		cmd := exec.CommandContext(ctx, kubeconformPath, args...)
		output, cmdErr := cmd.CombinedOutput()

		if cmdErr != nil {
			var kubeResult struct {
				Resources []struct {
					Filename string `json:"filename"`
					Kind     string `json:"kind"`
					Name     string `json:"name"`
					Status   string `json:"status"`
					Msg      string `json:"msg"`
				} `json:"resources"`
			}

			if jsonErr := json.Unmarshal(output, &kubeResult); jsonErr == nil {
				for _, r := range kubeResult.Resources {
					if r.Status != "ok" && r.Status != "skipped" {
						issue := Issue{
							File:     r.Filename,
							Category: CategorySchema,
							Severity: SeverityHigh,
							Rule:     "kubeconform",
							Message:  r.Msg,
						}
						catResult.Issues = append(catResult.Issues, issue)
						catResult.Failed++
					}
				}
			} else {
				catResult.Passed++
			}
		} else {
			catResult.Passed++
		}
	}

	result.Issues = append(result.Issues, catResult.Issues...)
	return nil
}

func (v *Validator) basicYAMLValidation(path string, catResult *CategoryResult) error {
	data, err := os.ReadFile(path)
	if err != nil {
		issue := Issue{
			File:     path,
			Category: CategorySchema,
			Severity: SeverityHigh,
			Rule:     "yaml-read",
			Message:  fmt.Sprintf("Cannot read file: %v", err),
		}
		catResult.Issues = append(catResult.Issues, issue)
		catResult.Failed++
		return err
	}

	var doc interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		issue := Issue{
			File:     path,
			Category: CategorySchema,
			Severity: SeverityHigh,
			Rule:     "yaml-syntax",
			Message:  fmt.Sprintf("Invalid YAML syntax: %v", err),
		}
		catResult.Issues = append(catResult.Issues, issue)
		catResult.Failed++
		return err
	}

	catResult.Passed++
	return nil
}

func (v *Validator) validateSecurity(ctx context.Context, result *ValidationResult) error {
	catResult := &CategoryResult{Issues: []Issue{}}
	result.Categories[CategorySecurity] = catResult

	trivyPath, trivyErr := exec.LookPath("trivy")
	if trivyErr == nil {
		if err := v.runTrivy(ctx, trivyPath, catResult); err != nil {
			return err
		}
	}

	checkovPath, checkovErr := exec.LookPath("checkov")
	if checkovErr == nil {
		if err := v.runCheckov(ctx, checkovPath, catResult); err != nil {
			return err
		}
	}

	if trivyErr != nil && checkovErr != nil {
		v.runBasicSecurityChecks(result, catResult)
	}

	result.Issues = append(result.Issues, catResult.Issues...)
	return nil
}

func (v *Validator) runTrivy(ctx context.Context, trivyPath string, catResult *CategoryResult) error {
	args := []string{
		"config",
		"--format", "json",
		"--severity", "HIGH,CRITICAL",
		v.opts.Path,
	}

	cmd := exec.CommandContext(ctx, trivyPath, args...)
	output, _ := cmd.CombinedOutput()

	var trivyResult struct {
		Results []struct {
			Target            string `json:"Target"`
			Misconfigurations []struct {
				ID          string `json:"ID"`
				Title       string `json:"Title"`
				Description string `json:"Description"`
				Severity    string `json:"Severity"`
				Resolution  string `json:"Resolution"`
			} `json:"Misconfigurations"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(output, &trivyResult); err == nil {
		for _, r := range trivyResult.Results {
			for _, m := range r.Misconfigurations {
				severity := SeverityMedium
				switch strings.ToLower(m.Severity) {
				case "critical":
					severity = SeverityCritical
				case "high":
					severity = SeverityHigh
				case "low":
					severity = SeverityLow
				}

				issue := Issue{
					File:       r.Target,
					Category:   CategorySecurity,
					Severity:   severity,
					Rule:       m.ID,
					Message:    m.Title,
					Suggestion: m.Resolution,
				}
				catResult.Issues = append(catResult.Issues, issue)
				catResult.Failed++
			}
		}
	}

	return nil
}

func (v *Validator) runCheckov(ctx context.Context, checkovPath string, catResult *CategoryResult) error {
	args := []string{
		"-d", v.opts.Path,
		"--framework", "kubernetes",
		"--output", "json",
		"--quiet",
	}

	cmd := exec.CommandContext(ctx, checkovPath, args...)
	output, _ := cmd.CombinedOutput()

	var checkovResult struct {
		Results struct {
			FailedChecks []struct {
				Check struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"check"`
				FilePath  string `json:"file_path"`
				Guideline string `json:"guideline"`
			} `json:"failed_checks"`
		} `json:"results"`
	}

	if err := json.Unmarshal(output, &checkovResult); err == nil {
		for _, fc := range checkovResult.Results.FailedChecks {
			issue := Issue{
				File:       fc.FilePath,
				Category:   CategorySecurity,
				Severity:   SeverityMedium,
				Rule:       fc.Check.ID,
				Message:    fc.Check.Name,
				Suggestion: fc.Guideline,
			}
			catResult.Issues = append(catResult.Issues, issue)
			catResult.Failed++
		}
	}

	return nil
}

func (v *Validator) runBasicSecurityChecks(result *ValidationResult, catResult *CategoryResult) {
	manifests, _ := v.findManifests()
	for _, manifest := range manifests {
		data, err := os.ReadFile(manifest)
		if err != nil {
			continue
		}

		content := string(data)

		if strings.Contains(content, "privileged: true") {
			issue := Issue{
				File:       manifest,
				Category:   CategorySecurity,
				Severity:   SeverityHigh,
				Rule:       "SEC001",
				Message:    "Container running in privileged mode",
				Suggestion: "Remove 'privileged: true' from securityContext",
			}
			catResult.Issues = append(catResult.Issues, issue)
			catResult.Failed++
		}

		if strings.Contains(content, "runAsUser: 0") || strings.Contains(content, "runAsNonRoot: false") {
			issue := Issue{
				File:       manifest,
				Category:   CategorySecurity,
				Severity:   SeverityMedium,
				Rule:       "SEC002",
				Message:    "Container may run as root",
				Suggestion: "Set runAsNonRoot: true and runAsUser to non-zero value",
			}
			catResult.Issues = append(catResult.Issues, issue)
			catResult.Failed++
		}

		if !strings.Contains(content, "resources:") && strings.Contains(content, "kind: Deployment") {
			issue := Issue{
				File:       manifest,
				Category:   CategorySecurity,
				Severity:   SeverityMedium,
				Rule:       "SEC003",
				Message:    "No resource limits defined",
				Suggestion: "Define resources.limits and resources.requests",
			}
			catResult.Issues = append(catResult.Issues, issue)
			catResult.Failed++
		}
	}

	catResult.Passed = result.TotalManifests - catResult.Failed
}

func (v *Validator) validateDeprecation(ctx context.Context, manifests []string, result *ValidationResult) error {
	catResult := &CategoryResult{Issues: []Issue{}}
	result.Categories[CategoryDeprecation] = catResult

	plutoPath, err := exec.LookPath("pluto")
	if err != nil {
		v.runBasicDeprecationChecks(manifests, catResult)
		result.Issues = append(result.Issues, catResult.Issues...)
		return nil
	}

	args := []string{
		"detect-files",
		"-d", v.opts.Path,
		"--target-versions", fmt.Sprintf("k8s=v%s", v.opts.K8sVersion),
		"-o", "json",
	}

	cmd := exec.CommandContext(ctx, plutoPath, args...)
	output, _ := cmd.CombinedOutput()

	var plutoResult struct {
		Items []struct {
			Name       string `json:"name"`
			FilePath   string `json:"filePath"`
			Kind       string `json:"kind"`
			APIVersion struct {
				Version        string `json:"version"`
				Kind           string `json:"kind"`
				Deprecated     bool   `json:"deprecated"`
				Removed        bool   `json:"removed"`
				ReplacementAPI string `json:"replacementAPI"`
			} `json:"apiVersion"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &plutoResult); err == nil {
		for _, item := range plutoResult.Items {
			if !item.APIVersion.Deprecated && !item.APIVersion.Removed {
				continue
			}

			severity := SeverityMedium
			if item.APIVersion.Removed {
				severity = SeverityHigh
			}

			msg := fmt.Sprintf("%s %s uses deprecated API %s", item.Kind, item.Name, item.APIVersion.Version)
			if item.APIVersion.Removed {
				msg = fmt.Sprintf("%s %s uses removed API %s", item.Kind, item.Name, item.APIVersion.Version)
			}

			issue := Issue{
				File:       item.FilePath,
				Category:   CategoryDeprecation,
				Severity:   severity,
				Rule:       "DEP001",
				Message:    msg,
				Suggestion: fmt.Sprintf("Migrate to %s", item.APIVersion.ReplacementAPI),
			}
			catResult.Issues = append(catResult.Issues, issue)
			catResult.Failed++
		}
	}

	catResult.Passed = len(manifests) - catResult.Failed
	result.Issues = append(result.Issues, catResult.Issues...)
	return nil
}

func (v *Validator) runBasicDeprecationChecks(manifests []string, catResult *CategoryResult) {
	deprecatedAPIs := map[string]string{
		"extensions/v1beta1":                   "apps/v1",
		"apps/v1beta1":                         "apps/v1",
		"apps/v1beta2":                         "apps/v1",
		"networking.k8s.io/v1beta1":            "networking.k8s.io/v1",
		"rbac.authorization.k8s.io/v1beta1":    "rbac.authorization.k8s.io/v1",
		"admissionregistration.k8s.io/v1beta1": "admissionregistration.k8s.io/v1",
	}

	for _, manifest := range manifests {
		data, err := os.ReadFile(manifest)
		if err != nil {
			continue
		}

		content := string(data)
		for deprecated, replacement := range deprecatedAPIs {
			if strings.Contains(content, "apiVersion: "+deprecated) {
				issue := Issue{
					File:       manifest,
					Category:   CategoryDeprecation,
					Severity:   SeverityMedium,
					Rule:       "DEP001",
					Message:    fmt.Sprintf("Uses deprecated API version: %s", deprecated),
					Suggestion: fmt.Sprintf("Migrate to %s", replacement),
					Fixable:    true,
				}
				catResult.Issues = append(catResult.Issues, issue)
				catResult.Failed++
			}
		}
	}

	catResult.Passed = len(manifests) - catResult.Failed
}

func (v *Validator) validateKustomize(ctx context.Context, result *ValidationResult) error {
	catResult := &CategoryResult{Issues: []Issue{}}
	result.Categories[CategoryKustomize] = catResult

	kustomizeFiles := []string{}
	_ = filepath.Walk(v.opts.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.Contains(filepath.Base(path), "kustomization") {
			kustomizeFiles = append(kustomizeFiles, filepath.Dir(path))
		}
		return nil
	})

	kustomizePath, err := exec.LookPath("kustomize")
	kubectlPath, kubectlErr := exec.LookPath("kubectl")

	buildCmd := ""
	if err == nil {
		buildCmd = kustomizePath
	} else if kubectlErr == nil {
		buildCmd = kubectlPath
	}

	for _, kDir := range kustomizeFiles {
		var cmd *exec.Cmd
		switch {
		case buildCmd == kustomizePath:
			cmd = exec.CommandContext(ctx, buildCmd, "build", kDir)
		case buildCmd == kubectlPath:
			cmd = exec.CommandContext(ctx, buildCmd, "kustomize", kDir)
		default:
			catResult.Passed++
			continue
		}

		output, buildErr := cmd.CombinedOutput()
		if buildErr != nil {
			issue := Issue{
				File:     filepath.Join(kDir, "kustomization.yaml"),
				Category: CategoryKustomize,
				Severity: SeverityHigh,
				Rule:     "KUS001",
				Message:  fmt.Sprintf("Kustomize build failed: %s", strings.TrimSpace(string(output))),
			}
			catResult.Issues = append(catResult.Issues, issue)
			catResult.Failed++
		} else {
			catResult.Passed++
		}
	}

	result.Issues = append(result.Issues, catResult.Issues...)
	return nil
}

func (v *Validator) calculateSummary(result *ValidationResult) {
	result.Passed = result.TotalManifests
	result.Warnings = 0
	result.Failed = 0

	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityCritical, SeverityHigh:
			result.Failed++
			result.Passed--
		case SeverityMedium, SeverityLow:
			result.Warnings++
		}
	}

	if result.Passed < 0 {
		result.Passed = 0
	}
}

func (v *Validator) ShouldFail(result *ValidationResult) bool {
	for _, issue := range result.Issues {
		switch v.opts.FailOn {
		case SeverityCritical:
			if issue.Severity == SeverityCritical {
				return true
			}
		case SeverityHigh:
			if issue.Severity == SeverityCritical || issue.Severity == SeverityHigh {
				return true
			}
		case SeverityMedium:
			if issue.Severity == SeverityCritical || issue.Severity == SeverityHigh || issue.Severity == SeverityMedium {
				return true
			}
		case SeverityLow:
			if issue.Severity != SeverityInfo {
				return true
			}
		}
	}
	return false
}

func (result *ValidationResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (result *ValidationResult) ToYAML() (string, error) {
	data, err := yaml.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
