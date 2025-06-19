.PHONY: build test lint clean install

# Build the application
build:
	go build -o bin/goldfish ./cmd/goldfish

# Run tests
test:
	gotestsum --format testname ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/

# Install locally
install:
	go install ./cmd/goldfish

# Run all checks
check: lint test

# Development build and test
dev: build test

# Release build
release:
	goreleaser release --clean