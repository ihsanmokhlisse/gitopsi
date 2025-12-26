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
