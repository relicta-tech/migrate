# Makefile for relicta-migrate

BINARY := migrate
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build clean test lint install release-build

all: build

# Build for current platform
build:
	@echo "Building $(BINARY)..."
	go build $(LDFLAGS) -o $(BINARY) .

# Install locally
install:
	@echo "Installing $(BINARY)..."
	go install $(LDFLAGS) .

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep total

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BINARY) dist/ coverage.out

# Build for all platforms (for releases)
release-build: clean
	@echo "Building release binaries..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		OUTPUT="dist/$(BINARY)_$${GOOS}_$${GOARCH}"; \
		if [ "$$GOOS" = "windows" ]; then OUTPUT="$$OUTPUT.exe"; fi; \
		echo "  Building $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o $$OUTPUT .; \
	done
	@echo "Creating archives..."
	@cd dist && for f in $(BINARY)_linux_* $(BINARY)_darwin_*; do \
		[ -f "$$f" ] && tar -czvf "$${f}.tar.gz" "$$f" && rm "$$f"; \
	done
	@cd dist && for f in $(BINARY)_windows_*.exe; do \
		[ -f "$$f" ] && zip "$${f%.exe}.zip" "$$f" && rm "$$f"; \
	done
	@echo "Creating checksums..."
	@cd dist && shasum -a 256 *.tar.gz *.zip > checksums.txt
	@echo "Done! Artifacts in dist/"
	@ls -lh dist/

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w .
	goimports -w .

# Run all checks (for CI)
check: fmt lint test build
	@echo "All checks passed!"
