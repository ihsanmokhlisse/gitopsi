BINARY_NAME=gitopsi
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Version=$(VERSION) -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Commit=$(COMMIT) -X github.com/ihsanmokhlisse/gitopsi/internal/cli.BuildDate=$(BUILD_DATE)"

CONTAINER_ENGINE=podman
IMAGE_NAME=gitopsi
IMAGE_TAG=dev

COVERAGE_THRESHOLD=70

PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build test clean lint fmt check help
.PHONY: container-build container-test container-shell container-run
.PHONY: ci-local pre-push release

all: check build

##@ Development

build: ## Build binary for current platform
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/gitopsi

build-all: ## Build binaries for all platforms
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=bin/$(BINARY_NAME)-$$os-$$arch; \
		[ "$$os" = "windows" ] && output=$$output.exe; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output ./cmd/gitopsi; \
	done

run: build ## Build and run
	./bin/$(BINARY_NAME) $(ARGS)

##@ Testing (Container-Based - REQUIRED)

container-build: ## Build development container
	$(CONTAINER_ENGINE) build -t $(IMAGE_NAME):$(IMAGE_TAG) --target dev -f Containerfile .

container-test: container-build ## Run all tests in container
	$(CONTAINER_ENGINE) run --rm -v $$(pwd):/app:Z $(IMAGE_NAME):$(IMAGE_TAG) go test -v -race -coverprofile=coverage.out ./...

container-shell: container-build ## Interactive development shell
	$(CONTAINER_ENGINE) run --rm -it -v $$(pwd):/app:Z $(IMAGE_NAME):$(IMAGE_TAG) /bin/bash

container-run: container-build ## Run command in container (use ARGS="...")
	$(CONTAINER_ENGINE) run --rm -v $$(pwd):/app:Z $(IMAGE_NAME):$(IMAGE_TAG) $(ARGS)

##@ Testing (Local - for CI only)

test: ## Run unit tests
	go test -v -race ./...

test-short: ## Run short tests only
	go test -short -v ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tail -1

test-coverage-html: test-coverage ## Generate HTML coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

coverage-check: test-coverage ## Check coverage meets threshold
	@coverage=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$coverage < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$coverage% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$coverage% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

test-e2e: ## Run end-to-end tests
	go test -v -tags=e2e ./test/e2e/...

test-verbose: ## Run tests with verbose output
	go test -v -race -count=1 ./...

##@ Code Quality

fmt: ## Format code
	gofmt -s -w .
	@command -v goimports >/dev/null && goimports -w . || true

fmt-check: ## Check code formatting
	@test -z "$$(gofmt -l .)" || (echo "❌ Code not formatted. Run 'make fmt'" && gofmt -l . && exit 1)
	@echo "✅ Code formatting OK"

imports-check: ## Check imports
	@command -v goimports >/dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	@test -z "$$(goimports -l .)" || (echo "❌ Imports not sorted. Run 'make fmt'" && goimports -l . && exit 1)
	@echo "✅ Imports OK"

lint: ## Run linter
	@command -v golangci-lint >/dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix ./...

security: ## Run security scanner
	@command -v gosec >/dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -quiet ./...

vuln: ## Check for vulnerabilities
	@command -v govulncheck >/dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

vet: ## Run go vet
	go vet ./...

##@ Pre-commit & CI

check: fmt-check vet lint test ## Run all checks (format, vet, lint, test)
	@echo "✅ All checks passed"

pre-push: ## Run before pushing (full validation)
	@echo "🔍 Running pre-push checks..."
	@make fmt-check
	@make vet
	@make lint
	@make test
	@make build
	@echo "✅ Pre-push checks passed"

ci-local: ## Simulate full CI pipeline locally
	@echo "🔄 Simulating CI pipeline..."
	@echo "\n📋 Step 1/6: Verify modules"
	go mod verify
	@echo "\n📋 Step 2/6: Check formatting"
	@make fmt-check
	@echo "\n📋 Step 3/6: Run vet"
	@make vet
	@echo "\n📋 Step 4/6: Run linter"
	@make lint
	@echo "\n📋 Step 5/6: Run tests with race detector"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "\n📋 Step 6/6: Build all platforms"
	@make build-all
	@echo "\n✅ CI simulation complete"

##@ Golden Files

golden-update: ## Update golden test files
	@echo "⚠️  Updating golden files..."
	go test -v -update ./...

golden-check: ## Verify golden files match
	go test -v ./... -run Golden

##@ Release

release-dry: ## Dry run release (no publish)
	@command -v goreleaser >/dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser release --snapshot --clean --skip=publish

release: ## Create release (requires GITHUB_TOKEN)
	goreleaser release --clean

##@ Container Images

container-build-release: ## Build release container
	$(CONTAINER_ENGINE) build -t $(IMAGE_NAME):$(VERSION) --target runtime -f Containerfile .

container-push: container-build-release ## Push container to registry
	$(CONTAINER_ENGINE) push $(IMAGE_NAME):$(VERSION)

##@ Dependencies

mod-download: ## Download dependencies
	go mod download

mod-tidy: ## Tidy dependencies
	go mod tidy

mod-verify: ## Verify dependencies
	go mod verify

mod-update: ## Update dependencies
	go get -u ./...
	go mod tidy

##@ Git Hooks

hooks-install: ## Install git hooks
	@echo "Installing git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/bash\nmake pre-push' > .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "✅ Git hooks installed"

hooks-uninstall: ## Remove git hooks
	@rm -f .git/hooks/pre-push
	@echo "✅ Git hooks removed"

##@ Cleanup

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html

clean-all: clean ## Clean everything including containers
	$(CONTAINER_ENGINE) rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
	$(CONTAINER_ENGINE) rmi $(IMAGE_NAME):$(VERSION) 2>/dev/null || true

##@ Help

help: ## Show this help
	@echo "gitopsi Makefile"
	@echo ""
	@echo "⚠️  TESTING RULE: All testing MUST use Podman containers"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
