#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║           gitopsi - Local CI Pipeline                      ║"
echo "║                                                            ║"
echo "║  ⚠️  Run this before pushing to catch issues early         ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

step=1
total=7

run_step() {
    echo ""
    echo -e "${YELLOW}[$step/$total] $1${NC}"
    echo "─────────────────────────────────────────────────"
    shift
    if "$@"; then
        echo -e "${GREEN}✅ Passed${NC}"
    else
        echo -e "${RED}❌ Failed${NC}"
        exit 1
    fi
    ((step++))
}

cd "$(dirname "$0")/.."

run_step "Verifying Go modules" go mod verify

run_step "Checking formatting" bash -c '
    if [ -n "$(gofmt -l .)" ]; then
        echo "Files not formatted:"
        gofmt -l .
        exit 1
    fi
'

run_step "Running go vet" go vet ./...

run_step "Running linter" bash -c '
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./...
    else
        echo "golangci-lint not installed, skipping..."
    fi
'

run_step "Running tests with race detector" go test -v -race -coverprofile=coverage.out ./...

run_step "Checking coverage" bash -c '
    coverage=$(go tool cover -func=coverage.out | tail -1 | awk "{print \$3}" | sed "s/%//")
    threshold=70
    echo "Coverage: ${coverage}%"
    if [ $(echo "$coverage < $threshold" | bc -l) -eq 1 ]; then
        echo "Coverage ${coverage}% is below threshold ${threshold}%"
        exit 1
    fi
'

run_step "Building binaries" bash -c '
    for os in linux darwin windows; do
        for arch in amd64 arm64; do
            [ "$os" = "windows" ] && [ "$arch" = "arm64" ] && continue
            output="bin/gitopsi-${os}-${arch}"
            [ "$os" = "windows" ] && output="${output}.exe"
            echo "Building ${os}/${arch}..."
            GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o "$output" ./cmd/gitopsi
        done
    done
'

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  ${GREEN}✅ All checks passed! Safe to push.${NC}                       ║"
echo "╚════════════════════════════════════════════════════════════╝"

