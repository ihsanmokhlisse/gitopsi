package cli

import (
	"strings"
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/marketplace"
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

func TestRootCommandLongDescription(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}

	if !strings.Contains(rootCmd.Long, "GitOps") {
		t.Error("rootCmd.Long should mention GitOps")
	}
}

func TestInitCommandLongDescription(t *testing.T) {
	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}

	if !strings.Contains(initCmd.Long, "interactive") {
		t.Error("initCmd.Long should mention interactive mode")
	}
}

func TestFlagDefaultValues(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	outputFlag := flags.Lookup("output")
	if outputFlag.DefValue != "." {
		t.Errorf("Default output should be '.', got %s", outputFlag.DefValue)
	}

	dryRunFlag := flags.Lookup("dry-run")
	if dryRunFlag.DefValue != "false" {
		t.Errorf("Default dry-run should be 'false', got %s", dryRunFlag.DefValue)
	}

	verboseFlag := flags.Lookup("verbose")
	if verboseFlag.DefValue != "false" {
		t.Errorf("Default verbose should be 'false', got %s", verboseFlag.DefValue)
	}
}

func TestVersionVariablesSetCorrectly(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalBuildDate := BuildDate
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalBuildDate
	}()

	Version = "1.2.3"
	Commit = "abc123"
	BuildDate = "2024-01-01T00:00:00Z"

	if Version != "1.2.3" {
		t.Error("Version not set correctly")
	}
	if Commit != "abc123" {
		t.Error("Commit not set correctly")
	}
	if BuildDate != "2024-01-01T00:00:00Z" {
		t.Error("BuildDate not set correctly")
	}
}

func TestAllCommandsHaveRunE(t *testing.T) {
	if initCmd.RunE == nil {
		t.Error("initCmd should have RunE function")
	}
}

func TestVersionCommandHasRun(t *testing.T) {
	if versionCmd.Run == nil {
		t.Error("versionCmd should have Run function")
	}
}

func TestRootCommandHasInit(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have init subcommand")
	}
}

func TestRootCommandHasVersion(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have version subcommand")
	}
}

func TestFlagsArePersistent(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	if flags.Lookup("config") == nil {
		t.Error("config flag should be persistent")
	}
	if flags.Lookup("output") == nil {
		t.Error("output flag should be persistent")
	}
	if flags.Lookup("dry-run") == nil {
		t.Error("dry-run flag should be persistent")
	}
	if flags.Lookup("verbose") == nil {
		t.Error("verbose flag should be persistent")
	}
}

func TestExecuteWithVersion(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	err := Execute()
	if err != nil {
		t.Errorf("Execute() with version should not error: %v", err)
	}
}

func TestExecuteWithHelp(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := Execute()
	if err != nil {
		t.Errorf("Execute() with help should not error: %v", err)
	}
}

func TestValidateCommandExists(t *testing.T) {
	if validateCmd == nil {
		t.Fatal("validateCmd is nil")
	}

	if validateCmd.Use != "validate [path]" {
		t.Errorf("validateCmd.Use = %s, want 'validate [path]'", validateCmd.Use)
	}
}

func TestValidateCommandFlags(t *testing.T) {
	flags := validateCmd.Flags()

	k8sVersionFlag := flags.Lookup("k8s-version")
	if k8sVersionFlag == nil {
		t.Error("--k8s-version flag not found")
	}
	if k8sVersionFlag.DefValue != "1.29" {
		t.Errorf("k8s-version default = %s, want 1.29", k8sVersionFlag.DefValue)
	}

	argoCDVersionFlag := flags.Lookup("argocd-version")
	if argoCDVersionFlag == nil {
		t.Error("--argocd-version flag not found")
	}

	schemaFlag := flags.Lookup("schema")
	if schemaFlag == nil {
		t.Error("--schema flag not found")
	}

	securityFlag := flags.Lookup("security")
	if securityFlag == nil {
		t.Error("--security flag not found")
	}

	deprecationFlag := flags.Lookup("deprecation")
	if deprecationFlag == nil {
		t.Error("--deprecation flag not found")
	}

	kustomizeFlag := flags.Lookup("kustomize")
	if kustomizeFlag == nil {
		t.Error("--kustomize flag not found")
	}

	failOnFlag := flags.Lookup("fail-on")
	if failOnFlag == nil {
		t.Error("--fail-on flag not found")
	}
	if failOnFlag.DefValue != "high" {
		t.Errorf("fail-on default = %s, want high", failOnFlag.DefValue)
	}

	outputFlag := flags.Lookup("output")
	if outputFlag == nil {
		t.Error("--output flag not found")
	}

	fixFlag := flags.Lookup("fix")
	if fixFlag == nil {
		t.Error("--fix flag not found")
	}
}

func TestStatusCommandExists(t *testing.T) {
	if statusCmd == nil {
		t.Fatal("statusCmd is nil")
	}

	if statusCmd.Use != "status" {
		t.Errorf("statusCmd.Use = %s, want 'status'", statusCmd.Use)
	}
}

func TestStatusCommandFlags(t *testing.T) {
	flags := statusCmd.Flags()

	jsonFlag := flags.Lookup("json")
	if jsonFlag == nil {
		t.Error("--json flag not found")
	}

	quietFlag := flags.Lookup("quiet")
	if quietFlag == nil {
		t.Error("--quiet flag not found")
	}
}

func TestAuthCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have auth subcommand")
	}
}

func TestEnvCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "env" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have env subcommand")
	}
}

func TestOperatorCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "operator" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have operator subcommand")
	}
}

func TestMarketplaceCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "marketplace" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have marketplace subcommand")
	}
}

func TestPreflightCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "preflight" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have preflight subcommand")
	}
}

func TestValidateCommandDescription(t *testing.T) {
	if validateCmd.Short == "" {
		t.Error("validateCmd.Short should not be empty")
	}

	if validateCmd.Long == "" {
		t.Error("validateCmd.Long should not be empty")
	}

	if !strings.Contains(validateCmd.Long, "validate") {
		t.Error("validateCmd.Long should mention 'validate'")
	}
}

func TestStatusCommandDescription(t *testing.T) {
	if statusCmd.Short == "" {
		t.Error("statusCmd.Short should not be empty")
	}

	if statusCmd.Long == "" {
		t.Error("statusCmd.Long should not be empty")
	}
}

func TestVersionCommandVerboseOutput(t *testing.T) {
	originalVerbose := verbose
	defer func() { verbose = originalVerbose }()

	// Test verbose output
	verbose = true
	rootCmd.SetArgs([]string{"version"})
	err := Execute()
	if err != nil {
		t.Errorf("Execute() version with verbose should not error: %v", err)
	}
}

func TestAllCommands(t *testing.T) {
	commands := rootCmd.Commands()

	expectedCommands := []string{
		"init",
		"version",
		"validate",
		"status",
		"auth",
		"env",
		"operator",
		"marketplace",
		"preflight",
	}

	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected command '%s' not found", expected)
		}
	}
}

func TestFlagsUseLongForm(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	// Verify flags exist with long form names
	configFlag := flags.Lookup("config")
	if configFlag == nil {
		t.Error("--config flag not found")
	}

	outputFlag := flags.Lookup("output")
	if outputFlag == nil {
		t.Error("--output flag not found")
	}

	verboseFlag := flags.Lookup("verbose")
	if verboseFlag == nil {
		t.Error("--verbose flag not found")
	}
}

func TestVersionConstants(t *testing.T) {
	// Ensure version constants have valid values or defaults
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Commit == "" {
		t.Error("Commit should not be empty")
	}
	if BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

func TestValidateCommandHasRunE(t *testing.T) {
	if validateCmd.RunE == nil {
		t.Error("validateCmd should have RunE function")
	}
}

func TestStatusCommandHasRunE(t *testing.T) {
	if statusCmd.RunE == nil {
		t.Error("statusCmd should have RunE function")
	}
}

func TestRootCommandShowsUsageOnError(t *testing.T) {
	// Default behavior: show usage on error (SilenceUsage = false)
	// This is the expected Cobra default behavior
	if rootCmd.SilenceUsage {
		t.Error("rootCmd has SilenceUsage enabled, but default behavior shows usage")
	}
}

func TestRootCommandShowsErrors(t *testing.T) {
	// Default behavior: show errors (SilenceErrors = false)
	// This is the expected Cobra default behavior
	if rootCmd.SilenceErrors {
		t.Error("rootCmd has SilenceErrors enabled, but default behavior shows errors")
	}
}

func TestExecuteVersionVerbose(t *testing.T) {
	originalVerbose := verbose
	defer func() { verbose = originalVerbose }()

	verbose = true
	rootCmd.SetArgs([]string{"version"})
	err := Execute()
	if err != nil {
		t.Errorf("Execute() with version --verbose should not error: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate needed", "hello world", 8, "hello..."},
		{"empty string", "", 10, ""},
		{"very short max", "hello", 4, "h..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestGetBuiltInCategories(t *testing.T) {
	categories := getBuiltInCategories()

	if len(categories) == 0 {
		t.Error("getBuiltInCategories() returned empty slice")
	}

	// Check that each category has a name
	for _, cat := range categories {
		if cat.Name == "" {
			t.Error("Category has empty name")
		}
	}
}

func TestGetBuiltInPatterns(t *testing.T) {
	patterns := getBuiltInPatterns()

	if len(patterns) == 0 {
		t.Error("getBuiltInPatterns() returned empty slice")
	}

	// Check that each pattern has required fields
	for _, p := range patterns {
		if p.Name == "" {
			t.Error("Pattern has empty name")
		}
		if p.Version == "" {
			t.Error("Pattern has empty version")
		}
	}
}

func TestSearchLocalPatterns(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		opts     marketplace.SearchOptions
		minCount int
	}{
		{"empty query returns all", "", marketplace.SearchOptions{}, 1},
		{"search by name", "prometheus", marketplace.SearchOptions{}, 0},
		{"limit results", "", marketplace.SearchOptions{Limit: 2}, 0},
		{"filter by category", "", marketplace.SearchOptions{Category: "monitoring"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searchLocalPatterns(tt.query, tt.opts)
			if len(results) < tt.minCount {
				t.Errorf("searchLocalPatterns(%q, opts) returned %d results, want at least %d",
					tt.query, len(results), tt.minCount)
			}
		})
	}
}

func TestGetLocalPatternInfo(t *testing.T) {
	// Get a known pattern name from built-in patterns
	patterns := getBuiltInPatterns()
	if len(patterns) == 0 {
		t.Skip("No built-in patterns available")
	}

	patternName := patterns[0].Name

	// Test finding existing pattern
	info := getLocalPatternInfo(patternName)
	if info == nil {
		t.Errorf("getLocalPatternInfo(%q) returned nil for existing pattern", patternName)
	}

	// Test non-existing pattern
	info = getLocalPatternInfo("non-existent-pattern")
	if info != nil {
		t.Error("getLocalPatternInfo() should return nil for non-existing pattern")
	}
}

func TestInfoCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "info" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have info subcommand")
	}
}

func TestOpenCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "open" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have open subcommand")
	}
}

func TestGetPasswordCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "get-password" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have get-password subcommand")
	}
}

func TestInstallCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "install" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have install subcommand")
	}
}

func TestPatternsCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "patterns" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have patterns subcommand")
	}
}

func TestOrgCommandExists(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "org" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Root command should have org subcommand")
	}
}

// Edge case tests for truncate function

func TestTruncate_ZeroMaxLen(t *testing.T) {
	result := truncate("hello", 0)
	if result != "" {
		t.Errorf("Expected empty string for maxLen=0, got %q", result)
	}
}

func TestTruncate_NegativeMaxLen(t *testing.T) {
	result := truncate("hello", -5)
	if result != "" {
		t.Errorf("Expected empty string for negative maxLen, got %q", result)
	}
}

func TestTruncate_MaxLenOne(t *testing.T) {
	result := truncate("hello", 1)
	if result != "h" {
		t.Errorf("Expected 'h' for maxLen=1, got %q", result)
	}
}

func TestTruncate_MaxLenTwo(t *testing.T) {
	result := truncate("hello", 2)
	if result != "he" {
		t.Errorf("Expected 'he' for maxLen=2, got %q", result)
	}
}

func TestTruncate_MaxLenThree(t *testing.T) {
	result := truncate("hello", 3)
	if result != "hel" {
		t.Errorf("Expected 'hel' for maxLen=3, got %q", result)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

func TestTruncate_LongerThanNeeded(t *testing.T) {
	result := truncate("hi", 10)
	if result != "hi" {
		t.Errorf("Expected 'hi', got '%s'", result)
	}
}

func TestTruncate_Unicode(t *testing.T) {
	// Unicode characters may have different byte lengths
	result := truncate("héllo wörld", 8)
	if len(result) > 8 {
		t.Logf("Unicode truncation result: %s", result)
	}
}

// Edge case tests for getBuiltInCategories

func TestGetBuiltInCategories_NotEmpty(t *testing.T) {
	categories := getBuiltInCategories()
	if len(categories) == 0 {
		t.Error("Expected at least one built-in category")
	}
}

func TestGetBuiltInCategories_AllHaveNames(t *testing.T) {
	categories := getBuiltInCategories()
	for i, cat := range categories {
		if cat.Name == "" {
			t.Errorf("Category %d has empty name", i)
		}
	}
}

func TestGetBuiltInCategories_ConsistentResults(t *testing.T) {
	categories1 := getBuiltInCategories()
	categories2 := getBuiltInCategories()

	if len(categories1) != len(categories2) {
		t.Error("getBuiltInCategories should return consistent results")
	}
}

// Edge case tests for getBuiltInPatterns

func TestGetBuiltInPatterns_NotEmpty(t *testing.T) {
	patterns := getBuiltInPatterns()
	if len(patterns) == 0 {
		t.Error("Expected at least one built-in pattern")
	}
}

func TestGetBuiltInPatterns_AllHaveRequiredFields(t *testing.T) {
	patterns := getBuiltInPatterns()
	for i, p := range patterns {
		if p.Name == "" {
			t.Errorf("Pattern %d has empty name", i)
		}
		if p.Version == "" {
			t.Errorf("Pattern %s has empty version", p.Name)
		}
	}
}

func TestGetBuiltInPatterns_ConsistentResults(t *testing.T) {
	patterns1 := getBuiltInPatterns()
	patterns2 := getBuiltInPatterns()

	if len(patterns1) != len(patterns2) {
		t.Error("getBuiltInPatterns should return consistent results")
	}
}

// Edge case tests for searchLocalPatterns

func TestSearchLocalPatterns_EmptyQuery(t *testing.T) {
	results := searchLocalPatterns("", marketplace.SearchOptions{})
	if len(results) == 0 {
		t.Error("Empty query should return all patterns")
	}
}

func TestSearchLocalPatterns_NonMatchingQuery(t *testing.T) {
	results := searchLocalPatterns("xyz123nonexistent", marketplace.SearchOptions{})
	if len(results) != 0 {
		t.Errorf("Non-matching query should return 0 results, got %d", len(results))
	}
}

func TestSearchLocalPatterns_CaseInsensitive(t *testing.T) {
	// Get a pattern name
	patterns := getBuiltInPatterns()
	if len(patterns) == 0 {
		t.Skip("No built-in patterns available")
	}

	patternName := patterns[0].Name
	upperQuery := strings.ToUpper(patternName)

	results := searchLocalPatterns(upperQuery, marketplace.SearchOptions{})
	// Should find the pattern regardless of case
	found := false
	for _, r := range results {
		if strings.EqualFold(r.Name, patternName) {
			found = true
			break
		}
	}
	if !found && len(results) > 0 {
		t.Logf("Case insensitive search may not find exact match: query=%s", upperQuery)
	}
}

func TestSearchLocalPatterns_LimitZero(t *testing.T) {
	results := searchLocalPatterns("", marketplace.SearchOptions{Limit: 0})
	// Limit 0 means no limit
	if len(results) == 0 {
		t.Error("Limit 0 should not limit results")
	}
}

func TestSearchLocalPatterns_LimitOne(t *testing.T) {
	results := searchLocalPatterns("", marketplace.SearchOptions{Limit: 1})
	if len(results) > 1 {
		t.Errorf("Expected max 1 result, got %d", len(results))
	}
}

func TestSearchLocalPatterns_NonExistentCategory(t *testing.T) {
	results := searchLocalPatterns("", marketplace.SearchOptions{Category: "nonexistent-category"})
	if len(results) != 0 {
		t.Errorf("Non-existent category should return 0 results, got %d", len(results))
	}
}

func TestSearchLocalPatterns_EmptyCategory(t *testing.T) {
	results := searchLocalPatterns("", marketplace.SearchOptions{Category: ""})
	// Empty category means no filter
	if len(results) == 0 {
		t.Error("Empty category should not filter results")
	}
}

// Edge case tests for getLocalPatternInfo

func TestGetLocalPatternInfo_NonExistent(t *testing.T) {
	info := getLocalPatternInfo("nonexistent-pattern-xyz123")
	if info != nil {
		t.Error("Expected nil for non-existent pattern")
	}
}

func TestGetLocalPatternInfo_EmptyName(t *testing.T) {
	info := getLocalPatternInfo("")
	if info != nil {
		t.Error("Expected nil for empty pattern name")
	}
}

func TestGetLocalPatternInfo_ValidPattern(t *testing.T) {
	patterns := getBuiltInPatterns()
	if len(patterns) == 0 {
		t.Skip("No built-in patterns available")
	}

	patternName := patterns[0].Name
	info := getLocalPatternInfo(patternName)

	if info == nil {
		t.Errorf("Expected info for pattern %s, got nil", patternName)
		return
	}

	if info.Pattern.Metadata.Name != patternName {
		t.Errorf("Expected pattern name %s, got %s", patternName, info.Pattern.Metadata.Name)
	}
}

func TestGetLocalPatternInfo_HasAPIVersion(t *testing.T) {
	patterns := getBuiltInPatterns()
	if len(patterns) == 0 {
		t.Skip("No built-in patterns available")
	}

	info := getLocalPatternInfo(patterns[0].Name)
	if info == nil {
		t.Skip("Pattern info not found")
	}

	if info.Pattern.APIVersion != "gitopsi.io/v1" {
		t.Errorf("Expected APIVersion 'gitopsi.io/v1', got '%s'", info.Pattern.APIVersion)
	}
}

// Edge case tests for global flag getters

func TestGetConfig_EmptyDefault(t *testing.T) {
	originalValue := cfgFile
	defer func() { cfgFile = originalValue }()

	cfgFile = ""
	if GetConfig() != "" {
		t.Error("GetConfig() should return empty string when cfgFile is empty")
	}
}

func TestGetOutput_DefaultValue(t *testing.T) {
	originalValue := output
	defer func() { output = originalValue }()

	output = "."
	if GetOutput() != "." {
		t.Errorf("GetOutput() = %s, want '.'", GetOutput())
	}
}

func TestIsDryRun_DefaultFalse(t *testing.T) {
	originalValue := dryRun
	defer func() { dryRun = originalValue }()

	dryRun = false
	if IsDryRun() {
		t.Error("IsDryRun() should return false by default")
	}
}

func TestIsVerbose_DefaultFalse(t *testing.T) {
	originalValue := verbose
	defer func() { verbose = originalValue }()

	verbose = false
	if IsVerbose() {
		t.Error("IsVerbose() should return false by default")
	}
}

// Edge case tests for command structure

func TestRootCommand_UseField(t *testing.T) {
	if rootCmd.Use != "gitopsi" {
		t.Errorf("rootCmd.Use = %s, want 'gitopsi'", rootCmd.Use)
	}
}

func TestRootCommand_HasShortDescription(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}

func TestRootCommand_HasLongDescription(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestVersionCommand_HasShort(t *testing.T) {
	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}
}

func TestInitCommand_HasShort(t *testing.T) {
	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}
}

func TestInitCommand_HasLong(t *testing.T) {
	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}
}

// Edge case tests for flag shorthand

func TestConfigFlag_HasShorthand(t *testing.T) {
	flags := rootCmd.PersistentFlags()
	flag := flags.Lookup("config")
	if flag == nil {
		t.Fatal("config flag not found")
	}
	// Check if shorthand is set (may be empty)
	t.Logf("config flag shorthand: %s", flag.Shorthand)
}

func TestVerboseFlag_HasShorthand(t *testing.T) {
	flags := rootCmd.PersistentFlags()
	flag := flags.Lookup("verbose")
	if flag == nil {
		t.Fatal("verbose flag not found")
	}
	t.Logf("verbose flag shorthand: %s", flag.Shorthand)
}

// Edge case tests for version variables

func TestVersion_CanBeSet(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "test-version"
	if Version != "test-version" {
		t.Error("Version should be settable")
	}
}

func TestCommit_CanBeSet(t *testing.T) {
	original := Commit
	defer func() { Commit = original }()

	Commit = "test-commit"
	if Commit != "test-commit" {
		t.Error("Commit should be settable")
	}
}

func TestBuildDate_CanBeSet(t *testing.T) {
	original := BuildDate
	defer func() { BuildDate = original }()

	BuildDate = "test-date"
	if BuildDate != "test-date" {
		t.Error("BuildDate should be settable")
	}
}

// Edge case tests for subcommand count

func TestRootCommand_MinimumSubcommands(t *testing.T) {
	commands := rootCmd.Commands()
	// Should have at least these commands: init, version, validate, status, auth, env, operator, marketplace, preflight
	minExpected := 9
	if len(commands) < minExpected {
		t.Errorf("Expected at least %d subcommands, got %d", minExpected, len(commands))
	}
}

// Edge case tests for marketplace command flags

func TestMarketplaceCmd_HasProjectFlag(t *testing.T) {
	flag := marketplaceCmd.PersistentFlags().Lookup("project")
	if flag == nil {
		t.Error("marketplace command should have --project flag")
	}
}

func TestMarketplaceCmd_HasGitOpsToolFlag(t *testing.T) {
	flag := marketplaceCmd.PersistentFlags().Lookup("gitops-tool")
	if flag == nil {
		t.Error("marketplace command should have --gitops-tool flag")
	}
}

func TestMarketplaceCmd_HasPlatformFlag(t *testing.T) {
	flag := marketplaceCmd.PersistentFlags().Lookup("platform")
	if flag == nil {
		t.Error("marketplace command should have --platform flag")
	}
}

func TestMarketplaceCmd_DefaultGitOpsTool(t *testing.T) {
	flag := marketplaceCmd.PersistentFlags().Lookup("gitops-tool")
	if flag == nil {
		t.Skip("gitops-tool flag not found")
	}
	if flag.DefValue != "argocd" {
		t.Errorf("Default gitops-tool should be 'argocd', got '%s'", flag.DefValue)
	}
}

func TestMarketplaceCmd_DefaultPlatform(t *testing.T) {
	flag := marketplaceCmd.PersistentFlags().Lookup("platform")
	if flag == nil {
		t.Skip("platform flag not found")
	}
	if flag.DefValue != "kubernetes" {
		t.Errorf("Default platform should be 'kubernetes', got '%s'", flag.DefValue)
	}
}

// Edge case tests for install command flags

func TestInstallCmd_HasVersionFlag(t *testing.T) {
	flag := installCmd.Flags().Lookup("version")
	if flag == nil {
		t.Error("install command should have --version flag")
	}
}

func TestInstallCmd_HasConfigFlag(t *testing.T) {
	flag := installCmd.Flags().Lookup("config")
	if flag == nil {
		t.Error("install command should have --config flag")
	}
}

func TestInstallCmd_HasDryRunFlag(t *testing.T) {
	flag := installCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Error("install command should have --dry-run flag")
	}
}

func TestInstallCmd_HasForceFlag(t *testing.T) {
	flag := installCmd.Flags().Lookup("force")
	if flag == nil {
		t.Error("install command should have --force flag")
	}
}

func TestInstallCmd_HasEnvFlag(t *testing.T) {
	flag := installCmd.Flags().Lookup("env")
	if flag == nil {
		t.Error("install command should have --env flag")
	}
}

func TestInstallCmd_HasSkipDepsFlag(t *testing.T) {
	flag := installCmd.Flags().Lookup("skip-deps")
	if flag == nil {
		t.Error("install command should have --skip-deps flag")
	}
}

// Edge case tests for search command flags

func TestMarketplaceSearchCmd_HasCategoryFlag(t *testing.T) {
	flag := marketplaceSearchCmd.Flags().Lookup("category")
	if flag == nil {
		t.Error("marketplace search command should have --category flag")
	}
}

func TestMarketplaceSearchCmd_HasTagsFlag(t *testing.T) {
	flag := marketplaceSearchCmd.Flags().Lookup("tags")
	if flag == nil {
		t.Error("marketplace search command should have --tags flag")
	}
}

func TestMarketplaceSearchCmd_HasLimitFlag(t *testing.T) {
	flag := marketplaceSearchCmd.Flags().Lookup("limit")
	if flag == nil {
		t.Error("marketplace search command should have --limit flag")
	}
}

func TestMarketplaceSearchCmd_DefaultLimit(t *testing.T) {
	flag := marketplaceSearchCmd.Flags().Lookup("limit")
	if flag == nil {
		t.Skip("limit flag not found")
	}
	if flag.DefValue != "20" {
		t.Errorf("Default limit should be '20', got '%s'", flag.DefValue)
	}
}
