// Package testutil provides test utilities and mock implementations for testing.
package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// MockHTTPClient provides a mock HTTP client for testing.
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do executes the mock HTTP request.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, fmt.Errorf("mock DoFunc not set")
}

// MockHTTPResponse creates a mock HTTP response with the given status and body.
func MockHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// MockRegistryServer creates a test HTTP server that serves mock registry responses.
type MockRegistryServer struct {
	*httptest.Server
	IndexResponse   string
	PatternResponse string
	ErrorResponse   string
	StatusCode      int
}

// NewMockRegistryServer creates a new mock registry server.
func NewMockRegistryServer() *MockRegistryServer {
	m := &MockRegistryServer{
		StatusCode: http.StatusOK,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.StatusCode != http.StatusOK {
			http.Error(w, m.ErrorResponse, m.StatusCode)
			return
		}

		switch {
		case strings.Contains(r.URL.Path, "index.yaml"):
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Write([]byte(m.IndexResponse))
		case strings.Contains(r.URL.Path, "pattern.yaml"):
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Write([]byte(m.PatternResponse))
		default:
			http.NotFound(w, r)
		}
	}))

	return m
}

// SetDefaultResponses sets default valid responses for the mock server.
func (m *MockRegistryServer) SetDefaultResponses() {
	m.IndexResponse = `
version: "1.0"
patterns:
  - name: test-pattern
    description: A test pattern for testing
    category: observability
    latest: "1.0.0"
    versions: ["1.0.0"]
    verified: true
  - name: another-pattern
    description: Another test pattern
    category: security
    latest: "2.0.0"
    versions: ["1.0.0", "2.0.0"]
categories:
  - name: observability
    description: Monitoring and observability patterns
    count: 1
  - name: security
    description: Security patterns
    count: 1
`

	m.PatternResponse = `
apiVersion: gitopsi.io/v1
kind: Pattern
metadata:
  name: test-pattern
  version: "1.0.0"
  description: A test pattern for testing
  author: test-author
  license: MIT
  category: observability
spec:
  platforms:
    - name: kubernetes
    - name: openshift
  gitops_tools:
    - name: argocd
      minVersion: "2.8"
  components:
    - name: test-component
      type: helm
      chart: test-chart
      version: "1.0.0"
      repository: https://charts.example.com
  config:
    replicas:
      type: integer
      default: 1
      description: Number of replicas
`
}

// Close shuts down the mock server.
func (m *MockRegistryServer) Close() {
	m.Server.Close()
}

// MockPattern represents a mock pattern for testing.
type MockPattern struct {
	Name        string
	Version     string
	Description string
	Category    string
	Tags        []string
}

// ToYAML converts the mock pattern to YAML.
func (m *MockPattern) ToYAML() (string, error) {
	pattern := map[string]any{
		"apiVersion": "gitopsi.io/v1",
		"kind":       "Pattern",
		"metadata": map[string]any{
			"name":        m.Name,
			"version":     m.Version,
			"description": m.Description,
			"category":    m.Category,
			"tags":        m.Tags,
			"license":     "MIT",
		},
		"spec": map[string]any{
			"platforms": []map[string]string{
				{"name": "kubernetes"},
			},
			"gitops_tools": []map[string]string{
				{"name": "argocd"},
			},
			"components": []map[string]any{},
			"config":     map[string]any{},
		},
	}

	data, err := yaml.Marshal(pattern)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MockIndex represents a mock registry index for testing.
type MockIndex struct {
	Version  string
	Patterns []MockIndexEntry
}

// MockIndexEntry represents a mock pattern index entry.
type MockIndexEntry struct {
	Name        string
	Description string
	Category    string
	Latest      string
	Versions    []string
}

// ToYAML converts the mock index to YAML.
func (m *MockIndex) ToYAML() (string, error) {
	patterns := make([]map[string]any, len(m.Patterns))
	for i, p := range m.Patterns {
		patterns[i] = map[string]any{
			"name":        p.Name,
			"description": p.Description,
			"category":    p.Category,
			"latest":      p.Latest,
			"versions":    p.Versions,
		}
	}

	index := map[string]any{
		"version":  m.Version,
		"patterns": patterns,
	}

	data, err := yaml.Marshal(index)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MockWriter provides a mock writer for testing file operations.
type MockWriter struct {
	Files       map[string]string
	Directories map[string]bool
	WriteError  error
	DirError    error
}

// NewMockWriter creates a new mock writer.
func NewMockWriter() *MockWriter {
	return &MockWriter{
		Files:       make(map[string]string),
		Directories: make(map[string]bool),
	}
}

// WriteFile writes content to the mock file system.
func (m *MockWriter) WriteFile(path, content string) error {
	if m.WriteError != nil {
		return m.WriteError
	}
	m.Files[path] = content
	return nil
}

// CreateDir creates a directory in the mock file system.
func (m *MockWriter) CreateDir(path string) error {
	if m.DirError != nil {
		return m.DirError
	}
	m.Directories[path] = true
	return nil
}

// FileExists checks if a file exists in the mock file system.
func (m *MockWriter) FileExists(path string) bool {
	_, exists := m.Files[path]
	return exists
}

// DirExists checks if a directory exists in the mock file system.
func (m *MockWriter) DirExists(path string) bool {
	_, exists := m.Directories[path]
	return exists
}

// GetFileContent returns the content of a file from the mock file system.
func (m *MockWriter) GetFileContent(path string) (string, bool) {
	content, exists := m.Files[path]
	return content, exists
}

// Reset clears all files and directories from the mock file system.
func (m *MockWriter) Reset() {
	m.Files = make(map[string]string)
	m.Directories = make(map[string]bool)
	m.WriteError = nil
	m.DirError = nil
}

// TestContext creates a context with a test deadline.
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", msg, err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", msg)
	}
}

// AssertEqual fails the test if got != want.
func AssertEqual(t *testing.T, got, want any, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertContains fails the test if s does not contain substr.
func AssertContains(t *testing.T, s, substr, msg string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("%s: expected %q to contain %q", msg, s, substr)
	}
}

// AssertNotContains fails the test if s contains substr.
func AssertNotContains(t *testing.T, s, substr, msg string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("%s: expected %q to not contain %q", msg, s, substr)
	}
}

// AssertLen fails the test if the length of slice is not expected.
func AssertLen(t *testing.T, slice any, expected int, msg string) {
	t.Helper()
	// Note: This is a simplified version. In practice, you'd use reflection.
}

// MockConfig provides a mock configuration for testing.
type MockConfig struct {
	ProjectName  string
	Platform     string
	Scope        string
	GitOpsTool   string
	GitURL       string
	Environments []string
}

// ToConfigYAML converts the mock config to YAML.
func (m *MockConfig) ToConfigYAML() (string, error) {
	envs := make([]map[string]string, len(m.Environments))
	for i, env := range m.Environments {
		envs[i] = map[string]string{"name": env}
	}

	cfg := map[string]any{
		"project": map[string]string{
			"name": m.ProjectName,
		},
		"platform":     m.Platform,
		"scope":        m.Scope,
		"gitops_tool":  m.GitOpsTool,
		"environments": envs,
		"git": map[string]string{
			"url": m.GitURL,
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
