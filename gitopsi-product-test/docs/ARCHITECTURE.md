# gitopsi-product-test - Architecture

## Overview

This document describes the architecture of the **gitopsi-product-test** GitOps repository.

## Repository Structure

```
gitopsi-product-test/
├── infrastructure/          # Cluster-level resources
│   ├── base/               # Base configurations
│   │   ├── namespaces/     # Namespace definitions
│   │   ├── rbac/           # RBAC policies
│   │   └── kustomization.yaml
│   └── overlays/           # Environment-specific overrides
│       └── dev/
├── applications/            # Application deployments
│   ├── base/               # Base application configs
│   └── overlays/           # Environment-specific overrides
├── argocd/        # GitOps tool configuration
│   ├── projects/           # ArgoCD/Flux projects
│   └── applicationsets/    # Application definitions
├── bootstrap/              # Initial cluster setup
│   └── argocd/    # GitOps tool installation
└── scripts/                # Automation scripts
```

## Platform

- **Target Platform**: kubernetes
- **GitOps Tool**: argocd
- **Scope**: both

## Environments

| Environment | Cluster |
|-------------|---------|
| dev |  |

## Components

### Infrastructure Layer

The infrastructure layer manages cluster-level resources:

- **Namespaces**: Logical isolation for environments
- **RBAC**: Role-based access control policies
- **Network Policies**: Network segmentation rules
- **Resource Quotas**: Resource limits per namespace

### Application Layer

The application layer manages workload deployments:

- **Deployments**: Container workloads
- **Services**: Network exposure
- **ConfigMaps/Secrets**: Configuration data

### GitOps Layer

The GitOps layer manages synchronization:

- **Projects**: Logical grouping of applications
- **ApplicationSets**: Dynamic application generation
- **Sync Policies**: Automated deployment rules

## Deployment Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Commit    │ ──▶ │  argocd  │ ──▶ │   Cluster   │
│   to Git    │     │    Sync     │     │   Applied   │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Security Considerations

1. **Least Privilege**: RBAC policies follow least privilege principle
2. **Network Segmentation**: Network policies isolate workloads
3. **Secret Management**: Sensitive data stored securely
4. **Audit Logging**: All changes tracked via Git history

