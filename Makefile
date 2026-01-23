# Makefile
.PHONY: build run clean install test test-unit test-coverage test-watch fmt vet lint deps deps-test check

BINARY_NAME=taozhai
BUILD_DIR=bin
GINKGO=$(shell go env GOPATH)/bin/ginkgo

# Build targets
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/taozhai

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
