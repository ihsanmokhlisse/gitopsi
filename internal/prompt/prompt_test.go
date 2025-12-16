package prompt

import (
	"errors"
	"testing"

	"github.com/AlecAivazis/survey/v2"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
)

type MockPrompter struct {
	AskFunc    func(qs []*survey.Question, response interface{}) error
	AskOneFunc func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error
	CallCount  int
}

func (m *MockPrompter) Ask(qs []*survey.Question, response interface{}) error {
	m.CallCount++
	if m.AskFunc != nil {
		return m.AskFunc(qs, response)
	}
	return nil
}

func (m *MockPrompter) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	m.CallCount++
	if m.AskOneFunc != nil {
		return m.AskOneFunc(p, response, opts...)
	}
	return nil
}

func TestRunWithMock_Success(t *testing.T) {
	askCallCount := 0
	askOneCallCount := 0

	mock := &MockPrompter{
		AskFunc: func(qs []*survey.Question, response interface{}) error {
			askCallCount++
			answers := response.(*struct {
				ProjectName string
				Platform    string
				Scope       string
				GitOpsTool  string
				OutputType  string
			})
			answers.ProjectName = "test-project"
			answers.Platform = "kubernetes"
			answers.Scope = "both"
			answers.GitOpsTool = "argocd"
			answers.OutputType = "local"
			return nil
		},
		AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			askOneCallCount++
			switch v := response.(type) {
			case *[]string:
				*v = []string{"dev", "prod"}
			case *bool:
				*v = true
			}
			return nil
		},
	}

	cfg, err := RunWith(mock)
	if err != nil {
		t.Fatalf("RunWith() error = %v", err)
	}

	if cfg.Project.Name != "test-project" {
		t.Errorf("Project.Name = %s, want test-project", cfg.Project.Name)
	}

	if cfg.Platform != "kubernetes" {
		t.Errorf("Platform = %s, want kubernetes", cfg.Platform)
	}

	if cfg.Scope != "both" {
		t.Errorf("Scope = %s, want both", cfg.Scope)
	}

	if cfg.GitOpsTool != "argocd" {
		t.Errorf("GitOpsTool = %s, want argocd", cfg.GitOpsTool)
	}

	if len(cfg.Environments) != 2 {
		t.Errorf("Environments count = %d, want 2", len(cfg.Environments))
	}

	if !cfg.Docs.Readme {
		t.Error("Docs.Readme should be true")
	}

	if askCallCount != 1 {
		t.Errorf("Ask() called %d times, want 1", askCallCount)
	}

	if askOneCallCount != 2 {
		t.Errorf("AskOne() called %d times, want 2", askOneCallCount)
	}
}

func TestRunWithMock_GitOutput(t *testing.T) {
	askOneCallCount := 0

	mock := &MockPrompter{
		AskFunc: func(qs []*survey.Question, response interface{}) error {
			answers := response.(*struct {
				ProjectName string
				Platform    string
				Scope       string
				GitOpsTool  string
				OutputType  string
			})
			answers.ProjectName = "git-project"
			answers.Platform = "kubernetes"
			answers.Scope = "both"
			answers.GitOpsTool = "argocd"
			answers.OutputType = "git"
			return nil
		},
		AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			askOneCallCount++
			switch v := response.(type) {
			case *string:
				*v = "https://github.com/test/repo.git"
			case *[]string:
				*v = []string{"dev"}
			case *bool:
				*v = true
			}
			return nil
		},
	}

	cfg, err := RunWith(mock)
	if err != nil {
		t.Fatalf("RunWith() error = %v", err)
	}

	if cfg.Output.Type != "git" {
		t.Errorf("Output.Type = %s, want git", cfg.Output.Type)
	}

	if cfg.Output.URL != "https://github.com/test/repo.git" {
		t.Errorf("Output.URL = %s, want https://github.com/test/repo.git", cfg.Output.URL)
	}

	if askOneCallCount != 3 {
		t.Errorf("AskOne() called %d times, want 3 (git URL + envs + docs)", askOneCallCount)
	}
}

func TestRunWithMock_AskError(t *testing.T) {
	expectedErr := errors.New("user canceled")

	mock := &MockPrompter{
		AskFunc: func(qs []*survey.Question, response interface{}) error {
			return expectedErr
		},
	}

	cfg, err := RunWith(mock)
	if err == nil {
		t.Error("RunWith() should return error")
	}

	if cfg != nil {
		t.Error("RunWith() should return nil config on error")
	}

	if err != expectedErr {
		t.Errorf("RunWith() error = %v, want %v", err, expectedErr)
	}
}

func TestRunWithMock_AskOneEnvError(t *testing.T) {
	expectedErr := errors.New("env selection failed")

	mock := &MockPrompter{
		AskFunc: func(qs []*survey.Question, response interface{}) error {
			answers := response.(*struct {
				ProjectName string
				Platform    string
				Scope       string
				GitOpsTool  string
				OutputType  string
			})
			answers.ProjectName = "test"
			answers.Platform = "kubernetes"
			answers.Scope = "both"
			answers.GitOpsTool = "argocd"
			answers.OutputType = "local"
			return nil
		},
		AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return expectedErr
		},
	}

	_, err := RunWith(mock)
	if err != expectedErr {
		t.Errorf("RunWith() error = %v, want %v", err, expectedErr)
	}
}

func TestRunWithMock_GitURLError(t *testing.T) {
	expectedErr := errors.New("git url input failed")

	mock := &MockPrompter{
		AskFunc: func(qs []*survey.Question, response interface{}) error {
			answers := response.(*struct {
				ProjectName string
				Platform    string
				Scope       string
				GitOpsTool  string
				OutputType  string
			})
			answers.ProjectName = "test"
			answers.Platform = "kubernetes"
			answers.Scope = "both"
			answers.GitOpsTool = "argocd"
			answers.OutputType = "git"
			return nil
		},
		AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return expectedErr
		},
	}

	_, err := RunWith(mock)
	if err != expectedErr {
		t.Errorf("RunWith() error = %v, want %v", err, expectedErr)
	}
}

func TestRunWithMock_DocsError(t *testing.T) {
	expectedErr := errors.New("docs prompt failed")
	callCount := 0

	mock := &MockPrompter{
		AskFunc: func(qs []*survey.Question, response interface{}) error {
			answers := response.(*struct {
				ProjectName string
				Platform    string
				Scope       string
				GitOpsTool  string
				OutputType  string
			})
			answers.ProjectName = "test"
			answers.Platform = "kubernetes"
			answers.Scope = "both"
			answers.GitOpsTool = "argocd"
			answers.OutputType = "local"
			return nil
		},
		AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			callCount++
			if callCount == 1 {
				if v, ok := response.(*[]string); ok {
					*v = []string{"dev"}
				}
				return nil
			}
			return expectedErr
		},
	}

	_, err := RunWith(mock)
	if err != expectedErr {
		t.Errorf("RunWith() error = %v, want %v", err, expectedErr)
	}
}

func TestRunWithMock_AllPlatforms(t *testing.T) {
	platforms := config.ValidPlatforms()

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			mock := &MockPrompter{
				AskFunc: func(qs []*survey.Question, response interface{}) error {
					answers := response.(*struct {
						ProjectName string
						Platform    string
						Scope       string
						GitOpsTool  string
						OutputType  string
					})
					answers.ProjectName = "test"
					answers.Platform = platform
					answers.Scope = "both"
					answers.GitOpsTool = "argocd"
					answers.OutputType = "local"
					return nil
				},
				AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
					switch v := response.(type) {
					case *[]string:
						*v = []string{"dev"}
					case *bool:
						*v = true
					}
					return nil
				},
			}

			cfg, err := RunWith(mock)
			if err != nil {
				t.Fatalf("RunWith() error = %v", err)
			}

			if cfg.Platform != platform {
				t.Errorf("Platform = %s, want %s", cfg.Platform, platform)
			}
		})
	}
}

func TestRunWithMock_AllScopes(t *testing.T) {
	scopes := config.ValidScopes()

	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			mock := &MockPrompter{
				AskFunc: func(qs []*survey.Question, response interface{}) error {
					answers := response.(*struct {
						ProjectName string
						Platform    string
						Scope       string
						GitOpsTool  string
						OutputType  string
					})
					answers.ProjectName = "test"
					answers.Platform = "kubernetes"
					answers.Scope = scope
					answers.GitOpsTool = "argocd"
					answers.OutputType = "local"
					return nil
				},
				AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
					switch v := response.(type) {
					case *[]string:
						*v = []string{"dev"}
					case *bool:
						*v = true
					}
					return nil
				},
			}

			cfg, err := RunWith(mock)
			if err != nil {
				t.Fatalf("RunWith() error = %v", err)
			}

			if cfg.Scope != scope {
				t.Errorf("Scope = %s, want %s", cfg.Scope, scope)
			}
		})
	}
}

func TestRunWithMock_AllGitOpsTools(t *testing.T) {
	tools := config.ValidGitOpsTools()

	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			mock := &MockPrompter{
				AskFunc: func(qs []*survey.Question, response interface{}) error {
					answers := response.(*struct {
						ProjectName string
						Platform    string
						Scope       string
						GitOpsTool  string
						OutputType  string
					})
					answers.ProjectName = "test"
					answers.Platform = "kubernetes"
					answers.Scope = "both"
					answers.GitOpsTool = tool
					answers.OutputType = "local"
					return nil
				},
				AskOneFunc: func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
					switch v := response.(type) {
					case *[]string:
						*v = []string{"dev"}
					case *bool:
						*v = true
					}
					return nil
				},
			}

			cfg, err := RunWith(mock)
			if err != nil {
				t.Fatalf("RunWith() error = %v", err)
			}

			if cfg.GitOpsTool != tool {
				t.Errorf("GitOpsTool = %s, want %s", cfg.GitOpsTool, tool)
			}
		})
	}
}

func TestSurveyPrompterImplementsInterface(t *testing.T) {
	var _ Prompter = &SurveyPrompter{}
}

func TestDefaultPrompterNotNil(t *testing.T) {
	if DefaultPrompter == nil {
		t.Error("DefaultPrompter should not be nil")
	}
}

func TestValidPlatformsUsedInPrompt(t *testing.T) {
	platforms := config.ValidPlatforms()
	if len(platforms) == 0 {
		t.Error("ValidPlatforms() should return platforms")
	}

	expected := []string{"kubernetes", "openshift", "aks", "eks"}
	for _, p := range expected {
		found := false
		for _, v := range platforms {
			if v == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing platform: %s", p)
		}
	}
}

func TestValidScopesUsedInPrompt(t *testing.T) {
	scopes := config.ValidScopes()
	if len(scopes) == 0 {
		t.Error("ValidScopes() should return scopes")
	}

	expected := []string{"infrastructure", "application", "both"}
	for _, s := range expected {
		found := false
		for _, v := range scopes {
			if v == s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing scope: %s", s)
		}
	}
}

func TestValidGitOpsToolsUsedInPrompt(t *testing.T) {
	tools := config.ValidGitOpsTools()
	if len(tools) == 0 {
		t.Error("ValidGitOpsTools() should return tools")
	}

	expected := []string{"argocd", "flux", "both"}
	for _, tool := range expected {
		found := false
		for _, v := range tools {
			if v == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing tool: %s", tool)
		}
	}
}

func TestDefaultConfigValuesForPrompt(t *testing.T) {
	cfg := config.NewDefaultConfig()

	if cfg.Platform != "kubernetes" {
		t.Errorf("Default platform should be kubernetes, got %s", cfg.Platform)
	}

	if cfg.Scope != "both" {
		t.Errorf("Default scope should be both, got %s", cfg.Scope)
	}

	if cfg.GitOpsTool != "argocd" {
		t.Errorf("Default gitops tool should be argocd, got %s", cfg.GitOpsTool)
	}

	if len(cfg.Environments) != 3 {
		t.Errorf("Default should have 3 environments, got %d", len(cfg.Environments))
	}
}
