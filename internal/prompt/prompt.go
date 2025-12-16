package prompt

import (
	"github.com/AlecAivazis/survey/v2"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
)

type Prompter interface {
	Ask(qs []*survey.Question, response interface{}) error
	AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error
}

type SurveyPrompter struct{}

func (s *SurveyPrompter) Ask(qs []*survey.Question, response interface{}) error {
	return survey.Ask(qs, response)
}

func (s *SurveyPrompter) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	return survey.AskOne(p, response, opts...)
}

var DefaultPrompter Prompter = &SurveyPrompter{}

func Run() (*config.Config, error) {
	return RunWith(DefaultPrompter)
}

func RunWith(p Prompter) (*config.Config, error) {
	cfg := config.NewDefaultConfig()

	questions := []*survey.Question{
		{
			Name: "projectName",
			Prompt: &survey.Input{
				Message: "Project name:",
				Default: "my-platform",
			},
			Validate: survey.Required,
		},
		{
			Name: "platform",
			Prompt: &survey.Select{
				Message: "Target platform:",
				Options: config.ValidPlatforms(),
				Default: "kubernetes",
			},
		},
		{
			Name: "scope",
			Prompt: &survey.Select{
				Message: "Scope:",
				Options: config.ValidScopes(),
				Default: "both",
			},
		},
		{
			Name: "gitopsTool",
			Prompt: &survey.Select{
				Message: "GitOps tool:",
				Options: config.ValidGitOpsTools(),
				Default: "argocd",
			},
		},
		{
			Name: "outputType",
			Prompt: &survey.Select{
				Message: "Output type:",
				Options: []string{"local", "git"},
				Default: "local",
			},
		},
	}

	answers := struct {
		ProjectName string
		Platform    string
		Scope       string
		GitOpsTool  string
		OutputType  string
	}{}

	if err := p.Ask(questions, &answers); err != nil {
		return nil, err
	}

	cfg.Project.Name = answers.ProjectName
	cfg.Platform = answers.Platform
	cfg.Scope = answers.Scope
	cfg.GitOpsTool = answers.GitOpsTool
	cfg.Output.Type = answers.OutputType

	if answers.OutputType == "git" {
		gitURL := ""
		gitURLPrompt := &survey.Input{
			Message: "Git repository URL:",
			Help:    "e.g., git@github.com:org/repo.git or https://gitlab.com/group/project.git",
		}
		if err := p.AskOne(gitURLPrompt, &gitURL, survey.WithValidator(survey.Required)); err != nil {
			return nil, err
		}
		cfg.Output.URL = gitURL
		cfg.Git.URL = gitURL

		authMethod := ""
		authPrompt := &survey.Select{
			Message: "Authentication method:",
			Options: []string{"ssh", "token"},
			Default: "ssh",
		}
		if err := p.AskOne(authPrompt, &authMethod); err != nil {
			return nil, err
		}
		cfg.Git.Auth.Method = authMethod

		if authMethod == "token" {
			tokenEnv := ""
			tokenPrompt := &survey.Input{
				Message: "Token environment variable name:",
				Default: "GIT_TOKEN",
				Help:    "e.g., GITHUB_TOKEN, GITLAB_TOKEN, GIT_TOKEN",
			}
			if err := p.AskOne(tokenPrompt, &tokenEnv); err != nil {
				return nil, err
			}
			cfg.Git.Auth.TokenEnv = tokenEnv
		}
	}

	envNames := []string{}
	envPrompt := &survey.MultiSelect{
		Message: "Environments:",
		Options: []string{"dev", "staging", "qa", "prod"},
		Default: []string{"dev", "staging", "prod"},
	}
	if err := p.AskOne(envPrompt, &envNames); err != nil {
		return nil, err
	}

	cfg.Environments = make([]config.Environment, len(envNames))
	for i, name := range envNames {
		cfg.Environments[i] = config.Environment{Name: name}
	}

	generateDocs := true
	docsPrompt := &survey.Confirm{
		Message: "Generate documentation?",
		Default: true,
	}
	if err := p.AskOne(docsPrompt, &generateDocs); err != nil {
		return nil, err
	}
	cfg.Docs.Readme = generateDocs
	cfg.Docs.Architecture = generateDocs
	cfg.Docs.Onboarding = generateDocs

	return cfg, nil
}
