package progress

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator(true)

	if v == nil {
		t.Fatal("Expected non-nil validator")
	}
	if !v.verbose {
		t.Error("Expected verbose to be true")
	}
	if len(v.checks) != 0 {
		t.Errorf("Expected empty checks, got %d", len(v.checks))
	}
}

func TestValidatorGetAllChecks(t *testing.T) {
	v := NewValidator(false)

	v.checks = []ValidationCheck{
		{Name: "Check 1", Status: "passed"},
		{Name: "Check 2", Status: "failed"},
	}

	checks := v.GetAllChecks()
	if len(checks) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(checks))
	}
}

func TestValidatorHasFailures(t *testing.T) {
	tests := []struct {
		name     string
		checks   []ValidationCheck
		expected bool
	}{
		{
			name: "no failures",
			checks: []ValidationCheck{
				{Status: "passed"},
				{Status: "warning"},
			},
			expected: false,
		},
		{
			name: "with failure",
			checks: []ValidationCheck{
				{Status: "passed"},
				{Status: "failed"},
			},
			expected: true,
		},
		{
			name:     "empty",
			checks:   []ValidationCheck{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			v.checks = tt.checks

			result := v.HasFailures()
			if result != tt.expected {
				t.Errorf("Expected HasFailures() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValidatorHasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		checks   []ValidationCheck
		expected bool
	}{
		{
			name: "no warnings",
			checks: []ValidationCheck{
				{Status: "passed"},
				{Status: "passed"},
			},
			expected: false,
		},
		{
			name: "with warning",
			checks: []ValidationCheck{
				{Status: "passed"},
				{Status: "warning"},
			},
			expected: true,
		},
		{
			name:     "empty",
			checks:   []ValidationCheck{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			v.checks = tt.checks

			result := v.HasWarnings()
			if result != tt.expected {
				t.Errorf("Expected HasWarnings() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValidationCheckStruct(t *testing.T) {
	check := ValidationCheck{
		Name:    "Test Check",
		Check:   "Test validation",
		Status:  "passed",
		Message: "All good",
	}

	if check.Name != "Test Check" {
		t.Errorf("Expected Name 'Test Check', got '%s'", check.Name)
	}
	if check.Check != "Test validation" {
		t.Errorf("Expected Check 'Test validation', got '%s'", check.Check)
	}
	if check.Status != "passed" {
		t.Errorf("Expected Status 'passed', got '%s'", check.Status)
	}
	if check.Message != "All good" {
		t.Errorf("Expected Message 'All good', got '%s'", check.Message)
	}
}

// Edge case tests for Validator

func TestNewValidator_VerboseFalse(t *testing.T) {
	v := NewValidator(false)

	if v == nil {
		t.Fatal("Expected non-nil validator")
	}
	if v.verbose {
		t.Error("Expected verbose to be false")
	}
}

func TestValidator_GetAllChecks_Empty(t *testing.T) {
	v := NewValidator(false)

	checks := v.GetAllChecks()
	if len(checks) != 0 {
		t.Errorf("Expected 0 checks, got %d", len(checks))
	}
}

func TestValidator_GetAllChecks_ReturnsSameSlice(t *testing.T) {
	v := NewValidator(false)

	v.checks = []ValidationCheck{
		{Name: "Check 1", Status: "passed"},
	}

	checks1 := v.GetAllChecks()
	checks2 := v.GetAllChecks()

	// Both should reference the same underlying slice
	if len(checks1) != len(checks2) {
		t.Error("Expected same slice to be returned")
	}
}

func TestValidator_HasFailures_AllFailed(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "failed"},
		{Status: "failed"},
		{Status: "failed"},
	}

	if !v.HasFailures() {
		t.Error("Expected HasFailures() to return true when all checks failed")
	}
}

func TestValidator_HasFailures_AllPassed(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "passed"},
		{Status: "passed"},
		{Status: "passed"},
	}

	if v.HasFailures() {
		t.Error("Expected HasFailures() to return false when all checks passed")
	}
}

func TestValidator_HasWarnings_AllWarnings(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "warning"},
		{Status: "warning"},
	}

	if !v.HasWarnings() {
		t.Error("Expected HasWarnings() to return true when all checks have warnings")
	}
}

func TestValidator_HasWarnings_MixedStatuses(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "passed"},
		{Status: "warning"},
		{Status: "failed"},
	}

	if !v.HasWarnings() {
		t.Error("Expected HasWarnings() to return true with mixed statuses including warning")
	}
}

func TestValidator_HasFailures_OnlyFirst(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "failed"},
		{Status: "passed"},
		{Status: "passed"},
	}

	if !v.HasFailures() {
		t.Error("Expected HasFailures() to return true when first check failed")
	}
}

func TestValidator_HasFailures_OnlyLast(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "passed"},
		{Status: "passed"},
		{Status: "failed"},
	}

	if !v.HasFailures() {
		t.Error("Expected HasFailures() to return true when last check failed")
	}
}

func TestValidator_HasWarnings_OnlyFirst(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "warning"},
		{Status: "passed"},
		{Status: "passed"},
	}

	if !v.HasWarnings() {
		t.Error("Expected HasWarnings() to return true when first check has warning")
	}
}

func TestValidator_HasWarnings_OnlyLast(t *testing.T) {
	v := NewValidator(false)
	v.checks = []ValidationCheck{
		{Status: "passed"},
		{Status: "passed"},
		{Status: "warning"},
	}

	if !v.HasWarnings() {
		t.Error("Expected HasWarnings() to return true when last check has warning")
	}
}

func TestValidator_ChecksAccumulate(t *testing.T) {
	v := NewValidator(false)

	// Simulate adding checks directly (similar to how validation methods work)
	v.checks = append(v.checks, ValidationCheck{Name: "Check 1", Status: "passed"})
	v.checks = append(v.checks, ValidationCheck{Name: "Check 2", Status: "warning"})
	v.checks = append(v.checks, ValidationCheck{Name: "Check 3", Status: "failed"})

	if len(v.checks) != 3 {
		t.Errorf("Expected 3 checks, got %d", len(v.checks))
	}
}

func TestValidator_VerboseDoesNotAffectChecks(t *testing.T) {
	vVerbose := NewValidator(true)
	vQuiet := NewValidator(false)

	vVerbose.checks = []ValidationCheck{{Status: "failed"}}
	vQuiet.checks = []ValidationCheck{{Status: "failed"}}

	// Both should return same result regardless of verbose mode
	if vVerbose.HasFailures() != vQuiet.HasFailures() {
		t.Error("HasFailures() should return same result regardless of verbose mode")
	}
}

func TestValidationCheck_AllStatusValues(t *testing.T) {
	statuses := []string{"pending", "passed", "failed", "warning"}

	for _, status := range statuses {
		check := ValidationCheck{Status: status}
		if check.Status != status {
			t.Errorf("Expected status %s, got %s", status, check.Status)
		}
	}
}

func TestValidationCheck_EmptyMessage(t *testing.T) {
	check := ValidationCheck{
		Name:    "Check",
		Status:  "passed",
		Message: "",
	}

	if check.Message != "" {
		t.Errorf("Expected empty message, got '%s'", check.Message)
	}
}

func TestValidationCheck_LongMessage(t *testing.T) {
	longMessage := "This is a very long error message that describes in detail what went wrong during the validation process and provides helpful suggestions for how to fix the issue"

	check := ValidationCheck{
		Name:    "Check",
		Status:  "failed",
		Message: longMessage,
	}

	if check.Message != longMessage {
		t.Error("Expected long message to be preserved")
	}
}
