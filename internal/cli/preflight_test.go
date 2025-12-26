package cli

import (
	"testing"
)

func TestPreflightResultStructure(t *testing.T) {
	result := PreflightResult{
		Name:    "Test Check",
		Status:  "ok",
		Message: "Test passed",
		Details: "Additional details",
	}

	if result.Name != "Test Check" {
		t.Errorf("Expected Name 'Test Check', got '%s'", result.Name)
	}

	if result.Status != "ok" {
		t.Errorf("Expected Status 'ok', got '%s'", result.Status)
	}

	if result.Message != "Test passed" {
		t.Errorf("Expected Message 'Test passed', got '%s'", result.Message)
	}

	if result.Details != "Additional details" {
		t.Errorf("Expected Details 'Additional details', got '%s'", result.Details)
	}
}

func TestPreflightResultStatuses(t *testing.T) {
	statuses := []string{"ok", "warn", "fail"}

	for _, status := range statuses {
		result := PreflightResult{
			Name:   "Status Test",
			Status: status,
		}

		if result.Status != status {
			t.Errorf("Expected status '%s', got '%s'", status, result.Status)
		}
	}
}

func TestPreflightCmdExists(t *testing.T) {
	if preflightCmd == nil {
		t.Fatal("preflightCmd should not be nil")
	}

	if preflightCmd.Use != "preflight" {
		t.Errorf("Expected Use 'preflight', got '%s'", preflightCmd.Use)
	}

	if preflightCmd.Short == "" {
		t.Error("preflightCmd should have a short description")
	}

	if preflightCmd.Long == "" {
		t.Error("preflightCmd should have a long description")
	}
}

func TestPreflightFlags(t *testing.T) {
	flags := []struct {
		name     string
		defValue string
	}{
		{"cluster", ""},
		{"kubeconfig", ""},
		{"context", ""},
		{"gitops-tool", "argocd"},
		{"timeout", "30"},
	}

	for _, flag := range flags {
		f := preflightCmd.Flags().Lookup(flag.name)
		if f == nil {
			t.Errorf("Flag '--%s' should exist", flag.name)
			continue
		}

		if f.DefValue != flag.defValue {
			t.Errorf("Flag '--%s' default value should be '%s', got '%s'", flag.name, flag.defValue, f.DefValue)
		}
	}
}

func TestPrintResultFormats(t *testing.T) {
	// Test that printResult doesn't panic for different statuses
	results := []PreflightResult{
		{Name: "OK Check", Status: "ok", Message: "Success"},
		{Name: "Warning Check", Status: "warn", Message: "Warning", Details: "Some detail"},
		{Name: "Fail Check", Status: "fail", Message: "Failed", Details: "Error detail"},
	}

	for _, result := range results {
		// This should not panic
		printResult(result)
	}
}

func TestPrintSummaryFormats(t *testing.T) {
	// Test that printSummary doesn't panic for different result combinations
	testCases := [][]PreflightResult{
		// All OK
		{
			{Name: "Check1", Status: "ok"},
			{Name: "Check2", Status: "ok"},
		},
		// Mix of OK and warnings
		{
			{Name: "Check1", Status: "ok"},
			{Name: "Check2", Status: "warn"},
		},
		// Contains failure
		{
			{Name: "Check1", Status: "ok"},
			{Name: "Check2", Status: "fail"},
		},
		// Empty results
		{},
	}

	for i, results := range testCases {
		// This should not panic
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			printSummary(results)
		})
	}
}




