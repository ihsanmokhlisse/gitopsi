# Testing Strategy

## Overview

gitopsi uses a containerized testing approach to ensure consistency across all environments.

**⚠️ STRICT RULE: All testing MUST use Podman Desktop containers. Direct local install/test is FORBIDDEN.**

---

## Testing Pyramid

```
                    ┌─────────────┐
                    │    E2E      │  ← Full workflow tests
                   ┌┴─────────────┴┐
                   │  Integration  │  ← Component interaction
                  ┌┴───────────────┴┐
                  │    Unit Tests    │  ← Individual functions
                 ┌┴─────────────────┴┐
                 │   Static Analysis  │  ← Linting, formatting
                └─────────────────────┘
```

---

## Test Categories

### 1. Static Analysis (Pre-commit)

| Check | Tool | Command |
|-------|------|---------|
| Formatting | gofmt | `make fmt-check` |
| Imports | goimports | `make imports-check` |
| Linting | golangci-lint | `make lint` |
| Security | gosec | `make security` |
| Vulnerabilities | govulncheck | `make vuln` |

### 2. Unit Tests

| Package | Coverage Target | Focus |
|---------|-----------------|-------|
| `internal/config` | >90% | Config loading, validation |
| `internal/output` | >90% | File writing, dry-run |
| `internal/templates` | >80% | Template rendering |
| `internal/generator` | >80% | Manifest generation |
| `internal/prompt` | >70% | User input (mocked) |
| `internal/cli` | >70% | Command execution |

### 3. Integration Tests

| Test | Description |
|------|-------------|
| Config → Generator | Load config, generate output |
| CLI → File System | Run init, verify files |
| Template → Output | Render templates, validate YAML |

### 4. End-to-End Tests

| Test | Description |
|------|-------------|
| Interactive Mode | Simulate user prompts |
| Config File Mode | Full generation from YAML |
| Dry Run Mode | Verify no files written |
| Platform Variants | K8s, OpenShift, AKS, EKS |

---

## Test Execution

### Local Development (Podman Required)

```bash
# Run all checks (recommended before commit)
make check

# Individual targets
make fmt-check      # Check formatting
make lint           # Run linter
make test           # Run unit tests
make test-coverage  # With coverage report
make test-e2e       # End-to-end tests

# Full CI simulation locally
make ci-local
```

### Container-Based Testing (REQUIRED)

```bash
# Build test container
make container-build

# Run all tests in container
make container-test

# Run specific test in container
make container-run ARGS="go test -v ./internal/config/..."

# Interactive shell for debugging
make container-shell
```

### Pre-Push Validation

```bash
# Run before every push (automated via git hook)
make pre-push

# This runs:
# 1. Format check
# 2. Lint
# 3. Unit tests
# 4. Build verification
```

---

## Coverage Requirements

| Phase | Minimum Coverage |
|-------|------------------|
| Phase 1 (MVP) | 70% |
| Phase 2 (Platforms) | 75% |
| Phase 3 (Features) | 80% |
| v1.0 Release | 85% |

### Coverage Commands

```bash
# Generate coverage report
make test-coverage

# View HTML report
make coverage-html

# Check coverage threshold
make coverage-check
```

---

## Test Data Management

### Golden Files

Location: `testdata/golden/`

```
testdata/
├── golden/
│   ├── kubernetes/          # Expected K8s output
│   ├── openshift/           # Expected OpenShift output
│   ├── argocd/              # Expected ArgoCD output
│   └── docs/                # Expected doc output
├── configs/
│   ├── valid/               # Valid test configs
│   └── invalid/             # Invalid test configs
└── fixtures/
    └── ...                  # Test fixtures
```

### Updating Golden Files

```bash
# Regenerate golden files (use with caution)
make golden-update

# Verify golden files match
make golden-check
```

---

## Platform Testing Matrix

| Platform | Architecture | Container | CI |
|----------|--------------|-----------|-----|
| Linux | amd64 | ✅ | ✅ |
| Linux | arm64 | ✅ | ✅ |
| macOS | amd64 | ✅ | ✅ |
| macOS | arm64 | ✅ | ✅ |
| Windows | amd64 | ✅ | ✅ |

---

## CI/CD Pipeline

### GitHub Actions Workflow

```
Push/PR → Lint → Test → Build → [Release]
                              ↓
                    Tag v*.*.* triggers:
                    - Multi-platform builds
                    - GitHub Release
                    - Homebrew formula
                    - Container images
```

### Local CI Simulation

```bash
# Simulate full CI pipeline locally
make ci-local

# This runs the same checks as GitHub Actions:
# 1. go mod verify
# 2. gofmt check
# 3. golangci-lint
# 4. go test with race detector
# 5. go build for all platforms
```

---

## Debugging Tests

### Verbose Output

```bash
# Verbose test output
make test-verbose

# Single package verbose
make container-run ARGS="go test -v -run TestConfig ./internal/config/..."
```

### Test Filtering

```bash
# Run specific test
make container-run ARGS="go test -v -run TestValidate ./..."

# Run tests matching pattern
make container-run ARGS="go test -v -run 'Test.*Config' ./..."
```

---

## E2E Full Testing (CI)

The E2E full testing workflow (`e2e-full.yml`) runs comprehensive end-to-end tests using a real Kubernetes cluster.

### Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                    E2E Full Test Workflow                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. SETUP                                                           │
│     ├── Build gitopsi binary                                        │
│     └── Create kind Kubernetes cluster                              │
│                                                                     │
│  2. GENERATE & VALIDATE                                             │
│     ├── Run gitopsi init with minimal/standard/enterprise presets   │
│     ├── Validate YAML syntax with yq                                │
│     ├── Run kustomize build on all overlays                         │
│     └── Run gitopsi validate                                        │
│                                                                     │
│  3. GIT PUSH                                                        │
│     ├── Create test branch: test/e2e-{version}-{commit}             │
│     └── Push generated manifests to GitHub                          │
│                                                                     │
│  4. BOOTSTRAP                                                       │
│     ├── Install ArgoCD via Helm                                     │
│     ├── Wait for ArgoCD to be ready                                 │
│     └── Configure argocd CLI                                        │
│                                                                     │
│  5. SYNC                                                            │
│     ├── Create ArgoCD Application pointing to test branch           │
│     ├── Trigger sync and wait for completion                        │
│     └── Verify resources created in cluster                         │
│                                                                     │
│  6. CLI TESTING                                                     │
│     ├── Test all major CLI commands                                 │
│     └── Verify exit codes and output                                │
│                                                                     │
│  7. CLEANUP                                                         │
│     ├── Delete test branch from GitHub                              │
│     └── Delete kind cluster                                         │
│                                                                     │
│  8. REPORTING (on failure)                                          │
│     ├── Create GitHub issue with failure details                    │
│     └── Upload debugging artifacts                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Trigger Conditions

| Trigger | Frequency | Purpose |
|---------|-----------|---------|
| Schedule | Weekly (Sun 2AM UTC) | Catch regressions |
| workflow_dispatch | Manual | On-demand testing |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GO_VERSION` | Go version (default: 1.23) |
| `KIND_VERSION` | Kind version (default: v0.24.0) |
| `ARGOCD_VERSION` | ArgoCD version (default: 2.13.2) |
| `KUSTOMIZE_VERSION` | Kustomize version (default: 5.5.0) |
| `HELM_VERSION` | Helm version (default: 3.16.3) |

### Manual Trigger Options

1. Go to Actions → E2E Full Test → Run workflow
2. Select options:
   - **preset**: Which preset to test (all, minimal, standard, enterprise)
   - **debug**: Enable debug logging

### Test Branch Naming

Test branches follow the pattern: `test/e2e-{version}-{commit}`

Example: `test/e2e-v0.2.0-abc1234`

These branches are automatically deleted after the test completes (success or failure).

---

## Running E2E Tests Locally

### Prerequisites

- Podman or Docker
- kind installed
- kubectl configured
- kustomize installed
- ArgoCD CLI (for sync tests)

### Basic E2E Tests (No Cluster)

```bash
# Build and run basic E2E tests
podman run --rm -v $(pwd):/app:Z golang:1.23 sh -c \
  "cd /app && go build -o bin/gitopsi ./cmd/gitopsi && \
   go test -tags=e2e ./test/e2e/..."
```

### Full E2E with kind Cluster

```bash
# 1. Create kind cluster
kind create cluster --name gitopsi-e2e

# 2. Build binary
go build -o bin/gitopsi ./cmd/gitopsi

# 3. Set environment variables
export KUBECONFIG=$(kind get kubeconfig --name gitopsi-e2e)
export E2E_CLUSTER=true

# 4. Run E2E tests
go test -tags=e2e -v ./test/e2e/...

# 5. Cleanup
kind delete cluster --name gitopsi-e2e
```

### E2E with ArgoCD

```bash
# Prerequisites: kind cluster running

# 1. Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl wait --for=condition=available -n argocd deployment/argocd-server --timeout=300s

# 2. Set environment
export E2E_CLUSTER=true
export E2E_ARGOCD=true
export ARGOCD_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)

# 3. Run bootstrap tests
go test -tags=e2e -v -run TestArgoCD ./test/e2e/...
```

---

## Debugging E2E Failures

### Viewing Logs

1. Go to Actions → E2E Full Test → [Failed Run]
2. Expand failed job to see logs
3. Download artifacts for kubeconfig and generated manifests

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| kind cluster creation timeout | Resource limits | Increase timeout or use smaller cluster |
| ArgoCD not ready | Pod startup time | Check pod logs in ArgoCD namespace |
| Sync failure | Invalid manifests | Run kustomize build locally |
| Branch push failed | Permissions | Check GitHub_TOKEN permissions |

### Reproducing Locally

```bash
# Use same versions as CI
export KIND_VERSION=v0.24.0
export ARGOCD_VERSION=2.13.2

# Create identical environment
kind create cluster --name e2e-test

# Follow workflow steps manually
./bin/gitopsi init --config test/e2e/fixtures/standard-config.yaml --output /tmp/test
kustomize build /tmp/test/test-standard/infrastructure/overlays/dev
./bin/gitopsi validate /tmp/test/test-standard
```

---

## Best Practices

1. **Always test in container** - Never test directly on host
2. **Run `make check` before commit** - Catches issues early
3. **Maintain golden files** - Update when output changes intentionally
4. **Table-driven tests** - Use for comprehensive coverage
5. **Mock external dependencies** - File system, user input
6. **Test error paths** - Invalid input, missing files
7. **Check E2E status** - Review weekly E2E runs for regressions
