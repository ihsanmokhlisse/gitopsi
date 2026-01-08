# E2E Testing Documentation

## Overview

The gitopsi E2E test suite validates the **complete gitopsi workflow** - from initial
setup to day-2 operations. Every operation is performed through gitopsi, and ArgoCD
syncs changes from Git.

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

## Test Suite Architecture

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                    COMPREHENSIVE E2E TEST SUITE                             │
└─────────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────┐
                              │    BUILD    │
                              │   Binary    │
                              └──────┬──────┘
                                     │
        ┌────────────────────────────┼────────────────────────────┐
        │                            │                            │
        ▼                            ▼                            ▼
┌───────────────┐          ┌─────────────────┐          ┌─────────────────┐
│   CLI TESTS   │          │ GENERATION TESTS│          │MARKETPLACE TESTS│
│               │          │                 │          │                 │
│ • version     │          │ Presets:        │          │ • list          │
│ • help        │          │ • minimal       │          │ • search        │
│ • all cmds    │          │ • standard      │          │ • info          │
└───────────────┘          │ • enterprise    │          │ • validate      │
                           │                 │          │ • install       │
                           │ Scopes:         │          └─────────────────┘
                           │ • infra-only    │
                           │ • app-only      │
                           │ • both          │
                           │                 │
                           │ Topologies:     │
                           │ • namespace     │
                           │ • cluster-per-env│
                           │ • multi-cluster │
                           └─────────────────┘
                                     │
        ┌────────────────────────────┼────────────────────────────┐
        │                            │                            │
        ▼                            ▼                            ▼
┌───────────────────┐      ┌─────────────────┐      ┌─────────────────────┐
│  BOOTSTRAP TESTS  │      │   SYNC TESTS    │      │  DAY 2 OPS TESTS    │
│                   │      │                 │      │                     │
│ Bootstrap modes:  │      │ Full GitOps:    │      │ • Add cluster       │
│ • helm            │      │ 1. Generate     │      │ • Add application   │
│ • manifest        │      │ 2. Push to Git  │      │ • Add operator      │
│                   │      │ 3. ArgoCD sync  │      │ • Promote app       │
│ Verifies:         │      │ 4. Deploy       │      │ • Multi-repo        │
│ • ArgoCD running  │      │ 5. Verify       │      │                     │
│ • CRDs installed  │      │                 │      │ All via gitopsi     │
│ • Manifests valid │      │                 │      │ Git push → sync     │
└───────────────────┘      └─────────────────┘      └─────────────────────┘
                                     │
                                     │
        ┌────────────────────────────┼────────────────────────────┐
        │                            │                            │
        ▼                            ▼                            ▼
┌───────────────────┐      ┌─────────────────┐      ┌─────────────────────┐
│   ORG MGMT TESTS  │      │  MULTI-REPO     │      │   SUMMARY REPORT    │
│                   │      │     TESTS       │      │                     │
│ • org init        │      │                 │      │ • Pass/Fail table   │
│ • team create     │      │ Repo patterns:  │      │ • Coverage details  │
│ • team quota      │      │ • infra repo    │      │ • Run metadata      │
│ • project create  │      │ • apps repo     │      │                     │
│ • project add-env │      │ • app-per-repo  │      │ On failure:         │
│                   │      │                 │      │ • Create GH issue   │
└───────────────────┘      └─────────────────┘      └─────────────────────┘
```

---

## Test Categories

### 1. CLI Tests

Basic CLI functionality validation.

| Test | Command | Purpose |
|------|---------|---------|
| Version | `gitopsi version` | Binary built correctly |
| Help | `gitopsi --help` | All commands registered |
| Init Help | `gitopsi init --help` | Init options documented |
| Validate Help | `gitopsi validate --help` | Validate options documented |
| Marketplace Help | `gitopsi marketplace --help` | Marketplace available |
| Auth Help | `gitopsi auth --help` | Auth commands available |
| Env Help | `gitopsi env --help` | Env commands available |

### 2. Generation Tests

Manifest generation without cluster (7 configurations):

| Config | Preset | Scope | Topology | Expected Dirs |
|--------|--------|-------|----------|---------------|
| minimal | minimal | both | namespace | infrastructure |
| standard | standard | both | namespace | infrastructure, ArgoCD |
| enterprise | enterprise | both | namespace | infrastructure, ArgoCD |
| infra-only | standard | infrastructure | namespace | infrastructure |
| app-only | standard | application | namespace | applications |
| namespace-scope | standard | both | namespace-based | infrastructure, ArgoCD |
| multi-cluster | standard | both | cluster-per-env | infrastructure, ArgoCD |

Each test:

1. Runs `gitopsi init --config <config>`
2. Verifies expected directories exist
3. Validates YAML syntax with yq
4. Validates with kustomize build
5. Runs `gitopsi validate`

### 3. Bootstrap Tests

Tests gitopsi's ability to install ArgoCD on a cluster (2 configurations):

| Mode | Command | Verifies |
|------|---------|----------|
| Helm | `gitopsi init --bootstrap --bootstrap-mode helm` | ArgoCD installed via Helm |
| Manifest | `gitopsi init --bootstrap --bootstrap-mode manifest` | ArgoCD installed via manifests |

Each test:

1. Creates kind cluster
2. Runs `gitopsi preflight`
3. Runs `gitopsi init --bootstrap`
4. Verifies ArgoCD namespace exists
5. Verifies ArgoCD pods are running
6. Verifies ArgoCD CRDs installed
7. Verifies generated manifests

### 4. Sync Tests

Full GitOps cycle - generate → push → sync:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SYNC TEST WORKFLOW                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Step 1: Generate Manifests                                                 │
│  ─────────────────────────────                                              │
│  gitopsi init --config standard-config.yaml --output /tmp/output            │
│                                                                             │
│  Step 2: Push to Git                                                        │
│  ───────────────────                                                        │
│  git checkout --orphan test/e2e-<version>-<commit>                          │
│  git add . && git commit && git push                                        │
│                                                                             │
│  Step 3: Create Cluster                                                     │
│  ─────────────────────                                                      │
│  kind create cluster --name e2e-sync-test                                   │
│                                                                             │
│  Step 4: Bootstrap ArgoCD                                                   │
│  ────────────────────────                                                   │
│  gitopsi init --bootstrap --bootstrap-mode helm                             │
│                                                                             │
│  Step 5: Create ArgoCD Application                                          │
│  ─────────────────────────────────                                          │
│  argocd app create e2e-sync-app                                             │
│    --repo https://github.com/<repo>.git                                     │
│    --revision test/e2e-<version>-<commit>                                   │
│    --path infrastructure/overlays/dev                                       │
│    --sync-policy automated                                                  │
│                                                                             │
│  Step 6: Verify Sync                                                        │
│  ────────────────────                                                       │
│  argocd app sync e2e-sync-app                                               │
│  argocd app wait e2e-sync-app --sync                                        │
│  kubectl get namespaces (verify dev namespace created)                      │
│                                                                             │
│  Step 7: Cleanup                                                            │
│  ───────────────                                                            │
│  kind delete cluster                                                        │
│  git push --delete origin test/e2e-<branch>                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5. Day 2 Operations Tests

Tests gitopsi's ability to manage resources after initial setup:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                      DAY 2 OPERATIONS TEST WORKFLOW                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Phase 1: Initial Setup                                                     │
│  ──────────────────────                                                     │
│  gitopsi init --bootstrap --bootstrap-mode helm                             │
│  (Creates cluster with ArgoCD + initial manifests)                          │
│                                                                             │
│  Phase 2: Add New Cluster                                                   │
│  ────────────────────────                                                   │
│  gitopsi env add-cluster prod --url https://prod.k8s.local:6443             │
│  git add . && git commit && git push                                        │
│  → Verify ArgoCD can see new cluster                                        │
│                                                                             │
│  Phase 3: Add Application                                                   │
│  ────────────────────────                                                   │
│  gitopsi install nginx --env dev,staging                                    │
│  git add . && git commit && git push                                        │
│  → Verify ArgoCD syncs nginx application                                    │
│                                                                             │
│  Phase 4: Add Operator                                                      │
│  ─────────────────────                                                      │
│  gitopsi operator add prometheus-operator                                   │
│  git add . && git commit && git push                                        │
│  → Verify operator manifests generated                                      │
│  → Verify ArgoCD syncs operator subscription                                │
│                                                                             │
│  Phase 5: Promote Application                                               │
│  ───────────────────────────                                                │
│  gitopsi promote nginx --from dev --to staging                              │
│  git add . && git commit && git push                                        │
│  → Verify staging now has nginx                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6. Multi-Repo Tests

Tests gitopsi's ability to manage multiple repositories:

| Pattern | Description | Use Case |
|---------|-------------|----------|
| Mono-repo | All in one repo | Small teams, simple projects |
| Infra + Apps | Separate repos for infra and apps | Medium teams, separation of concerns |
| App-per-repo | Each app in its own repo | Large teams, microservices |

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                      MULTI-REPO PATTERNS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Pattern 1: Mono-repo                                                       │
│  ────────────────────                                                       │
│  gitops-repo/                                                               │
│  ├── infrastructure/                                                        │
│  ├── applications/                                                          │
│  └── argocd/                                                                │
│                                                                             │
│  Pattern 2: Infra + Apps separation                                         │
│  ───────────────────────────────────                                        │
│  gitops-infra-repo/           gitops-apps-repo/                             │
│  ├── infrastructure/          ├── app1/                                     │
│  ├── argocd/                  ├── app2/                                     │
│  └── operators/               └── app3/                                     │
│                                                                             │
│  Pattern 3: App-per-repo                                                    │
│  ───────────────────────                                                    │
│  gitops-infra/    app1-repo/    app2-repo/    app3-repo/                    │
│  └── infra...     └── manifests └── manifests └── manifests                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 7. Organization Management Tests

Tests gitopsi's organization, team, and project management:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ORGANIZATION MANAGEMENT TEST                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Step 1: Initialize Organization                                            │
│  ───────────────────────────────                                            │
│  gitopsi org init acme-corp --domain acme.com                               │
│                                                                             │
│  Step 2: Create Teams                                                       │
│  ────────────────────                                                       │
│  gitopsi team create frontend --owners frontend@acme.com                    │
│  gitopsi team create backend --owners backend@acme.com                      │
│  gitopsi team create platform --owners platform@acme.com                    │
│                                                                             │
│  Step 3: Set Team Quotas                                                    │
│  ───────────────────────                                                    │
│  gitopsi team set-quota frontend --cpu 50 --memory 100Gi                    │
│  gitopsi team set-quota backend --cpu 100 --memory 200Gi                    │
│                                                                             │
│  Step 4: Create Projects                                                    │
│  ───────────────────────                                                    │
│  gitopsi project create web-app --team frontend                             │
│  gitopsi project create api-service --team backend                          │
│                                                                             │
│  Step 5: Add Environments to Projects                                       │
│  ─────────────────────────────────────                                      │
│  gitopsi project add-env web-app dev --team frontend                        │
│  gitopsi project add-env web-app staging --team frontend                    │
│  gitopsi project add-env web-app prod --team frontend                       │
│                                                                             │
│  Step 6: Verify Structure                                                   │
│  ────────────────────────                                                   │
│  gitopsi org show                                                           │
│  gitopsi team list                                                          │
│  gitopsi project list --team frontend                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 8. Marketplace Tests

Tests pattern installation from marketplace:

| Test | Command | Purpose |
|------|---------|---------|
| List | `gitopsi marketplace list` | List available patterns |
| Categories | `gitopsi marketplace categories` | List categories |
| Search | `gitopsi marketplace search monitoring` | Search patterns |
| Info | `gitopsi marketplace info prometheus` | Get pattern details |
| Validate | `gitopsi marketplace validate prometheus` | Validate pattern |
| Install | `gitopsi install prometheus --env dev` | Install pattern |

---

## Running E2E Tests

### GitHub Actions (Automatic)

The E2E test suite runs automatically:

- **Weekly**: Sunday at 2 AM UTC (scheduled)
- **Manual**: Via workflow_dispatch

### Manual Trigger

1. Go to Actions → E2E Full Test → Run workflow
2. Select test suite:
   - `all` - Run all tests
   - `generation` - Run only generation tests
   - `bootstrap` - Run only bootstrap tests
   - `sync` - Run only sync tests
   - `marketplace` - Run only marketplace tests
3. Enable debug if needed

### Local Testing

```bash
# Build binary
go build -o bin/gitopsi ./cmd/gitopsi

# Run generation tests
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml --output /tmp/test

# Run with kind cluster
kind create cluster --name e2e-test
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml \
  --output /tmp/test --bootstrap --bootstrap-mode helm

# Cleanup
kind delete cluster --name e2e-test
```

---

## Test Fixtures

Located in `test/e2e/fixtures/`:

| File | Purpose |
|------|---------|
| `minimal-config.yaml` | Minimal preset test |
| `standard-config.yaml` | Standard preset test |
| `enterprise-config.yaml` | Enterprise preset test |
| `infra-only-config.yaml` | Infrastructure-only scope |
| `app-only-config.yaml` | Application-only scope |
| `namespace-scope-config.yaml` | Namespace-scoped ArgoCD |
| `multi-cluster-config.yaml` | Multi-cluster topology |
| `day2-ops-config.yaml` | Day 2 operations test |
| `multi-repo-config.yaml` | Multi-repo test |

---

## Summary Report

After each run, a summary report is generated with:

```markdown
# E2E Test Summary Report

**Run ID:** 123456789
**Version:** v1.0.0
**Commit:** abc1234
**Date:** 2026-01-08T12:00:00Z

## Test Results

| Test Suite | Status |
|------------|--------|
| Build | ✅ Passed |
| CLI Tests | ✅ Passed |
| Generation Tests | ✅ Passed |
| Bootstrap Tests | ✅ Passed |
| Sync Tests | ✅ Passed |
| Day 2 Ops Tests | ✅ Passed |
| Multi-Repo Tests | ✅ Passed |
| Org Mgmt Tests | ✅ Passed |
| Marketplace Tests | ✅ Passed |

## Test Coverage

### Generation Tests
- ✅ Minimal preset
- ✅ Standard preset
- ✅ Enterprise preset
- ✅ Infrastructure-only scope
- ✅ Application-only scope
- ✅ Namespace-scope topology
- ✅ Multi-cluster topology

### Bootstrap Tests
- ✅ Helm bootstrap mode
- ✅ Manifest bootstrap mode

### Day 2 Operations
- ✅ Add cluster
- ✅ Add application
- ✅ Add operator
- ✅ Promote application

### Multi-Repo
- ✅ Infra + apps separation
- ✅ App-per-repo pattern

### Organization Management
- ✅ Organization init
- ✅ Team create
- ✅ Team quotas
- ✅ Project create
- ✅ Project environments
```

---

## Debugging Failures

### View Logs

1. Go to Actions → [Failed Run]
2. Expand failed job
3. Check step logs

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Kind cluster timeout | Resource limits | Increase timeout |
| ArgoCD not ready | Pod startup | Check pod logs |
| Sync failure | Invalid manifests | Run kustomize build locally |
| Branch push failed | Permissions | Check GitHub_TOKEN |

### Reproduce Locally

```bash
# Use same versions as CI
export KIND_VERSION=v0.24.0
export ARGOCD_VERSION=2.13.2

# Create cluster
kind create cluster --name e2e-test

# Follow workflow steps
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml --output /tmp/test
./bin/gitopsi validate /tmp/test/test-standard
```

---

## Adding New Tests

1. Create test fixture in `test/e2e/fixtures/`
2. Add matrix entry in `.github/workflows/e2e-full.yml`
3. Update documentation in `docs/E2E_TESTING.md`
4. Run tests locally first
5. Create PR with test changes
