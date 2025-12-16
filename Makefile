BINARY_NAME=gitopsi
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/ihsanmokhlisse/gitopsi/internal/cli.Version=$(VERSION) -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Commit=$(COMMIT) -X github.com/ihsanmokhlisse/gitopsi/internal/cli.BuildDate=$(BUILD_DATE)"

CONTAINER_ENGINE=podman
IMAGE_NAME=gitopsi
IMAGE_TAG=dev

.PHONY: all build test clean lint fmt container-build container-test container-shell container-run

all: fmt lint test build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/gitopsi

test:
	go test -v -race -coverprofile=coverage.out ./...

test-short:
	go test -short -v ./...

coverage: test
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

clean:
	rm -rf bin/ coverage.out coverage.html

container-build:
	$(CONTAINER_ENGINE) build -t $(IMAGE_NAME):$(IMAGE_TAG) --target dev -f Containerfile .

container-test: container-build
	$(CONTAINER_ENGINE) run --rm -v $(PWD):/app:Z $(IMAGE_NAME):$(IMAGE_TAG) make test

container-shell: container-build
	$(CONTAINER_ENGINE) run --rm -it -v $(PWD):/app:Z $(IMAGE_NAME):$(IMAGE_TAG) /bin/bash

container-run: container-build
	$(CONTAINER_ENGINE) run --rm -it -v $(PWD):/workspace:Z $(IMAGE_NAME):$(IMAGE_TAG) $(ARGS)

container-build-release:
	$(CONTAINER_ENGINE) build -t $(IMAGE_NAME):$(VERSION) --target runtime -f Containerfile .

mod-download:
	go mod download

mod-tidy:
	go mod tidy

install: build
	@echo "⚠️  WARNING: Direct installation is discouraged. Use container-build instead."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

help:
	@echo "gitopsi Makefile"
	@echo ""
	@echo "⚠️  TESTING RULE: Use Podman containers, not direct execution"
	@echo ""
	@echo "Container targets (RECOMMENDED):"
	@echo "  container-build   Build development container"
	@echo "  container-test    Run tests in container"
	@echo "  container-shell   Interactive dev shell"
	@echo "  container-run     Run gitopsi in container"
	@echo ""
	@echo "Local targets (for CI only):"
	@echo "  build             Build binary"
	@echo "  test              Run tests"
	@echo "  lint              Run linter"
	@echo "  fmt               Format code"
	@echo "  clean             Clean build artifacts"
