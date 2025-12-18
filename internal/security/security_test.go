package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSeverity(t *testing.T) {
	tests := []struct {
		severity ScanSeverity
		want     string
	}{
		{SeverityCritical, "CRITICAL"},
		{SeverityHigh, "HIGH"},
		{SeverityMedium, "MEDIUM"},
		{SeverityLow, "LOW"},
		{SeverityInfo, "INFO"},
	}

	for _, tt := range tests {
		if string(tt.severity) != tt.want {
			t.Errorf("Severity = %v, want %v", tt.severity, tt.want)
		}
	}
}

func TestNewScanner(t *testing.T) {
	s := NewScanner("/tmp/test", SeverityHigh, true)
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}
	if s.path != "/tmp/test" {
		t.Errorf("path = %v, want /tmp/test", s.path)
	}
	if s.severity != SeverityHigh {
		t.Errorf("severity = %v, want HIGH", s.severity)
	}
	if !s.verbose {
		t.Error("verbose should be true")
	}
}

func TestScanner_RunBuiltinChecks(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "security-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a deployment with security issues
	badDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: bad-deployment
spec:
  template:
    spec:
      containers:
        - name: bad
          image: nginx
          securityContext:
            privileged: true
            runAsUser: 0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "bad-deployment.yaml"), []byte(badDeployment), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a deployment with hardcoded secret
	secretDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: secret-deployment
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: API_KEY
              value: "secret123"
            - name: password
              value: "admin123"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "secret-deployment.yaml"), []byte(secretDeployment), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a pod with hostPath
	hostPathPod := `apiVersion: v1
kind: Pod
metadata:
  name: hostpath-pod
spec:
  hostNetwork: true
  volumes:
    - name: host
      hostPath:
        path: /etc
`
	if err := os.WriteFile(filepath.Join(tmpDir, "hostpath-pod.yaml"), []byte(hostPathPod), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner(tmpDir, SeverityLow, false)
	findings := s.runBuiltinChecks()

	if len(findings) == 0 {
		t.Error("Expected to find security issues")
	}

	// Check for specific findings
	foundPrivileged := false
	foundRunAsRoot := false
	foundSecret := false
	foundHostPath := false
	foundHostNetwork := false

	for _, f := range findings {
		switch f.ID {
		case "GITOPSI-001":
			foundPrivileged = true
		case "GITOPSI-002":
			foundRunAsRoot = true
		case "GITOPSI-004":
			foundSecret = true
		case "GITOPSI-005":
			foundHostPath = true
		case "GITOPSI-006":
			foundHostNetwork = true
		}
	}

	if !foundPrivileged {
		t.Error("Should detect privileged container")
	}
	if !foundRunAsRoot {
		t.Error("Should detect runAsRoot")
	}
	if !foundSecret {
		t.Error("Should detect hardcoded secret")
	}
	if !foundHostPath {
		t.Error("Should detect hostPath")
	}
	if !foundHostNetwork {
		t.Error("Should detect hostNetwork")
	}
}

func TestScanner_CalculateSummary(t *testing.T) {
	s := NewScanner("/tmp", SeverityLow, false)
	findings := []Finding{
		{Severity: SeverityCritical},
		{Severity: SeverityCritical},
		{Severity: SeverityHigh},
		{Severity: SeverityHigh},
		{Severity: SeverityHigh},
		{Severity: SeverityMedium},
		{Severity: SeverityLow},
		{Severity: SeverityInfo},
		{Severity: SeverityInfo},
	}

	summary := s.calculateSummary(findings)

	if summary.Critical != 2 {
		t.Errorf("Critical = %d, want 2", summary.Critical)
	}
	if summary.High != 3 {
		t.Errorf("High = %d, want 3", summary.High)
	}
	if summary.Medium != 1 {
		t.Errorf("Medium = %d, want 1", summary.Medium)
	}
	if summary.Low != 1 {
		t.Errorf("Low = %d, want 1", summary.Low)
	}
	if summary.Info != 2 {
		t.Errorf("Info = %d, want 2", summary.Info)
	}
	if summary.FailedChecks != 9 {
		t.Errorf("FailedChecks = %d, want 9", summary.FailedChecks)
	}
}

func TestScanResult_ShouldFail(t *testing.T) {
	tests := []struct {
		name       string
		summary    ScanSummary
		minSev     ScanSeverity
		shouldFail bool
	}{
		{
			name:       "critical with critical threshold",
			summary:    ScanSummary{Critical: 1},
			minSev:     SeverityCritical,
			shouldFail: true,
		},
		{
			name:       "high with critical threshold",
			summary:    ScanSummary{High: 1},
			minSev:     SeverityCritical,
			shouldFail: false,
		},
		{
			name:       "high with high threshold",
			summary:    ScanSummary{High: 1},
			minSev:     SeverityHigh,
			shouldFail: true,
		},
		{
			name:       "medium with high threshold",
			summary:    ScanSummary{Medium: 1},
			minSev:     SeverityHigh,
			shouldFail: false,
		},
		{
			name:       "medium with medium threshold",
			summary:    ScanSummary{Medium: 1},
			minSev:     SeverityMedium,
			shouldFail: true,
		},
		{
			name:       "low with medium threshold",
			summary:    ScanSummary{Low: 1},
			minSev:     SeverityMedium,
			shouldFail: false,
		},
		{
			name:       "clean with low threshold",
			summary:    ScanSummary{},
			minSev:     SeverityLow,
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ScanResult{Summary: tt.summary}
			if got := result.ShouldFail(tt.minSev); got != tt.shouldFail {
				t.Errorf("ShouldFail() = %v, want %v", got, tt.shouldFail)
			}
		})
	}
}

func TestGenerateProvenance(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "provenance-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte("test: content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test2.yaml"), []byte("another: file"), 0644); err != nil {
		t.Fatal(err)
	}

	prov, err := GenerateProvenance(tmpDir, "", "1.0.0")
	if err != nil {
		t.Fatalf("GenerateProvenance failed: %v", err)
	}

	if prov.Generator != "gitopsi" {
		t.Errorf("Generator = %v, want gitopsi", prov.Generator)
	}
	if prov.GeneratorVersion != "1.0.0" {
		t.Errorf("GeneratorVersion = %v, want 1.0.0", prov.GeneratorVersion)
	}
	if len(prov.Files) != 2 {
		t.Errorf("Files count = %d, want 2", len(prov.Files))
	}
	if prov.OutputHash == "" {
		t.Error("OutputHash should not be empty")
	}
	if prov.BuildTimestamp.IsZero() {
		t.Error("BuildTimestamp should not be zero")
	}
}

func TestProvenance_ToJSON(t *testing.T) {
	prov := &Provenance{
		Generator:        "gitopsi",
		GeneratorVersion: "1.0.0",
		GitCommit:        "abc123",
	}

	json, err := prov.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if json == "" {
		t.Error("JSON output should not be empty")
	}
	if !contains(json, "gitopsi") {
		t.Error("JSON should contain generator name")
	}
	if !contains(json, "abc123") {
		t.Error("JSON should contain git commit")
	}
}

func TestProvenance_Save(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "provenance-*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	prov := &Provenance{
		Generator:        "gitopsi",
		GeneratorVersion: "1.0.0",
	}

	if err := prov.Save(tmpFile.Name()); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !contains(string(content), "gitopsi") {
		t.Error("Saved file should contain generator name")
	}
}

func TestNewSanitizer(t *testing.T) {
	s := NewSanitizer()
	if s == nil {
		t.Fatal("NewSanitizer returned nil")
	}
	if s.maxLength != 1024 {
		t.Errorf("maxLength = %d, want 1024", s.maxLength)
	}
}

func TestSanitizer_SanitizeName(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		input string
		want  string
	}{
		{"my-app", "my-app"},
		{"My_App", "my-app"},
		{"my.app.name", "my-app-name"},
		{"MY APP", "my-app"},
		{"-leading-dash", "leading-dash"},
		{"trailing-dash-", "trailing-dash"},
		{"verylongnamethatshouldbetruncatedtosixtythreecharactersbecausek8s", "verylongnamethatshouldbetruncatedtosixtythreecharactersbecausek"},
	}

	for _, tt := range tests {
		got := s.SanitizeName(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizer_SanitizeLabel(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		input string
		want  string
	}{
		{"my-label", "my-label"},
		{"my_label", "my_label"},
		{"my.label", "my.label"},
		{"my label", "my-label"},
		{"my@label", "my-label"},
	}

	for _, tt := range tests {
		got := s.SanitizeLabel(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeLabel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizer_SanitizePath(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		input string
		want  string
	}{
		{"path/to/file", "path/to/file"},
		{"../etc/passwd", "etc/passwd"},
		{"path/../file", "path/file"}, // filepath.Clean normalizes but doesn't eliminate parent dir
		{"//path//file", "path/file"},
		{"/absolute/path", "absolute/path"},
	}

	for _, tt := range tests {
		got := s.SanitizePath(tt.input)
		if got != tt.want {
			t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizer_SanitizeURL(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/user/repo", "https://github.com/user/repo"},
		{"https://github.com/user/repo.git", "https://github.com/user/repo.git"},
		{"git@github.com:user/repo.git", "git@github.com:user/repo.git"},
	}

	for _, tt := range tests {
		got := s.SanitizeURL(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizer_SanitizeImage(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		input string
		want  string
	}{
		{"nginx:latest", "nginx:latest"},
		{"docker.io/library/nginx:1.25", "docker.io/library/nginx:1.25"},
		{"gcr.io/project/image@sha256:abc123", "gcr.io/project/image@sha256:abc123"},
		{"quay.io/org/image:v1.0.0", "quay.io/org/image:v1.0.0"},
	}

	for _, tt := range tests {
		got := s.SanitizeImage(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeImage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizer_ValidateInput(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"normal-input", false},
		{"my-app-name", false},
		{"{{ .Name }}", true},         // Template injection
		{"test; rm -rf /", true},      // Command injection
		{"test && echo hacked", true}, // Command injection
		{"test || echo hacked", true}, // Command injection
		{"$(whoami)", true},           // Command substitution
		{"`whoami`", true},            // Backtick command substitution
		{"test > /dev/null", true},    // Redirection
		{"test | cat", true},          // Pipe
	}

	for _, tt := range tests {
		err := s.ValidateInput(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestHashFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "hash-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "test content for hashing"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	hash, err := hashFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("hashFile failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}
	if len(hash) != 64 { // SHA256 hex length
		t.Errorf("Hash length = %d, want 64", len(hash))
	}
}

func TestFinding_Fields(t *testing.T) {
	f := Finding{
		ID:          "TEST-001",
		Title:       "Test Finding",
		Description: "A test security finding",
		Severity:    SeverityHigh,
		Category:    "test",
		File:        "test.yaml",
		Line:        42,
		Remediation: "Fix the issue",
		References:  []string{"https://example.com"},
	}

	if f.ID != "TEST-001" {
		t.Errorf("ID = %v, want TEST-001", f.ID)
	}
	if f.Severity != SeverityHigh {
		t.Errorf("Severity = %v, want HIGH", f.Severity)
	}
	if f.Line != 42 {
		t.Errorf("Line = %d, want 42", f.Line)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
