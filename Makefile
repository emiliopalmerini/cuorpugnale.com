# ==============================================================================
# Cuorpugnale - Go + HTMX
# ==============================================================================

# Build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go settings
GO          := go
GOFLAGS     := -v
CGO_ENABLED := 0
LDFLAGS     := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Directories
BIN_DIR     := bin
TMP_DIR     := tmp

# Binary
SERVER      := $(BIN_DIR)/server

# ==============================================================================
# MAIN TARGETS
# ==============================================================================

.PHONY: all
all: build ## Build everything (default)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ==============================================================================
# BUILD
# ==============================================================================

.PHONY: build
build: fmt vet ## Build the server binary
	@echo "Building server..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(SERVER) ./cmd/server
	@echo "Build complete: $(SERVER)"

# ==============================================================================
# RUN
# ==============================================================================

.PHONY: run
run: build ## Build and run the server binary
	@echo "Starting server..."
	@$(SERVER)

COUNTDOWN_DELAY ?= 10s

.PHONY: run-countdown
run-countdown: build ## Run locally with a countdown, defaulting to 10 seconds
	@echo "Starting server with a $(COUNTDOWN_DELAY) countdown..."
	@TRAILER_LAUNCH_DELAY=$(COUNTDOWN_DELAY) $(SERVER)

# ==============================================================================
# DEVELOPMENT (hot-reload with air)
# ==============================================================================

.PHONY: dev
dev: ## Run server with hot-reload (air)
	@mkdir -p tmp
	@air -c .air.toml

# ==============================================================================
# QUALITY
# ==============================================================================

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

# ==============================================================================
# TESTING
# ==============================================================================

.PHONY: test
test: fmt vet ## Run tests
	@echo "Running tests..."
	@$(GO) test -v ./...

.PHONY: test-short
test-short: ## Run tests (short mode)
	@$(GO) test -short -v ./...

.PHONY: test-cover
test-cover: fmt vet ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GO) test -coverprofile=coverage.out -covermode=atomic ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: test-cover-func
test-cover-func: ## Show coverage by function
	@$(GO) test -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out

.PHONY: bench
bench: ## Run benchmarks
	@$(GO) test -bench=. -benchmem ./...

# ==============================================================================
# DOCKER
# ==============================================================================

.PHONY: docker-build
docker-build: ## Build Docker images
	@docker compose build

.PHONY: docker-up
docker-up: ## Start Docker containers
	@docker compose up -d

.PHONY: docker-down
docker-down: ## Stop Docker containers
	@docker compose down

.PHONY: docker-logs
docker-logs: ## View Docker logs
	@docker compose logs -f

.PHONY: docker-clean
docker-clean: ## Remove Docker containers and volumes
	@docker compose down -v --remove-orphans

# ==============================================================================
# CLEANUP
# ==============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -rf $(TMP_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean!"

.PHONY: clean-all
clean-all: clean docker-clean ## Clean everything including Docker

# ==============================================================================
# INFO
# ==============================================================================

.PHONY: version
version: ## Show version info
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

.PHONY: info
info: ## Show project info
	@echo "Go version: $(shell $(GO) version)"
