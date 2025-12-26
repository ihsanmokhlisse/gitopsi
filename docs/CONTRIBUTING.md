# Contributing to gitopsi

Thank you for your interest in contributing to gitopsi! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.23+
- Git
- Python 3 (for pre-commit)
- Podman or Docker (for container tests)

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/ihsanmokhlisse/gitopsi.git
cd gitopsi

# Run developer setup (installs tools and hooks)
make setup
```

This will install:

- Go tools (golangci-lint, goimports, gosec, govulncheck)
- Pre-commit hooks
- Git hooks

### Manual Setup

If you prefer manual setup:

```bash
# Install Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install pre-commit
pip3 install pre-commit

# Install hooks
pre-commit install
pre-commit install --hook-type commit-msg
```

## Pre-commit Hooks

We use [pre-commit](https://pre-commit.com/) to ensure code quality before commits.

### Installed Hooks

| Hook | Description |
|------|-------------|
| `go-fmt` | Format Go code with gofmt |
| `goimports` | Sort imports and format |
| `go-vet` | Static analysis |
| `go-mod-tidy` | Keep go.mod clean |
| `golangci-lint` | Comprehensive linting |
| `gosec` | Security scanning |
| `go-test-short` | Quick tests |
| `go-build` | Verify compilation |
| `conventional-pre-commit` | Commit message format |
| `yamllint` | YAML linting |
| `markdownlint` | Markdown linting |
| `shellcheck` | Shell script linting |
| `hadolint` | Dockerfile linting |
| `gitleaks` | Secret detection |

### Running Hooks

```bash
# Run on staged files (automatic on commit)
pre-commit run

# Run on all files
make pre-commit-all
# or
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint --all-files
pre-commit run go-test-short --all-files

# Update hooks to latest versions
make pre-commit-update
```

### Skipping Hooks

In rare cases, you may need to skip hooks:

```bash
# Skip all hooks
git commit --no-verify -m "message"

# Skip specific hook
SKIP=golangci-lint git commit -m "message"

# Skip multiple hooks
SKIP=golangci-lint,go-test-short git commit -m "message"
```

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semi-colons, etc.
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding missing tests
- `build`: Changes to build system or dependencies
- `ci`: CI configuration changes
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

### Examples

```bash
feat(cli): add validate command
fix(generator): handle empty environment list
docs(readme): add installation instructions
test(bootstrap): add ArgoCD detection tests
ci: add integration test job
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout develop
git pull origin develop
git checkout -b feature/issue-XX-description
```

### 2. Make Changes

Write your code following our style guide:

- No comments in code (self-documenting)
- Use descriptive names
- Keep functions small and focused
- Write tests for new functionality

### 3. Run Local Checks

```bash
# Quick checks
make ci-quick

# Full CI simulation
make ci-local

# Or use pre-commit
pre-commit run --all-files
```

### 4. Run Tests

```bash
# All tests
make test-all

# Specific test types
make test-unit
make test-integration
make test-regression
```

### 5. Commit Changes

```bash
git add .
git commit -m "feat(scope): description"
```

Pre-commit hooks will run automatically. If they fail:

1. Fix the issues
2. Stage the fixes: `git add .`
3. Try committing again

### 6. Push and Create PR

```bash
git push origin feature/issue-XX-description
```

Then create a Pull Request on GitHub.

## Testing

### Test Types

| Type | Command | Description |
|------|---------|-------------|
| Unit | `make test-unit` | Test individual functions |
| Integration | `make test-integration` | Test component interactions |
| Regression | `make test-regression` | Verify bug fixes |
| E2E | `make test-e2e` | End-to-end scenarios |

### Writing Tests

```go
func TestFeatureName_Scenario(t *testing.T) {
    // Arrange
    input := ...

    // Act
    result, err := FunctionUnderTest(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Coverage

```bash
# Generate coverage report
make test-coverage

# HTML report
make test-coverage-html

# Check threshold
make coverage-check
```

## Code Style

### 12-Factor App Compliance (MANDATORY)

gitopsi follows [12-Factor App](https://12factor.net/) methodology. **Factor III (Config)** is critical:

#### ❌ NEVER Hardcode Values

```go
// BAD - Hardcoded URL
repoURL := "https://github.com/org/repo.git"

// BAD - Hardcoded fallback
if repoURL == "" {
    repoURL = "https://github.com/org/" + projectName + ".git"
}
```

#### ✅ ALWAYS Use Configuration

```go
// GOOD - From config (user-provided)
repoURL := cfg.Git.URL

// GOOD - Fail if required config is missing
if repoURL == "" {
    return fmt.Errorf("git.url is required - ArgoCD needs a Git repository to sync from")
}
```

#### Configuration Sources (Priority Order)

1. **CLI Flags**: `--git-url`, `--cluster`, etc.
2. **Environment Variables**: `GITOPSI_GIT_URL`, `GITOPSI_CLUSTER_URL`, etc.
3. **Config File**: `gitops.yaml`
4. **Auto-Detection**: kubeconfig, git remote (only for cluster info, never for Git URL)

#### Rules for Templates

```yaml
# BAD - Hardcoded in template
repoURL: https://github.com/org/{{ .ProjectName }}.git

# GOOD - Variable from config
repoURL: {{ .RepoURL }}
```

The `.RepoURL` value MUST come from user configuration, never constructed internally.

### Formatting

- Use `gofmt` (automatic via pre-commit)
- Use `goimports` for import sorting
- Maximum line length: 120 characters

### Naming

- Use camelCase for unexported identifiers
- Use PascalCase for exported identifiers
- Use descriptive names (no single letters except for loops)

### Error Handling

```go
// Return clear error when required config is missing
if cfg.Git.URL == "" {
    return fmt.Errorf("git.url is required: gitopsi generates GitOps manifests that sync from Git")
}

// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### Structure

```
internal/
├── cli/          # CLI commands
├── config/       # Configuration
├── generator/    # Code generation
├── bootstrap/    # GitOps bootstrapping
└── ...
```

## Pull Request Guidelines

### PR Checklist

- [ ] Code follows style guidelines
- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] Pre-commit hooks pass
- [ ] Commit messages follow conventional commits
- [ ] Documentation updated if needed
- [ ] PR description explains changes

### PR Title

Follow the same format as commit messages:

```
feat(cli): add validate command for manifest validation
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Related Issue
Fixes #XX

## Changes
- Change 1
- Change 2

## Testing
How was this tested?

## Checklist
- [ ] Tests added
- [ ] Documentation updated
- [ ] Pre-commit hooks pass
```

## Getting Help

- Open an issue for bugs or feature requests
- Use discussions for questions
- Review existing issues before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
