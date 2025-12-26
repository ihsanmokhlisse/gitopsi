package progress

import (
	"bytes"
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
