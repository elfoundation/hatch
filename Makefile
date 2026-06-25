.PHONY: all build test lint fmt vet clean help

all: lint test build

## build: Build the hatch binary
build:
	CGO_ENABLED=0 go build -o hatch ./cmd/hatch

## test: Run all tests with race detection and coverage
test:
	go test ./... -v -race -coverprofile=coverage.out

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code with gofmt
fmt:
	gofmt -s -w .

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -f hatch coverage.out

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
