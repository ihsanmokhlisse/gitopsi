package progress

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewProgress(t *testing.T) {
	p := New("Test Title", "test-project")

	if p.title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", p.title)
	}
	if p.projectName != "test-project" {
		t.Errorf("Expected projectName 'test-project', got '%s'", p.projectName)
	}
	if len(p.sections) != 0 {
		t.Errorf("Expected empty sections, got %d", len(p.sections))
	}
}

func TestProgressSetQuiet(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	if !p.quiet {
		t.Error("Expected quiet to be true")
	}
}

func TestProgressSetJSON(t *testing.T) {
	p := New("Test", "project")
	p.SetJSON(true)

	if !p.jsonOutput {
		t.Error("Expected jsonOutput to be true")
	}
}

func TestStartSection(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := p.StartSection("Test Section")

	if section.Name != "Test Section" {
		t.Errorf("Expected section name 'Test Section', got '%s'", section.Name)
	}
	if len(p.sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(p.sections))
	}
}

func TestStartStep(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := p.StartSection("Test Section")
	step := p.StartStep(section, "Test Step")

	if step.Name != "Test Step" {
		t.Errorf("Expected step name 'Test Step', got '%s'", step.Name)
	}
	if step.Status != StatusRunning {
		t.Errorf("Expected status Running, got %s", step.Status)
	}
	if len(section.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(section.Steps))
	}
}

func TestSuccessStep(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := p.StartSection("Test Section")
	step := p.StartStep(section, "Test Step")

	time.Sleep(10 * time.Millisecond)
	p.SuccessStep(section, step)

	if step.Status != StatusSuccess {
		t.Errorf("Expected status Success, got %s", step.Status)
	}
	if step.Duration < 10*time.Millisecond {
		t.Error("Expected duration to be at least 10ms")
	}
}

func TestFailStep(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := p.StartSection("Test Section")
	step := p.StartStep(section, "Test Step")

	testErr := os.ErrNotExist
	p.FailStep(section, step, testErr)

	if step.Status != StatusFailed {
		t.Errorf("Expected status Failed, got %s", step.Status)
	}
	if step.Message != testErr.Error() {
		t.Errorf("Expected message '%s', got '%s'", testErr.Error(), step.Message)
	}
}

func TestWarningStep(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := p.StartSection("Test Section")
	step := p.StartStep(section, "Test Step")

	p.WarningStep(section, step, "test warning")

	if step.Status != StatusWarning {
		t.Errorf("Expected status Warning, got %s", step.Status)
	}
	if step.Message != "test warning" {
		t.Errorf("Expected message 'test warning', got '%s'", step.Message)
	}
}

func TestAddSubStep(t *testing.T) {
	step := &Step{Name: "Parent"}
	step.AddSubStep("Child 1", StatusSuccess)
	step.AddSubStep("Child 2", StatusPending)

	if len(step.SubSteps) != 2 {
		t.Errorf("Expected 2 sub-steps, got %d", len(step.SubSteps))
	}
	if step.SubSteps[0].Name != "Child 1" {
		t.Errorf("Expected sub-step name 'Child 1', got '%s'", step.SubSteps[0].Name)
	}
	if step.SubSteps[0].Status != StatusSuccess {
		t.Errorf("Expected sub-step status Success, got %s", step.SubSteps[0].Status)
	}
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		status   StepStatus
		expected string
	}{
		{StatusSuccess, "✓"},
		{StatusFailed, "✗"},
		{StatusWarning, "⚠"},
		{StatusRunning, "●"},
		{StatusSkipped, "○"},
		{StatusPending, "○"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			icon := getStatusIcon(tt.status)
			if !bytes.Contains([]byte(icon), []byte(tt.expected)) {
				t.Errorf("Expected icon to contain '%s' for status %s", tt.expected, tt.status)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{100 * time.Millisecond, "ms"},
		{1500 * time.Millisecond, "s"},
		{5 * time.Second, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := formatDuration(tt.duration)
			if !bytes.Contains([]byte(result), []byte(tt.contains)) {
				t.Errorf("Expected '%s' to contain '%s'", result, tt.contains)
			}
		})
	}
}

func TestShowValidation(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	checks := []ValidationCheck{
		{Name: "Check 1", Status: "passed"},
		{Name: "Check 2", Status: "warning", Message: "test warning"},
		{Name: "Check 3", Status: "failed", Message: "test failure"},
	}

	p.ShowValidation(checks)
}

func TestShowError(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	err := os.ErrNotExist
	suggestions := []string{
		"Suggestion 1",
		"Suggestion 2",
	}

	p.ShowError(err, suggestions)
}

func TestSaveSummary(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(projectPath, 0750); err != nil {
		t.Fatal(err)
	}

	summary := &SetupSummary{
		Setup: SetupInfo{
			CompletedAt: time.Now(),
			Duration:    5 * time.Second,
			Version:     "0.1.0",
		},
		Git: GitInfo{
			URL:      "https://github.com/test/repo.git",
			Branch:   "main",
			Provider: "github",
			Status:   "connected",
		},
		Cluster: ClusterInfo{
			Name:     "test-cluster",
			URL:      "https://cluster.example.com",
			Platform: "kubernetes",
			Status:   "connected",
		},
		GitOpsTool: GitOpsToolInfo{
			Name:      "argocd",
			URL:       "https://argocd.example.com",
			Username:  "admin",
			Namespace: "argocd",
			Status:    "healthy",
		},
	}

	if err := SaveSummary(projectPath, summary); err != nil {
		t.Fatalf("SaveSummary failed: %v", err)
	}

	summaryPath := filepath.Join(projectPath, ".gitopsi", "setup-summary.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Summary file was not created")
	}
}

func TestLoadSummary(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(projectPath, 0750); err != nil {
		t.Fatal(err)
	}

	originalSummary := &SetupSummary{
		Setup: SetupInfo{
			CompletedAt: time.Now(),
			Duration:    5 * time.Second,
			Version:     "0.1.0",
		},
		Git: GitInfo{
			URL:      "https://github.com/test/repo.git",
			Branch:   "main",
			Provider: "github",
			Status:   "connected",
		},
	}

	if err := SaveSummary(projectPath, originalSummary); err != nil {
		t.Fatalf("SaveSummary failed: %v", err)
	}

	loadedSummary, err := LoadSummary(projectPath)
	if err != nil {
		t.Fatalf("LoadSummary failed: %v", err)
	}

	if loadedSummary.Git.URL != originalSummary.Git.URL {
		t.Errorf("Expected Git URL '%s', got '%s'", originalSummary.Git.URL, loadedSummary.Git.URL)
	}
	if loadedSummary.Git.Branch != originalSummary.Git.Branch {
		t.Errorf("Expected Git Branch '%s', got '%s'", originalSummary.Git.Branch, loadedSummary.Git.Branch)
	}
}

func TestLoadSummaryNotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, err := LoadSummary(tempDir)
	if err == nil {
		t.Error("Expected error when loading non-existent summary")
	}
}

func TestShowSummaryJSON(t *testing.T) {
	p := New("Test", "project")
	p.SetJSON(true)

	summary := &SetupSummary{
		Setup: SetupInfo{
			Version: "0.1.0",
		},
		Git: GitInfo{
			URL: "https://github.com/test/repo.git",
		},
	}

	p.ShowSummary(summary)
}

func TestShowSummaryQuiet(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	summary := &SetupSummary{
		GitOpsTool: GitOpsToolInfo{
			URL: "https://argocd.example.com",
		},
	}

	p.ShowSummary(summary)
}

func TestShowHeader(t *testing.T) {
	t.Run("normal mode", func(t *testing.T) {
		p := New("Test Title", "test-project")
		p.ShowHeader()
	})

	t.Run("quiet mode", func(t *testing.T) {
		p := New("Test Title", "test-project")
		p.SetQuiet(true)
		p.ShowHeader()
	})

	t.Run("json mode", func(t *testing.T) {
		p := New("Test Title", "test-project")
		p.SetJSON(true)
		p.ShowHeader()
	})
}

func TestUpdateStep(t *testing.T) {
	t.Run("with current step", func(t *testing.T) {
		p := New("Test", "project")
		p.SetQuiet(true)

		section := p.StartSection("Test Section")
		step := p.StartStep(section, "Test Step")

		p.UpdateStep("Updated message")

		if step.Message != "Updated message" {
			t.Errorf("Expected message 'Updated message', got '%s'", step.Message)
		}
	})

	t.Run("without current step", func(t *testing.T) {
		p := New("Test", "project")
		p.SetQuiet(true)

		p.UpdateStep("No effect")
	})

	t.Run("with active spinner", func(t *testing.T) {
		p := New("Test", "project")
		p.SetQuiet(true) // Disable spinner animation to avoid race condition

		section := p.StartSection("Test Section")
		_ = p.StartStep(section, "Test Step")

		p.UpdateStep("Spinner update")
	})
}

func TestShowSubSteps(t *testing.T) {
	t.Run("with sub-steps", func(t *testing.T) {
		p := New("Test", "project")

		step := &Step{Name: "Parent"}
		step.AddSubStep("Child 1", StatusSuccess)
		step.AddSubStep("Child 2", StatusFailed)
		step.AddSubStep("Child 3", StatusWarning)

		p.ShowSubSteps(step)
	})

	t.Run("without sub-steps", func(t *testing.T) {
		p := New("Test", "project")
		step := &Step{Name: "Parent"}

		p.ShowSubSteps(step)
	})

	t.Run("quiet mode", func(t *testing.T) {
		p := New("Test", "project")
		p.SetQuiet(true)

		step := &Step{Name: "Parent"}
		step.AddSubStep("Child", StatusSuccess)

		p.ShowSubSteps(step)
	})
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"connected", "connected"},
		{"synced", "synced"},
		{"healthy", "healthy"},
		{"ready", "ready"},
		{"warning", "warning"},
		{"degraded", "degraded"},
		{"failed", "failed"},
		{"error", "error"},
		{"disconnected", "disconnected"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := formatStatus(tt.status)
			if !bytes.Contains([]byte(result), []byte(tt.expected)) {
				t.Errorf("Expected result to contain '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestShowValidationAllStatuses(t *testing.T) {
	p := New("Test", "project")

	checks := []ValidationCheck{
		{Name: "Passed Check", Status: "passed", Message: ""},
		{Name: "Warning Check", Status: "warning", Message: "warning message"},
		{Name: "Failed Check", Status: "failed", Message: "failure message"},
	}

	p.ShowValidation(checks)
}

func TestShowErrorWithSuggestions(t *testing.T) {
	p := New("Test", "project")

	err := os.ErrPermission
	suggestions := []string{
		"Check file permissions",
		"Run with elevated privileges",
		"Verify the path exists",
	}

	p.ShowError(err, suggestions)
}

func TestShowErrorNoSuggestions(t *testing.T) {
	p := New("Test", "project")

	err := os.ErrNotExist
	p.ShowError(err, nil)
}

func TestSectionAddStep(t *testing.T) {
	section := &Section{
		Name:  "Test Section",
		Steps: make([]*Step, 0),
	}

	step1 := section.AddStep("Step 1")
	step2 := section.AddStep("Step 2")

	if len(section.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(section.Steps))
	}

	if step1.Name != "Step 1" {
		t.Errorf("Expected step name 'Step 1', got '%s'", step1.Name)
	}

	if step2.Status != StatusRunning {
		t.Errorf("Expected status Running, got %s", step2.Status)
	}
}

func TestShowSummaryFull(t *testing.T) {
	p := New("Test", "test-project")

	summary := &SetupSummary{
		Setup: SetupInfo{
			CompletedAt: time.Now(),
			Duration:    30 * time.Second,
			Version:     "0.2.0",
		},
		Git: GitInfo{
			URL:      "https://github.com/test/repo.git",
			Branch:   "main",
			WebURL:   "https://github.com/test/repo",
			Provider: "github",
			Status:   "connected",
		},
		Cluster: ClusterInfo{
			Name:       "production",
			URL:        "https://k8s.example.com",
			Platform:   "kubernetes",
			Version:    "1.28.0",
			Status:     "ready",
			Namespaces: []string{"app-dev", "app-staging", "app-prod"},
		},
		GitOpsTool: GitOpsToolInfo{
			Name:      "argocd",
			URL:       "https://argocd.example.com",
			Username:  "admin",
			Password:  "secret123",
			Namespace: "argocd",
			Version:   "2.10.0",
			Status:    "healthy",
			PodCount:  "5/5",
		},
		Environments: []EnvironmentInfo{
			{Name: "dev", Namespace: "app-dev", Status: "synced"},
			{Name: "staging", Namespace: "app-staging", Status: "synced"},
			{Name: "prod", Namespace: "app-prod", Status: "synced"},
		},
		Applications: []ApplicationInfo{
			{Name: "app-of-apps", Type: "ApplicationSet", Status: "synced", Children: []string{"infra-dev", "apps-dev"}},
			{Name: "standalone-app", Type: "Application", Status: "healthy"},
		},
	}

	p.ShowSummary(summary)
}

func TestShowSummaryWithNoPassword(t *testing.T) {
	p := New("Test", "test-project")

	summary := &SetupSummary{
		Setup: SetupInfo{
			Duration: 10 * time.Second,
		},
		Git: GitInfo{
			URL:    "https://github.com/test/repo.git",
			Branch: "main",
			Status: "connected",
		},
		GitOpsTool: GitOpsToolInfo{
			Name:      "argocd",
			URL:       "https://argocd.example.com",
			Username:  "admin",
			Password:  "",
			Namespace: "argocd",
			Status:    "healthy",
		},
	}

	p.ShowSummary(summary)
}

func TestGetStatusIconPassed(t *testing.T) {
	icon := getStatusIcon("passed")
	if !bytes.Contains([]byte(icon), []byte("✓")) {
		t.Error("Expected checkmark for 'passed' status")
	}
}

func TestStepStatus(t *testing.T) {
	statuses := []StepStatus{
		StatusPending,
		StatusRunning,
		StatusSuccess,
		StatusFailed,
		StatusSkipped,
		StatusWarning,
	}

	for _, status := range statuses {
		if status == "" {
			t.Errorf("Empty status found")
		}
	}
}

func TestLoadSummaryInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")
	gitopsiDir := filepath.Join(projectPath, ".gitopsi")

	if err := os.MkdirAll(gitopsiDir, 0750); err != nil {
		t.Fatal(err)
	}

	summaryPath := filepath.Join(gitopsiDir, "setup-summary.yaml")
	if err := os.WriteFile(summaryPath, []byte("invalid: yaml: content: ["), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSummary(projectPath)
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

// Edge case tests for StepStatus constants
func TestStepStatus_AllValues(t *testing.T) {
	statuses := map[StepStatus]string{
		StatusPending: "pending",
		StatusRunning: "running",
		StatusSuccess: "success",
		StatusFailed:  "failed",
		StatusSkipped: "skipped",
		StatusWarning: "warning",
	}

	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("Expected status %s to equal %s", status, expected)
		}
	}
}

// Edge case tests for Step struct
func TestStep_EmptyFields(t *testing.T) {
	step := &Step{}

	if step.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", step.Name)
	}
	if step.Status != "" {
		t.Errorf("Expected empty Status, got '%s'", step.Status)
	}
	if step.Duration != 0 {
		t.Errorf("Expected zero Duration, got %v", step.Duration)
	}
	if !step.StartTime.IsZero() {
		t.Errorf("Expected zero StartTime, got %v", step.StartTime)
	}
	if step.Message != "" {
		t.Errorf("Expected empty Message, got '%s'", step.Message)
	}
	if step.SubSteps != nil {
		t.Errorf("Expected nil SubSteps, got %v", step.SubSteps)
	}
}

func TestStep_AddSubStepToNilSlice(t *testing.T) {
	step := &Step{Name: "Parent"}
	// SubSteps is nil initially
	if step.SubSteps != nil {
		t.Fatal("Expected nil SubSteps initially")
	}

	step.AddSubStep("Child", StatusSuccess)

	if len(step.SubSteps) != 1 {
		t.Errorf("Expected 1 sub-step, got %d", len(step.SubSteps))
	}
}

func TestStep_AddMultipleSubSteps(t *testing.T) {
	step := &Step{Name: "Parent"}

	statuses := []StepStatus{StatusPending, StatusRunning, StatusSuccess, StatusFailed, StatusSkipped, StatusWarning}
	for i, status := range statuses {
		step.AddSubStep(fmt.Sprintf("Child %d", i), status)
	}

	if len(step.SubSteps) != 6 {
		t.Errorf("Expected 6 sub-steps, got %d", len(step.SubSteps))
	}

	for i, sub := range step.SubSteps {
		if sub.Status != statuses[i] {
			t.Errorf("SubStep %d: expected status %s, got %s", i, statuses[i], sub.Status)
		}
	}
}

// Edge case tests for Section struct
func TestSection_EmptySteps(t *testing.T) {
	section := &Section{
		Name:  "Empty Section",
		Steps: make([]*Step, 0),
	}

	if len(section.Steps) != 0 {
		t.Errorf("Expected 0 steps, got %d", len(section.Steps))
	}
	if section.spinner != nil {
		t.Error("Expected nil spinner")
	}
}

func TestSection_NilSteps(t *testing.T) {
	section := &Section{
		Name: "Nil Steps Section",
	}

	if section.Steps != nil {
		t.Error("Expected nil Steps")
	}
}

// Edge case tests for Progress struct
func TestNew_EmptyStrings(t *testing.T) {
	p := New("", "")

	if p.title != "" {
		t.Errorf("Expected empty title, got '%s'", p.title)
	}
	if p.projectName != "" {
		t.Errorf("Expected empty projectName, got '%s'", p.projectName)
	}
	if p.quiet {
		t.Error("Expected quiet to be false by default")
	}
	if p.jsonOutput {
		t.Error("Expected jsonOutput to be false by default")
	}
	if p.currentStep != nil {
		t.Error("Expected currentStep to be nil")
	}
}

func TestProgress_BothModes(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)
	p.SetJSON(true)

	if !p.quiet {
		t.Error("Expected quiet to be true")
	}
	if !p.jsonOutput {
		t.Error("Expected jsonOutput to be true")
	}
}

func TestProgress_ModeSwitching(t *testing.T) {
	p := New("Test", "project")

	p.SetQuiet(true)
	if !p.quiet {
		t.Error("Expected quiet true after SetQuiet(true)")
	}

	p.SetQuiet(false)
	if p.quiet {
		t.Error("Expected quiet false after SetQuiet(false)")
	}

	p.SetJSON(true)
	if !p.jsonOutput {
		t.Error("Expected jsonOutput true after SetJSON(true)")
	}

	p.SetJSON(false)
	if p.jsonOutput {
		t.Error("Expected jsonOutput false after SetJSON(false)")
	}
}

func TestProgress_MultipleSections(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section1 := p.StartSection("Section 1")
	section2 := p.StartSection("Section 2")
	section3 := p.StartSection("Section 3")

	if len(p.sections) != 3 {
		t.Errorf("Expected 3 sections, got %d", len(p.sections))
	}

	if section1.Name != "Section 1" {
		t.Errorf("Expected section1 name 'Section 1', got '%s'", section1.Name)
	}
	if section2.Name != "Section 2" {
		t.Errorf("Expected section2 name 'Section 2', got '%s'", section2.Name)
	}
	if section3.Name != "Section 3" {
		t.Errorf("Expected section3 name 'Section 3', got '%s'", section3.Name)
	}
}

func TestProgress_StepWithNoSection(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	// Create a section manually without using StartSection
	section := &Section{
		Name:  "Manual Section",
		Steps: make([]*Step, 0),
	}

	step := p.StartStep(section, "Manual Step")

	if step.Name != "Manual Step" {
		t.Errorf("Expected step name 'Manual Step', got '%s'", step.Name)
	}
	if p.currentStep != step {
		t.Error("Expected currentStep to be set")
	}
}

// Edge case tests for formatDuration
func TestFormatDuration_ExactlyOneSecond(t *testing.T) {
	result := formatDuration(1 * time.Second)
	if result != "1.0s" {
		t.Errorf("Expected '1.0s', got '%s'", result)
	}
}

func TestFormatDuration_JustUnderOneSecond(t *testing.T) {
	result := formatDuration(999 * time.Millisecond)
	if !bytes.Contains([]byte(result), []byte("ms")) {
		t.Errorf("Expected milliseconds format, got '%s'", result)
	}
}

func TestFormatDuration_ZeroDuration(t *testing.T) {
	result := formatDuration(0)
	if result != "0ms" {
		t.Errorf("Expected '0ms', got '%s'", result)
	}
}

func TestFormatDuration_LargeDuration(t *testing.T) {
	result := formatDuration(120 * time.Second)
	if !bytes.Contains([]byte(result), []byte("s")) {
		t.Errorf("Expected seconds format for 2 minutes, got '%s'", result)
	}
}

// Edge case tests for getStatusIcon
func TestGetStatusIcon_UnknownStatus(t *testing.T) {
	icon := getStatusIcon("unknown")
	if !bytes.Contains([]byte(icon), []byte("○")) {
		t.Errorf("Expected gray circle for unknown status, got '%s'", icon)
	}
}

func TestGetStatusIcon_EmptyStatus(t *testing.T) {
	icon := getStatusIcon("")
	if !bytes.Contains([]byte(icon), []byte("○")) {
		t.Errorf("Expected gray circle for empty status, got '%s'", icon)
	}
}

// Edge case tests for ValidationCheck
func TestValidationCheck_EmptyFields(t *testing.T) {
	check := ValidationCheck{}

	if check.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", check.Name)
	}
	if check.Check != "" {
		t.Errorf("Expected empty Check, got '%s'", check.Check)
	}
	if check.Status != "" {
		t.Errorf("Expected empty Status, got '%s'", check.Status)
	}
	if check.Message != "" {
		t.Errorf("Expected empty Message, got '%s'", check.Message)
	}
}

// Edge case tests for summary structs
func TestSetupInfo_EmptyFields(t *testing.T) {
	info := SetupInfo{}

	if !info.CompletedAt.IsZero() {
		t.Error("Expected zero CompletedAt")
	}
	if info.Duration != 0 {
		t.Error("Expected zero Duration")
	}
	if info.Version != "" {
		t.Error("Expected empty Version")
	}
}

func TestGitInfo_EmptyFields(t *testing.T) {
	info := GitInfo{}

	if info.URL != "" {
		t.Error("Expected empty URL")
	}
	if info.Branch != "" {
		t.Error("Expected empty Branch")
	}
	if info.WebURL != "" {
		t.Error("Expected empty WebURL")
	}
	if info.Provider != "" {
		t.Error("Expected empty Provider")
	}
	if info.Status != "" {
		t.Error("Expected empty Status")
	}
}

func TestClusterInfo_EmptyNamespaces(t *testing.T) {
	info := ClusterInfo{
		Name:       "test-cluster",
		Namespaces: []string{},
	}

	if len(info.Namespaces) != 0 {
		t.Errorf("Expected empty Namespaces, got %d", len(info.Namespaces))
	}
}

func TestClusterInfo_NilNamespaces(t *testing.T) {
	info := ClusterInfo{
		Name: "test-cluster",
	}

	if info.Namespaces != nil {
		t.Error("Expected nil Namespaces")
	}
}

func TestGitOpsToolInfo_AllFields(t *testing.T) {
	info := GitOpsToolInfo{
		Name:           "argocd",
		URL:            "https://argocd.example.com",
		Username:       "admin",
		Password:       "secret",
		PasswordSecret: "argocd-initial-admin-secret",
		Namespace:      "argocd",
		Version:        "2.10.0",
		Status:         "healthy",
		PodCount:       "5/5",
	}

	if info.Name != "argocd" {
		t.Errorf("Expected Name 'argocd', got '%s'", info.Name)
	}
	if info.PasswordSecret != "argocd-initial-admin-secret" {
		t.Errorf("Expected PasswordSecret, got '%s'", info.PasswordSecret)
	}
}

func TestEnvironmentInfo_AllFields(t *testing.T) {
	info := EnvironmentInfo{
		Name:      "production",
		Namespace: "prod",
		Status:    "synced",
	}

	if info.Name != "production" {
		t.Errorf("Expected Name 'production', got '%s'", info.Name)
	}
	if info.Namespace != "prod" {
		t.Errorf("Expected Namespace 'prod', got '%s'", info.Namespace)
	}
	if info.Status != "synced" {
		t.Errorf("Expected Status 'synced', got '%s'", info.Status)
	}
}

func TestApplicationInfo_EmptyChildren(t *testing.T) {
	info := ApplicationInfo{
		Name:     "test-app",
		Type:     "Application",
		Status:   "healthy",
		Children: []string{},
	}

	if len(info.Children) != 0 {
		t.Errorf("Expected empty Children, got %d", len(info.Children))
	}
}

func TestApplicationInfo_NilChildren(t *testing.T) {
	info := ApplicationInfo{
		Name:   "test-app",
		Type:   "Application",
		Status: "healthy",
	}

	if info.Children != nil {
		t.Error("Expected nil Children")
	}
}

func TestSetupSummary_EmptyFields(t *testing.T) {
	summary := SetupSummary{}

	if summary.Setup.Version != "" {
		t.Error("Expected empty Setup.Version")
	}
	if summary.Git.URL != "" {
		t.Error("Expected empty Git.URL")
	}
	if summary.Cluster.Name != "" {
		t.Error("Expected empty Cluster.Name")
	}
	if summary.GitOpsTool.Name != "" {
		t.Error("Expected empty GitOpsTool.Name")
	}
	if summary.Environments != nil {
		t.Error("Expected nil Environments")
	}
	if summary.Applications != nil {
		t.Error("Expected nil Applications")
	}
}

func TestSetupSummary_EmptySlices(t *testing.T) {
	summary := SetupSummary{
		Environments: []EnvironmentInfo{},
		Applications: []ApplicationInfo{},
	}

	if len(summary.Environments) != 0 {
		t.Errorf("Expected empty Environments, got %d", len(summary.Environments))
	}
	if len(summary.Applications) != 0 {
		t.Errorf("Expected empty Applications, got %d", len(summary.Applications))
	}
}

// Edge case tests for formatStatus
func TestFormatStatus_AllStatuses(t *testing.T) {
	tests := []struct {
		status   string
		contains string
	}{
		{"connected", "connected"},
		{"synced", "synced"},
		{"healthy", "healthy"},
		{"ready", "ready"},
		{"warning", "warning"},
		{"degraded", "degraded"},
		{"failed", "failed"},
		{"error", "error"},
		{"disconnected", "disconnected"},
		{"custom", "custom"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := formatStatus(tt.status)
			if !bytes.Contains([]byte(result), []byte(tt.contains)) {
				t.Errorf("Expected result to contain '%s', got '%s'", tt.contains, result)
			}
		})
	}
}

// Edge case tests for SaveSummary
func TestSaveSummary_EmptySummary(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(projectPath, 0750); err != nil {
		t.Fatal(err)
	}

	summary := &SetupSummary{}

	err := SaveSummary(projectPath, summary)
	if err != nil {
		t.Fatalf("SaveSummary failed with empty summary: %v", err)
	}

	// Verify file was created
	summaryPath := filepath.Join(projectPath, ".gitopsi", "setup-summary.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Summary file was not created")
	}
}

func TestSaveSummary_NestedDirCreation(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "deep", "nested", "path", "test-project")

	// SaveSummary uses MkdirAll which creates all parent directories
	summary := &SetupSummary{}

	err := SaveSummary(projectPath, summary)
	if err != nil {
		t.Errorf("SaveSummary() should create nested directories, got error: %v", err)
	}

	// Verify the file was created
	summaryPath := filepath.Join(projectPath, ".gitopsi", "setup-summary.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Summary file should exist after SaveSummary")
	}
}

// Edge case tests for ShowSummary
func TestShowSummary_MinimalSummary(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	summary := &SetupSummary{
		GitOpsTool: GitOpsToolInfo{
			URL: "https://argocd.example.com",
		},
	}

	// Should not panic
	p.ShowSummary(summary)
}

func TestShowSummary_WithEmptyGitOpsToolURL(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	summary := &SetupSummary{
		GitOpsTool: GitOpsToolInfo{
			Name: "argocd",
		},
	}

	// Should not panic
	p.ShowSummary(summary)
}

// Edge case tests for ShowValidation
func TestShowValidation_EmptyChecks(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	checks := []ValidationCheck{}

	// Should not panic
	p.ShowValidation(checks)
}

func TestShowValidation_NilChecks(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	// Should not panic
	p.ShowValidation(nil)
}

// Edge case tests for ShowError
func TestShowError_EmptySuggestions(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	err := os.ErrNotExist
	suggestions := []string{}

	// Should not panic
	p.ShowError(err, suggestions)
}

// Edge case tests for UpdateStep with nil currentStep
func TestUpdateStep_NilCurrentStep(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	// p.currentStep is nil
	p.UpdateStep("Should not panic")

	if p.currentStep != nil {
		t.Error("Expected currentStep to remain nil")
	}
}

// Edge case tests for SuccessStep, FailStep, WarningStep with nil spinner
func TestSuccessStep_NilSpinner(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := &Section{
		Name:    "Test Section",
		Steps:   make([]*Step, 0),
		spinner: nil,
	}

	step := &Step{
		Name:      "Test Step",
		Status:    StatusRunning,
		StartTime: time.Now().Add(-100 * time.Millisecond),
	}

	// Should not panic
	p.SuccessStep(section, step)

	if step.Status != StatusSuccess {
		t.Errorf("Expected status Success, got %s", step.Status)
	}
}

func TestFailStep_NilSpinner(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := &Section{
		Name:    "Test Section",
		Steps:   make([]*Step, 0),
		spinner: nil,
	}

	step := &Step{
		Name:      "Test Step",
		Status:    StatusRunning,
		StartTime: time.Now(),
	}

	testErr := os.ErrPermission

	// Should not panic
	p.FailStep(section, step, testErr)

	if step.Status != StatusFailed {
		t.Errorf("Expected status Failed, got %s", step.Status)
	}
}

func TestWarningStep_NilSpinner(t *testing.T) {
	p := New("Test", "project")
	p.SetQuiet(true)

	section := &Section{
		Name:    "Test Section",
		Steps:   make([]*Step, 0),
		spinner: nil,
	}

	step := &Step{
		Name:      "Test Step",
		Status:    StatusRunning,
		StartTime: time.Now(),
	}

	// Should not panic
	p.WarningStep(section, step, "warning message")

	if step.Status != StatusWarning {
		t.Errorf("Expected status Warning, got %s", step.Status)
	}
}

// Edge case tests for ShowSubSteps
func TestShowSubSteps_NilSubSteps(t *testing.T) {
	p := New("Test", "project")

	step := &Step{
		Name:     "Parent",
		SubSteps: nil,
	}

	// Should not panic
	p.ShowSubSteps(step)
}

func TestShowSubSteps_JSONMode(t *testing.T) {
	p := New("Test", "project")
	p.SetJSON(true)

	step := &Step{Name: "Parent"}
	step.AddSubStep("Child", StatusSuccess)

	// Should not panic and should not display anything
	p.ShowSubSteps(step)
}

// Edge case tests for LoadSummary
func TestLoadSummary_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")
	gitopsiDir := filepath.Join(projectPath, ".gitopsi")

	if err := os.MkdirAll(gitopsiDir, 0750); err != nil {
		t.Fatal(err)
	}

	summaryPath := filepath.Join(gitopsiDir, "setup-summary.yaml")
	if err := os.WriteFile(summaryPath, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}

	summary, err := LoadSummary(projectPath)
	if err != nil {
		t.Fatalf("LoadSummary should not fail with empty file: %v", err)
	}

	// Empty YAML should unmarshal to zero-value struct
	if summary.Setup.Version != "" {
		t.Error("Expected empty Version from empty file")
	}
}

// Edge case tests for ShowHeader
func TestShowHeader_BothModesEnabled(t *testing.T) {
	p := New("Test Title", "test-project")
	p.SetQuiet(true)
	p.SetJSON(true)

	// Should not panic and should not display anything
	p.ShowHeader()
}
