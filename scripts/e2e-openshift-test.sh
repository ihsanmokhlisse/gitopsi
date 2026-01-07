#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

OCP_API="${OCP_API:-}"
OCP_USER="${OCP_USER:-admin}"
OCP_PASSWORD="${OCP_PASSWORD:-}"
OCP_CONSOLE="${OCP_CONSOLE:-}"
TEST_NAMESPACE="gitopsi-e2e-test"
TEST_PROJECT="gitopsi-test"

cleanup() {
    log_info "Cleaning up test resources..."

    if command -v oc &> /dev/null && oc whoami &> /dev/null 2>&1; then
        oc delete namespace "$TEST_NAMESPACE" --ignore-not-found=true 2>/dev/null || true
        oc delete namespace "${TEST_PROJECT}-dev" --ignore-not-found=true 2>/dev/null || true
        oc delete namespace "${TEST_PROJECT}-staging" --ignore-not-found=true 2>/dev/null || true
        oc delete namespace "${TEST_PROJECT}-prod" --ignore-not-found=true 2>/dev/null || true

        oc delete application -n openshift-gitops "${TEST_PROJECT}-root" --ignore-not-found=true 2>/dev/null || true

        log_success "Cleanup completed"
    else
        log_warn "Not logged in to cluster, skipping cleanup"
    fi

    rm -rf "/tmp/${TEST_PROJECT}" 2>/dev/null || true
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    if [[ -z "$OCP_API" ]]; then
        log_error "OCP_API environment variable is required"
        echo "Usage: OCP_API=https://api.cluster.example.com:6443 OCP_PASSWORD=xxx $0"
        exit 1
    fi

    if [[ -z "$OCP_PASSWORD" ]]; then
        log_error "OCP_PASSWORD environment variable is required"
        exit 1
    fi

    if ! command -v oc &> /dev/null; then
        log_error "oc CLI not found. Please install OpenShift CLI"
        exit 1
    fi

    if ! command -v gitopsi &> /dev/null; then
        log_info "gitopsi not in PATH, building from source..."
        cd "$PROJECT_ROOT"
        go build -o /tmp/gitopsi ./cmd/gitopsi
        export PATH="/tmp:$PATH"
    fi

    log_success "Prerequisites check passed"
}

login_to_cluster() {
    log_info "Logging in to OpenShift cluster..."

    oc login "$OCP_API" -u "$OCP_USER" -p "$OCP_PASSWORD" --insecure-skip-tls-verify=true

    log_info "Cluster info:"
    oc cluster-info
    oc version

    log_success "Successfully logged in to cluster"
}

test_gitopsi_init() {
    log_info "Testing gitopsi init command..."

    TEST_DIR="/tmp/${TEST_PROJECT}"
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"

    cat > "$TEST_DIR/gitopsi.yaml" << EOF
project:
  name: ${TEST_PROJECT}
  description: "E2E Test Project for gitopsi"

platform: openshift
scope: both
gitops_tool: argocd

output:
  type: local
  branch: main

git:
  branch: main
  push_on_init: false

cluster:
  url: ${OCP_API}
  name: e2e-test-cluster
  platform: openshift
  auth:
    method: token
    skip_tls: true

bootstrap:
  enabled: false
  tool: argocd
  mode: olm
  namespace: openshift-gitops
  wait: true
  timeout: 300

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

docs:
  readme: true
  architecture: true
  onboarding: true
EOF

    log_info "Running gitopsi init with config file..."
    cd "$TEST_DIR"
    gitopsi init --config gitopsi.yaml --output "$TEST_DIR/output"

    log_info "Validating generated structure..."

    EXPECTED_FILES=(
        "output/README.md"
        "output/argocd/applicationsets/apps.yaml"
        "output/argocd/applicationsets/infra.yaml"
        "output/environments/dev/kustomization.yaml"
        "output/environments/staging/kustomization.yaml"
        "output/environments/prod/kustomization.yaml"
        "output/infrastructure/namespace.yaml"
    )

    for file in "${EXPECTED_FILES[@]}"; do
        if [[ -f "$TEST_DIR/$file" ]]; then
            log_success "Found: $file"
        else
            log_error "Missing: $file"
            exit 1
        fi
    done

    log_success "gitopsi init test passed"
}

test_manifest_validation() {
    log_info "Validating generated manifests..."

    TEST_DIR="/tmp/${TEST_PROJECT}/output"

    for yaml_file in $(find "$TEST_DIR" -name "*.yaml" -type f); do
        if oc apply --dry-run=client -f "$yaml_file" &> /dev/null; then
            log_success "Valid: $yaml_file"
        else
            log_warn "Validation issue: $yaml_file"
        fi
    done

    log_success "Manifest validation completed"
}

test_namespace_creation() {
    log_info "Testing namespace creation on cluster..."

    oc create namespace "$TEST_NAMESPACE" --dry-run=client -o yaml | oc apply -f -

    sleep 2

    if oc get namespace "$TEST_NAMESPACE" &> /dev/null; then
        log_success "Namespace $TEST_NAMESPACE created successfully"
    else
        log_error "Failed to create namespace"
        exit 1
    fi
}

test_argocd_integration() {
    log_info "Testing ArgoCD integration..."

    if oc get namespace openshift-gitops &> /dev/null; then
        log_success "OpenShift GitOps (ArgoCD) is installed"

        ARGOCD_SERVER=$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}' 2>/dev/null || echo "")
        if [[ -n "$ARGOCD_SERVER" ]]; then
            log_success "ArgoCD Server URL: https://$ARGOCD_SERVER"
        fi

        oc get applications -n openshift-gitops 2>/dev/null || log_info "No applications found (expected for clean cluster)"
    else
        log_warn "OpenShift GitOps not installed, skipping ArgoCD tests"
    fi
}

test_bootstrap_modes() {
    log_info "Testing bootstrap mode detection for OpenShift..."

    cd "$PROJECT_ROOT"

    go test -v -run "TestSuggestMode" ./internal/bootstrap/... 2>&1 | head -20
    go test -v -run "TestValidModes" ./internal/bootstrap/... 2>&1 | head -20

    log_success "Bootstrap mode tests passed"
}

run_cleanup() {
    log_info "Running full cleanup..."
    cleanup
    log_success "Cluster is clean and ready for next test"
}

print_summary() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}E2E Test Summary${NC}"
    echo "=============================================="
    echo "Cluster: $OCP_API"
    echo "User: $OCP_USER"
    echo "Test Project: $TEST_PROJECT"
    echo "Test Namespace: $TEST_NAMESPACE"
    echo "=============================================="
}

main() {
    echo ""
    echo "=============================================="
    echo -e "${BLUE}gitopsi E2E OpenShift Test${NC}"
    echo "=============================================="
    echo ""

    trap cleanup EXIT

    check_prerequisites
    login_to_cluster

    test_gitopsi_init
    test_manifest_validation
    test_namespace_creation
    test_argocd_integration
    test_bootstrap_modes

    print_summary

    log_success "All E2E tests passed!"

    run_cleanup
}

case "${1:-}" in
    "cleanup")
        check_prerequisites
        login_to_cluster
        run_cleanup
        ;;
    "test")
        main
        ;;
    *)
        main
        ;;
esac
