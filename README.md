# gitopsi

[![CI](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml/badge.svg)](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ihsanmokhlisse/gitopsi)](https://goreportcard.com/report/github.com/ihsanmokhlisse/gitopsi)
[![Security](https://github.com/ihsanmokhlisse/gitopsi/actions/workflows/ci.yaml/badge.svg?event=push)](https://github.com/ihsanmokhlisse/gitopsi/security)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Bootstrap production-ready GitOps repositories in seconds.**

gitopsi generates complete GitOps repository structures with Kubernetes manifests, Kustomize overlays, ArgoCD/Flux configurations, RBAC, Network Policies, Resource Quotas, and comprehensive documentation.

## 🚀 Features

| Feature | Description |
|---------|-------------|
| **Multi-Platform** | Kubernetes, OpenShift, AKS, EKS |
| **Multi-Tool** | ArgoCD, Flux, or both |
| **Flexible Scope** | Infrastructure, Applications, or both |
| **Interactive & Config** | CLI prompts or YAML config file |
| **Production Ready** | RBAC, NetworkPolicies, ResourceQuotas |
| **Documentation** | README, Architecture, Onboarding guides |
| **Dry-Run Mode** | Preview without writing files |

## 📦 Installation

### Homebrew (macOS/Linux)

```bash
brew install ihsanmokhlisse/tap/gitopsi
```

### Container (Recommended for CI/CD)

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/ihsanmokhlisse/gitopsi:latest

# Run with current directory mounted
docker run --rm -v $(pwd):/workspace ghcr.io/ihsanmokhlisse/gitopsi:latest init
```

### Binary Download

Download from [Releases](https://github.com/ihsanmokhlisse/gitopsi/releases):

```bash
# Linux amd64
curl -LO https://github.com/ihsanmokhlisse/gitopsi/releases/latest/download/gitopsi_linux_amd64.tar.gz
tar -xzf gitopsi_linux_amd64.tar.gz
sudo mv gitopsi /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -LO https://github.com/ihsanmokhlisse/gitopsi/releases/latest/download/gitopsi_darwin_arm64.tar.gz
tar -xzf gitopsi_darwin_arm64.tar.gz
sudo mv gitopsi /usr/local/bin/
```

### Verify Installation

```bash
gitopsi version
```

## 🎯 Quick Start

### Interactive Mode

Simply run `gitopsi init` and follow the prompts:

```bash
gitopsi init
```

You'll be asked to configure:
1. Project name
2. Target platform (kubernetes, openshift, aks, eks)
3. Scope (infrastructure, application, both)
4. GitOps tool (argocd, flux, both)
5. Output type (local, git)
6. Environments (dev, staging, qa, prod)
7. Documentation generation

### Config File Mode

Create a configuration file and run:

```bash
gitopsi init --config gitops.yaml
```

### Dry-Run Mode

Preview what will be generated without writing files:

```bash
gitopsi init --config gitops.yaml --dry-run
```

## 📝 Configuration

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

### Full Configuration

```yaml
project:
  name: my-platform
  description: "Production GitOps Platform"

output:
  type: git
  url: https://github.com/org/my-platform.git
  branch: main

platform: kubernetes      # kubernetes | openshift | aks | eks
scope: both               # infrastructure | application | both
gitops_tool: argocd       # argocd | flux | both

environments:
  - name: dev
    cluster: https://dev.k8s.local
  - name: staging
    cluster: https://staging.k8s.local
  - name: prod
    cluster: https://prod.k8s.local

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: api
    image: myregistry/api:latest
    port: 8080
    replicas: 3
  - name: web
    image: myregistry/web:latest
    port: 3000
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

## 📂 Generated Structure

```
my-platform/
├── README.md                           # Project overview
├── docs/
│   ├── ARCHITECTURE.md                 # System architecture
│   └── ONBOARDING.md                   # Developer onboarding
├── bootstrap/
│   └── argocd/
│       └── namespace.yaml              # GitOps tool namespace
├── infrastructure/
│   ├── base/
│   │   ├── kustomization.yaml
│   │   ├── namespaces/
│   │   │   ├── dev.yaml
│   │   │   ├── staging.yaml
│   │   │   └── prod.yaml
│   │   ├── rbac/                       # Role-Based Access Control
│   │   │   ├── dev.yaml
│   │   │   ├── staging.yaml
│   │   │   └── prod.yaml
│   │   ├── network-policies/           # Network isolation
│   │   │   ├── dev.yaml
│   │   │   ├── staging.yaml
│   │   │   └── prod.yaml
│   │   └── resource-quotas/            # Resource limits
│   │       ├── dev.yaml
│   │       ├── staging.yaml
│   │       └── prod.yaml
│   └── overlays/
│       ├── dev/kustomization.yaml
│       ├── staging/kustomization.yaml
│       └── prod/kustomization.yaml
├── applications/
│   ├── base/
│   │   ├── kustomization.yaml
│   │   └── {app-name}/
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       └── kustomization.yaml
│   └── overlays/
│       ├── dev/kustomization.yaml
│       ├── staging/kustomization.yaml
│       └── prod/kustomization.yaml
├── argocd/
│   ├── projects/
│   │   ├── infrastructure.yaml         # ArgoCD AppProject
│   │   └── applications.yaml
│   └── applicationsets/
│       ├── infra-dev.yaml              # ArgoCD Application
│       ├── infra-staging.yaml
│       ├── infra-prod.yaml
│       ├── apps-dev.yaml
│       ├── apps-staging.yaml
│       └── apps-prod.yaml
└── scripts/
    ├── bootstrap.sh                    # Initial setup script
    └── validate.sh                     # Validation script
```

## 🎮 Usage Examples

### Example 1: Kubernetes with ArgoCD

```yaml
# kubernetes-argocd.yaml
project:
  name: k8s-platform

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: backend
    image: nginx:latest
    port: 80
    replicas: 2
```

```bash
gitopsi init --config kubernetes-argocd.yaml
```

### Example 2: OpenShift Infrastructure Only

```yaml
# openshift-infra.yaml
project:
  name: ocp-infrastructure

platform: openshift
scope: infrastructure
gitops_tool: argocd

environments:
  - name: dev
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true
```

```bash
gitopsi init --config openshift-infra.yaml
```

### Example 3: AKS with Flux

```yaml
# aks-flux.yaml
project:
  name: aks-apps

platform: aks
scope: application
gitops_tool: flux

environments:
  - name: dev
    cluster: https://aks-dev.westus.azmk8s.io
  - name: prod
    cluster: https://aks-prod.westus.azmk8s.io

applications:
  - name: api
    image: myacr.azurecr.io/api:latest
    port: 8080
    replicas: 3
```

```bash
gitopsi init --config aks-flux.yaml
```

### Example 4: EKS Multi-App Setup

```yaml
# eks-multi-app.yaml
project:
  name: eks-platform
  description: "EKS Production Platform"

platform: eks
scope: both
gitops_tool: argocd

output:
  type: git
  url: https://github.com/myorg/eks-platform.git

environments:
  - name: dev
  - name: staging
  - name: prod

applications:
  - name: frontend
    image: 123456789.dkr.ecr.us-east-1.amazonaws.com/frontend:latest
    port: 3000
    replicas: 2
  - name: backend
    image: 123456789.dkr.ecr.us-east-1.amazonaws.com/backend:latest
    port: 8080
    replicas: 3
  - name: worker
    image: 123456789.dkr.ecr.us-east-1.amazonaws.com/worker:latest
    port: 9000
    replicas: 1
```

```bash
gitopsi init --config eks-multi-app.yaml
```

## 🔧 CLI Reference

```
gitopsi - GitOps Repository Generator

Usage:
  gitopsi [command]

Available Commands:
  init        Initialize a new GitOps repository
  version     Print version information
  help        Help about any command

Flags:
  --config string    Config file (default: gitops.yaml)
  --output string    Output directory (default: current directory)
  --dry-run          Preview without writing files
  --verbose          Verbose output
  -h, --help         Help for gitopsi

Examples:
  gitopsi init                           # Interactive mode
  gitopsi init --config gitops.yaml      # Config file mode
  gitopsi init --dry-run                 # Preview mode
  gitopsi init --output /path/to/output  # Custom output directory
```

## 🐳 Container Usage

### Basic Usage

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/ihsanmokhlisse/gitopsi:latest init
```

### With Config File

```bash
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/gitops.yaml:/workspace/gitops.yaml:ro \
  ghcr.io/ihsanmokhlisse/gitopsi:latest \
  init --config /workspace/gitops.yaml
```

### CI/CD Integration

```yaml
# GitHub Actions
- name: Generate GitOps Repository
  run: |
    docker run --rm \
      -v ${{ github.workspace }}:/workspace \
      ghcr.io/ihsanmokhlisse/gitopsi:latest \
      init --config /workspace/gitops.yaml
```

## 🔒 Security

gitopsi follows security best practices:

- **Container**: Uses distroless base image, runs as non-root
- **Binary Signing**: All releases are signed with cosign
- **SBOM**: Software Bill of Materials included with releases
- **Vulnerability Scanning**: Trivy, Grype, govulncheck, gosec
- **Code Analysis**: CodeQL security scanning

### Verify Container Signature

```bash
cosign verify ghcr.io/ihsanmokhlisse/gitopsi:latest
```

## 🛠️ Development

**⚠️ Testing Rule**: All testing must use Podman Desktop containers.

```bash
# Clone repository
git clone https://github.com/ihsanmokhlisse/gitopsi.git
cd gitopsi

# Build dev container
make container-build

# Run tests in container
make container-test

# Run linting
make container-lint

# Interactive shell
make container-shell

# Build all platforms
make build-all
```

## 📖 Documentation

- [Usage Guide](docs/USAGE.md) - Detailed usage instructions
- [Examples](docs/EXAMPLES.md) - Complete configuration examples
- [Contributing](docs/CONTRIBUTING.md) - How to contribute
- [Testing](docs/TESTING.md) - Testing strategy
- [Changelog](docs/CHANGELOG.md) - Version history

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/CONTRIBUTING.md) first.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [ArgoCD](https://argoproj.github.io/argo-cd/) - GitOps continuous delivery
- [Flux](https://fluxcd.io/) - GitOps toolkit
- [Kustomize](https://kustomize.io/) - Kubernetes native configuration
- [Cobra](https://github.com/spf13/cobra) - CLI framework
