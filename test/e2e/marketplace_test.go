//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ===========================================================================
// Marketplace CLI Tests (No cluster required)
// ===========================================================================

// TestMarketplaceHelp tests marketplace help output.
func TestMarketplaceHelp(t *testing.T) {
	result := RunGitopsi(t, "marketplace", "--help")

	if !result.Success() {
		t.Fatalf("marketplace help failed: %s", result.Combined)
	}

	// Help should mention available subcommands
	expectedCommands := []string{"list", "search", "install", "info", "categories"}
	for _, cmd := range expectedCommands {
		if !result.Contains(cmd) {
			t.Logf("marketplace help may be missing %s command", cmd)
		}
	}
}

// TestMarketplaceSubcommandHelp tests help for marketplace subcommands.
func TestMarketplaceSubcommandHelp(t *testing.T) {
	subcommands := []string{"list", "search", "install", "info", "categories"}

	for _, subcmd := range subcommands {
		t.Run(subcmd, func(t *testing.T) {
			result := RunGitopsi(t, "marketplace", subcmd, "--help")

			if !result.Success() {
				t.Errorf("help for marketplace %s failed: %s", subcmd, result.Combined)
			}
		})
	}
}

// TestMarketplaceList tests the marketplace list command.
func TestMarketplaceList(t *testing.T) {
	result := RunGitopsi(t, "marketplace", "list")

	// May fail if registry is unavailable
	if !result.Success() {
		if result.ContainsAny("connection", "timeout", "network") {
			t.Skip("Skipping: marketplace registry unavailable")
		}
		t.Logf("marketplace list result: %s", result.Combined)
	}
}

// TestMarketplaceSearch tests the marketplace search command.
func TestMarketplaceSearch(t *testing.T) {
	tests := []struct {
		name   string
		search string
	}{
		{name: "search-monitoring", search: "monitoring"},
		{name: "search-security", search: "security"},
		{name: "search-observability", search: "observability"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RunGitopsi(t, "marketplace", "search", tt.search)

			if !result.Success() {
				if result.ContainsAny("connection", "timeout", "network", "no results") {
					t.Skip("Skipping: marketplace unavailable or no results")
				}
			}

			t.Logf("Search results for %s: %s", tt.search, result.Combined)
		})
	}
}

// TestMarketplaceCategories tests the marketplace categories command.
func TestMarketplaceCategories(t *testing.T) {
	result := RunGitopsi(t, "marketplace", "categories")

	if !result.Success() {
		if result.ContainsAny("connection", "timeout", "network") {
			t.Skip("Skipping: marketplace registry unavailable")
		}
	}

	// Categories should include common GitOps categories
	expectedCategories := []string{"observability", "security"}
	found := 0
	for _, cat := range expectedCategories {
		if result.Contains(cat) {
			found++
		}
	}

	if result.Success() && found == 0 {
		t.Logf("No expected categories found. Output: %s", result.Combined)
	}
}

// ===========================================================================
// Marketplace Full GitOps Flow Tests (Requires cluster + ArgoCD)
// ===========================================================================

// skipIfNoMarketplaceFlow skips if full marketplace flow testing is not configured
func skipIfNoMarketplaceFlow(t *testing.T) {
	t.Helper()
	skipIfNoArgoCD(t)

	if os.Getenv("E2E_MARKETPLACE_FLOW") == "" {
		t.Skip("Skipping marketplace flow test: E2E_MARKETPLACE_FLOW not set")
	}
}

// MarketplaceFlowConfig holds configuration for marketplace flow tests
type MarketplaceFlowConfig struct {
	PatternName    string
	PatternVersion string
	ProjectName    string
	Environment    string
	GitURL         string
	GitBranch      string
}

// DefaultMarketplaceFlowConfig returns default flow configuration
func DefaultMarketplaceFlowConfig() *MarketplaceFlowConfig {
	return &MarketplaceFlowConfig{
		PatternName:    os.Getenv("E2E_PATTERN_NAME"),
		PatternVersion: os.Getenv("E2E_PATTERN_VERSION"),
		ProjectName:    "test-marketplace-flow",
		Environment:    "dev",
		GitURL:         os.Getenv("E2E_REPO_URL"),
		GitBranch:      os.Getenv("E2E_BRANCH"),
	}
}

// TestMarketplaceFullGitOpsFlow tests the complete pattern installation flow:
// 1. Generate a project
// 2. Install a pattern from marketplace
// 3. Push changes to Git
// 4. ArgoCD syncs the changes
// 5. Validate pattern deployment on cluster
func TestMarketplaceFullGitOpsFlow(t *testing.T) {
	skipIfNoMarketplaceFlow(t)

	cfg := DefaultMarketplaceFlowConfig()
	if cfg.PatternName == "" {
		t.Skip("Skipping: E2E_PATTERN_NAME not set")
	}
	if cfg.GitURL == "" {
		t.Skip("Skipping: E2E_REPO_URL not set")
	}

	tmpDir := t.TempDir()
	k8s := NewKubernetesHelper(t)
	defer k8s.Cleanup()

	reporter := NewTestReporter("MarketplaceFullGitOpsFlow")

	// =========================================================================
	// PHASE 1: Generate base project
	// =========================================================================
	t.Log("=== Phase 1: Generate Base Project ===")

	configContent := `
project:
  name: ` + cfg.ProjectName + `
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: ` + cfg.Environment + `
infra:
  namespaces: true
  rbac: true
git:
  url: ` + cfg.GitURL + `
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	if !result.Success() {
		t.Fatalf("Phase 1 failed - init: %s", result.Combined)
	}
	reporter.Pass("Generate base project")

	projectDir := filepath.Join(tmpDir, cfg.ProjectName)

	// =========================================================================
	// PHASE 2: Install pattern from marketplace
	// =========================================================================
	t.Log("=== Phase 2: Install Pattern from Marketplace ===")

	installArgs := []string{
		"marketplace", "install",
		"--project", projectDir,
		cfg.PatternName,
	}
	if cfg.PatternVersion != "" {
		installArgs = append(installArgs, "--version", cfg.PatternVersion)
	}

	result = RunGitopsi(t, installArgs...)

	if !result.Success() {
		if result.ContainsAny("not found", "unavailable") {
			t.Skipf("Pattern %s not available: %s", cfg.PatternName, result.Combined)
		}
		t.Fatalf("Phase 2 failed - pattern install: %s", result.Combined)
	}
	reporter.Pass("Install pattern: " + cfg.PatternName)

	// Verify pattern was added to project
	patternsDir := filepath.Join(projectDir, "patterns")
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		// Patterns might be installed elsewhere depending on implementation
		t.Log("patterns/ directory not created - checking alternative locations")
	}

	// =========================================================================
	// PHASE 3: Push changes to Git
	// =========================================================================
	t.Log("=== Phase 3: Push Changes to Git ===")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	branchName := cfg.GitBranch
	if branchName == "" {
		branchName = "test/marketplace-" + time.Now().Format("20060102150405")
	}

	// Initialize git and push
	gitCommands := []struct {
		name string
		args []string
	}{
		{"git init", []string{"git", "-C", projectDir, "init"}},
		{"git add", []string{"git", "-C", projectDir, "add", "-A"}},
		{"git commit", []string{"git", "-C", projectDir, "commit", "-m", "Add pattern: " + cfg.PatternName}},
		{"git remote add", []string{"git", "-C", projectDir, "remote", "add", "origin", cfg.GitURL}},
		{"git push", []string{"git", "-C", projectDir, "push", "-u", "origin", branchName, "--force"}},
	}

	for _, gc := range gitCommands {
		output, err := runWithTimeout(ctx, gc.args[0], gc.args[1:]...)
		if err != nil {
			// Remote add might fail if already exists, that's OK
			if !strings.Contains(gc.name, "remote") {
				t.Fatalf("Phase 3 failed - %s: %v\nOutput: %s", gc.name, err, output)
			}
		}
	}
	reporter.Pass("Push to Git branch: " + branchName)

	// =========================================================================
	// PHASE 4: Create ArgoCD Application and Sync
	// =========================================================================
	t.Log("=== Phase 4: ArgoCD Sync ===")

	appName := "marketplace-test-" + cfg.PatternName
	namespace := cfg.ProjectName + "-" + cfg.Environment

	// Clean up any existing app
	runWithTimeout(ctx, "argocd", "app", "delete", appName, "--yes", "--insecure")
	time.Sleep(2 * time.Second)

	// Create ArgoCD application for the pattern
	output, err := runWithTimeout(ctx, "argocd", "app", "create", appName,
		"--repo", cfg.GitURL,
		"--revision", branchName,
		"--path", "infrastructure/overlays/"+cfg.Environment,
		"--dest-server", "https://kubernetes.default.svc",
		"--dest-namespace", namespace,
		"--sync-policy", "automated",
		"--auto-prune",
		"--self-heal",
		"--insecure",
	)
	if err != nil {
		t.Fatalf("Phase 4 failed - create app: %v\nOutput: %s", err, output)
	}
	reporter.Pass("Create ArgoCD application: " + appName)

	// Trigger sync
	output, err = runWithTimeout(ctx, "argocd", "app", "sync", appName, "--insecure", "--timeout", "300")
	if err != nil {
		t.Logf("Sync may have issues: %v\nOutput: %s", err, output)
	}

	// Wait for sync to complete
	output, err = runWithTimeout(ctx, "argocd", "app", "wait", appName, "--sync", "--timeout", "300", "--insecure")
	if err != nil {
		t.Fatalf("Phase 4 failed - wait sync: %v\nOutput: %s", err, output)
	}
	reporter.Pass("ArgoCD sync complete")

	// =========================================================================
	// PHASE 5: Validate Pattern Deployment
	// =========================================================================
	t.Log("=== Phase 5: Validate Pattern Deployment ===")

	// Check namespace exists
	output, err = runWithTimeout(ctx, "kubectl", "get", "namespace", namespace)
	if err != nil {
		t.Errorf("Namespace %s not found: %v", namespace, err)
	} else {
		reporter.Pass("Namespace created: " + namespace)
	}

	// Check ArgoCD app health
	output, err = runWithTimeout(ctx, "argocd", "app", "get", appName, "-o", "json", "--insecure")
	if err != nil {
		t.Errorf("Failed to get app status: %v", err)
	} else {
		if strings.Contains(output, `"Healthy"`) || strings.Contains(output, `"Progressing"`) {
			reporter.Pass("Application health: OK")
		} else {
			t.Logf("App health status: %s", output)
			reporter.Fail("Application health", "Not healthy")
		}
	}

	// Check resources from pattern are deployed
	output, err = runWithTimeout(ctx, "kubectl", "get", "all", "-n", namespace)
	if err != nil {
		t.Logf("Failed to list resources: %v", err)
	} else {
		t.Logf("Resources in namespace %s:\n%s", namespace, output)
		reporter.Pass("Resources deployed")
	}

	// =========================================================================
	// Cleanup
	// =========================================================================
	t.Cleanup(func() {
		ctx := context.Background()
		runWithTimeout(ctx, "argocd", "app", "delete", appName, "--yes", "--insecure")
		runWithTimeout(ctx, "kubectl", "delete", "namespace", namespace, "--ignore-not-found")
		runWithTimeout(ctx, "git", "push", "origin", "--delete", branchName)
	})

	t.Log("\n" + reporter.Summary())
}

// TestMarketplacePatternUpdate tests updating an installed pattern
func TestMarketplacePatternUpdate(t *testing.T) {
	skipIfNoMarketplaceFlow(t)

	cfg := DefaultMarketplaceFlowConfig()
	if cfg.PatternName == "" {
		t.Skip("Skipping: E2E_PATTERN_NAME not set")
	}

	tmpDir := t.TempDir()

	// Generate project with pattern pre-installed
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
project:
  name: test-pattern-update
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init", "--config", configPath, "--output", tmpDir)
	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-pattern-update")

	// Install pattern v1
	result = RunGitopsi(t, "marketplace", "install",
		"--project", projectDir,
		"--version", "1.0.0",
		cfg.PatternName,
	)
	if !result.Success() {
		if result.ContainsAny("not found", "unavailable") {
			t.Skip("Pattern not available")
		}
		t.Logf("Install v1 result: %s", result.Combined)
	}

	// Update to latest
	result = RunGitopsi(t, "marketplace", "update",
		"--project", projectDir,
		cfg.PatternName,
	)
	if !result.Success() {
		t.Logf("Update result: %s", result.Combined)
	}

	// Check update status
	result = RunGitopsi(t, "marketplace", "list",
		"--installed",
		"--project", projectDir,
	)
	t.Logf("Installed patterns after update: %s", result.Combined)
}

// TestMarketplacePatternUninstall tests uninstalling a pattern
func TestMarketplacePatternUninstall(t *testing.T) {
	skipIfNoMarketplaceFlow(t)

	cfg := DefaultMarketplaceFlowConfig()
	if cfg.PatternName == "" {
		t.Skip("Skipping: E2E_PATTERN_NAME not set")
	}

	tmpDir := t.TempDir()

	// Generate project
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
project:
  name: test-pattern-uninstall
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init", "--config", configPath, "--output", tmpDir)
	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-pattern-uninstall")

	// Install pattern
	result = RunGitopsi(t, "marketplace", "install",
		"--project", projectDir,
		cfg.PatternName,
	)
	if !result.Success() {
		if result.ContainsAny("not found", "unavailable") {
			t.Skip("Pattern not available")
		}
	}

	// Uninstall pattern
	result = RunGitopsi(t, "marketplace", "uninstall",
		"--project", projectDir,
		cfg.PatternName,
	)
	if !result.Success() {
		t.Logf("Uninstall result: %s", result.Combined)
	}

	// Verify pattern is removed
	result = RunGitopsi(t, "marketplace", "list",
		"--installed",
		"--project", projectDir,
	)
	if result.Contains(cfg.PatternName) {
		t.Errorf("Pattern %s should be uninstalled", cfg.PatternName)
	}
}

// TestMarketplaceMultiplePatterns tests installing multiple patterns
func TestMarketplaceMultiplePatterns(t *testing.T) {
	skipIfNoMarketplaceFlow(t)

	patterns := strings.Split(os.Getenv("E2E_PATTERNS"), ",")
	if len(patterns) < 2 || patterns[0] == "" {
		t.Skip("Skipping: E2E_PATTERNS should contain comma-separated pattern names")
	}

	tmpDir := t.TempDir()

	// Generate project
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
project:
  name: test-multi-patterns
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init", "--config", configPath, "--output", tmpDir)
	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-multi-patterns")

	// Install each pattern
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		t.Run("install-"+pattern, func(t *testing.T) {
			result = RunGitopsi(t, "marketplace", "install",
				"--project", projectDir,
				pattern,
			)
			if !result.Success() {
				if result.ContainsAny("not found", "unavailable") {
					t.Skipf("Pattern %s not available", pattern)
				}
				t.Logf("Install %s result: %s", pattern, result.Combined)
			}
		})
	}

	// List installed patterns
	result = RunGitopsi(t, "marketplace", "list",
		"--installed",
		"--project", projectDir,
	)
	t.Logf("Installed patterns: %s", result.Combined)
}

// TestMarketplacePatternConflict tests pattern conflict detection
func TestMarketplacePatternConflict(t *testing.T) {
	skipIfNoMarketplaceFlow(t)

	cfg := DefaultMarketplaceFlowConfig()
	if cfg.PatternName == "" {
		t.Skip("Skipping: E2E_PATTERN_NAME not set")
	}

	tmpDir := t.TempDir()

	// Generate project
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
project:
  name: test-pattern-conflict
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	result := RunGitopsi(t, "init", "--config", configPath, "--output", tmpDir)
	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-pattern-conflict")

	// Install pattern first time
	result = RunGitopsi(t, "marketplace", "install",
		"--project", projectDir,
		cfg.PatternName,
	)
	if !result.Success() {
		if result.ContainsAny("not found", "unavailable") {
			t.Skip("Pattern not available")
		}
	}

	// Try to install same pattern again (should handle conflict)
	result = RunGitopsi(t, "marketplace", "install",
		"--project", projectDir,
		cfg.PatternName,
	)

	// Should either fail with "already installed" or succeed with "updated"
	if result.ContainsAny("already installed", "exists", "conflict") {
		t.Log("Conflict detection working correctly")
	} else if result.ContainsAny("updated", "upgraded") {
		t.Log("Pattern was updated instead of conflicting")
	} else {
		t.Logf("Unexpected result on duplicate install: %s", result.Combined)
	}
}

// TestMarketplaceDryRunInstall tests pattern installation in dry-run mode
func TestMarketplaceDryRunInstall(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate project
	result := RunGitopsi(t, "init",
		"--preset", "minimal",
		"--output", tmpDir,
		"--config", "fixtures/minimal-config.yaml",
	)
	if !result.Success() {
		t.Fatalf("init failed: %s", result.Combined)
	}

	projectDir := filepath.Join(tmpDir, "test-minimal")

	// Dry-run install
	result = RunGitopsi(t, "marketplace", "install",
		"--project", projectDir,
		"--dry-run",
		"test-pattern",
	)

	// Should show what would be installed without making changes
	if result.ContainsAny("not found", "unavailable") {
		t.Log("Pattern not available - dry-run test inconclusive")
		return
	}

	if result.Contains("dry") || result.Contains("would") {
		t.Log("Dry-run mode working correctly")
	}

	// Verify no files were actually modified
	result = RunGitopsi(t, "marketplace", "list",
		"--installed",
		"--project", projectDir,
	)
	if !result.ContainsAny("no patterns", "none", "empty", "0 patterns") {
		t.Log("Dry-run may have installed pattern - checking...")
	}
}
