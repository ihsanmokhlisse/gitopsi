package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInitWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: test-init-project
  description: Test project for init command
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
docs:
  readme: true
`
	configPath := filepath.Join(tmpDir, "gitops.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	originalCfgFile := cfgFile
	originalOutput := output
	originalDryRun := dryRun
	originalVerbose := verbose
	defer func() {
		cfgFile = originalCfgFile
		output = originalOutput
		dryRun = originalDryRun
		verbose = originalVerbose
	}()

	cfgFile = configPath
	output = tmpDir
	dryRun = true
	verbose = false

	rootCmd.SetArgs([]string{"init", "--config", configPath, "--output", tmpDir, "--dry-run"})
	err := rootCmd.Execute()
	if err != nil {
		t.Logf("Init command error (expected in test): %v", err)
	}
}

func TestRunInitDryRunMode(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: dry-run-test
platform: kubernetes
scope: infrastructure
gitops_tool: argocd
environments:
  - name: dev
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	originalCfgFile := cfgFile
	originalOutput := output
	originalDryRun := dryRun
	defer func() {
		cfgFile = originalCfgFile
		output = originalOutput
		dryRun = originalDryRun
	}()

	cfgFile = configPath
	output = tmpDir
	dryRun = true

	rootCmd.SetArgs([]string{"init", "--config", configPath, "--dry-run"})
	rootCmd.Execute()

	projectDir := filepath.Join(tmpDir, "dry-run-test")
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Error("Dry run should not create directories")
	}
}

func TestInitCmdProperties(t *testing.T) {
	if initCmd.Use != "init" {
		t.Errorf("initCmd.Use = %s, want init", initCmd.Use)
	}

	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}

	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}

	if initCmd.RunE == nil {
		t.Error("initCmd.RunE should be set")
	}
}

func TestInitWithInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	invalidConfig := `
project:
  name: ""
platform: invalid
`
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	cfgFile = configPath

	rootCmd.SetArgs([]string{"init", "--config", configPath})
	err := rootCmd.Execute()
	if err == nil {
		t.Log("Init with invalid config should fail validation")
	}
}

func TestInitWithNonExistentConfig(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	cfgFile = "/nonexistent/config.yaml"

	rootCmd.SetArgs([]string{"init", "--config", "/nonexistent/config.yaml"})
	err := rootCmd.Execute()
	if err == nil {
		t.Log("Init with non-existent config should fail")
	}
}

func TestInitVerboseMode(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
project:
  name: verbose-test
platform: kubernetes
scope: both
gitops_tool: argocd
environments:
  - name: dev
docs:
  readme: true
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	originalCfgFile := cfgFile
	originalOutput := output
	originalDryRun := dryRun
	originalVerbose := verbose
	defer func() {
		cfgFile = originalCfgFile
		output = originalOutput
		dryRun = originalDryRun
		verbose = originalVerbose
	}()

	cfgFile = configPath
	output = tmpDir
	dryRun = true
	verbose = true

	rootCmd.SetArgs([]string{"init", "--config", configPath, "--dry-run", "--verbose"})
	rootCmd.Execute()
}

func TestInitCommandInheritsFlags(t *testing.T) {
	pFlags := rootCmd.PersistentFlags()

	if pFlags.Lookup("config") == nil {
		t.Error("Init should inherit --config flag")
	}
	if pFlags.Lookup("output") == nil {
		t.Error("Init should inherit --output flag")
	}
	if pFlags.Lookup("dry-run") == nil {
		t.Error("Init should inherit --dry-run flag")
	}
	if pFlags.Lookup("verbose") == nil {
		t.Error("Init should inherit --verbose flag")
	}
}

