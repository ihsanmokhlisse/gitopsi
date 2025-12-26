#!/bin/bash
# Developer setup script for gitopsi
# Installs all necessary tools for development

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    gitopsi Developer Setup                             ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

cd "$(dirname "$0")/.."

check_command() {
    if command -v "$1" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 found"
        return 0
    else
        echo -e "${RED}✗${NC} $1 not found"
        return 1
    fi
}

install_go_tools() {
    echo ""
    echo -e "${YELLOW}Installing Go tools...${NC}"
    echo "────────────────────────────────────────────────────────────────────────"
    
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest
    
    echo -e "${GREEN}✓ Go tools installed${NC}"
}

install_pre_commit() {
    echo ""
    echo -e "${YELLOW}Setting up pre-commit hooks...${NC}"
    echo "────────────────────────────────────────────────────────────────────────"
    
    if ! check_command pre-commit; then
        echo "Installing pre-commit..."
        if check_command pip3; then
            pip3 install pre-commit
        elif check_command pip; then
            pip install pre-commit
        elif check_command brew; then
            brew install pre-commit
        else
            echo -e "${RED}Cannot install pre-commit. Please install Python/pip or Homebrew.${NC}"
            return 1
        fi
    fi
    
    echo "Installing pre-commit hooks..."
    pre-commit install
    pre-commit install --hook-type commit-msg
    
    echo -e "${GREEN}✓ Pre-commit hooks installed${NC}"
}

verify_setup() {
    echo ""
    echo -e "${YELLOW}Verifying setup...${NC}"
    echo "────────────────────────────────────────────────────────────────────────"
    
    local all_good=true
    
    check_command go || all_good=false
    check_command git || all_good=false
    check_command gofmt || all_good=false
    check_command goimports || all_good=false
    check_command golangci-lint || all_good=false
    check_command gosec || all_good=false
    check_command pre-commit || all_good=false
    
    # Optional tools
    echo ""
    echo "Optional tools:"
    check_command docker || echo "  (needed for container builds)"
    check_command podman || echo "  (alternative to docker)"
    check_command kubectl || echo "  (needed for E2E tests)"
    check_command oc || echo "  (needed for OpenShift tests)"
    
    if [ "$all_good" = true ]; then
        return 0
    else
        return 1
    fi
}

run_initial_checks() {
    echo ""
    echo -e "${YELLOW}Running initial checks...${NC}"
    echo "────────────────────────────────────────────────────────────────────────"
    
    echo "Downloading Go modules..."
    go mod download
    
    echo "Verifying modules..."
    go mod verify
    
    echo "Running pre-commit on all files..."
    pre-commit run --all-files || true
    
    echo -e "${GREEN}✓ Initial checks complete${NC}"
}

print_summary() {
    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${GREEN}✅ Developer setup complete!${NC}"
    echo ""
    echo "Available commands:"
    echo ""
    echo "  make build          Build gitopsi binary"
    echo "  make test           Run all tests"
    echo "  make lint           Run linter"
    echo "  make ci-local       Run full local CI pipeline"
    echo ""
    echo "Pre-commit hooks will automatically run on every commit."
    echo ""
    echo "To skip hooks temporarily:"
    echo "  git commit --no-verify -m 'message'"
    echo "  SKIP=golangci-lint git commit -m 'message'"
    echo ""
    echo "To run hooks manually:"
    echo "  pre-commit run --all-files"
    echo "  pre-commit run golangci-lint --all-files"
    echo ""
}

# Main
echo -e "${YELLOW}Checking prerequisites...${NC}"
echo "────────────────────────────────────────────────────────────────────────"

if ! check_command go; then
    echo -e "${RED}Go is required. Please install Go 1.23+ from https://go.dev/dl/${NC}"
    exit 1
fi

if ! check_command git; then
    echo -e "${RED}Git is required. Please install Git.${NC}"
    exit 1
fi

install_go_tools
install_pre_commit
verify_setup
run_initial_checks
print_summary





