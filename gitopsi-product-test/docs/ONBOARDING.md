# gitopsi-product-test - Onboarding Guide

## Prerequisites

Before you begin, ensure you have:

- [ ] Access to the Git repository
- [ ] `kubectl` configured for target cluster(s)
- [ ] argocd CLI installed
- [ ] Kustomize installed

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd gitopsi-product-test
```

### 2. Bootstrap the Cluster

```bash
# Run the bootstrap script
./scripts/bootstrap.sh
```

This will:
- Create the argocd namespace
- Install argocd components
- Configure initial synchronization

### 3. Verify Installation

```bash
# Check argocd status
kubectl get pods -n argocd

# Verify applications are syncing
argocd app list
```

## Environment Access

| Environment | Cluster URL | Access Level |
|-------------|-------------|--------------|
| dev |  | Developer |

## Common Tasks

### Adding a New Application

1. Create application manifests in `applications/base/<app-name>/`
2. Add to `applications/base/kustomization.yaml`
3. Create overlay if environment-specific config needed
4. Commit and push changes

### Modifying Infrastructure

1. Update manifests in `infrastructure/base/`
2. Test with `kustomize build infrastructure/base/`
3. Commit and push changes

### Promoting Changes

```bash
# Validate changes
./scripts/validate.sh

# Commit with conventional commit message
git commit -m "feat(<scope>): description"

# Push to trigger sync
git push origin main
```

## Troubleshooting

### Application Not Syncing

```bash
# Check application status
argocd app get <app-name>

# Force sync
argocd app sync <app-name>
```

### Viewing Logs

```bash
# ArgoCD logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-server

# Application logs
kubectl logs -n <namespace> -l app=<app-name>
```

## Support

- **Documentation**: See `docs/` folder
- **Issues**: Create a GitHub issue
- **Slack**: #gitopsi-product-test-support

