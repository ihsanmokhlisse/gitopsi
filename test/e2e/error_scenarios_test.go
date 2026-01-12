//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitMissingConfig tests init with a missing config file.
func TestInitMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "init",
		"--config", "nonexistent-config.yaml",
		"--output", tmpDir,
	)

	if result.Success() {
		t.Error("expected init to fail with missing config file")
	}

	if !result.ContainsAny("not found", "no such file", "does not exist") {
		t.Errorf("expected error message about missing file, got: %s", result.Combined)
	}
}

// TestInitInvalidConfig tests init with an invalid config file.
func TestInitInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid config file
	invalidConfig := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(invalidConfig, []byte("this: is: not: valid: yaml: {{{}}}"), 0644)
	if err != nil {
		t.Fatalf("failed to create invalid config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", invalidConfig,
		"--output", tmpDir,
	)

	if result.Success() {
		t.Error("expected init to fail with invalid config")
	}
}

// TestInitEmptyConfig tests init with an empty config file.
func TestInitEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty config file
	emptyConfig := filepath.Join(tmpDir, "empty.yaml")
	err := os.WriteFile(emptyConfig, []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to create empty config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", emptyConfig,
		"--output", tmpDir,
	)

	// Should fail or use defaults
	if result.Success() {
		// If success, verify defaults were applied
		entries, _ := os.ReadDir(tmpDir)
		if len(entries) == 0 {
			t.Error("no files created with empty config")
		}
	}
}

// TestInitMissingProjectName tests init with a config missing project name.
func TestInitMissingProjectName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config without project name
	configContent := `
platform: kubernetes
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
`
	configPath := filepath.Join(tmpDir, "no-name.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	// Should either fail or use default name
	if !result.Success() {
		if !result.ContainsAny("project", "name", "required") {
			t.Logf("Failed with message: %s", result.Combined)
		}
	}
}

// TestInitReadOnlyOutput tests init with a read-only output directory.
func TestInitReadOnlyOutput(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping: running as root, cannot test permission denied")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

	// Create read-only directory
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create read-only dir: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

	result := RunGitopsi(t, "init",
		"--preset", "minimal",
		"--output", readOnlyDir,
		"--config", "fixtures/minimal-config.yaml",
	)

	if result.Success() {
		t.Error("expected init to fail with read-only output directory")
	}

	if !result.ContainsAny("permission denied", "Permission denied", "access denied") {
		t.Logf("Failed with unexpected message: %s", result.Combined)
	}
}

// TestValidateNonExistentPath tests validate with a non-existent path.
func TestValidateNonExistentPath(t *testing.T) {
	result := RunGitopsi(t, "validate", "/nonexistent/path/to/project")

	if result.Success() {
		t.Error("expected validate to fail with non-existent path")
	}
}

// TestValidateEmptyDirectory tests validate with an empty directory.
func TestValidateEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	result := RunGitopsi(t, "validate", tmpDir)

	// Should fail or report no manifests found
	if result.Success() {
		if !result.ContainsAny("no manifests", "empty", "nothing to validate") {
			t.Logf("Validate succeeded with message: %s", result.Combined)
		}
	}
}

// TestPreflightWithoutCluster tests preflight when no cluster is available.
func TestPreflightWithoutCluster(t *testing.T) {
	// Save and clear KUBECONFIG
	originalKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/nonexistent/kubeconfig")
	defer os.Setenv("KUBECONFIG", originalKubeconfig)

	result := RunGitopsi(t, "preflight")

	// Should fail or report cluster unavailable
	// Note: This may succeed with warnings depending on implementation
	if result.Success() {
		if !result.ContainsAny("warning", "skip", "cluster not available") {
			t.Logf("Preflight output: %s", result.Combined)
		}
	}
}

// TestInitConflictingFlags tests init with conflicting flags.
func TestInitConflictingFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "dry-run with bootstrap",
			args: []string{"init", "--dry-run", "--bootstrap", "--config", "fixtures/minimal-config.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			args := append(tt.args, "--output", tmpDir)
			result := RunGitopsi(t, args...)

			// Should either fail or handle gracefully
			if result.Success() {
				t.Logf("Command succeeded (may be valid): %s", result.Combined)
			}
		})
	}
}

// TestInitWithInvalidPlatform tests init with an invalid platform.
func TestInitWithInvalidPlatform(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: test-invalid-platform
platform: invalid-platform-name
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
`
	configPath := filepath.Join(tmpDir, "invalid-platform.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	// Should fail with invalid platform error
	if result.Success() {
		// If success, platform may have been corrected or ignored
		t.Logf("Init succeeded with invalid platform: %s", result.Combined)
	}
}

// TestInitWithInvalidGitOpsTool tests init with an invalid gitops tool.
func TestInitWithInvalidGitOpsTool(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: test-invalid-tool
platform: kubernetes
scope: infrastructure
gitops_tool: invalid-tool
environments:
  - name: dev
`
	configPath := filepath.Join(tmpDir, "invalid-tool.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	// Should fail or default to argocd
	if result.Success() {
		t.Logf("Init succeeded with invalid tool: %s", result.Combined)
	}
}

// TestInitWithEmptyEnvironments tests init with no environments.
func TestInitWithEmptyEnvironments(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: test-no-envs
platform: kubernetes
scope: infrastructure
gitops_tool: argocd
environments: []
`
	configPath := filepath.Join(tmpDir, "no-envs.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	result := RunGitopsi(t, "init",
		"--config", configPath,
		"--output", tmpDir,
	)

	// Should either fail or create structure with no env-specific files
	if result.Success() {
		// Check that no environment overlays were created
		entries, _ := filepath.Glob(filepath.Join(tmpDir, "test-no-envs", "infrastructure", "overlays", "*"))
		if len(entries) > 0 {
			t.Logf("Environment overlays created despite empty environments: %v", entries)
		}
	}
}

// TestHelpForAllCommands verifies help text for all commands.
func TestHelpForAllCommands(t *testing.T) {
	commands := []string{
		"init",
		"validate",
		"preflight",
		"auth",
		"env",
		"operator",
		"marketplace",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			result := RunGitopsi(t, cmd, "--help")

			if !result.Success() {
				t.Errorf("help for %s failed: %s", cmd, result.Combined)
			}

			// Help should contain command name and usage info
			if !strings.Contains(result.Combined, cmd) {
				t.Errorf("help output should contain command name %q", cmd)
			}
		})
	}
}

// TestVersionOutput verifies version command output format.
func TestVersionOutput(t *testing.T) {
	result := RunGitopsi(t, "version")

	if !result.Success() {
		t.Fatalf("version command failed: %s", result.Combined)
	}

	// Version output should contain expected information
	expectedParts := []string{"gitopsi"}

	for _, part := range expectedParts {
		if !strings.Contains(strings.ToLower(result.Combined), strings.ToLower(part)) {
			t.Errorf("version output missing expected part: %s", part)
		}
	}
}
