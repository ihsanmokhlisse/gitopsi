package templates

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed all:files
var FS embed.FS

func Render(name string, data any) ([]byte, error) {
	content, err := FS.ReadFile("files/" + name)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

func RenderString(tmplContent string, data any) ([]byte, error) {
	tmpl, err := template.New("inline").Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func List() ([]string, error) {
	entries, err := FS.ReadDir("files")
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

