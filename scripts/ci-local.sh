#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════╗"
echo "║                    gitopsi - Local CI Pipeline                         ║"
echo "║                                                                        ║"
echo "║  ⚠️  Run this before pushing to catch issues early                      ║"
echo "║                                                                        ║"
echo "║  Usage: ./scripts/ci-local.sh [--quick|--full|--unit|--integration]    ║"
echo "╚════════════════════════════════════════════════════════════════════════╝"
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

MODE="${1:-full}"
step=0
failed=0
passed=0

cd "$(dirname "$0")/.."

run_step() {
    local name="$1"
    shift
    ((step++))
    echo ""
    echo -e "${BLUE}[$step] $name${NC}"
    echo "────────────────────────────────────────────────────────────────────────"
    
    if "$@"; then
        echo -e "${GREEN}✅ $name - Passed${NC}"
        ((passed++))
    else
        echo -e "${RED}❌ $name - Failed${NC}"
        ((failed++))
        if [ "$MODE" != "full" ]; then
            exit 1
        fi
    fi
}

skip_step() {
    ((step++))
    echo ""
    echo -e "${YELLOW}[$step] $1 - Skipped${NC}"
}

print_summary() {
    echo ""
    echo "════════════════════════════════════════════════════════════════════════"
    echo ""
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}╔════════════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║  ✅ All $passed checks passed! Safe to push.                             ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════════════════════════════════════╝${NC}"
    else
        echo -e "${RED}╔════════════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}║  ❌ $failed check(s) failed! Fix issues before pushing.                  ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════════════════════════════════════╝${NC}"
        exit 1
    fi
}

echo -e "${YELLOW}Running in $MODE mode${NC}"
echo ""

# ============================================================================
# Step 1: Go Modules
# ============================================================================
run_step "Verifying Go modules" go mod verify

# ============================================================================
# Step 2: Formatting
# ============================================================================
run_step "Checking code formatting" bash -c '
    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        echo "Files not formatted:"
        echo "$unformatted"
        echo ""
        echo "Run: gofmt -w ."
        exit 1
    fi
    echo "All files properly formatted"
'

# ============================================================================
# Step 3: Go Vet
# ============================================================================
run_step "Running go vet" go vet ./...

# ============================================================================
# Step 4: Linter
# ============================================================================
if command -v golangci-lint &> /dev/null; then
    run_step "Running golangci-lint" golangci-lint run --timeout=5m ./...
else
    skip_step "golangci-lint (not installed - run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)"
fi

# ============================================================================
# Step 5: Unit Tests
# ============================================================================
if [ "$MODE" = "quick" ]; then
    run_step "Running unit tests (quick)" go test -short ./...
else
    run_step "Running unit tests with race detector" bash -c '
        go test -v -race -coverprofile=coverage-unit.out \
            ./internal/auth/... \
            ./internal/bootstrap/... \
            ./internal/cli/... \
            ./internal/cluster/... \
            ./internal/config/... \
            ./internal/environment/... \
            ./internal/generator/... \
            ./internal/git/... \
            ./internal/marketplace/... \
            ./internal/multirepo/... \
            ./internal/operator/... \
            ./internal/organization/... \
            ./internal/output/... \
            ./internal/progress/... \
            ./internal/prompt/... \
            ./internal/security/... \
            ./internal/templates/... \
            ./internal/validate/... \
            ./internal/version/...
    '
fi

# ============================================================================
# Step 6: Integration Tests
# ============================================================================
if [ "$MODE" = "quick" ]; then
    skip_step "Integration tests (use --full to run)"
elif [ "$MODE" = "unit" ]; then
    skip_step "Integration tests (unit-only mode)"
else
    run_step "Running integration tests" bash -c '
        echo "Testing complete init → generation → validation flow..."
        go test -v -race -coverprofile=coverage-integration.out ./internal/integration/... -timeout 10m
    '
fi

# ============================================================================
# Step 7: Regression Tests
# ============================================================================
if [ "$MODE" = "quick" ]; then
    skip_step "Regression tests (use --full to run)"
elif [ "$MODE" = "unit" ]; then
    skip_step "Regression tests (unit-only mode)"
else
    run_step "Running regression tests" bash -c '
        echo "Verifying bug fixes remain fixed..."
        go test -v -race -coverprofile=coverage-regression.out ./internal/regression/... -timeout 5m
    '
fi

# ============================================================================
# Step 8: Coverage Check
# ============================================================================
if [ -f coverage-unit.out ]; then
    run_step "Checking test coverage" bash -c '
        coverage=$(go tool cover -func=coverage-unit.out | tail -1 | awk "{print \$3}" | sed "s/%//")
        threshold=40
        echo "Coverage: ${coverage}%"
        echo "Threshold: ${threshold}%"
        if (( $(echo "$coverage < $threshold" | bc -l) )); then
            echo "⚠️  Coverage ${coverage}% is below threshold ${threshold}%"
            exit 1
        fi
        echo "✅ Coverage meets threshold"
    '
fi

# ============================================================================
# Step 9: Build
# ============================================================================
if [ "$MODE" = "integration" ]; then
    skip_step "Build (integration-only mode)"
else
    run_step "Building binaries" bash -c '
        mkdir -p bin
        
        # Determine current platform
        current_os=$(go env GOOS)
        current_arch=$(go env GOARCH)
        
        if [ "$MODE" = "quick" ]; then
            # Quick mode: only build for current platform
            echo "Building for current platform: ${current_os}/${current_arch}..."
            output="bin/gitopsi"
            [ "$current_os" = "windows" ] && output="${output}.exe"
            go build -ldflags="-s -w" -o "$output" ./cmd/gitopsi
            echo "Built: $output"
        else
            # Full mode: build all platforms
            for os in linux darwin windows; do
                for arch in amd64 arm64; do
                    [ "$os" = "windows" ] && [ "$arch" = "arm64" ] && continue
                    output="bin/gitopsi-${os}-${arch}"
                    [ "$os" = "windows" ] && output="${output}.exe"
                    echo "Building ${os}/${arch}..."
                    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags="-s -w" -o "$output" ./cmd/gitopsi
                done
            done
            echo "Built all platform binaries"
        fi
    '
fi

# ============================================================================
# Step 10: E2E Smoke Test
# ============================================================================
if [ "$MODE" = "full" ]; then
    run_step "E2E smoke test" bash -c '
        # Find the binary for current platform
        binary="bin/gitopsi"
        if [ ! -f "$binary" ]; then
            binary="bin/gitopsi-$(go env GOOS)-$(go env GOARCH)"
        fi
        
        if [ ! -f "$binary" ]; then
            echo "Binary not found, skipping E2E test"
            exit 0
        fi
        
        chmod +x "$binary"
        
        # Test version
        echo "Testing version command..."
        "$binary" version
        
        # Test generation
        echo "Testing project generation..."
        tmpdir=$(mktemp -d)
        cd "$tmpdir"
        
        cat > test-config.yaml << EOF
project:
  name: ci-local-test
platform: kubernetes
gitops_tool: argocd
environments:
  - name: dev
infrastructure:
  namespaces: true
docs:
  readme: true
EOF
        
        "$OLDPWD/$binary" init --config test-config.yaml
        
        # Verify output
        if [ ! -f ci-local-test/README.md ]; then
            echo "Generation failed - README.md not found"
            exit 1
        fi
        
        echo "E2E smoke test passed"
        cd -
        rm -rf "$tmpdir"
    '
else
    skip_step "E2E smoke test (use --full to run)"
fi

# ============================================================================
# Summary
# ============================================================================
print_summary

# Cleanup coverage files
rm -f coverage-unit.out coverage-integration.out coverage-regression.out 2>/dev/null || true
