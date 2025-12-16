package prompt

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/ihsanmokhlisse/gitopsi/internal/config"
)

func Run() (*config.Config, error) {
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

	if err := survey.Ask(questions, &answers); err != nil {
		return nil, err
	}

	cfg.Project.Name = answers.ProjectName
	cfg.Platform = answers.Platform
	cfg.Scope = answers.Scope
	cfg.GitOpsTool = answers.GitOpsTool
	cfg.Output.Type = answers.OutputType

	if answers.OutputType == "git" {
		gitURL := ""
		prompt := &survey.Input{
			Message: "Git repository URL:",
		}
		if err := survey.AskOne(prompt, &gitURL, survey.WithValidator(survey.Required)); err != nil {
			return nil, err
		}
		cfg.Output.URL = gitURL
	}

	envNames := []string{}
	envPrompt := &survey.MultiSelect{
		Message: "Environments:",
		Options: []string{"dev", "staging", "qa", "prod"},
		Default: []string{"dev", "staging", "prod"},
	}
	if err := survey.AskOne(envPrompt, &envNames); err != nil {
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
	if err := survey.AskOne(docsPrompt, &generateDocs); err != nil {
		return nil, err
	}
	cfg.Docs.Readme = generateDocs
	cfg.Docs.Architecture = generateDocs
	cfg.Docs.Onboarding = generateDocs

	return cfg, nil
}

