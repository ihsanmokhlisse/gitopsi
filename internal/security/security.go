// Package security provides security scanning, provenance, and input sanitization.
package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ScanSeverity represents the severity level of a security finding.
type ScanSeverity string

const (
	SeverityCritical ScanSeverity = "CRITICAL"
	SeverityHigh     ScanSeverity = "HIGH"
	SeverityMedium   ScanSeverity = "MEDIUM"
	SeverityLow      ScanSeverity = "LOW"
	SeverityInfo     ScanSeverity = "INFO"
)

// ScanResult represents the result of a security scan.
type ScanResult struct {
	Tool      string         `json:"tool"`
	Version   string         `json:"version,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Path      string         `json:"path"`
	Findings  []Finding      `json:"findings"`
	Summary   ScanSummary    `json:"summary"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Finding represents a single security finding.
type Finding struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Severity    ScanSeverity `json:"severity"`
	Category    string       `json:"category"`
	File        string       `json:"file,omitempty"`
	Line        int          `json:"line,omitempty"`
	Remediation string       `json:"remediation,omitempty"`
	References  []string     `json:"references,omitempty"`
}

// ScanSummary provides a summary of scan results.
type ScanSummary struct {
	TotalFiles    int `json:"total_files"`
	ScannedFiles  int `json:"scanned_files"`
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Info          int `json:"info"`
	PassedChecks  int `json:"passed_checks"`
	FailedChecks  int `json:"failed_checks"`
	SkippedChecks int `json:"skipped_checks"`
}

// Scanner provides security scanning capabilities.
type Scanner struct {
	path     string
	tools    []string
	severity ScanSeverity
	verbose  bool
}

// NewScanner creates a new security scanner.
func NewScanner(path string, minSeverity ScanSeverity, verbose bool) *Scanner {
	return &Scanner{
		path:     path,
		tools:    []string{"trivy", "checkov", "kubesec"},
		severity: minSeverity,
		verbose:  verbose,
	}
}

// Scan runs security scanning using available tools.
func (s *Scanner) Scan(ctx context.Context) (*ScanResult, error) {
	result := &ScanResult{
		Timestamp: time.Now(),
		Path:      s.path,
		Findings:  []Finding{},
		Summary:   ScanSummary{},
		Metadata:  make(map[string]any),
	}

	// Try trivy first
	if s.isToolAvailable("trivy") {
		trivyResult, err := s.runTrivy(ctx)
		if err == nil {
			result.Tool = "trivy"
			result.Findings = append(result.Findings, trivyResult...)
		}
	}

	// Try checkov
	if s.isToolAvailable("checkov") {
		checkovResult, err := s.runCheckov(ctx)
		if err == nil {
			if result.Tool == "" {
				result.Tool = "checkov"
			} else {
				result.Tool += "+checkov"
			}
			result.Findings = append(result.Findings, checkovResult...)
		}
	}

	// Run built-in security checks
	builtinResult := s.runBuiltinChecks()
	result.Findings = append(result.Findings, builtinResult...)
	if result.Tool == "" {
		result.Tool = "builtin"
	}

	// Calculate summary
	result.Summary = s.calculateSummary(result.Findings)

	return result, nil
}

// isToolAvailable checks if a tool is available in PATH.
func (s *Scanner) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// runTrivy runs trivy config scanner.
func (s *Scanner) runTrivy(ctx context.Context) ([]Finding, error) {
	args := []string{
		"config",
		"--format", "json",
		"--severity", string(s.severity) + ",HIGH,CRITICAL",
		s.path,
	}

	cmd := exec.CommandContext(ctx, "trivy", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return s.parseTrivyOutput(output)
}

// parseTrivyOutput parses trivy JSON output.
func (s *Scanner) parseTrivyOutput(output []byte) ([]Finding, error) {
	var trivyResults struct {
		Results []struct {
			Target            string `json:"Target"`
			Misconfigurations []struct {
				ID          string   `json:"ID"`
				Title       string   `json:"Title"`
				Description string   `json:"Description"`
				Severity    string   `json:"Severity"`
				Resolution  string   `json:"Resolution"`
				References  []string `json:"References"`
			} `json:"Misconfigurations"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(output, &trivyResults); err != nil {
		return nil, err
	}

	var findings []Finding
	for _, result := range trivyResults.Results {
		for _, misconfig := range result.Misconfigurations {
			findings = append(findings, Finding{
				ID:          misconfig.ID,
				Title:       misconfig.Title,
				Description: misconfig.Description,
				Severity:    ScanSeverity(misconfig.Severity),
				Category:    "misconfiguration",
				File:        result.Target,
				Remediation: misconfig.Resolution,
				References:  misconfig.References,
			})
		}
	}

	return findings, nil
}

// runCheckov runs checkov scanner.
func (s *Scanner) runCheckov(ctx context.Context) ([]Finding, error) {
	args := []string{
		"-d", s.path,
		"--framework", "kubernetes",
		"--output", "json",
		"--quiet",
	}

	cmd := exec.CommandContext(ctx, "checkov", args...)
	output, err := cmd.Output()
	if err != nil {
		// checkov returns non-zero on findings
		if len(output) > 0 {
			return s.parseCheckovOutput(output)
		}
		return nil, err
	}

	return s.parseCheckovOutput(output)
}

// parseCheckovOutput parses checkov JSON output.
func (s *Scanner) parseCheckovOutput(output []byte) ([]Finding, error) {
	var checkovResults struct {
		Results struct {
			FailedChecks []struct {
				CheckID       string `json:"check_id"`
				CheckName     string `json:"check_name"`
				FilePath      string `json:"file_path"`
				FileLineRange []int  `json:"file_line_range"`
				Guideline     string `json:"guideline"`
			} `json:"failed_checks"`
		} `json:"results"`
	}

	if err := json.Unmarshal(output, &checkovResults); err != nil {
		return nil, err
	}

	findings := make([]Finding, 0, len(checkovResults.Results.FailedChecks))
	for i := range checkovResults.Results.FailedChecks {
		check := &checkovResults.Results.FailedChecks[i]
		line := 0
		if len(check.FileLineRange) > 0 {
			line = check.FileLineRange[0]
		}
		findings = append(findings, Finding{
			ID:          check.CheckID,
			Title:       check.CheckName,
			Severity:    SeverityMedium, // Checkov doesn't provide severity by default
			Category:    "compliance",
			File:        check.FilePath,
			Line:        line,
			Remediation: check.Guideline,
		})
	}

	return findings, nil
}

// runBuiltinChecks runs built-in security checks.
func (s *Scanner) runBuiltinChecks() []Finding {
	var findings []Finding

	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, relErr := filepath.Rel(s.path, path)
		if relErr != nil {
			relPath = path // Fallback to absolute path
		}
		findings = append(findings, s.checkFile(relPath, string(content))...)
		return nil
	})

	if err != nil {
		return findings
	}

	return findings
}

// checkFile performs built-in security checks on a file.
func (s *Scanner) checkFile(path, content string) []Finding {
	var findings []Finding

	// Check for privileged containers
	if strings.Contains(content, "privileged: true") {
		findings = append(findings, Finding{
			ID:          "GITOPSI-001",
			Title:       "Privileged container detected",
			Description: "Container is running in privileged mode which gives full access to the host",
			Severity:    SeverityHigh,
			Category:    "container-security",
			File:        path,
			Remediation: "Remove 'privileged: true' and use specific capabilities if needed",
		})
	}

	// Check for runAsRoot
	if strings.Contains(content, "runAsUser: 0") {
		findings = append(findings, Finding{
			ID:          "GITOPSI-002",
			Title:       "Container running as root",
			Description: "Container is configured to run as root user (UID 0)",
			Severity:    SeverityMedium,
			Category:    "container-security",
			File:        path,
			Remediation: "Use a non-root user by setting 'runAsUser' to a non-zero value",
		})
	}

	// Check for missing resource limits
	if strings.Contains(content, "kind: Deployment") || strings.Contains(content, "kind: StatefulSet") {
		if !strings.Contains(content, "resources:") {
			findings = append(findings, Finding{
				ID:          "GITOPSI-003",
				Title:       "Missing resource limits",
				Description: "Container does not have resource limits defined",
				Severity:    SeverityLow,
				Category:    "reliability",
				File:        path,
				Remediation: "Add 'resources.limits' and 'resources.requests' to container spec",
			})
		}
	}

	// Check for hardcoded secrets
	// Look for env values with sensitive names that have hardcoded values
	// Check for direct password/secret/token/key values
	hasHardcodedSecret := false

	// Pattern 1: password: value or password: "value"
	if regexp.MustCompile(`(?i)password:\s*["']?[a-zA-Z0-9]+["']?`).MatchString(content) {
		hasHardcodedSecret = true
	}
	// Pattern 2: value: "secretXXX" on line with PASSWORD/SECRET/TOKEN/KEY
	if regexp.MustCompile(`(?i)(password|secret|token|api.?key).*\n.*value:\s*["'][^"']+["']`).MatchString(content) {
		hasHardcodedSecret = true
	}
	// Pattern 3: name: PASSWORD/SECRET/TOKEN followed by value:
	if regexp.MustCompile(`(?i)name:\s*["']?(PASSWORD|SECRET|TOKEN|API_KEY)["']?\s*\n\s*value:\s*["'][^"']+["']`).MatchString(content) {
		hasHardcodedSecret = true
	}

	if hasHardcodedSecret && !strings.Contains(content, "secretKeyRef") && !strings.Contains(content, "valueFrom") {
		findings = append(findings, Finding{
			ID:          "GITOPSI-004",
			Title:       "Potential hardcoded secret",
			Description: "File may contain hardcoded credentials or secrets",
			Severity:    SeverityHigh,
			Category:    "secrets",
			File:        path,
			Remediation: "Use Kubernetes Secrets or external secret management",
		})
	}

	// Check for hostPath volumes
	if strings.Contains(content, "hostPath:") {
		findings = append(findings, Finding{
			ID:          "GITOPSI-005",
			Title:       "HostPath volume detected",
			Description: "Pod uses hostPath which exposes the host filesystem",
			Severity:    SeverityMedium,
			Category:    "container-security",
			File:        path,
			Remediation: "Use persistent volumes or emptyDir instead of hostPath",
		})
	}

	// Check for hostNetwork
	if strings.Contains(content, "hostNetwork: true") {
		findings = append(findings, Finding{
			ID:          "GITOPSI-006",
			Title:       "Host network enabled",
			Description: "Pod uses host network namespace",
			Severity:    SeverityMedium,
			Category:    "network-security",
			File:        path,
			Remediation: "Remove 'hostNetwork: true' unless absolutely necessary",
		})
	}

	// Check for missing network policies (directory level)
	if strings.Contains(path, "namespace") && !strings.Contains(content, "NetworkPolicy") {
		// This is a namespace file, check if network policy exists
		findings = append(findings, Finding{
			ID:          "GITOPSI-007",
			Title:       "Consider adding NetworkPolicy",
			Description: "Namespace may benefit from NetworkPolicy for traffic control",
			Severity:    SeverityInfo,
			Category:    "network-security",
			File:        path,
			Remediation: "Add NetworkPolicy to restrict pod-to-pod communication",
		})
	}

	return findings
}

// calculateSummary calculates the scan summary.
func (s *Scanner) calculateSummary(findings []Finding) ScanSummary {
	summary := ScanSummary{}

	for i := range findings {
		switch findings[i].Severity {
		case SeverityCritical:
			summary.Critical++
		case SeverityHigh:
			summary.High++
		case SeverityMedium:
			summary.Medium++
		case SeverityLow:
			summary.Low++
		case SeverityInfo:
			summary.Info++
		}
	}

	summary.FailedChecks = len(findings)
	return summary
}

// ShouldFail determines if the scan should fail based on severity threshold.
func (r *ScanResult) ShouldFail(minSeverity ScanSeverity) bool {
	switch minSeverity {
	case SeverityCritical:
		return r.Summary.Critical > 0
	case SeverityHigh:
		return r.Summary.Critical > 0 || r.Summary.High > 0
	case SeverityMedium:
		return r.Summary.Critical > 0 || r.Summary.High > 0 || r.Summary.Medium > 0
	case SeverityLow:
		return r.Summary.Critical > 0 || r.Summary.High > 0 || r.Summary.Medium > 0 || r.Summary.Low > 0
	default:
		return false
	}
}

// Provenance represents the provenance information for generated manifests.
type Provenance struct {
	BuildTimestamp   time.Time        `json:"build_timestamp"`
	GitCommit        string           `json:"git_commit,omitempty"`
	GitBranch        string           `json:"git_branch,omitempty"`
	GitRemote        string           `json:"git_remote,omitempty"`
	Generator        string           `json:"generator"`
	GeneratorVersion string           `json:"generator_version"`
	ConfigHash       string           `json:"config_hash"`
	OutputHash       string           `json:"output_hash"`
	Files            []FileProvenance `json:"files"`
	Signature        string           `json:"signature,omitempty"`
}

// FileProvenance represents provenance for a single file.
type FileProvenance struct {
	Path       string    `json:"path"`
	SHA256     string    `json:"sha256"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

// GenerateProvenance generates provenance information for the output directory.
func GenerateProvenance(outputPath, configPath, generatorVersion string) (*Provenance, error) {
	prov := &Provenance{
		BuildTimestamp:   time.Now(),
		Generator:        "gitopsi",
		GeneratorVersion: generatorVersion,
		Files:            []FileProvenance{},
	}

	// Get Git information
	if commit, err := getGitCommit(); err == nil {
		prov.GitCommit = commit
	}
	if branch, err := getGitBranch(); err == nil {
		prov.GitBranch = branch
	}
	if remote, err := getGitRemote(); err == nil {
		prov.GitRemote = remote
	}

	// Hash config file
	if configPath != "" {
		if hash, err := hashFile(configPath); err == nil {
			prov.ConfigHash = hash
		}
	}

	// Walk output directory and hash files
	err := filepath.Walk(outputPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}

		relPath, relErr := filepath.Rel(outputPath, path)
		if relErr != nil {
			relPath = path // Fallback to absolute path
		}
		hash, hashErr := hashFile(path)
		if hashErr != nil {
			return nil
		}

		prov.Files = append(prov.Files, FileProvenance{
			Path:       relPath,
			SHA256:     hash,
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Calculate overall output hash
	prov.OutputHash = calculateOutputHash(prov.Files)

	return prov, nil
}

// hashFile calculates SHA256 hash of a file.
func hashFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// calculateOutputHash calculates a combined hash of all output files.
func calculateOutputHash(files []FileProvenance) string {
	hashes := make([]string, 0, len(files))
	for i := range files {
		hashes = append(hashes, files[i].SHA256)
	}
	combined := strings.Join(hashes, "")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// getGitCommit returns the current Git commit hash.
func getGitCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getGitBranch returns the current Git branch.
func getGitBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getGitRemote returns the origin remote URL.
func getGitRemote() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// SaveProvenance saves provenance to a JSON file.
func (p *Provenance) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ToJSON returns provenance as JSON.
func (p *Provenance) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Sanitizer provides input sanitization to prevent code injection.
type Sanitizer struct {
	allowedCharsRegex *regexp.Regexp
	maxLength         int
}

// NewSanitizer creates a new input sanitizer.
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		// Allow alphanumeric, dash, underscore, dot, slash, colon, @
		allowedCharsRegex: regexp.MustCompile(`[^a-zA-Z0-9\-_./:@]`),
		maxLength:         1024,
	}
}

// SanitizeName sanitizes a Kubernetes resource name.
func (s *Sanitizer) SanitizeName(name string) string {
	// Kubernetes names must be lowercase, alphanumeric with dashes
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	// Max 63 characters for most K8s names
	if len(name) > 63 {
		name = name[:63]
	}

	return name
}

// SanitizeLabel sanitizes a Kubernetes label value.
func (s *Sanitizer) SanitizeLabel(value string) string {
	// Labels: alphanumeric, dash, underscore, dot; max 63 chars
	value = regexp.MustCompile(`[^a-zA-Z0-9\-_.]`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_.")

	if len(value) > 63 {
		value = value[:63]
	}

	return value
}

// SanitizeNamespace sanitizes a namespace name.
func (s *Sanitizer) SanitizeNamespace(ns string) string {
	return s.SanitizeName(ns)
}

// SanitizePath sanitizes a file path.
func (s *Sanitizer) SanitizePath(path string) string {
	// Remove any path traversal attempts
	path = strings.ReplaceAll(path, "..", "")
	path = strings.ReplaceAll(path, "//", "/")
	path = filepath.Clean(path)

	// Always use forward slashes for GitOps compatibility (YAML paths must be portable)
	path = filepath.ToSlash(path)

	// Remove leading slash for relative paths
	path = strings.TrimPrefix(path, "/")

	return path
}

// SanitizeURL sanitizes a URL.
func (s *Sanitizer) SanitizeURL(url string) string {
	// Allow common URL characters
	url = regexp.MustCompile(`[^a-zA-Z0-9\-_./:@?=&%]`).ReplaceAllString(url, "")
	return url
}

// SanitizeImage sanitizes a container image reference.
func (s *Sanitizer) SanitizeImage(image string) string {
	// Image references: registry/repo/name:tag@digest
	image = regexp.MustCompile(`[^a-zA-Z0-9\-_./:@]`).ReplaceAllString(image, "")
	return image
}

// ValidateInput validates that input doesn't contain injection patterns.
func (s *Sanitizer) ValidateInput(input string) error {
	// Check for template injection
	if strings.Contains(input, "{{") || strings.Contains(input, "}}") {
		return fmt.Errorf("template injection detected in input")
	}

	// Check for command injection
	dangerousPatterns := []string{
		"`", "$(", ";", "&&", "||", "|", ">", "<", "$(",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(input, pattern) {
			return fmt.Errorf("potential command injection detected: %s", pattern)
		}
	}

	// Check length
	if len(input) > s.maxLength {
		return fmt.Errorf("input exceeds maximum length of %d", s.maxLength)
	}

	return nil
}

// SanitizeConfig sanitizes all string fields in a configuration.
func (s *Sanitizer) SanitizeConfig(config map[string]any) map[string]any {
	sanitized := make(map[string]any)

	for key, value := range config {
		switch v := value.(type) {
		case string:
			sanitized[key] = s.sanitizeStringField(key, v)
		case map[string]any:
			sanitized[key] = s.SanitizeConfig(v)
		default:
			sanitized[key] = v
		}
	}

	return sanitized
}

// sanitizeStringField sanitizes a string field based on its key.
func (s *Sanitizer) sanitizeStringField(key, value string) string {
	switch {
	case key == "name" || strings.HasSuffix(key, "name"):
		return s.SanitizeName(value)
	case key == "namespace":
		return s.SanitizeNamespace(value)
	case key == "image":
		return s.SanitizeImage(value)
	case key == "url" || strings.HasSuffix(key, "url"):
		return s.SanitizeURL(value)
	case key == "path" || strings.HasSuffix(key, "path"):
		return s.SanitizePath(value)
	default:
		return s.SanitizeLabel(value)
	}
}
