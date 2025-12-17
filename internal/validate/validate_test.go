package validate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeverity(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityCritical, "critical"},
		{SeverityHigh, "high"},
		{SeverityMedium, "medium"},
		{SeverityLow, "low"},
		{SeverityInfo, "info"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, string(tt.severity))
	}
}

func TestCategory(t *testing.T) {
	tests := []struct {
		category Category
		expected string
	}{
		{CategorySchema, "schema"},
		{CategorySecurity, "security"},
		{CategoryDeprecation, "deprecation"},
		{CategoryBestPractice, "best-practice"},
		{CategoryKustomize, "kustomize"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, string(tt.category))
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "1.29", opts.K8sVersion)
	assert.Equal(t, "2.10", opts.ArgoCDVersion)
	assert.True(t, opts.Schema)
	assert.True(t, opts.Security)
	assert.True(t, opts.Deprecation)
	assert.True(t, opts.BestPractice)
	assert.True(t, opts.Kustomize)
	assert.Equal(t, SeverityHigh, opts.FailOn)
	assert.Equal(t, "table", opts.OutputFormat)
}

func TestNew(t *testing.T) {
	t.Run("with nil options", func(t *testing.T) {
		v := New(nil)
		assert.NotNil(t, v)
		assert.NotNil(t, v.opts)
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := &Options{
			K8sVersion: "1.28",
			Schema:     true,
		}
		v := New(opts)
		assert.NotNil(t, v)
		assert.Equal(t, "1.28", v.opts.K8sVersion)
	})
}

func TestValidateNonExistentPath(t *testing.T) {
	ctx := context.Background()
	opts := &Options{
		Path: "/nonexistent/path/that/does/not/exist",
	}
	v := New(opts)

	_, err := v.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestValidateValidManifests(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`
	manifestPath := filepath.Join(tmpDir, "configmap.yaml")
	err = os.WriteFile(manifestPath, []byte(manifest), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &Options{
		Path:        tmpDir,
		K8sVersion:  "1.29",
		Schema:      true,
		Security:    false,
		Deprecation: false,
		Kustomize:   false,
	}
	v := New(opts)

	result, err := v.Validate(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalManifests)
}

func TestValidateInvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	invalidYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
	tabs-and-spaces-mixed: bad
  labels:
    app: test
`
	manifestPath := filepath.Join(tmpDir, "invalid.yaml")
	err = os.WriteFile(manifestPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &Options{
		Path:        tmpDir,
		K8sVersion:  "1.29",
		Schema:      true,
		Security:    false,
		Deprecation: false,
		Kustomize:   false,
	}
	v := New(opts)

	result, err := v.Validate(ctx)
	require.NoError(t, err)

	assert.Greater(t, len(result.Categories[CategorySchema].Issues), 0)
}

func TestValidateDeprecatedAPI(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	deprecatedManifest := `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: test
        image: nginx
`
	manifestPath := filepath.Join(tmpDir, "deprecated.yaml")
	err = os.WriteFile(manifestPath, []byte(deprecatedManifest), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &Options{
		Path:        tmpDir,
		K8sVersion:  "1.29",
		Schema:      false,
		Security:    false,
		Deprecation: true,
		Kustomize:   false,
	}
	v := New(opts)

	result, err := v.Validate(ctx)
	require.NoError(t, err)

	assert.Greater(t, len(result.Categories[CategoryDeprecation].Issues), 0)
	assert.Contains(t, result.Categories[CategoryDeprecation].Issues[0].Message, "deprecated")
}

func TestValidateSecurityIssues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	insecureManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: insecure-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: insecure
  template:
    metadata:
      labels:
        app: insecure
    spec:
      containers:
      - name: test
        image: nginx
        securityContext:
          privileged: true
          runAsUser: 0
`
	manifestPath := filepath.Join(tmpDir, "insecure.yaml")
	err = os.WriteFile(manifestPath, []byte(insecureManifest), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &Options{
		Path:        tmpDir,
		K8sVersion:  "1.29",
		Schema:      false,
		Security:    true,
		Deprecation: false,
		Kustomize:   false,
	}
	v := New(opts)

	result, err := v.Validate(ctx)
	require.NoError(t, err)

	securityIssues := result.Categories[CategorySecurity]
	assert.Greater(t, len(securityIssues.Issues), 0)
}

func TestValidateKustomize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configmap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`
	err = os.WriteFile(filepath.Join(tmpDir, "configmap.yaml"), []byte(configmap), 0644)
	require.NoError(t, err)

	kustomization := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - configmap.yaml
`
	err = os.WriteFile(filepath.Join(tmpDir, "kustomization.yaml"), []byte(kustomization), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &Options{
		Path:        tmpDir,
		K8sVersion:  "1.29",
		Schema:      false,
		Security:    false,
		Deprecation: false,
		Kustomize:   true,
	}
	v := New(opts)

	result, err := v.Validate(ctx)
	require.NoError(t, err)

	kustomizeResult := result.Categories[CategoryKustomize]
	assert.NotNil(t, kustomizeResult)
}

func TestShouldFail(t *testing.T) {
	tests := []struct {
		name     string
		failOn   Severity
		issues   []Issue
		expected bool
	}{
		{
			name:     "no issues",
			failOn:   SeverityHigh,
			issues:   []Issue{},
			expected: false,
		},
		{
			name:     "critical issue with fail on critical",
			failOn:   SeverityCritical,
			issues:   []Issue{{Severity: SeverityCritical}},
			expected: true,
		},
		{
			name:     "high issue with fail on critical",
			failOn:   SeverityCritical,
			issues:   []Issue{{Severity: SeverityHigh}},
			expected: false,
		},
		{
			name:     "high issue with fail on high",
			failOn:   SeverityHigh,
			issues:   []Issue{{Severity: SeverityHigh}},
			expected: true,
		},
		{
			name:     "medium issue with fail on high",
			failOn:   SeverityHigh,
			issues:   []Issue{{Severity: SeverityMedium}},
			expected: false,
		},
		{
			name:     "medium issue with fail on medium",
			failOn:   SeverityMedium,
			issues:   []Issue{{Severity: SeverityMedium}},
			expected: true,
		},
		{
			name:     "low issue with fail on low",
			failOn:   SeverityLow,
			issues:   []Issue{{Severity: SeverityLow}},
			expected: true,
		},
		{
			name:     "info issue with fail on low",
			failOn:   SeverityLow,
			issues:   []Issue{{Severity: SeverityInfo}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(&Options{FailOn: tt.failOn})
			result := &ValidationResult{Issues: tt.issues}
			assert.Equal(t, tt.expected, v.ShouldFail(result))
		})
	}
}

func TestValidationResultToJSON(t *testing.T) {
	result := &ValidationResult{
		Path:           "/test/path",
		TotalManifests: 10,
		Passed:         8,
		Warnings:       1,
		Failed:         1,
		Issues: []Issue{
			{
				File:     "test.yaml",
				Category: CategorySecurity,
				Severity: SeverityHigh,
				Rule:     "SEC001",
				Message:  "Test issue",
			},
		},
	}

	json, err := result.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, json, "/test/path")
	assert.Contains(t, json, "SEC001")
	assert.Contains(t, json, "Test issue")
}

func TestValidationResultToYAML(t *testing.T) {
	result := &ValidationResult{
		Path:           "/test/path",
		TotalManifests: 10,
		Passed:         8,
		Warnings:       1,
		Failed:         1,
		Issues: []Issue{
			{
				File:     "test.yaml",
				Category: CategorySecurity,
				Severity: SeverityHigh,
				Rule:     "SEC001",
				Message:  "Test issue",
			},
		},
	}

	yamlOut, err := result.ToYAML()
	require.NoError(t, err)
	assert.Contains(t, yamlOut, "path: /test/path")
	assert.Contains(t, yamlOut, "rule: SEC001")
}

func TestFindManifests(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitopsi-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	files := []string{
		filepath.Join(tmpDir, "manifest1.yaml"),
		filepath.Join(tmpDir, "manifest2.yml"),
		filepath.Join(subDir, "manifest3.yaml"),
		filepath.Join(tmpDir, "kustomization.yaml"),
		filepath.Join(tmpDir, "readme.md"),
	}

	for _, f := range files {
		err = os.WriteFile(f, []byte("test"), 0644)
		require.NoError(t, err)
	}

	v := New(&Options{Path: tmpDir})
	manifests, err := v.findManifests()
	require.NoError(t, err)

	assert.Equal(t, 3, len(manifests))
}

func TestCalculateSummary(t *testing.T) {
	v := New(&Options{})
	result := &ValidationResult{
		TotalManifests: 10,
		Issues: []Issue{
			{Severity: SeverityCritical},
			{Severity: SeverityHigh},
			{Severity: SeverityMedium},
			{Severity: SeverityLow},
			{Severity: SeverityInfo},
		},
	}

	v.calculateSummary(result)

	assert.Equal(t, 8, result.Passed)
	assert.Equal(t, 2, result.Failed)
	assert.Equal(t, 2, result.Warnings)
}
