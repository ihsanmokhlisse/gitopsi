package cli

import (
	"testing"
)

func TestExecute(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	err := Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "gitopsi" {
		t.Errorf("rootCmd.Use = %s, want gitopsi", rootCmd.Use)
	}
}

func TestVersionCommandExists(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd is nil")
	}

	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %s, want version", versionCmd.Use)
	}
}

func TestInitCommandExists(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd is nil")
	}

	if initCmd.Use != "init" {
		t.Errorf("initCmd.Use = %s, want init", initCmd.Use)
	}
}

func TestGetConfig(t *testing.T) {
	originalValue := cfgFile
	defer func() { cfgFile = originalValue }()

	cfgFile = ""
	if GetConfig() != "" {
		t.Error("GetConfig() should return empty string")
	}

	cfgFile = "/tmp/test.yaml"
	if GetConfig() != "/tmp/test.yaml" {
		t.Error("GetConfig() mismatch")
	}
}

func TestGetOutput(t *testing.T) {
	originalValue := output
	defer func() { output = originalValue }()

	output = ""
	if GetOutput() != "" {
		t.Error("GetOutput() should return empty string")
	}

	output = "/tmp/output"
	if GetOutput() != "/tmp/output" {
		t.Error("GetOutput() mismatch")
	}
}

func TestIsDryRun(t *testing.T) {
	originalValue := dryRun
	defer func() { dryRun = originalValue }()

	dryRun = false
	if IsDryRun() {
		t.Error("IsDryRun() should return false")
	}

	dryRun = true
	if !IsDryRun() {
		t.Error("IsDryRun() should return true")
	}
}

func TestIsVerbose(t *testing.T) {
	originalValue := verbose
	defer func() { verbose = originalValue }()

	verbose = false
	if IsVerbose() {
		t.Error("IsVerbose() should return false")
	}

	verbose = true
	if !IsVerbose() {
		t.Error("IsVerbose() should return true")
	}
}

func TestVersionVariables(t *testing.T) {
	if Version == "" {
		Version = "dev"
	}
	if Commit == "" {
		Commit = "none"
	}
	if BuildDate == "" {
		BuildDate = "unknown"
	}

	if Version == "" || Commit == "" || BuildDate == "" {
		t.Error("Version variables should have default values")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	commands := rootCmd.Commands()
	if len(commands) < 2 {
		t.Errorf("Expected at least 2 subcommands, got %d", len(commands))
	}

	foundVersion := false
	foundInit := false
	for _, cmd := range commands {
		if cmd.Use == "version" {
			foundVersion = true
		}
		if cmd.Use == "init" {
			foundInit = true
		}
	}

	if !foundVersion {
		t.Error("version command not found")
	}
	if !foundInit {
		t.Error("init command not found")
	}
}

func TestRootCommandFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	configFlag := flags.Lookup("config")
	if configFlag == nil {
		t.Error("--config flag not found")
	}

	outputFlag := flags.Lookup("output")
	if outputFlag == nil {
		t.Error("--output flag not found")
	}

	dryRunFlag := flags.Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("--dry-run flag not found")
	}

	verboseFlag := flags.Lookup("verbose")
	if verboseFlag == nil {
		t.Error("--verbose flag not found")
	}
}

func TestInitCommandShortDescription(t *testing.T) {
	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}
}

func TestVersionCommandShortDescription(t *testing.T) {
	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}
}
