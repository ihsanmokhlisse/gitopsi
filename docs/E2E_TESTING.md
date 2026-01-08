# E2E Testing Documentation

## Overview

The gitopsi E2E test suite validates the **complete gitopsi workflow** - from manifest
generation through ArgoCD sync and cluster-side validation. Every operation is performed
through gitopsi, and ArgoCD syncs changes from Git.

**Key Principle**: gitopsi manages manifests, ArgoCD syncs from Git. No direct
kubectl/oc apply.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                         THE GITOPS WORKFLOW                                 │
│                                                                             │
│   gitopsi ────► Git Repository ◄──── ArgoCD ────► Kubernetes Cluster       │
│   (manages)     (source of truth)    (syncs)      (deploys)                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Complete 12-Step Test Flow

Each preset/configuration runs through a **complete 12-step flow**:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│              COMPLETE TEST FLOW (per preset/configuration)                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Step 1: Generate Manifests                                                 │
│  └── gitopsi init --config <preset-config>.yaml --output /tmp/output        │
│                                                                             │
│  Step 2: Create Kind Cluster                                                │
│  └── kind create cluster --name e2e-<preset>                                │
│                                                                             │
│  Step 3: Push to Git                                                        │
│  └── git push to test branch                                                │
│                                                                             │
│  Step 4: Bootstrap ArgoCD                                                   │
│  └── gitopsi init --bootstrap --bootstrap-mode helm                         │
│                                                                             │
│  Step 5: Configure ArgoCD to sync from Git                                  │
│  └── Create ArgoCD Application pointing to test branch                      │
│                                                                             │
│  Step 6: Verify Initial Sync                                                │
│  └── Check ArgoCD status shows "Synced"                                     │
│                                                                             │
│  Step 7: Validate Initial Deployment on Cluster                             │
│  └── kubectl get namespace dev (verify namespace exists)                    │
│  └── kubectl get resourcequota -n dev (verify quota applied)                │
│  └── kubectl get networkpolicy -n dev (verify netpol exists)                │
│                                                                             │
│  Step 8: Make Code Changes                                                  │
│  └── Modify manifest (e.g., add label "e2e-test-change: true")              │
│                                                                             │
│  Step 9: Push Changes to Git                                                │
│  └── git commit && git push                                                 │
│                                                                             │
│  Step 10: Verify ArgoCD Detects and Syncs Changes                           │
│  └── Wait for ArgoCD to show "Synced" status                                │
│                                                                             │
│  Step 11: Validate Changes Applied on Cluster                               │
│  └── kubectl get namespace dev -o yaml | grep "e2e-test-change: true"       │
│  └── Verify the ACTUAL resource matches what we pushed                      │
│                                                                             │
│  Step 12: Cleanup                                                           │
│  └── Delete kind cluster, delete test branch                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Test Configurations

### Preset Tests

| Test | Config | What It Validates |
|------|--------|-------------------|
| Minimal Preset | `minimal-config.yaml` | Namespaces only, basic structure |
| Standard Preset | `standard-config.yaml` | Full infra: namespaces, RBAC, netpol, quotas |
| Enterprise Preset | `enterprise-config.yaml` | Enterprise features, monitoring dirs |

### Scope Tests

| Test | Config | What It Validates |
|------|--------|-------------------|
| Infra-Only | `infra-only-config.yaml` | No applications/, only infrastructure/ |
| App-Only | `app-only-config.yaml` | No infrastructure/, only applications/ |

### Topology Tests

| Test | Config | What It Validates |
|------|--------|-------------------|
| Namespace-Scope | `namespace-scope-config.yaml` | ArgoCD in namespace-scoped mode |
| Multi-Cluster (2 from start) | `multi-cluster-2-from-start-config.yaml` | 2 clusters from day 1 |
| Multi-Cluster (add later) | `single-cluster-config.yaml` | Day 2: add cluster via gitopsi |

---

## Multi-Cluster Tests

### Test 1: 2 Clusters from the Start

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MULTI-CLUSTER TEST 1                                     │
│              "Multi-cluster from the start"                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Initial Setup: 1 ArgoCD managing 2 clusters from day 1                     │
│                                                                             │
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────┐        │
│  │   ArgoCD     │────────►│  Cluster 1   │         │  Cluster 2   │        │
│  │  (Hub)       │────────►│  (dev)       │         │  (prod)      │        │
│  └──────────────┘         └──────────────┘         └──────────────┘        │
│                                                                             │
│  Step 1: Create 2 kind clusters (dev + prod)                                │
│  Step 2: gitopsi init --config multi-cluster-2-envs.yaml                    │
│  Step 3: Push to Git                                                        │
│  Step 4: Bootstrap ArgoCD on cluster 1 (hub)                                │
│  Step 5: Register cluster 2 with ArgoCD                                     │
│  Step 6: ArgoCD syncs to BOTH clusters                                      │
│  Step 7: Validate resources exist on BOTH clusters                          │
│  Step 8-12: Change, push, verify sync on correct cluster                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Test 2: Add Cluster Later (Day 2 Operation)

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MULTI-CLUSTER TEST 2                                     │
│              "Add cluster later (Day 2 operation)"                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Phase A: Start with 1 cluster only                                         │
│  ─────────────────────────────────────                                      │
│  Step 1-7: Complete flow with single cluster                                │
│                                                                             │
│  Phase B: Add new cluster via gitopsi (Day 2)                               │
│  ────────────────────────────────────────────                               │
│  Step 8: Create 2nd kind cluster (prod)                                     │
│  Step 9: gitopsi env add-cluster prod --url <cluster2-url>                  │
│  Step 10: Push to Git                                                       │
│  Step 11: Verify ArgoCD detects and manages new cluster                     │
│  Step 12: Cleanup                                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Configuration-Specific Validation

Each configuration validates different resources:

| Configuration | Step 7 Validates | Step 8 Changes | Step 11 Validates |
|---------------|------------------|----------------|-------------------|
| Minimal | namespace exists | add label | label present |
| Standard | ns + RBAC + netpol + quota | modify quota | quota updated |
| Enterprise | all + enterprise dirs | add label | label present |
| Infra-Only | infra exists, NO apps | modify RBAC | RBAC updated |
| App-Only | apps exist, NO infra | change replicas | pods match |
| Namespace-Scope | ArgoCD ns-scoped | modify config | still ns-scoped |
| Multi-Cluster | both clusters | add to specific | only target updated |

---

## Test Fixtures

Located in `test/e2e/fixtures/`:

| File | Purpose |
|------|---------|
| `minimal-config.yaml` | Minimal preset (namespaces only) |
| `standard-config.yaml` | Standard preset (full infra) |
| `enterprise-config.yaml` | Enterprise preset |
| `infra-only-config.yaml` | Infrastructure-only scope |
| `app-only-config.yaml` | Application-only scope |
| `namespace-scope-config.yaml` | Namespace-scoped ArgoCD |
| `multi-cluster-config.yaml` | Multi-cluster topology |
| `multi-cluster-2-from-start-config.yaml` | 2 clusters from day 1 |
| `single-cluster-config.yaml` | Single cluster for day 2 test |
| `day2-ops-config.yaml` | Day 2 operations test |

---

## Running E2E Tests

### GitHub Actions (Automatic)

The E2E test suite runs automatically:

- **Weekly**: Sunday at 2 AM UTC (scheduled)
- **Manual**: Via workflow_dispatch

### Manual Trigger

1. Go to Actions -> E2E Full Test -> Run workflow
2. Select test suite:
   - `all` - Run all tests
   - `minimal` - Minimal preset flow only
   - `standard` - Standard preset flow only
   - `enterprise` - Enterprise preset flow only
   - `infra-only` - Infra-only scope flow only
   - `app-only` - App-only scope flow only
   - `namespace-scope` - Namespace-scope flow only
   - `multi-cluster` - Both multi-cluster tests
   - `marketplace` - Marketplace tests only
3. Enable debug if needed

### Local Testing

```bash
# Build binary
go build -o bin/gitopsi ./cmd/gitopsi

# Run generation test
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml --output /tmp/test

# Run with kind cluster (full flow)
kind create cluster --name e2e-test
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml \
  --output /tmp/test --bootstrap --bootstrap-mode helm

# Verify resources
kubectl get namespaces
kubectl get all -n dev

# Make a change
yq eval '.metadata.labels["test"] = "true"' -i /tmp/test/*/infrastructure/base/namespaces/dev.yaml

# Cleanup
kind delete cluster --name e2e-test
```

---

## Summary Report

After each run, a summary report is generated with:

```markdown
# E2E Test Summary Report

**Run ID:** 123456789
**Version:** v1.0.0
**Commit:** abc1234
**Date:** 2026-01-08T12:00:00Z

## Complete Flow Test Results

| Test Flow | Status | Description |
|-----------|--------|-------------|
| Build | ✅ | Binary compilation |
| Minimal Preset | ✅ | 12-step flow: namespace only |
| Standard Preset | ✅ | 12-step flow: full infra |
| Enterprise Preset | ✅ | 12-step flow: enterprise features |
| Infra-Only Scope | ✅ | 12-step flow: no applications |
| App-Only Scope | ✅ | 12-step flow: apps only |
| Namespace-Scope | ✅ | 12-step flow: ns-scoped ArgoCD |
| Multi-Cluster (2 start) | ✅ | 2 clusters from day 1 |
| Multi-Cluster (add later) | ✅ | Day 2: add cluster |
| Marketplace | ✅ | Pattern management |

## 12-Step Flow Validated

Each preset/scope test validates:
1. Generate manifests with gitopsi
2. Create Kind cluster
3. Push to Git
4. Bootstrap ArgoCD
5. Configure ArgoCD to sync from Git
6. Verify initial sync
7. Validate resources on cluster
8. Make code changes
9. Push changes to Git
10. Verify ArgoCD syncs changes
11. Validate changes applied on cluster
12. Cleanup
```

---

## Debugging Failures

### View Logs

1. Go to Actions -> [Failed Run]
2. Expand failed job
3. Check step logs

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Kind cluster timeout | Resource limits | Increase timeout |
| ArgoCD not ready | Pod startup | Check pod logs |
| Sync failure | Invalid manifests | Run kustomize build locally |
| Branch push failed | Permissions | Check GitHub_TOKEN |
| Cluster validation fails | Sync not complete | Increase sleep time |

### Reproduce Locally

```bash
# Use same versions as CI
export KIND_VERSION=v0.24.0
export ARGOCD_VERSION=2.13.2

# Create cluster
kind create cluster --name e2e-test

# Follow the 12-step flow manually
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml --output /tmp/test
./bin/gitopsi init --bootstrap --bootstrap-mode helm --output /tmp/bootstrap

# Check ArgoCD
kubectl get pods -n argocd
kubectl wait --for=condition=available deployment/argocd-server -n argocd

# Get password and login
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

---

## Adding New Tests

1. Create test fixture in `test/e2e/fixtures/`
2. Add new job in `.github/workflows/e2e-full.yml`
3. Follow the 12-step flow pattern
4. Add configuration-specific validation in Step 7 and Step 11
5. Update this documentation
6. Run tests locally first
7. Create PR with test changes
