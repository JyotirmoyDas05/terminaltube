# Makefile for TerminalTube

# Variables
APP_NAME = terminaltube
VERSION = 1.0.0
BUILD_DIR = build
DIST_DIR = dist

# Go build flags
LDFLAGS = -ldflags "-X main.version=$(VERSION)"
BUILD_FLAGS = $(LDFLAGS)

# Cross-compilation targets
PLATFORMS = windows/amd64 darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building $(APP_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT=$(DIST_DIR)/$(APP_NAME)-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then OUTPUT=$$OUTPUT.exe; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build $(BUILD_FLAGS) -o $$OUTPUT .; \
	done

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint code
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet instead"; \
		go vet ./...; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Install to $GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(APP_NAME)..."
	go install $(BUILD_FLAGS) .

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

# Development server (auto-rebuild on changes)
.PHONY: dev
dev:
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running normally instead..."; \
		go run . ; \
	fi

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Create release packages
.PHONY: release
release: build-all
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)/packages
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY=$(APP_NAME)-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then BINARY=$$BINARY.exe; fi; \
		PACKAGE=$(APP_NAME)-v$(VERSION)-$$OS-$$ARCH; \
		mkdir -p $(DIST_DIR)/packages/$$PACKAGE; \
		cp $(DIST_DIR)/$$BINARY $(DIST_DIR)/packages/$$PACKAGE/; \
		cp README.md $(DIST_DIR)/packages/$$PACKAGE/; \
		cp LICENSE $(DIST_DIR)/packages/$$PACKAGE/ 2>/dev/null || true; \
		if [ "$$OS" = "windows" ]; then \
			cd $(DIST_DIR)/packages && zip -r $$PACKAGE.zip $$PACKAGE/; \
		else \
			cd $(DIST_DIR)/packages && tar -czf $$PACKAGE.tar.gz $$PACKAGE/; \
		fi; \
		rm -rf $(DIST_DIR)/packages/$$PACKAGE; \
		echo "Created package: $$PACKAGE"; \
	done

# Check for required tools
.PHONY: check-tools
check-tools:
	@echo "Checking for required tools..."
	@command -v go >/dev/null 2>&1 || (echo "Go is not installed" && exit 1)
	@echo "Go: $$(go version)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint: $$(golangci-lint version)"; \
	else \
		echo "golangci-lint: not installed (optional)"; \
	fi
	@if command -v air >/dev/null 2>&1; then \
		echo "air: $$(air -v)"; \
	else \
		echo "air: not installed (optional)"; \
	fi

# Initialize development environment
.PHONY: init
init: check-tools tidy
	@echo "Initializing development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reloading..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@echo "Development environment ready!"

# Docker build (if Dockerfile exists)
.PHONY: docker-build
docker-build:
	@if [ -f Dockerfile ]; then \
		echo "Building Docker image..."; \
		docker build -t $(APP_NAME):$(VERSION) .; \
		docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest; \
	else \
		echo "Dockerfile not found"; \
	fi

# Help
.PHONY: help
help:
	@echo "TerminalTube Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build for current platform"
	@echo "  build-all     Build for all platforms"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  test-race     Run tests with race detection"
	@echo "  bench         Run benchmarks"
	@echo "  lint          Run linter"
	@echo "  fmt           Format code"
	@echo "  tidy          Tidy dependencies"
	@echo "  install       Install to GOPATH/bin"
	@echo "  clean         Clean build artifacts"
	@echo "  run           Build and run application"
	@echo "  dev           Start development mode with hot reload"
	@echo "  release       Create release packages"
	@echo "  check-tools   Check for required development tools"
	@echo "  init          Initialize development environment"
	@echo "  docker-build  Build Docker image (if Dockerfile exists)"
	@echo "  help          Show this help message"

# Print version
.PHONY: version
version:
	@echo "$(APP_NAME) v$(VERSION)"