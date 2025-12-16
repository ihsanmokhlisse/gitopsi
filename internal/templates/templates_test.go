package templates

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     any
		contains string
		wantErr  bool
	}{
		{
			name:     "kubernetes deployment",
			template: "kubernetes/deployment.yaml.tmpl",
			data: map[string]interface{}{
				"Name":     "test-app",
				"Image":    "nginx:latest",
				"Port":     80,
				"Replicas": 3,
			},
			contains: "name: test-app",
			wantErr:  false,
		},
		{
			name:     "kubernetes service",
			template: "kubernetes/service.yaml.tmpl",
			data: map[string]interface{}{
				"Name": "test-svc",
				"Port": 8080,
			},
			contains: "name: test-svc",
			wantErr:  false,
		},
		{
			name:     "kubernetes kustomization",
			template: "kubernetes/kustomization.yaml.tmpl",
			data: map[string]interface{}{
				"Resources": []string{"deployment.yaml", "service.yaml"},
			},
			contains: "kind: Kustomization",
			wantErr:  false,
		},
		{
			name:     "infrastructure namespace",
			template: "infrastructure/namespace.yaml.tmpl",
			data: map[string]string{
				"Name": "test-ns",
				"Env":  "dev",
			},
			contains: "name: test-ns",
			wantErr:  false,
		},
		{
			name:     "argocd application",
			template: "argocd/application.yaml.tmpl",
			data: map[string]string{
				"Name":      "test-app",
				"Project":   "default",
				"RepoURL":   "https://github.com/test/repo.git",
				"Path":      "apps/test",
				"Namespace": "test-ns",
			},
			contains: "kind: Application",
			wantErr:  false,
		},
		{
			name:     "argocd project",
			template: "argocd/project.yaml.tmpl",
			data: map[string]string{
				"Name":        "test-project",
				"Description": "Test project",
			},
			contains: "kind: AppProject",
			wantErr:  false,
		},
		{
			name:     "non-existent template",
			template: "nonexistent/template.yaml.tmpl",
			data:     nil,
			contains: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(result), tt.contains) {
				t.Errorf("Render() result does not contain %q, got: %s", tt.contains, string(result))
			}
		})
	}
}

func TestRenderString(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     any
		expected string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.Name}}!",
			data:     map[string]string{"Name": "World"},
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:     "template with number",
			template: "Port: {{.Port}}",
			data:     map[string]int{"Port": 8080},
			expected: "Port: 8080",
			wantErr:  false,
		},
		{
			name:     "invalid template syntax",
			template: "{{.Invalid",
			data:     nil,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "template with range",
			template: "{{range .Items}}{{.}} {{end}}",
			data:     map[string][]string{"Items": {"a", "b", "c"}},
			expected: "a b c ",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderString(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(result) != tt.expected {
				t.Errorf("RenderString() = %q, expected %q", string(result), tt.expected)
			}
		})
	}
}

func TestList(t *testing.T) {
	names, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(names) == 0 {
		t.Error("List() returned empty list")
	}

	expectedDirs := []string{"argocd", "docs", "infrastructure", "kubernetes"}
	for _, dir := range expectedDirs {
		found := false
		for _, name := range names {
			if name == dir {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() missing expected directory: %s", dir)
		}
	}
}

func TestRenderDocsReadme(t *testing.T) {
	data := map[string]interface{}{
		"Project": map[string]string{
			"Name":        "test-project",
			"Description": "A test project",
		},
		"Platform":   "kubernetes",
		"Scope":      "both",
		"GitOpsTool": "argocd",
		"Environments": []map[string]string{
			{"Name": "dev"},
			{"Name": "prod"},
		},
	}

	result, err := Render("docs/README.md.tmpl", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	checks := []string{
		"test-project",
		"kubernetes",
		"argocd",
		"dev",
		"prod",
	}

	for _, check := range checks {
		if !strings.Contains(string(result), check) {
			t.Errorf("README.md does not contain %q", check)
		}
	}
}

