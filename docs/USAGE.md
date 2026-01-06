# gitopsi Usage Guide

This guide covers all usage scenarios for gitopsi, from basic commands to advanced configurations.

> **v0.2.0 Note**: This release is tested with **ArgoCD + vanilla Kubernetes**. OpenShift, EKS, AKS, and Flux support exist but are not yet validated. See [README](../README.md) for the full support matrix.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Configuration Options](#configuration-options)
- [Platform Support](#platform-support)
- [Scope Options](#scope-options)
- [GitOps Tools](#gitops-tools)
- [Environment Configuration](#environment-configuration)
- [Infrastructure Components](#infrastructure-components)
- [Application Configuration](#application-configuration)
- [Output Options](#output-options)
- [Advanced Scenarios](#advanced-scenarios)

## Installation

### Option 1: Homebrew (Recommended for macOS/Linux)

```bash
brew install ihsanmokhlisse/tap/gitopsi
```

### Option 2: Container

```bash
docker pull ghcr.io/ihsanmokhlisse/gitopsi:latest
```

### Option 3: Binary Download

Visit [Releases](https://github.com/ihsanmokhlisse/gitopsi/releases) and download for your platform:

| OS | Architecture | File |
|----|--------------|------|
| Linux | amd64 | `gitopsi_VERSION_linux_amd64.tar.gz` |
| Linux | arm64 | `gitopsi_VERSION_linux_arm64.tar.gz` |
| macOS | amd64 (Intel) | `gitopsi_VERSION_darwin_amd64.tar.gz` |
| macOS | arm64 (Apple Silicon) | `gitopsi_VERSION_darwin_arm64.tar.gz` |
| Windows | amd64 | `gitopsi_VERSION_windows_amd64.zip` |

## Basic Usage

### Interactive Mode

The easiest way to get started:

```bash
gitopsi init
```

This will prompt you for:
1. **Project name**: Name of your GitOps repository
2. **Platform**: Target Kubernetes platform
3. **Scope**: What to generate (infrastructure, applications, or both)
4. **GitOps tool**: ArgoCD, Flux, or both
5. **Output type**: Local filesystem or Git repository
6. **Environments**: Which environments to configure
7. **Documentation**: Whether to generate docs

### Config File Mode

For repeatable, automated generation:

```bash
gitopsi init --config gitops.yaml
```

### Dry-Run Mode

Preview what will be generated without writing files:

```bash
gitopsi init --config gitops.yaml --dry-run
```

### Verbose Mode

See detailed output:

```bash
gitopsi init --config gitops.yaml --verbose
```

### Custom Output Directory

Generate to a specific directory:

```bash
gitopsi init --config gitops.yaml --output /path/to/output
```

## Configuration Options

### Minimal Configuration

```yaml
project:
  name: my-platform

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: prod
```

### Full Configuration Reference

```yaml
# Project metadata
project:
  name: my-platform              # Required: Project name
  description: "My GitOps Platform"  # Optional: Description

# Output configuration
output:
  type: local                    # local | git
  url: https://github.com/...    # Required if type is 'git'
  branch: main                   # Git branch (default: main)

# Platform selection
platform: kubernetes             # kubernetes | openshift | aks | eks

# Generation scope
scope: both                      # infrastructure | application | both

# GitOps tool
gitops_tool: argocd              # argocd | flux | both

# Environment definitions
environments:
  - name: dev
    cluster: https://dev.k8s.local    # Optional: Cluster URL
  - name: staging
    cluster: https://staging.k8s.local
  - name: prod
    cluster: https://prod.k8s.local

# Infrastructure components
infrastructure:
  namespaces: true               # Generate namespace manifests
  rbac: true                     # Generate RBAC (Role, RoleBinding)
  network_policies: true         # Generate NetworkPolicy
  resource_quotas: true          # Generate ResourceQuota

# Application definitions
applications:
  - name: api
    image: myregistry/api:latest
    port: 8080
    replicas: 3
  - name: web
    image: myregistry/web:latest
    port: 3000
    replicas: 2

# Documentation generation
docs:
  readme: true                   # Generate README.md
  architecture: true             # Generate docs/ARCHITECTURE.md
  onboarding: true               # Generate docs/ONBOARDING.md
```

## Platform Support

### Kubernetes (Generic)

```yaml
platform: kubernetes
```

Standard Kubernetes manifests compatible with any cluster.

### OpenShift

```yaml
platform: openshift
```

Generates OpenShift-compatible configurations including:
- SecurityContextConstraints awareness
- OpenShift-specific annotations
- Route support (future)

### Azure Kubernetes Service (AKS)

```yaml
platform: aks
```

Optimized for AKS with:
- Azure-specific annotations
- Azure AD integration patterns
- Azure Monitor compatibility

### Amazon Elastic Kubernetes Service (EKS)

```yaml
platform: eks
```

Optimized for EKS with:
- AWS-specific annotations
- IAM integration patterns
- AWS Load Balancer Controller compatibility

## Scope Options

### Infrastructure Only

Generate only cluster-level resources:

```yaml
scope: infrastructure
```

Generates:
- Namespaces
- RBAC (Roles, RoleBindings)
- Network Policies
- Resource Quotas

### Application Only

Generate only application deployments:

```yaml
scope: application
```

Generates:
- Deployments
- Services
- Kustomization files

### Both (Default)

Generate complete GitOps structure:

```yaml
scope: both
```

## GitOps Tools

### ArgoCD

```yaml
gitops_tool: argocd
```

Generates:
- `argocd/projects/` - AppProject resources
- `argocd/applicationsets/` - Application resources

### Flux

```yaml
gitops_tool: flux
```

Generates:
- `flux/` - Flux-specific configurations
- Kustomization controllers
- Source controllers

### Both

```yaml
gitops_tool: both
```

Generates configurations for both ArgoCD and Flux.

## Environment Configuration

### Basic Environments

```yaml
environments:
  - name: dev
  - name: staging
  - name: prod
```

### With Cluster URLs

```yaml
environments:
  - name: dev
    cluster: https://dev.k8s.local:6443
  - name: staging
    cluster: https://staging.k8s.local:6443
  - name: prod
    cluster: https://prod.k8s.local:6443
```

Cluster URLs are used in ArgoCD Application destinations.

## Infrastructure Components

### Namespaces

Creates environment-specific namespaces:

```yaml
infrastructure:
  namespaces: true
```

Generated: `{project}-{env}` (e.g., `my-platform-dev`)

### RBAC

Creates Role and RoleBinding for each environment:

```yaml
infrastructure:
  rbac: true
```

Includes permissions for:
- Pods, Services, ConfigMaps, Secrets (read)
- Deployments, ReplicaSets (read)

### Network Policies

Creates default network isolation:

```yaml
infrastructure:
  network_policies: true
```

Policies:
- Allow intra-namespace communication
- Allow DNS resolution
- Block external traffic by default

### Resource Quotas

Creates environment-appropriate quotas:

```yaml
infrastructure:
  resource_quotas: true
```

Default quotas by environment:

| Resource | Dev | Staging | Prod |
|----------|-----|---------|------|
| CPU Requests | 2 | 4 | 8 |
| Memory Requests | 4Gi | 8Gi | 16Gi |
| CPU Limits | 4 | 8 | 16 |
| Memory Limits | 8Gi | 16Gi | 32Gi |
| Pods | 20 | 50 | 100 |
| Services | 10 | 20 | 50 |

## Application Configuration

### Single Application

```yaml
applications:
  - name: api
    image: myregistry/api:latest
    port: 8080
    replicas: 2
```

### Multiple Applications

```yaml
applications:
  - name: frontend
    image: myregistry/frontend:latest
    port: 3000
    replicas: 2
  - name: backend
    image: myregistry/backend:latest
    port: 8080
    replicas: 3
  - name: worker
    image: myregistry/worker:latest
    port: 9000
    replicas: 1
```

### Application Fields

| Field | Description | Default |
|-------|-------------|---------|
| `name` | Application name (required) | - |
| `image` | Container image | - |
| `port` | Container port | - |
| `replicas` | Number of replicas | 1 |

## Output Options

### Local Output

Generate to local filesystem:

```yaml
output:
  type: local
```

### Git Output

Generate and configure for Git repository:

```yaml
output:
  type: git
  url: https://github.com/myorg/my-platform.git
  branch: main
```

## Advanced Scenarios

### CI/CD Pipeline Integration

```yaml
# .github/workflows/generate.yaml
name: Generate GitOps

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        default: 'dev'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Generate GitOps structure
        run: |
          docker run --rm \
            -v ${{ github.workspace }}:/workspace \
            ghcr.io/ihsanmokhlisse/gitopsi:latest \
            init --config /workspace/gitops.yaml
            
      - name: Commit and push
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add .
          git commit -m "chore: regenerate GitOps structure"
          git push
```

### Multi-Cluster Setup

```yaml
project:
  name: multi-cluster-platform

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
    cluster: https://dev-cluster.example.com
  - name: staging
    cluster: https://staging-cluster.example.com
  - name: prod-us
    cluster: https://prod-us.example.com
  - name: prod-eu
    cluster: https://prod-eu.example.com
```

### Microservices Architecture

```yaml
project:
  name: microservices-platform

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

applications:
  - name: api-gateway
    image: myregistry/api-gateway:latest
    port: 8080
    replicas: 2
  - name: user-service
    image: myregistry/user-service:latest
    port: 8081
    replicas: 3
  - name: order-service
    image: myregistry/order-service:latest
    port: 8082
    replicas: 3
  - name: payment-service
    image: myregistry/payment-service:latest
    port: 8083
    replicas: 2
  - name: notification-service
    image: myregistry/notification-service:latest
    port: 8084
    replicas: 2
```

## Troubleshooting

### Common Issues

**Error: "project name is required"**
- Ensure your config file has a `project.name` field

**Error: "invalid platform"**
- Valid platforms: `kubernetes`, `openshift`, `aks`, `eks`

**Error: "directory already exists"**
- The output directory already contains a project with that name
- Use `--output` to specify a different directory

**Error: "git URL is required when output type is 'git'"**
- When using `output.type: git`, you must provide `output.url`

### Getting Help

```bash
# General help
gitopsi --help

# Command-specific help
gitopsi init --help

# Version information
gitopsi version --verbose
```

