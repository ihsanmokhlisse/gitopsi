package testutil

import (
	"net/http"
	"strings"
	"testing"
)

func TestMockHTTPClient_Do(t *testing.T) {
	t.Run("with DoFunc set", func(t *testing.T) {
		client := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return MockHTTPResponse(http.StatusOK, "test body"), nil
			},
		}

		req, _ := http.NewRequest("GET", "http://example.com", nil)
		resp, err := client.Do(req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("without DoFunc set", func(t *testing.T) {
		client := &MockHTTPClient{}

		req, _ := http.NewRequest("GET", "http://example.com", nil)
		_, err := client.Do(req)

		if err == nil {
			t.Error("expected error when DoFunc not set")
		}
	})
}

func TestMockHTTPResponse(t *testing.T) {
	resp := MockHTTPResponse(http.StatusNotFound, "not found")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	buf := make([]byte, 100)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if body != "not found" {
		t.Errorf("Body = %s, want 'not found'", body)
	}
}

func TestMockRegistryServer(t *testing.T) {
	server := NewMockRegistryServer()
	defer server.Close()

	server.SetDefaultResponses()

	t.Run("index endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/index.yaml")
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("pattern endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/patterns/test/1.0.0/pattern.yaml")
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/nonexistent")
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("error status", func(t *testing.T) {
		server.StatusCode = http.StatusInternalServerError
		server.ErrorResponse = "internal error"

		resp, err := http.Get(server.URL + "/index.yaml")
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
		}

		// Reset for other tests
		server.StatusCode = http.StatusOK
	})
}

func TestMockPattern_ToYAML(t *testing.T) {
	pattern := &MockPattern{
		Name:        "test-pattern",
		Version:     "1.0.0",
		Description: "Test description",
		Category:    "security",
		Tags:        []string{"test", "security"},
	}

	yaml, err := pattern.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if !strings.Contains(yaml, "name: test-pattern") {
		t.Error("YAML should contain pattern name")
	}
	if !strings.Contains(yaml, "version: 1.0.0") {
		t.Error("YAML should contain version")
	}
	if !strings.Contains(yaml, "category: security") {
		t.Error("YAML should contain category")
	}
}

func TestMockIndex_ToYAML(t *testing.T) {
	index := &MockIndex{
		Version: "1.0",
		Patterns: []MockIndexEntry{
			{
				Name:        "pattern1",
				Description: "First pattern",
				Category:    "observability",
				Latest:      "1.0.0",
				Versions:    []string{"1.0.0"},
			},
			{
				Name:        "pattern2",
				Description: "Second pattern",
				Category:    "security",
				Latest:      "2.0.0",
				Versions:    []string{"1.0.0", "2.0.0"},
			},
		},
	}

	yaml, err := index.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if !strings.Contains(yaml, "version: \"1.0\"") {
		t.Error("YAML should contain version")
	}
	if !strings.Contains(yaml, "name: pattern1") {
		t.Error("YAML should contain pattern1")
	}
	if !strings.Contains(yaml, "name: pattern2") {
		t.Error("YAML should contain pattern2")
	}
}

func TestMockWriter(t *testing.T) {
	writer := NewMockWriter()

	t.Run("WriteFile", func(t *testing.T) {
		err := writer.WriteFile("test/file.yaml", "content")
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if !writer.FileExists("test/file.yaml") {
			t.Error("File should exist after WriteFile")
		}

		content, ok := writer.GetFileContent("test/file.yaml")
		if !ok {
			t.Error("GetFileContent should return true")
		}
		if content != "content" {
			t.Errorf("Content = %s, want 'content'", content)
		}
	})

	t.Run("CreateDir", func(t *testing.T) {
		err := writer.CreateDir("test/dir")
		if err != nil {
			t.Fatalf("CreateDir() error = %v", err)
		}

		if !writer.DirExists("test/dir") {
			t.Error("Directory should exist after CreateDir")
		}
	})

	t.Run("WriteError", func(t *testing.T) {
		writer.WriteError = http.ErrHandlerTimeout
		err := writer.WriteFile("error/file.yaml", "content")
		if err == nil {
			t.Error("WriteFile should return error when WriteError is set")
		}
		writer.WriteError = nil
	})

	t.Run("DirError", func(t *testing.T) {
		writer.DirError = http.ErrHandlerTimeout
		err := writer.CreateDir("error/dir")
		if err == nil {
			t.Error("CreateDir should return error when DirError is set")
		}
		writer.DirError = nil
	})

	t.Run("Reset", func(t *testing.T) {
		writer.WriteFile("another/file.yaml", "content")
		writer.CreateDir("another/dir")

		writer.Reset()

		if writer.FileExists("another/file.yaml") {
			t.Error("File should not exist after Reset")
		}
		if writer.DirExists("another/dir") {
			t.Error("Directory should not exist after Reset")
		}
	})

	t.Run("FileNotExists", func(t *testing.T) {
		writer.Reset()
		if writer.FileExists("nonexistent") {
			t.Error("FileExists should return false for non-existent file")
		}
	})

	t.Run("DirNotExists", func(t *testing.T) {
		writer.Reset()
		if writer.DirExists("nonexistent") {
			t.Error("DirExists should return false for non-existent directory")
		}
	})
}

func TestTestContext(t *testing.T) {
	ctx := TestContext(t)
	if ctx == nil {
		t.Error("TestContext should return non-nil context")
	}
}

func TestMockConfig_ToConfigYAML(t *testing.T) {
	cfg := &MockConfig{
		ProjectName:  "my-project",
		Platform:     "kubernetes",
		Scope:        "both",
		GitOpsTool:   "argocd",
		GitURL:       "https://github.com/org/repo.git",
		Environments: []string{"dev", "staging", "prod"},
	}

	yaml, err := cfg.ToConfigYAML()
	if err != nil {
		t.Fatalf("ToConfigYAML() error = %v", err)
	}

	if !strings.Contains(yaml, "name: my-project") {
		t.Error("YAML should contain project name")
	}
	if !strings.Contains(yaml, "platform: kubernetes") {
		t.Error("YAML should contain platform")
	}
	if !strings.Contains(yaml, "scope: both") {
		t.Error("YAML should contain scope")
	}
	if !strings.Contains(yaml, "gitops_tool: argocd") {
		t.Error("YAML should contain gitops_tool")
	}
}

func TestMockConfig_EmptyEnvironments(t *testing.T) {
	cfg := &MockConfig{
		ProjectName:  "minimal",
		Environments: []string{},
	}

	yaml, err := cfg.ToConfigYAML()
	if err != nil {
		t.Fatalf("ToConfigYAML() error = %v", err)
	}

	if !strings.Contains(yaml, "environments: []") {
		t.Error("YAML should contain empty environments")
	}
}

func TestMockPattern_EmptyTags(t *testing.T) {
	pattern := &MockPattern{
		Name:    "minimal",
		Version: "0.1.0",
		Tags:    nil,
	}

	yaml, err := pattern.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if !strings.Contains(yaml, "name: minimal") {
		t.Error("YAML should contain pattern name")
	}
}

func TestMockIndex_EmptyPatterns(t *testing.T) {
	index := &MockIndex{
		Version:  "1.0",
		Patterns: []MockIndexEntry{},
	}

	yaml, err := index.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if !strings.Contains(yaml, "version:") {
		t.Error("YAML should contain version")
	}
	if !strings.Contains(yaml, "patterns: []") {
		t.Error("YAML should contain empty patterns")
	}
}

func TestNewMockWriter_Initialization(t *testing.T) {
	writer := NewMockWriter()

	if writer.Files == nil {
		t.Error("Files map should be initialized")
	}
	if writer.Directories == nil {
		t.Error("Directories map should be initialized")
	}
	if writer.WriteError != nil {
		t.Error("WriteError should be nil")
	}
	if writer.DirError != nil {
		t.Error("DirError should be nil")
	}
}

func TestMockRegistryServer_SetDefaultResponses(t *testing.T) {
	server := NewMockRegistryServer()
	defer server.Close()

	server.SetDefaultResponses()

	if server.IndexResponse == "" {
		t.Error("IndexResponse should not be empty after SetDefaultResponses")
	}
	if server.PatternResponse == "" {
		t.Error("PatternResponse should not be empty after SetDefaultResponses")
	}
	if !strings.Contains(server.IndexResponse, "test-pattern") {
		t.Error("IndexResponse should contain test-pattern")
	}
	if !strings.Contains(server.PatternResponse, "test-pattern") {
		t.Error("PatternResponse should contain test-pattern")
	}
}
