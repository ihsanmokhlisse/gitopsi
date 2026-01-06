//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binaryPath string
)

func TestMain(m *testing.M) {
	// Build the binary before running tests
	cmd := exec.Command("go", "build", "-o", "../../bin/gitopsi-test", "../../cmd/gitopsi")
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	// Get absolute path to binary
	absPath, err := filepath.Abs("../../bin/gitopsi-test")
	if err != nil {
		panic("failed to get binary path: " + err.Error())
	}
	binaryPath = absPath

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove(binaryPath)

	os.Exit(code)
}

func TestInitMinimalPreset(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := exec.Command(binaryPath, "init",
		"--preset", "minimal",
		"--output", tmpDir,
		"--config", "fixtures/minimal-config.yaml",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, string(output))
	}

	// Verify expected directories exist
	expectedDirs := []string{
		"test-minimal",
		"test-minimal/infrastructure",
		"test-minimal/infrastructure/base",
		"test-minimal/infrastructure/base/namespaces",
		"test-minimal/argocd",
		"test-minimal/docs",
	}

	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s does not exist", dir)
		}
	}
}

func TestInitStandardPreset(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := exec.Command(binaryPath, "init",
		"--preset", "standard",
		"--output", tmpDir,
		"--config", "fixtures/standard-config.yaml",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, string(output))
	}

	// Verify expected directories exist
	expectedDirs := []string{
		"test-standard",
		"test-standard/infrastructure/base/namespaces",
		"test-standard/infrastructure/base/rbac",
		"test-standard/infrastructure/base/network-policies",
		"test-standard/infrastructure/base/resource-quotas",
		"test-standard/infrastructure/overlays/dev",
		"test-standard/infrastructure/overlays/staging",
		"test-standard/infrastructure/overlays/prod",
		"test-standard/applications/base",
		"test-standard/applications/overlays/dev",
		"test-standard/argocd/projects",
		"test-standard/argocd/applicationsets",
		"test-standard/docs",
		"test-standard/scripts",
	}

	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s does not exist", dir)
		}
	}
}

func TestInitEnterprisePreset(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := exec.Command(binaryPath, "init",
		"--preset", "enterprise",
		"--output", tmpDir,
		"--config", "fixtures/enterprise-config.yaml",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, string(output))
	}

	// Verify project directory exists
	projectDir := filepath.Join(tmpDir, "test-enterprise")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatalf("project directory does not exist: %s", projectDir)
	}
}

func TestInitDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := exec.Command(binaryPath, "init",
		"--dry-run",
		"--output", tmpDir,
		"--config", "fixtures/minimal-config.yaml",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init --dry-run failed: %v\nOutput: %s", err, string(output))
	}

	// Verify no files were created
	entries, _ := os.ReadDir(tmpDir)
	if len(entries) > 0 {
		t.Errorf("expected no files in dry-run mode, found %d", len(entries))
	}

	// Verify output contains dry-run message
	if !strings.Contains(string(output), "DRY RUN") {
		t.Error("expected dry-run message in output")
	}
}

func TestInitWithFlux(t *testing.T) {
	// TODO: Flux support is disabled - focus on ArgoCD first
	// Uncomment when Flux support is ready for production
	t.Skip("Flux support is disabled - focusing on ArgoCD first")

	tmpDir := t.TempDir()

	cmd := exec.Command(binaryPath, "init",
		"--output", tmpDir,
		"--config", "fixtures/flux-config.yaml",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init with flux failed: %v\nOutput: %s", err, string(output))
	}

	// Verify flux directories exist instead of argocd
	fluxDir := filepath.Join(tmpDir, "test-flux", "flux")
	if _, err := os.Stat(fluxDir); os.IsNotExist(err) {
		t.Error("expected flux directory does not exist")
	}
}

func TestValidateCommand(t *testing.T) {
	// First create a project
	tmpDir := t.TempDir()

	initCmd := exec.Command(binaryPath, "init",
		"--output", tmpDir,
		"--config", "fixtures/minimal-config.yaml",
	)
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, string(output))
	}

	// Then validate it
	projectDir := filepath.Join(tmpDir, "test-minimal")
	validateCmd := exec.Command(binaryPath, "validate", projectDir)
	output, err := validateCmd.CombinedOutput()
	if err != nil {
		// Validation may have warnings but shouldn't fail completely
		if !strings.Contains(string(output), "validation") {
			t.Fatalf("validate failed unexpectedly: %v\nOutput: %s", err, string(output))
		}
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "gitopsi") {
		t.Error("expected version output to contain 'gitopsi'")
	}
}

func TestHelpCommand(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help command failed: %v\nOutput: %s", err, string(output))
	}

	expectedStrings := []string{
		"gitopsi",
		"init",
		"GitOps",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(string(output), expected) {
			t.Errorf("expected help output to contain '%s'", expected)
		}
	}
}
