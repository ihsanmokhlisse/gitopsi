# gitopsi

[![CI](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml/badge.svg)](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ihsanmokhlisse/gitopsi)](https://goreportcard.com/report/github.com/ihsanmokhlisse/gitopsi)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A CLI tool for bootstrapping production-ready GitOps repositories.

## Features

- **Multi-Platform**: Kubernetes, OpenShift, AKS, EKS
- **Multi-Tool**: ArgoCD, Flux, or both
- **Full Scope**: Infrastructure + Applications
- **Dual Mode**: Interactive prompts or config file
- **Production Ready**: Docs, scripts, CI/CD included

## Installation

```bash
# From source
go install github.com/ihsanmokhlisse/gitopsi/cmd/gitopsi@latest

# Or build locally
git clone https://github.com/ihsanmokhlisse/gitopsi.git
cd gitopsi
make build
./bin/gitopsi --help
```

## Quick Start

### Interactive Mode

```bash
gitopsi init
```

### Config File Mode

```bash
gitopsi init --config gitops.yaml
```

## Usage

```bash
gitopsi [command] [flags]

Commands:
  init        Initialize new GitOps repository
  version     Show version

Flags:
  --config    Config file path
  --output    Output directory
  --dry-run   Preview without writing
  --verbose   Verbose output
```

## Configuration Example

```yaml
project:
  name: my-platform

platform: kubernetes    # kubernetes | openshift | aks | eks
scope: both            # infrastructure | application | both
gitops_tool: argocd    # argocd | flux | both

environments:
  - name: dev
  - name: staging
  - name: prod

applications:
  - name: frontend
    image: myregistry/frontend:latest
    port: 8080
```

## Generated Structure

```
my-platform/
├── README.md
├── Makefile
├── docs/
├── bootstrap/argocd/
├── infrastructure/
│   ├── base/
│   └── overlays/{env}/
├── applications/
│   ├── base/
│   └── overlays/{env}/
├── argocd/
│   └── applicationsets/
└── scripts/
```

## Documentation

- [Contributing Guide](docs/CONTRIBUTING.md)
- [Changelog](docs/CHANGELOG.md)

## Development

```bash
# Build
make build

# Test
make test

# Lint
make lint
```

## Contributing

Please read [CONTRIBUTING.md](docs/CONTRIBUTING.md) for details on our GitFlow workflow and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
