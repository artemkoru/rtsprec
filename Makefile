.PHONY: help tools fmt fmt-check vet lint golangci test test-unit test-integration build clean

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

test-integration: test-unit
	@echo "Integration tests (currently same as unit tests)"
	@echo "✓ Integration tests passed"

build:
	@echo "Building binary..."
	go build -o rtsprec ./cmd/rtsprec
	@test -f rtsprec || (echo "❌ Binary not created"; exit 1)
	@echo "✓ Binary created: $$(ls -lh rtsprec | awk '{print $$5}')"

clean:
	@echo "Cleaning..."
	rm -f rtsprec
	go clean ./cmd/rtsprec
	@echo "✓ Clean complete"
