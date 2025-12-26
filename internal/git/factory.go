package git

import (
	"fmt"
)

func NewProvider(providerType ProviderType, instance string) (Provider, error) {
	switch providerType {
	case ProviderGitHub:
		return NewGitHubProvider(), nil
	case ProviderGitLab:
		return NewGitLabProviderWithInstance(instance), nil
	case ProviderGitea:
		return NewGiteaProvider(instance), nil
	case ProviderAzureDevOps:
		return NewAzureDevOpsProvider("", ""), nil
	case ProviderBitbucket:
		return NewBitbucketProvider(""), nil
	case ProviderGeneric:
		return nil, fmt.Errorf("Generic provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

func NewProviderFromURL(gitURL string) (Provider, error) {
	parsed, err := ParseGitURL(gitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse git URL: %w", err)
	}

	return NewProvider(parsed.Provider, parsed.Instance)
}

func DetectProvider(gitURL string) (ProviderType, string, error) {
	parsed, err := ParseGitURL(gitURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse git URL: %w", err)
	}

	return parsed.Provider, parsed.Instance, nil
}

func GetSupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderGitHub,
		ProviderGitLab,
		ProviderGitea,
		ProviderAzureDevOps,
		ProviderBitbucket,
		ProviderGeneric,
	}
}

func GetImplementedProviders() []ProviderType {
	return []ProviderType{
		ProviderGitHub,
		ProviderGitLab,
		ProviderGitea,
		ProviderAzureDevOps,
		ProviderBitbucket,
	}
}

func IsProviderImplemented(providerType ProviderType) bool {
	implemented := GetImplementedProviders()
	for _, p := range implemented {
		if p == providerType {
			return true
		}
	}
	return false
}

func GetProviderDisplayName(providerType ProviderType) string {
	names := map[ProviderType]string{
		ProviderGitHub:      "GitHub",
		ProviderGitLab:      "GitLab",
		ProviderGitea:       "Gitea",
		ProviderAzureDevOps: "Azure DevOps",
		ProviderBitbucket:   "Bitbucket",
		ProviderGeneric:     "Generic Git",
	}

	if name, ok := names[providerType]; ok {
		return name
	}
	return string(providerType)
}

func GetAuthMethodsForProvider(providerType ProviderType) []AuthMethod {
	switch providerType {
	case ProviderGitHub:
		return []AuthMethod{AuthSSH, AuthToken, AuthOAuth}
	case ProviderGitLab:
		return []AuthMethod{AuthSSH, AuthToken, AuthOAuth}
	case ProviderGitea:
		return []AuthMethod{AuthSSH, AuthToken}
	case ProviderAzureDevOps:
		return []AuthMethod{AuthSSH, AuthToken, AuthOAuth}
	case ProviderBitbucket:
		return []AuthMethod{AuthSSH, AuthToken, AuthOAuth}
	default:
		return []AuthMethod{AuthSSH, AuthToken}
	}
}

func GetAuthMethodDisplayName(method AuthMethod) string {
	names := map[AuthMethod]string{
		AuthSSH:   "SSH Key",
		AuthToken: "Personal Access Token",
		AuthOAuth: "OAuth (Browser)",
	}

	if name, ok := names[method]; ok {
		return name
	}
	return string(method)
}
