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

func TestRenderWithMissingData(t *testing.T) {
	result, err := Render("kubernetes/deployment.yaml.tmpl", map[string]interface{}{
		"Name": "test",
	})
	if err != nil {
		t.Logf("Render with missing data: %v", err)
	}

	if result != nil && strings.Contains(string(result), "name: test") {
		t.Log("Template rendered with partial data")
	}
}

func TestRenderAllTemplateTypes(t *testing.T) {
	templates := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "kubernetes/deployment.yaml.tmpl",
			data: map[string]interface{}{
				"Name": "app", "Image": "nginx", "Port": 80, "Replicas": 1,
			},
		},
		{
			name: "kubernetes/service.yaml.tmpl",
			data: map[string]interface{}{
				"Name": "svc", "Port": 80,
			},
		},
		{
			name: "kubernetes/kustomization.yaml.tmpl",
			data: map[string]interface{}{
				"Resources": []string{"a.yaml", "b.yaml"},
			},
		},
		{
			name: "infrastructure/namespace.yaml.tmpl",
			data: map[string]interface{}{
				"Name": "ns", "Env": "dev",
			},
		},
		{
			name: "argocd/application.yaml.tmpl",
			data: map[string]interface{}{
				"Name": "app", "Project": "default", "RepoURL": "url",
				"Path": "path", "Namespace": "ns",
			},
		},
		{
			name: "argocd/project.yaml.tmpl",
			data: map[string]interface{}{
				"Name": "proj", "Description": "desc",
			},
		},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.name, tt.data)
			if err != nil {
				t.Fatalf("Render(%s) error = %v", tt.name, err)
			}
			if len(result) == 0 {
				t.Errorf("Render(%s) returned empty result", tt.name)
			}
		})
	}
}

func TestRenderStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     interface{}
		wantErr  bool
	}{
		{
			name:     "empty template",
			template: "",
			data:     nil,
			wantErr:  false,
		},
		{
			name:     "template with no variables",
			template: "static content",
			data:     nil,
			wantErr:  false,
		},
		{
			name:     "template with conditional",
			template: "{{if .Enabled}}yes{{else}}no{{end}}",
			data:     map[string]bool{"Enabled": true},
			wantErr:  false,
		},
		{
			name:     "template with nil data",
			template: "Hello",
			data:     nil,
			wantErr:  false,
		},
		{
			name:     "nested template data",
			template: "{{.Outer.Inner}}",
			data: map[string]interface{}{
				"Outer": map[string]string{"Inner": "value"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RenderString(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListDirectories(t *testing.T) {
	names, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(names) < 4 {
		t.Errorf("Expected at least 4 template directories, got %d", len(names))
	}

	for _, name := range names {
		if name == "" {
			t.Error("List() returned empty name")
		}
	}
}

func TestRenderDeploymentAllFields(t *testing.T) {
	data := map[string]interface{}{
		"Name":     "full-app",
		"Image":    "myregistry/myapp:v1.2.3",
		"Port":     8080,
		"Replicas": 5,
	}

	result, err := Render("kubernetes/deployment.yaml.tmpl", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	content := string(result)
	checks := []string{
		"apiVersion: apps/v1",
		"kind: Deployment",
		"name: full-app",
		"replicas: 5",
		"image: myregistry/myapp:v1.2.3",
		"containerPort: 8080",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Deployment missing: %s", check)
		}
	}
}

func TestRenderServiceAllFields(t *testing.T) {
	data := map[string]interface{}{
		"Name": "full-service",
		"Port": 3000,
	}

	result, err := Render("kubernetes/service.yaml.tmpl", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	content := string(result)
	checks := []string{
		"apiVersion: v1",
		"kind: Service",
		"name: full-service",
		"port: 3000",
		"targetPort: 3000",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Service missing: %s", check)
		}
	}
}

