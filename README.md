# gitopsi

Bootstrap production-ready GitOps repositories for Kubernetes clusters.

[![CI](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yml/badge.svg)](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yml)
[![E2E](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/e2e-full.yml/badge.svg)](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/e2e-full.yml)
[![Security](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/security.yml/badge.svg)](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ihsanmokhlisse/gitopsi)](https://goreportcard.com/report/github.com/ihsanmokhlisse/gitopsi)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

gitopsi is a CLI tool that generates production-ready GitOps repository structures and optionally bootstraps ArgoCD on your cluster. It creates a complete GitOps setup with:

- Infrastructure manifests (namespaces, RBAC, network policies, resource quotas)
- Application deployments with Kustomize overlays
- ArgoCD Projects and Applications
- Documentation and onboarding guides

## Supported Features (v0.2.0)

| Feature | Status | Notes |
|---------|--------|-------|
| **GitOps Tools** | | |
| ArgoCD | Tested | Full support for Projects, Applications, App-of-Apps |
| Flux | Not Tested | Templates exist, not validated |
| **Platforms** | | |
| Kubernetes (vanilla) | Tested | Primary target platform |
| OpenShift | Not Tested | Planned for v0.3.0 |
| EKS | Not Tested | Planned for v0.3.0 |
| AKS | Not Tested | Planned for v0.3.0 |
| **Bootstrap Modes** | | |
| Helm | Tested | Installs ArgoCD via Helm |
| Manifest | Tested | Applies raw YAML manifests |
| OLM | Not Tested | For OpenShift, planned for v0.3.0 |
| Kustomize | Not Tested | Planned |

## Installation

### Homebrew (macOS/Linux)

```bash
brew install ihsanmokhlisse/tap/gitopsi
```

### Container

```bash
docker pull ghcr.io/ihsanmokhlisse/gitopsi:latest
docker run --rm -v $(pwd):/workspace ghcr.io/ihsanmokhlisse/gitopsi init
```

### Binary Download

Download from [Releases](https://github.com/ihsanmokhlisse/gitopsi/releases):

| OS | Architecture | File |
|----|--------------|------|
| Linux | amd64 | `gitopsi_VERSION_linux_amd64.tar.gz` |
| Linux | arm64 | `gitopsi_VERSION_linux_arm64.tar.gz` |
| macOS | amd64 (Intel) | `gitopsi_VERSION_darwin_amd64.tar.gz` |
| macOS | arm64 (Apple Silicon) | `gitopsi_VERSION_darwin_arm64.tar.gz` |
| Windows | amd64 | `gitopsi_VERSION_windows_amd64.zip` |

## Quick Start

### Interactive Mode

```bash
gitopsi init
```

### With Config File

```bash
gitopsi init --config gitops.yaml
```

### Generate and Bootstrap ArgoCD

```bash
gitopsi init --config gitops.yaml --git-url https://github.com/org/repo.git --push --bootstrap
```

This will:

1. Generate the GitOps repository structure
2. Push to the Git repository
3. Install ArgoCD on your current cluster
4. Configure ArgoCD to sync from the repository

## Example Configuration

```yaml
project:
  name: my-platform
  description: "Production GitOps Platform"

platform: kubernetes
scope: both
gitops_tool: argocd

git:
  url: https://github.com/myorg/my-platform.git
  branch: main

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

docs:
  readme: true
  architecture: true
  onboarding: true
```

## Generated Structure

```
my-platform/
├── infrastructure/
│   ├── base/
│   │   ├── namespaces/
│   │   ├── rbac/
│   │   ├── network-policies/
│   │   └── resource-quotas/
│   └── overlays/
│       ├── dev/
│       ├── staging/
│       └── prod/
├── applications/
│   ├── base/
│   └── overlays/
├── argocd/
│   ├── projects/
│   └── applicationsets/
├── bootstrap/
├── docs/
└── scripts/
```

## Commands

| Command | Description |
|---------|-------------|
| `gitopsi init` | Generate GitOps repository structure |
| `gitopsi validate <path>` | Validate generated manifests |
| `gitopsi preflight` | Run pre-flight cluster checks |
| `gitopsi auth` | Manage credentials |
| `gitopsi env` | Manage environments |
| `gitopsi operator` | Manage OLM operators |
| `gitopsi marketplace` | Browse and install patterns |
| `gitopsi version` | Show version information |

## Documentation

- [Usage Guide](docs/USAGE.md)
- [Examples](docs/EXAMPLES.md)
- [Environment Variables](docs/ENVIRONMENT_VARIABLES.md)
- [Contributing](docs/CONTRIBUTING.md)
- [Testing](docs/TESTING.md)

## Roadmap

### v0.2.0 (Current)

- ArgoCD support with App-of-Apps pattern
- Vanilla Kubernetes bootstrap
- Helm and Manifest bootstrap modes
- Infrastructure generation (namespaces, RBAC, network policies, quotas)
- Pattern marketplace

### v0.3.0 (Planned)

- OpenShift support with OLM bootstrap
- EKS and AKS platform optimizations
- Flux support validation
- Multi-cluster ApplicationSets

## License

MIT License - see [LICENSE](LICENSE) for details.
