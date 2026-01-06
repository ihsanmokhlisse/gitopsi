# Bug and Error Tracking

This document tracks bugs, errors, and issues encountered during development. Always update this file when:
- A new bug is discovered
- An error occurs during development or testing
- A fix is applied

## Format

Each entry should include:
- **ID**: Sequential identifier (BUG-001, BUG-002, etc.)
- **Date**: When discovered
- **Status**: `open` | `in-progress` | `fixed` | `wont-fix`
- **Severity**: `critical` | `high` | `medium` | `low`
- **Description**: What the bug/error is
- **Steps to Reproduce**: How to trigger it
- **Root Cause**: Why it happens (if known)
- **Fix**: What was done to resolve it
- **Related Files**: Which files were affected

---

## Active Bugs

<!-- No active bugs -->

## Fixed Bugs

### BUG-001

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Status | fixed |
| Severity | high |
| Related Files | `internal/cli/marketplace.go` |

**Description:**
The `patternValidateCmd` was registered to `rootCmd`, overriding the main `validateCmd` from `validate.go`. Running `gitopsi validate <path>` would trigger the marketplace pattern validation instead of the manifest validation.

**Steps to Reproduce:**
1. Run `gitopsi init --config fixtures/minimal-config.yaml --output /tmp/test`
2. Run `gitopsi validate /tmp/test/test-minimal`
3. Error: "pattern.yaml not found" instead of validating K8s manifests

**Root Cause:**
In `internal/cli/marketplace.go:945`, `patternValidateCmd` was added to `rootCmd.AddCommand()` instead of `marketplaceCmd.AddCommand()`. Since both commands have `Use: "validate [path]"`, Cobra used the last-registered one.

**Fix Applied:**
Changed line 945 from `rootCmd.AddCommand(patternValidateCmd)` to `marketplaceCmd.AddCommand(patternValidateCmd)`. Now pattern validation is accessed via `gitopsi marketplace validate` and manifest validation via `gitopsi validate`.

---

### BUG-002

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Status | fixed |
| Severity | medium |
| Related Files | `test/e2e/fixtures/*.yaml` |

**Description:**
E2E tests failed because test fixtures lacked required `git.url` field.

**Steps to Reproduce:**
1. Run `go test -tags=e2e ./test/e2e/...`
2. Error: "git.url is required to generate ArgoCD applications"

**Root Cause:**
The `git.url` field became required for ArgoCD generation (since ArgoCD needs to sync from a Git repository), but the E2E test fixtures didn't include this field.

**Fix Applied:**
Added `git.url` field to all E2E test fixtures:
- `test/e2e/fixtures/minimal-config.yaml`
- `test/e2e/fixtures/standard-config.yaml`
- `test/e2e/fixtures/enterprise-config.yaml`

---

### BUG-003

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Status | fixed |
| Severity | low |
| Related Files | `internal/generator/generator_test.go`, `internal/integration/integration_test.go` |

**Description:**
Tests for Flux functionality failed because Flux support was disabled (focus on ArgoCD first).

**Steps to Reproduce:**
1. Run `go test -race ./...`
2. Tests `TestGenerateFluxTool`, `TestGenerateGitOpsForAllTools/flux`, `TestIntegration_GitOpsToolSelection/flux` fail

**Root Cause:**
Flux directory creation and generation were intentionally disabled to focus on ArgoCD. Tests still expected Flux directories to exist.

**Fix Applied:**
Added `t.Skip("Flux support is disabled - focusing on ArgoCD first")` to:
- `internal/generator/generator_test.go:TestGenerateFluxTool`
- `internal/generator/generator_test.go:TestGenerateGitOpsForAllTools` (for flux case)
- `internal/integration/integration_test.go:TestIntegration_GitOpsToolSelection` (for flux case)
- `test/e2e/init_test.go:TestInitWithFlux`

---

### BUG-006

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Status | fixed |
| Severity | critical |
| Related Files | `internal/bootstrap/bootstrap.go` |

**Description:**
Child ArgoCD Applications fail to sync with error "app is not allowed in project 'applications', or the project does not exist".

**Steps to Reproduce:**
1. Run `gitopsi init --bootstrap --push`
2. ArgoCD root app syncs successfully
3. Child apps (infra-dev, apps-dev, etc.) show "Unknown" status
4. Error: project doesn't exist

**Root Cause:**
The bootstrap process creates the root App-of-Apps, which syncs and creates child Applications. These child Applications reference projects named "infrastructure" and "applications". However, these AppProjects only exist in the Git repo and haven't been synced yet - creating a chicken-and-egg problem.

**Fix Applied:**
Added `createArgoCDProjects()` function in bootstrap.go that creates the "infrastructure" and "applications" AppProjects directly on the cluster BEFORE creating the App-of-Apps. This ensures the projects exist when child applications reference them.

---

### BUG-005

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Status | fixed |
| Severity | critical |
| Related Files | `internal/cli/init.go` |

**Description:**
Bootstrap never runs when `--bootstrap` flag is used without `--cluster` flag, even though auto-detection should work.

**Steps to Reproduce:**
1. Run `gitopsi init --bootstrap` (without --cluster)
2. ArgoCD is never deployed
3. No error message shown

**Root Cause:**
The `shouldBootstrap()` function at line 581 required BOTH `cfg.Bootstrap.Enabled` AND `cfg.Cluster.URL != ""`. But cluster URL auto-detection only happens INSIDE the `if shouldBootstrap()` block. This chicken-and-egg problem meant that if you didn't explicitly provide `--cluster`, bootstrap would never run.

**Fix Applied:**
Changed `shouldBootstrap()` to only check `cfg.Bootstrap.Enabled`:
```go
// Before: return cfg.Bootstrap.Enabled && cfg.Cluster.URL != ""
// After:  return cfg.Bootstrap.Enabled
```
The cluster URL auto-detection inside the bootstrap block handles the empty URL case.

---

### BUG-004

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Status | fixed |
| Severity | high |
| Related Files | `internal/cli/init.go` |

**Description:**
Git push fails with "src refspec main does not match any" when pushing to a new repository.

**Steps to Reproduce:**
1. Run `gitopsi init --git-url <url> --push`
2. Error: `error: src refspec main does not match any`

**Root Cause:**
The `git init` command at line 335 did not specify the initial branch name. On systems where git defaults to `master`, the local branch is `master`, but gitopsi tries to push `main`. The refspec mismatch causes the push to fail.

**Fix Applied:**
Added `-b` flag to `git init` command to explicitly set the branch name:
```go
// Before: runGitCommand(ctx, projectPath, "init")
// After:  runGitCommand(ctx, projectPath, "init", "-b", initBranch)
```
Where `initBranch` is taken from config or defaults to `main`.

---

### BUG-000 (Template)

| Field | Value |
|-------|-------|
| Date | YYYY-MM-DD |
| Status | fixed |
| Severity | medium |
| Related Files | `internal/example/file.go` |

**Description:**
Brief description of the bug.

**Steps to Reproduce:**
1. Step one
2. Step two
3. Error occurs

**Root Cause:**
Explanation of why the bug occurred.

**Fix Applied:**
Description of the fix, including commit hash if applicable.

---

<!--
Example entry:

### BUG-001

| Field | Value |
|-------|-------|
| Date | 2024-12-26 |
| Status | fixed |
| Severity | high |
| Related Files | `internal/bootstrap/bootstrap.go` |

**Description:**
Bootstrap fails on OpenShift when namespace already exists.

**Steps to Reproduce:**
1. Run `gitopsi init --bootstrap` on OpenShift cluster
2. Namespace `openshift-gitops` already exists from previous install
3. Error: "namespace already exists"

**Root Cause:**
The `CreateNamespace` function didn't check if namespace already exists before creating.

**Fix Applied:**
Added existence check in `bootstrap.go:131` - now uses `CreateNamespaceIfNotExists()`.
Commit: abc1234

-->
