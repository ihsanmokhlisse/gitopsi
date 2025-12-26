# gitopsi - Product Requirements Document

## Executive Summary

**gitopsi** is a Go CLI tool that bootstraps production-ready GitOps repository structures. Unlike existing tools (ArgoCD Autopilot, Flux CLI), gitopsi supports multiple platforms (Kubernetes, OpenShift, AKS, EKS), multiple GitOps tools (ArgoCD, Flux), and handles both infrastructure and application scopes.

**Vision:** One command to get from zero to a fully synced GitOps cluster.

```bash
gitopsi init --git-url git@github.com:org/repo.git --cluster https://api.ocp.com --bootstrap
```

---

## Core Philosophy

> *"Complex things should be simple to do, and simple things should be trivial."*

### Guiding Principles

1. **🎯 Simplicity First** - One command to start, zero config for basics
2. **📚 Best Practices by Default** - Every generated repo follows industry standards
3. **🔄 Two Modes, Same Output** - Interactive (human) and Config (automation)
4. **🏗️ Opinionated but Flexible** - Smart defaults with escape hatches
5. **📖 Self-Documenting** - Generated repos are understandable and maintainable
6. **🔧 12-Factor Compliant** - All configuration externalized, no hardcoded values

### 12-Factor App Compliance

gitopsi follows the [12-Factor App](https://12factor.net/) methodology:

| Factor | Implementation |
|--------|----------------|
| **I. Codebase** | One codebase in Git, multiple deploys |
| **II. Dependencies** | Go modules, explicit dependency declaration |
| **III. Config** | **All config from env vars, files, or flags - ZERO hardcoded values** |
| **IV. Backing Services** | Git providers, clusters as attached resources |
| **V. Build/Release/Run** | Separate build (binary), release (config), run stages |
| **VI. Processes** | Stateless CLI execution |
| **VII. Port Binding** | N/A (CLI tool) |
| **VIII. Concurrency** | Parallel file generation where possible |
| **IX. Disposability** | Fast startup, graceful shutdown |
| **X. Dev/Prod Parity** | Same config structure for all environments |
| **XI. Logs** | Stream to stdout/stderr |
| **XII. Admin Processes** | One-off commands (validate, migrate, etc.) |

#### Configuration Hierarchy (Factor III)

All values are configurable through three layers (highest to lowest priority):

```
1. CLI Flags         (--git-url, --cluster, etc.)
2. Environment Vars  (GITOPSI_GIT_URL, GITOPSI_CLUSTER_URL, etc.)
3. Config File       (gitops.yaml)
4. Auto-Detection    (kubeconfig, git remote, etc.)
```

**RULE: NO HARDCODED VALUES**
- ❌ Never: `repoURL: "https://github.com/org/repo.git"`
- ✅ Always: `repoURL: {{ .Config.Git.URL }}`

All placeholders MUST be populated from user-provided configuration.

### User Experience Goals

| User Type | Experience |
|-----------|------------|
| Beginners | 30 seconds to working GitOps |
| Teams | Standardized setup across projects |
| Enterprises | Full control with guardrails |

*See Issue [#20](https://github.com/ihsanmokhlisse/gitopsi/issues/20) for full philosophy documentation.*

---

## Competitive Advantage

### Market Gap

| Capability | Autopilot | Flux CLI | Cookiecutter | **gitopsi** |
|------------|-----------|----------|--------------|-------------|
| ArgoCD | ✅ | ❌ | ⚠️ | ✅ |
| Flux | ❌ | ✅ | ⚠️ | ✅ |
| OpenShift | ❌ | ❌ | ❌ | ✅ |
| AKS/EKS | ❌ | ❌ | ❌ | ✅ |
| Infrastructure | ❌ | ❌ | ⚠️ | ✅ |
| Applications | ✅ | ⚠️ | ⚠️ | ✅ |
| Doc Generation | ❌ | ❌ | ❌ | ✅ |
| Config File | ❌ | ❌ | ✅ | ✅ |
| Interactive | ⚠️ | ❌ | ✅ | ✅ |
| Add Commands | ❌ | ❌ | ❌ | ✅ |
| Validation | ❌ | ❌ | ❌ | ✅ |
| **End-to-End Setup** | ❌ | ❌ | ❌ | ✅ |
| **Multi-Git Provider** | ⚠️ | ⚠️ | ❌ | ✅ |
| **Pattern Marketplace** | ❌ | ❌ | ❌ | ✅ |
| **Version Compatibility** | ❌ | ❌ | ❌ | ✅ |
| **Security Scanning** | ❌ | ❌ | ❌ | ✅ |
| **Provenance/Signing** | ❌ | ❌ | ❌ | ✅ |

### Unique Selling Points

1. **"One Tool, All Platforms"** - K8s, OpenShift, AKS, EKS
2. **"Infrastructure + Applications"** - First tool for both scopes
3. **"ArgoCD or Flux"** - No vendor lock-in
4. **"Production Ready"** - Docs, scripts, CI/CD included
5. **"Living Repository"** - Add apps/envs incrementally
6. **"Zero to Synced"** - Auto-push and bootstrap in one command
7. **"Any Git Provider"** - GitHub, GitLab, Gitea, Azure DevOps, Bitbucket
8. **"Version-Aware"** - Compatible manifests for your exact K8s/ArgoCD version
9. **"Secure by Default"** - Security scanning, signing, and provenance built-in

See [COMPETITIVE_ANALYSIS.md](./COMPETITIVE_ANALYSIS.md) for full analysis.

---

## Testing Strategy

### Test Pyramid

```
                    ┌─────────────────┐
                    │    E2E Tests    │  ← OpenShift/Kind cluster validation
                   ─┤                 ├─
                  / │   Integration   │  ← Multi-component tests
                 /  │     Tests       │
                /   ├─────────────────┤
               /    │  Behavior/BDD   │  ← User story validation
              /     │     Tests       │
             /      ├─────────────────┤
            /       │   Unit Tests    │  ← Function-level tests
           /        │   (>80% cov)    │
          ──────────┴─────────────────┴──────────
```

### Test Types Required

| Test Type | Description | Tools | Required For |
|-----------|-------------|-------|--------------|
| **Unit Tests** | Function-level testing | Go `testing`, testify | All code changes |
| **Integration Tests** | Multi-package flow testing | Go `testing` | Core flows |
| **Behavior/BDD Tests** | User story validation | Ginkgo/Gomega | User-facing features |
| **E2E Tests (Cypress)** | UI verification | Cypress | ArgoCD UI features |
| **E2E Tests (Cluster)** | Cluster deployment | Shell scripts | Bootstrap features |
| **Regression Tests** | Bug fix verification | Go `testing` | All bug fixes |
| **Stability Tests** | Load/memory testing | Go benchmarks | Pre-release |
| **Performance Tests** | Speed benchmarks | Go benchmarks | Pre-release |

### Test Coverage Requirements

| Metric | Minimum | Target | Current |
|--------|---------|--------|---------|
| Unit Test Coverage | 70% | 85% | ~65% |
| Integration Test Coverage | 50% | 80% | ~20% |
| E2E Feature Coverage | 80% | 100% | ~60% |
| Regression Test Coverage | 100% | 100% | 0% |

---

## Testing Requirements Per Issue

### Phase 1 - MVP Issues Testing Matrix

| Issue | Feature | Unit | Integration | E2E | Behavior | Regression | Status |
|-------|---------|------|-------------|-----|----------|------------|--------|
| #1 | CLI structure (Cobra) | ✅ `cli_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #2 | Config parsing (Viper) | ✅ `config_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #3 | Interactive prompts | ✅ `prompt_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #4 | Init command | ✅ `init_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #5 | K8s manifest generation | ✅ `generator_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #6 | Kustomize structure | ✅ `generator_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #7 | ArgoCD generation | ✅ `generator_test.go` | ❌ Needed | ✅ Cypress | ❌ Needed | N/A | ⚠️ |
| #8 | Template embedding | ✅ `templates_test.go` | ❌ Needed | ✅ Shell | N/A | N/A | ⚠️ |
| #9 | File output | ✅ `writer_test.go` | ❌ Needed | ✅ Shell | N/A | N/A | ⚠️ |
| #10 | README generation | ✅ `generator_test.go` | ❌ Needed | ✅ Shell | N/A | N/A | ⚠️ |
| #11 | Unit tests | ✅ Multiple | N/A | N/A | N/A | N/A | ✅ |

#### Phase 1 Test Requirements Detail

**Issue #1: CLI Structure with Cobra**
```
Unit Tests (cli_test.go):
  ✅ TestRootCommand
  ✅ TestInitCommand
  ✅ TestVersionCommand
  ❌ TestAllSubcommands (needed)

Integration Tests (NEEDED):
  ❌ TestCLI_ConfigToInit_Flow
  ❌ TestCLI_FlagPrecedence

Behavior Tests (NEEDED):
  ❌ Given user runs 'gitopsi init', When no args, Then prompts interactively
  ❌ Given user provides --config, When file exists, Then uses config

E2E Tests:
  ✅ scripts/e2e-comprehensive-test.sh
```

**Issue #3: Interactive Prompts**
```
Unit Tests (prompt_test.go):
  ✅ TestPromptForString
  ✅ TestPromptForSelect
  ❌ TestPromptValidation (needed)

Behavior Tests (NEEDED):
  ❌ Given user is prompted for project name, When empty, Then shows error
  ❌ Given user selects platform, When OpenShift, Then enables OCP features

E2E Tests (NEEDED):
  ❌ Cypress tests for terminal interaction simulation
```

**Issue #4: Init Command**
```
Unit Tests (init_test.go):
  ✅ TestInitWithConfig
  ✅ TestInitFlags
  ❌ TestInitDryRun (needed)
  ❌ TestInitValidation (needed)

Integration Tests (NEEDED):
  ❌ TestInit_GeneratesCompleteStructure
  ❌ TestInit_WithBootstrap_InstallsArgoCD

E2E Tests:
  ✅ scripts/e2e-openshift-full.sh
  ✅ scripts/e2e-comprehensive-test.sh
```

---

### Phase 2 - Platform Issues Testing Matrix

| Issue | Feature | Unit | Integration | E2E | Behavior | Regression | Status |
|-------|---------|------|-------------|-----|----------|------------|--------|
| #15 | Operator management | ✅ `operator_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #16 | Bootstrap modes | ✅ `bootstrap_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #17 | Environment management | ✅ `environment_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #18 | Customizable generation | ✅ `preset_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #21 | E2E GitOps setup | ✅ `init_test.go` | ❌ Needed | ✅ Shell | ❌ Needed | N/A | ⚠️ |
| #22 | Live progress | ✅ `progress_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #24 | Multi-provider Git | ✅ `github_test.go` etc | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #28 | Version-aware manifests | ✅ `version_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #29 | Security scanning | ✅ `security_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #30 | Validate command | ✅ `validate_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |

#### Phase 2 Test Requirements Detail

**Issue #15: Operator Management**
```
Unit Tests (operator_test.go):
  ✅ TestOperatorStruct
  ✅ TestOperatorManager
  ✅ TestGenerateSubscription
  ❌ TestOperatorPresets (needed)

Integration Tests (NEEDED):
  ❌ TestOperator_WithOLM_Installs
  ❌ TestOperator_MultipleOperators

E2E Tests (NEEDED):
  ❌ Install prometheus-operator via gitopsi
  ❌ Verify operator pods running
  ❌ Verify CRDs created

Behavior Tests (NEEDED):
  ❌ Given operator added, When OLM available, Then installs via subscription
```

**Issue #16: Bootstrap Modes**
```
Unit Tests (bootstrap_test.go):
  ✅ TestModeHelm
  ✅ TestModeOLM
  ✅ TestModeManifest
  ✅ TestModeKustomize

Integration Tests (NEEDED):
  ❌ TestBootstrap_Helm_InstallsArgoCD
  ❌ TestBootstrap_OLM_InstallsArgoCD
  ❌ TestBootstrap_DetectsExisting

E2E Tests:
  ✅ scripts/e2e-openshift-full.sh (OLM mode)
  ✅ scripts/e2e-comprehensive-test.sh (Helm mode)
  ❌ Test manifest mode on air-gapped cluster

Behavior Tests (NEEDED):
  ❌ Given cluster is OpenShift, When bootstrap, Then uses OLM by default
  ❌ Given ArgoCD exists, When bootstrap, Then skips installation
```

**Issue #17: Environment Management**
```
Unit Tests (environment_test.go, manager_test.go):
  ✅ TestEnvironmentStruct
  ✅ TestClusterInfo
  ✅ TestAddEnvironment
  ✅ TestPromote

Integration Tests (NEEDED):
  ❌ TestEnv_Create_GeneratesNamespaces
  ❌ TestEnv_Promote_MovesVersions

E2E Tests (NEEDED):
  ❌ Create dev/staging/prod environments
  ❌ Promote app from dev to staging
  ❌ Verify ArgoCD ApplicationSets update

Behavior Tests (NEEDED):
  ❌ Given multi-cluster topology, When env created, Then generates cluster secrets
```

**Issue #24: Multi-Provider Git**
```
Unit Tests:
  ✅ github_test.go
  ✅ gitlab_test.go
  ✅ gitea_test.go
  ✅ azure_test.go
  ✅ bitbucket_test.go
  ✅ factory_test.go

Integration Tests (NEEDED):
  ❌ TestGitHub_Push_TriggersSync
  ❌ TestGitLab_CreateRepo_SetupWebhook
  ❌ TestGitea_Authenticate_Push

E2E Tests (NEEDED):
  ❌ Full flow with GitHub
  ❌ Full flow with GitLab
  ❌ Full flow with Gitea

Behavior Tests (NEEDED):
  ❌ Given GitHub URL, When init, Then auto-detects provider
  ❌ Given SSH key, When push, Then authenticates correctly
```

**Issue #28: Version-Aware Manifests**
```
Unit Tests (version_test.go):
  ✅ TestKubernetesVersion
  ✅ TestAPIVersionMapping
  ✅ TestVersionMapper
  ❌ TestDeprecationDetection (needed)

Integration Tests (NEEDED):
  ❌ TestGenerate_K8s127_UsesCorrectAPIs
  ❌ TestGenerate_K8s129_NoDeprecations

E2E Tests (NEEDED):
  ❌ Generate for K8s 1.27 cluster
  ❌ Generate for K8s 1.30 cluster
  ❌ Verify no deprecated APIs

Behavior Tests (NEEDED):
  ❌ Given K8s 1.29 target, When generating, Then uses apps/v1
```

**Issue #29: Security Scanning**
```
Unit Tests (security_test.go):
  ✅ TestSecurityScanner
  ✅ TestProvenanceGeneration
  ✅ TestInputSanitization
  ❌ TestManifestSigning (needed)

Integration Tests (NEEDED):
  ❌ TestScan_FindsPrivileged
  ❌ TestScan_FindsMissingLimits
  ❌ TestProvenance_GeneratesSBOM

E2E Tests (NEEDED):
  ❌ Run gitopsi init --scan
  ❌ Verify SBOM generated
  ❌ Verify cosign signature

Behavior Tests (NEEDED):
  ❌ Given privileged container, When scan, Then reports HIGH severity
```

**Issue #30: Validate Command**
```
Unit Tests (validate_test.go):
  ✅ TestValidator
  ✅ TestSchemaValidation
  ✅ TestSecurityScan
  ✅ TestDeprecationCheck

Integration Tests (NEEDED):
  ❌ TestValidate_FullProject
  ❌ TestValidate_WithKubeconform
  ❌ TestValidate_WithTrivy

E2E Tests (NEEDED):
  ❌ Run gitopsi validate on generated project
  ❌ Verify JSON/YAML output
  ❌ Test --fail-on flag

Behavior Tests (NEEDED):
  ❌ Given invalid manifest, When validate, Then fails with details
```

---

### Phase 3 - Enterprise Issues Testing Matrix

| Issue | Feature | Unit | Integration | E2E | Behavior | Regression | Status |
|-------|---------|------|-------------|-----|----------|------------|--------|
| #13 | Multi-repository | ✅ `multirepo_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #14 | Auth management | ✅ `auth_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #19 | Organization mgmt | ✅ `organization_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |
| #20 | Philosophy docs | N/A (docs) | N/A | N/A | N/A | N/A | ✅ |
| #23 | Marketplace | ✅ `marketplace_test.go` | ❌ Needed | ❌ Needed | ❌ Needed | N/A | ⚠️ |

#### Phase 3 Test Requirements Detail

**Issue #23: Marketplace**
```
Unit Tests (marketplace_test.go):
  ✅ TestPattern
  ✅ TestRegistry
  ✅ TestInstaller
  ❌ TestPatternValidation (needed)

Integration Tests (NEEDED):
  ❌ TestMarketplace_Search_ReturnsResults
  ❌ TestMarketplace_Install_GeneratesManifests
  ❌ TestMarketplace_Update_AppliesDiff

E2E Tests (NEEDED):
  ❌ gitopsi marketplace list
  ❌ gitopsi install prometheus-stack
  ❌ Verify Prometheus running in cluster

Behavior Tests (NEEDED):
  ❌ Given pattern installed, When update available, Then shows notification
```

---

### Bug Fix Issues - Regression Test Matrix

| Issue | Bug | Unit | Regression | E2E | Status |
|-------|-----|------|------------|-----|--------|
| #34 | ArgoCD namespace | ✅ | ❌ NEEDED | ❌ NEEDED | ⚠️ |
| #35 | Pre-flight check | ✅ | ❌ NEEDED | ❌ NEEDED | ⚠️ |
| #36 | ArgoCD detection | ✅ | ❌ NEEDED | ❌ NEEDED | ⚠️ |
| #40 | Missing kustomization | ✅ | ❌ NEEDED | ❌ NEEDED | ⚠️ |
| #41 | Bootstrap auto-apply | ✅ | ❌ NEEDED | ✅ | ⚠️ |

#### Regression Test Requirements

**Bug #34: ArgoCD Namespace Configurable**
```
Regression Test (NEEDED):
  ❌ TestRegression_34_OpenshiftUsesOpenshiftGitops
  
  Given: Platform is OpenShift
  When: gitopsi init generates ArgoCD manifests
  Then: namespace is 'openshift-gitops' NOT 'argocd'
  
E2E Test (NEEDED):
  ❌ Generate on OpenShift, verify namespace in all ArgoCD YAML files
```

**Bug #40: Missing Kustomization Files**
```
Regression Test (NEEDED):
  ❌ TestRegression_40_AllSubdirsHaveKustomization
  
  Given: gitopsi init with infrastructure scope
  When: generates infrastructure/base/
  Then: namespaces/, rbac/, network-policies/, resource-quotas/ all have kustomization.yaml
  
E2E Test (NEEDED):
  ❌ Generate project, run 'kustomize build' on all overlays, verify no errors
```

---

### Test Implementation Priority

| Priority | Test Type | Issues | Effort | Impact |
|----------|-----------|--------|--------|--------|
| 🔴 P0 | Regression Tests | #34, #35, #36, #40, #41 | Low | High |
| 🔴 P0 | Integration Tests | Core flows (#4, #16, #21) | Medium | High |
| 🟡 P1 | E2E Tests | Phase 2 features (#15, #17, #18) | High | High |
| 🟡 P1 | Behavior Tests | User-facing features | Medium | Medium |
| 🟢 P2 | Stability Tests | All | High | Medium |
| 🟢 P2 | Performance Tests | All | Medium | Low |

---

### E2E Test Requirements (MANDATORY)

All code changes MUST pass E2E tests on OpenShift before merge:

```bash
# Run E2E tests
export OCP_API="https://api.cluster.example.com:6443"
export OCP_USER="admin"
export OCP_PASSWORD="password"
./scripts/e2e-openshift-full.sh
```

### E2E Test Coverage

| Test | Description | Validates |
|------|-------------|-----------|
| **GitOps Detection** | Detect installed ArgoCD | Type (community/RedHat), method (Helm/OLM), status |
| **Manifest Generation** | Generate and validate manifests | Server-side validation against cluster API |
| **Infrastructure Deploy** | Apply namespaces, RBAC | Resources created successfully |
| **Bootstrap Validation** | Verify ArgoCD installation | Components running, health status |
| **Cleanup** | Remove all test resources | Cluster left clean for next run |

### ArgoCD Detection Tests

The E2E tests validate ArgoCD installation details:

```go
// Detection result includes:
type ArgoCDDetectionResult struct {
    Installed        bool            // Is ArgoCD installed?
    Type             ArgoCDType      // community | redhat | unknown
    InstallMethod    InstallMethod   // helm | olm | manifest | operator
    OperatorSource   OperatorSource  // redhat-operators | community-operators
    Namespace        string          // openshift-gitops | argocd
    Version          string          // v2.9.0
    Running          bool            // All components healthy?
    Components       []ArgoCDComponent
    Issues           []string        // Problems detected
    Recommendations  []string        // Suggested fixes
}
```

### Test Output Storage

All E2E test output is saved for review:

```
test-output/
└── YYYYMMDD_HHMMSS/
    ├── test.log              # Full execution log
    ├── results.csv           # Test pass/fail summary
    ├── summary.txt           # Human-readable summary
    ├── issues-to-create.md   # GitHub issues to create
    ├── generated/            # Generated GitOps files
    ├── cluster-state/        # Cluster snapshots
    └── validation/           # Manifest validation results
```

### Automatic Issue Creation

E2E tests automatically detect and log issues:

1. **Invalid manifests** → Creates bug issue
2. **Failed validations** → Creates enhancement issue
3. **Missing features** → Creates feature request

### Local Testing Before CI/CD

> ⚠️ **ALWAYS run tests locally before pushing**

```bash
# 1. Run unit tests
make test

# 2. Run linters
make lint

# 3. Run E2E tests (if cluster available)
./scripts/e2e-openshift-full.sh

# 4. Only push if all pass
git push
```

### Post-E2E Analysis (MANDATORY RULE)

> ⚠️ **After EVERY E2E test run, you MUST perform post-analysis**

#### Post-E2E Checklist

After running E2E tests, ALWAYS:

1. **Read and analyze test output**
   ```bash
   # Review summary
   cat test-output/TIMESTAMP/summary.txt
   
   # Review issues found
   cat test-output/TIMESTAMP/issues-to-create.md
   
   # Review validation results
   cat test-output/TIMESTAMP/validation/manifest-validation.log
   cat test-output/TIMESTAMP/cluster-state/argocd-detection.txt
   ```

2. **Create GitHub issues for findings**
   - ❌ Failed tests → Create bug issue
   - ⚠️ Warnings → Create enhancement issue
   - 💡 Improvements identified → Create feature request

3. **Update PRD to track issues**
   - Add new issues to Issue Summary section
   - Update roadmap if needed
   - Document any blockers

4. **Clean up test output**
   ```bash
   # Keep latest for reference, delete old runs
   ls -la test-output/
   rm -rf test-output/OLD_TIMESTAMP/
   ```

#### E2E Finding Categories

| Finding | Action | Priority |
|---------|--------|----------|
| Manifest validation FAIL | Create bug issue | 🔴 High |
| Bootstrap validation FAIL | Investigate cluster state | 🟡 Medium |
| ArgoCD detection issues | Create enhancement issue | 🟢 Low |
| Missing feature detected | Create feature request | 🟢 Low |

#### Example Post-E2E Analysis

From E2E run `20251217_184029`:

| Finding | Issue Created |
|---------|---------------|
| ArgoCD namespace mismatch | [#34](https://github.com/ihsanmokhlisse/gitopsi/issues/34) |
| Pre-flight check needed | [#35](https://github.com/ihsanmokhlisse/gitopsi/issues/35) |
| Better ArgoCD detection | [#36](https://github.com/ihsanmokhlisse/gitopsi/issues/36) |

---

## Development Workflow

### GitFlow Branching Strategy

```
main        ← Production releases (tagged v*.*.*)
  │
develop     ← Integration branch (all PRs target here)
  │
feature/*   ← New features (from GitHub Issues)
bugfix/*    ← Bug fixes (from GitHub Issues)
release/*   ← Release preparation
hotfix/*    ← Emergency production fixes
```

### Branch Naming Convention

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feature/issue-{id}-{short-desc}` | `feature/issue-12-add-flux-support` |
| Bug Fix | `bugfix/issue-{id}-{short-desc}` | `bugfix/issue-15-fix-template` |
| Release | `release/v{major}.{minor}.{patch}` | `release/v0.1.0` |
| Hotfix | `hotfix/v{version}-{desc}` | `hotfix/v0.1.1-critical` |

### Commit Convention

```
type(scope): description

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- refactor: Code refactoring
- test: Adding tests
- chore: Maintenance

Examples:
feat(cli): add interactive prompts for init
fix(generator): resolve template path issue
docs(readme): update installation guide
```

### Version Numbering

Semantic Versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Development Rules (MANDATORY)

> ⚠️ **These rules must be followed strictly for all development work**

#### 1. One Issue = One Branch
```
✅ Each GitHub issue MUST have its own dedicated branch
✅ Branch naming: feature/issue-{id}-{short-desc}
✅ No mixing of multiple issues in one branch
```

#### 2. Sequential Issue Completion
```
🚫 DO NOT start a new issue until the current one is:
   - ✅ Fully implemented
   - ✅ All tests passing (unit + E2E)
   - ✅ PR reviewed and merged
   - ✅ Issue closed
   - ✅ Branch deleted

⚠️ If a bug or new issue is discovered during development:
   - 📝 Document it in a new GitHub issue
   - 🚫 DO NOT work on it immediately
   - ⏳ Wait for permission to work on new issues
   - ✅ Complete and close current issue first
   - 🔄 Then proceed to next issue with approval
```

#### 3. Test Requirements (Non-Negotiable)
```
Every feature/fix MUST include ALL applicable test types:

📋 Unit Tests (MANDATORY):
   - All new functions/methods tested
   - Edge cases covered
   - Error scenarios tested
   - Coverage must not decrease
   - File: internal/<package>/<package>_test.go

🔗 Integration Tests (MANDATORY for core features):
   - Multi-package flow tested
   - Config → Generator → Output flow verified
   - CLI → Cluster → Bootstrap flow verified
   - File: internal/<package>/integration_test.go

🔄 End-to-End Tests (MANDATORY):
   - Complete workflow tested on real cluster
   - Cypress UI tests for ArgoCD features
   - Shell scripts for cluster operations
   - Files: e2e-tests/cypress/, scripts/e2e-*.sh

📖 Behavior/BDD Tests (MANDATORY for user-facing features):
   - Given/When/Then scenarios documented
   - User story acceptance criteria verified
   - File: internal/<package>/behavior_test.go

🔒 Regression Tests (MANDATORY for bug fixes):
   - Specific test that would have caught the bug
   - Test name: TestRegression_<IssueNumber>_<Description>
   - File: internal/<package>/regression_test.go

✅ All tests must pass before:
   - Creating a PR
   - Requesting review
   - Merging to develop
```

#### Test Acceptance Criteria Per Issue Type

**New Feature:**
```
□ Unit tests for all new functions (>80% coverage)
□ Integration test for feature flow
□ E2E test on Kind/OpenShift cluster
□ Behavior test with Given/When/Then
□ Documentation updated
```

**Bug Fix:**
```
□ Unit test that reproduces the bug
□ Regression test: TestRegression_<IssueNumber>_<Description>
□ E2E test to verify fix on cluster
□ Root cause documented in issue
```

**Enhancement:**
```
□ Unit tests for modified functions
□ Integration test if affects multiple packages
□ E2E test if user-visible change
□ Performance benchmark if applicable
```

#### 4. Quality Checklist (Before PR)
```
□ All unit tests pass
□ All E2E tests pass
□ No linting errors
□ Documentation updated
□ CHANGELOG updated (if user-facing)
□ Commit messages follow convention
□ PR description complete
□ Issue linked in PR
```

#### 5. Issue Lifecycle
```
1. 📋 Issue created (TODO)
2. 🔀 Branch created from develop
3. 💻 Development in progress
4. ✅ Tests written and passing
5. 📝 PR created targeting develop
6. 👀 Code review
7. 🔄 Address review comments
8. ✅ Final approval
9. 🔀 Merge to develop
10. 🗑️ Branch deleted
11. ✅ Issue closed
12. 🎯 Next issue can begin
```

---

## Problem Statement

Setting up GitOps repositories is:

1. **Time-consuming** - Manual folder/manifest creation
2. **Error-prone** - Missing files, syntax errors
3. **Inconsistent** - Different structures per team
4. **Platform-ignorant** - No OpenShift/AKS/EKS specifics
5. **Undocumented** - Missing README, runbooks
6. **Disconnected** - Manual git push and cluster bootstrap

---

## Solution

A CLI that:

- Collects requirements **interactively** or via **config file**
- **Generates** complete GitOps repository structure
- Supports **multiple platforms** and **scopes**
- Creates **manifests, docs, scripts, and CI/CD**
- Enables **incremental updates** to existing repos
- **Authenticates** to any Git provider
- **Pushes** code and **bootstraps** clusters automatically
- Provides **live progress** and **setup summaries**
- Offers a **marketplace** of reusable patterns

---

## Target Users

| User | Use Case |
|------|----------|
| Platform Engineers | Bootstrap repos for new clusters |
| DevOps Engineers | Standardize infra/app repos |
| Developers | Quick app deployment setup |
| SREs | Consistent infrastructure management |
| Enterprise Teams | Multi-tenant GitOps organization |

---

## Core Concepts

### Scope

| Scope | Description | Content |
|-------|-------------|---------|
| **infrastructure** | Cluster resources | Namespaces, RBAC, Quotas, Policies |
| **application** | App deployments | Deployments, Services, ConfigMaps |
| **both** | Full stack | Infrastructure + Applications |

### Platform

| Platform | Specifics |
|----------|-----------|
| **kubernetes** | Standard K8s (Ingress, Deployments) |
| **openshift** | Routes, SCCs, DeploymentConfigs |
| **aks** | Azure LB, AAD Pod Identity, Key Vault |
| **eks** | AWS LB Controller, IRSA, Secrets Manager |

### GitOps Tool

| Tool | Resources Generated |
|------|---------------------|
| **argocd** | Application, ApplicationSet, AppProject |
| **flux** | GitRepository, Kustomization, HelmRelease |
| **both** | All of the above |

### Git Providers

| Provider | Auth Methods | Self-Hosted |
|----------|--------------|-------------|
| **GitHub** | SSH, Token, OAuth, GitHub App | ✅ Enterprise |
| **GitLab** | SSH, Token, OAuth, Deploy Token | ✅ |
| **Gitea** | SSH, Token | ✅ |
| **Azure DevOps** | SSH, PAT, OAuth | ✅ Server |
| **Bitbucket** | SSH, App Password, OAuth | ✅ Server |

---

## Functional Requirements

### FR-001: Interactive Mode

#### Phase 1 MVP Interactive Mode (v0.1.0)

```bash
$ gitopsi init

🎯 gitopsi - GitOps Repository Generator

? Project name: my-platform
? Target platform: [kubernetes, openshift, aks, eks]
? Scope: [infrastructure, application, both]
? GitOps tool: [argocd, flux, both]
? Output type: [local, git]
? Environments: [dev, staging, qa, prod]
? Generate documentation? Yes

🚀 Generating GitOps repository: my-platform

📁 Creating directory structure...
🏗️  Generating infrastructure...
📦 Generating applications...
🔄 Generating ArgoCD configuration...
📚 Generating documentation...
🔧 Generating bootstrap...
📜 Generating scripts...

✅ Generated: my-platform/
```

#### Phase 2+ Interactive Mode (Additional Prompts)

```bash
$ gitopsi init

? Project name: my-platform
? Git repository URL: git@github.com:org/gitops-repo.git    # [Phase 2]
? Platform: [Kubernetes, OpenShift, AKS, EKS]
? Scope: [infrastructure, application, both]
? GitOps tool: [argocd, flux, both]
? Environments: [dev, staging, prod]
? Push to repository? Yes                                     # [Phase 2]
? Bootstrap cluster? Yes                                      # [Phase 2]

✅ Generated: my-platform/
✅ Pushed to: git@github.com:org/gitops-repo.git              # [Phase 2]
✅ ArgoCD installed and syncing!                              # [Phase 2]
```

### FR-002: Config File Mode

#### Phase 1 MVP Config (v0.1.0)

```yaml
# gitops.yaml - MVP supported fields
project:
  name: my-platform
  description: "Platform GitOps repository"

output:
  type: local                           # local | git
  url: ""                               # Git URL (if type: git)
  branch: main

platform: kubernetes                    # kubernetes | openshift | aks | eks
scope: both                             # infrastructure | application | both
gitops_tool: argocd                     # argocd | flux | both

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
  - name: frontend
    image: registry.io/frontend:latest
    port: 8080
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

#### Phase 2+ Additional Config Fields

```yaml
# Additional fields for Phase 2+
git:                                    # [Phase 2] Git provider config
  url: git@github.com:org/gitops-repo.git
  branch: main
  auth:
    method: ssh
    ssh_key: ~/.ssh/id_rsa
  push_on_init: true

cluster:                                # [Phase 2] Cluster bootstrap config
  url: https://api.cluster.example.com:6443
  auth:
    method: token
    token_env: CLUSTER_TOKEN

bootstrap:                              # [Phase 2] Bootstrap config
  enabled: true
  tool: argocd
  mode: helm

infrastructure:
  operators:                            # [Phase 2] Operator management
    - name: prometheus-operator
      source: community-operators

patterns:                               # [Phase 3] Marketplace patterns
  - name: prometheus-stack
    config:
      retention: 30d
```

### FR-003: Generated Structure

#### Phase 1 MVP Structure (v0.1.0)

```
my-platform/
├── README.md                           # Project documentation
├── docs/
│   ├── ARCHITECTURE.md                 # Architecture overview
│   └── ONBOARDING.md                   # Getting started guide
├── bootstrap/
│   └── argocd/
│       └── namespace.yaml              # ArgoCD namespace
├── infrastructure/
│   ├── base/
│   │   ├── kustomization.yaml          # Base kustomization
│   │   ├── namespaces/                 # Namespace manifests
│   │   │   ├── dev.yaml
│   │   │   ├── staging.yaml
│   │   │   └── prod.yaml
│   │   ├── rbac/                       # RBAC manifests (if enabled)
│   │   ├── network-policies/           # NetworkPolicy manifests (if enabled)
│   │   └── resource-quotas/            # ResourceQuota manifests (if enabled)
│   └── overlays/
│       ├── dev/
│       │   └── kustomization.yaml
│       ├── staging/
│       │   └── kustomization.yaml
│       └── prod/
│           └── kustomization.yaml
├── applications/
│   ├── base/
│   │   ├── kustomization.yaml
│   │   └── sample-app/                 # Sample application
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       └── kustomization.yaml
│   └── overlays/
│       ├── dev/
│       │   └── kustomization.yaml
│       ├── staging/
│       │   └── kustomization.yaml
│       └── prod/
│           └── kustomization.yaml
├── argocd/
│   ├── projects/
│   │   ├── infrastructure.yaml         # Infrastructure AppProject
│   │   └── applications.yaml           # Applications AppProject
│   └── applicationsets/
│       ├── infra-dev.yaml              # Per-environment ArgoCD Apps
│       ├── infra-staging.yaml
│       ├── infra-prod.yaml
│       ├── apps-dev.yaml
│       ├── apps-staging.yaml
│       └── apps-prod.yaml
└── scripts/
    ├── bootstrap.sh                    # Bootstrap script
    └── validate.sh                     # Validation script
```

#### Phase 2+ Additional Structure

```
my-platform/
├── .gitopsi/
│   └── setup-summary.yaml              # [Phase 2] Saved credentials & URLs
├── bootstrap/
│   └── argocd/
│       └── install.yaml                # [Phase 2] ArgoCD installation
└── infrastructure/
    └── base/
        └── operators/                  # [Phase 2] Operator manifests
```

### FR-004: End-to-End Setup (Issue #21)

```bash
$ gitopsi init --git-url <repo> --cluster <url> --bootstrap

🔐 Authenticating to Git...       ✓
📁 Generating GitOps repository... ✓
📤 Pushing to repository...        ✓
🔐 Authenticating to cluster...    ✓
🚀 Bootstrapping ArgoCD...         ✓
🔗 Configuring repository...       ✓
📦 Creating root application...    ✓

✅ GitOps setup complete!
   ArgoCD UI: https://argocd.apps.cluster.com
   Username: admin
   Password: xK9mP2vL8nQ4rT6w
```

### FR-005: Live Progress Display (Issue #22)

```
┌─ ArgoCD Bootstrap ───────────────────────────────────────────┐
│ ● Creating argocd namespace...                   [✓] 0.2s    │
│ ● Installing ArgoCD via Helm...                  [⠋] 23s     │
│   ├── argocd-server                              [✓] Ready   │
│   ├── argocd-repo-server                         [⠋] 1/1     │
│   ├── argocd-application-controller              [✓] Ready   │
│   └── argocd-redis                               [✓] Ready   │
└──────────────────────────────────────────────────────────────┘
```

### FR-006: Pattern Marketplace (Issue #23)

```bash
$ gitopsi marketplace search monitoring

📦 GitOps Patterns - Monitoring

┌────────────────────────────────────────────────────────────────┐
│ prometheus-stack                              ⭐ 4.8 (234)     │
│ Complete Prometheus + Grafana monitoring stack                 │
│ Install: gitopsi install prometheus-stack                      │
└────────────────────────────────────────────────────────────────┘

$ gitopsi install prometheus-stack
```

### FR-007: Add Commands

```bash
# Add application
$ gitopsi add app --name api-gateway --image registry.io/api-gw --port 8080

# Add environment
$ gitopsi add env --name qa --cluster https://qa.k8s.local

# Add infrastructure component
$ gitopsi add infra --type network-policy --name deny-all

# Add operator
$ gitopsi add operator prometheus-operator --source community-operators
```

### FR-008: Organization Management (Issue #19)

```bash
# Initialize organization
$ gitopsi org init acme-corp

# Onboard teams
$ gitopsi team onboard frontend --quota-cpu 50 --quota-memory 100Gi

# Generate multi-tenant structure
acme-corp/
├── platform/          # Platform team manages
├── teams/
│   ├── frontend/      # Team-specific
│   └── backend/
└── shared/            # Shared services
```

---

## CLI Interface

```bash
gitopsi [command] [subcommand] [flags]

Commands:
  init              Initialize new GitOps repository
  add               Add resources to existing repo
    app             Add application
    env             Add environment
    infra           Add infrastructure component
    operator        Add Kubernetes operator
  validate          Validate repository structure
  status            Show sync status
  
  # Git Provider Integration
  connect           Connect to Git repository
  test git          Test Git connection
  test cluster      Test cluster connection
  
  # Bootstrap & Management
  bootstrap         Bootstrap GitOps tool on cluster
  info              Show setup information
    argocd          ArgoCD URL & credentials
    git             Git repository info
    cluster         Cluster info
  get-password      Get service password
  open              Open in browser
    argocd          Open ArgoCD UI
    git             Open Git repository
  
  # Marketplace
  marketplace       Browse pattern marketplace
    list            List all patterns
    search          Search patterns
    info            Pattern details
  install           Install pattern
  patterns          Manage installed patterns
    list            List installed
    update          Update pattern
    remove          Remove pattern
  
  # Organization (Phase 3)
  org               Organization management
    init            Initialize organization
    status          Organization status
  team              Team management
    onboard         Onboard new team
    list            List teams
  
  # Utilities
  export            Export configuration
  version           Show version

Global Flags:
  --config string   Config file path
  --output string   Output directory (default: .)
  --dry-run         Preview without writing
  --verbose         Verbose output
  --quiet           Minimal output
  --json            JSON output format
```

---

## Architecture

### Components

```
┌────────────────────────────────────────────────────────────────┐
│                        gitopsi CLI                              │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐  ┌─────────┐  ┌───────────┐  ┌──────────────┐    │
│  │  Cobra  │  │  Viper  │  │  Survey   │  │   Progress   │    │
│  │  (CLI)  │  │(Config) │  │ (Prompts) │  │   Display    │    │
│  └────┬────┘  └────┬────┘  └─────┬─────┘  └──────┬───────┘    │
│       └────────────┴─────────────┴───────────────┘             │
│                           │                                     │
│                    ┌──────▼──────┐                              │
│                    │   Engine    │                              │
│                    └──────┬──────┘                              │
│                           │                                     │
│    ┌──────────────────────┼──────────────────────┐             │
│    │                      │                      │             │
│    ▼                      ▼                      ▼             │
│ ┌──────────┐       ┌──────────────┐       ┌───────────┐       │
│ │Generator │       │ Git Provider │       │  Cluster  │       │
│ └────┬─────┘       │   Adapter    │       │  Manager  │       │
│      │             └──────┬───────┘       └─────┬─────┘       │
│      │                    │                     │              │
│ ┌────▼────┐         ┌─────▼─────┐         ┌────▼────┐        │
│ │Templates│         │  GitHub   │         │ ArgoCD  │        │
│ │(embed)  │         │  GitLab   │         │  Flux   │        │
│ └─────────┘         │  Gitea    │         │Bootstrap│        │
│                     │Azure/Bitb │         └─────────┘        │
│                     └───────────┘                             │
│                                                                │
│ ┌──────────────────────────────────────────────────────────┐  │
│ │                    Pattern Marketplace                    │  │
│ │  ┌─────────┐  ┌──────────┐  ┌─────────┐  ┌───────────┐  │  │
│ │  │Official │  │Community │  │Private  │  │  Local    │  │  │
│ │  │Patterns │  │Patterns  │  │Registry │  │ Patterns  │  │  │
│ │  └─────────┘  └──────────┘  └─────────┘  └───────────┘  │  │
│ └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
gitopsi/
├── cmd/gitopsi/
│   └── main.go
├── internal/
│   ├── cli/                    # CLI commands
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── add.go
│   │   ├── validate.go
│   │   ├── marketplace.go
│   │   ├── org.go
│   │   └── team.go
│   ├── config/                 # Configuration
│   │   ├── config.go
│   │   ├── loader.go
│   │   └── validate.go
│   ├── generator/              # Generation engine
│   │   ├── generator.go
│   │   ├── infrastructure.go
│   │   ├── applications.go
│   │   ├── argocd.go
│   │   ├── flux.go
│   │   └── docs.go
│   ├── git/                    # Git provider adapters
│   │   ├── provider.go         # Interface
│   │   ├── github.go
│   │   ├── gitlab.go
│   │   ├── gitea.go
│   │   ├── azure.go
│   │   └── bitbucket.go
│   ├── cluster/                # Cluster management
│   │   ├── client.go
│   │   ├── bootstrap.go
│   │   └── validate.go
│   ├── marketplace/            # Pattern marketplace
│   │   ├── registry.go
│   │   ├── pattern.go
│   │   └── installer.go
│   ├── platform/               # Platform specifics
│   │   ├── kubernetes.go
│   │   ├── openshift.go
│   │   ├── aks.go
│   │   └── eks.go
│   ├── prompt/                 # Interactive prompts
│   │   └── prompt.go
│   ├── progress/               # Progress display
│   │   └── display.go
│   ├── templates/              # Embedded templates
│   │   └── files/
│   └── output/                 # File writing
│       └── writer.go
├── templates/                  # Source templates
├── testdata/
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yaml
└── README.md
```

---

## Technology Stack

| Component | Choice | Reason |
|-----------|--------|--------|
| Language | Go 1.22+ | Single binary, cross-platform |
| CLI | Cobra | Industry standard |
| Config | Viper | YAML/JSON/ENV support |
| Prompts | Survey v2 | Rich interactive prompts |
| Templates | text/template | Go native, embed.FS |
| YAML | gopkg.in/yaml.v3 | Full YAML support |
| Git | go-git | Pure Go Git implementation |
| K8s | client-go | Official Kubernetes client |
| Testing | testify | Assertions, mocks |
| Build | GoReleaser | Multi-platform releases |
| Progress | pterm/spinner | Beautiful terminal output |

---

## Roadmap

### Phase 1: MVP (v0.1.0) ✅ COMPLETED

**Goal:** Basic repository generation for Kubernetes + ArgoCD
**Status:** ✅ Completed December 2025

| Issue | Title | Status |
|-------|-------|--------|
| #1 | CLI structure with Cobra | ✅ Done |
| #2 | Config file parsing with Viper | ✅ Done |
| #3 | Interactive prompts with Survey | ✅ Done |
| #4 | Init command implementation | ✅ Done |
| #5 | Kubernetes manifest generation | ✅ Done |
| #6 | Kustomize base/overlay structure | ✅ Done |
| #7 | ArgoCD Application generation | ✅ Done |
| #8 | Template embedding with embed.FS | ✅ Done |
| #9 | File output system | ✅ Done |
| #10 | README.md generation | ✅ Done |
| #11 | Unit tests (>80% coverage) | ✅ Done |

**Deliverable:** `gitopsi init` generates working K8s + ArgoCD repo

---

### Phase 2: End-to-End & Platforms (v0.2.0)

**Goal:** Complete end-to-end workflow from init to synced cluster
**Target:** 4 weeks after Phase 1

| Issue | Title | Priority | Status |
|-------|-------|----------|--------|
| [#24](https://github.com/ihsanmokhlisse/gitopsi/issues/24) | Multi-provider Git support (GitHub, GitLab, Gitea, Azure, Bitbucket) | 🔴 High | 🔲 TODO |
| [#21](https://github.com/ihsanmokhlisse/gitopsi/issues/21) | End-to-end setup with auto-push and cluster bootstrap | 🔴 High | 🔲 TODO |
| [#22](https://github.com/ihsanmokhlisse/gitopsi/issues/22) | Live progress display with validation and setup summary | 🔴 High | 🔲 TODO |
| [#16](https://github.com/ihsanmokhlisse/gitopsi/issues/16) | Multiple bootstrap modes (Helm, OLM, Manifest) | 🔴 High | 🔲 TODO |
| [#17](https://github.com/ihsanmokhlisse/gitopsi/issues/17) | Flexible environment and cluster management | 🔴 High | 🔲 TODO |
| [#15](https://github.com/ihsanmokhlisse/gitopsi/issues/15) | Kubernetes Operator management | 🟡 Medium | 🔲 TODO |
| [#18](https://github.com/ihsanmokhlisse/gitopsi/issues/18) | Customizable project generation from config | 🟡 Medium | 🔲 TODO |
| [#28](https://github.com/ihsanmokhlisse/gitopsi/issues/28) | Version-aware manifest generation | 🔴 High | 🔲 TODO |
| [#29](https://github.com/ihsanmokhlisse/gitopsi/issues/29) | Security scanning and provenance | 🔴 High | 🔲 TODO |
| [#30](https://github.com/ihsanmokhlisse/gitopsi/issues/30) | Validate command for manifest validation | 🔴 High | 🔲 TODO |

#### Phase 2 Features

**Version-Aware Manifest Generation (#28)**

Ensures generated manifests are compatible with target platform versions:

```yaml
# gitops.yaml - Version configuration
platform:
  type: kubernetes
  version: "1.29"          # Target K8s version
gitops_tool:
  type: argocd
  version: "2.10"          # Target ArgoCD version
openshift:
  version: "4.14"          # If using OpenShift
```

Features:
- API version mapping per Kubernetes release
- Deprecated API detection with `pluto`
- Schema validation with `kubeconform`
- Version auto-detection from cluster (optional)

**Security Scanning & Provenance (#29)**

Ensures generated manifests are secure and verifiable:

```bash
# Scan generated manifests
gitopsi init --scan

# Verify provenance
gitopsi verify ./my-platform/
```

Security Features:
| Tool | Purpose |
|------|---------|
| `checkov` | IaC security scanner |
| `trivy config` | Vulnerability scanning |
| `kube-score` | Security best practices |
| `kubesec` | Risk analysis |
| `cosign` | Manifest signing |

Provenance:
- Generation metadata in files
- SLSA attestations
- Cosign signatures
- Input sanitization (code injection prevention)

**Validate Command (#30)**

Comprehensive manifest validation:

```bash
gitopsi validate ./my-platform/ --all
gitopsi validate ./my-platform/ --security --fail-on high
gitopsi validate ./my-platform/ --k8s-version 1.29
```

Output:
```
🔍 Validating: ./my-platform/

📋 Schema Validation
  ✅ 45 manifests validated against Kubernetes 1.29

🔒 Security Scan  
  ⚠️  3 medium issues found
    
⚠️  Deprecation Check
  ⚠️  1 deprecated API found

📊 Summary: 41 passed, 4 warnings, 0 failed
```

**Multi-Provider Git Support (#24)**
- Auto-detect provider from URL
- Support SSH, Token, OAuth authentication
- Provider-specific CI/CD generation
- Repository creation if doesn't exist
- Webhook configuration for GitOps sync

**End-to-End Setup (#21)**
```bash
gitopsi init --git-url <url> --cluster <url> --bootstrap
# One command: Generate → Push → Bootstrap → Sync
```

**Live Progress (#22)**
- Real-time step-by-step status
- Pod/resource status during bootstrap
- Validation after setup
- Summary with all credentials and URLs
- Error recovery suggestions

**Bootstrap Modes (#16)**
| Mode | Description | Use Case |
|------|-------------|----------|
| Helm | Helm chart installation | Standard Kubernetes |
| OLM | Operator Lifecycle Manager | OpenShift |
| Manifest | Raw YAML manifests | Air-gapped environments |

**Environment Management (#17)**
| Pattern | Description |
|---------|-------------|
| Single cluster, multi-namespace | Environments as namespaces |
| Multi-cluster | One cluster per environment |
| Hybrid | Mix of both approaches |

---

### Phase 3: Enterprise & Marketplace (v0.3.0)

**Goal:** Enterprise features and community marketplace
**Target:** 4 weeks after Phase 2

| Issue | Title | Priority | Status |
|-------|-------|----------|--------|
| [#14](https://github.com/ihsanmokhlisse/gitopsi/issues/14) | Authentication management (credentials store) | 🔴 High | 🔲 TODO |
| [#23](https://github.com/ihsanmokhlisse/gitopsi/issues/23) | GitOps Pattern Marketplace | 🔴 High | 🔲 TODO |
| [#13](https://github.com/ihsanmokhlisse/gitopsi/issues/13) | Multi-repository support | 🟡 Medium | 🔲 TODO |
| [#19](https://github.com/ihsanmokhlisse/gitopsi/issues/19) | Organization and multi-tenancy management | 🟡 Medium | 🔲 TODO |

#### Phase 3 Features

**Pattern Marketplace (#23)**
```bash
gitopsi marketplace search monitoring
gitopsi install prometheus-stack
```

Categories:
- 🏗️ Infrastructure (networking, security, storage)
- 📊 Observability (monitoring, logging, tracing)
- 🔐 Security (secrets, policies, scanning)
- 🌐 Networking (ingress, service mesh)
- 💾 Data (databases, caching, messaging)
- 🚀 CI/CD (pipelines, gitops addons)

**Multi-Repository Support (#13)**
- Applications from different repos
- Clusters managed separately
- Components with independent lifecycles

**Organization Management (#19)**
```bash
gitopsi org init acme-corp --config enterprise.yaml
gitopsi team onboard data-science --quota-cpu 50
```

---

### Phase 4: Polish & Release (v1.0.0)

**Goal:** Production-ready release
**Target:** 2 weeks after Phase 3

| Feature | Description |
|---------|-------------|
| OpenShift support | Routes, SCCs, DeploymentConfigs |
| AKS support | Azure-specific annotations |
| EKS support | AWS-specific annotations |
| Flux support | GitRepository, Kustomization |
| Validation command | `gitopsi validate` |
| Integration tests | Full workflow tests |
| Homebrew formula | `brew install gitopsi` |
| Container image | `ghcr.io/ihsanmokhlisse/gitopsi` |
| Documentation site | docs.gitopsi.io |

---

## Issue Summary

### All Closed Issues with Test Status

| Phase | Issue | Title | Code | Unit | Integ | E2E | Regress | Complete |
|-------|-------|-------|------|------|-------|-----|---------|----------|
| **P1** | #1 | CLI structure | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #2 | Config parsing | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #3 | Interactive prompts | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P1** | #4 | Init command | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #5 | K8s manifests | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #6 | Kustomize structure | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #7 | ArgoCD generation | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #8 | Template embedding | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #9 | File output | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #10 | README generation | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P1** | #11 | Unit tests | ✅ | ✅ | N/A | N/A | N/A | ✅ 100% |
| **P2** | #15 | Operator management | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #16 | Bootstrap modes | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P2** | #17 | Environment management | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #18 | Customizable generation | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #21 | E2E GitOps setup | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P2** | #22 | Live progress | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #24 | Multi-provider Git | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #28 | Version-aware manifests | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #29 | Security scanning | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #30 | Validate command | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #34 | ArgoCD namespace | ✅ | ✅ | ❌ | ❌ | ❌ | ⚠️ 30% |
| **P2** | #35 | Pre-flight check | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #36 | ArgoCD detection | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P2** | #40 | Missing kustomization | ✅ | ✅ | ❌ | ❌ | ❌ | ⚠️ 30% |
| **P2** | #41 | Bootstrap auto-apply | ✅ | ✅ | ❌ | ✅ | N/A | ⚠️ 60% |
| **P3** | #13 | Multi-repository | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P3** | #14 | Auth management | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P3** | #19 | Organization mgmt | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |
| **P3** | #20 | Philosophy docs | ✅ | N/A | N/A | N/A | N/A | ✅ 100% |
| **P3** | #23 | Marketplace | ✅ | ✅ | ❌ | ❌ | N/A | ⚠️ 40% |

### Progress Tracking

```
FEATURE IMPLEMENTATION:
Phase 1: [██████████] 11/11 (100%) ✅ COMPLETE
Phase 2: [██████████] 13/13 (100%) ✅ COMPLETE
Phase 3: [██████████]  5/5  (100%) ✅ COMPLETE
────────────────────────────────────
Feature:  [██████████] 29/29 (100%) ✅

TESTING COVERAGE:
Unit Tests:        [██████████] 30/30 (100%) ✅
Integration Tests: [██░░░░░░░░]  3/30 (10%)  ❌ CRITICAL GAP
E2E Tests:         [█████░░░░░] 15/30 (50%)  ⚠️ 
Behavior Tests:    [░░░░░░░░░░]  0/30 (0%)   ❌ NOT STARTED
Regression Tests:  [░░░░░░░░░░]  0/5  (0%)   ❌ CRITICAL GAP
Stability Tests:   [░░░░░░░░░░]  0/30 (0%)   ❌ NOT STARTED
────────────────────────────────────
Overall Test:      [███░░░░░░░] 48/155 (31%) ⚠️ NEEDS WORK
```

### Test Gap Analysis

| Test Type | Required | Implemented | Gap | Priority |
|-----------|----------|-------------|-----|----------|
| Unit Tests | 30 | 30 | 0 | ✅ |
| Integration Tests | 30 | 3 | 27 | 🔴 P0 |
| E2E Tests | 30 | 15 | 15 | 🟡 P1 |
| Behavior Tests | 30 | 0 | 30 | 🟡 P1 |
| Regression Tests | 5 | 0 | 5 | 🔴 P0 |
| Stability Tests | 30 | 0 | 30 | 🟢 P2 |
| Performance Tests | 30 | 0 | 30 | 🟢 P2 |

### Immediate Testing Priorities

| Priority | Action | Issues | Effort |
|----------|--------|--------|--------|
| 🔴 P0 | Add regression tests | #34, #40 | 1 day |
| 🔴 P0 | Add integration tests for core flows | #4, #16, #21 | 2 days |
| 🟡 P1 | Add E2E for Phase 2/3 features | #15, #17, #18, #23 | 3 days |
| 🟡 P1 | Add behavior tests | All user-facing | 3 days |
| 🟢 P2 | Add stability tests | All | 2 days |

### E2E Test Findings Tracked

| E2E Run | Issues Created |
|---------|----------------|
| `20251217_184029` | #34, #35, #36 |

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Time to first deploy | < 5 minutes |
| Learning curve | Productive in 30 minutes |
| Config lines | 80% reduction vs manual |
| Test coverage | > 80% |
| Supported platforms | 4 (K8s, OCP, AKS, EKS) |
| Git providers | 5 (GitHub, GitLab, Gitea, Azure, Bitbucket) |
| Pattern marketplace | 20+ patterns at launch |

---

## Out of Scope (v1.0)

- Secret management (use external tools like Vault)
- CI/CD execution (generate configs only)
- GUI/Web interface
- Cloud provider authentication (use native CLIs)
- Cluster provisioning (use Terraform, eksctl, etc.)

---

## Quick Links

- **Repository:** https://github.com/ihsanmokhlisse/gitopsi
- **Issues:** https://github.com/ihsanmokhlisse/gitopsi/issues
- **Milestones:** https://github.com/ihsanmokhlisse/gitopsi/milestones

---

## Test Commands Quick Reference

```bash
# ═══════════════════════════════════════════════════════════════════════
# UNIT TESTS
# ═══════════════════════════════════════════════════════════════════════
# Run all unit tests
go test ./internal/... -v

# Run with coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./internal/generator/... -v

# Run specific test
go test ./internal/generator/... -run TestGenerateArgoCD -v

# ═══════════════════════════════════════════════════════════════════════
# INTEGRATION TESTS
# ═══════════════════════════════════════════════════════════════════════
# Run integration tests (tagged)
go test ./internal/... -tags=integration -v

# ═══════════════════════════════════════════════════════════════════════
# E2E TESTS - CLUSTER
# ═══════════════════════════════════════════════════════════════════════
# Kind cluster (local)
./scripts/e2e-comprehensive-test.sh

# OpenShift cluster
export OCP_API="https://api.cluster.example.com:6443"
export OCP_USER="admin"
export OCP_PASSWORD="password"
./scripts/e2e-openshift-full.sh

# ═══════════════════════════════════════════════════════════════════════
# E2E TESTS - CYPRESS UI
# ═══════════════════════════════════════════════════════════════════════
# Run Cypress tests
cd e2e-tests && npm test

# Run specific Cypress spec
cd e2e-tests && npx cypress run --spec "cypress/e2e/01-argocd-ui.cy.js"

# Open Cypress interactive mode
cd e2e-tests && npx cypress open

# ═══════════════════════════════════════════════════════════════════════
# BEHAVIOR/BDD TESTS (Ginkgo)
# ═══════════════════════════════════════════════════════════════════════
# Install Ginkgo
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run BDD tests
ginkgo ./internal/...

# ═══════════════════════════════════════════════════════════════════════
# REGRESSION TESTS
# ═══════════════════════════════════════════════════════════════════════
# Run regression tests only
go test ./internal/... -run "TestRegression" -v

# ═══════════════════════════════════════════════════════════════════════
# BENCHMARKS & STABILITY
# ═══════════════════════════════════════════════════════════════════════
# Run benchmarks
go test ./internal/... -bench=. -benchmem

# Memory profiling
go test ./internal/generator/... -memprofile=mem.out
go tool pprof mem.out

# ═══════════════════════════════════════════════════════════════════════
# LINTING
# ═══════════════════════════════════════════════════════════════════════
make lint
# or
golangci-lint run ./...

# ═══════════════════════════════════════════════════════════════════════
# FULL VALIDATION (Before PR)
# ═══════════════════════════════════════════════════════════════════════
make test lint
./scripts/e2e-comprehensive-test.sh  # If cluster available
```

---

*Version: 4.0*
*Date: December 2025*
*Status: All Phases Complete - Testing Strategy Upgrade*
*Last Updated: Comprehensive testing requirements added for all issues*
*Test Coverage: Unit 100% | Integration 10% | E2E 50% | Behavior 0% | Regression 0%*
