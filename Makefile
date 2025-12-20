BINARY_NAME=gitopsi
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Version=$(VERSION) -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Commit=$(COMMIT) -X github.com/ihsanmokhlisse/gitopsi/internal/cli.BuildDate=$(BUILD_DATE)"

CONTAINER_ENGINE=podman
IMAGE_NAME=gitopsi
IMAGE_TAG=dev

COVERAGE_THRESHOLD=40

PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build test clean lint fmt check help
.PHONY: container-build container-test container-shell container-run
.PHONY: ci-local pre-push release
.PHONY: setup pre-commit pre-commit-all pre-commit-install

all: check build

##@ Development Setup

setup: ## Initial developer setup (installs tools and hooks)
	@chmod +x scripts/setup-dev.sh
	@./scripts/setup-dev.sh

pre-commit-install: ## Install pre-commit hooks
	@command -v pre-commit >/dev/null || pip3 install pre-commit
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "✅ Pre-commit hooks installed"

pre-commit: ## Run pre-commit on staged files
	pre-commit run

pre-commit-all: ## Run pre-commit on all files
	pre-commit run --all-files

pre-commit-update: ## Update pre-commit hooks to latest versions
	pre-commit autoupdate

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
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) -o $$output ./cmd/gitopsi; \
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

##@ Testing

test: ## Run all tests
	go test -v -race ./...

test-unit: ## Run unit tests only
	go test -v -race ./internal/auth/... ./internal/bootstrap/... ./internal/cli/... \
		./internal/cluster/... ./internal/config/... ./internal/environment/... \
		./internal/generator/... ./internal/git/... ./internal/marketplace/... \
		./internal/multirepo/... ./internal/operator/... ./internal/organization/... \
		./internal/output/... ./internal/progress/... ./internal/prompt/... \
		./internal/security/... ./internal/templates/... ./internal/validate/... \
		./internal/version/...

test-integration: ## Run integration tests
	go test -v -race ./internal/integration/... -timeout 10m

test-regression: ## Run regression tests for bug fixes
	go test -v -race ./internal/regression/... -timeout 5m

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

test-all: test-unit test-integration test-regression ## Run all test types
	@echo "✅ All tests passed"

##@ Code Quality

fmt: ## Format code
	gofmt -s -w .
	@command -v goimports >/dev/null && goimports -w -local github.com/ihsanmokhlisse/gitopsi . || true

fmt-check: ## Check code formatting
	@test -z "$$(gofmt -l .)" || (echo "❌ Code not formatted. Run 'make fmt'" && gofmt -l . && exit 1)
	@echo "✅ Code formatting OK"

imports-check: ## Check imports
	@command -v goimports >/dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	@test -z "$$(goimports -l .)" || (echo "❌ Imports not sorted. Run 'make fmt'" && goimports -l . && exit 1)
	@echo "✅ Imports OK"

lint: ## Run linter
	@command -v golangci-lint >/dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m ./...

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix ./...

security: ## Run security scanner
	@command -v gosec >/dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -exclude-dir=internal/git -exclude=G204 ./...

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
	@make test-unit
	@make build
	@echo "✅ Pre-push checks passed"

ci-local: ## Simulate full CI pipeline locally
	@chmod +x scripts/ci-local.sh
	@./scripts/ci-local.sh --full

ci-quick: ## Quick CI checks (faster)
	@chmod +x scripts/ci-local.sh
	@./scripts/ci-local.sh --quick

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

##@ Cleanup

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html *.out

clean-all: clean ## Clean everything including containers
	$(CONTAINER_ENGINE) rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
	$(CONTAINER_ENGINE) rmi $(IMAGE_NAME):$(VERSION) 2>/dev/null || true

##@ Help

help: ## Show this help
	@echo "gitopsi Makefile"
	@echo ""
	@echo "Quick Start:"
	@echo "  make setup           Run initial developer setup"
	@echo "  make pre-commit-all  Run all pre-commit hooks"
	@echo "  make ci-local        Run full local CI"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
