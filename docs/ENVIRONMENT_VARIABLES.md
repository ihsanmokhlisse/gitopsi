# Environment Variables

gitopsi follows [12-Factor App](https://12factor.net/) methodology. All configuration can be provided via environment variables.

## Configuration Priority

gitopsi reads configuration from these sources (highest to lowest priority):

1. **CLI Flags** - `--git-url`, `--cluster`, etc.
2. **Environment Variables** - `GITOPSI_*`
3. **Config File** - `gitops.yaml`
4. **Auto-Detection** - kubeconfig, git remote

## Supported Environment Variables

### Git Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `GITOPSI_GIT_URL` | Git repository URL | `https://github.com/org/repo.git` |
| `GITOPSI_GIT_TOKEN` | Git authentication token | `ghp_xxxx` |
| `GITOPSI_GIT_BRANCH` | Default branch | `main` |

### Cluster Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `GITOPSI_CLUSTER_URL` | Kubernetes API server URL | `https://api.cluster.com:6443` |
| `GITOPSI_CLUSTER_TOKEN` | Kubernetes authentication token | `eyJhbG...` |
| `GITOPSI_CLUSTER_PLATFORM` | Platform type | `kubernetes`, `openshift`, `eks`, `aks` |

### Git Provider Tokens

| Variable | Provider | Scope Required |
|----------|----------|----------------|
| `GITHUB_TOKEN` | GitHub | `repo` |
| `GITLAB_TOKEN` | GitLab | `api`, `write_repository` |
| `BITBUCKET_TOKEN` | Bitbucket | Repository write |
| `GITEA_TOKEN` | Gitea | Repository write |
| `AZURE_DEVOPS_TOKEN` | Azure DevOps | Code (Read & Write) |

### General Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GITOPSI_CONFIG` | Path to config file | `gitops.yaml` |
| `GITOPSI_OUTPUT` | Output directory | `.` |
| `GITOPSI_VERBOSE` | Enable verbose output | `false` |
| `GITOPSI_DRY_RUN` | Preview without writing | `false` |

## Usage Examples

### Basic Setup

```bash
export GITOPSI_GIT_URL="https://github.com/myorg/gitops-repo.git"
export GITOPSI_GIT_TOKEN="ghp_your_github_token"
export GITOPSI_CLUSTER_URL="https://api.mycluster.com:6443"

gitopsi init --bootstrap --push
```

### CI/CD Pipeline

```yaml
# GitHub Actions
env:
  GITOPSI_GIT_TOKEN: ${{ secrets.GIT_TOKEN }}
  GITOPSI_CLUSTER_TOKEN: ${{ secrets.CLUSTER_TOKEN }}

steps:
  - run: gitopsi init --config gitops.yaml --bootstrap --push
```

```yaml
# GitLab CI
variables:
  GITOPSI_GIT_TOKEN: $GITLAB_TOKEN
  GITOPSI_CLUSTER_TOKEN: $KUBE_TOKEN

script:
  - gitopsi init --config gitops.yaml --bootstrap --push
```

### Multiple Clusters

```bash
# Development
export GITOPSI_CLUSTER_URL="https://dev-cluster:6443"
export GITOPSI_CLUSTER_TOKEN="dev-token"
gitopsi init --config dev.yaml

# Production
export GITOPSI_CLUSTER_URL="https://prod-cluster:6443"
export GITOPSI_CLUSTER_TOKEN="prod-token"
gitopsi init --config prod.yaml
```

## Security Best Practices

1. **Never commit tokens** - Use environment variables or secrets management
2. **Use minimal scopes** - Only request permissions you need
3. **Rotate tokens regularly** - Set up automated rotation
4. **Use secret managers** - Vault, AWS Secrets Manager, Azure Key Vault

```bash
# Using Vault
export GITOPSI_GIT_TOKEN=$(vault kv get -field=token secret/gitopsi)

# Using AWS Secrets Manager
export GITOPSI_GIT_TOKEN=$(aws secretsmanager get-secret-value --secret-id gitopsi-token --query SecretString --output text)
```

## 12-Factor Compliance

gitopsi ensures:

- **No hardcoded values** - All URLs, tokens, and configuration come from user input
- **Environment parity** - Same config structure works across dev/staging/prod
- **Explicit dependencies** - All required config is validated upfront
- **Fail fast** - Clear errors when required config is missing

```bash
# This will fail with a clear error if GITOPSI_GIT_URL is not set
gitopsi init --bootstrap --push
# Error: git.url is required - gitopsi generates GitOps manifests that sync from Git
```




