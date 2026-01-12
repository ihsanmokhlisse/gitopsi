//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E2ETestConfig holds configuration for E2E tests.
type E2ETestConfig struct {
	BinaryPath    string
	Timeout       time.Duration
	KubeAvailable bool
	ArgoCDEnabled bool
	SyncEnabled   bool
	Verbose       bool
}

// DefaultE2EConfig returns default E2E test configuration.
func DefaultE2EConfig() *E2ETestConfig {
	return &E2ETestConfig{
		BinaryPath:    binaryPath,
		Timeout:       2 * time.Minute,
		KubeAvailable: os.Getenv("KUBECONFIG") != "" || os.Getenv("E2E_CLUSTER") != "",
		ArgoCDEnabled: os.Getenv("E2E_ARGOCD") != "",
		SyncEnabled:   os.Getenv("E2E_SYNC_TEST") != "",
		Verbose:       os.Getenv("E2E_VERBOSE") != "",
	}
}

// CommandResult holds the result of a command execution.
type CommandResult struct {
	Stdout   string
	Stderr   string
	Combined string
	ExitCode int
	Duration time.Duration
	Error    error
}

// Success returns true if the command succeeded.
func (r *CommandResult) Success() bool {
	return r.Error == nil && r.ExitCode == 0
}

// Contains checks if the combined output contains a substring.
func (r *CommandResult) Contains(substr string) bool {
	return strings.Contains(r.Combined, substr)
}

// ContainsAll checks if the combined output contains all substrings.
func (r *CommandResult) ContainsAll(substrs ...string) bool {
	for _, s := range substrs {
		if !strings.Contains(r.Combined, s) {
			return false
		}
	}
	return true
}

// ContainsAny checks if the combined output contains any of the substrings.
func (r *CommandResult) ContainsAny(substrs ...string) bool {
	for _, s := range substrs {
		if strings.Contains(r.Combined, s) {
			return true
		}
	}
	return false
}

// RunGitopsi runs the gitopsi binary with the given arguments.
func RunGitopsi(t *testing.T, args ...string) *CommandResult {
	t.Helper()
	return RunGitopsiWithTimeout(t, 2*time.Minute, args...)
}

// RunGitopsiWithTimeout runs the gitopsi binary with a custom timeout.
func RunGitopsiWithTimeout(t *testing.T, timeout time.Duration, args ...string) *CommandResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunGitopsiWithContext(t, ctx, args...)
}

// RunGitopsiWithContext runs the gitopsi binary with a context.
func RunGitopsiWithContext(t *testing.T, ctx context.Context, args ...string) *CommandResult {
	t.Helper()

	start := time.Now()
	cmd := exec.CommandContext(ctx, binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	combined := stdout.String() + stderr.String()

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Combined: combined,
		ExitCode: exitCode,
		Duration: duration,
		Error:    err,
	}
}

// RunCommand runs an arbitrary command with timeout.
func RunCommand(t *testing.T, name string, args ...string) *CommandResult {
	t.Helper()
	return RunCommandWithTimeout(t, 2*time.Minute, name, args...)
}

// RunCommandWithTimeout runs an arbitrary command with custom timeout.
func RunCommandWithTimeout(t *testing.T, timeout time.Duration, name string, args ...string) *CommandResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	combined := stdout.String() + stderr.String()

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Combined: combined,
		ExitCode: exitCode,
		Duration: duration,
		Error:    err,
	}
}

// ProjectValidator validates a generated project structure.
type ProjectValidator struct {
	t          *testing.T
	projectDir string
	errors     []string
}

// NewProjectValidator creates a new project validator.
func NewProjectValidator(t *testing.T, projectDir string) *ProjectValidator {
	return &ProjectValidator{
		t:          t,
		projectDir: projectDir,
		errors:     []string{},
	}
}

// DirectoryExists checks if a directory exists.
func (v *ProjectValidator) DirectoryExists(relPath string) *ProjectValidator {
	v.t.Helper()
	fullPath := filepath.Join(v.projectDir, relPath)
	if info, err := os.Stat(fullPath); err != nil || !info.IsDir() {
		v.errors = append(v.errors, fmt.Sprintf("directory not found: %s", relPath))
	}
	return v
}

// FileExists checks if a file exists.
func (v *ProjectValidator) FileExists(relPath string) *ProjectValidator {
	v.t.Helper()
	fullPath := filepath.Join(v.projectDir, relPath)
	if info, err := os.Stat(fullPath); err != nil || info.IsDir() {
		v.errors = append(v.errors, fmt.Sprintf("file not found: %s", relPath))
	}
	return v
}

// FileContains checks if a file contains a substring.
func (v *ProjectValidator) FileContains(relPath, substr string) *ProjectValidator {
	v.t.Helper()
	fullPath := filepath.Join(v.projectDir, relPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("cannot read file %s: %v", relPath, err))
		return v
	}
	if !strings.Contains(string(content), substr) {
		v.errors = append(v.errors, fmt.Sprintf("file %s does not contain: %s", relPath, substr))
	}
	return v
}

// FileNotEmpty checks if a file is not empty.
func (v *ProjectValidator) FileNotEmpty(relPath string) *ProjectValidator {
	v.t.Helper()
	fullPath := filepath.Join(v.projectDir, relPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("file not found: %s", relPath))
		return v
	}
	if info.Size() == 0 {
		v.errors = append(v.errors, fmt.Sprintf("file is empty: %s", relPath))
	}
	return v
}

// HasErrors returns true if validation errors occurred.
func (v *ProjectValidator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors.
func (v *ProjectValidator) Errors() []string {
	return v.errors
}

// Assert fails the test if any validation errors occurred.
func (v *ProjectValidator) Assert() {
	v.t.Helper()
	if len(v.errors) > 0 {
		v.t.Errorf("Project validation failed with %d errors:\n  - %s",
			len(v.errors), strings.Join(v.errors, "\n  - "))
	}
}

// ValidateMinimalPreset validates a minimal preset project structure.
func ValidateMinimalPreset(t *testing.T, projectDir string) {
	t.Helper()
	v := NewProjectValidator(t, projectDir)
	v.DirectoryExists("infrastructure").
		DirectoryExists("infrastructure/base").
		DirectoryExists("infrastructure/base/namespaces").
		DirectoryExists("argocd").
		DirectoryExists("docs").
		FileExists("docs/README.md").
		Assert()
}

// ValidateStandardPreset validates a standard preset project structure.
func ValidateStandardPreset(t *testing.T, projectDir string) {
	t.Helper()
	v := NewProjectValidator(t, projectDir)
	v.DirectoryExists("infrastructure").
		DirectoryExists("infrastructure/base/namespaces").
		DirectoryExists("infrastructure/base/rbac").
		DirectoryExists("infrastructure/base/network-policies").
		DirectoryExists("infrastructure/base/resource-quotas").
		DirectoryExists("infrastructure/overlays/dev").
		DirectoryExists("infrastructure/overlays/staging").
		DirectoryExists("infrastructure/overlays/prod").
		DirectoryExists("applications/base").
		DirectoryExists("applications/overlays/dev").
		DirectoryExists("argocd/projects").
		DirectoryExists("argocd/applicationsets").
		DirectoryExists("docs").
		DirectoryExists("scripts").
		FileExists("docs/README.md").
		FileExists("docs/ARCHITECTURE.md").
		FileExists("docs/ONBOARDING.md").
		Assert()
}

// ValidateEnterprisePreset validates an enterprise preset project structure.
func ValidateEnterprisePreset(t *testing.T, projectDir string) {
	t.Helper()
	// Enterprise includes everything in standard
	ValidateStandardPreset(t, projectDir)

	// Plus additional directories
	v := NewProjectValidator(t, projectDir)
	v.DirectoryExists("bootstrap").
		Assert()
}

// FixtureConfig represents a test fixture configuration.
type FixtureConfig struct {
	Name        string
	ConfigFile  string
	ProjectName string
	Preset      string
}

// GetFixture returns a fixture configuration by name.
func GetFixture(name string) FixtureConfig {
	fixtures := map[string]FixtureConfig{
		"minimal": {
			Name:        "minimal",
			ConfigFile:  "fixtures/minimal-config.yaml",
			ProjectName: "test-minimal",
			Preset:      "minimal",
		},
		"standard": {
			Name:        "standard",
			ConfigFile:  "fixtures/standard-config.yaml",
			ProjectName: "test-standard",
			Preset:      "standard",
		},
		"enterprise": {
			Name:        "enterprise",
			ConfigFile:  "fixtures/enterprise-config.yaml",
			ProjectName: "test-enterprise",
			Preset:      "enterprise",
		},
		"infra-only": {
			Name:        "infra-only",
			ConfigFile:  "fixtures/infra-only-config.yaml",
			ProjectName: "test-infra",
			Preset:      "standard",
		},
		"app-only": {
			Name:        "app-only",
			ConfigFile:  "fixtures/app-only-config.yaml",
			ProjectName: "test-app",
			Preset:      "standard",
		},
		"multi-cluster": {
			Name:        "multi-cluster",
			ConfigFile:  "fixtures/multi-cluster-config.yaml",
			ProjectName: "test-multi-cluster",
			Preset:      "standard",
		},
	}

	if f, ok := fixtures[name]; ok {
		return f
	}
	return FixtureConfig{Name: name}
}

// ListAllFixtures returns all available fixtures.
func ListAllFixtures() []FixtureConfig {
	return []FixtureConfig{
		GetFixture("minimal"),
		GetFixture("standard"),
		GetFixture("enterprise"),
		GetFixture("infra-only"),
		GetFixture("app-only"),
		GetFixture("multi-cluster"),
	}
}

// CleanupFunc represents a cleanup function to be called after test.
type CleanupFunc func()

// KubernetesHelper provides helper methods for Kubernetes operations in tests.
type KubernetesHelper struct {
	t       *testing.T
	cleanup []CleanupFunc
}

// NewKubernetesHelper creates a new Kubernetes helper.
func NewKubernetesHelper(t *testing.T) *KubernetesHelper {
	return &KubernetesHelper{
		t:       t,
		cleanup: []CleanupFunc{},
	}
}

// SkipIfNoCluster skips the test if no cluster is available.
func (k *KubernetesHelper) SkipIfNoCluster() {
	k.t.Helper()
	skipIfNoCluster(k.t)
}

// CreateNamespace creates a namespace and registers cleanup.
func (k *KubernetesHelper) CreateNamespace(name string) error {
	result := RunCommand(k.t, "kubectl", "create", "namespace", name)
	if result.Success() {
		k.cleanup = append(k.cleanup, func() {
			RunCommand(k.t, "kubectl", "delete", "namespace", name, "--ignore-not-found")
		})
	}
	return result.Error
}

// DeleteNamespace deletes a namespace.
func (k *KubernetesHelper) DeleteNamespace(name string) error {
	result := RunCommand(k.t, "kubectl", "delete", "namespace", name, "--ignore-not-found")
	return result.Error
}

// NamespaceExists checks if a namespace exists.
func (k *KubernetesHelper) NamespaceExists(name string) bool {
	result := RunCommand(k.t, "kubectl", "get", "namespace", name)
	return result.Success()
}

// ApplyManifest applies a manifest file.
func (k *KubernetesHelper) ApplyManifest(path string) error {
	result := RunCommand(k.t, "kubectl", "apply", "-f", path)
	return result.Error
}

// ApplyDryRun performs a dry-run apply.
func (k *KubernetesHelper) ApplyDryRun(path string) error {
	result := RunCommand(k.t, "kubectl", "apply", "-f", path, "--dry-run=server")
	return result.Error
}

// Cleanup runs all registered cleanup functions.
func (k *KubernetesHelper) Cleanup() {
	for i := len(k.cleanup) - 1; i >= 0; i-- {
		k.cleanup[i]()
	}
}

// TestReporter collects and reports test results.
type TestReporter struct {
	testName  string
	startTime time.Time
	steps     []testStep
}

type testStep struct {
	name     string
	status   string
	duration time.Duration
	message  string
}

// NewTestReporter creates a new test reporter.
func NewTestReporter(testName string) *TestReporter {
	return &TestReporter{
		testName:  testName,
		startTime: time.Now(),
		steps:     []testStep{},
	}
}

// Step records a test step.
func (r *TestReporter) Step(name, status, message string) {
	r.steps = append(r.steps, testStep{
		name:     name,
		status:   status,
		duration: time.Since(r.startTime),
		message:  message,
	})
}

// Pass records a passing step.
func (r *TestReporter) Pass(name string) {
	r.Step(name, "PASS", "")
}

// Fail records a failing step.
func (r *TestReporter) Fail(name, message string) {
	r.Step(name, "FAIL", message)
}

// Skip records a skipped step.
func (r *TestReporter) Skip(name, reason string) {
	r.Step(name, "SKIP", reason)
}

// Summary returns a summary of the test.
func (r *TestReporter) Summary() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Test: %s\n", r.testName))
	buf.WriteString(fmt.Sprintf("Duration: %v\n", time.Since(r.startTime)))
	buf.WriteString("Steps:\n")
	for _, s := range r.steps {
		buf.WriteString(fmt.Sprintf("  [%s] %s", s.status, s.name))
		if s.message != "" {
			buf.WriteString(fmt.Sprintf(" - %s", s.message))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// JSONReport returns a JSON report of the test.
func (r *TestReporter) JSONReport() ([]byte, error) {
	report := map[string]interface{}{
		"testName": r.testName,
		"duration": time.Since(r.startTime).String(),
		"steps":    r.steps,
	}
	return json.MarshalIndent(report, "", "  ")
}
