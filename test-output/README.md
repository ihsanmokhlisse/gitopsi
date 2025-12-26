# Test Output Directory

This folder contains E2E test output from OpenShift cluster testing.

## Structure

Each test run creates a timestamped folder with:

```
test-output/
└── YYYYMMDD_HHMMSS/
    ├── test.log              # Full test execution log
    ├── results.csv           # Test results in CSV format
    ├── summary.txt           # Summary of test run
    ├── blockers.txt          # Any blockers found
    ├── issues-to-create.md   # GitHub issues to create
    ├── build.log             # gitopsi build log
    ├── generated/            # Generated GitOps files
    │   ├── gitopsi-config.yaml
    │   ├── init-output.log
    │   ├── file-list.txt
    │   └── project/          # Full generated project
    ├── cluster-state/        # Cluster snapshots
    │   ├── login.log
    │   ├── cluster-info.txt
    │   ├── version.txt
    │   ├── nodes.txt
    │   ├── clusterversion.txt
    │   ├── cluster-operators.txt
    │   ├── gitops-resources.txt
    │   ├── gitops-routes.txt
    │   ├── argocd-url.txt
    │   └── test-namespaces.txt
    └── validation/           # Validation results
        └── manifest-validation.log
```

## Running Tests

```bash
# Set credentials
export OCP_API="https://api.cluster.example.com:6443"
export OCP_USER="admin"
export OCP_PASSWORD="your-password"

# Run full E2E test (with cleanup)
./scripts/e2e-openshift-full.sh

# Run without cleanup (to inspect resources)
./scripts/e2e-openshift-full.sh no-cleanup

# Only cleanup existing test resources
./scripts/e2e-openshift-full.sh cleanup-only
```

## Reviewing Results

After a test run:

1. Check `summary.txt` for overall results
2. Review `results.csv` for detailed test status
3. Check `issues-to-create.md` for any issues found
4. Examine `validation/manifest-validation.log` for manifest issues
5. Review `generated/` folder for generated files

## Cleanup

Test output folders are **not** committed to git. You can safely delete them:

```bash
# Delete all test output
rm -rf test-output/*/

# Delete specific run
rm -rf test-output/YYYYMMDD_HHMMSS/
```

## Auto-Issue Creation

If the `gh` CLI is installed and authenticated, issues are automatically created for:
- Test failures
- Invalid manifests
- Missing features detected during testing

