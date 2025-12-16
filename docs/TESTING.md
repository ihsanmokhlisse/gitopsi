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

## Best Practices

1. **Always test in container** - Never test directly on host
2. **Run `make check` before commit** - Catches issues early
3. **Maintain golden files** - Update when output changes intentionally
4. **Table-driven tests** - Use for comprehensive coverage
5. **Mock external dependencies** - File system, user input
6. **Test error paths** - Invalid input, missing files

