.PHONY: build run test clean install

BINARY=gitopsi
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/gitopsi

run: build
	./bin/$(BINARY)

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/

fmt:
	go fmt ./...
	goimports -w .

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod download

