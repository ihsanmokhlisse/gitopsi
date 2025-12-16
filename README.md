# gitopsi

[![CI](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml/badge.svg)](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ihsanmokhlisse/gitopsi)](https://goreportcard.com/report/github.com/ihsanmokhlisse/gitopsi)

**Bootstrap production-ready GitOps repositories in seconds.**

gitopsi generates complete GitOps repository structures with Kubernetes manifests, Kustomize overlays, ArgoCD/Flux configurations, and documentation.

## Features

- **Multi-Platform**: Kubernetes, OpenShift, AKS, EKS
- **Multi-Tool**: ArgoCD, Flux, or both
- **Flexible Scope**: Infrastructure, Applications, or both
- **Interactive & Config**: CLI prompts or YAML config file
- **Production Ready**: Includes docs, scripts, and CI/CD

## Installation

### Container (Recommended)

```bash
podman pull ghcr.io/ihsanmokhlisse/gitopsi:latest
podman run --rm -v $(pwd):/workspace gitopsi init
```

### Binary

Download from [Releases](https://github.com/ihsanmokhlisse/gitopsi/releases)

## Quick Start

### Interactive Mode

```bash
gitopsi init
```

### Config File Mode

```bash
gitopsi init --config gitops.yaml
```

### Example Config

```yaml
project:
  name: my-platform

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

applications:
  - name: api
    image: myregistry/api:latest
    port: 8080
```

## Generated Structure

```
my-platform/
├── README.md
├── bootstrap/argocd/
├── infrastructure/
│   ├── base/
│   └── overlays/{dev,staging,prod}/
├── applications/
│   ├── base/
│   └── overlays/{dev,staging,prod}/
├── argocd/
│   ├── projects/
│   └── applicationsets/
└── scripts/
```

## Development

**⚠️ Testing Rule**: All testing must use Podman Desktop containers.

```bash
# Build dev container
make container-build

# Run tests in container
make container-test

# Interactive shell
make container-shell
```

## License

MIT
