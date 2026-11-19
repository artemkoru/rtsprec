.PHONY: help tools fmt fmt-check vet lint golangci test test-unit test-integration web-build build clean docker-check docker-build docker-up docker-down docker-logs

# Local Docker stack (compose.yaml). HTTP_PORT is exported so compose publishes the same port.
HTTP_PORT ?= 8080
export HTTP_PORT
COMPOSE := docker compose

# Packages checked by lint and tests. CI runs the same targets, so this list is the single source.
GO_PKGS := ./internal/domain ./internal/repository ./internal/service ./internal/app

# Pinned tool versions. CI installs the same version via `make tools`.
GOLANGCI_LINT_VERSION := v2.15.0
GOBIN_DIR := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
GOLANGCI_LINT := $(GOBIN_DIR)/golangci-lint

help:
	@echo "Available targets:"
	@echo "  make tools            - Install pinned tools (golangci-lint $(GOLANGCI_LINT_VERSION))"
	@echo "  make fmt              - Format code with gofmt"
	@echo "  make fmt-check        - Fail if gofmt would change files"
	@echo "  make vet              - Run go vet"
	@echo "  make lint             - fmt-check, vet and golangci-lint (same as CI lint job)"
	@echo "  make test             - Run all tests (unit + integration)"
	@echo "  make test-unit        - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make build            - Build binary (fails on error)"
	@echo "  make docker-build     - Build the backend image (compose)"
	@echo "  make docker-up        - Start backend + S3 in Docker and wait until the API answers"
	@echo "  make docker-down      - Stop the Docker stack (volumes are kept)"
	@echo "  make docker-logs      - Follow backend logs"
	@echo "  make clean            - Clean build artifacts"

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@echo "✓ golangci-lint $(GOLANGCI_LINT_VERSION) installed to $(GOBIN_DIR)"

fmt:
	gofmt -w $(GO_PKGS)
	@echo "✓ Code formatted"

fmt-check:
	@echo "Checking gofmt..."
	@out=$$(gofmt -l $(GO_PKGS)); \
	if [ -n "$$out" ]; then \
		echo "❌ gofmt issues found:"; \
		echo "$$out"; \
		gofmt -d $(GO_PKGS); \
		exit 1; \
	fi
	@echo "✓ All files properly formatted"

vet:
	@echo "Running go vet..."
	go vet $(GO_PKGS)
	@echo "✓ go vet passed"

golangci:
	@test -x $(GOLANGCI_LINT) || (echo "❌ golangci-lint not found at $(GOLANGCI_LINT). Run: make tools"; exit 1)
	@$(GOLANGCI_LINT) version | grep -q "has version $(patsubst v%,%,$(GOLANGCI_LINT_VERSION))" || (echo "❌ wrong golangci-lint version. Expected $(GOLANGCI_LINT_VERSION). Run: make tools"; exit 1)
	@echo "Running golangci-lint $(GOLANGCI_LINT_VERSION)..."
	$(GOLANGCI_LINT) run $(GO_PKGS)
	@echo "✓ golangci-lint passed"

lint: fmt-check vet golangci
	@echo "✓ Lint passed"

test: test-unit test-integration
	@echo "✓ All tests passed"

test-unit:
	@echo "Running unit tests..."
	go test -count=1 -v $(GO_PKGS)
	@echo "✓ Unit tests passed"

test-integration:
	@echo "Checking Docker..."
	@docker info > /dev/null 2>&1 || (echo "❌ Docker is not running. Integration tests start containers: start Docker and retry."; exit 1)
	@echo "Running integration tests..."
	cd integration && go test -count=1 -v ./...
	@echo "✓ Integration tests passed"

# Build the web UI into web/dist; keeps the committed placeholder so the Go embed always has a file.
web-build:
	@echo "Building web UI..."
	find web/dist -mindepth 1 ! -name .gitkeep -delete 2>/dev/null || true
	cd web && npm ci && npm run build
	touch web/dist/.gitkeep
	@test -f web/dist/index.html || (echo "❌ Web UI build did not produce web/dist/index.html"; exit 1)
	@echo "✓ Web UI built"

build: web-build
	@echo "Building binary..."
	go build -o rtsprec ./cmd/rtsprec
	@test -f rtsprec || (echo "❌ Binary not created"; exit 1)
	@echo "✓ Binary created: $$(ls -lh rtsprec | awk '{print $$5}')"

docker-check:
	@docker info > /dev/null 2>&1 || (echo "❌ Docker is not running. Start Docker Desktop (or the Docker daemon) and retry."; exit 1)

docker-build: docker-check
	$(COMPOSE) build backend
	@echo "✓ Backend image built"

docker-up: docker-check
	$(COMPOSE) up -d
	@echo "Waiting for the backend on http://localhost:$(HTTP_PORT) ..."
	@for i in $$(seq 1 60); do \
		code=$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:$(HTTP_PORT)/api/cameras); \
		if [ "$$code" = "401" ]; then echo "✓ Backend is up: http://localhost:$(HTTP_PORT)  (sign in as the AUTH_BASIC_USERNAME user)"; exit 0; fi; \
		sleep 2; \
	done; \
	echo "❌ Backend did not answer within 120s. See: make docker-logs"; exit 1

docker-down: docker-check
	$(COMPOSE) down
	@echo "✓ Docker stack stopped (volumes kept; use 'docker compose down -v' to reset data)"

docker-logs: docker-check
	$(COMPOSE) logs -f backend

clean:
	@echo "Cleaning..."
	rm -f rtsprec
	go clean ./cmd/rtsprec
	@echo "✓ Clean complete"
