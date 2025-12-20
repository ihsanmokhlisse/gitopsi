#!/bin/bash
# Product Owner Validation Test for gitopsi
# This script sets up a local Kubernetes cluster for testing

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

CLUSTER_NAME="gitopsi-test"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
GITOPSI_BIN="$PROJECT_DIR/bin/gitopsi"

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║              gitopsi - Product Owner Validation Test                   ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    kind delete cluster --name $CLUSTER_NAME 2>/dev/null || true
    rm -rf "$PROJECT_DIR/gitopsi-product-test" 2>/dev/null || true
    rm -f /tmp/gitopsi-product-test.yaml 2>/dev/null || true
    echo -e "${GREEN}✅ Cleanup complete${NC}"
}

case "${1:-}" in
    cleanup|clean)
        cleanup
        exit 0
        ;;
    *)
        ;;
esac

echo -e "${BLUE}Step 1: Checking prerequisites...${NC}"
echo "────────────────────────────────────────────────────────────────────────"

if ! command -v kind &> /dev/null; then
    echo -e "${RED}❌ Kind not found. Install with: brew install kind${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} Kind available"

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl not found. Install with: brew install kubectl${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} kubectl available"

if [ ! -f "$GITOPSI_BIN" ]; then
    echo -e "${RED}❌ gitopsi binary not found at $GITOPSI_BIN${NC}"
    echo "  Run: make build"
    exit 1
fi
chmod +x "$GITOPSI_BIN"
echo -e "${GREEN}✓${NC} gitopsi binary ready"

echo ""
echo -e "${BLUE}Step 2: Setting up local Kubernetes cluster...${NC}"
echo "────────────────────────────────────────────────────────────────────────"

kind delete cluster --name $CLUSTER_NAME 2>/dev/null || true
echo "Creating Kind cluster '$CLUSTER_NAME'..."

kind create cluster --name $CLUSTER_NAME --wait 5m

kubectl cluster-info --context kind-$CLUSTER_NAME
echo -e "${GREEN}✅ Kubernetes cluster ready${NC}"

echo ""
echo -e "${BLUE}Step 3: Creating test configuration...${NC}"
echo "────────────────────────────────────────────────────────────────────────"

cat > /tmp/gitopsi-product-test.yaml << 'EOF'
project:
  name: gitopsi-product-test
  description: "Product Owner Validation Test"

platform: kubernetes
scope: both
gitops_tool: argocd

git:
  url: https://github.com/ihsanmokhlisse/gitopsitest
  branch: main
  push_on_init: true

environments:
  - name: dev
    namespace: product-test-dev
  - name: staging
    namespace: product-test-staging

applications:
  - name: nginx
    type: deployment
    replicas: 2
    image: nginxinc/nginx-unprivileged:1.25-alpine
    port: 8080

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

docs:
  readme: true
  architecture: true

bootstrap:
  enabled: true
  mode: helm
  namespace: argocd
  wait: true
  timeout: 300
  configure_repo: true
  create_app_of_apps: true
EOF

echo -e "${GREEN}✅ Configuration created at /tmp/gitopsi-product-test.yaml${NC}"

echo ""
echo -e "${CYAN}════════════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}🎉 Environment is ready for testing!${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo ""
echo "1. Set your GitHub token (generate at https://github.com/settings/tokens):"
echo ""
echo -e "   ${CYAN}export GITOPSI_GIT_TOKEN=\"ghp_your_token_here\"${NC}"
echo ""
echo "2. Run gitopsi to test the full flow:"
echo ""
echo -e "   ${CYAN}cd $PROJECT_DIR${NC}"
echo -e "   ${CYAN}./bin/gitopsi init --config /tmp/gitopsi-product-test.yaml --bootstrap --verbose${NC}"
echo ""
echo "3. After completion, gitopsi will display:"
echo "   - ArgoCD URL"
echo "   - Admin credentials"
echo "   - Generated file location"
echo ""
echo "4. Access ArgoCD in your browser to verify:"
echo "   - Applications are synced"
echo "   - Infrastructure is deployed"
echo ""
echo "5. When done, run cleanup:"
echo ""
echo -e "   ${CYAN}$0 cleanup${NC}"
echo ""
echo -e "${CYAN}════════════════════════════════════════════════════════════════════════${NC}"

