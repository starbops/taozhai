# Makefile
.PHONY: build run clean install test test-unit test-coverage test-watch fmt vet lint deps deps-test check build-cross checksums release-build

BINARY_NAME=taozhai
BUILD_DIR=bin
GINKGO=$(shell go env GOPATH)/bin/ginkgo

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-s -w \
	-X github.com/starbops/taozhai/pkg/version.Version=$(VERSION) \
	-X github.com/starbops/taozhai/pkg/version.Commit=$(COMMIT) \
	-X github.com/starbops/taozhai/pkg/version.BuildDate=$(BUILD_DATE)"

# Build targets
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/taozhai

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html
	go clean

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Test targets
test: test-unit

test-unit:
	$(GINKGO) -r -v --cover --coverprofile=coverage.out ./...

test-coverage: test-unit
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-watch:
	$(GINKGO) watch -r ./...

# Code quality targets
fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

check: fmt vet lint test

# Dependencies
deps:
	go mod download
	go mod tidy

deps-test:
	go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Cross-platform build targets
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

build-cross:
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch; \
		echo "Building $$output..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output ./cmd/taozhai; \
	done

checksums:
	@cd $(BUILD_DIR) && \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum $(BINARY_NAME)-* > checksums.txt; \
	else \
		shasum -a 256 $(BINARY_NAME)-* > checksums.txt; \
	fi
	@echo "Generated $(BUILD_DIR)/checksums.txt"

release-build: clean build-cross checksums
