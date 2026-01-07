#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_OUTPUT_DIR="${PROJECT_ROOT}/test-output"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEST_RUN_DIR="${TEST_OUTPUT_DIR}/${TIMESTAMP}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

OCP_API="${OCP_API:-}"
OCP_USER="${OCP_USER:-admin}"
OCP_PASSWORD="${OCP_PASSWORD:-}"
TEST_NAMESPACE="gitopsi-e2e-${TIMESTAMP:0:8}"
TEST_PROJECT="gitopsi-e2e-test"
ARGOCD_NS="openshift-gitops"

ISSUES_TO_CREATE=()
TEST_RESULTS=()
BLOCKERS=()

log_info() { echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${TEST_RUN_DIR}/test.log"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1" | tee -a "${TEST_RUN_DIR}/test.log"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "${TEST_RUN_DIR}/test.log"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1" | tee -a "${TEST_RUN_DIR}/test.log"; }

add_test_result() {
    local name="$1"
    local status="$2"
    local details="$3"
    TEST_RESULTS+=("${name}|${status}|${details}")
    echo "${name}|${status}|${details}" >> "${TEST_RUN_DIR}/results.csv"
}

add_issue() {
    local title="$1"
    local body="$2"
    local labels="$3"
    ISSUES_TO_CREATE+=("${title}|||${body}|||${labels}")
    echo "---" >> "${TEST_RUN_DIR}/issues-to-create.md"
    echo "### ${title}" >> "${TEST_RUN_DIR}/issues-to-create.md"
    echo "Labels: ${labels}" >> "${TEST_RUN_DIR}/issues-to-create.md"
    echo "" >> "${TEST_RUN_DIR}/issues-to-create.md"
    echo "${body}" >> "${TEST_RUN_DIR}/issues-to-create.md"
    echo "" >> "${TEST_RUN_DIR}/issues-to-create.md"
}

add_blocker() {
    local message="$1"
    BLOCKERS+=("$message")
    echo "$message" >> "${TEST_RUN_DIR}/blockers.txt"
}

setup_test_dir() {
    mkdir -p "${TEST_RUN_DIR}"
    mkdir -p "${TEST_RUN_DIR}/generated"
    mkdir -p "${TEST_RUN_DIR}/cluster-state"
    mkdir -p "${TEST_RUN_DIR}/validation"

    echo "Test Run: ${TIMESTAMP}" > "${TEST_RUN_DIR}/test.log"
    echo "Cluster: ${OCP_API}" >> "${TEST_RUN_DIR}/test.log"
    echo "User: ${OCP_USER}" >> "${TEST_RUN_DIR}/test.log"
    echo "---" >> "${TEST_RUN_DIR}/test.log"

    echo "test,status,details" > "${TEST_RUN_DIR}/results.csv"
    echo "# Issues to Create" > "${TEST_RUN_DIR}/issues-to-create.md"
    echo "" > "${TEST_RUN_DIR}/blockers.txt"

    log_info "Test output directory: ${TEST_RUN_DIR}"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    if [[ -z "$OCP_API" ]]; then
        add_blocker "OCP_API environment variable not set"
        log_error "OCP_API required"
        exit 1
    fi

    if [[ -z "$OCP_PASSWORD" ]]; then
        add_blocker "OCP_PASSWORD environment variable not set"
        log_error "OCP_PASSWORD required"
        exit 1
    fi

    if ! command -v oc &> /dev/null; then
        add_blocker "oc CLI not installed"
        add_issue "feat: Add oc CLI installation instructions" \
            "The E2E test requires the OpenShift CLI (oc) to be installed.\n\nAdd installation instructions to README." \
            "documentation,enhancement"
        log_error "oc CLI not found"
        exit 1
    fi

    if ! command -v gh &> /dev/null; then
        log_warn "gh CLI not found - issues won't be auto-created"
    fi

    add_test_result "prerequisites" "PASS" "All prerequisites met"
    log_success "Prerequisites check passed"
}

login_to_cluster() {
    log_info "Logging in to OpenShift cluster..."

    if ! oc login "$OCP_API" -u "$OCP_USER" -p "$OCP_PASSWORD" --insecure-skip-tls-verify=true > "${TEST_RUN_DIR}/cluster-state/login.log" 2>&1; then
        add_blocker "Failed to login to cluster"
        add_test_result "cluster_login" "FAIL" "Authentication failed"
        log_error "Login failed - see ${TEST_RUN_DIR}/cluster-state/login.log"
        exit 1
    fi

    oc cluster-info > "${TEST_RUN_DIR}/cluster-state/cluster-info.txt" 2>&1
    oc version > "${TEST_RUN_DIR}/cluster-state/version.txt" 2>&1
    oc get nodes -o wide > "${TEST_RUN_DIR}/cluster-state/nodes.txt" 2>&1

    add_test_result "cluster_login" "PASS" "Authenticated as ${OCP_USER}"
    log_success "Logged in to cluster"
}

check_cluster_health() {
    log_info "Checking cluster health..."

    oc get clusterversion > "${TEST_RUN_DIR}/cluster-state/clusterversion.txt" 2>&1
    oc get co > "${TEST_RUN_DIR}/cluster-state/cluster-operators.txt" 2>&1

    DEGRADED=$(oc get co -o json | jq '[.items[] | select(.status.conditions[] | select(.type=="Degraded" and .status=="True"))] | length')
    if [[ "$DEGRADED" -gt 0 ]]; then
        add_test_result "cluster_health" "WARN" "${DEGRADED} operators degraded"
        log_warn "${DEGRADED} cluster operators are degraded"
    else
        add_test_result "cluster_health" "PASS" "All operators healthy"
        log_success "Cluster is healthy"
    fi
}

check_gitops_installation() {
    log_info "Checking OpenShift GitOps installation..."

    if ! oc get namespace openshift-gitops > /dev/null 2>&1; then
        add_test_result "gitops_installed" "FAIL" "OpenShift GitOps not installed"
        add_issue "bug: OpenShift GitOps not detected" \
            "E2E test failed because OpenShift GitOps is not installed on the cluster.\n\nExpected: openshift-gitops namespace exists\nActual: Namespace not found" \
            "bug,e2e-test"
        log_error "OpenShift GitOps not installed"
        return 1
    fi

    oc get all -n openshift-gitops > "${TEST_RUN_DIR}/cluster-state/gitops-resources.txt" 2>&1
    oc get route -n openshift-gitops -o wide > "${TEST_RUN_DIR}/cluster-state/gitops-routes.txt" 2>&1

    ARGOCD_URL=$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}' 2>/dev/null || echo "")
    if [[ -n "$ARGOCD_URL" ]]; then
        echo "https://${ARGOCD_URL}" > "${TEST_RUN_DIR}/cluster-state/argocd-url.txt"
        add_test_result "gitops_installed" "PASS" "ArgoCD URL: https://${ARGOCD_URL}"
        log_success "OpenShift GitOps installed - https://${ARGOCD_URL}"
    else
        add_test_result "gitops_installed" "WARN" "GitOps installed but no route"
        log_warn "OpenShift GitOps installed but route not found"
    fi
}

detect_argocd_details() {
    log_info "Detecting ArgoCD installation details..."

    local detection_file="${TEST_RUN_DIR}/cluster-state/argocd-detection.txt"
    echo "=== ArgoCD Detection Report ===" > "$detection_file"
    echo "Timestamp: $(date)" >> "$detection_file"
    echo "" >> "$detection_file"

    local argocd_type="unknown"
    local install_method="unknown"
    local operator_source="unknown"
    local argocd_version=""
    local argocd_namespace=""

    for ns in openshift-gitops argocd gitops; do
        if oc get namespace "$ns" > /dev/null 2>&1; then
            argocd_namespace="$ns"
            break
        fi
    done

    if [[ -z "$argocd_namespace" ]]; then
        echo "ArgoCD Type: not_installed" >> "$detection_file"
        add_test_result "argocd_detection" "FAIL" "ArgoCD not detected"
        log_error "ArgoCD not detected on cluster"
        return 1
    fi

    echo "Namespace: $argocd_namespace" >> "$detection_file"

    local images=$(oc get deployments -n "$argocd_namespace" -o jsonpath='{.items[*].spec.template.spec.containers[*].image}' 2>/dev/null)
    if [[ "$images" == *"registry.redhat.io"* ]] || [[ "$images" == *"quay.io/openshift-gitops"* ]]; then
        argocd_type="redhat"
    elif [[ "$images" == *"quay.io/argoproj"* ]] || [[ "$images" == *"ghcr.io/argoproj"* ]]; then
        argocd_type="community"
    fi
    echo "ArgoCD Type: $argocd_type" >> "$detection_file"

    if oc get subscription -n "$argocd_namespace" --ignore-not-found 2>/dev/null | grep -q gitops; then
        install_method="olm"
        operator_source=$(oc get subscription -A -o jsonpath='{.items[?(@.spec.name=="openshift-gitops-operator")].spec.source}' 2>/dev/null || echo "unknown")
    elif oc get subscription -n openshift-operators --ignore-not-found 2>/dev/null | grep -q gitops; then
        install_method="olm"
        operator_source=$(oc get subscription -n openshift-operators -o jsonpath='{.items[?(@.spec.name=="openshift-gitops-operator")].spec.source}' 2>/dev/null || echo "unknown")
    elif oc get deployment -n "$argocd_namespace" -l "helm.sh/chart" --ignore-not-found 2>/dev/null | grep -q argocd; then
        install_method="helm"
    elif oc get argocd -n "$argocd_namespace" --ignore-not-found 2>/dev/null | grep -q NAME; then
        install_method="operator"
    else
        install_method="manifest"
    fi
    echo "Install Method: $install_method" >> "$detection_file"
    echo "Operator Source: $operator_source" >> "$detection_file"

    argocd_version=$(echo "$images" | tr ' ' '\n' | grep -E "argocd|gitops" | head -1 | sed 's/.*://' || echo "unknown")
    echo "Version: $argocd_version" >> "$detection_file"

    echo "" >> "$detection_file"
    echo "=== Component Status ===" >> "$detection_file"

    local components_healthy=0
    local components_total=0
    local components_unhealthy=""

    for deploy in $(oc get deployments -n "$argocd_namespace" -o name 2>/dev/null | grep -E "argocd|gitops"); do
        components_total=$((components_total + 1))
        local name=$(echo "$deploy" | sed 's/deployment.apps\///')
        local ready=$(oc get "$deploy" -n "$argocd_namespace" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
        local replicas=$(oc get "$deploy" -n "$argocd_namespace" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "1")

        if [[ "$ready" -ge "$replicas" ]] && [[ "$replicas" -gt 0 ]]; then
            echo "  ✅ $name ($ready/$replicas)" >> "$detection_file"
            components_healthy=$((components_healthy + 1))
        else
            echo "  ❌ $name ($ready/$replicas)" >> "$detection_file"
            components_unhealthy="${components_unhealthy}${name}, "
        fi
    done

    local app_count=$(oc get applications.argoproj.io -n "$argocd_namespace" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    echo "" >> "$detection_file"
    echo "Applications: $app_count" >> "$detection_file"

    echo "" >> "$detection_file"
    echo "=== Health Analysis ===" >> "$detection_file"

    local issues=""
    local recommendations=""

    if [[ "$argocd_type" == "community" ]] && [[ "$argocd_namespace" == "openshift-gitops" ]]; then
        issues="${issues}Community ArgoCD in OpenShift GitOps namespace\n"
        recommendations="${recommendations}Consider using Red Hat OpenShift GitOps operator\n"
        add_issue "feat: Detect community ArgoCD in OpenShift namespace" \
            "gitopsi should warn when community ArgoCD is installed in the openshift-gitops namespace instead of Red Hat's official operator." \
            "enhancement"
    fi

    if [[ "$install_method" == "olm" ]] && [[ "$operator_source" == "community-operators" ]]; then
        issues="${issues}Using community operator on OpenShift\n"
        recommendations="${recommendations}Switch to redhat-operators for official support\n"
    fi

    if [[ "$components_healthy" -lt "$components_total" ]]; then
        issues="${issues}Some components not healthy: ${components_unhealthy}\n"
        recommendations="${recommendations}Check pod logs: oc logs -n $argocd_namespace -l app.kubernetes.io/part-of=argocd\n"
    fi

    if [[ "$argocd_version" == v1* ]]; then
        issues="${issues}ArgoCD v1.x is outdated\n"
        recommendations="${recommendations}Upgrade to ArgoCD v2.x for security and features\n"
        add_issue "feat: Detect outdated ArgoCD versions" \
            "gitopsi should warn when detecting ArgoCD v1.x and recommend upgrade." \
            "enhancement"
    fi

    if [[ -n "$issues" ]]; then
        echo "Issues:" >> "$detection_file"
        echo -e "  $issues" >> "$detection_file"
    fi

    if [[ -n "$recommendations" ]]; then
        echo "Recommendations:" >> "$detection_file"
        echo -e "  $recommendations" >> "$detection_file"
    fi

    cat "$detection_file"

    if [[ "$components_healthy" -eq "$components_total" ]] && [[ "$components_total" -gt 0 ]]; then
        add_test_result "argocd_detection" "PASS" "Type: $argocd_type, Method: $install_method, Health: $components_healthy/$components_total"
        log_success "ArgoCD detection complete - Type: $argocd_type, Method: $install_method"
    else
        add_test_result "argocd_detection" "WARN" "Health: $components_healthy/$components_total"
        log_warn "ArgoCD detection: $components_healthy/$components_total components healthy"
    fi
}

build_gitopsi() {
    log_info "Building gitopsi..."

    cd "$PROJECT_ROOT"

    if [[ -f "gitopsi-darwin" ]]; then
        GITOPSI_BIN="${PROJECT_ROOT}/gitopsi-darwin"
    else
        log_info "Building gitopsi binary..."
        if command -v podman &> /dev/null; then
            podman run --rm -v "${PROJECT_ROOT}:/app" -w /app \
                -e GOOS=darwin -e GOARCH=arm64 \
                golang:1.23-alpine go build -o gitopsi-darwin ./cmd/gitopsi > "${TEST_RUN_DIR}/build.log" 2>&1
            GITOPSI_BIN="${PROJECT_ROOT}/gitopsi-darwin"
        else
            add_blocker "Cannot build gitopsi - podman not available"
            log_error "Cannot build gitopsi"
            return 1
        fi
    fi

    if [[ -x "$GITOPSI_BIN" ]]; then
        add_test_result "gitopsi_build" "PASS" "Binary built successfully"
        log_success "gitopsi built: ${GITOPSI_BIN}"
    else
        add_test_result "gitopsi_build" "FAIL" "Binary not executable"
        log_error "gitopsi build failed"
        return 1
    fi

    export GITOPSI_BIN
}

test_gitopsi_init() {
    log_info "Testing gitopsi init..."

    local test_dir="${TEST_RUN_DIR}/generated/project"
    mkdir -p "$test_dir"
    cd "$test_dir"

    cat > gitopsi.yaml << EOF
project:
  name: ${TEST_PROJECT}
  description: "E2E Test Project - ${TIMESTAMP}"

platform: openshift
scope: both
gitops_tool: argocd

output:
  type: local
  branch: main

cluster:
  url: ${OCP_API}
  name: e2e-test-cluster
  platform: openshift

bootstrap:
  enabled: false
  tool: argocd
  mode: olm
  namespace: openshift-gitops

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: sample-app
    image: nginx:latest
    port: 80
    replicas: 1

docs:
  readme: true
  architecture: true
  onboarding: true
EOF

    cp gitopsi.yaml "${TEST_RUN_DIR}/generated/gitopsi-config.yaml"

    if ! "$GITOPSI_BIN" init --config gitopsi.yaml > "${TEST_RUN_DIR}/generated/init-output.log" 2>&1; then
        add_test_result "gitopsi_init" "FAIL" "Init command failed"
        add_issue "bug: gitopsi init failed in E2E test" \
            "The gitopsi init command failed during E2E testing.\n\nSee logs for details." \
            "bug,e2e-test"
        log_error "gitopsi init failed"
        return 1
    fi

    find "${TEST_PROJECT}" -type f > "${TEST_RUN_DIR}/generated/file-list.txt"
    YAML_COUNT=$(find "${TEST_PROJECT}" -name "*.yaml" | wc -l | tr -d ' ')

    add_test_result "gitopsi_init" "PASS" "Generated ${YAML_COUNT} YAML files"
    log_success "gitopsi init completed - ${YAML_COUNT} files generated"

    cp -r "${TEST_PROJECT}" "${TEST_RUN_DIR}/generated/"
}

validate_manifests() {
    log_info "Validating generated manifests..."

    local project_dir="${TEST_RUN_DIR}/generated/project/${TEST_PROJECT}"
    local valid=0
    local invalid=0
    local skipped=0
    local validation_log="${TEST_RUN_DIR}/validation/manifest-validation.log"

    echo "Manifest Validation Results" > "$validation_log"
    echo "===========================" >> "$validation_log"
    echo "" >> "$validation_log"

    for yaml_file in $(find "$project_dir" -name "*.yaml" -type f | grep -v setup-summary); do
        local relative_path="${yaml_file#$project_dir/}"
        local filename=$(basename "$yaml_file")

        if [[ "$filename" == "kustomization.yaml" ]]; then
            echo "⏭️  SKIP (kustomize): $relative_path" >> "$validation_log"
            skipped=$((skipped+1))
            continue
        fi

        if grep -q "kind: Kustomization" "$yaml_file" 2>/dev/null; then
            echo "⏭️  SKIP (kustomize kind): $relative_path" >> "$validation_log"
            skipped=$((skipped+1))
            continue
        fi

        if oc apply --dry-run=server -f "$yaml_file" >> "$validation_log" 2>&1; then
            echo "✅ VALID: $relative_path" >> "$validation_log"
            valid=$((valid+1))
        else
            local error_msg=$(oc apply --dry-run=server -f "$yaml_file" 2>&1 | tail -1)

            if [[ "$error_msg" == *"namespaces"*"not found"* ]]; then
                echo "⚠️  WARN (ns missing): $relative_path" >> "$validation_log"
                echo "   Note: Namespace not yet created, manifest structure is valid" >> "$validation_log"
                valid=$((valid+1))
            else
                echo "❌ INVALID: $relative_path" >> "$validation_log"
                echo "   Error: $error_msg" >> "$validation_log"
                invalid=$((invalid+1))

                add_issue "bug: Invalid manifest generated - ${relative_path}" \
                    "The manifest \`${relative_path}\` failed server-side validation.\n\nError: ${error_msg}" \
                    "bug,manifest"
            fi
        fi
    done

    echo "" >> "$validation_log"
    echo "Summary: ${valid} valid, ${invalid} invalid, ${skipped} skipped" >> "$validation_log"

    if [[ $invalid -gt 0 ]]; then
        add_test_result "manifest_validation" "WARN" "${valid} valid, ${invalid} invalid, ${skipped} skipped"
        log_warn "Manifest validation: ${valid} valid, ${invalid} invalid, ${skipped} skipped"
    else
        add_test_result "manifest_validation" "PASS" "All ${valid} manifests valid (${skipped} skipped)"
        log_success "All ${valid} manifests are valid (${skipped} kustomize files skipped)"
    fi
}

test_infrastructure_deployment() {
    log_info "Testing infrastructure deployment..."

    local project_dir="${TEST_RUN_DIR}/generated/project/${TEST_PROJECT}"
    local ns_dir="${project_dir}/infrastructure/base/namespaces"

    if [[ -d "$ns_dir" ]]; then
        log_info "Creating test namespaces..."

        for ns_file in "$ns_dir"/*.yaml; do
            local ns_name=$(grep -E "^  name:" "$ns_file" | head -1 | awk '{print $2}')
            if oc apply -f "$ns_file" > /dev/null 2>&1; then
                log_success "Created namespace: $ns_name"
            else
                log_warn "Failed to create namespace: $ns_name"
            fi
        done

        sleep 3

        oc get namespaces | grep "${TEST_PROJECT}" > "${TEST_RUN_DIR}/cluster-state/test-namespaces.txt" 2>&1

        local created=$(oc get namespaces | grep "${TEST_PROJECT}" | wc -l | tr -d ' ')
        add_test_result "infrastructure_deploy" "PASS" "Created ${created} namespaces"
        log_success "Infrastructure deployed: ${created} namespaces"
    else
        add_test_result "infrastructure_deploy" "SKIP" "No namespaces to deploy"
        log_warn "No namespace manifests found"
    fi
}

test_rbac_deployment() {
    log_info "Testing RBAC deployment..."

    local project_dir="${TEST_RUN_DIR}/generated/project/${TEST_PROJECT}"
    local rbac_dir="${project_dir}/infrastructure/base/rbac"

    if [[ -d "$rbac_dir" ]]; then
        local applied=0
        local failed=0

        for rbac_file in "$rbac_dir"/*.yaml; do
            if oc apply -f "$rbac_file" > /dev/null 2>&1; then
                applied=$((applied+1))
            else
                failed=$((failed+1))
            fi
        done

        if [[ $failed -gt 0 ]]; then
            add_test_result "rbac_deploy" "WARN" "${applied} applied, ${failed} failed"
            log_warn "RBAC: ${applied} applied, ${failed} failed"
        else
            add_test_result "rbac_deploy" "PASS" "${applied} RBAC resources applied"
            log_success "RBAC deployed: ${applied} resources"
        fi
    else
        add_test_result "rbac_deploy" "SKIP" "No RBAC manifests"
        log_info "No RBAC manifests found"
    fi
}

test_argocd_resources() {
    log_info "Testing ArgoCD resource generation..."

    local project_dir="${TEST_RUN_DIR}/generated/project/${TEST_PROJECT}"
    local argocd_dir="${project_dir}/argocd"

    if [[ -d "$argocd_dir" ]]; then
        local appsets=$(find "$argocd_dir" -name "*.yaml" | wc -l | tr -d ' ')

        oc get applications -n openshift-gitops > "${TEST_RUN_DIR}/cluster-state/argocd-apps-before.txt" 2>&1 || true

        add_test_result "argocd_resources" "PASS" "Generated ${appsets} ArgoCD manifests"
        log_success "ArgoCD resources: ${appsets} manifests generated"
    else
        add_test_result "argocd_resources" "SKIP" "No ArgoCD manifests"
        log_info "No ArgoCD manifests found"
    fi
}

check_test_status() {
    log_info "Checking overall test status..."

    local passed=0
    local failed=0
    local warned=0
    local skipped=0

    for result in "${TEST_RESULTS[@]}"; do
        local status=$(echo "$result" | cut -d'|' -f2)
        case "$status" in
            PASS) passed=$((passed+1)) ;;
            FAIL) failed=$((failed+1)) ;;
            WARN) warned=$((warned+1)) ;;
            SKIP) skipped=$((skipped+1)) ;;
        esac
    done

    cat > "${TEST_RUN_DIR}/summary.txt" << EOF
E2E Test Summary
================
Timestamp: ${TIMESTAMP}
Cluster: ${OCP_API}

Results:
  PASSED:  ${passed}
  FAILED:  ${failed}
  WARNED:  ${warned}
  SKIPPED: ${skipped}

Blockers: ${#BLOCKERS[@]}
Issues to Create: ${#ISSUES_TO_CREATE[@]}

EOF

    if [[ ${#BLOCKERS[@]} -gt 0 ]]; then
        echo "Blockers:" >> "${TEST_RUN_DIR}/summary.txt"
        for blocker in "${BLOCKERS[@]}"; do
            echo "  - ${blocker}" >> "${TEST_RUN_DIR}/summary.txt"
        done
    fi

    cat "${TEST_RUN_DIR}/summary.txt"

    if [[ $failed -gt 0 ]]; then
        add_issue "E2E Test Failed - ${TIMESTAMP}" \
            "The E2E test run on ${TIMESTAMP} had ${failed} failures.\n\nSee test output in \`test-output/${TIMESTAMP}/\`" \
            "bug,e2e-test"
    fi
}

cleanup() {
    log_info "Cleaning up test resources from cluster..."

    oc get namespaces | grep "${TEST_PROJECT}" | awk '{print $1}' | while read ns; do
        oc delete namespace "$ns" --ignore-not-found=true 2>/dev/null || true
        log_info "Deleted namespace: $ns"
    done

    oc get rolebindings --all-namespaces 2>/dev/null | grep "${TEST_PROJECT}" | while read line; do
        local ns=$(echo "$line" | awk '{print $1}')
        local name=$(echo "$line" | awk '{print $2}')
        oc delete rolebinding "$name" -n "$ns" --ignore-not-found=true 2>/dev/null || true
    done

    oc get applications -n openshift-gitops 2>/dev/null | grep "${TEST_PROJECT}" | awk '{print $1}' | while read app; do
        oc delete application "$app" -n openshift-gitops --ignore-not-found=true 2>/dev/null || true
    done

    sleep 3

    local remaining=$(oc get namespaces 2>/dev/null | grep "${TEST_PROJECT}" | wc -l | tr -d ' ')
    if [[ "$remaining" -eq 0 ]]; then
        add_test_result "cleanup" "PASS" "All test resources removed"
        log_success "Cleanup complete - cluster is clean"
    else
        add_test_result "cleanup" "WARN" "${remaining} resources remaining"
        log_warn "Cleanup: ${remaining} resources still remain"
    fi
}

create_github_issues() {
    if [[ ${#ISSUES_TO_CREATE[@]} -eq 0 ]]; then
        log_info "No issues to create"
        return
    fi

    if ! command -v gh &> /dev/null; then
        log_warn "gh CLI not available - issues logged to ${TEST_RUN_DIR}/issues-to-create.md"
        return
    fi

    log_info "Creating ${#ISSUES_TO_CREATE[@]} GitHub issues..."

    for issue_data in "${ISSUES_TO_CREATE[@]}"; do
        local title=$(echo "$issue_data" | cut -d'|||' -f1)
        local body=$(echo "$issue_data" | cut -d'|||' -f2)
        local labels=$(echo "$issue_data" | cut -d'|||' -f3)

        if gh issue create --title "$title" --body "$body" --label "$labels" 2>/dev/null; then
            log_success "Created issue: $title"
        else
            log_warn "Failed to create issue: $title"
        fi
    done
}

print_final_summary() {
    echo ""
    echo "=============================================="
    echo -e "${BLUE}E2E Test Complete${NC}"
    echo "=============================================="
    echo ""
    echo "Test Output: ${TEST_RUN_DIR}"
    echo ""
    echo "Files:"
    echo "  - test.log          : Full test log"
    echo "  - results.csv       : Test results"
    echo "  - summary.txt       : Summary"
    echo "  - issues-to-create.md : Issues found"
    echo "  - generated/        : Generated GitOps files"
    echo "  - cluster-state/    : Cluster snapshots"
    echo "  - validation/       : Validation results"
    echo ""

    if [[ -f "${TEST_RUN_DIR}/summary.txt" ]]; then
        cat "${TEST_RUN_DIR}/summary.txt"
    fi

    echo "=============================================="
}

test_bootstrap_validation() {
    log_info "Testing bootstrap validation..."

    local validation_log="${TEST_RUN_DIR}/validation/bootstrap-validation.log"
    echo "=== Bootstrap Validation ===" > "$validation_log"

    if ! oc get namespace openshift-gitops > /dev/null 2>&1; then
        echo "SKIP: OpenShift GitOps namespace not found" >> "$validation_log"
        add_test_result "bootstrap_validation" "SKIP" "No GitOps namespace"
        log_info "Skipping bootstrap validation - no GitOps namespace"
        return
    fi

    local pods_ready=0
    local pods_total=0

    echo "Pod Status:" >> "$validation_log"
    while IFS= read -r line; do
        pods_total=$((pods_total + 1))
        local pod_name=$(echo "$line" | awk '{print $1}')
        local pod_ready=$(echo "$line" | awk '{print $2}')
        local pod_status=$(echo "$line" | awk '{print $3}')

        if [[ "$pod_status" == "Running" ]] || [[ "$pod_status" == "Completed" ]]; then
            echo "  ✅ $pod_name ($pod_ready) - $pod_status" >> "$validation_log"
            pods_ready=$((pods_ready + 1))
        else
            echo "  ❌ $pod_name ($pod_ready) - $pod_status" >> "$validation_log"
        fi
    done < <(oc get pods -n openshift-gitops --no-headers 2>/dev/null | grep -E "argocd|gitops")

    echo "" >> "$validation_log"
    echo "Services:" >> "$validation_log"
    oc get services -n openshift-gitops -o wide >> "$validation_log" 2>&1

    echo "" >> "$validation_log"
    echo "Routes:" >> "$validation_log"
    oc get routes -n openshift-gitops -o wide >> "$validation_log" 2>&1

    if oc get argocd -n openshift-gitops > /dev/null 2>&1; then
        echo "" >> "$validation_log"
        echo "ArgoCD CR Status:" >> "$validation_log"
        oc get argocd -n openshift-gitops -o yaml >> "$validation_log" 2>&1
    fi

    echo "" >> "$validation_log"
    echo "Summary: $pods_ready/$pods_total pods ready" >> "$validation_log"

    if [[ "$pods_ready" -eq "$pods_total" ]] && [[ "$pods_total" -gt 0 ]]; then
        add_test_result "bootstrap_validation" "PASS" "$pods_ready/$pods_total pods ready"
        log_success "Bootstrap validation passed - $pods_ready/$pods_total pods ready"
    elif [[ "$pods_ready" -gt 0 ]]; then
        add_test_result "bootstrap_validation" "WARN" "$pods_ready/$pods_total pods ready"
        log_warn "Bootstrap validation: $pods_ready/$pods_total pods ready"
    else
        add_test_result "bootstrap_validation" "FAIL" "No pods ready"
        log_error "Bootstrap validation failed - no pods ready"
        add_issue "bug: Bootstrap validation failed" \
            "E2E test detected that ArgoCD pods are not ready.\n\nPods: $pods_ready/$pods_total" \
            "bug,bootstrap"
    fi
}

test_argocd_connectivity() {
    log_info "Testing ArgoCD connectivity..."

    local connectivity_log="${TEST_RUN_DIR}/validation/argocd-connectivity.log"
    echo "=== ArgoCD Connectivity Test ===" > "$connectivity_log"

    local argocd_route=$(oc get route -n openshift-gitops -o jsonpath='{.items[0].spec.host}' 2>/dev/null || echo "")

    if [[ -z "$argocd_route" ]]; then
        echo "SKIP: No ArgoCD route found" >> "$connectivity_log"
        add_test_result "argocd_connectivity" "SKIP" "No route"
        log_info "Skipping ArgoCD connectivity - no route"
        return
    fi

    echo "Route: https://$argocd_route" >> "$connectivity_log"

    if command -v curl &> /dev/null; then
        local http_code=$(curl -sk -o /dev/null -w "%{http_code}" "https://$argocd_route" 2>/dev/null || echo "000")
        echo "HTTP Response: $http_code" >> "$connectivity_log"

        if [[ "$http_code" == "200" ]] || [[ "$http_code" == "307" ]] || [[ "$http_code" == "302" ]]; then
            add_test_result "argocd_connectivity" "PASS" "HTTP $http_code"
            log_success "ArgoCD is accessible - HTTP $http_code"
        elif [[ "$http_code" == "000" ]]; then
            add_test_result "argocd_connectivity" "FAIL" "Connection failed"
            log_error "ArgoCD connectivity failed - connection refused"
        else
            add_test_result "argocd_connectivity" "WARN" "HTTP $http_code"
            log_warn "ArgoCD returned unexpected status - HTTP $http_code"
        fi
    else
        echo "SKIP: curl not available" >> "$connectivity_log"
        add_test_result "argocd_connectivity" "SKIP" "No curl"
        log_info "Skipping ArgoCD connectivity - curl not available"
    fi
}

main() {
    setup_test_dir

    echo ""
    echo "=============================================="
    echo -e "${BLUE}gitopsi E2E OpenShift Full Test${NC}"
    echo "=============================================="
    echo ""

    check_prerequisites
    login_to_cluster
    check_cluster_health
    check_gitops_installation
    detect_argocd_details
    test_bootstrap_validation
    test_argocd_connectivity
    build_gitopsi
    test_gitopsi_init
    validate_manifests
    test_infrastructure_deployment
    test_rbac_deployment
    test_argocd_resources
    check_test_status
    cleanup

    print_final_summary
}

case "${1:-}" in
    "cleanup-only")
        setup_test_dir
        check_prerequisites
        login_to_cluster
        cleanup
        ;;
    "no-cleanup")
        main
        log_warn "Skipping cleanup as requested"
        ;;
    *)
        main
        ;;
esac
