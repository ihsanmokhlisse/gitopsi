# Contributing to gitopsi

Thank you for your interest in contributing to gitopsi!

## GitFlow Workflow

We use GitFlow for development:

```
main        ← Production releases (tagged)
  │
develop     ← Integration branch
  │
feature/*   ← New features
bugfix/*    ← Bug fixes
release/*   ← Release preparation
hotfix/*    ← Emergency fixes
```

### Branch Naming

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feature/issue-{id}-{description}` | `feature/issue-12-add-flux-support` |
| Bug Fix | `bugfix/issue-{id}-{description}` | `bugfix/issue-15-fix-template-error` |
| Release | `release/v{version}` | `release/v0.1.0` |
| Hotfix | `hotfix/v{version}-{description}` | `hotfix/v0.1.1-critical-fix` |

## Development Process

### 1. Create an Issue

Before starting work, create a GitHub Issue:

- Use the **Feature Request** template for new features
- Use the **Bug Report** template for bugs
- Wait for issue to be triaged and assigned

### 2. Create a Branch

```bash
# Sync with develop
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/issue-{id}-{description}
```

### 3. Make Changes

- Follow Go coding standards
- Write tests for new functionality
- Update documentation if needed
- Use conventional commits

### 4. Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

**Examples:**
```
feat(cli): add interactive prompts for init command
fix(generator): correct template path resolution
docs(readme): update installation instructions
```

### 5. Create Pull Request

- Target the `develop` branch
- Fill out the PR template completely
- Link the related issue
- Request review

### 6. Code Review

- Address all review comments
- Keep discussions focused
- Update PR as needed

### 7. Merge

Once approved:
- Squash and merge to develop
- Delete the feature branch

## Release Process

### Version Numbering

We use [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

### Creating a Release

1. Create release branch from develop:
   ```bash
   git checkout develop
   git checkout -b release/v0.1.0
   ```

2. Update version in code and CHANGELOG.md

3. Create PR to main

4. After merge, tag the release:
   ```bash
   git checkout main
   git pull
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

5. GitHub Actions will automatically:
   - Build binaries
   - Create GitHub Release
   - Generate changelog

6. Merge main back to develop:
   ```bash
   git checkout develop
   git merge main
   git push origin develop
   ```

## Development Setup

```bash
# Clone
git clone https://github.com/ihsanmokhlisse/gitopsi.git
cd gitopsi

# Install dependencies
go mod download

# Build
make build

# Test
make test

# Lint
make lint
```

## Code Style

- Use `gofmt` and `goimports`
- Follow Effective Go guidelines
- No comments in code (self-documenting)
- Error wrapping with context

## Testing

- Write table-driven tests
- Use testify for assertions
- Golden files in `testdata/`
- Aim for >80% coverage

## Questions?

Open a GitHub Issue with the `question` label.

