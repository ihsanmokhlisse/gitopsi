# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- N/A

### Changed
- N/A

### Fixed
- N/A

## [0.2.0] - 2026-01-06

### Tested & Validated Features

This release focuses on **ArgoCD + vanilla Kubernetes** as the primary tested configuration.

| Feature | Status |
|---------|--------|
| ArgoCD generation | Tested |
| Kubernetes (vanilla) platform | Tested |
| Bootstrap (Helm mode) | Tested |
| Bootstrap (Manifest mode) | Tested |
| OpenShift, EKS, AKS | Not tested (planned for v0.3.0) |
| Flux | Not tested (planned for v0.3.0) |
| OLM bootstrap mode | Not tested (planned for v0.3.0) |

### Added

#### Bootstrap System
- Full ArgoCD bootstrap with `--bootstrap` flag
- Helm bootstrap mode for vanilla Kubernetes
- Manifest bootstrap mode for raw YAML installation
- Auto-detection of cluster platform
- Cluster URL auto-detection from kubeconfig
- Pre-creation of ArgoCD Projects before App-of-Apps (fixes sync issues)
- App-of-Apps pattern for managing infrastructure and applications

#### CLI Enhancements
- `gitopsi preflight` command for cluster readiness checks
- `gitopsi validate` command for manifest validation (schema, security, deprecation)
- `gitopsi marketplace` command for pattern discovery and installation
- `gitopsi auth` command for credential management
- `gitopsi env` command for environment management
- `gitopsi operator` command for OLM operator management

#### E2E Testing
- Comprehensive E2E test suite with fixtures
- Preset validation tests (minimal, standard, enterprise)
- CI pipeline with E2E and regression tests

#### Pattern Marketplace
- Remote pattern registry support
- Pattern installation into projects
- Built-in patterns for common use cases

### Fixed
- Bootstrap now runs without explicit `--cluster` flag (auto-detection)
- Child ArgoCD apps now sync correctly (projects created before App-of-Apps)
- Git push works correctly with new repositories (branch name initialization)
- Validate command no longer conflicts with marketplace validate
- E2E test fixtures include required git.url field

### Changed
- Updated README with accurate feature support matrix
- CI workflow now includes E2E tests and preset validation
- Improved error messages and progress reporting

## [0.1.0] - 2024-12-16

### 🚀 Initial Release - MVP Phase 1

First public release of gitopsi, a CLI tool for bootstrapping production-ready GitOps repositories.

### Added

#### CLI & Configuration
- `gitopsi init` command with interactive mode using Survey library
- `gitopsi init --config <file>` for YAML configuration file mode
- `gitopsi init --dry-run` for preview without writing files
- `gitopsi version` command with build information
- Multi-platform support: Kubernetes, OpenShift, AKS, EKS
- Multi-scope support: infrastructure, application, or both
- GitOps tool support: ArgoCD, Flux, or both

#### Generated Structure
- Kubernetes manifests (Deployments, Services)
- Kustomize base/overlay structure per environment
- ArgoCD resources (AppProject, Application per environment)
- Infrastructure components:
  - Namespace manifests
  - RBAC (Role, RoleBinding)
  - NetworkPolicies (default deny with allow intra-namespace)
  - ResourceQuotas (environment-appropriate limits)
- Bootstrap scripts (`bootstrap.sh`, `validate.sh`)
- Documentation generation (README, Architecture, Onboarding)

#### Templates
- 12 embedded templates using Go 1.16+ embed.FS
- Kubernetes: deployment, service, kustomization
- Infrastructure: namespace, rbac, networkpolicy, resourcequota
- ArgoCD: application, project
- Documentation: README, architecture, onboarding

#### CI/CD Pipeline
- GitHub Actions workflows for CI and Release
- Multi-platform testing: Linux, macOS, Windows
- Security scanning: CodeQL, gosec, govulncheck, Trivy, Grype, nancy
- Multi-platform builds: linux/darwin/windows × amd64/arm64
- E2E tests with full generation validation
- Container build with distroless base, non-root user
- SBOM generation: SPDX + CycloneDX formats
- Binary signing with cosign
- Container image signing with cosign

#### Documentation
- Comprehensive README with installation and usage
- Detailed USAGE.md with all configuration options
- EXAMPLES.md with 15+ real-world configuration examples
- CONTRIBUTING.md with GitFlow workflow
- TESTING.md with testing strategy

### Test Coverage
| Package | Coverage |
|---------|----------|
| internal/cli | 81.2% |
| internal/config | 97.1% |
| internal/generator | 78.0% |
| internal/output | 87.0% |
| internal/prompt | 91.2% |
| internal/templates | 83.3% |

### Platforms
- **Container**: `ghcr.io/ihsanmokhlisse/gitopsi:0.1.0`
- **Homebrew**: `brew install ihsanmokhlisse/tap/gitopsi`
- **Binaries**: Linux, macOS, Windows (amd64, arm64)
- **Packages**: deb, rpm, apk

### Closed Issues
- #1 CLI structure with Cobra
- #2 Config file parsing with Viper
- #3 Interactive prompts with Survey
- #4 Init command implementation
- #5 Kubernetes manifest generation
- #6 Kustomize base/overlay structure
- #7 ArgoCD Application generation
- #8 Template embedding with embed.FS
- #9 File output system
- #10 README.md generation
- #11 Unit tests (>80% coverage)

---

[Unreleased]: https://github.com/ihsanmokhlisse/gitopsi/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ihsanmokhlisse/gitopsi/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ihsanmokhlisse/gitopsi/releases/tag/v0.1.0
